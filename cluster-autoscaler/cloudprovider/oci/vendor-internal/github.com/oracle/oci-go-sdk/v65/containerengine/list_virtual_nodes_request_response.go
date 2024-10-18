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

// ListVirtualNodesRequest wrapper for the ListVirtualNodes operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListVirtualNodes.go.html to see an example of how to use ListVirtualNodesRequest.
type ListVirtualNodesRequest struct {

	// The OCID of the virtual node pool.
	VirtualNodePoolId *string `mandatory:"true" contributesTo:"path" name:"virtualNodePoolId"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

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
	SortOrder ListVirtualNodesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the results by.
	SortBy ListVirtualNodesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListVirtualNodesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListVirtualNodesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListVirtualNodesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListVirtualNodesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListVirtualNodesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListVirtualNodesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListVirtualNodesSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListVirtualNodesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListVirtualNodesSortByEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListVirtualNodesResponse wrapper for the ListVirtualNodes operation
type ListVirtualNodesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []VirtualNodeSummary instances
	Items []VirtualNodeSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of results remain.
	// For important details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListVirtualNodesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListVirtualNodesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListVirtualNodesSortOrderEnum Enum with underlying type: string
type ListVirtualNodesSortOrderEnum string

// Set of constants representing the allowable values for ListVirtualNodesSortOrderEnum
const (
	ListVirtualNodesSortOrderAsc  ListVirtualNodesSortOrderEnum = "ASC"
	ListVirtualNodesSortOrderDesc ListVirtualNodesSortOrderEnum = "DESC"
)

var mappingListVirtualNodesSortOrderEnum = map[string]ListVirtualNodesSortOrderEnum{
	"ASC":  ListVirtualNodesSortOrderAsc,
	"DESC": ListVirtualNodesSortOrderDesc,
}

var mappingListVirtualNodesSortOrderEnumLowerCase = map[string]ListVirtualNodesSortOrderEnum{
	"asc":  ListVirtualNodesSortOrderAsc,
	"desc": ListVirtualNodesSortOrderDesc,
}

// GetListVirtualNodesSortOrderEnumValues Enumerates the set of values for ListVirtualNodesSortOrderEnum
func GetListVirtualNodesSortOrderEnumValues() []ListVirtualNodesSortOrderEnum {
	values := make([]ListVirtualNodesSortOrderEnum, 0)
	for _, v := range mappingListVirtualNodesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListVirtualNodesSortOrderEnumStringValues Enumerates the set of values in String for ListVirtualNodesSortOrderEnum
func GetListVirtualNodesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListVirtualNodesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVirtualNodesSortOrderEnum(val string) (ListVirtualNodesSortOrderEnum, bool) {
	enum, ok := mappingListVirtualNodesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListVirtualNodesSortByEnum Enum with underlying type: string
type ListVirtualNodesSortByEnum string

// Set of constants representing the allowable values for ListVirtualNodesSortByEnum
const (
	ListVirtualNodesSortById          ListVirtualNodesSortByEnum = "ID"
	ListVirtualNodesSortByName        ListVirtualNodesSortByEnum = "NAME"
	ListVirtualNodesSortByTimeCreated ListVirtualNodesSortByEnum = "TIME_CREATED"
)

var mappingListVirtualNodesSortByEnum = map[string]ListVirtualNodesSortByEnum{
	"ID":           ListVirtualNodesSortById,
	"NAME":         ListVirtualNodesSortByName,
	"TIME_CREATED": ListVirtualNodesSortByTimeCreated,
}

var mappingListVirtualNodesSortByEnumLowerCase = map[string]ListVirtualNodesSortByEnum{
	"id":           ListVirtualNodesSortById,
	"name":         ListVirtualNodesSortByName,
	"time_created": ListVirtualNodesSortByTimeCreated,
}

// GetListVirtualNodesSortByEnumValues Enumerates the set of values for ListVirtualNodesSortByEnum
func GetListVirtualNodesSortByEnumValues() []ListVirtualNodesSortByEnum {
	values := make([]ListVirtualNodesSortByEnum, 0)
	for _, v := range mappingListVirtualNodesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListVirtualNodesSortByEnumStringValues Enumerates the set of values in String for ListVirtualNodesSortByEnum
func GetListVirtualNodesSortByEnumStringValues() []string {
	return []string{
		"ID",
		"NAME",
		"TIME_CREATED",
	}
}

// GetMappingListVirtualNodesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVirtualNodesSortByEnum(val string) (ListVirtualNodesSortByEnum, bool) {
	enum, ok := mappingListVirtualNodesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
