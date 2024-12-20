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

// UpdateLaunchOptions Options for tuning the compatibility and performance of VM shapes.
type UpdateLaunchOptions struct {

	// Emulation type for the boot volume.
	// * `ISCSI` - ISCSI attached block storage device.
	// * `PARAVIRTUALIZED` - Paravirtualized disk. This is the default for boot volumes and remote block
	// storage volumes on platform images.
	// Before you change the boot volume attachment type, detach all block volumes and VNICs except for
	// the boot volume and the primary VNIC.
	// If the instance is running when you change the boot volume attachment type, it will be rebooted.
	// **Note:** Some instances might not function properly if you change the boot volume attachment type. After
	// the instance reboots and is running, connect to it. If the connection fails or the OS doesn't behave
	// as expected, the changes are not supported. Revert the instance to the original boot volume attachment type.
	BootVolumeType UpdateLaunchOptionsBootVolumeTypeEnum `mandatory:"false" json:"bootVolumeType,omitempty"`

	// Emulation type for the physical network interface card (NIC).
	// * `VFIO` - Direct attached Virtual Function network controller. This is the networking type
	// when you launch an instance using hardware-assisted (SR-IOV) networking.
	// * `PARAVIRTUALIZED` - VM instances launch with paravirtualized devices using VirtIO drivers.
	// Before you change the networking type, detach all VNICs and block volumes except for the primary
	// VNIC and the boot volume.
	// The image must have paravirtualized drivers installed. For more information, see
	// Editing an Instance (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/resizinginstances.htm).
	// If the instance is running when you change the network type, it will be rebooted.
	// **Note:** Some instances might not function properly if you change the networking type. After
	// the instance reboots and is running, connect to it. If the connection fails or the OS doesn't behave
	// as expected, the changes are not supported. Revert the instance to the original networking type.
	NetworkType UpdateLaunchOptionsNetworkTypeEnum `mandatory:"false" json:"networkType,omitempty"`

	// Whether to enable in-transit encryption for the volume's paravirtualized attachment.
	// To enable in-transit encryption for block volumes and boot volumes, this field must be set to `true`.
	// Data in transit is transferred over an internal and highly secure network. If you have specific
	// compliance requirements related to the encryption of the data while it is moving between the
	// instance and the boot volume or the block volume, you can enable in-transit encryption.
	// In-transit encryption is not enabled by default.
	// All boot volumes and block volumes are encrypted at rest.
	// For more information, see Block Volume Encryption (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm#Encrypti).
	IsPvEncryptionInTransitEnabled *bool `mandatory:"false" json:"isPvEncryptionInTransitEnabled"`
}

func (m UpdateLaunchOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateLaunchOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingUpdateLaunchOptionsBootVolumeTypeEnum(string(m.BootVolumeType)); !ok && m.BootVolumeType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BootVolumeType: %s. Supported values are: %s.", m.BootVolumeType, strings.Join(GetUpdateLaunchOptionsBootVolumeTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingUpdateLaunchOptionsNetworkTypeEnum(string(m.NetworkType)); !ok && m.NetworkType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for NetworkType: %s. Supported values are: %s.", m.NetworkType, strings.Join(GetUpdateLaunchOptionsNetworkTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UpdateLaunchOptionsBootVolumeTypeEnum Enum with underlying type: string
type UpdateLaunchOptionsBootVolumeTypeEnum string

// Set of constants representing the allowable values for UpdateLaunchOptionsBootVolumeTypeEnum
const (
	UpdateLaunchOptionsBootVolumeTypeIscsi           UpdateLaunchOptionsBootVolumeTypeEnum = "ISCSI"
	UpdateLaunchOptionsBootVolumeTypeParavirtualized UpdateLaunchOptionsBootVolumeTypeEnum = "PARAVIRTUALIZED"
)

var mappingUpdateLaunchOptionsBootVolumeTypeEnum = map[string]UpdateLaunchOptionsBootVolumeTypeEnum{
	"ISCSI":           UpdateLaunchOptionsBootVolumeTypeIscsi,
	"PARAVIRTUALIZED": UpdateLaunchOptionsBootVolumeTypeParavirtualized,
}

var mappingUpdateLaunchOptionsBootVolumeTypeEnumLowerCase = map[string]UpdateLaunchOptionsBootVolumeTypeEnum{
	"iscsi":           UpdateLaunchOptionsBootVolumeTypeIscsi,
	"paravirtualized": UpdateLaunchOptionsBootVolumeTypeParavirtualized,
}

// GetUpdateLaunchOptionsBootVolumeTypeEnumValues Enumerates the set of values for UpdateLaunchOptionsBootVolumeTypeEnum
func GetUpdateLaunchOptionsBootVolumeTypeEnumValues() []UpdateLaunchOptionsBootVolumeTypeEnum {
	values := make([]UpdateLaunchOptionsBootVolumeTypeEnum, 0)
	for _, v := range mappingUpdateLaunchOptionsBootVolumeTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateLaunchOptionsBootVolumeTypeEnumStringValues Enumerates the set of values in String for UpdateLaunchOptionsBootVolumeTypeEnum
func GetUpdateLaunchOptionsBootVolumeTypeEnumStringValues() []string {
	return []string{
		"ISCSI",
		"PARAVIRTUALIZED",
	}
}

// GetMappingUpdateLaunchOptionsBootVolumeTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateLaunchOptionsBootVolumeTypeEnum(val string) (UpdateLaunchOptionsBootVolumeTypeEnum, bool) {
	enum, ok := mappingUpdateLaunchOptionsBootVolumeTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// UpdateLaunchOptionsNetworkTypeEnum Enum with underlying type: string
type UpdateLaunchOptionsNetworkTypeEnum string

// Set of constants representing the allowable values for UpdateLaunchOptionsNetworkTypeEnum
const (
	UpdateLaunchOptionsNetworkTypeVfio            UpdateLaunchOptionsNetworkTypeEnum = "VFIO"
	UpdateLaunchOptionsNetworkTypeParavirtualized UpdateLaunchOptionsNetworkTypeEnum = "PARAVIRTUALIZED"
)

var mappingUpdateLaunchOptionsNetworkTypeEnum = map[string]UpdateLaunchOptionsNetworkTypeEnum{
	"VFIO":            UpdateLaunchOptionsNetworkTypeVfio,
	"PARAVIRTUALIZED": UpdateLaunchOptionsNetworkTypeParavirtualized,
}

var mappingUpdateLaunchOptionsNetworkTypeEnumLowerCase = map[string]UpdateLaunchOptionsNetworkTypeEnum{
	"vfio":            UpdateLaunchOptionsNetworkTypeVfio,
	"paravirtualized": UpdateLaunchOptionsNetworkTypeParavirtualized,
}

// GetUpdateLaunchOptionsNetworkTypeEnumValues Enumerates the set of values for UpdateLaunchOptionsNetworkTypeEnum
func GetUpdateLaunchOptionsNetworkTypeEnumValues() []UpdateLaunchOptionsNetworkTypeEnum {
	values := make([]UpdateLaunchOptionsNetworkTypeEnum, 0)
	for _, v := range mappingUpdateLaunchOptionsNetworkTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateLaunchOptionsNetworkTypeEnumStringValues Enumerates the set of values in String for UpdateLaunchOptionsNetworkTypeEnum
func GetUpdateLaunchOptionsNetworkTypeEnumStringValues() []string {
	return []string{
		"VFIO",
		"PARAVIRTUALIZED",
	}
}

// GetMappingUpdateLaunchOptionsNetworkTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateLaunchOptionsNetworkTypeEnum(val string) (UpdateLaunchOptionsNetworkTypeEnum, bool) {
	enum, ok := mappingUpdateLaunchOptionsNetworkTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
