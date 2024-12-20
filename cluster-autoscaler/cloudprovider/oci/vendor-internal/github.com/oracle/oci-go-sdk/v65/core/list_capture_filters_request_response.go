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

// ListCaptureFiltersRequest wrapper for the ListCaptureFilters operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListCaptureFilters.go.html to see an example of how to use ListCaptureFiltersRequest.
type ListCaptureFiltersRequest struct {

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
	SortBy ListCaptureFiltersSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListCaptureFiltersSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to return only resources that match the given capture filter lifecycle state.
	// The state value is case-insensitive.
	LifecycleState CaptureFilterLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// A filter to only return resources that match the given capture `filterType`. The `filterType` value is the string representation of enum - `VTAP`, `FLOWLOG`.
	FilterType CaptureFilterFilterTypeEnum `mandatory:"false" contributesTo:"query" name:"filterType" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListCaptureFiltersRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListCaptureFiltersRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListCaptureFiltersRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListCaptureFiltersRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListCaptureFiltersRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListCaptureFiltersSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListCaptureFiltersSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListCaptureFiltersSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListCaptureFiltersSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCaptureFilterLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetCaptureFilterLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCaptureFilterFilterTypeEnum(string(request.FilterType)); !ok && request.FilterType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for FilterType: %s. Supported values are: %s.", request.FilterType, strings.Join(GetCaptureFilterFilterTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListCaptureFiltersResponse wrapper for the ListCaptureFilters operation
type ListCaptureFiltersResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []CaptureFilter instances
	Items []CaptureFilter `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListCaptureFiltersResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListCaptureFiltersResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListCaptureFiltersSortByEnum Enum with underlying type: string
type ListCaptureFiltersSortByEnum string

// Set of constants representing the allowable values for ListCaptureFiltersSortByEnum
const (
	ListCaptureFiltersSortByTimecreated ListCaptureFiltersSortByEnum = "TIMECREATED"
	ListCaptureFiltersSortByDisplayname ListCaptureFiltersSortByEnum = "DISPLAYNAME"
)

var mappingListCaptureFiltersSortByEnum = map[string]ListCaptureFiltersSortByEnum{
	"TIMECREATED": ListCaptureFiltersSortByTimecreated,
	"DISPLAYNAME": ListCaptureFiltersSortByDisplayname,
}

var mappingListCaptureFiltersSortByEnumLowerCase = map[string]ListCaptureFiltersSortByEnum{
	"timecreated": ListCaptureFiltersSortByTimecreated,
	"displayname": ListCaptureFiltersSortByDisplayname,
}

// GetListCaptureFiltersSortByEnumValues Enumerates the set of values for ListCaptureFiltersSortByEnum
func GetListCaptureFiltersSortByEnumValues() []ListCaptureFiltersSortByEnum {
	values := make([]ListCaptureFiltersSortByEnum, 0)
	for _, v := range mappingListCaptureFiltersSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListCaptureFiltersSortByEnumStringValues Enumerates the set of values in String for ListCaptureFiltersSortByEnum
func GetListCaptureFiltersSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListCaptureFiltersSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListCaptureFiltersSortByEnum(val string) (ListCaptureFiltersSortByEnum, bool) {
	enum, ok := mappingListCaptureFiltersSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListCaptureFiltersSortOrderEnum Enum with underlying type: string
type ListCaptureFiltersSortOrderEnum string

// Set of constants representing the allowable values for ListCaptureFiltersSortOrderEnum
const (
	ListCaptureFiltersSortOrderAsc  ListCaptureFiltersSortOrderEnum = "ASC"
	ListCaptureFiltersSortOrderDesc ListCaptureFiltersSortOrderEnum = "DESC"
)

var mappingListCaptureFiltersSortOrderEnum = map[string]ListCaptureFiltersSortOrderEnum{
	"ASC":  ListCaptureFiltersSortOrderAsc,
	"DESC": ListCaptureFiltersSortOrderDesc,
}

var mappingListCaptureFiltersSortOrderEnumLowerCase = map[string]ListCaptureFiltersSortOrderEnum{
	"asc":  ListCaptureFiltersSortOrderAsc,
	"desc": ListCaptureFiltersSortOrderDesc,
}

// GetListCaptureFiltersSortOrderEnumValues Enumerates the set of values for ListCaptureFiltersSortOrderEnum
func GetListCaptureFiltersSortOrderEnumValues() []ListCaptureFiltersSortOrderEnum {
	values := make([]ListCaptureFiltersSortOrderEnum, 0)
	for _, v := range mappingListCaptureFiltersSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListCaptureFiltersSortOrderEnumStringValues Enumerates the set of values in String for ListCaptureFiltersSortOrderEnum
func GetListCaptureFiltersSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListCaptureFiltersSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListCaptureFiltersSortOrderEnum(val string) (ListCaptureFiltersSortOrderEnum, bool) {
	enum, ok := mappingListCaptureFiltersSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
