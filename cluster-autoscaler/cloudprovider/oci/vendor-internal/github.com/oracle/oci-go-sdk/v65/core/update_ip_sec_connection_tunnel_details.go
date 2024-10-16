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

// UpdateIpSecConnectionTunnelDetails The representation of UpdateIpSecConnectionTunnelDetails
type UpdateIpSecConnectionTunnelDetails struct {

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The type of routing to use for this tunnel (BGP dynamic routing, static routing, or policy-based routing).
	Routing UpdateIpSecConnectionTunnelDetailsRoutingEnum `mandatory:"false" json:"routing,omitempty"`

	// Internet Key Exchange protocol version.
	IkeVersion UpdateIpSecConnectionTunnelDetailsIkeVersionEnum `mandatory:"false" json:"ikeVersion,omitempty"`

	BgpSessionConfig *UpdateIpSecTunnelBgpSessionDetails `mandatory:"false" json:"bgpSessionConfig"`

	// Indicates whether the Oracle end of the IPSec connection is able to initiate starting up the IPSec tunnel.
	OracleInitiation UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum `mandatory:"false" json:"oracleInitiation,omitempty"`

	// By default (the `AUTO` setting), IKE sends packets with a source and destination port set to 500,
	// and when it detects that the port used to forward packets has changed (most likely because a NAT device
	// is between the CPE device and the Oracle VPN headend) it will try to negotiate the use of NAT-T.
	// The `ENABLED` option sets the IKE protocol to use port 4500 instead of 500 and forces encapsulating traffic with the ESP protocol inside UDP packets.
	// The `DISABLED` option directs IKE to completely refuse to negotiate NAT-T
	// even if it senses there may be a NAT device in use.
	NatTranslationEnabled UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum `mandatory:"false" json:"natTranslationEnabled,omitempty"`

	PhaseOneConfig *PhaseOneConfigDetails `mandatory:"false" json:"phaseOneConfig"`

	PhaseTwoConfig *PhaseTwoConfigDetails `mandatory:"false" json:"phaseTwoConfig"`

	DpdConfig *DpdConfig `mandatory:"false" json:"dpdConfig"`

	EncryptionDomainConfig *UpdateIpSecTunnelEncryptionDomainDetails `mandatory:"false" json:"encryptionDomainConfig"`
}

func (m UpdateIpSecConnectionTunnelDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateIpSecConnectionTunnelDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingUpdateIpSecConnectionTunnelDetailsRoutingEnum(string(m.Routing)); !ok && m.Routing != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Routing: %s. Supported values are: %s.", m.Routing, strings.Join(GetUpdateIpSecConnectionTunnelDetailsRoutingEnumStringValues(), ",")))
	}
	if _, ok := GetMappingUpdateIpSecConnectionTunnelDetailsIkeVersionEnum(string(m.IkeVersion)); !ok && m.IkeVersion != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for IkeVersion: %s. Supported values are: %s.", m.IkeVersion, strings.Join(GetUpdateIpSecConnectionTunnelDetailsIkeVersionEnumStringValues(), ",")))
	}
	if _, ok := GetMappingUpdateIpSecConnectionTunnelDetailsOracleInitiationEnum(string(m.OracleInitiation)); !ok && m.OracleInitiation != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for OracleInitiation: %s. Supported values are: %s.", m.OracleInitiation, strings.Join(GetUpdateIpSecConnectionTunnelDetailsOracleInitiationEnumStringValues(), ",")))
	}
	if _, ok := GetMappingUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum(string(m.NatTranslationEnabled)); !ok && m.NatTranslationEnabled != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for NatTranslationEnabled: %s. Supported values are: %s.", m.NatTranslationEnabled, strings.Join(GetUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UpdateIpSecConnectionTunnelDetailsRoutingEnum Enum with underlying type: string
type UpdateIpSecConnectionTunnelDetailsRoutingEnum string

// Set of constants representing the allowable values for UpdateIpSecConnectionTunnelDetailsRoutingEnum
const (
	UpdateIpSecConnectionTunnelDetailsRoutingBgp    UpdateIpSecConnectionTunnelDetailsRoutingEnum = "BGP"
	UpdateIpSecConnectionTunnelDetailsRoutingStatic UpdateIpSecConnectionTunnelDetailsRoutingEnum = "STATIC"
	UpdateIpSecConnectionTunnelDetailsRoutingPolicy UpdateIpSecConnectionTunnelDetailsRoutingEnum = "POLICY"
)

var mappingUpdateIpSecConnectionTunnelDetailsRoutingEnum = map[string]UpdateIpSecConnectionTunnelDetailsRoutingEnum{
	"BGP":    UpdateIpSecConnectionTunnelDetailsRoutingBgp,
	"STATIC": UpdateIpSecConnectionTunnelDetailsRoutingStatic,
	"POLICY": UpdateIpSecConnectionTunnelDetailsRoutingPolicy,
}

var mappingUpdateIpSecConnectionTunnelDetailsRoutingEnumLowerCase = map[string]UpdateIpSecConnectionTunnelDetailsRoutingEnum{
	"bgp":    UpdateIpSecConnectionTunnelDetailsRoutingBgp,
	"static": UpdateIpSecConnectionTunnelDetailsRoutingStatic,
	"policy": UpdateIpSecConnectionTunnelDetailsRoutingPolicy,
}

// GetUpdateIpSecConnectionTunnelDetailsRoutingEnumValues Enumerates the set of values for UpdateIpSecConnectionTunnelDetailsRoutingEnum
func GetUpdateIpSecConnectionTunnelDetailsRoutingEnumValues() []UpdateIpSecConnectionTunnelDetailsRoutingEnum {
	values := make([]UpdateIpSecConnectionTunnelDetailsRoutingEnum, 0)
	for _, v := range mappingUpdateIpSecConnectionTunnelDetailsRoutingEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateIpSecConnectionTunnelDetailsRoutingEnumStringValues Enumerates the set of values in String for UpdateIpSecConnectionTunnelDetailsRoutingEnum
func GetUpdateIpSecConnectionTunnelDetailsRoutingEnumStringValues() []string {
	return []string{
		"BGP",
		"STATIC",
		"POLICY",
	}
}

// GetMappingUpdateIpSecConnectionTunnelDetailsRoutingEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateIpSecConnectionTunnelDetailsRoutingEnum(val string) (UpdateIpSecConnectionTunnelDetailsRoutingEnum, bool) {
	enum, ok := mappingUpdateIpSecConnectionTunnelDetailsRoutingEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// UpdateIpSecConnectionTunnelDetailsIkeVersionEnum Enum with underlying type: string
type UpdateIpSecConnectionTunnelDetailsIkeVersionEnum string

// Set of constants representing the allowable values for UpdateIpSecConnectionTunnelDetailsIkeVersionEnum
const (
	UpdateIpSecConnectionTunnelDetailsIkeVersionV1 UpdateIpSecConnectionTunnelDetailsIkeVersionEnum = "V1"
	UpdateIpSecConnectionTunnelDetailsIkeVersionV2 UpdateIpSecConnectionTunnelDetailsIkeVersionEnum = "V2"
)

var mappingUpdateIpSecConnectionTunnelDetailsIkeVersionEnum = map[string]UpdateIpSecConnectionTunnelDetailsIkeVersionEnum{
	"V1": UpdateIpSecConnectionTunnelDetailsIkeVersionV1,
	"V2": UpdateIpSecConnectionTunnelDetailsIkeVersionV2,
}

var mappingUpdateIpSecConnectionTunnelDetailsIkeVersionEnumLowerCase = map[string]UpdateIpSecConnectionTunnelDetailsIkeVersionEnum{
	"v1": UpdateIpSecConnectionTunnelDetailsIkeVersionV1,
	"v2": UpdateIpSecConnectionTunnelDetailsIkeVersionV2,
}

// GetUpdateIpSecConnectionTunnelDetailsIkeVersionEnumValues Enumerates the set of values for UpdateIpSecConnectionTunnelDetailsIkeVersionEnum
func GetUpdateIpSecConnectionTunnelDetailsIkeVersionEnumValues() []UpdateIpSecConnectionTunnelDetailsIkeVersionEnum {
	values := make([]UpdateIpSecConnectionTunnelDetailsIkeVersionEnum, 0)
	for _, v := range mappingUpdateIpSecConnectionTunnelDetailsIkeVersionEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateIpSecConnectionTunnelDetailsIkeVersionEnumStringValues Enumerates the set of values in String for UpdateIpSecConnectionTunnelDetailsIkeVersionEnum
func GetUpdateIpSecConnectionTunnelDetailsIkeVersionEnumStringValues() []string {
	return []string{
		"V1",
		"V2",
	}
}

// GetMappingUpdateIpSecConnectionTunnelDetailsIkeVersionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateIpSecConnectionTunnelDetailsIkeVersionEnum(val string) (UpdateIpSecConnectionTunnelDetailsIkeVersionEnum, bool) {
	enum, ok := mappingUpdateIpSecConnectionTunnelDetailsIkeVersionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum Enum with underlying type: string
type UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum string

// Set of constants representing the allowable values for UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum
const (
	UpdateIpSecConnectionTunnelDetailsOracleInitiationInitiatorOrResponder UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum = "INITIATOR_OR_RESPONDER"
	UpdateIpSecConnectionTunnelDetailsOracleInitiationResponderOnly        UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum = "RESPONDER_ONLY"
)

var mappingUpdateIpSecConnectionTunnelDetailsOracleInitiationEnum = map[string]UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum{
	"INITIATOR_OR_RESPONDER": UpdateIpSecConnectionTunnelDetailsOracleInitiationInitiatorOrResponder,
	"RESPONDER_ONLY":         UpdateIpSecConnectionTunnelDetailsOracleInitiationResponderOnly,
}

var mappingUpdateIpSecConnectionTunnelDetailsOracleInitiationEnumLowerCase = map[string]UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum{
	"initiator_or_responder": UpdateIpSecConnectionTunnelDetailsOracleInitiationInitiatorOrResponder,
	"responder_only":         UpdateIpSecConnectionTunnelDetailsOracleInitiationResponderOnly,
}

// GetUpdateIpSecConnectionTunnelDetailsOracleInitiationEnumValues Enumerates the set of values for UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum
func GetUpdateIpSecConnectionTunnelDetailsOracleInitiationEnumValues() []UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum {
	values := make([]UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum, 0)
	for _, v := range mappingUpdateIpSecConnectionTunnelDetailsOracleInitiationEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateIpSecConnectionTunnelDetailsOracleInitiationEnumStringValues Enumerates the set of values in String for UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum
func GetUpdateIpSecConnectionTunnelDetailsOracleInitiationEnumStringValues() []string {
	return []string{
		"INITIATOR_OR_RESPONDER",
		"RESPONDER_ONLY",
	}
}

// GetMappingUpdateIpSecConnectionTunnelDetailsOracleInitiationEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateIpSecConnectionTunnelDetailsOracleInitiationEnum(val string) (UpdateIpSecConnectionTunnelDetailsOracleInitiationEnum, bool) {
	enum, ok := mappingUpdateIpSecConnectionTunnelDetailsOracleInitiationEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum Enum with underlying type: string
type UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum string

// Set of constants representing the allowable values for UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum
const (
	UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnabled  UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum = "ENABLED"
	UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledDisabled UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum = "DISABLED"
	UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledAuto     UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum = "AUTO"
)

var mappingUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum = map[string]UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum{
	"ENABLED":  UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnabled,
	"DISABLED": UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledDisabled,
	"AUTO":     UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledAuto,
}

var mappingUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumLowerCase = map[string]UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum{
	"enabled":  UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnabled,
	"disabled": UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledDisabled,
	"auto":     UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledAuto,
}

// GetUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumValues Enumerates the set of values for UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum
func GetUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumValues() []UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum {
	values := make([]UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum, 0)
	for _, v := range mappingUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumStringValues Enumerates the set of values in String for UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum
func GetUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumStringValues() []string {
	return []string{
		"ENABLED",
		"DISABLED",
		"AUTO",
	}
}

// GetMappingUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum(val string) (UpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum, bool) {
	enum, ok := mappingUpdateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
