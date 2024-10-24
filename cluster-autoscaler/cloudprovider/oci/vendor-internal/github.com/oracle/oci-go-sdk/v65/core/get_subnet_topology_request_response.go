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

// GetSubnetTopologyRequest wrapper for the GetSubnetTopology operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetSubnetTopology.go.html to see an example of how to use GetSubnetTopologyRequest.
type GetSubnetTopologyRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet.
	SubnetId *string `mandatory:"true" contributesTo:"query" name:"subnetId"`

	// Valid values are `ANY` and `ACCESSIBLE`. The default is `ANY`.
	// Setting this to `ACCESSIBLE` returns only compartments for which a
	// user has INSPECT permissions, either directly or indirectly (permissions can be on a
	// resource in a subcompartment). A restricted set of fields is returned for compartments in which a user has
	// indirect INSPECT permissions.
	// When set to `ANY` permissions are not checked.
	AccessLevel GetSubnetTopologyAccessLevelEnum `mandatory:"false" contributesTo:"query" name:"accessLevel" omitEmpty:"true"`

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

func (request GetSubnetTopologyRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetSubnetTopologyRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request GetSubnetTopologyRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetSubnetTopologyRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request GetSubnetTopologyRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingGetSubnetTopologyAccessLevelEnum(string(request.AccessLevel)); !ok && request.AccessLevel != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AccessLevel: %s. Supported values are: %s.", request.AccessLevel, strings.Join(GetGetSubnetTopologyAccessLevelEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// GetSubnetTopologyResponse wrapper for the GetSubnetTopology operation
type GetSubnetTopologyResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The SubnetTopology instance
	SubnetTopology `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetSubnetTopologyResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetSubnetTopologyResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// GetSubnetTopologyAccessLevelEnum Enum with underlying type: string
type GetSubnetTopologyAccessLevelEnum string

// Set of constants representing the allowable values for GetSubnetTopologyAccessLevelEnum
const (
	GetSubnetTopologyAccessLevelAny        GetSubnetTopologyAccessLevelEnum = "ANY"
	GetSubnetTopologyAccessLevelAccessible GetSubnetTopologyAccessLevelEnum = "ACCESSIBLE"
)

var mappingGetSubnetTopologyAccessLevelEnum = map[string]GetSubnetTopologyAccessLevelEnum{
	"ANY":        GetSubnetTopologyAccessLevelAny,
	"ACCESSIBLE": GetSubnetTopologyAccessLevelAccessible,
}

var mappingGetSubnetTopologyAccessLevelEnumLowerCase = map[string]GetSubnetTopologyAccessLevelEnum{
	"any":        GetSubnetTopologyAccessLevelAny,
	"accessible": GetSubnetTopologyAccessLevelAccessible,
}

// GetGetSubnetTopologyAccessLevelEnumValues Enumerates the set of values for GetSubnetTopologyAccessLevelEnum
func GetGetSubnetTopologyAccessLevelEnumValues() []GetSubnetTopologyAccessLevelEnum {
	values := make([]GetSubnetTopologyAccessLevelEnum, 0)
	for _, v := range mappingGetSubnetTopologyAccessLevelEnum {
		values = append(values, v)
	}
	return values
}

// GetGetSubnetTopologyAccessLevelEnumStringValues Enumerates the set of values in String for GetSubnetTopologyAccessLevelEnum
func GetGetSubnetTopologyAccessLevelEnumStringValues() []string {
	return []string{
		"ANY",
		"ACCESSIBLE",
	}
}

// GetMappingGetSubnetTopologyAccessLevelEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingGetSubnetTopologyAccessLevelEnum(val string) (GetSubnetTopologyAccessLevelEnum, bool) {
	enum, ok := mappingGetSubnetTopologyAccessLevelEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
