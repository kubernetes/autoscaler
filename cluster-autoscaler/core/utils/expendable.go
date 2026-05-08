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
)

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
