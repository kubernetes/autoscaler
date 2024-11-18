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

// InstancePool An instance pool is a set of instances within the same region that are managed as a group.
// For more information about instance pools and instance configurations, see
// Managing Compute Instances (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/instancemanagement.htm).
type InstancePool struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the instance pool.
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the instance
	// pool.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the instance configuration associated
	// with the instance pool.
	InstanceConfigurationId *string `mandatory:"true" json:"instanceConfigurationId"`

	// The current state of the instance pool.
	LifecycleState InstancePoolLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The placement configurations for the instance pool.
	PlacementConfigurations []InstancePoolPlacementConfiguration `mandatory:"true" json:"placementConfigurations"`

	// The number of instances that should be in the instance pool.
	Size *int `mandatory:"true" json:"size"`

	// The date and time the instance pool was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

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

	// The load balancers attached to the instance pool.
	LoadBalancers []InstancePoolLoadBalancerAttachment `mandatory:"false" json:"loadBalancers"`

	// A user-friendly formatter for the instance pool's instances. Instance displaynames follow the format.
	// The formatter does not retroactively change instance's displaynames, only instance displaynames in the future follow the format
	InstanceDisplayNameFormatter *string `mandatory:"false" json:"instanceDisplayNameFormatter"`

	// A user-friendly formatter for the instance pool's instances. Instance hostnames follow the format.
	// The formatter does not retroactively change instance's hostnames, only instance hostnames in the future follow the format
	InstanceHostnameFormatter *string `mandatory:"false" json:"instanceHostnameFormatter"`
}

func (m InstancePool) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstancePool) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingInstancePoolLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInstancePoolLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstancePoolLifecycleStateEnum Enum with underlying type: string
type InstancePoolLifecycleStateEnum string

// Set of constants representing the allowable values for InstancePoolLifecycleStateEnum
const (
	InstancePoolLifecycleStateProvisioning InstancePoolLifecycleStateEnum = "PROVISIONING"
	InstancePoolLifecycleStateScaling      InstancePoolLifecycleStateEnum = "SCALING"
	InstancePoolLifecycleStateStarting     InstancePoolLifecycleStateEnum = "STARTING"
	InstancePoolLifecycleStateStopping     InstancePoolLifecycleStateEnum = "STOPPING"
	InstancePoolLifecycleStateTerminating  InstancePoolLifecycleStateEnum = "TERMINATING"
	InstancePoolLifecycleStateStopped      InstancePoolLifecycleStateEnum = "STOPPED"
	InstancePoolLifecycleStateTerminated   InstancePoolLifecycleStateEnum = "TERMINATED"
	InstancePoolLifecycleStateRunning      InstancePoolLifecycleStateEnum = "RUNNING"
)

var mappingInstancePoolLifecycleStateEnum = map[string]InstancePoolLifecycleStateEnum{
	"PROVISIONING": InstancePoolLifecycleStateProvisioning,
	"SCALING":      InstancePoolLifecycleStateScaling,
	"STARTING":     InstancePoolLifecycleStateStarting,
	"STOPPING":     InstancePoolLifecycleStateStopping,
	"TERMINATING":  InstancePoolLifecycleStateTerminating,
	"STOPPED":      InstancePoolLifecycleStateStopped,
	"TERMINATED":   InstancePoolLifecycleStateTerminated,
	"RUNNING":      InstancePoolLifecycleStateRunning,
}

var mappingInstancePoolLifecycleStateEnumLowerCase = map[string]InstancePoolLifecycleStateEnum{
	"provisioning": InstancePoolLifecycleStateProvisioning,
	"scaling":      InstancePoolLifecycleStateScaling,
	"starting":     InstancePoolLifecycleStateStarting,
	"stopping":     InstancePoolLifecycleStateStopping,
	"terminating":  InstancePoolLifecycleStateTerminating,
	"stopped":      InstancePoolLifecycleStateStopped,
	"terminated":   InstancePoolLifecycleStateTerminated,
	"running":      InstancePoolLifecycleStateRunning,
}

// GetInstancePoolLifecycleStateEnumValues Enumerates the set of values for InstancePoolLifecycleStateEnum
func GetInstancePoolLifecycleStateEnumValues() []InstancePoolLifecycleStateEnum {
	values := make([]InstancePoolLifecycleStateEnum, 0)
	for _, v := range mappingInstancePoolLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInstancePoolLifecycleStateEnumStringValues Enumerates the set of values in String for InstancePoolLifecycleStateEnum
func GetInstancePoolLifecycleStateEnumStringValues() []string {
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

// GetMappingInstancePoolLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstancePoolLifecycleStateEnum(val string) (InstancePoolLifecycleStateEnum, bool) {
	enum, ok := mappingInstancePoolLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
