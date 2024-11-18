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

// ListServiceGatewaysRequest wrapper for the ListServiceGateways operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListServiceGateways.go.html to see an example of how to use ListServiceGatewaysRequest.
type ListServiceGatewaysRequest struct {

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

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListServiceGatewaysSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListServiceGatewaysSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only resources that match the given lifecycle
	// state. The state value is case-insensitive.
	LifecycleState ServiceGatewayLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListServiceGatewaysRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListServiceGatewaysRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListServiceGatewaysRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListServiceGatewaysRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListServiceGatewaysRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListServiceGatewaysSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListServiceGatewaysSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListServiceGatewaysSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListServiceGatewaysSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingServiceGatewayLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetServiceGatewayLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListServiceGatewaysResponse wrapper for the ListServiceGateways operation
type ListServiceGatewaysResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ServiceGateway instances
	Items []ServiceGateway `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListServiceGatewaysResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListServiceGatewaysResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListServiceGatewaysSortByEnum Enum with underlying type: string
type ListServiceGatewaysSortByEnum string

// Set of constants representing the allowable values for ListServiceGatewaysSortByEnum
const (
	ListServiceGatewaysSortByTimecreated ListServiceGatewaysSortByEnum = "TIMECREATED"
	ListServiceGatewaysSortByDisplayname ListServiceGatewaysSortByEnum = "DISPLAYNAME"
)

var mappingListServiceGatewaysSortByEnum = map[string]ListServiceGatewaysSortByEnum{
	"TIMECREATED": ListServiceGatewaysSortByTimecreated,
	"DISPLAYNAME": ListServiceGatewaysSortByDisplayname,
}

var mappingListServiceGatewaysSortByEnumLowerCase = map[string]ListServiceGatewaysSortByEnum{
	"timecreated": ListServiceGatewaysSortByTimecreated,
	"displayname": ListServiceGatewaysSortByDisplayname,
}

// GetListServiceGatewaysSortByEnumValues Enumerates the set of values for ListServiceGatewaysSortByEnum
func GetListServiceGatewaysSortByEnumValues() []ListServiceGatewaysSortByEnum {
	values := make([]ListServiceGatewaysSortByEnum, 0)
	for _, v := range mappingListServiceGatewaysSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListServiceGatewaysSortByEnumStringValues Enumerates the set of values in String for ListServiceGatewaysSortByEnum
func GetListServiceGatewaysSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListServiceGatewaysSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListServiceGatewaysSortByEnum(val string) (ListServiceGatewaysSortByEnum, bool) {
	enum, ok := mappingListServiceGatewaysSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListServiceGatewaysSortOrderEnum Enum with underlying type: string
type ListServiceGatewaysSortOrderEnum string

// Set of constants representing the allowable values for ListServiceGatewaysSortOrderEnum
const (
	ListServiceGatewaysSortOrderAsc  ListServiceGatewaysSortOrderEnum = "ASC"
	ListServiceGatewaysSortOrderDesc ListServiceGatewaysSortOrderEnum = "DESC"
)

var mappingListServiceGatewaysSortOrderEnum = map[string]ListServiceGatewaysSortOrderEnum{
	"ASC":  ListServiceGatewaysSortOrderAsc,
	"DESC": ListServiceGatewaysSortOrderDesc,
}

var mappingListServiceGatewaysSortOrderEnumLowerCase = map[string]ListServiceGatewaysSortOrderEnum{
	"asc":  ListServiceGatewaysSortOrderAsc,
	"desc": ListServiceGatewaysSortOrderDesc,
}

// GetListServiceGatewaysSortOrderEnumValues Enumerates the set of values for ListServiceGatewaysSortOrderEnum
func GetListServiceGatewaysSortOrderEnumValues() []ListServiceGatewaysSortOrderEnum {
	values := make([]ListServiceGatewaysSortOrderEnum, 0)
	for _, v := range mappingListServiceGatewaysSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListServiceGatewaysSortOrderEnumStringValues Enumerates the set of values in String for ListServiceGatewaysSortOrderEnum
func GetListServiceGatewaysSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListServiceGatewaysSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListServiceGatewaysSortOrderEnum(val string) (ListServiceGatewaysSortOrderEnum, bool) {
	enum, ok := mappingListServiceGatewaysSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
