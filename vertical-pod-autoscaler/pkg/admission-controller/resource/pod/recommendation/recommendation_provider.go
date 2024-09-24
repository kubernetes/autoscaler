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

package recommendation

import (
	"fmt"

	core "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// Provider gets current recommendation, annotations and vpaName for the given pod.
type Provider interface {
	GetContainersResourcesForPod(pod *core.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]vpa_api_util.ContainerResources, vpa_api_util.ContainerToAnnotationsMap, error)
}

type recommendationProvider struct {
	limitsRangeCalculator   limitrange.LimitRangeCalculator
	recommendationProcessor vpa_api_util.RecommendationProcessor
}

// NewProvider constructs the recommendation provider that can be used to determine recommendations for pods.
func NewProvider(calculator limitrange.LimitRangeCalculator,
	recommendationProcessor vpa_api_util.RecommendationProcessor) Provider {
	return &recommendationProvider{
		limitsRangeCalculator:   calculator,
		recommendationProcessor: recommendationProcessor,
	}
}

// GetContainersResources returns the recommended resources for each container in the given pod in the same order they are specified in the pod.Spec.
// If addAll is set to true, containers w/o a recommendation are also added to the list (and their non-recommended requests and limits will always be preserved if present),
// otherwise they're skipped (default behaviour).
func GetContainersResources(pod *core.Pod, vpaResourcePolicy *vpa_types.PodResourcePolicy, podRecommendation vpa_types.RecommendedPodResources, limitRange *core.LimitRangeItem,
	addAll bool, annotations vpa_api_util.ContainerToAnnotationsMap) []vpa_api_util.ContainerResources {
	resources := make([]vpa_api_util.ContainerResources, len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {
		recommendation := vpa_api_util.GetRecommendationForContainer(container.Name, &podRecommendation)
		if recommendation == nil {
			if !addAll {
				klog.V(2).InfoS("No recommendation found for container, skipping", "container", container.Name)
				continue
			}
			klog.V(2).InfoS("No match found for container, using Pod request", "container", container.Name)
			resources[i].Requests = container.Resources.Requests
		} else {
			resources[i].Requests = recommendation.Target
		}
		defaultLimit := core.ResourceList{}
		if limitRange != nil {
			defaultLimit = limitRange.Default
		}
		containerControlledValues := vpa_api_util.GetContainerControlledValues(container.Name, vpaResourcePolicy)
		if containerControlledValues == vpa_types.ContainerControlledValuesRequestsAndLimits {
			proportionalLimits, limitAnnotations := vpa_api_util.GetProportionalLimit(container.Resources.Limits, container.Resources.Requests, resources[i].Requests, defaultLimit)
			if proportionalLimits != nil {
				resources[i].Limits = proportionalLimits
				if len(limitAnnotations) > 0 {
					annotations[container.Name] = append(annotations[container.Name], limitAnnotations...)
				}
			}
		}
		// If the recommendation only contains CPU or Memory (if the VPA was configured this way), we need to make sure we "backfill" the other.
		// Only do this when the addAll flag is true.
		if addAll {
			cpuRequest, hasCpuRequest := container.Resources.Requests[core.ResourceCPU]
			if _, ok := resources[i].Requests[core.ResourceCPU]; !ok && hasCpuRequest {
				resources[i].Requests[core.ResourceCPU] = cpuRequest
			}
			memRequest, hasMemRequest := container.Resources.Requests[core.ResourceMemory]
			if _, ok := resources[i].Requests[core.ResourceMemory]; !ok && hasMemRequest {
				resources[i].Requests[core.ResourceMemory] = memRequest
			}
			cpuLimit, hasCpuLimit := container.Resources.Limits[core.ResourceCPU]
			if _, ok := resources[i].Limits[core.ResourceCPU]; !ok && hasCpuLimit {
				resources[i].Limits[core.ResourceCPU] = cpuLimit
			}
			memLimit, hasMemLimit := container.Resources.Limits[core.ResourceMemory]
			if _, ok := resources[i].Limits[core.ResourceMemory]; !ok && hasMemLimit {
				resources[i].Limits[core.ResourceMemory] = memLimit
			}
		}
	}
	return resources
}

// GetContainersResourcesForPod returns recommended request for a given pod and associated annotations.
// The returned slice corresponds 1-1 to containers in the Pod.
func (p *recommendationProvider) GetContainersResourcesForPod(pod *core.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]vpa_api_util.ContainerResources, vpa_api_util.ContainerToAnnotationsMap, error) {
	if vpa == nil || pod == nil {
		klog.V(2).InfoS("Can't calculate recommendations, one of VPA or Pod is nil", "vpa", vpa, "pod", pod)
		return nil, nil, nil
	}
	klog.V(2).InfoS("Updating requirements for pod", "pod", pod.Name)

	var annotations vpa_api_util.ContainerToAnnotationsMap
	recommendedPodResources := &vpa_types.RecommendedPodResources{}

	if vpa.Status.Recommendation != nil {
		var err error
		recommendedPodResources, annotations, err = p.recommendationProcessor.Apply(vpa.Status.Recommendation, vpa.Spec.ResourcePolicy, vpa.Status.Conditions, pod)
		if err != nil {
			klog.V(2).InfoS("Cannot process recommendation for pod", "pod", klog.KObj(pod))
			return nil, annotations, err
		}
	}
	containerLimitRange, err := p.limitsRangeCalculator.GetContainerLimitRangeItem(pod.Namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting containerLimitRange: %s", err)
	}
	var resourcePolicy *vpa_types.PodResourcePolicy
	if vpa.Spec.UpdatePolicy == nil || vpa.Spec.UpdatePolicy.UpdateMode == nil || *vpa.Spec.UpdatePolicy.UpdateMode != vpa_types.UpdateModeOff {
		resourcePolicy = vpa.Spec.ResourcePolicy
	}
	containerResources := GetContainersResources(pod, resourcePolicy, *recommendedPodResources, containerLimitRange, false, annotations)

	// Ensure that we are not propagating empty resource key if any.
	for _, resource := range containerResources {
		if resource.RemoveEmptyResourceKeyIfAny() {
			klog.InfoS("An empty resource key was found and purged", "pod", klog.KObj(pod), "vpa", klog.KObj(vpa))
		}
	}

	return containerResources, annotations, nil
}
