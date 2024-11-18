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

// InstanceConfigurationCreateVnicDetails Contains the properties of the VNIC for an instance configuration. See CreateVnicDetails
// and Instance Configurations (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/instancemanagement.htm#config) for more information.
type InstanceConfigurationCreateVnicDetails struct {

	// Whether to allocate an IPv6 address at instance and VNIC creation from an IPv6 enabled
	// subnet. Default: False. When provided you may optionally provide an IPv6 prefix
	// (`ipv6SubnetCidr`) of your choice to assign the IPv6 address from. If `ipv6SubnetCidr`
	// is not provided then an IPv6 prefix is chosen
	// for you.
	AssignIpv6Ip *bool `mandatory:"false" json:"assignIpv6Ip"`

	// Whether the VNIC should be assigned a public IP address. See the `assignPublicIp` attribute of CreateVnicDetails
	// for more information.
	AssignPublicIp *bool `mandatory:"false" json:"assignPublicIp"`

	// Whether the VNIC should be assigned a private DNS record. See the `assignPrivateDnsRecord` attribute of CreateVnicDetails
	// for more information.
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

	// A list of IPv6 prefixes from which the VNIC should be assigned an IPv6 address.
	// You can provide only the prefix and OCI selects an available
	// address from the range. You can optionally choose to leave the prefix range empty
	// and instead provide the specific IPv6 address that should be used from within that range.
	Ipv6AddressIpv6SubnetCidrPairDetails []InstanceConfigurationIpv6AddressIpv6SubnetCidrPairDetails `mandatory:"false" json:"ipv6AddressIpv6SubnetCidrPairDetails"`

	// The hostname for the VNIC's primary private IP.
	// See the `hostnameLabel` attribute of CreateVnicDetails for more information.
	HostnameLabel *string `mandatory:"false" json:"hostnameLabel"`

	// A list of the OCIDs of the network security groups (NSGs) to add the VNIC to. For more
	// information about NSGs, see
	// NetworkSecurityGroup.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// A private IP address of your choice to assign to the VNIC.
	// See the `privateIp` attribute of CreateVnicDetails for more information.
	PrivateIp *string `mandatory:"false" json:"privateIp"`

	// Whether the source/destination check is disabled on the VNIC.
	// See the `skipSourceDestCheck` attribute of CreateVnicDetails for more information.
	SkipSourceDestCheck *bool `mandatory:"false" json:"skipSourceDestCheck"`

	// The OCID of the subnet to create the VNIC in.
	// See the `subnetId` attribute of CreateVnicDetails for more information.
	SubnetId *string `mandatory:"false" json:"subnetId"`
}

func (m InstanceConfigurationCreateVnicDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstanceConfigurationCreateVnicDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
