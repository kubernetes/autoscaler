/*
Copyright 2018 The Kubernetes Authors.

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

package tpu

import (
	"strings"

	apiv1 "k8s.io/api/core/v1"
)

const (
	// ResourceTPUPrefix is the prefix of the TPU resource names.
	ResourceTPUPrefix = "cloud-tpus.google.com/"
)

func hasTPURequest(pod *apiv1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		for name := range container.Resources.Requests {
			if strings.HasPrefix(string(name), ResourceTPUPrefix) {
				return true
			}
		}
	}

	return false
}

func clearTPURequest(pod *apiv1.Pod) *apiv1.Pod {
	sanitized := pod.DeepCopy()
	for _, container := range sanitized.Spec.Containers {
		for name := range container.Resources.Requests {
			if strings.HasPrefix(string(name), ResourceTPUPrefix) {
				delete(container.Resources.Requests, name)
			}
		}
	}

	return sanitized
}

// ClearTPURequests clears TPU requests for pods, so they don't interfere with scheduling
// simulations. This isn't yet zone-aware.
func ClearTPURequests(pods []*apiv1.Pod) []*apiv1.Pod {
	podsWithTPU := make(map[int]*apiv1.Pod)
	for i, pod := range pods {
		if hasTPURequest(pod) {
			podsWithTPU[i] = clearTPURequest(pod)
		}
	}

	if len(podsWithTPU) == 0 {
		return pods
	}

	// Copy slice only if we need to replace some pods.
	sanitizedPods := make([]*apiv1.Pod, len(pods))
	for i, pod := range pods {
		if sanitized, found := podsWithTPU[i]; found {
			sanitizedPods[i] = sanitized
		} else {
			sanitizedPods[i] = pod
		}
	}
	return sanitizedPods
}
