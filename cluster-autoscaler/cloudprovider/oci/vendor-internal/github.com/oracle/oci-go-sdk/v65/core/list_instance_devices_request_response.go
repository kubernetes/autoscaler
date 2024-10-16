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

// ListInstanceDevicesRequest wrapper for the ListInstanceDevices operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListInstanceDevices.go.html to see an example of how to use ListInstanceDevicesRequest.
type ListInstanceDevicesRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the instance.
	InstanceId *string `mandatory:"true" contributesTo:"path" name:"instanceId"`

	// A filter to return only available devices or only used devices.
	IsAvailable *bool `mandatory:"false" contributesTo:"query" name:"isAvailable"`

	// A filter to return only devices that match the given name exactly.
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

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
	SortBy ListInstanceDevicesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListInstanceDevicesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListInstanceDevicesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListInstanceDevicesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListInstanceDevicesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListInstanceDevicesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListInstanceDevicesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListInstanceDevicesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListInstanceDevicesSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListInstanceDevicesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListInstanceDevicesSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListInstanceDevicesResponse wrapper for the ListInstanceDevices operation
type ListInstanceDevicesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []Device instances
	Items []Device `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListInstanceDevicesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListInstanceDevicesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListInstanceDevicesSortByEnum Enum with underlying type: string
type ListInstanceDevicesSortByEnum string

// Set of constants representing the allowable values for ListInstanceDevicesSortByEnum
const (
	ListInstanceDevicesSortByTimecreated ListInstanceDevicesSortByEnum = "TIMECREATED"
	ListInstanceDevicesSortByDisplayname ListInstanceDevicesSortByEnum = "DISPLAYNAME"
)

var mappingListInstanceDevicesSortByEnum = map[string]ListInstanceDevicesSortByEnum{
	"TIMECREATED": ListInstanceDevicesSortByTimecreated,
	"DISPLAYNAME": ListInstanceDevicesSortByDisplayname,
}

var mappingListInstanceDevicesSortByEnumLowerCase = map[string]ListInstanceDevicesSortByEnum{
	"timecreated": ListInstanceDevicesSortByTimecreated,
	"displayname": ListInstanceDevicesSortByDisplayname,
}

// GetListInstanceDevicesSortByEnumValues Enumerates the set of values for ListInstanceDevicesSortByEnum
func GetListInstanceDevicesSortByEnumValues() []ListInstanceDevicesSortByEnum {
	values := make([]ListInstanceDevicesSortByEnum, 0)
	for _, v := range mappingListInstanceDevicesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListInstanceDevicesSortByEnumStringValues Enumerates the set of values in String for ListInstanceDevicesSortByEnum
func GetListInstanceDevicesSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListInstanceDevicesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListInstanceDevicesSortByEnum(val string) (ListInstanceDevicesSortByEnum, bool) {
	enum, ok := mappingListInstanceDevicesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListInstanceDevicesSortOrderEnum Enum with underlying type: string
type ListInstanceDevicesSortOrderEnum string

// Set of constants representing the allowable values for ListInstanceDevicesSortOrderEnum
const (
	ListInstanceDevicesSortOrderAsc  ListInstanceDevicesSortOrderEnum = "ASC"
	ListInstanceDevicesSortOrderDesc ListInstanceDevicesSortOrderEnum = "DESC"
)

var mappingListInstanceDevicesSortOrderEnum = map[string]ListInstanceDevicesSortOrderEnum{
	"ASC":  ListInstanceDevicesSortOrderAsc,
	"DESC": ListInstanceDevicesSortOrderDesc,
}

var mappingListInstanceDevicesSortOrderEnumLowerCase = map[string]ListInstanceDevicesSortOrderEnum{
	"asc":  ListInstanceDevicesSortOrderAsc,
	"desc": ListInstanceDevicesSortOrderDesc,
}

// GetListInstanceDevicesSortOrderEnumValues Enumerates the set of values for ListInstanceDevicesSortOrderEnum
func GetListInstanceDevicesSortOrderEnumValues() []ListInstanceDevicesSortOrderEnum {
	values := make([]ListInstanceDevicesSortOrderEnum, 0)
	for _, v := range mappingListInstanceDevicesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListInstanceDevicesSortOrderEnumStringValues Enumerates the set of values in String for ListInstanceDevicesSortOrderEnum
func GetListInstanceDevicesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListInstanceDevicesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListInstanceDevicesSortOrderEnum(val string) (ListInstanceDevicesSortOrderEnum, bool) {
	enum, ok := mappingListInstanceDevicesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
