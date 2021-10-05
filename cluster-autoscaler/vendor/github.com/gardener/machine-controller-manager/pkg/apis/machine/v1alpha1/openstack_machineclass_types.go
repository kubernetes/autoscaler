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
	// OpenStackAuthURL is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackAuthURL string = "authURL"
	// OpenStackCACert is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackCACert string = "caCert"
	// OpenStackInsecure is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackInsecure string = "insecure"
	// OpenStackDomainName is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackDomainName string = "domainName"
	// OpenStackDomainID is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackDomainID string = "domainID"
	// OpenStackTenantName is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackTenantName string = "tenantName"
	// OpenStackTenantID is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackTenantID string = "tenantID"
	// OpenStackUserDomainName is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackUserDomainName string = "userDomainName"
	// OpenStackUserDomainID is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackUserDomainID string = "userDomainID"
	// OpenStackUsername is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackUsername string = "username"
	// OpenStackPassword is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackPassword string = "password"
	// OpenStackClientCert is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackClientCert string = "clientCert"
	// OpenStackClientKey is a constant for a key name that is part of the OpenStack cloud credentials.
	OpenStackClientKey string = "clientKey"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Flavor",type=string,JSONPath=`.spec.flavorName`
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.imageName`
// +kubebuilder:printcolumn:name="Region",type=string,JSONPath=`.spec.region`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata"
// OpenStackMachineClass TODO
type OpenStackMachineClass struct {
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	metav1.TypeMeta `json:",inline"`

	// +optional
	Spec OpenStackMachineClassSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// OpenStackMachineClassList is a collection of OpenStackMachineClasses.
type OpenStackMachineClassList struct {
	// +optional
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// +optional
	Items []OpenStackMachineClass `json:"items"`
}

// OpenStackMachineClassSpec is the specification of a OpenStackMachineClass.
type OpenStackMachineClassSpec struct {
	ImageID              string                  `json:"imageID"`
	ImageName            string                  `json:"imageName"`
	Region               string                  `json:"region"`
	AvailabilityZone     string                  `json:"availabilityZone"`
	FlavorName           string                  `json:"flavorName"`
	KeyName              string                  `json:"keyName"`
	SecurityGroups       []string                `json:"securityGroups"`
	Tags                 map[string]string       `json:"tags,omitempty"`
	NetworkID            string                  `json:"networkID"`
	Networks             []OpenStackNetwork      `json:"networks,omitempty"`
	SubnetID             *string                 `json:"subnetID,omitempty"`
	SecretRef            *corev1.SecretReference `json:"secretRef,omitempty"`
	CredentialsSecretRef *corev1.SecretReference `json:"credentialsSecretRef,omitempty"`
	PodNetworkCidr       string                  `json:"podNetworkCidr"`
	RootDiskSize         int                     `json:"rootDiskSize,omitempty"` // in GB
	UseConfigDrive       *bool                   `json:"useConfigDrive,omitempty"`
	ServerGroupID        *string                 `json:"serverGroupID,omitempty"`
}

type OpenStackNetwork struct {
	Id         string `json:"id,omitempty"` // takes priority before name
	Name       string `json:"name,omitempty"`
	PodNetwork bool   `json:"podNetwork,omitempty"`
}
