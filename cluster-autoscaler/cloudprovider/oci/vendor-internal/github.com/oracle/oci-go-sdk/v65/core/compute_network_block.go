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

// ComputeNetworkBlock A compute network block.
type ComputeNetworkBlock struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute capacity topology.
	ComputeCapacityTopologyId *string `mandatory:"true" json:"computeCapacityTopologyId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute HPC island.
	ComputeHpcIslandId *string `mandatory:"true" json:"computeHpcIslandId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute network block.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the compute network block.
	LifecycleState ComputeNetworkBlockLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time that the compute network block was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The date and time that the compute network block was updated, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeUpdated *common.SDKTime `mandatory:"true" json:"timeUpdated"`

	// The total number of compute bare metal hosts located in this compute network block.
	TotalComputeBareMetalHostCount *int64 `mandatory:"true" json:"totalComputeBareMetalHostCount"`
}

func (m ComputeNetworkBlock) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ComputeNetworkBlock) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingComputeNetworkBlockLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetComputeNetworkBlockLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ComputeNetworkBlockLifecycleStateEnum Enum with underlying type: string
type ComputeNetworkBlockLifecycleStateEnum string

// Set of constants representing the allowable values for ComputeNetworkBlockLifecycleStateEnum
const (
	ComputeNetworkBlockLifecycleStateActive   ComputeNetworkBlockLifecycleStateEnum = "ACTIVE"
	ComputeNetworkBlockLifecycleStateInactive ComputeNetworkBlockLifecycleStateEnum = "INACTIVE"
)

var mappingComputeNetworkBlockLifecycleStateEnum = map[string]ComputeNetworkBlockLifecycleStateEnum{
	"ACTIVE":   ComputeNetworkBlockLifecycleStateActive,
	"INACTIVE": ComputeNetworkBlockLifecycleStateInactive,
}

var mappingComputeNetworkBlockLifecycleStateEnumLowerCase = map[string]ComputeNetworkBlockLifecycleStateEnum{
	"active":   ComputeNetworkBlockLifecycleStateActive,
	"inactive": ComputeNetworkBlockLifecycleStateInactive,
}

// GetComputeNetworkBlockLifecycleStateEnumValues Enumerates the set of values for ComputeNetworkBlockLifecycleStateEnum
func GetComputeNetworkBlockLifecycleStateEnumValues() []ComputeNetworkBlockLifecycleStateEnum {
	values := make([]ComputeNetworkBlockLifecycleStateEnum, 0)
	for _, v := range mappingComputeNetworkBlockLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetComputeNetworkBlockLifecycleStateEnumStringValues Enumerates the set of values in String for ComputeNetworkBlockLifecycleStateEnum
func GetComputeNetworkBlockLifecycleStateEnumStringValues() []string {
	return []string{
		"ACTIVE",
		"INACTIVE",
	}
}

// GetMappingComputeNetworkBlockLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingComputeNetworkBlockLifecycleStateEnum(val string) (ComputeNetworkBlockLifecycleStateEnum, bool) {
	enum, ok := mappingComputeNetworkBlockLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
