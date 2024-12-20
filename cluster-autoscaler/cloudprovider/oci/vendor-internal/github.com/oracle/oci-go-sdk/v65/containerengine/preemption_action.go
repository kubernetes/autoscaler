// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Kubernetes Engine API
//
// API for the Kubernetes Engine service (also known as the Container Engine for Kubernetes service). Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Kubernetes Engine (https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// PreemptionAction The action to run when the preemptible node is interrupted for eviction.
type PreemptionAction interface {
}

type preemptionaction struct {
	JsonData []byte
	Type     string `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *preemptionaction) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerpreemptionaction preemptionaction
	s := struct {
		Model Unmarshalerpreemptionaction
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Type = s.Model.Type

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *preemptionaction) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "TERMINATE":
		mm := TerminatePreemptionAction{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for PreemptionAction: %s.", m.Type)
		return *m, nil
	}
}

func (m preemptionaction) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m preemptionaction) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// PreemptionActionTypeEnum Enum with underlying type: string
type PreemptionActionTypeEnum string

// Set of constants representing the allowable values for PreemptionActionTypeEnum
const (
	PreemptionActionTypeTerminate PreemptionActionTypeEnum = "TERMINATE"
)

var mappingPreemptionActionTypeEnum = map[string]PreemptionActionTypeEnum{
	"TERMINATE": PreemptionActionTypeTerminate,
}

var mappingPreemptionActionTypeEnumLowerCase = map[string]PreemptionActionTypeEnum{
	"terminate": PreemptionActionTypeTerminate,
}

// GetPreemptionActionTypeEnumValues Enumerates the set of values for PreemptionActionTypeEnum
func GetPreemptionActionTypeEnumValues() []PreemptionActionTypeEnum {
	values := make([]PreemptionActionTypeEnum, 0)
	for _, v := range mappingPreemptionActionTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetPreemptionActionTypeEnumStringValues Enumerates the set of values in String for PreemptionActionTypeEnum
func GetPreemptionActionTypeEnumStringValues() []string {
	return []string{
		"TERMINATE",
	}
}

// GetMappingPreemptionActionTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingPreemptionActionTypeEnum(val string) (PreemptionActionTypeEnum, bool) {
	enum, ok := mappingPreemptionActionTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
