// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// ComputeGpuMemoryCluster The customer facing object includes GPU memory cluster details.
type ComputeGpuMemoryCluster struct {

	// The availability domain of the GPU memory cluster.
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for the Customer-unique GPU memory cluster
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the Instance Configuration used to source launch details for this instance.
	InstanceConfigurationId *string `mandatory:"true" json:"instanceConfigurationId"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment that contains the compute GPU
	// memory cluster.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The lifecycle state of the GPU memory cluster
	LifecycleState ComputeGpuMemoryClusterLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute cluster.
	ComputeClusterId *string `mandatory:"true" json:"computeClusterId"`

	// The number of instances currently running in the GpuMemoryCluster
	Size *int64 `mandatory:"true" json:"size"`

	// The date and time the GPU memory cluster was created.
	// Example: `2016-09-15T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the GPU memory fabric.
	GpuMemoryFabricId *string `mandatory:"false" json:"gpuMemoryFabricId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Usage of system tag keys. These predefined keys are scoped to namespaces.
	// Example: `{ "orcl-cloud": { "free-tier-retained": "true" } }`
	SystemTags map[string]map[string]interface{} `mandatory:"false" json:"systemTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m ComputeGpuMemoryCluster) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ComputeGpuMemoryCluster) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingComputeGpuMemoryClusterLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetComputeGpuMemoryClusterLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ComputeGpuMemoryClusterLifecycleStateEnum Enum with underlying type: string
type ComputeGpuMemoryClusterLifecycleStateEnum string

// Set of constants representing the allowable values for ComputeGpuMemoryClusterLifecycleStateEnum
const (
	ComputeGpuMemoryClusterLifecycleStateCreating ComputeGpuMemoryClusterLifecycleStateEnum = "CREATING"
	ComputeGpuMemoryClusterLifecycleStateActive   ComputeGpuMemoryClusterLifecycleStateEnum = "ACTIVE"
	ComputeGpuMemoryClusterLifecycleStateUpdating ComputeGpuMemoryClusterLifecycleStateEnum = "UPDATING"
	ComputeGpuMemoryClusterLifecycleStateDeleting ComputeGpuMemoryClusterLifecycleStateEnum = "DELETING"
	ComputeGpuMemoryClusterLifecycleStateDeleted  ComputeGpuMemoryClusterLifecycleStateEnum = "DELETED"
)

var mappingComputeGpuMemoryClusterLifecycleStateEnum = map[string]ComputeGpuMemoryClusterLifecycleStateEnum{
	"CREATING": ComputeGpuMemoryClusterLifecycleStateCreating,
	"ACTIVE":   ComputeGpuMemoryClusterLifecycleStateActive,
	"UPDATING": ComputeGpuMemoryClusterLifecycleStateUpdating,
	"DELETING": ComputeGpuMemoryClusterLifecycleStateDeleting,
	"DELETED":  ComputeGpuMemoryClusterLifecycleStateDeleted,
}

var mappingComputeGpuMemoryClusterLifecycleStateEnumLowerCase = map[string]ComputeGpuMemoryClusterLifecycleStateEnum{
	"creating": ComputeGpuMemoryClusterLifecycleStateCreating,
	"active":   ComputeGpuMemoryClusterLifecycleStateActive,
	"updating": ComputeGpuMemoryClusterLifecycleStateUpdating,
	"deleting": ComputeGpuMemoryClusterLifecycleStateDeleting,
	"deleted":  ComputeGpuMemoryClusterLifecycleStateDeleted,
}

// GetComputeGpuMemoryClusterLifecycleStateEnumValues Enumerates the set of values for ComputeGpuMemoryClusterLifecycleStateEnum
func GetComputeGpuMemoryClusterLifecycleStateEnumValues() []ComputeGpuMemoryClusterLifecycleStateEnum {
	values := make([]ComputeGpuMemoryClusterLifecycleStateEnum, 0)
	for _, v := range mappingComputeGpuMemoryClusterLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetComputeGpuMemoryClusterLifecycleStateEnumStringValues Enumerates the set of values in String for ComputeGpuMemoryClusterLifecycleStateEnum
func GetComputeGpuMemoryClusterLifecycleStateEnumStringValues() []string {
	return []string{
		"CREATING",
		"ACTIVE",
		"UPDATING",
		"DELETING",
		"DELETED",
	}
}

// GetMappingComputeGpuMemoryClusterLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingComputeGpuMemoryClusterLifecycleStateEnum(val string) (ComputeGpuMemoryClusterLifecycleStateEnum, bool) {
	enum, ok := mappingComputeGpuMemoryClusterLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
