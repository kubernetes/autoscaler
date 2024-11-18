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

// ListSecurityListsRequest wrapper for the ListSecurityLists operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListSecurityLists.go.html to see an example of how to use ListSecurityListsRequest.
type ListSecurityListsRequest struct {

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
	SortBy ListSecurityListsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListSecurityListsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle
	// state. The state value is case-insensitive.
	LifecycleState SecurityListLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListSecurityListsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListSecurityListsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListSecurityListsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListSecurityListsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListSecurityListsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListSecurityListsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListSecurityListsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListSecurityListsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListSecurityListsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingSecurityListLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetSecurityListLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListSecurityListsResponse wrapper for the ListSecurityLists operation
type ListSecurityListsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []SecurityList instances
	Items []SecurityList `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListSecurityListsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListSecurityListsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListSecurityListsSortByEnum Enum with underlying type: string
type ListSecurityListsSortByEnum string

// Set of constants representing the allowable values for ListSecurityListsSortByEnum
const (
	ListSecurityListsSortByTimecreated ListSecurityListsSortByEnum = "TIMECREATED"
	ListSecurityListsSortByDisplayname ListSecurityListsSortByEnum = "DISPLAYNAME"
)

var mappingListSecurityListsSortByEnum = map[string]ListSecurityListsSortByEnum{
	"TIMECREATED": ListSecurityListsSortByTimecreated,
	"DISPLAYNAME": ListSecurityListsSortByDisplayname,
}

var mappingListSecurityListsSortByEnumLowerCase = map[string]ListSecurityListsSortByEnum{
	"timecreated": ListSecurityListsSortByTimecreated,
	"displayname": ListSecurityListsSortByDisplayname,
}

// GetListSecurityListsSortByEnumValues Enumerates the set of values for ListSecurityListsSortByEnum
func GetListSecurityListsSortByEnumValues() []ListSecurityListsSortByEnum {
	values := make([]ListSecurityListsSortByEnum, 0)
	for _, v := range mappingListSecurityListsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListSecurityListsSortByEnumStringValues Enumerates the set of values in String for ListSecurityListsSortByEnum
func GetListSecurityListsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListSecurityListsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListSecurityListsSortByEnum(val string) (ListSecurityListsSortByEnum, bool) {
	enum, ok := mappingListSecurityListsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListSecurityListsSortOrderEnum Enum with underlying type: string
type ListSecurityListsSortOrderEnum string

// Set of constants representing the allowable values for ListSecurityListsSortOrderEnum
const (
	ListSecurityListsSortOrderAsc  ListSecurityListsSortOrderEnum = "ASC"
	ListSecurityListsSortOrderDesc ListSecurityListsSortOrderEnum = "DESC"
)

var mappingListSecurityListsSortOrderEnum = map[string]ListSecurityListsSortOrderEnum{
	"ASC":  ListSecurityListsSortOrderAsc,
	"DESC": ListSecurityListsSortOrderDesc,
}

var mappingListSecurityListsSortOrderEnumLowerCase = map[string]ListSecurityListsSortOrderEnum{
	"asc":  ListSecurityListsSortOrderAsc,
	"desc": ListSecurityListsSortOrderDesc,
}

// GetListSecurityListsSortOrderEnumValues Enumerates the set of values for ListSecurityListsSortOrderEnum
func GetListSecurityListsSortOrderEnumValues() []ListSecurityListsSortOrderEnum {
	values := make([]ListSecurityListsSortOrderEnum, 0)
	for _, v := range mappingListSecurityListsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListSecurityListsSortOrderEnumStringValues Enumerates the set of values in String for ListSecurityListsSortOrderEnum
func GetListSecurityListsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListSecurityListsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListSecurityListsSortOrderEnum(val string) (ListSecurityListsSortOrderEnum, bool) {
	enum, ok := mappingListSecurityListsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
