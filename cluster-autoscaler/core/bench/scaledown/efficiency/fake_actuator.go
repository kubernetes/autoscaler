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
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	caerrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
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

// fakeActuator may embed scaledown.Actuator interface in future to ensure all methods are implemented.
type fakeActuator struct {
	autoscalingCtx  context.AutoscalingContext
	t               testing.TB
	actuationStatus *fakeActuationStatus
}

// NewFakeActuator returns new instance of fakeActuator.
func NewFakeActuator(ctx context.AutoscalingContext, status *fakeActuationStatus, t testing.TB) *fakeActuator {
	return &fakeActuator{
		autoscalingCtx:  ctx,
		t:               t,
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
		err := f.autoscalingCtx.ClusterSnapshot.RemoveNodeInfo(node.Name)
		if err != nil {
			f.t.Logf("cannot remove node info, err: %v", err.Error())
			return status.ScaleDownError, nil, nil
		}
		scaledDownNodes = append(scaledDownNodes, &status.ScaleDownNode{
			Node:        node,
			EvictedPods: nil,
		})
	}

	needDrain = needDrain[:1]
	schedSimulator := scheduling.NewHintingSimulator()

	for _, n := range needDrain {
		nodeInfo, err := f.autoscalingCtx.ClusterSnapshot.GetNodeInfo(n.Name)
		assert.NoError(f.t, err)

		var podsToReschedule []*apiv1.Pod
		for _, podInfo := range nodeInfo.Pods() {
			podInfo.Pod.Spec.NodeName = ""
			podsToReschedule = append(podsToReschedule, podInfo.Pod)
		}

		err = f.autoscalingCtx.ClusterSnapshot.RemoveNodeInfo(n.Name)
		if err != nil {
			f.t.Errorf("error removing node info from cluster snapshot, error: %s", err.Error())
		}

		if len(podsToReschedule) > 0 {
			statuses, _, err := schedSimulator.TrySchedulePods(f.autoscalingCtx.ClusterSnapshot, podsToReschedule, false,
				clustersnapshot.SchedulingOptions{
					IsNodeAcceptable: scheduling.ScheduleAnywhere,
					//NodeOrdering:     clustersnapshot.NewLastIndexOrderMapping(0),
				})
			if err != nil {
				f.t.Errorf("cannot scale down, an unexpected error occurred: %s", err.Error())
			}

			scheduledPods := make(map[string]bool)
			for _, s := range statuses {
				scheduledPods[s.Pod.Namespace+"/"+s.Pod.Name] = true
			}

			// Any pod not in scheduledPods was not successful and is registered as eviction.
			// Theoretically, all Pods evicted from a drained node could be registered as an eviction.
			// Depends on the interpretation of eviction.
			for _, pod := range podsToReschedule {
				if !scheduledPods[pod.Namespace+"/"+pod.Name] {
					f.t.Logf("Pod %s/%s failed to reschedule!", pod.Namespace, pod.Name)
					if f.actuationStatus != nil {
						f.actuationStatus.RegisterEviction(pod)
					}
				}
			}

			if len(statuses) != len(podsToReschedule) {
				f.t.Logf("failed to reschedule all pods from node %s: scheduled %d out of %d pods", n.Name, len(statuses), len(podsToReschedule))
				continue
			}
		}
		scaledDownNodes = append(scaledDownNodes, &status.ScaleDownNode{
			Node:        n,
			EvictedPods: podsToReschedule,
		})
	}
	return status.ScaleDownNodeDeleteStarted, scaledDownNodes, nil
}
