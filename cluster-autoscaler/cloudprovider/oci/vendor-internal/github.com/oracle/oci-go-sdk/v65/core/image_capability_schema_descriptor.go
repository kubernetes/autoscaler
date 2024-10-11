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

// ImageCapabilitySchemaDescriptor Image Capability Schema Descriptor is a type of capability for an image.
type ImageCapabilitySchemaDescriptor interface {
	GetSource() ImageCapabilitySchemaDescriptorSourceEnum
}

type imagecapabilityschemadescriptor struct {
	JsonData       []byte
	Source         ImageCapabilitySchemaDescriptorSourceEnum `mandatory:"true" json:"source"`
	DescriptorType string                                    `json:"descriptorType"`
}

// UnmarshalJSON unmarshals json
func (m *imagecapabilityschemadescriptor) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerimagecapabilityschemadescriptor imagecapabilityschemadescriptor
	s := struct {
		Model Unmarshalerimagecapabilityschemadescriptor
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Source = s.Model.Source
	m.DescriptorType = s.Model.DescriptorType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *imagecapabilityschemadescriptor) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.DescriptorType {
	case "enumstring":
		mm := EnumStringImageCapabilitySchemaDescriptor{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "enuminteger":
		mm := EnumIntegerImageCapabilityDescriptor{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "boolean":
		mm := BooleanImageCapabilitySchemaDescriptor{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for ImageCapabilitySchemaDescriptor: %s.", m.DescriptorType)
		return *m, nil
	}
}

// GetSource returns Source
func (m imagecapabilityschemadescriptor) GetSource() ImageCapabilitySchemaDescriptorSourceEnum {
	return m.Source
}

func (m imagecapabilityschemadescriptor) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m imagecapabilityschemadescriptor) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingImageCapabilitySchemaDescriptorSourceEnum(string(m.Source)); !ok && m.Source != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Source: %s. Supported values are: %s.", m.Source, strings.Join(GetImageCapabilitySchemaDescriptorSourceEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ImageCapabilitySchemaDescriptorSourceEnum Enum with underlying type: string
type ImageCapabilitySchemaDescriptorSourceEnum string

// Set of constants representing the allowable values for ImageCapabilitySchemaDescriptorSourceEnum
const (
	ImageCapabilitySchemaDescriptorSourceGlobal ImageCapabilitySchemaDescriptorSourceEnum = "GLOBAL"
	ImageCapabilitySchemaDescriptorSourceImage  ImageCapabilitySchemaDescriptorSourceEnum = "IMAGE"
)

var mappingImageCapabilitySchemaDescriptorSourceEnum = map[string]ImageCapabilitySchemaDescriptorSourceEnum{
	"GLOBAL": ImageCapabilitySchemaDescriptorSourceGlobal,
	"IMAGE":  ImageCapabilitySchemaDescriptorSourceImage,
}

var mappingImageCapabilitySchemaDescriptorSourceEnumLowerCase = map[string]ImageCapabilitySchemaDescriptorSourceEnum{
	"global": ImageCapabilitySchemaDescriptorSourceGlobal,
	"image":  ImageCapabilitySchemaDescriptorSourceImage,
}

// GetImageCapabilitySchemaDescriptorSourceEnumValues Enumerates the set of values for ImageCapabilitySchemaDescriptorSourceEnum
func GetImageCapabilitySchemaDescriptorSourceEnumValues() []ImageCapabilitySchemaDescriptorSourceEnum {
	values := make([]ImageCapabilitySchemaDescriptorSourceEnum, 0)
	for _, v := range mappingImageCapabilitySchemaDescriptorSourceEnum {
		values = append(values, v)
	}
	return values
}

// GetImageCapabilitySchemaDescriptorSourceEnumStringValues Enumerates the set of values in String for ImageCapabilitySchemaDescriptorSourceEnum
func GetImageCapabilitySchemaDescriptorSourceEnumStringValues() []string {
	return []string{
		"GLOBAL",
		"IMAGE",
	}
}

// GetMappingImageCapabilitySchemaDescriptorSourceEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingImageCapabilitySchemaDescriptorSourceEnum(val string) (ImageCapabilitySchemaDescriptorSourceEnum, bool) {
	enum, ok := mappingImageCapabilitySchemaDescriptorSourceEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
