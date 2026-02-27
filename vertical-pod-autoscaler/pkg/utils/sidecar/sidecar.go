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

package sidecar

import (
	corev1 "k8s.io/api/core/v1"
)

// IsNativeSidecar returns true if the given container is a native sidecar
// (an init container with restartPolicy set to Always).
func IsNativeSidecar(container *corev1.Container) bool {
	return container.RestartPolicy != nil &&
		*container.RestartPolicy == corev1.ContainerRestartPolicyAlways
}

// IsNativeSidecarByName returns true if the pod has an init container with the
// given name that has restartPolicy set to Always.
func IsNativeSidecarByName(name string, pod *corev1.Pod) bool {
	for i, c := range pod.Spec.InitContainers {
		if c.Name == name {
			return IsNativeSidecar(&pod.Spec.InitContainers[i])
		}
	}
	return false
}
