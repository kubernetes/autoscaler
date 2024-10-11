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

// InstanceConfigurationAutotunePolicy An autotune policy automatically tunes the volume's performace based on the type of the policy.
type InstanceConfigurationAutotunePolicy interface {
}

type instanceconfigurationautotunepolicy struct {
	JsonData     []byte
	AutotuneType string `json:"autotuneType"`
}

// UnmarshalJSON unmarshals json
func (m *instanceconfigurationautotunepolicy) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerinstanceconfigurationautotunepolicy instanceconfigurationautotunepolicy
	s := struct {
		Model Unmarshalerinstanceconfigurationautotunepolicy
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.AutotuneType = s.Model.AutotuneType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *instanceconfigurationautotunepolicy) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.AutotuneType {
	case "PERFORMANCE_BASED":
		mm := InstanceConfigurationPerformanceBasedAutotunePolicy{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "DETACHED_VOLUME":
		mm := InstanceConfigurationDetachedVolumeAutotunePolicy{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for InstanceConfigurationAutotunePolicy: %s.", m.AutotuneType)
		return *m, nil
	}
}

func (m instanceconfigurationautotunepolicy) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m instanceconfigurationautotunepolicy) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstanceConfigurationAutotunePolicyAutotuneTypeEnum Enum with underlying type: string
type InstanceConfigurationAutotunePolicyAutotuneTypeEnum string

// Set of constants representing the allowable values for InstanceConfigurationAutotunePolicyAutotuneTypeEnum
const (
	InstanceConfigurationAutotunePolicyAutotuneTypeDetachedVolume   InstanceConfigurationAutotunePolicyAutotuneTypeEnum = "DETACHED_VOLUME"
	InstanceConfigurationAutotunePolicyAutotuneTypePerformanceBased InstanceConfigurationAutotunePolicyAutotuneTypeEnum = "PERFORMANCE_BASED"
)

var mappingInstanceConfigurationAutotunePolicyAutotuneTypeEnum = map[string]InstanceConfigurationAutotunePolicyAutotuneTypeEnum{
	"DETACHED_VOLUME":   InstanceConfigurationAutotunePolicyAutotuneTypeDetachedVolume,
	"PERFORMANCE_BASED": InstanceConfigurationAutotunePolicyAutotuneTypePerformanceBased,
}

var mappingInstanceConfigurationAutotunePolicyAutotuneTypeEnumLowerCase = map[string]InstanceConfigurationAutotunePolicyAutotuneTypeEnum{
	"detached_volume":   InstanceConfigurationAutotunePolicyAutotuneTypeDetachedVolume,
	"performance_based": InstanceConfigurationAutotunePolicyAutotuneTypePerformanceBased,
}

// GetInstanceConfigurationAutotunePolicyAutotuneTypeEnumValues Enumerates the set of values for InstanceConfigurationAutotunePolicyAutotuneTypeEnum
func GetInstanceConfigurationAutotunePolicyAutotuneTypeEnumValues() []InstanceConfigurationAutotunePolicyAutotuneTypeEnum {
	values := make([]InstanceConfigurationAutotunePolicyAutotuneTypeEnum, 0)
	for _, v := range mappingInstanceConfigurationAutotunePolicyAutotuneTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceConfigurationAutotunePolicyAutotuneTypeEnumStringValues Enumerates the set of values in String for InstanceConfigurationAutotunePolicyAutotuneTypeEnum
func GetInstanceConfigurationAutotunePolicyAutotuneTypeEnumStringValues() []string {
	return []string{
		"DETACHED_VOLUME",
		"PERFORMANCE_BASED",
	}
}

// GetMappingInstanceConfigurationAutotunePolicyAutotuneTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceConfigurationAutotunePolicyAutotuneTypeEnum(val string) (InstanceConfigurationAutotunePolicyAutotuneTypeEnum, bool) {
	enum, ok := mappingInstanceConfigurationAutotunePolicyAutotuneTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
