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

// AddonOptionSummary The properties that define addon summary.
type AddonOptionSummary struct {

	// Name of the addon and it would be unique.
	Name *string `mandatory:"true" json:"name"`

	// The life cycle state of the addon.
	LifecycleState AddonOptionSummaryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Is it an essential addon for cluster operation or not.
	IsEssential *bool `mandatory:"true" json:"isEssential"`

	// The resources this work request affects.
	Versions []AddonVersions `mandatory:"true" json:"versions"`

	// Addon definition schema version to validate addon.
	AddonSchemaVersion *string `mandatory:"false" json:"addonSchemaVersion"`

	// Addon group info, a namespace concept that groups addons with similar functionalities.
	AddonGroup *string `mandatory:"false" json:"addonGroup"`

	// Description on the addon.
	Description *string `mandatory:"false" json:"description"`

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

	// The time the work request was created.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m AddonOptionSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m AddonOptionSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingAddonOptionSummaryLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetAddonOptionSummaryLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// AddonOptionSummaryLifecycleStateEnum Enum with underlying type: string
type AddonOptionSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for AddonOptionSummaryLifecycleStateEnum
const (
	AddonOptionSummaryLifecycleStateActive   AddonOptionSummaryLifecycleStateEnum = "ACTIVE"
	AddonOptionSummaryLifecycleStateInactive AddonOptionSummaryLifecycleStateEnum = "INACTIVE"
)

var mappingAddonOptionSummaryLifecycleStateEnum = map[string]AddonOptionSummaryLifecycleStateEnum{
	"ACTIVE":   AddonOptionSummaryLifecycleStateActive,
	"INACTIVE": AddonOptionSummaryLifecycleStateInactive,
}

var mappingAddonOptionSummaryLifecycleStateEnumLowerCase = map[string]AddonOptionSummaryLifecycleStateEnum{
	"active":   AddonOptionSummaryLifecycleStateActive,
	"inactive": AddonOptionSummaryLifecycleStateInactive,
}

// GetAddonOptionSummaryLifecycleStateEnumValues Enumerates the set of values for AddonOptionSummaryLifecycleStateEnum
func GetAddonOptionSummaryLifecycleStateEnumValues() []AddonOptionSummaryLifecycleStateEnum {
	values := make([]AddonOptionSummaryLifecycleStateEnum, 0)
	for _, v := range mappingAddonOptionSummaryLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetAddonOptionSummaryLifecycleStateEnumStringValues Enumerates the set of values in String for AddonOptionSummaryLifecycleStateEnum
func GetAddonOptionSummaryLifecycleStateEnumStringValues() []string {
	return []string{
		"ACTIVE",
		"INACTIVE",
	}
}

// GetMappingAddonOptionSummaryLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingAddonOptionSummaryLifecycleStateEnum(val string) (AddonOptionSummaryLifecycleStateEnum, bool) {
	enum, ok := mappingAddonOptionSummaryLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
