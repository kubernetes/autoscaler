// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the machine.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
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
