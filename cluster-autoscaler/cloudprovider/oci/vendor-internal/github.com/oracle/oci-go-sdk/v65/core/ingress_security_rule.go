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

// IngressSecurityRule A rule for allowing inbound IP packets.
type IngressSecurityRule struct {

	// The transport protocol. Specify either `all` or an IPv4 protocol number as
	// defined in
	// Protocol Numbers (http://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml).
	// Options are supported only for ICMP ("1"), TCP ("6"), UDP ("17"), and ICMPv6 ("58").
	Protocol *string `mandatory:"true" json:"protocol"`

	// Conceptually, this is the range of IP addresses that a packet coming into the instance
	// can come from.
	// Allowed values:
	//   * IP address range in CIDR notation. For example: `192.168.1.0/24` or `2001:0db8:0123:45::/56`.
	//     IPv6 addressing is supported for all commercial and government regions. See
	//     IPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	//   * The `cidrBlock` value for a Service, if you're
	//     setting up a security list rule for traffic coming from a particular `Service` through
	//     a service gateway. For example: `oci-phx-objectstorage`.
	Source *string `mandatory:"true" json:"source"`

	IcmpOptions *IcmpOptions `mandatory:"false" json:"icmpOptions"`

	// A stateless rule allows traffic in one direction. Remember to add a corresponding
	// stateless rule in the other direction if you need to support bidirectional traffic. For
	// example, if ingress traffic allows TCP destination port 80, there should be an egress
	// rule to allow TCP source port 80. Defaults to false, which means the rule is stateful
	// and a corresponding rule is not necessary for bidirectional traffic.
	IsStateless *bool `mandatory:"false" json:"isStateless"`

	// Type of source for the rule. The default is `CIDR_BLOCK`.
	//   * `CIDR_BLOCK`: If the rule's `source` is an IP address range in CIDR notation.
	//   * `SERVICE_CIDR_BLOCK`: If the rule's `source` is the `cidrBlock` value for a
	//     Service (the rule is for traffic coming from a
	//     particular `Service` through a service gateway).
	SourceType IngressSecurityRuleSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`

	TcpOptions *TcpOptions `mandatory:"false" json:"tcpOptions"`

	UdpOptions *UdpOptions `mandatory:"false" json:"udpOptions"`

	// An optional description of your choice for the rule.
	Description *string `mandatory:"false" json:"description"`
}

func (m IngressSecurityRule) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m IngressSecurityRule) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingIngressSecurityRuleSourceTypeEnum(string(m.SourceType)); !ok && m.SourceType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SourceType: %s. Supported values are: %s.", m.SourceType, strings.Join(GetIngressSecurityRuleSourceTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// IngressSecurityRuleSourceTypeEnum Enum with underlying type: string
type IngressSecurityRuleSourceTypeEnum string

// Set of constants representing the allowable values for IngressSecurityRuleSourceTypeEnum
const (
	IngressSecurityRuleSourceTypeCidrBlock        IngressSecurityRuleSourceTypeEnum = "CIDR_BLOCK"
	IngressSecurityRuleSourceTypeServiceCidrBlock IngressSecurityRuleSourceTypeEnum = "SERVICE_CIDR_BLOCK"
)

var mappingIngressSecurityRuleSourceTypeEnum = map[string]IngressSecurityRuleSourceTypeEnum{
	"CIDR_BLOCK":         IngressSecurityRuleSourceTypeCidrBlock,
	"SERVICE_CIDR_BLOCK": IngressSecurityRuleSourceTypeServiceCidrBlock,
}

var mappingIngressSecurityRuleSourceTypeEnumLowerCase = map[string]IngressSecurityRuleSourceTypeEnum{
	"cidr_block":         IngressSecurityRuleSourceTypeCidrBlock,
	"service_cidr_block": IngressSecurityRuleSourceTypeServiceCidrBlock,
}

// GetIngressSecurityRuleSourceTypeEnumValues Enumerates the set of values for IngressSecurityRuleSourceTypeEnum
func GetIngressSecurityRuleSourceTypeEnumValues() []IngressSecurityRuleSourceTypeEnum {
	values := make([]IngressSecurityRuleSourceTypeEnum, 0)
	for _, v := range mappingIngressSecurityRuleSourceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetIngressSecurityRuleSourceTypeEnumStringValues Enumerates the set of values in String for IngressSecurityRuleSourceTypeEnum
func GetIngressSecurityRuleSourceTypeEnumStringValues() []string {
	return []string{
		"CIDR_BLOCK",
		"SERVICE_CIDR_BLOCK",
	}
}

// GetMappingIngressSecurityRuleSourceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIngressSecurityRuleSourceTypeEnum(val string) (IngressSecurityRuleSourceTypeEnum, bool) {
	enum, ok := mappingIngressSecurityRuleSourceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
