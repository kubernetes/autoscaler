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
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
)

// GetPodsToMove returns a list of pods that should be moved elsewhere and a
// list of DaemonSet pods that should be evicted if the node is drained.
// Raises error if there is an unreplicated pod.
// Based on kubectl drain code. If listers is nil it makes an assumption that
// RC, DS, Jobs and RS were deleted along with their pods (no abandoned pods
// with dangling created-by annotation).
// If listers is not nil it checks whether RC, DS, Jobs and RS that created
// these pods still exist.
func GetPodsToMove(nodeInfo *framework.NodeInfo, deleteOptions options.NodeDeleteOptions, drainabilityRules rules.Rules, listers kube_util.ListerRegistry, remainingPdbTracker pdb.RemainingPdbTracker, timestamp time.Time) (pods []*apiv1.Pod, daemonSetPods []*apiv1.Pod, blockingPod *drain.BlockingPod, err error) {
	if drainabilityRules == nil {
		drainabilityRules = rules.Default(deleteOptions)
	}
	if remainingPdbTracker == nil {
		remainingPdbTracker = pdb.NewBasicRemainingPdbTracker()
	}
	drainCtx := &drainability.DrainContext{
		RemainingPdbTracker: remainingPdbTracker,
		Listers:             listers,
		Timestamp:           timestamp,
	}
	for _, podInfo := range nodeInfo.Pods {
		pod := podInfo.Pod
		status := drainabilityRules.Drainable(drainCtx, pod, nodeInfo)
		switch status.Outcome {
		case drainability.UndefinedOutcome, drainability.DrainOk:
			if pod_util.IsDaemonSetPod(pod) {
				daemonSetPods = append(daemonSetPods, pod)
			} else {
				pods = append(pods, pod)
			}
		case drainability.BlockDrain:
			return nil, nil, &drain.BlockingPod{
				Pod:    pod,
				Reason: status.BlockingReason,
			}, status.Error
		}
	}
	return pods, daemonSetPods, nil, nil
}
