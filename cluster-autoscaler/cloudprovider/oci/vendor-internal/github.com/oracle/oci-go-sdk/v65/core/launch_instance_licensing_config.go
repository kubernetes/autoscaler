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
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// LaunchInstanceLicensingConfig The license config requested for the instance.
type LaunchInstanceLicensingConfig interface {

	// License Type for the OS license.
	// * `OCI_PROVIDED` - OCI provided license (e.g. metered $/OCPU-hour).
	// * `BRING_YOUR_OWN_LICENSE` - Bring your own license.
	GetLicenseType() LaunchInstanceLicensingConfigLicenseTypeEnum
}

type launchinstancelicensingconfig struct {
	JsonData    []byte
	LicenseType LaunchInstanceLicensingConfigLicenseTypeEnum `mandatory:"false" json:"licenseType,omitempty"`
	Type        string                                       `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *launchinstancelicensingconfig) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerlaunchinstancelicensingconfig launchinstancelicensingconfig
	s := struct {
		Model Unmarshalerlaunchinstancelicensingconfig
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.LicenseType = s.Model.LicenseType
	m.Type = s.Model.Type

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *launchinstancelicensingconfig) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "WINDOWS":
		mm := LaunchInstanceWindowsLicensingConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Received unsupported enum value for LaunchInstanceLicensingConfig: %s.", m.Type)
		return *m, nil
	}
}

// GetLicenseType returns LicenseType
func (m launchinstancelicensingconfig) GetLicenseType() LaunchInstanceLicensingConfigLicenseTypeEnum {
	return m.LicenseType
}

func (m launchinstancelicensingconfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m launchinstancelicensingconfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingLaunchInstanceLicensingConfigLicenseTypeEnum(string(m.LicenseType)); !ok && m.LicenseType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LicenseType: %s. Supported values are: %s.", m.LicenseType, strings.Join(GetLaunchInstanceLicensingConfigLicenseTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// LaunchInstanceLicensingConfigLicenseTypeEnum Enum with underlying type: string
type LaunchInstanceLicensingConfigLicenseTypeEnum string

// Set of constants representing the allowable values for LaunchInstanceLicensingConfigLicenseTypeEnum
const (
	LaunchInstanceLicensingConfigLicenseTypeOciProvided         LaunchInstanceLicensingConfigLicenseTypeEnum = "OCI_PROVIDED"
	LaunchInstanceLicensingConfigLicenseTypeBringYourOwnLicense LaunchInstanceLicensingConfigLicenseTypeEnum = "BRING_YOUR_OWN_LICENSE"
)

var mappingLaunchInstanceLicensingConfigLicenseTypeEnum = map[string]LaunchInstanceLicensingConfigLicenseTypeEnum{
	"OCI_PROVIDED":           LaunchInstanceLicensingConfigLicenseTypeOciProvided,
	"BRING_YOUR_OWN_LICENSE": LaunchInstanceLicensingConfigLicenseTypeBringYourOwnLicense,
}

var mappingLaunchInstanceLicensingConfigLicenseTypeEnumLowerCase = map[string]LaunchInstanceLicensingConfigLicenseTypeEnum{
	"oci_provided":           LaunchInstanceLicensingConfigLicenseTypeOciProvided,
	"bring_your_own_license": LaunchInstanceLicensingConfigLicenseTypeBringYourOwnLicense,
}

// GetLaunchInstanceLicensingConfigLicenseTypeEnumValues Enumerates the set of values for LaunchInstanceLicensingConfigLicenseTypeEnum
func GetLaunchInstanceLicensingConfigLicenseTypeEnumValues() []LaunchInstanceLicensingConfigLicenseTypeEnum {
	values := make([]LaunchInstanceLicensingConfigLicenseTypeEnum, 0)
	for _, v := range mappingLaunchInstanceLicensingConfigLicenseTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetLaunchInstanceLicensingConfigLicenseTypeEnumStringValues Enumerates the set of values in String for LaunchInstanceLicensingConfigLicenseTypeEnum
func GetLaunchInstanceLicensingConfigLicenseTypeEnumStringValues() []string {
	return []string{
		"OCI_PROVIDED",
		"BRING_YOUR_OWN_LICENSE",
	}
}

// GetMappingLaunchInstanceLicensingConfigLicenseTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingLaunchInstanceLicensingConfigLicenseTypeEnum(val string) (LaunchInstanceLicensingConfigLicenseTypeEnum, bool) {
	enum, ok := mappingLaunchInstanceLicensingConfigLicenseTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// LaunchInstanceLicensingConfigTypeEnum Enum with underlying type: string
type LaunchInstanceLicensingConfigTypeEnum string

// Set of constants representing the allowable values for LaunchInstanceLicensingConfigTypeEnum
const (
	LaunchInstanceLicensingConfigTypeWindows LaunchInstanceLicensingConfigTypeEnum = "WINDOWS"
)

var mappingLaunchInstanceLicensingConfigTypeEnum = map[string]LaunchInstanceLicensingConfigTypeEnum{
	"WINDOWS": LaunchInstanceLicensingConfigTypeWindows,
}

var mappingLaunchInstanceLicensingConfigTypeEnumLowerCase = map[string]LaunchInstanceLicensingConfigTypeEnum{
	"windows": LaunchInstanceLicensingConfigTypeWindows,
}

// GetLaunchInstanceLicensingConfigTypeEnumValues Enumerates the set of values for LaunchInstanceLicensingConfigTypeEnum
func GetLaunchInstanceLicensingConfigTypeEnumValues() []LaunchInstanceLicensingConfigTypeEnum {
	values := make([]LaunchInstanceLicensingConfigTypeEnum, 0)
	for _, v := range mappingLaunchInstanceLicensingConfigTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetLaunchInstanceLicensingConfigTypeEnumStringValues Enumerates the set of values in String for LaunchInstanceLicensingConfigTypeEnum
func GetLaunchInstanceLicensingConfigTypeEnumStringValues() []string {
	return []string{
		"WINDOWS",
	}
}

// GetMappingLaunchInstanceLicensingConfigTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingLaunchInstanceLicensingConfigTypeEnum(val string) (LaunchInstanceLicensingConfigTypeEnum, bool) {
	enum, ok := mappingLaunchInstanceLicensingConfigTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
