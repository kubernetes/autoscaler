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

// VirtualCircuitIpMtuEnum Enum with underlying type: string
type VirtualCircuitIpMtuEnum string

// Set of constants representing the allowable values for VirtualCircuitIpMtuEnum
const (
	VirtualCircuitIpMtuMtu1500 VirtualCircuitIpMtuEnum = "MTU_1500"
	VirtualCircuitIpMtuMtu9000 VirtualCircuitIpMtuEnum = "MTU_9000"
)

var mappingVirtualCircuitIpMtuEnum = map[string]VirtualCircuitIpMtuEnum{
	"MTU_1500": VirtualCircuitIpMtuMtu1500,
	"MTU_9000": VirtualCircuitIpMtuMtu9000,
}

var mappingVirtualCircuitIpMtuEnumLowerCase = map[string]VirtualCircuitIpMtuEnum{
	"mtu_1500": VirtualCircuitIpMtuMtu1500,
	"mtu_9000": VirtualCircuitIpMtuMtu9000,
}

// GetVirtualCircuitIpMtuEnumValues Enumerates the set of values for VirtualCircuitIpMtuEnum
func GetVirtualCircuitIpMtuEnumValues() []VirtualCircuitIpMtuEnum {
	values := make([]VirtualCircuitIpMtuEnum, 0)
	for _, v := range mappingVirtualCircuitIpMtuEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitIpMtuEnumStringValues Enumerates the set of values in String for VirtualCircuitIpMtuEnum
func GetVirtualCircuitIpMtuEnumStringValues() []string {
	return []string{
		"MTU_1500",
		"MTU_9000",
	}
}

// GetMappingVirtualCircuitIpMtuEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitIpMtuEnum(val string) (VirtualCircuitIpMtuEnum, bool) {
	enum, ok := mappingVirtualCircuitIpMtuEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
