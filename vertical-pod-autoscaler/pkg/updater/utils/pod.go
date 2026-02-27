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

package utils

import (
	corev1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/sidecar"
)

// GetPodCondition will get Pod's condition.
func GetPodCondition(pod *corev1.Pod, conditionType corev1.PodConditionType) (corev1.PodCondition, bool) {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == conditionType {
			return cond, true
		}
	}
	return corev1.PodCondition{}, false
}

// IsNonDisruptiveResize checks if all containers in the pod have NotRequired
// resize policy for the resources being resized. If any container requires
// restart for any resource, returns false.
func IsNonDisruptiveResize(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		for _, policy := range container.ResizePolicy {
			// If any resource has RestartContainer policy, it's disruptive
			if policy.RestartPolicy == corev1.RestartContainer {
				return false
			}
		}
	}
	// Also check native sidecar init containers for resize policy
	if features.Enabled(features.NativeSidecar) {
		for i, container := range pod.Spec.InitContainers {
			if !sidecar.IsNativeSidecar(&pod.Spec.InitContainers[i]) {
				continue
			}
			for _, policy := range container.ResizePolicy {
				if policy.RestartPolicy == corev1.RestartContainer {
					return false
				}
			}
		}
	}
	return true
}
