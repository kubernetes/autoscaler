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

// ListComputeGpuMemoryClusterInstancesRequest wrapper for the ListComputeGpuMemoryClusterInstances operation
//
// # See also
//
// Click https://docs.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeGpuMemoryClusterInstances.go.html to see an example of how to use ListComputeGpuMemoryClusterInstancesRequest.
type ListComputeGpuMemoryClusterInstancesRequest struct {

	// The OCID of the compute GPU memory cluster.
	ComputeGpuMemoryClusterId *string `mandatory:"true" contributesTo:"path" name:"computeGpuMemoryClusterId"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

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
	SortBy ListComputeGpuMemoryClusterInstancesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeGpuMemoryClusterInstancesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeGpuMemoryClusterInstancesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeGpuMemoryClusterInstancesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeGpuMemoryClusterInstancesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeGpuMemoryClusterInstancesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeGpuMemoryClusterInstancesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListComputeGpuMemoryClusterInstancesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeGpuMemoryClusterInstancesSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeGpuMemoryClusterInstancesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeGpuMemoryClusterInstancesSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeGpuMemoryClusterInstancesResponse wrapper for the ListComputeGpuMemoryClusterInstances operation
type ListComputeGpuMemoryClusterInstancesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ComputeGpuMemoryClusterInstanceCollection instances
	ComputeGpuMemoryClusterInstanceCollection `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeGpuMemoryClusterInstancesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeGpuMemoryClusterInstancesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeGpuMemoryClusterInstancesSortByEnum Enum with underlying type: string
type ListComputeGpuMemoryClusterInstancesSortByEnum string

// Set of constants representing the allowable values for ListComputeGpuMemoryClusterInstancesSortByEnum
const (
	ListComputeGpuMemoryClusterInstancesSortByTimecreated ListComputeGpuMemoryClusterInstancesSortByEnum = "TIMECREATED"
	ListComputeGpuMemoryClusterInstancesSortByDisplayname ListComputeGpuMemoryClusterInstancesSortByEnum = "DISPLAYNAME"
)

var mappingListComputeGpuMemoryClusterInstancesSortByEnum = map[string]ListComputeGpuMemoryClusterInstancesSortByEnum{
	"TIMECREATED": ListComputeGpuMemoryClusterInstancesSortByTimecreated,
	"DISPLAYNAME": ListComputeGpuMemoryClusterInstancesSortByDisplayname,
}

var mappingListComputeGpuMemoryClusterInstancesSortByEnumLowerCase = map[string]ListComputeGpuMemoryClusterInstancesSortByEnum{
	"timecreated": ListComputeGpuMemoryClusterInstancesSortByTimecreated,
	"displayname": ListComputeGpuMemoryClusterInstancesSortByDisplayname,
}

// GetListComputeGpuMemoryClusterInstancesSortByEnumValues Enumerates the set of values for ListComputeGpuMemoryClusterInstancesSortByEnum
func GetListComputeGpuMemoryClusterInstancesSortByEnumValues() []ListComputeGpuMemoryClusterInstancesSortByEnum {
	values := make([]ListComputeGpuMemoryClusterInstancesSortByEnum, 0)
	for _, v := range mappingListComputeGpuMemoryClusterInstancesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeGpuMemoryClusterInstancesSortByEnumStringValues Enumerates the set of values in String for ListComputeGpuMemoryClusterInstancesSortByEnum
func GetListComputeGpuMemoryClusterInstancesSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeGpuMemoryClusterInstancesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeGpuMemoryClusterInstancesSortByEnum(val string) (ListComputeGpuMemoryClusterInstancesSortByEnum, bool) {
	enum, ok := mappingListComputeGpuMemoryClusterInstancesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeGpuMemoryClusterInstancesSortOrderEnum Enum with underlying type: string
type ListComputeGpuMemoryClusterInstancesSortOrderEnum string

// Set of constants representing the allowable values for ListComputeGpuMemoryClusterInstancesSortOrderEnum
const (
	ListComputeGpuMemoryClusterInstancesSortOrderAsc  ListComputeGpuMemoryClusterInstancesSortOrderEnum = "ASC"
	ListComputeGpuMemoryClusterInstancesSortOrderDesc ListComputeGpuMemoryClusterInstancesSortOrderEnum = "DESC"
)

var mappingListComputeGpuMemoryClusterInstancesSortOrderEnum = map[string]ListComputeGpuMemoryClusterInstancesSortOrderEnum{
	"ASC":  ListComputeGpuMemoryClusterInstancesSortOrderAsc,
	"DESC": ListComputeGpuMemoryClusterInstancesSortOrderDesc,
}

var mappingListComputeGpuMemoryClusterInstancesSortOrderEnumLowerCase = map[string]ListComputeGpuMemoryClusterInstancesSortOrderEnum{
	"asc":  ListComputeGpuMemoryClusterInstancesSortOrderAsc,
	"desc": ListComputeGpuMemoryClusterInstancesSortOrderDesc,
}

// GetListComputeGpuMemoryClusterInstancesSortOrderEnumValues Enumerates the set of values for ListComputeGpuMemoryClusterInstancesSortOrderEnum
func GetListComputeGpuMemoryClusterInstancesSortOrderEnumValues() []ListComputeGpuMemoryClusterInstancesSortOrderEnum {
	values := make([]ListComputeGpuMemoryClusterInstancesSortOrderEnum, 0)
	for _, v := range mappingListComputeGpuMemoryClusterInstancesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeGpuMemoryClusterInstancesSortOrderEnumStringValues Enumerates the set of values in String for ListComputeGpuMemoryClusterInstancesSortOrderEnum
func GetListComputeGpuMemoryClusterInstancesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeGpuMemoryClusterInstancesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeGpuMemoryClusterInstancesSortOrderEnum(val string) (ListComputeGpuMemoryClusterInstancesSortOrderEnum, bool) {
	enum, ok := mappingListComputeGpuMemoryClusterInstancesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
