// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// DrgUpgradeState A dynamic routing gateway (DRG) can be in different states during upgrade -  Classical, Migrated, Upgraded(Transit-Hub)
type DrgUpgradeState struct {

	// The type of the DRG.
	State DrgUpgradeStateStateEnum `mandatory:"false" json:"state,omitempty"`
}

func (m DrgUpgradeState) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DrgUpgradeState) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingDrgUpgradeStateStateEnum[string(m.State)]; !ok && m.State != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for State: %s. Supported values are: %s.", m.State, strings.Join(GetDrgUpgradeStateStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// DrgUpgradeStateStateEnum Enum with underlying type: string
type DrgUpgradeStateStateEnum string

// Set of constants representing the allowable values for DrgUpgradeStateStateEnum
const (
	DrgUpgradeStateStateClassical DrgUpgradeStateStateEnum = "CLASSICAL"
	DrgUpgradeStateStateMigrated  DrgUpgradeStateStateEnum = "MIGRATED"
	DrgUpgradeStateStateUpgraded  DrgUpgradeStateStateEnum = "UPGRADED"
)

var mappingDrgUpgradeStateStateEnum = map[string]DrgUpgradeStateStateEnum{
	"CLASSICAL": DrgUpgradeStateStateClassical,
	"MIGRATED":  DrgUpgradeStateStateMigrated,
	"UPGRADED":  DrgUpgradeStateStateUpgraded,
}

// GetDrgUpgradeStateStateEnumValues Enumerates the set of values for DrgUpgradeStateStateEnum
func GetDrgUpgradeStateStateEnumValues() []DrgUpgradeStateStateEnum {
	values := make([]DrgUpgradeStateStateEnum, 0)
	for _, v := range mappingDrgUpgradeStateStateEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgUpgradeStateStateEnumStringValues Enumerates the set of values in String for DrgUpgradeStateStateEnum
func GetDrgUpgradeStateStateEnumStringValues() []string {
	return []string{
		"CLASSICAL",
		"MIGRATED",
		"UPGRADED",
	}
}
