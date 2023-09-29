/*
Copyright 2015 The Kubernetes Authors.

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

package simulator

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// GetPodsToMove returns a list of pods that should be moved elsewhere
// and a list of DaemonSet pods that should be evicted if the node
// is drained. Raises error if there is an unreplicated pod.
// Based on kubectl drain code. If listers is nil it makes an assumption that RC, DS, Jobs and RS were deleted
// along with their pods (no abandoned pods with dangling created-by annotation).
// If listers is not nil it checks whether RC, DS, Jobs and RS that created these pods
// still exist.
// TODO(x13n): Rewrite GetPodsForDeletionOnNodeDrain into a set of DrainabilityRules.
func GetPodsToMove(nodeInfo *schedulerframework.NodeInfo, deleteOptions options.NodeDeleteOptions, drainabilityRules rules.Rules, listers kube_util.ListerRegistry, remainingPdbTracker pdb.RemainingPdbTracker, timestamp time.Time) (pods []*apiv1.Pod, daemonSetPods []*apiv1.Pod, blockingPod *drain.BlockingPod, err error) {
	var drainPods, drainDs []*apiv1.Pod
	if drainabilityRules == nil {
		drainabilityRules = rules.Default()
	}
	if remainingPdbTracker == nil {
		remainingPdbTracker = pdb.NewBasicRemainingPdbTracker()
	}
	drainCtx := &drainability.DrainContext{
		RemainingPdbTracker: remainingPdbTracker,
		DeleteOptions:       deleteOptions,
	}
	for _, podInfo := range nodeInfo.Pods {
		pod := podInfo.Pod
		status := drainabilityRules.Drainable(drainCtx, pod)
		switch status.Outcome {
		case drainability.UndefinedOutcome:
			pods = append(pods, podInfo.Pod)
		case drainability.DrainOk:
			if pod_util.IsDaemonSetPod(pod) {
				drainDs = append(drainDs, pod)
			} else {
				drainPods = append(drainPods, pod)
			}
		case drainability.BlockDrain:
			blockingPod = &drain.BlockingPod{
				Pod:    pod,
				Reason: status.BlockingReason,
			}
			err = status.Error
			return
		}
	}

	pods, daemonSetPods, blockingPod, err = drain.GetPodsForDeletionOnNodeDrain(
		pods,
		remainingPdbTracker.GetPdbs(),
		deleteOptions.SkipNodesWithSystemPods,
		deleteOptions.SkipNodesWithLocalStorage,
		deleteOptions.SkipNodesWithCustomControllerPods,
		listers,
		int32(deleteOptions.MinReplicaCount),
		timestamp)
	pods = append(pods, drainPods...)
	daemonSetPods = append(daemonSetPods, drainDs...)
	if err != nil {
		return pods, daemonSetPods, blockingPod, err
	}
	if canRemove, _, blockingPodInfo := remainingPdbTracker.CanRemovePods(pods); !canRemove {
		pod := blockingPodInfo.Pod
		return []*apiv1.Pod{}, []*apiv1.Pod{}, blockingPodInfo, fmt.Errorf("not enough pod disruption budget to move %s/%s", pod.Namespace, pod.Name)
	}

	return pods, daemonSetPods, nil, nil
}
