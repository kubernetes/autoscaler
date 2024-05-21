// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// WARNING!
// IF YOU MODIFY ANY OF THE TYPES HERE COPY THEM TO ../types.go
// AND RUN `make generate`

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachineTemplateSpec describes the data a machine should have when created from a template
type MachineTemplateSpec struct {
	// +kubebuilder:validation:XPreserveUnknownFields
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the machine.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec MachineSpec `json:"spec,omitempty"`
}

// MachineConfiguration describes the configurations useful for the machine-controller.
type MachineConfiguration struct {
	// MachineDraintimeout is the timeout after which machine is forcefully deleted.
	// +optional
	MachineDrainTimeout *metav1.Duration `json:"drainTimeout,omitempty"`

	// MachineHealthTimeout is the timeout after which machine is declared unhealhty/failed.
	// +optional
	MachineHealthTimeout *metav1.Duration `json:"healthTimeout,omitempty"`

	// MachineCreationTimeout is the timeout after which machinie creation is declared failed.
	// +optional
	MachineCreationTimeout *metav1.Duration `json:"creationTimeout,omitempty"`

	// MaxEvictRetries is the number of retries that will be attempted while draining the node.
	// +optional
	MaxEvictRetries *int32 `json:"maxEvictRetries,omitempty"`

	// NodeConditions are the set of conditions if set to true for MachineHealthTimeOut, machine will be declared failed.
	// +optional
	NodeConditions *string `json:"nodeConditions,omitempty"`
}

// MachineSummary store the summary of machine.
type MachineSummary struct {
	// Name of the machine object
	Name string `json:"name,omitempty"`

	// ProviderID represents the provider's unique ID given to a machine
	ProviderID string `json:"providerID,omitempty"`

	// Last operation refers to the status of the last operation performed
	LastOperation LastOperation `json:"lastOperation,omitempty"`

	// OwnerRef
	OwnerRef string `json:"ownerRef,omitempty"`
}
