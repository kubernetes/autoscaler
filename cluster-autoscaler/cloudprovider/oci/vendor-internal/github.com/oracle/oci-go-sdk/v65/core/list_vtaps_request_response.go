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

// ListVtapsRequest wrapper for the ListVtaps operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListVtaps.go.html to see an example of how to use ListVtapsRequest.
type ListVtapsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN.
	VcnId *string `mandatory:"false" contributesTo:"query" name:"vcnId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VTAP source.
	Source *string `mandatory:"false" contributesTo:"query" name:"source"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VTAP target.
	TargetId *string `mandatory:"false" contributesTo:"query" name:"targetId"`

	// The IP address of the VTAP target.
	TargetIp *string `mandatory:"false" contributesTo:"query" name:"targetIp"`

	// Indicates whether to list all VTAPs or only running VTAPs.
	// * When `FALSE`, lists ALL running and stopped VTAPs.
	// * When `TRUE`, lists only running VTAPs (VTAPs where isVtapEnabled = `TRUE`).
	IsVtapEnabled *bool `mandatory:"false" contributesTo:"query" name:"isVtapEnabled"`

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
	SortBy ListVtapsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListVtapsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to return only resources that match the given VTAP administrative lifecycle state.
	// The state value is case-insensitive.
	LifecycleState VtapLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListVtapsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListVtapsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListVtapsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListVtapsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListVtapsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListVtapsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListVtapsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListVtapsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListVtapsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVtapLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetVtapLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListVtapsResponse wrapper for the ListVtaps operation
type ListVtapsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []Vtap instances
	Items []Vtap `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListVtapsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListVtapsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListVtapsSortByEnum Enum with underlying type: string
type ListVtapsSortByEnum string

// Set of constants representing the allowable values for ListVtapsSortByEnum
const (
	ListVtapsSortByTimecreated ListVtapsSortByEnum = "TIMECREATED"
	ListVtapsSortByDisplayname ListVtapsSortByEnum = "DISPLAYNAME"
)

var mappingListVtapsSortByEnum = map[string]ListVtapsSortByEnum{
	"TIMECREATED": ListVtapsSortByTimecreated,
	"DISPLAYNAME": ListVtapsSortByDisplayname,
}

var mappingListVtapsSortByEnumLowerCase = map[string]ListVtapsSortByEnum{
	"timecreated": ListVtapsSortByTimecreated,
	"displayname": ListVtapsSortByDisplayname,
}

// GetListVtapsSortByEnumValues Enumerates the set of values for ListVtapsSortByEnum
func GetListVtapsSortByEnumValues() []ListVtapsSortByEnum {
	values := make([]ListVtapsSortByEnum, 0)
	for _, v := range mappingListVtapsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListVtapsSortByEnumStringValues Enumerates the set of values in String for ListVtapsSortByEnum
func GetListVtapsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListVtapsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVtapsSortByEnum(val string) (ListVtapsSortByEnum, bool) {
	enum, ok := mappingListVtapsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListVtapsSortOrderEnum Enum with underlying type: string
type ListVtapsSortOrderEnum string

// Set of constants representing the allowable values for ListVtapsSortOrderEnum
const (
	ListVtapsSortOrderAsc  ListVtapsSortOrderEnum = "ASC"
	ListVtapsSortOrderDesc ListVtapsSortOrderEnum = "DESC"
)

var mappingListVtapsSortOrderEnum = map[string]ListVtapsSortOrderEnum{
	"ASC":  ListVtapsSortOrderAsc,
	"DESC": ListVtapsSortOrderDesc,
}

var mappingListVtapsSortOrderEnumLowerCase = map[string]ListVtapsSortOrderEnum{
	"asc":  ListVtapsSortOrderAsc,
	"desc": ListVtapsSortOrderDesc,
}

// GetListVtapsSortOrderEnumValues Enumerates the set of values for ListVtapsSortOrderEnum
func GetListVtapsSortOrderEnumValues() []ListVtapsSortOrderEnum {
	values := make([]ListVtapsSortOrderEnum, 0)
	for _, v := range mappingListVtapsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListVtapsSortOrderEnumStringValues Enumerates the set of values in String for ListVtapsSortOrderEnum
func GetListVtapsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListVtapsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVtapsSortOrderEnum(val string) (ListVtapsSortOrderEnum, bool) {
	enum, ok := mappingListVtapsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
