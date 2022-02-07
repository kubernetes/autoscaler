// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"encoding/json"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// PreemptionAction The action to run when the preemptible instance is interrupted for eviction.
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
		return *m, nil
	}
}

func (m preemptionaction) String() string {
	return common.PointerString(m)
}

// PreemptionActionTypeEnum Enum with underlying type: string
type PreemptionActionTypeEnum string

// Set of constants representing the allowable values for PreemptionActionTypeEnum
const (
	PreemptionActionTypeTerminate PreemptionActionTypeEnum = "TERMINATE"
)

var mappingPreemptionActionType = map[string]PreemptionActionTypeEnum{
	"TERMINATE": PreemptionActionTypeTerminate,
}

// GetPreemptionActionTypeEnumValues Enumerates the set of values for PreemptionActionTypeEnum
func GetPreemptionActionTypeEnumValues() []PreemptionActionTypeEnum {
	values := make([]PreemptionActionTypeEnum, 0)
	for _, v := range mappingPreemptionActionType {
		values = append(values, v)
	}
	return values
}
