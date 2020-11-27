/*
Copyright 2016 The Kubernetes Authors.

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
	apiv1 "k8s.io/api/core/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
)

// getRequiredPodsForNode returns a list of pods that would appear on the node if the
// node was just created (like daemonset and manifest-run pods). It reuses kubectl
// drain command to get the list.
func getRequiredPodsForNode(nodename string, podsForNodes map[string][]*apiv1.Pod) ([]*apiv1.Pod, errors.AutoscalerError) {
	allPods := podsForNodes[nodename]

	return filterRequiredPodsForNode(allPods), nil
}

// BuildNodeInfoForNode build a NodeInfo structure for the given node as if the node was just created.
func BuildNodeInfoForNode(node *apiv1.Node, podsForNodes map[string][]*apiv1.Pod) (*schedulerframework.NodeInfo, errors.AutoscalerError) {
	requiredPods, err := getRequiredPodsForNode(node.Name, podsForNodes)
	if err != nil {
		return nil, err
	}
	result := schedulerframework.NewNodeInfo(requiredPods...)
	if err := result.SetNode(node); err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	return result, nil
}

func filterRequiredPodsForNode(allPods []*apiv1.Pod) []*apiv1.Pod {
	var selectedPods []*apiv1.Pod

	for id, pod := range allPods {
		// Ignore pod in deletion phase
		if pod.DeletionTimestamp != nil {
			continue
		}

		if pod_util.IsMirrorPod(pod) || pod_util.IsDaemonSetPod(pod) {
			selectedPods = append(selectedPods, allPods[id])
		}
	}

	return selectedPods
}
