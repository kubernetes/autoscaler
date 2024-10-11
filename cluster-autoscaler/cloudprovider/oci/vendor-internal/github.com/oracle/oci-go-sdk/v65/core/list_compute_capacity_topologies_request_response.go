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

// ListComputeCapacityTopologiesRequest wrapper for the ListComputeCapacityTopologies operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeCapacityTopologies.go.html to see an example of how to use ListComputeCapacityTopologiesRequest.
type ListComputeCapacityTopologiesRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

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

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListComputeCapacityTopologiesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeCapacityTopologiesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeCapacityTopologiesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeCapacityTopologiesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeCapacityTopologiesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeCapacityTopologiesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeCapacityTopologiesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListComputeCapacityTopologiesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeCapacityTopologiesSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeCapacityTopologiesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeCapacityTopologiesSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeCapacityTopologiesResponse wrapper for the ListComputeCapacityTopologies operation
type ListComputeCapacityTopologiesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ComputeCapacityTopologyCollection instances
	ComputeCapacityTopologyCollection `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeCapacityTopologiesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeCapacityTopologiesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeCapacityTopologiesSortByEnum Enum with underlying type: string
type ListComputeCapacityTopologiesSortByEnum string

// Set of constants representing the allowable values for ListComputeCapacityTopologiesSortByEnum
const (
	ListComputeCapacityTopologiesSortByTimecreated ListComputeCapacityTopologiesSortByEnum = "TIMECREATED"
	ListComputeCapacityTopologiesSortByDisplayname ListComputeCapacityTopologiesSortByEnum = "DISPLAYNAME"
)

var mappingListComputeCapacityTopologiesSortByEnum = map[string]ListComputeCapacityTopologiesSortByEnum{
	"TIMECREATED": ListComputeCapacityTopologiesSortByTimecreated,
	"DISPLAYNAME": ListComputeCapacityTopologiesSortByDisplayname,
}

var mappingListComputeCapacityTopologiesSortByEnumLowerCase = map[string]ListComputeCapacityTopologiesSortByEnum{
	"timecreated": ListComputeCapacityTopologiesSortByTimecreated,
	"displayname": ListComputeCapacityTopologiesSortByDisplayname,
}

// GetListComputeCapacityTopologiesSortByEnumValues Enumerates the set of values for ListComputeCapacityTopologiesSortByEnum
func GetListComputeCapacityTopologiesSortByEnumValues() []ListComputeCapacityTopologiesSortByEnum {
	values := make([]ListComputeCapacityTopologiesSortByEnum, 0)
	for _, v := range mappingListComputeCapacityTopologiesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeCapacityTopologiesSortByEnumStringValues Enumerates the set of values in String for ListComputeCapacityTopologiesSortByEnum
func GetListComputeCapacityTopologiesSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeCapacityTopologiesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeCapacityTopologiesSortByEnum(val string) (ListComputeCapacityTopologiesSortByEnum, bool) {
	enum, ok := mappingListComputeCapacityTopologiesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeCapacityTopologiesSortOrderEnum Enum with underlying type: string
type ListComputeCapacityTopologiesSortOrderEnum string

// Set of constants representing the allowable values for ListComputeCapacityTopologiesSortOrderEnum
const (
	ListComputeCapacityTopologiesSortOrderAsc  ListComputeCapacityTopologiesSortOrderEnum = "ASC"
	ListComputeCapacityTopologiesSortOrderDesc ListComputeCapacityTopologiesSortOrderEnum = "DESC"
)

var mappingListComputeCapacityTopologiesSortOrderEnum = map[string]ListComputeCapacityTopologiesSortOrderEnum{
	"ASC":  ListComputeCapacityTopologiesSortOrderAsc,
	"DESC": ListComputeCapacityTopologiesSortOrderDesc,
}

var mappingListComputeCapacityTopologiesSortOrderEnumLowerCase = map[string]ListComputeCapacityTopologiesSortOrderEnum{
	"asc":  ListComputeCapacityTopologiesSortOrderAsc,
	"desc": ListComputeCapacityTopologiesSortOrderDesc,
}

// GetListComputeCapacityTopologiesSortOrderEnumValues Enumerates the set of values for ListComputeCapacityTopologiesSortOrderEnum
func GetListComputeCapacityTopologiesSortOrderEnumValues() []ListComputeCapacityTopologiesSortOrderEnum {
	values := make([]ListComputeCapacityTopologiesSortOrderEnum, 0)
	for _, v := range mappingListComputeCapacityTopologiesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeCapacityTopologiesSortOrderEnumStringValues Enumerates the set of values in String for ListComputeCapacityTopologiesSortOrderEnum
func GetListComputeCapacityTopologiesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeCapacityTopologiesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeCapacityTopologiesSortOrderEnum(val string) (ListComputeCapacityTopologiesSortOrderEnum, bool) {
	enum, ok := mappingListComputeCapacityTopologiesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
