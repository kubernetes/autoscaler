/*
Copyright 2019 The Kubernetes Authors.

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

// Package v1 contains definitions of Vertical Pod Autoscaler related objects.
package v1

import (
	autoscaling "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VerticalPodAutoscalerList is a list of VerticalPodAutoscaler objects.
type VerticalPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	// items is the list of vertical pod autoscaler objects.
	Items []VerticalPodAutoscaler `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
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
	Spec VerticalPodAutoscalerSpec `json:"spec" protobuf:"bytes,2,name=spec"`

	// Current information about the autoscaler.
	// +optional
	Status VerticalPodAutoscalerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// VerticalPodAutoscalerSpec is the specification of the behavior of the autoscaler.
type VerticalPodAutoscalerSpec struct {

	// TargetRef points to the controller managing the set of pods for the
	// autoscaler to control - e.g. Deployment, StatefulSet. VerticalPodAutoscaler
	// can be targeted at controller implementing scale subresource (the pod set is
	// retrieved from the controller's ScaleStatus) or some well known controllers
	// (e.g. for DaemonSet the pod set is read from the controller's spec).
	// If VerticalPodAutoscaler cannot use specified target it will report
	// ConfigUnsupported condition.
	// Note that VerticalPodAutoscaler does not require full implementation
	// of scale subresource - it will not use it to modify the replica count.
	// The only thing retrieved is a label selector matching pods grouped by
	// the target resource.
	TargetRef *autoscaling.CrossVersionObjectReference `json:"targetRef" protobuf:"bytes,1,name=targetRef"`

	// Describes the rules on how changes are applied to the pods.
	// If not specified, all fields in the `PodUpdatePolicy` are set to their
	// default values.
	// +optional
	UpdatePolicy *PodUpdatePolicy `json:"updatePolicy,omitempty" protobuf:"bytes,2,opt,name=updatePolicy"`

	// Controls how the autoscaler computes recommended resources.
	// The resource policy may be used to set constraints on the recommendations
	// for individual containers. If not specified, the autoscaler computes recommended
	// resources for all containers in the pod, without additional constraints.
	// +optional
	ResourcePolicy *PodResourcePolicy `json:"resourcePolicy,omitempty" protobuf:"bytes,3,opt,name=resourcePolicy"`
}

// PodUpdatePolicy describes the rules on how changes are applied to the pods.
type PodUpdatePolicy struct {
	// Controls when autoscaler applies changes to the pod resources.
	// The default is 'Auto'.
	// +optional
	UpdateMode *UpdateMode `json:"updateMode,omitempty" protobuf:"bytes,1,opt,name=updateMode"`
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
	// UpdateModeRecreate means that autoscaler assigns resources on pod
	// creation and additionally can update them during the lifetime of the
	// pod by deleting and recreating the pod.
	UpdateModeRecreate UpdateMode = "Recreate"
	// UpdateModeAuto means that autoscaler assigns resources on pod creation
	// and additionally can update them during the lifetime of the pod,
	// using any available update method. Currently this is equivalent to
	// Recreate, which is the only available update method.
	UpdateModeAuto UpdateMode = "Auto"
)

// PodResourcePolicy controls how autoscaler computes the recommended resources
// for containers belonging to the pod. There can be at most one entry for every
// named container and optionally a single wildcard entry with `containerName` = '*',
// which handles all containers that don't have individual policies.
type PodResourcePolicy struct {
	// Per-container resource policies.
	// +optional
	// +patchMergeKey=containerName
	// +patchStrategy=merge
	ContainerPolicies []ContainerResourcePolicy `json:"containerPolicies,omitempty" patchStrategy:"merge" patchMergeKey:"containerName" protobuf:"bytes,1,rep,name=containerPolicies"`
}

// ContainerResourcePolicy controls how autoscaler computes the recommended
// resources for a specific container.
type ContainerResourcePolicy struct {
	// Name of the container or DefaultContainerResourcePolicy, in which
	// case the policy is used by the containers that don't have their own
	// policy specified.
	ContainerName string `json:"containerName,omitempty" protobuf:"bytes,1,opt,name=containerName"`
	// Whether autoscaler is enabled for the container. The default is "Auto".
	// +optional
	Mode *ContainerScalingMode `json:"mode,omitempty" protobuf:"bytes,2,opt,name=mode"`
	// Specifies the minimal amount of resources that will be recommended
	// for the container. The default is no minimum.
	// +optional
	MinAllowed v1.ResourceList `json:"minAllowed,omitempty" protobuf:"bytes,3,rep,name=minAllowed,casttype=ResourceList,castkey=ResourceName"`
	// Specifies the maximum amount of resources that will be recommended
	// for the container. The default is no maximum.
	// +optional
	MaxAllowed v1.ResourceList `json:"maxAllowed,omitempty" protobuf:"bytes,4,rep,name=maxAllowed,casttype=ResourceList,castkey=ResourceName"`
}

const (
	// DefaultContainerResourcePolicy can be passed as
	// ContainerResourcePolicy.ContainerName to specify the default policy.
	DefaultContainerResourcePolicy = "*"
)

// ContainerScalingMode controls whether autoscaler is enabled for a specific
// container.
type ContainerScalingMode string

const (
	// ContainerScalingModeAuto means autoscaling is enabled for a container.
	ContainerScalingModeAuto ContainerScalingMode = "Auto"
	// ContainerScalingModeOff means autoscaling is disabled for a container.
	ContainerScalingModeOff ContainerScalingMode = "Off"
)

// VerticalPodAutoscalerStatus describes the runtime state of the autoscaler.
type VerticalPodAutoscalerStatus struct {
	// The most recently computed amount of resources recommended by the
	// autoscaler for the controlled pods.
	// +optional
	Recommendation *RecommendedPodResources `json:"recommendation,omitempty" protobuf:"bytes,1,opt,name=recommendation"`

	// Conditions is the set of conditions required for this autoscaler to scale its target,
	// and indicates whether or not those conditions are met.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []VerticalPodAutoscalerCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

// RecommendedPodResources is the recommendation of resources computed by
// autoscaler. It contains a recommendation for each container in the pod
// (except for those with `ContainerScalingMode` set to 'Off').
type RecommendedPodResources struct {
	// Resources recommended by the autoscaler for each container.
	// +optional
	ContainerRecommendations []RecommendedContainerResources `json:"containerRecommendations,omitempty" protobuf:"bytes,1,rep,name=containerRecommendations"`
}

// RecommendedContainerResources is the recommendation of resources computed by
// autoscaler for a specific container. Respects the container resource policy
// if present in the spec. In particular the recommendation is not produced for
// containers with `ContainerScalingMode` set to 'Off'.
type RecommendedContainerResources struct {
	// Name of the container.
	ContainerName string `json:"containerName,omitempty" protobuf:"bytes,1,opt,name=containerName"`
	// Recommended amount of resources. Observes ContainerResourcePolicy.
	Target v1.ResourceList `json:"target" protobuf:"bytes,2,rep,name=target,casttype=ResourceList,castkey=ResourceName"`
	// Minimum recommended amount of resources. Observes ContainerResourcePolicy.
	// This amount is not guaranteed to be sufficient for the application to operate in a stable way, however
	// running with less resources is likely to have significant impact on performance/availability.
	// +optional
	LowerBound v1.ResourceList `json:"lowerBound,omitempty" protobuf:"bytes,3,rep,name=lowerBound,casttype=ResourceList,castkey=ResourceName"`
	// Maximum recommended amount of resources. Observes ContainerResourcePolicy.
	// Any resources allocated beyond this value are likely wasted. This value may be larger than the maximum
	// amount of application is actually capable of consuming.
	// +optional
	UpperBound v1.ResourceList `json:"upperBound,omitempty" protobuf:"bytes,4,rep,name=upperBound,casttype=ResourceList,castkey=ResourceName"`
	// The most recent recommended resources target computed by the autoscaler
	// for the controlled pods, based only on actual resource usage, not taking
	// into account the ContainerResourcePolicy.
	// May differ from the Recommendation if the actual resource usage causes
	// the target to violate the ContainerResourcePolicy (lower than MinAllowed
	// or higher that MaxAllowed).
	// Used only as status indication, will not affect actual resource assignment.
	// +optional
	UncappedTarget v1.ResourceList `json:"uncappedTarget,omitempty" protobuf:"bytes,5,opt,name=uncappedTarget"`
}

// VerticalPodAutoscalerConditionType are the valid conditions of
// a VerticalPodAutoscaler.
type VerticalPodAutoscalerConditionType string

var (
	// RecommendationProvided indicates whether the VPA recommender was able to calculate a recommendation.
	RecommendationProvided VerticalPodAutoscalerConditionType = "RecommendationProvided"
	// LowConfidence indicates whether the VPA recommender has low confidence in the recommendation for
	// some of containers.
	LowConfidence VerticalPodAutoscalerConditionType = "LowConfidence"
	// NoPodsMatched indicates that label selector used with VPA object didn't match any pods.
	NoPodsMatched VerticalPodAutoscalerConditionType = "NoPodsMatched"
	// FetchingHistory indicates that VPA recommender is in the process of loading additional history samples.
	FetchingHistory VerticalPodAutoscalerConditionType = "FetchingHistory"
	// ConfigDeprecated indicates that this VPA configuration is deprecated
	// and will stop being supported soon.
	ConfigDeprecated VerticalPodAutoscalerConditionType = "ConfigDeprecated"
	// ConfigUnsupported indicates that this VPA configuration is unsupported
	// and recommendations will not be provided for it.
	ConfigUnsupported VerticalPodAutoscalerConditionType = "ConfigUnsupported"
)

// VerticalPodAutoscalerCondition describes the state of
// a VerticalPodAutoscaler at a certain point.
type VerticalPodAutoscalerCondition struct {
	// type describes the current condition
	Type VerticalPodAutoscalerConditionType `json:"type" protobuf:"bytes,1,name=type"`
	// status is the status of the condition (True, False, Unknown)
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,name=status"`
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

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VerticalPodAutoscalerCheckpoint is the checkpoint of the internal state of VPA that
// is used for recovery after recommender's restart.
type VerticalPodAutoscalerCheckpoint struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the checkpoint.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	// +optional
	Spec VerticalPodAutoscalerCheckpointSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Data of the checkpoint.
	// +optional
	Status VerticalPodAutoscalerCheckpointStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VerticalPodAutoscalerCheckpointList is a list of VerticalPodAutoscalerCheckpoint objects.
type VerticalPodAutoscalerCheckpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []VerticalPodAutoscalerCheckpoint `json:"items"`
}

// VerticalPodAutoscalerCheckpointSpec is the specification of the checkpoint object.
type VerticalPodAutoscalerCheckpointSpec struct {
	// Name of the VPA object that stored VerticalPodAutoscalerCheckpoint object.
	VPAObjectName string `json:"vpaObjectName,omitempty" protobuf:"bytes,1,opt,name=vpaObjectName"`

	// Name of the checkpointed container.
	ContainerName string `json:"containerName,omitempty" protobuf:"bytes,2,opt,name=containerName"`
}

// VerticalPodAutoscalerCheckpointStatus contains data of the checkpoint.
type VerticalPodAutoscalerCheckpointStatus struct {
	// The time when the status was last refreshed.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,1,opt,name=lastUpdateTime"`

	// Version of the format of the stored data.
	Version string `json:"version,omitempty" protobuf:"bytes,2,opt,name=version"`

	// Checkpoint of histogram for consumption of CPU.
	CPUHistogram HistogramCheckpoint `json:"cpuHistogram,omitempty" protobuf:"bytes,3,rep,name=cpuHistograms"`

	// Checkpoint of histogram for consumption of memory.
	MemoryHistogram HistogramCheckpoint `json:"memoryHistogram,omitempty" protobuf:"bytes,4,rep,name=memoryHistogram"`

	// Timestamp of the fist sample from the histograms.
	FirstSampleStart metav1.Time `json:"firstSampleStart,omitempty" protobuf:"bytes,5,opt,name=firstSampleStart"`

	// Timestamp of the last sample from the histograms.
	LastSampleStart metav1.Time `json:"lastSampleStart,omitempty" protobuf:"bytes,6,opt,name=lastSampleStart"`

	// Total number of samples in the histograms.
	TotalSamplesCount int `json:"totalSamplesCount,omitempty" protobuf:"bytes,7,opt,name=totalSamplesCount"`
}

// HistogramCheckpoint contains data needed to reconstruct the histogram.
type HistogramCheckpoint struct {
	// Reference timestamp for samples collected within this histogram.
	ReferenceTimestamp metav1.Time `json:"referenceTimestamp,omitempty" protobuf:"bytes,1,opt,name=referenceTimestamp"`

	// Map from bucket index to bucket weight.
	BucketWeights map[int]uint32 `json:"bucketWeights,omitempty" protobuf:"bytes,2,opt,name=bucketWeights"`

	// Sum of samples to be used as denominator for weights from BucketWeights.
	TotalWeight float64 `json:"totalWeight,omitempty" protobuf:"bytes,3,opt,name=totalWeight"`
}
