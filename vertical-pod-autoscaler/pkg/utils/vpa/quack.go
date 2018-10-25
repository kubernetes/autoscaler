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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
)

// TODO: Rename to ScalingPolicy

// ScalerDuck defines the common interface for scaling policies
type ScalerDuck interface {
	GetNamespace() string
	GetName() string
	GetCreationTimestamp() meta.Time

	GetSelector() *meta.LabelSelector

	GetRecommendation() *vpa_types.RecommendedPodResources

	// GetContainerResourcePolicy returns the ContainerResourcePolicy for a given policy
	// and container name. It returns nil if there is no policy specified for the container.
	GetContainerResourcePolicy(name string) *vpa_types.ContainerResourcePolicy

	//GetConditions() []vpa_types.VerticalPodAutoscalerCondition
}

/*type ContainerResourcePolicy interface {
	GetContainerName() string
	GetMaxAllowed(name v1.ResourceName) (resource.Quantity, bool)
	GetMinAllowed(name v1.ResourceName) (resource.Quantity, bool)
}*/

type VPAAdapter struct {
	*vpa_types.VerticalPodAutoscaler
}

func (a *VPAAdapter) GetSelector() *meta.LabelSelector {
	return a.VerticalPodAutoscaler.Spec.Selector
}

func (a *VPAAdapter) GetRecommendation() *vpa_types.RecommendedPodResources {
	return a.VerticalPodAutoscaler.Status.Recommendation
}

func (a *VPAAdapter) GetContainerResourcePolicy(containerName string) *vpa_types.ContainerResourcePolicy {
	var defaultPolicy *vpa_types.ContainerResourcePolicy
	if a != nil && a.VerticalPodAutoscaler.Spec.ResourcePolicy != nil {
		return GetContainerResourcePolicy(containerName, a.VerticalPodAutoscaler.Spec.ResourcePolicy)
	}
	return defaultPolicy
}

/*
func (a *VPAAdapter) GetConditions() []vpa_types.VerticalPodAutoscalerCondition {
	return a.VerticalPodAutoscaler.Status.Conditions
}
*/

var _ ScalerDuck = &VPAAdapter{}

type CPSAdapter struct {
	*vpa_types.ClusterProportionalScaler
}

func (a *CPSAdapter) GetSelector() *meta.LabelSelector {
	return a.ClusterProportionalScaler.Spec.Selector
}

func (a *CPSAdapter) GetRecommendation() *vpa_types.RecommendedPodResources {
	return a.ClusterProportionalScaler.Status.Recommendation
}

func (a *CPSAdapter) GetContainerResourcePolicy(containerName string) *vpa_types.ContainerResourcePolicy {
	if a != nil && a.ClusterProportionalScaler.Spec.ResourcePolicy != nil {
		return GetCPSContainerResourcePolicy(containerName, a.ClusterProportionalScaler.Spec.ResourcePolicy)
	}
	return nil
}

/*
func (a *CPSAdapter) GetConditions() []vpa_types.VerticalPodAutoscalerCondition {
	return a.ClusterProportionalScaler.Status.Conditions
}
*/

var _ ScalerDuck = &CPSAdapter{}
