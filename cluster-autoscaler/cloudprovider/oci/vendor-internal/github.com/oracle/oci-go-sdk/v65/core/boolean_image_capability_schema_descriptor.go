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

// BooleanImageCapabilitySchemaDescriptor Boolean type ImageCapabilitySchemaDescriptor
type BooleanImageCapabilitySchemaDescriptor struct {

	// the default value
	DefaultValue *bool `mandatory:"false" json:"defaultValue"`

	Source ImageCapabilitySchemaDescriptorSourceEnum `mandatory:"true" json:"source"`
}

// GetSource returns Source
func (m BooleanImageCapabilitySchemaDescriptor) GetSource() ImageCapabilitySchemaDescriptorSourceEnum {
	return m.Source
}

func (m BooleanImageCapabilitySchemaDescriptor) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m BooleanImageCapabilitySchemaDescriptor) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingImageCapabilitySchemaDescriptorSourceEnum(string(m.Source)); !ok && m.Source != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Source: %s. Supported values are: %s.", m.Source, strings.Join(GetImageCapabilitySchemaDescriptorSourceEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m BooleanImageCapabilitySchemaDescriptor) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeBooleanImageCapabilitySchemaDescriptor BooleanImageCapabilitySchemaDescriptor
	s := struct {
		DiscriminatorParam string `json:"descriptorType"`
		MarshalTypeBooleanImageCapabilitySchemaDescriptor
	}{
		"boolean",
		(MarshalTypeBooleanImageCapabilitySchemaDescriptor)(m),
	}

	return json.Marshal(&s)
}
