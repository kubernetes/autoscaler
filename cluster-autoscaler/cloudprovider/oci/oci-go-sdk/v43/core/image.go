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

// Image A boot disk image for launching an instance. For more information, see
// Overview of the Compute Service (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
type Image struct {

	// The OCID of the compartment containing the instance you want to use as the basis for the image.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Whether instances launched with this image can be used to create new images.
	// For example, you cannot create an image of an Oracle Database instance.
	// Example: `true`
	CreateImageAllowed *bool `mandatory:"true" json:"createImageAllowed"`

	// The OCID of the image.
	Id *string `mandatory:"true" json:"id"`

	LifecycleState ImageLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The image's operating system.
	// Example: `Oracle Linux`
	OperatingSystem *string `mandatory:"true" json:"operatingSystem"`

	// The image's operating system version.
	// Example: `7.2`
	OperatingSystemVersion *string `mandatory:"true" json:"operatingSystemVersion"`

	// The date and time the image was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID of the image originally used to launch the instance.
	BaseImageId *string `mandatory:"false" json:"baseImageId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name for the image. It does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	// You cannot use a platform image name as a custom image name.
	// Example: `My custom Oracle Linux image`
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Specifies the configuration mode for launching virtual machine (VM) instances. The configuration modes are:
	// * `NATIVE` - VM instances launch with iSCSI boot and VFIO devices. The default value for platform images.
	// * `EMULATED` - VM instances launch with emulated devices, such as the E1000 network driver and emulated SCSI disk controller.
	// * `PARAVIRTUALIZED` - VM instances launch with paravirtualized devices using VirtIO drivers.
	// * `CUSTOM` - VM instances launch with custom configuration settings specified in the `LaunchOptions` parameter.
	LaunchMode ImageLaunchModeEnum `mandatory:"false" json:"launchMode,omitempty"`

	LaunchOptions *LaunchOptions `mandatory:"false" json:"launchOptions"`

	AgentFeatures *InstanceAgentFeatures `mandatory:"false" json:"agentFeatures"`

	// The listing type of the image. The default value is "NONE".
	ListingType ImageListingTypeEnum `mandatory:"false" json:"listingType,omitempty"`

	// The boot volume size for an instance launched from this image (1 MB = 1,048,576 bytes).
	// Note this is not the same as the size of the image when it was exported or the actual size of the image.
	// Example: `47694`
	SizeInMBs *int64 `mandatory:"false" json:"sizeInMBs"`

	// The size of the internal storage for this image that is subject to billing (1 GB = 1,073,741,824 bytes).
	// Example: `100`
	BillableSizeInGBs *int64 `mandatory:"false" json:"billableSizeInGBs"`
}

func (m Image) String() string {
	return common.PointerString(m)
}

// ImageLaunchModeEnum Enum with underlying type: string
type ImageLaunchModeEnum string

// Set of constants representing the allowable values for ImageLaunchModeEnum
const (
	ImageLaunchModeNative          ImageLaunchModeEnum = "NATIVE"
	ImageLaunchModeEmulated        ImageLaunchModeEnum = "EMULATED"
	ImageLaunchModeParavirtualized ImageLaunchModeEnum = "PARAVIRTUALIZED"
	ImageLaunchModeCustom          ImageLaunchModeEnum = "CUSTOM"
)

var mappingImageLaunchMode = map[string]ImageLaunchModeEnum{
	"NATIVE":          ImageLaunchModeNative,
	"EMULATED":        ImageLaunchModeEmulated,
	"PARAVIRTUALIZED": ImageLaunchModeParavirtualized,
	"CUSTOM":          ImageLaunchModeCustom,
}

// GetImageLaunchModeEnumValues Enumerates the set of values for ImageLaunchModeEnum
func GetImageLaunchModeEnumValues() []ImageLaunchModeEnum {
	values := make([]ImageLaunchModeEnum, 0)
	for _, v := range mappingImageLaunchMode {
		values = append(values, v)
	}
	return values
}

// ImageLifecycleStateEnum Enum with underlying type: string
type ImageLifecycleStateEnum string

// Set of constants representing the allowable values for ImageLifecycleStateEnum
const (
	ImageLifecycleStateProvisioning ImageLifecycleStateEnum = "PROVISIONING"
	ImageLifecycleStateImporting    ImageLifecycleStateEnum = "IMPORTING"
	ImageLifecycleStateAvailable    ImageLifecycleStateEnum = "AVAILABLE"
	ImageLifecycleStateExporting    ImageLifecycleStateEnum = "EXPORTING"
	ImageLifecycleStateDisabled     ImageLifecycleStateEnum = "DISABLED"
	ImageLifecycleStateDeleted      ImageLifecycleStateEnum = "DELETED"
)

var mappingImageLifecycleState = map[string]ImageLifecycleStateEnum{
	"PROVISIONING": ImageLifecycleStateProvisioning,
	"IMPORTING":    ImageLifecycleStateImporting,
	"AVAILABLE":    ImageLifecycleStateAvailable,
	"EXPORTING":    ImageLifecycleStateExporting,
	"DISABLED":     ImageLifecycleStateDisabled,
	"DELETED":      ImageLifecycleStateDeleted,
}

// GetImageLifecycleStateEnumValues Enumerates the set of values for ImageLifecycleStateEnum
func GetImageLifecycleStateEnumValues() []ImageLifecycleStateEnum {
	values := make([]ImageLifecycleStateEnum, 0)
	for _, v := range mappingImageLifecycleState {
		values = append(values, v)
	}
	return values
}

// ImageListingTypeEnum Enum with underlying type: string
type ImageListingTypeEnum string

// Set of constants representing the allowable values for ImageListingTypeEnum
const (
	ImageListingTypeCommunity ImageListingTypeEnum = "COMMUNITY"
	ImageListingTypeNone      ImageListingTypeEnum = "NONE"
)

var mappingImageListingType = map[string]ImageListingTypeEnum{
	"COMMUNITY": ImageListingTypeCommunity,
	"NONE":      ImageListingTypeNone,
}

// GetImageListingTypeEnumValues Enumerates the set of values for ImageListingTypeEnum
func GetImageListingTypeEnumValues() []ImageListingTypeEnum {
	values := make([]ImageListingTypeEnum, 0)
	for _, v := range mappingImageListingType {
		values = append(values, v)
	}
	return values
}
