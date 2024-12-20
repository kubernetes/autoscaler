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

// ListDhcpOptionsRequest wrapper for the ListDhcpOptions operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListDhcpOptions.go.html to see an example of how to use ListDhcpOptionsRequest.
type ListDhcpOptionsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN.
	VcnId *string `mandatory:"false" contributesTo:"query" name:"vcnId"`

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
	SortBy ListDhcpOptionsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListDhcpOptionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle
	// state. The state value is case-insensitive.
	LifecycleState DhcpOptionsLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListDhcpOptionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListDhcpOptionsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListDhcpOptionsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListDhcpOptionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListDhcpOptionsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListDhcpOptionsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListDhcpOptionsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListDhcpOptionsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListDhcpOptionsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingDhcpOptionsLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetDhcpOptionsLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListDhcpOptionsResponse wrapper for the ListDhcpOptions operation
type ListDhcpOptionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []DhcpOptions instances
	Items []DhcpOptions `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListDhcpOptionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListDhcpOptionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListDhcpOptionsSortByEnum Enum with underlying type: string
type ListDhcpOptionsSortByEnum string

// Set of constants representing the allowable values for ListDhcpOptionsSortByEnum
const (
	ListDhcpOptionsSortByTimecreated ListDhcpOptionsSortByEnum = "TIMECREATED"
	ListDhcpOptionsSortByDisplayname ListDhcpOptionsSortByEnum = "DISPLAYNAME"
)

var mappingListDhcpOptionsSortByEnum = map[string]ListDhcpOptionsSortByEnum{
	"TIMECREATED": ListDhcpOptionsSortByTimecreated,
	"DISPLAYNAME": ListDhcpOptionsSortByDisplayname,
}

var mappingListDhcpOptionsSortByEnumLowerCase = map[string]ListDhcpOptionsSortByEnum{
	"timecreated": ListDhcpOptionsSortByTimecreated,
	"displayname": ListDhcpOptionsSortByDisplayname,
}

// GetListDhcpOptionsSortByEnumValues Enumerates the set of values for ListDhcpOptionsSortByEnum
func GetListDhcpOptionsSortByEnumValues() []ListDhcpOptionsSortByEnum {
	values := make([]ListDhcpOptionsSortByEnum, 0)
	for _, v := range mappingListDhcpOptionsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListDhcpOptionsSortByEnumStringValues Enumerates the set of values in String for ListDhcpOptionsSortByEnum
func GetListDhcpOptionsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListDhcpOptionsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListDhcpOptionsSortByEnum(val string) (ListDhcpOptionsSortByEnum, bool) {
	enum, ok := mappingListDhcpOptionsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListDhcpOptionsSortOrderEnum Enum with underlying type: string
type ListDhcpOptionsSortOrderEnum string

// Set of constants representing the allowable values for ListDhcpOptionsSortOrderEnum
const (
	ListDhcpOptionsSortOrderAsc  ListDhcpOptionsSortOrderEnum = "ASC"
	ListDhcpOptionsSortOrderDesc ListDhcpOptionsSortOrderEnum = "DESC"
)

var mappingListDhcpOptionsSortOrderEnum = map[string]ListDhcpOptionsSortOrderEnum{
	"ASC":  ListDhcpOptionsSortOrderAsc,
	"DESC": ListDhcpOptionsSortOrderDesc,
}

var mappingListDhcpOptionsSortOrderEnumLowerCase = map[string]ListDhcpOptionsSortOrderEnum{
	"asc":  ListDhcpOptionsSortOrderAsc,
	"desc": ListDhcpOptionsSortOrderDesc,
}

// GetListDhcpOptionsSortOrderEnumValues Enumerates the set of values for ListDhcpOptionsSortOrderEnum
func GetListDhcpOptionsSortOrderEnumValues() []ListDhcpOptionsSortOrderEnum {
	values := make([]ListDhcpOptionsSortOrderEnum, 0)
	for _, v := range mappingListDhcpOptionsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListDhcpOptionsSortOrderEnumStringValues Enumerates the set of values in String for ListDhcpOptionsSortOrderEnum
func GetListDhcpOptionsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListDhcpOptionsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListDhcpOptionsSortOrderEnum(val string) (ListDhcpOptionsSortOrderEnum, bool) {
	enum, ok := mappingListDhcpOptionsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
