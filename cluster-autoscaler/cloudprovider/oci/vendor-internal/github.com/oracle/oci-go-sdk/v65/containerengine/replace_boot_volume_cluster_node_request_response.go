// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package containerengine

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"net/http"
	"strings"
)

// ReplaceBootVolumeClusterNodeRequest wrapper for the ReplaceBootVolumeClusterNode operation
//
// # See also
//
// Click https://docs.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ReplaceBootVolumeClusterNode.go.html to see an example of how to use ReplaceBootVolumeClusterNodeRequest.
type ReplaceBootVolumeClusterNodeRequest struct {

	// The OCID of the cluster.
	ClusterId *string `mandatory:"true" contributesTo:"path" name:"clusterId"`

	// The OCID of the compute instance.
	NodeId *string `mandatory:"true" contributesTo:"path" name:"nodeId"`

	// The fields to replace boot volume of a node.
	ReplaceBootVolumeClusterNodeDetails `contributesTo:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// A token you supply to uniquely identify the request and provide idempotency if
	// the request is retried. Idempotency tokens expire after 24 hours.
	OpcRetryToken *string `mandatory:"false" contributesTo:"header" name:"opc-retry-token"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ReplaceBootVolumeClusterNodeRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ReplaceBootVolumeClusterNodeRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ReplaceBootVolumeClusterNodeRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ReplaceBootVolumeClusterNodeRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ReplaceBootVolumeClusterNodeRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ReplaceBootVolumeClusterNodeResponse wrapper for the ReplaceBootVolumeClusterNode operation
type ReplaceBootVolumeClusterNodeResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The OCID of the work request handling the operation.
	OpcWorkRequestId *string `presentIn:"header" name:"opc-work-request-id"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ReplaceBootVolumeClusterNodeResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ReplaceBootVolumeClusterNodeResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
