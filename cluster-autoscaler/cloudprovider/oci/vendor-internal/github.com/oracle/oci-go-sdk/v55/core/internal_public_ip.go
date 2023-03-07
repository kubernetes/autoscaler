// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// InternalPublicIp A model for public IPs that are for internal use
// A *public IP* is a conceptual term that refers to a public IP address and related properties.
// The `publicIp` object is the API representation of a public IP.
// There are two types of public IPs:
// 1. Ephemeral
// 2. Reserved
// For more information and comparison of the two types,
// see Public IP Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingpublicIPs.htm).
type InternalPublicIp struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the public IP. For an ephemeral public IP, this is
	// the compartment of its assigned entity (which can be a private IP or a regional entity such
	// as a NAT gateway). For a reserved public IP that is currently assigned,
	// its compartment can be different from the assigned private IP's.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The public IP's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// Defines when the public IP is deleted and released back to Oracle's public IP pool.
	// * `EPHEMERAL`: The lifetime is tied to the lifetime of its assigned entity. An ephemeral
	// public IP must always be assigned to an entity. If the assigned entity is a private IP,
	// the ephemeral public IP is automatically deleted when the private IP is deleted, when
	// the VNIC is terminated, or when the instance is terminated. If the assigned entity is a
	// NatGateway, the ephemeral public IP is automatically
	// deleted when the NAT gateway is terminated.
	// * `RESERVED`: You control the public IP's lifetime. You can delete a reserved public IP
	// whenever you like. It does not need to be assigned to a private IP at all times.
	// For more information and comparison of the two types,
	// see Public IP Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingpublicIPs.htm).
	Lifetime InternalPublicIpLifetimeEnum `mandatory:"true" json:"lifetime"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the entity the public IP is assigned to, or in the process of
	// being assigned to.
	AssignedEntityId *string `mandatory:"false" json:"assignedEntityId"`

	// The type of entity the public IP is assigned to, or in the process of being
	// assigned to.
	AssignedEntityType InternalPublicIpAssignedEntityTypeEnum `mandatory:"false" json:"assignedEntityType,omitempty"`

	// The public IP's availability domain. This property is set only for ephemeral public IPs
	// that are assigned to a private IP (that is, when the `scope` of the public IP is set to
	// AVAILABILITY_DOMAIN). The value is the availability domain of the assigned private IP.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The public IP address of the `publicIp` object.
	// Example: `203.0.113.2`
	IpAddress *string `mandatory:"false" json:"ipAddress"`

	// The public IP's current state.
	LifecycleState InternalPublicIpLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Deprecated. Use `assignedEntityId` instead.
	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the private IP that the public IP is currently assigned to, or in the
	// process of being assigned to.
	// **Note:** This is `null` if the public IP is not assigned to a private IP, or is
	// in the process of being assigned to one.
	PrivateIpId *string `mandatory:"false" json:"privateIpId"`

	// Whether the public IP is regional or specific to a particular availability domain.
	// * `REGION`: The public IP exists within a region and is assigned to a regional entity
	// (such as a NatGateway), or can be assigned to a private
	// IP in any availability domain in the region. Reserved public IPs and ephemeral public IPs
	// assigned to a regional entity have `scope` = `REGION`.
	// * `AVAILABILITY_DOMAIN`: The public IP exists within the availability domain of the entity
	// it's assigned to, which is specified by the `availabilityDomain` property of the public IP object.
	// Ephemeral public IPs that are assigned to private IPs have `scope` = `AVAILABILITY_DOMAIN`.
	Scope InternalPublicIpScopeEnum `mandatory:"false" json:"scope,omitempty"`

	// The date and time the public IP was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the pool object created in the current tenancy.
	PublicIpPoolId *string `mandatory:"false" json:"publicIpPoolId"`
}

func (m InternalPublicIp) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InternalPublicIp) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingInternalPublicIpLifetimeEnum[string(m.Lifetime)]; !ok && m.Lifetime != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Lifetime: %s. Supported values are: %s.", m.Lifetime, strings.Join(GetInternalPublicIpLifetimeEnumStringValues(), ",")))
	}

	if _, ok := mappingInternalPublicIpAssignedEntityTypeEnum[string(m.AssignedEntityType)]; !ok && m.AssignedEntityType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AssignedEntityType: %s. Supported values are: %s.", m.AssignedEntityType, strings.Join(GetInternalPublicIpAssignedEntityTypeEnumStringValues(), ",")))
	}
	if _, ok := mappingInternalPublicIpLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInternalPublicIpLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := mappingInternalPublicIpScopeEnum[string(m.Scope)]; !ok && m.Scope != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Scope: %s. Supported values are: %s.", m.Scope, strings.Join(GetInternalPublicIpScopeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InternalPublicIpAssignedEntityTypeEnum Enum with underlying type: string
type InternalPublicIpAssignedEntityTypeEnum string

// Set of constants representing the allowable values for InternalPublicIpAssignedEntityTypeEnum
const (
	InternalPublicIpAssignedEntityTypePrivateIp  InternalPublicIpAssignedEntityTypeEnum = "PRIVATE_IP"
	InternalPublicIpAssignedEntityTypeNatGateway InternalPublicIpAssignedEntityTypeEnum = "NAT_GATEWAY"
)

var mappingInternalPublicIpAssignedEntityTypeEnum = map[string]InternalPublicIpAssignedEntityTypeEnum{
	"PRIVATE_IP":  InternalPublicIpAssignedEntityTypePrivateIp,
	"NAT_GATEWAY": InternalPublicIpAssignedEntityTypeNatGateway,
}

// GetInternalPublicIpAssignedEntityTypeEnumValues Enumerates the set of values for InternalPublicIpAssignedEntityTypeEnum
func GetInternalPublicIpAssignedEntityTypeEnumValues() []InternalPublicIpAssignedEntityTypeEnum {
	values := make([]InternalPublicIpAssignedEntityTypeEnum, 0)
	for _, v := range mappingInternalPublicIpAssignedEntityTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalPublicIpAssignedEntityTypeEnumStringValues Enumerates the set of values in String for InternalPublicIpAssignedEntityTypeEnum
func GetInternalPublicIpAssignedEntityTypeEnumStringValues() []string {
	return []string{
		"PRIVATE_IP",
		"NAT_GATEWAY",
	}
}

// InternalPublicIpLifecycleStateEnum Enum with underlying type: string
type InternalPublicIpLifecycleStateEnum string

// Set of constants representing the allowable values for InternalPublicIpLifecycleStateEnum
const (
	InternalPublicIpLifecycleStateProvisioning InternalPublicIpLifecycleStateEnum = "PROVISIONING"
	InternalPublicIpLifecycleStateAvailable    InternalPublicIpLifecycleStateEnum = "AVAILABLE"
	InternalPublicIpLifecycleStateAssigning    InternalPublicIpLifecycleStateEnum = "ASSIGNING"
	InternalPublicIpLifecycleStateAssigned     InternalPublicIpLifecycleStateEnum = "ASSIGNED"
	InternalPublicIpLifecycleStateUnassigning  InternalPublicIpLifecycleStateEnum = "UNASSIGNING"
	InternalPublicIpLifecycleStateUnassigned   InternalPublicIpLifecycleStateEnum = "UNASSIGNED"
	InternalPublicIpLifecycleStateTerminating  InternalPublicIpLifecycleStateEnum = "TERMINATING"
	InternalPublicIpLifecycleStateTerminated   InternalPublicIpLifecycleStateEnum = "TERMINATED"
)

var mappingInternalPublicIpLifecycleStateEnum = map[string]InternalPublicIpLifecycleStateEnum{
	"PROVISIONING": InternalPublicIpLifecycleStateProvisioning,
	"AVAILABLE":    InternalPublicIpLifecycleStateAvailable,
	"ASSIGNING":    InternalPublicIpLifecycleStateAssigning,
	"ASSIGNED":     InternalPublicIpLifecycleStateAssigned,
	"UNASSIGNING":  InternalPublicIpLifecycleStateUnassigning,
	"UNASSIGNED":   InternalPublicIpLifecycleStateUnassigned,
	"TERMINATING":  InternalPublicIpLifecycleStateTerminating,
	"TERMINATED":   InternalPublicIpLifecycleStateTerminated,
}

// GetInternalPublicIpLifecycleStateEnumValues Enumerates the set of values for InternalPublicIpLifecycleStateEnum
func GetInternalPublicIpLifecycleStateEnumValues() []InternalPublicIpLifecycleStateEnum {
	values := make([]InternalPublicIpLifecycleStateEnum, 0)
	for _, v := range mappingInternalPublicIpLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalPublicIpLifecycleStateEnumStringValues Enumerates the set of values in String for InternalPublicIpLifecycleStateEnum
func GetInternalPublicIpLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"ASSIGNING",
		"ASSIGNED",
		"UNASSIGNING",
		"UNASSIGNED",
		"TERMINATING",
		"TERMINATED",
	}
}

// InternalPublicIpLifetimeEnum Enum with underlying type: string
type InternalPublicIpLifetimeEnum string

// Set of constants representing the allowable values for InternalPublicIpLifetimeEnum
const (
	InternalPublicIpLifetimeEphemeral InternalPublicIpLifetimeEnum = "EPHEMERAL"
	InternalPublicIpLifetimeReserved  InternalPublicIpLifetimeEnum = "RESERVED"
)

var mappingInternalPublicIpLifetimeEnum = map[string]InternalPublicIpLifetimeEnum{
	"EPHEMERAL": InternalPublicIpLifetimeEphemeral,
	"RESERVED":  InternalPublicIpLifetimeReserved,
}

// GetInternalPublicIpLifetimeEnumValues Enumerates the set of values for InternalPublicIpLifetimeEnum
func GetInternalPublicIpLifetimeEnumValues() []InternalPublicIpLifetimeEnum {
	values := make([]InternalPublicIpLifetimeEnum, 0)
	for _, v := range mappingInternalPublicIpLifetimeEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalPublicIpLifetimeEnumStringValues Enumerates the set of values in String for InternalPublicIpLifetimeEnum
func GetInternalPublicIpLifetimeEnumStringValues() []string {
	return []string{
		"EPHEMERAL",
		"RESERVED",
	}
}

// InternalPublicIpScopeEnum Enum with underlying type: string
type InternalPublicIpScopeEnum string

// Set of constants representing the allowable values for InternalPublicIpScopeEnum
const (
	InternalPublicIpScopeRegion             InternalPublicIpScopeEnum = "REGION"
	InternalPublicIpScopeAvailabilityDomain InternalPublicIpScopeEnum = "AVAILABILITY_DOMAIN"
)

var mappingInternalPublicIpScopeEnum = map[string]InternalPublicIpScopeEnum{
	"REGION":              InternalPublicIpScopeRegion,
	"AVAILABILITY_DOMAIN": InternalPublicIpScopeAvailabilityDomain,
}

// GetInternalPublicIpScopeEnumValues Enumerates the set of values for InternalPublicIpScopeEnum
func GetInternalPublicIpScopeEnumValues() []InternalPublicIpScopeEnum {
	values := make([]InternalPublicIpScopeEnum, 0)
	for _, v := range mappingInternalPublicIpScopeEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalPublicIpScopeEnumStringValues Enumerates the set of values in String for InternalPublicIpScopeEnum
func GetInternalPublicIpScopeEnumStringValues() []string {
	return []string{
		"REGION",
		"AVAILABILITY_DOMAIN",
	}
}
