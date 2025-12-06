// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Kubernetes Engine API
//
// API for the Kubernetes Engine service (also known as the Container Engine for Kubernetes service). Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Kubernetes Engine (https://docs.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// ClusterCreateOptions The properties that define extra options for a cluster.
type ClusterCreateOptions struct {

	// The OCIDs of the subnets used for Kubernetes services load balancers.
	ServiceLbSubnetIds []string `mandatory:"false" json:"serviceLbSubnetIds"`

	// IP family to use for single stack or define the order of IP families for dual-stack
	IpFamilies []ClusterCreateOptionsIpFamiliesEnum `mandatory:"false" json:"ipFamilies,omitempty"`

	// Network configuration for Kubernetes.
	KubernetesNetworkConfig *KubernetesNetworkConfig `mandatory:"false" json:"kubernetesNetworkConfig"`

	// Configurable cluster add-ons
	AddOns *AddOnOptions `mandatory:"false" json:"addOns"`

	// Configurable cluster admission controllers
	AdmissionControllerOptions *AdmissionControllerOptions `mandatory:"false" json:"admissionControllerOptions"`

	PersistentVolumeConfig *PersistentVolumeConfigDetails `mandatory:"false" json:"persistentVolumeConfig"`

	ServiceLbConfig *ServiceLbConfigDetails `mandatory:"false" json:"serviceLbConfig"`

	OpenIdConnectTokenAuthenticationConfig *OpenIdConnectTokenAuthenticationConfig `mandatory:"false" json:"openIdConnectTokenAuthenticationConfig"`

	OpenIdConnectDiscovery *OpenIdConnectDiscovery `mandatory:"false" json:"openIdConnectDiscovery"`
}

func (m ClusterCreateOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ClusterCreateOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	for _, val := range m.IpFamilies {
		if _, ok := GetMappingClusterCreateOptionsIpFamiliesEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for IpFamilies: %s. Supported values are: %s.", val, strings.Join(GetClusterCreateOptionsIpFamiliesEnumStringValues(), ",")))
		}
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ClusterCreateOptionsIpFamiliesEnum Enum with underlying type: string
type ClusterCreateOptionsIpFamiliesEnum string

// Set of constants representing the allowable values for ClusterCreateOptionsIpFamiliesEnum
const (
	ClusterCreateOptionsIpFamiliesIpv4 ClusterCreateOptionsIpFamiliesEnum = "IPv4"
	ClusterCreateOptionsIpFamiliesIpv6 ClusterCreateOptionsIpFamiliesEnum = "IPv6"
)

var mappingClusterCreateOptionsIpFamiliesEnum = map[string]ClusterCreateOptionsIpFamiliesEnum{
	"IPv4": ClusterCreateOptionsIpFamiliesIpv4,
	"IPv6": ClusterCreateOptionsIpFamiliesIpv6,
}

var mappingClusterCreateOptionsIpFamiliesEnumLowerCase = map[string]ClusterCreateOptionsIpFamiliesEnum{
	"ipv4": ClusterCreateOptionsIpFamiliesIpv4,
	"ipv6": ClusterCreateOptionsIpFamiliesIpv6,
}

// GetClusterCreateOptionsIpFamiliesEnumValues Enumerates the set of values for ClusterCreateOptionsIpFamiliesEnum
func GetClusterCreateOptionsIpFamiliesEnumValues() []ClusterCreateOptionsIpFamiliesEnum {
	values := make([]ClusterCreateOptionsIpFamiliesEnum, 0)
	for _, v := range mappingClusterCreateOptionsIpFamiliesEnum {
		values = append(values, v)
	}
	return values
}

// GetClusterCreateOptionsIpFamiliesEnumStringValues Enumerates the set of values in String for ClusterCreateOptionsIpFamiliesEnum
func GetClusterCreateOptionsIpFamiliesEnumStringValues() []string {
	return []string{
		"IPv4",
		"IPv6",
	}
}

// GetMappingClusterCreateOptionsIpFamiliesEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingClusterCreateOptionsIpFamiliesEnum(val string) (ClusterCreateOptionsIpFamiliesEnum, bool) {
	enum, ok := mappingClusterCreateOptionsIpFamiliesEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
