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

// ListClientVpnUsersRequest wrapper for the ListClientVpnUsers operation
type ListClientVpnUsersRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the ClientVpn.
	ClientVpnId *string `mandatory:"true" contributesTo:"path" name:"clientVpnId"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListClientVpnUsersSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListClientVpnUsersSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

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

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListClientVpnUsersRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListClientVpnUsersRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListClientVpnUsersRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListClientVpnUsersRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListClientVpnUsersRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingListClientVpnUsersSortByEnum[string(request.SortBy)]; !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListClientVpnUsersSortByEnumStringValues(), ",")))
	}
	if _, ok := mappingListClientVpnUsersSortOrderEnum[string(request.SortOrder)]; !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListClientVpnUsersSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListClientVpnUsersResponse wrapper for the ListClientVpnUsers operation
type ListClientVpnUsersResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of ClientVpnUserSummaryCollection instances
	ClientVpnUserSummaryCollection `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListClientVpnUsersResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListClientVpnUsersResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListClientVpnUsersSortByEnum Enum with underlying type: string
type ListClientVpnUsersSortByEnum string

// Set of constants representing the allowable values for ListClientVpnUsersSortByEnum
const (
	ListClientVpnUsersSortByTimecreated ListClientVpnUsersSortByEnum = "TIMECREATED"
	ListClientVpnUsersSortByDisplayname ListClientVpnUsersSortByEnum = "DISPLAYNAME"
)

var mappingListClientVpnUsersSortByEnum = map[string]ListClientVpnUsersSortByEnum{
	"TIMECREATED": ListClientVpnUsersSortByTimecreated,
	"DISPLAYNAME": ListClientVpnUsersSortByDisplayname,
}

// GetListClientVpnUsersSortByEnumValues Enumerates the set of values for ListClientVpnUsersSortByEnum
func GetListClientVpnUsersSortByEnumValues() []ListClientVpnUsersSortByEnum {
	values := make([]ListClientVpnUsersSortByEnum, 0)
	for _, v := range mappingListClientVpnUsersSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListClientVpnUsersSortByEnumStringValues Enumerates the set of values in String for ListClientVpnUsersSortByEnum
func GetListClientVpnUsersSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// ListClientVpnUsersSortOrderEnum Enum with underlying type: string
type ListClientVpnUsersSortOrderEnum string

// Set of constants representing the allowable values for ListClientVpnUsersSortOrderEnum
const (
	ListClientVpnUsersSortOrderAsc  ListClientVpnUsersSortOrderEnum = "ASC"
	ListClientVpnUsersSortOrderDesc ListClientVpnUsersSortOrderEnum = "DESC"
)

var mappingListClientVpnUsersSortOrderEnum = map[string]ListClientVpnUsersSortOrderEnum{
	"ASC":  ListClientVpnUsersSortOrderAsc,
	"DESC": ListClientVpnUsersSortOrderDesc,
}

// GetListClientVpnUsersSortOrderEnumValues Enumerates the set of values for ListClientVpnUsersSortOrderEnum
func GetListClientVpnUsersSortOrderEnumValues() []ListClientVpnUsersSortOrderEnum {
	values := make([]ListClientVpnUsersSortOrderEnum, 0)
	for _, v := range mappingListClientVpnUsersSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListClientVpnUsersSortOrderEnumStringValues Enumerates the set of values in String for ListClientVpnUsersSortOrderEnum
func GetListClientVpnUsersSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}
