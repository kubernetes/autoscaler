// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Container Engine for Kubernetes API
//
// API for the Container Engine for Kubernetes service. Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Container Engine for Kubernetes (https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
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

var mappingClusterMigrateToNativeVcnStatusState = map[string]ClusterMigrateToNativeVcnStatusStateEnum{
	"NOT_STARTED":          ClusterMigrateToNativeVcnStatusStateNotStarted,
	"REQUESTED":            ClusterMigrateToNativeVcnStatusStateRequested,
	"IN_PROGRESS":          ClusterMigrateToNativeVcnStatusStateInProgress,
	"PENDING_DECOMMISSION": ClusterMigrateToNativeVcnStatusStatePendingDecommission,
	"COMPLETED":            ClusterMigrateToNativeVcnStatusStateCompleted,
}

// GetClusterMigrateToNativeVcnStatusStateEnumValues Enumerates the set of values for ClusterMigrateToNativeVcnStatusStateEnum
func GetClusterMigrateToNativeVcnStatusStateEnumValues() []ClusterMigrateToNativeVcnStatusStateEnum {
	values := make([]ClusterMigrateToNativeVcnStatusStateEnum, 0)
	for _, v := range mappingClusterMigrateToNativeVcnStatusState {
		values = append(values, v)
	}
	return values
}
