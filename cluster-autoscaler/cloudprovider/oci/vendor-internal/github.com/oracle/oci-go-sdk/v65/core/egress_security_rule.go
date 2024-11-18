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

// EgressSecurityRule A rule for allowing outbound IP packets.
type EgressSecurityRule struct {

	// Conceptually, this is the range of IP addresses that a packet originating from the instance
	// can go to.
	// Allowed values:
	//   * IP address range in CIDR notation. For example: `192.168.1.0/24` or `2001:0db8:0123:45::/56`
	//     Note that IPv6 addressing is currently supported only in certain regions. See
	//     IPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	//   * The `cidrBlock` value for a Service, if you're
	//     setting up a security list rule for traffic destined for a particular `Service` through
	//     a service gateway. For example: `oci-phx-objectstorage`.
	Destination *string `mandatory:"true" json:"destination"`

	// The transport protocol. Specify either `all` or an IPv4 protocol number as
	// defined in
	// Protocol Numbers (http://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml).
	// Options are supported only for ICMP ("1"), TCP ("6"), UDP ("17"), and ICMPv6 ("58").
	Protocol *string `mandatory:"true" json:"protocol"`

	// Type of destination for the rule. The default is `CIDR_BLOCK`.
	// Allowed values:
	//   * `CIDR_BLOCK`: If the rule's `destination` is an IP address range in CIDR notation.
	//   * `SERVICE_CIDR_BLOCK`: If the rule's `destination` is the `cidrBlock` value for a
	//     Service (the rule is for traffic destined for a
	//     particular `Service` through a service gateway).
	DestinationType EgressSecurityRuleDestinationTypeEnum `mandatory:"false" json:"destinationType,omitempty"`

	IcmpOptions *IcmpOptions `mandatory:"false" json:"icmpOptions"`

	// A stateless rule allows traffic in one direction. Remember to add a corresponding
	// stateless rule in the other direction if you need to support bidirectional traffic. For
	// example, if egress traffic allows TCP destination port 80, there should be an ingress
	// rule to allow TCP source port 80. Defaults to false, which means the rule is stateful
	// and a corresponding rule is not necessary for bidirectional traffic.
	IsStateless *bool `mandatory:"false" json:"isStateless"`

	TcpOptions *TcpOptions `mandatory:"false" json:"tcpOptions"`

	UdpOptions *UdpOptions `mandatory:"false" json:"udpOptions"`

	// An optional description of your choice for the rule.
	Description *string `mandatory:"false" json:"description"`
}

func (m EgressSecurityRule) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m EgressSecurityRule) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingEgressSecurityRuleDestinationTypeEnum(string(m.DestinationType)); !ok && m.DestinationType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DestinationType: %s. Supported values are: %s.", m.DestinationType, strings.Join(GetEgressSecurityRuleDestinationTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// EgressSecurityRuleDestinationTypeEnum Enum with underlying type: string
type EgressSecurityRuleDestinationTypeEnum string

// Set of constants representing the allowable values for EgressSecurityRuleDestinationTypeEnum
const (
	EgressSecurityRuleDestinationTypeCidrBlock        EgressSecurityRuleDestinationTypeEnum = "CIDR_BLOCK"
	EgressSecurityRuleDestinationTypeServiceCidrBlock EgressSecurityRuleDestinationTypeEnum = "SERVICE_CIDR_BLOCK"
)

var mappingEgressSecurityRuleDestinationTypeEnum = map[string]EgressSecurityRuleDestinationTypeEnum{
	"CIDR_BLOCK":         EgressSecurityRuleDestinationTypeCidrBlock,
	"SERVICE_CIDR_BLOCK": EgressSecurityRuleDestinationTypeServiceCidrBlock,
}

var mappingEgressSecurityRuleDestinationTypeEnumLowerCase = map[string]EgressSecurityRuleDestinationTypeEnum{
	"cidr_block":         EgressSecurityRuleDestinationTypeCidrBlock,
	"service_cidr_block": EgressSecurityRuleDestinationTypeServiceCidrBlock,
}

// GetEgressSecurityRuleDestinationTypeEnumValues Enumerates the set of values for EgressSecurityRuleDestinationTypeEnum
func GetEgressSecurityRuleDestinationTypeEnumValues() []EgressSecurityRuleDestinationTypeEnum {
	values := make([]EgressSecurityRuleDestinationTypeEnum, 0)
	for _, v := range mappingEgressSecurityRuleDestinationTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetEgressSecurityRuleDestinationTypeEnumStringValues Enumerates the set of values in String for EgressSecurityRuleDestinationTypeEnum
func GetEgressSecurityRuleDestinationTypeEnumStringValues() []string {
	return []string{
		"CIDR_BLOCK",
		"SERVICE_CIDR_BLOCK",
	}
}

// GetMappingEgressSecurityRuleDestinationTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingEgressSecurityRuleDestinationTypeEnum(val string) (EgressSecurityRuleDestinationTypeEnum, bool) {
	enum, ok := mappingEgressSecurityRuleDestinationTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
