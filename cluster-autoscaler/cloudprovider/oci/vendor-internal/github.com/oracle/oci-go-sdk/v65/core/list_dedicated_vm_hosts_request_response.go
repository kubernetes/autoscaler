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

// ListDedicatedVmHostsRequest wrapper for the ListDedicatedVmHosts operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListDedicatedVmHosts.go.html to see an example of how to use ListDedicatedVmHostsRequest.
type ListDedicatedVmHostsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// A filter to only return resources that match the given lifecycle state.
	LifecycleState ListDedicatedVmHostsLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The name for the instance's shape.
	InstanceShapeName *string `mandatory:"false" contributesTo:"query" name:"instanceShapeName"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListDedicatedVmHostsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListDedicatedVmHostsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The remaining memory of the dedicated VM host, in GBs.
	RemainingMemoryInGBsGreaterThanOrEqualTo *float32 `mandatory:"false" contributesTo:"query" name:"remainingMemoryInGBsGreaterThanOrEqualTo"`

	// The available OCPUs of the dedicated VM host.
	RemainingOcpusGreaterThanOrEqualTo *float32 `mandatory:"false" contributesTo:"query" name:"remainingOcpusGreaterThanOrEqualTo"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListDedicatedVmHostsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListDedicatedVmHostsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListDedicatedVmHostsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListDedicatedVmHostsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListDedicatedVmHostsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListDedicatedVmHostsLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetListDedicatedVmHostsLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListDedicatedVmHostsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListDedicatedVmHostsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListDedicatedVmHostsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListDedicatedVmHostsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListDedicatedVmHostsResponse wrapper for the ListDedicatedVmHosts operation
type ListDedicatedVmHostsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []DedicatedVmHostSummary instances
	Items []DedicatedVmHostSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListDedicatedVmHostsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListDedicatedVmHostsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListDedicatedVmHostsLifecycleStateEnum Enum with underlying type: string
type ListDedicatedVmHostsLifecycleStateEnum string

// Set of constants representing the allowable values for ListDedicatedVmHostsLifecycleStateEnum
const (
	ListDedicatedVmHostsLifecycleStateCreating ListDedicatedVmHostsLifecycleStateEnum = "CREATING"
	ListDedicatedVmHostsLifecycleStateActive   ListDedicatedVmHostsLifecycleStateEnum = "ACTIVE"
	ListDedicatedVmHostsLifecycleStateUpdating ListDedicatedVmHostsLifecycleStateEnum = "UPDATING"
	ListDedicatedVmHostsLifecycleStateDeleting ListDedicatedVmHostsLifecycleStateEnum = "DELETING"
	ListDedicatedVmHostsLifecycleStateDeleted  ListDedicatedVmHostsLifecycleStateEnum = "DELETED"
	ListDedicatedVmHostsLifecycleStateFailed   ListDedicatedVmHostsLifecycleStateEnum = "FAILED"
)

var mappingListDedicatedVmHostsLifecycleStateEnum = map[string]ListDedicatedVmHostsLifecycleStateEnum{
	"CREATING": ListDedicatedVmHostsLifecycleStateCreating,
	"ACTIVE":   ListDedicatedVmHostsLifecycleStateActive,
	"UPDATING": ListDedicatedVmHostsLifecycleStateUpdating,
	"DELETING": ListDedicatedVmHostsLifecycleStateDeleting,
	"DELETED":  ListDedicatedVmHostsLifecycleStateDeleted,
	"FAILED":   ListDedicatedVmHostsLifecycleStateFailed,
}

var mappingListDedicatedVmHostsLifecycleStateEnumLowerCase = map[string]ListDedicatedVmHostsLifecycleStateEnum{
	"creating": ListDedicatedVmHostsLifecycleStateCreating,
	"active":   ListDedicatedVmHostsLifecycleStateActive,
	"updating": ListDedicatedVmHostsLifecycleStateUpdating,
	"deleting": ListDedicatedVmHostsLifecycleStateDeleting,
	"deleted":  ListDedicatedVmHostsLifecycleStateDeleted,
	"failed":   ListDedicatedVmHostsLifecycleStateFailed,
}

// GetListDedicatedVmHostsLifecycleStateEnumValues Enumerates the set of values for ListDedicatedVmHostsLifecycleStateEnum
func GetListDedicatedVmHostsLifecycleStateEnumValues() []ListDedicatedVmHostsLifecycleStateEnum {
	values := make([]ListDedicatedVmHostsLifecycleStateEnum, 0)
	for _, v := range mappingListDedicatedVmHostsLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetListDedicatedVmHostsLifecycleStateEnumStringValues Enumerates the set of values in String for ListDedicatedVmHostsLifecycleStateEnum
func GetListDedicatedVmHostsLifecycleStateEnumStringValues() []string {
	return []string{
		"CREATING",
		"ACTIVE",
		"UPDATING",
		"DELETING",
		"DELETED",
		"FAILED",
	}
}

// GetMappingListDedicatedVmHostsLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListDedicatedVmHostsLifecycleStateEnum(val string) (ListDedicatedVmHostsLifecycleStateEnum, bool) {
	enum, ok := mappingListDedicatedVmHostsLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListDedicatedVmHostsSortByEnum Enum with underlying type: string
type ListDedicatedVmHostsSortByEnum string

// Set of constants representing the allowable values for ListDedicatedVmHostsSortByEnum
const (
	ListDedicatedVmHostsSortByTimecreated ListDedicatedVmHostsSortByEnum = "TIMECREATED"
	ListDedicatedVmHostsSortByDisplayname ListDedicatedVmHostsSortByEnum = "DISPLAYNAME"
)

var mappingListDedicatedVmHostsSortByEnum = map[string]ListDedicatedVmHostsSortByEnum{
	"TIMECREATED": ListDedicatedVmHostsSortByTimecreated,
	"DISPLAYNAME": ListDedicatedVmHostsSortByDisplayname,
}

var mappingListDedicatedVmHostsSortByEnumLowerCase = map[string]ListDedicatedVmHostsSortByEnum{
	"timecreated": ListDedicatedVmHostsSortByTimecreated,
	"displayname": ListDedicatedVmHostsSortByDisplayname,
}

// GetListDedicatedVmHostsSortByEnumValues Enumerates the set of values for ListDedicatedVmHostsSortByEnum
func GetListDedicatedVmHostsSortByEnumValues() []ListDedicatedVmHostsSortByEnum {
	values := make([]ListDedicatedVmHostsSortByEnum, 0)
	for _, v := range mappingListDedicatedVmHostsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListDedicatedVmHostsSortByEnumStringValues Enumerates the set of values in String for ListDedicatedVmHostsSortByEnum
func GetListDedicatedVmHostsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListDedicatedVmHostsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListDedicatedVmHostsSortByEnum(val string) (ListDedicatedVmHostsSortByEnum, bool) {
	enum, ok := mappingListDedicatedVmHostsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListDedicatedVmHostsSortOrderEnum Enum with underlying type: string
type ListDedicatedVmHostsSortOrderEnum string

// Set of constants representing the allowable values for ListDedicatedVmHostsSortOrderEnum
const (
	ListDedicatedVmHostsSortOrderAsc  ListDedicatedVmHostsSortOrderEnum = "ASC"
	ListDedicatedVmHostsSortOrderDesc ListDedicatedVmHostsSortOrderEnum = "DESC"
)

var mappingListDedicatedVmHostsSortOrderEnum = map[string]ListDedicatedVmHostsSortOrderEnum{
	"ASC":  ListDedicatedVmHostsSortOrderAsc,
	"DESC": ListDedicatedVmHostsSortOrderDesc,
}

var mappingListDedicatedVmHostsSortOrderEnumLowerCase = map[string]ListDedicatedVmHostsSortOrderEnum{
	"asc":  ListDedicatedVmHostsSortOrderAsc,
	"desc": ListDedicatedVmHostsSortOrderDesc,
}

// GetListDedicatedVmHostsSortOrderEnumValues Enumerates the set of values for ListDedicatedVmHostsSortOrderEnum
func GetListDedicatedVmHostsSortOrderEnumValues() []ListDedicatedVmHostsSortOrderEnum {
	values := make([]ListDedicatedVmHostsSortOrderEnum, 0)
	for _, v := range mappingListDedicatedVmHostsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListDedicatedVmHostsSortOrderEnumStringValues Enumerates the set of values in String for ListDedicatedVmHostsSortOrderEnum
func GetListDedicatedVmHostsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListDedicatedVmHostsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListDedicatedVmHostsSortOrderEnum(val string) (ListDedicatedVmHostsSortOrderEnum, bool) {
	enum, ok := mappingListDedicatedVmHostsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
