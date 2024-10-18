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

// RouteRule A mapping between a destination IP address range and a virtual device to route matching
// packets to (a target).
type RouteRule struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) for the route rule's target. For information about the type of
	// targets you can specify, see
	// Route Tables (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm).
	NetworkEntityId *string `mandatory:"true" json:"networkEntityId"`

	// Deprecated. Instead use `destination` and `destinationType`. Requests that include both
	// `cidrBlock` and `destination` will be rejected.
	// A destination IP address range in CIDR notation. Matching packets will
	// be routed to the indicated network entity (the target).
	// Cannot be an IPv6 prefix.
	// Example: `0.0.0.0/0`
	CidrBlock *string `mandatory:"false" json:"cidrBlock"`

	// Conceptually, this is the range of IP addresses used for matching when routing
	// traffic. Required if you provide a `destinationType`.
	// Allowed values:
	//   * IP address range in CIDR notation. Can be an IPv4 CIDR block or IPv6 prefix. For example: `192.168.1.0/24`
	//   or `2001:0db8:0123:45::/56`. If you set this to an IPv6 prefix, the route rule's target
	//   can only be a DRG or internet gateway.
	//   IPv6 addressing is supported for all commercial and government regions.
	//   See IPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	//   * The `cidrBlock` value for a Service, if you're
	//     setting up a route rule for traffic destined for a particular `Service` through
	//     a service gateway. For example: `oci-phx-objectstorage`.
	Destination *string `mandatory:"false" json:"destination"`

	// Type of destination for the rule. Required if you provide a `destination`.
	//   * `CIDR_BLOCK`: If the rule's `destination` is an IP address range in CIDR notation.
	//   * `SERVICE_CIDR_BLOCK`: If the rule's `destination` is the `cidrBlock` value for a
	//     Service (the rule is for traffic destined for a
	//     particular `Service` through a service gateway).
	DestinationType RouteRuleDestinationTypeEnum `mandatory:"false" json:"destinationType,omitempty"`

	// An optional description of your choice for the rule.
	Description *string `mandatory:"false" json:"description"`

	// A route rule can be STATIC if manually added to the route table, LOCAL if added by OCI to the route table.
	RouteType RouteRuleRouteTypeEnum `mandatory:"false" json:"routeType,omitempty"`
}

func (m RouteRule) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m RouteRule) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingRouteRuleDestinationTypeEnum(string(m.DestinationType)); !ok && m.DestinationType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DestinationType: %s. Supported values are: %s.", m.DestinationType, strings.Join(GetRouteRuleDestinationTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingRouteRuleRouteTypeEnum(string(m.RouteType)); !ok && m.RouteType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RouteType: %s. Supported values are: %s.", m.RouteType, strings.Join(GetRouteRuleRouteTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// RouteRuleDestinationTypeEnum Enum with underlying type: string
type RouteRuleDestinationTypeEnum string

// Set of constants representing the allowable values for RouteRuleDestinationTypeEnum
const (
	RouteRuleDestinationTypeCidrBlock        RouteRuleDestinationTypeEnum = "CIDR_BLOCK"
	RouteRuleDestinationTypeServiceCidrBlock RouteRuleDestinationTypeEnum = "SERVICE_CIDR_BLOCK"
)

var mappingRouteRuleDestinationTypeEnum = map[string]RouteRuleDestinationTypeEnum{
	"CIDR_BLOCK":         RouteRuleDestinationTypeCidrBlock,
	"SERVICE_CIDR_BLOCK": RouteRuleDestinationTypeServiceCidrBlock,
}

var mappingRouteRuleDestinationTypeEnumLowerCase = map[string]RouteRuleDestinationTypeEnum{
	"cidr_block":         RouteRuleDestinationTypeCidrBlock,
	"service_cidr_block": RouteRuleDestinationTypeServiceCidrBlock,
}

// GetRouteRuleDestinationTypeEnumValues Enumerates the set of values for RouteRuleDestinationTypeEnum
func GetRouteRuleDestinationTypeEnumValues() []RouteRuleDestinationTypeEnum {
	values := make([]RouteRuleDestinationTypeEnum, 0)
	for _, v := range mappingRouteRuleDestinationTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetRouteRuleDestinationTypeEnumStringValues Enumerates the set of values in String for RouteRuleDestinationTypeEnum
func GetRouteRuleDestinationTypeEnumStringValues() []string {
	return []string{
		"CIDR_BLOCK",
		"SERVICE_CIDR_BLOCK",
	}
}

// GetMappingRouteRuleDestinationTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingRouteRuleDestinationTypeEnum(val string) (RouteRuleDestinationTypeEnum, bool) {
	enum, ok := mappingRouteRuleDestinationTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// RouteRuleRouteTypeEnum Enum with underlying type: string
type RouteRuleRouteTypeEnum string

// Set of constants representing the allowable values for RouteRuleRouteTypeEnum
const (
	RouteRuleRouteTypeStatic RouteRuleRouteTypeEnum = "STATIC"
	RouteRuleRouteTypeLocal  RouteRuleRouteTypeEnum = "LOCAL"
)

var mappingRouteRuleRouteTypeEnum = map[string]RouteRuleRouteTypeEnum{
	"STATIC": RouteRuleRouteTypeStatic,
	"LOCAL":  RouteRuleRouteTypeLocal,
}

var mappingRouteRuleRouteTypeEnumLowerCase = map[string]RouteRuleRouteTypeEnum{
	"static": RouteRuleRouteTypeStatic,
	"local":  RouteRuleRouteTypeLocal,
}

// GetRouteRuleRouteTypeEnumValues Enumerates the set of values for RouteRuleRouteTypeEnum
func GetRouteRuleRouteTypeEnumValues() []RouteRuleRouteTypeEnum {
	values := make([]RouteRuleRouteTypeEnum, 0)
	for _, v := range mappingRouteRuleRouteTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetRouteRuleRouteTypeEnumStringValues Enumerates the set of values in String for RouteRuleRouteTypeEnum
func GetRouteRuleRouteTypeEnumStringValues() []string {
	return []string{
		"STATIC",
		"LOCAL",
	}
}

// GetMappingRouteRuleRouteTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingRouteRuleRouteTypeEnum(val string) (RouteRuleRouteTypeEnum, bool) {
	enum, ok := mappingRouteRuleRouteTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
