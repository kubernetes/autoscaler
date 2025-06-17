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

// CreateIpv6Details The representation of CreateIpv6Details
type CreateIpv6Details struct {

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

	// An IPv6 address of your choice. Must be an available IP address within
	// the subnet's CIDR. If you don't specify a value, Oracle automatically
	// assigns an IPv6 address from the subnet. The subnet is the one that
	// contains the VNIC you specify in `vnicId`.
	// Example: `2001:DB8::`
	IpAddress *string `mandatory:"false" json:"ipAddress"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VNIC to assign the IPv6 to. The
	// IPv6 will be in the VNIC's subnet.
	VnicId *string `mandatory:"false" json:"vnicId"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet from which the IPv6 is to be drawn. The IP address,
	// *if supplied*, must be valid for the given subnet, only valid for reserved IPs currently.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// Lifetime of the IP address.
	// There are two types of IPv6 IPs:
	//  - Ephemeral
	//  - Reserved
	Lifetime CreateIpv6DetailsLifetimeEnum `mandatory:"false" json:"lifetime,omitempty"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the route table the IP address or VNIC will use. For more information, see
	// Source Based Routing (https://docs.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm#Overview_of_Routing_for_Your_VCN__source_routing).
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// The IPv6 prefix allocated to the subnet. This is required if more than one IPv6 prefix exists on the subnet.
	Ipv6SubnetCidr *string `mandatory:"false" json:"ipv6SubnetCidr"`
}

func (m CreateIpv6Details) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateIpv6Details) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCreateIpv6DetailsLifetimeEnum(string(m.Lifetime)); !ok && m.Lifetime != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Lifetime: %s. Supported values are: %s.", m.Lifetime, strings.Join(GetCreateIpv6DetailsLifetimeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateIpv6DetailsLifetimeEnum Enum with underlying type: string
type CreateIpv6DetailsLifetimeEnum string

// Set of constants representing the allowable values for CreateIpv6DetailsLifetimeEnum
const (
	CreateIpv6DetailsLifetimeEphemeral CreateIpv6DetailsLifetimeEnum = "EPHEMERAL"
	CreateIpv6DetailsLifetimeReserved  CreateIpv6DetailsLifetimeEnum = "RESERVED"
)

var mappingCreateIpv6DetailsLifetimeEnum = map[string]CreateIpv6DetailsLifetimeEnum{
	"EPHEMERAL": CreateIpv6DetailsLifetimeEphemeral,
	"RESERVED":  CreateIpv6DetailsLifetimeReserved,
}

var mappingCreateIpv6DetailsLifetimeEnumLowerCase = map[string]CreateIpv6DetailsLifetimeEnum{
	"ephemeral": CreateIpv6DetailsLifetimeEphemeral,
	"reserved":  CreateIpv6DetailsLifetimeReserved,
}

// GetCreateIpv6DetailsLifetimeEnumValues Enumerates the set of values for CreateIpv6DetailsLifetimeEnum
func GetCreateIpv6DetailsLifetimeEnumValues() []CreateIpv6DetailsLifetimeEnum {
	values := make([]CreateIpv6DetailsLifetimeEnum, 0)
	for _, v := range mappingCreateIpv6DetailsLifetimeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateIpv6DetailsLifetimeEnumStringValues Enumerates the set of values in String for CreateIpv6DetailsLifetimeEnum
func GetCreateIpv6DetailsLifetimeEnumStringValues() []string {
	return []string{
		"EPHEMERAL",
		"RESERVED",
	}
}

// GetMappingCreateIpv6DetailsLifetimeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateIpv6DetailsLifetimeEnum(val string) (CreateIpv6DetailsLifetimeEnum, bool) {
	enum, ok := mappingCreateIpv6DetailsLifetimeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
