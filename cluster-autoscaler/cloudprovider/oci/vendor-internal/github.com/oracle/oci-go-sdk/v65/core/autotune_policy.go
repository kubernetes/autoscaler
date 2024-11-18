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

// AutotunePolicy An autotune policy automatically tunes the volume's performace based on the type of the policy.
type AutotunePolicy interface {
}

type autotunepolicy struct {
	JsonData     []byte
	AutotuneType string `json:"autotuneType"`
}

// UnmarshalJSON unmarshals json
func (m *autotunepolicy) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerautotunepolicy autotunepolicy
	s := struct {
		Model Unmarshalerautotunepolicy
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.AutotuneType = s.Model.AutotuneType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *autotunepolicy) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.AutotuneType {
	case "DETACHED_VOLUME":
		mm := DetachedVolumeAutotunePolicy{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "PERFORMANCE_BASED":
		mm := PerformanceBasedAutotunePolicy{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for AutotunePolicy: %s.", m.AutotuneType)
		return *m, nil
	}
}

func (m autotunepolicy) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m autotunepolicy) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// AutotunePolicyAutotuneTypeEnum Enum with underlying type: string
type AutotunePolicyAutotuneTypeEnum string

// Set of constants representing the allowable values for AutotunePolicyAutotuneTypeEnum
const (
	AutotunePolicyAutotuneTypeDetachedVolume   AutotunePolicyAutotuneTypeEnum = "DETACHED_VOLUME"
	AutotunePolicyAutotuneTypePerformanceBased AutotunePolicyAutotuneTypeEnum = "PERFORMANCE_BASED"
)

var mappingAutotunePolicyAutotuneTypeEnum = map[string]AutotunePolicyAutotuneTypeEnum{
	"DETACHED_VOLUME":   AutotunePolicyAutotuneTypeDetachedVolume,
	"PERFORMANCE_BASED": AutotunePolicyAutotuneTypePerformanceBased,
}

var mappingAutotunePolicyAutotuneTypeEnumLowerCase = map[string]AutotunePolicyAutotuneTypeEnum{
	"detached_volume":   AutotunePolicyAutotuneTypeDetachedVolume,
	"performance_based": AutotunePolicyAutotuneTypePerformanceBased,
}

// GetAutotunePolicyAutotuneTypeEnumValues Enumerates the set of values for AutotunePolicyAutotuneTypeEnum
func GetAutotunePolicyAutotuneTypeEnumValues() []AutotunePolicyAutotuneTypeEnum {
	values := make([]AutotunePolicyAutotuneTypeEnum, 0)
	for _, v := range mappingAutotunePolicyAutotuneTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetAutotunePolicyAutotuneTypeEnumStringValues Enumerates the set of values in String for AutotunePolicyAutotuneTypeEnum
func GetAutotunePolicyAutotuneTypeEnumStringValues() []string {
	return []string{
		"DETACHED_VOLUME",
		"PERFORMANCE_BASED",
	}
}

// GetMappingAutotunePolicyAutotuneTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingAutotunePolicyAutotuneTypeEnum(val string) (AutotunePolicyAutotuneTypeEnum, bool) {
	enum, ok := mappingAutotunePolicyAutotuneTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
