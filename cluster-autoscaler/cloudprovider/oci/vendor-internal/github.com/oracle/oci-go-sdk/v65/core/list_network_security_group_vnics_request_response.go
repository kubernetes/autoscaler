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

// ListNetworkSecurityGroupVnicsRequest wrapper for the ListNetworkSecurityGroupVnics operation
//
// # See also
//
// Click https://docs.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListNetworkSecurityGroupVnics.go.html to see an example of how to use ListNetworkSecurityGroupVnicsRequest.
type ListNetworkSecurityGroupVnicsRequest struct {

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the network security group.
	NetworkSecurityGroupId *string `mandatory:"true" contributesTo:"path" name:"networkSecurityGroupId"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The field to sort by.
	SortBy ListNetworkSecurityGroupVnicsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListNetworkSecurityGroupVnicsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListNetworkSecurityGroupVnicsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListNetworkSecurityGroupVnicsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListNetworkSecurityGroupVnicsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListNetworkSecurityGroupVnicsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListNetworkSecurityGroupVnicsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListNetworkSecurityGroupVnicsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListNetworkSecurityGroupVnicsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListNetworkSecurityGroupVnicsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListNetworkSecurityGroupVnicsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListNetworkSecurityGroupVnicsResponse wrapper for the ListNetworkSecurityGroupVnics operation
type ListNetworkSecurityGroupVnicsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []NetworkSecurityGroupVnic instances
	Items []NetworkSecurityGroupVnic `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListNetworkSecurityGroupVnicsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListNetworkSecurityGroupVnicsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListNetworkSecurityGroupVnicsSortByEnum Enum with underlying type: string
type ListNetworkSecurityGroupVnicsSortByEnum string

// Set of constants representing the allowable values for ListNetworkSecurityGroupVnicsSortByEnum
const (
	ListNetworkSecurityGroupVnicsSortByTimeassociated ListNetworkSecurityGroupVnicsSortByEnum = "TIMEASSOCIATED"
)

var mappingListNetworkSecurityGroupVnicsSortByEnum = map[string]ListNetworkSecurityGroupVnicsSortByEnum{
	"TIMEASSOCIATED": ListNetworkSecurityGroupVnicsSortByTimeassociated,
}

var mappingListNetworkSecurityGroupVnicsSortByEnumLowerCase = map[string]ListNetworkSecurityGroupVnicsSortByEnum{
	"timeassociated": ListNetworkSecurityGroupVnicsSortByTimeassociated,
}

// GetListNetworkSecurityGroupVnicsSortByEnumValues Enumerates the set of values for ListNetworkSecurityGroupVnicsSortByEnum
func GetListNetworkSecurityGroupVnicsSortByEnumValues() []ListNetworkSecurityGroupVnicsSortByEnum {
	values := make([]ListNetworkSecurityGroupVnicsSortByEnum, 0)
	for _, v := range mappingListNetworkSecurityGroupVnicsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListNetworkSecurityGroupVnicsSortByEnumStringValues Enumerates the set of values in String for ListNetworkSecurityGroupVnicsSortByEnum
func GetListNetworkSecurityGroupVnicsSortByEnumStringValues() []string {
	return []string{
		"TIMEASSOCIATED",
	}
}

// GetMappingListNetworkSecurityGroupVnicsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListNetworkSecurityGroupVnicsSortByEnum(val string) (ListNetworkSecurityGroupVnicsSortByEnum, bool) {
	enum, ok := mappingListNetworkSecurityGroupVnicsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListNetworkSecurityGroupVnicsSortOrderEnum Enum with underlying type: string
type ListNetworkSecurityGroupVnicsSortOrderEnum string

// Set of constants representing the allowable values for ListNetworkSecurityGroupVnicsSortOrderEnum
const (
	ListNetworkSecurityGroupVnicsSortOrderAsc  ListNetworkSecurityGroupVnicsSortOrderEnum = "ASC"
	ListNetworkSecurityGroupVnicsSortOrderDesc ListNetworkSecurityGroupVnicsSortOrderEnum = "DESC"
)

var mappingListNetworkSecurityGroupVnicsSortOrderEnum = map[string]ListNetworkSecurityGroupVnicsSortOrderEnum{
	"ASC":  ListNetworkSecurityGroupVnicsSortOrderAsc,
	"DESC": ListNetworkSecurityGroupVnicsSortOrderDesc,
}

var mappingListNetworkSecurityGroupVnicsSortOrderEnumLowerCase = map[string]ListNetworkSecurityGroupVnicsSortOrderEnum{
	"asc":  ListNetworkSecurityGroupVnicsSortOrderAsc,
	"desc": ListNetworkSecurityGroupVnicsSortOrderDesc,
}

// GetListNetworkSecurityGroupVnicsSortOrderEnumValues Enumerates the set of values for ListNetworkSecurityGroupVnicsSortOrderEnum
func GetListNetworkSecurityGroupVnicsSortOrderEnumValues() []ListNetworkSecurityGroupVnicsSortOrderEnum {
	values := make([]ListNetworkSecurityGroupVnicsSortOrderEnum, 0)
	for _, v := range mappingListNetworkSecurityGroupVnicsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListNetworkSecurityGroupVnicsSortOrderEnumStringValues Enumerates the set of values in String for ListNetworkSecurityGroupVnicsSortOrderEnum
func GetListNetworkSecurityGroupVnicsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListNetworkSecurityGroupVnicsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListNetworkSecurityGroupVnicsSortOrderEnum(val string) (ListNetworkSecurityGroupVnicsSortOrderEnum, bool) {
	enum, ok := mappingListNetworkSecurityGroupVnicsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
