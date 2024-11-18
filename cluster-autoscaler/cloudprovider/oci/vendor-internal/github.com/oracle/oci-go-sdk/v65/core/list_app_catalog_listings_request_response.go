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

// ListAppCatalogListingsRequest wrapper for the ListAppCatalogListings operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListAppCatalogListings.go.html to see an example of how to use ListAppCatalogListingsRequest.
type ListAppCatalogListingsRequest struct {

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListAppCatalogListingsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only the publisher that matches the given publisher name exactly.
	PublisherName *string `mandatory:"false" contributesTo:"query" name:"publisherName"`

	// A filter to return only publishers that match the given publisher type exactly. Valid types are OCI, ORACLE, TRUSTED, STANDARD.
	PublisherType *string `mandatory:"false" contributesTo:"query" name:"publisherType"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListAppCatalogListingsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListAppCatalogListingsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListAppCatalogListingsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListAppCatalogListingsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListAppCatalogListingsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListAppCatalogListingsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListAppCatalogListingsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListAppCatalogListingsResponse wrapper for the ListAppCatalogListings operation
type ListAppCatalogListingsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []AppCatalogListingSummary instances
	Items []AppCatalogListingSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListAppCatalogListingsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListAppCatalogListingsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListAppCatalogListingsSortOrderEnum Enum with underlying type: string
type ListAppCatalogListingsSortOrderEnum string

// Set of constants representing the allowable values for ListAppCatalogListingsSortOrderEnum
const (
	ListAppCatalogListingsSortOrderAsc  ListAppCatalogListingsSortOrderEnum = "ASC"
	ListAppCatalogListingsSortOrderDesc ListAppCatalogListingsSortOrderEnum = "DESC"
)

var mappingListAppCatalogListingsSortOrderEnum = map[string]ListAppCatalogListingsSortOrderEnum{
	"ASC":  ListAppCatalogListingsSortOrderAsc,
	"DESC": ListAppCatalogListingsSortOrderDesc,
}

var mappingListAppCatalogListingsSortOrderEnumLowerCase = map[string]ListAppCatalogListingsSortOrderEnum{
	"asc":  ListAppCatalogListingsSortOrderAsc,
	"desc": ListAppCatalogListingsSortOrderDesc,
}

// GetListAppCatalogListingsSortOrderEnumValues Enumerates the set of values for ListAppCatalogListingsSortOrderEnum
func GetListAppCatalogListingsSortOrderEnumValues() []ListAppCatalogListingsSortOrderEnum {
	values := make([]ListAppCatalogListingsSortOrderEnum, 0)
	for _, v := range mappingListAppCatalogListingsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListAppCatalogListingsSortOrderEnumStringValues Enumerates the set of values in String for ListAppCatalogListingsSortOrderEnum
func GetListAppCatalogListingsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListAppCatalogListingsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListAppCatalogListingsSortOrderEnum(val string) (ListAppCatalogListingsSortOrderEnum, bool) {
	enum, ok := mappingListAppCatalogListingsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
