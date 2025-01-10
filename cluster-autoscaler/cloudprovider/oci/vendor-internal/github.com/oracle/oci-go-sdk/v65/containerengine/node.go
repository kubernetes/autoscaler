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

// Node The properties that define a node.
type Node struct {

	// The OCID of the compute instance backing this node.
	Id *string `mandatory:"false" json:"id"`

	// The name of the node.
	Name *string `mandatory:"false" json:"name"`

	// The version of Kubernetes this node is running.
	KubernetesVersion *string `mandatory:"false" json:"kubernetesVersion"`

	// The name of the availability domain in which this node is placed.
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// The OCID of the subnet in which this node is placed.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// The OCID of the node pool to which this node belongs.
	NodePoolId *string `mandatory:"false" json:"nodePoolId"`

	// The fault domain of this node.
	FaultDomain *string `mandatory:"false" json:"faultDomain"`

	// The private IP address of this node.
	PrivateIp *string `mandatory:"false" json:"privateIp"`

	// The public IP address of this node.
	PublicIp *string `mandatory:"false" json:"publicIp"`

	// An error that may be associated with the node.
	NodeError *NodeError `mandatory:"false" json:"nodeError"`

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

	// The state of the node. For more information, see Monitoring Clusters (https://docs.cloud.oracle.com/Content/ContEng/Tasks/contengmonitoringclusters.htm)
	LifecycleState NodeLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Details about the state of the node.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`
}

func (m Node) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m Node) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingNodeLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetNodeLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// NodeLifecycleStateEnum Enum with underlying type: string
type NodeLifecycleStateEnum string

// Set of constants representing the allowable values for NodeLifecycleStateEnum
const (
	NodeLifecycleStateCreating NodeLifecycleStateEnum = "CREATING"
	NodeLifecycleStateActive   NodeLifecycleStateEnum = "ACTIVE"
	NodeLifecycleStateUpdating NodeLifecycleStateEnum = "UPDATING"
	NodeLifecycleStateDeleting NodeLifecycleStateEnum = "DELETING"
	NodeLifecycleStateDeleted  NodeLifecycleStateEnum = "DELETED"
	NodeLifecycleStateFailing  NodeLifecycleStateEnum = "FAILING"
	NodeLifecycleStateInactive NodeLifecycleStateEnum = "INACTIVE"
)

var mappingNodeLifecycleStateEnum = map[string]NodeLifecycleStateEnum{
	"CREATING": NodeLifecycleStateCreating,
	"ACTIVE":   NodeLifecycleStateActive,
	"UPDATING": NodeLifecycleStateUpdating,
	"DELETING": NodeLifecycleStateDeleting,
	"DELETED":  NodeLifecycleStateDeleted,
	"FAILING":  NodeLifecycleStateFailing,
	"INACTIVE": NodeLifecycleStateInactive,
}

var mappingNodeLifecycleStateEnumLowerCase = map[string]NodeLifecycleStateEnum{
	"creating": NodeLifecycleStateCreating,
	"active":   NodeLifecycleStateActive,
	"updating": NodeLifecycleStateUpdating,
	"deleting": NodeLifecycleStateDeleting,
	"deleted":  NodeLifecycleStateDeleted,
	"failing":  NodeLifecycleStateFailing,
	"inactive": NodeLifecycleStateInactive,
}

// GetNodeLifecycleStateEnumValues Enumerates the set of values for NodeLifecycleStateEnum
func GetNodeLifecycleStateEnumValues() []NodeLifecycleStateEnum {
	values := make([]NodeLifecycleStateEnum, 0)
	for _, v := range mappingNodeLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetNodeLifecycleStateEnumStringValues Enumerates the set of values in String for NodeLifecycleStateEnum
func GetNodeLifecycleStateEnumStringValues() []string {
	return []string{
		"CREATING",
		"ACTIVE",
		"UPDATING",
		"DELETING",
		"DELETED",
		"FAILING",
		"INACTIVE",
	}
}

// GetMappingNodeLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingNodeLifecycleStateEnum(val string) (NodeLifecycleStateEnum, bool) {
	enum, ok := mappingNodeLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
