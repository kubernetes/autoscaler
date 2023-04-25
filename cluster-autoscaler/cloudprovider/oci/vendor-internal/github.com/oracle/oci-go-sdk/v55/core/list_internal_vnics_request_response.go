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

// ListInternalVnicsRequest wrapper for the ListInternalVnics operation
type ListInternalVnicsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the subnet.
	SubnetId *string `mandatory:"true" contributesTo:"query" name:"subnetId"`

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
	SortBy ListInternalVnicsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListInternalVnicsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// A filter to return only resources that match the specified lifecycle state.
	LifecycleState InternalVnicLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListInternalVnicsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListInternalVnicsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListInternalVnicsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListInternalVnicsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListInternalVnicsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingListInternalVnicsSortByEnum[string(request.SortBy)]; !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListInternalVnicsSortByEnumStringValues(), ",")))
	}
	if _, ok := mappingListInternalVnicsSortOrderEnum[string(request.SortOrder)]; !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListInternalVnicsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := mappingInternalVnicLifecycleStateEnum[string(request.LifecycleState)]; !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetInternalVnicLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListInternalVnicsResponse wrapper for the ListInternalVnics operation
type ListInternalVnicsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []InternalVnic instances
	Items []InternalVnic `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListInternalVnicsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListInternalVnicsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListInternalVnicsSortByEnum Enum with underlying type: string
type ListInternalVnicsSortByEnum string

// Set of constants representing the allowable values for ListInternalVnicsSortByEnum
const (
	ListInternalVnicsSortByTimecreated ListInternalVnicsSortByEnum = "TIMECREATED"
	ListInternalVnicsSortByDisplayname ListInternalVnicsSortByEnum = "DISPLAYNAME"
)

var mappingListInternalVnicsSortByEnum = map[string]ListInternalVnicsSortByEnum{
	"TIMECREATED": ListInternalVnicsSortByTimecreated,
	"DISPLAYNAME": ListInternalVnicsSortByDisplayname,
}

// GetListInternalVnicsSortByEnumValues Enumerates the set of values for ListInternalVnicsSortByEnum
func GetListInternalVnicsSortByEnumValues() []ListInternalVnicsSortByEnum {
	values := make([]ListInternalVnicsSortByEnum, 0)
	for _, v := range mappingListInternalVnicsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListInternalVnicsSortByEnumStringValues Enumerates the set of values in String for ListInternalVnicsSortByEnum
func GetListInternalVnicsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// ListInternalVnicsSortOrderEnum Enum with underlying type: string
type ListInternalVnicsSortOrderEnum string

// Set of constants representing the allowable values for ListInternalVnicsSortOrderEnum
const (
	ListInternalVnicsSortOrderAsc  ListInternalVnicsSortOrderEnum = "ASC"
	ListInternalVnicsSortOrderDesc ListInternalVnicsSortOrderEnum = "DESC"
)

var mappingListInternalVnicsSortOrderEnum = map[string]ListInternalVnicsSortOrderEnum{
	"ASC":  ListInternalVnicsSortOrderAsc,
	"DESC": ListInternalVnicsSortOrderDesc,
}

// GetListInternalVnicsSortOrderEnumValues Enumerates the set of values for ListInternalVnicsSortOrderEnum
func GetListInternalVnicsSortOrderEnumValues() []ListInternalVnicsSortOrderEnum {
	values := make([]ListInternalVnicsSortOrderEnum, 0)
	for _, v := range mappingListInternalVnicsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListInternalVnicsSortOrderEnumStringValues Enumerates the set of values in String for ListInternalVnicsSortOrderEnum
func GetListInternalVnicsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}
