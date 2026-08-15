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

// UpdateIpv6Details The representation of UpdateIpv6Details
type UpdateIpv6Details struct {

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

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VNIC to reassign the IPv6 to.
	// The VNIC must be in the same subnet as the current VNIC.
	VnicId *string `mandatory:"false" json:"vnicId"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the route table the IP address or VNIC will use. For more information, see
	// Source Based Routing (https://docs.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm#Overview_of_Routing_for_Your_VCN__source_routing).
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// Lifetime of the IP address.
	// There are two types of IPv6 IPs:
	//  - Ephemeral
	//  - Reserved
	Lifetime UpdateIpv6DetailsLifetimeEnum `mandatory:"false" json:"lifetime,omitempty"`
}

func (m UpdateIpv6Details) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateIpv6Details) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingUpdateIpv6DetailsLifetimeEnum(string(m.Lifetime)); !ok && m.Lifetime != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Lifetime: %s. Supported values are: %s.", m.Lifetime, strings.Join(GetUpdateIpv6DetailsLifetimeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UpdateIpv6DetailsLifetimeEnum Enum with underlying type: string
type UpdateIpv6DetailsLifetimeEnum string

// Set of constants representing the allowable values for UpdateIpv6DetailsLifetimeEnum
const (
	UpdateIpv6DetailsLifetimeEphemeral UpdateIpv6DetailsLifetimeEnum = "EPHEMERAL"
	UpdateIpv6DetailsLifetimeReserved  UpdateIpv6DetailsLifetimeEnum = "RESERVED"
)

var mappingUpdateIpv6DetailsLifetimeEnum = map[string]UpdateIpv6DetailsLifetimeEnum{
	"EPHEMERAL": UpdateIpv6DetailsLifetimeEphemeral,
	"RESERVED":  UpdateIpv6DetailsLifetimeReserved,
}

var mappingUpdateIpv6DetailsLifetimeEnumLowerCase = map[string]UpdateIpv6DetailsLifetimeEnum{
	"ephemeral": UpdateIpv6DetailsLifetimeEphemeral,
	"reserved":  UpdateIpv6DetailsLifetimeReserved,
}

// GetUpdateIpv6DetailsLifetimeEnumValues Enumerates the set of values for UpdateIpv6DetailsLifetimeEnum
func GetUpdateIpv6DetailsLifetimeEnumValues() []UpdateIpv6DetailsLifetimeEnum {
	values := make([]UpdateIpv6DetailsLifetimeEnum, 0)
	for _, v := range mappingUpdateIpv6DetailsLifetimeEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateIpv6DetailsLifetimeEnumStringValues Enumerates the set of values in String for UpdateIpv6DetailsLifetimeEnum
func GetUpdateIpv6DetailsLifetimeEnumStringValues() []string {
	return []string{
		"EPHEMERAL",
		"RESERVED",
	}
}

// GetMappingUpdateIpv6DetailsLifetimeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateIpv6DetailsLifetimeEnum(val string) (UpdateIpv6DetailsLifetimeEnum, bool) {
	enum, ok := mappingUpdateIpv6DetailsLifetimeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
