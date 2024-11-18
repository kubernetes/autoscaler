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

// NodePoolLifecycleStateEnum Enum with underlying type: string
type NodePoolLifecycleStateEnum string

// Set of constants representing the allowable values for NodePoolLifecycleStateEnum
const (
	NodePoolLifecycleStateDeleted        NodePoolLifecycleStateEnum = "DELETED"
	NodePoolLifecycleStateCreating       NodePoolLifecycleStateEnum = "CREATING"
	NodePoolLifecycleStateActive         NodePoolLifecycleStateEnum = "ACTIVE"
	NodePoolLifecycleStateUpdating       NodePoolLifecycleStateEnum = "UPDATING"
	NodePoolLifecycleStateDeleting       NodePoolLifecycleStateEnum = "DELETING"
	NodePoolLifecycleStateFailed         NodePoolLifecycleStateEnum = "FAILED"
	NodePoolLifecycleStateInactive       NodePoolLifecycleStateEnum = "INACTIVE"
	NodePoolLifecycleStateNeedsAttention NodePoolLifecycleStateEnum = "NEEDS_ATTENTION"
)

var mappingNodePoolLifecycleStateEnum = map[string]NodePoolLifecycleStateEnum{
	"DELETED":         NodePoolLifecycleStateDeleted,
	"CREATING":        NodePoolLifecycleStateCreating,
	"ACTIVE":          NodePoolLifecycleStateActive,
	"UPDATING":        NodePoolLifecycleStateUpdating,
	"DELETING":        NodePoolLifecycleStateDeleting,
	"FAILED":          NodePoolLifecycleStateFailed,
	"INACTIVE":        NodePoolLifecycleStateInactive,
	"NEEDS_ATTENTION": NodePoolLifecycleStateNeedsAttention,
}

var mappingNodePoolLifecycleStateEnumLowerCase = map[string]NodePoolLifecycleStateEnum{
	"deleted":         NodePoolLifecycleStateDeleted,
	"creating":        NodePoolLifecycleStateCreating,
	"active":          NodePoolLifecycleStateActive,
	"updating":        NodePoolLifecycleStateUpdating,
	"deleting":        NodePoolLifecycleStateDeleting,
	"failed":          NodePoolLifecycleStateFailed,
	"inactive":        NodePoolLifecycleStateInactive,
	"needs_attention": NodePoolLifecycleStateNeedsAttention,
}

// GetNodePoolLifecycleStateEnumValues Enumerates the set of values for NodePoolLifecycleStateEnum
func GetNodePoolLifecycleStateEnumValues() []NodePoolLifecycleStateEnum {
	values := make([]NodePoolLifecycleStateEnum, 0)
	for _, v := range mappingNodePoolLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetNodePoolLifecycleStateEnumStringValues Enumerates the set of values in String for NodePoolLifecycleStateEnum
func GetNodePoolLifecycleStateEnumStringValues() []string {
	return []string{
		"DELETED",
		"CREATING",
		"ACTIVE",
		"UPDATING",
		"DELETING",
		"FAILED",
		"INACTIVE",
		"NEEDS_ATTENTION",
	}
}

// GetMappingNodePoolLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingNodePoolLifecycleStateEnum(val string) (NodePoolLifecycleStateEnum, bool) {
	enum, ok := mappingNodePoolLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
