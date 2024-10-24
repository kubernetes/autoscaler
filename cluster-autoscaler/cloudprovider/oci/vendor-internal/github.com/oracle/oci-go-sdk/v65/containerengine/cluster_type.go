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

// ClusterTypeEnum Enum with underlying type: string
type ClusterTypeEnum string

// Set of constants representing the allowable values for ClusterTypeEnum
const (
	ClusterTypeBasicCluster    ClusterTypeEnum = "BASIC_CLUSTER"
	ClusterTypeEnhancedCluster ClusterTypeEnum = "ENHANCED_CLUSTER"
)

var mappingClusterTypeEnum = map[string]ClusterTypeEnum{
	"BASIC_CLUSTER":    ClusterTypeBasicCluster,
	"ENHANCED_CLUSTER": ClusterTypeEnhancedCluster,
}

var mappingClusterTypeEnumLowerCase = map[string]ClusterTypeEnum{
	"basic_cluster":    ClusterTypeBasicCluster,
	"enhanced_cluster": ClusterTypeEnhancedCluster,
}

// GetClusterTypeEnumValues Enumerates the set of values for ClusterTypeEnum
func GetClusterTypeEnumValues() []ClusterTypeEnum {
	values := make([]ClusterTypeEnum, 0)
	for _, v := range mappingClusterTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetClusterTypeEnumStringValues Enumerates the set of values in String for ClusterTypeEnum
func GetClusterTypeEnumStringValues() []string {
	return []string{
		"BASIC_CLUSTER",
		"ENHANCED_CLUSTER",
	}
}

// GetMappingClusterTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingClusterTypeEnum(val string) (ClusterTypeEnum, bool) {
	enum, ok := mappingClusterTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
