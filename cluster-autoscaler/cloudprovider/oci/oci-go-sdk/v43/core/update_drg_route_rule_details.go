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

// UpdateDrgRouteRuleDetails Details used to update a route rule in the DRG route table.
type UpdateDrgRouteRuleDetails struct {

	// The Oracle-assigned ID of each DRG route rule to update.
	Id *string `mandatory:"true" json:"id"`

	// The range of IP addresses used for matching when routing traffic.
	// Potential values:
	//   * IP address range in CIDR notation. Can be an IPv4 or IPv6 CIDR. For example: `192.168.1.0/24`
	//   or `2001:0db8:0123:45::/56`.
	Destination *string `mandatory:"false" json:"destination"`

	// Type of destination for the rule. Required if `direction` = `EGRESS`.
	// Allowed values:
	//   * `CIDR_BLOCK`: If the rule's `destination` is an IP address range in CIDR notation.
	DestinationType UpdateDrgRouteRuleDetailsDestinationTypeEnum `mandatory:"false" json:"destinationType,omitempty"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the next hop DRG attachment. The next hop DRG attachment is responsible
	// for reaching the network destination.
	NextHopDrgAttachmentId *string `mandatory:"false" json:"nextHopDrgAttachmentId"`
}

func (m UpdateDrgRouteRuleDetails) String() string {
	return common.PointerString(m)
}

// UpdateDrgRouteRuleDetailsDestinationTypeEnum Enum with underlying type: string
type UpdateDrgRouteRuleDetailsDestinationTypeEnum string

// Set of constants representing the allowable values for UpdateDrgRouteRuleDetailsDestinationTypeEnum
const (
	UpdateDrgRouteRuleDetailsDestinationTypeCidrBlock UpdateDrgRouteRuleDetailsDestinationTypeEnum = "CIDR_BLOCK"
)

var mappingUpdateDrgRouteRuleDetailsDestinationType = map[string]UpdateDrgRouteRuleDetailsDestinationTypeEnum{
	"CIDR_BLOCK": UpdateDrgRouteRuleDetailsDestinationTypeCidrBlock,
}

// GetUpdateDrgRouteRuleDetailsDestinationTypeEnumValues Enumerates the set of values for UpdateDrgRouteRuleDetailsDestinationTypeEnum
func GetUpdateDrgRouteRuleDetailsDestinationTypeEnumValues() []UpdateDrgRouteRuleDetailsDestinationTypeEnum {
	values := make([]UpdateDrgRouteRuleDetailsDestinationTypeEnum, 0)
	for _, v := range mappingUpdateDrgRouteRuleDetailsDestinationType {
		values = append(values, v)
	}
	return values
}
