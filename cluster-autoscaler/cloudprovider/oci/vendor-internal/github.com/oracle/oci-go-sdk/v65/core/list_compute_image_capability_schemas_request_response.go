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

// ListComputeImageCapabilitySchemasRequest wrapper for the ListComputeImageCapabilitySchemas operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeImageCapabilitySchemas.go.html to see an example of how to use ListComputeImageCapabilitySchemasRequest.
type ListComputeImageCapabilitySchemasRequest struct {

	// A filter to return only resources that match the given compartment OCID exactly.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of an image.
	ImageId *string `mandatory:"false" contributesTo:"query" name:"imageId"`

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
	SortBy ListComputeImageCapabilitySchemasSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeImageCapabilitySchemasSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeImageCapabilitySchemasRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeImageCapabilitySchemasRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeImageCapabilitySchemasRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeImageCapabilitySchemasRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeImageCapabilitySchemasRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListComputeImageCapabilitySchemasSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeImageCapabilitySchemasSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeImageCapabilitySchemasSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeImageCapabilitySchemasSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeImageCapabilitySchemasResponse wrapper for the ListComputeImageCapabilitySchemas operation
type ListComputeImageCapabilitySchemasResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ComputeImageCapabilitySchemaSummary instances
	Items []ComputeImageCapabilitySchemaSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeImageCapabilitySchemasResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeImageCapabilitySchemasResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeImageCapabilitySchemasSortByEnum Enum with underlying type: string
type ListComputeImageCapabilitySchemasSortByEnum string

// Set of constants representing the allowable values for ListComputeImageCapabilitySchemasSortByEnum
const (
	ListComputeImageCapabilitySchemasSortByTimecreated ListComputeImageCapabilitySchemasSortByEnum = "TIMECREATED"
	ListComputeImageCapabilitySchemasSortByDisplayname ListComputeImageCapabilitySchemasSortByEnum = "DISPLAYNAME"
)

var mappingListComputeImageCapabilitySchemasSortByEnum = map[string]ListComputeImageCapabilitySchemasSortByEnum{
	"TIMECREATED": ListComputeImageCapabilitySchemasSortByTimecreated,
	"DISPLAYNAME": ListComputeImageCapabilitySchemasSortByDisplayname,
}

var mappingListComputeImageCapabilitySchemasSortByEnumLowerCase = map[string]ListComputeImageCapabilitySchemasSortByEnum{
	"timecreated": ListComputeImageCapabilitySchemasSortByTimecreated,
	"displayname": ListComputeImageCapabilitySchemasSortByDisplayname,
}

// GetListComputeImageCapabilitySchemasSortByEnumValues Enumerates the set of values for ListComputeImageCapabilitySchemasSortByEnum
func GetListComputeImageCapabilitySchemasSortByEnumValues() []ListComputeImageCapabilitySchemasSortByEnum {
	values := make([]ListComputeImageCapabilitySchemasSortByEnum, 0)
	for _, v := range mappingListComputeImageCapabilitySchemasSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeImageCapabilitySchemasSortByEnumStringValues Enumerates the set of values in String for ListComputeImageCapabilitySchemasSortByEnum
func GetListComputeImageCapabilitySchemasSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeImageCapabilitySchemasSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeImageCapabilitySchemasSortByEnum(val string) (ListComputeImageCapabilitySchemasSortByEnum, bool) {
	enum, ok := mappingListComputeImageCapabilitySchemasSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeImageCapabilitySchemasSortOrderEnum Enum with underlying type: string
type ListComputeImageCapabilitySchemasSortOrderEnum string

// Set of constants representing the allowable values for ListComputeImageCapabilitySchemasSortOrderEnum
const (
	ListComputeImageCapabilitySchemasSortOrderAsc  ListComputeImageCapabilitySchemasSortOrderEnum = "ASC"
	ListComputeImageCapabilitySchemasSortOrderDesc ListComputeImageCapabilitySchemasSortOrderEnum = "DESC"
)

var mappingListComputeImageCapabilitySchemasSortOrderEnum = map[string]ListComputeImageCapabilitySchemasSortOrderEnum{
	"ASC":  ListComputeImageCapabilitySchemasSortOrderAsc,
	"DESC": ListComputeImageCapabilitySchemasSortOrderDesc,
}

var mappingListComputeImageCapabilitySchemasSortOrderEnumLowerCase = map[string]ListComputeImageCapabilitySchemasSortOrderEnum{
	"asc":  ListComputeImageCapabilitySchemasSortOrderAsc,
	"desc": ListComputeImageCapabilitySchemasSortOrderDesc,
}

// GetListComputeImageCapabilitySchemasSortOrderEnumValues Enumerates the set of values for ListComputeImageCapabilitySchemasSortOrderEnum
func GetListComputeImageCapabilitySchemasSortOrderEnumValues() []ListComputeImageCapabilitySchemasSortOrderEnum {
	values := make([]ListComputeImageCapabilitySchemasSortOrderEnum, 0)
	for _, v := range mappingListComputeImageCapabilitySchemasSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeImageCapabilitySchemasSortOrderEnumStringValues Enumerates the set of values in String for ListComputeImageCapabilitySchemasSortOrderEnum
func GetListComputeImageCapabilitySchemasSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeImageCapabilitySchemasSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeImageCapabilitySchemasSortOrderEnum(val string) (ListComputeImageCapabilitySchemasSortOrderEnum, bool) {
	enum, ok := mappingListComputeImageCapabilitySchemasSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
