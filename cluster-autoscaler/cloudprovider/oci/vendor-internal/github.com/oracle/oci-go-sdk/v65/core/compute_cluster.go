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

// ComputeCluster A remote direct memory access (RDMA) network group.
// A cluster network on a compute cluster (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/compute-clusters.htm) is a group of
// high performance computing (HPC), GPU, or optimized instances that are connected with an ultra low-latency network.
// Use compute clusters when you want to manage instances in the cluster individually in the RDMA network group.
// For details about cluster networks that use instance pools to manage groups of identical instances,
// see ClusterNetwork.
type ComputeCluster struct {

	// The availability domain the compute cluster is running in.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment that contains the compute cluster.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute cluster.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the compute cluster.
	LifecycleState ComputeClusterLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the compute cluster was created,
	// in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m ComputeCluster) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ComputeCluster) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingComputeClusterLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetComputeClusterLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ComputeClusterLifecycleStateEnum Enum with underlying type: string
type ComputeClusterLifecycleStateEnum string

// Set of constants representing the allowable values for ComputeClusterLifecycleStateEnum
const (
	ComputeClusterLifecycleStateActive  ComputeClusterLifecycleStateEnum = "ACTIVE"
	ComputeClusterLifecycleStateDeleted ComputeClusterLifecycleStateEnum = "DELETED"
)

var mappingComputeClusterLifecycleStateEnum = map[string]ComputeClusterLifecycleStateEnum{
	"ACTIVE":  ComputeClusterLifecycleStateActive,
	"DELETED": ComputeClusterLifecycleStateDeleted,
}

var mappingComputeClusterLifecycleStateEnumLowerCase = map[string]ComputeClusterLifecycleStateEnum{
	"active":  ComputeClusterLifecycleStateActive,
	"deleted": ComputeClusterLifecycleStateDeleted,
}

// GetComputeClusterLifecycleStateEnumValues Enumerates the set of values for ComputeClusterLifecycleStateEnum
func GetComputeClusterLifecycleStateEnumValues() []ComputeClusterLifecycleStateEnum {
	values := make([]ComputeClusterLifecycleStateEnum, 0)
	for _, v := range mappingComputeClusterLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetComputeClusterLifecycleStateEnumStringValues Enumerates the set of values in String for ComputeClusterLifecycleStateEnum
func GetComputeClusterLifecycleStateEnumStringValues() []string {
	return []string{
		"ACTIVE",
		"DELETED",
	}
}

// GetMappingComputeClusterLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingComputeClusterLifecycleStateEnum(val string) (ComputeClusterLifecycleStateEnum, bool) {
	enum, ok := mappingComputeClusterLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
