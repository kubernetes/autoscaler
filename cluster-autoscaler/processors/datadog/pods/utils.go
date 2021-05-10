/*
Copyright 2021 The Kubernetes Authors.

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

package pods

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ageCondition int

const (
	longPendingCutoff = time.Hour * 2
	youngerThan       = iota
	olderThan         = iota
)

func countDistinctOwnerReferences(pods []*apiv1.Pod) int {
	distinctOwners := make(map[types.UID]struct{})
	for _, pod := range pods {
		controllerRef := metav1.GetControllerOf(pod)
		if controllerRef == nil {
			continue
		}
		distinctOwners[controllerRef.UID] = struct{}{}
	}

	return len(distinctOwners)
}

func filterByAge(pods []*apiv1.Pod, condition ageCondition, age time.Duration) []*apiv1.Pod {
	var filtered []*apiv1.Pod
	for _, pod := range pods {
		cutoff := pod.GetCreationTimestamp().Time.Add(age)
		if condition == youngerThan && cutoff.After(time.Now()) {
			filtered = append(filtered, pod)
		}
		if condition == olderThan && cutoff.Before(time.Now()) {
			filtered = append(filtered, pod)
		}
	}
	return filtered
}
