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

// ListVirtualCircuitsRequest wrapper for the ListVirtualCircuits operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListVirtualCircuits.go.html to see an example of how to use ListVirtualCircuitsRequest.
type ListVirtualCircuitsRequest struct {

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

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListVirtualCircuitsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListVirtualCircuitsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only resources that match the specified lifecycle
	// state. The value is case insensitive.
	LifecycleState VirtualCircuitLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListVirtualCircuitsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListVirtualCircuitsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListVirtualCircuitsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListVirtualCircuitsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListVirtualCircuitsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListVirtualCircuitsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListVirtualCircuitsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListVirtualCircuitsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListVirtualCircuitsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetVirtualCircuitLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListVirtualCircuitsResponse wrapper for the ListVirtualCircuits operation
type ListVirtualCircuitsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []VirtualCircuit instances
	Items []VirtualCircuit `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListVirtualCircuitsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListVirtualCircuitsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListVirtualCircuitsSortByEnum Enum with underlying type: string
type ListVirtualCircuitsSortByEnum string

// Set of constants representing the allowable values for ListVirtualCircuitsSortByEnum
const (
	ListVirtualCircuitsSortByTimecreated ListVirtualCircuitsSortByEnum = "TIMECREATED"
	ListVirtualCircuitsSortByDisplayname ListVirtualCircuitsSortByEnum = "DISPLAYNAME"
)

var mappingListVirtualCircuitsSortByEnum = map[string]ListVirtualCircuitsSortByEnum{
	"TIMECREATED": ListVirtualCircuitsSortByTimecreated,
	"DISPLAYNAME": ListVirtualCircuitsSortByDisplayname,
}

var mappingListVirtualCircuitsSortByEnumLowerCase = map[string]ListVirtualCircuitsSortByEnum{
	"timecreated": ListVirtualCircuitsSortByTimecreated,
	"displayname": ListVirtualCircuitsSortByDisplayname,
}

// GetListVirtualCircuitsSortByEnumValues Enumerates the set of values for ListVirtualCircuitsSortByEnum
func GetListVirtualCircuitsSortByEnumValues() []ListVirtualCircuitsSortByEnum {
	values := make([]ListVirtualCircuitsSortByEnum, 0)
	for _, v := range mappingListVirtualCircuitsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListVirtualCircuitsSortByEnumStringValues Enumerates the set of values in String for ListVirtualCircuitsSortByEnum
func GetListVirtualCircuitsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListVirtualCircuitsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVirtualCircuitsSortByEnum(val string) (ListVirtualCircuitsSortByEnum, bool) {
	enum, ok := mappingListVirtualCircuitsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListVirtualCircuitsSortOrderEnum Enum with underlying type: string
type ListVirtualCircuitsSortOrderEnum string

// Set of constants representing the allowable values for ListVirtualCircuitsSortOrderEnum
const (
	ListVirtualCircuitsSortOrderAsc  ListVirtualCircuitsSortOrderEnum = "ASC"
	ListVirtualCircuitsSortOrderDesc ListVirtualCircuitsSortOrderEnum = "DESC"
)

var mappingListVirtualCircuitsSortOrderEnum = map[string]ListVirtualCircuitsSortOrderEnum{
	"ASC":  ListVirtualCircuitsSortOrderAsc,
	"DESC": ListVirtualCircuitsSortOrderDesc,
}

var mappingListVirtualCircuitsSortOrderEnumLowerCase = map[string]ListVirtualCircuitsSortOrderEnum{
	"asc":  ListVirtualCircuitsSortOrderAsc,
	"desc": ListVirtualCircuitsSortOrderDesc,
}

// GetListVirtualCircuitsSortOrderEnumValues Enumerates the set of values for ListVirtualCircuitsSortOrderEnum
func GetListVirtualCircuitsSortOrderEnumValues() []ListVirtualCircuitsSortOrderEnum {
	values := make([]ListVirtualCircuitsSortOrderEnum, 0)
	for _, v := range mappingListVirtualCircuitsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListVirtualCircuitsSortOrderEnumStringValues Enumerates the set of values in String for ListVirtualCircuitsSortOrderEnum
func GetListVirtualCircuitsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListVirtualCircuitsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVirtualCircuitsSortOrderEnum(val string) (ListVirtualCircuitsSortOrderEnum, bool) {
	enum, ok := mappingListVirtualCircuitsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
