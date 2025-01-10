// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.cloud.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
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

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpgradeStatus) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingUpgradeStatusStatusEnum(string(m.Status)); !ok && m.Status != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Status: %s. Supported values are: %s.", m.Status, strings.Join(GetUpgradeStatusStatusEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UpgradeStatusStatusEnum Enum with underlying type: string
type UpgradeStatusStatusEnum string

// Set of constants representing the allowable values for UpgradeStatusStatusEnum
const (
	UpgradeStatusStatusNotUpgraded UpgradeStatusStatusEnum = "NOT_UPGRADED"
	UpgradeStatusStatusInProgress  UpgradeStatusStatusEnum = "IN_PROGRESS"
	UpgradeStatusStatusUpgraded    UpgradeStatusStatusEnum = "UPGRADED"
)

var mappingUpgradeStatusStatusEnum = map[string]UpgradeStatusStatusEnum{
	"NOT_UPGRADED": UpgradeStatusStatusNotUpgraded,
	"IN_PROGRESS":  UpgradeStatusStatusInProgress,
	"UPGRADED":     UpgradeStatusStatusUpgraded,
}

var mappingUpgradeStatusStatusEnumLowerCase = map[string]UpgradeStatusStatusEnum{
	"not_upgraded": UpgradeStatusStatusNotUpgraded,
	"in_progress":  UpgradeStatusStatusInProgress,
	"upgraded":     UpgradeStatusStatusUpgraded,
}

// GetUpgradeStatusStatusEnumValues Enumerates the set of values for UpgradeStatusStatusEnum
func GetUpgradeStatusStatusEnumValues() []UpgradeStatusStatusEnum {
	values := make([]UpgradeStatusStatusEnum, 0)
	for _, v := range mappingUpgradeStatusStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetUpgradeStatusStatusEnumStringValues Enumerates the set of values in String for UpgradeStatusStatusEnum
func GetUpgradeStatusStatusEnumStringValues() []string {
	return []string{
		"NOT_UPGRADED",
		"IN_PROGRESS",
		"UPGRADED",
	}
}

// GetMappingUpgradeStatusStatusEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpgradeStatusStatusEnum(val string) (UpgradeStatusStatusEnum, bool) {
	enum, ok := mappingUpgradeStatusStatusEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
