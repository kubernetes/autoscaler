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

// ListDrgRouteTablesRequest wrapper for the ListDrgRouteTables operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListDrgRouteTables.go.html to see an example of how to use ListDrgRouteTablesRequest.
type ListDrgRouteTablesRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DRG.
	DrgId *string `mandatory:"true" contributesTo:"query" name:"drgId"`

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
	SortBy ListDrgRouteTablesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListDrgRouteTablesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the import route distribution.
	ImportDrgRouteDistributionId *string `mandatory:"false" contributesTo:"query" name:"importDrgRouteDistributionId"`

	// A filter that only returns matches for the specified lifecycle
	// state. The value is case insensitive.
	LifecycleState DrgRouteTableLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListDrgRouteTablesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListDrgRouteTablesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListDrgRouteTablesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListDrgRouteTablesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListDrgRouteTablesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListDrgRouteTablesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListDrgRouteTablesSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListDrgRouteTablesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListDrgRouteTablesSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingDrgRouteTableLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetDrgRouteTableLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListDrgRouteTablesResponse wrapper for the ListDrgRouteTables operation
type ListDrgRouteTablesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []DrgRouteTable instances
	Items []DrgRouteTable `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListDrgRouteTablesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListDrgRouteTablesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListDrgRouteTablesSortByEnum Enum with underlying type: string
type ListDrgRouteTablesSortByEnum string

// Set of constants representing the allowable values for ListDrgRouteTablesSortByEnum
const (
	ListDrgRouteTablesSortByTimecreated ListDrgRouteTablesSortByEnum = "TIMECREATED"
	ListDrgRouteTablesSortByDisplayname ListDrgRouteTablesSortByEnum = "DISPLAYNAME"
)

var mappingListDrgRouteTablesSortByEnum = map[string]ListDrgRouteTablesSortByEnum{
	"TIMECREATED": ListDrgRouteTablesSortByTimecreated,
	"DISPLAYNAME": ListDrgRouteTablesSortByDisplayname,
}

var mappingListDrgRouteTablesSortByEnumLowerCase = map[string]ListDrgRouteTablesSortByEnum{
	"timecreated": ListDrgRouteTablesSortByTimecreated,
	"displayname": ListDrgRouteTablesSortByDisplayname,
}

// GetListDrgRouteTablesSortByEnumValues Enumerates the set of values for ListDrgRouteTablesSortByEnum
func GetListDrgRouteTablesSortByEnumValues() []ListDrgRouteTablesSortByEnum {
	values := make([]ListDrgRouteTablesSortByEnum, 0)
	for _, v := range mappingListDrgRouteTablesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListDrgRouteTablesSortByEnumStringValues Enumerates the set of values in String for ListDrgRouteTablesSortByEnum
func GetListDrgRouteTablesSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListDrgRouteTablesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListDrgRouteTablesSortByEnum(val string) (ListDrgRouteTablesSortByEnum, bool) {
	enum, ok := mappingListDrgRouteTablesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListDrgRouteTablesSortOrderEnum Enum with underlying type: string
type ListDrgRouteTablesSortOrderEnum string

// Set of constants representing the allowable values for ListDrgRouteTablesSortOrderEnum
const (
	ListDrgRouteTablesSortOrderAsc  ListDrgRouteTablesSortOrderEnum = "ASC"
	ListDrgRouteTablesSortOrderDesc ListDrgRouteTablesSortOrderEnum = "DESC"
)

var mappingListDrgRouteTablesSortOrderEnum = map[string]ListDrgRouteTablesSortOrderEnum{
	"ASC":  ListDrgRouteTablesSortOrderAsc,
	"DESC": ListDrgRouteTablesSortOrderDesc,
}

var mappingListDrgRouteTablesSortOrderEnumLowerCase = map[string]ListDrgRouteTablesSortOrderEnum{
	"asc":  ListDrgRouteTablesSortOrderAsc,
	"desc": ListDrgRouteTablesSortOrderDesc,
}

// GetListDrgRouteTablesSortOrderEnumValues Enumerates the set of values for ListDrgRouteTablesSortOrderEnum
func GetListDrgRouteTablesSortOrderEnumValues() []ListDrgRouteTablesSortOrderEnum {
	values := make([]ListDrgRouteTablesSortOrderEnum, 0)
	for _, v := range mappingListDrgRouteTablesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListDrgRouteTablesSortOrderEnumStringValues Enumerates the set of values in String for ListDrgRouteTablesSortOrderEnum
func GetListDrgRouteTablesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListDrgRouteTablesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListDrgRouteTablesSortOrderEnum(val string) (ListDrgRouteTablesSortOrderEnum, bool) {
	enum, ok := mappingListDrgRouteTablesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
