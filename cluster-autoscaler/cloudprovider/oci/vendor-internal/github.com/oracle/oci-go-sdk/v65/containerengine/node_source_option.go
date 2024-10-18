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

// NodeSourceOption The source option for the node.
type NodeSourceOption interface {

	// The user-friendly name of the entity corresponding to the OCID.
	GetSourceName() *string
}

type nodesourceoption struct {
	JsonData   []byte
	SourceName *string `mandatory:"false" json:"sourceName"`
	SourceType string  `json:"sourceType"`
}

// UnmarshalJSON unmarshals json
func (m *nodesourceoption) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalernodesourceoption nodesourceoption
	s := struct {
		Model Unmarshalernodesourceoption
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.SourceName = s.Model.SourceName
	m.SourceType = s.Model.SourceType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *nodesourceoption) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.SourceType {
	case "IMAGE":
		mm := NodeSourceViaImageOption{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for NodeSourceOption: %s.", m.SourceType)
		return *m, nil
	}
}

// GetSourceName returns SourceName
func (m nodesourceoption) GetSourceName() *string {
	return m.SourceName
}

func (m nodesourceoption) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m nodesourceoption) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
