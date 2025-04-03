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

// ClusterMigrateToNativeVcnStatus Information regarding a cluster's move to Native VCN.
type ClusterMigrateToNativeVcnStatus struct {

	// The current migration status of the cluster.
	State ClusterMigrateToNativeVcnStatusStateEnum `mandatory:"true" json:"state"`

	// The date and time the non-native VCN is due to be decommissioned.
	TimeDecommissionScheduled *common.SDKTime `mandatory:"false" json:"timeDecommissionScheduled"`
}

func (m ClusterMigrateToNativeVcnStatus) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ClusterMigrateToNativeVcnStatus) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingClusterMigrateToNativeVcnStatusStateEnum(string(m.State)); !ok && m.State != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for State: %s. Supported values are: %s.", m.State, strings.Join(GetClusterMigrateToNativeVcnStatusStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ClusterMigrateToNativeVcnStatusStateEnum Enum with underlying type: string
type ClusterMigrateToNativeVcnStatusStateEnum string

// Set of constants representing the allowable values for ClusterMigrateToNativeVcnStatusStateEnum
const (
	ClusterMigrateToNativeVcnStatusStateNotStarted          ClusterMigrateToNativeVcnStatusStateEnum = "NOT_STARTED"
	ClusterMigrateToNativeVcnStatusStateRequested           ClusterMigrateToNativeVcnStatusStateEnum = "REQUESTED"
	ClusterMigrateToNativeVcnStatusStateInProgress          ClusterMigrateToNativeVcnStatusStateEnum = "IN_PROGRESS"
	ClusterMigrateToNativeVcnStatusStatePendingDecommission ClusterMigrateToNativeVcnStatusStateEnum = "PENDING_DECOMMISSION"
	ClusterMigrateToNativeVcnStatusStateCompleted           ClusterMigrateToNativeVcnStatusStateEnum = "COMPLETED"
)

var mappingClusterMigrateToNativeVcnStatusStateEnum = map[string]ClusterMigrateToNativeVcnStatusStateEnum{
	"NOT_STARTED":          ClusterMigrateToNativeVcnStatusStateNotStarted,
	"REQUESTED":            ClusterMigrateToNativeVcnStatusStateRequested,
	"IN_PROGRESS":          ClusterMigrateToNativeVcnStatusStateInProgress,
	"PENDING_DECOMMISSION": ClusterMigrateToNativeVcnStatusStatePendingDecommission,
	"COMPLETED":            ClusterMigrateToNativeVcnStatusStateCompleted,
}

var mappingClusterMigrateToNativeVcnStatusStateEnumLowerCase = map[string]ClusterMigrateToNativeVcnStatusStateEnum{
	"not_started":          ClusterMigrateToNativeVcnStatusStateNotStarted,
	"requested":            ClusterMigrateToNativeVcnStatusStateRequested,
	"in_progress":          ClusterMigrateToNativeVcnStatusStateInProgress,
	"pending_decommission": ClusterMigrateToNativeVcnStatusStatePendingDecommission,
	"completed":            ClusterMigrateToNativeVcnStatusStateCompleted,
}

// GetClusterMigrateToNativeVcnStatusStateEnumValues Enumerates the set of values for ClusterMigrateToNativeVcnStatusStateEnum
func GetClusterMigrateToNativeVcnStatusStateEnumValues() []ClusterMigrateToNativeVcnStatusStateEnum {
	values := make([]ClusterMigrateToNativeVcnStatusStateEnum, 0)
	for _, v := range mappingClusterMigrateToNativeVcnStatusStateEnum {
		values = append(values, v)
	}
	return values
}

// GetClusterMigrateToNativeVcnStatusStateEnumStringValues Enumerates the set of values in String for ClusterMigrateToNativeVcnStatusStateEnum
func GetClusterMigrateToNativeVcnStatusStateEnumStringValues() []string {
	return []string{
		"NOT_STARTED",
		"REQUESTED",
		"IN_PROGRESS",
		"PENDING_DECOMMISSION",
		"COMPLETED",
	}
}

// GetMappingClusterMigrateToNativeVcnStatusStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingClusterMigrateToNativeVcnStatusStateEnum(val string) (ClusterMigrateToNativeVcnStatusStateEnum, bool) {
	enum, ok := mappingClusterMigrateToNativeVcnStatusStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
