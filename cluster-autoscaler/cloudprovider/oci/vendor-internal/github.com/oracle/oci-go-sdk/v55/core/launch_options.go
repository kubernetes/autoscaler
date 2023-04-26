// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// LaunchOptions Options for tuning the compatibility and performance of VM shapes. The values that you specify override any
// default values.
type LaunchOptions struct {

	// Emulation type for the boot volume.
	// * `ISCSI` - ISCSI attached block storage device.
	// * `SCSI` - Emulated SCSI disk.
	// * `IDE` - Emulated IDE disk.
	// * `VFIO` - Direct attached Virtual Function storage. This is the default option for local data
	// volumes on platform images.
	// * `PARAVIRTUALIZED` - Paravirtualized disk. This is the default for boot volumes and remote block
	// storage volumes on platform images.
	BootVolumeType LaunchOptionsBootVolumeTypeEnum `mandatory:"false" json:"bootVolumeType,omitempty"`

	// Firmware used to boot VM. Select the option that matches your operating system.
	// * `BIOS` - Boot VM using BIOS style firmware. This is compatible with both 32 bit and 64 bit operating
	// systems that boot using MBR style bootloaders.
	// * `UEFI_64` - Boot VM using UEFI style firmware compatible with 64 bit operating systems. This is the
	// default for platform images.
	Firmware LaunchOptionsFirmwareEnum `mandatory:"false" json:"firmware,omitempty"`

	// Emulation type for the physical network interface card (NIC).
	// * `E1000` - Emulated Gigabit ethernet controller. Compatible with Linux e1000 network driver.
	// * `VFIO` - Direct attached Virtual Function network controller. This is the networking type
	// when you launch an instance using hardware-assisted (SR-IOV) networking.
	// * `PARAVIRTUALIZED` - VM instances launch with paravirtualized devices using VirtIO drivers.
	NetworkType LaunchOptionsNetworkTypeEnum `mandatory:"false" json:"networkType,omitempty"`

	// Emulation type for volume.
	// * `ISCSI` - ISCSI attached block storage device.
	// * `SCSI` - Emulated SCSI disk.
	// * `IDE` - Emulated IDE disk.
	// * `VFIO` - Direct attached Virtual Function storage. This is the default option for local data
	// volumes on platform images.
	// * `PARAVIRTUALIZED` - Paravirtualized disk. This is the default for boot volumes and remote block
	// storage volumes on platform images.
	RemoteDataVolumeType LaunchOptionsRemoteDataVolumeTypeEnum `mandatory:"false" json:"remoteDataVolumeType,omitempty"`

	// Deprecated. Instead use `isPvEncryptionInTransitEnabled` in
	// LaunchInstanceDetails.
	IsPvEncryptionInTransitEnabled *bool `mandatory:"false" json:"isPvEncryptionInTransitEnabled"`

	// Whether to enable consistent volume naming feature. Defaults to false.
	IsConsistentVolumeNamingEnabled *bool `mandatory:"false" json:"isConsistentVolumeNamingEnabled"`
}

func (m LaunchOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m LaunchOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingLaunchOptionsBootVolumeTypeEnum[string(m.BootVolumeType)]; !ok && m.BootVolumeType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BootVolumeType: %s. Supported values are: %s.", m.BootVolumeType, strings.Join(GetLaunchOptionsBootVolumeTypeEnumStringValues(), ",")))
	}
	if _, ok := mappingLaunchOptionsFirmwareEnum[string(m.Firmware)]; !ok && m.Firmware != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Firmware: %s. Supported values are: %s.", m.Firmware, strings.Join(GetLaunchOptionsFirmwareEnumStringValues(), ",")))
	}
	if _, ok := mappingLaunchOptionsNetworkTypeEnum[string(m.NetworkType)]; !ok && m.NetworkType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for NetworkType: %s. Supported values are: %s.", m.NetworkType, strings.Join(GetLaunchOptionsNetworkTypeEnumStringValues(), ",")))
	}
	if _, ok := mappingLaunchOptionsRemoteDataVolumeTypeEnum[string(m.RemoteDataVolumeType)]; !ok && m.RemoteDataVolumeType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RemoteDataVolumeType: %s. Supported values are: %s.", m.RemoteDataVolumeType, strings.Join(GetLaunchOptionsRemoteDataVolumeTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// LaunchOptionsBootVolumeTypeEnum Enum with underlying type: string
type LaunchOptionsBootVolumeTypeEnum string

// Set of constants representing the allowable values for LaunchOptionsBootVolumeTypeEnum
const (
	LaunchOptionsBootVolumeTypeIscsi           LaunchOptionsBootVolumeTypeEnum = "ISCSI"
	LaunchOptionsBootVolumeTypeScsi            LaunchOptionsBootVolumeTypeEnum = "SCSI"
	LaunchOptionsBootVolumeTypeIde             LaunchOptionsBootVolumeTypeEnum = "IDE"
	LaunchOptionsBootVolumeTypeVfio            LaunchOptionsBootVolumeTypeEnum = "VFIO"
	LaunchOptionsBootVolumeTypeParavirtualized LaunchOptionsBootVolumeTypeEnum = "PARAVIRTUALIZED"
)

var mappingLaunchOptionsBootVolumeTypeEnum = map[string]LaunchOptionsBootVolumeTypeEnum{
	"ISCSI":           LaunchOptionsBootVolumeTypeIscsi,
	"SCSI":            LaunchOptionsBootVolumeTypeScsi,
	"IDE":             LaunchOptionsBootVolumeTypeIde,
	"VFIO":            LaunchOptionsBootVolumeTypeVfio,
	"PARAVIRTUALIZED": LaunchOptionsBootVolumeTypeParavirtualized,
}

// GetLaunchOptionsBootVolumeTypeEnumValues Enumerates the set of values for LaunchOptionsBootVolumeTypeEnum
func GetLaunchOptionsBootVolumeTypeEnumValues() []LaunchOptionsBootVolumeTypeEnum {
	values := make([]LaunchOptionsBootVolumeTypeEnum, 0)
	for _, v := range mappingLaunchOptionsBootVolumeTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetLaunchOptionsBootVolumeTypeEnumStringValues Enumerates the set of values in String for LaunchOptionsBootVolumeTypeEnum
func GetLaunchOptionsBootVolumeTypeEnumStringValues() []string {
	return []string{
		"ISCSI",
		"SCSI",
		"IDE",
		"VFIO",
		"PARAVIRTUALIZED",
	}
}

// LaunchOptionsFirmwareEnum Enum with underlying type: string
type LaunchOptionsFirmwareEnum string

// Set of constants representing the allowable values for LaunchOptionsFirmwareEnum
const (
	LaunchOptionsFirmwareBios   LaunchOptionsFirmwareEnum = "BIOS"
	LaunchOptionsFirmwareUefi64 LaunchOptionsFirmwareEnum = "UEFI_64"
)

var mappingLaunchOptionsFirmwareEnum = map[string]LaunchOptionsFirmwareEnum{
	"BIOS":    LaunchOptionsFirmwareBios,
	"UEFI_64": LaunchOptionsFirmwareUefi64,
}

// GetLaunchOptionsFirmwareEnumValues Enumerates the set of values for LaunchOptionsFirmwareEnum
func GetLaunchOptionsFirmwareEnumValues() []LaunchOptionsFirmwareEnum {
	values := make([]LaunchOptionsFirmwareEnum, 0)
	for _, v := range mappingLaunchOptionsFirmwareEnum {
		values = append(values, v)
	}
	return values
}

// GetLaunchOptionsFirmwareEnumStringValues Enumerates the set of values in String for LaunchOptionsFirmwareEnum
func GetLaunchOptionsFirmwareEnumStringValues() []string {
	return []string{
		"BIOS",
		"UEFI_64",
	}
}

// LaunchOptionsNetworkTypeEnum Enum with underlying type: string
type LaunchOptionsNetworkTypeEnum string

// Set of constants representing the allowable values for LaunchOptionsNetworkTypeEnum
const (
	LaunchOptionsNetworkTypeE1000           LaunchOptionsNetworkTypeEnum = "E1000"
	LaunchOptionsNetworkTypeVfio            LaunchOptionsNetworkTypeEnum = "VFIO"
	LaunchOptionsNetworkTypeParavirtualized LaunchOptionsNetworkTypeEnum = "PARAVIRTUALIZED"
)

var mappingLaunchOptionsNetworkTypeEnum = map[string]LaunchOptionsNetworkTypeEnum{
	"E1000":           LaunchOptionsNetworkTypeE1000,
	"VFIO":            LaunchOptionsNetworkTypeVfio,
	"PARAVIRTUALIZED": LaunchOptionsNetworkTypeParavirtualized,
}

// GetLaunchOptionsNetworkTypeEnumValues Enumerates the set of values for LaunchOptionsNetworkTypeEnum
func GetLaunchOptionsNetworkTypeEnumValues() []LaunchOptionsNetworkTypeEnum {
	values := make([]LaunchOptionsNetworkTypeEnum, 0)
	for _, v := range mappingLaunchOptionsNetworkTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetLaunchOptionsNetworkTypeEnumStringValues Enumerates the set of values in String for LaunchOptionsNetworkTypeEnum
func GetLaunchOptionsNetworkTypeEnumStringValues() []string {
	return []string{
		"E1000",
		"VFIO",
		"PARAVIRTUALIZED",
	}
}

// LaunchOptionsRemoteDataVolumeTypeEnum Enum with underlying type: string
type LaunchOptionsRemoteDataVolumeTypeEnum string

// Set of constants representing the allowable values for LaunchOptionsRemoteDataVolumeTypeEnum
const (
	LaunchOptionsRemoteDataVolumeTypeIscsi           LaunchOptionsRemoteDataVolumeTypeEnum = "ISCSI"
	LaunchOptionsRemoteDataVolumeTypeScsi            LaunchOptionsRemoteDataVolumeTypeEnum = "SCSI"
	LaunchOptionsRemoteDataVolumeTypeIde             LaunchOptionsRemoteDataVolumeTypeEnum = "IDE"
	LaunchOptionsRemoteDataVolumeTypeVfio            LaunchOptionsRemoteDataVolumeTypeEnum = "VFIO"
	LaunchOptionsRemoteDataVolumeTypeParavirtualized LaunchOptionsRemoteDataVolumeTypeEnum = "PARAVIRTUALIZED"
)

var mappingLaunchOptionsRemoteDataVolumeTypeEnum = map[string]LaunchOptionsRemoteDataVolumeTypeEnum{
	"ISCSI":           LaunchOptionsRemoteDataVolumeTypeIscsi,
	"SCSI":            LaunchOptionsRemoteDataVolumeTypeScsi,
	"IDE":             LaunchOptionsRemoteDataVolumeTypeIde,
	"VFIO":            LaunchOptionsRemoteDataVolumeTypeVfio,
	"PARAVIRTUALIZED": LaunchOptionsRemoteDataVolumeTypeParavirtualized,
}

// GetLaunchOptionsRemoteDataVolumeTypeEnumValues Enumerates the set of values for LaunchOptionsRemoteDataVolumeTypeEnum
func GetLaunchOptionsRemoteDataVolumeTypeEnumValues() []LaunchOptionsRemoteDataVolumeTypeEnum {
	values := make([]LaunchOptionsRemoteDataVolumeTypeEnum, 0)
	for _, v := range mappingLaunchOptionsRemoteDataVolumeTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetLaunchOptionsRemoteDataVolumeTypeEnumStringValues Enumerates the set of values in String for LaunchOptionsRemoteDataVolumeTypeEnum
func GetLaunchOptionsRemoteDataVolumeTypeEnumStringValues() []string {
	return []string{
		"ISCSI",
		"SCSI",
		"IDE",
		"VFIO",
		"PARAVIRTUALIZED",
	}
}
