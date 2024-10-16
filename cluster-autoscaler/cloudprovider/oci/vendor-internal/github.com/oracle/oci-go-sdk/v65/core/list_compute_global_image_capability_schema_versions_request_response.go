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

// ListComputeGlobalImageCapabilitySchemaVersionsRequest wrapper for the ListComputeGlobalImageCapabilitySchemaVersions operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeGlobalImageCapabilitySchemaVersions.go.html to see an example of how to use ListComputeGlobalImageCapabilitySchemaVersionsRequest.
type ListComputeGlobalImageCapabilitySchemaVersionsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute global image capability schema
	ComputeGlobalImageCapabilitySchemaId *string `mandatory:"true" contributesTo:"path" name:"computeGlobalImageCapabilitySchemaId"`

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
	SortBy ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeGlobalImageCapabilitySchemaVersionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeGlobalImageCapabilitySchemaVersionsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeGlobalImageCapabilitySchemaVersionsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeGlobalImageCapabilitySchemaVersionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeGlobalImageCapabilitySchemaVersionsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListComputeGlobalImageCapabilitySchemaVersionsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeGlobalImageCapabilitySchemaVersionsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeGlobalImageCapabilitySchemaVersionsResponse wrapper for the ListComputeGlobalImageCapabilitySchemaVersions operation
type ListComputeGlobalImageCapabilitySchemaVersionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ComputeGlobalImageCapabilitySchemaVersionSummary instances
	Items []ComputeGlobalImageCapabilitySchemaVersionSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeGlobalImageCapabilitySchemaVersionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeGlobalImageCapabilitySchemaVersionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum Enum with underlying type: string
type ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum string

// Set of constants representing the allowable values for ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum
const (
	ListComputeGlobalImageCapabilitySchemaVersionsSortByTimecreated ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum = "TIMECREATED"
	ListComputeGlobalImageCapabilitySchemaVersionsSortByDisplayname ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum = "DISPLAYNAME"
)

var mappingListComputeGlobalImageCapabilitySchemaVersionsSortByEnum = map[string]ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum{
	"TIMECREATED": ListComputeGlobalImageCapabilitySchemaVersionsSortByTimecreated,
	"DISPLAYNAME": ListComputeGlobalImageCapabilitySchemaVersionsSortByDisplayname,
}

var mappingListComputeGlobalImageCapabilitySchemaVersionsSortByEnumLowerCase = map[string]ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum{
	"timecreated": ListComputeGlobalImageCapabilitySchemaVersionsSortByTimecreated,
	"displayname": ListComputeGlobalImageCapabilitySchemaVersionsSortByDisplayname,
}

// GetListComputeGlobalImageCapabilitySchemaVersionsSortByEnumValues Enumerates the set of values for ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum
func GetListComputeGlobalImageCapabilitySchemaVersionsSortByEnumValues() []ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum {
	values := make([]ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum, 0)
	for _, v := range mappingListComputeGlobalImageCapabilitySchemaVersionsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeGlobalImageCapabilitySchemaVersionsSortByEnumStringValues Enumerates the set of values in String for ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum
func GetListComputeGlobalImageCapabilitySchemaVersionsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeGlobalImageCapabilitySchemaVersionsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeGlobalImageCapabilitySchemaVersionsSortByEnum(val string) (ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum, bool) {
	enum, ok := mappingListComputeGlobalImageCapabilitySchemaVersionsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum Enum with underlying type: string
type ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum string

// Set of constants representing the allowable values for ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum
const (
	ListComputeGlobalImageCapabilitySchemaVersionsSortOrderAsc  ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum = "ASC"
	ListComputeGlobalImageCapabilitySchemaVersionsSortOrderDesc ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum = "DESC"
)

var mappingListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum = map[string]ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum{
	"ASC":  ListComputeGlobalImageCapabilitySchemaVersionsSortOrderAsc,
	"DESC": ListComputeGlobalImageCapabilitySchemaVersionsSortOrderDesc,
}

var mappingListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnumLowerCase = map[string]ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum{
	"asc":  ListComputeGlobalImageCapabilitySchemaVersionsSortOrderAsc,
	"desc": ListComputeGlobalImageCapabilitySchemaVersionsSortOrderDesc,
}

// GetListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnumValues Enumerates the set of values for ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum
func GetListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnumValues() []ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum {
	values := make([]ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum, 0)
	for _, v := range mappingListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnumStringValues Enumerates the set of values in String for ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum
func GetListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum(val string) (ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum, bool) {
	enum, ok := mappingListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
