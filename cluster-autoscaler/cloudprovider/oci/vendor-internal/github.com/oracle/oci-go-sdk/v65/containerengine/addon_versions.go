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

// AddonVersions The properties that define a work request resource.
type AddonVersions struct {

	// Current state of the addon, only active will be visible to customer, visibility of versions in other status will be filtered  based on limits property.
	Status AddonVersionsStatusEnum `mandatory:"false" json:"status,omitempty"`

	// Version number, need be comparable within an addon.
	VersionNumber *string `mandatory:"false" json:"versionNumber"`

	// Information about the addon version.
	Description *string `mandatory:"false" json:"description"`

	// The range of kubernetes versions an addon can be configured.
	KubernetesVersionFilters *KubernetesVersionsFilters `mandatory:"false" json:"kubernetesVersionFilters"`

	// Addon version configuration details.
	Configurations []AddonVersionConfiguration `mandatory:"false" json:"configurations"`
}

func (m AddonVersions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m AddonVersions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingAddonVersionsStatusEnum(string(m.Status)); !ok && m.Status != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Status: %s. Supported values are: %s.", m.Status, strings.Join(GetAddonVersionsStatusEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// AddonVersionsStatusEnum Enum with underlying type: string
type AddonVersionsStatusEnum string

// Set of constants representing the allowable values for AddonVersionsStatusEnum
const (
	AddonVersionsStatusActive     AddonVersionsStatusEnum = "ACTIVE"
	AddonVersionsStatusDeprecated AddonVersionsStatusEnum = "DEPRECATED"
	AddonVersionsStatusPreview    AddonVersionsStatusEnum = "PREVIEW"
	AddonVersionsStatusRecalled   AddonVersionsStatusEnum = "RECALLED"
)

var mappingAddonVersionsStatusEnum = map[string]AddonVersionsStatusEnum{
	"ACTIVE":     AddonVersionsStatusActive,
	"DEPRECATED": AddonVersionsStatusDeprecated,
	"PREVIEW":    AddonVersionsStatusPreview,
	"RECALLED":   AddonVersionsStatusRecalled,
}

var mappingAddonVersionsStatusEnumLowerCase = map[string]AddonVersionsStatusEnum{
	"active":     AddonVersionsStatusActive,
	"deprecated": AddonVersionsStatusDeprecated,
	"preview":    AddonVersionsStatusPreview,
	"recalled":   AddonVersionsStatusRecalled,
}

// GetAddonVersionsStatusEnumValues Enumerates the set of values for AddonVersionsStatusEnum
func GetAddonVersionsStatusEnumValues() []AddonVersionsStatusEnum {
	values := make([]AddonVersionsStatusEnum, 0)
	for _, v := range mappingAddonVersionsStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetAddonVersionsStatusEnumStringValues Enumerates the set of values in String for AddonVersionsStatusEnum
func GetAddonVersionsStatusEnumStringValues() []string {
	return []string{
		"ACTIVE",
		"DEPRECATED",
		"PREVIEW",
		"RECALLED",
	}
}

// GetMappingAddonVersionsStatusEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingAddonVersionsStatusEnum(val string) (AddonVersionsStatusEnum, bool) {
	enum, ok := mappingAddonVersionsStatusEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
