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

// ListNetworkSecurityGroupSecurityRulesRequest wrapper for the ListNetworkSecurityGroupSecurityRules operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListNetworkSecurityGroupSecurityRules.go.html to see an example of how to use ListNetworkSecurityGroupSecurityRulesRequest.
type ListNetworkSecurityGroupSecurityRulesRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the network security group.
	NetworkSecurityGroupId *string `mandatory:"true" contributesTo:"path" name:"networkSecurityGroupId"`

	// Direction of the security rule. Set to `EGRESS` for rules that allow outbound IP packets,
	// or `INGRESS` for rules that allow inbound IP packets.
	Direction ListNetworkSecurityGroupSecurityRulesDirectionEnum `mandatory:"false" contributesTo:"query" name:"direction" omitEmpty:"true"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The field to sort by.
	SortBy ListNetworkSecurityGroupSecurityRulesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListNetworkSecurityGroupSecurityRulesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListNetworkSecurityGroupSecurityRulesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListNetworkSecurityGroupSecurityRulesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListNetworkSecurityGroupSecurityRulesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListNetworkSecurityGroupSecurityRulesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListNetworkSecurityGroupSecurityRulesRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListNetworkSecurityGroupSecurityRulesDirectionEnum(string(request.Direction)); !ok && request.Direction != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Direction: %s. Supported values are: %s.", request.Direction, strings.Join(GetListNetworkSecurityGroupSecurityRulesDirectionEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListNetworkSecurityGroupSecurityRulesSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListNetworkSecurityGroupSecurityRulesSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListNetworkSecurityGroupSecurityRulesSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListNetworkSecurityGroupSecurityRulesSortOrderEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListNetworkSecurityGroupSecurityRulesResponse wrapper for the ListNetworkSecurityGroupSecurityRules operation
type ListNetworkSecurityGroupSecurityRulesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []SecurityRule instances
	Items []SecurityRule `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListNetworkSecurityGroupSecurityRulesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListNetworkSecurityGroupSecurityRulesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListNetworkSecurityGroupSecurityRulesDirectionEnum Enum with underlying type: string
type ListNetworkSecurityGroupSecurityRulesDirectionEnum string

// Set of constants representing the allowable values for ListNetworkSecurityGroupSecurityRulesDirectionEnum
const (
	ListNetworkSecurityGroupSecurityRulesDirectionEgress  ListNetworkSecurityGroupSecurityRulesDirectionEnum = "EGRESS"
	ListNetworkSecurityGroupSecurityRulesDirectionIngress ListNetworkSecurityGroupSecurityRulesDirectionEnum = "INGRESS"
)

var mappingListNetworkSecurityGroupSecurityRulesDirectionEnum = map[string]ListNetworkSecurityGroupSecurityRulesDirectionEnum{
	"EGRESS":  ListNetworkSecurityGroupSecurityRulesDirectionEgress,
	"INGRESS": ListNetworkSecurityGroupSecurityRulesDirectionIngress,
}

var mappingListNetworkSecurityGroupSecurityRulesDirectionEnumLowerCase = map[string]ListNetworkSecurityGroupSecurityRulesDirectionEnum{
	"egress":  ListNetworkSecurityGroupSecurityRulesDirectionEgress,
	"ingress": ListNetworkSecurityGroupSecurityRulesDirectionIngress,
}

// GetListNetworkSecurityGroupSecurityRulesDirectionEnumValues Enumerates the set of values for ListNetworkSecurityGroupSecurityRulesDirectionEnum
func GetListNetworkSecurityGroupSecurityRulesDirectionEnumValues() []ListNetworkSecurityGroupSecurityRulesDirectionEnum {
	values := make([]ListNetworkSecurityGroupSecurityRulesDirectionEnum, 0)
	for _, v := range mappingListNetworkSecurityGroupSecurityRulesDirectionEnum {
		values = append(values, v)
	}
	return values
}

// GetListNetworkSecurityGroupSecurityRulesDirectionEnumStringValues Enumerates the set of values in String for ListNetworkSecurityGroupSecurityRulesDirectionEnum
func GetListNetworkSecurityGroupSecurityRulesDirectionEnumStringValues() []string {
	return []string{
		"EGRESS",
		"INGRESS",
	}
}

// GetMappingListNetworkSecurityGroupSecurityRulesDirectionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListNetworkSecurityGroupSecurityRulesDirectionEnum(val string) (ListNetworkSecurityGroupSecurityRulesDirectionEnum, bool) {
	enum, ok := mappingListNetworkSecurityGroupSecurityRulesDirectionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListNetworkSecurityGroupSecurityRulesSortByEnum Enum with underlying type: string
type ListNetworkSecurityGroupSecurityRulesSortByEnum string

// Set of constants representing the allowable values for ListNetworkSecurityGroupSecurityRulesSortByEnum
const (
	ListNetworkSecurityGroupSecurityRulesSortByTimecreated ListNetworkSecurityGroupSecurityRulesSortByEnum = "TIMECREATED"
)

var mappingListNetworkSecurityGroupSecurityRulesSortByEnum = map[string]ListNetworkSecurityGroupSecurityRulesSortByEnum{
	"TIMECREATED": ListNetworkSecurityGroupSecurityRulesSortByTimecreated,
}

var mappingListNetworkSecurityGroupSecurityRulesSortByEnumLowerCase = map[string]ListNetworkSecurityGroupSecurityRulesSortByEnum{
	"timecreated": ListNetworkSecurityGroupSecurityRulesSortByTimecreated,
}

// GetListNetworkSecurityGroupSecurityRulesSortByEnumValues Enumerates the set of values for ListNetworkSecurityGroupSecurityRulesSortByEnum
func GetListNetworkSecurityGroupSecurityRulesSortByEnumValues() []ListNetworkSecurityGroupSecurityRulesSortByEnum {
	values := make([]ListNetworkSecurityGroupSecurityRulesSortByEnum, 0)
	for _, v := range mappingListNetworkSecurityGroupSecurityRulesSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListNetworkSecurityGroupSecurityRulesSortByEnumStringValues Enumerates the set of values in String for ListNetworkSecurityGroupSecurityRulesSortByEnum
func GetListNetworkSecurityGroupSecurityRulesSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
	}
}

// GetMappingListNetworkSecurityGroupSecurityRulesSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListNetworkSecurityGroupSecurityRulesSortByEnum(val string) (ListNetworkSecurityGroupSecurityRulesSortByEnum, bool) {
	enum, ok := mappingListNetworkSecurityGroupSecurityRulesSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListNetworkSecurityGroupSecurityRulesSortOrderEnum Enum with underlying type: string
type ListNetworkSecurityGroupSecurityRulesSortOrderEnum string

// Set of constants representing the allowable values for ListNetworkSecurityGroupSecurityRulesSortOrderEnum
const (
	ListNetworkSecurityGroupSecurityRulesSortOrderAsc  ListNetworkSecurityGroupSecurityRulesSortOrderEnum = "ASC"
	ListNetworkSecurityGroupSecurityRulesSortOrderDesc ListNetworkSecurityGroupSecurityRulesSortOrderEnum = "DESC"
)

var mappingListNetworkSecurityGroupSecurityRulesSortOrderEnum = map[string]ListNetworkSecurityGroupSecurityRulesSortOrderEnum{
	"ASC":  ListNetworkSecurityGroupSecurityRulesSortOrderAsc,
	"DESC": ListNetworkSecurityGroupSecurityRulesSortOrderDesc,
}

var mappingListNetworkSecurityGroupSecurityRulesSortOrderEnumLowerCase = map[string]ListNetworkSecurityGroupSecurityRulesSortOrderEnum{
	"asc":  ListNetworkSecurityGroupSecurityRulesSortOrderAsc,
	"desc": ListNetworkSecurityGroupSecurityRulesSortOrderDesc,
}

// GetListNetworkSecurityGroupSecurityRulesSortOrderEnumValues Enumerates the set of values for ListNetworkSecurityGroupSecurityRulesSortOrderEnum
func GetListNetworkSecurityGroupSecurityRulesSortOrderEnumValues() []ListNetworkSecurityGroupSecurityRulesSortOrderEnum {
	values := make([]ListNetworkSecurityGroupSecurityRulesSortOrderEnum, 0)
	for _, v := range mappingListNetworkSecurityGroupSecurityRulesSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListNetworkSecurityGroupSecurityRulesSortOrderEnumStringValues Enumerates the set of values in String for ListNetworkSecurityGroupSecurityRulesSortOrderEnum
func GetListNetworkSecurityGroupSecurityRulesSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListNetworkSecurityGroupSecurityRulesSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListNetworkSecurityGroupSecurityRulesSortOrderEnum(val string) (ListNetworkSecurityGroupSecurityRulesSortOrderEnum, bool) {
	enum, ok := mappingListNetworkSecurityGroupSecurityRulesSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
