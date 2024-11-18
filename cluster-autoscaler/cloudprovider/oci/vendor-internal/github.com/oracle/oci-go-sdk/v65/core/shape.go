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

// Shape A compute instance shape that can be used in LaunchInstance.
// For more information, see Overview of the Compute Service (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm) and
// Compute Shapes (https://docs.cloud.oracle.com/iaas/Content/Compute/References/computeshapes.htm).
type Shape struct {

	// The name of the shape. You can enumerate all available shapes by calling
	// ListShapes.
	Shape *string `mandatory:"true" json:"shape"`

	// For a subcore burstable VM, the supported baseline OCPU utilization for instances that use this shape.
	BaselineOcpuUtilizations []ShapeBaselineOcpuUtilizationsEnum `mandatory:"false" json:"baselineOcpuUtilizations,omitempty"`

	// For a subcore burstable VM, the minimum total baseline OCPUs required. The total baseline OCPUs is equal to
	// baselineOcpuUtilization chosen multiplied by the number of OCPUs chosen.
	MinTotalBaselineOcpusRequired *float32 `mandatory:"false" json:"minTotalBaselineOcpusRequired"`

	// A short description of the shape's processor (CPU).
	ProcessorDescription *string `mandatory:"false" json:"processorDescription"`

	// The default number of OCPUs available for this shape.
	Ocpus *float32 `mandatory:"false" json:"ocpus"`

	// The default amount of memory available for this shape, in gigabytes.
	MemoryInGBs *float32 `mandatory:"false" json:"memoryInGBs"`

	// The number of physical network interface card (NIC) ports available for this shape.
	NetworkPorts *int `mandatory:"false" json:"networkPorts"`

	// The networking bandwidth available for this shape, in gigabits per second.
	NetworkingBandwidthInGbps *float32 `mandatory:"false" json:"networkingBandwidthInGbps"`

	// The maximum number of VNIC attachments available for this shape.
	MaxVnicAttachments *int `mandatory:"false" json:"maxVnicAttachments"`

	// The number of GPUs available for this shape.
	Gpus *int `mandatory:"false" json:"gpus"`

	// A short description of the graphics processing unit (GPU) available for this shape.
	// If the shape does not have any GPUs, this field is `null`.
	GpuDescription *string `mandatory:"false" json:"gpuDescription"`

	// The number of local disks available for this shape.
	LocalDisks *int `mandatory:"false" json:"localDisks"`

	// The aggregate size of the local disks available for this shape, in gigabytes.
	// If the shape does not have any local disks, this field is `null`.
	LocalDisksTotalSizeInGBs *float32 `mandatory:"false" json:"localDisksTotalSizeInGBs"`

	// A short description of the local disks available for this shape.
	// If the shape does not have any local disks, this field is `null`.
	LocalDiskDescription *string `mandatory:"false" json:"localDiskDescription"`

	// The number of networking ports available for the remote direct memory access (RDMA) network between nodes in
	// a high performance computing (HPC) cluster network. If the shape does not support cluster networks, this
	// value is `0`.
	RdmaPorts *int `mandatory:"false" json:"rdmaPorts"`

	// The networking bandwidth available for the remote direct memory access (RDMA) network for this shape, in
	// gigabits per second.
	RdmaBandwidthInGbps *int `mandatory:"false" json:"rdmaBandwidthInGbps"`

	// Whether live migration is supported for this shape.
	IsLiveMigrationSupported *bool `mandatory:"false" json:"isLiveMigrationSupported"`

	OcpuOptions *ShapeOcpuOptions `mandatory:"false" json:"ocpuOptions"`

	MemoryOptions *ShapeMemoryOptions `mandatory:"false" json:"memoryOptions"`

	NetworkingBandwidthOptions *ShapeNetworkingBandwidthOptions `mandatory:"false" json:"networkingBandwidthOptions"`

	MaxVnicAttachmentOptions *ShapeMaxVnicAttachmentOptions `mandatory:"false" json:"maxVnicAttachmentOptions"`

	PlatformConfigOptions *ShapePlatformConfigOptions `mandatory:"false" json:"platformConfigOptions"`

	// Whether billing continues when the instances that use this shape are in the stopped state.
	IsBilledForStoppedInstance *bool `mandatory:"false" json:"isBilledForStoppedInstance"`

	// How instances that use this shape are charged.
	BillingType ShapeBillingTypeEnum `mandatory:"false" json:"billingType,omitempty"`

	// The list of of compartment quotas for the shape.
	QuotaNames []string `mandatory:"false" json:"quotaNames"`

	// Whether the shape supports creating subcore or burstable instances. A burstable instance (https://docs.cloud.oracle.com/iaas/Content/Compute/References/burstable-instances.htm)
	// is a virtual machine (VM) instance that provides a baseline level of CPU performance with the ability to burst to a higher level to support occasional
	// spikes in usage.
	IsSubcore *bool `mandatory:"false" json:"isSubcore"`

	// Whether the shape supports creating flexible instances. A flexible shape (https://docs.cloud.oracle.com/iaas/Content/Compute/References/computeshapes.htm#flexible)
	// is a shape that lets you customize the number of OCPUs and the amount of memory when launching or resizing your instance.
	IsFlexible *bool `mandatory:"false" json:"isFlexible"`

	// The list of compatible shapes that this shape can be changed to. For more information,
	// see Changing the Shape of an Instance (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/resizinginstances.htm).
	ResizeCompatibleShapes []string `mandatory:"false" json:"resizeCompatibleShapes"`

	// The list of shapes and shape details (if applicable) that Oracle recommends that you use as an alternative to the current shape.
	RecommendedAlternatives []ShapeAlternativeObject `mandatory:"false" json:"recommendedAlternatives"`
}

func (m Shape) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m Shape) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	for _, val := range m.BaselineOcpuUtilizations {
		if _, ok := GetMappingShapeBaselineOcpuUtilizationsEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BaselineOcpuUtilizations: %s. Supported values are: %s.", val, strings.Join(GetShapeBaselineOcpuUtilizationsEnumStringValues(), ",")))
		}
	}

	if _, ok := GetMappingShapeBillingTypeEnum(string(m.BillingType)); !ok && m.BillingType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BillingType: %s. Supported values are: %s.", m.BillingType, strings.Join(GetShapeBillingTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ShapeBaselineOcpuUtilizationsEnum Enum with underlying type: string
type ShapeBaselineOcpuUtilizationsEnum string

// Set of constants representing the allowable values for ShapeBaselineOcpuUtilizationsEnum
const (
	ShapeBaselineOcpuUtilizations8 ShapeBaselineOcpuUtilizationsEnum = "BASELINE_1_8"
	ShapeBaselineOcpuUtilizations2 ShapeBaselineOcpuUtilizationsEnum = "BASELINE_1_2"
	ShapeBaselineOcpuUtilizations1 ShapeBaselineOcpuUtilizationsEnum = "BASELINE_1_1"
)

var mappingShapeBaselineOcpuUtilizationsEnum = map[string]ShapeBaselineOcpuUtilizationsEnum{
	"BASELINE_1_8": ShapeBaselineOcpuUtilizations8,
	"BASELINE_1_2": ShapeBaselineOcpuUtilizations2,
	"BASELINE_1_1": ShapeBaselineOcpuUtilizations1,
}

var mappingShapeBaselineOcpuUtilizationsEnumLowerCase = map[string]ShapeBaselineOcpuUtilizationsEnum{
	"baseline_1_8": ShapeBaselineOcpuUtilizations8,
	"baseline_1_2": ShapeBaselineOcpuUtilizations2,
	"baseline_1_1": ShapeBaselineOcpuUtilizations1,
}

// GetShapeBaselineOcpuUtilizationsEnumValues Enumerates the set of values for ShapeBaselineOcpuUtilizationsEnum
func GetShapeBaselineOcpuUtilizationsEnumValues() []ShapeBaselineOcpuUtilizationsEnum {
	values := make([]ShapeBaselineOcpuUtilizationsEnum, 0)
	for _, v := range mappingShapeBaselineOcpuUtilizationsEnum {
		values = append(values, v)
	}
	return values
}

// GetShapeBaselineOcpuUtilizationsEnumStringValues Enumerates the set of values in String for ShapeBaselineOcpuUtilizationsEnum
func GetShapeBaselineOcpuUtilizationsEnumStringValues() []string {
	return []string{
		"BASELINE_1_8",
		"BASELINE_1_2",
		"BASELINE_1_1",
	}
}

// GetMappingShapeBaselineOcpuUtilizationsEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingShapeBaselineOcpuUtilizationsEnum(val string) (ShapeBaselineOcpuUtilizationsEnum, bool) {
	enum, ok := mappingShapeBaselineOcpuUtilizationsEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ShapeBillingTypeEnum Enum with underlying type: string
type ShapeBillingTypeEnum string

// Set of constants representing the allowable values for ShapeBillingTypeEnum
const (
	ShapeBillingTypeAlwaysFree  ShapeBillingTypeEnum = "ALWAYS_FREE"
	ShapeBillingTypeLimitedFree ShapeBillingTypeEnum = "LIMITED_FREE"
	ShapeBillingTypePaid        ShapeBillingTypeEnum = "PAID"
)

var mappingShapeBillingTypeEnum = map[string]ShapeBillingTypeEnum{
	"ALWAYS_FREE":  ShapeBillingTypeAlwaysFree,
	"LIMITED_FREE": ShapeBillingTypeLimitedFree,
	"PAID":         ShapeBillingTypePaid,
}

var mappingShapeBillingTypeEnumLowerCase = map[string]ShapeBillingTypeEnum{
	"always_free":  ShapeBillingTypeAlwaysFree,
	"limited_free": ShapeBillingTypeLimitedFree,
	"paid":         ShapeBillingTypePaid,
}

// GetShapeBillingTypeEnumValues Enumerates the set of values for ShapeBillingTypeEnum
func GetShapeBillingTypeEnumValues() []ShapeBillingTypeEnum {
	values := make([]ShapeBillingTypeEnum, 0)
	for _, v := range mappingShapeBillingTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetShapeBillingTypeEnumStringValues Enumerates the set of values in String for ShapeBillingTypeEnum
func GetShapeBillingTypeEnumStringValues() []string {
	return []string{
		"ALWAYS_FREE",
		"LIMITED_FREE",
		"PAID",
	}
}

// GetMappingShapeBillingTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingShapeBillingTypeEnum(val string) (ShapeBillingTypeEnum, bool) {
	enum, ok := mappingShapeBillingTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
