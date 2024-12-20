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

// CreateVnicDetails Contains properties for a VNIC. You use this object when creating the
// primary VNIC during instance launch or when creating a secondary VNIC.
// For more information about VNICs, see
// Virtual Network Interface Cards (VNICs) (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingVNICs.htm).
type CreateVnicDetails struct {

	// Whether to allocate an IPv6 address at instance and VNIC creation from an IPv6 enabled
	// subnet. Default: False. When provided you may optionally provide an IPv6 prefix
	// (`ipv6SubnetCidr`) of your choice to assign the IPv6 address from. If `ipv6SubnetCidr`
	// is not provided then an IPv6 prefix is chosen
	// for you.
	AssignIpv6Ip *bool `mandatory:"false" json:"assignIpv6Ip"`

	// Whether the VNIC should be assigned a public IP address. Defaults to whether
	// the subnet is public or private. If not set and the VNIC is being created
	// in a private subnet (that is, where `prohibitPublicIpOnVnic` = true in the
	// Subnet), then no public IP address is assigned.
	// If not set and the subnet is public (`prohibitPublicIpOnVnic` = false), then
	// a public IP address is assigned. If set to true and
	// `prohibitPublicIpOnVnic` = true, an error is returned.
	// **Note:** This public IP address is associated with the primary private IP
	// on the VNIC. For more information, see
	// IP Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingIPaddresses.htm).
	// **Note:** There's a limit to the number of PublicIp
	// a VNIC or instance can have. If you try to create a secondary VNIC
	// with an assigned public IP for an instance that has already
	// reached its public IP limit, an error is returned. For information
	// about the public IP limits, see
	// Public IP Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingpublicIPs.htm).
	// Example: `false`
	// If you specify a `vlanId`, then `assignPublicIp` must be set to false. See
	// Vlan.
	AssignPublicIp *bool `mandatory:"false" json:"assignPublicIp"`

	// Whether the VNIC should be assigned a DNS record. If set to false, there will be no DNS record
	// registration for the VNIC. If set to true, the DNS record will be registered. The default
	// value is true.
	// If you specify a `hostnameLabel`, then `assignPrivateDnsRecord` must be set to true.
	AssignPrivateDnsRecord *bool `mandatory:"false" json:"assignPrivateDnsRecord"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Security Attributes for this resource. This is unique to ZPR, and helps identify which resources are allowed to be accessed by what permission controls.
	// Example: `{"Oracle-DataSecurity-ZPR": {"MaxEgressCount": {"value":"42","mode":"audit"}}}`
	SecurityAttributes map[string]map[string]interface{} `mandatory:"false" json:"securityAttributes"`

	// The hostname for the VNIC's primary private IP. Used for DNS. The value is the hostname
	// portion of the primary private IP's fully qualified domain name (FQDN)
	// (for example, `bminstance1` in FQDN `bminstance1.subnet123.vcn1.oraclevcn.com`).
	// Must be unique across all VNICs in the subnet and comply with
	// RFC 952 (https://tools.ietf.org/html/rfc952) and
	// RFC 1123 (https://tools.ietf.org/html/rfc1123).
	// The value appears in the `Vnic` object and also the
	// `PrivateIp` object returned by
	// `ListPrivateIps` and
	// `GetPrivateIp`.
	// For more information, see
	// DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm).
	// When launching an instance, use this `hostnameLabel` instead
	// of the deprecated `hostnameLabel` in
	// `LaunchInstanceDetails`.
	// If you provide both, the values must match.
	// Example: `bminstance1`
	// If you specify a `vlanId`, the `hostnameLabel` cannot be specified. VNICs on a VLAN
	// can not be assigned a hostname. See Vlan.
	HostnameLabel *string `mandatory:"false" json:"hostnameLabel"`

	// A list of IPv6 prefix ranges from which the VNIC is assigned an IPv6 address.
	// You can provide only the prefix ranges from which OCI selects an available
	// address from the range. You can optionally choose to leave the prefix range empty
	// and instead provide the specific IPv6 address within that range to use.
	Ipv6AddressIpv6SubnetCidrPairDetails []Ipv6AddressIpv6SubnetCidrPairDetails `mandatory:"false" json:"ipv6AddressIpv6SubnetCidrPairDetails"`

	// A list of the OCIDs of the network security groups (NSGs) to add the VNIC to. For more
	// information about NSGs, see
	// NetworkSecurityGroup.
	// If a `vlanId` is specified, the `nsgIds` cannot be specified. The `vlanId`
	// indicates that the VNIC will belong to a VLAN instead of a subnet. With VLANs,
	// all VNICs in the VLAN belong to the NSGs that are associated with the VLAN.
	// See Vlan.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// A private IP address of your choice to assign to the VNIC. Must be an
	// available IP address within the subnet's CIDR. If you don't specify a
	// value, Oracle automatically assigns a private IP address from the subnet.
	// This is the VNIC's *primary* private IP address. The value appears in
	// the `Vnic` object and also the
	// `PrivateIp` object returned by
	// `ListPrivateIps` and
	// `GetPrivateIp`.
	//
	// If you specify a `vlanId`, the `privateIp` cannot be specified.
	// See Vlan.
	// Example: `10.0.3.3`
	PrivateIp *string `mandatory:"false" json:"privateIp"`

	// Whether the source/destination check is disabled on the VNIC.
	// Defaults to `false`, which means the check is performed. For information
	// about why you would skip the source/destination check, see
	// Using a Private IP as a Route Target (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm#privateip).
	//
	// If you specify a `vlanId`, the `skipSourceDestCheck` cannot be specified because the
	// source/destination check is always disabled for VNICs in a VLAN. See
	// Vlan.
	// Example: `true`
	SkipSourceDestCheck *bool `mandatory:"false" json:"skipSourceDestCheck"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet to create the VNIC in. When launching an instance,
	// use this `subnetId` instead of the deprecated `subnetId` in
	// LaunchInstanceDetails.
	// At least one of them is required; if you provide both, the values must match.
	// If you are an Oracle Cloud VMware Solution customer and creating a secondary
	// VNIC in a VLAN instead of a subnet, provide a `vlanId` instead of a `subnetId`.
	// If you provide both a `vlanId` and `subnetId`, the request fails.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// Provide this attribute only if you are an Oracle Cloud VMware Solution
	// customer and creating a secondary VNIC in a VLAN. The value is the OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VLAN.
	// See Vlan.
	// Provide a `vlanId` instead of a `subnetId`. If you provide both a
	// `vlanId` and `subnetId`, the request fails.
	VlanId *string `mandatory:"false" json:"vlanId"`
}

func (m CreateVnicDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateVnicDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
