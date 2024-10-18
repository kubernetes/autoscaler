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

// ShapeNumaNodesPerSocketPlatformOptions Configuration options for NUMA nodes per socket.
type ShapeNumaNodesPerSocketPlatformOptions struct {

	// The supported values for this platform configuration property.
	AllowedValues []ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum `mandatory:"false" json:"allowedValues,omitempty"`

	// The default NUMA nodes per socket configuration.
	DefaultValue *string `mandatory:"false" json:"defaultValue"`
}

func (m ShapeNumaNodesPerSocketPlatformOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ShapeNumaNodesPerSocketPlatformOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	for _, val := range m.AllowedValues {
		if _, ok := GetMappingShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AllowedValues: %s. Supported values are: %s.", val, strings.Join(GetShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnumStringValues(), ",")))
		}
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum Enum with underlying type: string
type ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum string

// Set of constants representing the allowable values for ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum
const (
	ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps0 ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum = "NPS0"
	ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps1 ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum = "NPS1"
	ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps2 ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum = "NPS2"
	ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps4 ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum = "NPS4"
)

var mappingShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum = map[string]ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum{
	"NPS0": ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps0,
	"NPS1": ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps1,
	"NPS2": ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps2,
	"NPS4": ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps4,
}

var mappingShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnumLowerCase = map[string]ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum{
	"nps0": ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps0,
	"nps1": ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps1,
	"nps2": ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps2,
	"nps4": ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesNps4,
}

// GetShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnumValues Enumerates the set of values for ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum
func GetShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnumValues() []ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum {
	values := make([]ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum, 0)
	for _, v := range mappingShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum {
		values = append(values, v)
	}
	return values
}

// GetShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnumStringValues Enumerates the set of values in String for ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum
func GetShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnumStringValues() []string {
	return []string{
		"NPS0",
		"NPS1",
		"NPS2",
		"NPS4",
	}
}

// GetMappingShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum(val string) (ShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnum, bool) {
	enum, ok := mappingShapeNumaNodesPerSocketPlatformOptionsAllowedValuesEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
