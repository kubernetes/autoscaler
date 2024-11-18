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

// ListWorkloadMappingsRequest wrapper for the ListWorkloadMappings operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListWorkloadMappings.go.html to see an example of how to use ListWorkloadMappingsRequest.
type ListWorkloadMappingsRequest struct {

	// The OCID of the cluster.
	ClusterId *string `mandatory:"true" contributesTo:"path" name:"clusterId"`

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
	SortOrder ListWorkloadMappingsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the results by.
	SortBy ListWorkloadMappingsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListWorkloadMappingsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListWorkloadMappingsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListWorkloadMappingsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListWorkloadMappingsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListWorkloadMappingsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListWorkloadMappingsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListWorkloadMappingsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListWorkloadMappingsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListWorkloadMappingsSortByEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListWorkloadMappingsResponse wrapper for the ListWorkloadMappings operation
type ListWorkloadMappingsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []WorkloadMappingSummary instances
	Items []WorkloadMappingSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of results remain.
	// For important details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListWorkloadMappingsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListWorkloadMappingsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListWorkloadMappingsSortOrderEnum Enum with underlying type: string
type ListWorkloadMappingsSortOrderEnum string

// Set of constants representing the allowable values for ListWorkloadMappingsSortOrderEnum
const (
	ListWorkloadMappingsSortOrderAsc  ListWorkloadMappingsSortOrderEnum = "ASC"
	ListWorkloadMappingsSortOrderDesc ListWorkloadMappingsSortOrderEnum = "DESC"
)

var mappingListWorkloadMappingsSortOrderEnum = map[string]ListWorkloadMappingsSortOrderEnum{
	"ASC":  ListWorkloadMappingsSortOrderAsc,
	"DESC": ListWorkloadMappingsSortOrderDesc,
}

var mappingListWorkloadMappingsSortOrderEnumLowerCase = map[string]ListWorkloadMappingsSortOrderEnum{
	"asc":  ListWorkloadMappingsSortOrderAsc,
	"desc": ListWorkloadMappingsSortOrderDesc,
}

// GetListWorkloadMappingsSortOrderEnumValues Enumerates the set of values for ListWorkloadMappingsSortOrderEnum
func GetListWorkloadMappingsSortOrderEnumValues() []ListWorkloadMappingsSortOrderEnum {
	values := make([]ListWorkloadMappingsSortOrderEnum, 0)
	for _, v := range mappingListWorkloadMappingsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListWorkloadMappingsSortOrderEnumStringValues Enumerates the set of values in String for ListWorkloadMappingsSortOrderEnum
func GetListWorkloadMappingsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListWorkloadMappingsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListWorkloadMappingsSortOrderEnum(val string) (ListWorkloadMappingsSortOrderEnum, bool) {
	enum, ok := mappingListWorkloadMappingsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListWorkloadMappingsSortByEnum Enum with underlying type: string
type ListWorkloadMappingsSortByEnum string

// Set of constants representing the allowable values for ListWorkloadMappingsSortByEnum
const (
	ListWorkloadMappingsSortByNamespace   ListWorkloadMappingsSortByEnum = "NAMESPACE"
	ListWorkloadMappingsSortByTimecreated ListWorkloadMappingsSortByEnum = "TIMECREATED"
)

var mappingListWorkloadMappingsSortByEnum = map[string]ListWorkloadMappingsSortByEnum{
	"NAMESPACE":   ListWorkloadMappingsSortByNamespace,
	"TIMECREATED": ListWorkloadMappingsSortByTimecreated,
}

var mappingListWorkloadMappingsSortByEnumLowerCase = map[string]ListWorkloadMappingsSortByEnum{
	"namespace":   ListWorkloadMappingsSortByNamespace,
	"timecreated": ListWorkloadMappingsSortByTimecreated,
}

// GetListWorkloadMappingsSortByEnumValues Enumerates the set of values for ListWorkloadMappingsSortByEnum
func GetListWorkloadMappingsSortByEnumValues() []ListWorkloadMappingsSortByEnum {
	values := make([]ListWorkloadMappingsSortByEnum, 0)
	for _, v := range mappingListWorkloadMappingsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListWorkloadMappingsSortByEnumStringValues Enumerates the set of values in String for ListWorkloadMappingsSortByEnum
func GetListWorkloadMappingsSortByEnumStringValues() []string {
	return []string{
		"NAMESPACE",
		"TIMECREATED",
	}
}

// GetMappingListWorkloadMappingsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListWorkloadMappingsSortByEnum(val string) (ListWorkloadMappingsSortByEnum, bool) {
	enum, ok := mappingListWorkloadMappingsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
