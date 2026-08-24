/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package efficiency

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	cacontext "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/status"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/clustersnapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/scheduling"
	caerrors "sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
)

type fakeActuationStatus struct {
	recentEvictions []*apiv1.Pod
}

// RecentEvictions implements the scaledown.ActuationStatus interface, demanded by UpdateClusterState.
func (f *fakeActuationStatus) RecentEvictions() []*apiv1.Pod {
	return f.recentEvictions
}

// RegisterEviction implements the scaledown.ActuationStatus interface, demanded by UpdateClusterState.
func (f *fakeActuationStatus) RegisterEviction(pod *apiv1.Pod) {
	f.recentEvictions = append(f.recentEvictions, pod)
}

// DeletionsInProgress implements the scaledown.ActuationStatus interface, demanded by UpdateClusterState.
func (f *fakeActuationStatus) DeletionsInProgress() ([]string, []string) {
	return nil, nil
}

// DeletionsCount implements the scaledown.ActuationStatus interface, demanded by UpdateClusterState.
func (f *fakeActuationStatus) DeletionsCount(nodeGroup string) int {
	return 0
}

type fakeActuator struct {
	autoscalingCtx  cacontext.AutoscalingContext
	actuationStatus *fakeActuationStatus
}

// NewFakeActuator returns new instance of fakeActuator.
func NewFakeActuator(ctx cacontext.AutoscalingContext, status *fakeActuationStatus) *fakeActuator {
	return &fakeActuator{
		autoscalingCtx:  ctx,
		actuationStatus: status,
	}
}

// startDeletion "manually" removes nodes and ensures pods rescheduling.
// Clustersnapshot.fork/commit/revert not present as Pod rescheduling should succeed if nodes were marked as candidates (consistent with planner's decisions on candidate nodes).
func (f *fakeActuator) startDeletion(empty, needDrain []*apiv1.Node) (status.ScaleDownResult, []*status.ScaleDownNode, caerrors.AutoscalerError) {
	if len(empty) == 0 && len(needDrain) == 0 {
		return status.ScaleDownNoNodeDeleted, nil, nil
	}

	var scaledDownNodes []*status.ScaleDownNode
	for _, node := range empty {
		err := f.autoscalingCtx.ClusterSnapshot.RemoveNodeInfo(context.TODO(), node.Name)
		if err != nil {
			return status.ScaleDownError, nil, nil
		}
		scaledDownNodes = append(scaledDownNodes, &status.ScaleDownNode{
			Node:        node,
			EvictedPods: nil,
		})
	}

	schedSimulator := scheduling.NewHintingSimulator()

	for _, n := range needDrain {
		nodeInfo, err := f.autoscalingCtx.ClusterSnapshot.GetNodeInfo(n.Name)
		if err != nil {
			return status.ScaleDownError, nil, caerrors.ToAutoscalerError(caerrors.TransientError, err)
		}

		var podsToReschedule []*apiv1.Pod
		for _, podInfo := range nodeInfo.Pods() {
			podInfo.Pod.Spec.NodeName = ""
			podsToReschedule = append(podsToReschedule, podInfo.Pod)
		}

		err = f.autoscalingCtx.ClusterSnapshot.RemoveNodeInfo(context.TODO(), n.Name)
		if err != nil {
			return status.ScaleDownError, nil, caerrors.ToAutoscalerError(caerrors.TransientError, err)
		}

		// If there are Pods to be rescheduled, try rescheduling them.
		if len(podsToReschedule) > 0 {
			schedulingResult, err := schedSimulator.TrySchedulePods(context.Background(), f.autoscalingCtx.ClusterSnapshot, podsToReschedule, false,
				clustersnapshot.SchedulingOptions{
					// Currently evaluated node to drain is not considered as scheduling target as it was already removed from cluster snapshot.
					IsNodeAcceptable: scheduling.ScheduleAnywhere,
				})
			if err != nil {
				return status.ScaleDownError, nil, caerrors.ToAutoscalerError(caerrors.TransientError, err)
			}

			scheduledPods := sets.Set[string]{}
			for _, s := range schedulingResult.Statuses {
				scheduledPods.Insert(fmt.Sprintf("%s/%s", s.Pod.Namespace, s.Pod.Name))
			}

			// If Pod is not present in scheduledPods, it was not successfully rescheduled.
			// This is an error because Planner's decision on nodes to drain should be successful in static cluster snapshot.
			for _, pod := range podsToReschedule {
				if !scheduledPods.Has(fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)) {
					return status.ScaleDownError, nil, caerrors.NewAutoscalerError(caerrors.InternalError, "failed to reschedule all Pods in actuator")
				}
			}
		}
		scaledDownNodes = append(scaledDownNodes, &status.ScaleDownNode{
			Node:        n,
			EvictedPods: podsToReschedule,
		})
	}
	return status.ScaleDownNodeDeleteStarted, scaledDownNodes, nil
}
