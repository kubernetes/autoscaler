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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpalister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/poc.autoscaling.k8s.io/v1alpha1"
)

// RecommendationProvider gets current recommendation for the given pod.
type RecommendationProvider interface {
	GetRequestForPod(pod *v1.Pod) ([]v1.ResourceList, error)
}

type recommendationProvider struct {
	vpaLister vpalister.VerticalPodAutoscalerLister
}

// NewRecommendationProvider constructs the recommendation provider that list VPAs and can be used to determine recommendations for pods.
func NewRecommendationProvider(vpaLister vpalister.VerticalPodAutoscalerLister) *recommendationProvider {
	return &recommendationProvider{vpaLister: vpaLister}
}

// getRecomendedResources overwrites pod resources Request field with recommended values.
func getRecomendedResources(pod *v1.Pod, podRecommendation v1alpha1.RecommendedPodResources, policy v1alpha1.PodResourcePolicy) []v1.ResourceList {
	res := make([]v1.ResourceList, len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {
		containerRecommendation := getRecommendationForContainer(podRecommendation, container)
		if containerRecommendation == nil {
			continue
		}
		containerPolicy := getContainerPolicy(container.Name, &policy)
		applyVPAPolicy(containerRecommendation, containerPolicy)
		res[i] = make(v1.ResourceList)
		for resource, recommended := range containerRecommendation.Target {
			requested, exists := container.Resources.Requests[resource]
			if exists {
				// overwriting existing resource spec
				glog.V(2).Infof("updating resources request for pod %v container %v resource %v old value: %v new value: %v",
					pod.Name, container.Name, resource, requested, recommended)
			} else {
				// adding new resource spec
				glog.V(2).Infof("updating resources request for pod %v container %v resource %v old value: none new value: %v",
					pod.Name, container.Name, resource, recommended)
			}

			res[i][resource] = recommended
		}
	}
	return res
}

// applyVPAPolicy updates recommendation if recommended resources exceed limits defined in VPA resources policy
func applyVPAPolicy(recommendation *v1alpha1.RecommendedContainerResources, policy *v1alpha1.ContainerResourcePolicy) {
	for resourceName, recommended := range recommendation.Target {
		if policy == nil {
			continue
		}
		min, found := policy.MinAllowed[resourceName]
		if found && !min.IsZero() && recommended.Value() < min.Value() {
			glog.Warningf("recommendation outside of policy bounds : min value : %v recommended : %v",
				min.Value(), recommended)
			recommendation.Target[resourceName] = min
		}
		max, found := policy.MaxAllowed[resourceName]
		if found && !max.IsZero() && recommended.Value() > max.Value() {
			glog.Warningf("recommendation outside of policy bounds : max value : %v recommended : %v",
				max.Value(), recommended)
			recommendation.Target[resourceName] = max
		}
	}
}

func getRecommendationForContainer(recommendation v1alpha1.RecommendedPodResources, container v1.Container) *v1alpha1.RecommendedContainerResources {
	for i, containerRec := range recommendation.ContainerRecommendations {
		if containerRec.Name == container.Name {
			return &recommendation.ContainerRecommendations[i]
		}
	}
	return nil
}

func getContainerPolicy(containerName string, policy *v1alpha1.PodResourcePolicy) *v1alpha1.ContainerResourcePolicy {
	if policy != nil {
		for i, container := range policy.ContainerPolicies {
			if containerName == container.Name {
				return &policy.ContainerPolicies[i]
			}
		}
	}
	return nil
}

func (p *recommendationProvider) getMatchingVPA(pod *v1.Pod) *v1alpha1.VerticalPodAutoscaler {
	configs, err := p.vpaLister.VerticalPodAutoscalers(pod.Namespace).List(labels.Everything())
	if err != nil {
		glog.Error("failed to get vpa configs: %v", err)
		return nil
	}
	for _, vpaConfig := range configs {
		selector, err := metav1.LabelSelectorAsSelector(vpaConfig.Spec.Selector)
		if err != nil {
			continue
		}
		glog.Infof("trying to match %v - selector %v", vpaConfig.Name, selector)
		if selector.Matches(labels.Set(pod.GetLabels())) {
			return vpaConfig
		}
		glog.V(4).Infof("pod %v didn't match selector. Selector: %+v, labels: %+v", pod.Name, selector, pod.GetLabels())
	}
	return nil
}

// GetRequestForPod returns recommended request for a given pod. The returned slice corresponds 1-1 to containers in the Pod.
func (p *recommendationProvider) GetRequestForPod(pod *v1.Pod) ([]v1.ResourceList, error) {
	glog.V(2).Infof("updating requirements for pod %s.", pod.Name)

	vpaConfig := p.getMatchingVPA(pod)
	if vpaConfig == nil {
		glog.V(2).Infof("no matching VPA found for pod %s", pod.Name)
		return nil, nil
	}

	return getRecomendedResources(pod, vpaConfig.Status.Recommendation, vpaConfig.Spec.ResourcePolicy), nil
}
