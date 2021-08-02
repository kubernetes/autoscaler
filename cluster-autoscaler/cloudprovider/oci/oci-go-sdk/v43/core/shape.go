// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
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

	// Whether live migration is supported for this shape.
	IsLiveMigrationSupported *bool `mandatory:"false" json:"isLiveMigrationSupported"`

	OcpuOptions *ShapeOcpuOptions `mandatory:"false" json:"ocpuOptions"`

	MemoryOptions *ShapeMemoryOptions `mandatory:"false" json:"memoryOptions"`

	NetworkingBandwidthOptions *ShapeNetworkingBandwidthOptions `mandatory:"false" json:"networkingBandwidthOptions"`

	MaxVnicAttachmentOptions *ShapeMaxVnicAttachmentOptions `mandatory:"false" json:"maxVnicAttachmentOptions"`
}

func (m Shape) String() string {
	return common.PointerString(m)
}

// ShapeBaselineOcpuUtilizationsEnum Enum with underlying type: string
type ShapeBaselineOcpuUtilizationsEnum string

// Set of constants representing the allowable values for ShapeBaselineOcpuUtilizationsEnum
const (
	ShapeBaselineOcpuUtilizations8 ShapeBaselineOcpuUtilizationsEnum = "BASELINE_1_8"
	ShapeBaselineOcpuUtilizations2 ShapeBaselineOcpuUtilizationsEnum = "BASELINE_1_2"
	ShapeBaselineOcpuUtilizations1 ShapeBaselineOcpuUtilizationsEnum = "BASELINE_1_1"
)

var mappingShapeBaselineOcpuUtilizations = map[string]ShapeBaselineOcpuUtilizationsEnum{
	"BASELINE_1_8": ShapeBaselineOcpuUtilizations8,
	"BASELINE_1_2": ShapeBaselineOcpuUtilizations2,
	"BASELINE_1_1": ShapeBaselineOcpuUtilizations1,
}

// GetShapeBaselineOcpuUtilizationsEnumValues Enumerates the set of values for ShapeBaselineOcpuUtilizationsEnum
func GetShapeBaselineOcpuUtilizationsEnumValues() []ShapeBaselineOcpuUtilizationsEnum {
	values := make([]ShapeBaselineOcpuUtilizationsEnum, 0)
	for _, v := range mappingShapeBaselineOcpuUtilizations {
		values = append(values, v)
	}
	return values
}
