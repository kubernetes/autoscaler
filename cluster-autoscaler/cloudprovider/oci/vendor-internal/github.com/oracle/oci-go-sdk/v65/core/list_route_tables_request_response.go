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

// ListRouteTablesRequest wrapper for the ListRouteTables operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListRouteTables.go.html to see an example of how to use ListRouteTablesRequest.
type ListRouteTablesRequest struct {

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

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN.
	VcnId *string `mandatory:"false" contributesTo:"query" name:"vcnId"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListRouteTablesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListRouteTablesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle
	// state. The state value is case-insensitive.
	LifecycleState RouteTableLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListRouteTablesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListRouteTablesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListRouteTablesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListRouteTablesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListRouteTablesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListRouteTablesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListRouteTablesSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListRouteTablesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListRouteTablesSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingRouteTableLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetRouteTableLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListRouteTablesResponse wrapper for the ListRouteTables operation
type ListRouteTablesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []RouteTable instances
	Items []RouteTable `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListRouteTablesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListRouteTablesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListRouteTablesSortByEnum Enum with underlying type: string
type ListRouteTablesSortByEnum string

// Set of constants representing the allowable values for ListRouteTablesSortByEnum
const (
	ListRouteTablesSortByTimecreated ListRouteTablesSortByEnum = "TIMECREATED"
	ListRouteTablesSortByDisplayname ListRouteTablesSortByEnum = "DISPLAYNAME"
)

var mappingListRouteTablesSortByEnum = map[string]ListRouteTablesSortByEnum{
	"TIMECREATED": ListRouteTablesSortByTimecreated,
	"DISPLAYNAME": ListRouteTablesSortByDisplayname,
}

var mappingListRouteTablesSortByEnumLowerCase = map[string]ListRouteTablesSortByEnum{
	"timecreated": ListRouteTablesSortByTimecreated,
	"displayname": ListRouteTablesSortByDisplayname,
}

// GetListRouteTablesSortByEnumValues Enumerates the set of values for ListRouteTablesSortByEnum
func GetListRouteTablesSortByEnumValues() []ListRouteTablesSortByEnum {
	values := make([]ListRouteTablesSortByEnum, 0)
	for _, v := range mappingListRouteTablesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListRouteTablesSortByEnumStringValues Enumerates the set of values in String for ListRouteTablesSortByEnum
func GetListRouteTablesSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListRouteTablesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListRouteTablesSortByEnum(val string) (ListRouteTablesSortByEnum, bool) {
	enum, ok := mappingListRouteTablesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListRouteTablesSortOrderEnum Enum with underlying type: string
type ListRouteTablesSortOrderEnum string

// Set of constants representing the allowable values for ListRouteTablesSortOrderEnum
const (
	ListRouteTablesSortOrderAsc  ListRouteTablesSortOrderEnum = "ASC"
	ListRouteTablesSortOrderDesc ListRouteTablesSortOrderEnum = "DESC"
)

var mappingListRouteTablesSortOrderEnum = map[string]ListRouteTablesSortOrderEnum{
	"ASC":  ListRouteTablesSortOrderAsc,
	"DESC": ListRouteTablesSortOrderDesc,
}

var mappingListRouteTablesSortOrderEnumLowerCase = map[string]ListRouteTablesSortOrderEnum{
	"asc":  ListRouteTablesSortOrderAsc,
	"desc": ListRouteTablesSortOrderDesc,
}

// GetListRouteTablesSortOrderEnumValues Enumerates the set of values for ListRouteTablesSortOrderEnum
func GetListRouteTablesSortOrderEnumValues() []ListRouteTablesSortOrderEnum {
	values := make([]ListRouteTablesSortOrderEnum, 0)
	for _, v := range mappingListRouteTablesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListRouteTablesSortOrderEnumStringValues Enumerates the set of values in String for ListRouteTablesSortOrderEnum
func GetListRouteTablesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListRouteTablesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListRouteTablesSortOrderEnum(val string) (ListRouteTablesSortOrderEnum, bool) {
	enum, ok := mappingListRouteTablesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
