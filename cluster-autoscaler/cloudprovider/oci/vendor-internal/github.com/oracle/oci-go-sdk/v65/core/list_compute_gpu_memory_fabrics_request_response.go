// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"net/http"
	"strings"
)

// ListComputeGpuMemoryFabricsRequest wrapper for the ListComputeGpuMemoryFabrics operation
//
// # See also
//
// Click https://docs.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeGpuMemoryFabrics.go.html to see an example of how to use ListComputeGpuMemoryFabricsRequest.
type ListComputeGpuMemoryFabricsRequest struct {

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// A filter to return only the listings that matches the given GPU memory fabric id.
	ComputeGpuMemoryFabricId *string `mandatory:"false" contributesTo:"query" name:"computeGpuMemoryFabricId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute HPC island.
	ComputeHpcIslandId *string `mandatory:"false" contributesTo:"query" name:"computeHpcIslandId"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute network block.
	ComputeNetworkBlockId *string `mandatory:"false" contributesTo:"query" name:"computeNetworkBlockId"`

	// A filter to return ComputeGpuMemoryFabricSummary resources that match the given lifecycle state.
	ComputeGpuMemoryFabricLifecycleState ComputeGpuMemoryFabricLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"computeGpuMemoryFabricLifecycleState" omitEmpty:"true"`

	// A filter to return ComputeGpuMemoryFabricSummary resources that match the given fabric health.
	ComputeGpuMemoryFabricHealth ComputeGpuMemoryFabricFabricHealthEnum `mandatory:"false" contributesTo:"query" name:"computeGpuMemoryFabricHealth" omitEmpty:"true"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListComputeGpuMemoryFabricsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeGpuMemoryFabricsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeGpuMemoryFabricsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeGpuMemoryFabricsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeGpuMemoryFabricsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeGpuMemoryFabricsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeGpuMemoryFabricsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingComputeGpuMemoryFabricLifecycleStateEnum(string(request.ComputeGpuMemoryFabricLifecycleState)); !ok && request.ComputeGpuMemoryFabricLifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ComputeGpuMemoryFabricLifecycleState: %s. Supported values are: %s.", request.ComputeGpuMemoryFabricLifecycleState, strings.Join(GetComputeGpuMemoryFabricLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingComputeGpuMemoryFabricFabricHealthEnum(string(request.ComputeGpuMemoryFabricHealth)); !ok && request.ComputeGpuMemoryFabricHealth != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ComputeGpuMemoryFabricHealth: %s. Supported values are: %s.", request.ComputeGpuMemoryFabricHealth, strings.Join(GetComputeGpuMemoryFabricFabricHealthEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeGpuMemoryFabricsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeGpuMemoryFabricsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeGpuMemoryFabricsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeGpuMemoryFabricsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeGpuMemoryFabricsResponse wrapper for the ListComputeGpuMemoryFabrics operation
type ListComputeGpuMemoryFabricsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ComputeGpuMemoryFabricCollection instances
	ComputeGpuMemoryFabricCollection `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeGpuMemoryFabricsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeGpuMemoryFabricsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeGpuMemoryFabricsSortByEnum Enum with underlying type: string
type ListComputeGpuMemoryFabricsSortByEnum string

// Set of constants representing the allowable values for ListComputeGpuMemoryFabricsSortByEnum
const (
	ListComputeGpuMemoryFabricsSortByTimecreated ListComputeGpuMemoryFabricsSortByEnum = "TIMECREATED"
	ListComputeGpuMemoryFabricsSortByDisplayname ListComputeGpuMemoryFabricsSortByEnum = "DISPLAYNAME"
)

var mappingListComputeGpuMemoryFabricsSortByEnum = map[string]ListComputeGpuMemoryFabricsSortByEnum{
	"TIMECREATED": ListComputeGpuMemoryFabricsSortByTimecreated,
	"DISPLAYNAME": ListComputeGpuMemoryFabricsSortByDisplayname,
}

var mappingListComputeGpuMemoryFabricsSortByEnumLowerCase = map[string]ListComputeGpuMemoryFabricsSortByEnum{
	"timecreated": ListComputeGpuMemoryFabricsSortByTimecreated,
	"displayname": ListComputeGpuMemoryFabricsSortByDisplayname,
}

// GetListComputeGpuMemoryFabricsSortByEnumValues Enumerates the set of values for ListComputeGpuMemoryFabricsSortByEnum
func GetListComputeGpuMemoryFabricsSortByEnumValues() []ListComputeGpuMemoryFabricsSortByEnum {
	values := make([]ListComputeGpuMemoryFabricsSortByEnum, 0)
	for _, v := range mappingListComputeGpuMemoryFabricsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeGpuMemoryFabricsSortByEnumStringValues Enumerates the set of values in String for ListComputeGpuMemoryFabricsSortByEnum
func GetListComputeGpuMemoryFabricsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeGpuMemoryFabricsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeGpuMemoryFabricsSortByEnum(val string) (ListComputeGpuMemoryFabricsSortByEnum, bool) {
	enum, ok := mappingListComputeGpuMemoryFabricsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeGpuMemoryFabricsSortOrderEnum Enum with underlying type: string
type ListComputeGpuMemoryFabricsSortOrderEnum string

// Set of constants representing the allowable values for ListComputeGpuMemoryFabricsSortOrderEnum
const (
	ListComputeGpuMemoryFabricsSortOrderAsc  ListComputeGpuMemoryFabricsSortOrderEnum = "ASC"
	ListComputeGpuMemoryFabricsSortOrderDesc ListComputeGpuMemoryFabricsSortOrderEnum = "DESC"
)

var mappingListComputeGpuMemoryFabricsSortOrderEnum = map[string]ListComputeGpuMemoryFabricsSortOrderEnum{
	"ASC":  ListComputeGpuMemoryFabricsSortOrderAsc,
	"DESC": ListComputeGpuMemoryFabricsSortOrderDesc,
}

var mappingListComputeGpuMemoryFabricsSortOrderEnumLowerCase = map[string]ListComputeGpuMemoryFabricsSortOrderEnum{
	"asc":  ListComputeGpuMemoryFabricsSortOrderAsc,
	"desc": ListComputeGpuMemoryFabricsSortOrderDesc,
}

// GetListComputeGpuMemoryFabricsSortOrderEnumValues Enumerates the set of values for ListComputeGpuMemoryFabricsSortOrderEnum
func GetListComputeGpuMemoryFabricsSortOrderEnumValues() []ListComputeGpuMemoryFabricsSortOrderEnum {
	values := make([]ListComputeGpuMemoryFabricsSortOrderEnum, 0)
	for _, v := range mappingListComputeGpuMemoryFabricsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeGpuMemoryFabricsSortOrderEnumStringValues Enumerates the set of values in String for ListComputeGpuMemoryFabricsSortOrderEnum
func GetListComputeGpuMemoryFabricsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeGpuMemoryFabricsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeGpuMemoryFabricsSortOrderEnum(val string) (ListComputeGpuMemoryFabricsSortOrderEnum, bool) {
	enum, ok := mappingListComputeGpuMemoryFabricsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
