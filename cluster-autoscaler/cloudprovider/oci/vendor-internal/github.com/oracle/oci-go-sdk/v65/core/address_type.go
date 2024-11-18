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
	"strings"
)

// AddressTypeEnum Enum with underlying type: string
type AddressTypeEnum string

// Set of constants representing the allowable values for AddressTypeEnum
const (
	AddressTypePrivateIPv4               AddressTypeEnum = "Private_IPv4"
	AddressTypeOracleAllocatedPublicIPv4 AddressTypeEnum = "Oracle_Allocated_Public_IPv4"
	AddressTypeByoipIPv4                 AddressTypeEnum = "BYOIP_IPv4"
	AddressTypeUlaIPv6                   AddressTypeEnum = "ULA_IPv6"
	AddressTypeOracleAllocatedGuaIPv6    AddressTypeEnum = "Oracle_Allocated_GUA_IPv6"
	AddressTypeByoipIPv6                 AddressTypeEnum = "BYOIP_IPv6"
)

var mappingAddressTypeEnum = map[string]AddressTypeEnum{
	"Private_IPv4":                 AddressTypePrivateIPv4,
	"Oracle_Allocated_Public_IPv4": AddressTypeOracleAllocatedPublicIPv4,
	"BYOIP_IPv4":                   AddressTypeByoipIPv4,
	"ULA_IPv6":                     AddressTypeUlaIPv6,
	"Oracle_Allocated_GUA_IPv6":    AddressTypeOracleAllocatedGuaIPv6,
	"BYOIP_IPv6":                   AddressTypeByoipIPv6,
}

var mappingAddressTypeEnumLowerCase = map[string]AddressTypeEnum{
	"private_ipv4":                 AddressTypePrivateIPv4,
	"oracle_allocated_public_ipv4": AddressTypeOracleAllocatedPublicIPv4,
	"byoip_ipv4":                   AddressTypeByoipIPv4,
	"ula_ipv6":                     AddressTypeUlaIPv6,
	"oracle_allocated_gua_ipv6":    AddressTypeOracleAllocatedGuaIPv6,
	"byoip_ipv6":                   AddressTypeByoipIPv6,
}

// GetAddressTypeEnumValues Enumerates the set of values for AddressTypeEnum
func GetAddressTypeEnumValues() []AddressTypeEnum {
	values := make([]AddressTypeEnum, 0)
	for _, v := range mappingAddressTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetAddressTypeEnumStringValues Enumerates the set of values in String for AddressTypeEnum
func GetAddressTypeEnumStringValues() []string {
	return []string{
		"Private_IPv4",
		"Oracle_Allocated_Public_IPv4",
		"BYOIP_IPv4",
		"ULA_IPv6",
		"Oracle_Allocated_GUA_IPv6",
		"BYOIP_IPv6",
	}
}

// GetMappingAddressTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingAddressTypeEnum(val string) (AddressTypeEnum, bool) {
	enum, ok := mappingAddressTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
