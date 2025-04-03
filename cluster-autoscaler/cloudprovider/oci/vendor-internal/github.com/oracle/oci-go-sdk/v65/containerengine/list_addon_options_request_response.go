// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package containerengine

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"net/http"
	"strings"
)

// ListAddonOptionsRequest wrapper for the ListAddonOptions operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListAddonOptions.go.html to see an example of how to use ListAddonOptionsRequest.
type ListAddonOptionsRequest struct {

	// The kubernetes version to fetch the addons.
	KubernetesVersion *string `mandatory:"true" contributesTo:"query" name:"kubernetesVersion"`

	// The name of the addon.
	AddonName *string `mandatory:"false" contributesTo:"query" name:"addonName"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated "List" call.
	// 1 is the minimum, 1000 is the maximum. For important details about how pagination works,
	// see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List" call.
	// For important details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The optional order in which to sort the results.
	SortOrder ListAddonOptionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the results by.
	SortBy ListAddonOptionsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListAddonOptionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListAddonOptionsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListAddonOptionsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListAddonOptionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListAddonOptionsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListAddonOptionsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListAddonOptionsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListAddonOptionsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListAddonOptionsSortByEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListAddonOptionsResponse wrapper for the ListAddonOptions operation
type ListAddonOptionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []AddonOptionSummary instances
	Items []AddonOptionSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of results remain.
	// For important details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListAddonOptionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListAddonOptionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListAddonOptionsSortOrderEnum Enum with underlying type: string
type ListAddonOptionsSortOrderEnum string

// Set of constants representing the allowable values for ListAddonOptionsSortOrderEnum
const (
	ListAddonOptionsSortOrderAsc  ListAddonOptionsSortOrderEnum = "ASC"
	ListAddonOptionsSortOrderDesc ListAddonOptionsSortOrderEnum = "DESC"
)

var mappingListAddonOptionsSortOrderEnum = map[string]ListAddonOptionsSortOrderEnum{
	"ASC":  ListAddonOptionsSortOrderAsc,
	"DESC": ListAddonOptionsSortOrderDesc,
}

var mappingListAddonOptionsSortOrderEnumLowerCase = map[string]ListAddonOptionsSortOrderEnum{
	"asc":  ListAddonOptionsSortOrderAsc,
	"desc": ListAddonOptionsSortOrderDesc,
}

// GetListAddonOptionsSortOrderEnumValues Enumerates the set of values for ListAddonOptionsSortOrderEnum
func GetListAddonOptionsSortOrderEnumValues() []ListAddonOptionsSortOrderEnum {
	values := make([]ListAddonOptionsSortOrderEnum, 0)
	for _, v := range mappingListAddonOptionsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListAddonOptionsSortOrderEnumStringValues Enumerates the set of values in String for ListAddonOptionsSortOrderEnum
func GetListAddonOptionsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListAddonOptionsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListAddonOptionsSortOrderEnum(val string) (ListAddonOptionsSortOrderEnum, bool) {
	enum, ok := mappingListAddonOptionsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListAddonOptionsSortByEnum Enum with underlying type: string
type ListAddonOptionsSortByEnum string

// Set of constants representing the allowable values for ListAddonOptionsSortByEnum
const (
	ListAddonOptionsSortByName        ListAddonOptionsSortByEnum = "NAME"
	ListAddonOptionsSortByTimeCreated ListAddonOptionsSortByEnum = "TIME_CREATED"
)

var mappingListAddonOptionsSortByEnum = map[string]ListAddonOptionsSortByEnum{
	"NAME":         ListAddonOptionsSortByName,
	"TIME_CREATED": ListAddonOptionsSortByTimeCreated,
}

var mappingListAddonOptionsSortByEnumLowerCase = map[string]ListAddonOptionsSortByEnum{
	"name":         ListAddonOptionsSortByName,
	"time_created": ListAddonOptionsSortByTimeCreated,
}

// GetListAddonOptionsSortByEnumValues Enumerates the set of values for ListAddonOptionsSortByEnum
func GetListAddonOptionsSortByEnumValues() []ListAddonOptionsSortByEnum {
	values := make([]ListAddonOptionsSortByEnum, 0)
	for _, v := range mappingListAddonOptionsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListAddonOptionsSortByEnumStringValues Enumerates the set of values in String for ListAddonOptionsSortByEnum
func GetListAddonOptionsSortByEnumStringValues() []string {
	return []string{
		"NAME",
		"TIME_CREATED",
	}
}

// GetMappingListAddonOptionsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListAddonOptionsSortByEnum(val string) (ListAddonOptionsSortByEnum, bool) {
	enum, ok := mappingListAddonOptionsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
