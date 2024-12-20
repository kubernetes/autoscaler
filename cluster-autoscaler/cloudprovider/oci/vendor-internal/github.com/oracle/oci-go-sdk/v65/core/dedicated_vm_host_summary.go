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

// DedicatedVmHostSummary A dedicated virtual machine (VM) host lets you host multiple instances on a dedicated server that is not shared with other tenancies.
type DedicatedVmHostSummary struct {

	// The availability domain the dedicated VM host is running in.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment that contains the dedicated VM host.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The shape of the dedicated VM host. The shape determines the number of CPUs and
	// other resources available for VMs.
	DedicatedVmHostShape *string `mandatory:"true" json:"dedicatedVmHostShape"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the dedicated VM host.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the dedicated VM host.
	LifecycleState DedicatedVmHostSummaryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the dedicated VM host was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The current available OCPUs of the dedicated VM host.
	RemainingOcpus *float32 `mandatory:"true" json:"remainingOcpus"`

	// The current total OCPUs of the dedicated VM host.
	TotalOcpus *float32 `mandatory:"true" json:"totalOcpus"`

	// The fault domain for the dedicated VM host's assigned instances. For more information, see Fault Domains.
	// If you do not specify the fault domain, the system selects one for you. To change the fault domain for a dedicated VM host,
	// delete it and create a new dedicated VM host in the preferred fault domain.
	// To get a list of fault domains, use the ListFaultDomains operation in the Identity and Access Management Service API.
	// Example: `FAULT-DOMAIN-1`
	FaultDomain *string `mandatory:"false" json:"faultDomain"`

	// The current total memory of the dedicated VM host, in GBs.
	TotalMemoryInGBs *float32 `mandatory:"false" json:"totalMemoryInGBs"`

	// The current available memory of the dedicated VM host, in GBs.
	RemainingMemoryInGBs *float32 `mandatory:"false" json:"remainingMemoryInGBs"`
}

func (m DedicatedVmHostSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DedicatedVmHostSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingDedicatedVmHostSummaryLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetDedicatedVmHostSummaryLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// DedicatedVmHostSummaryLifecycleStateEnum Enum with underlying type: string
type DedicatedVmHostSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for DedicatedVmHostSummaryLifecycleStateEnum
const (
	DedicatedVmHostSummaryLifecycleStateCreating DedicatedVmHostSummaryLifecycleStateEnum = "CREATING"
	DedicatedVmHostSummaryLifecycleStateActive   DedicatedVmHostSummaryLifecycleStateEnum = "ACTIVE"
	DedicatedVmHostSummaryLifecycleStateUpdating DedicatedVmHostSummaryLifecycleStateEnum = "UPDATING"
	DedicatedVmHostSummaryLifecycleStateDeleting DedicatedVmHostSummaryLifecycleStateEnum = "DELETING"
	DedicatedVmHostSummaryLifecycleStateDeleted  DedicatedVmHostSummaryLifecycleStateEnum = "DELETED"
	DedicatedVmHostSummaryLifecycleStateFailed   DedicatedVmHostSummaryLifecycleStateEnum = "FAILED"
)

var mappingDedicatedVmHostSummaryLifecycleStateEnum = map[string]DedicatedVmHostSummaryLifecycleStateEnum{
	"CREATING": DedicatedVmHostSummaryLifecycleStateCreating,
	"ACTIVE":   DedicatedVmHostSummaryLifecycleStateActive,
	"UPDATING": DedicatedVmHostSummaryLifecycleStateUpdating,
	"DELETING": DedicatedVmHostSummaryLifecycleStateDeleting,
	"DELETED":  DedicatedVmHostSummaryLifecycleStateDeleted,
	"FAILED":   DedicatedVmHostSummaryLifecycleStateFailed,
}

var mappingDedicatedVmHostSummaryLifecycleStateEnumLowerCase = map[string]DedicatedVmHostSummaryLifecycleStateEnum{
	"creating": DedicatedVmHostSummaryLifecycleStateCreating,
	"active":   DedicatedVmHostSummaryLifecycleStateActive,
	"updating": DedicatedVmHostSummaryLifecycleStateUpdating,
	"deleting": DedicatedVmHostSummaryLifecycleStateDeleting,
	"deleted":  DedicatedVmHostSummaryLifecycleStateDeleted,
	"failed":   DedicatedVmHostSummaryLifecycleStateFailed,
}

// GetDedicatedVmHostSummaryLifecycleStateEnumValues Enumerates the set of values for DedicatedVmHostSummaryLifecycleStateEnum
func GetDedicatedVmHostSummaryLifecycleStateEnumValues() []DedicatedVmHostSummaryLifecycleStateEnum {
	values := make([]DedicatedVmHostSummaryLifecycleStateEnum, 0)
	for _, v := range mappingDedicatedVmHostSummaryLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetDedicatedVmHostSummaryLifecycleStateEnumStringValues Enumerates the set of values in String for DedicatedVmHostSummaryLifecycleStateEnum
func GetDedicatedVmHostSummaryLifecycleStateEnumStringValues() []string {
	return []string{
		"CREATING",
		"ACTIVE",
		"UPDATING",
		"DELETING",
		"DELETED",
		"FAILED",
	}
}

// GetMappingDedicatedVmHostSummaryLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDedicatedVmHostSummaryLifecycleStateEnum(val string) (DedicatedVmHostSummaryLifecycleStateEnum, bool) {
	enum, ok := mappingDedicatedVmHostSummaryLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
