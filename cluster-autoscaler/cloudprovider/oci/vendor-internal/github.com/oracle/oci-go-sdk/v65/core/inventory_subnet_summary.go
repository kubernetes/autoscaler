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

// InventorySubnetSummary Lists subnet and its associated resources.
type InventorySubnetSummary struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// Name of the subnet within a VCN.
	SubnetName *string `mandatory:"false" json:"subnetName"`

	// Resource types of the subnet.
	ResourceType InventorySubnetSummaryResourceTypeEnum `mandatory:"false" json:"resourceType,omitempty"`

	// Lists CIDRs and utilization within the subnet.
	InventorySubnetCidrCollection []InventorySubnetCidrBlockSummary `mandatory:"false" json:"inventorySubnetCidrCollection"`

	// DNS domain name of the subnet.
	DnsDomainName *string `mandatory:"false" json:"dnsDomainName"`

	// Region name of the subnet.
	Region *string `mandatory:"false" json:"region"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// Lists the `ResourceCollection` object.
	InventoryResourceSummary []InventoryResourceSummary `mandatory:"false" json:"inventoryResourceSummary"`
}

func (m InventorySubnetSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InventorySubnetSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingInventorySubnetSummaryResourceTypeEnum(string(m.ResourceType)); !ok && m.ResourceType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ResourceType: %s. Supported values are: %s.", m.ResourceType, strings.Join(GetInventorySubnetSummaryResourceTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InventorySubnetSummaryResourceTypeEnum Enum with underlying type: string
type InventorySubnetSummaryResourceTypeEnum string

// Set of constants representing the allowable values for InventorySubnetSummaryResourceTypeEnum
const (
	InventorySubnetSummaryResourceTypeSubnet InventorySubnetSummaryResourceTypeEnum = "Subnet"
)

var mappingInventorySubnetSummaryResourceTypeEnum = map[string]InventorySubnetSummaryResourceTypeEnum{
	"Subnet": InventorySubnetSummaryResourceTypeSubnet,
}

var mappingInventorySubnetSummaryResourceTypeEnumLowerCase = map[string]InventorySubnetSummaryResourceTypeEnum{
	"subnet": InventorySubnetSummaryResourceTypeSubnet,
}

// GetInventorySubnetSummaryResourceTypeEnumValues Enumerates the set of values for InventorySubnetSummaryResourceTypeEnum
func GetInventorySubnetSummaryResourceTypeEnumValues() []InventorySubnetSummaryResourceTypeEnum {
	values := make([]InventorySubnetSummaryResourceTypeEnum, 0)
	for _, v := range mappingInventorySubnetSummaryResourceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetInventorySubnetSummaryResourceTypeEnumStringValues Enumerates the set of values in String for InventorySubnetSummaryResourceTypeEnum
func GetInventorySubnetSummaryResourceTypeEnumStringValues() []string {
	return []string{
		"Subnet",
	}
}

// GetMappingInventorySubnetSummaryResourceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInventorySubnetSummaryResourceTypeEnum(val string) (InventorySubnetSummaryResourceTypeEnum, bool) {
	enum, ok := mappingInventorySubnetSummaryResourceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
