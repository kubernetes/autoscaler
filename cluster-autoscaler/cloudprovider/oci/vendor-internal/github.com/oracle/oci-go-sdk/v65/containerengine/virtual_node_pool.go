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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// VirtualNodePool A pool of virtual nodes attached to a cluster.
type VirtualNodePool struct {

	// The OCID of the virtual node pool.
	Id *string `mandatory:"true" json:"id"`

	// Compartment of the virtual node pool.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The cluster the virtual node pool is associated with. A virtual node pool can only be associated with one cluster.
	ClusterId *string `mandatory:"true" json:"clusterId"`

	// Display name of the virtual node pool. This is a non-unique value.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The version of Kubernetes running on the nodes in the node pool.
	KubernetesVersion *string `mandatory:"true" json:"kubernetesVersion"`

	// The list of placement configurations which determines where Virtual Nodes will be provisioned across as it relates to the subnet and availability domains. The size attribute determines how many we evenly spread across these placement configurations
	PlacementConfigurations []PlacementConfiguration `mandatory:"true" json:"placementConfigurations"`

	// Initial labels that will be added to the Kubernetes Virtual Node object when it registers. This is the same as virtualNodePool resources.
	InitialVirtualNodeLabels []InitialVirtualNodeLabel `mandatory:"false" json:"initialVirtualNodeLabels"`

	// A taint is a collection of <key, value, effect>. These taints will be applied to the Virtual Nodes of this Virtual Node Pool for Kubernetes scheduling.
	Taints []Taint `mandatory:"false" json:"taints"`

	// The number of Virtual Nodes that should be in the Virtual Node Pool. The placement configurations determine where these virtual nodes are placed.
	Size *int `mandatory:"false" json:"size"`

	// List of network security group id's applied to the Virtual Node VNIC.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// The pod configuration for pods run on virtual nodes of this virtual node pool.
	PodConfiguration *PodConfiguration `mandatory:"false" json:"podConfiguration"`

	// The state of the Virtual Node Pool.
	LifecycleState VirtualNodePoolLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Details about the state of the Virtual Node Pool.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// The time the virtual node pool was created.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The time the virtual node pool was updated.
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

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

	VirtualNodeTags *VirtualNodeTags `mandatory:"false" json:"virtualNodeTags"`
}

func (m VirtualNodePool) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VirtualNodePool) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingVirtualNodePoolLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetVirtualNodePoolLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
