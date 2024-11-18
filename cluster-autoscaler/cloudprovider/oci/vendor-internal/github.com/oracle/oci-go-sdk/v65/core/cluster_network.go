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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// ClusterNetwork A cluster network is a group of high performance computing (HPC), GPU, or optimized bare metal
// instances that are connected with an ultra low-latency remote direct memory access (RDMA)
// network. Cluster networks with instance pools (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/managingclusternetworks.htm)
// use instance pools to manage groups of identical instances.
// Use cluster networks with instance pools when you want predictable capacity for a specific number of identical
// instances that are managed as a group.
// If you want to manage instances in the RDMA network independently of each other or use different types of instances
// in the network group, use compute clusters instead. For details, see ComputeCluster.
type ClusterNetwork struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the cluster network.
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the cluster network.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The current state of the cluster network.
	LifecycleState ClusterNetworkLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the resource was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The date and time the resource was updated, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeUpdated *common.SDKTime `mandatory:"true" json:"timeUpdated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the HPC island used by the cluster network.
	HpcIslandId *string `mandatory:"false" json:"hpcIslandId"`

	// The list of network block OCIDs of the HPC island.
	NetworkBlockIds []string `mandatory:"false" json:"networkBlockIds"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The instance pools in the cluster network.
	// Each cluster network can have one instance pool.
	InstancePools []InstancePool `mandatory:"false" json:"instancePools"`

	PlacementConfiguration *ClusterNetworkPlacementConfigurationDetails `mandatory:"false" json:"placementConfiguration"`
}

func (m ClusterNetwork) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ClusterNetwork) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingClusterNetworkLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetClusterNetworkLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ClusterNetworkLifecycleStateEnum Enum with underlying type: string
type ClusterNetworkLifecycleStateEnum string

// Set of constants representing the allowable values for ClusterNetworkLifecycleStateEnum
const (
	ClusterNetworkLifecycleStateProvisioning ClusterNetworkLifecycleStateEnum = "PROVISIONING"
	ClusterNetworkLifecycleStateScaling      ClusterNetworkLifecycleStateEnum = "SCALING"
	ClusterNetworkLifecycleStateStarting     ClusterNetworkLifecycleStateEnum = "STARTING"
	ClusterNetworkLifecycleStateStopping     ClusterNetworkLifecycleStateEnum = "STOPPING"
	ClusterNetworkLifecycleStateTerminating  ClusterNetworkLifecycleStateEnum = "TERMINATING"
	ClusterNetworkLifecycleStateStopped      ClusterNetworkLifecycleStateEnum = "STOPPED"
	ClusterNetworkLifecycleStateTerminated   ClusterNetworkLifecycleStateEnum = "TERMINATED"
	ClusterNetworkLifecycleStateRunning      ClusterNetworkLifecycleStateEnum = "RUNNING"
)

var mappingClusterNetworkLifecycleStateEnum = map[string]ClusterNetworkLifecycleStateEnum{
	"PROVISIONING": ClusterNetworkLifecycleStateProvisioning,
	"SCALING":      ClusterNetworkLifecycleStateScaling,
	"STARTING":     ClusterNetworkLifecycleStateStarting,
	"STOPPING":     ClusterNetworkLifecycleStateStopping,
	"TERMINATING":  ClusterNetworkLifecycleStateTerminating,
	"STOPPED":      ClusterNetworkLifecycleStateStopped,
	"TERMINATED":   ClusterNetworkLifecycleStateTerminated,
	"RUNNING":      ClusterNetworkLifecycleStateRunning,
}

var mappingClusterNetworkLifecycleStateEnumLowerCase = map[string]ClusterNetworkLifecycleStateEnum{
	"provisioning": ClusterNetworkLifecycleStateProvisioning,
	"scaling":      ClusterNetworkLifecycleStateScaling,
	"starting":     ClusterNetworkLifecycleStateStarting,
	"stopping":     ClusterNetworkLifecycleStateStopping,
	"terminating":  ClusterNetworkLifecycleStateTerminating,
	"stopped":      ClusterNetworkLifecycleStateStopped,
	"terminated":   ClusterNetworkLifecycleStateTerminated,
	"running":      ClusterNetworkLifecycleStateRunning,
}

// GetClusterNetworkLifecycleStateEnumValues Enumerates the set of values for ClusterNetworkLifecycleStateEnum
func GetClusterNetworkLifecycleStateEnumValues() []ClusterNetworkLifecycleStateEnum {
	values := make([]ClusterNetworkLifecycleStateEnum, 0)
	for _, v := range mappingClusterNetworkLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetClusterNetworkLifecycleStateEnumStringValues Enumerates the set of values in String for ClusterNetworkLifecycleStateEnum
func GetClusterNetworkLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"SCALING",
		"STARTING",
		"STOPPING",
		"TERMINATING",
		"STOPPED",
		"TERMINATED",
		"RUNNING",
	}
}

// GetMappingClusterNetworkLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingClusterNetworkLifecycleStateEnum(val string) (ClusterNetworkLifecycleStateEnum, bool) {
	enum, ok := mappingClusterNetworkLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
