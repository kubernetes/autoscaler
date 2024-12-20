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

// ListBootVolumeReplicasRequest wrapper for the ListBootVolumeReplicas operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListBootVolumeReplicas.go.html to see an example of how to use ListBootVolumeReplicasRequest.
type ListBootVolumeReplicasRequest struct {

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// The OCID of the volume group replica.
	VolumeGroupReplicaId *string `mandatory:"false" contributesTo:"query" name:"volumeGroupReplicaId"`

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
	SortBy ListBootVolumeReplicasSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListBootVolumeReplicasSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle state. The state value is case-insensitive.
	LifecycleState BootVolumeReplicaLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListBootVolumeReplicasRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListBootVolumeReplicasRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListBootVolumeReplicasRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListBootVolumeReplicasRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListBootVolumeReplicasRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListBootVolumeReplicasSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListBootVolumeReplicasSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListBootVolumeReplicasSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListBootVolumeReplicasSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingBootVolumeReplicaLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetBootVolumeReplicaLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListBootVolumeReplicasResponse wrapper for the ListBootVolumeReplicas operation
type ListBootVolumeReplicasResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []BootVolumeReplica instances
	Items []BootVolumeReplica `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListBootVolumeReplicasResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListBootVolumeReplicasResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListBootVolumeReplicasSortByEnum Enum with underlying type: string
type ListBootVolumeReplicasSortByEnum string

// Set of constants representing the allowable values for ListBootVolumeReplicasSortByEnum
const (
	ListBootVolumeReplicasSortByTimecreated ListBootVolumeReplicasSortByEnum = "TIMECREATED"
	ListBootVolumeReplicasSortByDisplayname ListBootVolumeReplicasSortByEnum = "DISPLAYNAME"
)

var mappingListBootVolumeReplicasSortByEnum = map[string]ListBootVolumeReplicasSortByEnum{
	"TIMECREATED": ListBootVolumeReplicasSortByTimecreated,
	"DISPLAYNAME": ListBootVolumeReplicasSortByDisplayname,
}

var mappingListBootVolumeReplicasSortByEnumLowerCase = map[string]ListBootVolumeReplicasSortByEnum{
	"timecreated": ListBootVolumeReplicasSortByTimecreated,
	"displayname": ListBootVolumeReplicasSortByDisplayname,
}

// GetListBootVolumeReplicasSortByEnumValues Enumerates the set of values for ListBootVolumeReplicasSortByEnum
func GetListBootVolumeReplicasSortByEnumValues() []ListBootVolumeReplicasSortByEnum {
	values := make([]ListBootVolumeReplicasSortByEnum, 0)
	for _, v := range mappingListBootVolumeReplicasSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListBootVolumeReplicasSortByEnumStringValues Enumerates the set of values in String for ListBootVolumeReplicasSortByEnum
func GetListBootVolumeReplicasSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListBootVolumeReplicasSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListBootVolumeReplicasSortByEnum(val string) (ListBootVolumeReplicasSortByEnum, bool) {
	enum, ok := mappingListBootVolumeReplicasSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListBootVolumeReplicasSortOrderEnum Enum with underlying type: string
type ListBootVolumeReplicasSortOrderEnum string

// Set of constants representing the allowable values for ListBootVolumeReplicasSortOrderEnum
const (
	ListBootVolumeReplicasSortOrderAsc  ListBootVolumeReplicasSortOrderEnum = "ASC"
	ListBootVolumeReplicasSortOrderDesc ListBootVolumeReplicasSortOrderEnum = "DESC"
)

var mappingListBootVolumeReplicasSortOrderEnum = map[string]ListBootVolumeReplicasSortOrderEnum{
	"ASC":  ListBootVolumeReplicasSortOrderAsc,
	"DESC": ListBootVolumeReplicasSortOrderDesc,
}

var mappingListBootVolumeReplicasSortOrderEnumLowerCase = map[string]ListBootVolumeReplicasSortOrderEnum{
	"asc":  ListBootVolumeReplicasSortOrderAsc,
	"desc": ListBootVolumeReplicasSortOrderDesc,
}

// GetListBootVolumeReplicasSortOrderEnumValues Enumerates the set of values for ListBootVolumeReplicasSortOrderEnum
func GetListBootVolumeReplicasSortOrderEnumValues() []ListBootVolumeReplicasSortOrderEnum {
	values := make([]ListBootVolumeReplicasSortOrderEnum, 0)
	for _, v := range mappingListBootVolumeReplicasSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListBootVolumeReplicasSortOrderEnumStringValues Enumerates the set of values in String for ListBootVolumeReplicasSortOrderEnum
func GetListBootVolumeReplicasSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListBootVolumeReplicasSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListBootVolumeReplicasSortOrderEnum(val string) (ListBootVolumeReplicasSortOrderEnum, bool) {
	enum, ok := mappingListBootVolumeReplicasSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
