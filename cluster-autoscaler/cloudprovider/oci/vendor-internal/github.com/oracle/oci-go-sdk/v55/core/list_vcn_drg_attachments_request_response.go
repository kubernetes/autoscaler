// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"net/http"
	"strings"
)

// ListVcnDrgAttachmentsRequest wrapper for the ListVcnDrgAttachments operation
type ListVcnDrgAttachmentsRequest struct {

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
	AttachmentType ListVcnDrgAttachmentsAttachmentTypeEnum `mandatory:"false" contributesTo:"query" name:"attachmentType" omitEmpty:"true"`

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
	SortBy ListVcnDrgAttachmentsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListVcnDrgAttachmentsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

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

func (request ListVcnDrgAttachmentsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListVcnDrgAttachmentsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListVcnDrgAttachmentsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListVcnDrgAttachmentsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListVcnDrgAttachmentsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingListVcnDrgAttachmentsAttachmentTypeEnum[string(request.AttachmentType)]; !ok && request.AttachmentType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AttachmentType: %s. Supported values are: %s.", request.AttachmentType, strings.Join(GetListVcnDrgAttachmentsAttachmentTypeEnumStringValues(), ",")))
	}
	if _, ok := mappingListVcnDrgAttachmentsSortByEnum[string(request.SortBy)]; !ok && request.SortBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortBy: %s. Supported values are: %s.", request.SortBy, strings.Join(GetListVcnDrgAttachmentsSortByEnumStringValues(), ",")))
	}
	if _, ok := mappingListVcnDrgAttachmentsSortOrderEnum[string(request.SortOrder)]; !ok && request.SortOrder != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SortOrder: %s. Supported values are: %s.", request.SortOrder, strings.Join(GetListVcnDrgAttachmentsSortOrderEnumStringValues(), ",")))
	}
	if _, ok := mappingDrgAttachmentLifecycleStateEnum[string(request.LifecycleState)]; !ok && request.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", request.LifecycleState, strings.Join(GetDrgAttachmentLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListVcnDrgAttachmentsResponse wrapper for the ListVcnDrgAttachments operation
type ListVcnDrgAttachmentsResponse struct {

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

func (response ListVcnDrgAttachmentsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListVcnDrgAttachmentsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListVcnDrgAttachmentsAttachmentTypeEnum Enum with underlying type: string
type ListVcnDrgAttachmentsAttachmentTypeEnum string

// Set of constants representing the allowable values for ListVcnDrgAttachmentsAttachmentTypeEnum
const (
	ListVcnDrgAttachmentsAttachmentTypeVcn                     ListVcnDrgAttachmentsAttachmentTypeEnum = "VCN"
	ListVcnDrgAttachmentsAttachmentTypeVirtualCircuit          ListVcnDrgAttachmentsAttachmentTypeEnum = "VIRTUAL_CIRCUIT"
	ListVcnDrgAttachmentsAttachmentTypeRemotePeeringConnection ListVcnDrgAttachmentsAttachmentTypeEnum = "REMOTE_PEERING_CONNECTION"
	ListVcnDrgAttachmentsAttachmentTypeIpsecTunnel             ListVcnDrgAttachmentsAttachmentTypeEnum = "IPSEC_TUNNEL"
	ListVcnDrgAttachmentsAttachmentTypeAll                     ListVcnDrgAttachmentsAttachmentTypeEnum = "ALL"
)

var mappingListVcnDrgAttachmentsAttachmentTypeEnum = map[string]ListVcnDrgAttachmentsAttachmentTypeEnum{
	"VCN":                       ListVcnDrgAttachmentsAttachmentTypeVcn,
	"VIRTUAL_CIRCUIT":           ListVcnDrgAttachmentsAttachmentTypeVirtualCircuit,
	"REMOTE_PEERING_CONNECTION": ListVcnDrgAttachmentsAttachmentTypeRemotePeeringConnection,
	"IPSEC_TUNNEL":              ListVcnDrgAttachmentsAttachmentTypeIpsecTunnel,
	"ALL":                       ListVcnDrgAttachmentsAttachmentTypeAll,
}

// GetListVcnDrgAttachmentsAttachmentTypeEnumValues Enumerates the set of values for ListVcnDrgAttachmentsAttachmentTypeEnum
func GetListVcnDrgAttachmentsAttachmentTypeEnumValues() []ListVcnDrgAttachmentsAttachmentTypeEnum {
	values := make([]ListVcnDrgAttachmentsAttachmentTypeEnum, 0)
	for _, v := range mappingListVcnDrgAttachmentsAttachmentTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetListVcnDrgAttachmentsAttachmentTypeEnumStringValues Enumerates the set of values in String for ListVcnDrgAttachmentsAttachmentTypeEnum
func GetListVcnDrgAttachmentsAttachmentTypeEnumStringValues() []string {
	return []string{
		"VCN",
		"VIRTUAL_CIRCUIT",
		"REMOTE_PEERING_CONNECTION",
		"IPSEC_TUNNEL",
		"ALL",
	}
}

// ListVcnDrgAttachmentsSortByEnum Enum with underlying type: string
type ListVcnDrgAttachmentsSortByEnum string

// Set of constants representing the allowable values for ListVcnDrgAttachmentsSortByEnum
const (
	ListVcnDrgAttachmentsSortByTimecreated ListVcnDrgAttachmentsSortByEnum = "TIMECREATED"
	ListVcnDrgAttachmentsSortByDisplayname ListVcnDrgAttachmentsSortByEnum = "DISPLAYNAME"
)

var mappingListVcnDrgAttachmentsSortByEnum = map[string]ListVcnDrgAttachmentsSortByEnum{
	"TIMECREATED": ListVcnDrgAttachmentsSortByTimecreated,
	"DISPLAYNAME": ListVcnDrgAttachmentsSortByDisplayname,
}

// GetListVcnDrgAttachmentsSortByEnumValues Enumerates the set of values for ListVcnDrgAttachmentsSortByEnum
func GetListVcnDrgAttachmentsSortByEnumValues() []ListVcnDrgAttachmentsSortByEnum {
	values := make([]ListVcnDrgAttachmentsSortByEnum, 0)
	for _, v := range mappingListVcnDrgAttachmentsSortByEnum {
		values = append(values, v)
	}
	return values
}

// GetListVcnDrgAttachmentsSortByEnumStringValues Enumerates the set of values in String for ListVcnDrgAttachmentsSortByEnum
func GetListVcnDrgAttachmentsSortByEnumStringValues() []string {
	return []string{
		"TIMECREATED",
		"DISPLAYNAME",
	}
}

// ListVcnDrgAttachmentsSortOrderEnum Enum with underlying type: string
type ListVcnDrgAttachmentsSortOrderEnum string

// Set of constants representing the allowable values for ListVcnDrgAttachmentsSortOrderEnum
const (
	ListVcnDrgAttachmentsSortOrderAsc  ListVcnDrgAttachmentsSortOrderEnum = "ASC"
	ListVcnDrgAttachmentsSortOrderDesc ListVcnDrgAttachmentsSortOrderEnum = "DESC"
)

var mappingListVcnDrgAttachmentsSortOrderEnum = map[string]ListVcnDrgAttachmentsSortOrderEnum{
	"ASC":  ListVcnDrgAttachmentsSortOrderAsc,
	"DESC": ListVcnDrgAttachmentsSortOrderDesc,
}

// GetListVcnDrgAttachmentsSortOrderEnumValues Enumerates the set of values for ListVcnDrgAttachmentsSortOrderEnum
func GetListVcnDrgAttachmentsSortOrderEnumValues() []ListVcnDrgAttachmentsSortOrderEnum {
	values := make([]ListVcnDrgAttachmentsSortOrderEnum, 0)
	for _, v := range mappingListVcnDrgAttachmentsSortOrderEnum {
		values = append(values, v)
	}
	return values
}

// GetListVcnDrgAttachmentsSortOrderEnumStringValues Enumerates the set of values in String for ListVcnDrgAttachmentsSortOrderEnum
func GetListVcnDrgAttachmentsSortOrderEnumStringValues() []string {
	return []string{
		"ASC",
		"DESC",
	}
}
