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

// GetVcnOverlapRequest wrapper for the GetVcnOverlap operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetVcnOverlap.go.html to see an example of how to use GetVcnOverlapRequest.
type GetVcnOverlapRequest struct {

	// Specify the OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN.
	VcnId *string `mandatory:"true" contributesTo:"path" name:"vcnId"`

	// Lists details of the IP Inventory VCN overlap data.
	GetVcnOverlapDetails GetIpInventoryVcnOverlapDetails `contributesTo:"body"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// A token that uniquely identifies a request so it can be retried in case of a timeout or
	// server error without risk of executing that same action again. Retry tokens expire after 24
	// hours, but can be invalidated before then due to conflicting operations (for example, if a resource
	// has been deleted and purged from the system, then a retry of the original creation request
	// may be rejected).
	OpcRetryToken *string `mandatory:"false" contributesTo:"header" name:"opc-retry-token"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetVcnOverlapRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetVcnOverlapRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request GetVcnOverlapRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetVcnOverlapRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request GetVcnOverlapRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// GetVcnOverlapResponse wrapper for the GetVcnOverlap operation
type GetVcnOverlapResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The IpInventoryVcnOverlapCollection instance
	IpInventoryVcnOverlapCollection `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For list pagination. A pagination token to get the total number of results available.
	OpcTotalItems *int `presentIn:"header" name:"opc-total-items"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the work request.
	// Use GetWorkRequest (https://docs.cloud.oracle.com/api/#/en/workrequests/latest/WorkRequest/GetWorkRequest)
	// with this ID to track the status of the request.
	OpcWorkRequestId *string `presentIn:"header" name:"opc-work-request-id"`

	// The IpInventory API current state.
	LifecycleState GetVcnOverlapLifecycleStateEnum `presentIn:"header" name:"lifecycle-state"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the resource.
	// Use GetWorkRequest (https://docs.cloud.oracle.com/api/#/en/workrequests/latest/WorkRequest/GetWorkRequest)
	// with this ID to track the status of the resource.
	DataRequestId *string `presentIn:"header" name:"data-request-id"`
}

func (response GetVcnOverlapResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetVcnOverlapResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// GetVcnOverlapLifecycleStateEnum Enum with underlying type: string
type GetVcnOverlapLifecycleStateEnum string

// Set of constants representing the allowable values for GetVcnOverlapLifecycleStateEnum
const (
	GetVcnOverlapLifecycleStateInProgress GetVcnOverlapLifecycleStateEnum = "IN_PROGRESS"
	GetVcnOverlapLifecycleStateDone       GetVcnOverlapLifecycleStateEnum = "DONE"
)

var mappingGetVcnOverlapLifecycleStateEnum = map[string]GetVcnOverlapLifecycleStateEnum{
	"IN_PROGRESS": GetVcnOverlapLifecycleStateInProgress,
	"DONE":        GetVcnOverlapLifecycleStateDone,
}

var mappingGetVcnOverlapLifecycleStateEnumLowerCase = map[string]GetVcnOverlapLifecycleStateEnum{
	"in_progress": GetVcnOverlapLifecycleStateInProgress,
	"done":        GetVcnOverlapLifecycleStateDone,
}

// GetGetVcnOverlapLifecycleStateEnumValues Enumerates the set of values for GetVcnOverlapLifecycleStateEnum
func GetGetVcnOverlapLifecycleStateEnumValues() []GetVcnOverlapLifecycleStateEnum {
	values := make([]GetVcnOverlapLifecycleStateEnum, 0)
	for _, v := range mappingGetVcnOverlapLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetGetVcnOverlapLifecycleStateEnumStringValues Enumerates the set of values in String for GetVcnOverlapLifecycleStateEnum
func GetGetVcnOverlapLifecycleStateEnumStringValues() []string {
	return []string{
		"IN_PROGRESS",
		"DONE",
	}
}

// GetMappingGetVcnOverlapLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingGetVcnOverlapLifecycleStateEnum(val string) (GetVcnOverlapLifecycleStateEnum, bool) {
	enum, ok := mappingGetVcnOverlapLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
