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

// TopologyContainsEntityRelationship Defines the `contains` relationship between virtual network topology entities. A `Contains` relationship
// is defined when an entity fully owns, contains or manages another entity.
// For example, a subnet is contained and managed in the scope of a VCN, therefore a VCN has a
// `contains` relationship to a subnet.
type TopologyContainsEntityRelationship struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the first entity in the relationship.
	Id1 *string `mandatory:"true" json:"id1"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the second entity in the relationship.
	Id2 *string `mandatory:"true" json:"id2"`
}

// GetId1 returns Id1
func (m TopologyContainsEntityRelationship) GetId1() *string {
	return m.Id1
}

// GetId2 returns Id2
func (m TopologyContainsEntityRelationship) GetId2() *string {
	return m.Id2
}

func (m TopologyContainsEntityRelationship) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m TopologyContainsEntityRelationship) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m TopologyContainsEntityRelationship) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeTopologyContainsEntityRelationship TopologyContainsEntityRelationship
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeTopologyContainsEntityRelationship
	}{
		"CONTAINS",
		(MarshalTypeTopologyContainsEntityRelationship)(m),
	}

	return json.Marshal(&s)
}
