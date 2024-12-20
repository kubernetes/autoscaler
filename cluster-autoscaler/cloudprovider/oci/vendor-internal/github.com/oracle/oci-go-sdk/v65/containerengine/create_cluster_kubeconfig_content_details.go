// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Kubernetes Engine API
//
// API for the Kubernetes Engine service (also known as the Container Engine for Kubernetes service). Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Kubernetes Engine (https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// CreateClusterKubeconfigContentDetails The properties that define a request to create a cluster kubeconfig.
type CreateClusterKubeconfigContentDetails struct {

	// The version of the kubeconfig token. Supported value 2.0.0
	TokenVersion *string `mandatory:"false" json:"tokenVersion"`

	// Deprecated. This field is no longer used.
	Expiration *int `mandatory:"false" json:"expiration"`

	// The endpoint to target. A cluster may have multiple endpoints exposed but the kubeconfig can only target one at a time.
	Endpoint CreateClusterKubeconfigContentDetailsEndpointEnum `mandatory:"false" json:"endpoint,omitempty"`
}

func (m CreateClusterKubeconfigContentDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateClusterKubeconfigContentDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCreateClusterKubeconfigContentDetailsEndpointEnum(string(m.Endpoint)); !ok && m.Endpoint != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Endpoint: %s. Supported values are: %s.", m.Endpoint, strings.Join(GetCreateClusterKubeconfigContentDetailsEndpointEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateClusterKubeconfigContentDetailsEndpointEnum Enum with underlying type: string
type CreateClusterKubeconfigContentDetailsEndpointEnum string

// Set of constants representing the allowable values for CreateClusterKubeconfigContentDetailsEndpointEnum
const (
	CreateClusterKubeconfigContentDetailsEndpointLegacyKubernetes CreateClusterKubeconfigContentDetailsEndpointEnum = "LEGACY_KUBERNETES"
	CreateClusterKubeconfigContentDetailsEndpointPublicEndpoint   CreateClusterKubeconfigContentDetailsEndpointEnum = "PUBLIC_ENDPOINT"
	CreateClusterKubeconfigContentDetailsEndpointPrivateEndpoint  CreateClusterKubeconfigContentDetailsEndpointEnum = "PRIVATE_ENDPOINT"
	CreateClusterKubeconfigContentDetailsEndpointVcnHostname      CreateClusterKubeconfigContentDetailsEndpointEnum = "VCN_HOSTNAME"
)

var mappingCreateClusterKubeconfigContentDetailsEndpointEnum = map[string]CreateClusterKubeconfigContentDetailsEndpointEnum{
	"LEGACY_KUBERNETES": CreateClusterKubeconfigContentDetailsEndpointLegacyKubernetes,
	"PUBLIC_ENDPOINT":   CreateClusterKubeconfigContentDetailsEndpointPublicEndpoint,
	"PRIVATE_ENDPOINT":  CreateClusterKubeconfigContentDetailsEndpointPrivateEndpoint,
	"VCN_HOSTNAME":      CreateClusterKubeconfigContentDetailsEndpointVcnHostname,
}

var mappingCreateClusterKubeconfigContentDetailsEndpointEnumLowerCase = map[string]CreateClusterKubeconfigContentDetailsEndpointEnum{
	"legacy_kubernetes": CreateClusterKubeconfigContentDetailsEndpointLegacyKubernetes,
	"public_endpoint":   CreateClusterKubeconfigContentDetailsEndpointPublicEndpoint,
	"private_endpoint":  CreateClusterKubeconfigContentDetailsEndpointPrivateEndpoint,
	"vcn_hostname":      CreateClusterKubeconfigContentDetailsEndpointVcnHostname,
}

// GetCreateClusterKubeconfigContentDetailsEndpointEnumValues Enumerates the set of values for CreateClusterKubeconfigContentDetailsEndpointEnum
func GetCreateClusterKubeconfigContentDetailsEndpointEnumValues() []CreateClusterKubeconfigContentDetailsEndpointEnum {
	values := make([]CreateClusterKubeconfigContentDetailsEndpointEnum, 0)
	for _, v := range mappingCreateClusterKubeconfigContentDetailsEndpointEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateClusterKubeconfigContentDetailsEndpointEnumStringValues Enumerates the set of values in String for CreateClusterKubeconfigContentDetailsEndpointEnum
func GetCreateClusterKubeconfigContentDetailsEndpointEnumStringValues() []string {
	return []string{
		"LEGACY_KUBERNETES",
		"PUBLIC_ENDPOINT",
		"PRIVATE_ENDPOINT",
		"VCN_HOSTNAME",
	}
}

// GetMappingCreateClusterKubeconfigContentDetailsEndpointEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateClusterKubeconfigContentDetailsEndpointEnum(val string) (CreateClusterKubeconfigContentDetailsEndpointEnum, bool) {
	enum, ok := mappingCreateClusterKubeconfigContentDetailsEndpointEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
