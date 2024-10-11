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

// ImageSourceDetails The representation of ImageSourceDetails
type ImageSourceDetails interface {
	GetOperatingSystem() *string

	GetOperatingSystemVersion() *string

	// The format of the image to be imported. Only monolithic
	// images are supported. This attribute is not used for exported Oracle images with the OCI image format.
	GetSourceImageType() ImageSourceDetailsSourceImageTypeEnum
}

type imagesourcedetails struct {
	JsonData               []byte
	OperatingSystem        *string                               `mandatory:"false" json:"operatingSystem"`
	OperatingSystemVersion *string                               `mandatory:"false" json:"operatingSystemVersion"`
	SourceImageType        ImageSourceDetailsSourceImageTypeEnum `mandatory:"false" json:"sourceImageType,omitempty"`
	SourceType             string                                `json:"sourceType"`
}

// UnmarshalJSON unmarshals json
func (m *imagesourcedetails) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerimagesourcedetails imagesourcedetails
	s := struct {
		Model Unmarshalerimagesourcedetails
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.OperatingSystem = s.Model.OperatingSystem
	m.OperatingSystemVersion = s.Model.OperatingSystemVersion
	m.SourceImageType = s.Model.SourceImageType
	m.SourceType = s.Model.SourceType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *imagesourcedetails) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.SourceType {
	case "objectStorageTuple":
		mm := ImageSourceViaObjectStorageTupleDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "objectStorageUri":
		mm := ImageSourceViaObjectStorageUriDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for ImageSourceDetails: %s.", m.SourceType)
		return *m, nil
	}
}

// GetOperatingSystem returns OperatingSystem
func (m imagesourcedetails) GetOperatingSystem() *string {
	return m.OperatingSystem
}

// GetOperatingSystemVersion returns OperatingSystemVersion
func (m imagesourcedetails) GetOperatingSystemVersion() *string {
	return m.OperatingSystemVersion
}

// GetSourceImageType returns SourceImageType
func (m imagesourcedetails) GetSourceImageType() ImageSourceDetailsSourceImageTypeEnum {
	return m.SourceImageType
}

func (m imagesourcedetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m imagesourcedetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingImageSourceDetailsSourceImageTypeEnum(string(m.SourceImageType)); !ok && m.SourceImageType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SourceImageType: %s. Supported values are: %s.", m.SourceImageType, strings.Join(GetImageSourceDetailsSourceImageTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ImageSourceDetailsSourceImageTypeEnum Enum with underlying type: string
type ImageSourceDetailsSourceImageTypeEnum string

// Set of constants representing the allowable values for ImageSourceDetailsSourceImageTypeEnum
const (
	ImageSourceDetailsSourceImageTypeQcow2 ImageSourceDetailsSourceImageTypeEnum = "QCOW2"
	ImageSourceDetailsSourceImageTypeVmdk  ImageSourceDetailsSourceImageTypeEnum = "VMDK"
)

var mappingImageSourceDetailsSourceImageTypeEnum = map[string]ImageSourceDetailsSourceImageTypeEnum{
	"QCOW2": ImageSourceDetailsSourceImageTypeQcow2,
	"VMDK":  ImageSourceDetailsSourceImageTypeVmdk,
}

var mappingImageSourceDetailsSourceImageTypeEnumLowerCase = map[string]ImageSourceDetailsSourceImageTypeEnum{
	"qcow2": ImageSourceDetailsSourceImageTypeQcow2,
	"vmdk":  ImageSourceDetailsSourceImageTypeVmdk,
}

// GetImageSourceDetailsSourceImageTypeEnumValues Enumerates the set of values for ImageSourceDetailsSourceImageTypeEnum
func GetImageSourceDetailsSourceImageTypeEnumValues() []ImageSourceDetailsSourceImageTypeEnum {
	values := make([]ImageSourceDetailsSourceImageTypeEnum, 0)
	for _, v := range mappingImageSourceDetailsSourceImageTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetImageSourceDetailsSourceImageTypeEnumStringValues Enumerates the set of values in String for ImageSourceDetailsSourceImageTypeEnum
func GetImageSourceDetailsSourceImageTypeEnumStringValues() []string {
	return []string{
		"QCOW2",
		"VMDK",
	}
}

// GetMappingImageSourceDetailsSourceImageTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingImageSourceDetailsSourceImageTypeEnum(val string) (ImageSourceDetailsSourceImageTypeEnum, bool) {
	enum, ok := mappingImageSourceDetailsSourceImageTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
