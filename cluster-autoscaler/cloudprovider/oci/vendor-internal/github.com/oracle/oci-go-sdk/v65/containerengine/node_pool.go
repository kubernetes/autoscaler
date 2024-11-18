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

// NodePool A pool of compute nodes attached to a cluster. Avoid entering confidential information.
type NodePool struct {

	// The OCID of the node pool.
	Id *string `mandatory:"false" json:"id"`

	// The state of the nodepool. For more information, see Monitoring Clusters (https://docs.cloud.oracle.com/Content/ContEng/Tasks/contengmonitoringclusters.htm)
	LifecycleState NodePoolLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Details about the state of the nodepool.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// The OCID of the compartment in which the node pool exists.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The OCID of the cluster to which this node pool is attached.
	ClusterId *string `mandatory:"false" json:"clusterId"`

	// The name of the node pool.
	Name *string `mandatory:"false" json:"name"`

	// The version of Kubernetes running on the nodes in the node pool.
	KubernetesVersion *string `mandatory:"false" json:"kubernetesVersion"`

	// A list of key/value pairs to add to each underlying OCI instance in the node pool on launch.
	NodeMetadata map[string]string `mandatory:"false" json:"nodeMetadata"`

	// Deprecated. see `nodeSource`. The OCID of the image running on the nodes in the node pool.
	NodeImageId *string `mandatory:"false" json:"nodeImageId"`

	// Deprecated. see `nodeSource`. The name of the image running on the nodes in the node pool.
	NodeImageName *string `mandatory:"false" json:"nodeImageName"`

	// The shape configuration of the nodes.
	NodeShapeConfig *NodeShapeConfig `mandatory:"false" json:"nodeShapeConfig"`

	// Deprecated. see `nodeSourceDetails`. Source running on the nodes in the node pool.
	NodeSource NodeSourceOption `mandatory:"false" json:"nodeSource"`

	// Source running on the nodes in the node pool.
	NodeSourceDetails NodeSourceDetails `mandatory:"false" json:"nodeSourceDetails"`

	// The name of the node shape of the nodes in the node pool.
	NodeShape *string `mandatory:"false" json:"nodeShape"`

	// A list of key/value pairs to add to nodes after they join the Kubernetes cluster.
	InitialNodeLabels []KeyValue `mandatory:"false" json:"initialNodeLabels"`

	// The SSH public key on each node in the node pool on launch.
	SshPublicKey *string `mandatory:"false" json:"sshPublicKey"`

	// The number of nodes in each subnet.
	QuantityPerSubnet *int `mandatory:"false" json:"quantityPerSubnet"`

	// The OCIDs of the subnets in which to place nodes for this node pool.
	SubnetIds []string `mandatory:"false" json:"subnetIds"`

	// The nodes in the node pool.
	Nodes []Node `mandatory:"false" json:"nodes"`

	// The configuration of nodes in the node pool.
	NodeConfigDetails *NodePoolNodeConfigDetails `mandatory:"false" json:"nodeConfigDetails"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Usage of system tag keys. These predefined keys are scoped to namespaces.
	// Example: `{"orcl-cloud": {"free-tier-retained": "true"}}`
	SystemTags map[string]map[string]interface{} `mandatory:"false" json:"systemTags"`

	NodeEvictionNodePoolSettings *NodeEvictionNodePoolSettings `mandatory:"false" json:"nodeEvictionNodePoolSettings"`

	NodePoolCyclingDetails *NodePoolCyclingDetails `mandatory:"false" json:"nodePoolCyclingDetails"`
}

func (m NodePool) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m NodePool) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingNodePoolLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetNodePoolLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *NodePool) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		Id                           *string                           `json:"id"`
		LifecycleState               NodePoolLifecycleStateEnum        `json:"lifecycleState"`
		LifecycleDetails             *string                           `json:"lifecycleDetails"`
		CompartmentId                *string                           `json:"compartmentId"`
		ClusterId                    *string                           `json:"clusterId"`
		Name                         *string                           `json:"name"`
		KubernetesVersion            *string                           `json:"kubernetesVersion"`
		NodeMetadata                 map[string]string                 `json:"nodeMetadata"`
		NodeImageId                  *string                           `json:"nodeImageId"`
		NodeImageName                *string                           `json:"nodeImageName"`
		NodeShapeConfig              *NodeShapeConfig                  `json:"nodeShapeConfig"`
		NodeSource                   nodesourceoption                  `json:"nodeSource"`
		NodeSourceDetails            nodesourcedetails                 `json:"nodeSourceDetails"`
		NodeShape                    *string                           `json:"nodeShape"`
		InitialNodeLabels            []KeyValue                        `json:"initialNodeLabels"`
		SshPublicKey                 *string                           `json:"sshPublicKey"`
		QuantityPerSubnet            *int                              `json:"quantityPerSubnet"`
		SubnetIds                    []string                          `json:"subnetIds"`
		Nodes                        []Node                            `json:"nodes"`
		NodeConfigDetails            *NodePoolNodeConfigDetails        `json:"nodeConfigDetails"`
		FreeformTags                 map[string]string                 `json:"freeformTags"`
		DefinedTags                  map[string]map[string]interface{} `json:"definedTags"`
		SystemTags                   map[string]map[string]interface{} `json:"systemTags"`
		NodeEvictionNodePoolSettings *NodeEvictionNodePoolSettings     `json:"nodeEvictionNodePoolSettings"`
		NodePoolCyclingDetails       *NodePoolCyclingDetails           `json:"nodePoolCyclingDetails"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.Id = model.Id

	m.LifecycleState = model.LifecycleState

	m.LifecycleDetails = model.LifecycleDetails

	m.CompartmentId = model.CompartmentId

	m.ClusterId = model.ClusterId

	m.Name = model.Name

	m.KubernetesVersion = model.KubernetesVersion

	m.NodeMetadata = model.NodeMetadata

	m.NodeImageId = model.NodeImageId

	m.NodeImageName = model.NodeImageName

	m.NodeShapeConfig = model.NodeShapeConfig

	nn, e = model.NodeSource.UnmarshalPolymorphicJSON(model.NodeSource.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.NodeSource = nn.(NodeSourceOption)
	} else {
		m.NodeSource = nil
	}

	nn, e = model.NodeSourceDetails.UnmarshalPolymorphicJSON(model.NodeSourceDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.NodeSourceDetails = nn.(NodeSourceDetails)
	} else {
		m.NodeSourceDetails = nil
	}

	m.NodeShape = model.NodeShape

	m.InitialNodeLabels = make([]KeyValue, len(model.InitialNodeLabels))
	copy(m.InitialNodeLabels, model.InitialNodeLabels)
	m.SshPublicKey = model.SshPublicKey

	m.QuantityPerSubnet = model.QuantityPerSubnet

	m.SubnetIds = make([]string, len(model.SubnetIds))
	copy(m.SubnetIds, model.SubnetIds)
	m.Nodes = make([]Node, len(model.Nodes))
	copy(m.Nodes, model.Nodes)
	m.NodeConfigDetails = model.NodeConfigDetails

	m.FreeformTags = model.FreeformTags

	m.DefinedTags = model.DefinedTags

	m.SystemTags = model.SystemTags

	m.NodeEvictionNodePoolSettings = model.NodeEvictionNodePoolSettings

	m.NodePoolCyclingDetails = model.NodePoolCyclingDetails

	return
}
