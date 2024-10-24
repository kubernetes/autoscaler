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

// InstanceShapeConfig The shape configuration for an instance. The shape configuration determines
// the resources allocated to an instance.
type InstanceShapeConfig struct {

	// The total number of OCPUs available to the instance.
	Ocpus *float32 `mandatory:"false" json:"ocpus"`

	// The total amount of memory available to the instance, in gigabytes.
	MemoryInGBs *float32 `mandatory:"false" json:"memoryInGBs"`

	// The baseline OCPU utilization for a subcore burstable VM instance. Leave this attribute blank for a
	// non-burstable instance, or explicitly specify non-burstable with `BASELINE_1_1`.
	// The following values are supported:
	// - `BASELINE_1_8` - baseline usage is 1/8 of an OCPU.
	// - `BASELINE_1_2` - baseline usage is 1/2 of an OCPU.
	// - `BASELINE_1_1` - baseline usage is the entire OCPU. This represents a non-burstable instance.
	BaselineOcpuUtilization InstanceShapeConfigBaselineOcpuUtilizationEnum `mandatory:"false" json:"baselineOcpuUtilization,omitempty"`

	// A short description of the instance's processor (CPU).
	ProcessorDescription *string `mandatory:"false" json:"processorDescription"`

	// The networking bandwidth available to the instance, in gigabits per second.
	NetworkingBandwidthInGbps *float32 `mandatory:"false" json:"networkingBandwidthInGbps"`

	// The maximum number of VNIC attachments for the instance.
	MaxVnicAttachments *int `mandatory:"false" json:"maxVnicAttachments"`

	// The number of GPUs available to the instance.
	Gpus *int `mandatory:"false" json:"gpus"`

	// A short description of the instance's graphics processing unit (GPU).
	// If the instance does not have any GPUs, this field is `null`.
	GpuDescription *string `mandatory:"false" json:"gpuDescription"`

	// The number of local disks available to the instance.
	LocalDisks *int `mandatory:"false" json:"localDisks"`

	// The aggregate size of all local disks, in gigabytes.
	// If the instance does not have any local disks, this field is `null`.
	LocalDisksTotalSizeInGBs *float32 `mandatory:"false" json:"localDisksTotalSizeInGBs"`

	// A short description of the local disks available to this instance.
	// If the instance does not have any local disks, this field is `null`.
	LocalDiskDescription *string `mandatory:"false" json:"localDiskDescription"`

	// The total number of VCPUs available to the instance. This can be used instead of OCPUs,
	// in which case the actual number of OCPUs will be calculated based on this value
	// and the actual hardware. This must be a multiple of 2.
	Vcpus *int `mandatory:"false" json:"vcpus"`
}

func (m InstanceShapeConfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstanceShapeConfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingInstanceShapeConfigBaselineOcpuUtilizationEnum(string(m.BaselineOcpuUtilization)); !ok && m.BaselineOcpuUtilization != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BaselineOcpuUtilization: %s. Supported values are: %s.", m.BaselineOcpuUtilization, strings.Join(GetInstanceShapeConfigBaselineOcpuUtilizationEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstanceShapeConfigBaselineOcpuUtilizationEnum Enum with underlying type: string
type InstanceShapeConfigBaselineOcpuUtilizationEnum string

// Set of constants representing the allowable values for InstanceShapeConfigBaselineOcpuUtilizationEnum
const (
	InstanceShapeConfigBaselineOcpuUtilization8 InstanceShapeConfigBaselineOcpuUtilizationEnum = "BASELINE_1_8"
	InstanceShapeConfigBaselineOcpuUtilization2 InstanceShapeConfigBaselineOcpuUtilizationEnum = "BASELINE_1_2"
	InstanceShapeConfigBaselineOcpuUtilization1 InstanceShapeConfigBaselineOcpuUtilizationEnum = "BASELINE_1_1"
)

var mappingInstanceShapeConfigBaselineOcpuUtilizationEnum = map[string]InstanceShapeConfigBaselineOcpuUtilizationEnum{
	"BASELINE_1_8": InstanceShapeConfigBaselineOcpuUtilization8,
	"BASELINE_1_2": InstanceShapeConfigBaselineOcpuUtilization2,
	"BASELINE_1_1": InstanceShapeConfigBaselineOcpuUtilization1,
}

var mappingInstanceShapeConfigBaselineOcpuUtilizationEnumLowerCase = map[string]InstanceShapeConfigBaselineOcpuUtilizationEnum{
	"baseline_1_8": InstanceShapeConfigBaselineOcpuUtilization8,
	"baseline_1_2": InstanceShapeConfigBaselineOcpuUtilization2,
	"baseline_1_1": InstanceShapeConfigBaselineOcpuUtilization1,
}

// GetInstanceShapeConfigBaselineOcpuUtilizationEnumValues Enumerates the set of values for InstanceShapeConfigBaselineOcpuUtilizationEnum
func GetInstanceShapeConfigBaselineOcpuUtilizationEnumValues() []InstanceShapeConfigBaselineOcpuUtilizationEnum {
	values := make([]InstanceShapeConfigBaselineOcpuUtilizationEnum, 0)
	for _, v := range mappingInstanceShapeConfigBaselineOcpuUtilizationEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceShapeConfigBaselineOcpuUtilizationEnumStringValues Enumerates the set of values in String for InstanceShapeConfigBaselineOcpuUtilizationEnum
func GetInstanceShapeConfigBaselineOcpuUtilizationEnumStringValues() []string {
	return []string{
		"BASELINE_1_8",
		"BASELINE_1_2",
		"BASELINE_1_1",
	}
}

// GetMappingInstanceShapeConfigBaselineOcpuUtilizationEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceShapeConfigBaselineOcpuUtilizationEnum(val string) (InstanceShapeConfigBaselineOcpuUtilizationEnum, bool) {
	enum, ok := mappingInstanceShapeConfigBaselineOcpuUtilizationEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
