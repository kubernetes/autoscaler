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

// ListInstanceMaintenanceEventsRequest wrapper for the ListInstanceMaintenanceEvents operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListInstanceMaintenanceEvents.go.html to see an example of how to use ListInstanceMaintenanceEventsRequest.
type ListInstanceMaintenanceEventsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID of the instance.
	InstanceId *string `mandatory:"false" contributesTo:"query" name:"instanceId"`

	// A filter to only return resources that match the given lifecycle state.
	LifecycleState InstanceMaintenanceEventLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// A filter to only return resources that have a matching correlationToken.
	CorrelationToken *string `mandatory:"false" contributesTo:"query" name:"correlationToken"`

	// A filter to only return resources that match the given instance action.
	InstanceAction *string `mandatory:"false" contributesTo:"query" name:"instanceAction"`

	// Starting range to return the maintenances which are not completed (date-time is in RFC3339 (https://tools.ietf.org/html/rfc3339) format).
	TimeWindowStartGreaterThanOrEqualTo *common.SDKTime `mandatory:"false" contributesTo:"query" name:"timeWindowStartGreaterThanOrEqualTo"`

	// Ending range to return the maintenances which are not completed (date-time is in RFC3339 (https://tools.ietf.org/html/rfc3339) format).
	TimeWindowStartLessThanOrEqualTo *common.SDKTime `mandatory:"false" contributesTo:"query" name:"timeWindowStartLessThanOrEqualTo"`

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
	SortBy ListInstanceMaintenanceEventsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListInstanceMaintenanceEventsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListInstanceMaintenanceEventsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListInstanceMaintenanceEventsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListInstanceMaintenanceEventsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListInstanceMaintenanceEventsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListInstanceMaintenanceEventsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingInstanceMaintenanceEventLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetInstanceMaintenanceEventLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListInstanceMaintenanceEventsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListInstanceMaintenanceEventsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListInstanceMaintenanceEventsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListInstanceMaintenanceEventsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListInstanceMaintenanceEventsResponse wrapper for the ListInstanceMaintenanceEvents operation
type ListInstanceMaintenanceEventsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []InstanceMaintenanceEventSummary instances
	Items []InstanceMaintenanceEventSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListInstanceMaintenanceEventsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListInstanceMaintenanceEventsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListInstanceMaintenanceEventsSortByEnum Enum with underlying type: string
type ListInstanceMaintenanceEventsSortByEnum string

// Set of constants representing the allowable values for ListInstanceMaintenanceEventsSortByEnum
const (
	ListInstanceMaintenanceEventsSortByTimecreated ListInstanceMaintenanceEventsSortByEnum = "TIMECREATED"
	ListInstanceMaintenanceEventsSortByDisplayname ListInstanceMaintenanceEventsSortByEnum = "DISPLAYNAME"
)

var mappingListInstanceMaintenanceEventsSortByEnum = map[string]ListInstanceMaintenanceEventsSortByEnum{
	"TIMECREATED": ListInstanceMaintenanceEventsSortByTimecreated,
	"DISPLAYNAME": ListInstanceMaintenanceEventsSortByDisplayname,
}

var mappingListInstanceMaintenanceEventsSortByEnumLowerCase = map[string]ListInstanceMaintenanceEventsSortByEnum{
	"timecreated": ListInstanceMaintenanceEventsSortByTimecreated,
	"displayname": ListInstanceMaintenanceEventsSortByDisplayname,
}

// GetListInstanceMaintenanceEventsSortByEnumValues Enumerates the set of values for ListInstanceMaintenanceEventsSortByEnum
func GetListInstanceMaintenanceEventsSortByEnumValues() []ListInstanceMaintenanceEventsSortByEnum {
	values := make([]ListInstanceMaintenanceEventsSortByEnum, 0)
	for _, v := range mappingListInstanceMaintenanceEventsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListInstanceMaintenanceEventsSortByEnumStringValues Enumerates the set of values in String for ListInstanceMaintenanceEventsSortByEnum
func GetListInstanceMaintenanceEventsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListInstanceMaintenanceEventsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListInstanceMaintenanceEventsSortByEnum(val string) (ListInstanceMaintenanceEventsSortByEnum, bool) {
	enum, ok := mappingListInstanceMaintenanceEventsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListInstanceMaintenanceEventsSortOrderEnum Enum with underlying type: string
type ListInstanceMaintenanceEventsSortOrderEnum string

// Set of constants representing the allowable values for ListInstanceMaintenanceEventsSortOrderEnum
const (
	ListInstanceMaintenanceEventsSortOrderAsc  ListInstanceMaintenanceEventsSortOrderEnum = "ASC"
	ListInstanceMaintenanceEventsSortOrderDesc ListInstanceMaintenanceEventsSortOrderEnum = "DESC"
)

var mappingListInstanceMaintenanceEventsSortOrderEnum = map[string]ListInstanceMaintenanceEventsSortOrderEnum{
	"ASC":  ListInstanceMaintenanceEventsSortOrderAsc,
	"DESC": ListInstanceMaintenanceEventsSortOrderDesc,
}

var mappingListInstanceMaintenanceEventsSortOrderEnumLowerCase = map[string]ListInstanceMaintenanceEventsSortOrderEnum{
	"asc":  ListInstanceMaintenanceEventsSortOrderAsc,
	"desc": ListInstanceMaintenanceEventsSortOrderDesc,
}

// GetListInstanceMaintenanceEventsSortOrderEnumValues Enumerates the set of values for ListInstanceMaintenanceEventsSortOrderEnum
func GetListInstanceMaintenanceEventsSortOrderEnumValues() []ListInstanceMaintenanceEventsSortOrderEnum {
	values := make([]ListInstanceMaintenanceEventsSortOrderEnum, 0)
	for _, v := range mappingListInstanceMaintenanceEventsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListInstanceMaintenanceEventsSortOrderEnumStringValues Enumerates the set of values in String for ListInstanceMaintenanceEventsSortOrderEnum
func GetListInstanceMaintenanceEventsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListInstanceMaintenanceEventsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListInstanceMaintenanceEventsSortOrderEnum(val string) (ListInstanceMaintenanceEventsSortOrderEnum, bool) {
	enum, ok := mappingListInstanceMaintenanceEventsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
