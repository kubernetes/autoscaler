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

package logic

import (
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"math"
	"math/big"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1beta2"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/klog"
)

// ContainerResources holds resources request for container
type ContainerResources struct {
	Limits   v1.ResourceList
	Requests v1.ResourceList
}

func newContainerResources() ContainerResources {
	return ContainerResources{
		Requests: v1.ResourceList{},
		Limits:   v1.ResourceList{},
	}
}

// RecommendationProvider gets current recommendation, annotations and vpaName for the given pod.
type RecommendationProvider interface {
	GetContainersResourcesForPod(pod *v1.Pod) ([]ContainerResources, vpa_api_util.ContainerToAnnotationsMap, string, error)
}

type recommendationProvider struct {
	limitsRangeCalculator   LimitsRangeCalculator
	recommendationProcessor vpa_api_util.RecommendationProcessor
	selectorFetcher         target.VpaTargetSelectorFetcher
	vpaLister               vpa_lister.VerticalPodAutoscalerLister
}

// NewRecommendationProvider constructs the recommendation provider that list VPAs and can be used to determine recommendations for pods.
func NewRecommendationProvider(calculator LimitsRangeCalculator, recommendationProcessor vpa_api_util.RecommendationProcessor,
	selectorFetcher target.VpaTargetSelectorFetcher, vpaLister vpa_lister.VerticalPodAutoscalerLister) *recommendationProvider {
	return &recommendationProvider{
		limitsRangeCalculator:   calculator,
		recommendationProcessor: recommendationProcessor,
		selectorFetcher:         selectorFetcher,
		vpaLister:               vpaLister,
	}
}

// scaleQuantityProportionally returns value which has the same proportion to scaledQuantity as scaleResult has to scaleBase
// It also returns a bool indicating if it had to cap result to MaxInt64 milliunits.
func scaleQuantityProportionally(scaledQuantity, scaleBase, scaleResult *resource.Quantity) (*resource.Quantity, bool) {
	originalMilli := big.NewInt(scaledQuantity.MilliValue())
	scaleBaseMilli := big.NewInt(scaleBase.MilliValue())
	scaleResultMilli := big.NewInt(scaleResult.MilliValue())
	var scaledOriginal big.Int
	scaledOriginal.Mul(originalMilli, scaleResultMilli)
	scaledOriginal.Div(&scaledOriginal, scaleBaseMilli)
	if scaledOriginal.IsInt64() {
		return resource.NewMilliQuantity(scaledOriginal.Int64(), scaledQuantity.Format), false
	}
	return resource.NewMilliQuantity(math.MaxInt64, scaledQuantity.Format), true
}

func getProportionalLimit(originalLimit, originalRequest, recommendedRequest, defaultLimit *resource.Quantity) (limit *resource.Quantity, capped bool) {
	if originalLimit == nil || originalLimit.Value() == 0 && defaultLimit != nil {
		originalLimit = defaultLimit
	}
	// originalLimit not set, don't set limit.
	if originalLimit == nil || originalLimit.Value() == 0 {
		return nil, false
	}
	// originalLimit set but originalRequest not set - K8s will treat the pod as if they were equal,
	// recommend limit equal to request
	if originalRequest == nil || originalRequest.Value() == 0 {
		result := *recommendedRequest
		return &result, false
	}
	// originalLimit and originalRequest are set. If they are equal recommend limit equal to request.
	if originalRequest.MilliValue() == originalLimit.MilliValue() {
		result := *recommendedRequest
		return &result, false
	}

	// Input and output milli values should fit in int64 but intermediate values might be bigger.
	return scaleQuantityProportionally( /*scaledQuantity=*/ originalLimit /*scaleBase=*/, originalRequest /*scaleResult=*/, recommendedRequest)
}

func proportionallyCapLimitToMax(recommendedRequest, recommendedLimit, maxLimit *resource.Quantity) (request, limit *resource.Quantity) {
	if recommendedLimit == nil || maxLimit == nil || maxLimit.IsZero() {
		return recommendedRequest, recommendedLimit
	}
	if recommendedLimit.Cmp(*maxLimit) <= 0 {
		return recommendedRequest, recommendedLimit
	}
	scaledRequest, _ := scaleQuantityProportionally(recommendedRequest, recommendedLimit, maxLimit)
	return scaledRequest, maxLimit
}

func proportionallyCapLimitsToMax(recommendedRequests v1.ResourceList, cpuLimit, memLimit, maxCpuLimit, maxMemLimit *resource.Quantity) ContainerResources {
	scaledCpuRequest, scaledCpuLimit := proportionallyCapLimitToMax(recommendedRequests.Cpu(), cpuLimit, maxCpuLimit)
	scaledMemRequest, scaledMemLimit := proportionallyCapLimitToMax(recommendedRequests.Memory(), memLimit, maxMemLimit)
	result := newContainerResources()
	if scaledCpuRequest != nil {
		result.Requests[v1.ResourceCPU] = *scaledCpuRequest
	}
	if scaledCpuLimit != nil {
		result.Limits[v1.ResourceCPU] = *scaledCpuLimit
	}
	if scaledMemRequest != nil {
		result.Requests[v1.ResourceMemory] = *scaledMemRequest
	}
	if scaledMemLimit != nil {
		result.Limits[v1.ResourceMemory] = *scaledMemLimit
	}
	return result
}

// GetContainersResources returns the recommended resources for each container in the given pod in the same order they are specified in the pod.Spec.
func GetContainersResources(pod *v1.Pod, podRecommendation vpa_types.RecommendedPodResources, limitRange *v1.LimitRangeItem,
	annotations vpa_api_util.ContainerToAnnotationsMap) []ContainerResources {
	resources := make([]ContainerResources, len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {

		recommendation := vpa_api_util.GetRecommendationForContainer(container.Name, &podRecommendation)
		if recommendation == nil {
			klog.V(2).Infof("no matching recommendation found for container %s", container.Name)
			continue
		}

		var defaultCpu, defaultMem, maxCpuLimit, maxMemLimit *resource.Quantity
		if limitRange != nil {
			defaultCpu = limitRange.Default.Cpu()
			defaultMem = limitRange.Default.Memory()
			maxCpuLimit = limitRange.Max.Cpu()
			maxMemLimit = limitRange.Max.Memory()
		}
		cpuLimit, capped := getProportionalLimit(container.Resources.Limits.Cpu(), container.Resources.Requests.Cpu(), recommendation.Target.Cpu(), defaultCpu)
		if capped {
			annotations[container.Name] = append(
				annotations[container.Name],
				fmt.Sprintf(
					"Failed to keep CPU limit to request proportion of %d to %d with recommended request of %d milliCPU; doesn't fit in int64. Capping limit to MaxInt64",
					container.Resources.Limits.Cpu().MilliValue(), container.Resources.Requests.Cpu().MilliValue(), recommendation.Target.Cpu().MilliValue()))
		}
		memLimit, capped := getProportionalLimit(container.Resources.Limits.Memory(), container.Resources.Requests.Memory(), recommendation.Target.Memory(), defaultMem)
		if capped {
			annotations[container.Name] = append(
				annotations[container.Name],
				fmt.Sprintf(
					"Failed to keep memory limit to request proportion of %d to %d with recommended request of %d milliBytes; doesn't fit in int64. Capping limit to MaxInt64",
					container.Resources.Limits.Memory().MilliValue(), container.Resources.Requests.Memory().MilliValue(), recommendation.Target.Memory().MilliValue()))
		}
		resources[i] = proportionallyCapLimitsToMax(recommendation.Target, cpuLimit, memLimit, maxCpuLimit, maxMemLimit)
	}
	return resources
}

func (p *recommendationProvider) getMatchingVPA(pod *v1.Pod) *vpa_types.VerticalPodAutoscaler {
	configs, err := p.vpaLister.VerticalPodAutoscalers(pod.Namespace).List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to get vpa configs: %v", err)
		return nil
	}
	onConfigs := make([]*vpa_api_util.VpaWithSelector, 0)
	for _, vpaConfig := range configs {
		if vpa_api_util.GetUpdateMode(vpaConfig) == vpa_types.UpdateModeOff {
			continue
		}
		selector, err := p.selectorFetcher.Fetch(vpaConfig)
		if err != nil {
			klog.V(3).Infof("skipping VPA object %v because we cannot fetch selector", vpaConfig.Name)
			continue
		}
		onConfigs = append(onConfigs, &vpa_api_util.VpaWithSelector{
			Vpa:      vpaConfig,
			Selector: selector,
		})
	}
	klog.V(2).Infof("Let's choose from %d configs for pod %s/%s", len(onConfigs), pod.Namespace, pod.Name)
	result := vpa_api_util.GetControllingVPAForPod(pod, onConfigs)
	if result != nil {
		return result.Vpa
	}
	return nil
}

// GetContainersResourcesForPod returns recommended request for a given pod, annotations and name of controlling VPA.
// The returned slice corresponds 1-1 to containers in the Pod.
func (p *recommendationProvider) GetContainersResourcesForPod(pod *v1.Pod) ([]ContainerResources, vpa_api_util.ContainerToAnnotationsMap, string, error) {
	klog.V(2).Infof("updating requirements for pod %s.", pod.Name)
	vpaConfig := p.getMatchingVPA(pod)
	if vpaConfig == nil {
		klog.V(2).Infof("no matching VPA found for pod %s", pod.Name)
		return nil, nil, "", nil
	}

	var annotations vpa_api_util.ContainerToAnnotationsMap
	recommendedPodResources := &vpa_types.RecommendedPodResources{}

	if vpaConfig.Status.Recommendation != nil {
		var err error
		recommendedPodResources, annotations, err = p.recommendationProcessor.Apply(vpaConfig.Status.Recommendation, vpaConfig.Spec.ResourcePolicy, vpaConfig.Status.Conditions, pod)
		if err != nil {
			klog.V(2).Infof("cannot process recommendation for pod %s", pod.Name)
			return nil, annotations, vpaConfig.Name, err
		}
	}
	podLimitRange, err := p.limitsRangeCalculator.GetContainerLimitRangeItem(pod.Namespace)
	// TODO: Support limit range on pod level.
	if err != nil {
		return nil, nil, "", fmt.Errorf("error getting podLimitRange: %s", err)
	}
	containerResources := GetContainersResources(pod, *recommendedPodResources, podLimitRange, annotations)
	return containerResources, annotations, vpaConfig.Name, nil
}
