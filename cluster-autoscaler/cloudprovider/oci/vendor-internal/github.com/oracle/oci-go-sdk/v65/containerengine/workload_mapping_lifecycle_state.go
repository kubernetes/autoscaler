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

// WorkloadMappingLifecycleStateEnum Enum with underlying type: string
type WorkloadMappingLifecycleStateEnum string

// Set of constants representing the allowable values for WorkloadMappingLifecycleStateEnum
const (
	WorkloadMappingLifecycleStateCreating WorkloadMappingLifecycleStateEnum = "CREATING"
	WorkloadMappingLifecycleStateActive   WorkloadMappingLifecycleStateEnum = "ACTIVE"
	WorkloadMappingLifecycleStateFailed   WorkloadMappingLifecycleStateEnum = "FAILED"
	WorkloadMappingLifecycleStateDeleting WorkloadMappingLifecycleStateEnum = "DELETING"
	WorkloadMappingLifecycleStateDeleted  WorkloadMappingLifecycleStateEnum = "DELETED"
	WorkloadMappingLifecycleStateUpdating WorkloadMappingLifecycleStateEnum = "UPDATING"
)

var mappingWorkloadMappingLifecycleStateEnum = map[string]WorkloadMappingLifecycleStateEnum{
	"CREATING": WorkloadMappingLifecycleStateCreating,
	"ACTIVE":   WorkloadMappingLifecycleStateActive,
	"FAILED":   WorkloadMappingLifecycleStateFailed,
	"DELETING": WorkloadMappingLifecycleStateDeleting,
	"DELETED":  WorkloadMappingLifecycleStateDeleted,
	"UPDATING": WorkloadMappingLifecycleStateUpdating,
}

var mappingWorkloadMappingLifecycleStateEnumLowerCase = map[string]WorkloadMappingLifecycleStateEnum{
	"creating": WorkloadMappingLifecycleStateCreating,
	"active":   WorkloadMappingLifecycleStateActive,
	"failed":   WorkloadMappingLifecycleStateFailed,
	"deleting": WorkloadMappingLifecycleStateDeleting,
	"deleted":  WorkloadMappingLifecycleStateDeleted,
	"updating": WorkloadMappingLifecycleStateUpdating,
}

// GetWorkloadMappingLifecycleStateEnumValues Enumerates the set of values for WorkloadMappingLifecycleStateEnum
func GetWorkloadMappingLifecycleStateEnumValues() []WorkloadMappingLifecycleStateEnum {
	values := make([]WorkloadMappingLifecycleStateEnum, 0)
	for _, v := range mappingWorkloadMappingLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetWorkloadMappingLifecycleStateEnumStringValues Enumerates the set of values in String for WorkloadMappingLifecycleStateEnum
func GetWorkloadMappingLifecycleStateEnumStringValues() []string {
	return []string{
		"CREATING",
		"ACTIVE",
		"FAILED",
		"DELETING",
		"DELETED",
		"UPDATING",
	}
}

// GetMappingWorkloadMappingLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingWorkloadMappingLifecycleStateEnum(val string) (WorkloadMappingLifecycleStateEnum, bool) {
	enum, ok := mappingWorkloadMappingLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
