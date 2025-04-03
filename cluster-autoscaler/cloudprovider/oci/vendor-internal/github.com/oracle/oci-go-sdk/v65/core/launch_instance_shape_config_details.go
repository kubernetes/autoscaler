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

// LaunchInstanceShapeConfigDetails The shape configuration requested for the instance.
// If the parameter is provided, the instance is created with the resources that you specify. If some
// properties are missing or the entire parameter is not provided, the instance is created
// with the default configuration values for the `shape` that you specify.
// Each shape only supports certain configurable values. If the values that you provide are not valid for the
// specified `shape`, an error is returned.
type LaunchInstanceShapeConfigDetails struct {

	// The total number of OCPUs available to the instance.
	Ocpus *float32 `mandatory:"false" json:"ocpus"`

	// The total number of VCPUs available to the instance. This can be used instead of OCPUs,
	// in which case the actual number of OCPUs will be calculated based on this value
	// and the actual hardware. This must be a multiple of 2.
	Vcpus *int `mandatory:"false" json:"vcpus"`

	// The total amount of memory available to the instance, in gigabytes.
	MemoryInGBs *float32 `mandatory:"false" json:"memoryInGBs"`

	// The baseline OCPU utilization for a subcore burstable VM instance. Leave this attribute blank for a
	// non-burstable instance, or explicitly specify non-burstable with `BASELINE_1_1`.
	// The following values are supported:
	// - `BASELINE_1_8` - baseline usage is 1/8 of an OCPU.
	// - `BASELINE_1_2` - baseline usage is 1/2 of an OCPU.
	// - `BASELINE_1_1` - baseline usage is an entire OCPU. This represents a non-burstable instance.
	BaselineOcpuUtilization LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum `mandatory:"false" json:"baselineOcpuUtilization,omitempty"`

	// The number of NVMe drives to be used for storage. A single drive has 6.8 TB available.
	Nvmes *int `mandatory:"false" json:"nvmes"`
}

func (m LaunchInstanceShapeConfigDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m LaunchInstanceShapeConfigDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum(string(m.BaselineOcpuUtilization)); !ok && m.BaselineOcpuUtilization != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BaselineOcpuUtilization: %s. Supported values are: %s.", m.BaselineOcpuUtilization, strings.Join(GetLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum Enum with underlying type: string
type LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum string

// Set of constants representing the allowable values for LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum
const (
	LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization8 LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum = "BASELINE_1_8"
	LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization2 LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum = "BASELINE_1_2"
	LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization1 LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum = "BASELINE_1_1"
)

var mappingLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum = map[string]LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum{
	"BASELINE_1_8": LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization8,
	"BASELINE_1_2": LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization2,
	"BASELINE_1_1": LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization1,
}

var mappingLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumLowerCase = map[string]LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum{
	"baseline_1_8": LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization8,
	"baseline_1_2": LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization2,
	"baseline_1_1": LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization1,
}

// GetLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumValues Enumerates the set of values for LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum
func GetLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumValues() []LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum {
	values := make([]LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum, 0)
	for _, v := range mappingLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum {
		values = append(values, v)
	}
	return values
}

// GetLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumStringValues Enumerates the set of values in String for LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum
func GetLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumStringValues() []string {
	return []string{
		"BASELINE_1_8",
		"BASELINE_1_2",
		"BASELINE_1_1",
	}
}

// GetMappingLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum(val string) (LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum, bool) {
	enum, ok := mappingLaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
