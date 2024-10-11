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

// InstanceConfigurationAvailabilityConfig Options for defining the availabiity of a VM instance after a maintenance event that impacts the underlying hardware.
type InstanceConfigurationAvailabilityConfig struct {

	// Whether to live migrate supported VM instances to a healthy physical VM host without
	// disrupting running instances during infrastructure maintenance events. If null, Oracle
	// chooses the best option for migrating the VM during infrastructure maintenance events.
	IsLiveMigrationPreferred *bool `mandatory:"false" json:"isLiveMigrationPreferred"`

	// The lifecycle state for an instance when it is recovered after infrastructure maintenance.
	// * `RESTORE_INSTANCE` - The instance is restored to the lifecycle state it was in before the maintenance event.
	// If the instance was running, it is automatically rebooted. This is the default action when a value is not set.
	// * `STOP_INSTANCE` - The instance is recovered in the stopped state.
	RecoveryAction InstanceConfigurationAvailabilityConfigRecoveryActionEnum `mandatory:"false" json:"recoveryAction,omitempty"`
}

func (m InstanceConfigurationAvailabilityConfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstanceConfigurationAvailabilityConfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingInstanceConfigurationAvailabilityConfigRecoveryActionEnum(string(m.RecoveryAction)); !ok && m.RecoveryAction != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RecoveryAction: %s. Supported values are: %s.", m.RecoveryAction, strings.Join(GetInstanceConfigurationAvailabilityConfigRecoveryActionEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstanceConfigurationAvailabilityConfigRecoveryActionEnum Enum with underlying type: string
type InstanceConfigurationAvailabilityConfigRecoveryActionEnum string

// Set of constants representing the allowable values for InstanceConfigurationAvailabilityConfigRecoveryActionEnum
const (
	InstanceConfigurationAvailabilityConfigRecoveryActionRestoreInstance InstanceConfigurationAvailabilityConfigRecoveryActionEnum = "RESTORE_INSTANCE"
	InstanceConfigurationAvailabilityConfigRecoveryActionStopInstance    InstanceConfigurationAvailabilityConfigRecoveryActionEnum = "STOP_INSTANCE"
)

var mappingInstanceConfigurationAvailabilityConfigRecoveryActionEnum = map[string]InstanceConfigurationAvailabilityConfigRecoveryActionEnum{
	"RESTORE_INSTANCE": InstanceConfigurationAvailabilityConfigRecoveryActionRestoreInstance,
	"STOP_INSTANCE":    InstanceConfigurationAvailabilityConfigRecoveryActionStopInstance,
}

var mappingInstanceConfigurationAvailabilityConfigRecoveryActionEnumLowerCase = map[string]InstanceConfigurationAvailabilityConfigRecoveryActionEnum{
	"restore_instance": InstanceConfigurationAvailabilityConfigRecoveryActionRestoreInstance,
	"stop_instance":    InstanceConfigurationAvailabilityConfigRecoveryActionStopInstance,
}

// GetInstanceConfigurationAvailabilityConfigRecoveryActionEnumValues Enumerates the set of values for InstanceConfigurationAvailabilityConfigRecoveryActionEnum
func GetInstanceConfigurationAvailabilityConfigRecoveryActionEnumValues() []InstanceConfigurationAvailabilityConfigRecoveryActionEnum {
	values := make([]InstanceConfigurationAvailabilityConfigRecoveryActionEnum, 0)
	for _, v := range mappingInstanceConfigurationAvailabilityConfigRecoveryActionEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceConfigurationAvailabilityConfigRecoveryActionEnumStringValues Enumerates the set of values in String for InstanceConfigurationAvailabilityConfigRecoveryActionEnum
func GetInstanceConfigurationAvailabilityConfigRecoveryActionEnumStringValues() []string {
	return []string{
		"RESTORE_INSTANCE",
		"STOP_INSTANCE",
	}
}

// GetMappingInstanceConfigurationAvailabilityConfigRecoveryActionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceConfigurationAvailabilityConfigRecoveryActionEnum(val string) (InstanceConfigurationAvailabilityConfigRecoveryActionEnum, bool) {
	enum, ok := mappingInstanceConfigurationAvailabilityConfigRecoveryActionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
