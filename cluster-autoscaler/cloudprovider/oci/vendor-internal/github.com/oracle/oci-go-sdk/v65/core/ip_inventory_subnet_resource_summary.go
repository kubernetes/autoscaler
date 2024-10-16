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

// IpInventorySubnetResourceSummary Provides the IP Inventory details of a subnet and its associated resources.
type IpInventorySubnetResourceSummary struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the IP address.
	IpId *string `mandatory:"false" json:"ipId"`

	// Lists the allocated private IP address.
	IpAddress *string `mandatory:"false" json:"ipAddress"`

	// Lifetime of the allocated private IP address.
	IpAddressLifetime IpInventorySubnetResourceSummaryIpAddressLifetimeEnum `mandatory:"false" json:"ipAddressLifetime,omitempty"`

	// The address range the IP address is assigned from.
	ParentCidr *string `mandatory:"false" json:"parentCidr"`

	// Associated public IP address for the private IP address.
	AssociatedPublicIp *string `mandatory:"false" json:"associatedPublicIp"`

	// Lifetime of the assigned public IP address.
	PublicIpLifetime IpInventorySubnetResourceSummaryPublicIpLifetimeEnum `mandatory:"false" json:"publicIpLifetime,omitempty"`

	// Public IP address Pool the IP address is allocated from.
	AssociatedPublicIpPool IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum `mandatory:"false" json:"associatedPublicIpPool,omitempty"`

	// DNS hostname of the IP address.
	DnsHostName *string `mandatory:"false" json:"dnsHostName"`

	// Name of the created resource.
	AssignedResourceName *string `mandatory:"false" json:"assignedResourceName"`

	// Type of the resource.
	AssignedResourceType IpInventorySubnetResourceSummaryAssignedResourceTypeEnum `mandatory:"false" json:"assignedResourceType,omitempty"`

	// Address type of the allocated private IP address.
	AddressType *string `mandatory:"false" json:"addressType"`

	// Assigned time of the private IP address.
	AssignedTime *common.SDKTime `mandatory:"false" json:"assignedTime"`
}

func (m IpInventorySubnetResourceSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m IpInventorySubnetResourceSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingIpInventorySubnetResourceSummaryIpAddressLifetimeEnum(string(m.IpAddressLifetime)); !ok && m.IpAddressLifetime != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for IpAddressLifetime: %s. Supported values are: %s.", m.IpAddressLifetime, strings.Join(GetIpInventorySubnetResourceSummaryIpAddressLifetimeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingIpInventorySubnetResourceSummaryPublicIpLifetimeEnum(string(m.PublicIpLifetime)); !ok && m.PublicIpLifetime != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for PublicIpLifetime: %s. Supported values are: %s.", m.PublicIpLifetime, strings.Join(GetIpInventorySubnetResourceSummaryPublicIpLifetimeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum(string(m.AssociatedPublicIpPool)); !ok && m.AssociatedPublicIpPool != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AssociatedPublicIpPool: %s. Supported values are: %s.", m.AssociatedPublicIpPool, strings.Join(GetIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnumStringValues(), ",")))
	}
	if _, ok := GetMappingIpInventorySubnetResourceSummaryAssignedResourceTypeEnum(string(m.AssignedResourceType)); !ok && m.AssignedResourceType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AssignedResourceType: %s. Supported values are: %s.", m.AssignedResourceType, strings.Join(GetIpInventorySubnetResourceSummaryAssignedResourceTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// IpInventorySubnetResourceSummaryIpAddressLifetimeEnum Enum with underlying type: string
type IpInventorySubnetResourceSummaryIpAddressLifetimeEnum string

// Set of constants representing the allowable values for IpInventorySubnetResourceSummaryIpAddressLifetimeEnum
const (
	IpInventorySubnetResourceSummaryIpAddressLifetimeEphemeral IpInventorySubnetResourceSummaryIpAddressLifetimeEnum = "Ephemeral"
	IpInventorySubnetResourceSummaryIpAddressLifetimeReserved  IpInventorySubnetResourceSummaryIpAddressLifetimeEnum = "Reserved"
)

var mappingIpInventorySubnetResourceSummaryIpAddressLifetimeEnum = map[string]IpInventorySubnetResourceSummaryIpAddressLifetimeEnum{
	"Ephemeral": IpInventorySubnetResourceSummaryIpAddressLifetimeEphemeral,
	"Reserved":  IpInventorySubnetResourceSummaryIpAddressLifetimeReserved,
}

var mappingIpInventorySubnetResourceSummaryIpAddressLifetimeEnumLowerCase = map[string]IpInventorySubnetResourceSummaryIpAddressLifetimeEnum{
	"ephemeral": IpInventorySubnetResourceSummaryIpAddressLifetimeEphemeral,
	"reserved":  IpInventorySubnetResourceSummaryIpAddressLifetimeReserved,
}

// GetIpInventorySubnetResourceSummaryIpAddressLifetimeEnumValues Enumerates the set of values for IpInventorySubnetResourceSummaryIpAddressLifetimeEnum
func GetIpInventorySubnetResourceSummaryIpAddressLifetimeEnumValues() []IpInventorySubnetResourceSummaryIpAddressLifetimeEnum {
	values := make([]IpInventorySubnetResourceSummaryIpAddressLifetimeEnum, 0)
	for _, v := range mappingIpInventorySubnetResourceSummaryIpAddressLifetimeEnum {
		values = append(values, v)
	}
	return values
}

// GetIpInventorySubnetResourceSummaryIpAddressLifetimeEnumStringValues Enumerates the set of values in String for IpInventorySubnetResourceSummaryIpAddressLifetimeEnum
func GetIpInventorySubnetResourceSummaryIpAddressLifetimeEnumStringValues() []string {
	return []string{
		"Ephemeral",
		"Reserved",
	}
}

// GetMappingIpInventorySubnetResourceSummaryIpAddressLifetimeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpInventorySubnetResourceSummaryIpAddressLifetimeEnum(val string) (IpInventorySubnetResourceSummaryIpAddressLifetimeEnum, bool) {
	enum, ok := mappingIpInventorySubnetResourceSummaryIpAddressLifetimeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpInventorySubnetResourceSummaryPublicIpLifetimeEnum Enum with underlying type: string
type IpInventorySubnetResourceSummaryPublicIpLifetimeEnum string

// Set of constants representing the allowable values for IpInventorySubnetResourceSummaryPublicIpLifetimeEnum
const (
	IpInventorySubnetResourceSummaryPublicIpLifetimeEphemeral IpInventorySubnetResourceSummaryPublicIpLifetimeEnum = "Ephemeral"
	IpInventorySubnetResourceSummaryPublicIpLifetimeReserved  IpInventorySubnetResourceSummaryPublicIpLifetimeEnum = "Reserved"
)

var mappingIpInventorySubnetResourceSummaryPublicIpLifetimeEnum = map[string]IpInventorySubnetResourceSummaryPublicIpLifetimeEnum{
	"Ephemeral": IpInventorySubnetResourceSummaryPublicIpLifetimeEphemeral,
	"Reserved":  IpInventorySubnetResourceSummaryPublicIpLifetimeReserved,
}

var mappingIpInventorySubnetResourceSummaryPublicIpLifetimeEnumLowerCase = map[string]IpInventorySubnetResourceSummaryPublicIpLifetimeEnum{
	"ephemeral": IpInventorySubnetResourceSummaryPublicIpLifetimeEphemeral,
	"reserved":  IpInventorySubnetResourceSummaryPublicIpLifetimeReserved,
}

// GetIpInventorySubnetResourceSummaryPublicIpLifetimeEnumValues Enumerates the set of values for IpInventorySubnetResourceSummaryPublicIpLifetimeEnum
func GetIpInventorySubnetResourceSummaryPublicIpLifetimeEnumValues() []IpInventorySubnetResourceSummaryPublicIpLifetimeEnum {
	values := make([]IpInventorySubnetResourceSummaryPublicIpLifetimeEnum, 0)
	for _, v := range mappingIpInventorySubnetResourceSummaryPublicIpLifetimeEnum {
		values = append(values, v)
	}
	return values
}

// GetIpInventorySubnetResourceSummaryPublicIpLifetimeEnumStringValues Enumerates the set of values in String for IpInventorySubnetResourceSummaryPublicIpLifetimeEnum
func GetIpInventorySubnetResourceSummaryPublicIpLifetimeEnumStringValues() []string {
	return []string{
		"Ephemeral",
		"Reserved",
	}
}

// GetMappingIpInventorySubnetResourceSummaryPublicIpLifetimeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpInventorySubnetResourceSummaryPublicIpLifetimeEnum(val string) (IpInventorySubnetResourceSummaryPublicIpLifetimeEnum, bool) {
	enum, ok := mappingIpInventorySubnetResourceSummaryPublicIpLifetimeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum Enum with underlying type: string
type IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum string

// Set of constants representing the allowable values for IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum
const (
	IpInventorySubnetResourceSummaryAssociatedPublicIpPoolOracle IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum = "ORACLE"
	IpInventorySubnetResourceSummaryAssociatedPublicIpPoolByoip  IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum = "BYOIP"
)

var mappingIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum = map[string]IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum{
	"ORACLE": IpInventorySubnetResourceSummaryAssociatedPublicIpPoolOracle,
	"BYOIP":  IpInventorySubnetResourceSummaryAssociatedPublicIpPoolByoip,
}

var mappingIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnumLowerCase = map[string]IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum{
	"oracle": IpInventorySubnetResourceSummaryAssociatedPublicIpPoolOracle,
	"byoip":  IpInventorySubnetResourceSummaryAssociatedPublicIpPoolByoip,
}

// GetIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnumValues Enumerates the set of values for IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum
func GetIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnumValues() []IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum {
	values := make([]IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum, 0)
	for _, v := range mappingIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum {
		values = append(values, v)
	}
	return values
}

// GetIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnumStringValues Enumerates the set of values in String for IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum
func GetIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnumStringValues() []string {
	return []string{
		"ORACLE",
		"BYOIP",
	}
}

// GetMappingIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum(val string) (IpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnum, bool) {
	enum, ok := mappingIpInventorySubnetResourceSummaryAssociatedPublicIpPoolEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpInventorySubnetResourceSummaryAssignedResourceTypeEnum Enum with underlying type: string
type IpInventorySubnetResourceSummaryAssignedResourceTypeEnum string

// Set of constants representing the allowable values for IpInventorySubnetResourceSummaryAssignedResourceTypeEnum
const (
	IpInventorySubnetResourceSummaryAssignedResourceTypeResource IpInventorySubnetResourceSummaryAssignedResourceTypeEnum = "Resource"
)

var mappingIpInventorySubnetResourceSummaryAssignedResourceTypeEnum = map[string]IpInventorySubnetResourceSummaryAssignedResourceTypeEnum{
	"Resource": IpInventorySubnetResourceSummaryAssignedResourceTypeResource,
}

var mappingIpInventorySubnetResourceSummaryAssignedResourceTypeEnumLowerCase = map[string]IpInventorySubnetResourceSummaryAssignedResourceTypeEnum{
	"resource": IpInventorySubnetResourceSummaryAssignedResourceTypeResource,
}

// GetIpInventorySubnetResourceSummaryAssignedResourceTypeEnumValues Enumerates the set of values for IpInventorySubnetResourceSummaryAssignedResourceTypeEnum
func GetIpInventorySubnetResourceSummaryAssignedResourceTypeEnumValues() []IpInventorySubnetResourceSummaryAssignedResourceTypeEnum {
	values := make([]IpInventorySubnetResourceSummaryAssignedResourceTypeEnum, 0)
	for _, v := range mappingIpInventorySubnetResourceSummaryAssignedResourceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetIpInventorySubnetResourceSummaryAssignedResourceTypeEnumStringValues Enumerates the set of values in String for IpInventorySubnetResourceSummaryAssignedResourceTypeEnum
func GetIpInventorySubnetResourceSummaryAssignedResourceTypeEnumStringValues() []string {
	return []string{
		"Resource",
	}
}

// GetMappingIpInventorySubnetResourceSummaryAssignedResourceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpInventorySubnetResourceSummaryAssignedResourceTypeEnum(val string) (IpInventorySubnetResourceSummaryAssignedResourceTypeEnum, bool) {
	enum, ok := mappingIpInventorySubnetResourceSummaryAssignedResourceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
