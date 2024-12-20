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

// ListInstancePoolsRequest wrapper for the ListInstancePools operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListInstancePools.go.html to see an example of how to use ListInstancePoolsRequest.
type ListInstancePoolsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

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
	SortBy ListInstancePoolsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListInstancePoolsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle state. The state
	// value is case-insensitive.
	LifecycleState InstancePoolSummaryLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListInstancePoolsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListInstancePoolsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListInstancePoolsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListInstancePoolsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListInstancePoolsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListInstancePoolsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListInstancePoolsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListInstancePoolsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListInstancePoolsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingInstancePoolSummaryLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetInstancePoolSummaryLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListInstancePoolsResponse wrapper for the ListInstancePools operation
type ListInstancePoolsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []InstancePoolSummary instances
	Items []InstancePoolSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListInstancePoolsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListInstancePoolsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListInstancePoolsSortByEnum Enum with underlying type: string
type ListInstancePoolsSortByEnum string

// Set of constants representing the allowable values for ListInstancePoolsSortByEnum
const (
	ListInstancePoolsSortByTimecreated ListInstancePoolsSortByEnum = "TIMECREATED"
	ListInstancePoolsSortByDisplayname ListInstancePoolsSortByEnum = "DISPLAYNAME"
)

var mappingListInstancePoolsSortByEnum = map[string]ListInstancePoolsSortByEnum{
	"TIMECREATED": ListInstancePoolsSortByTimecreated,
	"DISPLAYNAME": ListInstancePoolsSortByDisplayname,
}

var mappingListInstancePoolsSortByEnumLowerCase = map[string]ListInstancePoolsSortByEnum{
	"timecreated": ListInstancePoolsSortByTimecreated,
	"displayname": ListInstancePoolsSortByDisplayname,
}

// GetListInstancePoolsSortByEnumValues Enumerates the set of values for ListInstancePoolsSortByEnum
func GetListInstancePoolsSortByEnumValues() []ListInstancePoolsSortByEnum {
	values := make([]ListInstancePoolsSortByEnum, 0)
	for _, v := range mappingListInstancePoolsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListInstancePoolsSortByEnumStringValues Enumerates the set of values in String for ListInstancePoolsSortByEnum
func GetListInstancePoolsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListInstancePoolsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListInstancePoolsSortByEnum(val string) (ListInstancePoolsSortByEnum, bool) {
	enum, ok := mappingListInstancePoolsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListInstancePoolsSortOrderEnum Enum with underlying type: string
type ListInstancePoolsSortOrderEnum string

// Set of constants representing the allowable values for ListInstancePoolsSortOrderEnum
const (
	ListInstancePoolsSortOrderAsc  ListInstancePoolsSortOrderEnum = "ASC"
	ListInstancePoolsSortOrderDesc ListInstancePoolsSortOrderEnum = "DESC"
)

var mappingListInstancePoolsSortOrderEnum = map[string]ListInstancePoolsSortOrderEnum{
	"ASC":  ListInstancePoolsSortOrderAsc,
	"DESC": ListInstancePoolsSortOrderDesc,
}

var mappingListInstancePoolsSortOrderEnumLowerCase = map[string]ListInstancePoolsSortOrderEnum{
	"asc":  ListInstancePoolsSortOrderAsc,
	"desc": ListInstancePoolsSortOrderDesc,
}

// GetListInstancePoolsSortOrderEnumValues Enumerates the set of values for ListInstancePoolsSortOrderEnum
func GetListInstancePoolsSortOrderEnumValues() []ListInstancePoolsSortOrderEnum {
	values := make([]ListInstancePoolsSortOrderEnum, 0)
	for _, v := range mappingListInstancePoolsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListInstancePoolsSortOrderEnumStringValues Enumerates the set of values in String for ListInstancePoolsSortOrderEnum
func GetListInstancePoolsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListInstancePoolsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListInstancePoolsSortOrderEnum(val string) (ListInstancePoolsSortOrderEnum, bool) {
	enum, ok := mappingListInstancePoolsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
