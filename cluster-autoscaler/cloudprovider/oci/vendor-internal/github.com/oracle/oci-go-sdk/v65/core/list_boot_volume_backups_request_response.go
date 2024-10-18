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

// ListBootVolumeBackupsRequest wrapper for the ListBootVolumeBackups operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListBootVolumeBackups.go.html to see an example of how to use ListBootVolumeBackupsRequest.
type ListBootVolumeBackupsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID of the boot volume.
	BootVolumeId *string `mandatory:"false" contributesTo:"query" name:"bootVolumeId"`

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

	// A filter to return only resources that originated from the given source boot volume backup.
	SourceBootVolumeBackupId *string `mandatory:"false" contributesTo:"query" name:"sourceBootVolumeBackupId"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListBootVolumeBackupsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListBootVolumeBackupsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle state. The state value is
	// case-insensitive.
	LifecycleState BootVolumeBackupLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListBootVolumeBackupsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListBootVolumeBackupsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListBootVolumeBackupsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListBootVolumeBackupsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListBootVolumeBackupsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListBootVolumeBackupsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListBootVolumeBackupsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListBootVolumeBackupsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListBootVolumeBackupsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingBootVolumeBackupLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetBootVolumeBackupLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListBootVolumeBackupsResponse wrapper for the ListBootVolumeBackups operation
type ListBootVolumeBackupsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []BootVolumeBackup instances
	Items []BootVolumeBackup `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListBootVolumeBackupsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListBootVolumeBackupsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListBootVolumeBackupsSortByEnum Enum with underlying type: string
type ListBootVolumeBackupsSortByEnum string

// Set of constants representing the allowable values for ListBootVolumeBackupsSortByEnum
const (
	ListBootVolumeBackupsSortByTimecreated ListBootVolumeBackupsSortByEnum = "TIMECREATED"
	ListBootVolumeBackupsSortByDisplayname ListBootVolumeBackupsSortByEnum = "DISPLAYNAME"
)

var mappingListBootVolumeBackupsSortByEnum = map[string]ListBootVolumeBackupsSortByEnum{
	"TIMECREATED": ListBootVolumeBackupsSortByTimecreated,
	"DISPLAYNAME": ListBootVolumeBackupsSortByDisplayname,
}

var mappingListBootVolumeBackupsSortByEnumLowerCase = map[string]ListBootVolumeBackupsSortByEnum{
	"timecreated": ListBootVolumeBackupsSortByTimecreated,
	"displayname": ListBootVolumeBackupsSortByDisplayname,
}

// GetListBootVolumeBackupsSortByEnumValues Enumerates the set of values for ListBootVolumeBackupsSortByEnum
func GetListBootVolumeBackupsSortByEnumValues() []ListBootVolumeBackupsSortByEnum {
	values := make([]ListBootVolumeBackupsSortByEnum, 0)
	for _, v := range mappingListBootVolumeBackupsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListBootVolumeBackupsSortByEnumStringValues Enumerates the set of values in String for ListBootVolumeBackupsSortByEnum
func GetListBootVolumeBackupsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListBootVolumeBackupsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListBootVolumeBackupsSortByEnum(val string) (ListBootVolumeBackupsSortByEnum, bool) {
	enum, ok := mappingListBootVolumeBackupsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListBootVolumeBackupsSortOrderEnum Enum with underlying type: string
type ListBootVolumeBackupsSortOrderEnum string

// Set of constants representing the allowable values for ListBootVolumeBackupsSortOrderEnum
const (
	ListBootVolumeBackupsSortOrderAsc  ListBootVolumeBackupsSortOrderEnum = "ASC"
	ListBootVolumeBackupsSortOrderDesc ListBootVolumeBackupsSortOrderEnum = "DESC"
)

var mappingListBootVolumeBackupsSortOrderEnum = map[string]ListBootVolumeBackupsSortOrderEnum{
	"ASC":  ListBootVolumeBackupsSortOrderAsc,
	"DESC": ListBootVolumeBackupsSortOrderDesc,
}

var mappingListBootVolumeBackupsSortOrderEnumLowerCase = map[string]ListBootVolumeBackupsSortOrderEnum{
	"asc":  ListBootVolumeBackupsSortOrderAsc,
	"desc": ListBootVolumeBackupsSortOrderDesc,
}

// GetListBootVolumeBackupsSortOrderEnumValues Enumerates the set of values for ListBootVolumeBackupsSortOrderEnum
func GetListBootVolumeBackupsSortOrderEnumValues() []ListBootVolumeBackupsSortOrderEnum {
	values := make([]ListBootVolumeBackupsSortOrderEnum, 0)
	for _, v := range mappingListBootVolumeBackupsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListBootVolumeBackupsSortOrderEnumStringValues Enumerates the set of values in String for ListBootVolumeBackupsSortOrderEnum
func GetListBootVolumeBackupsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListBootVolumeBackupsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListBootVolumeBackupsSortOrderEnum(val string) (ListBootVolumeBackupsSortOrderEnum, bool) {
	enum, ok := mappingListBootVolumeBackupsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
