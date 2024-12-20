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

// ListWorkRequestsRequest wrapper for the ListWorkRequests operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListWorkRequests.go.html to see an example of how to use ListWorkRequestsRequest.
type ListWorkRequestsRequest struct {

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID of the cluster.
	ClusterId *string `mandatory:"false" contributesTo:"query" name:"clusterId"`

	// The OCID of the resource associated with a work request
	ResourceId *string `mandatory:"false" contributesTo:"query" name:"resourceId"`

	// Type of the resource associated with a work request
	ResourceType ListWorkRequestsResourceTypeEnum `mandatory:"false" contributesTo:"query" name:"resourceType" omitEmpty:"true"`

	// A work request status to filter on. Can have multiple parameters of this name.
	Status []string `contributesTo:"query" name:"status" collectionFormat:"multi"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated "List" call.
	// 1 is the minimum, 1000 is the maximum. For important details about how pagination works,
	// see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List" call.
	// For important details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The optional order in which to sort the results.
	SortOrder ListWorkRequestsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the results by.
	SortBy ListWorkRequestsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListWorkRequestsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListWorkRequestsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListWorkRequestsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListWorkRequestsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListWorkRequestsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListWorkRequestsResourceTypeEnum(string(request.ResourceType)); !ok && request.ResourceType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ResourceType: %s. Supported values are: %s.", request.ResourceType, strings.Join(GetListWorkRequestsResourceTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListWorkRequestsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListWorkRequestsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListWorkRequestsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListWorkRequestsSortByEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListWorkRequestsResponse wrapper for the ListWorkRequests operation
type ListWorkRequestsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []WorkRequestSummary instances
	Items []WorkRequestSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of results remain.
	// For important details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListWorkRequestsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListWorkRequestsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListWorkRequestsResourceTypeEnum Enum with underlying type: string
type ListWorkRequestsResourceTypeEnum string

// Set of constants representing the allowable values for ListWorkRequestsResourceTypeEnum
const (
	ListWorkRequestsResourceTypeCluster  ListWorkRequestsResourceTypeEnum = "CLUSTER"
	ListWorkRequestsResourceTypeNodepool ListWorkRequestsResourceTypeEnum = "NODEPOOL"
)

var mappingListWorkRequestsResourceTypeEnum = map[string]ListWorkRequestsResourceTypeEnum{
	"CLUSTER":  ListWorkRequestsResourceTypeCluster,
	"NODEPOOL": ListWorkRequestsResourceTypeNodepool,
}

var mappingListWorkRequestsResourceTypeEnumLowerCase = map[string]ListWorkRequestsResourceTypeEnum{
	"cluster":  ListWorkRequestsResourceTypeCluster,
	"nodepool": ListWorkRequestsResourceTypeNodepool,
}

// GetListWorkRequestsResourceTypeEnumValues Enumerates the set of values for ListWorkRequestsResourceTypeEnum
func GetListWorkRequestsResourceTypeEnumValues() []ListWorkRequestsResourceTypeEnum {
	values := make([]ListWorkRequestsResourceTypeEnum, 0)
	for _, v := range mappingListWorkRequestsResourceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetListWorkRequestsResourceTypeEnumStringValues Enumerates the set of values in String for ListWorkRequestsResourceTypeEnum
func GetListWorkRequestsResourceTypeEnumStringValues() []string {
	return []string{
		"CLUSTER",
		"NODEPOOL",
	}
}

// GetMappingListWorkRequestsResourceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListWorkRequestsResourceTypeEnum(val string) (ListWorkRequestsResourceTypeEnum, bool) {
	enum, ok := mappingListWorkRequestsResourceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListWorkRequestsSortOrderEnum Enum with underlying type: string
type ListWorkRequestsSortOrderEnum string

// Set of constants representing the allowable values for ListWorkRequestsSortOrderEnum
const (
	ListWorkRequestsSortOrderAsc  ListWorkRequestsSortOrderEnum = "ASC"
	ListWorkRequestsSortOrderDesc ListWorkRequestsSortOrderEnum = "DESC"
)

var mappingListWorkRequestsSortOrderEnum = map[string]ListWorkRequestsSortOrderEnum{
	"ASC":  ListWorkRequestsSortOrderAsc,
	"DESC": ListWorkRequestsSortOrderDesc,
}

var mappingListWorkRequestsSortOrderEnumLowerCase = map[string]ListWorkRequestsSortOrderEnum{
	"asc":  ListWorkRequestsSortOrderAsc,
	"desc": ListWorkRequestsSortOrderDesc,
}

// GetListWorkRequestsSortOrderEnumValues Enumerates the set of values for ListWorkRequestsSortOrderEnum
func GetListWorkRequestsSortOrderEnumValues() []ListWorkRequestsSortOrderEnum {
	values := make([]ListWorkRequestsSortOrderEnum, 0)
	for _, v := range mappingListWorkRequestsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListWorkRequestsSortOrderEnumStringValues Enumerates the set of values in String for ListWorkRequestsSortOrderEnum
func GetListWorkRequestsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListWorkRequestsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListWorkRequestsSortOrderEnum(val string) (ListWorkRequestsSortOrderEnum, bool) {
	enum, ok := mappingListWorkRequestsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListWorkRequestsSortByEnum Enum with underlying type: string
type ListWorkRequestsSortByEnum string

// Set of constants representing the allowable values for ListWorkRequestsSortByEnum
const (
	ListWorkRequestsSortById            ListWorkRequestsSortByEnum = "ID"
	ListWorkRequestsSortByOperationType ListWorkRequestsSortByEnum = "OPERATION_TYPE"
	ListWorkRequestsSortByStatus        ListWorkRequestsSortByEnum = "STATUS"
	ListWorkRequestsSortByTimeAccepted  ListWorkRequestsSortByEnum = "TIME_ACCEPTED"
	ListWorkRequestsSortByTimeStarted   ListWorkRequestsSortByEnum = "TIME_STARTED"
	ListWorkRequestsSortByTimeFinished  ListWorkRequestsSortByEnum = "TIME_FINISHED"
)

var mappingListWorkRequestsSortByEnum = map[string]ListWorkRequestsSortByEnum{
	"ID":             ListWorkRequestsSortById,
	"OPERATION_TYPE": ListWorkRequestsSortByOperationType,
	"STATUS":         ListWorkRequestsSortByStatus,
	"TIME_ACCEPTED":  ListWorkRequestsSortByTimeAccepted,
	"TIME_STARTED":   ListWorkRequestsSortByTimeStarted,
	"TIME_FINISHED":  ListWorkRequestsSortByTimeFinished,
}

var mappingListWorkRequestsSortByEnumLowerCase = map[string]ListWorkRequestsSortByEnum{
	"id":             ListWorkRequestsSortById,
	"operation_type": ListWorkRequestsSortByOperationType,
	"status":         ListWorkRequestsSortByStatus,
	"time_accepted":  ListWorkRequestsSortByTimeAccepted,
	"time_started":   ListWorkRequestsSortByTimeStarted,
	"time_finished":  ListWorkRequestsSortByTimeFinished,
}

// GetListWorkRequestsSortByEnumValues Enumerates the set of values for ListWorkRequestsSortByEnum
func GetListWorkRequestsSortByEnumValues() []ListWorkRequestsSortByEnum {
	values := make([]ListWorkRequestsSortByEnum, 0)
	for _, v := range mappingListWorkRequestsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListWorkRequestsSortByEnumStringValues Enumerates the set of values in String for ListWorkRequestsSortByEnum
func GetListWorkRequestsSortByEnumStringValues() []string {
	return []string{
		"ID",
		"OPERATION_TYPE",
		"STATUS",
		"TIME_ACCEPTED",
		"TIME_STARTED",
		"TIME_FINISHED",
	}
}

// GetMappingListWorkRequestsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListWorkRequestsSortByEnum(val string) (ListWorkRequestsSortByEnum, bool) {
	enum, ok := mappingListWorkRequestsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
