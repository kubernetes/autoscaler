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

// ShapePlatformConfigOptions The list of supported platform configuration options for this shape.
type ShapePlatformConfigOptions struct {

	// The type of platform being configured.
	Type ShapePlatformConfigOptionsTypeEnum `mandatory:"false" json:"type,omitempty"`

	SecureBootOptions *ShapeSecureBootOptions `mandatory:"false" json:"secureBootOptions"`

	MeasuredBootOptions *ShapeMeasuredBootOptions `mandatory:"false" json:"measuredBootOptions"`

	TrustedPlatformModuleOptions *ShapeTrustedPlatformModuleOptions `mandatory:"false" json:"trustedPlatformModuleOptions"`

	NumaNodesPerSocketPlatformOptions *ShapeNumaNodesPerSocketPlatformOptions `mandatory:"false" json:"numaNodesPerSocketPlatformOptions"`

	MemoryEncryptionOptions *ShapeMemoryEncryptionOptions `mandatory:"false" json:"memoryEncryptionOptions"`

	SymmetricMultiThreadingOptions *ShapeSymmetricMultiThreadingEnabledPlatformOptions `mandatory:"false" json:"symmetricMultiThreadingOptions"`

	AccessControlServiceOptions *ShapeAccessControlServiceEnabledPlatformOptions `mandatory:"false" json:"accessControlServiceOptions"`

	VirtualInstructionsOptions *ShapeVirtualInstructionsEnabledPlatformOptions `mandatory:"false" json:"virtualInstructionsOptions"`

	InputOutputMemoryManagementUnitOptions *ShapeInputOutputMemoryManagementUnitEnabledPlatformOptions `mandatory:"false" json:"inputOutputMemoryManagementUnitOptions"`

	PercentageOfCoresEnabledOptions *PercentageOfCoresEnabledOptions `mandatory:"false" json:"percentageOfCoresEnabledOptions"`
}

func (m ShapePlatformConfigOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ShapePlatformConfigOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingShapePlatformConfigOptionsTypeEnum(string(m.Type)); !ok && m.Type != "" {
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
	ShapePlatformConfigOptionsTypeAmdMilanBmGpu  ShapePlatformConfigOptionsTypeEnum = "AMD_MILAN_BM_GPU"
	ShapePlatformConfigOptionsTypeAmdRomeBm      ShapePlatformConfigOptionsTypeEnum = "AMD_ROME_BM"
	ShapePlatformConfigOptionsTypeAmdRomeBmGpu   ShapePlatformConfigOptionsTypeEnum = "AMD_ROME_BM_GPU"
	ShapePlatformConfigOptionsTypeGenericBm      ShapePlatformConfigOptionsTypeEnum = "GENERIC_BM"
	ShapePlatformConfigOptionsTypeIntelIcelakeBm ShapePlatformConfigOptionsTypeEnum = "INTEL_ICELAKE_BM"
	ShapePlatformConfigOptionsTypeIntelSkylakeBm ShapePlatformConfigOptionsTypeEnum = "INTEL_SKYLAKE_BM"
	ShapePlatformConfigOptionsTypeAmdVm          ShapePlatformConfigOptionsTypeEnum = "AMD_VM"
	ShapePlatformConfigOptionsTypeIntelVm        ShapePlatformConfigOptionsTypeEnum = "INTEL_VM"
)

var mappingShapePlatformConfigOptionsTypeEnum = map[string]ShapePlatformConfigOptionsTypeEnum{
	"AMD_MILAN_BM":     ShapePlatformConfigOptionsTypeAmdMilanBm,
	"AMD_MILAN_BM_GPU": ShapePlatformConfigOptionsTypeAmdMilanBmGpu,
	"AMD_ROME_BM":      ShapePlatformConfigOptionsTypeAmdRomeBm,
	"AMD_ROME_BM_GPU":  ShapePlatformConfigOptionsTypeAmdRomeBmGpu,
	"GENERIC_BM":       ShapePlatformConfigOptionsTypeGenericBm,
	"INTEL_ICELAKE_BM": ShapePlatformConfigOptionsTypeIntelIcelakeBm,
	"INTEL_SKYLAKE_BM": ShapePlatformConfigOptionsTypeIntelSkylakeBm,
	"AMD_VM":           ShapePlatformConfigOptionsTypeAmdVm,
	"INTEL_VM":         ShapePlatformConfigOptionsTypeIntelVm,
}

var mappingShapePlatformConfigOptionsTypeEnumLowerCase = map[string]ShapePlatformConfigOptionsTypeEnum{
	"amd_milan_bm":     ShapePlatformConfigOptionsTypeAmdMilanBm,
	"amd_milan_bm_gpu": ShapePlatformConfigOptionsTypeAmdMilanBmGpu,
	"amd_rome_bm":      ShapePlatformConfigOptionsTypeAmdRomeBm,
	"amd_rome_bm_gpu":  ShapePlatformConfigOptionsTypeAmdRomeBmGpu,
	"generic_bm":       ShapePlatformConfigOptionsTypeGenericBm,
	"intel_icelake_bm": ShapePlatformConfigOptionsTypeIntelIcelakeBm,
	"intel_skylake_bm": ShapePlatformConfigOptionsTypeIntelSkylakeBm,
	"amd_vm":           ShapePlatformConfigOptionsTypeAmdVm,
	"intel_vm":         ShapePlatformConfigOptionsTypeIntelVm,
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
		"AMD_MILAN_BM_GPU",
		"AMD_ROME_BM",
		"AMD_ROME_BM_GPU",
		"GENERIC_BM",
		"INTEL_ICELAKE_BM",
		"INTEL_SKYLAKE_BM",
		"AMD_VM",
		"INTEL_VM",
	}
}

// GetMappingShapePlatformConfigOptionsTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingShapePlatformConfigOptionsTypeEnum(val string) (ShapePlatformConfigOptionsTypeEnum, bool) {
	enum, ok := mappingShapePlatformConfigOptionsTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
