// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
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

	// The type of destination for the rule. the type is required if `direction` = `EGRESS`.
	// Allowed values:
	//   * `CIDR_BLOCK`: If the rule's `destination` is an IP address range in CIDR notation.
	//   * `SERVICE_CIDR_BLOCK`: If the rule's `destination` is the `cidrBlock` value for a
	//     Service (the rule is for traffic destined for a
	//     particular `Service` through a service gateway).
	DestinationType DrgRouteRuleDestinationTypeEnum `mandatory:"true" json:"destinationType"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the next hop DRG attachment responsible
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
}

func (m DrgRouteRule) String() string {
	return common.PointerString(m)
}

// DrgRouteRuleDestinationTypeEnum Enum with underlying type: string
type DrgRouteRuleDestinationTypeEnum string

// Set of constants representing the allowable values for DrgRouteRuleDestinationTypeEnum
const (
	DrgRouteRuleDestinationTypeCidrBlock        DrgRouteRuleDestinationTypeEnum = "CIDR_BLOCK"
	DrgRouteRuleDestinationTypeServiceCidrBlock DrgRouteRuleDestinationTypeEnum = "SERVICE_CIDR_BLOCK"
)

var mappingDrgRouteRuleDestinationType = map[string]DrgRouteRuleDestinationTypeEnum{
	"CIDR_BLOCK":         DrgRouteRuleDestinationTypeCidrBlock,
	"SERVICE_CIDR_BLOCK": DrgRouteRuleDestinationTypeServiceCidrBlock,
}

// GetDrgRouteRuleDestinationTypeEnumValues Enumerates the set of values for DrgRouteRuleDestinationTypeEnum
func GetDrgRouteRuleDestinationTypeEnumValues() []DrgRouteRuleDestinationTypeEnum {
	values := make([]DrgRouteRuleDestinationTypeEnum, 0)
	for _, v := range mappingDrgRouteRuleDestinationType {
		values = append(values, v)
	}
	return values
}

// DrgRouteRuleRouteTypeEnum Enum with underlying type: string
type DrgRouteRuleRouteTypeEnum string

// Set of constants representing the allowable values for DrgRouteRuleRouteTypeEnum
const (
	DrgRouteRuleRouteTypeStatic  DrgRouteRuleRouteTypeEnum = "STATIC"
	DrgRouteRuleRouteTypeDynamic DrgRouteRuleRouteTypeEnum = "DYNAMIC"
)

var mappingDrgRouteRuleRouteType = map[string]DrgRouteRuleRouteTypeEnum{
	"STATIC":  DrgRouteRuleRouteTypeStatic,
	"DYNAMIC": DrgRouteRuleRouteTypeDynamic,
}

// GetDrgRouteRuleRouteTypeEnumValues Enumerates the set of values for DrgRouteRuleRouteTypeEnum
func GetDrgRouteRuleRouteTypeEnumValues() []DrgRouteRuleRouteTypeEnum {
	values := make([]DrgRouteRuleRouteTypeEnum, 0)
	for _, v := range mappingDrgRouteRuleRouteType {
		values = append(values, v)
	}
	return values
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

var mappingDrgRouteRuleRouteProvenance = map[string]DrgRouteRuleRouteProvenanceEnum{
	"STATIC":          DrgRouteRuleRouteProvenanceStatic,
	"VCN":             DrgRouteRuleRouteProvenanceVcn,
	"VIRTUAL_CIRCUIT": DrgRouteRuleRouteProvenanceVirtualCircuit,
	"IPSEC_TUNNEL":    DrgRouteRuleRouteProvenanceIpsecTunnel,
}

// GetDrgRouteRuleRouteProvenanceEnumValues Enumerates the set of values for DrgRouteRuleRouteProvenanceEnum
func GetDrgRouteRuleRouteProvenanceEnumValues() []DrgRouteRuleRouteProvenanceEnum {
	values := make([]DrgRouteRuleRouteProvenanceEnum, 0)
	for _, v := range mappingDrgRouteRuleRouteProvenance {
		values = append(values, v)
	}
	return values
}
