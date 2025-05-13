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

// ListComputeGpuMemoryClustersRequest wrapper for the ListComputeGpuMemoryClusters operation
//
// # See also
//
// Click https://docs.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeGpuMemoryClusters.go.html to see an example of how to use ListComputeGpuMemoryClustersRequest.
type ListComputeGpuMemoryClustersRequest struct {

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// A filter to return only the listings that matches the given GPU memory cluster id.
	ComputeGpuMemoryClusterId *string `mandatory:"false" contributesTo:"query" name:"computeGpuMemoryClusterId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute cluster.
	// A compute cluster (https://docs.oracle.com/iaas/Content/Compute/Tasks/compute-clusters.htm) is a remote direct memory
	// access (RDMA) network group.
	ComputeClusterId *string `mandatory:"false" contributesTo:"query" name:"computeClusterId"`

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
	SortBy ListComputeGpuMemoryClustersSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeGpuMemoryClustersSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeGpuMemoryClustersRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeGpuMemoryClustersRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeGpuMemoryClustersRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeGpuMemoryClustersRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeGpuMemoryClustersRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListComputeGpuMemoryClustersSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeGpuMemoryClustersSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeGpuMemoryClustersSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeGpuMemoryClustersSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeGpuMemoryClustersResponse wrapper for the ListComputeGpuMemoryClusters operation
type ListComputeGpuMemoryClustersResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ComputeGpuMemoryClusterCollection instances
	ComputeGpuMemoryClusterCollection `presentIn:"body"`

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

func (response ListComputeGpuMemoryClustersResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeGpuMemoryClustersResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeGpuMemoryClustersSortByEnum Enum with underlying type: string
type ListComputeGpuMemoryClustersSortByEnum string

// Set of constants representing the allowable values for ListComputeGpuMemoryClustersSortByEnum
const (
	ListComputeGpuMemoryClustersSortByTimecreated ListComputeGpuMemoryClustersSortByEnum = "TIMECREATED"
	ListComputeGpuMemoryClustersSortByDisplayname ListComputeGpuMemoryClustersSortByEnum = "DISPLAYNAME"
)

var mappingListComputeGpuMemoryClustersSortByEnum = map[string]ListComputeGpuMemoryClustersSortByEnum{
	"TIMECREATED": ListComputeGpuMemoryClustersSortByTimecreated,
	"DISPLAYNAME": ListComputeGpuMemoryClustersSortByDisplayname,
}

var mappingListComputeGpuMemoryClustersSortByEnumLowerCase = map[string]ListComputeGpuMemoryClustersSortByEnum{
	"timecreated": ListComputeGpuMemoryClustersSortByTimecreated,
	"displayname": ListComputeGpuMemoryClustersSortByDisplayname,
}

// GetListComputeGpuMemoryClustersSortByEnumValues Enumerates the set of values for ListComputeGpuMemoryClustersSortByEnum
func GetListComputeGpuMemoryClustersSortByEnumValues() []ListComputeGpuMemoryClustersSortByEnum {
	values := make([]ListComputeGpuMemoryClustersSortByEnum, 0)
	for _, v := range mappingListComputeGpuMemoryClustersSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeGpuMemoryClustersSortByEnumStringValues Enumerates the set of values in String for ListComputeGpuMemoryClustersSortByEnum
func GetListComputeGpuMemoryClustersSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeGpuMemoryClustersSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeGpuMemoryClustersSortByEnum(val string) (ListComputeGpuMemoryClustersSortByEnum, bool) {
	enum, ok := mappingListComputeGpuMemoryClustersSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeGpuMemoryClustersSortOrderEnum Enum with underlying type: string
type ListComputeGpuMemoryClustersSortOrderEnum string

// Set of constants representing the allowable values for ListComputeGpuMemoryClustersSortOrderEnum
const (
	ListComputeGpuMemoryClustersSortOrderAsc  ListComputeGpuMemoryClustersSortOrderEnum = "ASC"
	ListComputeGpuMemoryClustersSortOrderDesc ListComputeGpuMemoryClustersSortOrderEnum = "DESC"
)

var mappingListComputeGpuMemoryClustersSortOrderEnum = map[string]ListComputeGpuMemoryClustersSortOrderEnum{
	"ASC":  ListComputeGpuMemoryClustersSortOrderAsc,
	"DESC": ListComputeGpuMemoryClustersSortOrderDesc,
}

var mappingListComputeGpuMemoryClustersSortOrderEnumLowerCase = map[string]ListComputeGpuMemoryClustersSortOrderEnum{
	"asc":  ListComputeGpuMemoryClustersSortOrderAsc,
	"desc": ListComputeGpuMemoryClustersSortOrderDesc,
}

// GetListComputeGpuMemoryClustersSortOrderEnumValues Enumerates the set of values for ListComputeGpuMemoryClustersSortOrderEnum
func GetListComputeGpuMemoryClustersSortOrderEnumValues() []ListComputeGpuMemoryClustersSortOrderEnum {
	values := make([]ListComputeGpuMemoryClustersSortOrderEnum, 0)
	for _, v := range mappingListComputeGpuMemoryClustersSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeGpuMemoryClustersSortOrderEnumStringValues Enumerates the set of values in String for ListComputeGpuMemoryClustersSortOrderEnum
func GetListComputeGpuMemoryClustersSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeGpuMemoryClustersSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeGpuMemoryClustersSortOrderEnum(val string) (ListComputeGpuMemoryClustersSortOrderEnum, bool) {
	enum, ok := mappingListComputeGpuMemoryClustersSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
