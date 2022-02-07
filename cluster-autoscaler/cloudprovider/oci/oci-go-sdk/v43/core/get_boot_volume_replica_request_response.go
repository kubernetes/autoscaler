// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
	"net/http"
)

// GetBootVolumeReplicaRequest wrapper for the GetBootVolumeReplica operation
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetBootVolumeReplica.go.html to see an example of how to use GetBootVolumeReplicaRequest.
type GetBootVolumeReplicaRequest struct {

	// The OCID of the boot volume replica.
	BootVolumeReplicaId *string `mandatory:"true" contributesTo:"path" name:"bootVolumeReplicaId"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetBootVolumeReplicaRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetBootVolumeReplicaRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser) (http.Request, error) {

	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// BinaryRequestBody implements the OCIRequest interface
func (request GetBootVolumeReplicaRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetBootVolumeReplicaRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// GetBootVolumeReplicaResponse wrapper for the GetBootVolumeReplica operation
type GetBootVolumeReplicaResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The BootVolumeReplica instance
	BootVolumeReplica `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetBootVolumeReplicaResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetBootVolumeReplicaResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
