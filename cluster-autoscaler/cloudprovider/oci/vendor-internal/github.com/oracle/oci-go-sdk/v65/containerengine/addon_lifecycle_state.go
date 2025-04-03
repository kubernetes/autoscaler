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
	"strings"
)

// AddonLifecycleStateEnum Enum with underlying type: string
type AddonLifecycleStateEnum string

// Set of constants representing the allowable values for AddonLifecycleStateEnum
const (
	AddonLifecycleStateCreating       AddonLifecycleStateEnum = "CREATING"
	AddonLifecycleStateActive         AddonLifecycleStateEnum = "ACTIVE"
	AddonLifecycleStateDeleting       AddonLifecycleStateEnum = "DELETING"
	AddonLifecycleStateDeleted        AddonLifecycleStateEnum = "DELETED"
	AddonLifecycleStateUpdating       AddonLifecycleStateEnum = "UPDATING"
	AddonLifecycleStateNeedsAttention AddonLifecycleStateEnum = "NEEDS_ATTENTION"
	AddonLifecycleStateFailed         AddonLifecycleStateEnum = "FAILED"
)

var mappingAddonLifecycleStateEnum = map[string]AddonLifecycleStateEnum{
	"CREATING":        AddonLifecycleStateCreating,
	"ACTIVE":          AddonLifecycleStateActive,
	"DELETING":        AddonLifecycleStateDeleting,
	"DELETED":         AddonLifecycleStateDeleted,
	"UPDATING":        AddonLifecycleStateUpdating,
	"NEEDS_ATTENTION": AddonLifecycleStateNeedsAttention,
	"FAILED":          AddonLifecycleStateFailed,
}

var mappingAddonLifecycleStateEnumLowerCase = map[string]AddonLifecycleStateEnum{
	"creating":        AddonLifecycleStateCreating,
	"active":          AddonLifecycleStateActive,
	"deleting":        AddonLifecycleStateDeleting,
	"deleted":         AddonLifecycleStateDeleted,
	"updating":        AddonLifecycleStateUpdating,
	"needs_attention": AddonLifecycleStateNeedsAttention,
	"failed":          AddonLifecycleStateFailed,
}

// GetAddonLifecycleStateEnumValues Enumerates the set of values for AddonLifecycleStateEnum
func GetAddonLifecycleStateEnumValues() []AddonLifecycleStateEnum {
	values := make([]AddonLifecycleStateEnum, 0)
	for _, v := range mappingAddonLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetAddonLifecycleStateEnumStringValues Enumerates the set of values in String for AddonLifecycleStateEnum
func GetAddonLifecycleStateEnumStringValues() []string {
	return []string{
		"CREATING",
		"ACTIVE",
		"DELETING",
		"DELETED",
		"UPDATING",
		"NEEDS_ATTENTION",
		"FAILED",
	}
}

// GetMappingAddonLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingAddonLifecycleStateEnum(val string) (AddonLifecycleStateEnum, bool) {
	enum, ok := mappingAddonLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
