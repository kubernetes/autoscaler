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

// CapacitySource A capacity source of bare metal hosts.
type CapacitySource interface {
}

type capacitysource struct {
	JsonData     []byte
	CapacityType string `json:"capacityType"`
}

// UnmarshalJSON unmarshals json
func (m *capacitysource) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalercapacitysource capacitysource
	s := struct {
		Model Unmarshalercapacitysource
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.CapacityType = s.Model.CapacityType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *capacitysource) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.CapacityType {
	case "DEDICATED":
		mm := DedicatedCapacitySource{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for CapacitySource: %s.", m.CapacityType)
		return *m, nil
	}
}

func (m capacitysource) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m capacitysource) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CapacitySourceCapacityTypeEnum Enum with underlying type: string
type CapacitySourceCapacityTypeEnum string

// Set of constants representing the allowable values for CapacitySourceCapacityTypeEnum
const (
	CapacitySourceCapacityTypeDedicated CapacitySourceCapacityTypeEnum = "DEDICATED"
)

var mappingCapacitySourceCapacityTypeEnum = map[string]CapacitySourceCapacityTypeEnum{
	"DEDICATED": CapacitySourceCapacityTypeDedicated,
}

var mappingCapacitySourceCapacityTypeEnumLowerCase = map[string]CapacitySourceCapacityTypeEnum{
	"dedicated": CapacitySourceCapacityTypeDedicated,
}

// GetCapacitySourceCapacityTypeEnumValues Enumerates the set of values for CapacitySourceCapacityTypeEnum
func GetCapacitySourceCapacityTypeEnumValues() []CapacitySourceCapacityTypeEnum {
	values := make([]CapacitySourceCapacityTypeEnum, 0)
	for _, v := range mappingCapacitySourceCapacityTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCapacitySourceCapacityTypeEnumStringValues Enumerates the set of values in String for CapacitySourceCapacityTypeEnum
func GetCapacitySourceCapacityTypeEnumStringValues() []string {
	return []string{
		"DEDICATED",
	}
}

// GetMappingCapacitySourceCapacityTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCapacitySourceCapacityTypeEnum(val string) (CapacitySourceCapacityTypeEnum, bool) {
	enum, ok := mappingCapacitySourceCapacityTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
