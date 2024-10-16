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

// ListAppCatalogSubscriptionsRequest wrapper for the ListAppCatalogSubscriptions operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListAppCatalogSubscriptions.go.html to see an example of how to use ListAppCatalogSubscriptionsRequest.
type ListAppCatalogSubscriptionsRequest struct {

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

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListAppCatalogSubscriptionsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListAppCatalogSubscriptionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only the listings that matches the given listing id.
	ListingId *string `mandatory:"false" contributesTo:"query" name:"listingId"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListAppCatalogSubscriptionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListAppCatalogSubscriptionsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListAppCatalogSubscriptionsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListAppCatalogSubscriptionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListAppCatalogSubscriptionsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListAppCatalogSubscriptionsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListAppCatalogSubscriptionsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListAppCatalogSubscriptionsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListAppCatalogSubscriptionsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListAppCatalogSubscriptionsResponse wrapper for the ListAppCatalogSubscriptions operation
type ListAppCatalogSubscriptionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []AppCatalogSubscriptionSummary instances
	Items []AppCatalogSubscriptionSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListAppCatalogSubscriptionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListAppCatalogSubscriptionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListAppCatalogSubscriptionsSortByEnum Enum with underlying type: string
type ListAppCatalogSubscriptionsSortByEnum string

// Set of constants representing the allowable values for ListAppCatalogSubscriptionsSortByEnum
const (
	ListAppCatalogSubscriptionsSortByTimecreated ListAppCatalogSubscriptionsSortByEnum = "TIMECREATED"
	ListAppCatalogSubscriptionsSortByDisplayname ListAppCatalogSubscriptionsSortByEnum = "DISPLAYNAME"
)

var mappingListAppCatalogSubscriptionsSortByEnum = map[string]ListAppCatalogSubscriptionsSortByEnum{
	"TIMECREATED": ListAppCatalogSubscriptionsSortByTimecreated,
	"DISPLAYNAME": ListAppCatalogSubscriptionsSortByDisplayname,
}

var mappingListAppCatalogSubscriptionsSortByEnumLowerCase = map[string]ListAppCatalogSubscriptionsSortByEnum{
	"timecreated": ListAppCatalogSubscriptionsSortByTimecreated,
	"displayname": ListAppCatalogSubscriptionsSortByDisplayname,
}

// GetListAppCatalogSubscriptionsSortByEnumValues Enumerates the set of values for ListAppCatalogSubscriptionsSortByEnum
func GetListAppCatalogSubscriptionsSortByEnumValues() []ListAppCatalogSubscriptionsSortByEnum {
	values := make([]ListAppCatalogSubscriptionsSortByEnum, 0)
	for _, v := range mappingListAppCatalogSubscriptionsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListAppCatalogSubscriptionsSortByEnumStringValues Enumerates the set of values in String for ListAppCatalogSubscriptionsSortByEnum
func GetListAppCatalogSubscriptionsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListAppCatalogSubscriptionsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListAppCatalogSubscriptionsSortByEnum(val string) (ListAppCatalogSubscriptionsSortByEnum, bool) {
	enum, ok := mappingListAppCatalogSubscriptionsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListAppCatalogSubscriptionsSortOrderEnum Enum with underlying type: string
type ListAppCatalogSubscriptionsSortOrderEnum string

// Set of constants representing the allowable values for ListAppCatalogSubscriptionsSortOrderEnum
const (
	ListAppCatalogSubscriptionsSortOrderAsc  ListAppCatalogSubscriptionsSortOrderEnum = "ASC"
	ListAppCatalogSubscriptionsSortOrderDesc ListAppCatalogSubscriptionsSortOrderEnum = "DESC"
)

var mappingListAppCatalogSubscriptionsSortOrderEnum = map[string]ListAppCatalogSubscriptionsSortOrderEnum{
	"ASC":  ListAppCatalogSubscriptionsSortOrderAsc,
	"DESC": ListAppCatalogSubscriptionsSortOrderDesc,
}

var mappingListAppCatalogSubscriptionsSortOrderEnumLowerCase = map[string]ListAppCatalogSubscriptionsSortOrderEnum{
	"asc":  ListAppCatalogSubscriptionsSortOrderAsc,
	"desc": ListAppCatalogSubscriptionsSortOrderDesc,
}

// GetListAppCatalogSubscriptionsSortOrderEnumValues Enumerates the set of values for ListAppCatalogSubscriptionsSortOrderEnum
func GetListAppCatalogSubscriptionsSortOrderEnumValues() []ListAppCatalogSubscriptionsSortOrderEnum {
	values := make([]ListAppCatalogSubscriptionsSortOrderEnum, 0)
	for _, v := range mappingListAppCatalogSubscriptionsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListAppCatalogSubscriptionsSortOrderEnumStringValues Enumerates the set of values in String for ListAppCatalogSubscriptionsSortOrderEnum
func GetListAppCatalogSubscriptionsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListAppCatalogSubscriptionsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListAppCatalogSubscriptionsSortOrderEnum(val string) (ListAppCatalogSubscriptionsSortOrderEnum, bool) {
	enum, ok := mappingListAppCatalogSubscriptionsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
