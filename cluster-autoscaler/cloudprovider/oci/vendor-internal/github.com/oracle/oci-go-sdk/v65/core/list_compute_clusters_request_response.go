// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"net/http"
	"strings"
)

// ListComputeClustersRequest wrapper for the ListComputeClusters operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeClusters.go.html to see an example of how to use ListComputeClustersRequest.
type ListComputeClustersRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListComputeClustersSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeClustersSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeClustersRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeClustersRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeClustersRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeClustersRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeClustersRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListComputeClustersSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeClustersSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeClustersSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeClustersSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeClustersResponse wrapper for the ListComputeClusters operation
type ListComputeClustersResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ComputeClusterCollection instances
	ComputeClusterCollection `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeClustersResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeClustersResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeClustersSortByEnum Enum with underlying type: string
type ListComputeClustersSortByEnum string

// Set of constants representing the allowable values for ListComputeClustersSortByEnum
const (
	ListComputeClustersSortByTimecreated ListComputeClustersSortByEnum = "TIMECREATED"
	ListComputeClustersSortByDisplayname ListComputeClustersSortByEnum = "DISPLAYNAME"
)

var mappingListComputeClustersSortByEnum = map[string]ListComputeClustersSortByEnum{
	"TIMECREATED": ListComputeClustersSortByTimecreated,
	"DISPLAYNAME": ListComputeClustersSortByDisplayname,
}

var mappingListComputeClustersSortByEnumLowerCase = map[string]ListComputeClustersSortByEnum{
	"timecreated": ListComputeClustersSortByTimecreated,
	"displayname": ListComputeClustersSortByDisplayname,
}

// GetListComputeClustersSortByEnumValues Enumerates the set of values for ListComputeClustersSortByEnum
func GetListComputeClustersSortByEnumValues() []ListComputeClustersSortByEnum {
	values := make([]ListComputeClustersSortByEnum, 0)
	for _, v := range mappingListComputeClustersSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeClustersSortByEnumStringValues Enumerates the set of values in String for ListComputeClustersSortByEnum
func GetListComputeClustersSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeClustersSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeClustersSortByEnum(val string) (ListComputeClustersSortByEnum, bool) {
	enum, ok := mappingListComputeClustersSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeClustersSortOrderEnum Enum with underlying type: string
type ListComputeClustersSortOrderEnum string

// Set of constants representing the allowable values for ListComputeClustersSortOrderEnum
const (
	ListComputeClustersSortOrderAsc  ListComputeClustersSortOrderEnum = "ASC"
	ListComputeClustersSortOrderDesc ListComputeClustersSortOrderEnum = "DESC"
)

var mappingListComputeClustersSortOrderEnum = map[string]ListComputeClustersSortOrderEnum{
	"ASC":  ListComputeClustersSortOrderAsc,
	"DESC": ListComputeClustersSortOrderDesc,
}

var mappingListComputeClustersSortOrderEnumLowerCase = map[string]ListComputeClustersSortOrderEnum{
	"asc":  ListComputeClustersSortOrderAsc,
	"desc": ListComputeClustersSortOrderDesc,
}

// GetListComputeClustersSortOrderEnumValues Enumerates the set of values for ListComputeClustersSortOrderEnum
func GetListComputeClustersSortOrderEnumValues() []ListComputeClustersSortOrderEnum {
	values := make([]ListComputeClustersSortOrderEnum, 0)
	for _, v := range mappingListComputeClustersSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeClustersSortOrderEnumStringValues Enumerates the set of values in String for ListComputeClustersSortOrderEnum
func GetListComputeClustersSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeClustersSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeClustersSortOrderEnum(val string) (ListComputeClustersSortOrderEnum, bool) {
	enum, ok := mappingListComputeClustersSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
