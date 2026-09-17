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
	containers := pod.Spec.Containers
	if features.Enabled(features.NativeSidecar) {
		containers = make([]corev1.Container, 0, len(pod.Spec.Containers)+len(pod.Spec.InitContainers))
		containers = append(containers, pod.Spec.Containers...)
		for i := range pod.Spec.InitContainers {
			if c := &pod.Spec.InitContainers[i]; c.RestartPolicy != nil && *c.RestartPolicy == corev1.ContainerRestartPolicyAlways {
				containers = append(containers, *c)
			}
		}
	}
	for _, container := range containers {
		for _, policy := range container.ResizePolicy {
			// If any resource has RestartContainer policy, it's disruptive
			if policy.RestartPolicy == corev1.RestartContainer {
				return false
			}
		}
	}
	return true
}
