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

// ListVolumeBackupsRequest wrapper for the ListVolumeBackups operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListVolumeBackups.go.html to see an example of how to use ListVolumeBackupsRequest.
type ListVolumeBackupsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID of the volume.
	VolumeId *string `mandatory:"false" contributesTo:"query" name:"volumeId"`

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

	// A filter to return only resources that originated from the given source volume backup.
	SourceVolumeBackupId *string `mandatory:"false" contributesTo:"query" name:"sourceVolumeBackupId"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListVolumeBackupsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListVolumeBackupsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle state. The state
	// value is case-insensitive.
	LifecycleState VolumeBackupLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListVolumeBackupsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListVolumeBackupsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListVolumeBackupsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListVolumeBackupsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListVolumeBackupsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListVolumeBackupsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListVolumeBackupsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListVolumeBackupsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListVolumeBackupsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVolumeBackupLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetVolumeBackupLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListVolumeBackupsResponse wrapper for the ListVolumeBackups operation
type ListVolumeBackupsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []VolumeBackup instances
	Items []VolumeBackup `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListVolumeBackupsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListVolumeBackupsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListVolumeBackupsSortByEnum Enum with underlying type: string
type ListVolumeBackupsSortByEnum string

// Set of constants representing the allowable values for ListVolumeBackupsSortByEnum
const (
	ListVolumeBackupsSortByTimecreated ListVolumeBackupsSortByEnum = "TIMECREATED"
	ListVolumeBackupsSortByDisplayname ListVolumeBackupsSortByEnum = "DISPLAYNAME"
)

var mappingListVolumeBackupsSortByEnum = map[string]ListVolumeBackupsSortByEnum{
	"TIMECREATED": ListVolumeBackupsSortByTimecreated,
	"DISPLAYNAME": ListVolumeBackupsSortByDisplayname,
}

var mappingListVolumeBackupsSortByEnumLowerCase = map[string]ListVolumeBackupsSortByEnum{
	"timecreated": ListVolumeBackupsSortByTimecreated,
	"displayname": ListVolumeBackupsSortByDisplayname,
}

// GetListVolumeBackupsSortByEnumValues Enumerates the set of values for ListVolumeBackupsSortByEnum
func GetListVolumeBackupsSortByEnumValues() []ListVolumeBackupsSortByEnum {
	values := make([]ListVolumeBackupsSortByEnum, 0)
	for _, v := range mappingListVolumeBackupsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListVolumeBackupsSortByEnumStringValues Enumerates the set of values in String for ListVolumeBackupsSortByEnum
func GetListVolumeBackupsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListVolumeBackupsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVolumeBackupsSortByEnum(val string) (ListVolumeBackupsSortByEnum, bool) {
	enum, ok := mappingListVolumeBackupsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListVolumeBackupsSortOrderEnum Enum with underlying type: string
type ListVolumeBackupsSortOrderEnum string

// Set of constants representing the allowable values for ListVolumeBackupsSortOrderEnum
const (
	ListVolumeBackupsSortOrderAsc  ListVolumeBackupsSortOrderEnum = "ASC"
	ListVolumeBackupsSortOrderDesc ListVolumeBackupsSortOrderEnum = "DESC"
)

var mappingListVolumeBackupsSortOrderEnum = map[string]ListVolumeBackupsSortOrderEnum{
	"ASC":  ListVolumeBackupsSortOrderAsc,
	"DESC": ListVolumeBackupsSortOrderDesc,
}

var mappingListVolumeBackupsSortOrderEnumLowerCase = map[string]ListVolumeBackupsSortOrderEnum{
	"asc":  ListVolumeBackupsSortOrderAsc,
	"desc": ListVolumeBackupsSortOrderDesc,
}

// GetListVolumeBackupsSortOrderEnumValues Enumerates the set of values for ListVolumeBackupsSortOrderEnum
func GetListVolumeBackupsSortOrderEnumValues() []ListVolumeBackupsSortOrderEnum {
	values := make([]ListVolumeBackupsSortOrderEnum, 0)
	for _, v := range mappingListVolumeBackupsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListVolumeBackupsSortOrderEnumStringValues Enumerates the set of values in String for ListVolumeBackupsSortOrderEnum
func GetListVolumeBackupsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListVolumeBackupsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListVolumeBackupsSortOrderEnum(val string) (ListVolumeBackupsSortOrderEnum, bool) {
	enum, ok := mappingListVolumeBackupsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
