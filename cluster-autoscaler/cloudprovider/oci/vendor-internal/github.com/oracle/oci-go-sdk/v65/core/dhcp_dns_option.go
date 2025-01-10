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
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// DhcpDnsOption DHCP option for specifying how DNS (hostname resolution) is handled in the subnets in the VCN.
// For more information, see
// DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm).
type DhcpDnsOption struct {

	// If you set `serverType` to `CustomDnsServer`, specify the
	// IP address of at least one DNS server of your choice (three maximum).
	CustomDnsServers []string `mandatory:"false" json:"customDnsServers"`

	// * **VcnLocal:** Reserved for future use.
	// * **VcnLocalPlusInternet:** Also referred to as "Internet and VCN Resolver".
	// Instances can resolve internet hostnames (no internet gateway is required),
	// and can resolve hostnames of instances in the VCN. This is the default
	// value in the default set of DHCP options in the VCN. For the Internet and
	// VCN Resolver to work across the VCN, there must also be a DNS label set for
	// the VCN, a DNS label set for each subnet, and a hostname for each instance.
	// The Internet and VCN Resolver also enables reverse DNS lookup, which lets
	// you determine the hostname corresponding to the private IP address. For more
	// information, see
	// DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm).
	// * **CustomDnsServer:** Instances use a DNS server of your choice (three
	// maximum).
	ServerType DhcpDnsOptionServerTypeEnum `mandatory:"true" json:"serverType"`
}

func (m DhcpDnsOption) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DhcpDnsOption) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingDhcpDnsOptionServerTypeEnum(string(m.ServerType)); !ok && m.ServerType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ServerType: %s. Supported values are: %s.", m.ServerType, strings.Join(GetDhcpDnsOptionServerTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m DhcpDnsOption) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeDhcpDnsOption DhcpDnsOption
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeDhcpDnsOption
	}{
		"DomainNameServer",
		(MarshalTypeDhcpDnsOption)(m),
	}

	return json.Marshal(&s)
}

// DhcpDnsOptionServerTypeEnum Enum with underlying type: string
type DhcpDnsOptionServerTypeEnum string

// Set of constants representing the allowable values for DhcpDnsOptionServerTypeEnum
const (
	DhcpDnsOptionServerTypeVcnlocal             DhcpDnsOptionServerTypeEnum = "VcnLocal"
	DhcpDnsOptionServerTypeVcnlocalplusinternet DhcpDnsOptionServerTypeEnum = "VcnLocalPlusInternet"
	DhcpDnsOptionServerTypeCustomdnsserver      DhcpDnsOptionServerTypeEnum = "CustomDnsServer"
)

var mappingDhcpDnsOptionServerTypeEnum = map[string]DhcpDnsOptionServerTypeEnum{
	"VcnLocal":             DhcpDnsOptionServerTypeVcnlocal,
	"VcnLocalPlusInternet": DhcpDnsOptionServerTypeVcnlocalplusinternet,
	"CustomDnsServer":      DhcpDnsOptionServerTypeCustomdnsserver,
}

var mappingDhcpDnsOptionServerTypeEnumLowerCase = map[string]DhcpDnsOptionServerTypeEnum{
	"vcnlocal":             DhcpDnsOptionServerTypeVcnlocal,
	"vcnlocalplusinternet": DhcpDnsOptionServerTypeVcnlocalplusinternet,
	"customdnsserver":      DhcpDnsOptionServerTypeCustomdnsserver,
}

// GetDhcpDnsOptionServerTypeEnumValues Enumerates the set of values for DhcpDnsOptionServerTypeEnum
func GetDhcpDnsOptionServerTypeEnumValues() []DhcpDnsOptionServerTypeEnum {
	values := make([]DhcpDnsOptionServerTypeEnum, 0)
	for _, v := range mappingDhcpDnsOptionServerTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetDhcpDnsOptionServerTypeEnumStringValues Enumerates the set of values in String for DhcpDnsOptionServerTypeEnum
func GetDhcpDnsOptionServerTypeEnumStringValues() []string {
	return []string{
		"VcnLocal",
		"VcnLocalPlusInternet",
		"CustomDnsServer",
	}
}

// GetMappingDhcpDnsOptionServerTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDhcpDnsOptionServerTypeEnum(val string) (DhcpDnsOptionServerTypeEnum, bool) {
	enum, ok := mappingDhcpDnsOptionServerTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
