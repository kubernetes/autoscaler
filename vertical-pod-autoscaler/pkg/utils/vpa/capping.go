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

// GetCappedRecommendationForContainer returns a recommendation extracted for the given container, adjusted to obey policy and limits.
func GetCappedRecommendationForContainer(
	container apiv1.Container,
	podRecommendation *vpa_types.RecommendedPodResources,
	policy *vpa_types.PodResourcePolicy) (apiv1.ResourceList, error) {
	containerRecommendation := getRecommendationForContainer(podRecommendation, container)
	if containerRecommendation == nil {
		return nil, fmt.Errorf("no recommendation available for container name %v", container.Name)
	}
	capRecommendationToContainerLimit(containerRecommendation, container)
	containerPolicy := getContainerPolicy(container.Name, policy)
	applyVPAPolicy(containerRecommendation, containerPolicy)
	res := make(apiv1.ResourceList)
	for resource, recommended := range containerRecommendation.Target {
		requested, exists := container.Resources.Requests[resource]
		if exists {
			// overwriting existing resource spec
			glog.V(2).Infof("updating resources request for pod %v container %v resource %v old value: %v new value: %v",
				"", container.Name, resource, requested, recommended)
		} else {
			// adding new resource spec
			glog.V(2).Infof("updating resources request for pod %v container %v resource %v old value: none new value: %v",
				"", container.Name, resource, recommended)
		}

		res[resource] = recommended
	}
	return res, nil
}

// capRecommendationToContainerLimit makes sure recommendation is not above current limit for the container.
func capRecommendationToContainerLimit(recommendation *vpa_types.RecommendedContainerResources, container apiv1.Container) {
	// Iterate over limits set in the container. Unset means Infinite limit.
	for resourceName, limit := range container.Resources.Limits {
		target, found := recommendation.Target[resourceName]
		if found && target.MilliValue() > limit.MilliValue() {
			recommendation.Target[resourceName] = limit
		}
	}
}

// applyVPAPolicy updates recommendation if recommended resources exceed limits defined in VPA resources policy
func applyVPAPolicy(recommendation *vpa_types.RecommendedContainerResources, policy *vpa_types.ContainerResourcePolicy) {
	for resourceName, recommended := range recommendation.Target {
		if policy == nil {
			continue
		}
		min, found := policy.MinAllowed[resourceName]
		if found && !min.IsZero() && recommended.MilliValue() < min.MilliValue() {
			glog.Warningf("recommendation outside of policy bounds : min value : %v recommended : %v",
				min.MilliValue(), recommended)
			recommendation.Target[resourceName] = min
		}
		max, found := policy.MaxAllowed[resourceName]
		if found && !max.IsZero() && recommended.MilliValue() > max.MilliValue() {
			glog.Warningf("recommendation outside of policy bounds : max value : %v recommended : %v",
				max.MilliValue(), recommended)
			recommendation.Target[resourceName] = max
		}
	}
}

func getRecommendationForContainer(recommendation *vpa_types.RecommendedPodResources, container apiv1.Container) *vpa_types.RecommendedContainerResources {
	if recommendation != nil {
		for i, containerRec := range recommendation.ContainerRecommendations {
			if containerRec.Name == container.Name {
				return &recommendation.ContainerRecommendations[i]
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
