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

// InstanceConfigurationLaunchOptions Options for tuning the compatibility and performance of VM shapes. The values that you specify override any
// default values.
type InstanceConfigurationLaunchOptions struct {

	// Emulation type for the boot volume.
	// * `ISCSI` - ISCSI attached block storage device.
	// * `SCSI` - Emulated SCSI disk.
	// * `IDE` - Emulated IDE disk.
	// * `VFIO` - Direct attached Virtual Function storage. This is the default option for local data
	// volumes on platform images.
	// * `PARAVIRTUALIZED` - Paravirtualized disk. This is the default for boot volumes and remote block
	// storage volumes on platform images.
	BootVolumeType InstanceConfigurationLaunchOptionsBootVolumeTypeEnum `mandatory:"false" json:"bootVolumeType,omitempty"`

	// Firmware used to boot VM. Select the option that matches your operating system.
	// * `BIOS` - Boot VM using BIOS style firmware. This is compatible with both 32 bit and 64 bit operating
	// systems that boot using MBR style bootloaders.
	// * `UEFI_64` - Boot VM using UEFI style firmware compatible with 64 bit operating systems. This is the
	// default for platform images.
	Firmware InstanceConfigurationLaunchOptionsFirmwareEnum `mandatory:"false" json:"firmware,omitempty"`

	// Emulation type for the physical network interface card (NIC).
	// * `E1000` - Emulated Gigabit ethernet controller. Compatible with Linux e1000 network driver.
	// * `VFIO` - Direct attached Virtual Function network controller. This is the networking type
	// when you launch an instance using hardware-assisted (SR-IOV) networking.
	// * `PARAVIRTUALIZED` - VM instances launch with paravirtualized devices using VirtIO drivers.
	NetworkType InstanceConfigurationLaunchOptionsNetworkTypeEnum `mandatory:"false" json:"networkType,omitempty"`

	// Emulation type for volume.
	// * `ISCSI` - ISCSI attached block storage device.
	// * `SCSI` - Emulated SCSI disk.
	// * `IDE` - Emulated IDE disk.
	// * `VFIO` - Direct attached Virtual Function storage. This is the default option for local data
	// volumes on platform images.
	// * `PARAVIRTUALIZED` - Paravirtualized disk. This is the default for boot volumes and remote block
	// storage volumes on platform images.
	RemoteDataVolumeType InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum `mandatory:"false" json:"remoteDataVolumeType,omitempty"`

	// Deprecated. Instead use `isPvEncryptionInTransitEnabled` in
	// InstanceConfigurationLaunchInstanceDetails.
	IsPvEncryptionInTransitEnabled *bool `mandatory:"false" json:"isPvEncryptionInTransitEnabled"`

	// Whether to enable consistent volume naming feature. Defaults to false.
	IsConsistentVolumeNamingEnabled *bool `mandatory:"false" json:"isConsistentVolumeNamingEnabled"`
}

func (m InstanceConfigurationLaunchOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstanceConfigurationLaunchOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingInstanceConfigurationLaunchOptionsBootVolumeTypeEnum(string(m.BootVolumeType)); !ok && m.BootVolumeType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BootVolumeType: %s. Supported values are: %s.", m.BootVolumeType, strings.Join(GetInstanceConfigurationLaunchOptionsBootVolumeTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingInstanceConfigurationLaunchOptionsFirmwareEnum(string(m.Firmware)); !ok && m.Firmware != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Firmware: %s. Supported values are: %s.", m.Firmware, strings.Join(GetInstanceConfigurationLaunchOptionsFirmwareEnumStringValues(), ",")))
	}
	if _, ok := GetMappingInstanceConfigurationLaunchOptionsNetworkTypeEnum(string(m.NetworkType)); !ok && m.NetworkType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for NetworkType: %s. Supported values are: %s.", m.NetworkType, strings.Join(GetInstanceConfigurationLaunchOptionsNetworkTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum(string(m.RemoteDataVolumeType)); !ok && m.RemoteDataVolumeType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RemoteDataVolumeType: %s. Supported values are: %s.", m.RemoteDataVolumeType, strings.Join(GetInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstanceConfigurationLaunchOptionsBootVolumeTypeEnum Enum with underlying type: string
type InstanceConfigurationLaunchOptionsBootVolumeTypeEnum string

// Set of constants representing the allowable values for InstanceConfigurationLaunchOptionsBootVolumeTypeEnum
const (
	InstanceConfigurationLaunchOptionsBootVolumeTypeIscsi           InstanceConfigurationLaunchOptionsBootVolumeTypeEnum = "ISCSI"
	InstanceConfigurationLaunchOptionsBootVolumeTypeScsi            InstanceConfigurationLaunchOptionsBootVolumeTypeEnum = "SCSI"
	InstanceConfigurationLaunchOptionsBootVolumeTypeIde             InstanceConfigurationLaunchOptionsBootVolumeTypeEnum = "IDE"
	InstanceConfigurationLaunchOptionsBootVolumeTypeVfio            InstanceConfigurationLaunchOptionsBootVolumeTypeEnum = "VFIO"
	InstanceConfigurationLaunchOptionsBootVolumeTypeParavirtualized InstanceConfigurationLaunchOptionsBootVolumeTypeEnum = "PARAVIRTUALIZED"
)

var mappingInstanceConfigurationLaunchOptionsBootVolumeTypeEnum = map[string]InstanceConfigurationLaunchOptionsBootVolumeTypeEnum{
	"ISCSI":           InstanceConfigurationLaunchOptionsBootVolumeTypeIscsi,
	"SCSI":            InstanceConfigurationLaunchOptionsBootVolumeTypeScsi,
	"IDE":             InstanceConfigurationLaunchOptionsBootVolumeTypeIde,
	"VFIO":            InstanceConfigurationLaunchOptionsBootVolumeTypeVfio,
	"PARAVIRTUALIZED": InstanceConfigurationLaunchOptionsBootVolumeTypeParavirtualized,
}

var mappingInstanceConfigurationLaunchOptionsBootVolumeTypeEnumLowerCase = map[string]InstanceConfigurationLaunchOptionsBootVolumeTypeEnum{
	"iscsi":           InstanceConfigurationLaunchOptionsBootVolumeTypeIscsi,
	"scsi":            InstanceConfigurationLaunchOptionsBootVolumeTypeScsi,
	"ide":             InstanceConfigurationLaunchOptionsBootVolumeTypeIde,
	"vfio":            InstanceConfigurationLaunchOptionsBootVolumeTypeVfio,
	"paravirtualized": InstanceConfigurationLaunchOptionsBootVolumeTypeParavirtualized,
}

// GetInstanceConfigurationLaunchOptionsBootVolumeTypeEnumValues Enumerates the set of values for InstanceConfigurationLaunchOptionsBootVolumeTypeEnum
func GetInstanceConfigurationLaunchOptionsBootVolumeTypeEnumValues() []InstanceConfigurationLaunchOptionsBootVolumeTypeEnum {
	values := make([]InstanceConfigurationLaunchOptionsBootVolumeTypeEnum, 0)
	for _, v := range mappingInstanceConfigurationLaunchOptionsBootVolumeTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceConfigurationLaunchOptionsBootVolumeTypeEnumStringValues Enumerates the set of values in String for InstanceConfigurationLaunchOptionsBootVolumeTypeEnum
func GetInstanceConfigurationLaunchOptionsBootVolumeTypeEnumStringValues() []string {
	return []string{
		"ISCSI",
		"SCSI",
		"IDE",
		"VFIO",
		"PARAVIRTUALIZED",
	}
}

// GetMappingInstanceConfigurationLaunchOptionsBootVolumeTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceConfigurationLaunchOptionsBootVolumeTypeEnum(val string) (InstanceConfigurationLaunchOptionsBootVolumeTypeEnum, bool) {
	enum, ok := mappingInstanceConfigurationLaunchOptionsBootVolumeTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// InstanceConfigurationLaunchOptionsFirmwareEnum Enum with underlying type: string
type InstanceConfigurationLaunchOptionsFirmwareEnum string

// Set of constants representing the allowable values for InstanceConfigurationLaunchOptionsFirmwareEnum
const (
	InstanceConfigurationLaunchOptionsFirmwareBios   InstanceConfigurationLaunchOptionsFirmwareEnum = "BIOS"
	InstanceConfigurationLaunchOptionsFirmwareUefi64 InstanceConfigurationLaunchOptionsFirmwareEnum = "UEFI_64"
)

var mappingInstanceConfigurationLaunchOptionsFirmwareEnum = map[string]InstanceConfigurationLaunchOptionsFirmwareEnum{
	"BIOS":    InstanceConfigurationLaunchOptionsFirmwareBios,
	"UEFI_64": InstanceConfigurationLaunchOptionsFirmwareUefi64,
}

var mappingInstanceConfigurationLaunchOptionsFirmwareEnumLowerCase = map[string]InstanceConfigurationLaunchOptionsFirmwareEnum{
	"bios":    InstanceConfigurationLaunchOptionsFirmwareBios,
	"uefi_64": InstanceConfigurationLaunchOptionsFirmwareUefi64,
}

// GetInstanceConfigurationLaunchOptionsFirmwareEnumValues Enumerates the set of values for InstanceConfigurationLaunchOptionsFirmwareEnum
func GetInstanceConfigurationLaunchOptionsFirmwareEnumValues() []InstanceConfigurationLaunchOptionsFirmwareEnum {
	values := make([]InstanceConfigurationLaunchOptionsFirmwareEnum, 0)
	for _, v := range mappingInstanceConfigurationLaunchOptionsFirmwareEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceConfigurationLaunchOptionsFirmwareEnumStringValues Enumerates the set of values in String for InstanceConfigurationLaunchOptionsFirmwareEnum
func GetInstanceConfigurationLaunchOptionsFirmwareEnumStringValues() []string {
	return []string{
		"BIOS",
		"UEFI_64",
	}
}

// GetMappingInstanceConfigurationLaunchOptionsFirmwareEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceConfigurationLaunchOptionsFirmwareEnum(val string) (InstanceConfigurationLaunchOptionsFirmwareEnum, bool) {
	enum, ok := mappingInstanceConfigurationLaunchOptionsFirmwareEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// InstanceConfigurationLaunchOptionsNetworkTypeEnum Enum with underlying type: string
type InstanceConfigurationLaunchOptionsNetworkTypeEnum string

// Set of constants representing the allowable values for InstanceConfigurationLaunchOptionsNetworkTypeEnum
const (
	InstanceConfigurationLaunchOptionsNetworkTypeE1000           InstanceConfigurationLaunchOptionsNetworkTypeEnum = "E1000"
	InstanceConfigurationLaunchOptionsNetworkTypeVfio            InstanceConfigurationLaunchOptionsNetworkTypeEnum = "VFIO"
	InstanceConfigurationLaunchOptionsNetworkTypeParavirtualized InstanceConfigurationLaunchOptionsNetworkTypeEnum = "PARAVIRTUALIZED"
)

var mappingInstanceConfigurationLaunchOptionsNetworkTypeEnum = map[string]InstanceConfigurationLaunchOptionsNetworkTypeEnum{
	"E1000":           InstanceConfigurationLaunchOptionsNetworkTypeE1000,
	"VFIO":            InstanceConfigurationLaunchOptionsNetworkTypeVfio,
	"PARAVIRTUALIZED": InstanceConfigurationLaunchOptionsNetworkTypeParavirtualized,
}

var mappingInstanceConfigurationLaunchOptionsNetworkTypeEnumLowerCase = map[string]InstanceConfigurationLaunchOptionsNetworkTypeEnum{
	"e1000":           InstanceConfigurationLaunchOptionsNetworkTypeE1000,
	"vfio":            InstanceConfigurationLaunchOptionsNetworkTypeVfio,
	"paravirtualized": InstanceConfigurationLaunchOptionsNetworkTypeParavirtualized,
}

// GetInstanceConfigurationLaunchOptionsNetworkTypeEnumValues Enumerates the set of values for InstanceConfigurationLaunchOptionsNetworkTypeEnum
func GetInstanceConfigurationLaunchOptionsNetworkTypeEnumValues() []InstanceConfigurationLaunchOptionsNetworkTypeEnum {
	values := make([]InstanceConfigurationLaunchOptionsNetworkTypeEnum, 0)
	for _, v := range mappingInstanceConfigurationLaunchOptionsNetworkTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceConfigurationLaunchOptionsNetworkTypeEnumStringValues Enumerates the set of values in String for InstanceConfigurationLaunchOptionsNetworkTypeEnum
func GetInstanceConfigurationLaunchOptionsNetworkTypeEnumStringValues() []string {
	return []string{
		"E1000",
		"VFIO",
		"PARAVIRTUALIZED",
	}
}

// GetMappingInstanceConfigurationLaunchOptionsNetworkTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceConfigurationLaunchOptionsNetworkTypeEnum(val string) (InstanceConfigurationLaunchOptionsNetworkTypeEnum, bool) {
	enum, ok := mappingInstanceConfigurationLaunchOptionsNetworkTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum Enum with underlying type: string
type InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum string

// Set of constants representing the allowable values for InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum
const (
	InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeIscsi           InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum = "ISCSI"
	InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeScsi            InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum = "SCSI"
	InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeIde             InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum = "IDE"
	InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeVfio            InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum = "VFIO"
	InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeParavirtualized InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum = "PARAVIRTUALIZED"
)

var mappingInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum = map[string]InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum{
	"ISCSI":           InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeIscsi,
	"SCSI":            InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeScsi,
	"IDE":             InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeIde,
	"VFIO":            InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeVfio,
	"PARAVIRTUALIZED": InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeParavirtualized,
}

var mappingInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnumLowerCase = map[string]InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum{
	"iscsi":           InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeIscsi,
	"scsi":            InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeScsi,
	"ide":             InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeIde,
	"vfio":            InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeVfio,
	"paravirtualized": InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeParavirtualized,
}

// GetInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnumValues Enumerates the set of values for InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum
func GetInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnumValues() []InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum {
	values := make([]InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum, 0)
	for _, v := range mappingInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnumStringValues Enumerates the set of values in String for InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum
func GetInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnumStringValues() []string {
	return []string{
		"ISCSI",
		"SCSI",
		"IDE",
		"VFIO",
		"PARAVIRTUALIZED",
	}
}

// GetMappingInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum(val string) (InstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnum, bool) {
	enum, ok := mappingInstanceConfigurationLaunchOptionsRemoteDataVolumeTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
