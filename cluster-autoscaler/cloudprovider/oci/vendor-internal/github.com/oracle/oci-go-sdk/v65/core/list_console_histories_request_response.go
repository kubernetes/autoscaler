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

// ListConsoleHistoriesRequest wrapper for the ListConsoleHistories operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListConsoleHistories.go.html to see an example of how to use ListConsoleHistoriesRequest.
type ListConsoleHistoriesRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The OCID of the instance.
	InstanceId *string `mandatory:"false" contributesTo:"query" name:"instanceId"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListConsoleHistoriesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListConsoleHistoriesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle state. The state
	// value is case-insensitive.
	LifecycleState ConsoleHistoryLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListConsoleHistoriesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListConsoleHistoriesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListConsoleHistoriesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListConsoleHistoriesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListConsoleHistoriesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListConsoleHistoriesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListConsoleHistoriesSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListConsoleHistoriesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListConsoleHistoriesSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingConsoleHistoryLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetConsoleHistoryLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListConsoleHistoriesResponse wrapper for the ListConsoleHistories operation
type ListConsoleHistoriesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ConsoleHistory instances
	Items []ConsoleHistory `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListConsoleHistoriesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListConsoleHistoriesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListConsoleHistoriesSortByEnum Enum with underlying type: string
type ListConsoleHistoriesSortByEnum string

// Set of constants representing the allowable values for ListConsoleHistoriesSortByEnum
const (
	ListConsoleHistoriesSortByTimecreated ListConsoleHistoriesSortByEnum = "TIMECREATED"
	ListConsoleHistoriesSortByDisplayname ListConsoleHistoriesSortByEnum = "DISPLAYNAME"
)

var mappingListConsoleHistoriesSortByEnum = map[string]ListConsoleHistoriesSortByEnum{
	"TIMECREATED": ListConsoleHistoriesSortByTimecreated,
	"DISPLAYNAME": ListConsoleHistoriesSortByDisplayname,
}

var mappingListConsoleHistoriesSortByEnumLowerCase = map[string]ListConsoleHistoriesSortByEnum{
	"timecreated": ListConsoleHistoriesSortByTimecreated,
	"displayname": ListConsoleHistoriesSortByDisplayname,
}

// GetListConsoleHistoriesSortByEnumValues Enumerates the set of values for ListConsoleHistoriesSortByEnum
func GetListConsoleHistoriesSortByEnumValues() []ListConsoleHistoriesSortByEnum {
	values := make([]ListConsoleHistoriesSortByEnum, 0)
	for _, v := range mappingListConsoleHistoriesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListConsoleHistoriesSortByEnumStringValues Enumerates the set of values in String for ListConsoleHistoriesSortByEnum
func GetListConsoleHistoriesSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListConsoleHistoriesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListConsoleHistoriesSortByEnum(val string) (ListConsoleHistoriesSortByEnum, bool) {
	enum, ok := mappingListConsoleHistoriesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListConsoleHistoriesSortOrderEnum Enum with underlying type: string
type ListConsoleHistoriesSortOrderEnum string

// Set of constants representing the allowable values for ListConsoleHistoriesSortOrderEnum
const (
	ListConsoleHistoriesSortOrderAsc  ListConsoleHistoriesSortOrderEnum = "ASC"
	ListConsoleHistoriesSortOrderDesc ListConsoleHistoriesSortOrderEnum = "DESC"
)

var mappingListConsoleHistoriesSortOrderEnum = map[string]ListConsoleHistoriesSortOrderEnum{
	"ASC":  ListConsoleHistoriesSortOrderAsc,
	"DESC": ListConsoleHistoriesSortOrderDesc,
}

var mappingListConsoleHistoriesSortOrderEnumLowerCase = map[string]ListConsoleHistoriesSortOrderEnum{
	"asc":  ListConsoleHistoriesSortOrderAsc,
	"desc": ListConsoleHistoriesSortOrderDesc,
}

// GetListConsoleHistoriesSortOrderEnumValues Enumerates the set of values for ListConsoleHistoriesSortOrderEnum
func GetListConsoleHistoriesSortOrderEnumValues() []ListConsoleHistoriesSortOrderEnum {
	values := make([]ListConsoleHistoriesSortOrderEnum, 0)
	for _, v := range mappingListConsoleHistoriesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListConsoleHistoriesSortOrderEnumStringValues Enumerates the set of values in String for ListConsoleHistoriesSortOrderEnum
func GetListConsoleHistoriesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListConsoleHistoriesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListConsoleHistoriesSortOrderEnum(val string) (ListConsoleHistoriesSortOrderEnum, bool) {
	enum, ok := mappingListConsoleHistoriesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
