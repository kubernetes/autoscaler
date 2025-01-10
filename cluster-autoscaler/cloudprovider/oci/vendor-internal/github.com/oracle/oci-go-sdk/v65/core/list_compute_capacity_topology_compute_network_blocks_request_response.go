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

// ListComputeCapacityTopologyComputeNetworkBlocksRequest wrapper for the ListComputeCapacityTopologyComputeNetworkBlocks operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeCapacityTopologyComputeNetworkBlocks.go.html to see an example of how to use ListComputeCapacityTopologyComputeNetworkBlocksRequest.
type ListComputeCapacityTopologyComputeNetworkBlocksRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute capacity topology.
	ComputeCapacityTopologyId *string `mandatory:"true" contributesTo:"path" name:"computeCapacityTopologyId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute HPC island.
	ComputeHpcIslandId *string `mandatory:"false" contributesTo:"query" name:"computeHpcIslandId"`

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
	SortBy ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeCapacityTopologyComputeNetworkBlocksRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeCapacityTopologyComputeNetworkBlocksRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeCapacityTopologyComputeNetworkBlocksRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeCapacityTopologyComputeNetworkBlocksRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeCapacityTopologyComputeNetworkBlocksRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListComputeCapacityTopologyComputeNetworkBlocksSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeCapacityTopologyComputeNetworkBlocksSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeCapacityTopologyComputeNetworkBlocksResponse wrapper for the ListComputeCapacityTopologyComputeNetworkBlocks operation
type ListComputeCapacityTopologyComputeNetworkBlocksResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ComputeNetworkBlockCollection instances
	ComputeNetworkBlockCollection `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeCapacityTopologyComputeNetworkBlocksResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeCapacityTopologyComputeNetworkBlocksResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum Enum with underlying type: string
type ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum string

// Set of constants representing the allowable values for ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum
const (
	ListComputeCapacityTopologyComputeNetworkBlocksSortByTimecreated ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum = "TIMECREATED"
	ListComputeCapacityTopologyComputeNetworkBlocksSortByDisplayname ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum = "DISPLAYNAME"
)

var mappingListComputeCapacityTopologyComputeNetworkBlocksSortByEnum = map[string]ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum{
	"TIMECREATED": ListComputeCapacityTopologyComputeNetworkBlocksSortByTimecreated,
	"DISPLAYNAME": ListComputeCapacityTopologyComputeNetworkBlocksSortByDisplayname,
}

var mappingListComputeCapacityTopologyComputeNetworkBlocksSortByEnumLowerCase = map[string]ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum{
	"timecreated": ListComputeCapacityTopologyComputeNetworkBlocksSortByTimecreated,
	"displayname": ListComputeCapacityTopologyComputeNetworkBlocksSortByDisplayname,
}

// GetListComputeCapacityTopologyComputeNetworkBlocksSortByEnumValues Enumerates the set of values for ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum
func GetListComputeCapacityTopologyComputeNetworkBlocksSortByEnumValues() []ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum {
	values := make([]ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum, 0)
	for _, v := range mappingListComputeCapacityTopologyComputeNetworkBlocksSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeCapacityTopologyComputeNetworkBlocksSortByEnumStringValues Enumerates the set of values in String for ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum
func GetListComputeCapacityTopologyComputeNetworkBlocksSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeCapacityTopologyComputeNetworkBlocksSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeCapacityTopologyComputeNetworkBlocksSortByEnum(val string) (ListComputeCapacityTopologyComputeNetworkBlocksSortByEnum, bool) {
	enum, ok := mappingListComputeCapacityTopologyComputeNetworkBlocksSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum Enum with underlying type: string
type ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum string

// Set of constants representing the allowable values for ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum
const (
	ListComputeCapacityTopologyComputeNetworkBlocksSortOrderAsc  ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum = "ASC"
	ListComputeCapacityTopologyComputeNetworkBlocksSortOrderDesc ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum = "DESC"
)

var mappingListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum = map[string]ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum{
	"ASC":  ListComputeCapacityTopologyComputeNetworkBlocksSortOrderAsc,
	"DESC": ListComputeCapacityTopologyComputeNetworkBlocksSortOrderDesc,
}

var mappingListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnumLowerCase = map[string]ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum{
	"asc":  ListComputeCapacityTopologyComputeNetworkBlocksSortOrderAsc,
	"desc": ListComputeCapacityTopologyComputeNetworkBlocksSortOrderDesc,
}

// GetListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnumValues Enumerates the set of values for ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum
func GetListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnumValues() []ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum {
	values := make([]ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum, 0)
	for _, v := range mappingListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnumStringValues Enumerates the set of values in String for ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum
func GetListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum(val string) (ListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnum, bool) {
	enum, ok := mappingListComputeCapacityTopologyComputeNetworkBlocksSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
