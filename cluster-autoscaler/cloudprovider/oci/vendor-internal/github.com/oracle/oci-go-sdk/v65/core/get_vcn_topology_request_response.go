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

// GetVcnTopologyRequest wrapper for the GetVcnTopology operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetVcnTopology.go.html to see an example of how to use GetVcnTopologyRequest.
type GetVcnTopologyRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN.
	VcnId *string `mandatory:"true" contributesTo:"query" name:"vcnId"`

	// Valid values are `ANY` and `ACCESSIBLE`. The default is `ANY`.
	// Setting this to `ACCESSIBLE` returns only compartments for which a
	// user has INSPECT permissions, either directly or indirectly (permissions can be on a
	// resource in a subcompartment). A restricted set of fields is returned for compartments in which a user has
	// indirect INSPECT permissions.
	// When set to `ANY` permissions are not checked.
	AccessLevel GetVcnTopologyAccessLevelEnum `mandatory:"false" contributesTo:"query" name:"accessLevel" omitEmpty:"true"`

	// When set to true, the hierarchy of compartments is traversed
	// and the specified compartment and its subcompartments are
	// inspected depending on the the setting of `accessLevel`.
	// Default is false.
	QueryCompartmentSubtree *bool `mandatory:"false" contributesTo:"query" name:"queryCompartmentSubtree"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// For querying if there is a cached value on the server. The If-None-Match HTTP request header
	// makes the request conditional. For GET and HEAD methods, the server will send back the requested
	// resource, with a 200 status, only if it doesn't have an ETag matching the given ones.
	// For other methods, the request will be processed only if the eventually existing resource's
	// ETag doesn't match any of the values listed.
	IfNoneMatch *string `mandatory:"false" contributesTo:"header" name:"if-none-match"`

	// The Cache-Control HTTP header holds directives (instructions)
	// for caching in both requests and responses.
	CacheControl *string `mandatory:"false" contributesTo:"header" name:"cache-control"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetVcnTopologyRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetVcnTopologyRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request GetVcnTopologyRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetVcnTopologyRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request GetVcnTopologyRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingGetVcnTopologyAccessLevelEnum(string(request.AccessLevel)); !ok && request.AccessLevel != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AccessLevel: %s. Supported values are: %s.", request.AccessLevel, strings.Join(GetGetVcnTopologyAccessLevelEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// GetVcnTopologyResponse wrapper for the GetVcnTopology operation
type GetVcnTopologyResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The VcnTopology instance
	VcnTopology `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetVcnTopologyResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetVcnTopologyResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// GetVcnTopologyAccessLevelEnum Enum with underlying type: string
type GetVcnTopologyAccessLevelEnum string

// Set of constants representing the allowable values for GetVcnTopologyAccessLevelEnum
const (
	GetVcnTopologyAccessLevelAny        GetVcnTopologyAccessLevelEnum = "ANY"
	GetVcnTopologyAccessLevelAccessible GetVcnTopologyAccessLevelEnum = "ACCESSIBLE"
)

var mappingGetVcnTopologyAccessLevelEnum = map[string]GetVcnTopologyAccessLevelEnum{
	"ANY":        GetVcnTopologyAccessLevelAny,
	"ACCESSIBLE": GetVcnTopologyAccessLevelAccessible,
}

var mappingGetVcnTopologyAccessLevelEnumLowerCase = map[string]GetVcnTopologyAccessLevelEnum{
	"any":        GetVcnTopologyAccessLevelAny,
	"accessible": GetVcnTopologyAccessLevelAccessible,
}

// GetGetVcnTopologyAccessLevelEnumValues Enumerates the set of values for GetVcnTopologyAccessLevelEnum
func GetGetVcnTopologyAccessLevelEnumValues() []GetVcnTopologyAccessLevelEnum {
	values := make([]GetVcnTopologyAccessLevelEnum, 0)
	for _, v := range mappingGetVcnTopologyAccessLevelEnum {
		values = append(values, v)
	}
	return values
}

// GetGetVcnTopologyAccessLevelEnumStringValues Enumerates the set of values in String for GetVcnTopologyAccessLevelEnum
func GetGetVcnTopologyAccessLevelEnumStringValues() []string {
	return []string{
		"ANY",
		"ACCESSIBLE",
	}
}

// GetMappingGetVcnTopologyAccessLevelEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingGetVcnTopologyAccessLevelEnum(val string) (GetVcnTopologyAccessLevelEnum, bool) {
	enum, ok := mappingGetVcnTopologyAccessLevelEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
