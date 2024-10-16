// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.cloud.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// InstancePoolInstanceLoadBalancerBackend Represents the load balancer Backend that is configured for an instance pool instance.
type InstancePoolInstanceLoadBalancerBackend struct {

	// The OCID of the load balancer attached to the instance pool.
	LoadBalancerId *string `mandatory:"true" json:"loadBalancerId"`

	// The name of the backend set on the load balancer.
	BackendSetName *string `mandatory:"true" json:"backendSetName"`

	// The name of the backend in the backend set.
	BackendName *string `mandatory:"true" json:"backendName"`

	// The health of the backend as observed by the load balancer.
	BackendHealthStatus InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum `mandatory:"true" json:"backendHealthStatus"`
}

func (m InstancePoolInstanceLoadBalancerBackend) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstancePoolInstanceLoadBalancerBackend) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum(string(m.BackendHealthStatus)); !ok && m.BackendHealthStatus != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BackendHealthStatus: %s. Supported values are: %s.", m.BackendHealthStatus, strings.Join(GetInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum Enum with underlying type: string
type InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum string

// Set of constants representing the allowable values for InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum
const (
	InstancePoolInstanceLoadBalancerBackendBackendHealthStatusOk       InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum = "OK"
	InstancePoolInstanceLoadBalancerBackendBackendHealthStatusWarning  InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum = "WARNING"
	InstancePoolInstanceLoadBalancerBackendBackendHealthStatusCritical InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum = "CRITICAL"
	InstancePoolInstanceLoadBalancerBackendBackendHealthStatusUnknown  InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum = "UNKNOWN"
)

var mappingInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum = map[string]InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum{
	"OK":       InstancePoolInstanceLoadBalancerBackendBackendHealthStatusOk,
	"WARNING":  InstancePoolInstanceLoadBalancerBackendBackendHealthStatusWarning,
	"CRITICAL": InstancePoolInstanceLoadBalancerBackendBackendHealthStatusCritical,
	"UNKNOWN":  InstancePoolInstanceLoadBalancerBackendBackendHealthStatusUnknown,
}

var mappingInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnumLowerCase = map[string]InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum{
	"ok":       InstancePoolInstanceLoadBalancerBackendBackendHealthStatusOk,
	"warning":  InstancePoolInstanceLoadBalancerBackendBackendHealthStatusWarning,
	"critical": InstancePoolInstanceLoadBalancerBackendBackendHealthStatusCritical,
	"unknown":  InstancePoolInstanceLoadBalancerBackendBackendHealthStatusUnknown,
}

// GetInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnumValues Enumerates the set of values for InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum
func GetInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnumValues() []InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum {
	values := make([]InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum, 0)
	for _, v := range mappingInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnumStringValues Enumerates the set of values in String for InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum
func GetInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnumStringValues() []string {
	return []string{
		"OK",
		"WARNING",
		"CRITICAL",
		"UNKNOWN",
	}
}

// GetMappingInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum(val string) (InstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnum, bool) {
	enum, ok := mappingInstancePoolInstanceLoadBalancerBackendBackendHealthStatusEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
