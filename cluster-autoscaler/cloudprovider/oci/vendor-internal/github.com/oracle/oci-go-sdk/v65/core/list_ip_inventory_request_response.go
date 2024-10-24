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

// ListIpInventoryRequest wrapper for the ListIpInventory operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListIpInventory.go.html to see an example of how to use ListIpInventoryRequest.
type ListIpInventoryRequest struct {

	// Details required to list the IP Inventory data.
	ListIpInventoryDetails `contributesTo:"body"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListIpInventoryRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListIpInventoryRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListIpInventoryRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListIpInventoryRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListIpInventoryRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListIpInventoryResponse wrapper for the ListIpInventory operation
type ListIpInventoryResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The IpInventoryCollection instance
	IpInventoryCollection `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

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
	LifecycleState ListIpInventoryLifecycleStateEnum `presentIn:"header" name:"lifecycle-state"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the resource.
	// Use GetWorkRequest (https://docs.cloud.oracle.com/api/#/en/workrequests/latest/WorkRequest/GetWorkRequest)
	// with this ID to track the status of the resource.
	DataRequestId *string `presentIn:"header" name:"data-request-id"`
}

func (response ListIpInventoryResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListIpInventoryResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListIpInventoryLifecycleStateEnum Enum with underlying type: string
type ListIpInventoryLifecycleStateEnum string

// Set of constants representing the allowable values for ListIpInventoryLifecycleStateEnum
const (
	ListIpInventoryLifecycleStateInProgress ListIpInventoryLifecycleStateEnum = "IN_PROGRESS"
	ListIpInventoryLifecycleStateDone       ListIpInventoryLifecycleStateEnum = "DONE"
)

var mappingListIpInventoryLifecycleStateEnum = map[string]ListIpInventoryLifecycleStateEnum{
	"IN_PROGRESS": ListIpInventoryLifecycleStateInProgress,
	"DONE":        ListIpInventoryLifecycleStateDone,
}

var mappingListIpInventoryLifecycleStateEnumLowerCase = map[string]ListIpInventoryLifecycleStateEnum{
	"in_progress": ListIpInventoryLifecycleStateInProgress,
	"done":        ListIpInventoryLifecycleStateDone,
}

// GetListIpInventoryLifecycleStateEnumValues Enumerates the set of values for ListIpInventoryLifecycleStateEnum
func GetListIpInventoryLifecycleStateEnumValues() []ListIpInventoryLifecycleStateEnum {
	values := make([]ListIpInventoryLifecycleStateEnum, 0)
	for _, v := range mappingListIpInventoryLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetListIpInventoryLifecycleStateEnumStringValues Enumerates the set of values in String for ListIpInventoryLifecycleStateEnum
func GetListIpInventoryLifecycleStateEnumStringValues() []string {
	return []string{
		"IN_PROGRESS",
		"DONE",
	}
}

// GetMappingListIpInventoryLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListIpInventoryLifecycleStateEnum(val string) (ListIpInventoryLifecycleStateEnum, bool) {
	enum, ok := mappingListIpInventoryLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
