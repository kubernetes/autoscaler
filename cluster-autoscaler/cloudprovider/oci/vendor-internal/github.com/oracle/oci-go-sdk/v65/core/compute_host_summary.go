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

// ComputeHostSummary Summary information for a compute host.
type ComputeHostSummary struct {

	// The availability domain of the compute host.
	// Example: `Uocm:US-CHICAGO-1-AD-2`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for the compartment. This should always be the root
	// compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for the Customer-unique host
	Id *string `mandatory:"true" json:"id"`

	// A fault domain is a grouping of hardware and infrastructure within an availability domain.
	// Each availability domain contains three fault domains. Fault domains let you distribute your
	// instances so that they are not on the same physical hardware within a single availability domain.
	// A hardware failure or Compute hardware maintenance that affects one fault domain does not affect
	// instances in other fault domains.
	// This field is the Fault domain of the host
	FaultDomain *string `mandatory:"true" json:"faultDomain"`

	// The shape of host
	Shape *string `mandatory:"true" json:"shape"`

	// The heathy state of the host
	Health ComputeHostHealthEnum `mandatory:"true" json:"health"`

	// The lifecycle state of the host
	LifecycleState ComputeHostLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// While listing a host the user will know if they have an impacted component or not.
	// The user will have to issue a get host to see details.
	HasImpactedComponents *bool `mandatory:"true" json:"hasImpactedComponents"`

	// The date and time that the compute host record was created, in the format defined by RFC3339 (https://tools
	// .ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The date and time that the compute host record was updated, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	//
	// Example: `2016-08-25T21:10:29.600Z`
	TimeUpdated *common.SDKTime `mandatory:"true" json:"timeUpdated"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for Customer-unique HPC Island
	HpcIslandId *string `mandatory:"false" json:"hpcIslandId"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for Customer-unique Network Block
	NetworkBlockId *string `mandatory:"false" json:"networkBlockId"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for Customer-unique Local Block
	LocalBlockId *string `mandatory:"false" json:"localBlockId"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for Customer-unique GPU Memory Fabric
	GpuMemoryFabricId *string `mandatory:"false" json:"gpuMemoryFabricId"`

	// The public OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for the Virtual Machine or Bare Metal instance
	InstanceId *string `mandatory:"false" json:"instanceId"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for the Capacity Reserver that is currently on host
	CapacityReservationId *string `mandatory:"false" json:"capacityReservationId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m ComputeHostSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ComputeHostSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingComputeHostHealthEnum(string(m.Health)); !ok && m.Health != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Health: %s. Supported values are: %s.", m.Health, strings.Join(GetComputeHostHealthEnumStringValues(), ",")))
	}
	if _, ok := GetMappingComputeHostLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetComputeHostLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
