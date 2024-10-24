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

// UpdateVirtualNodePoolDetails The properties that define a request to update a virtual node pool.
type UpdateVirtualNodePoolDetails struct {

	// Display name of the virtual node pool. This is a non-unique value.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Initial labels that will be added to the Kubernetes Virtual Node object when it registers.
	InitialVirtualNodeLabels []InitialVirtualNodeLabel `mandatory:"false" json:"initialVirtualNodeLabels"`

	// A taint is a collection of <key, value, effect>. These taints will be applied to the Virtual Nodes of this Virtual Node Pool for Kubernetes scheduling.
	Taints []Taint `mandatory:"false" json:"taints"`

	// The number of Virtual Nodes that should be in the Virtual Node Pool. The placement configurations determine where these virtual nodes are placed.
	Size *int `mandatory:"false" json:"size"`

	// The list of placement configurations which determines where Virtual Nodes will be provisioned across as it relates to the subnet and availability domains. The size attribute determines how many we evenly spread across these placement configurations
	PlacementConfigurations []PlacementConfiguration `mandatory:"false" json:"placementConfigurations"`

	// List of network security group id's applied to the Virtual Node VNIC.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// The pod configuration for pods run on virtual nodes of this virtual node pool.
	PodConfiguration *PodConfiguration `mandatory:"false" json:"podConfiguration"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	VirtualNodeTags *VirtualNodeTags `mandatory:"false" json:"virtualNodeTags"`
}

func (m UpdateVirtualNodePoolDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateVirtualNodePoolDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
