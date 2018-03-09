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

	apiv1 "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
)

// GetCappedRecommendationForContainer returns a recommendation extracted for the given container, adjusted to obey policy and limits.
func GetCappedRecommendationForContainer(
	container apiv1.Container,
	podRecommendation *vpa_types.RecommendedPodResources,
	policy *vpa_types.PodResourcePolicy) (*vpa_types.RecommendedContainerResources, error) {
	containerRecommendation := getRecommendationForContainer(podRecommendation, container)
	if containerRecommendation == nil {
		return nil, fmt.Errorf("no recommendation available for container name %v", container.Name)
	}
	// containerPolicy can be nil (user does not have to configure it).
	containerPolicy := getContainerPolicy(container.Name, policy)

	cappedRecommendations := containerRecommendation.DeepCopy()
	cappedRecommendationsList := []apiv1.ResourceList{
		cappedRecommendations.Target,
		cappedRecommendations.MinRecommended,
		cappedRecommendations.MaxRecommended,
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

func getRecommendationForContainer(recommendation *vpa_types.RecommendedPodResources, container apiv1.Container) *vpa_types.RecommendedContainerResources {
	if recommendation != nil {
		for i, containerRec := range recommendation.ContainerRecommendations {
			if containerRec.Name == container.Name {
				recommendationCopy := recommendation.ContainerRecommendations[i]
				return &recommendationCopy
			}
		}
	}
	return nil
}

func getContainerPolicy(containerName string, policy *vpa_types.PodResourcePolicy) *vpa_types.ContainerResourcePolicy {
	if policy != nil {
		for i, container := range policy.ContainerPolicies {
			if containerName == container.Name {
				return &policy.ContainerPolicies[i]
			}
		}
	}
	return nil
}
