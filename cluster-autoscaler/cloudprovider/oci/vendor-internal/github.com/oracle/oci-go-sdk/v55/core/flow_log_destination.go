// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
//

package core

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// FlowLogDestination Where to store the flow logs.
type FlowLogDestination interface {
}

type flowlogdestination struct {
	JsonData        []byte
	DestinationType string `json:"destinationType"`
}

// UnmarshalJSON unmarshals json
func (m *flowlogdestination) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerflowlogdestination flowlogdestination
	s := struct {
		Model Unmarshalerflowlogdestination
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.DestinationType = s.Model.DestinationType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *flowlogdestination) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.DestinationType {
	case "OBJECT_STORAGE":
		mm := FlowLogObjectStorageDestination{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

func (m flowlogdestination) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m flowlogdestination) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// FlowLogDestinationDestinationTypeEnum Enum with underlying type: string
type FlowLogDestinationDestinationTypeEnum string

// Set of constants representing the allowable values for FlowLogDestinationDestinationTypeEnum
const (
	FlowLogDestinationDestinationTypeObjectStorage FlowLogDestinationDestinationTypeEnum = "OBJECT_STORAGE"
)

var mappingFlowLogDestinationDestinationTypeEnum = map[string]FlowLogDestinationDestinationTypeEnum{
	"OBJECT_STORAGE": FlowLogDestinationDestinationTypeObjectStorage,
}

// GetFlowLogDestinationDestinationTypeEnumValues Enumerates the set of values for FlowLogDestinationDestinationTypeEnum
func GetFlowLogDestinationDestinationTypeEnumValues() []FlowLogDestinationDestinationTypeEnum {
	values := make([]FlowLogDestinationDestinationTypeEnum, 0)
	for _, v := range mappingFlowLogDestinationDestinationTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetFlowLogDestinationDestinationTypeEnumStringValues Enumerates the set of values in String for FlowLogDestinationDestinationTypeEnum
func GetFlowLogDestinationDestinationTypeEnumStringValues() []string {
	return []string{
		"OBJECT_STORAGE",
	}
}
