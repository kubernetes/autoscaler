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

// InstancePoolSummary Summary information for an instance pool.
type InstancePoolSummary struct {

	// The OCID of the instance pool.
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the compartment containing the instance pool.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the instance configuration associated with the instance pool.
	InstanceConfigurationId *string `mandatory:"true" json:"instanceConfigurationId"`

	// The current state of the instance pool.
	LifecycleState InstancePoolSummaryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The availability domains for the instance pool.
	AvailabilityDomains []string `mandatory:"true" json:"availabilityDomains"`

	// The number of instances that should be in the instance pool.
	Size *int `mandatory:"true" json:"size"`

	// The date and time the instance pool was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m InstancePoolSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstancePoolSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingInstancePoolSummaryLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInstancePoolSummaryLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstancePoolSummaryLifecycleStateEnum Enum with underlying type: string
type InstancePoolSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for InstancePoolSummaryLifecycleStateEnum
const (
	InstancePoolSummaryLifecycleStateProvisioning InstancePoolSummaryLifecycleStateEnum = "PROVISIONING"
	InstancePoolSummaryLifecycleStateScaling      InstancePoolSummaryLifecycleStateEnum = "SCALING"
	InstancePoolSummaryLifecycleStateStarting     InstancePoolSummaryLifecycleStateEnum = "STARTING"
	InstancePoolSummaryLifecycleStateStopping     InstancePoolSummaryLifecycleStateEnum = "STOPPING"
	InstancePoolSummaryLifecycleStateTerminating  InstancePoolSummaryLifecycleStateEnum = "TERMINATING"
	InstancePoolSummaryLifecycleStateStopped      InstancePoolSummaryLifecycleStateEnum = "STOPPED"
	InstancePoolSummaryLifecycleStateTerminated   InstancePoolSummaryLifecycleStateEnum = "TERMINATED"
	InstancePoolSummaryLifecycleStateRunning      InstancePoolSummaryLifecycleStateEnum = "RUNNING"
)

var mappingInstancePoolSummaryLifecycleStateEnum = map[string]InstancePoolSummaryLifecycleStateEnum{
	"PROVISIONING": InstancePoolSummaryLifecycleStateProvisioning,
	"SCALING":      InstancePoolSummaryLifecycleStateScaling,
	"STARTING":     InstancePoolSummaryLifecycleStateStarting,
	"STOPPING":     InstancePoolSummaryLifecycleStateStopping,
	"TERMINATING":  InstancePoolSummaryLifecycleStateTerminating,
	"STOPPED":      InstancePoolSummaryLifecycleStateStopped,
	"TERMINATED":   InstancePoolSummaryLifecycleStateTerminated,
	"RUNNING":      InstancePoolSummaryLifecycleStateRunning,
}

var mappingInstancePoolSummaryLifecycleStateEnumLowerCase = map[string]InstancePoolSummaryLifecycleStateEnum{
	"provisioning": InstancePoolSummaryLifecycleStateProvisioning,
	"scaling":      InstancePoolSummaryLifecycleStateScaling,
	"starting":     InstancePoolSummaryLifecycleStateStarting,
	"stopping":     InstancePoolSummaryLifecycleStateStopping,
	"terminating":  InstancePoolSummaryLifecycleStateTerminating,
	"stopped":      InstancePoolSummaryLifecycleStateStopped,
	"terminated":   InstancePoolSummaryLifecycleStateTerminated,
	"running":      InstancePoolSummaryLifecycleStateRunning,
}

// GetInstancePoolSummaryLifecycleStateEnumValues Enumerates the set of values for InstancePoolSummaryLifecycleStateEnum
func GetInstancePoolSummaryLifecycleStateEnumValues() []InstancePoolSummaryLifecycleStateEnum {
	values := make([]InstancePoolSummaryLifecycleStateEnum, 0)
	for _, v := range mappingInstancePoolSummaryLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInstancePoolSummaryLifecycleStateEnumStringValues Enumerates the set of values in String for InstancePoolSummaryLifecycleStateEnum
func GetInstancePoolSummaryLifecycleStateEnumStringValues() []string {
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

// GetMappingInstancePoolSummaryLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstancePoolSummaryLifecycleStateEnum(val string) (InstancePoolSummaryLifecycleStateEnum, bool) {
	enum, ok := mappingInstancePoolSummaryLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
