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

// ListClusterNetworksRequest wrapper for the ListClusterNetworks operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListClusterNetworks.go.html to see an example of how to use ListClusterNetworksRequest.
type ListClusterNetworksRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

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
	SortBy ListClusterNetworksSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListClusterNetworksSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle
	// state. The state value is case-insensitive.
	LifecycleState ClusterNetworkSummaryLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListClusterNetworksRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListClusterNetworksRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListClusterNetworksRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListClusterNetworksRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListClusterNetworksRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListClusterNetworksSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListClusterNetworksSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListClusterNetworksSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListClusterNetworksSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingClusterNetworkSummaryLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetClusterNetworkSummaryLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListClusterNetworksResponse wrapper for the ListClusterNetworks operation
type ListClusterNetworksResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ClusterNetworkSummary instances
	Items []ClusterNetworkSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListClusterNetworksResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListClusterNetworksResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListClusterNetworksSortByEnum Enum with underlying type: string
type ListClusterNetworksSortByEnum string

// Set of constants representing the allowable values for ListClusterNetworksSortByEnum
const (
	ListClusterNetworksSortByTimecreated ListClusterNetworksSortByEnum = "TIMECREATED"
	ListClusterNetworksSortByDisplayname ListClusterNetworksSortByEnum = "DISPLAYNAME"
)

var mappingListClusterNetworksSortByEnum = map[string]ListClusterNetworksSortByEnum{
	"TIMECREATED": ListClusterNetworksSortByTimecreated,
	"DISPLAYNAME": ListClusterNetworksSortByDisplayname,
}

var mappingListClusterNetworksSortByEnumLowerCase = map[string]ListClusterNetworksSortByEnum{
	"timecreated": ListClusterNetworksSortByTimecreated,
	"displayname": ListClusterNetworksSortByDisplayname,
}

// GetListClusterNetworksSortByEnumValues Enumerates the set of values for ListClusterNetworksSortByEnum
func GetListClusterNetworksSortByEnumValues() []ListClusterNetworksSortByEnum {
	values := make([]ListClusterNetworksSortByEnum, 0)
	for _, v := range mappingListClusterNetworksSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListClusterNetworksSortByEnumStringValues Enumerates the set of values in String for ListClusterNetworksSortByEnum
func GetListClusterNetworksSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListClusterNetworksSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListClusterNetworksSortByEnum(val string) (ListClusterNetworksSortByEnum, bool) {
	enum, ok := mappingListClusterNetworksSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListClusterNetworksSortOrderEnum Enum with underlying type: string
type ListClusterNetworksSortOrderEnum string

// Set of constants representing the allowable values for ListClusterNetworksSortOrderEnum
const (
	ListClusterNetworksSortOrderAsc  ListClusterNetworksSortOrderEnum = "ASC"
	ListClusterNetworksSortOrderDesc ListClusterNetworksSortOrderEnum = "DESC"
)

var mappingListClusterNetworksSortOrderEnum = map[string]ListClusterNetworksSortOrderEnum{
	"ASC":  ListClusterNetworksSortOrderAsc,
	"DESC": ListClusterNetworksSortOrderDesc,
}

var mappingListClusterNetworksSortOrderEnumLowerCase = map[string]ListClusterNetworksSortOrderEnum{
	"asc":  ListClusterNetworksSortOrderAsc,
	"desc": ListClusterNetworksSortOrderDesc,
}

// GetListClusterNetworksSortOrderEnumValues Enumerates the set of values for ListClusterNetworksSortOrderEnum
func GetListClusterNetworksSortOrderEnumValues() []ListClusterNetworksSortOrderEnum {
	values := make([]ListClusterNetworksSortOrderEnum, 0)
	for _, v := range mappingListClusterNetworksSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListClusterNetworksSortOrderEnumStringValues Enumerates the set of values in String for ListClusterNetworksSortOrderEnum
func GetListClusterNetworksSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListClusterNetworksSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListClusterNetworksSortOrderEnum(val string) (ListClusterNetworksSortOrderEnum, bool) {
	enum, ok := mappingListClusterNetworksSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
