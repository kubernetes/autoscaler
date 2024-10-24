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

// ListComputeCapacityReservationsRequest wrapper for the ListComputeCapacityReservations operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeCapacityReservations.go.html to see an example of how to use ListComputeCapacityReservationsRequest.
type ListComputeCapacityReservationsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// A filter to only return resources that match the given lifecycle state.
	LifecycleState ComputeCapacityReservationLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

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
	SortBy ListComputeCapacityReservationsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeCapacityReservationsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeCapacityReservationsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeCapacityReservationsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeCapacityReservationsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeCapacityReservationsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeCapacityReservationsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingComputeCapacityReservationLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetComputeCapacityReservationLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeCapacityReservationsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeCapacityReservationsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeCapacityReservationsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeCapacityReservationsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeCapacityReservationsResponse wrapper for the ListComputeCapacityReservations operation
type ListComputeCapacityReservationsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ComputeCapacityReservationSummary instances
	Items []ComputeCapacityReservationSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeCapacityReservationsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeCapacityReservationsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeCapacityReservationsSortByEnum Enum with underlying type: string
type ListComputeCapacityReservationsSortByEnum string

// Set of constants representing the allowable values for ListComputeCapacityReservationsSortByEnum
const (
	ListComputeCapacityReservationsSortByTimecreated ListComputeCapacityReservationsSortByEnum = "TIMECREATED"
	ListComputeCapacityReservationsSortByDisplayname ListComputeCapacityReservationsSortByEnum = "DISPLAYNAME"
)

var mappingListComputeCapacityReservationsSortByEnum = map[string]ListComputeCapacityReservationsSortByEnum{
	"TIMECREATED": ListComputeCapacityReservationsSortByTimecreated,
	"DISPLAYNAME": ListComputeCapacityReservationsSortByDisplayname,
}

var mappingListComputeCapacityReservationsSortByEnumLowerCase = map[string]ListComputeCapacityReservationsSortByEnum{
	"timecreated": ListComputeCapacityReservationsSortByTimecreated,
	"displayname": ListComputeCapacityReservationsSortByDisplayname,
}

// GetListComputeCapacityReservationsSortByEnumValues Enumerates the set of values for ListComputeCapacityReservationsSortByEnum
func GetListComputeCapacityReservationsSortByEnumValues() []ListComputeCapacityReservationsSortByEnum {
	values := make([]ListComputeCapacityReservationsSortByEnum, 0)
	for _, v := range mappingListComputeCapacityReservationsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeCapacityReservationsSortByEnumStringValues Enumerates the set of values in String for ListComputeCapacityReservationsSortByEnum
func GetListComputeCapacityReservationsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeCapacityReservationsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeCapacityReservationsSortByEnum(val string) (ListComputeCapacityReservationsSortByEnum, bool) {
	enum, ok := mappingListComputeCapacityReservationsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeCapacityReservationsSortOrderEnum Enum with underlying type: string
type ListComputeCapacityReservationsSortOrderEnum string

// Set of constants representing the allowable values for ListComputeCapacityReservationsSortOrderEnum
const (
	ListComputeCapacityReservationsSortOrderAsc  ListComputeCapacityReservationsSortOrderEnum = "ASC"
	ListComputeCapacityReservationsSortOrderDesc ListComputeCapacityReservationsSortOrderEnum = "DESC"
)

var mappingListComputeCapacityReservationsSortOrderEnum = map[string]ListComputeCapacityReservationsSortOrderEnum{
	"ASC":  ListComputeCapacityReservationsSortOrderAsc,
	"DESC": ListComputeCapacityReservationsSortOrderDesc,
}

var mappingListComputeCapacityReservationsSortOrderEnumLowerCase = map[string]ListComputeCapacityReservationsSortOrderEnum{
	"asc":  ListComputeCapacityReservationsSortOrderAsc,
	"desc": ListComputeCapacityReservationsSortOrderDesc,
}

// GetListComputeCapacityReservationsSortOrderEnumValues Enumerates the set of values for ListComputeCapacityReservationsSortOrderEnum
func GetListComputeCapacityReservationsSortOrderEnumValues() []ListComputeCapacityReservationsSortOrderEnum {
	values := make([]ListComputeCapacityReservationsSortOrderEnum, 0)
	for _, v := range mappingListComputeCapacityReservationsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeCapacityReservationsSortOrderEnumStringValues Enumerates the set of values in String for ListComputeCapacityReservationsSortOrderEnum
func GetListComputeCapacityReservationsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeCapacityReservationsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeCapacityReservationsSortOrderEnum(val string) (ListComputeCapacityReservationsSortOrderEnum, bool) {
	enum, ok := mappingListComputeCapacityReservationsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
