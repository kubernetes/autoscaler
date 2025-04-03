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

// ListPublicIpsRequest wrapper for the ListPublicIps operation
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListPublicIps.go.html to see an example of how to use ListPublicIpsRequest.
type ListPublicIpsRequest struct {

	// Whether the public IP is regional or specific to a particular availability domain.
	// * `REGION`: The public IP exists within a region and is assigned to a regional entity
	// (such as a NatGateway), or can be assigned to a private IP
	// in any availability domain in the region. Reserved public IPs have `scope` = `REGION`, as do
	// ephemeral public IPs assigned to a regional entity.
	// * `AVAILABILITY_DOMAIN`: The public IP exists within the availability domain of the entity
	// it's assigned to, which is specified by the `availabilityDomain` property of the public IP object.
	// Ephemeral public IPs that are assigned to private IPs have `scope` = `AVAILABILITY_DOMAIN`.
	Scope ListPublicIpsScopeEnum `mandatory:"true" contributesTo:"query" name:"scope" omitEmpty:"true"`

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

	// A filter to return only public IPs that match given lifetime.
	Lifetime ListPublicIpsLifetimeEnum `mandatory:"false" contributesTo:"query" name:"lifetime" omitEmpty:"true"`

	// A filter to return only resources that belong to the given public IP pool.
	PublicIpPoolId *string `mandatory:"false" contributesTo:"query" name:"publicIpPoolId"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListPublicIpsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListPublicIpsRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	_, err := request.ValidateEnumValue()
	if err != nil {
		return http.Request{}, err
	}
	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListPublicIpsRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListPublicIpsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (request ListPublicIpsRequest) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingListPublicIpsScopeEnum(string(request.Scope)); !ok && request.Scope != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Scope: %s. Supported values are: %s.", request.Scope, strings.Join(GetListPublicIpsScopeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingListPublicIpsLifetimeEnum(string(request.Lifetime)); !ok && request.Lifetime != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Lifetime: %s. Supported values are: %s.", request.Lifetime, strings.Join(GetListPublicIpsLifetimeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ListPublicIpsResponse wrapper for the ListPublicIps operation
type ListPublicIpsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []PublicIp instances
	Items []PublicIp `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListPublicIpsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListPublicIpsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListPublicIpsScopeEnum Enum with underlying type: string
type ListPublicIpsScopeEnum string

// Set of constants representing the allowable values for ListPublicIpsScopeEnum
const (
	ListPublicIpsScopeRegion             ListPublicIpsScopeEnum = "REGION"
	ListPublicIpsScopeAvailabilityDomain ListPublicIpsScopeEnum = "AVAILABILITY_DOMAIN"
)

var mappingListPublicIpsScopeEnum = map[string]ListPublicIpsScopeEnum{
	"REGION":              ListPublicIpsScopeRegion,
	"AVAILABILITY_DOMAIN": ListPublicIpsScopeAvailabilityDomain,
}

var mappingListPublicIpsScopeEnumLowerCase = map[string]ListPublicIpsScopeEnum{
	"region":              ListPublicIpsScopeRegion,
	"availability_domain": ListPublicIpsScopeAvailabilityDomain,
}

// GetListPublicIpsScopeEnumValues Enumerates the set of values for ListPublicIpsScopeEnum
func GetListPublicIpsScopeEnumValues() []ListPublicIpsScopeEnum {
	values := make([]ListPublicIpsScopeEnum, 0)
	for _, v := range mappingListPublicIpsScopeEnum {
		values = append(values, v)
	}
	return values
}

// GetListPublicIpsScopeEnumStringValues Enumerates the set of values in String for ListPublicIpsScopeEnum
func GetListPublicIpsScopeEnumStringValues() []string {
	return []string{
		"REGION",
		"AVAILABILITY_DOMAIN",
	}
}

// GetMappingListPublicIpsScopeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListPublicIpsScopeEnum(val string) (ListPublicIpsScopeEnum, bool) {
	enum, ok := mappingListPublicIpsScopeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ListPublicIpsLifetimeEnum Enum with underlying type: string
type ListPublicIpsLifetimeEnum string

// Set of constants representing the allowable values for ListPublicIpsLifetimeEnum
const (
	ListPublicIpsLifetimeEphemeral ListPublicIpsLifetimeEnum = "EPHEMERAL"
	ListPublicIpsLifetimeReserved  ListPublicIpsLifetimeEnum = "RESERVED"
)

var mappingListPublicIpsLifetimeEnum = map[string]ListPublicIpsLifetimeEnum{
	"EPHEMERAL": ListPublicIpsLifetimeEphemeral,
	"RESERVED":  ListPublicIpsLifetimeReserved,
}

var mappingListPublicIpsLifetimeEnumLowerCase = map[string]ListPublicIpsLifetimeEnum{
	"ephemeral": ListPublicIpsLifetimeEphemeral,
	"reserved":  ListPublicIpsLifetimeReserved,
}

// GetListPublicIpsLifetimeEnumValues Enumerates the set of values for ListPublicIpsLifetimeEnum
func GetListPublicIpsLifetimeEnumValues() []ListPublicIpsLifetimeEnum {
	values := make([]ListPublicIpsLifetimeEnum, 0)
	for _, v := range mappingListPublicIpsLifetimeEnum {
		values = append(values, v)
	}
	return values
}

// GetListPublicIpsLifetimeEnumStringValues Enumerates the set of values in String for ListPublicIpsLifetimeEnum
func GetListPublicIpsLifetimeEnumStringValues() []string {
	return []string{
		"EPHEMERAL",
		"RESERVED",
	}
}

// GetMappingListPublicIpsLifetimeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingListPublicIpsLifetimeEnum(val string) (ListPublicIpsLifetimeEnum, bool) {
	enum, ok := mappingListPublicIpsLifetimeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
