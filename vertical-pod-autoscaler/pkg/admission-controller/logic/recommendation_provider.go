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
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1beta2"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"

	"github.com/golang/glog"
)

// RecommendationProvider gets current recommendation, annotations and vpaName for the given pod.
type RecommendationProvider interface {
	GetContainersResourcesForPod(pod *v1.Pod) ([]vpa_api_util.ContainerResources, vpa_api_util.ContainerToAnnotationsMap, string, error)
}

type recommendationProvider struct {
	limitsRangeCalculator   LimitRangeCalculator
	recommendationProcessor vpa_api_util.RecommendationProcessor
	selectorFetcher         target.VpaTargetSelectorFetcher
	vpaLister               vpa_lister.VerticalPodAutoscalerLister
}

// NewRecommendationProvider constructs the recommendation provider that list VPAs and can be used to determine recommendations for pods.
func NewRecommendationProvider(calculator LimitRangeCalculator, recommendationProcessor vpa_api_util.RecommendationProcessor,
	selectorFetcher target.VpaTargetSelectorFetcher, vpaLister vpa_lister.VerticalPodAutoscalerLister) *recommendationProvider {
	return &recommendationProvider{
		limitsRangeCalculator:   calculator,
		recommendationProcessor: recommendationProcessor,
		selectorFetcher:         selectorFetcher,
		vpaLister:               vpaLister,
	}
}

// GetContainersResources returns the recommended resources for each container in the given pod in the same order they are specified in the pod.Spec.
func GetContainersResources(pod *v1.Pod, podRecommendation vpa_types.RecommendedPodResources, limitRange *v1.LimitRangeItem,
	annotations vpa_api_util.ContainerToAnnotationsMap) []vpa_api_util.ContainerResources {
	resources := make([]vpa_api_util.ContainerResources, len(pod.Spec.Containers))
	var defaultCpu, defaultMem, maxCpuLimit, maxMemLimit *resource.Quantity
	if limitRange != nil {
		defaultCpu = limitRange.Default.Cpu()
		defaultMem = limitRange.Default.Memory()
		maxCpuLimit = limitRange.Max.Cpu()
		maxMemLimit = limitRange.Max.Memory()
	}
	for i, container := range pod.Spec.Containers {
		recommendation := vpa_api_util.GetRecommendationForContainer(container.Name, &podRecommendation)
		if recommendation == nil {
			glog.V(2).Infof("no matching recommendation found for container %s", container.Name)
			continue
		}
		cpuLimit, annotation := vpa_api_util.GetProportionalLimit(container.Resources.Limits.Cpu(), container.Resources.Requests.Cpu(), recommendation.Target.Cpu(), defaultCpu)
		if annotation != "" {
			annotations[container.Name] = append(annotations[container.Name], fmt.Sprintf("CPU: %s", annotation))
		}
		memLimit, annotation := vpa_api_util.GetProportionalLimit(container.Resources.Limits.Memory(), container.Resources.Requests.Memory(), recommendation.Target.Memory(), defaultMem)
		if annotation != "" {
			annotations[container.Name] = append(annotations[container.Name], fmt.Sprintf("memory: %s", annotation))
		}
		resources[i] = vpa_api_util.ProportionallyCapResourcesToMaxLimit(recommendation.Target, cpuLimit, memLimit, maxCpuLimit, maxMemLimit)
	}
	return resources
}

func (p *recommendationProvider) getMatchingVPA(pod *v1.Pod) *vpa_types.VerticalPodAutoscaler {
	configs, err := p.vpaLister.VerticalPodAutoscalers(pod.Namespace).List(labels.Everything())
	if err != nil {
		glog.Errorf("failed to get vpa configs: %v", err)
		return nil
	}
	onConfigs := make([]*vpa_api_util.VpaWithSelector, 0)
	for _, vpaConfig := range configs {
		if vpa_api_util.GetUpdateMode(vpaConfig) == vpa_types.UpdateModeOff {
			continue
		}
		selector, err := p.selectorFetcher.Fetch(vpaConfig)
		if err != nil {
			glog.V(3).Infof("skipping VPA object %v because we cannot fetch selector", vpaConfig.Name)
			continue
		}
		onConfigs = append(onConfigs, &vpa_api_util.VpaWithSelector{
			Vpa:      vpaConfig,
			Selector: selector,
		})
	}
	glog.V(2).Infof("Let's choose from %d configs for pod %s/%s", len(onConfigs), pod.Namespace, pod.Name)
	result := vpa_api_util.GetControllingVPAForPod(pod, onConfigs)
	if result != nil {
		return result.Vpa
	}
	return nil
}

// GetContainersResourcesForPod returns recommended request for a given pod, annotations and name of controlling VPA.
// The returned slice corresponds 1-1 to containers in the Pod.
func (p *recommendationProvider) GetContainersResourcesForPod(pod *v1.Pod) ([]vpa_api_util.ContainerResources, vpa_api_util.ContainerToAnnotationsMap, string, error) {
	glog.V(2).Infof("updating requirements for pod %s.", pod.Name)
	vpaConfig := p.getMatchingVPA(pod)
	if vpaConfig == nil {
		glog.V(2).Infof("no matching VPA found for pod %s", pod.Name)
		return nil, nil, "", nil
	}

	var annotations vpa_api_util.ContainerToAnnotationsMap
	recommendedPodResources := &vpa_types.RecommendedPodResources{}

	if vpaConfig.Status.Recommendation != nil {
		var err error
		recommendedPodResources, annotations, err = p.recommendationProcessor.Apply(vpaConfig.Status.Recommendation, vpaConfig.Spec.ResourcePolicy, vpaConfig.Status.Conditions, pod)
		if err != nil {
			glog.V(2).Infof("cannot process recommendation for pod %s", pod.Name)
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
