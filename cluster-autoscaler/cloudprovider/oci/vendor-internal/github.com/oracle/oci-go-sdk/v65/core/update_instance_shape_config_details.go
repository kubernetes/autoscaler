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

// UpdateInstanceShapeConfigDetails The shape configuration requested for the instance. If provided, the instance will be updated
// with the resources specified. In the case where some properties are missing,
// the missing values will be set to the default for the provided `shape`.
// Each shape only supports certain configurable values. If the `shape` is provided
// and the configuration values are invalid for that new `shape`, an error will be returned.
// If no `shape` is provided and the configuration values are invalid for the instance's
// existing shape, an error will be returned.
type UpdateInstanceShapeConfigDetails struct {

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
	BaselineOcpuUtilization UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum `mandatory:"false" json:"baselineOcpuUtilization,omitempty"`

	// The number of NVMe drives to be used for storage. A single drive has 6.8 TB available.
	Nvmes *int `mandatory:"false" json:"nvmes"`
}

func (m UpdateInstanceShapeConfigDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateInstanceShapeConfigDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum(string(m.BaselineOcpuUtilization)); !ok && m.BaselineOcpuUtilization != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BaselineOcpuUtilization: %s. Supported values are: %s.", m.BaselineOcpuUtilization, strings.Join(GetUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum Enum with underlying type: string
type UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum string

// Set of constants representing the allowable values for UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum
const (
	UpdateInstanceShapeConfigDetailsBaselineOcpuUtilization8 UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum = "BASELINE_1_8"
	UpdateInstanceShapeConfigDetailsBaselineOcpuUtilization2 UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum = "BASELINE_1_2"
	UpdateInstanceShapeConfigDetailsBaselineOcpuUtilization1 UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum = "BASELINE_1_1"
)

var mappingUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum = map[string]UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum{
	"BASELINE_1_8": UpdateInstanceShapeConfigDetailsBaselineOcpuUtilization8,
	"BASELINE_1_2": UpdateInstanceShapeConfigDetailsBaselineOcpuUtilization2,
	"BASELINE_1_1": UpdateInstanceShapeConfigDetailsBaselineOcpuUtilization1,
}

var mappingUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumLowerCase = map[string]UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum{
	"baseline_1_8": UpdateInstanceShapeConfigDetailsBaselineOcpuUtilization8,
	"baseline_1_2": UpdateInstanceShapeConfigDetailsBaselineOcpuUtilization2,
	"baseline_1_1": UpdateInstanceShapeConfigDetailsBaselineOcpuUtilization1,
}

// GetUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumValues Enumerates the set of values for UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum
func GetUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumValues() []UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum {
	values := make([]UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum, 0)
	for _, v := range mappingUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumStringValues Enumerates the set of values in String for UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum
func GetUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumStringValues() []string {
	return []string{
		"BASELINE_1_8",
		"BASELINE_1_2",
		"BASELINE_1_1",
	}
}

// GetMappingUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum(val string) (UpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum, bool) {
	enum, ok := mappingUpdateInstanceShapeConfigDetailsBaselineOcpuUtilizationEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
