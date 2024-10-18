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

// ClusterLifecycleStateEnum Enum with underlying type: string
type ClusterLifecycleStateEnum string

// Set of constants representing the allowable values for ClusterLifecycleStateEnum
const (
	ClusterLifecycleStateCreating ClusterLifecycleStateEnum = "CREATING"
	ClusterLifecycleStateActive   ClusterLifecycleStateEnum = "ACTIVE"
	ClusterLifecycleStateFailed   ClusterLifecycleStateEnum = "FAILED"
	ClusterLifecycleStateDeleting ClusterLifecycleStateEnum = "DELETING"
	ClusterLifecycleStateDeleted  ClusterLifecycleStateEnum = "DELETED"
	ClusterLifecycleStateUpdating ClusterLifecycleStateEnum = "UPDATING"
)

var mappingClusterLifecycleStateEnum = map[string]ClusterLifecycleStateEnum{
	"CREATING": ClusterLifecycleStateCreating,
	"ACTIVE":   ClusterLifecycleStateActive,
	"FAILED":   ClusterLifecycleStateFailed,
	"DELETING": ClusterLifecycleStateDeleting,
	"DELETED":  ClusterLifecycleStateDeleted,
	"UPDATING": ClusterLifecycleStateUpdating,
}

var mappingClusterLifecycleStateEnumLowerCase = map[string]ClusterLifecycleStateEnum{
	"creating": ClusterLifecycleStateCreating,
	"active":   ClusterLifecycleStateActive,
	"failed":   ClusterLifecycleStateFailed,
	"deleting": ClusterLifecycleStateDeleting,
	"deleted":  ClusterLifecycleStateDeleted,
	"updating": ClusterLifecycleStateUpdating,
}

// GetClusterLifecycleStateEnumValues Enumerates the set of values for ClusterLifecycleStateEnum
func GetClusterLifecycleStateEnumValues() []ClusterLifecycleStateEnum {
	values := make([]ClusterLifecycleStateEnum, 0)
	for _, v := range mappingClusterLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetClusterLifecycleStateEnumStringValues Enumerates the set of values in String for ClusterLifecycleStateEnum
func GetClusterLifecycleStateEnumStringValues() []string {
	return []string{
		"CREATING",
		"ACTIVE",
		"FAILED",
		"DELETING",
		"DELETED",
		"UPDATING",
	}
}

// GetMappingClusterLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingClusterLifecycleStateEnum(val string) (ClusterLifecycleStateEnum, bool) {
	enum, ok := mappingClusterLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
