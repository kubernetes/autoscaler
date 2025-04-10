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

package resourcehelpers

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// ContainerRequestsAndLimits returns a copy of the actual resource requests and
// limits of a given container:
//
//   - If in-place pod updates feature [1] is enabled, the actual resource requests
//     are stored in the container status field.
//   - Otherwise, fallback to the resource requests defined in the pod spec.
//
// [1] https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources
func ContainerRequestsAndLimits(containerName string, pod *v1.Pod) (v1.ResourceList, v1.ResourceList) {
	cs := containerStatusForContainer(containerName, pod)
	if cs != nil && cs.Resources != nil {
		return cs.Resources.Requests.DeepCopy(), cs.Resources.Limits.DeepCopy()
	}

	klog.V(6).InfoS("Container resources not found in containerStatus for container. Falling back to resources defined in the pod spec. This is expected for clusters with in-place pod updates feature disabled.", "container", containerName, "containerStatus", cs)
	container := findContainer(containerName, pod)
	if container != nil {
		return container.Resources.Requests.DeepCopy(), container.Resources.Limits.DeepCopy()
	}

	return nil, nil
}

func findContainer(containerName string, pod *v1.Pod) *v1.Container {
	for i, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return &pod.Spec.Containers[i]
		}
	}
	return nil
}

func containerStatusForContainer(containerName string, pod *v1.Pod) *v1.ContainerStatus {
	for i, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			return &pod.Status.ContainerStatuses[i]
		}
	}
	return nil
}
