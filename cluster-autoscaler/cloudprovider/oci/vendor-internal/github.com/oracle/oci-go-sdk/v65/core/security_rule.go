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

// SecurityRule A security rule is one of the items in a NetworkSecurityGroup.
// It is a virtual firewall rule for the VNICs in the network security group. A rule can be for
// either inbound (`direction`= INGRESS) or outbound (`direction`= EGRESS) IP packets.
type SecurityRule struct {

	// Direction of the security rule. Set to `EGRESS` for rules to allow outbound IP packets,
	// or `INGRESS` for rules to allow inbound IP packets.
	Direction SecurityRuleDirectionEnum `mandatory:"true" json:"direction"`

	// The transport protocol. Specify either `all` or an IPv4 protocol number as
	// defined in
	// Protocol Numbers (http://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml).
	// Options are supported only for ICMP ("1"), TCP ("6"), UDP ("17"), and ICMPv6 ("58").
	Protocol *string `mandatory:"true" json:"protocol"`

	// An optional description of your choice for the rule.
	Description *string `mandatory:"false" json:"description"`

	// Conceptually, this is the range of IP addresses that a packet originating from the instance
	// can go to.
	// Allowed values:
	//   * An IP address range in CIDR notation. For example: `192.168.1.0/24` or `2001:0db8:0123:45::/56`
	//     IPv6 addressing is supported for all commercial and government regions.
	//     See IPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	//   * The `cidrBlock` value for a Service, if you're
	//     setting up a security rule for traffic destined for a particular `Service` through
	//     a service gateway. For example: `oci-phx-objectstorage`.
	//   * The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of a NetworkSecurityGroup in the same
	//     VCN. The value can be the NSG that the rule belongs to if the rule's intent is to control
	//     traffic between VNICs in the same NSG.
	Destination *string `mandatory:"false" json:"destination"`

	// Type of destination for the rule. Required if `direction` = `EGRESS`.
	// Allowed values:
	//   * `CIDR_BLOCK`: If the rule's `destination` is an IP address range in CIDR notation.
	//   * `SERVICE_CIDR_BLOCK`: If the rule's `destination` is the `cidrBlock` value for a
	//     Service (the rule is for traffic destined for a
	//     particular `Service` through a service gateway).
	//   * `NETWORK_SECURITY_GROUP`: If the rule's `destination` is the OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of a
	//     NetworkSecurityGroup.
	DestinationType SecurityRuleDestinationTypeEnum `mandatory:"false" json:"destinationType,omitempty"`

	IcmpOptions *IcmpOptions `mandatory:"false" json:"icmpOptions"`

	// An Oracle-assigned identifier for the security rule. You specify this ID when you want to
	// update or delete the rule.
	// Example: `04ABEC`
	Id *string `mandatory:"false" json:"id"`

	// A stateless rule allows traffic in one direction. Remember to add a corresponding
	// stateless rule in the other direction if you need to support bidirectional traffic. For
	// example, if egress traffic allows TCP destination port 80, there should be an ingress
	// rule to allow TCP source port 80. Defaults to false, which means the rule is stateful
	// and a corresponding rule is not necessary for bidirectional traffic.
	IsStateless *bool `mandatory:"false" json:"isStateless"`

	// Whether the rule is valid. The value is `True` when the rule is first created. If
	// the rule's `source` or `destination` is a network security group, the value changes to
	// `False` if that network security group is deleted.
	IsValid *bool `mandatory:"false" json:"isValid"`

	// Conceptually, this is the range of IP addresses that a packet coming into the instance
	// can come from.
	// Allowed values:
	//   * An IP address range in CIDR notation. For example: `192.168.1.0/24` or `2001:0db8:0123:45::/56`
	//     IPv6 addressing is supported for all commercial and government regions.
	//     See IPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	//   * The `cidrBlock` value for a Service, if you're
	//     setting up a security rule for traffic coming from a particular `Service` through
	//     a service gateway. For example: `oci-phx-objectstorage`.
	//   * The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of a NetworkSecurityGroup in the same
	//     VCN. The value can be the NSG that the rule belongs to if the rule's intent is to control
	//     traffic between VNICs in the same NSG.
	Source *string `mandatory:"false" json:"source"`

	// Type of source for the rule. Required if `direction` = `INGRESS`.
	//   * `CIDR_BLOCK`: If the rule's `source` is an IP address range in CIDR notation.
	//   * `SERVICE_CIDR_BLOCK`: If the rule's `source` is the `cidrBlock` value for a
	//     Service (the rule is for traffic coming from a
	//     particular `Service` through a service gateway).
	//   * `NETWORK_SECURITY_GROUP`: If the rule's `source` is the OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of a
	//     NetworkSecurityGroup.
	SourceType SecurityRuleSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`

	TcpOptions *TcpOptions `mandatory:"false" json:"tcpOptions"`

	// The date and time the security rule was created. Format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	UdpOptions *UdpOptions `mandatory:"false" json:"udpOptions"`
}

func (m SecurityRule) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m SecurityRule) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingSecurityRuleDirectionEnum(string(m.Direction)); !ok && m.Direction != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Direction: %s. Supported values are: %s.", m.Direction, strings.Join(GetSecurityRuleDirectionEnumStringValues(), ",")))
	}

	if _, ok := GetMappingSecurityRuleDestinationTypeEnum(string(m.DestinationType)); !ok && m.DestinationType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DestinationType: %s. Supported values are: %s.", m.DestinationType, strings.Join(GetSecurityRuleDestinationTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingSecurityRuleSourceTypeEnum(string(m.SourceType)); !ok && m.SourceType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SourceType: %s. Supported values are: %s.", m.SourceType, strings.Join(GetSecurityRuleSourceTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// SecurityRuleDestinationTypeEnum Enum with underlying type: string
type SecurityRuleDestinationTypeEnum string

// Set of constants representing the allowable values for SecurityRuleDestinationTypeEnum
const (
	SecurityRuleDestinationTypeCidrBlock            SecurityRuleDestinationTypeEnum = "CIDR_BLOCK"
	SecurityRuleDestinationTypeServiceCidrBlock     SecurityRuleDestinationTypeEnum = "SERVICE_CIDR_BLOCK"
	SecurityRuleDestinationTypeNetworkSecurityGroup SecurityRuleDestinationTypeEnum = "NETWORK_SECURITY_GROUP"
)

var mappingSecurityRuleDestinationTypeEnum = map[string]SecurityRuleDestinationTypeEnum{
	"CIDR_BLOCK":             SecurityRuleDestinationTypeCidrBlock,
	"SERVICE_CIDR_BLOCK":     SecurityRuleDestinationTypeServiceCidrBlock,
	"NETWORK_SECURITY_GROUP": SecurityRuleDestinationTypeNetworkSecurityGroup,
}

var mappingSecurityRuleDestinationTypeEnumLowerCase = map[string]SecurityRuleDestinationTypeEnum{
	"cidr_block":             SecurityRuleDestinationTypeCidrBlock,
	"service_cidr_block":     SecurityRuleDestinationTypeServiceCidrBlock,
	"network_security_group": SecurityRuleDestinationTypeNetworkSecurityGroup,
}

// GetSecurityRuleDestinationTypeEnumValues Enumerates the set of values for SecurityRuleDestinationTypeEnum
func GetSecurityRuleDestinationTypeEnumValues() []SecurityRuleDestinationTypeEnum {
	values := make([]SecurityRuleDestinationTypeEnum, 0)
	for _, v := range mappingSecurityRuleDestinationTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetSecurityRuleDestinationTypeEnumStringValues Enumerates the set of values in String for SecurityRuleDestinationTypeEnum
func GetSecurityRuleDestinationTypeEnumStringValues() []string {
	return []string{
		"CIDR_BLOCK",
		"SERVICE_CIDR_BLOCK",
		"NETWORK_SECURITY_GROUP",
	}
}

// GetMappingSecurityRuleDestinationTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingSecurityRuleDestinationTypeEnum(val string) (SecurityRuleDestinationTypeEnum, bool) {
	enum, ok := mappingSecurityRuleDestinationTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// SecurityRuleDirectionEnum Enum with underlying type: string
type SecurityRuleDirectionEnum string

// Set of constants representing the allowable values for SecurityRuleDirectionEnum
const (
	SecurityRuleDirectionEgress  SecurityRuleDirectionEnum = "EGRESS"
	SecurityRuleDirectionIngress SecurityRuleDirectionEnum = "INGRESS"
)

var mappingSecurityRuleDirectionEnum = map[string]SecurityRuleDirectionEnum{
	"EGRESS":  SecurityRuleDirectionEgress,
	"INGRESS": SecurityRuleDirectionIngress,
}

var mappingSecurityRuleDirectionEnumLowerCase = map[string]SecurityRuleDirectionEnum{
	"egress":  SecurityRuleDirectionEgress,
	"ingress": SecurityRuleDirectionIngress,
}

// GetSecurityRuleDirectionEnumValues Enumerates the set of values for SecurityRuleDirectionEnum
func GetSecurityRuleDirectionEnumValues() []SecurityRuleDirectionEnum {
	values := make([]SecurityRuleDirectionEnum, 0)
	for _, v := range mappingSecurityRuleDirectionEnum {
		values = append(values, v)
	}
	return values
}

// GetSecurityRuleDirectionEnumStringValues Enumerates the set of values in String for SecurityRuleDirectionEnum
func GetSecurityRuleDirectionEnumStringValues() []string {
	return []string{
		"EGRESS",
		"INGRESS",
	}
}

// GetMappingSecurityRuleDirectionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingSecurityRuleDirectionEnum(val string) (SecurityRuleDirectionEnum, bool) {
	enum, ok := mappingSecurityRuleDirectionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// SecurityRuleSourceTypeEnum Enum with underlying type: string
type SecurityRuleSourceTypeEnum string

// Set of constants representing the allowable values for SecurityRuleSourceTypeEnum
const (
	SecurityRuleSourceTypeCidrBlock            SecurityRuleSourceTypeEnum = "CIDR_BLOCK"
	SecurityRuleSourceTypeServiceCidrBlock     SecurityRuleSourceTypeEnum = "SERVICE_CIDR_BLOCK"
	SecurityRuleSourceTypeNetworkSecurityGroup SecurityRuleSourceTypeEnum = "NETWORK_SECURITY_GROUP"
)

var mappingSecurityRuleSourceTypeEnum = map[string]SecurityRuleSourceTypeEnum{
	"CIDR_BLOCK":             SecurityRuleSourceTypeCidrBlock,
	"SERVICE_CIDR_BLOCK":     SecurityRuleSourceTypeServiceCidrBlock,
	"NETWORK_SECURITY_GROUP": SecurityRuleSourceTypeNetworkSecurityGroup,
}

var mappingSecurityRuleSourceTypeEnumLowerCase = map[string]SecurityRuleSourceTypeEnum{
	"cidr_block":             SecurityRuleSourceTypeCidrBlock,
	"service_cidr_block":     SecurityRuleSourceTypeServiceCidrBlock,
	"network_security_group": SecurityRuleSourceTypeNetworkSecurityGroup,
}

// GetSecurityRuleSourceTypeEnumValues Enumerates the set of values for SecurityRuleSourceTypeEnum
func GetSecurityRuleSourceTypeEnumValues() []SecurityRuleSourceTypeEnum {
	values := make([]SecurityRuleSourceTypeEnum, 0)
	for _, v := range mappingSecurityRuleSourceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetSecurityRuleSourceTypeEnumStringValues Enumerates the set of values in String for SecurityRuleSourceTypeEnum
func GetSecurityRuleSourceTypeEnumStringValues() []string {
	return []string{
		"CIDR_BLOCK",
		"SERVICE_CIDR_BLOCK",
		"NETWORK_SECURITY_GROUP",
	}
}

// GetMappingSecurityRuleSourceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingSecurityRuleSourceTypeEnum(val string) (SecurityRuleSourceTypeEnum, bool) {
	enum, ok := mappingSecurityRuleSourceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
