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

// ListFlowLogConfigAttachmentsRequest wrapper for the ListFlowLogConfigAttachments operation
type ListFlowLogConfigAttachmentsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of a resource that has flow logs enabled.
	TargetEntityId *string `mandatory:"false" contributesTo:"query" name:"targetEntityId"`

	// The type of resource that has flow logs enabled.
	TargetEntityType ListFlowLogConfigAttachmentsTargetEntityTypeEnum `mandatory:"false" contributesTo:"query" name:"targetEntityType" omitEmpty:"true"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListFlowLogConfigAttachmentsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListFlowLogConfigAttachmentsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListFlowLogConfigAttachmentsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListFlowLogConfigAttachmentsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListFlowLogConfigAttachmentsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingListFlowLogConfigAttachmentsTargetEntityTypeEnum[string(request.TargetEntityType)]; !ok && request.TargetEntityType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TargetEntityType: %s. Supported values are: %s.", request.TargetEntityType, strings.Join(GetListFlowLogConfigAttachmentsTargetEntityTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListFlowLogConfigAttachmentsResponse wrapper for the ListFlowLogConfigAttachments operation
type ListFlowLogConfigAttachmentsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []FlowLogConfigAttachment instances
	Items []FlowLogConfigAttachment `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListFlowLogConfigAttachmentsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListFlowLogConfigAttachmentsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListFlowLogConfigAttachmentsTargetEntityTypeEnum Enum with underlying type: string
type ListFlowLogConfigAttachmentsTargetEntityTypeEnum string

// Set of constants representing the allowable values for ListFlowLogConfigAttachmentsTargetEntityTypeEnum
const (
	ListFlowLogConfigAttachmentsTargetEntityTypeSubnet ListFlowLogConfigAttachmentsTargetEntityTypeEnum = "SUBNET"
)

var mappingListFlowLogConfigAttachmentsTargetEntityTypeEnum = map[string]ListFlowLogConfigAttachmentsTargetEntityTypeEnum{
	"SUBNET": ListFlowLogConfigAttachmentsTargetEntityTypeSubnet,
}

// GetListFlowLogConfigAttachmentsTargetEntityTypeEnumValues Enumerates the set of values for ListFlowLogConfigAttachmentsTargetEntityTypeEnum
func GetListFlowLogConfigAttachmentsTargetEntityTypeEnumValues() []ListFlowLogConfigAttachmentsTargetEntityTypeEnum {
	values := make([]ListFlowLogConfigAttachmentsTargetEntityTypeEnum, 0)
	for _, v := range mappingListFlowLogConfigAttachmentsTargetEntityTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetListFlowLogConfigAttachmentsTargetEntityTypeEnumStringValues Enumerates the set of values in String for ListFlowLogConfigAttachmentsTargetEntityTypeEnum
func GetListFlowLogConfigAttachmentsTargetEntityTypeEnumStringValues() []string {
	return []string{
		"SUBNET",
	}
}
