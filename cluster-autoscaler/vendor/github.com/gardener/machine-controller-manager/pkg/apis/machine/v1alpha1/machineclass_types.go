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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// MachineClass can be used to templatize and re-use provider configuration
// across multiple Machines / MachineSets / MachineDeployments.
// +k8s:openapi-gen=true
// +resource:path=machineclasses
type MachineClass struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XPreserveUnknownFields

	// CredentialsSecretRef can optionally store the credentials (in this case the SecretRef does not need to store them).
	// This might be useful if multiple machine classes with the same credentials but different user-datas are used.
	CredentialsSecretRef *corev1.SecretReference `json:"credentialsSecretRef,omitempty"`
	// ExpectedNodeAttributes contains subfields the track all scale from zero resources
	// +optional
	ExpectedNodeAttributes *ExpectedNodeAttributes `json:"expectedNodeAttributes,omitempty"`
	// Provider-specific configuration to use during node creation.
	ProviderSpec runtime.RawExtension `json:"providerSpec"`
	// Provider is the combination of name and location of cloud-specific drivers.
	Provider string `json:"provider,omitempty"`
	// SecretRef stores the necessary secrets such as credentials or userdata.
	SecretRef *corev1.SecretReference `json:"secretRef,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// MachineClassList contains a list of MachineClasses
type MachineClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineClass `json:"items"`
}

// ExpectedNodeAttributes contains subfields to track all scale from zero resources
type ExpectedNodeAttributes struct {
	// CPU cores of the expected node
	// +optional
	CPU *int64 `json:"cpu,omitempty"`
	// GPU cores of the expected node
	// +optional
	GPU *int64 `json:"gpu,omitempty"`
	// Instance type of the expected node
	// +optional
	InstanceType *string `json:"instanceType,omitempty"`
	// Memory of the expected node
	// +optional
	Memory *int64 `json:"memory,omitempty"`
	// Region of the expected node
	// +optional
	Region *string `json:"region,omitempty"`
	// Zone of the expected node
	// +optional
	Zone *string `json:"zone,omitempty"`
}
