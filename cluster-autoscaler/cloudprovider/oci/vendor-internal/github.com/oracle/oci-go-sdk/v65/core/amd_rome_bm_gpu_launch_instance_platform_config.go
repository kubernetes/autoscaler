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

// AmdRomeBmGpuLaunchInstancePlatformConfig The platform configuration used when launching a bare metal GPU instance with the BM.GPU4.8 shape
// (the AMD Rome platform).
type AmdRomeBmGpuLaunchInstancePlatformConfig struct {

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

	// Instance Platform Configuration Configuration Map for flexible setting input.
	ConfigMap map[string]string `mandatory:"false" json:"configMap"`

	// The number of NUMA nodes per socket (NPS).
	NumaNodesPerSocket AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum `mandatory:"false" json:"numaNodesPerSocket,omitempty"`
}

// GetIsSecureBootEnabled returns IsSecureBootEnabled
func (m AmdRomeBmGpuLaunchInstancePlatformConfig) GetIsSecureBootEnabled() *bool {
	return m.IsSecureBootEnabled
}

// GetIsTrustedPlatformModuleEnabled returns IsTrustedPlatformModuleEnabled
func (m AmdRomeBmGpuLaunchInstancePlatformConfig) GetIsTrustedPlatformModuleEnabled() *bool {
	return m.IsTrustedPlatformModuleEnabled
}

// GetIsMeasuredBootEnabled returns IsMeasuredBootEnabled
func (m AmdRomeBmGpuLaunchInstancePlatformConfig) GetIsMeasuredBootEnabled() *bool {
	return m.IsMeasuredBootEnabled
}

// GetIsMemoryEncryptionEnabled returns IsMemoryEncryptionEnabled
func (m AmdRomeBmGpuLaunchInstancePlatformConfig) GetIsMemoryEncryptionEnabled() *bool {
	return m.IsMemoryEncryptionEnabled
}

func (m AmdRomeBmGpuLaunchInstancePlatformConfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m AmdRomeBmGpuLaunchInstancePlatformConfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum(string(m.NumaNodesPerSocket)); !ok && m.NumaNodesPerSocket != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for NumaNodesPerSocket: %s. Supported values are: %s.", m.NumaNodesPerSocket, strings.Join(GetAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m AmdRomeBmGpuLaunchInstancePlatformConfig) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeAmdRomeBmGpuLaunchInstancePlatformConfig AmdRomeBmGpuLaunchInstancePlatformConfig
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeAmdRomeBmGpuLaunchInstancePlatformConfig
	}{
		"AMD_ROME_BM_GPU",
		(MarshalTypeAmdRomeBmGpuLaunchInstancePlatformConfig)(m),
	}

	return json.Marshal(&s)
}

// AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum Enum with underlying type: string
type AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum string

// Set of constants representing the allowable values for AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum
const (
	AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps0 AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum = "NPS0"
	AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps1 AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum = "NPS1"
	AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps2 AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum = "NPS2"
	AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps4 AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum = "NPS4"
)

var mappingAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum = map[string]AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum{
	"NPS0": AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps0,
	"NPS1": AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps1,
	"NPS2": AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps2,
	"NPS4": AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps4,
}

var mappingAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnumLowerCase = map[string]AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum{
	"nps0": AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps0,
	"nps1": AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps1,
	"nps2": AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps2,
	"nps4": AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketNps4,
}

// GetAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnumValues Enumerates the set of values for AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum
func GetAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnumValues() []AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum {
	values := make([]AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum, 0)
	for _, v := range mappingAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum {
		values = append(values, v)
	}
	return values
}

// GetAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnumStringValues Enumerates the set of values in String for AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum
func GetAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnumStringValues() []string {
	return []string{
		"NPS0",
		"NPS1",
		"NPS2",
		"NPS4",
	}
}

// GetMappingAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum(val string) (AmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnum, bool) {
	enum, ok := mappingAmdRomeBmGpuLaunchInstancePlatformConfigNumaNodesPerSocketEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
