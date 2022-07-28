// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// UpgradeStatus The upgrade status of a DRG.
type UpgradeStatus struct {

	// The `drgId` of the upgraded DRG.
	DrgId *string `mandatory:"true" json:"drgId"`

	// The current upgrade status of the DRG attachment.
	Status UpgradeStatusStatusEnum `mandatory:"true" json:"status"`

	// The number of upgraded connections.
	UpgradedConnections *string `mandatory:"true" json:"upgradedConnections"`
}

func (m UpgradeStatus) String() string {
	return common.PointerString(m)
}

// UpgradeStatusStatusEnum Enum with underlying type: string
type UpgradeStatusStatusEnum string

// Set of constants representing the allowable values for UpgradeStatusStatusEnum
const (
	UpgradeStatusStatusNotUpgraded UpgradeStatusStatusEnum = "NOT_UPGRADED"
	UpgradeStatusStatusInProgress  UpgradeStatusStatusEnum = "IN_PROGRESS"
	UpgradeStatusStatusUpgraded    UpgradeStatusStatusEnum = "UPGRADED"
)

var mappingUpgradeStatusStatus = map[string]UpgradeStatusStatusEnum{
	"NOT_UPGRADED": UpgradeStatusStatusNotUpgraded,
	"IN_PROGRESS":  UpgradeStatusStatusInProgress,
	"UPGRADED":     UpgradeStatusStatusUpgraded,
}

// GetUpgradeStatusStatusEnumValues Enumerates the set of values for UpgradeStatusStatusEnum
func GetUpgradeStatusStatusEnumValues() []UpgradeStatusStatusEnum {
	values := make([]UpgradeStatusStatusEnum, 0)
	for _, v := range mappingUpgradeStatusStatus {
		values = append(values, v)
	}
	return values
}
