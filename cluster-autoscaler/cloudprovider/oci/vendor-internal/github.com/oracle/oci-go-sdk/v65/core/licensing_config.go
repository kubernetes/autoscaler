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

// LicensingConfig Configuration of the Operating System license.
type LicensingConfig struct {

	// Operating System type of the Configuration.
	Type LicensingConfigTypeEnum `mandatory:"false" json:"type,omitempty"`

	// License Type for the OS license.
	// * `OCI_PROVIDED` - OCI provided license (e.g. metered $/OCPU-hour).
	// * `BRING_YOUR_OWN_LICENSE` - Bring your own license.
	LicenseType LicensingConfigLicenseTypeEnum `mandatory:"false" json:"licenseType,omitempty"`

	// The Operating System version of the license config.
	OsVersion *string `mandatory:"false" json:"osVersion"`
}

func (m LicensingConfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m LicensingConfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingLicensingConfigTypeEnum(string(m.Type)); !ok && m.Type != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Type: %s. Supported values are: %s.", m.Type, strings.Join(GetLicensingConfigTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingLicensingConfigLicenseTypeEnum(string(m.LicenseType)); !ok && m.LicenseType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LicenseType: %s. Supported values are: %s.", m.LicenseType, strings.Join(GetLicensingConfigLicenseTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// LicensingConfigTypeEnum Enum with underlying type: string
type LicensingConfigTypeEnum string

// Set of constants representing the allowable values for LicensingConfigTypeEnum
const (
	LicensingConfigTypeWindows LicensingConfigTypeEnum = "WINDOWS"
)

var mappingLicensingConfigTypeEnum = map[string]LicensingConfigTypeEnum{
	"WINDOWS": LicensingConfigTypeWindows,
}

var mappingLicensingConfigTypeEnumLowerCase = map[string]LicensingConfigTypeEnum{
	"windows": LicensingConfigTypeWindows,
}

// GetLicensingConfigTypeEnumValues Enumerates the set of values for LicensingConfigTypeEnum
func GetLicensingConfigTypeEnumValues() []LicensingConfigTypeEnum {
	values := make([]LicensingConfigTypeEnum, 0)
	for _, v := range mappingLicensingConfigTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetLicensingConfigTypeEnumStringValues Enumerates the set of values in String for LicensingConfigTypeEnum
func GetLicensingConfigTypeEnumStringValues() []string {
	return []string{
		"WINDOWS",
	}
}

// GetMappingLicensingConfigTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingLicensingConfigTypeEnum(val string) (LicensingConfigTypeEnum, bool) {
	enum, ok := mappingLicensingConfigTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// LicensingConfigLicenseTypeEnum Enum with underlying type: string
type LicensingConfigLicenseTypeEnum string

// Set of constants representing the allowable values for LicensingConfigLicenseTypeEnum
const (
	LicensingConfigLicenseTypeOciProvided         LicensingConfigLicenseTypeEnum = "OCI_PROVIDED"
	LicensingConfigLicenseTypeBringYourOwnLicense LicensingConfigLicenseTypeEnum = "BRING_YOUR_OWN_LICENSE"
)

var mappingLicensingConfigLicenseTypeEnum = map[string]LicensingConfigLicenseTypeEnum{
	"OCI_PROVIDED":           LicensingConfigLicenseTypeOciProvided,
	"BRING_YOUR_OWN_LICENSE": LicensingConfigLicenseTypeBringYourOwnLicense,
}

var mappingLicensingConfigLicenseTypeEnumLowerCase = map[string]LicensingConfigLicenseTypeEnum{
	"oci_provided":           LicensingConfigLicenseTypeOciProvided,
	"bring_your_own_license": LicensingConfigLicenseTypeBringYourOwnLicense,
}

// GetLicensingConfigLicenseTypeEnumValues Enumerates the set of values for LicensingConfigLicenseTypeEnum
func GetLicensingConfigLicenseTypeEnumValues() []LicensingConfigLicenseTypeEnum {
	values := make([]LicensingConfigLicenseTypeEnum, 0)
	for _, v := range mappingLicensingConfigLicenseTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetLicensingConfigLicenseTypeEnumStringValues Enumerates the set of values in String for LicensingConfigLicenseTypeEnum
func GetLicensingConfigLicenseTypeEnumStringValues() []string {
	return []string{
		"OCI_PROVIDED",
		"BRING_YOUR_OWN_LICENSE",
	}
}

// GetMappingLicensingConfigLicenseTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingLicensingConfigLicenseTypeEnum(val string) (LicensingConfigLicenseTypeEnum, bool) {
	enum, ok := mappingLicensingConfigLicenseTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
