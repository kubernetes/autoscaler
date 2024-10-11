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

// VirtualCircuit For use with Oracle Cloud Infrastructure FastConnect.
// A virtual circuit is an isolated network path that runs over one or more physical
// network connections to provide a single, logical connection between the edge router
// on the customer's existing network and Oracle Cloud Infrastructure. *Private*
// virtual circuits support private peering, and *public* virtual circuits support
// public peering. For more information, see FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
// Each virtual circuit is made up of information shared between a customer, Oracle,
// and a provider (if the customer is using FastConnect via a provider). Who fills in
// a given property of a virtual circuit depends on whether the BGP session related to
// that virtual circuit goes from the customer's edge router to Oracle, or from the provider's
// edge router to Oracle. Also, in the case where the customer is using a provider, values
// for some of the properties may not be present immediately, but may get filled in as the
// provider and Oracle each do their part to provision the virtual circuit.
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type VirtualCircuit struct {

	// The provisioned data rate of the connection. To get a list of the
	// available bandwidth levels (that is, shapes), see
	// ListFastConnectProviderVirtualCircuitBandwidthShapes.
	// Example: `10 Gbps`
	BandwidthShapeName *string `mandatory:"false" json:"bandwidthShapeName"`

	// Deprecated. Instead use the information in
	// FastConnectProviderService.
	BgpManagement VirtualCircuitBgpManagementEnum `mandatory:"false" json:"bgpManagement,omitempty"`

	// The state of the Ipv4 BGP session associated with the virtual circuit.
	BgpSessionState VirtualCircuitBgpSessionStateEnum `mandatory:"false" json:"bgpSessionState,omitempty"`

	// The state of the Ipv6 BGP session associated with the virtual circuit.
	BgpIpv6SessionState VirtualCircuitBgpIpv6SessionStateEnum `mandatory:"false" json:"bgpIpv6SessionState,omitempty"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the virtual circuit.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// An array of mappings, each containing properties for a
	// cross-connect or cross-connect group that is associated with this
	// virtual circuit.
	CrossConnectMappings []CrossConnectMapping `mandatory:"false" json:"crossConnectMappings"`

	// The routing policy sets how routing information about the Oracle cloud is shared over a public virtual circuit.
	// Policies available are: `ORACLE_SERVICE_NETWORK`, `REGIONAL`, `MARKET_LEVEL`, and `GLOBAL`.
	// See Route Filtering (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/routingonprem.htm#route_filtering) for details.
	// By default, routing information is shared for all routes in the same market.
	RoutingPolicy []VirtualCircuitRoutingPolicyEnum `mandatory:"false" json:"routingPolicy,omitempty"`

	// Set to `ENABLED` (the default) to activate the BGP session of the virtual circuit, set to `DISABLED` to deactivate the virtual circuit.
	BgpAdminState VirtualCircuitBgpAdminStateEnum `mandatory:"false" json:"bgpAdminState,omitempty"`

	// Set to `true` to enable BFD for IPv4 BGP peering, or set to `false` to disable BFD. If this is not set, the default is `false`.
	IsBfdEnabled *bool `mandatory:"false" json:"isBfdEnabled"`

	// Set to `true` for the virtual circuit to carry only encrypted traffic, or set to `false` for the virtual circuit to carry unencrypted traffic. If this is not set, the default is `false`.
	IsTransportMode *bool `mandatory:"false" json:"isTransportMode"`

	// Deprecated. Instead use `customerAsn`.
	// If you specify values for both, the request will be rejected.
	CustomerBgpAsn *int `mandatory:"false" json:"customerBgpAsn"`

	// The BGP ASN of the network at the other end of the BGP
	// session from Oracle. If the session is between the customer's
	// edge router and Oracle, the value is the customer's ASN. If the BGP
	// session is between the provider's edge router and Oracle, the value
	// is the provider's ASN.
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

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the customer's Drg
	// that this virtual circuit uses. Applicable only to private virtual circuits.
	GatewayId *string `mandatory:"false" json:"gatewayId"`

	// The virtual circuit's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"false" json:"id"`

	// The virtual circuit's current state. For information about
	// the different states, see
	// FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
	LifecycleState VirtualCircuitLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The Oracle BGP ASN.
	OracleBgpAsn *int `mandatory:"false" json:"oracleBgpAsn"`

	// Deprecated. Instead use `providerServiceId`.
	ProviderName *string `mandatory:"false" json:"providerName"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the service offered by the provider (if the customer is connecting via a provider).
	ProviderServiceId *string `mandatory:"false" json:"providerServiceId"`

	// The service key name offered by the provider (if the customer is connecting via a provider).
	ProviderServiceKeyName *string `mandatory:"false" json:"providerServiceKeyName"`

	// Deprecated. Instead use `providerServiceId`.
	ProviderServiceName *string `mandatory:"false" json:"providerServiceName"`

	// The provider's state in relation to this virtual circuit (if the
	// customer is connecting via a provider). ACTIVE means
	// the provider has provisioned the virtual circuit from their end.
	// INACTIVE means the provider has not yet provisioned the virtual
	// circuit, or has de-provisioned it.
	ProviderState VirtualCircuitProviderStateEnum `mandatory:"false" json:"providerState,omitempty"`

	// For a public virtual circuit. The public IP prefixes (CIDRs) the customer wants to
	// advertise across the connection. All prefix sizes are allowed.
	PublicPrefixes []string `mandatory:"false" json:"publicPrefixes"`

	// Provider-supplied reference information about this virtual circuit
	// (if the customer is connecting via a provider).
	ReferenceComment *string `mandatory:"false" json:"referenceComment"`

	// The Oracle Cloud Infrastructure region where this virtual
	// circuit is located.
	Region *string `mandatory:"false" json:"region"`

	// Provider service type.
	ServiceType VirtualCircuitServiceTypeEnum `mandatory:"false" json:"serviceType,omitempty"`

	// The date and time the virtual circuit was created,
	// in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Whether the virtual circuit supports private or public peering. For more information,
	// see FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
	Type VirtualCircuitTypeEnum `mandatory:"false" json:"type,omitempty"`

	// The layer 3 IP MTU to use on this virtual circuit.
	IpMtu VirtualCircuitIpMtuEnum `mandatory:"false" json:"ipMtu,omitempty"`
}

func (m VirtualCircuit) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VirtualCircuit) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingVirtualCircuitBgpManagementEnum(string(m.BgpManagement)); !ok && m.BgpManagement != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BgpManagement: %s. Supported values are: %s.", m.BgpManagement, strings.Join(GetVirtualCircuitBgpManagementEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitBgpSessionStateEnum(string(m.BgpSessionState)); !ok && m.BgpSessionState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BgpSessionState: %s. Supported values are: %s.", m.BgpSessionState, strings.Join(GetVirtualCircuitBgpSessionStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitBgpIpv6SessionStateEnum(string(m.BgpIpv6SessionState)); !ok && m.BgpIpv6SessionState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BgpIpv6SessionState: %s. Supported values are: %s.", m.BgpIpv6SessionState, strings.Join(GetVirtualCircuitBgpIpv6SessionStateEnumStringValues(), ",")))
	}
	for _, val := range m.RoutingPolicy {
		if _, ok := GetMappingVirtualCircuitRoutingPolicyEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RoutingPolicy: %s. Supported values are: %s.", val, strings.Join(GetVirtualCircuitRoutingPolicyEnumStringValues(), ",")))
		}
	}

	if _, ok := GetMappingVirtualCircuitBgpAdminStateEnum(string(m.BgpAdminState)); !ok && m.BgpAdminState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BgpAdminState: %s. Supported values are: %s.", m.BgpAdminState, strings.Join(GetVirtualCircuitBgpAdminStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetVirtualCircuitLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitProviderStateEnum(string(m.ProviderState)); !ok && m.ProviderState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ProviderState: %s. Supported values are: %s.", m.ProviderState, strings.Join(GetVirtualCircuitProviderStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitServiceTypeEnum(string(m.ServiceType)); !ok && m.ServiceType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ServiceType: %s. Supported values are: %s.", m.ServiceType, strings.Join(GetVirtualCircuitServiceTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitTypeEnum(string(m.Type)); !ok && m.Type != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Type: %s. Supported values are: %s.", m.Type, strings.Join(GetVirtualCircuitTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVirtualCircuitIpMtuEnum(string(m.IpMtu)); !ok && m.IpMtu != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for IpMtu: %s. Supported values are: %s.", m.IpMtu, strings.Join(GetVirtualCircuitIpMtuEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// VirtualCircuitBgpManagementEnum Enum with underlying type: string
type VirtualCircuitBgpManagementEnum string

// Set of constants representing the allowable values for VirtualCircuitBgpManagementEnum
const (
	VirtualCircuitBgpManagementCustomerManaged VirtualCircuitBgpManagementEnum = "CUSTOMER_MANAGED"
	VirtualCircuitBgpManagementProviderManaged VirtualCircuitBgpManagementEnum = "PROVIDER_MANAGED"
	VirtualCircuitBgpManagementOracleManaged   VirtualCircuitBgpManagementEnum = "ORACLE_MANAGED"
)

var mappingVirtualCircuitBgpManagementEnum = map[string]VirtualCircuitBgpManagementEnum{
	"CUSTOMER_MANAGED": VirtualCircuitBgpManagementCustomerManaged,
	"PROVIDER_MANAGED": VirtualCircuitBgpManagementProviderManaged,
	"ORACLE_MANAGED":   VirtualCircuitBgpManagementOracleManaged,
}

var mappingVirtualCircuitBgpManagementEnumLowerCase = map[string]VirtualCircuitBgpManagementEnum{
	"customer_managed": VirtualCircuitBgpManagementCustomerManaged,
	"provider_managed": VirtualCircuitBgpManagementProviderManaged,
	"oracle_managed":   VirtualCircuitBgpManagementOracleManaged,
}

// GetVirtualCircuitBgpManagementEnumValues Enumerates the set of values for VirtualCircuitBgpManagementEnum
func GetVirtualCircuitBgpManagementEnumValues() []VirtualCircuitBgpManagementEnum {
	values := make([]VirtualCircuitBgpManagementEnum, 0)
	for _, v := range mappingVirtualCircuitBgpManagementEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitBgpManagementEnumStringValues Enumerates the set of values in String for VirtualCircuitBgpManagementEnum
func GetVirtualCircuitBgpManagementEnumStringValues() []string {
	return []string{
		"CUSTOMER_MANAGED",
		"PROVIDER_MANAGED",
		"ORACLE_MANAGED",
	}
}

// GetMappingVirtualCircuitBgpManagementEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitBgpManagementEnum(val string) (VirtualCircuitBgpManagementEnum, bool) {
	enum, ok := mappingVirtualCircuitBgpManagementEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VirtualCircuitBgpSessionStateEnum Enum with underlying type: string
type VirtualCircuitBgpSessionStateEnum string

// Set of constants representing the allowable values for VirtualCircuitBgpSessionStateEnum
const (
	VirtualCircuitBgpSessionStateUp   VirtualCircuitBgpSessionStateEnum = "UP"
	VirtualCircuitBgpSessionStateDown VirtualCircuitBgpSessionStateEnum = "DOWN"
)

var mappingVirtualCircuitBgpSessionStateEnum = map[string]VirtualCircuitBgpSessionStateEnum{
	"UP":   VirtualCircuitBgpSessionStateUp,
	"DOWN": VirtualCircuitBgpSessionStateDown,
}

var mappingVirtualCircuitBgpSessionStateEnumLowerCase = map[string]VirtualCircuitBgpSessionStateEnum{
	"up":   VirtualCircuitBgpSessionStateUp,
	"down": VirtualCircuitBgpSessionStateDown,
}

// GetVirtualCircuitBgpSessionStateEnumValues Enumerates the set of values for VirtualCircuitBgpSessionStateEnum
func GetVirtualCircuitBgpSessionStateEnumValues() []VirtualCircuitBgpSessionStateEnum {
	values := make([]VirtualCircuitBgpSessionStateEnum, 0)
	for _, v := range mappingVirtualCircuitBgpSessionStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitBgpSessionStateEnumStringValues Enumerates the set of values in String for VirtualCircuitBgpSessionStateEnum
func GetVirtualCircuitBgpSessionStateEnumStringValues() []string {
	return []string{
		"UP",
		"DOWN",
	}
}

// GetMappingVirtualCircuitBgpSessionStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitBgpSessionStateEnum(val string) (VirtualCircuitBgpSessionStateEnum, bool) {
	enum, ok := mappingVirtualCircuitBgpSessionStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VirtualCircuitBgpIpv6SessionStateEnum Enum with underlying type: string
type VirtualCircuitBgpIpv6SessionStateEnum string

// Set of constants representing the allowable values for VirtualCircuitBgpIpv6SessionStateEnum
const (
	VirtualCircuitBgpIpv6SessionStateUp   VirtualCircuitBgpIpv6SessionStateEnum = "UP"
	VirtualCircuitBgpIpv6SessionStateDown VirtualCircuitBgpIpv6SessionStateEnum = "DOWN"
)

var mappingVirtualCircuitBgpIpv6SessionStateEnum = map[string]VirtualCircuitBgpIpv6SessionStateEnum{
	"UP":   VirtualCircuitBgpIpv6SessionStateUp,
	"DOWN": VirtualCircuitBgpIpv6SessionStateDown,
}

var mappingVirtualCircuitBgpIpv6SessionStateEnumLowerCase = map[string]VirtualCircuitBgpIpv6SessionStateEnum{
	"up":   VirtualCircuitBgpIpv6SessionStateUp,
	"down": VirtualCircuitBgpIpv6SessionStateDown,
}

// GetVirtualCircuitBgpIpv6SessionStateEnumValues Enumerates the set of values for VirtualCircuitBgpIpv6SessionStateEnum
func GetVirtualCircuitBgpIpv6SessionStateEnumValues() []VirtualCircuitBgpIpv6SessionStateEnum {
	values := make([]VirtualCircuitBgpIpv6SessionStateEnum, 0)
	for _, v := range mappingVirtualCircuitBgpIpv6SessionStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitBgpIpv6SessionStateEnumStringValues Enumerates the set of values in String for VirtualCircuitBgpIpv6SessionStateEnum
func GetVirtualCircuitBgpIpv6SessionStateEnumStringValues() []string {
	return []string{
		"UP",
		"DOWN",
	}
}

// GetMappingVirtualCircuitBgpIpv6SessionStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitBgpIpv6SessionStateEnum(val string) (VirtualCircuitBgpIpv6SessionStateEnum, bool) {
	enum, ok := mappingVirtualCircuitBgpIpv6SessionStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VirtualCircuitRoutingPolicyEnum Enum with underlying type: string
type VirtualCircuitRoutingPolicyEnum string

// Set of constants representing the allowable values for VirtualCircuitRoutingPolicyEnum
const (
	VirtualCircuitRoutingPolicyOracleServiceNetwork VirtualCircuitRoutingPolicyEnum = "ORACLE_SERVICE_NETWORK"
	VirtualCircuitRoutingPolicyRegional             VirtualCircuitRoutingPolicyEnum = "REGIONAL"
	VirtualCircuitRoutingPolicyMarketLevel          VirtualCircuitRoutingPolicyEnum = "MARKET_LEVEL"
	VirtualCircuitRoutingPolicyGlobal               VirtualCircuitRoutingPolicyEnum = "GLOBAL"
)

var mappingVirtualCircuitRoutingPolicyEnum = map[string]VirtualCircuitRoutingPolicyEnum{
	"ORACLE_SERVICE_NETWORK": VirtualCircuitRoutingPolicyOracleServiceNetwork,
	"REGIONAL":               VirtualCircuitRoutingPolicyRegional,
	"MARKET_LEVEL":           VirtualCircuitRoutingPolicyMarketLevel,
	"GLOBAL":                 VirtualCircuitRoutingPolicyGlobal,
}

var mappingVirtualCircuitRoutingPolicyEnumLowerCase = map[string]VirtualCircuitRoutingPolicyEnum{
	"oracle_service_network": VirtualCircuitRoutingPolicyOracleServiceNetwork,
	"regional":               VirtualCircuitRoutingPolicyRegional,
	"market_level":           VirtualCircuitRoutingPolicyMarketLevel,
	"global":                 VirtualCircuitRoutingPolicyGlobal,
}

// GetVirtualCircuitRoutingPolicyEnumValues Enumerates the set of values for VirtualCircuitRoutingPolicyEnum
func GetVirtualCircuitRoutingPolicyEnumValues() []VirtualCircuitRoutingPolicyEnum {
	values := make([]VirtualCircuitRoutingPolicyEnum, 0)
	for _, v := range mappingVirtualCircuitRoutingPolicyEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitRoutingPolicyEnumStringValues Enumerates the set of values in String for VirtualCircuitRoutingPolicyEnum
func GetVirtualCircuitRoutingPolicyEnumStringValues() []string {
	return []string{
		"ORACLE_SERVICE_NETWORK",
		"REGIONAL",
		"MARKET_LEVEL",
		"GLOBAL",
	}
}

// GetMappingVirtualCircuitRoutingPolicyEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitRoutingPolicyEnum(val string) (VirtualCircuitRoutingPolicyEnum, bool) {
	enum, ok := mappingVirtualCircuitRoutingPolicyEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VirtualCircuitBgpAdminStateEnum Enum with underlying type: string
type VirtualCircuitBgpAdminStateEnum string

// Set of constants representing the allowable values for VirtualCircuitBgpAdminStateEnum
const (
	VirtualCircuitBgpAdminStateEnabled  VirtualCircuitBgpAdminStateEnum = "ENABLED"
	VirtualCircuitBgpAdminStateDisabled VirtualCircuitBgpAdminStateEnum = "DISABLED"
)

var mappingVirtualCircuitBgpAdminStateEnum = map[string]VirtualCircuitBgpAdminStateEnum{
	"ENABLED":  VirtualCircuitBgpAdminStateEnabled,
	"DISABLED": VirtualCircuitBgpAdminStateDisabled,
}

var mappingVirtualCircuitBgpAdminStateEnumLowerCase = map[string]VirtualCircuitBgpAdminStateEnum{
	"enabled":  VirtualCircuitBgpAdminStateEnabled,
	"disabled": VirtualCircuitBgpAdminStateDisabled,
}

// GetVirtualCircuitBgpAdminStateEnumValues Enumerates the set of values for VirtualCircuitBgpAdminStateEnum
func GetVirtualCircuitBgpAdminStateEnumValues() []VirtualCircuitBgpAdminStateEnum {
	values := make([]VirtualCircuitBgpAdminStateEnum, 0)
	for _, v := range mappingVirtualCircuitBgpAdminStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitBgpAdminStateEnumStringValues Enumerates the set of values in String for VirtualCircuitBgpAdminStateEnum
func GetVirtualCircuitBgpAdminStateEnumStringValues() []string {
	return []string{
		"ENABLED",
		"DISABLED",
	}
}

// GetMappingVirtualCircuitBgpAdminStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitBgpAdminStateEnum(val string) (VirtualCircuitBgpAdminStateEnum, bool) {
	enum, ok := mappingVirtualCircuitBgpAdminStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VirtualCircuitLifecycleStateEnum Enum with underlying type: string
type VirtualCircuitLifecycleStateEnum string

// Set of constants representing the allowable values for VirtualCircuitLifecycleStateEnum
const (
	VirtualCircuitLifecycleStatePendingProvider VirtualCircuitLifecycleStateEnum = "PENDING_PROVIDER"
	VirtualCircuitLifecycleStateVerifying       VirtualCircuitLifecycleStateEnum = "VERIFYING"
	VirtualCircuitLifecycleStateProvisioning    VirtualCircuitLifecycleStateEnum = "PROVISIONING"
	VirtualCircuitLifecycleStateProvisioned     VirtualCircuitLifecycleStateEnum = "PROVISIONED"
	VirtualCircuitLifecycleStateFailed          VirtualCircuitLifecycleStateEnum = "FAILED"
	VirtualCircuitLifecycleStateInactive        VirtualCircuitLifecycleStateEnum = "INACTIVE"
	VirtualCircuitLifecycleStateTerminating     VirtualCircuitLifecycleStateEnum = "TERMINATING"
	VirtualCircuitLifecycleStateTerminated      VirtualCircuitLifecycleStateEnum = "TERMINATED"
)

var mappingVirtualCircuitLifecycleStateEnum = map[string]VirtualCircuitLifecycleStateEnum{
	"PENDING_PROVIDER": VirtualCircuitLifecycleStatePendingProvider,
	"VERIFYING":        VirtualCircuitLifecycleStateVerifying,
	"PROVISIONING":     VirtualCircuitLifecycleStateProvisioning,
	"PROVISIONED":      VirtualCircuitLifecycleStateProvisioned,
	"FAILED":           VirtualCircuitLifecycleStateFailed,
	"INACTIVE":         VirtualCircuitLifecycleStateInactive,
	"TERMINATING":      VirtualCircuitLifecycleStateTerminating,
	"TERMINATED":       VirtualCircuitLifecycleStateTerminated,
}

var mappingVirtualCircuitLifecycleStateEnumLowerCase = map[string]VirtualCircuitLifecycleStateEnum{
	"pending_provider": VirtualCircuitLifecycleStatePendingProvider,
	"verifying":        VirtualCircuitLifecycleStateVerifying,
	"provisioning":     VirtualCircuitLifecycleStateProvisioning,
	"provisioned":      VirtualCircuitLifecycleStateProvisioned,
	"failed":           VirtualCircuitLifecycleStateFailed,
	"inactive":         VirtualCircuitLifecycleStateInactive,
	"terminating":      VirtualCircuitLifecycleStateTerminating,
	"terminated":       VirtualCircuitLifecycleStateTerminated,
}

// GetVirtualCircuitLifecycleStateEnumValues Enumerates the set of values for VirtualCircuitLifecycleStateEnum
func GetVirtualCircuitLifecycleStateEnumValues() []VirtualCircuitLifecycleStateEnum {
	values := make([]VirtualCircuitLifecycleStateEnum, 0)
	for _, v := range mappingVirtualCircuitLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitLifecycleStateEnumStringValues Enumerates the set of values in String for VirtualCircuitLifecycleStateEnum
func GetVirtualCircuitLifecycleStateEnumStringValues() []string {
	return []string{
		"PENDING_PROVIDER",
		"VERIFYING",
		"PROVISIONING",
		"PROVISIONED",
		"FAILED",
		"INACTIVE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingVirtualCircuitLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitLifecycleStateEnum(val string) (VirtualCircuitLifecycleStateEnum, bool) {
	enum, ok := mappingVirtualCircuitLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VirtualCircuitProviderStateEnum Enum with underlying type: string
type VirtualCircuitProviderStateEnum string

// Set of constants representing the allowable values for VirtualCircuitProviderStateEnum
const (
	VirtualCircuitProviderStateActive   VirtualCircuitProviderStateEnum = "ACTIVE"
	VirtualCircuitProviderStateInactive VirtualCircuitProviderStateEnum = "INACTIVE"
)

var mappingVirtualCircuitProviderStateEnum = map[string]VirtualCircuitProviderStateEnum{
	"ACTIVE":   VirtualCircuitProviderStateActive,
	"INACTIVE": VirtualCircuitProviderStateInactive,
}

var mappingVirtualCircuitProviderStateEnumLowerCase = map[string]VirtualCircuitProviderStateEnum{
	"active":   VirtualCircuitProviderStateActive,
	"inactive": VirtualCircuitProviderStateInactive,
}

// GetVirtualCircuitProviderStateEnumValues Enumerates the set of values for VirtualCircuitProviderStateEnum
func GetVirtualCircuitProviderStateEnumValues() []VirtualCircuitProviderStateEnum {
	values := make([]VirtualCircuitProviderStateEnum, 0)
	for _, v := range mappingVirtualCircuitProviderStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitProviderStateEnumStringValues Enumerates the set of values in String for VirtualCircuitProviderStateEnum
func GetVirtualCircuitProviderStateEnumStringValues() []string {
	return []string{
		"ACTIVE",
		"INACTIVE",
	}
}

// GetMappingVirtualCircuitProviderStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitProviderStateEnum(val string) (VirtualCircuitProviderStateEnum, bool) {
	enum, ok := mappingVirtualCircuitProviderStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VirtualCircuitServiceTypeEnum Enum with underlying type: string
type VirtualCircuitServiceTypeEnum string

// Set of constants representing the allowable values for VirtualCircuitServiceTypeEnum
const (
	VirtualCircuitServiceTypeColocated VirtualCircuitServiceTypeEnum = "COLOCATED"
	VirtualCircuitServiceTypeLayer2    VirtualCircuitServiceTypeEnum = "LAYER2"
	VirtualCircuitServiceTypeLayer3    VirtualCircuitServiceTypeEnum = "LAYER3"
)

var mappingVirtualCircuitServiceTypeEnum = map[string]VirtualCircuitServiceTypeEnum{
	"COLOCATED": VirtualCircuitServiceTypeColocated,
	"LAYER2":    VirtualCircuitServiceTypeLayer2,
	"LAYER3":    VirtualCircuitServiceTypeLayer3,
}

var mappingVirtualCircuitServiceTypeEnumLowerCase = map[string]VirtualCircuitServiceTypeEnum{
	"colocated": VirtualCircuitServiceTypeColocated,
	"layer2":    VirtualCircuitServiceTypeLayer2,
	"layer3":    VirtualCircuitServiceTypeLayer3,
}

// GetVirtualCircuitServiceTypeEnumValues Enumerates the set of values for VirtualCircuitServiceTypeEnum
func GetVirtualCircuitServiceTypeEnumValues() []VirtualCircuitServiceTypeEnum {
	values := make([]VirtualCircuitServiceTypeEnum, 0)
	for _, v := range mappingVirtualCircuitServiceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitServiceTypeEnumStringValues Enumerates the set of values in String for VirtualCircuitServiceTypeEnum
func GetVirtualCircuitServiceTypeEnumStringValues() []string {
	return []string{
		"COLOCATED",
		"LAYER2",
		"LAYER3",
	}
}

// GetMappingVirtualCircuitServiceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitServiceTypeEnum(val string) (VirtualCircuitServiceTypeEnum, bool) {
	enum, ok := mappingVirtualCircuitServiceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VirtualCircuitTypeEnum Enum with underlying type: string
type VirtualCircuitTypeEnum string

// Set of constants representing the allowable values for VirtualCircuitTypeEnum
const (
	VirtualCircuitTypePublic  VirtualCircuitTypeEnum = "PUBLIC"
	VirtualCircuitTypePrivate VirtualCircuitTypeEnum = "PRIVATE"
)

var mappingVirtualCircuitTypeEnum = map[string]VirtualCircuitTypeEnum{
	"PUBLIC":  VirtualCircuitTypePublic,
	"PRIVATE": VirtualCircuitTypePrivate,
}

var mappingVirtualCircuitTypeEnumLowerCase = map[string]VirtualCircuitTypeEnum{
	"public":  VirtualCircuitTypePublic,
	"private": VirtualCircuitTypePrivate,
}

// GetVirtualCircuitTypeEnumValues Enumerates the set of values for VirtualCircuitTypeEnum
func GetVirtualCircuitTypeEnumValues() []VirtualCircuitTypeEnum {
	values := make([]VirtualCircuitTypeEnum, 0)
	for _, v := range mappingVirtualCircuitTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitTypeEnumStringValues Enumerates the set of values in String for VirtualCircuitTypeEnum
func GetVirtualCircuitTypeEnumStringValues() []string {
	return []string{
		"PUBLIC",
		"PRIVATE",
	}
}

// GetMappingVirtualCircuitTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitTypeEnum(val string) (VirtualCircuitTypeEnum, bool) {
	enum, ok := mappingVirtualCircuitTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
