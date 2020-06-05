/*
Copyright 2019 The Kubernetes Authors.

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

package utils

import (
	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

// FilterOutExpendableAndSplit filters out expendable pods and splits into:
//   - waiting for lower priority pods preemption
//   - other pods.
func FilterOutExpendableAndSplit(unschedulableCandidates []*apiv1.Pod, nodes []*apiv1.Node, expendablePodsPriorityCutoff int) ([]*apiv1.Pod, []*apiv1.Pod) {
	var unschedulableNonExpendable []*apiv1.Pod
	var waitingForLowerPriorityPreemption []*apiv1.Pod

	nodeNames := make(map[string]bool)
	for _, node := range nodes {
		nodeNames[node.Name] = true
	}

	for _, pod := range unschedulableCandidates {
		if pod.Spec.Priority != nil && int(*pod.Spec.Priority) < expendablePodsPriorityCutoff {
			klog.V(4).Infof("Pod %s has priority below %d (%d) and will scheduled when enough resources is free. Ignoring in scale up.", pod.Name, expendablePodsPriorityCutoff, *pod.Spec.Priority)
		} else if nominatedNodeName := pod.Status.NominatedNodeName; nominatedNodeName != "" {
			if nodeNames[nominatedNodeName] {
				klog.V(4).Infof("Pod %s will be scheduled after low priority pods are preempted on %s. Ignoring in scale up.", pod.Name, nominatedNodeName)
				waitingForLowerPriorityPreemption = append(waitingForLowerPriorityPreemption, pod)
			} else {
				klog.V(4).Infof("Pod %s has nominatedNodeName set to %s but node is gone", pod.Name, nominatedNodeName)
				unschedulableNonExpendable = append(unschedulableNonExpendable, pod)
			}
		} else {
			unschedulableNonExpendable = append(unschedulableNonExpendable, pod)
		}
	}
	return unschedulableNonExpendable, waitingForLowerPriorityPreemption
}

// FilterOutExpendablePods filters out expendable pods.
func FilterOutExpendablePods(pods []*apiv1.Pod, expendablePodsPriorityCutoff int) []*apiv1.Pod {
	var result []*apiv1.Pod
	for _, pod := range pods {
		if !IsExpendablePod(pod, expendablePodsPriorityCutoff) {
			result = append(result, pod)
		}
	}
	return result
}

// IsExpendablePod tests if pod is expendable for give priority cutoff
func IsExpendablePod(pod *apiv1.Pod, expendablePodsPriorityCutoff int) bool {
	return pod.Spec.Priority != nil && int(*pod.Spec.Priority) < expendablePodsPriorityCutoff
}
