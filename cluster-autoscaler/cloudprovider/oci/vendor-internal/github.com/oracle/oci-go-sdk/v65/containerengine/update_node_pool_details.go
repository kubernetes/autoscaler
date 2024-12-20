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

// UpdateNodePoolDetails The properties that define a request to update a node pool.
type UpdateNodePoolDetails struct {

	// The new name for the cluster. Avoid entering confidential information.
	Name *string `mandatory:"false" json:"name"`

	// The version of Kubernetes to which the nodes in the node pool should be upgraded.
	KubernetesVersion *string `mandatory:"false" json:"kubernetesVersion"`

	// A list of key/value pairs to add to nodes after they join the Kubernetes cluster.
	InitialNodeLabels []KeyValue `mandatory:"false" json:"initialNodeLabels"`

	// The number of nodes to have in each subnet specified in the subnetIds property. This property is deprecated,
	// use nodeConfigDetails instead. If the current value of quantityPerSubnet is greater than 0, you can only
	// use quantityPerSubnet to scale the node pool. If the current value of quantityPerSubnet is equal to 0 and
	// the current value of size in nodeConfigDetails is greater than 0, before you can use quantityPerSubnet,
	// you must first scale the node pool to 0 nodes using nodeConfigDetails.
	QuantityPerSubnet *int `mandatory:"false" json:"quantityPerSubnet"`

	// The OCIDs of the subnets in which to place nodes for this node pool. This property is deprecated,
	// use nodeConfigDetails instead. Only one of the subnetIds or nodeConfigDetails
	// properties can be specified.
	SubnetIds []string `mandatory:"false" json:"subnetIds"`

	// The configuration of nodes in the node pool. Only one of the subnetIds or nodeConfigDetails
	// properties should be specified. If the current value of quantityPerSubnet is greater than 0, the node
	// pool may still be scaled using quantityPerSubnet. Before you can use nodeConfigDetails,
	// you must first scale the node pool to 0 nodes using quantityPerSubnet.
	NodeConfigDetails *UpdateNodePoolNodeConfigDetails `mandatory:"false" json:"nodeConfigDetails"`

	// A list of key/value pairs to add to each underlying OCI instance in the node pool on launch.
	NodeMetadata map[string]string `mandatory:"false" json:"nodeMetadata"`

	// Specify the source to use to launch nodes in the node pool. Currently, image is the only supported source.
	NodeSourceDetails NodeSourceDetails `mandatory:"false" json:"nodeSourceDetails"`

	// The SSH public key to add to each node in the node pool on launch.
	SshPublicKey *string `mandatory:"false" json:"sshPublicKey"`

	// The name of the node shape of the nodes in the node pool used on launch.
	NodeShape *string `mandatory:"false" json:"nodeShape"`

	// Specify the configuration of the shape to launch nodes in the node pool.
	NodeShapeConfig *UpdateNodeShapeConfigDetails `mandatory:"false" json:"nodeShapeConfig"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	NodeEvictionNodePoolSettings *NodeEvictionNodePoolSettings `mandatory:"false" json:"nodeEvictionNodePoolSettings"`

	NodePoolCyclingDetails *NodePoolCyclingDetails `mandatory:"false" json:"nodePoolCyclingDetails"`
}

func (m UpdateNodePoolDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateNodePoolDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *UpdateNodePoolDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		Name                         *string                           `json:"name"`
		KubernetesVersion            *string                           `json:"kubernetesVersion"`
		InitialNodeLabels            []KeyValue                        `json:"initialNodeLabels"`
		QuantityPerSubnet            *int                              `json:"quantityPerSubnet"`
		SubnetIds                    []string                          `json:"subnetIds"`
		NodeConfigDetails            *UpdateNodePoolNodeConfigDetails  `json:"nodeConfigDetails"`
		NodeMetadata                 map[string]string                 `json:"nodeMetadata"`
		NodeSourceDetails            nodesourcedetails                 `json:"nodeSourceDetails"`
		SshPublicKey                 *string                           `json:"sshPublicKey"`
		NodeShape                    *string                           `json:"nodeShape"`
		NodeShapeConfig              *UpdateNodeShapeConfigDetails     `json:"nodeShapeConfig"`
		FreeformTags                 map[string]string                 `json:"freeformTags"`
		DefinedTags                  map[string]map[string]interface{} `json:"definedTags"`
		NodeEvictionNodePoolSettings *NodeEvictionNodePoolSettings     `json:"nodeEvictionNodePoolSettings"`
		NodePoolCyclingDetails       *NodePoolCyclingDetails           `json:"nodePoolCyclingDetails"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.Name = model.Name

	m.KubernetesVersion = model.KubernetesVersion

	m.InitialNodeLabels = make([]KeyValue, len(model.InitialNodeLabels))
	copy(m.InitialNodeLabels, model.InitialNodeLabels)
	m.QuantityPerSubnet = model.QuantityPerSubnet

	m.SubnetIds = make([]string, len(model.SubnetIds))
	copy(m.SubnetIds, model.SubnetIds)
	m.NodeConfigDetails = model.NodeConfigDetails

	m.NodeMetadata = model.NodeMetadata

	nn, e = model.NodeSourceDetails.UnmarshalPolymorphicJSON(model.NodeSourceDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.NodeSourceDetails = nn.(NodeSourceDetails)
	} else {
		m.NodeSourceDetails = nil
	}

	m.SshPublicKey = model.SshPublicKey

	m.NodeShape = model.NodeShape

	m.NodeShapeConfig = model.NodeShapeConfig

	m.FreeformTags = model.FreeformTags

	m.DefinedTags = model.DefinedTags

	m.NodeEvictionNodePoolSettings = model.NodeEvictionNodePoolSettings

	m.NodePoolCyclingDetails = model.NodePoolCyclingDetails

	return
}
