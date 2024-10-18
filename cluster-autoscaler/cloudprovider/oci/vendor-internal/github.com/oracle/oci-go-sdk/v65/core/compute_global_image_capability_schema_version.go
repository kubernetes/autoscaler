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

// ComputeGlobalImageCapabilitySchemaVersion Compute Global Image Capability Schema Version is a set of all possible capabilities for a collection of images.
type ComputeGlobalImageCapabilitySchemaVersion struct {

	// The name of the compute global image capability schema version
	Name *string `mandatory:"true" json:"name"`

	// The ocid of the compute global image capability schema
	ComputeGlobalImageCapabilitySchemaId *string `mandatory:"true" json:"computeGlobalImageCapabilitySchemaId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The map of each capability name to its ImageCapabilityDescriptor.
	SchemaData map[string]ImageCapabilitySchemaDescriptor `mandatory:"true" json:"schemaData"`

	// The date and time the compute global image capability schema version was created, in the format defined by
	// RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`
}

func (m ComputeGlobalImageCapabilitySchemaVersion) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ComputeGlobalImageCapabilitySchemaVersion) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *ComputeGlobalImageCapabilitySchemaVersion) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		Name                                 *string                                    `json:"name"`
		ComputeGlobalImageCapabilitySchemaId *string                                    `json:"computeGlobalImageCapabilitySchemaId"`
		DisplayName                          *string                                    `json:"displayName"`
		SchemaData                           map[string]imagecapabilityschemadescriptor `json:"schemaData"`
		TimeCreated                          *common.SDKTime                            `json:"timeCreated"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.Name = model.Name

	m.ComputeGlobalImageCapabilitySchemaId = model.ComputeGlobalImageCapabilitySchemaId

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
