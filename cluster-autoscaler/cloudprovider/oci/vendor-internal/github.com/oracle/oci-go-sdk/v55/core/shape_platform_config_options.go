// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// ShapePlatformConfigOptions The list of supported platform configuration options for this shape.
type ShapePlatformConfigOptions struct {

	// The type of platform being configured.
	Type ShapePlatformConfigOptionsTypeEnum `mandatory:"false" json:"type,omitempty"`

	SecureBootOptions *ShapeSecureBootOptions `mandatory:"false" json:"secureBootOptions"`

	MeasuredBootOptions *ShapeMeasuredBootOptions `mandatory:"false" json:"measuredBootOptions"`

	TrustedPlatformModuleOptions *ShapeTrustedPlatformModuleOptions `mandatory:"false" json:"trustedPlatformModuleOptions"`

	NumaNodesPerSocketPlatformOptions *ShapeNumaNodesPerSocketPlatformOptions `mandatory:"false" json:"numaNodesPerSocketPlatformOptions"`

	MemoryEncryptionOptions *ShapeMemoryEncryptionOptions `mandatory:"false" json:"memoryEncryptionOptions"`
}

func (m ShapePlatformConfigOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ShapePlatformConfigOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingShapePlatformConfigOptionsTypeEnum[string(m.Type)]; !ok && m.Type != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Type: %s. Supported values are: %s.", m.Type, strings.Join(GetShapePlatformConfigOptionsTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ShapePlatformConfigOptionsTypeEnum Enum with underlying type: string
type ShapePlatformConfigOptionsTypeEnum string

// Set of constants representing the allowable values for ShapePlatformConfigOptionsTypeEnum
const (
	ShapePlatformConfigOptionsTypeAmdMilanBm     ShapePlatformConfigOptionsTypeEnum = "AMD_MILAN_BM"
	ShapePlatformConfigOptionsTypeAmdRomeBm      ShapePlatformConfigOptionsTypeEnum = "AMD_ROME_BM"
	ShapePlatformConfigOptionsTypeIntelSkylakeBm ShapePlatformConfigOptionsTypeEnum = "INTEL_SKYLAKE_BM"
	ShapePlatformConfigOptionsTypeAmdVm          ShapePlatformConfigOptionsTypeEnum = "AMD_VM"
	ShapePlatformConfigOptionsTypeIntelVm        ShapePlatformConfigOptionsTypeEnum = "INTEL_VM"
)

var mappingShapePlatformConfigOptionsTypeEnum = map[string]ShapePlatformConfigOptionsTypeEnum{
	"AMD_MILAN_BM":     ShapePlatformConfigOptionsTypeAmdMilanBm,
	"AMD_ROME_BM":      ShapePlatformConfigOptionsTypeAmdRomeBm,
	"INTEL_SKYLAKE_BM": ShapePlatformConfigOptionsTypeIntelSkylakeBm,
	"AMD_VM":           ShapePlatformConfigOptionsTypeAmdVm,
	"INTEL_VM":         ShapePlatformConfigOptionsTypeIntelVm,
}

// GetShapePlatformConfigOptionsTypeEnumValues Enumerates the set of values for ShapePlatformConfigOptionsTypeEnum
func GetShapePlatformConfigOptionsTypeEnumValues() []ShapePlatformConfigOptionsTypeEnum {
	values := make([]ShapePlatformConfigOptionsTypeEnum, 0)
	for _, v := range mappingShapePlatformConfigOptionsTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetShapePlatformConfigOptionsTypeEnumStringValues Enumerates the set of values in String for ShapePlatformConfigOptionsTypeEnum
func GetShapePlatformConfigOptionsTypeEnumStringValues() []string {
	return []string{
		"AMD_MILAN_BM",
		"AMD_ROME_BM",
		"INTEL_SKYLAKE_BM",
		"AMD_VM",
		"INTEL_VM",
	}
}
