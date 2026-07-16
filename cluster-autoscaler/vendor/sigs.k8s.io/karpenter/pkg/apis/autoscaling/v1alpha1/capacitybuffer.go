/*
Copyright The Kubernetes Authors.

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

// Upstream source: k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1
// Used this commit - d75b23c039d8596e70565cb8c5186e15a1f21301
// These types are copied from the upstream CapacityBuffer API to avoid pulling in the entire
// autoscaler module and its transitive dependencies. Keep in sync with upstream as needed.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CapacityBuffer is the Schema for the CapacityBuffer API.
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=capacitybuffers,scope=Namespaced,shortName=cb,categories=autoscaling
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Strategy",type="string",JSONPath=".spec.provisioningStrategy",description="The strategy to be used."
// +kubebuilder:printcolumn:name="PodTemplate",type="string",JSONPath=".status.podTemplateRef.name",description="The name of the PodTemplate used."
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".status.replicas",description="The actual number of buffer chunks."
// +kubebuilder:printcolumn:name="ConditionsType",type="string",JSONPath=".status.conditions[*].type",description="List of all condition types."
// +kubebuilder:printcolumn:name="ConditionsStatus",type="string",JSONPath=".status.conditions[*].status",description="List of all condition statuses."
// +kubebuilder:printcolumn:name="ConditionsReason",type="string",JSONPath=".status.conditions[*].reason",description="List of all condition reasons."
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="The age of the CapacityBuffer."

// CapacityBuffer is the configuration that an autoscaler can use to provision buffer capacity within a cluster.
// This buffer is represented by placeholder pods that trigger the Cluster Autoscaler to scale up nodes in advance,
// ensuring that there is always spare capacity available to handle sudden workload spikes or to speed up scaling events.
type CapacityBuffer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"` //nolint:kubeapilinter

	//nolint:kubeapilinter
	// spec defines the desired characteristics of the buffer.
	// +required
	Spec CapacityBufferSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`

	// status represents the current state of the buffer and its readiness for autoprovisioning.
	// +optional
	Status CapacityBufferStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"` //nolint:kubeapilinter
}

type Limits v1.ResourceList

// LocalObjectRef is a reference to an object in the same namespace.
type LocalObjectRef struct {
	//nolint:kubeapilinter
	// name of the referent.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
}

// ScalableRef is a reference to an object with a scale subresource.
type ScalableRef struct {
	//nolint:kubeapilinter
	// apiGroup is the API group of the referent.
	// Empty string for the core API group.
	// +optional
	APIGroup string `json:"apiGroup,omitempty" protobuf:"bytes,1,opt,name=apiGroup"`

	//nolint:kubeapilinter
	// kind is the kind of the referent.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Kind string `json:"kind" protobuf:"bytes,2,opt,name=kind"`

	//nolint:kubeapilinter
	// name is the name of the referent.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
}

// CapacityBufferSpec defines the desired state of CapacityBuffer.
// +kubebuilder:validation:XValidation:rule="!has(self.podTemplateRef) || has(self.replicas) || has(self.limits)",message="If podTemplateRef is set, replicas or limits must also be set"
// +kubebuilder:validation:XValidation:rule="!(has(self.podTemplateRef) && has(self.scalableRef))",message="You must define either PodTemplateRef or ScalableRef, but not both"
type CapacityBufferSpec struct {
	// provisioningStrategy defines how the buffer is utilized.
	// "buffer.x-k8s.io/active-capacity" is the default strategy, where the buffer actively scales up the cluster by creating placeholder pods.
	// +default="buffer.x-k8s.io/active-capacity"
	// +optional
	ProvisioningStrategy *string `json:"provisioningStrategy,omitempty" protobuf:"bytes,1,opt,name=provisioningStrategy"`

	// podTemplateRef is a reference to a PodTemplate resource in the same namespace
	// that declares the shape of a single chunk of the buffer. The pods created
	// from this template will be used as placeholder pods for the buffer capacity.
	// Exactly one of `podTemplateRef`, `scalableRef` should be specified.
	// +optional
	PodTemplateRef *LocalObjectRef `json:"podTemplateRef,omitempty" protobuf:"bytes,2,opt,name=podTemplateRef"`

	// scalableRef is a reference to an object of a kind that has a scale subresource
	// and specifies its label selector field. This allows the CapacityBuffer to
	// manage the buffer by scaling an existing scalable resource.
	// Exactly one of `podTemplateRef`, `scalableRef` should be specified.
	// +optional
	ScalableRef *ScalableRef `json:"scalableRef,omitempty" protobuf:"bytes,3,opt,name=scalableRef"`

	// replicas defines the desired number of buffer chunks to provision.
	// If neither `replicas` nor `percentage` is set, as many chunks as fit within
	// defined resource limits (if any) will be created. If both are set, the minimum
	// of the two will be used.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMinimum=false
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,4,opt,name=replicas"`

	// percentage defines the desired buffer capacity as a percentage of the
	// `scalableRef`'s current replicas. This is only applicable if `scalableRef` is set.
	// The absolute number of replicas is calculated from the percentage by rounding up to a minimum of 1.
	// For example, if `scalableRef` has 10 replicas and `percentage` is 20, 2 buffer chunks will be created.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:ExclusiveMinimum=false
	Percentage *int32 `json:"percentage,omitempty" protobuf:"varint,5,opt,name=percentage"`

	// limits specifies resource constraints that limit the number of chunks created for this buffer
	// based on total resource requests (e.g., CPU, memory). If there are no other
	// limitations for the number of chunks (i.e., `replicas` or `percentage` are not set),
	// this will be used to create as many chunks as fit into these limits.
	// +optional
	Limits Limits `json:"limits,omitempty" protobuf:"bytes,6,opt,name=limits"` //nolint:kubeapilinter
}

// CapacityBufferStatus defines the observed state of CapacityBuffer.
type CapacityBufferStatus struct {
	// podTemplateRef is the observed reference to the PodTemplate that was used
	// to provision the buffer. If this field is not set, and the `conditions`
	// indicate an error, it provides details about the error state.
	// +optional
	PodTemplateRef *LocalObjectRef `json:"podTemplateRef,omitempty" protobuf:"bytes,1,opt,name=podTemplateRef"`

	// replicas is the actual number of buffer chunks currently provisioned.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,2,opt,name=replicas"`

	// podTemplateGeneration is the observed generation of the PodTemplate, used
	// to determine if the status is up-to-date with the desired `spec.podTemplateRef`.
	// +optional
	PodTemplateGeneration *int64 `json:"podTemplateGeneration,omitempty" protobuf:"varint,3,opt,name=podTemplateGeneration"`

	// conditions provide a standard mechanism for reporting the buffer's state.
	// The "Ready" condition indicates if the buffer is successfully provisioned
	// and active. Other conditions may report on various aspects of the buffer's
	// health and provisioning process.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,4,rep,name=conditions"` //nolint:kubeapilinter

	// provisioningStrategy defines how the buffer should be utilized.
	// +optional
	ProvisioningStrategy *string `json:"provisioningStrategy,omitempty" protobuf:"bytes,5,opt,name=provisioningStrategy"`
}

// SetCondition sets or updates a status condition on the CapacityBuffer.
func (cb *CapacityBuffer) SetCondition(condType string, condStatus metav1.ConditionStatus, reason, message string) {
	apimeta.SetStatusCondition(&cb.Status.Conditions, metav1.Condition{
		Type:               condType,
		Status:             condStatus,
		ObservedGeneration: cb.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})
}

// CapacityBufferList contains a list of CapacityBuffer resources.
// +kubebuilder:object:root=true
type CapacityBufferList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CapacityBuffer `json:"items"`
}
