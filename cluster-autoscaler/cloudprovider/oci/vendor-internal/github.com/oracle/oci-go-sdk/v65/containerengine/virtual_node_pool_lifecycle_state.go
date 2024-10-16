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

// VirtualNodePoolLifecycleStateEnum Enum with underlying type: string
type VirtualNodePoolLifecycleStateEnum string

// Set of constants representing the allowable values for VirtualNodePoolLifecycleStateEnum
const (
	VirtualNodePoolLifecycleStateCreating       VirtualNodePoolLifecycleStateEnum = "CREATING"
	VirtualNodePoolLifecycleStateActive         VirtualNodePoolLifecycleStateEnum = "ACTIVE"
	VirtualNodePoolLifecycleStateUpdating       VirtualNodePoolLifecycleStateEnum = "UPDATING"
	VirtualNodePoolLifecycleStateDeleting       VirtualNodePoolLifecycleStateEnum = "DELETING"
	VirtualNodePoolLifecycleStateDeleted        VirtualNodePoolLifecycleStateEnum = "DELETED"
	VirtualNodePoolLifecycleStateFailed         VirtualNodePoolLifecycleStateEnum = "FAILED"
	VirtualNodePoolLifecycleStateNeedsAttention VirtualNodePoolLifecycleStateEnum = "NEEDS_ATTENTION"
)

var mappingVirtualNodePoolLifecycleStateEnum = map[string]VirtualNodePoolLifecycleStateEnum{
	"CREATING":        VirtualNodePoolLifecycleStateCreating,
	"ACTIVE":          VirtualNodePoolLifecycleStateActive,
	"UPDATING":        VirtualNodePoolLifecycleStateUpdating,
	"DELETING":        VirtualNodePoolLifecycleStateDeleting,
	"DELETED":         VirtualNodePoolLifecycleStateDeleted,
	"FAILED":          VirtualNodePoolLifecycleStateFailed,
	"NEEDS_ATTENTION": VirtualNodePoolLifecycleStateNeedsAttention,
}

var mappingVirtualNodePoolLifecycleStateEnumLowerCase = map[string]VirtualNodePoolLifecycleStateEnum{
	"creating":        VirtualNodePoolLifecycleStateCreating,
	"active":          VirtualNodePoolLifecycleStateActive,
	"updating":        VirtualNodePoolLifecycleStateUpdating,
	"deleting":        VirtualNodePoolLifecycleStateDeleting,
	"deleted":         VirtualNodePoolLifecycleStateDeleted,
	"failed":          VirtualNodePoolLifecycleStateFailed,
	"needs_attention": VirtualNodePoolLifecycleStateNeedsAttention,
}

// GetVirtualNodePoolLifecycleStateEnumValues Enumerates the set of values for VirtualNodePoolLifecycleStateEnum
func GetVirtualNodePoolLifecycleStateEnumValues() []VirtualNodePoolLifecycleStateEnum {
	values := make([]VirtualNodePoolLifecycleStateEnum, 0)
	for _, v := range mappingVirtualNodePoolLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualNodePoolLifecycleStateEnumStringValues Enumerates the set of values in String for VirtualNodePoolLifecycleStateEnum
func GetVirtualNodePoolLifecycleStateEnumStringValues() []string {
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

// GetMappingVirtualNodePoolLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualNodePoolLifecycleStateEnum(val string) (VirtualNodePoolLifecycleStateEnum, bool) {
	enum, ok := mappingVirtualNodePoolLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
