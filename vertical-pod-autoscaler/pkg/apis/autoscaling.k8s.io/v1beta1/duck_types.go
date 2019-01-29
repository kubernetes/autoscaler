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

	//GetConditions() []VerticalPodAutoscalerCondition
}

/*type ContainerResourcePolicy interface {
	GetContainerName() string
	GetMaxAllowed(name v1.ResourceName) (resource.Quantity, bool)
	GetMinAllowed(name v1.ResourceName) (resource.Quantity, bool)
}*/

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

/*
type VPAAdapter struct {
	*VerticalPodAutoscaler
}

func (a *VPAAdapter) GetSelector() *meta.LabelSelector {
	return a.VerticalPodAutoscaler.Spec.Selector
}

func (a *VPAAdapter) GetRecommendation() *RecommendedPodResources {
	return a.VerticalPodAutoscaler.Status.Recommendation
}

func (a *VPAAdapter) GetContainerResourcePolicy(containerName string) *ContainerResourcePolicy {
	var defaultPolicy *ContainerResourcePolicy
	if a != nil && a.VerticalPodAutoscaler.Spec.ResourcePolicy != nil {
		return GetContainerResourcePolicy(containerName, a.VerticalPodAutoscaler.Spec.ResourcePolicy)
	}
	return defaultPolicy
}
*/
/*
func (a *VPAAdapter) GetConditions() []VerticalPodAutoscalerCondition {
	return a.VerticalPodAutoscaler.Status.Conditions
}
*/

func (v *ClusterProportionalScaler) GetSelector() *meta.LabelSelector {
	return v.Spec.Selector
}

func (v *ClusterProportionalScaler) GetRecommendation() *RecommendedPodResources {
	return v.Status.Recommendation
}

func (v *ClusterProportionalScaler) GetContainerResourcePolicy(containerName string) *ContainerResourcePolicy {
	var defaultPolicy *ContainerResourcePolicy
	if v != nil && v.Spec.ResourcePolicy != nil {
		containerPolicies := v.Spec.ResourcePolicy.ContainerPolicies
		for i := range containerPolicies {
			if containerPolicies[i].ContainerName == containerName {
				return &containerPolicies[i].ContainerResourcePolicy
			}
			if containerPolicies[i].ContainerName == DefaultContainerResourcePolicy {
				defaultPolicy = &containerPolicies[i].ContainerResourcePolicy
			}
		}
	}
	return defaultPolicy
}

/*
var _ ScalingPolicy = &VPAAdapter{}

type CPSAdapter struct {
	*ClusterProportionalScaler
}

func (a *CPSAdapter) GetSelector() *meta.LabelSelector {
	return a.ClusterProportionalScaler.Spec.Selector
}

func (a *CPSAdapter) GetRecommendation() *RecommendedPodResources {
	return a.ClusterProportionalScaler.Status.Recommendation
}

func (a *CPSAdapter) GetContainerResourcePolicy(containerName string) *ContainerResourcePolicy {
	if a != nil && a.ClusterProportionalScaler.Spec.ResourcePolicy != nil {
		return GetCPSContainerResourcePolicy(containerName, a.ClusterProportionalScaler.Spec.ResourcePolicy)
	}
	return nil
}
*/
/*
func (a *CPSAdapter) GetConditions() []VerticalPodAutoscalerCondition {
	return a.ClusterProportionalScaler.Status.Conditions
}
*/

/*
var _ ScalingPolicy = &CPSAdapter{}
*/

// GetContainerResourcePolicy returns the ContainerResourcePolicy for a given policy
// and container name. It returns nil if there is no policy specified for the container.
func GetContainerResourcePolicy(containerName string, policy ScalingPolicy) *ContainerResourcePolicy {
	var defaultPolicy *ContainerResourcePolicy
	if policy != nil {
		return policy.GetContainerResourcePolicy(containerName)
	}
	return defaultPolicy
}
