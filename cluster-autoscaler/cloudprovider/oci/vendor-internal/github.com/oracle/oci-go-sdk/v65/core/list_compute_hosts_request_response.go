// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"net/http"
	"strings"
)

// ListComputeHostsRequest wrapper for the ListComputeHosts operation
//
// # See also
//
// Click https://docs.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeHosts.go.html to see an example of how to use ListComputeHostsRequest.
type ListComputeHostsRequest struct {

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute host network resoruce.
	// - Customer-unique HPC island ID
	// - Customer-unique network block ID
	// - Customer-unique local block ID
	NetworkResourceId *string `mandatory:"false" contributesTo:"query" name:"networkResourceId"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListComputeHostsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeHostsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only ComputeHostSummary resources that match the given Compute Host lifecycle State OCID exactly.
	ComputeHostLifecycleState *string `mandatory:"false" contributesTo:"query" name:"computeHostLifecycleState"`

	// A filter to return only ComputeHostSummary resources that match the given Compute Host health State OCID exactly.
	ComputeHostHealth *string `mandatory:"false" contributesTo:"query" name:"computeHostHealth"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeHostsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeHostsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeHostsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeHostsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeHostsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListComputeHostsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeHostsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeHostsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeHostsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeHostsResponse wrapper for the ListComputeHosts operation
type ListComputeHostsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ComputeHostCollection instances
	ComputeHostCollection `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeHostsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeHostsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeHostsSortByEnum Enum with underlying type: string
type ListComputeHostsSortByEnum string

// Set of constants representing the allowable values for ListComputeHostsSortByEnum
const (
	ListComputeHostsSortByTimecreated ListComputeHostsSortByEnum = "TIMECREATED"
	ListComputeHostsSortByDisplayname ListComputeHostsSortByEnum = "DISPLAYNAME"
)

var mappingListComputeHostsSortByEnum = map[string]ListComputeHostsSortByEnum{
	"TIMECREATED": ListComputeHostsSortByTimecreated,
	"DISPLAYNAME": ListComputeHostsSortByDisplayname,
}

var mappingListComputeHostsSortByEnumLowerCase = map[string]ListComputeHostsSortByEnum{
	"timecreated": ListComputeHostsSortByTimecreated,
	"displayname": ListComputeHostsSortByDisplayname,
}

// GetListComputeHostsSortByEnumValues Enumerates the set of values for ListComputeHostsSortByEnum
func GetListComputeHostsSortByEnumValues() []ListComputeHostsSortByEnum {
	values := make([]ListComputeHostsSortByEnum, 0)
	for _, v := range mappingListComputeHostsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeHostsSortByEnumStringValues Enumerates the set of values in String for ListComputeHostsSortByEnum
func GetListComputeHostsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeHostsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeHostsSortByEnum(val string) (ListComputeHostsSortByEnum, bool) {
	enum, ok := mappingListComputeHostsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeHostsSortOrderEnum Enum with underlying type: string
type ListComputeHostsSortOrderEnum string

// Set of constants representing the allowable values for ListComputeHostsSortOrderEnum
const (
	ListComputeHostsSortOrderAsc  ListComputeHostsSortOrderEnum = "ASC"
	ListComputeHostsSortOrderDesc ListComputeHostsSortOrderEnum = "DESC"
)

var mappingListComputeHostsSortOrderEnum = map[string]ListComputeHostsSortOrderEnum{
	"ASC":  ListComputeHostsSortOrderAsc,
	"DESC": ListComputeHostsSortOrderDesc,
}

var mappingListComputeHostsSortOrderEnumLowerCase = map[string]ListComputeHostsSortOrderEnum{
	"asc":  ListComputeHostsSortOrderAsc,
	"desc": ListComputeHostsSortOrderDesc,
}

// GetListComputeHostsSortOrderEnumValues Enumerates the set of values for ListComputeHostsSortOrderEnum
func GetListComputeHostsSortOrderEnumValues() []ListComputeHostsSortOrderEnum {
	values := make([]ListComputeHostsSortOrderEnum, 0)
	for _, v := range mappingListComputeHostsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeHostsSortOrderEnumStringValues Enumerates the set of values in String for ListComputeHostsSortOrderEnum
func GetListComputeHostsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeHostsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeHostsSortOrderEnum(val string) (ListComputeHostsSortOrderEnum, bool) {
	enum, ok := mappingListComputeHostsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
