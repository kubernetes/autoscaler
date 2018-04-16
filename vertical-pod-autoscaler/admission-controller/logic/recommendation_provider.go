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
	"math"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/poc.autoscaling.k8s.io/v1alpha1"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	// MemoryLimitBumpUpRatio specifies how much memory will be available above memory request.
	// Limit is set to trigger OOMs.
	MemoryLimitBumpUpRatio float64 = 1.2
	// MemoryLimitMinBumpUp specifies minimal amount of memory above memory request.
	// Limit is set to trigger OOMs.
	MemoryLimitMinBumpUp float64 = 100 * 1024 * 1024 // 100MB
)

// ContainerResources holds request and limit resources for container
type ContainerResources struct {
	Requests, Limits v1.ResourceList
}

func newContainerResources() ContainerResources {
	return ContainerResources{Requests: v1.ResourceList{}, Limits: v1.ResourceList{}}
}

// RecommendationProvider gets current recommendation and limits for the given pod.
type RecommendationProvider interface {
	GetContainersResourcesForPod(pod *v1.Pod) ([]ContainerResources, error)
}

type recommendationProvider struct {
	vpaLister vpa_lister.VerticalPodAutoscalerLister
}

// NewRecommendationProvider constructs the recommendation provider that list VPAs and can be used to determine recommendations for pods.
func NewRecommendationProvider(vpaLister vpa_lister.VerticalPodAutoscalerLister) *recommendationProvider {
	return &recommendationProvider{vpaLister: vpaLister}
}

// getMemoryLimit returns a limit that is proportionally (with min step) higher than the request. Limit is set to trigger OOMs.
func getMemoryLimit(resources v1.ResourceList) *resource.Quantity {
	memory, found := resources[v1.ResourceMemory]
	if !found {
		return nil
	}
	limit := float64(memory.Value())
	limit = math.Max(limit+MemoryLimitMinBumpUp, limit*MemoryLimitBumpUpRatio)
	return resource.NewQuantity(int64(limit), memory.Format)
}

// getContainersResources returns the recommended resources and limits for each container in the given pod in the same order they are specified in the pod.Spec.
func getContainersResources(pod *v1.Pod, podRecommendation vpa_types.RecommendedPodResources, policy vpa_types.PodResourcePolicy) []ContainerResources {
	resources := make([]ContainerResources, len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {
		resources[i] = newContainerResources()
		setLimit := func(r v1.ResourceList) {
			memoryLimit := getMemoryLimit(r)
			if memoryLimit != nil {
				resources[i].Limits[v1.ResourceMemory] = *memoryLimit
				vpa_api_util.ApplyVPAContainerPolicy(resources[i].Limits, container, &policy)
			}
		}

		recommendation, err := vpa_api_util.GetCappedRecommendationForContainer(container, &podRecommendation, &policy)
		if err != nil {
			glog.V(2).Infof("%v", err)
			setLimit(pod.Spec.Containers[i].Resources.Requests)
			continue
		}
		resources[i].Requests = recommendation.Target
		setLimit(resources[i].Requests)
	}
	return resources
}

func (p *recommendationProvider) getMatchingVPA(pod *v1.Pod) *vpa_types.VerticalPodAutoscaler {
	configs, err := p.vpaLister.VerticalPodAutoscalers(pod.Namespace).List(labels.Everything())
	if err != nil {
		glog.Error("failed to get vpa configs: %v", err)
		return nil
	}
	onConfigs := make([]*vpa_types.VerticalPodAutoscaler, 0)
	for _, vpaConfig := range configs {
		if vpaConfig.Spec.UpdatePolicy.UpdateMode == vpa_types.UpdateModeOff {
			continue
		}
		onConfigs = append(onConfigs, vpaConfig)
	}
	glog.Infof("Let's choose from %d configs", len(onConfigs))
	return vpa_api_util.GetControllingVPAForPod(pod, onConfigs)
}

// GetContainersResourcesForPod returns recommended request and limits for a given pod. The returned slice corresponds 1-1 to containers in the Pod.
func (p *recommendationProvider) GetContainersResourcesForPod(pod *v1.Pod) ([]ContainerResources, error) {
	glog.V(2).Infof("updating requirements for pod %s.", pod.Name)
	vpaConfig := p.getMatchingVPA(pod)
	if vpaConfig == nil {
		glog.V(2).Infof("no matching VPA found for pod %s", pod.Name)
		return nil, nil
	}
	containerResources := getContainersResources(pod, vpaConfig.Status.Recommendation, vpaConfig.Spec.ResourcePolicy)
	return containerResources, nil
}
