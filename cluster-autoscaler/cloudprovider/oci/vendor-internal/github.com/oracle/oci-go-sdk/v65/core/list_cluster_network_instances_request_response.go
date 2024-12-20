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

// ListClusterNetworkInstancesRequest wrapper for the ListClusterNetworkInstances operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListClusterNetworkInstances.go.html to see an example of how to use ListClusterNetworkInstancesRequest.
type ListClusterNetworkInstancesRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the cluster network.
	ClusterNetworkId *string `mandatory:"true" contributesTo:"path" name:"clusterNetworkId"`

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
	SortBy ListClusterNetworkInstancesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListClusterNetworkInstancesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListClusterNetworkInstancesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListClusterNetworkInstancesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListClusterNetworkInstancesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListClusterNetworkInstancesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListClusterNetworkInstancesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListClusterNetworkInstancesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListClusterNetworkInstancesSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListClusterNetworkInstancesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListClusterNetworkInstancesSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListClusterNetworkInstancesResponse wrapper for the ListClusterNetworkInstances operation
type ListClusterNetworkInstancesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []InstanceSummary instances
	Items []InstanceSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListClusterNetworkInstancesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListClusterNetworkInstancesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListClusterNetworkInstancesSortByEnum Enum with underlying type: string
type ListClusterNetworkInstancesSortByEnum string

// Set of constants representing the allowable values for ListClusterNetworkInstancesSortByEnum
const (
	ListClusterNetworkInstancesSortByTimecreated ListClusterNetworkInstancesSortByEnum = "TIMECREATED"
	ListClusterNetworkInstancesSortByDisplayname ListClusterNetworkInstancesSortByEnum = "DISPLAYNAME"
)

var mappingListClusterNetworkInstancesSortByEnum = map[string]ListClusterNetworkInstancesSortByEnum{
	"TIMECREATED": ListClusterNetworkInstancesSortByTimecreated,
	"DISPLAYNAME": ListClusterNetworkInstancesSortByDisplayname,
}

var mappingListClusterNetworkInstancesSortByEnumLowerCase = map[string]ListClusterNetworkInstancesSortByEnum{
	"timecreated": ListClusterNetworkInstancesSortByTimecreated,
	"displayname": ListClusterNetworkInstancesSortByDisplayname,
}

// GetListClusterNetworkInstancesSortByEnumValues Enumerates the set of values for ListClusterNetworkInstancesSortByEnum
func GetListClusterNetworkInstancesSortByEnumValues() []ListClusterNetworkInstancesSortByEnum {
	values := make([]ListClusterNetworkInstancesSortByEnum, 0)
	for _, v := range mappingListClusterNetworkInstancesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListClusterNetworkInstancesSortByEnumStringValues Enumerates the set of values in String for ListClusterNetworkInstancesSortByEnum
func GetListClusterNetworkInstancesSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListClusterNetworkInstancesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListClusterNetworkInstancesSortByEnum(val string) (ListClusterNetworkInstancesSortByEnum, bool) {
	enum, ok := mappingListClusterNetworkInstancesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListClusterNetworkInstancesSortOrderEnum Enum with underlying type: string
type ListClusterNetworkInstancesSortOrderEnum string

// Set of constants representing the allowable values for ListClusterNetworkInstancesSortOrderEnum
const (
	ListClusterNetworkInstancesSortOrderAsc  ListClusterNetworkInstancesSortOrderEnum = "ASC"
	ListClusterNetworkInstancesSortOrderDesc ListClusterNetworkInstancesSortOrderEnum = "DESC"
)

var mappingListClusterNetworkInstancesSortOrderEnum = map[string]ListClusterNetworkInstancesSortOrderEnum{
	"ASC":  ListClusterNetworkInstancesSortOrderAsc,
	"DESC": ListClusterNetworkInstancesSortOrderDesc,
}

var mappingListClusterNetworkInstancesSortOrderEnumLowerCase = map[string]ListClusterNetworkInstancesSortOrderEnum{
	"asc":  ListClusterNetworkInstancesSortOrderAsc,
	"desc": ListClusterNetworkInstancesSortOrderDesc,
}

// GetListClusterNetworkInstancesSortOrderEnumValues Enumerates the set of values for ListClusterNetworkInstancesSortOrderEnum
func GetListClusterNetworkInstancesSortOrderEnumValues() []ListClusterNetworkInstancesSortOrderEnum {
	values := make([]ListClusterNetworkInstancesSortOrderEnum, 0)
	for _, v := range mappingListClusterNetworkInstancesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListClusterNetworkInstancesSortOrderEnumStringValues Enumerates the set of values in String for ListClusterNetworkInstancesSortOrderEnum
func GetListClusterNetworkInstancesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListClusterNetworkInstancesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListClusterNetworkInstancesSortOrderEnum(val string) (ListClusterNetworkInstancesSortOrderEnum, bool) {
	enum, ok := mappingListClusterNetworkInstancesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
