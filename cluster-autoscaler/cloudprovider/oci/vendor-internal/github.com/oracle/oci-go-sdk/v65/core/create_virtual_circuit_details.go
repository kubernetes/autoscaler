// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.cloud.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// CreateVirtualCircuitDetails The representation of CreateVirtualCircuitDetails
type CreateVirtualCircuitDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment to contain the virtual circuit.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The type of IP addresses used in this virtual circuit. PRIVATE
	// means RFC 1918 (https://tools.ietf.org/html/rfc1918) addresses
	// (10.0.0.0/8, 172.16/12, and 192.168/16).
	Type CreateVirtualCircuitDetailsTypeEnum `mandatory:"true" json:"type"`

	// The provisioned data rate of the connection. To get a list of the
	// available bandwidth levels (that is, shapes), see
	// ListFastConnectProviderVirtualCircuitBandwidthShapes.
	// Example: `10 Gbps`
	BandwidthShapeName *string `mandatory:"false" json:"bandwidthShapeName"`

	// Create a `CrossConnectMapping` for each cross-connect or cross-connect
	// group this virtual circuit will run on.
	CrossConnectMappings []CrossConnectMapping `mandatory:"false" json:"crossConnectMappings"`

	// The routing policy sets how routing information about the Oracle cloud is shared over a public virtual circuit.
	// Policies available are: `ORACLE_SERVICE_NETWORK`, `REGIONAL`, `MARKET_LEVEL`, and `GLOBAL`.
	// See Route Filtering (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/routingonprem.htm#route_filtering) for details.
	// By default, routing information is shared for all routes in the same market.
	RoutingPolicy []CreateVirtualCircuitDetailsRoutingPolicyEnum `mandatory:"false" json:"routingPolicy,omitempty"`

	// Set to `ENABLED` (the default) to activate the BGP session of the virtual circuit, set to `DISABLED` to deactivate the virtual circuit.
	BgpAdminState CreateVirtualCircuitDetailsBgpAdminStateEnum `mandatory:"false" json:"bgpAdminState,omitempty"`

	// Set to `true` to enable BFD for IPv4 BGP peering, or set to `false` to disable BFD. If this is not set, the default is `false`.
	IsBfdEnabled *bool `mandatory:"false" json:"isBfdEnabled"`

	// Set to `true` for the virtual circuit to carry only encrypted traffic, or set to `false` for the virtual circuit to carry unencrypted traffic. If this is not set, the default is `false`.
	IsTransportMode *bool `mandatory:"false" json:"isTransportMode"`

	// Deprecated. Instead use `customerAsn`.
	// If you specify values for both, the request will be rejected.
	CustomerBgpAsn *int `mandatory:"false" json:"customerBgpAsn"`

	// Your BGP ASN (either public or private). Provide this value only if
	// there's a BGP session that goes from your edge router to Oracle.
	// Otherwise, leave this empty or null.
	// Can be a 2-byte or 4-byte ASN. Uses "asplain" format.
	// Example: `12345` (2-byte) or `1587232876` (4-byte)
	CustomerAsn *int64 `mandatory:"false" json:"customerAsn"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// For private virtual circuits only. The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the Drg
	// that this virtual circuit uses.
	GatewayId *string `mandatory:"false" json:"gatewayId"`

	// Deprecated. Instead use `providerServiceId`.
	// To get a list of the provider names, see
	// ListFastConnectProviderServices.
	ProviderName *string `mandatory:"false" json:"providerName"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the service offered by the provider (if you're connecting
	// via a provider). To get a list of the available service offerings, see
	// ListFastConnectProviderServices.
	ProviderServiceId *string `mandatory:"false" json:"providerServiceId"`

	// The service key name offered by the provider (if the customer is connecting via a provider).
	ProviderServiceKeyName *string `mandatory:"false" json:"providerServiceKeyName"`

	// Deprecated. Instead use `providerServiceId`.
	// To get a list of the provider names, see
	// ListFastConnectProviderServices.
	ProviderServiceName *string `mandatory:"false" json:"providerServiceName"`

	// For a public virtual circuit. The public IP prefixes (CIDRs) the customer wants to
	// advertise across the connection.
	PublicPrefixes []CreateVirtualCircuitPublicPrefixDetails `mandatory:"false" json:"publicPrefixes"`

	// The Oracle Cloud Infrastructure region where this virtual
	// circuit is located.
	// Example: `phx`
	Region *string `mandatory:"false" json:"region"`

	// The layer 3 IP MTU to use with this virtual circuit.
	IpMtu VirtualCircuitIpMtuEnum `mandatory:"false" json:"ipMtu,omitempty"`
}

func (m CreateVirtualCircuitDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateVirtualCircuitDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingCreateVirtualCircuitDetailsTypeEnum(string(m.Type)); !ok && m.Type != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Type: %s. Supported values are: %s.", m.Type, strings.Join(GetCreateVirtualCircuitDetailsTypeEnumStringValues(), ",")))
	}

	for _, val := range m.RoutingPolicy {
		if _, ok := GetMappingCreateVirtualCircuitDetailsRoutingPolicyEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RoutingPolicy: %s. Supported values are: %s.", val, strings.Join(GetCreateVirtualCircuitDetailsRoutingPolicyEnumStringValues(), ",")))
		}
	}

	if _, ok := GetMappingCreateVirtualCircuitDetailsBgpAdminStateEnum(string(m.BgpAdminState)); !ok && m.BgpAdminState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BgpAdminState: %s. Supported values are: %s.", m.BgpAdminState, strings.Join(GetCreateVirtualCircuitDetailsBgpAdminStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitIpMtuEnum(string(m.IpMtu)); !ok && m.IpMtu != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for IpMtu: %s. Supported values are: %s.", m.IpMtu, strings.Join(GetVirtualCircuitIpMtuEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateVirtualCircuitDetailsRoutingPolicyEnum Enum with underlying type: string
type CreateVirtualCircuitDetailsRoutingPolicyEnum string

// Set of constants representing the allowable values for CreateVirtualCircuitDetailsRoutingPolicyEnum
const (
	CreateVirtualCircuitDetailsRoutingPolicyOracleServiceNetwork CreateVirtualCircuitDetailsRoutingPolicyEnum = "ORACLE_SERVICE_NETWORK"
	CreateVirtualCircuitDetailsRoutingPolicyRegional             CreateVirtualCircuitDetailsRoutingPolicyEnum = "REGIONAL"
	CreateVirtualCircuitDetailsRoutingPolicyMarketLevel          CreateVirtualCircuitDetailsRoutingPolicyEnum = "MARKET_LEVEL"
	CreateVirtualCircuitDetailsRoutingPolicyGlobal               CreateVirtualCircuitDetailsRoutingPolicyEnum = "GLOBAL"
)

var mappingCreateVirtualCircuitDetailsRoutingPolicyEnum = map[string]CreateVirtualCircuitDetailsRoutingPolicyEnum{
	"ORACLE_SERVICE_NETWORK": CreateVirtualCircuitDetailsRoutingPolicyOracleServiceNetwork,
	"REGIONAL":               CreateVirtualCircuitDetailsRoutingPolicyRegional,
	"MARKET_LEVEL":           CreateVirtualCircuitDetailsRoutingPolicyMarketLevel,
	"GLOBAL":                 CreateVirtualCircuitDetailsRoutingPolicyGlobal,
}

var mappingCreateVirtualCircuitDetailsRoutingPolicyEnumLowerCase = map[string]CreateVirtualCircuitDetailsRoutingPolicyEnum{
	"oracle_service_network": CreateVirtualCircuitDetailsRoutingPolicyOracleServiceNetwork,
	"regional":               CreateVirtualCircuitDetailsRoutingPolicyRegional,
	"market_level":           CreateVirtualCircuitDetailsRoutingPolicyMarketLevel,
	"global":                 CreateVirtualCircuitDetailsRoutingPolicyGlobal,
}

// GetCreateVirtualCircuitDetailsRoutingPolicyEnumValues Enumerates the set of values for CreateVirtualCircuitDetailsRoutingPolicyEnum
func GetCreateVirtualCircuitDetailsRoutingPolicyEnumValues() []CreateVirtualCircuitDetailsRoutingPolicyEnum {
	values := make([]CreateVirtualCircuitDetailsRoutingPolicyEnum, 0)
	for _, v := range mappingCreateVirtualCircuitDetailsRoutingPolicyEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateVirtualCircuitDetailsRoutingPolicyEnumStringValues Enumerates the set of values in String for CreateVirtualCircuitDetailsRoutingPolicyEnum
func GetCreateVirtualCircuitDetailsRoutingPolicyEnumStringValues() []string {
	return []string{
		"ORACLE_SERVICE_NETWORK",
		"REGIONAL",
		"MARKET_LEVEL",
		"GLOBAL",
	}
}

// GetMappingCreateVirtualCircuitDetailsRoutingPolicyEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateVirtualCircuitDetailsRoutingPolicyEnum(val string) (CreateVirtualCircuitDetailsRoutingPolicyEnum, bool) {
	enum, ok := mappingCreateVirtualCircuitDetailsRoutingPolicyEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CreateVirtualCircuitDetailsBgpAdminStateEnum Enum with underlying type: string
type CreateVirtualCircuitDetailsBgpAdminStateEnum string

// Set of constants representing the allowable values for CreateVirtualCircuitDetailsBgpAdminStateEnum
const (
	CreateVirtualCircuitDetailsBgpAdminStateEnabled  CreateVirtualCircuitDetailsBgpAdminStateEnum = "ENABLED"
	CreateVirtualCircuitDetailsBgpAdminStateDisabled CreateVirtualCircuitDetailsBgpAdminStateEnum = "DISABLED"
)

var mappingCreateVirtualCircuitDetailsBgpAdminStateEnum = map[string]CreateVirtualCircuitDetailsBgpAdminStateEnum{
	"ENABLED":  CreateVirtualCircuitDetailsBgpAdminStateEnabled,
	"DISABLED": CreateVirtualCircuitDetailsBgpAdminStateDisabled,
}

var mappingCreateVirtualCircuitDetailsBgpAdminStateEnumLowerCase = map[string]CreateVirtualCircuitDetailsBgpAdminStateEnum{
	"enabled":  CreateVirtualCircuitDetailsBgpAdminStateEnabled,
	"disabled": CreateVirtualCircuitDetailsBgpAdminStateDisabled,
}

// GetCreateVirtualCircuitDetailsBgpAdminStateEnumValues Enumerates the set of values for CreateVirtualCircuitDetailsBgpAdminStateEnum
func GetCreateVirtualCircuitDetailsBgpAdminStateEnumValues() []CreateVirtualCircuitDetailsBgpAdminStateEnum {
	values := make([]CreateVirtualCircuitDetailsBgpAdminStateEnum, 0)
	for _, v := range mappingCreateVirtualCircuitDetailsBgpAdminStateEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateVirtualCircuitDetailsBgpAdminStateEnumStringValues Enumerates the set of values in String for CreateVirtualCircuitDetailsBgpAdminStateEnum
func GetCreateVirtualCircuitDetailsBgpAdminStateEnumStringValues() []string {
	return []string{
		"ENABLED",
		"DISABLED",
	}
}

// GetMappingCreateVirtualCircuitDetailsBgpAdminStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateVirtualCircuitDetailsBgpAdminStateEnum(val string) (CreateVirtualCircuitDetailsBgpAdminStateEnum, bool) {
	enum, ok := mappingCreateVirtualCircuitDetailsBgpAdminStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CreateVirtualCircuitDetailsTypeEnum Enum with underlying type: string
type CreateVirtualCircuitDetailsTypeEnum string

// Set of constants representing the allowable values for CreateVirtualCircuitDetailsTypeEnum
const (
	CreateVirtualCircuitDetailsTypePublic  CreateVirtualCircuitDetailsTypeEnum = "PUBLIC"
	CreateVirtualCircuitDetailsTypePrivate CreateVirtualCircuitDetailsTypeEnum = "PRIVATE"
)

var mappingCreateVirtualCircuitDetailsTypeEnum = map[string]CreateVirtualCircuitDetailsTypeEnum{
	"PUBLIC":  CreateVirtualCircuitDetailsTypePublic,
	"PRIVATE": CreateVirtualCircuitDetailsTypePrivate,
}

var mappingCreateVirtualCircuitDetailsTypeEnumLowerCase = map[string]CreateVirtualCircuitDetailsTypeEnum{
	"public":  CreateVirtualCircuitDetailsTypePublic,
	"private": CreateVirtualCircuitDetailsTypePrivate,
}

// GetCreateVirtualCircuitDetailsTypeEnumValues Enumerates the set of values for CreateVirtualCircuitDetailsTypeEnum
func GetCreateVirtualCircuitDetailsTypeEnumValues() []CreateVirtualCircuitDetailsTypeEnum {
	values := make([]CreateVirtualCircuitDetailsTypeEnum, 0)
	for _, v := range mappingCreateVirtualCircuitDetailsTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateVirtualCircuitDetailsTypeEnumStringValues Enumerates the set of values in String for CreateVirtualCircuitDetailsTypeEnum
func GetCreateVirtualCircuitDetailsTypeEnumStringValues() []string {
	return []string{
		"PUBLIC",
		"PRIVATE",
	}
}

// GetMappingCreateVirtualCircuitDetailsTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateVirtualCircuitDetailsTypeEnum(val string) (CreateVirtualCircuitDetailsTypeEnum, bool) {
	enum, ok := mappingCreateVirtualCircuitDetailsTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
