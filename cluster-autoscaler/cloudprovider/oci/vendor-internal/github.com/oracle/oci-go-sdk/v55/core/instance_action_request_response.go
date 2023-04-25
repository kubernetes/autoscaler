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

// InstanceActionRequest wrapper for the InstanceAction operation
type InstanceActionRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the instance.
	InstanceId *string `mandatory:"true" contributesTo:"path" name:"instanceId"`

	// The action to perform on the instance.
	Action InstanceActionActionEnum `mandatory:"true" contributesTo:"query" name:"action" omitEmpty:"true"`

	// A token that uniquely identifies a request so it can be retried in case of a timeout or
	// server error without risk of executing that same action again. Retry tokens expire after 24
	// hours, but can be invalidated before then due to conflicting operations (for example, if a resource
	// has been deleted and purged from the system, then a retry of the original creation request
	// may be rejected).
	OpcRetryToken *string `mandatory:"false" contributesTo:"header" name:"opc-retry-token"`

	// For optimistic concurrency control. In the PUT or DELETE call for a resource, set the `if-match`
	// parameter to the value of the etag from a previous GET or POST response for that resource. The resource
	// will be updated or deleted only if the etag you provide matches the resource's current etag value.
	IfMatch *string `mandatory:"false" contributesTo:"header" name:"if-match"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request InstanceActionRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request InstanceActionRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request InstanceActionRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request InstanceActionRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request InstanceActionRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingInstanceActionActionEnum[string(request.Action)]; !ok && request.Action != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Action: %s. Supported values are: %s.", request.Action, strings.Join(GetInstanceActionActionEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstanceActionResponse wrapper for the InstanceAction operation
type InstanceActionResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Instance instance
	Instance `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response InstanceActionResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response InstanceActionResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// InstanceActionActionEnum Enum with underlying type: string
type InstanceActionActionEnum string

// Set of constants representing the allowable values for InstanceActionActionEnum
const (
	InstanceActionActionStop                    InstanceActionActionEnum = "STOP"
	InstanceActionActionStart                   InstanceActionActionEnum = "START"
	InstanceActionActionSoftreset               InstanceActionActionEnum = "SOFTRESET"
	InstanceActionActionReset                   InstanceActionActionEnum = "RESET"
	InstanceActionActionSoftstop                InstanceActionActionEnum = "SOFTSTOP"
	InstanceActionActionValidatelivemigrate     InstanceActionActionEnum = "VALIDATELIVEMIGRATE"
	InstanceActionActionSenddiagnosticinterrupt InstanceActionActionEnum = "SENDDIAGNOSTICINTERRUPT"
	InstanceActionActionExtendscheduledstop     InstanceActionActionEnum = "EXTENDSCHEDULEDSTOP"
)

var mappingInstanceActionActionEnum = map[string]InstanceActionActionEnum{
	"STOP":                    InstanceActionActionStop,
	"START":                   InstanceActionActionStart,
	"SOFTRESET":               InstanceActionActionSoftreset,
	"RESET":                   InstanceActionActionReset,
	"SOFTSTOP":                InstanceActionActionSoftstop,
	"VALIDATELIVEMIGRATE":     InstanceActionActionValidatelivemigrate,
	"SENDDIAGNOSTICINTERRUPT": InstanceActionActionSenddiagnosticinterrupt,
	"EXTENDSCHEDULEDSTOP":     InstanceActionActionExtendscheduledstop,
}

// GetInstanceActionActionEnumValues Enumerates the set of values for InstanceActionActionEnum
func GetInstanceActionActionEnumValues() []InstanceActionActionEnum {
	values := make([]InstanceActionActionEnum, 0)
	for _, v := range mappingInstanceActionActionEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceActionActionEnumStringValues Enumerates the set of values in String for InstanceActionActionEnum
func GetInstanceActionActionEnumStringValues() []string {
	return []string{
		"STOP",
		"START",
		"SOFTRESET",
		"RESET",
		"SOFTSTOP",
		"VALIDATELIVEMIGRATE",
		"SENDDIAGNOSTICINTERRUPT",
		"EXTENDSCHEDULEDSTOP",
	}
}
