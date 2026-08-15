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

// ComputeGpuMemoryClusterInstanceSummary The customer facing GPU memory cluster instance object details.
type ComputeGpuMemoryClusterInstanceSummary struct {

	// The availability domain of the GPU memory cluster instance.
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for the Customer-unique GPU memory cluster instance
	Id *string `mandatory:"false" json:"id"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for the compartment
	// compartment.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The fault domain the GPU memory cluster instance is running in.
	FaultDomain *string `mandatory:"false" json:"faultDomain"`

	// Configuration to be used for this GPU Memory Cluster instance.
	InstanceConfigurationId *string `mandatory:"false" json:"instanceConfigurationId"`

	// The region that contains the availability domain the instance is running in.
	Region *string `mandatory:"false" json:"region"`

	// The shape of an instance. The shape determines the number of CPUs, amount of memory,
	// and other resources allocated to the instance. The shape determines the number of CPUs,
	// the amount of memory, and other resources allocated to the instance.
	// You can list all available shapes by calling ListShapes.
	InstanceShape *string `mandatory:"false" json:"instanceShape"`

	// The lifecycle state of the GPU memory cluster instance
	LifecycleState ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The date and time the GPU memory cluster instance was created.
	// Example: `2016-09-15T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m ComputeGpuMemoryClusterInstanceSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ComputeGpuMemoryClusterInstanceSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum Enum with underlying type: string
type ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum
const (
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateMoving        ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "MOVING"
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateProvisioning  ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "PROVISIONING"
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateRunning       ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "RUNNING"
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateStarting      ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "STARTING"
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateStopping      ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "STOPPING"
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateStopped       ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "STOPPED"
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateSuspending    ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "SUSPENDING"
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateSuspended     ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "SUSPENDED"
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateCreatingImage ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "CREATING_IMAGE"
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateTerminating   ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "TERMINATING"
	ComputeGpuMemoryClusterInstanceSummaryLifecycleStateTerminated    ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = "TERMINATED"
)

var mappingComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum = map[string]ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum{
	"MOVING":         ComputeGpuMemoryClusterInstanceSummaryLifecycleStateMoving,
	"PROVISIONING":   ComputeGpuMemoryClusterInstanceSummaryLifecycleStateProvisioning,
	"RUNNING":        ComputeGpuMemoryClusterInstanceSummaryLifecycleStateRunning,
	"STARTING":       ComputeGpuMemoryClusterInstanceSummaryLifecycleStateStarting,
	"STOPPING":       ComputeGpuMemoryClusterInstanceSummaryLifecycleStateStopping,
	"STOPPED":        ComputeGpuMemoryClusterInstanceSummaryLifecycleStateStopped,
	"SUSPENDING":     ComputeGpuMemoryClusterInstanceSummaryLifecycleStateSuspending,
	"SUSPENDED":      ComputeGpuMemoryClusterInstanceSummaryLifecycleStateSuspended,
	"CREATING_IMAGE": ComputeGpuMemoryClusterInstanceSummaryLifecycleStateCreatingImage,
	"TERMINATING":    ComputeGpuMemoryClusterInstanceSummaryLifecycleStateTerminating,
	"TERMINATED":     ComputeGpuMemoryClusterInstanceSummaryLifecycleStateTerminated,
}

var mappingComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnumLowerCase = map[string]ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum{
	"moving":         ComputeGpuMemoryClusterInstanceSummaryLifecycleStateMoving,
	"provisioning":   ComputeGpuMemoryClusterInstanceSummaryLifecycleStateProvisioning,
	"running":        ComputeGpuMemoryClusterInstanceSummaryLifecycleStateRunning,
	"starting":       ComputeGpuMemoryClusterInstanceSummaryLifecycleStateStarting,
	"stopping":       ComputeGpuMemoryClusterInstanceSummaryLifecycleStateStopping,
	"stopped":        ComputeGpuMemoryClusterInstanceSummaryLifecycleStateStopped,
	"suspending":     ComputeGpuMemoryClusterInstanceSummaryLifecycleStateSuspending,
	"suspended":      ComputeGpuMemoryClusterInstanceSummaryLifecycleStateSuspended,
	"creating_image": ComputeGpuMemoryClusterInstanceSummaryLifecycleStateCreatingImage,
	"terminating":    ComputeGpuMemoryClusterInstanceSummaryLifecycleStateTerminating,
	"terminated":     ComputeGpuMemoryClusterInstanceSummaryLifecycleStateTerminated,
}

// GetComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnumValues Enumerates the set of values for ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum
func GetComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnumValues() []ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum {
	values := make([]ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum, 0)
	for _, v := range mappingComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnumStringValues Enumerates the set of values in String for ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum
func GetComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnumStringValues() []string {
	return []string{
		"MOVING",
		"PROVISIONING",
		"RUNNING",
		"STARTING",
		"STOPPING",
		"STOPPED",
		"SUSPENDING",
		"SUSPENDED",
		"CREATING_IMAGE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum(val string) (ComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnum, bool) {
	enum, ok := mappingComputeGpuMemoryClusterInstanceSummaryLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
