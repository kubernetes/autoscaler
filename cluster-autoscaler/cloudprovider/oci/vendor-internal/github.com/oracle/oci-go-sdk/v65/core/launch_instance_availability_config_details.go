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

// LaunchInstanceAvailabilityConfigDetails Options for VM migration during infrastructure maintenance events and for defining
// the availability of a VM instance after a maintenance event that impacts the underlying hardware.
type LaunchInstanceAvailabilityConfigDetails struct {

	// Whether to live migrate supported VM instances to a healthy physical VM host without
	// disrupting running instances during infrastructure maintenance events. If null, Oracle
	// chooses the best option for migrating the VM during infrastructure maintenance events.
	IsLiveMigrationPreferred *bool `mandatory:"false" json:"isLiveMigrationPreferred"`

	// The lifecycle state for an instance when it is recovered after infrastructure maintenance.
	// * `RESTORE_INSTANCE` - The instance is restored to the lifecycle state it was in before the maintenance event.
	// If the instance was running, it is automatically rebooted. This is the default action when a value is not set.
	// * `STOP_INSTANCE` - The instance is recovered in the stopped state.
	RecoveryAction LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum `mandatory:"false" json:"recoveryAction,omitempty"`
}

func (m LaunchInstanceAvailabilityConfigDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m LaunchInstanceAvailabilityConfigDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum(string(m.RecoveryAction)); !ok && m.RecoveryAction != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RecoveryAction: %s. Supported values are: %s.", m.RecoveryAction, strings.Join(GetLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum Enum with underlying type: string
type LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum string

// Set of constants representing the allowable values for LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum
const (
	LaunchInstanceAvailabilityConfigDetailsRecoveryActionRestoreInstance LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum = "RESTORE_INSTANCE"
	LaunchInstanceAvailabilityConfigDetailsRecoveryActionStopInstance    LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum = "STOP_INSTANCE"
)

var mappingLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum = map[string]LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum{
	"RESTORE_INSTANCE": LaunchInstanceAvailabilityConfigDetailsRecoveryActionRestoreInstance,
	"STOP_INSTANCE":    LaunchInstanceAvailabilityConfigDetailsRecoveryActionStopInstance,
}

var mappingLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnumLowerCase = map[string]LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum{
	"restore_instance": LaunchInstanceAvailabilityConfigDetailsRecoveryActionRestoreInstance,
	"stop_instance":    LaunchInstanceAvailabilityConfigDetailsRecoveryActionStopInstance,
}

// GetLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnumValues Enumerates the set of values for LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum
func GetLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnumValues() []LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum {
	values := make([]LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum, 0)
	for _, v := range mappingLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum {
		values = append(values, v)
	}
	return values
}

// GetLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnumStringValues Enumerates the set of values in String for LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum
func GetLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnumStringValues() []string {
	return []string{
		"RESTORE_INSTANCE",
		"STOP_INSTANCE",
	}
}

// GetMappingLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum(val string) (LaunchInstanceAvailabilityConfigDetailsRecoveryActionEnum, bool) {
	enum, ok := mappingLaunchInstanceAvailabilityConfigDetailsRecoveryActionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
