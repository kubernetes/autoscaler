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

// CreateSubnetDetails The representation of CreateSubnetDetails
type CreateSubnetDetails struct {

	// The CIDR IP address range of the subnet. The CIDR must maintain the following rules -
	// a. The CIDR block is valid and correctly formatted.
	// b. The new range is within one of the parent VCN ranges.
	// Example: `10.0.1.0/24`
	CidrBlock *string `mandatory:"true" json:"cidrBlock"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment to contain the subnet.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN to contain the subnet.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// Controls whether the subnet is regional or specific to an availability domain. Oracle
	// recommends creating regional subnets because they're more flexible and make it easier to
	// implement failover across availability domains. Originally, AD-specific subnets were the
	// only kind available to use.
	// To create a regional subnet, omit this attribute. Then any resources later created in this
	// subnet (such as a Compute instance) can be created in any availability domain in the region.
	// To instead create an AD-specific subnet, set this attribute to the availability domain you
	// want this subnet to be in. Then any resources later created in this subnet can only be
	// created in that availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the set of DHCP options the subnet will use. If you don't
	// provide a value, the subnet uses the VCN's default set of DHCP options.
	DhcpOptionsId *string `mandatory:"false" json:"dhcpOptionsId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// A DNS label for the subnet, used in conjunction with the VNIC's hostname and
	// VCN's DNS label to form a fully qualified domain name (FQDN) for each VNIC
	// within this subnet (for example, `bminstance1.subnet123.vcn1.oraclevcn.com`).
	// Must be an alphanumeric string that begins with a letter and is unique within the VCN.
	// The value cannot be changed.
	// This value must be set if you want to use the Internet and VCN Resolver to resolve the
	// hostnames of instances in the subnet. It can only be set if the VCN itself
	// was created with a DNS label.
	// For more information, see
	// DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm).
	// Example: `subnet123`
	DnsLabel *string `mandatory:"false" json:"dnsLabel"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Use this to enable IPv6 addressing for this subnet. The VCN must be enabled for IPv6.
	// You can't change this subnet characteristic later. All subnets are /64 in size. The subnet
	// portion of the IPv6 address is the fourth hextet from the left (1111 in the following example).
	// For important details about IPv6 addressing in a VCN, see IPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	// Example: `2001:0db8:0123:1111::/64`
	Ipv6CidrBlock *string `mandatory:"false" json:"ipv6CidrBlock"`

	// The list of all IPv6 prefixes (Oracle allocated IPv6 GUA, ULA or private IPv6 prefixes, BYOIPv6 prefixes) for the subnet that meets the following criteria:
	// - The prefixes must be valid.
	// - Multiple prefixes must not overlap each other or the on-premises network prefix.
	// - The number of prefixes must not exceed the limit of IPv6 prefixes allowed to a subnet.
	Ipv6CidrBlocks []string `mandatory:"false" json:"ipv6CidrBlocks"`

	// Whether to disallow ingress internet traffic to VNICs within this subnet. Defaults to false.
	// For IPv6, if `prohibitInternetIngress` is set to `true`, internet access is not allowed for any
	// IPv6s assigned to VNICs in the subnet. Otherwise, ingress internet traffic is allowed by default.
	// `prohibitPublicIpOnVnic` will be set to the value of `prohibitInternetIngress` to dictate IPv4
	// behavior in this subnet. Only one or the other flag should be specified.
	// Example: `true`
	ProhibitInternetIngress *bool `mandatory:"false" json:"prohibitInternetIngress"`

	// Whether VNICs within this subnet can have public IP addresses.
	// Defaults to false, which means VNICs created in this subnet will
	// automatically be assigned public IP addresses unless specified
	// otherwise during instance launch or VNIC creation (with the
	// `assignPublicIp` flag in CreateVnicDetails).
	// If `prohibitPublicIpOnVnic` is set to true, VNICs created in this
	// subnet cannot have public IP addresses (that is, it's a private
	// subnet).
	// If you intend to use an IPv6 prefix, you should use the flag `prohibitInternetIngress` to
	// specify ingress internet traffic behavior of the subnet.
	// Example: `true`
	ProhibitPublicIpOnVnic *bool `mandatory:"false" json:"prohibitPublicIpOnVnic"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the route table the subnet will use. If you don't provide a value,
	// the subnet uses the VCN's default route table.
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// The OCIDs of the security list or lists the subnet will use. If you don't
	// provide a value, the subnet uses the VCN's default security list.
	// Remember that security lists are associated *with the subnet*, but the
	// rules are applied to the individual VNICs in the subnet.
	SecurityListIds []string `mandatory:"false" json:"securityListIds"`
}

func (m CreateSubnetDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateSubnetDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
