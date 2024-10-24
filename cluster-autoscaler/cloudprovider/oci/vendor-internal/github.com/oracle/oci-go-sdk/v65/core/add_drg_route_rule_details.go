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

// AddDrgRouteRuleDetails Details needed when adding a DRG route rule.
type AddDrgRouteRuleDetails struct {

	// Type of destination for the rule.
	// Allowed values:
	//   * `CIDR_BLOCK`: If the rule's `destination` is an IP address range in CIDR notation.
	DestinationType AddDrgRouteRuleDetailsDestinationTypeEnum `mandatory:"true" json:"destinationType"`

	// This is the range of IP addresses used for matching when routing
	// traffic. Only CIDR_BLOCK values are allowed.
	// Potential values:
	//   * IP address range in CIDR notation. This can be an IPv4 CIDR block or IPv6 prefix. For example: `192.168.1.0/24`
	//   or `2001:0db8:0123:45::/56`.
	Destination *string `mandatory:"true" json:"destination"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the next hop DRG attachment. The next hop DRG attachment is responsible
	// for reaching the network destination.
	NextHopDrgAttachmentId *string `mandatory:"true" json:"nextHopDrgAttachmentId"`
}

func (m AddDrgRouteRuleDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m AddDrgRouteRuleDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingAddDrgRouteRuleDetailsDestinationTypeEnum(string(m.DestinationType)); !ok && m.DestinationType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DestinationType: %s. Supported values are: %s.", m.DestinationType, strings.Join(GetAddDrgRouteRuleDetailsDestinationTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// AddDrgRouteRuleDetailsDestinationTypeEnum Enum with underlying type: string
type AddDrgRouteRuleDetailsDestinationTypeEnum string

// Set of constants representing the allowable values for AddDrgRouteRuleDetailsDestinationTypeEnum
const (
	AddDrgRouteRuleDetailsDestinationTypeCidrBlock AddDrgRouteRuleDetailsDestinationTypeEnum = "CIDR_BLOCK"
)

var mappingAddDrgRouteRuleDetailsDestinationTypeEnum = map[string]AddDrgRouteRuleDetailsDestinationTypeEnum{
	"CIDR_BLOCK": AddDrgRouteRuleDetailsDestinationTypeCidrBlock,
}

var mappingAddDrgRouteRuleDetailsDestinationTypeEnumLowerCase = map[string]AddDrgRouteRuleDetailsDestinationTypeEnum{
	"cidr_block": AddDrgRouteRuleDetailsDestinationTypeCidrBlock,
}

// GetAddDrgRouteRuleDetailsDestinationTypeEnumValues Enumerates the set of values for AddDrgRouteRuleDetailsDestinationTypeEnum
func GetAddDrgRouteRuleDetailsDestinationTypeEnumValues() []AddDrgRouteRuleDetailsDestinationTypeEnum {
	values := make([]AddDrgRouteRuleDetailsDestinationTypeEnum, 0)
	for _, v := range mappingAddDrgRouteRuleDetailsDestinationTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetAddDrgRouteRuleDetailsDestinationTypeEnumStringValues Enumerates the set of values in String for AddDrgRouteRuleDetailsDestinationTypeEnum
func GetAddDrgRouteRuleDetailsDestinationTypeEnumStringValues() []string {
	return []string{
		"CIDR_BLOCK",
	}
}

// GetMappingAddDrgRouteRuleDetailsDestinationTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingAddDrgRouteRuleDetailsDestinationTypeEnum(val string) (AddDrgRouteRuleDetailsDestinationTypeEnum, bool) {
	enum, ok := mappingAddDrgRouteRuleDetailsDestinationTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
