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

// CreatePrivateIpDetails The representation of CreatePrivateIpDetails
type CreatePrivateIpDetails struct {

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

	// The hostname for the private IP. Used for DNS. The value
	// is the hostname portion of the private IP's fully qualified domain name (FQDN)
	// (for example, `bminstance1` in FQDN `bminstance1.subnet123.vcn1.oraclevcn.com`).
	// Must be unique across all VNICs in the subnet and comply with
	// RFC 952 (https://tools.ietf.org/html/rfc952) and
	// RFC 1123 (https://tools.ietf.org/html/rfc1123).
	// For more information, see
	// DNS in Your Virtual Cloud Network (https://docs.oracle.com/iaas/Content/Network/Concepts/dns.htm).
	// Example: `bminstance1`
	HostnameLabel *string `mandatory:"false" json:"hostnameLabel"`

	// A private IP address of your choice. Must be an available IP address within
	// the subnet's CIDR. If you don't specify a value, Oracle automatically
	// assigns a private IP address from the subnet.
	// Example: `10.0.3.3`
	IpAddress *string `mandatory:"false" json:"ipAddress"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VNIC to assign the private IP to. The VNIC and private IP
	// must be in the same subnet.
	VnicId *string `mandatory:"false" json:"vnicId"`

	// Use this attribute only with the Oracle Cloud VMware Solution.
	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VLAN from which the private IP is to be drawn. The IP address,
	// *if supplied*, must be valid for the given VLAN. See Vlan.
	VlanId *string `mandatory:"false" json:"vlanId"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet from which the private IP is to be drawn. The IP address,
	// *if supplied*, must be valid for the given subnet.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// Lifetime of the IP address.
	// There are two types of IPv6 IPs:
	//  - Ephemeral
	//  - Reserved
	Lifetime CreatePrivateIpDetailsLifetimeEnum `mandatory:"false" json:"lifetime,omitempty"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the route table the IP address or VNIC will use. For more information, see
	// Source Based Routing (https://docs.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm#Overview_of_Routing_for_Your_VCN__source_routing).
	RouteTableId *string `mandatory:"false" json:"routeTableId"`
}

func (m CreatePrivateIpDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreatePrivateIpDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCreatePrivateIpDetailsLifetimeEnum(string(m.Lifetime)); !ok && m.Lifetime != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Lifetime: %s. Supported values are: %s.", m.Lifetime, strings.Join(GetCreatePrivateIpDetailsLifetimeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreatePrivateIpDetailsLifetimeEnum Enum with underlying type: string
type CreatePrivateIpDetailsLifetimeEnum string

// Set of constants representing the allowable values for CreatePrivateIpDetailsLifetimeEnum
const (
	CreatePrivateIpDetailsLifetimeEphemeral CreatePrivateIpDetailsLifetimeEnum = "EPHEMERAL"
	CreatePrivateIpDetailsLifetimeReserved  CreatePrivateIpDetailsLifetimeEnum = "RESERVED"
)

var mappingCreatePrivateIpDetailsLifetimeEnum = map[string]CreatePrivateIpDetailsLifetimeEnum{
	"EPHEMERAL": CreatePrivateIpDetailsLifetimeEphemeral,
	"RESERVED":  CreatePrivateIpDetailsLifetimeReserved,
}

var mappingCreatePrivateIpDetailsLifetimeEnumLowerCase = map[string]CreatePrivateIpDetailsLifetimeEnum{
	"ephemeral": CreatePrivateIpDetailsLifetimeEphemeral,
	"reserved":  CreatePrivateIpDetailsLifetimeReserved,
}

// GetCreatePrivateIpDetailsLifetimeEnumValues Enumerates the set of values for CreatePrivateIpDetailsLifetimeEnum
func GetCreatePrivateIpDetailsLifetimeEnumValues() []CreatePrivateIpDetailsLifetimeEnum {
	values := make([]CreatePrivateIpDetailsLifetimeEnum, 0)
	for _, v := range mappingCreatePrivateIpDetailsLifetimeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreatePrivateIpDetailsLifetimeEnumStringValues Enumerates the set of values in String for CreatePrivateIpDetailsLifetimeEnum
func GetCreatePrivateIpDetailsLifetimeEnumStringValues() []string {
	return []string{
		"EPHEMERAL",
		"RESERVED",
	}
}

// GetMappingCreatePrivateIpDetailsLifetimeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreatePrivateIpDetailsLifetimeEnum(val string) (CreatePrivateIpDetailsLifetimeEnum, bool) {
	enum, ok := mappingCreatePrivateIpDetailsLifetimeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
