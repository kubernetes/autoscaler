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

// AppCatalogListingResourceVersion Listing Resource Version
type AppCatalogListingResourceVersion struct {

	// The OCID of the listing this resource version belongs to.
	ListingId *string `mandatory:"false" json:"listingId"`

	// Date and time the listing resource version was published, in RFC3339 (https://tools.ietf.org/html/rfc3339) format.
	// Example: `2018-03-20T12:32:53.532Z`
	TimePublished *common.SDKTime `mandatory:"false" json:"timePublished"`

	// OCID of the listing resource.
	ListingResourceId *string `mandatory:"false" json:"listingResourceId"`

	// Resource Version.
	ListingResourceVersion *string `mandatory:"false" json:"listingResourceVersion"`

	// List of regions that this listing resource version is available.
	// For information about regions, see
	// Regions and Availability Domains (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/regions.htm).
	// Example: `["us-ashburn-1", "us-phoenix-1"]`
	AvailableRegions []string `mandatory:"false" json:"availableRegions"`

	// Array of shapes compatible with this resource.
	// You can enumerate all available shapes by calling ListShapes.
	// Example: `["VM.Standard1.1", "VM.Standard1.2"]`
	CompatibleShapes []string `mandatory:"false" json:"compatibleShapes"`

	// List of accessible ports for instances launched with this listing resource version.
	AccessiblePorts []int `mandatory:"false" json:"accessiblePorts"`

	// Allowed actions for the listing resource.
	AllowedActions []AppCatalogListingResourceVersionAllowedActionsEnum `mandatory:"false" json:"allowedActions,omitempty"`
}

func (m AppCatalogListingResourceVersion) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m AppCatalogListingResourceVersion) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	for _, val := range m.AllowedActions {
		if _, ok := GetMappingAppCatalogListingResourceVersionAllowedActionsEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AllowedActions: %s. Supported values are: %s.", val, strings.Join(GetAppCatalogListingResourceVersionAllowedActionsEnumStringValues(), ",")))
		}
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// AppCatalogListingResourceVersionAllowedActionsEnum Enum with underlying type: string
type AppCatalogListingResourceVersionAllowedActionsEnum string

// Set of constants representing the allowable values for AppCatalogListingResourceVersionAllowedActionsEnum
const (
	AppCatalogListingResourceVersionAllowedActionsSnapshot              AppCatalogListingResourceVersionAllowedActionsEnum = "SNAPSHOT"
	AppCatalogListingResourceVersionAllowedActionsBootVolumeDetach      AppCatalogListingResourceVersionAllowedActionsEnum = "BOOT_VOLUME_DETACH"
	AppCatalogListingResourceVersionAllowedActionsPreserveBootVolume    AppCatalogListingResourceVersionAllowedActionsEnum = "PRESERVE_BOOT_VOLUME"
	AppCatalogListingResourceVersionAllowedActionsSerialConsoleAccess   AppCatalogListingResourceVersionAllowedActionsEnum = "SERIAL_CONSOLE_ACCESS"
	AppCatalogListingResourceVersionAllowedActionsBootRecovery          AppCatalogListingResourceVersionAllowedActionsEnum = "BOOT_RECOVERY"
	AppCatalogListingResourceVersionAllowedActionsBackupBootVolume      AppCatalogListingResourceVersionAllowedActionsEnum = "BACKUP_BOOT_VOLUME"
	AppCatalogListingResourceVersionAllowedActionsCaptureConsoleHistory AppCatalogListingResourceVersionAllowedActionsEnum = "CAPTURE_CONSOLE_HISTORY"
)

var mappingAppCatalogListingResourceVersionAllowedActionsEnum = map[string]AppCatalogListingResourceVersionAllowedActionsEnum{
	"SNAPSHOT":                AppCatalogListingResourceVersionAllowedActionsSnapshot,
	"BOOT_VOLUME_DETACH":      AppCatalogListingResourceVersionAllowedActionsBootVolumeDetach,
	"PRESERVE_BOOT_VOLUME":    AppCatalogListingResourceVersionAllowedActionsPreserveBootVolume,
	"SERIAL_CONSOLE_ACCESS":   AppCatalogListingResourceVersionAllowedActionsSerialConsoleAccess,
	"BOOT_RECOVERY":           AppCatalogListingResourceVersionAllowedActionsBootRecovery,
	"BACKUP_BOOT_VOLUME":      AppCatalogListingResourceVersionAllowedActionsBackupBootVolume,
	"CAPTURE_CONSOLE_HISTORY": AppCatalogListingResourceVersionAllowedActionsCaptureConsoleHistory,
}

var mappingAppCatalogListingResourceVersionAllowedActionsEnumLowerCase = map[string]AppCatalogListingResourceVersionAllowedActionsEnum{
	"snapshot":                AppCatalogListingResourceVersionAllowedActionsSnapshot,
	"boot_volume_detach":      AppCatalogListingResourceVersionAllowedActionsBootVolumeDetach,
	"preserve_boot_volume":    AppCatalogListingResourceVersionAllowedActionsPreserveBootVolume,
	"serial_console_access":   AppCatalogListingResourceVersionAllowedActionsSerialConsoleAccess,
	"boot_recovery":           AppCatalogListingResourceVersionAllowedActionsBootRecovery,
	"backup_boot_volume":      AppCatalogListingResourceVersionAllowedActionsBackupBootVolume,
	"capture_console_history": AppCatalogListingResourceVersionAllowedActionsCaptureConsoleHistory,
}

// GetAppCatalogListingResourceVersionAllowedActionsEnumValues Enumerates the set of values for AppCatalogListingResourceVersionAllowedActionsEnum
func GetAppCatalogListingResourceVersionAllowedActionsEnumValues() []AppCatalogListingResourceVersionAllowedActionsEnum {
	values := make([]AppCatalogListingResourceVersionAllowedActionsEnum, 0)
	for _, v := range mappingAppCatalogListingResourceVersionAllowedActionsEnum {
		values = append(values, v)
	}
	return values
}

// GetAppCatalogListingResourceVersionAllowedActionsEnumStringValues Enumerates the set of values in String for AppCatalogListingResourceVersionAllowedActionsEnum
func GetAppCatalogListingResourceVersionAllowedActionsEnumStringValues() []string {
	return []string{
		"SNAPSHOT",
		"BOOT_VOLUME_DETACH",
		"PRESERVE_BOOT_VOLUME",
		"SERIAL_CONSOLE_ACCESS",
		"BOOT_RECOVERY",
		"BACKUP_BOOT_VOLUME",
		"CAPTURE_CONSOLE_HISTORY",
	}
}

// GetMappingAppCatalogListingResourceVersionAllowedActionsEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingAppCatalogListingResourceVersionAllowedActionsEnum(val string) (AppCatalogListingResourceVersionAllowedActionsEnum, bool) {
	enum, ok := mappingAppCatalogListingResourceVersionAllowedActionsEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
