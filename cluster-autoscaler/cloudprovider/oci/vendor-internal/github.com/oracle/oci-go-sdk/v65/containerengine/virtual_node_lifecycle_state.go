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

// VirtualNodeLifecycleStateEnum Enum with underlying type: string
type VirtualNodeLifecycleStateEnum string

// Set of constants representing the allowable values for VirtualNodeLifecycleStateEnum
const (
	VirtualNodeLifecycleStateCreating       VirtualNodeLifecycleStateEnum = "CREATING"
	VirtualNodeLifecycleStateActive         VirtualNodeLifecycleStateEnum = "ACTIVE"
	VirtualNodeLifecycleStateUpdating       VirtualNodeLifecycleStateEnum = "UPDATING"
	VirtualNodeLifecycleStateDeleting       VirtualNodeLifecycleStateEnum = "DELETING"
	VirtualNodeLifecycleStateDeleted        VirtualNodeLifecycleStateEnum = "DELETED"
	VirtualNodeLifecycleStateFailed         VirtualNodeLifecycleStateEnum = "FAILED"
	VirtualNodeLifecycleStateNeedsAttention VirtualNodeLifecycleStateEnum = "NEEDS_ATTENTION"
)

var mappingVirtualNodeLifecycleStateEnum = map[string]VirtualNodeLifecycleStateEnum{
	"CREATING":        VirtualNodeLifecycleStateCreating,
	"ACTIVE":          VirtualNodeLifecycleStateActive,
	"UPDATING":        VirtualNodeLifecycleStateUpdating,
	"DELETING":        VirtualNodeLifecycleStateDeleting,
	"DELETED":         VirtualNodeLifecycleStateDeleted,
	"FAILED":          VirtualNodeLifecycleStateFailed,
	"NEEDS_ATTENTION": VirtualNodeLifecycleStateNeedsAttention,
}

var mappingVirtualNodeLifecycleStateEnumLowerCase = map[string]VirtualNodeLifecycleStateEnum{
	"creating":        VirtualNodeLifecycleStateCreating,
	"active":          VirtualNodeLifecycleStateActive,
	"updating":        VirtualNodeLifecycleStateUpdating,
	"deleting":        VirtualNodeLifecycleStateDeleting,
	"deleted":         VirtualNodeLifecycleStateDeleted,
	"failed":          VirtualNodeLifecycleStateFailed,
	"needs_attention": VirtualNodeLifecycleStateNeedsAttention,
}

// GetVirtualNodeLifecycleStateEnumValues Enumerates the set of values for VirtualNodeLifecycleStateEnum
func GetVirtualNodeLifecycleStateEnumValues() []VirtualNodeLifecycleStateEnum {
	values := make([]VirtualNodeLifecycleStateEnum, 0)
	for _, v := range mappingVirtualNodeLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualNodeLifecycleStateEnumStringValues Enumerates the set of values in String for VirtualNodeLifecycleStateEnum
func GetVirtualNodeLifecycleStateEnumStringValues() []string {
	return []string{
		"CREATING",
		"ACTIVE",
		"UPDATING",
		"DELETING",
		"DELETED",
		"FAILED",
		"NEEDS_ATTENTION",
	}
}

// GetMappingVirtualNodeLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualNodeLifecycleStateEnum(val string) (VirtualNodeLifecycleStateEnum, bool) {
	enum, ok := mappingVirtualNodeLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
