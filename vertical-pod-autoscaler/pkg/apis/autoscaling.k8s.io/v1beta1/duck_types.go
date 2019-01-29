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

package v1beta1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScalingPolicy defines the common interface for scaling policies
type ScalingPolicy interface {
	GetNamespace() string
	GetName() string
	GetCreationTimestamp() meta.Time

	GetSelector() *meta.LabelSelector

	GetRecommendation() *RecommendedPodResources

	// GetContainerResourcePolicy returns the ContainerResourcePolicy for a given policy
	// and container name. It returns nil if there is no policy specified for the container.
	GetContainerResourcePolicy(name string) *ContainerResourcePolicy
}

func (v *VerticalPodAutoscaler) GetSelector() *meta.LabelSelector {
	return v.Spec.Selector
}

func (v *VerticalPodAutoscaler) GetRecommendation() *RecommendedPodResources {
	return v.Status.Recommendation
}

func (v *VerticalPodAutoscaler) GetContainerResourcePolicy(containerName string) *ContainerResourcePolicy {
	var defaultPolicy *ContainerResourcePolicy
	if v != nil && v.Spec.ResourcePolicy != nil {
		containerPolicies := v.Spec.ResourcePolicy.ContainerPolicies
		for i := range containerPolicies {
			if containerPolicies[i].ContainerName == containerName {
				return &containerPolicies[i]
			}
			if containerPolicies[i].ContainerName == DefaultContainerResourcePolicy {
				defaultPolicy = &containerPolicies[i]
			}
		}
	}
	return defaultPolicy
}

// GetContainerResourcePolicy returns the ContainerResourcePolicy for a given policy
// and container name. It returns nil if there is no policy specified for the container.
func GetContainerResourcePolicy(containerName string, policy ScalingPolicy) *ContainerResourcePolicy {
	var defaultPolicy *ContainerResourcePolicy
	if policy != nil {
		return policy.GetContainerResourcePolicy(containerName)
	}
	return defaultPolicy
}
