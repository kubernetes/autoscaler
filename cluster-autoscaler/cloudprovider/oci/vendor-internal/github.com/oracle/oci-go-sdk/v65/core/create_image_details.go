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
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// CreateImageDetails Either instanceId or imageSourceDetails must be provided in addition to other required parameters.
type CreateImageDetails struct {

	// The OCID of the compartment you want the image to be created in.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name for the image. It does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	// You cannot use a platform image name as a custom image name.
	// Example: `My Oracle Linux image`
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	ImageSourceDetails ImageSourceDetails `mandatory:"false" json:"imageSourceDetails"`

	// The OCID of the instance you want to use as the basis for the image.
	InstanceId *string `mandatory:"false" json:"instanceId"`

	// Specifies the configuration mode for launching virtual machine (VM) instances. The configuration modes are:
	// * `NATIVE` - VM instances launch with iSCSI boot and VFIO devices. The default value for platform images.
	// * `EMULATED` - VM instances launch with emulated devices, such as the E1000 network driver and emulated SCSI disk controller.
	// * `PARAVIRTUALIZED` - VM instances launch with paravirtualized devices using VirtIO drivers.
	// * `CUSTOM` - VM instances launch with custom configuration settings specified in the `LaunchOptions` parameter.
	LaunchMode CreateImageDetailsLaunchModeEnum `mandatory:"false" json:"launchMode,omitempty"`
}

func (m CreateImageDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateImageDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCreateImageDetailsLaunchModeEnum(string(m.LaunchMode)); !ok && m.LaunchMode != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LaunchMode: %s. Supported values are: %s.", m.LaunchMode, strings.Join(GetCreateImageDetailsLaunchModeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *CreateImageDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DefinedTags        map[string]map[string]interface{} `json:"definedTags"`
		DisplayName        *string                           `json:"displayName"`
		FreeformTags       map[string]string                 `json:"freeformTags"`
		ImageSourceDetails imagesourcedetails                `json:"imageSourceDetails"`
		InstanceId         *string                           `json:"instanceId"`
		LaunchMode         CreateImageDetailsLaunchModeEnum  `json:"launchMode"`
		CompartmentId      *string                           `json:"compartmentId"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DefinedTags = model.DefinedTags

	m.DisplayName = model.DisplayName

	m.FreeformTags = model.FreeformTags

	nn, e = model.ImageSourceDetails.UnmarshalPolymorphicJSON(model.ImageSourceDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.ImageSourceDetails = nn.(ImageSourceDetails)
	} else {
		m.ImageSourceDetails = nil
	}

	m.InstanceId = model.InstanceId

	m.LaunchMode = model.LaunchMode

	m.CompartmentId = model.CompartmentId

	return
}

// CreateImageDetailsLaunchModeEnum Enum with underlying type: string
type CreateImageDetailsLaunchModeEnum string

// Set of constants representing the allowable values for CreateImageDetailsLaunchModeEnum
const (
	CreateImageDetailsLaunchModeNative          CreateImageDetailsLaunchModeEnum = "NATIVE"
	CreateImageDetailsLaunchModeEmulated        CreateImageDetailsLaunchModeEnum = "EMULATED"
	CreateImageDetailsLaunchModeParavirtualized CreateImageDetailsLaunchModeEnum = "PARAVIRTUALIZED"
	CreateImageDetailsLaunchModeCustom          CreateImageDetailsLaunchModeEnum = "CUSTOM"
)

var mappingCreateImageDetailsLaunchModeEnum = map[string]CreateImageDetailsLaunchModeEnum{
	"NATIVE":          CreateImageDetailsLaunchModeNative,
	"EMULATED":        CreateImageDetailsLaunchModeEmulated,
	"PARAVIRTUALIZED": CreateImageDetailsLaunchModeParavirtualized,
	"CUSTOM":          CreateImageDetailsLaunchModeCustom,
}

var mappingCreateImageDetailsLaunchModeEnumLowerCase = map[string]CreateImageDetailsLaunchModeEnum{
	"native":          CreateImageDetailsLaunchModeNative,
	"emulated":        CreateImageDetailsLaunchModeEmulated,
	"paravirtualized": CreateImageDetailsLaunchModeParavirtualized,
	"custom":          CreateImageDetailsLaunchModeCustom,
}

// GetCreateImageDetailsLaunchModeEnumValues Enumerates the set of values for CreateImageDetailsLaunchModeEnum
func GetCreateImageDetailsLaunchModeEnumValues() []CreateImageDetailsLaunchModeEnum {
	values := make([]CreateImageDetailsLaunchModeEnum, 0)
	for _, v := range mappingCreateImageDetailsLaunchModeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateImageDetailsLaunchModeEnumStringValues Enumerates the set of values in String for CreateImageDetailsLaunchModeEnum
func GetCreateImageDetailsLaunchModeEnumStringValues() []string {
	return []string{
		"NATIVE",
		"EMULATED",
		"PARAVIRTUALIZED",
		"CUSTOM",
	}
}

// GetMappingCreateImageDetailsLaunchModeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateImageDetailsLaunchModeEnum(val string) (CreateImageDetailsLaunchModeEnum, bool) {
	enum, ok := mappingCreateImageDetailsLaunchModeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
