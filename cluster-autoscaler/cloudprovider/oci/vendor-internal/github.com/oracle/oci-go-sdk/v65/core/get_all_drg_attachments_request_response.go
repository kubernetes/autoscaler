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

// GetAllDrgAttachmentsRequest wrapper for the GetAllDrgAttachments operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetAllDrgAttachments.go.html to see an example of how to use GetAllDrgAttachmentsRequest.
type GetAllDrgAttachmentsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DRG.
	DrgId *string `mandatory:"true" contributesTo:"path" name:"drgId"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The type for the network resource attached to the DRG.
	AttachmentType GetAllDrgAttachmentsAttachmentTypeEnum `mandatory:"false" contributesTo:"query" name:"attachmentType" omitEmpty:"true"`

	// Whether the DRG attachment lives in a different tenancy than the DRG.
	IsCrossTenancy *bool `mandatory:"false" contributesTo:"query" name:"isCrossTenancy"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetAllDrgAttachmentsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetAllDrgAttachmentsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request GetAllDrgAttachmentsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetAllDrgAttachmentsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request GetAllDrgAttachmentsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingGetAllDrgAttachmentsAttachmentTypeEnum(string(request.AttachmentType)); !ok && request.AttachmentType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AttachmentType: %s. Supported values are: %s.", request.AttachmentType, strings.Join(GetGetAllDrgAttachmentsAttachmentTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// GetAllDrgAttachmentsResponse wrapper for the GetAllDrgAttachments operation
type GetAllDrgAttachmentsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []DrgAttachmentInfo instances
	Items []DrgAttachmentInfo `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetAllDrgAttachmentsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetAllDrgAttachmentsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// GetAllDrgAttachmentsAttachmentTypeEnum Enum with underlying type: string
type GetAllDrgAttachmentsAttachmentTypeEnum string

// Set of constants representing the allowable values for GetAllDrgAttachmentsAttachmentTypeEnum
const (
	GetAllDrgAttachmentsAttachmentTypeVcn                     GetAllDrgAttachmentsAttachmentTypeEnum = "VCN"
	GetAllDrgAttachmentsAttachmentTypeVirtualCircuit          GetAllDrgAttachmentsAttachmentTypeEnum = "VIRTUAL_CIRCUIT"
	GetAllDrgAttachmentsAttachmentTypeRemotePeeringConnection GetAllDrgAttachmentsAttachmentTypeEnum = "REMOTE_PEERING_CONNECTION"
	GetAllDrgAttachmentsAttachmentTypeIpsecTunnel             GetAllDrgAttachmentsAttachmentTypeEnum = "IPSEC_TUNNEL"
	GetAllDrgAttachmentsAttachmentTypeAll                     GetAllDrgAttachmentsAttachmentTypeEnum = "ALL"
)

var mappingGetAllDrgAttachmentsAttachmentTypeEnum = map[string]GetAllDrgAttachmentsAttachmentTypeEnum{
	"VCN":                       GetAllDrgAttachmentsAttachmentTypeVcn,
	"VIRTUAL_CIRCUIT":           GetAllDrgAttachmentsAttachmentTypeVirtualCircuit,
	"REMOTE_PEERING_CONNECTION": GetAllDrgAttachmentsAttachmentTypeRemotePeeringConnection,
	"IPSEC_TUNNEL":              GetAllDrgAttachmentsAttachmentTypeIpsecTunnel,
	"ALL":                       GetAllDrgAttachmentsAttachmentTypeAll,
}

var mappingGetAllDrgAttachmentsAttachmentTypeEnumLowerCase = map[string]GetAllDrgAttachmentsAttachmentTypeEnum{
	"vcn":                       GetAllDrgAttachmentsAttachmentTypeVcn,
	"virtual_circuit":           GetAllDrgAttachmentsAttachmentTypeVirtualCircuit,
	"remote_peering_connection": GetAllDrgAttachmentsAttachmentTypeRemotePeeringConnection,
	"ipsec_tunnel":              GetAllDrgAttachmentsAttachmentTypeIpsecTunnel,
	"all":                       GetAllDrgAttachmentsAttachmentTypeAll,
}

// GetGetAllDrgAttachmentsAttachmentTypeEnumValues Enumerates the set of values for GetAllDrgAttachmentsAttachmentTypeEnum
func GetGetAllDrgAttachmentsAttachmentTypeEnumValues() []GetAllDrgAttachmentsAttachmentTypeEnum {
	values := make([]GetAllDrgAttachmentsAttachmentTypeEnum, 0)
	for _, v := range mappingGetAllDrgAttachmentsAttachmentTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetGetAllDrgAttachmentsAttachmentTypeEnumStringValues Enumerates the set of values in String for GetAllDrgAttachmentsAttachmentTypeEnum
func GetGetAllDrgAttachmentsAttachmentTypeEnumStringValues() []string {
	return []string{
		"VCN",
		"VIRTUAL_CIRCUIT",
		"REMOTE_PEERING_CONNECTION",
		"IPSEC_TUNNEL",
		"ALL",
	}
}

// GetMappingGetAllDrgAttachmentsAttachmentTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingGetAllDrgAttachmentsAttachmentTypeEnum(val string) (GetAllDrgAttachmentsAttachmentTypeEnum, bool) {
	enum, ok := mappingGetAllDrgAttachmentsAttachmentTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
