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

// AmdRomeBmPlatformConfig The platform configuration of a bare metal instance that uses the BM.Standard.E3.128 shape (the AMD Rome platform).
type AmdRomeBmPlatformConfig struct {

	// Whether Secure Boot is enabled on the instance.
	IsSecureBootEnabled *bool `mandatory:"false" json:"isSecureBootEnabled"`

	// Whether the Trusted Platform Module (TPM) is enabled on the instance.
	IsTrustedPlatformModuleEnabled *bool `mandatory:"false" json:"isTrustedPlatformModuleEnabled"`

	// Whether the Measured Boot feature is enabled on the instance.
	IsMeasuredBootEnabled *bool `mandatory:"false" json:"isMeasuredBootEnabled"`

	// Whether the instance is a confidential instance. If this value is `true`, the instance is a confidential instance. The default value is `false`.
	IsMemoryEncryptionEnabled *bool `mandatory:"false" json:"isMemoryEncryptionEnabled"`

	// Whether symmetric multithreading is enabled on the instance. Symmetric multithreading is also
	// called simultaneous multithreading (SMT) or Intel Hyper-Threading.
	// Intel and AMD processors have two hardware execution threads per core (OCPU). SMT permits multiple
	// independent threads of execution, to better use the resources and increase the efficiency
	// of the CPU. When multithreading is disabled, only one thread is permitted to run on each core, which
	// can provide higher or more predictable performance for some workloads.
	IsSymmetricMultiThreadingEnabled *bool `mandatory:"false" json:"isSymmetricMultiThreadingEnabled"`

	// Whether the Access Control Service is enabled on the instance. When enabled,
	// the platform can enforce PCIe device isolation, required for VFIO device pass-through.
	IsAccessControlServiceEnabled *bool `mandatory:"false" json:"isAccessControlServiceEnabled"`

	// Whether virtualization instructions are available. For example, Secure Virtual Machine for AMD shapes
	// or VT-x for Intel shapes.
	AreVirtualInstructionsEnabled *bool `mandatory:"false" json:"areVirtualInstructionsEnabled"`

	// Whether the input-output memory management unit is enabled.
	IsInputOutputMemoryManagementUnitEnabled *bool `mandatory:"false" json:"isInputOutputMemoryManagementUnitEnabled"`

	// The percentage of cores enabled. Value must be a multiple of 25%. If the requested percentage
	// results in a fractional number of cores, the system rounds up the number of cores across processors
	// and provisions an instance with a whole number of cores.
	// If the applications that you run on the instance use a core-based licensing model and need fewer cores
	// than the full size of the shape, you can disable cores to reduce your licensing costs. The instance
	// itself is billed for the full shape, regardless of whether all cores are enabled.
	PercentageOfCoresEnabled *int `mandatory:"false" json:"percentageOfCoresEnabled"`

	// Instance Platform Configuration Configuration Map for flexible setting input.
	ConfigMap map[string]string `mandatory:"false" json:"configMap"`

	// The number of NUMA nodes per socket (NPS).
	NumaNodesPerSocket AmdRomeBmPlatformConfigNumaNodesPerSocketEnum `mandatory:"false" json:"numaNodesPerSocket,omitempty"`
}

// GetIsSecureBootEnabled returns IsSecureBootEnabled
func (m AmdRomeBmPlatformConfig) GetIsSecureBootEnabled() *bool {
	return m.IsSecureBootEnabled
}

// GetIsTrustedPlatformModuleEnabled returns IsTrustedPlatformModuleEnabled
func (m AmdRomeBmPlatformConfig) GetIsTrustedPlatformModuleEnabled() *bool {
	return m.IsTrustedPlatformModuleEnabled
}

// GetIsMeasuredBootEnabled returns IsMeasuredBootEnabled
func (m AmdRomeBmPlatformConfig) GetIsMeasuredBootEnabled() *bool {
	return m.IsMeasuredBootEnabled
}

// GetIsMemoryEncryptionEnabled returns IsMemoryEncryptionEnabled
func (m AmdRomeBmPlatformConfig) GetIsMemoryEncryptionEnabled() *bool {
	return m.IsMemoryEncryptionEnabled
}

func (m AmdRomeBmPlatformConfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m AmdRomeBmPlatformConfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingAmdRomeBmPlatformConfigNumaNodesPerSocketEnum(string(m.NumaNodesPerSocket)); !ok && m.NumaNodesPerSocket != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for NumaNodesPerSocket: %s. Supported values are: %s.", m.NumaNodesPerSocket, strings.Join(GetAmdRomeBmPlatformConfigNumaNodesPerSocketEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m AmdRomeBmPlatformConfig) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeAmdRomeBmPlatformConfig AmdRomeBmPlatformConfig
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeAmdRomeBmPlatformConfig
	}{
		"AMD_ROME_BM",
		(MarshalTypeAmdRomeBmPlatformConfig)(m),
	}

	return json.Marshal(&s)
}

// AmdRomeBmPlatformConfigNumaNodesPerSocketEnum Enum with underlying type: string
type AmdRomeBmPlatformConfigNumaNodesPerSocketEnum string

// Set of constants representing the allowable values for AmdRomeBmPlatformConfigNumaNodesPerSocketEnum
const (
	AmdRomeBmPlatformConfigNumaNodesPerSocketNps0 AmdRomeBmPlatformConfigNumaNodesPerSocketEnum = "NPS0"
	AmdRomeBmPlatformConfigNumaNodesPerSocketNps1 AmdRomeBmPlatformConfigNumaNodesPerSocketEnum = "NPS1"
	AmdRomeBmPlatformConfigNumaNodesPerSocketNps2 AmdRomeBmPlatformConfigNumaNodesPerSocketEnum = "NPS2"
	AmdRomeBmPlatformConfigNumaNodesPerSocketNps4 AmdRomeBmPlatformConfigNumaNodesPerSocketEnum = "NPS4"
)

var mappingAmdRomeBmPlatformConfigNumaNodesPerSocketEnum = map[string]AmdRomeBmPlatformConfigNumaNodesPerSocketEnum{
	"NPS0": AmdRomeBmPlatformConfigNumaNodesPerSocketNps0,
	"NPS1": AmdRomeBmPlatformConfigNumaNodesPerSocketNps1,
	"NPS2": AmdRomeBmPlatformConfigNumaNodesPerSocketNps2,
	"NPS4": AmdRomeBmPlatformConfigNumaNodesPerSocketNps4,
}

var mappingAmdRomeBmPlatformConfigNumaNodesPerSocketEnumLowerCase = map[string]AmdRomeBmPlatformConfigNumaNodesPerSocketEnum{
	"nps0": AmdRomeBmPlatformConfigNumaNodesPerSocketNps0,
	"nps1": AmdRomeBmPlatformConfigNumaNodesPerSocketNps1,
	"nps2": AmdRomeBmPlatformConfigNumaNodesPerSocketNps2,
	"nps4": AmdRomeBmPlatformConfigNumaNodesPerSocketNps4,
}

// GetAmdRomeBmPlatformConfigNumaNodesPerSocketEnumValues Enumerates the set of values for AmdRomeBmPlatformConfigNumaNodesPerSocketEnum
func GetAmdRomeBmPlatformConfigNumaNodesPerSocketEnumValues() []AmdRomeBmPlatformConfigNumaNodesPerSocketEnum {
	values := make([]AmdRomeBmPlatformConfigNumaNodesPerSocketEnum, 0)
	for _, v := range mappingAmdRomeBmPlatformConfigNumaNodesPerSocketEnum {
		values = append(values, v)
	}
	return values
}

// GetAmdRomeBmPlatformConfigNumaNodesPerSocketEnumStringValues Enumerates the set of values in String for AmdRomeBmPlatformConfigNumaNodesPerSocketEnum
func GetAmdRomeBmPlatformConfigNumaNodesPerSocketEnumStringValues() []string {
	return []string{
		"NPS0",
		"NPS1",
		"NPS2",
		"NPS4",
	}
}

// GetMappingAmdRomeBmPlatformConfigNumaNodesPerSocketEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingAmdRomeBmPlatformConfigNumaNodesPerSocketEnum(val string) (AmdRomeBmPlatformConfigNumaNodesPerSocketEnum, bool) {
	enum, ok := mappingAmdRomeBmPlatformConfigNumaNodesPerSocketEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
