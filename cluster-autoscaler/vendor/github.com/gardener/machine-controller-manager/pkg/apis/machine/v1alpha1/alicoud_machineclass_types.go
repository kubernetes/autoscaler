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
	// AlicloudAccessKeyID is a constant for a key name that is part of the Alibaba cloud credentials.
	AlicloudAccessKeyID string = "alicloudAccessKeyID"
	// AlicloudAccessKeySecret is a constant for a key name that is part of the Alibaba cloud credentials.
	AlicloudAccessKeySecret string = "alicloudAccessKeySecret"

	// AlicloudAlternativeAccessKeyID is a constant for a key name of a secret containing the Alibaba cloud
	// credentials (access key id).
	AlicloudAlternativeAccessKeyID = "accessKeyID"
	// AlicloudAlternativeAccessKeySecret is a constant for a key name of a secret containing the Alibaba cloud
	// credentials (access key secret).
	AlicloudAlternativeAccessKeySecret = "accessKeySecret"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Instance Type",type=string,JSONPath=`.spec.instanceType`
// +kubebuilder:printcolumn:name="Region",type=string,JSONPath=`.spec.region`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata"

// AlicloudMachineClass TODO
type AlicloudMachineClass struct {
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	metav1.TypeMeta `json:",inline"`

	// +optional
	Spec AlicloudMachineClassSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// AlicloudMachineClassList is a collection of AlicloudMachineClasses.
type AlicloudMachineClassList struct {
	// +optional
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// +optional
	Items []AlicloudMachineClass `json:"items"`
}

// AlicloudMachineClassSpec is the specification of a AlicloudMachineClass.
type AlicloudMachineClassSpec struct {
	ImageID                 string                  `json:"imageID"`
	InstanceType            string                  `json:"instanceType"`
	Region                  string                  `json:"region"`
	ZoneID                  string                  `json:"zoneID,omitempty"`
	SecurityGroupID         string                  `json:"securityGroupID,omitempty"`
	VSwitchID               string                  `json:"vSwitchID"`
	PrivateIPAddress        string                  `json:"privateIPAddress,omitempty"`
	SystemDisk              *AlicloudSystemDisk     `json:"systemDisk,omitempty"`
	DataDisks               []AlicloudDataDisk      `json:"dataDisks,omitempty"`
	InstanceChargeType      string                  `json:"instanceChargeType,omitempty"`
	InternetChargeType      string                  `json:"internetChargeType,omitempty"`
	InternetMaxBandwidthIn  *int                    `json:"internetMaxBandwidthIn,omitempty"`
	InternetMaxBandwidthOut *int                    `json:"internetMaxBandwidthOut,omitempty"`
	SpotStrategy            string                  `json:"spotStrategy,omitempty"`
	IoOptimized             string                  `json:"IoOptimized,omitempty"`
	Tags                    map[string]string       `json:"tags,omitempty"`
	KeyPairName             string                  `json:"keyPairName"`
	SecretRef               *corev1.SecretReference `json:"secretRef,omitempty"`
	CredentialsSecretRef    *corev1.SecretReference `json:"credentialsSecretRef,omitempty"`
}

type AlicloudDataDisk struct {
	Name     string `json:"name,omitempty"`
	Category string `json:"category,omitempty"`
	// +optional
	Description        string `json:"description,omitempty"`
	Encrypted          bool   `json:"encrypted,omitempty"`
	DeleteWithInstance *bool  `json:"deleteWithInstance,omitempty"`
	Size               int    `json:"size,omitempty"`
}

// AlicloudSystemDisk describes SystemDisk for Alicloud.
type AlicloudSystemDisk struct {
	Category string `json:"category"`
	Size     int    `json:"size"`
}
