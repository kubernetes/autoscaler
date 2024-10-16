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

// IpSecConnectionTunnel Information about a single IPSec tunnel in an IPSec connection. This object does not include the tunnel's
// shared secret (pre-shared key), which is found in the
// IPSecConnectionTunnelSharedSecret object.
type IpSecConnectionTunnel struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the tunnel.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the tunnel.
	Id *string `mandatory:"true" json:"id"`

	// The tunnel's lifecycle state.
	LifecycleState IpSecConnectionTunnelLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The IP address of the Oracle VPN headend for the connection.
	// Example: `203.0.113.21`
	VpnIp *string `mandatory:"false" json:"vpnIp"`

	// The IP address of the CPE device's VPN headend.
	// Example: `203.0.113.22`
	CpeIp *string `mandatory:"false" json:"cpeIp"`

	// The status of the tunnel based on IPSec protocol characteristics.
	Status IpSecConnectionTunnelStatusEnum `mandatory:"false" json:"status,omitempty"`

	// Internet Key Exchange protocol version.
	IkeVersion IpSecConnectionTunnelIkeVersionEnum `mandatory:"false" json:"ikeVersion,omitempty"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	BgpSessionInfo *BgpSessionInfo `mandatory:"false" json:"bgpSessionInfo"`

	EncryptionDomainConfig *EncryptionDomainConfig `mandatory:"false" json:"encryptionDomainConfig"`

	// The type of routing used for this tunnel (BGP dynamic routing, static routing, or policy-based routing).
	Routing IpSecConnectionTunnelRoutingEnum `mandatory:"false" json:"routing,omitempty"`

	// The date and time the IPSec tunnel was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// When the status of the IPSec tunnel last changed, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeStatusUpdated *common.SDKTime `mandatory:"false" json:"timeStatusUpdated"`

	// Indicates whether Oracle can only respond to a request to start an IPSec tunnel from the CPE device, or both respond to and initiate requests.
	OracleCanInitiate IpSecConnectionTunnelOracleCanInitiateEnum `mandatory:"false" json:"oracleCanInitiate,omitempty"`

	// By default (the `AUTO` setting), IKE sends packets with a source and destination port set to 500,
	// and when it detects that the port used to forward packets has changed (most likely because a NAT device
	// is between the CPE device and the Oracle VPN headend) it will try to negotiate the use of NAT-T.
	// The `ENABLED` option sets the IKE protocol to use port 4500 instead of 500 and forces encapsulating traffic with the ESP protocol inside UDP packets.
	// The `DISABLED` option directs IKE to completely refuse to negotiate NAT-T
	// even if it senses there may be a NAT device in use.
	//
	// .
	NatTranslationEnabled IpSecConnectionTunnelNatTranslationEnabledEnum `mandatory:"false" json:"natTranslationEnabled,omitempty"`

	// Dead peer detection (DPD) mode set on the Oracle side of the connection.
	// This mode sets whether Oracle can only respond to a request from the CPE device to start DPD,
	// or both respond to and initiate requests.
	DpdMode IpSecConnectionTunnelDpdModeEnum `mandatory:"false" json:"dpdMode,omitempty"`

	// DPD timeout in seconds.
	DpdTimeoutInSec *int `mandatory:"false" json:"dpdTimeoutInSec"`

	PhaseOneDetails *TunnelPhaseOneDetails `mandatory:"false" json:"phaseOneDetails"`

	PhaseTwoDetails *TunnelPhaseTwoDetails `mandatory:"false" json:"phaseTwoDetails"`

	// The list of virtual circuit OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)s over which your network can reach this tunnel.
	AssociatedVirtualCircuits []string `mandatory:"false" json:"associatedVirtualCircuits"`
}

func (m IpSecConnectionTunnel) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m IpSecConnectionTunnel) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingIpSecConnectionTunnelLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetIpSecConnectionTunnelLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := GetMappingIpSecConnectionTunnelStatusEnum(string(m.Status)); !ok && m.Status != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Status: %s. Supported values are: %s.", m.Status, strings.Join(GetIpSecConnectionTunnelStatusEnumStringValues(), ",")))
	}
	if _, ok := GetMappingIpSecConnectionTunnelIkeVersionEnum(string(m.IkeVersion)); !ok && m.IkeVersion != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for IkeVersion: %s. Supported values are: %s.", m.IkeVersion, strings.Join(GetIpSecConnectionTunnelIkeVersionEnumStringValues(), ",")))
	}
	if _, ok := GetMappingIpSecConnectionTunnelRoutingEnum(string(m.Routing)); !ok && m.Routing != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Routing: %s. Supported values are: %s.", m.Routing, strings.Join(GetIpSecConnectionTunnelRoutingEnumStringValues(), ",")))
	}
	if _, ok := GetMappingIpSecConnectionTunnelOracleCanInitiateEnum(string(m.OracleCanInitiate)); !ok && m.OracleCanInitiate != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for OracleCanInitiate: %s. Supported values are: %s.", m.OracleCanInitiate, strings.Join(GetIpSecConnectionTunnelOracleCanInitiateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingIpSecConnectionTunnelNatTranslationEnabledEnum(string(m.NatTranslationEnabled)); !ok && m.NatTranslationEnabled != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for NatTranslationEnabled: %s. Supported values are: %s.", m.NatTranslationEnabled, strings.Join(GetIpSecConnectionTunnelNatTranslationEnabledEnumStringValues(), ",")))
	}
	if _, ok := GetMappingIpSecConnectionTunnelDpdModeEnum(string(m.DpdMode)); !ok && m.DpdMode != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DpdMode: %s. Supported values are: %s.", m.DpdMode, strings.Join(GetIpSecConnectionTunnelDpdModeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// IpSecConnectionTunnelStatusEnum Enum with underlying type: string
type IpSecConnectionTunnelStatusEnum string

// Set of constants representing the allowable values for IpSecConnectionTunnelStatusEnum
const (
	IpSecConnectionTunnelStatusUp                 IpSecConnectionTunnelStatusEnum = "UP"
	IpSecConnectionTunnelStatusDown               IpSecConnectionTunnelStatusEnum = "DOWN"
	IpSecConnectionTunnelStatusDownForMaintenance IpSecConnectionTunnelStatusEnum = "DOWN_FOR_MAINTENANCE"
	IpSecConnectionTunnelStatusPartialUp          IpSecConnectionTunnelStatusEnum = "PARTIAL_UP"
)

var mappingIpSecConnectionTunnelStatusEnum = map[string]IpSecConnectionTunnelStatusEnum{
	"UP":                   IpSecConnectionTunnelStatusUp,
	"DOWN":                 IpSecConnectionTunnelStatusDown,
	"DOWN_FOR_MAINTENANCE": IpSecConnectionTunnelStatusDownForMaintenance,
	"PARTIAL_UP":           IpSecConnectionTunnelStatusPartialUp,
}

var mappingIpSecConnectionTunnelStatusEnumLowerCase = map[string]IpSecConnectionTunnelStatusEnum{
	"up":                   IpSecConnectionTunnelStatusUp,
	"down":                 IpSecConnectionTunnelStatusDown,
	"down_for_maintenance": IpSecConnectionTunnelStatusDownForMaintenance,
	"partial_up":           IpSecConnectionTunnelStatusPartialUp,
}

// GetIpSecConnectionTunnelStatusEnumValues Enumerates the set of values for IpSecConnectionTunnelStatusEnum
func GetIpSecConnectionTunnelStatusEnumValues() []IpSecConnectionTunnelStatusEnum {
	values := make([]IpSecConnectionTunnelStatusEnum, 0)
	for _, v := range mappingIpSecConnectionTunnelStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetIpSecConnectionTunnelStatusEnumStringValues Enumerates the set of values in String for IpSecConnectionTunnelStatusEnum
func GetIpSecConnectionTunnelStatusEnumStringValues() []string {
	return []string{
		"UP",
		"DOWN",
		"DOWN_FOR_MAINTENANCE",
		"PARTIAL_UP",
	}
}

// GetMappingIpSecConnectionTunnelStatusEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpSecConnectionTunnelStatusEnum(val string) (IpSecConnectionTunnelStatusEnum, bool) {
	enum, ok := mappingIpSecConnectionTunnelStatusEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpSecConnectionTunnelIkeVersionEnum Enum with underlying type: string
type IpSecConnectionTunnelIkeVersionEnum string

// Set of constants representing the allowable values for IpSecConnectionTunnelIkeVersionEnum
const (
	IpSecConnectionTunnelIkeVersionV1 IpSecConnectionTunnelIkeVersionEnum = "V1"
	IpSecConnectionTunnelIkeVersionV2 IpSecConnectionTunnelIkeVersionEnum = "V2"
)

var mappingIpSecConnectionTunnelIkeVersionEnum = map[string]IpSecConnectionTunnelIkeVersionEnum{
	"V1": IpSecConnectionTunnelIkeVersionV1,
	"V2": IpSecConnectionTunnelIkeVersionV2,
}

var mappingIpSecConnectionTunnelIkeVersionEnumLowerCase = map[string]IpSecConnectionTunnelIkeVersionEnum{
	"v1": IpSecConnectionTunnelIkeVersionV1,
	"v2": IpSecConnectionTunnelIkeVersionV2,
}

// GetIpSecConnectionTunnelIkeVersionEnumValues Enumerates the set of values for IpSecConnectionTunnelIkeVersionEnum
func GetIpSecConnectionTunnelIkeVersionEnumValues() []IpSecConnectionTunnelIkeVersionEnum {
	values := make([]IpSecConnectionTunnelIkeVersionEnum, 0)
	for _, v := range mappingIpSecConnectionTunnelIkeVersionEnum {
		values = append(values, v)
	}
	return values
}

// GetIpSecConnectionTunnelIkeVersionEnumStringValues Enumerates the set of values in String for IpSecConnectionTunnelIkeVersionEnum
func GetIpSecConnectionTunnelIkeVersionEnumStringValues() []string {
	return []string{
		"V1",
		"V2",
	}
}

// GetMappingIpSecConnectionTunnelIkeVersionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpSecConnectionTunnelIkeVersionEnum(val string) (IpSecConnectionTunnelIkeVersionEnum, bool) {
	enum, ok := mappingIpSecConnectionTunnelIkeVersionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpSecConnectionTunnelLifecycleStateEnum Enum with underlying type: string
type IpSecConnectionTunnelLifecycleStateEnum string

// Set of constants representing the allowable values for IpSecConnectionTunnelLifecycleStateEnum
const (
	IpSecConnectionTunnelLifecycleStateProvisioning IpSecConnectionTunnelLifecycleStateEnum = "PROVISIONING"
	IpSecConnectionTunnelLifecycleStateAvailable    IpSecConnectionTunnelLifecycleStateEnum = "AVAILABLE"
	IpSecConnectionTunnelLifecycleStateTerminating  IpSecConnectionTunnelLifecycleStateEnum = "TERMINATING"
	IpSecConnectionTunnelLifecycleStateTerminated   IpSecConnectionTunnelLifecycleStateEnum = "TERMINATED"
)

var mappingIpSecConnectionTunnelLifecycleStateEnum = map[string]IpSecConnectionTunnelLifecycleStateEnum{
	"PROVISIONING": IpSecConnectionTunnelLifecycleStateProvisioning,
	"AVAILABLE":    IpSecConnectionTunnelLifecycleStateAvailable,
	"TERMINATING":  IpSecConnectionTunnelLifecycleStateTerminating,
	"TERMINATED":   IpSecConnectionTunnelLifecycleStateTerminated,
}

var mappingIpSecConnectionTunnelLifecycleStateEnumLowerCase = map[string]IpSecConnectionTunnelLifecycleStateEnum{
	"provisioning": IpSecConnectionTunnelLifecycleStateProvisioning,
	"available":    IpSecConnectionTunnelLifecycleStateAvailable,
	"terminating":  IpSecConnectionTunnelLifecycleStateTerminating,
	"terminated":   IpSecConnectionTunnelLifecycleStateTerminated,
}

// GetIpSecConnectionTunnelLifecycleStateEnumValues Enumerates the set of values for IpSecConnectionTunnelLifecycleStateEnum
func GetIpSecConnectionTunnelLifecycleStateEnumValues() []IpSecConnectionTunnelLifecycleStateEnum {
	values := make([]IpSecConnectionTunnelLifecycleStateEnum, 0)
	for _, v := range mappingIpSecConnectionTunnelLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetIpSecConnectionTunnelLifecycleStateEnumStringValues Enumerates the set of values in String for IpSecConnectionTunnelLifecycleStateEnum
func GetIpSecConnectionTunnelLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingIpSecConnectionTunnelLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpSecConnectionTunnelLifecycleStateEnum(val string) (IpSecConnectionTunnelLifecycleStateEnum, bool) {
	enum, ok := mappingIpSecConnectionTunnelLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpSecConnectionTunnelRoutingEnum Enum with underlying type: string
type IpSecConnectionTunnelRoutingEnum string

// Set of constants representing the allowable values for IpSecConnectionTunnelRoutingEnum
const (
	IpSecConnectionTunnelRoutingBgp    IpSecConnectionTunnelRoutingEnum = "BGP"
	IpSecConnectionTunnelRoutingStatic IpSecConnectionTunnelRoutingEnum = "STATIC"
	IpSecConnectionTunnelRoutingPolicy IpSecConnectionTunnelRoutingEnum = "POLICY"
)

var mappingIpSecConnectionTunnelRoutingEnum = map[string]IpSecConnectionTunnelRoutingEnum{
	"BGP":    IpSecConnectionTunnelRoutingBgp,
	"STATIC": IpSecConnectionTunnelRoutingStatic,
	"POLICY": IpSecConnectionTunnelRoutingPolicy,
}

var mappingIpSecConnectionTunnelRoutingEnumLowerCase = map[string]IpSecConnectionTunnelRoutingEnum{
	"bgp":    IpSecConnectionTunnelRoutingBgp,
	"static": IpSecConnectionTunnelRoutingStatic,
	"policy": IpSecConnectionTunnelRoutingPolicy,
}

// GetIpSecConnectionTunnelRoutingEnumValues Enumerates the set of values for IpSecConnectionTunnelRoutingEnum
func GetIpSecConnectionTunnelRoutingEnumValues() []IpSecConnectionTunnelRoutingEnum {
	values := make([]IpSecConnectionTunnelRoutingEnum, 0)
	for _, v := range mappingIpSecConnectionTunnelRoutingEnum {
		values = append(values, v)
	}
	return values
}

// GetIpSecConnectionTunnelRoutingEnumStringValues Enumerates the set of values in String for IpSecConnectionTunnelRoutingEnum
func GetIpSecConnectionTunnelRoutingEnumStringValues() []string {
	return []string{
		"BGP",
		"STATIC",
		"POLICY",
	}
}

// GetMappingIpSecConnectionTunnelRoutingEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpSecConnectionTunnelRoutingEnum(val string) (IpSecConnectionTunnelRoutingEnum, bool) {
	enum, ok := mappingIpSecConnectionTunnelRoutingEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpSecConnectionTunnelOracleCanInitiateEnum Enum with underlying type: string
type IpSecConnectionTunnelOracleCanInitiateEnum string

// Set of constants representing the allowable values for IpSecConnectionTunnelOracleCanInitiateEnum
const (
	IpSecConnectionTunnelOracleCanInitiateInitiatorOrResponder IpSecConnectionTunnelOracleCanInitiateEnum = "INITIATOR_OR_RESPONDER"
	IpSecConnectionTunnelOracleCanInitiateResponderOnly        IpSecConnectionTunnelOracleCanInitiateEnum = "RESPONDER_ONLY"
)

var mappingIpSecConnectionTunnelOracleCanInitiateEnum = map[string]IpSecConnectionTunnelOracleCanInitiateEnum{
	"INITIATOR_OR_RESPONDER": IpSecConnectionTunnelOracleCanInitiateInitiatorOrResponder,
	"RESPONDER_ONLY":         IpSecConnectionTunnelOracleCanInitiateResponderOnly,
}

var mappingIpSecConnectionTunnelOracleCanInitiateEnumLowerCase = map[string]IpSecConnectionTunnelOracleCanInitiateEnum{
	"initiator_or_responder": IpSecConnectionTunnelOracleCanInitiateInitiatorOrResponder,
	"responder_only":         IpSecConnectionTunnelOracleCanInitiateResponderOnly,
}

// GetIpSecConnectionTunnelOracleCanInitiateEnumValues Enumerates the set of values for IpSecConnectionTunnelOracleCanInitiateEnum
func GetIpSecConnectionTunnelOracleCanInitiateEnumValues() []IpSecConnectionTunnelOracleCanInitiateEnum {
	values := make([]IpSecConnectionTunnelOracleCanInitiateEnum, 0)
	for _, v := range mappingIpSecConnectionTunnelOracleCanInitiateEnum {
		values = append(values, v)
	}
	return values
}

// GetIpSecConnectionTunnelOracleCanInitiateEnumStringValues Enumerates the set of values in String for IpSecConnectionTunnelOracleCanInitiateEnum
func GetIpSecConnectionTunnelOracleCanInitiateEnumStringValues() []string {
	return []string{
		"INITIATOR_OR_RESPONDER",
		"RESPONDER_ONLY",
	}
}

// GetMappingIpSecConnectionTunnelOracleCanInitiateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpSecConnectionTunnelOracleCanInitiateEnum(val string) (IpSecConnectionTunnelOracleCanInitiateEnum, bool) {
	enum, ok := mappingIpSecConnectionTunnelOracleCanInitiateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpSecConnectionTunnelNatTranslationEnabledEnum Enum with underlying type: string
type IpSecConnectionTunnelNatTranslationEnabledEnum string

// Set of constants representing the allowable values for IpSecConnectionTunnelNatTranslationEnabledEnum
const (
	IpSecConnectionTunnelNatTranslationEnabledEnabled  IpSecConnectionTunnelNatTranslationEnabledEnum = "ENABLED"
	IpSecConnectionTunnelNatTranslationEnabledDisabled IpSecConnectionTunnelNatTranslationEnabledEnum = "DISABLED"
	IpSecConnectionTunnelNatTranslationEnabledAuto     IpSecConnectionTunnelNatTranslationEnabledEnum = "AUTO"
)

var mappingIpSecConnectionTunnelNatTranslationEnabledEnum = map[string]IpSecConnectionTunnelNatTranslationEnabledEnum{
	"ENABLED":  IpSecConnectionTunnelNatTranslationEnabledEnabled,
	"DISABLED": IpSecConnectionTunnelNatTranslationEnabledDisabled,
	"AUTO":     IpSecConnectionTunnelNatTranslationEnabledAuto,
}

var mappingIpSecConnectionTunnelNatTranslationEnabledEnumLowerCase = map[string]IpSecConnectionTunnelNatTranslationEnabledEnum{
	"enabled":  IpSecConnectionTunnelNatTranslationEnabledEnabled,
	"disabled": IpSecConnectionTunnelNatTranslationEnabledDisabled,
	"auto":     IpSecConnectionTunnelNatTranslationEnabledAuto,
}

// GetIpSecConnectionTunnelNatTranslationEnabledEnumValues Enumerates the set of values for IpSecConnectionTunnelNatTranslationEnabledEnum
func GetIpSecConnectionTunnelNatTranslationEnabledEnumValues() []IpSecConnectionTunnelNatTranslationEnabledEnum {
	values := make([]IpSecConnectionTunnelNatTranslationEnabledEnum, 0)
	for _, v := range mappingIpSecConnectionTunnelNatTranslationEnabledEnum {
		values = append(values, v)
	}
	return values
}

// GetIpSecConnectionTunnelNatTranslationEnabledEnumStringValues Enumerates the set of values in String for IpSecConnectionTunnelNatTranslationEnabledEnum
func GetIpSecConnectionTunnelNatTranslationEnabledEnumStringValues() []string {
	return []string{
		"ENABLED",
		"DISABLED",
		"AUTO",
	}
}

// GetMappingIpSecConnectionTunnelNatTranslationEnabledEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpSecConnectionTunnelNatTranslationEnabledEnum(val string) (IpSecConnectionTunnelNatTranslationEnabledEnum, bool) {
	enum, ok := mappingIpSecConnectionTunnelNatTranslationEnabledEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpSecConnectionTunnelDpdModeEnum Enum with underlying type: string
type IpSecConnectionTunnelDpdModeEnum string

// Set of constants representing the allowable values for IpSecConnectionTunnelDpdModeEnum
const (
	IpSecConnectionTunnelDpdModeInitiateAndRespond IpSecConnectionTunnelDpdModeEnum = "INITIATE_AND_RESPOND"
	IpSecConnectionTunnelDpdModeRespondOnly        IpSecConnectionTunnelDpdModeEnum = "RESPOND_ONLY"
)

var mappingIpSecConnectionTunnelDpdModeEnum = map[string]IpSecConnectionTunnelDpdModeEnum{
	"INITIATE_AND_RESPOND": IpSecConnectionTunnelDpdModeInitiateAndRespond,
	"RESPOND_ONLY":         IpSecConnectionTunnelDpdModeRespondOnly,
}

var mappingIpSecConnectionTunnelDpdModeEnumLowerCase = map[string]IpSecConnectionTunnelDpdModeEnum{
	"initiate_and_respond": IpSecConnectionTunnelDpdModeInitiateAndRespond,
	"respond_only":         IpSecConnectionTunnelDpdModeRespondOnly,
}

// GetIpSecConnectionTunnelDpdModeEnumValues Enumerates the set of values for IpSecConnectionTunnelDpdModeEnum
func GetIpSecConnectionTunnelDpdModeEnumValues() []IpSecConnectionTunnelDpdModeEnum {
	values := make([]IpSecConnectionTunnelDpdModeEnum, 0)
	for _, v := range mappingIpSecConnectionTunnelDpdModeEnum {
		values = append(values, v)
	}
	return values
}

// GetIpSecConnectionTunnelDpdModeEnumStringValues Enumerates the set of values in String for IpSecConnectionTunnelDpdModeEnum
func GetIpSecConnectionTunnelDpdModeEnumStringValues() []string {
	return []string{
		"INITIATE_AND_RESPOND",
		"RESPOND_ONLY",
	}
}

// GetMappingIpSecConnectionTunnelDpdModeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpSecConnectionTunnelDpdModeEnum(val string) (IpSecConnectionTunnelDpdModeEnum, bool) {
	enum, ok := mappingIpSecConnectionTunnelDpdModeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
