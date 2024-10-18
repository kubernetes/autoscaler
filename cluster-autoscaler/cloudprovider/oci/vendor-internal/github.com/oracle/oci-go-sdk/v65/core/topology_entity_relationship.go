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

// TopologyEntityRelationship Defines the relationship between Virtual Network topology entities.
type TopologyEntityRelationship interface {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the first entity in the relationship.
	GetId1() *string

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the second entity in the relationship.
	GetId2() *string
}

type topologyentityrelationship struct {
	JsonData []byte
	Id1      *string `mandatory:"true" json:"id1"`
	Id2      *string `mandatory:"true" json:"id2"`
	Type     string  `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *topologyentityrelationship) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalertopologyentityrelationship topologyentityrelationship
	s := struct {
		Model Unmarshalertopologyentityrelationship
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Id1 = s.Model.Id1
	m.Id2 = s.Model.Id2
	m.Type = s.Model.Type

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *topologyentityrelationship) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "ROUTES_TO":
		mm := TopologyRoutesToEntityRelationship{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "ASSOCIATED_WITH":
		mm := TopologyAssociatedWithEntityRelationship{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "CONTAINS":
		mm := TopologyContainsEntityRelationship{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for TopologyEntityRelationship: %s.", m.Type)
		return *m, nil
	}
}

// GetId1 returns Id1
func (m topologyentityrelationship) GetId1() *string {
	return m.Id1
}

// GetId2 returns Id2
func (m topologyentityrelationship) GetId2() *string {
	return m.Id2
}

func (m topologyentityrelationship) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m topologyentityrelationship) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// TopologyEntityRelationshipTypeEnum Enum with underlying type: string
type TopologyEntityRelationshipTypeEnum string

// Set of constants representing the allowable values for TopologyEntityRelationshipTypeEnum
const (
	TopologyEntityRelationshipTypeContains       TopologyEntityRelationshipTypeEnum = "CONTAINS"
	TopologyEntityRelationshipTypeAssociatedWith TopologyEntityRelationshipTypeEnum = "ASSOCIATED_WITH"
	TopologyEntityRelationshipTypeRoutesTo       TopologyEntityRelationshipTypeEnum = "ROUTES_TO"
)

var mappingTopologyEntityRelationshipTypeEnum = map[string]TopologyEntityRelationshipTypeEnum{
	"CONTAINS":        TopologyEntityRelationshipTypeContains,
	"ASSOCIATED_WITH": TopologyEntityRelationshipTypeAssociatedWith,
	"ROUTES_TO":       TopologyEntityRelationshipTypeRoutesTo,
}

var mappingTopologyEntityRelationshipTypeEnumLowerCase = map[string]TopologyEntityRelationshipTypeEnum{
	"contains":        TopologyEntityRelationshipTypeContains,
	"associated_with": TopologyEntityRelationshipTypeAssociatedWith,
	"routes_to":       TopologyEntityRelationshipTypeRoutesTo,
}

// GetTopologyEntityRelationshipTypeEnumValues Enumerates the set of values for TopologyEntityRelationshipTypeEnum
func GetTopologyEntityRelationshipTypeEnumValues() []TopologyEntityRelationshipTypeEnum {
	values := make([]TopologyEntityRelationshipTypeEnum, 0)
	for _, v := range mappingTopologyEntityRelationshipTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetTopologyEntityRelationshipTypeEnumStringValues Enumerates the set of values in String for TopologyEntityRelationshipTypeEnum
func GetTopologyEntityRelationshipTypeEnumStringValues() []string {
	return []string{
		"CONTAINS",
		"ASSOCIATED_WITH",
		"ROUTES_TO",
	}
}

// GetMappingTopologyEntityRelationshipTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingTopologyEntityRelationshipTypeEnum(val string) (TopologyEntityRelationshipTypeEnum, bool) {
	enum, ok := mappingTopologyEntityRelationshipTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
