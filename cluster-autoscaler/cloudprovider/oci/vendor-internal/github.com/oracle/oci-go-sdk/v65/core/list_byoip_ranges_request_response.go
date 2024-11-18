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

// ListByoipRangesRequest wrapper for the ListByoipRanges operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListByoipRanges.go.html to see an example of how to use ListByoipRangesRequest.
type ListByoipRangesRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

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

	// A filter to return only resources that match the given lifecycle state name exactly.
	LifecycleState *string `mandatory:"false" contributesTo:"query" name:"lifecycleState"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListByoipRangesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListByoipRangesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListByoipRangesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListByoipRangesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListByoipRangesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListByoipRangesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListByoipRangesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListByoipRangesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListByoipRangesSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListByoipRangesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListByoipRangesSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListByoipRangesResponse wrapper for the ListByoipRanges operation
type ListByoipRangesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ByoipRangeCollection instances
	ByoipRangeCollection `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListByoipRangesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListByoipRangesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListByoipRangesSortByEnum Enum with underlying type: string
type ListByoipRangesSortByEnum string

// Set of constants representing the allowable values for ListByoipRangesSortByEnum
const (
	ListByoipRangesSortByTimecreated ListByoipRangesSortByEnum = "TIMECREATED"
	ListByoipRangesSortByDisplayname ListByoipRangesSortByEnum = "DISPLAYNAME"
)

var mappingListByoipRangesSortByEnum = map[string]ListByoipRangesSortByEnum{
	"TIMECREATED": ListByoipRangesSortByTimecreated,
	"DISPLAYNAME": ListByoipRangesSortByDisplayname,
}

var mappingListByoipRangesSortByEnumLowerCase = map[string]ListByoipRangesSortByEnum{
	"timecreated": ListByoipRangesSortByTimecreated,
	"displayname": ListByoipRangesSortByDisplayname,
}

// GetListByoipRangesSortByEnumValues Enumerates the set of values for ListByoipRangesSortByEnum
func GetListByoipRangesSortByEnumValues() []ListByoipRangesSortByEnum {
	values := make([]ListByoipRangesSortByEnum, 0)
	for _, v := range mappingListByoipRangesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListByoipRangesSortByEnumStringValues Enumerates the set of values in String for ListByoipRangesSortByEnum
func GetListByoipRangesSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListByoipRangesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListByoipRangesSortByEnum(val string) (ListByoipRangesSortByEnum, bool) {
	enum, ok := mappingListByoipRangesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListByoipRangesSortOrderEnum Enum with underlying type: string
type ListByoipRangesSortOrderEnum string

// Set of constants representing the allowable values for ListByoipRangesSortOrderEnum
const (
	ListByoipRangesSortOrderAsc  ListByoipRangesSortOrderEnum = "ASC"
	ListByoipRangesSortOrderDesc ListByoipRangesSortOrderEnum = "DESC"
)

var mappingListByoipRangesSortOrderEnum = map[string]ListByoipRangesSortOrderEnum{
	"ASC":  ListByoipRangesSortOrderAsc,
	"DESC": ListByoipRangesSortOrderDesc,
}

var mappingListByoipRangesSortOrderEnumLowerCase = map[string]ListByoipRangesSortOrderEnum{
	"asc":  ListByoipRangesSortOrderAsc,
	"desc": ListByoipRangesSortOrderDesc,
}

// GetListByoipRangesSortOrderEnumValues Enumerates the set of values for ListByoipRangesSortOrderEnum
func GetListByoipRangesSortOrderEnumValues() []ListByoipRangesSortOrderEnum {
	values := make([]ListByoipRangesSortOrderEnum, 0)
	for _, v := range mappingListByoipRangesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListByoipRangesSortOrderEnumStringValues Enumerates the set of values in String for ListByoipRangesSortOrderEnum
func GetListByoipRangesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListByoipRangesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListByoipRangesSortOrderEnum(val string) (ListByoipRangesSortOrderEnum, bool) {
	enum, ok := mappingListByoipRangesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
