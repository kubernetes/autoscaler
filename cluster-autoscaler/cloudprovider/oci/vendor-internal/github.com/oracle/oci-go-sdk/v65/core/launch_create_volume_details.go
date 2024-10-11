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

// LaunchCreateVolumeDetails Define a volume that will be created and attached or attached to an instance on creation.
type LaunchCreateVolumeDetails interface {
}

type launchcreatevolumedetails struct {
	JsonData           []byte
	VolumeCreationType string `json:"volumeCreationType"`
}

// UnmarshalJSON unmarshals json
func (m *launchcreatevolumedetails) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerlaunchcreatevolumedetails launchcreatevolumedetails
	s := struct {
		Model Unmarshalerlaunchcreatevolumedetails
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.VolumeCreationType = s.Model.VolumeCreationType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *launchcreatevolumedetails) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.VolumeCreationType {
	case "ATTRIBUTES":
		mm := LaunchCreateVolumeFromAttributes{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for LaunchCreateVolumeDetails: %s.", m.VolumeCreationType)
		return *m, nil
	}
}

func (m launchcreatevolumedetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m launchcreatevolumedetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// LaunchCreateVolumeDetailsVolumeCreationTypeEnum Enum with underlying type: string
type LaunchCreateVolumeDetailsVolumeCreationTypeEnum string

// Set of constants representing the allowable values for LaunchCreateVolumeDetailsVolumeCreationTypeEnum
const (
	LaunchCreateVolumeDetailsVolumeCreationTypeAttributes LaunchCreateVolumeDetailsVolumeCreationTypeEnum = "ATTRIBUTES"
)

var mappingLaunchCreateVolumeDetailsVolumeCreationTypeEnum = map[string]LaunchCreateVolumeDetailsVolumeCreationTypeEnum{
	"ATTRIBUTES": LaunchCreateVolumeDetailsVolumeCreationTypeAttributes,
}

var mappingLaunchCreateVolumeDetailsVolumeCreationTypeEnumLowerCase = map[string]LaunchCreateVolumeDetailsVolumeCreationTypeEnum{
	"attributes": LaunchCreateVolumeDetailsVolumeCreationTypeAttributes,
}

// GetLaunchCreateVolumeDetailsVolumeCreationTypeEnumValues Enumerates the set of values for LaunchCreateVolumeDetailsVolumeCreationTypeEnum
func GetLaunchCreateVolumeDetailsVolumeCreationTypeEnumValues() []LaunchCreateVolumeDetailsVolumeCreationTypeEnum {
	values := make([]LaunchCreateVolumeDetailsVolumeCreationTypeEnum, 0)
	for _, v := range mappingLaunchCreateVolumeDetailsVolumeCreationTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetLaunchCreateVolumeDetailsVolumeCreationTypeEnumStringValues Enumerates the set of values in String for LaunchCreateVolumeDetailsVolumeCreationTypeEnum
func GetLaunchCreateVolumeDetailsVolumeCreationTypeEnumStringValues() []string {
	return []string{
		"ATTRIBUTES",
	}
}

// GetMappingLaunchCreateVolumeDetailsVolumeCreationTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingLaunchCreateVolumeDetailsVolumeCreationTypeEnum(val string) (LaunchCreateVolumeDetailsVolumeCreationTypeEnum, bool) {
	enum, ok := mappingLaunchCreateVolumeDetailsVolumeCreationTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
