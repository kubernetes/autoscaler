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
	cs := containerStatusFor(containerName, pod)
	if cs != nil && cs.Resources != nil {
		metrics_resources.RecordGetResourcesCount(metrics_resources.ContainerStatus)
		return cs.Resources.Requests.DeepCopy(), cs.Resources.Limits.DeepCopy()
	}

	klog.V(6).InfoS("Container resources not found in containerStatus for container. Falling back to resources defined in the pod spec. This is expected for clusters with in-place pod updates feature disabled.", "container", containerName, "containerStatus", cs)
	container := findContainer(containerName, pod)
	if container != nil {
		metrics_resources.RecordGetResourcesCount(metrics_resources.PodSpecContainer)
		return container.Resources.Requests.DeepCopy(), container.Resources.Limits.DeepCopy()
	}

	return nil, nil
}

// InitContainerRequestsAndLimits returns a copy of the actual resource requests
// and limits of a given initContainer:
//
//   - If in-place pod updates feature [1] is enabled, the actual resource requests
//     are stored in the initContainer status field.
//   - Otherwise, fallback to the resource requests defined in the pod spec.
//
// [1] https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources
func InitContainerRequestsAndLimits(initContainerName string, pod *corev1.Pod) (corev1.ResourceList, corev1.ResourceList) {
	cs := initContainerStatusFor(initContainerName, pod)
	if cs != nil && cs.Resources != nil {
		metrics_resources.RecordGetResourcesCount(metrics_resources.InitContainerStatus)
		return cs.Resources.Requests.DeepCopy(), cs.Resources.Limits.DeepCopy()
	}

	klog.V(6).InfoS("initContainer resources not found in initContainerStatus for initContainer. Falling back to resources defined in the pod spec. This is expected for clusters with in-place pod updates feature disabled.", "initContainer", initContainerName, "initContainerStatus", cs)
	initContainer := findInitContainer(initContainerName, pod)
	if initContainer != nil {
		metrics_resources.RecordGetResourcesCount(metrics_resources.PodSpecInitContainer)
		return initContainer.Resources.Requests.DeepCopy(), initContainer.Resources.Limits.DeepCopy()
	}

	return nil, nil
}

func findContainer(containerName string, pod *corev1.Pod) *corev1.Container {
	for i, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return &pod.Spec.Containers[i]
		}
	}
	return nil
}

func findInitContainer(initContainerName string, pod *corev1.Pod) *corev1.Container {
	for i, initContainer := range pod.Spec.InitContainers {
		if initContainer.Name == initContainerName {
			return &pod.Spec.InitContainers[i]
		}
	}
	return nil
}

func containerStatusFor(containerName string, pod *corev1.Pod) *corev1.ContainerStatus {
	for i, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			return &pod.Status.ContainerStatuses[i]
		}
	}
	return nil
}

func initContainerStatusFor(initContainerName string, pod *corev1.Pod) *corev1.ContainerStatus {
	for i, initContainerStatus := range pod.Status.InitContainerStatuses {
		if initContainerStatus.Name == initContainerName {
			return &pod.Status.InitContainerStatuses[i]
		}
	}
	return nil
}

// SumContainerLevelRecommendations implements the algorithm that calculates Pod-level recommendations.
// It sums all container-level recommendations, such as Target LowerBound, UpperBound and UncappedTarget
func SumContainerLevelRecommendations(containerRecommendations []vpa_types.RecommendedContainerResources) *vpa_types.PodRecommendations {
	if len(containerRecommendations) == 0 {
		return nil
	}

	add := func(dst, src corev1.ResourceList) corev1.ResourceList {
		for name, q := range src {
			if existing, ok := dst[name]; ok {
				existing.Add(q)
				dst[name] = existing
				continue
			}
			dst[name] = q.DeepCopy()
		}
		return dst
	}

	podRecommendations := vpa_types.PodRecommendations{
		Target:         make(corev1.ResourceList),
		LowerBound:     make(corev1.ResourceList),
		UpperBound:     make(corev1.ResourceList),
		UncappedTarget: make(corev1.ResourceList),
	}

	for _, container := range containerRecommendations {
		podRecommendations.LowerBound = add(podRecommendations.LowerBound, container.LowerBound)
		podRecommendations.Target = add(podRecommendations.Target, container.Target)
		podRecommendations.UpperBound = add(podRecommendations.UpperBound, container.UpperBound)
		podRecommendations.UncappedTarget = add(podRecommendations.UncappedTarget, container.UncappedTarget)
	}

	return &podRecommendations
}

// PodRequestsAndLimits returns a copy of the actual Pod-level resource
// requests and limits:
//
//   - If the "In-Place Pod Resize" (IPPR) feature at the Pod-level [1] is enabled, the actual
//     resource requests are stored in PodStatus.
//   - Otherwise, fall back to the resource requests defined in PodSpec.
//
// [1] https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/5419-pod-level-resources-in-place-resize/
func PodRequestsAndLimits(pod *corev1.Pod) (corev1.ResourceList, corev1.ResourceList) {
	// TODO: Comment out this part once https://github.com/kubernetes/kubernetes/issues/137628 is fixed and backported to Kubernetes versions.
	// ps := podStatus(pod)
	// if ps != nil && ps.Resources != nil {
	// 	return ps.Resources.Requests.DeepCopy(), ps.Resources.Limits.DeepCopy()
	// }
	// klog.V(6).InfoS("Resources are not found in PodStatus. Fall back to resources defined in the PodSpec. This behavior is expected when InPlacePodLevelResourcesVerticalScaling feature gate is disabled.", "podStatus", ps)
	if pod.Spec.Resources != nil {
		return pod.Spec.Resources.Requests.DeepCopy(), pod.Spec.Resources.Limits.DeepCopy()
	}
	return nil, nil
}

// func podStatus(pod *corev1.Pod) *corev1.PodStatus {
// 	if pod.Status.Resources != nil {
// 		return &pod.Status
// 	}
// 	return nil
// }
