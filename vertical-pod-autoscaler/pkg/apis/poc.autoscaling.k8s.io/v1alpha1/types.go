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

// Package v1alpha1 contains definitions of Vertical Pod Autoscaler related objects.
package v1alpha1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VerticalPodAutoscaler is the configuration for a vertical pod
// autoscaler, which automatically manages pod resources based on historical and
// real time resource utilization.
type VerticalPodAutoscaler struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the behavior of the autoscaler.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	// +optional
	Spec VerticalPodAutoscalerSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Current information about the autoscaler.
	// +optional
	Status VerticalPodAutoscalerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VerticalPodAutoscalerList is a list of VerticalPodAutoscaler objects.
type VerticalPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []VerticalPodAutoscaler `json:"items"`
}

// VerticalPodAutoscalerSpec is the specification of the behavior of the autoscaler.
type VerticalPodAutoscalerSpec struct {
	// A label query that determines the set of pods controlled by the Autoscaler.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,1,opt,name=selector"`

	// Describes the rules on how changes are applied to the pods.
	// +optional
	UpdatePolicy PodUpdatePolicy `json:"updatePolicy,omitempty" protobuf:"bytes,2,opt,name=updatePolicy"`

	// Controls how the autoscaler computes recommended resources.
	// +optional
	ResourcePolicy PodResourcePolicy `json:"podResourcePolicy,omitempty" protobuf:"bytes,3,opt,name=podResourcePolicy"`
}

// VerticalPodAutoscalerStatus describes the runtime state of the autoscaler.
type VerticalPodAutoscalerStatus struct {
	// The time when the status was last refreshed.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,1,opt,name=lastUpdateTime"`
	// The most recently computed amount of resources recommended by the
	// autoscaler for the controlled pods.
	// +optional
	Recommendation RecommendedPodResources `json:"recommendation,omitempty" protobuf:"bytes,2,opt,name=recommendation"`

	// Conditions is the set of conditions required for this autoscaler to scale its target,
	// and indicates whether or not those conditions are met.
	Conditions []VerticalPodAutoscalerCondition `json:"conditions" protobuf:"bytes,3,rep,name=conditions"`
}

// VerticalPodAutoscalerConditionType are the valid conditions of
// a VerticalPodAutoscaler.
type VerticalPodAutoscalerConditionType string

var (
	// RecommendationProvided indicates whether the VPA recommender was able to calculate a recommendation.
	RecommendationProvided VerticalPodAutoscalerConditionType = "RecommendationProvided"
)

// VerticalPodAutoscalerCondition describes the state of
// a VerticalPodAutoscaler at a certain point.
type VerticalPodAutoscalerCondition struct {
	// type describes the current condition
	Type VerticalPodAutoscalerConditionType `json:"type" protobuf:"bytes,1,name=type"`
	// status is the status of the condition (True, False, Unknown)
	Status apiv1.ConditionStatus `json:"status" protobuf:"bytes,2,name=status"`
	// lastTransitionTime is the last time the condition transitioned from
	// one status to another
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// reason is the reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// message is a human-readable explanation containing details about
	// the transition
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// UpdateMode controls when autoscaler applies changes to the pod resoures.
type UpdateMode string

const (
	// UpdateModeOff means that autoscaler never changes Pod resources.
	// The recommender still sets the recommended resources in the
	// VerticalPodAutoscaler object. This can be used for a "dry run".
	UpdateModeOff UpdateMode = "Off"
	// UpdateModeInitial means that autoscaler only assigns resources on pod
	// creation and does not change them during the lifetime of the pod.
	UpdateModeInitial UpdateMode = "Initial"
	// UpdateModeAuto means that autoscaler assigns resources on pod creation
	// and additionally can update them during the lifetime of the pod,
	// including evicting / rescheduling the pod.
	UpdateModeAuto UpdateMode = "Auto"
)

// PodUpdatePolicy describes the rules on how changes are applied to the pods.
type PodUpdatePolicy struct {
	// Controls when autoscaler applies changes to the pod resoures.
	// +optional
	UpdateMode UpdateMode `json:"updateMode,omitempty" protobuf:"bytes,1,opt,name=updateMode"`
}

const (
	// DefaultContainerResourcePolicy can be passed as
	// ContainerResourcePolicy.Name to specify the default policy.
	DefaultContainerResourcePolicy = "*"
)

// ContainerResourcePolicy controls how autoscaler computes the recommended
// resources for a specific container.
type ContainerResourcePolicy struct {
	// Name of the container or DefaultContainerResourcePolicy, in which
	// case the policy is used by the containers that don't have their own
	// policy specified.
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Whether autoscaler is enabled for the container. Defaults to "On".
	// +optional
	Mode ContainerScalingMode `json:"mode,omitempty" protobuf:"bytes,2,opt,name=mode"`
	// Specifies the minimal amount of resources that will be recommended
	// for the container.
	// +optional
	MinAllowed apiv1.ResourceList `json:"minAllowed,omitempty" protobuf:"bytes,3,rep,name=minAllowed,casttype=ResourceList,castkey=ResourceName"`
	// Specifies the maximum amount of resources that will be recommended
	// for the container.
	// +optional
	MaxAllowed apiv1.ResourceList `json:"maxAllowed,omitempty" protobuf:"bytes,4,rep,name=maxAllowed,casttype=ResourceList,castkey=ResourceName"`
}

// PodResourcePolicy controls how autoscaler computes the recommended resources
// for containers belonging to the pod.
type PodResourcePolicy struct {
	// Per-container resource policies.
	ContainerPolicies []ContainerResourcePolicy `json:"containerPolicies" protobuf:"bytes,1,rep,name=containerPolicies"`
}

// ContainerScalingMode controls whether autoscaler is enabled for a specific
// container.
type ContainerScalingMode string

const (
	// ContainerScalingModeOn means autoscaling is enabled for a container.
	ContainerScalingModeOn ContainerScalingMode = "On"
	// ContainerScalingModeOff means autoscaling is disabled for a container.
	ContainerScalingModeOff ContainerScalingMode = "Off"
)

// RecommendedPodResources is the recommendation of resources computed by
// autoscaler.
type RecommendedPodResources struct {
	// Resources recommended by the autoscaler for each container.
	ContainerRecommendations []RecommendedContainerResources `json:"containerRecommendations" protobuf:"bytes,1,rep,name=containerRecommendations"`
}

// RecommendedContainerResources is the recommendation of resources computed by
// autoscaler for a specific container. Respects the container resource policy
// if present in the spec.
type RecommendedContainerResources struct {
	// Name of the container.
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Recommended amount of resources.
	Target apiv1.ResourceList `json:"target,omitempty" protobuf:"bytes,2,rep,name=target,casttype=ResourceList,castkey=ResourceName"`
	// Minimum recommended amount of resources.
	// Running the application with less resources is likely to have
	// significant impact on performance/availability.
	// +optional
	MinRecommended apiv1.ResourceList `json:"minRecommended,omitempty" protobuf:"bytes,3,rep,name=minRecommended,casttype=ResourceList,castkey=ResourceName"`
	// Maximum recommended amount of resources.
	// Any resources allocated beyond this value are likely wasted.
	// +optional
	MaxRecommended apiv1.ResourceList `json:"maxRecommended,omitempty" protobuf:"bytes,3,rep,name=maxRecommended,casttype=ResourceList,castkey=ResourceName"`
}
