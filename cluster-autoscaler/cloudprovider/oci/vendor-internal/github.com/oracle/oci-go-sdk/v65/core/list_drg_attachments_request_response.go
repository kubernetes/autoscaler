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

// ListDrgAttachmentsRequest wrapper for the ListDrgAttachments operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListDrgAttachments.go.html to see an example of how to use ListDrgAttachmentsRequest.
type ListDrgAttachmentsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN.
	VcnId *string `mandatory:"false" contributesTo:"query" name:"vcnId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DRG.
	DrgId *string `mandatory:"false" contributesTo:"query" name:"drgId"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the resource (virtual circuit, VCN, IPSec tunnel, or remote peering connection) attached to the DRG.
	NetworkId *string `mandatory:"false" contributesTo:"query" name:"networkId"`

	// The type for the network resource attached to the DRG.
	AttachmentType ListDrgAttachmentsAttachmentTypeEnum `mandatory:"false" contributesTo:"query" name:"attachmentType" omitEmpty:"true"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DRG route table assigned to the DRG attachment.
	DrgRouteTableId *string `mandatory:"false" contributesTo:"query" name:"drgRouteTableId"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListDrgAttachmentsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListDrgAttachmentsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only resources that match the specified lifecycle
	// state. The value is case insensitive.
	LifecycleState DrgAttachmentLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListDrgAttachmentsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListDrgAttachmentsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListDrgAttachmentsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListDrgAttachmentsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListDrgAttachmentsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListDrgAttachmentsAttachmentTypeEnum(string(request.AttachmentType)); !ok && request.AttachmentType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AttachmentType: %s. Supported values are: %s.", request.AttachmentType, strings.Join(GetListDrgAttachmentsAttachmentTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListDrgAttachmentsSortByEnum(string(request.SortBy)); !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListDrgAttachmentsSortByEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListDrgAttachmentsSortOrderEnum(string(request.SortOrder)); !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListDrgAttachmentsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := GetMappingDrgAttachmentLifecycleStateEnum(string(request.LifecycleState)); !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetDrgAttachmentLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListDrgAttachmentsResponse wrapper for the ListDrgAttachments operation
type ListDrgAttachmentsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []DrgAttachment instances
	Items []DrgAttachment `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListDrgAttachmentsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListDrgAttachmentsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListDrgAttachmentsAttachmentTypeEnum Enum with underlying type: string
type ListDrgAttachmentsAttachmentTypeEnum string

// Set of constants representing the allowable values for ListDrgAttachmentsAttachmentTypeEnum
const (
	ListDrgAttachmentsAttachmentTypeVcn                     ListDrgAttachmentsAttachmentTypeEnum = "VCN"
	ListDrgAttachmentsAttachmentTypeVirtualCircuit          ListDrgAttachmentsAttachmentTypeEnum = "VIRTUAL_CIRCUIT"
	ListDrgAttachmentsAttachmentTypeRemotePeeringConnection ListDrgAttachmentsAttachmentTypeEnum = "REMOTE_PEERING_CONNECTION"
	ListDrgAttachmentsAttachmentTypeIpsecTunnel             ListDrgAttachmentsAttachmentTypeEnum = "IPSEC_TUNNEL"
	ListDrgAttachmentsAttachmentTypeAll                     ListDrgAttachmentsAttachmentTypeEnum = "ALL"
)

var mappingListDrgAttachmentsAttachmentTypeEnum = map[string]ListDrgAttachmentsAttachmentTypeEnum{
	"VCN":                       ListDrgAttachmentsAttachmentTypeVcn,
	"VIRTUAL_CIRCUIT":           ListDrgAttachmentsAttachmentTypeVirtualCircuit,
	"REMOTE_PEERING_CONNECTION": ListDrgAttachmentsAttachmentTypeRemotePeeringConnection,
	"IPSEC_TUNNEL":              ListDrgAttachmentsAttachmentTypeIpsecTunnel,
	"ALL":                       ListDrgAttachmentsAttachmentTypeAll,
}

var mappingListDrgAttachmentsAttachmentTypeEnumLowerCase = map[string]ListDrgAttachmentsAttachmentTypeEnum{
	"vcn":                       ListDrgAttachmentsAttachmentTypeVcn,
	"virtual_circuit":           ListDrgAttachmentsAttachmentTypeVirtualCircuit,
	"remote_peering_connection": ListDrgAttachmentsAttachmentTypeRemotePeeringConnection,
	"ipsec_tunnel":              ListDrgAttachmentsAttachmentTypeIpsecTunnel,
	"all":                       ListDrgAttachmentsAttachmentTypeAll,
}

// GetListDrgAttachmentsAttachmentTypeEnumValues Enumerates the set of values for ListDrgAttachmentsAttachmentTypeEnum
func GetListDrgAttachmentsAttachmentTypeEnumValues() []ListDrgAttachmentsAttachmentTypeEnum {
	values := make([]ListDrgAttachmentsAttachmentTypeEnum, 0)
	for _, v := range mappingListDrgAttachmentsAttachmentTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetListDrgAttachmentsAttachmentTypeEnumStringValues Enumerates the set of values in String for ListDrgAttachmentsAttachmentTypeEnum
func GetListDrgAttachmentsAttachmentTypeEnumStringValues() []string {
	return []string{
		"VCN",
		"VIRTUAL_CIRCUIT",
		"REMOTE_PEERING_CONNECTION",
		"IPSEC_TUNNEL",
		"ALL",
	}
}

// GetMappingListDrgAttachmentsAttachmentTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListDrgAttachmentsAttachmentTypeEnum(val string) (ListDrgAttachmentsAttachmentTypeEnum, bool) {
	enum, ok := mappingListDrgAttachmentsAttachmentTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListDrgAttachmentsSortByEnum Enum with underlying type: string
type ListDrgAttachmentsSortByEnum string

// Set of constants representing the allowable values for ListDrgAttachmentsSortByEnum
const (
	ListDrgAttachmentsSortByTimecreated ListDrgAttachmentsSortByEnum = "TIMECREATED"
	ListDrgAttachmentsSortByDisplayname ListDrgAttachmentsSortByEnum = "DISPLAYNAME"
)

var mappingListDrgAttachmentsSortByEnum = map[string]ListDrgAttachmentsSortByEnum{
	"TIMECREATED": ListDrgAttachmentsSortByTimecreated,
	"DISPLAYNAME": ListDrgAttachmentsSortByDisplayname,
}

var mappingListDrgAttachmentsSortByEnumLowerCase = map[string]ListDrgAttachmentsSortByEnum{
	"timecreated": ListDrgAttachmentsSortByTimecreated,
	"displayname": ListDrgAttachmentsSortByDisplayname,
}

// GetListDrgAttachmentsSortByEnumValues Enumerates the set of values for ListDrgAttachmentsSortByEnum
func GetListDrgAttachmentsSortByEnumValues() []ListDrgAttachmentsSortByEnum {
	values := make([]ListDrgAttachmentsSortByEnum, 0)
	for _, v := range mappingListDrgAttachmentsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListDrgAttachmentsSortByEnumStringValues Enumerates the set of values in String for ListDrgAttachmentsSortByEnum
func GetListDrgAttachmentsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// GetMappingListDrgAttachmentsSortByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListDrgAttachmentsSortByEnum(val string) (ListDrgAttachmentsSortByEnum, bool) {
	enum, ok := mappingListDrgAttachmentsSortByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListDrgAttachmentsSortOrderEnum Enum with underlying type: string
type ListDrgAttachmentsSortOrderEnum string

// Set of constants representing the allowable values for ListDrgAttachmentsSortOrderEnum
const (
	ListDrgAttachmentsSortOrderAsc  ListDrgAttachmentsSortOrderEnum = "ASC"
	ListDrgAttachmentsSortOrderDesc ListDrgAttachmentsSortOrderEnum = "DESC"
)

var mappingListDrgAttachmentsSortOrderEnum = map[string]ListDrgAttachmentsSortOrderEnum{
	"ASC":  ListDrgAttachmentsSortOrderAsc,
	"DESC": ListDrgAttachmentsSortOrderDesc,
}

var mappingListDrgAttachmentsSortOrderEnumLowerCase = map[string]ListDrgAttachmentsSortOrderEnum{
	"asc":  ListDrgAttachmentsSortOrderAsc,
	"desc": ListDrgAttachmentsSortOrderDesc,
}

// GetListDrgAttachmentsSortOrderEnumValues Enumerates the set of values for ListDrgAttachmentsSortOrderEnum
func GetListDrgAttachmentsSortOrderEnumValues() []ListDrgAttachmentsSortOrderEnum {
	values := make([]ListDrgAttachmentsSortOrderEnum, 0)
	for _, v := range mappingListDrgAttachmentsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListDrgAttachmentsSortOrderEnumStringValues Enumerates the set of values in String for ListDrgAttachmentsSortOrderEnum
func GetListDrgAttachmentsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}

// GetMappingListDrgAttachmentsSortOrderEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListDrgAttachmentsSortOrderEnum(val string) (ListDrgAttachmentsSortOrderEnum, bool) {
	enum, ok := mappingListDrgAttachmentsSortOrderEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
