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

// UpdateInstanceLicensingConfig The target license config to be updated on the instance.
type UpdateInstanceLicensingConfig interface {

	// License Type for the OS license.
	// * `OCI_PROVIDED` - OCI provided license (e.g. metered $/OCPU-hour).
	// * `BRING_YOUR_OWN_LICENSE` - Bring your own license.
	GetLicenseType() UpdateInstanceLicensingConfigLicenseTypeEnum
}

type updateinstancelicensingconfig struct {
	JsonData    []byte
	LicenseType UpdateInstanceLicensingConfigLicenseTypeEnum `mandatory:"false" json:"licenseType,omitempty"`
	Type        string                                       `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *updateinstancelicensingconfig) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerupdateinstancelicensingconfig updateinstancelicensingconfig
	s := struct {
		Model Unmarshalerupdateinstancelicensingconfig
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
func (m *updateinstancelicensingconfig) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "WINDOWS":
		mm := UpdateInstanceWindowsLicensingConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Received unsupported enum value for UpdateInstanceLicensingConfig: %s.", m.Type)
		return *m, nil
	}
}

// GetLicenseType returns LicenseType
func (m updateinstancelicensingconfig) GetLicenseType() UpdateInstanceLicensingConfigLicenseTypeEnum {
	return m.LicenseType
}

func (m updateinstancelicensingconfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m updateinstancelicensingconfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingUpdateInstanceLicensingConfigLicenseTypeEnum(string(m.LicenseType)); !ok && m.LicenseType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LicenseType: %s. Supported values are: %s.", m.LicenseType, strings.Join(GetUpdateInstanceLicensingConfigLicenseTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UpdateInstanceLicensingConfigLicenseTypeEnum Enum with underlying type: string
type UpdateInstanceLicensingConfigLicenseTypeEnum string

// Set of constants representing the allowable values for UpdateInstanceLicensingConfigLicenseTypeEnum
const (
	UpdateInstanceLicensingConfigLicenseTypeOciProvided         UpdateInstanceLicensingConfigLicenseTypeEnum = "OCI_PROVIDED"
	UpdateInstanceLicensingConfigLicenseTypeBringYourOwnLicense UpdateInstanceLicensingConfigLicenseTypeEnum = "BRING_YOUR_OWN_LICENSE"
)

var mappingUpdateInstanceLicensingConfigLicenseTypeEnum = map[string]UpdateInstanceLicensingConfigLicenseTypeEnum{
	"OCI_PROVIDED":           UpdateInstanceLicensingConfigLicenseTypeOciProvided,
	"BRING_YOUR_OWN_LICENSE": UpdateInstanceLicensingConfigLicenseTypeBringYourOwnLicense,
}

var mappingUpdateInstanceLicensingConfigLicenseTypeEnumLowerCase = map[string]UpdateInstanceLicensingConfigLicenseTypeEnum{
	"oci_provided":           UpdateInstanceLicensingConfigLicenseTypeOciProvided,
	"bring_your_own_license": UpdateInstanceLicensingConfigLicenseTypeBringYourOwnLicense,
}

// GetUpdateInstanceLicensingConfigLicenseTypeEnumValues Enumerates the set of values for UpdateInstanceLicensingConfigLicenseTypeEnum
func GetUpdateInstanceLicensingConfigLicenseTypeEnumValues() []UpdateInstanceLicensingConfigLicenseTypeEnum {
	values := make([]UpdateInstanceLicensingConfigLicenseTypeEnum, 0)
	for _, v := range mappingUpdateInstanceLicensingConfigLicenseTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateInstanceLicensingConfigLicenseTypeEnumStringValues Enumerates the set of values in String for UpdateInstanceLicensingConfigLicenseTypeEnum
func GetUpdateInstanceLicensingConfigLicenseTypeEnumStringValues() []string {
	return []string{
		"OCI_PROVIDED",
		"BRING_YOUR_OWN_LICENSE",
	}
}

// GetMappingUpdateInstanceLicensingConfigLicenseTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateInstanceLicensingConfigLicenseTypeEnum(val string) (UpdateInstanceLicensingConfigLicenseTypeEnum, bool) {
	enum, ok := mappingUpdateInstanceLicensingConfigLicenseTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// UpdateInstanceLicensingConfigTypeEnum Enum with underlying type: string
type UpdateInstanceLicensingConfigTypeEnum string

// Set of constants representing the allowable values for UpdateInstanceLicensingConfigTypeEnum
const (
	UpdateInstanceLicensingConfigTypeWindows UpdateInstanceLicensingConfigTypeEnum = "WINDOWS"
)

var mappingUpdateInstanceLicensingConfigTypeEnum = map[string]UpdateInstanceLicensingConfigTypeEnum{
	"WINDOWS": UpdateInstanceLicensingConfigTypeWindows,
}

var mappingUpdateInstanceLicensingConfigTypeEnumLowerCase = map[string]UpdateInstanceLicensingConfigTypeEnum{
	"windows": UpdateInstanceLicensingConfigTypeWindows,
}

// GetUpdateInstanceLicensingConfigTypeEnumValues Enumerates the set of values for UpdateInstanceLicensingConfigTypeEnum
func GetUpdateInstanceLicensingConfigTypeEnumValues() []UpdateInstanceLicensingConfigTypeEnum {
	values := make([]UpdateInstanceLicensingConfigTypeEnum, 0)
	for _, v := range mappingUpdateInstanceLicensingConfigTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateInstanceLicensingConfigTypeEnumStringValues Enumerates the set of values in String for UpdateInstanceLicensingConfigTypeEnum
func GetUpdateInstanceLicensingConfigTypeEnumStringValues() []string {
	return []string{
		"WINDOWS",
	}
}

// GetMappingUpdateInstanceLicensingConfigTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateInstanceLicensingConfigTypeEnum(val string) (UpdateInstanceLicensingConfigTypeEnum, bool) {
	enum, ok := mappingUpdateInstanceLicensingConfigTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
