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

// ListVirtualNodePoolsRequest wrapper for the ListVirtualNodePools operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListVirtualNodePools.go.html to see an example of how to use ListVirtualNodePoolsRequest.
type ListVirtualNodePoolsRequest struct {

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The OCID of the cluster.
	ClusterId *string `mandatory:"false" contributesTo:"query" name:"clusterId"`

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
	SortOrder ListVirtualNodePoolsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the results by.
	SortBy ListVirtualNodePoolsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// A virtual node pool lifecycle state to filter on. Can have multiple parameters of this name.
	LifecycleState []VirtualNodePoolLifecycleStateEnum `contributesTo:"query" name:"lifecycleState" omitEmpty:"true" collectionFormat:"multi"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListVirtualNodePoolsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListVirtualNodePoolsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListVirtualNodePoolsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListVirtualNodePoolsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListVirtualNodePoolsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListVirtualNodePoolsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListVirtualNodePoolsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListVirtualNodePoolsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListVirtualNodePoolsSortByEnumStringValues(), ",")))
	}
	for _, val := range request.LifecycleState {
		if _, ok := GetMappingVirtualNodePoolLifecycleStateEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", val, strings.Join(GetVirtualNodePoolLifecycleStateEnumStringValues(), ",")))
		}
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListVirtualNodePoolsResponse wrapper for the ListVirtualNodePools operation
type ListVirtualNodePoolsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []VirtualNodePoolSummary instances
	Items []VirtualNodePoolSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of results remain.
	// For important details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListVirtualNodePoolsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListVirtualNodePoolsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListVirtualNodePoolsSortOrderEnum Enum with underlying type: string
type ListVirtualNodePoolsSortOrderEnum string

// Set of constants representing the allowable values for ListVirtualNodePoolsSortOrderEnum
const (
	ListVirtualNodePoolsSortOrderAsc  ListVirtualNodePoolsSortOrderEnum = "ASC"
	ListVirtualNodePoolsSortOrderDesc ListVirtualNodePoolsSortOrderEnum = "DESC"
)

var mappingListVirtualNodePoolsSortOrderEnum = map[string]ListVirtualNodePoolsSortOrderEnum{
	"ASC":  ListVirtualNodePoolsSortOrderAsc,
	"DESC": ListVirtualNodePoolsSortOrderDesc,
}

var mappingListVirtualNodePoolsSortOrderEnumLowerCase = map[string]ListVirtualNodePoolsSortOrderEnum{
	"asc":  ListVirtualNodePoolsSortOrderAsc,
	"desc": ListVirtualNodePoolsSortOrderDesc,
}

// GetListVirtualNodePoolsSortOrderEnumValues Enumerates the set of values for ListVirtualNodePoolsSortOrderEnum
func GetListVirtualNodePoolsSortOrderEnumValues() []ListVirtualNodePoolsSortOrderEnum {
	values := make([]ListVirtualNodePoolsSortOrderEnum, 0)
	for _, v := range mappingListVirtualNodePoolsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListVirtualNodePoolsSortOrderEnumStringValues Enumerates the set of values in String for ListVirtualNodePoolsSortOrderEnum
func GetListVirtualNodePoolsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListVirtualNodePoolsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVirtualNodePoolsSortOrderEnum(val string) (ListVirtualNodePoolsSortOrderEnum, bool) {
	enum, ok := mappingListVirtualNodePoolsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListVirtualNodePoolsSortByEnum Enum with underlying type: string
type ListVirtualNodePoolsSortByEnum string

// Set of constants representing the allowable values for ListVirtualNodePoolsSortByEnum
const (
	ListVirtualNodePoolsSortById          ListVirtualNodePoolsSortByEnum = "ID"
	ListVirtualNodePoolsSortByName        ListVirtualNodePoolsSortByEnum = "NAME"
	ListVirtualNodePoolsSortByTimeCreated ListVirtualNodePoolsSortByEnum = "TIME_CREATED"
)

var mappingListVirtualNodePoolsSortByEnum = map[string]ListVirtualNodePoolsSortByEnum{
	"ID":           ListVirtualNodePoolsSortById,
	"NAME":         ListVirtualNodePoolsSortByName,
	"TIME_CREATED": ListVirtualNodePoolsSortByTimeCreated,
}

var mappingListVirtualNodePoolsSortByEnumLowerCase = map[string]ListVirtualNodePoolsSortByEnum{
	"id":           ListVirtualNodePoolsSortById,
	"name":         ListVirtualNodePoolsSortByName,
	"time_created": ListVirtualNodePoolsSortByTimeCreated,
}

// GetListVirtualNodePoolsSortByEnumValues Enumerates the set of values for ListVirtualNodePoolsSortByEnum
func GetListVirtualNodePoolsSortByEnumValues() []ListVirtualNodePoolsSortByEnum {
	values := make([]ListVirtualNodePoolsSortByEnum, 0)
	for _, v := range mappingListVirtualNodePoolsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListVirtualNodePoolsSortByEnumStringValues Enumerates the set of values in String for ListVirtualNodePoolsSortByEnum
func GetListVirtualNodePoolsSortByEnumStringValues() []string {
	return []string{
		"ID",
		"NAME",
		"TIME_CREATED",
	}
}

// GetMappingListVirtualNodePoolsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVirtualNodePoolsSortByEnum(val string) (ListVirtualNodePoolsSortByEnum, bool) {
	enum, ok := mappingListVirtualNodePoolsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
