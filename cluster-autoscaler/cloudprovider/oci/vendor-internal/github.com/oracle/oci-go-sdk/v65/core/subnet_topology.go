// Copyright (c) 2016, 2018, 2023, Oracle and/or its affiliates.  All rights reserved.
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

// SubnetTopology Defines the visualization of a subnet in a VCN.
type SubnetTopology struct {

	// Lists entities comprising the virtual network topology.
	Entities []interface{} `mandatory:"true" json:"entities"`

	// Lists relationships between entities in the virtual network topology.
	Relationships []TopologyEntityRelationship `mandatory:"true" json:"relationships"`

	// Records when the virtual network topology was created, in RFC3339 (https://tools.ietf.org/html/rfc3339) format for date and time.
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet for which the visualization is generated.
	SubnetId *string `mandatory:"false" json:"subnetId"`
}

// GetEntities returns Entities
func (m SubnetTopology) GetEntities() []interface{} {
	return m.Entities
}

// GetRelationships returns Relationships
func (m SubnetTopology) GetRelationships() []TopologyEntityRelationship {
	return m.Relationships
}

// GetTimeCreated returns TimeCreated
func (m SubnetTopology) GetTimeCreated() *common.SDKTime {
	return m.TimeCreated
}

func (m SubnetTopology) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m SubnetTopology) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m SubnetTopology) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeSubnetTopology SubnetTopology
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeSubnetTopology
	}{
		"SUBNET",
		(MarshalTypeSubnetTopology)(m),
	}

	return json.Marshal(&s)
}

// UnmarshalJSON unmarshals from json
func (m *SubnetTopology) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		SubnetId      *string                      `json:"subnetId"`
		Entities      []interface{}                `json:"entities"`
		Relationships []topologyentityrelationship `json:"relationships"`
		TimeCreated   *common.SDKTime              `json:"timeCreated"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.SubnetId = model.SubnetId

	m.Entities = make([]interface{}, len(model.Entities))
	for i, n := range model.Entities {
		m.Entities[i] = n
	}

	m.Relationships = make([]TopologyEntityRelationship, len(model.Relationships))
	for i, n := range model.Relationships {
		nn, e = n.UnmarshalPolymorphicJSON(n.JsonData)
		if e != nil {
			return e
		}
		if nn != nil {
			m.Relationships[i] = nn.(TopologyEntityRelationship)
		} else {
			m.Relationships[i] = nil
		}
	}

	m.TimeCreated = model.TimeCreated

	return
}
