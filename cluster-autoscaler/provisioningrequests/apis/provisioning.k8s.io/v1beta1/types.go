/*
Copyright 2023 The Kubernetes Authors.

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

// Package v1beta1 contains definitions of Provisioning Request related objects.
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// Dependencies for the generation of the code:
	_ "github.com/onsi/ginkgo/v2"
	_ "github.com/onsi/gomega"
	_ "k8s.io/code-generator"
)

// +genclient
// +kubebuilder:storageversions
// +kubebuilder:resource:shortName=provreq;provreqs

// ProvisioningRequest is a way to express additional capacity
// that we would like to provision in the cluster. Cluster Autoscaler
// can use this information in its calculations and signal if the capacity
// is available in the cluster or actively add capacity if needed.
//
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:metadata:annotations="api-approved.kubernetes.io=https://github.com/kubernetes/autoscaler/pull/5848"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ProvisioningRequest struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	//
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec contains specification of the ProvisioningRequest object.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	//
	// +kubebuilder:validation:Required
	Spec ProvisioningRequestSpec `json:"spec"`
	// Status of the ProvisioningRequest. CA constantly reconciles this field.
	//
	// +optional
	Status ProvisioningRequestStatus `json:"status,omitempty"`
}

// ProvisioningRequestList is a object for list of ProvisioningRequest.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ProvisioningRequestList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	//
	// +optional
	metav1.ListMeta `json:"metadata"`
	// Items, list of ProvisioningRequest returned from API.
	//
	// +optional
	Items []ProvisioningRequest `json:"items"`
}

// ProvisioningRequestSpec is a specification of additional pods for which we
// would like to provision additional resources in the cluster.
type ProvisioningRequestSpec struct {
	// PodSets lists groups of pods for which we would like to provision
	// resources.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=32
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	PodSets []PodSet `json:"podSets"`

	// ProvisioningClass describes the different modes of provisioning the resources.
	// Supported values:
	// * check-capacity.kubernetes.io - check if current cluster state can fullfil this request,
	//   do not reserve the capacity.
	// * atomic-scale-up.kubernetes.io - provision the resources in an atomic manner
	// * ... - potential other classes that are specific to the cloud providers
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ProvisioningClass string `json:"provisioningClass"`

	// AdditionalParameters contains all other parameters custom classes may require.
	//
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	AdditionalParameters map[string]string `json:"additionalParameters"`
}

// PodSet represents one group of pods for Provisioning Request to provision capacity.
type PodSet struct {
	// PodTemplateRef is a reference to a PodTemplate object that is representing pods
	// that will consume this reservation (must be within the same namespace).
	// Users need to make sure that the  fields relevant to scheduler (e.g. node selector tolerations)
	// are consistent between this template and actual pods consuming the Provisioning Request.
	//
	// +kubebuilder:validation:Required
	PodTemplateRef Reference `json:"podTemplateRef"`
	// Count contains the number of pods that will be created with a given
	// template.
	//
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=16384
	Count int32 `json:"count"`
}

// Reference represents reference to an object within the same namespace.
type Reference struct {
	// Name of the referenced object.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names
	//
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
}

// ProvisioningRequestStatus represents the status of the resource reservation.
type ProvisioningRequestStatus struct {
	// Conditions represent the observations of a Provisioning Request's
	// current state. Those will contain information whether the capacity
	// was found/created or if there were any issues. The condition types
	// may differ between different provisioning classes.
	//
	// +listType=map
	// +listMapKey=type
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions"`

	// AdditionalStatus contains all other status values custom provisioning classes may require.
	//
	// +optional
	// +kubebuilder:validation:MinProperties=64
	AdditionalStatus map[string]string `json:"additionalStatus"`
}

// The following constants list all currently available Conditions Type values.
// See: https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Condition
const (
	// CapacityAvailable indicates that all of the requested resources were
	// already available in the cluster.
	CapacityAvailable string = "CapacityAvailable"
	// Provisioned indicates that all of the requested resources were created
	// and are available in the cluster. CA will set this condition when the
	// VM creation finishes successfully.
	Provisioned string = "Provisioned"
	// Failed indicates that it is impossible to obtain resources to fulfill
	// this ProvisioningRequest.
	// Condition Reason and Message will contain more details about what failed.
	Failed string = "Failed"
)

const (
	// ProvisioningClassCheckCapacity denotes that CA will check if free capacity
	// is available in the cluster.
	ProvisioningClassCheckCapacity string = "check-capacity.kubernetes.io"
	// ProvisioningClassAtomicScaleUp denotes that CA try to provision the capacity
	// in an atomic manner.
	ProvisioningClassAtomicScaleUp string = "atomic-scale-up.kubernetes.io"
)
