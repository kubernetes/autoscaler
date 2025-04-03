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

// ListCrossConnectGroupsRequest wrapper for the ListCrossConnectGroups operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListCrossConnectGroups.go.html to see an example of how to use ListCrossConnectGroupsRequest.
type ListCrossConnectGroupsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListCrossConnectGroupsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListCrossConnectGroupsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only resources that match the specified lifecycle
	// state. The value is case insensitive.
	LifecycleState CrossConnectGroupLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListCrossConnectGroupsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListCrossConnectGroupsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListCrossConnectGroupsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListCrossConnectGroupsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListCrossConnectGroupsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListCrossConnectGroupsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListCrossConnectGroupsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListCrossConnectGroupsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListCrossConnectGroupsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCrossConnectGroupLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetCrossConnectGroupLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListCrossConnectGroupsResponse wrapper for the ListCrossConnectGroups operation
type ListCrossConnectGroupsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []CrossConnectGroup instances
	Items []CrossConnectGroup `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListCrossConnectGroupsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListCrossConnectGroupsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListCrossConnectGroupsSortByEnum Enum with underlying type: string
type ListCrossConnectGroupsSortByEnum string

// Set of constants representing the allowable values for ListCrossConnectGroupsSortByEnum
const (
	ListCrossConnectGroupsSortByTimecreated ListCrossConnectGroupsSortByEnum = "TIMECREATED"
	ListCrossConnectGroupsSortByDisplayname ListCrossConnectGroupsSortByEnum = "DISPLAYNAME"
)

var mappingListCrossConnectGroupsSortByEnum = map[string]ListCrossConnectGroupsSortByEnum{
	"TIMECREATED": ListCrossConnectGroupsSortByTimecreated,
	"DISPLAYNAME": ListCrossConnectGroupsSortByDisplayname,
}

var mappingListCrossConnectGroupsSortByEnumLowerCase = map[string]ListCrossConnectGroupsSortByEnum{
	"timecreated": ListCrossConnectGroupsSortByTimecreated,
	"displayname": ListCrossConnectGroupsSortByDisplayname,
}

// GetListCrossConnectGroupsSortByEnumValues Enumerates the set of values for ListCrossConnectGroupsSortByEnum
func GetListCrossConnectGroupsSortByEnumValues() []ListCrossConnectGroupsSortByEnum {
	values := make([]ListCrossConnectGroupsSortByEnum, 0)
	for _, v := range mappingListCrossConnectGroupsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListCrossConnectGroupsSortByEnumStringValues Enumerates the set of values in String for ListCrossConnectGroupsSortByEnum
func GetListCrossConnectGroupsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListCrossConnectGroupsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListCrossConnectGroupsSortByEnum(val string) (ListCrossConnectGroupsSortByEnum, bool) {
	enum, ok := mappingListCrossConnectGroupsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListCrossConnectGroupsSortOrderEnum Enum with underlying type: string
type ListCrossConnectGroupsSortOrderEnum string

// Set of constants representing the allowable values for ListCrossConnectGroupsSortOrderEnum
const (
	ListCrossConnectGroupsSortOrderAsc  ListCrossConnectGroupsSortOrderEnum = "ASC"
	ListCrossConnectGroupsSortOrderDesc ListCrossConnectGroupsSortOrderEnum = "DESC"
)

var mappingListCrossConnectGroupsSortOrderEnum = map[string]ListCrossConnectGroupsSortOrderEnum{
	"ASC":  ListCrossConnectGroupsSortOrderAsc,
	"DESC": ListCrossConnectGroupsSortOrderDesc,
}

var mappingListCrossConnectGroupsSortOrderEnumLowerCase = map[string]ListCrossConnectGroupsSortOrderEnum{
	"asc":  ListCrossConnectGroupsSortOrderAsc,
	"desc": ListCrossConnectGroupsSortOrderDesc,
}

// GetListCrossConnectGroupsSortOrderEnumValues Enumerates the set of values for ListCrossConnectGroupsSortOrderEnum
func GetListCrossConnectGroupsSortOrderEnumValues() []ListCrossConnectGroupsSortOrderEnum {
	values := make([]ListCrossConnectGroupsSortOrderEnum, 0)
	for _, v := range mappingListCrossConnectGroupsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListCrossConnectGroupsSortOrderEnumStringValues Enumerates the set of values in String for ListCrossConnectGroupsSortOrderEnum
func GetListCrossConnectGroupsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListCrossConnectGroupsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListCrossConnectGroupsSortOrderEnum(val string) (ListCrossConnectGroupsSortOrderEnum, bool) {
	enum, ok := mappingListCrossConnectGroupsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
