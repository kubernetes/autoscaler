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

// TopologyRoutesToRelationshipDetails Defines route rule details for a `routesTo` relationship.
type TopologyRoutesToRelationshipDetails struct {

	// The destinationType can be set to one of two values:
	// * Use `CIDR_BLOCK` if the rule's `destination` is an IP address range in CIDR notation.
	// * Use `SERVICE_CIDR_BLOCK` if the rule's `destination` is the `cidrBlock` value for a Service.
	DestinationType *string `mandatory:"true" json:"destinationType"`

	// An IP address range in CIDR notation or the `cidrBlock` value for a Service.
	Destination *string `mandatory:"true" json:"destination"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the routing table that contains the route rule.
	RouteTableId *string `mandatory:"true" json:"routeTableId"`

	// A route rule can be `STATIC` if manually added to the route table or `DYNAMIC` if imported from another route table.
	RouteType TopologyRoutesToRelationshipDetailsRouteTypeEnum `mandatory:"false" json:"routeType,omitempty"`
}

func (m TopologyRoutesToRelationshipDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m TopologyRoutesToRelationshipDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingTopologyRoutesToRelationshipDetailsRouteTypeEnum(string(m.RouteType)); !ok && m.RouteType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RouteType: %s. Supported values are: %s.", m.RouteType, strings.Join(GetTopologyRoutesToRelationshipDetailsRouteTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// TopologyRoutesToRelationshipDetailsRouteTypeEnum Enum with underlying type: string
type TopologyRoutesToRelationshipDetailsRouteTypeEnum string

// Set of constants representing the allowable values for TopologyRoutesToRelationshipDetailsRouteTypeEnum
const (
	TopologyRoutesToRelationshipDetailsRouteTypeStatic  TopologyRoutesToRelationshipDetailsRouteTypeEnum = "STATIC"
	TopologyRoutesToRelationshipDetailsRouteTypeDynamic TopologyRoutesToRelationshipDetailsRouteTypeEnum = "DYNAMIC"
)

var mappingTopologyRoutesToRelationshipDetailsRouteTypeEnum = map[string]TopologyRoutesToRelationshipDetailsRouteTypeEnum{
	"STATIC":  TopologyRoutesToRelationshipDetailsRouteTypeStatic,
	"DYNAMIC": TopologyRoutesToRelationshipDetailsRouteTypeDynamic,
}

var mappingTopologyRoutesToRelationshipDetailsRouteTypeEnumLowerCase = map[string]TopologyRoutesToRelationshipDetailsRouteTypeEnum{
	"static":  TopologyRoutesToRelationshipDetailsRouteTypeStatic,
	"dynamic": TopologyRoutesToRelationshipDetailsRouteTypeDynamic,
}

// GetTopologyRoutesToRelationshipDetailsRouteTypeEnumValues Enumerates the set of values for TopologyRoutesToRelationshipDetailsRouteTypeEnum
func GetTopologyRoutesToRelationshipDetailsRouteTypeEnumValues() []TopologyRoutesToRelationshipDetailsRouteTypeEnum {
	values := make([]TopologyRoutesToRelationshipDetailsRouteTypeEnum, 0)
	for _, v := range mappingTopologyRoutesToRelationshipDetailsRouteTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetTopologyRoutesToRelationshipDetailsRouteTypeEnumStringValues Enumerates the set of values in String for TopologyRoutesToRelationshipDetailsRouteTypeEnum
func GetTopologyRoutesToRelationshipDetailsRouteTypeEnumStringValues() []string {
	return []string{
		"STATIC",
		"DYNAMIC",
	}
}

// GetMappingTopologyRoutesToRelationshipDetailsRouteTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingTopologyRoutesToRelationshipDetailsRouteTypeEnum(val string) (TopologyRoutesToRelationshipDetailsRouteTypeEnum, bool) {
	enum, ok := mappingTopologyRoutesToRelationshipDetailsRouteTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
