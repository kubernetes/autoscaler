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

// ComputeImageCapabilitySchema Compute Image Capability Schema is a set of capabilities that filter the compute global capability schema
// version for an image.
type ComputeImageCapabilitySchema struct {

	// The id of the compute global image capability schema version
	Id *string `mandatory:"true" json:"id"`

	// The ocid of the compute global image capability schema
	ComputeGlobalImageCapabilitySchemaId *string `mandatory:"true" json:"computeGlobalImageCapabilitySchemaId"`

	// The name of the compute global image capability schema version
	ComputeGlobalImageCapabilitySchemaVersionName *string `mandatory:"true" json:"computeGlobalImageCapabilitySchemaVersionName"`

	// The OCID of the image associated with this compute image capability schema
	ImageId *string `mandatory:"true" json:"imageId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The map of each capability name to its ImageCapabilityDescriptor.
	SchemaData map[string]ImageCapabilitySchemaDescriptor `mandatory:"true" json:"schemaData"`

	// The date and time the compute image capability schema was created, in the format defined by
	// RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID of the compartment that contains the resource.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m ComputeImageCapabilitySchema) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ComputeImageCapabilitySchema) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *ComputeImageCapabilitySchema) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		CompartmentId                                 *string                                    `json:"compartmentId"`
		DefinedTags                                   map[string]map[string]interface{}          `json:"definedTags"`
		FreeformTags                                  map[string]string                          `json:"freeformTags"`
		Id                                            *string                                    `json:"id"`
		ComputeGlobalImageCapabilitySchemaId          *string                                    `json:"computeGlobalImageCapabilitySchemaId"`
		ComputeGlobalImageCapabilitySchemaVersionName *string                                    `json:"computeGlobalImageCapabilitySchemaVersionName"`
		ImageId                                       *string                                    `json:"imageId"`
		DisplayName                                   *string                                    `json:"displayName"`
		SchemaData                                    map[string]imagecapabilityschemadescriptor `json:"schemaData"`
		TimeCreated                                   *common.SDKTime                            `json:"timeCreated"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.CompartmentId = model.CompartmentId

	m.DefinedTags = model.DefinedTags

	m.FreeformTags = model.FreeformTags

	m.Id = model.Id

	m.ComputeGlobalImageCapabilitySchemaId = model.ComputeGlobalImageCapabilitySchemaId

	m.ComputeGlobalImageCapabilitySchemaVersionName = model.ComputeGlobalImageCapabilitySchemaVersionName

	m.ImageId = model.ImageId

	m.DisplayName = model.DisplayName

	m.SchemaData = make(map[string]ImageCapabilitySchemaDescriptor)
	for k, v := range model.SchemaData {
		nn, e = v.UnmarshalPolymorphicJSON(v.JsonData)
		if e != nil {
			return e
		}
		if nn != nil {
			m.SchemaData[k] = nn.(ImageCapabilitySchemaDescriptor)
		} else {
			m.SchemaData[k] = nil
		}
	}

	m.TimeCreated = model.TimeCreated

	return
}
