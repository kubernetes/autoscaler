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

// ListByoasnsRequest wrapper for the ListByoasns operation
//
// # See also
//
// Click https://docs.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListByoasns.go.html to see an example of how to use ListByoasnsRequest.
type ListByoasnsRequest struct {

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to return only resources that match the given lifecycle state name exactly.
	LifecycleState *string `mandatory:"false" contributesTo:"query" name:"lifecycleState"`

	// The field to sort by, for byoasn List operation.
	SortBy ListByoasnsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListByoasnsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListByoasnsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListByoasnsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListByoasnsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListByoasnsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListByoasnsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListByoasnsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListByoasnsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListByoasnsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListByoasnsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListByoasnsResponse wrapper for the ListByoasns operation
type ListByoasnsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ByoasnCollection instances
	ByoasnCollection `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListByoasnsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListByoasnsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListByoasnsSortByEnum Enum with underlying type: string
type ListByoasnsSortByEnum string

// Set of constants representing the allowable values for ListByoasnsSortByEnum
const (
	ListByoasnsSortByTimecreated ListByoasnsSortByEnum = "TIMECREATED"
	ListByoasnsSortByDisplayname ListByoasnsSortByEnum = "DISPLAYNAME"
)

var mappingListByoasnsSortByEnum = map[string]ListByoasnsSortByEnum{
	"TIMECREATED": ListByoasnsSortByTimecreated,
	"DISPLAYNAME": ListByoasnsSortByDisplayname,
}

var mappingListByoasnsSortByEnumLowerCase = map[string]ListByoasnsSortByEnum{
	"timecreated": ListByoasnsSortByTimecreated,
	"displayname": ListByoasnsSortByDisplayname,
}

// GetListByoasnsSortByEnumValues Enumerates the set of values for ListByoasnsSortByEnum
func GetListByoasnsSortByEnumValues() []ListByoasnsSortByEnum {
	values := make([]ListByoasnsSortByEnum, 0)
	for _, v := range mappingListByoasnsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListByoasnsSortByEnumStringValues Enumerates the set of values in String for ListByoasnsSortByEnum
func GetListByoasnsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListByoasnsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListByoasnsSortByEnum(val string) (ListByoasnsSortByEnum, bool) {
	enum, ok := mappingListByoasnsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListByoasnsSortOrderEnum Enum with underlying type: string
type ListByoasnsSortOrderEnum string

// Set of constants representing the allowable values for ListByoasnsSortOrderEnum
const (
	ListByoasnsSortOrderAsc  ListByoasnsSortOrderEnum = "ASC"
	ListByoasnsSortOrderDesc ListByoasnsSortOrderEnum = "DESC"
)

var mappingListByoasnsSortOrderEnum = map[string]ListByoasnsSortOrderEnum{
	"ASC":  ListByoasnsSortOrderAsc,
	"DESC": ListByoasnsSortOrderDesc,
}

var mappingListByoasnsSortOrderEnumLowerCase = map[string]ListByoasnsSortOrderEnum{
	"asc":  ListByoasnsSortOrderAsc,
	"desc": ListByoasnsSortOrderDesc,
}

// GetListByoasnsSortOrderEnumValues Enumerates the set of values for ListByoasnsSortOrderEnum
func GetListByoasnsSortOrderEnumValues() []ListByoasnsSortOrderEnum {
	values := make([]ListByoasnsSortOrderEnum, 0)
	for _, v := range mappingListByoasnsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListByoasnsSortOrderEnumStringValues Enumerates the set of values in String for ListByoasnsSortOrderEnum
func GetListByoasnsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListByoasnsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListByoasnsSortOrderEnum(val string) (ListByoasnsSortOrderEnum, bool) {
	enum, ok := mappingListByoasnsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
