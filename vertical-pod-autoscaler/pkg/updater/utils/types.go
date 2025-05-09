/*
Copyright 2025 The Kubernetes Authors.

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

// InPlaceDecision is the type of decision that can be made for a pod.
type InPlaceDecision string

const (
	// InPlaceApproved means we can in-place update the pod.
	InPlaceApproved InPlaceDecision = "InPlaceApproved"
	// InPlaceDeferred means we can't in-place update the pod right now, but we will wait for the next loop to check for in-placeability again
	InPlaceDeferred InPlaceDecision = "InPlaceDeferred"
	// InPlaceEvict means we will attempt to evict the pod.
	InPlaceEvict InPlaceDecision = "InPlaceEvict"
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
