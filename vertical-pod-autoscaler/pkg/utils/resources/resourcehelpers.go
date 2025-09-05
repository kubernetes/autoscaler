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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	metrics_resources "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/resources"
)

// ContainerRequestsAndLimits returns a copy of the actual resource requests and
// limits of a given container:
//
//   - If in-place pod updates feature [1] is enabled, the actual resource requests
//     are stored in the container status field.
//   - Otherwise, fallback to the resource requests defined in the pod spec.
//
// [1] https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources
func ContainerRequestsAndLimits(containerName string, pod *corev1.Pod) (corev1.ResourceList, corev1.ResourceList) {
	containerStatuses, containers := pod.Status.ContainerStatuses, pod.Spec.Containers
	containerStatusSource, containerSource := metrics_resources.ContainerStatus, metrics_resources.PodSpecContainer
	if isInitContainer(containerName, pod) {
		containerStatuses, containers = pod.Status.InitContainerStatuses, pod.Spec.InitContainers
		containerStatusSource, containerSource = metrics_resources.InitContainerStatus, metrics_resources.PodSpecInitContainer
	}

	cs := containerStatusFor(containerName, containerStatuses)
	if cs != nil && cs.Resources != nil {
		metrics_resources.RecordGetResourcesCount(containerStatusSource)
		return cs.Resources.Requests.DeepCopy(), cs.Resources.Limits.DeepCopy()
	}

	klog.V(6).InfoS("Container resources not found in containerStatus for container. Falling back to resources defined in the pod spec. This is expected for clusters with in-place pod updates feature disabled.", "container", containerName, "containerStatus", cs)
	container := findContainer(containerName, containers)
	if container != nil {
		metrics_resources.RecordGetResourcesCount(containerSource)
		return container.Resources.Requests.DeepCopy(), container.Resources.Limits.DeepCopy()
	}

	return nil, nil
}

func findContainer(containerName string, containers []corev1.Container) *corev1.Container {
	for i, container := range containers {
		if container.Name == containerName {
			return &containers[i]
		}
	}
	return nil
}

func containerStatusFor(containerName string, containerStatuses []corev1.ContainerStatus) *corev1.ContainerStatus {
	for i, containerStatus := range containerStatuses {
		if containerStatus.Name == containerName {
			return &containerStatuses[i]
		}
	}
	return nil
}

func isInitContainer(containerName string, pod *corev1.Pod) bool {
	// A regular container with the same name takes precedence over an init container.
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return false
		}
	}
	for _, container := range pod.Spec.InitContainers {
		if container.Name == containerName {
			return true
		}
	}
	return false
}

// RecommendationHasLowerResource returns true if recommendation b has at least one
// resource target lower than a for any matching container. This is used for infeasible
// retry logic: we don't know which resource causes infeasibility, so any reduction
// is worth retrying.
func RecommendationHasLowerResource(a, b *vpa_types.RecommendedPodResources) bool {
	if a == nil || b == nil {
		return false
	}
	for _, aRec := range a.ContainerRecommendations {
		for _, bRec := range b.ContainerRecommendations {
			if aRec.ContainerName == bRec.ContainerName {
				if HasLowerResource(aRec.Target, bRec.Target) {
					return true
				}
			}
		}
	}
	return false
}

// HasLowerResource returns true if any resource in b is lower than the
// corresponding resource in a.
func HasLowerResource(a, b corev1.ResourceList) bool {
	for key, aVal := range a {
		if bVal, exists := b[key]; exists && bVal.Cmp(aVal) < 0 {
			return true
		}
	}
	return false
}
