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

// RouteTable A collection of `RouteRule` objects, which are used to route packets
// based on destination IP to a particular network entity. For more information, see
// Overview of the Networking Service (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type RouteTable struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the route table.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The route table's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The route table's current state.
	LifecycleState RouteTableLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The collection of rules for routing destination IPs to network devices.
	RouteRules []RouteRule `mandatory:"true" json:"routeRules"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN the route table list belongs to.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The date and time the route table was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m RouteTable) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m RouteTable) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingRouteTableLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetRouteTableLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// RouteTableLifecycleStateEnum Enum with underlying type: string
type RouteTableLifecycleStateEnum string

// Set of constants representing the allowable values for RouteTableLifecycleStateEnum
const (
	RouteTableLifecycleStateProvisioning RouteTableLifecycleStateEnum = "PROVISIONING"
	RouteTableLifecycleStateAvailable    RouteTableLifecycleStateEnum = "AVAILABLE"
	RouteTableLifecycleStateTerminating  RouteTableLifecycleStateEnum = "TERMINATING"
	RouteTableLifecycleStateTerminated   RouteTableLifecycleStateEnum = "TERMINATED"
)

var mappingRouteTableLifecycleStateEnum = map[string]RouteTableLifecycleStateEnum{
	"PROVISIONING": RouteTableLifecycleStateProvisioning,
	"AVAILABLE":    RouteTableLifecycleStateAvailable,
	"TERMINATING":  RouteTableLifecycleStateTerminating,
	"TERMINATED":   RouteTableLifecycleStateTerminated,
}

var mappingRouteTableLifecycleStateEnumLowerCase = map[string]RouteTableLifecycleStateEnum{
	"provisioning": RouteTableLifecycleStateProvisioning,
	"available":    RouteTableLifecycleStateAvailable,
	"terminating":  RouteTableLifecycleStateTerminating,
	"terminated":   RouteTableLifecycleStateTerminated,
}

// GetRouteTableLifecycleStateEnumValues Enumerates the set of values for RouteTableLifecycleStateEnum
func GetRouteTableLifecycleStateEnumValues() []RouteTableLifecycleStateEnum {
	values := make([]RouteTableLifecycleStateEnum, 0)
	for _, v := range mappingRouteTableLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetRouteTableLifecycleStateEnumStringValues Enumerates the set of values in String for RouteTableLifecycleStateEnum
func GetRouteTableLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingRouteTableLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingRouteTableLifecycleStateEnum(val string) (RouteTableLifecycleStateEnum, bool) {
	enum, ok := mappingRouteTableLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
