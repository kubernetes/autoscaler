// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// VirtualCircuitRedundancyMetadata This resource provides redundancy level details for the virtual circuit. For more about redundancy, see FastConnect Redundancy Best Practices (https://docs.oracle.com/iaas/Content/Network/Concepts/fastconnectresiliency.htm).
type VirtualCircuitRedundancyMetadata struct {

	// The configured redundancy level of the virtual circuit.
	ConfiguredRedundancyLevel VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum `mandatory:"false" json:"configuredRedundancyLevel,omitempty"`

	// Indicates if the configured level is met for IPv4 BGP redundancy.
	Ipv4bgpSessionRedundancyStatus VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum `mandatory:"false" json:"ipv4bgpSessionRedundancyStatus,omitempty"`

	// Indicates if the configured level is met for IPv6 BGP redundancy.
	Ipv6bgpSessionRedundancyStatus VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum `mandatory:"false" json:"ipv6bgpSessionRedundancyStatus,omitempty"`
}

func (m VirtualCircuitRedundancyMetadata) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VirtualCircuitRedundancyMetadata) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum(string(m.ConfiguredRedundancyLevel)); !ok && m.ConfiguredRedundancyLevel != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ConfiguredRedundancyLevel: %s. Supported values are: %s.", m.ConfiguredRedundancyLevel, strings.Join(GetVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum(string(m.Ipv4bgpSessionRedundancyStatus)); !ok && m.Ipv4bgpSessionRedundancyStatus != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Ipv4bgpSessionRedundancyStatus: %s. Supported values are: %s.", m.Ipv4bgpSessionRedundancyStatus, strings.Join(GetVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum(string(m.Ipv6bgpSessionRedundancyStatus)); !ok && m.Ipv6bgpSessionRedundancyStatus != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Ipv6bgpSessionRedundancyStatus: %s. Supported values are: %s.", m.Ipv6bgpSessionRedundancyStatus, strings.Join(GetVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum Enum with underlying type: string
type VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum string

// Set of constants representing the allowable values for VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum
const (
	VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelDevice       VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum = "DEVICE"
	VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelPop          VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum = "POP"
	VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelRegion       VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum = "REGION"
	VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelNonRedundant VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum = "NON_REDUNDANT"
	VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelPending      VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum = "PENDING"
)

var mappingVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum = map[string]VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum{
	"DEVICE":        VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelDevice,
	"POP":           VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelPop,
	"REGION":        VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelRegion,
	"NON_REDUNDANT": VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelNonRedundant,
	"PENDING":       VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelPending,
}

var mappingVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnumLowerCase = map[string]VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum{
	"device":        VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelDevice,
	"pop":           VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelPop,
	"region":        VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelRegion,
	"non_redundant": VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelNonRedundant,
	"pending":       VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelPending,
}

// GetVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnumValues Enumerates the set of values for VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum
func GetVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnumValues() []VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum {
	values := make([]VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum, 0)
	for _, v := range mappingVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnumStringValues Enumerates the set of values in String for VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum
func GetVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnumStringValues() []string {
	return []string{
		"DEVICE",
		"POP",
		"REGION",
		"NON_REDUNDANT",
		"PENDING",
	}
}

// GetMappingVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum(val string) (VirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnum, bool) {
	enum, ok := mappingVirtualCircuitRedundancyMetadataConfiguredRedundancyLevelEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum Enum with underlying type: string
type VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum string

// Set of constants representing the allowable values for VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum
const (
	VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusConfigurationMatch    VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum = "CONFIGURATION_MATCH"
	VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusConfigurationMismatch VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum = "CONFIGURATION_MISMATCH"
	VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusNotMetSla             VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum = "NOT_MET_SLA"
)

var mappingVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum = map[string]VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum{
	"CONFIGURATION_MATCH":    VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusConfigurationMatch,
	"CONFIGURATION_MISMATCH": VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusConfigurationMismatch,
	"NOT_MET_SLA":            VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusNotMetSla,
}

var mappingVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnumLowerCase = map[string]VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum{
	"configuration_match":    VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusConfigurationMatch,
	"configuration_mismatch": VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusConfigurationMismatch,
	"not_met_sla":            VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusNotMetSla,
}

// GetVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnumValues Enumerates the set of values for VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum
func GetVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnumValues() []VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum {
	values := make([]VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum, 0)
	for _, v := range mappingVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnumStringValues Enumerates the set of values in String for VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum
func GetVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnumStringValues() []string {
	return []string{
		"CONFIGURATION_MATCH",
		"CONFIGURATION_MISMATCH",
		"NOT_MET_SLA",
	}
}

// GetMappingVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum(val string) (VirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnum, bool) {
	enum, ok := mappingVirtualCircuitRedundancyMetadataIpv4bgpSessionRedundancyStatusEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum Enum with underlying type: string
type VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum string

// Set of constants representing the allowable values for VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum
const (
	VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusConfigurationMatch    VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum = "CONFIGURATION_MATCH"
	VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusConfigurationMismatch VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum = "CONFIGURATION_MISMATCH"
	VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusNotMetSla             VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum = "NOT_MET_SLA"
)

var mappingVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum = map[string]VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum{
	"CONFIGURATION_MATCH":    VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusConfigurationMatch,
	"CONFIGURATION_MISMATCH": VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusConfigurationMismatch,
	"NOT_MET_SLA":            VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusNotMetSla,
}

var mappingVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnumLowerCase = map[string]VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum{
	"configuration_match":    VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusConfigurationMatch,
	"configuration_mismatch": VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusConfigurationMismatch,
	"not_met_sla":            VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusNotMetSla,
}

// GetVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnumValues Enumerates the set of values for VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum
func GetVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnumValues() []VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum {
	values := make([]VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum, 0)
	for _, v := range mappingVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnumStringValues Enumerates the set of values in String for VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum
func GetVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnumStringValues() []string {
	return []string{
		"CONFIGURATION_MATCH",
		"CONFIGURATION_MISMATCH",
		"NOT_MET_SLA",
	}
}

// GetMappingVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum(val string) (VirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnum, bool) {
	enum, ok := mappingVirtualCircuitRedundancyMetadataIpv6bgpSessionRedundancyStatusEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
