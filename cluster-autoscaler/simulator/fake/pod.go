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

package fake

import (
	apiv1 "k8s.io/api/core/v1"
)

const (
	// FakePodAnnotationKey the key for pod type.
	FakePodAnnotationKey = "podtype"
	// FakePodAnnotationValue the value for a fake pod,
	FakePodAnnotationValue = "fakepod"
)

// IsFake returns true if the a pod is marked as fake and false otherwise,
func IsFake(pod *apiv1.Pod) bool {
	if pod.Annotations == nil {
		return false
	}
	return pod.Annotations[FakePodAnnotationKey] == FakePodAnnotationValue
}

// WithFakePodAnnotation adds annotation of key `FakePodAnnotationKey` with value `FakePodAnnotationValue` to passed pod.
// WithFakePodAnnotation also creates a new annotations map if original pod.Annotations is nil.
func WithFakePodAnnotation(pod *apiv1.Pod) *apiv1.Pod {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string, 1)
	}
	pod.Annotations[FakePodAnnotationKey] = FakePodAnnotationValue
	return pod
}
