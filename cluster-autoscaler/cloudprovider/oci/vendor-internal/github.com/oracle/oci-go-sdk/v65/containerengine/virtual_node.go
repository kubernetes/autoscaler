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

// VirtualNode The properties that define a virtual node.
type VirtualNode struct {

	// The ocid of the virtual node.
	Id *string `mandatory:"true" json:"id"`

	// The name of the virtual node.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The ocid of the virtual node pool this virtual node belongs to.
	VirtualNodePoolId *string `mandatory:"true" json:"virtualNodePoolId"`

	// The version of Kubernetes this virtual node is running.
	KubernetesVersion *string `mandatory:"false" json:"kubernetesVersion"`

	// The name of the availability domain in which this virtual node is placed
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// The fault domain of this virtual node.
	FaultDomain *string `mandatory:"false" json:"faultDomain"`

	// The OCID of the subnet in which this Virtual Node is placed.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// NSG Ids applied to virtual node vnic.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// The private IP address of this Virtual Node.
	PrivateIp *string `mandatory:"false" json:"privateIp"`

	// An error that may be associated with the virtual node.
	VirtualNodeError *string `mandatory:"false" json:"virtualNodeError"`

	// The state of the Virtual Node.
	LifecycleState VirtualNodeLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Details about the state of the Virtual Node.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// The time at which the virtual node was created.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

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
}

func (m VirtualNode) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VirtualNode) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingVirtualNodeLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetVirtualNodeLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
