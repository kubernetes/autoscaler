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

package annotations

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

const (
	// StartupCPUBoostAnnotation is the annotation set on a pod when a CPU boost is applied.
	// The value of the annotation is the original resource specification of the container.
	StartupCPUBoostAnnotation = "startup-cpu-boost"
)

// OriginalResources contains the original resources of a container.
type OriginalResources struct {
	Requests corev1.ResourceList `json:"requests"`
	Limits   corev1.ResourceList `json:"limits"`
}

// GetOriginalResourcesAnnotationValue returns the annotation value for the original resources.
func GetOriginalResourcesAnnotationValue(container *corev1.Container) (string, error) {
	original := OriginalResources{
		Requests: corev1.ResourceList{},
		Limits:   corev1.ResourceList{},
	}
	if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
		original.Requests[corev1.ResourceCPU] = cpu
	}
	if mem, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
		original.Requests[corev1.ResourceMemory] = mem
	}
	if cpu, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
		original.Limits[corev1.ResourceCPU] = cpu
	}
	if mem, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
		original.Limits[corev1.ResourceMemory] = mem
	}
	b, err := json.Marshal(original)
	return string(b), err
}

// GetOriginalResourcesFromAnnotation returns the original resources from the annotation.
func GetOriginalResourcesFromAnnotation(pod *corev1.Pod) (*OriginalResources, error) {
	val, ok := pod.Annotations[StartupCPUBoostAnnotation]
	if !ok {
		return nil, nil
	}
	var original OriginalResources
	err := json.Unmarshal([]byte(val), &original)
	if err != nil {
		return nil, err
	}
	return &original, nil
}
