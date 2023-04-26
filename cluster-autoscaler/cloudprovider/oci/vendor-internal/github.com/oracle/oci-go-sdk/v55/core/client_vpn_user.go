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

// ClientVpnUser The user of a certain clientVpn.
type ClientVpnUser struct {

	// The unique username of the user want to create.
	UserName *string `mandatory:"true" json:"userName"`

	// The current state of the ClientVPNendpointUser.
	LifecycleState ClientVpnUserLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Whether to log in the user by cert-authentication only or not.
	IsCertAuthOnly *bool `mandatory:"false" json:"isCertAuthOnly"`

	// The time ClientVpnUser was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m ClientVpnUser) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ClientVpnUser) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingClientVpnUserLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetClientVpnUserLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ClientVpnUserLifecycleStateEnum Enum with underlying type: string
type ClientVpnUserLifecycleStateEnum string

// Set of constants representing the allowable values for ClientVpnUserLifecycleStateEnum
const (
	ClientVpnUserLifecycleStateCreating ClientVpnUserLifecycleStateEnum = "CREATING"
	ClientVpnUserLifecycleStateActive   ClientVpnUserLifecycleStateEnum = "ACTIVE"
	ClientVpnUserLifecycleStateInactive ClientVpnUserLifecycleStateEnum = "INACTIVE"
	ClientVpnUserLifecycleStateFailed   ClientVpnUserLifecycleStateEnum = "FAILED"
	ClientVpnUserLifecycleStateDeleted  ClientVpnUserLifecycleStateEnum = "DELETED"
	ClientVpnUserLifecycleStateDeleting ClientVpnUserLifecycleStateEnum = "DELETING"
	ClientVpnUserLifecycleStateUpdating ClientVpnUserLifecycleStateEnum = "UPDATING"
)

var mappingClientVpnUserLifecycleStateEnum = map[string]ClientVpnUserLifecycleStateEnum{
	"CREATING": ClientVpnUserLifecycleStateCreating,
	"ACTIVE":   ClientVpnUserLifecycleStateActive,
	"INACTIVE": ClientVpnUserLifecycleStateInactive,
	"FAILED":   ClientVpnUserLifecycleStateFailed,
	"DELETED":  ClientVpnUserLifecycleStateDeleted,
	"DELETING": ClientVpnUserLifecycleStateDeleting,
	"UPDATING": ClientVpnUserLifecycleStateUpdating,
}

// GetClientVpnUserLifecycleStateEnumValues Enumerates the set of values for ClientVpnUserLifecycleStateEnum
func GetClientVpnUserLifecycleStateEnumValues() []ClientVpnUserLifecycleStateEnum {
	values := make([]ClientVpnUserLifecycleStateEnum, 0)
	for _, v := range mappingClientVpnUserLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetClientVpnUserLifecycleStateEnumStringValues Enumerates the set of values in String for ClientVpnUserLifecycleStateEnum
func GetClientVpnUserLifecycleStateEnumStringValues() []string {
	return []string{
		"CREATING",
		"ACTIVE",
		"INACTIVE",
		"FAILED",
		"DELETED",
		"DELETING",
		"UPDATING",
	}
}
