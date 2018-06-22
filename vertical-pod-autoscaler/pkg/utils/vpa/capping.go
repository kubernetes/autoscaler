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

package api

import (
	"fmt"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
)

// NewCappingRecommendationProcessor constructs new RecommendationsProcessor that adjusts recommendation
// for given pod to obey VPA resources policy and container limits
func NewCappingRecommendationProcessor() RecommendationProcessor {
	return &cappingRecommendationProcessor{}
}

type cappingRecommendationProcessor struct{}

// Apply returns a recommendation for the given pod, adjusted to obey policy and limits.
func (c *cappingRecommendationProcessor) Apply(
	podRecommendation *vpa_types.RecommendedPodResources, policy *vpa_types.PodResourcePolicy, pod *apiv1.Pod) (*vpa_types.RecommendedPodResources, error) {

	if podRecommendation == nil && policy == nil {
		// If there is no recommendation and no policies have been defined then no recommendation can be computed.
		return nil, nil
	}
	if podRecommendation == nil {
		// Policies have been specified. Create an empty recommendation so that the policies can be applied correctly.
		podRecommendation = new(vpa_types.RecommendedPodResources)
	}
	updatedRecommendations := []vpa_types.RecommendedContainerResources{}
	for _, containerRecommendation := range podRecommendation.ContainerRecommendations {
		container := getContainer(containerRecommendation.ContainerName, pod)
		if container == nil {
			glog.V(2).Infof("no matching Container found for recommendation %s", containerRecommendation.ContainerName)
			continue
		}
		updatedContainerResources, err := getCappedRecommendationForContainer(*container, &containerRecommendation, policy)
		if err != nil {
			return nil, fmt.Errorf("cannot update recommendation for container name %v", container.Name)
		}
		updatedRecommendations = append(updatedRecommendations, *updatedContainerResources)
	}
	return &vpa_types.RecommendedPodResources{updatedRecommendations}, nil
}

// getCappedRecommendationForContainer returns a recommendation for the given container, adjusted to obey policy and limits.
func getCappedRecommendationForContainer(
	container apiv1.Container,
	containerRecommendation *vpa_types.RecommendedContainerResources,
	policy *vpa_types.PodResourcePolicy) (*vpa_types.RecommendedContainerResources, error) {
	if containerRecommendation == nil {
		return nil, fmt.Errorf("no recommendation available for container name %v", container.Name)
	}
	// containerPolicy can be nil (user does not have to configure it).
	containerPolicy := GetContainerResourcePolicy(container.Name, policy)

	cappedRecommendations := containerRecommendation.DeepCopy()
	cappedRecommendationsList := []apiv1.ResourceList{
		cappedRecommendations.Target,
		cappedRecommendations.LowerBound,
		cappedRecommendations.UpperBound,
	}

	for _, cappedRecommendation := range cappedRecommendationsList {
		applyVPAPolicy(cappedRecommendation, containerPolicy)
		// TODO: If limits and policy are conflicting, set some condition on the VPA.
		capRecommendationToContainerLimit(cappedRecommendation, container)
	}
	return cappedRecommendations, nil
}

// capRecommendationToContainerLimit makes sure recommendation is not above current limit for the container.
func capRecommendationToContainerLimit(recommendation apiv1.ResourceList, container apiv1.Container) {
	// Iterate over limits set in the container. Unset means Infinite limit.
	for resourceName, limit := range container.Resources.Limits {
		recommendedValue, found := recommendation[resourceName]
		if found && recommendedValue.MilliValue() > limit.MilliValue() {
			recommendation[resourceName] = limit
		}
	}
}

// applyVPAPolicy updates recommendation if recommended resources are outside of limits defined in VPA resources policy
func applyVPAPolicy(recommendation apiv1.ResourceList, policy *vpa_types.ContainerResourcePolicy) {
	if policy == nil {
		return
	}
	for resourceName, recommended := range recommendation {
		min, found := policy.MinAllowed[resourceName]
		if found && !min.IsZero() && recommended.MilliValue() < min.MilliValue() {
			recommendation[resourceName] = min
		}
		max, found := policy.MaxAllowed[resourceName]
		if found && !max.IsZero() && recommended.MilliValue() > max.MilliValue() {
			recommendation[resourceName] = max
		}
	}
}

// ApplyVPAContainerPolicy enforces min/max policy on resources
func ApplyVPAContainerPolicy(resources apiv1.ResourceList, container apiv1.Container, policy *vpa_types.PodResourcePolicy) {
	// containerPolicy can be nil (user does not have to configure it).
	containerPolicy := GetContainerResourcePolicy(container.Name, policy)
	applyVPAPolicy(resources, containerPolicy)
}

// GetRecommendationForContainer returns recommendation for given container name
func GetRecommendationForContainer(containerName string, recommendation *vpa_types.RecommendedPodResources) *vpa_types.RecommendedContainerResources {
	if recommendation != nil {
		for i, containerRec := range recommendation.ContainerRecommendations {
			if containerRec.ContainerName == containerName {
				recommendationCopy := recommendation.ContainerRecommendations[i]
				return &recommendationCopy
			}
		}
	}
	return nil
}

func getContainer(containerName string, pod *apiv1.Pod) *apiv1.Container {
	for i, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return &pod.Spec.Containers[i]
		}
	}
	return nil
}
