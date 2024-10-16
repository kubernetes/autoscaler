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

// TunnelStatus Deprecated. For tunnel information, instead see IPSecConnectionTunnel.
type TunnelStatus struct {

	// The IP address of Oracle's VPN headend.
	// Example: `203.0.113.50`
	IpAddress *string `mandatory:"true" json:"ipAddress"`

	// The tunnel's current state.
	LifecycleState TunnelStatusLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the IPSec connection was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// When the state of the tunnel last changed, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeStateModified *common.SDKTime `mandatory:"false" json:"timeStateModified"`
}

func (m TunnelStatus) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m TunnelStatus) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingTunnelStatusLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetTunnelStatusLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// TunnelStatusLifecycleStateEnum Enum with underlying type: string
type TunnelStatusLifecycleStateEnum string

// Set of constants representing the allowable values for TunnelStatusLifecycleStateEnum
const (
	TunnelStatusLifecycleStateUp                 TunnelStatusLifecycleStateEnum = "UP"
	TunnelStatusLifecycleStateDown               TunnelStatusLifecycleStateEnum = "DOWN"
	TunnelStatusLifecycleStateDownForMaintenance TunnelStatusLifecycleStateEnum = "DOWN_FOR_MAINTENANCE"
	TunnelStatusLifecycleStatePartialUp          TunnelStatusLifecycleStateEnum = "PARTIAL_UP"
)

var mappingTunnelStatusLifecycleStateEnum = map[string]TunnelStatusLifecycleStateEnum{
	"UP":                   TunnelStatusLifecycleStateUp,
	"DOWN":                 TunnelStatusLifecycleStateDown,
	"DOWN_FOR_MAINTENANCE": TunnelStatusLifecycleStateDownForMaintenance,
	"PARTIAL_UP":           TunnelStatusLifecycleStatePartialUp,
}

var mappingTunnelStatusLifecycleStateEnumLowerCase = map[string]TunnelStatusLifecycleStateEnum{
	"up":                   TunnelStatusLifecycleStateUp,
	"down":                 TunnelStatusLifecycleStateDown,
	"down_for_maintenance": TunnelStatusLifecycleStateDownForMaintenance,
	"partial_up":           TunnelStatusLifecycleStatePartialUp,
}

// GetTunnelStatusLifecycleStateEnumValues Enumerates the set of values for TunnelStatusLifecycleStateEnum
func GetTunnelStatusLifecycleStateEnumValues() []TunnelStatusLifecycleStateEnum {
	values := make([]TunnelStatusLifecycleStateEnum, 0)
	for _, v := range mappingTunnelStatusLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetTunnelStatusLifecycleStateEnumStringValues Enumerates the set of values in String for TunnelStatusLifecycleStateEnum
func GetTunnelStatusLifecycleStateEnumStringValues() []string {
	return []string{
		"UP",
		"DOWN",
		"DOWN_FOR_MAINTENANCE",
		"PARTIAL_UP",
	}
}

// GetMappingTunnelStatusLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingTunnelStatusLifecycleStateEnum(val string) (TunnelStatusLifecycleStateEnum, bool) {
	enum, ok := mappingTunnelStatusLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
