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

// UpdateVirtualCircuitDetails The representation of UpdateVirtualCircuitDetails
type UpdateVirtualCircuitDetails struct {

	// The provisioned data rate of the connection. To get a list of the
	// available bandwidth levels (that is, shapes), see
	// ListFastConnectProviderVirtualCircuitBandwidthShapes.
	// To be updated only by the customer who owns the virtual circuit.
	BandwidthShapeName *string `mandatory:"false" json:"bandwidthShapeName"`

	// An array of mappings, each containing properties for a cross-connect or
	// cross-connect group associated with this virtual circuit.
	// The customer and provider can update different properties in the mapping
	// depending on the situation. See the description of the
	// CrossConnectMapping.
	CrossConnectMappings []CrossConnectMapping `mandatory:"false" json:"crossConnectMappings"`

	// The routing policy sets how routing information about the Oracle cloud is shared over a public virtual circuit.
	// Policies available are: `ORACLE_SERVICE_NETWORK`, `REGIONAL`, `MARKET_LEVEL`, and `GLOBAL`.
	// See Route Filtering (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/routingonprem.htm#route_filtering) for details.
	// By default, routing information is shared for all routes in the same market.
	RoutingPolicy []UpdateVirtualCircuitDetailsRoutingPolicyEnum `mandatory:"false" json:"routingPolicy,omitempty"`

	// Set to `ENABLED` (the default) to activate the BGP session of the virtual circuit, set to `DISABLED` to deactivate the virtual circuit.
	BgpAdminState UpdateVirtualCircuitDetailsBgpAdminStateEnum `mandatory:"false" json:"bgpAdminState,omitempty"`

	// Set to `true` to enable BFD for IPv4 BGP peering, or set to `false` to disable BFD. If this is not set, the default is `false`.
	IsBfdEnabled *bool `mandatory:"false" json:"isBfdEnabled"`

	// Set to `true` for the virtual circuit to carry only encrypted traffic, or set to `false` for the virtual circuit to carry unencrypted traffic. If this is not set, the default is `false`.
	IsTransportMode *bool `mandatory:"false" json:"isTransportMode"`

	// Deprecated. Instead use `customerAsn`.
	// If you specify values for both, the request will be rejected.
	CustomerBgpAsn *int `mandatory:"false" json:"customerBgpAsn"`

	// The BGP ASN of the network at the other end of the BGP
	// session from Oracle.
	// If the BGP session is from the customer's edge router to Oracle, the
	// required value is the customer's ASN, and it can be updated only
	// by the customer.
	// If the BGP session is from the provider's edge router to Oracle, the
	// required value is the provider's ASN, and it can be updated only
	// by the provider.
	// Can be a 2-byte or 4-byte ASN. Uses "asplain" format.
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

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the Drg
	// that this private virtual circuit uses.
	// To be updated only by the customer who owns the virtual circuit.
	GatewayId *string `mandatory:"false" json:"gatewayId"`

	// The provider's state in relation to this virtual circuit. Relevant only
	// if the customer is using FastConnect via a provider. ACTIVE
	// means the provider has provisioned the virtual circuit from their
	// end. INACTIVE means the provider has not yet provisioned the virtual
	// circuit, or has de-provisioned it.
	// To be updated only by the provider.
	ProviderState UpdateVirtualCircuitDetailsProviderStateEnum `mandatory:"false" json:"providerState,omitempty"`

	// The service key name offered by the provider (if the customer is connecting via a provider).
	ProviderServiceKeyName *string `mandatory:"false" json:"providerServiceKeyName"`

	// Provider-supplied reference information about this virtual circuit.
	// Relevant only if the customer is using FastConnect via a provider.
	// To be updated only by the provider.
	ReferenceComment *string `mandatory:"false" json:"referenceComment"`

	// The layer 3 IP MTU to use on this virtual circuit.
	IpMtu VirtualCircuitIpMtuEnum `mandatory:"false" json:"ipMtu,omitempty"`
}

func (m UpdateVirtualCircuitDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateVirtualCircuitDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	for _, val := range m.RoutingPolicy {
		if _, ok := GetMappingUpdateVirtualCircuitDetailsRoutingPolicyEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RoutingPolicy: %s. Supported values are: %s.", val, strings.Join(GetUpdateVirtualCircuitDetailsRoutingPolicyEnumStringValues(), ",")))
		}
	}

	if _, ok := GetMappingUpdateVirtualCircuitDetailsBgpAdminStateEnum(string(m.BgpAdminState)); !ok && m.BgpAdminState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BgpAdminState: %s. Supported values are: %s.", m.BgpAdminState, strings.Join(GetUpdateVirtualCircuitDetailsBgpAdminStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingUpdateVirtualCircuitDetailsProviderStateEnum(string(m.ProviderState)); !ok && m.ProviderState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ProviderState: %s. Supported values are: %s.", m.ProviderState, strings.Join(GetUpdateVirtualCircuitDetailsProviderStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitIpMtuEnum(string(m.IpMtu)); !ok && m.IpMtu != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for IpMtu: %s. Supported values are: %s.", m.IpMtu, strings.Join(GetVirtualCircuitIpMtuEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UpdateVirtualCircuitDetailsRoutingPolicyEnum Enum with underlying type: string
type UpdateVirtualCircuitDetailsRoutingPolicyEnum string

// Set of constants representing the allowable values for UpdateVirtualCircuitDetailsRoutingPolicyEnum
const (
	UpdateVirtualCircuitDetailsRoutingPolicyOracleServiceNetwork UpdateVirtualCircuitDetailsRoutingPolicyEnum = "ORACLE_SERVICE_NETWORK"
	UpdateVirtualCircuitDetailsRoutingPolicyRegional             UpdateVirtualCircuitDetailsRoutingPolicyEnum = "REGIONAL"
	UpdateVirtualCircuitDetailsRoutingPolicyMarketLevel          UpdateVirtualCircuitDetailsRoutingPolicyEnum = "MARKET_LEVEL"
	UpdateVirtualCircuitDetailsRoutingPolicyGlobal               UpdateVirtualCircuitDetailsRoutingPolicyEnum = "GLOBAL"
)

var mappingUpdateVirtualCircuitDetailsRoutingPolicyEnum = map[string]UpdateVirtualCircuitDetailsRoutingPolicyEnum{
	"ORACLE_SERVICE_NETWORK": UpdateVirtualCircuitDetailsRoutingPolicyOracleServiceNetwork,
	"REGIONAL":               UpdateVirtualCircuitDetailsRoutingPolicyRegional,
	"MARKET_LEVEL":           UpdateVirtualCircuitDetailsRoutingPolicyMarketLevel,
	"GLOBAL":                 UpdateVirtualCircuitDetailsRoutingPolicyGlobal,
}

var mappingUpdateVirtualCircuitDetailsRoutingPolicyEnumLowerCase = map[string]UpdateVirtualCircuitDetailsRoutingPolicyEnum{
	"oracle_service_network": UpdateVirtualCircuitDetailsRoutingPolicyOracleServiceNetwork,
	"regional":               UpdateVirtualCircuitDetailsRoutingPolicyRegional,
	"market_level":           UpdateVirtualCircuitDetailsRoutingPolicyMarketLevel,
	"global":                 UpdateVirtualCircuitDetailsRoutingPolicyGlobal,
}

// GetUpdateVirtualCircuitDetailsRoutingPolicyEnumValues Enumerates the set of values for UpdateVirtualCircuitDetailsRoutingPolicyEnum
func GetUpdateVirtualCircuitDetailsRoutingPolicyEnumValues() []UpdateVirtualCircuitDetailsRoutingPolicyEnum {
	values := make([]UpdateVirtualCircuitDetailsRoutingPolicyEnum, 0)
	for _, v := range mappingUpdateVirtualCircuitDetailsRoutingPolicyEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateVirtualCircuitDetailsRoutingPolicyEnumStringValues Enumerates the set of values in String for UpdateVirtualCircuitDetailsRoutingPolicyEnum
func GetUpdateVirtualCircuitDetailsRoutingPolicyEnumStringValues() []string {
	return []string{
		"ORACLE_SERVICE_NETWORK",
		"REGIONAL",
		"MARKET_LEVEL",
		"GLOBAL",
	}
}

// GetMappingUpdateVirtualCircuitDetailsRoutingPolicyEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateVirtualCircuitDetailsRoutingPolicyEnum(val string) (UpdateVirtualCircuitDetailsRoutingPolicyEnum, bool) {
	enum, ok := mappingUpdateVirtualCircuitDetailsRoutingPolicyEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// UpdateVirtualCircuitDetailsBgpAdminStateEnum Enum with underlying type: string
type UpdateVirtualCircuitDetailsBgpAdminStateEnum string

// Set of constants representing the allowable values for UpdateVirtualCircuitDetailsBgpAdminStateEnum
const (
	UpdateVirtualCircuitDetailsBgpAdminStateEnabled  UpdateVirtualCircuitDetailsBgpAdminStateEnum = "ENABLED"
	UpdateVirtualCircuitDetailsBgpAdminStateDisabled UpdateVirtualCircuitDetailsBgpAdminStateEnum = "DISABLED"
)

var mappingUpdateVirtualCircuitDetailsBgpAdminStateEnum = map[string]UpdateVirtualCircuitDetailsBgpAdminStateEnum{
	"ENABLED":  UpdateVirtualCircuitDetailsBgpAdminStateEnabled,
	"DISABLED": UpdateVirtualCircuitDetailsBgpAdminStateDisabled,
}

var mappingUpdateVirtualCircuitDetailsBgpAdminStateEnumLowerCase = map[string]UpdateVirtualCircuitDetailsBgpAdminStateEnum{
	"enabled":  UpdateVirtualCircuitDetailsBgpAdminStateEnabled,
	"disabled": UpdateVirtualCircuitDetailsBgpAdminStateDisabled,
}

// GetUpdateVirtualCircuitDetailsBgpAdminStateEnumValues Enumerates the set of values for UpdateVirtualCircuitDetailsBgpAdminStateEnum
func GetUpdateVirtualCircuitDetailsBgpAdminStateEnumValues() []UpdateVirtualCircuitDetailsBgpAdminStateEnum {
	values := make([]UpdateVirtualCircuitDetailsBgpAdminStateEnum, 0)
	for _, v := range mappingUpdateVirtualCircuitDetailsBgpAdminStateEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateVirtualCircuitDetailsBgpAdminStateEnumStringValues Enumerates the set of values in String for UpdateVirtualCircuitDetailsBgpAdminStateEnum
func GetUpdateVirtualCircuitDetailsBgpAdminStateEnumStringValues() []string {
	return []string{
		"ENABLED",
		"DISABLED",
	}
}

// GetMappingUpdateVirtualCircuitDetailsBgpAdminStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateVirtualCircuitDetailsBgpAdminStateEnum(val string) (UpdateVirtualCircuitDetailsBgpAdminStateEnum, bool) {
	enum, ok := mappingUpdateVirtualCircuitDetailsBgpAdminStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// UpdateVirtualCircuitDetailsProviderStateEnum Enum with underlying type: string
type UpdateVirtualCircuitDetailsProviderStateEnum string

// Set of constants representing the allowable values for UpdateVirtualCircuitDetailsProviderStateEnum
const (
	UpdateVirtualCircuitDetailsProviderStateActive   UpdateVirtualCircuitDetailsProviderStateEnum = "ACTIVE"
	UpdateVirtualCircuitDetailsProviderStateInactive UpdateVirtualCircuitDetailsProviderStateEnum = "INACTIVE"
)

var mappingUpdateVirtualCircuitDetailsProviderStateEnum = map[string]UpdateVirtualCircuitDetailsProviderStateEnum{
	"ACTIVE":   UpdateVirtualCircuitDetailsProviderStateActive,
	"INACTIVE": UpdateVirtualCircuitDetailsProviderStateInactive,
}

var mappingUpdateVirtualCircuitDetailsProviderStateEnumLowerCase = map[string]UpdateVirtualCircuitDetailsProviderStateEnum{
	"active":   UpdateVirtualCircuitDetailsProviderStateActive,
	"inactive": UpdateVirtualCircuitDetailsProviderStateInactive,
}

// GetUpdateVirtualCircuitDetailsProviderStateEnumValues Enumerates the set of values for UpdateVirtualCircuitDetailsProviderStateEnum
func GetUpdateVirtualCircuitDetailsProviderStateEnumValues() []UpdateVirtualCircuitDetailsProviderStateEnum {
	values := make([]UpdateVirtualCircuitDetailsProviderStateEnum, 0)
	for _, v := range mappingUpdateVirtualCircuitDetailsProviderStateEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateVirtualCircuitDetailsProviderStateEnumStringValues Enumerates the set of values in String for UpdateVirtualCircuitDetailsProviderStateEnum
func GetUpdateVirtualCircuitDetailsProviderStateEnumStringValues() []string {
	return []string{
		"ACTIVE",
		"INACTIVE",
	}
}

// GetMappingUpdateVirtualCircuitDetailsProviderStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateVirtualCircuitDetailsProviderStateEnum(val string) (UpdateVirtualCircuitDetailsProviderStateEnum, bool) {
	enum, ok := mappingUpdateVirtualCircuitDetailsProviderStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
