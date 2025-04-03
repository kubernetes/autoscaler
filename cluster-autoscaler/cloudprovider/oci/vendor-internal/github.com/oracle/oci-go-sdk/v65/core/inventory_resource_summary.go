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

// InventoryResourceSummary Lists resources and its properties under a given subnet.
type InventoryResourceSummary struct {

	// The name of the resource created.
	ResourceName *string `mandatory:"false" json:"resourceName"`

	// Resource types of the resource.
	ResourceType InventoryResourceSummaryResourceTypeEnum `mandatory:"false" json:"resourceType,omitempty"`

	// Lists the 'IpAddressCollection' object.
	IpAddressCollection []InventoryIpAddressSummary `mandatory:"false" json:"ipAddressCollection"`

	// The region name of the corresponding resource.
	Region *string `mandatory:"false" json:"region"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`
}

func (m InventoryResourceSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InventoryResourceSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingInventoryResourceSummaryResourceTypeEnum(string(m.ResourceType)); !ok && m.ResourceType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ResourceType: %s. Supported values are: %s.", m.ResourceType, strings.Join(GetInventoryResourceSummaryResourceTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InventoryResourceSummaryResourceTypeEnum Enum with underlying type: string
type InventoryResourceSummaryResourceTypeEnum string

// Set of constants representing the allowable values for InventoryResourceSummaryResourceTypeEnum
const (
	InventoryResourceSummaryResourceTypeResource InventoryResourceSummaryResourceTypeEnum = "Resource"
)

var mappingInventoryResourceSummaryResourceTypeEnum = map[string]InventoryResourceSummaryResourceTypeEnum{
	"Resource": InventoryResourceSummaryResourceTypeResource,
}

var mappingInventoryResourceSummaryResourceTypeEnumLowerCase = map[string]InventoryResourceSummaryResourceTypeEnum{
	"resource": InventoryResourceSummaryResourceTypeResource,
}

// GetInventoryResourceSummaryResourceTypeEnumValues Enumerates the set of values for InventoryResourceSummaryResourceTypeEnum
func GetInventoryResourceSummaryResourceTypeEnumValues() []InventoryResourceSummaryResourceTypeEnum {
	values := make([]InventoryResourceSummaryResourceTypeEnum, 0)
	for _, v := range mappingInventoryResourceSummaryResourceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetInventoryResourceSummaryResourceTypeEnumStringValues Enumerates the set of values in String for InventoryResourceSummaryResourceTypeEnum
func GetInventoryResourceSummaryResourceTypeEnumStringValues() []string {
	return []string{
		"Resource",
	}
}

// GetMappingInventoryResourceSummaryResourceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInventoryResourceSummaryResourceTypeEnum(val string) (InventoryResourceSummaryResourceTypeEnum, bool) {
	enum, ok := mappingInventoryResourceSummaryResourceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
