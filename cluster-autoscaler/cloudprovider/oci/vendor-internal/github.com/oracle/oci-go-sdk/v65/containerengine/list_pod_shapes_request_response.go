// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package containerengine

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"net/http"
	"strings"
)

// ListPodShapesRequest wrapper for the ListPodShapes operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListPodShapes.go.html to see an example of how to use ListPodShapesRequest.
type ListPodShapesRequest struct {

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The availability domain of the pod shape.
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// The name to filter on.
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated "List" call.
	// 1 is the minimum, 1000 is the maximum. For important details about how pagination works,
	// see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List" call.
	// For important details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The optional order in which to sort the results.
	SortOrder ListPodShapesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the results by.
	SortBy ListPodShapesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListPodShapesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListPodShapesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListPodShapesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListPodShapesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListPodShapesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListPodShapesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListPodShapesSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListPodShapesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListPodShapesSortByEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListPodShapesResponse wrapper for the ListPodShapes operation
type ListPodShapesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []PodShapeSummary instances
	Items []PodShapeSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of results remain.
	// For important details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListPodShapesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListPodShapesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListPodShapesSortOrderEnum Enum with underlying type: string
type ListPodShapesSortOrderEnum string

// Set of constants representing the allowable values for ListPodShapesSortOrderEnum
const (
	ListPodShapesSortOrderAsc  ListPodShapesSortOrderEnum = "ASC"
	ListPodShapesSortOrderDesc ListPodShapesSortOrderEnum = "DESC"
)

var mappingListPodShapesSortOrderEnum = map[string]ListPodShapesSortOrderEnum{
	"ASC":  ListPodShapesSortOrderAsc,
	"DESC": ListPodShapesSortOrderDesc,
}

var mappingListPodShapesSortOrderEnumLowerCase = map[string]ListPodShapesSortOrderEnum{
	"asc":  ListPodShapesSortOrderAsc,
	"desc": ListPodShapesSortOrderDesc,
}

// GetListPodShapesSortOrderEnumValues Enumerates the set of values for ListPodShapesSortOrderEnum
func GetListPodShapesSortOrderEnumValues() []ListPodShapesSortOrderEnum {
	values := make([]ListPodShapesSortOrderEnum, 0)
	for _, v := range mappingListPodShapesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListPodShapesSortOrderEnumStringValues Enumerates the set of values in String for ListPodShapesSortOrderEnum
func GetListPodShapesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListPodShapesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListPodShapesSortOrderEnum(val string) (ListPodShapesSortOrderEnum, bool) {
	enum, ok := mappingListPodShapesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListPodShapesSortByEnum Enum with underlying type: string
type ListPodShapesSortByEnum string

// Set of constants representing the allowable values for ListPodShapesSortByEnum
const (
	ListPodShapesSortById          ListPodShapesSortByEnum = "ID"
	ListPodShapesSortByName        ListPodShapesSortByEnum = "NAME"
	ListPodShapesSortByTimeCreated ListPodShapesSortByEnum = "TIME_CREATED"
)

var mappingListPodShapesSortByEnum = map[string]ListPodShapesSortByEnum{
	"ID":           ListPodShapesSortById,
	"NAME":         ListPodShapesSortByName,
	"TIME_CREATED": ListPodShapesSortByTimeCreated,
}

var mappingListPodShapesSortByEnumLowerCase = map[string]ListPodShapesSortByEnum{
	"id":           ListPodShapesSortById,
	"name":         ListPodShapesSortByName,
	"time_created": ListPodShapesSortByTimeCreated,
}

// GetListPodShapesSortByEnumValues Enumerates the set of values for ListPodShapesSortByEnum
func GetListPodShapesSortByEnumValues() []ListPodShapesSortByEnum {
	values := make([]ListPodShapesSortByEnum, 0)
	for _, v := range mappingListPodShapesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListPodShapesSortByEnumStringValues Enumerates the set of values in String for ListPodShapesSortByEnum
func GetListPodShapesSortByEnumStringValues() []string {
	return []string{
		"ID",
		"NAME",
		"TIME_CREATED",
	}
}

// GetMappingListPodShapesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListPodShapesSortByEnum(val string) (ListPodShapesSortByEnum, bool) {
	enum, ok := mappingListPodShapesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
