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

// ComputeBareMetalHost A compute bare metal host.
type ComputeBareMetalHost struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute capacity topology.
	ComputeCapacityTopologyId *string `mandatory:"true" json:"computeCapacityTopologyId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute bare metal host.
	Id *string `mandatory:"true" json:"id"`

	// The shape of the compute instance that runs on the compute bare metal host.
	InstanceShape *string `mandatory:"true" json:"instanceShape"`

	// The current state of the compute bare metal host.
	LifecycleState ComputeBareMetalHostLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time that the compute bare metal host was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The date and time that the compute bare metal host was updated, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeUpdated *common.SDKTime `mandatory:"true" json:"timeUpdated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute HPC island.
	ComputeHpcIslandId *string `mandatory:"false" json:"computeHpcIslandId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute local block.
	ComputeLocalBlockId *string `mandatory:"false" json:"computeLocalBlockId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute network block.
	ComputeNetworkBlockId *string `mandatory:"false" json:"computeNetworkBlockId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute instance that runs on the compute bare metal host.
	InstanceId *string `mandatory:"false" json:"instanceId"`

	// The lifecycle state details of the compute bare metal host.
	LifecycleDetails ComputeBareMetalHostLifecycleDetailsEnum `mandatory:"false" json:"lifecycleDetails,omitempty"`
}

func (m ComputeBareMetalHost) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ComputeBareMetalHost) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingComputeBareMetalHostLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetComputeBareMetalHostLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := GetMappingComputeBareMetalHostLifecycleDetailsEnum(string(m.LifecycleDetails)); !ok && m.LifecycleDetails != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleDetails: %s. Supported values are: %s.", m.LifecycleDetails, strings.Join(GetComputeBareMetalHostLifecycleDetailsEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ComputeBareMetalHostLifecycleDetailsEnum Enum with underlying type: string
type ComputeBareMetalHostLifecycleDetailsEnum string

// Set of constants representing the allowable values for ComputeBareMetalHostLifecycleDetailsEnum
const (
	ComputeBareMetalHostLifecycleDetailsAvailable   ComputeBareMetalHostLifecycleDetailsEnum = "AVAILABLE"
	ComputeBareMetalHostLifecycleDetailsDegraded    ComputeBareMetalHostLifecycleDetailsEnum = "DEGRADED"
	ComputeBareMetalHostLifecycleDetailsUnavailable ComputeBareMetalHostLifecycleDetailsEnum = "UNAVAILABLE"
)

var mappingComputeBareMetalHostLifecycleDetailsEnum = map[string]ComputeBareMetalHostLifecycleDetailsEnum{
	"AVAILABLE":   ComputeBareMetalHostLifecycleDetailsAvailable,
	"DEGRADED":    ComputeBareMetalHostLifecycleDetailsDegraded,
	"UNAVAILABLE": ComputeBareMetalHostLifecycleDetailsUnavailable,
}

var mappingComputeBareMetalHostLifecycleDetailsEnumLowerCase = map[string]ComputeBareMetalHostLifecycleDetailsEnum{
	"available":   ComputeBareMetalHostLifecycleDetailsAvailable,
	"degraded":    ComputeBareMetalHostLifecycleDetailsDegraded,
	"unavailable": ComputeBareMetalHostLifecycleDetailsUnavailable,
}

// GetComputeBareMetalHostLifecycleDetailsEnumValues Enumerates the set of values for ComputeBareMetalHostLifecycleDetailsEnum
func GetComputeBareMetalHostLifecycleDetailsEnumValues() []ComputeBareMetalHostLifecycleDetailsEnum {
	values := make([]ComputeBareMetalHostLifecycleDetailsEnum, 0)
	for _, v := range mappingComputeBareMetalHostLifecycleDetailsEnum {
		values = append(values, v)
	}
	return values
}

// GetComputeBareMetalHostLifecycleDetailsEnumStringValues Enumerates the set of values in String for ComputeBareMetalHostLifecycleDetailsEnum
func GetComputeBareMetalHostLifecycleDetailsEnumStringValues() []string {
	return []string{
		"AVAILABLE",
		"DEGRADED",
		"UNAVAILABLE",
	}
}

// GetMappingComputeBareMetalHostLifecycleDetailsEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingComputeBareMetalHostLifecycleDetailsEnum(val string) (ComputeBareMetalHostLifecycleDetailsEnum, bool) {
	enum, ok := mappingComputeBareMetalHostLifecycleDetailsEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ComputeBareMetalHostLifecycleStateEnum Enum with underlying type: string
type ComputeBareMetalHostLifecycleStateEnum string

// Set of constants representing the allowable values for ComputeBareMetalHostLifecycleStateEnum
const (
	ComputeBareMetalHostLifecycleStateActive   ComputeBareMetalHostLifecycleStateEnum = "ACTIVE"
	ComputeBareMetalHostLifecycleStateInactive ComputeBareMetalHostLifecycleStateEnum = "INACTIVE"
)

var mappingComputeBareMetalHostLifecycleStateEnum = map[string]ComputeBareMetalHostLifecycleStateEnum{
	"ACTIVE":   ComputeBareMetalHostLifecycleStateActive,
	"INACTIVE": ComputeBareMetalHostLifecycleStateInactive,
}

var mappingComputeBareMetalHostLifecycleStateEnumLowerCase = map[string]ComputeBareMetalHostLifecycleStateEnum{
	"active":   ComputeBareMetalHostLifecycleStateActive,
	"inactive": ComputeBareMetalHostLifecycleStateInactive,
}

// GetComputeBareMetalHostLifecycleStateEnumValues Enumerates the set of values for ComputeBareMetalHostLifecycleStateEnum
func GetComputeBareMetalHostLifecycleStateEnumValues() []ComputeBareMetalHostLifecycleStateEnum {
	values := make([]ComputeBareMetalHostLifecycleStateEnum, 0)
	for _, v := range mappingComputeBareMetalHostLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetComputeBareMetalHostLifecycleStateEnumStringValues Enumerates the set of values in String for ComputeBareMetalHostLifecycleStateEnum
func GetComputeBareMetalHostLifecycleStateEnumStringValues() []string {
	return []string{
		"ACTIVE",
		"INACTIVE",
	}
}

// GetMappingComputeBareMetalHostLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingComputeBareMetalHostLifecycleStateEnum(val string) (ComputeBareMetalHostLifecycleStateEnum, bool) {
	enum, ok := mappingComputeBareMetalHostLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
