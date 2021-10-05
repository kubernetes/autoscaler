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
)

const (
	// GCPServiceAccountJSON is a constant for a key name that is part of the GCP cloud credentials.
	GCPServiceAccountJSON string = "serviceAccountJSON"

	// GCPAlternativeServiceAccountJSON is a constant for a key name of a secret containing the GCP credentials (service
	// account json).
	GCPAlternativeServiceAccountJSON = "serviceaccount.json"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Machine Type",type=string,JSONPath=`.spec.machineType`
// +kubebuilder:printcolumn:name="Region",type=string,JSONPath=`.spec.region`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata"

// GCPMachineClass TODO
type GCPMachineClass struct {
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	metav1.TypeMeta `json:",inline"`

	// +optional
	Spec GCPMachineClassSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// GCPMachineClassList is a collection of GCPMachineClasses.
type GCPMachineClassList struct {
	// +optional
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// +optional
	Items []GCPMachineClass `json:"items"`
}

// GCPMachineClassSpec is the specification of a GCPMachineClass.
type GCPMachineClassSpec struct {
	CanIpForward         bool                    `json:"canIpForward"`
	DeletionProtection   bool                    `json:"deletionProtection"`
	Description          *string                 `json:"description,omitempty"`
	Disks                []*GCPDisk              `json:"disks,omitempty"`
	Labels               map[string]string       `json:"labels,omitempty"`
	MachineType          string                  `json:"machineType"`
	Metadata             []*GCPMetadata          `json:"metadata,omitempty"`
	NetworkInterfaces    []*GCPNetworkInterface  `json:"networkInterfaces,omitempty"`
	Scheduling           GCPScheduling           `json:"scheduling"`
	SecretRef            *corev1.SecretReference `json:"secretRef,omitempty"`
	CredentialsSecretRef *corev1.SecretReference `json:"credentialsSecretRef,omitempty"`
	ServiceAccounts      []GCPServiceAccount     `json:"serviceAccounts"`
	Tags                 []string                `json:"tags,omitempty"`
	Region               string                  `json:"region"`
	Zone                 string                  `json:"zone"`
}

// GCPDisk describes disks for GCP.
type GCPDisk struct {
	AutoDelete *bool             `json:"autoDelete"`
	Boot       bool              `json:"boot"`
	SizeGb     int64             `json:"sizeGb"`
	Type       string            `json:"type"`
	Interface  string            `json:"interface"`
	Image      string            `json:"image"`
	Labels     map[string]string `json:"labels"`
}

// GCPMetadata describes metadata for GCP.
type GCPMetadata struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
}

// GCPNetworkInterface describes network interfaces for GCP
type GCPNetworkInterface struct {
	DisableExternalIP bool   `json:"disableExternalIP,omitempty"`
	Network           string `json:"network,omitempty"`
	Subnetwork        string `json:"subnetwork,omitempty"`
}

// GCPScheduling describes scheduling configuration for GCP.
type GCPScheduling struct {
	AutomaticRestart  bool   `json:"automaticRestart"`
	OnHostMaintenance string `json:"onHostMaintenance"`
	Preemptible       bool   `json:"preemptible"`
}

// GCPServiceAccount describes service accounts for GCP.
type GCPServiceAccount struct {
	Email  string   `json:"email"`
	Scopes []string `json:"scopes"`
}
