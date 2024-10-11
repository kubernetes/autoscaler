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

// ListComputeCapacityTopologyComputeBareMetalHostsRequest wrapper for the ListComputeCapacityTopologyComputeBareMetalHosts operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListComputeCapacityTopologyComputeBareMetalHosts.go.html to see an example of how to use ListComputeCapacityTopologyComputeBareMetalHostsRequest.
type ListComputeCapacityTopologyComputeBareMetalHostsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute capacity topology.
	ComputeCapacityTopologyId *string `mandatory:"true" contributesTo:"path" name:"computeCapacityTopologyId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute HPC island.
	ComputeHpcIslandId *string `mandatory:"false" contributesTo:"query" name:"computeHpcIslandId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute network block.
	ComputeNetworkBlockId *string `mandatory:"false" contributesTo:"query" name:"computeNetworkBlockId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute local block.
	ComputeLocalBlockId *string `mandatory:"false" contributesTo:"query" name:"computeLocalBlockId"`

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
	SortBy ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeCapacityTopologyComputeBareMetalHostsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeCapacityTopologyComputeBareMetalHostsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListComputeCapacityTopologyComputeBareMetalHostsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeCapacityTopologyComputeBareMetalHostsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListComputeCapacityTopologyComputeBareMetalHostsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListComputeCapacityTopologyComputeBareMetalHostsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListComputeCapacityTopologyComputeBareMetalHostsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListComputeCapacityTopologyComputeBareMetalHostsResponse wrapper for the ListComputeCapacityTopologyComputeBareMetalHosts operation
type ListComputeCapacityTopologyComputeBareMetalHostsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ComputeBareMetalHostCollection instances
	ComputeBareMetalHostCollection `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeCapacityTopologyComputeBareMetalHostsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeCapacityTopologyComputeBareMetalHostsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum Enum with underlying type: string
type ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum string

// Set of constants representing the allowable values for ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum
const (
	ListComputeCapacityTopologyComputeBareMetalHostsSortByTimecreated ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum = "TIMECREATED"
	ListComputeCapacityTopologyComputeBareMetalHostsSortByDisplayname ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum = "DISPLAYNAME"
)

var mappingListComputeCapacityTopologyComputeBareMetalHostsSortByEnum = map[string]ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum{
	"TIMECREATED": ListComputeCapacityTopologyComputeBareMetalHostsSortByTimecreated,
	"DISPLAYNAME": ListComputeCapacityTopologyComputeBareMetalHostsSortByDisplayname,
}

var mappingListComputeCapacityTopologyComputeBareMetalHostsSortByEnumLowerCase = map[string]ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum{
	"timecreated": ListComputeCapacityTopologyComputeBareMetalHostsSortByTimecreated,
	"displayname": ListComputeCapacityTopologyComputeBareMetalHostsSortByDisplayname,
}

// GetListComputeCapacityTopologyComputeBareMetalHostsSortByEnumValues Enumerates the set of values for ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum
func GetListComputeCapacityTopologyComputeBareMetalHostsSortByEnumValues() []ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum {
	values := make([]ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum, 0)
	for _, v := range mappingListComputeCapacityTopologyComputeBareMetalHostsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeCapacityTopologyComputeBareMetalHostsSortByEnumStringValues Enumerates the set of values in String for ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum
func GetListComputeCapacityTopologyComputeBareMetalHostsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListComputeCapacityTopologyComputeBareMetalHostsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeCapacityTopologyComputeBareMetalHostsSortByEnum(val string) (ListComputeCapacityTopologyComputeBareMetalHostsSortByEnum, bool) {
	enum, ok := mappingListComputeCapacityTopologyComputeBareMetalHostsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum Enum with underlying type: string
type ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum string

// Set of constants representing the allowable values for ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum
const (
	ListComputeCapacityTopologyComputeBareMetalHostsSortOrderAsc  ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum = "ASC"
	ListComputeCapacityTopologyComputeBareMetalHostsSortOrderDesc ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum = "DESC"
)

var mappingListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum = map[string]ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum{
	"ASC":  ListComputeCapacityTopologyComputeBareMetalHostsSortOrderAsc,
	"DESC": ListComputeCapacityTopologyComputeBareMetalHostsSortOrderDesc,
}

var mappingListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnumLowerCase = map[string]ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum{
	"asc":  ListComputeCapacityTopologyComputeBareMetalHostsSortOrderAsc,
	"desc": ListComputeCapacityTopologyComputeBareMetalHostsSortOrderDesc,
}

// GetListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnumValues Enumerates the set of values for ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum
func GetListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnumValues() []ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum {
	values := make([]ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum, 0)
	for _, v := range mappingListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnumStringValues Enumerates the set of values in String for ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum
func GetListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum(val string) (ListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnum, bool) {
	enum, ok := mappingListComputeCapacityTopologyComputeBareMetalHostsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
