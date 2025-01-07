/*
Copyright 2024 The Kubernetes Authors.

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

package inplace

import (
	"fmt"

	core "k8s.io/api/core/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/recommendation"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

type inPlaceRecommendationProvider struct {
	limitsRangeCalculator   limitrange.LimitRangeCalculator
	recommendationProcessor vpa_api_util.RecommendationProcessor
}

// NewInPlaceRecommendationProvider constructs the recommendation provider that can be used to determine recommendations for pods.
func NewInPlaceRecommendationProvider(calculator limitrange.LimitRangeCalculator,
	recommendationProcessor vpa_api_util.RecommendationProcessor) recommendation.Provider {
	return &inPlaceRecommendationProvider{
		limitsRangeCalculator:   calculator,
		recommendationProcessor: recommendationProcessor,
	}
}

// GetContainersResourcesForPod returns recommended request for a given pod.
// The returned slice corresponds 1-1 to containers in the Pod.
func (p *inPlaceRecommendationProvider) GetContainersResourcesForPod(pod *core.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]vpa_api_util.ContainerResources, vpa_api_util.ContainerToAnnotationsMap, error) {
	if vpa == nil || pod == nil {
		klog.V(2).InfoS("Can't calculate recommendations, one of VPA or Pod is nil", "vpa", vpa, "pod", pod)
		return nil, nil, nil
	}
	klog.V(2).InfoS("Updating requirements for pod", "pod", pod.Name)

	recommendedPodResources := &vpa_types.RecommendedPodResources{}

	if vpa.Status.Recommendation != nil {
		var err error
		// ignore annotations as they are cannot be used when patching resize subresource
		recommendedPodResources, _, err = p.recommendationProcessor.Apply(vpa, pod)
		if err != nil {
			klog.V(2).InfoS("Cannot process recommendation for pod", "pod", klog.KObj(pod))
			return nil, nil, err
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
	containerResources := recommendation.GetContainersResources(pod, resourcePolicy, *recommendedPodResources, containerLimitRange, false, nil)

	// Ensure that we are not propagating empty resource key if any.
	for _, resource := range containerResources {
		if resource.RemoveEmptyResourceKeyIfAny() {
			klog.InfoS("An empty resource key was found and purged", "pod", klog.KObj(pod), "vpa", klog.KObj(vpa))
		}
	}

	return containerResources, nil, nil
}
