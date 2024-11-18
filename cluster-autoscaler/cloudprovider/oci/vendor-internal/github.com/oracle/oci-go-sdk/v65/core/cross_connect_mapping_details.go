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

// CrossConnectMappingDetails For use with Oracle Cloud Infrastructure FastConnect. Each
// VirtualCircuit runs on one or
// more cross-connects or cross-connect groups. A `CrossConnectMappingDetails`
// contains the properties for an individual cross-connect or cross-connect group
// associated with a given virtual circuit.
// The details includes information about the cross-connect or
// cross-connect group, the VLAN, and the BGP peering session.
type CrossConnectMappingDetails struct {

	// The key for BGP MD5 authentication. Only applicable if your system
	// requires MD5 authentication. If empty or not set (null), that
	// means you don't use BGP MD5 authentication.
	BgpMd5AuthKey *string `mandatory:"false" json:"bgpMd5AuthKey"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the cross-connect or cross-connect group for this mapping.
	// Specified by the owner of the cross-connect or cross-connect group (the
	// customer if the customer is colocated with Oracle, or the provider if the
	// customer is connecting via provider).
	CrossConnectOrCrossConnectGroupId *string `mandatory:"false" json:"crossConnectOrCrossConnectGroupId"`

	// The BGP IPv4 address for the router on the other end of the BGP session from
	// Oracle. Specified by the owner of that router. If the session goes from Oracle
	// to a customer, this is the BGP IPv4 address of the customer's edge router. If the
	// session goes from Oracle to a provider, this is the BGP IPv4 address of the
	// provider's edge router. Must use a subnet mask from /28 to /31.
	// There's one exception: for a public virtual circuit, Oracle specifies the BGP IPv4 addresses.
	// Example: `10.0.0.18/31`
	CustomerBgpPeeringIp *string `mandatory:"false" json:"customerBgpPeeringIp"`

	// The IPv4 address for Oracle's end of the BGP session. Must use a subnet mask from /28 to /31.
	// If the session goes from Oracle to a customer's edge router,
	// the customer specifies this information. If the session goes from Oracle to
	// a provider's edge router, the provider specifies this.
	// There's one exception: for a public virtual circuit, Oracle specifies the BGP IPv4 addresses.
	// Example: `10.0.0.19/31`
	OracleBgpPeeringIp *string `mandatory:"false" json:"oracleBgpPeeringIp"`

	// The BGP IPv6 address for the router on the other end of the BGP session from
	// Oracle. Specified by the owner of that router. If the session goes from Oracle
	// to a customer, this is the BGP IPv6 address of the customer's edge router. If the
	// session goes from Oracle to a provider, this is the BGP IPv6 address of the
	// provider's edge router. Only subnet masks from /64 up to /127 are allowed.
	// There's one exception: for a public virtual circuit, Oracle specifies the BGP IPv6 addresses.
	// Example: `2001:db8::1/64`
	CustomerBgpPeeringIpv6 *string `mandatory:"false" json:"customerBgpPeeringIpv6"`

	// The IPv6 address for Oracle's end of the BGP session. Only subnet masks from /64 up to /127 are allowed.
	// If the session goes from Oracle to a customer's edge router,
	// the customer specifies this information. If the session goes from Oracle to
	// a provider's edge router, the provider specifies this.
	// There's one exception: for a public virtual circuit, Oracle specifies the BGP IPv6 addresses.
	// Example: `2001:db8::2/64`
	OracleBgpPeeringIpv6 *string `mandatory:"false" json:"oracleBgpPeeringIpv6"`

	// The number of the specific VLAN (on the cross-connect or cross-connect group)
	// that is assigned to this virtual circuit. Specified by the owner of the cross-connect
	// or cross-connect group (the customer if the customer is colocated with Oracle, or
	// the provider if the customer is connecting via provider).
	// Example: `200`
	Vlan *int `mandatory:"false" json:"vlan"`

	// The state of the Ipv4 BGP session.
	Ipv4BgpStatus CrossConnectMappingDetailsIpv4BgpStatusEnum `mandatory:"false" json:"ipv4BgpStatus,omitempty"`

	// The state of the Ipv6 BGP session.
	Ipv6BgpStatus CrossConnectMappingDetailsIpv6BgpStatusEnum `mandatory:"false" json:"ipv6BgpStatus,omitempty"`

	// The FastConnect device that terminates the logical connection.
	OciLogicalDeviceName *string `mandatory:"false" json:"ociLogicalDeviceName"`
}

func (m CrossConnectMappingDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CrossConnectMappingDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCrossConnectMappingDetailsIpv4BgpStatusEnum(string(m.Ipv4BgpStatus)); !ok && m.Ipv4BgpStatus != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Ipv4BgpStatus: %s. Supported values are: %s.", m.Ipv4BgpStatus, strings.Join(GetCrossConnectMappingDetailsIpv4BgpStatusEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCrossConnectMappingDetailsIpv6BgpStatusEnum(string(m.Ipv6BgpStatus)); !ok && m.Ipv6BgpStatus != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Ipv6BgpStatus: %s. Supported values are: %s.", m.Ipv6BgpStatus, strings.Join(GetCrossConnectMappingDetailsIpv6BgpStatusEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CrossConnectMappingDetailsIpv4BgpStatusEnum Enum with underlying type: string
type CrossConnectMappingDetailsIpv4BgpStatusEnum string

// Set of constants representing the allowable values for CrossConnectMappingDetailsIpv4BgpStatusEnum
const (
	CrossConnectMappingDetailsIpv4BgpStatusUp   CrossConnectMappingDetailsIpv4BgpStatusEnum = "UP"
	CrossConnectMappingDetailsIpv4BgpStatusDown CrossConnectMappingDetailsIpv4BgpStatusEnum = "DOWN"
)

var mappingCrossConnectMappingDetailsIpv4BgpStatusEnum = map[string]CrossConnectMappingDetailsIpv4BgpStatusEnum{
	"UP":   CrossConnectMappingDetailsIpv4BgpStatusUp,
	"DOWN": CrossConnectMappingDetailsIpv4BgpStatusDown,
}

var mappingCrossConnectMappingDetailsIpv4BgpStatusEnumLowerCase = map[string]CrossConnectMappingDetailsIpv4BgpStatusEnum{
	"up":   CrossConnectMappingDetailsIpv4BgpStatusUp,
	"down": CrossConnectMappingDetailsIpv4BgpStatusDown,
}

// GetCrossConnectMappingDetailsIpv4BgpStatusEnumValues Enumerates the set of values for CrossConnectMappingDetailsIpv4BgpStatusEnum
func GetCrossConnectMappingDetailsIpv4BgpStatusEnumValues() []CrossConnectMappingDetailsIpv4BgpStatusEnum {
	values := make([]CrossConnectMappingDetailsIpv4BgpStatusEnum, 0)
	for _, v := range mappingCrossConnectMappingDetailsIpv4BgpStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetCrossConnectMappingDetailsIpv4BgpStatusEnumStringValues Enumerates the set of values in String for CrossConnectMappingDetailsIpv4BgpStatusEnum
func GetCrossConnectMappingDetailsIpv4BgpStatusEnumStringValues() []string {
	return []string{
		"UP",
		"DOWN",
	}
}

// GetMappingCrossConnectMappingDetailsIpv4BgpStatusEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCrossConnectMappingDetailsIpv4BgpStatusEnum(val string) (CrossConnectMappingDetailsIpv4BgpStatusEnum, bool) {
	enum, ok := mappingCrossConnectMappingDetailsIpv4BgpStatusEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CrossConnectMappingDetailsIpv6BgpStatusEnum Enum with underlying type: string
type CrossConnectMappingDetailsIpv6BgpStatusEnum string

// Set of constants representing the allowable values for CrossConnectMappingDetailsIpv6BgpStatusEnum
const (
	CrossConnectMappingDetailsIpv6BgpStatusUp   CrossConnectMappingDetailsIpv6BgpStatusEnum = "UP"
	CrossConnectMappingDetailsIpv6BgpStatusDown CrossConnectMappingDetailsIpv6BgpStatusEnum = "DOWN"
)

var mappingCrossConnectMappingDetailsIpv6BgpStatusEnum = map[string]CrossConnectMappingDetailsIpv6BgpStatusEnum{
	"UP":   CrossConnectMappingDetailsIpv6BgpStatusUp,
	"DOWN": CrossConnectMappingDetailsIpv6BgpStatusDown,
}

var mappingCrossConnectMappingDetailsIpv6BgpStatusEnumLowerCase = map[string]CrossConnectMappingDetailsIpv6BgpStatusEnum{
	"up":   CrossConnectMappingDetailsIpv6BgpStatusUp,
	"down": CrossConnectMappingDetailsIpv6BgpStatusDown,
}

// GetCrossConnectMappingDetailsIpv6BgpStatusEnumValues Enumerates the set of values for CrossConnectMappingDetailsIpv6BgpStatusEnum
func GetCrossConnectMappingDetailsIpv6BgpStatusEnumValues() []CrossConnectMappingDetailsIpv6BgpStatusEnum {
	values := make([]CrossConnectMappingDetailsIpv6BgpStatusEnum, 0)
	for _, v := range mappingCrossConnectMappingDetailsIpv6BgpStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetCrossConnectMappingDetailsIpv6BgpStatusEnumStringValues Enumerates the set of values in String for CrossConnectMappingDetailsIpv6BgpStatusEnum
func GetCrossConnectMappingDetailsIpv6BgpStatusEnumStringValues() []string {
	return []string{
		"UP",
		"DOWN",
	}
}

// GetMappingCrossConnectMappingDetailsIpv6BgpStatusEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCrossConnectMappingDetailsIpv6BgpStatusEnum(val string) (CrossConnectMappingDetailsIpv6BgpStatusEnum, bool) {
	enum, ok := mappingCrossConnectMappingDetailsIpv6BgpStatusEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
