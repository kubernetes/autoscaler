// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// CreateVolumeBackupDetails The representation of CreateVolumeBackupDetails
type CreateVolumeBackupDetails struct {

	// The OCID of the volume that needs to be backed up.
	VolumeId *string `mandatory:"true" json:"volumeId"`

	// The OCID of the Vault service key which is the master encryption key for the volume backup.
	// For more information about the Vault service and encryption keys, see
	// Overview of Vault service (https://docs.oracle.com/iaas/Content/KeyManagement/Concepts/keyoverview.htm) and
	// Using Keys (https://docs.oracle.com/iaas/Content/KeyManagement/Tasks/usingkeys.htm).
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The type of backup to create. If omitted, defaults to INCREMENTAL.
	Type CreateVolumeBackupDetailsTypeEnum `mandatory:"false" json:"type,omitempty"`
}

func (m CreateVolumeBackupDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateVolumeBackupDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCreateVolumeBackupDetailsTypeEnum(string(m.Type)); !ok && m.Type != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Type: %s. Supported values are: %s.", m.Type, strings.Join(GetCreateVolumeBackupDetailsTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateVolumeBackupDetailsTypeEnum Enum with underlying type: string
type CreateVolumeBackupDetailsTypeEnum string

// Set of constants representing the allowable values for CreateVolumeBackupDetailsTypeEnum
const (
	CreateVolumeBackupDetailsTypeFull        CreateVolumeBackupDetailsTypeEnum = "FULL"
	CreateVolumeBackupDetailsTypeIncremental CreateVolumeBackupDetailsTypeEnum = "INCREMENTAL"
)

var mappingCreateVolumeBackupDetailsTypeEnum = map[string]CreateVolumeBackupDetailsTypeEnum{
	"FULL":        CreateVolumeBackupDetailsTypeFull,
	"INCREMENTAL": CreateVolumeBackupDetailsTypeIncremental,
}

var mappingCreateVolumeBackupDetailsTypeEnumLowerCase = map[string]CreateVolumeBackupDetailsTypeEnum{
	"full":        CreateVolumeBackupDetailsTypeFull,
	"incremental": CreateVolumeBackupDetailsTypeIncremental,
}

// GetCreateVolumeBackupDetailsTypeEnumValues Enumerates the set of values for CreateVolumeBackupDetailsTypeEnum
func GetCreateVolumeBackupDetailsTypeEnumValues() []CreateVolumeBackupDetailsTypeEnum {
	values := make([]CreateVolumeBackupDetailsTypeEnum, 0)
	for _, v := range mappingCreateVolumeBackupDetailsTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateVolumeBackupDetailsTypeEnumStringValues Enumerates the set of values in String for CreateVolumeBackupDetailsTypeEnum
func GetCreateVolumeBackupDetailsTypeEnumStringValues() []string {
	return []string{
		"FULL",
		"INCREMENTAL",
	}
}

// GetMappingCreateVolumeBackupDetailsTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateVolumeBackupDetailsTypeEnum(val string) (CreateVolumeBackupDetailsTypeEnum, bool) {
	enum, ok := mappingCreateVolumeBackupDetailsTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
