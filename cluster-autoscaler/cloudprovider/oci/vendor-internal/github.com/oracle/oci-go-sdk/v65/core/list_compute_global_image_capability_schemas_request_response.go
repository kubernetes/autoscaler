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

// ListComputeGlobalImageCapabilitySchemasRequest wrapper for the ListComputeGlobalImageCapabilitySchemas operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeGlobalImageCapabilitySchemas.go.html to see an example of how to use ListComputeGlobalImageCapabilitySchemasRequest.
type ListComputeGlobalImageCapabilitySchemasRequest struct {

	// A filter to return only resources that match the given compartment OCID exactly.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

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
	SortBy ListComputeGlobalImageCapabilitySchemasSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeGlobalImageCapabilitySchemasSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeGlobalImageCapabilitySchemasRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeGlobalImageCapabilitySchemasRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeGlobalImageCapabilitySchemasRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeGlobalImageCapabilitySchemasRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeGlobalImageCapabilitySchemasRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListComputeGlobalImageCapabilitySchemasSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeGlobalImageCapabilitySchemasSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeGlobalImageCapabilitySchemasSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeGlobalImageCapabilitySchemasSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeGlobalImageCapabilitySchemasResponse wrapper for the ListComputeGlobalImageCapabilitySchemas operation
type ListComputeGlobalImageCapabilitySchemasResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ComputeGlobalImageCapabilitySchemaSummary instances
	Items []ComputeGlobalImageCapabilitySchemaSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeGlobalImageCapabilitySchemasResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeGlobalImageCapabilitySchemasResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeGlobalImageCapabilitySchemasSortByEnum Enum with underlying type: string
type ListComputeGlobalImageCapabilitySchemasSortByEnum string

// Set of constants representing the allowable values for ListComputeGlobalImageCapabilitySchemasSortByEnum
const (
	ListComputeGlobalImageCapabilitySchemasSortByTimecreated ListComputeGlobalImageCapabilitySchemasSortByEnum = "TIMECREATED"
	ListComputeGlobalImageCapabilitySchemasSortByDisplayname ListComputeGlobalImageCapabilitySchemasSortByEnum = "DISPLAYNAME"
)

var mappingListComputeGlobalImageCapabilitySchemasSortByEnum = map[string]ListComputeGlobalImageCapabilitySchemasSortByEnum{
	"TIMECREATED": ListComputeGlobalImageCapabilitySchemasSortByTimecreated,
	"DISPLAYNAME": ListComputeGlobalImageCapabilitySchemasSortByDisplayname,
}

var mappingListComputeGlobalImageCapabilitySchemasSortByEnumLowerCase = map[string]ListComputeGlobalImageCapabilitySchemasSortByEnum{
	"timecreated": ListComputeGlobalImageCapabilitySchemasSortByTimecreated,
	"displayname": ListComputeGlobalImageCapabilitySchemasSortByDisplayname,
}

// GetListComputeGlobalImageCapabilitySchemasSortByEnumValues Enumerates the set of values for ListComputeGlobalImageCapabilitySchemasSortByEnum
func GetListComputeGlobalImageCapabilitySchemasSortByEnumValues() []ListComputeGlobalImageCapabilitySchemasSortByEnum {
	values := make([]ListComputeGlobalImageCapabilitySchemasSortByEnum, 0)
	for _, v := range mappingListComputeGlobalImageCapabilitySchemasSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeGlobalImageCapabilitySchemasSortByEnumStringValues Enumerates the set of values in String for ListComputeGlobalImageCapabilitySchemasSortByEnum
func GetListComputeGlobalImageCapabilitySchemasSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeGlobalImageCapabilitySchemasSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeGlobalImageCapabilitySchemasSortByEnum(val string) (ListComputeGlobalImageCapabilitySchemasSortByEnum, bool) {
	enum, ok := mappingListComputeGlobalImageCapabilitySchemasSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeGlobalImageCapabilitySchemasSortOrderEnum Enum with underlying type: string
type ListComputeGlobalImageCapabilitySchemasSortOrderEnum string

// Set of constants representing the allowable values for ListComputeGlobalImageCapabilitySchemasSortOrderEnum
const (
	ListComputeGlobalImageCapabilitySchemasSortOrderAsc  ListComputeGlobalImageCapabilitySchemasSortOrderEnum = "ASC"
	ListComputeGlobalImageCapabilitySchemasSortOrderDesc ListComputeGlobalImageCapabilitySchemasSortOrderEnum = "DESC"
)

var mappingListComputeGlobalImageCapabilitySchemasSortOrderEnum = map[string]ListComputeGlobalImageCapabilitySchemasSortOrderEnum{
	"ASC":  ListComputeGlobalImageCapabilitySchemasSortOrderAsc,
	"DESC": ListComputeGlobalImageCapabilitySchemasSortOrderDesc,
}

var mappingListComputeGlobalImageCapabilitySchemasSortOrderEnumLowerCase = map[string]ListComputeGlobalImageCapabilitySchemasSortOrderEnum{
	"asc":  ListComputeGlobalImageCapabilitySchemasSortOrderAsc,
	"desc": ListComputeGlobalImageCapabilitySchemasSortOrderDesc,
}

// GetListComputeGlobalImageCapabilitySchemasSortOrderEnumValues Enumerates the set of values for ListComputeGlobalImageCapabilitySchemasSortOrderEnum
func GetListComputeGlobalImageCapabilitySchemasSortOrderEnumValues() []ListComputeGlobalImageCapabilitySchemasSortOrderEnum {
	values := make([]ListComputeGlobalImageCapabilitySchemasSortOrderEnum, 0)
	for _, v := range mappingListComputeGlobalImageCapabilitySchemasSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeGlobalImageCapabilitySchemasSortOrderEnumStringValues Enumerates the set of values in String for ListComputeGlobalImageCapabilitySchemasSortOrderEnum
func GetListComputeGlobalImageCapabilitySchemasSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeGlobalImageCapabilitySchemasSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeGlobalImageCapabilitySchemasSortOrderEnum(val string) (ListComputeGlobalImageCapabilitySchemasSortOrderEnum, bool) {
	enum, ok := mappingListComputeGlobalImageCapabilitySchemasSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
