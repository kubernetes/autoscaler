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

// CreateVolumeGroupBackupDetails The representation of CreateVolumeGroupBackupDetails
type CreateVolumeGroupBackupDetails struct {

	// The OCID of the volume group that needs to be backed up.
	VolumeGroupId *string `mandatory:"true" json:"volumeGroupId"`

	// The OCID of the compartment that will contain the volume group
	// backup. This parameter is optional, by default backup will be created in
	// the same compartment and source volume group.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The type of backup to create. If omitted, defaults to incremental.
	Type CreateVolumeGroupBackupDetailsTypeEnum `mandatory:"false" json:"type,omitempty"`
}

func (m CreateVolumeGroupBackupDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateVolumeGroupBackupDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCreateVolumeGroupBackupDetailsTypeEnum(string(m.Type)); !ok && m.Type != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Type: %s. Supported values are: %s.", m.Type, strings.Join(GetCreateVolumeGroupBackupDetailsTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateVolumeGroupBackupDetailsTypeEnum Enum with underlying type: string
type CreateVolumeGroupBackupDetailsTypeEnum string

// Set of constants representing the allowable values for CreateVolumeGroupBackupDetailsTypeEnum
const (
	CreateVolumeGroupBackupDetailsTypeFull        CreateVolumeGroupBackupDetailsTypeEnum = "FULL"
	CreateVolumeGroupBackupDetailsTypeIncremental CreateVolumeGroupBackupDetailsTypeEnum = "INCREMENTAL"
)

var mappingCreateVolumeGroupBackupDetailsTypeEnum = map[string]CreateVolumeGroupBackupDetailsTypeEnum{
	"FULL":        CreateVolumeGroupBackupDetailsTypeFull,
	"INCREMENTAL": CreateVolumeGroupBackupDetailsTypeIncremental,
}

var mappingCreateVolumeGroupBackupDetailsTypeEnumLowerCase = map[string]CreateVolumeGroupBackupDetailsTypeEnum{
	"full":        CreateVolumeGroupBackupDetailsTypeFull,
	"incremental": CreateVolumeGroupBackupDetailsTypeIncremental,
}

// GetCreateVolumeGroupBackupDetailsTypeEnumValues Enumerates the set of values for CreateVolumeGroupBackupDetailsTypeEnum
func GetCreateVolumeGroupBackupDetailsTypeEnumValues() []CreateVolumeGroupBackupDetailsTypeEnum {
	values := make([]CreateVolumeGroupBackupDetailsTypeEnum, 0)
	for _, v := range mappingCreateVolumeGroupBackupDetailsTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateVolumeGroupBackupDetailsTypeEnumStringValues Enumerates the set of values in String for CreateVolumeGroupBackupDetailsTypeEnum
func GetCreateVolumeGroupBackupDetailsTypeEnumStringValues() []string {
	return []string{
		"FULL",
		"INCREMENTAL",
	}
}

// GetMappingCreateVolumeGroupBackupDetailsTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateVolumeGroupBackupDetailsTypeEnum(val string) (CreateVolumeGroupBackupDetailsTypeEnum, bool) {
	enum, ok := mappingCreateVolumeGroupBackupDetailsTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
