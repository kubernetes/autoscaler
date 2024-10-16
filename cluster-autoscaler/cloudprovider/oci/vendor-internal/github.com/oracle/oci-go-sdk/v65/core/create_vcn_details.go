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

// CreateVcnDetails The representation of CreateVcnDetails
type CreateVcnDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment to contain the VCN.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// **Deprecated.** Do *not* set this value. Use `cidrBlocks` instead.
	// Example: `10.0.0.0/16`
	CidrBlock *string `mandatory:"false" json:"cidrBlock"`

	// The list of one or more IPv4 CIDR blocks for the VCN that meet the following criteria:
	// - The CIDR blocks must be valid.
	// - They must not overlap with each other or with the on-premises network CIDR block.
	// - The number of CIDR blocks must not exceed the limit of CIDR blocks allowed per VCN.
	// **Important:** Do *not* specify a value for `cidrBlock`. Use this parameter instead.
	CidrBlocks []string `mandatory:"false" json:"cidrBlocks"`

	// The list of one or more ULA or Private IPv6 prefixes for the VCN that meets the following criteria:
	// - The CIDR blocks must be valid.
	// - Multiple CIDR blocks must not overlap each other or the on-premises network prefix.
	// - The number of CIDR blocks must not exceed the limit of IPv6 prefixes allowed to a VCN.
	// **Important:** Do *not* specify a value for `ipv6CidrBlock`. Use this parameter instead.
	Ipv6PrivateCidrBlocks []string `mandatory:"false" json:"ipv6PrivateCidrBlocks"`

	// Specifies whether to skip Oracle allocated IPv6 GUA. By default, Oracle will allocate one GUA of /56
	// size for an IPv6 enabled VCN.
	IsOracleGuaAllocationEnabled *bool `mandatory:"false" json:"isOracleGuaAllocationEnabled"`

	// The list of BYOIPv6 OCIDs and BYOIPv6 prefixes required to create a VCN that uses BYOIPv6 address ranges.
	Byoipv6CidrDetails []Byoipv6CidrDetails `mandatory:"false" json:"byoipv6CidrDetails"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// A DNS label for the VCN, used in conjunction with the VNIC's hostname and
	// subnet's DNS label to form a fully qualified domain name (FQDN) for each VNIC
	// within this subnet (for example, `bminstance1.subnet123.vcn1.oraclevcn.com`).
	// Not required to be unique, but it's a best practice to set unique DNS labels
	// for VCNs in your tenancy. Must be an alphanumeric string that begins with a letter.
	// The value cannot be changed.
	// You must set this value if you want instances to be able to use hostnames to
	// resolve other instances in the VCN. Otherwise the Internet and VCN Resolver
	// will not work.
	// For more information, see
	// DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm).
	// Example: `vcn1`
	DnsLabel *string `mandatory:"false" json:"dnsLabel"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Security Attributes for this resource. This is unique to ZPR, and helps identify which resources are allowed to be accessed by what permission controls.
	// Example: `{"Oracle-DataSecurity-ZPR": {"MaxEgressCount": {"value":"42","mode":"audit"}}}`
	SecurityAttributes map[string]map[string]interface{} `mandatory:"false" json:"securityAttributes"`

	// Whether IPv6 is enabled for the VCN. Default is `false`.
	// If enabled, Oracle will assign the VCN a IPv6 /56 CIDR block.
	// You may skip having Oracle allocate the VCN a IPv6 /56 CIDR block by setting isOracleGuaAllocationEnabled to `false`.
	// For important details about IPv6 addressing in a VCN, see IPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	// Example: `true`
	IsIpv6Enabled *bool `mandatory:"false" json:"isIpv6Enabled"`
}

func (m CreateVcnDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateVcnDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
