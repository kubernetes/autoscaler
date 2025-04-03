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

// ListInternetGatewaysRequest wrapper for the ListInternetGateways operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListInternetGateways.go.html to see an example of how to use ListInternetGatewaysRequest.
type ListInternetGatewaysRequest struct {

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
	SortBy ListInternetGatewaysSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListInternetGatewaysSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle
	// state. The state value is case-insensitive.
	LifecycleState InternetGatewayLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListInternetGatewaysRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListInternetGatewaysRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListInternetGatewaysRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListInternetGatewaysRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListInternetGatewaysRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListInternetGatewaysSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListInternetGatewaysSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListInternetGatewaysSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListInternetGatewaysSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingInternetGatewayLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetInternetGatewayLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListInternetGatewaysResponse wrapper for the ListInternetGateways operation
type ListInternetGatewaysResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []InternetGateway instances
	Items []InternetGateway `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListInternetGatewaysResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListInternetGatewaysResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListInternetGatewaysSortByEnum Enum with underlying type: string
type ListInternetGatewaysSortByEnum string

// Set of constants representing the allowable values for ListInternetGatewaysSortByEnum
const (
	ListInternetGatewaysSortByTimecreated ListInternetGatewaysSortByEnum = "TIMECREATED"
	ListInternetGatewaysSortByDisplayname ListInternetGatewaysSortByEnum = "DISPLAYNAME"
)

var mappingListInternetGatewaysSortByEnum = map[string]ListInternetGatewaysSortByEnum{
	"TIMECREATED": ListInternetGatewaysSortByTimecreated,
	"DISPLAYNAME": ListInternetGatewaysSortByDisplayname,
}

var mappingListInternetGatewaysSortByEnumLowerCase = map[string]ListInternetGatewaysSortByEnum{
	"timecreated": ListInternetGatewaysSortByTimecreated,
	"displayname": ListInternetGatewaysSortByDisplayname,
}

// GetListInternetGatewaysSortByEnumValues Enumerates the set of values for ListInternetGatewaysSortByEnum
func GetListInternetGatewaysSortByEnumValues() []ListInternetGatewaysSortByEnum {
	values := make([]ListInternetGatewaysSortByEnum, 0)
	for _, v := range mappingListInternetGatewaysSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListInternetGatewaysSortByEnumStringValues Enumerates the set of values in String for ListInternetGatewaysSortByEnum
func GetListInternetGatewaysSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListInternetGatewaysSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListInternetGatewaysSortByEnum(val string) (ListInternetGatewaysSortByEnum, bool) {
	enum, ok := mappingListInternetGatewaysSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListInternetGatewaysSortOrderEnum Enum with underlying type: string
type ListInternetGatewaysSortOrderEnum string

// Set of constants representing the allowable values for ListInternetGatewaysSortOrderEnum
const (
	ListInternetGatewaysSortOrderAsc  ListInternetGatewaysSortOrderEnum = "ASC"
	ListInternetGatewaysSortOrderDesc ListInternetGatewaysSortOrderEnum = "DESC"
)

var mappingListInternetGatewaysSortOrderEnum = map[string]ListInternetGatewaysSortOrderEnum{
	"ASC":  ListInternetGatewaysSortOrderAsc,
	"DESC": ListInternetGatewaysSortOrderDesc,
}

var mappingListInternetGatewaysSortOrderEnumLowerCase = map[string]ListInternetGatewaysSortOrderEnum{
	"asc":  ListInternetGatewaysSortOrderAsc,
	"desc": ListInternetGatewaysSortOrderDesc,
}

// GetListInternetGatewaysSortOrderEnumValues Enumerates the set of values for ListInternetGatewaysSortOrderEnum
func GetListInternetGatewaysSortOrderEnumValues() []ListInternetGatewaysSortOrderEnum {
	values := make([]ListInternetGatewaysSortOrderEnum, 0)
	for _, v := range mappingListInternetGatewaysSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListInternetGatewaysSortOrderEnumStringValues Enumerates the set of values in String for ListInternetGatewaysSortOrderEnum
func GetListInternetGatewaysSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListInternetGatewaysSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListInternetGatewaysSortOrderEnum(val string) (ListInternetGatewaysSortOrderEnum, bool) {
	enum, ok := mappingListInternetGatewaysSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
