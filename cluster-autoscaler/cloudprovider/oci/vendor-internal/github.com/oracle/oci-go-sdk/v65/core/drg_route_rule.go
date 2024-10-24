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

// DrgRouteRule A DRG route rule is a mapping between a destination IP address range and a DRG attachment.
// The map is used to route matching packets. Traffic will be routed across the attachments using Equal-cost multi-path routing (ECMP)
// if there are multiple rules with identical destinations and none of the rules conflict.
type DrgRouteRule struct {

	// Represents the range of IP addresses to match against when routing traffic.
	// Potential values:
	//   * An IP address range (IPv4 or IPv6) in CIDR notation. For example: `192.168.1.0/24`
	//   or `2001:0db8:0123:45::/56`.
	//   * When you're setting up a security rule for traffic destined for a particular `Service` through
	//   a service gateway, this is the `cidrBlock` value associated with that Service. For example: `oci-phx-objectstorage`.
	Destination *string `mandatory:"true" json:"destination"`

	// The type of destination for the rule.
	// Allowed values:
	//   * `CIDR_BLOCK`: If the rule's `destination` is an IP address range in CIDR notation.
	//   * `SERVICE_CIDR_BLOCK`: If the rule's `destination` is the `cidrBlock` value for a
	//     Service (the rule is for traffic destined for a
	//     particular `Service` through a service gateway).
	DestinationType DrgRouteRuleDestinationTypeEnum `mandatory:"true" json:"destinationType"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the next hop DRG attachment responsible
	// for reaching the network destination.
	// A value of `BLACKHOLE` means traffic for this route is discarded without notification.
	NextHopDrgAttachmentId *string `mandatory:"true" json:"nextHopDrgAttachmentId"`

	// The Oracle-assigned ID of the DRG route rule.
	Id *string `mandatory:"true" json:"id"`

	// The earliest origin of a route. If a route is advertised to a DRG through an IPsec tunnel attachment,
	// and is propagated to peered DRGs via RPC attachments, the route's provenance in the peered DRGs remains `IPSEC_TUNNEL`,
	// because that is the earliest origin.
	// No routes with a provenance `IPSEC_TUNNEL` or `VIRTUAL_CIRCUIT` will be exported to IPsec tunnel or virtual circuit attachments,
	// regardless of the attachment's export distribution.
	RouteProvenance DrgRouteRuleRouteProvenanceEnum `mandatory:"true" json:"routeProvenance"`

	// You can specify static routes for the DRG route table using the API.
	// The DRG learns dynamic routes from the DRG attachments using various routing protocols.
	RouteType DrgRouteRuleRouteTypeEnum `mandatory:"false" json:"routeType,omitempty"`

	// Indicates that the route was not imported due to a conflict between route rules.
	IsConflict *bool `mandatory:"false" json:"isConflict"`

	// Indicates that if the next hop attachment does not exist, so traffic for this route is discarded without notification.
	IsBlackhole *bool `mandatory:"false" json:"isBlackhole"`

	// Additional properties for the route, computed by the service.
	Attributes *interface{} `mandatory:"false" json:"attributes"`
}

func (m DrgRouteRule) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DrgRouteRule) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingDrgRouteRuleDestinationTypeEnum(string(m.DestinationType)); !ok && m.DestinationType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DestinationType: %s. Supported values are: %s.", m.DestinationType, strings.Join(GetDrgRouteRuleDestinationTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingDrgRouteRuleRouteProvenanceEnum(string(m.RouteProvenance)); !ok && m.RouteProvenance != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RouteProvenance: %s. Supported values are: %s.", m.RouteProvenance, strings.Join(GetDrgRouteRuleRouteProvenanceEnumStringValues(), ",")))
	}

	if _, ok := GetMappingDrgRouteRuleRouteTypeEnum(string(m.RouteType)); !ok && m.RouteType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RouteType: %s. Supported values are: %s.", m.RouteType, strings.Join(GetDrgRouteRuleRouteTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// DrgRouteRuleDestinationTypeEnum Enum with underlying type: string
type DrgRouteRuleDestinationTypeEnum string

// Set of constants representing the allowable values for DrgRouteRuleDestinationTypeEnum
const (
	DrgRouteRuleDestinationTypeCidrBlock        DrgRouteRuleDestinationTypeEnum = "CIDR_BLOCK"
	DrgRouteRuleDestinationTypeServiceCidrBlock DrgRouteRuleDestinationTypeEnum = "SERVICE_CIDR_BLOCK"
)

var mappingDrgRouteRuleDestinationTypeEnum = map[string]DrgRouteRuleDestinationTypeEnum{
	"CIDR_BLOCK":         DrgRouteRuleDestinationTypeCidrBlock,
	"SERVICE_CIDR_BLOCK": DrgRouteRuleDestinationTypeServiceCidrBlock,
}

var mappingDrgRouteRuleDestinationTypeEnumLowerCase = map[string]DrgRouteRuleDestinationTypeEnum{
	"cidr_block":         DrgRouteRuleDestinationTypeCidrBlock,
	"service_cidr_block": DrgRouteRuleDestinationTypeServiceCidrBlock,
}

// GetDrgRouteRuleDestinationTypeEnumValues Enumerates the set of values for DrgRouteRuleDestinationTypeEnum
func GetDrgRouteRuleDestinationTypeEnumValues() []DrgRouteRuleDestinationTypeEnum {
	values := make([]DrgRouteRuleDestinationTypeEnum, 0)
	for _, v := range mappingDrgRouteRuleDestinationTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgRouteRuleDestinationTypeEnumStringValues Enumerates the set of values in String for DrgRouteRuleDestinationTypeEnum
func GetDrgRouteRuleDestinationTypeEnumStringValues() []string {
	return []string{
		"CIDR_BLOCK",
		"SERVICE_CIDR_BLOCK",
	}
}

// GetMappingDrgRouteRuleDestinationTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgRouteRuleDestinationTypeEnum(val string) (DrgRouteRuleDestinationTypeEnum, bool) {
	enum, ok := mappingDrgRouteRuleDestinationTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// DrgRouteRuleRouteTypeEnum Enum with underlying type: string
type DrgRouteRuleRouteTypeEnum string

// Set of constants representing the allowable values for DrgRouteRuleRouteTypeEnum
const (
	DrgRouteRuleRouteTypeStatic  DrgRouteRuleRouteTypeEnum = "STATIC"
	DrgRouteRuleRouteTypeDynamic DrgRouteRuleRouteTypeEnum = "DYNAMIC"
)

var mappingDrgRouteRuleRouteTypeEnum = map[string]DrgRouteRuleRouteTypeEnum{
	"STATIC":  DrgRouteRuleRouteTypeStatic,
	"DYNAMIC": DrgRouteRuleRouteTypeDynamic,
}

var mappingDrgRouteRuleRouteTypeEnumLowerCase = map[string]DrgRouteRuleRouteTypeEnum{
	"static":  DrgRouteRuleRouteTypeStatic,
	"dynamic": DrgRouteRuleRouteTypeDynamic,
}

// GetDrgRouteRuleRouteTypeEnumValues Enumerates the set of values for DrgRouteRuleRouteTypeEnum
func GetDrgRouteRuleRouteTypeEnumValues() []DrgRouteRuleRouteTypeEnum {
	values := make([]DrgRouteRuleRouteTypeEnum, 0)
	for _, v := range mappingDrgRouteRuleRouteTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgRouteRuleRouteTypeEnumStringValues Enumerates the set of values in String for DrgRouteRuleRouteTypeEnum
func GetDrgRouteRuleRouteTypeEnumStringValues() []string {
	return []string{
		"STATIC",
		"DYNAMIC",
	}
}

// GetMappingDrgRouteRuleRouteTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgRouteRuleRouteTypeEnum(val string) (DrgRouteRuleRouteTypeEnum, bool) {
	enum, ok := mappingDrgRouteRuleRouteTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// DrgRouteRuleRouteProvenanceEnum Enum with underlying type: string
type DrgRouteRuleRouteProvenanceEnum string

// Set of constants representing the allowable values for DrgRouteRuleRouteProvenanceEnum
const (
	DrgRouteRuleRouteProvenanceStatic         DrgRouteRuleRouteProvenanceEnum = "STATIC"
	DrgRouteRuleRouteProvenanceVcn            DrgRouteRuleRouteProvenanceEnum = "VCN"
	DrgRouteRuleRouteProvenanceVirtualCircuit DrgRouteRuleRouteProvenanceEnum = "VIRTUAL_CIRCUIT"
	DrgRouteRuleRouteProvenanceIpsecTunnel    DrgRouteRuleRouteProvenanceEnum = "IPSEC_TUNNEL"
)

var mappingDrgRouteRuleRouteProvenanceEnum = map[string]DrgRouteRuleRouteProvenanceEnum{
	"STATIC":          DrgRouteRuleRouteProvenanceStatic,
	"VCN":             DrgRouteRuleRouteProvenanceVcn,
	"VIRTUAL_CIRCUIT": DrgRouteRuleRouteProvenanceVirtualCircuit,
	"IPSEC_TUNNEL":    DrgRouteRuleRouteProvenanceIpsecTunnel,
}

var mappingDrgRouteRuleRouteProvenanceEnumLowerCase = map[string]DrgRouteRuleRouteProvenanceEnum{
	"static":          DrgRouteRuleRouteProvenanceStatic,
	"vcn":             DrgRouteRuleRouteProvenanceVcn,
	"virtual_circuit": DrgRouteRuleRouteProvenanceVirtualCircuit,
	"ipsec_tunnel":    DrgRouteRuleRouteProvenanceIpsecTunnel,
}

// GetDrgRouteRuleRouteProvenanceEnumValues Enumerates the set of values for DrgRouteRuleRouteProvenanceEnum
func GetDrgRouteRuleRouteProvenanceEnumValues() []DrgRouteRuleRouteProvenanceEnum {
	values := make([]DrgRouteRuleRouteProvenanceEnum, 0)
	for _, v := range mappingDrgRouteRuleRouteProvenanceEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgRouteRuleRouteProvenanceEnumStringValues Enumerates the set of values in String for DrgRouteRuleRouteProvenanceEnum
func GetDrgRouteRuleRouteProvenanceEnumStringValues() []string {
	return []string{
		"STATIC",
		"VCN",
		"VIRTUAL_CIRCUIT",
		"IPSEC_TUNNEL",
	}
}

// GetMappingDrgRouteRuleRouteProvenanceEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgRouteRuleRouteProvenanceEnum(val string) (DrgRouteRuleRouteProvenanceEnum, bool) {
	enum, ok := mappingDrgRouteRuleRouteProvenanceEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
