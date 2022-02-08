// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
	"net/http"
)

// ListComputeCapacityReservationInstanceShapesRequest wrapper for the ListComputeCapacityReservationInstanceShapes operation
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeCapacityReservationInstanceShapes.go.html to see an example of how to use ListComputeCapacityReservationInstanceShapesRequest.
type ListComputeCapacityReservationInstanceShapesRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

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

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListComputeCapacityReservationInstanceShapesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeCapacityReservationInstanceShapesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeCapacityReservationInstanceShapesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeCapacityReservationInstanceShapesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser) (http.Request, error) {

	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeCapacityReservationInstanceShapesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeCapacityReservationInstanceShapesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListComputeCapacityReservationInstanceShapesResponse wrapper for the ListComputeCapacityReservationInstanceShapes operation
type ListComputeCapacityReservationInstanceShapesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ComputeCapacityReservationInstanceShapeSummary instances
	Items []ComputeCapacityReservationInstanceShapeSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeCapacityReservationInstanceShapesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeCapacityReservationInstanceShapesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeCapacityReservationInstanceShapesSortByEnum Enum with underlying type: string
type ListComputeCapacityReservationInstanceShapesSortByEnum string

// Set of constants representing the allowable values for ListComputeCapacityReservationInstanceShapesSortByEnum
const (
	ListComputeCapacityReservationInstanceShapesSortByTimecreated ListComputeCapacityReservationInstanceShapesSortByEnum = "TIMECREATED"
	ListComputeCapacityReservationInstanceShapesSortByDisplayname ListComputeCapacityReservationInstanceShapesSortByEnum = "DISPLAYNAME"
)

var mappingListComputeCapacityReservationInstanceShapesSortBy = map[string]ListComputeCapacityReservationInstanceShapesSortByEnum{
	"TIMECREATED": ListComputeCapacityReservationInstanceShapesSortByTimecreated,
	"DISPLAYNAME": ListComputeCapacityReservationInstanceShapesSortByDisplayname,
}

// GetListComputeCapacityReservationInstanceShapesSortByEnumValues Enumerates the set of values for ListComputeCapacityReservationInstanceShapesSortByEnum
func GetListComputeCapacityReservationInstanceShapesSortByEnumValues() []ListComputeCapacityReservationInstanceShapesSortByEnum {
	values := make([]ListComputeCapacityReservationInstanceShapesSortByEnum, 0)
	for _, v := range mappingListComputeCapacityReservationInstanceShapesSortBy {
		values = append(values, v)
	}
	return values
}

// ListComputeCapacityReservationInstanceShapesSortOrderEnum Enum with underlying type: string
type ListComputeCapacityReservationInstanceShapesSortOrderEnum string

// Set of constants representing the allowable values for ListComputeCapacityReservationInstanceShapesSortOrderEnum
const (
	ListComputeCapacityReservationInstanceShapesSortOrderAsc  ListComputeCapacityReservationInstanceShapesSortOrderEnum = "ASC"
	ListComputeCapacityReservationInstanceShapesSortOrderDesc ListComputeCapacityReservationInstanceShapesSortOrderEnum = "DESC"
)

var mappingListComputeCapacityReservationInstanceShapesSortOrder = map[string]ListComputeCapacityReservationInstanceShapesSortOrderEnum{
	"ASC":  ListComputeCapacityReservationInstanceShapesSortOrderAsc,
	"DESC": ListComputeCapacityReservationInstanceShapesSortOrderDesc,
}

// GetListComputeCapacityReservationInstanceShapesSortOrderEnumValues Enumerates the set of values for ListComputeCapacityReservationInstanceShapesSortOrderEnum
func GetListComputeCapacityReservationInstanceShapesSortOrderEnumValues() []ListComputeCapacityReservationInstanceShapesSortOrderEnum {
	values := make([]ListComputeCapacityReservationInstanceShapesSortOrderEnum, 0)
	for _, v := range mappingListComputeCapacityReservationInstanceShapesSortOrder {
		values = append(values, v)
	}
	return values
}
