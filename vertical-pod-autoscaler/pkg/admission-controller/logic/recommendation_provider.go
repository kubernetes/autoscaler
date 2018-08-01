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
	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/poc.autoscaling.k8s.io/v1alpha1"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// ContainerResources holds request and limit resources for container
type ContainerResources struct {
	Requests v1.ResourceList
}

func newContainerResources() ContainerResources {
	return ContainerResources{Requests: v1.ResourceList{}}
}

// RecommendationProvider gets current recommendation and limits for the given pod.
type RecommendationProvider interface {
	GetContainersResourcesForPod(pod *v1.Pod) ([]ContainerResources, string, error)
}

type recommendationProvider struct {
	vpaLister               vpa_lister.VerticalPodAutoscalerLister
	recommendationProcessor vpa_api_util.RecommendationProcessor
}

// NewRecommendationProvider constructs the recommendation provider that list VPAs and can be used to determine recommendations for pods.
func NewRecommendationProvider(vpaLister vpa_lister.VerticalPodAutoscalerLister, recommendationProcessor vpa_api_util.RecommendationProcessor) *recommendationProvider {
	return &recommendationProvider{vpaLister: vpaLister,
		recommendationProcessor: recommendationProcessor}
}

// getContainersResources returns the recommended resources and limits for each container in the given pod in the same order they are specified in the pod.Spec.
func getContainersResources(pod *v1.Pod, podRecommendation vpa_types.RecommendedPodResources, policy *vpa_types.PodResourcePolicy) []ContainerResources {
	resources := make([]ContainerResources, len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {
		resources[i] = newContainerResources()

		recommendation := vpa_api_util.GetRecommendationForContainer(container.Name, &podRecommendation)
		if recommendation == nil {
			glog.V(2).Infof("no matching recommendation found for container %s", container.Name)
			continue
		}
		resources[i].Requests = recommendation.Target
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
		if vpa_api_util.GetUpdateMode(vpaConfig) == vpa_types.UpdateModeOff {
			continue
		}
		onConfigs = append(onConfigs, vpaConfig)
	}
	glog.Infof("Let's choose from %d configs", len(onConfigs))
	return vpa_api_util.GetControllingVPAForPod(pod, onConfigs)
}

// GetContainersResourcesForPod returns recommended request and limits for a given pod and name of controlling VPA.
// The returned slice corresponds 1-1 to containers in the Pod.
func (p *recommendationProvider) GetContainersResourcesForPod(pod *v1.Pod) ([]ContainerResources, string, error) {
	glog.V(2).Infof("updating requirements for pod %s.", pod.Name)
	vpaConfig := p.getMatchingVPA(pod)
	if vpaConfig == nil {
		glog.V(2).Infof("no matching VPA found for pod %s", pod.Name)
		return nil, "", nil
	}
	recommendedPodResources := &vpa_types.RecommendedPodResources{}
	if vpaConfig.Status.Recommendation != nil {
		var err error
		recommendedPodResources, err = p.recommendationProcessor.Apply(vpaConfig.Status.Recommendation, vpaConfig.Spec.ResourcePolicy, pod)
		if err != nil {
			glog.V(2).Infof("cannot process recommendation for pod %s", pod.Name)
			return nil, "", err
		}
	}
	containerResources := getContainersResources(pod, *recommendedPodResources, vpaConfig.Spec.ResourcePolicy)
	return containerResources, vpaConfig.Name, nil
}
