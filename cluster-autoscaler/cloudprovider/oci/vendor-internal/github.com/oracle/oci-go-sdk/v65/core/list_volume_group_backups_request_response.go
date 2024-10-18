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

// ListVolumeGroupBackupsRequest wrapper for the ListVolumeGroupBackups operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListVolumeGroupBackups.go.html to see an example of how to use ListVolumeGroupBackupsRequest.
type ListVolumeGroupBackupsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID of the volume group.
	VolumeGroupId *string `mandatory:"false" contributesTo:"query" name:"volumeGroupId"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListVolumeGroupBackupsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListVolumeGroupBackupsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListVolumeGroupBackupsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListVolumeGroupBackupsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListVolumeGroupBackupsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListVolumeGroupBackupsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListVolumeGroupBackupsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListVolumeGroupBackupsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListVolumeGroupBackupsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListVolumeGroupBackupsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListVolumeGroupBackupsSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListVolumeGroupBackupsResponse wrapper for the ListVolumeGroupBackups operation
type ListVolumeGroupBackupsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []VolumeGroupBackup instances
	Items []VolumeGroupBackup `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListVolumeGroupBackupsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListVolumeGroupBackupsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListVolumeGroupBackupsSortByEnum Enum with underlying type: string
type ListVolumeGroupBackupsSortByEnum string

// Set of constants representing the allowable values for ListVolumeGroupBackupsSortByEnum
const (
	ListVolumeGroupBackupsSortByTimecreated ListVolumeGroupBackupsSortByEnum = "TIMECREATED"
	ListVolumeGroupBackupsSortByDisplayname ListVolumeGroupBackupsSortByEnum = "DISPLAYNAME"
)

var mappingListVolumeGroupBackupsSortByEnum = map[string]ListVolumeGroupBackupsSortByEnum{
	"TIMECREATED": ListVolumeGroupBackupsSortByTimecreated,
	"DISPLAYNAME": ListVolumeGroupBackupsSortByDisplayname,
}

var mappingListVolumeGroupBackupsSortByEnumLowerCase = map[string]ListVolumeGroupBackupsSortByEnum{
	"timecreated": ListVolumeGroupBackupsSortByTimecreated,
	"displayname": ListVolumeGroupBackupsSortByDisplayname,
}

// GetListVolumeGroupBackupsSortByEnumValues Enumerates the set of values for ListVolumeGroupBackupsSortByEnum
func GetListVolumeGroupBackupsSortByEnumValues() []ListVolumeGroupBackupsSortByEnum {
	values := make([]ListVolumeGroupBackupsSortByEnum, 0)
	for _, v := range mappingListVolumeGroupBackupsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListVolumeGroupBackupsSortByEnumStringValues Enumerates the set of values in String for ListVolumeGroupBackupsSortByEnum
func GetListVolumeGroupBackupsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListVolumeGroupBackupsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVolumeGroupBackupsSortByEnum(val string) (ListVolumeGroupBackupsSortByEnum, bool) {
	enum, ok := mappingListVolumeGroupBackupsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListVolumeGroupBackupsSortOrderEnum Enum with underlying type: string
type ListVolumeGroupBackupsSortOrderEnum string

// Set of constants representing the allowable values for ListVolumeGroupBackupsSortOrderEnum
const (
	ListVolumeGroupBackupsSortOrderAsc  ListVolumeGroupBackupsSortOrderEnum = "ASC"
	ListVolumeGroupBackupsSortOrderDesc ListVolumeGroupBackupsSortOrderEnum = "DESC"
)

var mappingListVolumeGroupBackupsSortOrderEnum = map[string]ListVolumeGroupBackupsSortOrderEnum{
	"ASC":  ListVolumeGroupBackupsSortOrderAsc,
	"DESC": ListVolumeGroupBackupsSortOrderDesc,
}

var mappingListVolumeGroupBackupsSortOrderEnumLowerCase = map[string]ListVolumeGroupBackupsSortOrderEnum{
	"asc":  ListVolumeGroupBackupsSortOrderAsc,
	"desc": ListVolumeGroupBackupsSortOrderDesc,
}

// GetListVolumeGroupBackupsSortOrderEnumValues Enumerates the set of values for ListVolumeGroupBackupsSortOrderEnum
func GetListVolumeGroupBackupsSortOrderEnumValues() []ListVolumeGroupBackupsSortOrderEnum {
	values := make([]ListVolumeGroupBackupsSortOrderEnum, 0)
	for _, v := range mappingListVolumeGroupBackupsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListVolumeGroupBackupsSortOrderEnumStringValues Enumerates the set of values in String for ListVolumeGroupBackupsSortOrderEnum
func GetListVolumeGroupBackupsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListVolumeGroupBackupsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVolumeGroupBackupsSortOrderEnum(val string) (ListVolumeGroupBackupsSortOrderEnum, bool) {
	enum, ok := mappingListVolumeGroupBackupsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
