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

// ListDrgsByStatesRequest wrapper for the ListDrgsByStates operation
type ListDrgsByStatesRequest struct {

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// A query param to return resources that match the given compartment OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) exactly.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// The State of the DRG (Classical/Migrated/Upgraded) of the DRG.
	DrgState DrgUpgradeStateStateEnum `mandatory:"false" contributesTo:"query" name:"drgState" omitEmpty:"true"`

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
	SortBy ListDrgsByStatesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListDrgsByStatesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListDrgsByStatesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListDrgsByStatesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListDrgsByStatesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListDrgsByStatesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListDrgsByStatesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingDrgUpgradeStateStateEnum[string(request.DrgState)]; !ok && request.DrgState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DrgState: %s. Supported values are: %s.", request.DrgState, strings.Join(GetDrgUpgradeStateStateEnumStringValues(), ",")))
	}
	if _, ok := mappingListDrgsByStatesSortByEnum[string(request.SortBy)]; !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListDrgsByStatesSortByEnumStringValues(), ",")))
	}
	if _, ok := mappingListDrgsByStatesSortOrderEnum[string(request.SortOrder)]; !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListDrgsByStatesSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListDrgsByStatesResponse wrapper for the ListDrgsByStates operation
type ListDrgsByStatesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []Drg instances
	Items []Drg `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListDrgsByStatesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListDrgsByStatesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListDrgsByStatesSortByEnum Enum with underlying type: string
type ListDrgsByStatesSortByEnum string

// Set of constants representing the allowable values for ListDrgsByStatesSortByEnum
const (
	ListDrgsByStatesSortByTimecreated ListDrgsByStatesSortByEnum = "TIMECREATED"
	ListDrgsByStatesSortByDisplayname ListDrgsByStatesSortByEnum = "DISPLAYNAME"
)

var mappingListDrgsByStatesSortByEnum = map[string]ListDrgsByStatesSortByEnum{
	"TIMECREATED": ListDrgsByStatesSortByTimecreated,
	"DISPLAYNAME": ListDrgsByStatesSortByDisplayname,
}

// GetListDrgsByStatesSortByEnumValues Enumerates the set of values for ListDrgsByStatesSortByEnum
func GetListDrgsByStatesSortByEnumValues() []ListDrgsByStatesSortByEnum {
	values := make([]ListDrgsByStatesSortByEnum, 0)
	for _, v := range mappingListDrgsByStatesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListDrgsByStatesSortByEnumStringValues Enumerates the set of values in String for ListDrgsByStatesSortByEnum
func GetListDrgsByStatesSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// ListDrgsByStatesSortOrderEnum Enum with underlying type: string
type ListDrgsByStatesSortOrderEnum string

// Set of constants representing the allowable values for ListDrgsByStatesSortOrderEnum
const (
	ListDrgsByStatesSortOrderAsc  ListDrgsByStatesSortOrderEnum = "ASC"
	ListDrgsByStatesSortOrderDesc ListDrgsByStatesSortOrderEnum = "DESC"
)

var mappingListDrgsByStatesSortOrderEnum = map[string]ListDrgsByStatesSortOrderEnum{
	"ASC":  ListDrgsByStatesSortOrderAsc,
	"DESC": ListDrgsByStatesSortOrderDesc,
}

// GetListDrgsByStatesSortOrderEnumValues Enumerates the set of values for ListDrgsByStatesSortOrderEnum
func GetListDrgsByStatesSortOrderEnumValues() []ListDrgsByStatesSortOrderEnum {
	values := make([]ListDrgsByStatesSortOrderEnum, 0)
	for _, v := range mappingListDrgsByStatesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListDrgsByStatesSortOrderEnumStringValues Enumerates the set of values in String for ListDrgsByStatesSortOrderEnum
func GetListDrgsByStatesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}
