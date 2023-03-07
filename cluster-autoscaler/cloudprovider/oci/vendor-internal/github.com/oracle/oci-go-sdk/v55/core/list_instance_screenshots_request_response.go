// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"net/http"
	"strings"
)

// ListInstanceScreenshotsRequest wrapper for the ListInstanceScreenshots operation
type ListInstanceScreenshotsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the instance.
	InstanceId *string `mandatory:"true" contributesTo:"path" name:"instanceId"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

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
	SortBy ListInstanceScreenshotsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListInstanceScreenshotsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle state. The state
	// value is case-insensitive.
	LifecycleState InstanceScreenshotLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListInstanceScreenshotsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListInstanceScreenshotsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListInstanceScreenshotsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListInstanceScreenshotsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListInstanceScreenshotsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingListInstanceScreenshotsSortByEnum[string(request.SortBy)]; !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListInstanceScreenshotsSortByEnumStringValues(), ",")))
	}
	if _, ok := mappingListInstanceScreenshotsSortOrderEnum[string(request.SortOrder)]; !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListInstanceScreenshotsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := mappingInstanceScreenshotLifecycleStateEnum[string(request.LifecycleState)]; !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetInstanceScreenshotLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListInstanceScreenshotsResponse wrapper for the ListInstanceScreenshots operation
type ListInstanceScreenshotsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []InstanceScreenshotSummary instances
	Items []InstanceScreenshotSummary `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListInstanceScreenshotsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListInstanceScreenshotsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListInstanceScreenshotsSortByEnum Enum with underlying type: string
type ListInstanceScreenshotsSortByEnum string

// Set of constants representing the allowable values for ListInstanceScreenshotsSortByEnum
const (
	ListInstanceScreenshotsSortByTimecreated ListInstanceScreenshotsSortByEnum = "TIMECREATED"
	ListInstanceScreenshotsSortByDisplayname ListInstanceScreenshotsSortByEnum = "DISPLAYNAME"
)

var mappingListInstanceScreenshotsSortByEnum = map[string]ListInstanceScreenshotsSortByEnum{
	"TIMECREATED": ListInstanceScreenshotsSortByTimecreated,
	"DISPLAYNAME": ListInstanceScreenshotsSortByDisplayname,
}

// GetListInstanceScreenshotsSortByEnumValues Enumerates the set of values for ListInstanceScreenshotsSortByEnum
func GetListInstanceScreenshotsSortByEnumValues() []ListInstanceScreenshotsSortByEnum {
	values := make([]ListInstanceScreenshotsSortByEnum, 0)
	for _, v := range mappingListInstanceScreenshotsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListInstanceScreenshotsSortByEnumStringValues Enumerates the set of values in String for ListInstanceScreenshotsSortByEnum
func GetListInstanceScreenshotsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// ListInstanceScreenshotsSortOrderEnum Enum with underlying type: string
type ListInstanceScreenshotsSortOrderEnum string

// Set of constants representing the allowable values for ListInstanceScreenshotsSortOrderEnum
const (
	ListInstanceScreenshotsSortOrderAsc  ListInstanceScreenshotsSortOrderEnum = "ASC"
	ListInstanceScreenshotsSortOrderDesc ListInstanceScreenshotsSortOrderEnum = "DESC"
)

var mappingListInstanceScreenshotsSortOrderEnum = map[string]ListInstanceScreenshotsSortOrderEnum{
	"ASC":  ListInstanceScreenshotsSortOrderAsc,
	"DESC": ListInstanceScreenshotsSortOrderDesc,
}

// GetListInstanceScreenshotsSortOrderEnumValues Enumerates the set of values for ListInstanceScreenshotsSortOrderEnum
func GetListInstanceScreenshotsSortOrderEnumValues() []ListInstanceScreenshotsSortOrderEnum {
	values := make([]ListInstanceScreenshotsSortOrderEnum, 0)
	for _, v := range mappingListInstanceScreenshotsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListInstanceScreenshotsSortOrderEnumStringValues Enumerates the set of values in String for ListInstanceScreenshotsSortOrderEnum
func GetListInstanceScreenshotsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}
