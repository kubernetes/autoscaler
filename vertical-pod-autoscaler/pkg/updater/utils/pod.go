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
	apiv1 "k8s.io/api/core/v1"
)

// GetPodCondition will get Pod's condition.
func GetPodCondition(pod *apiv1.Pod, conditionType apiv1.PodConditionType) (apiv1.PodCondition, bool) {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == conditionType {
			return cond, true
		}
	}
	return apiv1.PodCondition{}, false
}

// IsNonDisruptiveResize checks if all containers in the pod have NotRequired
// resize policy for the resources being resized. If any container requires
// restart for any resource, returns false.
func IsNonDisruptiveResize(pod *apiv1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		for _, policy := range container.ResizePolicy {
			// If any resource has RestartContainer policy, it's disruptive
			if policy.RestartPolicy == apiv1.RestartContainer {
				return false
			}
		}
	}
	// TODO(omerap12): do we want to check here for InitContainers/InitContainers+restartPolicy Always/
	// Also check init containers if they can be resized
	return true
}
