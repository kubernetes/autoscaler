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

// InstanceAgentPluginConfigDetails The configuration of plugins associated with this instance.
type InstanceAgentPluginConfigDetails struct {

	// The plugin name. To get a list of available plugins, use the
	// ListInstanceagentAvailablePlugins
	// operation in the Oracle Cloud Agent API. For more information about the available plugins, see
	// Managing Plugins with Oracle Cloud Agent (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/manage-plugins.htm).
	Name *string `mandatory:"true" json:"name"`

	// Whether the plugin should be enabled or disabled.
	// To enable the monitoring and management plugins, the `isMonitoringDisabled` and
	// `isManagementDisabled` attributes must also be set to false.
	DesiredState InstanceAgentPluginConfigDetailsDesiredStateEnum `mandatory:"true" json:"desiredState"`
}

func (m InstanceAgentPluginConfigDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstanceAgentPluginConfigDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingInstanceAgentPluginConfigDetailsDesiredStateEnum(string(m.DesiredState)); !ok && m.DesiredState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DesiredState: %s. Supported values are: %s.", m.DesiredState, strings.Join(GetInstanceAgentPluginConfigDetailsDesiredStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstanceAgentPluginConfigDetailsDesiredStateEnum Enum with underlying type: string
type InstanceAgentPluginConfigDetailsDesiredStateEnum string

// Set of constants representing the allowable values for InstanceAgentPluginConfigDetailsDesiredStateEnum
const (
	InstanceAgentPluginConfigDetailsDesiredStateEnabled  InstanceAgentPluginConfigDetailsDesiredStateEnum = "ENABLED"
	InstanceAgentPluginConfigDetailsDesiredStateDisabled InstanceAgentPluginConfigDetailsDesiredStateEnum = "DISABLED"
)

var mappingInstanceAgentPluginConfigDetailsDesiredStateEnum = map[string]InstanceAgentPluginConfigDetailsDesiredStateEnum{
	"ENABLED":  InstanceAgentPluginConfigDetailsDesiredStateEnabled,
	"DISABLED": InstanceAgentPluginConfigDetailsDesiredStateDisabled,
}

var mappingInstanceAgentPluginConfigDetailsDesiredStateEnumLowerCase = map[string]InstanceAgentPluginConfigDetailsDesiredStateEnum{
	"enabled":  InstanceAgentPluginConfigDetailsDesiredStateEnabled,
	"disabled": InstanceAgentPluginConfigDetailsDesiredStateDisabled,
}

// GetInstanceAgentPluginConfigDetailsDesiredStateEnumValues Enumerates the set of values for InstanceAgentPluginConfigDetailsDesiredStateEnum
func GetInstanceAgentPluginConfigDetailsDesiredStateEnumValues() []InstanceAgentPluginConfigDetailsDesiredStateEnum {
	values := make([]InstanceAgentPluginConfigDetailsDesiredStateEnum, 0)
	for _, v := range mappingInstanceAgentPluginConfigDetailsDesiredStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceAgentPluginConfigDetailsDesiredStateEnumStringValues Enumerates the set of values in String for InstanceAgentPluginConfigDetailsDesiredStateEnum
func GetInstanceAgentPluginConfigDetailsDesiredStateEnumStringValues() []string {
	return []string{
		"ENABLED",
		"DISABLED",
	}
}

// GetMappingInstanceAgentPluginConfigDetailsDesiredStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceAgentPluginConfigDetailsDesiredStateEnum(val string) (InstanceAgentPluginConfigDetailsDesiredStateEnum, bool) {
	enum, ok := mappingInstanceAgentPluginConfigDetailsDesiredStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
