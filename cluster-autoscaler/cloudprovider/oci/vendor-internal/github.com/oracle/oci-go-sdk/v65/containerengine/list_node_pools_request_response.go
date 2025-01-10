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

// ListNodePoolsRequest wrapper for the ListNodePools operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListNodePools.go.html to see an example of how to use ListNodePoolsRequest.
type ListNodePoolsRequest struct {

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

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
	SortOrder ListNodePoolsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the results by.
	SortBy ListNodePoolsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// A list of nodepool lifecycle states on which to filter on, matching any of the list items (OR logic). eg. ACTIVE, DELETING. For more information, see Monitoring Clusters (https://docs.cloud.oracle.com/Content/ContEng/Tasks/contengmonitoringclusters.htm)
	LifecycleState []NodePoolLifecycleStateEnum `contributesTo:"query" name:"lifecycleState" omitEmpty:"true" collectionFormat:"multi"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListNodePoolsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListNodePoolsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListNodePoolsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListNodePoolsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListNodePoolsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListNodePoolsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListNodePoolsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListNodePoolsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListNodePoolsSortByEnumStringValues(), ",")))
	}
	for _, val := range request.LifecycleState {
		if _, ok := GetMappingNodePoolLifecycleStateEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", val, strings.Join(GetNodePoolLifecycleStateEnumStringValues(), ",")))
		}
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListNodePoolsResponse wrapper for the ListNodePools operation
type ListNodePoolsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []NodePoolSummary instances
	Items []NodePoolSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of results remain.
	// For important details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListNodePoolsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListNodePoolsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListNodePoolsSortOrderEnum Enum with underlying type: string
type ListNodePoolsSortOrderEnum string

// Set of constants representing the allowable values for ListNodePoolsSortOrderEnum
const (
	ListNodePoolsSortOrderAsc  ListNodePoolsSortOrderEnum = "ASC"
	ListNodePoolsSortOrderDesc ListNodePoolsSortOrderEnum = "DESC"
)

var mappingListNodePoolsSortOrderEnum = map[string]ListNodePoolsSortOrderEnum{
	"ASC":  ListNodePoolsSortOrderAsc,
	"DESC": ListNodePoolsSortOrderDesc,
}

var mappingListNodePoolsSortOrderEnumLowerCase = map[string]ListNodePoolsSortOrderEnum{
	"asc":  ListNodePoolsSortOrderAsc,
	"desc": ListNodePoolsSortOrderDesc,
}

// GetListNodePoolsSortOrderEnumValues Enumerates the set of values for ListNodePoolsSortOrderEnum
func GetListNodePoolsSortOrderEnumValues() []ListNodePoolsSortOrderEnum {
	values := make([]ListNodePoolsSortOrderEnum, 0)
	for _, v := range mappingListNodePoolsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListNodePoolsSortOrderEnumStringValues Enumerates the set of values in String for ListNodePoolsSortOrderEnum
func GetListNodePoolsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListNodePoolsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListNodePoolsSortOrderEnum(val string) (ListNodePoolsSortOrderEnum, bool) {
	enum, ok := mappingListNodePoolsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListNodePoolsSortByEnum Enum with underlying type: string
type ListNodePoolsSortByEnum string

// Set of constants representing the allowable values for ListNodePoolsSortByEnum
const (
	ListNodePoolsSortById          ListNodePoolsSortByEnum = "ID"
	ListNodePoolsSortByName        ListNodePoolsSortByEnum = "NAME"
	ListNodePoolsSortByTimeCreated ListNodePoolsSortByEnum = "TIME_CREATED"
)

var mappingListNodePoolsSortByEnum = map[string]ListNodePoolsSortByEnum{
	"ID":           ListNodePoolsSortById,
	"NAME":         ListNodePoolsSortByName,
	"TIME_CREATED": ListNodePoolsSortByTimeCreated,
}

var mappingListNodePoolsSortByEnumLowerCase = map[string]ListNodePoolsSortByEnum{
	"id":           ListNodePoolsSortById,
	"name":         ListNodePoolsSortByName,
	"time_created": ListNodePoolsSortByTimeCreated,
}

// GetListNodePoolsSortByEnumValues Enumerates the set of values for ListNodePoolsSortByEnum
func GetListNodePoolsSortByEnumValues() []ListNodePoolsSortByEnum {
	values := make([]ListNodePoolsSortByEnum, 0)
	for _, v := range mappingListNodePoolsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListNodePoolsSortByEnumStringValues Enumerates the set of values in String for ListNodePoolsSortByEnum
func GetListNodePoolsSortByEnumStringValues() []string {
	return []string{
		"ID",
		"NAME",
		"TIME_CREATED",
	}
}

// GetMappingListNodePoolsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListNodePoolsSortByEnum(val string) (ListNodePoolsSortByEnum, bool) {
	enum, ok := mappingListNodePoolsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
