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

// NodePoolNodeConfigDetails The size and placement configuration of nodes in the node pool.
type NodePoolNodeConfigDetails struct {

	// The number of nodes in the node pool.
	Size *int `mandatory:"false" json:"size"`

	// The OCIDs of the Network Security Group(s) to associate nodes for this node pool with. For more information about NSGs, see NetworkSecurityGroup.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// The OCID of the Key Management Service key assigned to the boot volume.
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`

	// Whether to enable in-transit encryption for the data volume's paravirtualized attachment. This field applies to both block volumes and boot volumes. The default value is false.
	IsPvEncryptionInTransitEnabled *bool `mandatory:"false" json:"isPvEncryptionInTransitEnabled"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// The placement configurations for the node pool. Provide one placement
	// configuration for each availability domain in which you intend to launch a node.
	// To use the node pool with a regional subnet, provide a placement configuration for
	// each availability domain, and include the regional subnet in each placement
	// configuration.
	PlacementConfigs []NodePoolPlacementConfigDetails `mandatory:"false" json:"placementConfigs"`

	// The CNI related configuration of pods in the node pool.
	NodePoolPodNetworkOptionDetails NodePoolPodNetworkOptionDetails `mandatory:"false" json:"nodePoolPodNetworkOptionDetails"`
}

func (m NodePoolNodeConfigDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m NodePoolNodeConfigDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *NodePoolNodeConfigDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		Size                            *int                              `json:"size"`
		NsgIds                          []string                          `json:"nsgIds"`
		KmsKeyId                        *string                           `json:"kmsKeyId"`
		IsPvEncryptionInTransitEnabled  *bool                             `json:"isPvEncryptionInTransitEnabled"`
		FreeformTags                    map[string]string                 `json:"freeformTags"`
		DefinedTags                     map[string]map[string]interface{} `json:"definedTags"`
		PlacementConfigs                []NodePoolPlacementConfigDetails  `json:"placementConfigs"`
		NodePoolPodNetworkOptionDetails nodepoolpodnetworkoptiondetails   `json:"nodePoolPodNetworkOptionDetails"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.Size = model.Size

	m.NsgIds = make([]string, len(model.NsgIds))
	copy(m.NsgIds, model.NsgIds)
	m.KmsKeyId = model.KmsKeyId

	m.IsPvEncryptionInTransitEnabled = model.IsPvEncryptionInTransitEnabled

	m.FreeformTags = model.FreeformTags

	m.DefinedTags = model.DefinedTags

	m.PlacementConfigs = make([]NodePoolPlacementConfigDetails, len(model.PlacementConfigs))
	copy(m.PlacementConfigs, model.PlacementConfigs)
	nn, e = model.NodePoolPodNetworkOptionDetails.UnmarshalPolymorphicJSON(model.NodePoolPodNetworkOptionDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.NodePoolPodNetworkOptionDetails = nn.(NodePoolPodNetworkOptionDetails)
	} else {
		m.NodePoolPodNetworkOptionDetails = nil
	}

	return
}
