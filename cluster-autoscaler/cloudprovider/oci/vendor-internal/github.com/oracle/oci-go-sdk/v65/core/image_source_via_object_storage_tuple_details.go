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

// ImageSourceViaObjectStorageTupleDetails The representation of ImageSourceViaObjectStorageTupleDetails
type ImageSourceViaObjectStorageTupleDetails struct {

	// The Object Storage bucket for the image.
	BucketName *string `mandatory:"true" json:"bucketName"`

	// The Object Storage namespace for the image.
	NamespaceName *string `mandatory:"true" json:"namespaceName"`

	// The Object Storage name for the image.
	ObjectName *string `mandatory:"true" json:"objectName"`

	OperatingSystem *string `mandatory:"false" json:"operatingSystem"`

	OperatingSystemVersion *string `mandatory:"false" json:"operatingSystemVersion"`

	// The format of the image to be imported. Only monolithic
	// images are supported. This attribute is not used for exported Oracle images with the OCI image format.
	SourceImageType ImageSourceDetailsSourceImageTypeEnum `mandatory:"false" json:"sourceImageType,omitempty"`
}

// GetOperatingSystem returns OperatingSystem
func (m ImageSourceViaObjectStorageTupleDetails) GetOperatingSystem() *string {
	return m.OperatingSystem
}

// GetOperatingSystemVersion returns OperatingSystemVersion
func (m ImageSourceViaObjectStorageTupleDetails) GetOperatingSystemVersion() *string {
	return m.OperatingSystemVersion
}

// GetSourceImageType returns SourceImageType
func (m ImageSourceViaObjectStorageTupleDetails) GetSourceImageType() ImageSourceDetailsSourceImageTypeEnum {
	return m.SourceImageType
}

func (m ImageSourceViaObjectStorageTupleDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ImageSourceViaObjectStorageTupleDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingImageSourceDetailsSourceImageTypeEnum(string(m.SourceImageType)); !ok && m.SourceImageType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SourceImageType: %s. Supported values are: %s.", m.SourceImageType, strings.Join(GetImageSourceDetailsSourceImageTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m ImageSourceViaObjectStorageTupleDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeImageSourceViaObjectStorageTupleDetails ImageSourceViaObjectStorageTupleDetails
	s := struct {
		DiscriminatorParam string `json:"sourceType"`
		MarshalTypeImageSourceViaObjectStorageTupleDetails
	}{
		"objectStorageTuple",
		(MarshalTypeImageSourceViaObjectStorageTupleDetails)(m),
	}

	return json.Marshal(&s)
}
