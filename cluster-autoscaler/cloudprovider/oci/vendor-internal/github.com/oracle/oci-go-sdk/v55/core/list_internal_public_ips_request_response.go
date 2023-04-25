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

// ListInternalPublicIpsRequest wrapper for the ListInternalPublicIps operation
type ListInternalPublicIpsRequest struct {

	// Whether the public IP is regional or specific to a particular availability domain.
	// * `REGION`: The public IP exists within a region and is assigned to a regional entity
	// (such as a NatGateway), or can be assigned to a private IP
	// in any availability domain in the region. Reserved public IPs have `scope` = `REGION`, as do
	// ephemeral public IPs assigned to a regional entity.
	// * `AVAILABILITY_DOMAIN`: The public IP exists within the availability domain of the entity
	// it's assigned to, which is specified by the `availabilityDomain` property of the public IP object.
	// Ephemeral public IPs that are assigned to private IPs have `scope` = `AVAILABILITY_DOMAIN`.
	Scope ListInternalPublicIpsScopeEnum `mandatory:"true" contributesTo:"query" name:"scope" omitEmpty:"true"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" contributesTo:"query" name:"availabilityDomain"`

	// A filter to return internal public IPs that match given lifetime.
	Lifetime ListInternalPublicIpsLifetimeEnum `mandatory:"false" contributesTo:"query" name:"lifetime" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListInternalPublicIpsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListInternalPublicIpsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListInternalPublicIpsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListInternalPublicIpsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListInternalPublicIpsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingListInternalPublicIpsScopeEnum[string(request.Scope)]; !ok && request.Scope != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Scope: %s. Supported values are: %s.", request.Scope, strings.Join(GetListInternalPublicIpsScopeEnumStringValues(), ",")))
	}
	if _, ok := mappingListInternalPublicIpsLifetimeEnum[string(request.Lifetime)]; !ok && request.Lifetime != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Lifetime: %s. Supported values are: %s.", request.Lifetime, strings.Join(GetListInternalPublicIpsLifetimeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListInternalPublicIpsResponse wrapper for the ListInternalPublicIps operation
type ListInternalPublicIpsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []InternalPublicIp instances
	Items []InternalPublicIp `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListInternalPublicIpsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListInternalPublicIpsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListInternalPublicIpsScopeEnum Enum with underlying type: string
type ListInternalPublicIpsScopeEnum string

// Set of constants representing the allowable values for ListInternalPublicIpsScopeEnum
const (
	ListInternalPublicIpsScopeRegion             ListInternalPublicIpsScopeEnum = "REGION"
	ListInternalPublicIpsScopeAvailabilityDomain ListInternalPublicIpsScopeEnum = "AVAILABILITY_DOMAIN"
)

var mappingListInternalPublicIpsScopeEnum = map[string]ListInternalPublicIpsScopeEnum{
	"REGION":              ListInternalPublicIpsScopeRegion,
	"AVAILABILITY_DOMAIN": ListInternalPublicIpsScopeAvailabilityDomain,
}

// GetListInternalPublicIpsScopeEnumValues Enumerates the set of values for ListInternalPublicIpsScopeEnum
func GetListInternalPublicIpsScopeEnumValues() []ListInternalPublicIpsScopeEnum {
	values := make([]ListInternalPublicIpsScopeEnum, 0)
	for _, v := range mappingListInternalPublicIpsScopeEnum {
		values = append(values, v)
	}
	return values
}

// GetListInternalPublicIpsScopeEnumStringValues Enumerates the set of values in String for ListInternalPublicIpsScopeEnum
func GetListInternalPublicIpsScopeEnumStringValues() []string {
	return []string{
		"REGION",
		"AVAILABILITY_DOMAIN",
	}
}

// ListInternalPublicIpsLifetimeEnum Enum with underlying type: string
type ListInternalPublicIpsLifetimeEnum string

// Set of constants representing the allowable values for ListInternalPublicIpsLifetimeEnum
const (
	ListInternalPublicIpsLifetimeEphemeral ListInternalPublicIpsLifetimeEnum = "EPHEMERAL"
	ListInternalPublicIpsLifetimeReserved  ListInternalPublicIpsLifetimeEnum = "RESERVED"
)

var mappingListInternalPublicIpsLifetimeEnum = map[string]ListInternalPublicIpsLifetimeEnum{
	"EPHEMERAL": ListInternalPublicIpsLifetimeEphemeral,
	"RESERVED":  ListInternalPublicIpsLifetimeReserved,
}

// GetListInternalPublicIpsLifetimeEnumValues Enumerates the set of values for ListInternalPublicIpsLifetimeEnum
func GetListInternalPublicIpsLifetimeEnumValues() []ListInternalPublicIpsLifetimeEnum {
	values := make([]ListInternalPublicIpsLifetimeEnum, 0)
	for _, v := range mappingListInternalPublicIpsLifetimeEnum {
		values = append(values, v)
	}
	return values
}

// GetListInternalPublicIpsLifetimeEnumStringValues Enumerates the set of values in String for ListInternalPublicIpsLifetimeEnum
func GetListInternalPublicIpsLifetimeEnumStringValues() []string {
	return []string{
		"EPHEMERAL",
		"RESERVED",
	}
}
