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

// ClientVpnStatus The status of ClientVpn.
type ClientVpnStatus struct {

	// The number of current connections on given clientVpn.
	CurrentConnections *int `mandatory:"true" json:"currentConnections"`

	// The list of active users.
	ActiveUsers []ClientVpnActiveUser `mandatory:"true" json:"activeUsers"`

	// The current state of the ClientVPNendpoint.
	LifecycleState ClientVpnStatusLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`
}

func (m ClientVpnStatus) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ClientVpnStatus) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingClientVpnStatusLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetClientVpnStatusLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ClientVpnStatusLifecycleStateEnum Enum with underlying type: string
type ClientVpnStatusLifecycleStateEnum string

// Set of constants representing the allowable values for ClientVpnStatusLifecycleStateEnum
const (
	ClientVpnStatusLifecycleStateCreating ClientVpnStatusLifecycleStateEnum = "CREATING"
	ClientVpnStatusLifecycleStateActive   ClientVpnStatusLifecycleStateEnum = "ACTIVE"
	ClientVpnStatusLifecycleStateInactive ClientVpnStatusLifecycleStateEnum = "INACTIVE"
	ClientVpnStatusLifecycleStateFailed   ClientVpnStatusLifecycleStateEnum = "FAILED"
	ClientVpnStatusLifecycleStateDeleted  ClientVpnStatusLifecycleStateEnum = "DELETED"
	ClientVpnStatusLifecycleStateDeleting ClientVpnStatusLifecycleStateEnum = "DELETING"
	ClientVpnStatusLifecycleStateUpdating ClientVpnStatusLifecycleStateEnum = "UPDATING"
)

var mappingClientVpnStatusLifecycleStateEnum = map[string]ClientVpnStatusLifecycleStateEnum{
	"CREATING": ClientVpnStatusLifecycleStateCreating,
	"ACTIVE":   ClientVpnStatusLifecycleStateActive,
	"INACTIVE": ClientVpnStatusLifecycleStateInactive,
	"FAILED":   ClientVpnStatusLifecycleStateFailed,
	"DELETED":  ClientVpnStatusLifecycleStateDeleted,
	"DELETING": ClientVpnStatusLifecycleStateDeleting,
	"UPDATING": ClientVpnStatusLifecycleStateUpdating,
}

// GetClientVpnStatusLifecycleStateEnumValues Enumerates the set of values for ClientVpnStatusLifecycleStateEnum
func GetClientVpnStatusLifecycleStateEnumValues() []ClientVpnStatusLifecycleStateEnum {
	values := make([]ClientVpnStatusLifecycleStateEnum, 0)
	for _, v := range mappingClientVpnStatusLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetClientVpnStatusLifecycleStateEnumStringValues Enumerates the set of values in String for ClientVpnStatusLifecycleStateEnum
func GetClientVpnStatusLifecycleStateEnumStringValues() []string {
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
