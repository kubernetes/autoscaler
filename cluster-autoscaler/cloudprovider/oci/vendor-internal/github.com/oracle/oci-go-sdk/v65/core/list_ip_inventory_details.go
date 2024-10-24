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

// ListIpInventoryDetails Required input parameters for retrieving IP Inventory data within the specified compartments of a region.
type ListIpInventoryDetails struct {

	// Lists the selected regions.
	RegionList []string `mandatory:"true" json:"regionList"`

	// List the OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartments.
	CompartmentList []string `mandatory:"true" json:"compartmentList"`

	// List of selected filters.
	OverrideFilters *bool `mandatory:"false" json:"overrideFilters"`

	// The CIDR utilization of a VCN.
	Utilization *float32 `mandatory:"false" json:"utilization"`

	// List of overlapping VCNs.
	OverlappingVcnsOnly *bool `mandatory:"false" json:"overlappingVcnsOnly"`

	// List of IP address types used.
	AddressTypeList []AddressTypeEnum `mandatory:"false" json:"addressTypeList"`

	// List of VCN resource types.
	ResourceTypeList []ListIpInventoryDetailsResourceTypeListEnum `mandatory:"false" json:"resourceTypeList,omitempty"`

	// Filters the results for the specified string.
	SearchKeyword *string `mandatory:"false" json:"searchKeyword"`

	// Provide the sort order (`sortOrder`) to sort the fields such as TIMECREATED in descending or descending order, and DISPLAYNAME in case sensitive.
	// **Note:** For some "List" operations (for example, `ListInstances`), sort resources by an availability domain when the resources belong to a single availability domain.
	// If you sort the "List" operations without specifying
	// an availability domain, the resources are grouped by availability domains and then sorted.
	SortBy ListIpInventoryDetailsSortByEnum `mandatory:"false" json:"sortBy,omitempty"`

	// Specifies the sort order to use. Select either ascending (`ASC`) or descending (`DESC`) order. The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListIpInventoryDetailsSortOrderEnum `mandatory:"false" json:"sortOrder,omitempty"`

	// Most List operations paginate results. Results are paginated for the ListInstances operations. When you call a paginated List operation, the response indicates more pages of results by including the opc-next-page header.
	// For more information, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	PaginationOffset *int `mandatory:"false" json:"paginationOffset"`

	// Specifies the maximum number of results displayed per page for a paginated "List" call. For more information, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	PaginationLimit *int `mandatory:"false" json:"paginationLimit"`
}

func (m ListIpInventoryDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ListIpInventoryDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	for _, val := range m.ResourceTypeList {
		if _, ok := GetMappingListIpInventoryDetailsResourceTypeListEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ResourceTypeList: %s. Supported values are: %s.", val, strings.Join(GetListIpInventoryDetailsResourceTypeListEnumStringValues(), ",")))
		}
	}

	if _, ok := GetMappingListIpInventoryDetailsSortByEnum(string(m.SortBy)); !ok && m.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", m.SortBy, strings.Join(GetListIpInventoryDetailsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListIpInventoryDetailsSortOrderEnum(string(m.SortOrder)); !ok && m.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", m.SortOrder, strings.Join(GetListIpInventoryDetailsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListIpInventoryDetailsResourceTypeListEnum Enum with underlying type: string
type ListIpInventoryDetailsResourceTypeListEnum string

// Set of constants representing the allowable values for ListIpInventoryDetailsResourceTypeListEnum
const (
	ListIpInventoryDetailsResourceTypeListResource ListIpInventoryDetailsResourceTypeListEnum = "Resource"
)

var mappingListIpInventoryDetailsResourceTypeListEnum = map[string]ListIpInventoryDetailsResourceTypeListEnum{
	"Resource": ListIpInventoryDetailsResourceTypeListResource,
}

var mappingListIpInventoryDetailsResourceTypeListEnumLowerCase = map[string]ListIpInventoryDetailsResourceTypeListEnum{
	"resource": ListIpInventoryDetailsResourceTypeListResource,
}

// GetListIpInventoryDetailsResourceTypeListEnumValues Enumerates the set of values for ListIpInventoryDetailsResourceTypeListEnum
func GetListIpInventoryDetailsResourceTypeListEnumValues() []ListIpInventoryDetailsResourceTypeListEnum {
	values := make([]ListIpInventoryDetailsResourceTypeListEnum, 0)
	for _, v := range mappingListIpInventoryDetailsResourceTypeListEnum {
		values = append(values, v)
	}
	return values
}

// GetListIpInventoryDetailsResourceTypeListEnumStringValues Enumerates the set of values in String for ListIpInventoryDetailsResourceTypeListEnum
func GetListIpInventoryDetailsResourceTypeListEnumStringValues() []string {
	return []string{
		"Resource",
	}
}

// GetMappingListIpInventoryDetailsResourceTypeListEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListIpInventoryDetailsResourceTypeListEnum(val string) (ListIpInventoryDetailsResourceTypeListEnum, bool) {
	enum, ok := mappingListIpInventoryDetailsResourceTypeListEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListIpInventoryDetailsSortByEnum Enum with underlying type: string
type ListIpInventoryDetailsSortByEnum string

// Set of constants representing the allowable values for ListIpInventoryDetailsSortByEnum
const (
	ListIpInventoryDetailsSortByDisplayname ListIpInventoryDetailsSortByEnum = "DISPLAYNAME"
	ListIpInventoryDetailsSortByUtilization ListIpInventoryDetailsSortByEnum = "UTILIZATION"
	ListIpInventoryDetailsSortByDnsHostname ListIpInventoryDetailsSortByEnum = "DNS_HOSTNAME"
	ListIpInventoryDetailsSortByRegion      ListIpInventoryDetailsSortByEnum = "REGION"
)

var mappingListIpInventoryDetailsSortByEnum = map[string]ListIpInventoryDetailsSortByEnum{
	"DISPLAYNAME":  ListIpInventoryDetailsSortByDisplayname,
	"UTILIZATION":  ListIpInventoryDetailsSortByUtilization,
	"DNS_HOSTNAME": ListIpInventoryDetailsSortByDnsHostname,
	"REGION":       ListIpInventoryDetailsSortByRegion,
}

var mappingListIpInventoryDetailsSortByEnumLowerCase = map[string]ListIpInventoryDetailsSortByEnum{
	"displayname":  ListIpInventoryDetailsSortByDisplayname,
	"utilization":  ListIpInventoryDetailsSortByUtilization,
	"dns_hostname": ListIpInventoryDetailsSortByDnsHostname,
	"region":       ListIpInventoryDetailsSortByRegion,
}

// GetListIpInventoryDetailsSortByEnumValues Enumerates the set of values for ListIpInventoryDetailsSortByEnum
func GetListIpInventoryDetailsSortByEnumValues() []ListIpInventoryDetailsSortByEnum {
	values := make([]ListIpInventoryDetailsSortByEnum, 0)
	for _, v := range mappingListIpInventoryDetailsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListIpInventoryDetailsSortByEnumStringValues Enumerates the set of values in String for ListIpInventoryDetailsSortByEnum
func GetListIpInventoryDetailsSortByEnumStringValues() []string {
	return []string{
		"DISPLAYNAME",
		"UTILIZATION",
		"DNS_HOSTNAME",
		"REGION",
	}
}

// GetMappingListIpInventoryDetailsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListIpInventoryDetailsSortByEnum(val string) (ListIpInventoryDetailsSortByEnum, bool) {
	enum, ok := mappingListIpInventoryDetailsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListIpInventoryDetailsSortOrderEnum Enum with underlying type: string
type ListIpInventoryDetailsSortOrderEnum string

// Set of constants representing the allowable values for ListIpInventoryDetailsSortOrderEnum
const (
	ListIpInventoryDetailsSortOrderAsc  ListIpInventoryDetailsSortOrderEnum = "ASC"
	ListIpInventoryDetailsSortOrderDesc ListIpInventoryDetailsSortOrderEnum = "DESC"
)

var mappingListIpInventoryDetailsSortOrderEnum = map[string]ListIpInventoryDetailsSortOrderEnum{
	"ASC":  ListIpInventoryDetailsSortOrderAsc,
	"DESC": ListIpInventoryDetailsSortOrderDesc,
}

var mappingListIpInventoryDetailsSortOrderEnumLowerCase = map[string]ListIpInventoryDetailsSortOrderEnum{
	"asc":  ListIpInventoryDetailsSortOrderAsc,
	"desc": ListIpInventoryDetailsSortOrderDesc,
}

// GetListIpInventoryDetailsSortOrderEnumValues Enumerates the set of values for ListIpInventoryDetailsSortOrderEnum
func GetListIpInventoryDetailsSortOrderEnumValues() []ListIpInventoryDetailsSortOrderEnum {
	values := make([]ListIpInventoryDetailsSortOrderEnum, 0)
	for _, v := range mappingListIpInventoryDetailsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListIpInventoryDetailsSortOrderEnumStringValues Enumerates the set of values in String for ListIpInventoryDetailsSortOrderEnum
func GetListIpInventoryDetailsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListIpInventoryDetailsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListIpInventoryDetailsSortOrderEnum(val string) (ListIpInventoryDetailsSortOrderEnum, bool) {
	enum, ok := mappingListIpInventoryDetailsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
