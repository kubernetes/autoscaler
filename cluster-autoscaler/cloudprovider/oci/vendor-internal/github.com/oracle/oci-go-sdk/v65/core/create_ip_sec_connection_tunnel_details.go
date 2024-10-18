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

// CreateIpSecConnectionTunnelDetails The representation of CreateIpSecConnectionTunnelDetails
type CreateIpSecConnectionTunnelDetails struct {

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The type of routing to use for this tunnel (BGP dynamic routing, static routing, or policy-based routing).
	Routing CreateIpSecConnectionTunnelDetailsRoutingEnum `mandatory:"false" json:"routing,omitempty"`

	// Internet Key Exchange protocol version.
	IkeVersion CreateIpSecConnectionTunnelDetailsIkeVersionEnum `mandatory:"false" json:"ikeVersion,omitempty"`

	// The shared secret (pre-shared key) to use for the IPSec tunnel. Only numbers, letters, and
	// spaces are allowed. If you don't provide a value,
	// Oracle generates a value for you. You can specify your own shared secret later if
	// you like with UpdateIPSecConnectionTunnelSharedSecret.
	SharedSecret *string `mandatory:"false" json:"sharedSecret"`

	BgpSessionConfig *CreateIpSecTunnelBgpSessionDetails `mandatory:"false" json:"bgpSessionConfig"`

	// Indicates whether the Oracle end of the IPSec connection is able to initiate starting up the IPSec tunnel.
	OracleInitiation CreateIpSecConnectionTunnelDetailsOracleInitiationEnum `mandatory:"false" json:"oracleInitiation,omitempty"`

	// By default (the `AUTO` setting), IKE sends packets with a source and destination port set to 500,
	// and when it detects that the port used to forward packets has changed (most likely because a NAT device
	// is between the CPE device and the Oracle VPN headend) it will try to negotiate the use of NAT-T.
	// The `ENABLED` option sets the IKE protocol to use port 4500 instead of 500 and forces encapsulating traffic with the ESP protocol inside UDP packets.
	// The `DISABLED` option directs IKE to completely refuse to negotiate NAT-T
	// even if it senses there may be a NAT device in use.
	NatTranslationEnabled CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum `mandatory:"false" json:"natTranslationEnabled,omitempty"`

	PhaseOneConfig *PhaseOneConfigDetails `mandatory:"false" json:"phaseOneConfig"`

	PhaseTwoConfig *PhaseTwoConfigDetails `mandatory:"false" json:"phaseTwoConfig"`

	DpdConfig *DpdConfig `mandatory:"false" json:"dpdConfig"`

	// The headend IP that you can choose on the Oracle side to terminate your private IPSec tunnel.
	OracleTunnelIp *string `mandatory:"false" json:"oracleTunnelIp"`

	// The list of virtual circuit OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)s over which your network can reach this tunnel.
	AssociatedVirtualCircuits []string `mandatory:"false" json:"associatedVirtualCircuits"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DRG route table assigned to this attachment.
	// The DRG route table manages traffic inside the DRG.
	DrgRouteTableId *string `mandatory:"false" json:"drgRouteTableId"`

	EncryptionDomainConfig *CreateIpSecTunnelEncryptionDomainDetails `mandatory:"false" json:"encryptionDomainConfig"`
}

func (m CreateIpSecConnectionTunnelDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateIpSecConnectionTunnelDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCreateIpSecConnectionTunnelDetailsRoutingEnum(string(m.Routing)); !ok && m.Routing != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Routing: %s. Supported values are: %s.", m.Routing, strings.Join(GetCreateIpSecConnectionTunnelDetailsRoutingEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCreateIpSecConnectionTunnelDetailsIkeVersionEnum(string(m.IkeVersion)); !ok && m.IkeVersion != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for IkeVersion: %s. Supported values are: %s.", m.IkeVersion, strings.Join(GetCreateIpSecConnectionTunnelDetailsIkeVersionEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCreateIpSecConnectionTunnelDetailsOracleInitiationEnum(string(m.OracleInitiation)); !ok && m.OracleInitiation != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for OracleInitiation: %s. Supported values are: %s.", m.OracleInitiation, strings.Join(GetCreateIpSecConnectionTunnelDetailsOracleInitiationEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum(string(m.NatTranslationEnabled)); !ok && m.NatTranslationEnabled != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for NatTranslationEnabled: %s. Supported values are: %s.", m.NatTranslationEnabled, strings.Join(GetCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateIpSecConnectionTunnelDetailsRoutingEnum Enum with underlying type: string
type CreateIpSecConnectionTunnelDetailsRoutingEnum string

// Set of constants representing the allowable values for CreateIpSecConnectionTunnelDetailsRoutingEnum
const (
	CreateIpSecConnectionTunnelDetailsRoutingBgp    CreateIpSecConnectionTunnelDetailsRoutingEnum = "BGP"
	CreateIpSecConnectionTunnelDetailsRoutingStatic CreateIpSecConnectionTunnelDetailsRoutingEnum = "STATIC"
	CreateIpSecConnectionTunnelDetailsRoutingPolicy CreateIpSecConnectionTunnelDetailsRoutingEnum = "POLICY"
)

var mappingCreateIpSecConnectionTunnelDetailsRoutingEnum = map[string]CreateIpSecConnectionTunnelDetailsRoutingEnum{
	"BGP":    CreateIpSecConnectionTunnelDetailsRoutingBgp,
	"STATIC": CreateIpSecConnectionTunnelDetailsRoutingStatic,
	"POLICY": CreateIpSecConnectionTunnelDetailsRoutingPolicy,
}

var mappingCreateIpSecConnectionTunnelDetailsRoutingEnumLowerCase = map[string]CreateIpSecConnectionTunnelDetailsRoutingEnum{
	"bgp":    CreateIpSecConnectionTunnelDetailsRoutingBgp,
	"static": CreateIpSecConnectionTunnelDetailsRoutingStatic,
	"policy": CreateIpSecConnectionTunnelDetailsRoutingPolicy,
}

// GetCreateIpSecConnectionTunnelDetailsRoutingEnumValues Enumerates the set of values for CreateIpSecConnectionTunnelDetailsRoutingEnum
func GetCreateIpSecConnectionTunnelDetailsRoutingEnumValues() []CreateIpSecConnectionTunnelDetailsRoutingEnum {
	values := make([]CreateIpSecConnectionTunnelDetailsRoutingEnum, 0)
	for _, v := range mappingCreateIpSecConnectionTunnelDetailsRoutingEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateIpSecConnectionTunnelDetailsRoutingEnumStringValues Enumerates the set of values in String for CreateIpSecConnectionTunnelDetailsRoutingEnum
func GetCreateIpSecConnectionTunnelDetailsRoutingEnumStringValues() []string {
	return []string{
		"BGP",
		"STATIC",
		"POLICY",
	}
}

// GetMappingCreateIpSecConnectionTunnelDetailsRoutingEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateIpSecConnectionTunnelDetailsRoutingEnum(val string) (CreateIpSecConnectionTunnelDetailsRoutingEnum, bool) {
	enum, ok := mappingCreateIpSecConnectionTunnelDetailsRoutingEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CreateIpSecConnectionTunnelDetailsIkeVersionEnum Enum with underlying type: string
type CreateIpSecConnectionTunnelDetailsIkeVersionEnum string

// Set of constants representing the allowable values for CreateIpSecConnectionTunnelDetailsIkeVersionEnum
const (
	CreateIpSecConnectionTunnelDetailsIkeVersionV1 CreateIpSecConnectionTunnelDetailsIkeVersionEnum = "V1"
	CreateIpSecConnectionTunnelDetailsIkeVersionV2 CreateIpSecConnectionTunnelDetailsIkeVersionEnum = "V2"
)

var mappingCreateIpSecConnectionTunnelDetailsIkeVersionEnum = map[string]CreateIpSecConnectionTunnelDetailsIkeVersionEnum{
	"V1": CreateIpSecConnectionTunnelDetailsIkeVersionV1,
	"V2": CreateIpSecConnectionTunnelDetailsIkeVersionV2,
}

var mappingCreateIpSecConnectionTunnelDetailsIkeVersionEnumLowerCase = map[string]CreateIpSecConnectionTunnelDetailsIkeVersionEnum{
	"v1": CreateIpSecConnectionTunnelDetailsIkeVersionV1,
	"v2": CreateIpSecConnectionTunnelDetailsIkeVersionV2,
}

// GetCreateIpSecConnectionTunnelDetailsIkeVersionEnumValues Enumerates the set of values for CreateIpSecConnectionTunnelDetailsIkeVersionEnum
func GetCreateIpSecConnectionTunnelDetailsIkeVersionEnumValues() []CreateIpSecConnectionTunnelDetailsIkeVersionEnum {
	values := make([]CreateIpSecConnectionTunnelDetailsIkeVersionEnum, 0)
	for _, v := range mappingCreateIpSecConnectionTunnelDetailsIkeVersionEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateIpSecConnectionTunnelDetailsIkeVersionEnumStringValues Enumerates the set of values in String for CreateIpSecConnectionTunnelDetailsIkeVersionEnum
func GetCreateIpSecConnectionTunnelDetailsIkeVersionEnumStringValues() []string {
	return []string{
		"V1",
		"V2",
	}
}

// GetMappingCreateIpSecConnectionTunnelDetailsIkeVersionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateIpSecConnectionTunnelDetailsIkeVersionEnum(val string) (CreateIpSecConnectionTunnelDetailsIkeVersionEnum, bool) {
	enum, ok := mappingCreateIpSecConnectionTunnelDetailsIkeVersionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CreateIpSecConnectionTunnelDetailsOracleInitiationEnum Enum with underlying type: string
type CreateIpSecConnectionTunnelDetailsOracleInitiationEnum string

// Set of constants representing the allowable values for CreateIpSecConnectionTunnelDetailsOracleInitiationEnum
const (
	CreateIpSecConnectionTunnelDetailsOracleInitiationInitiatorOrResponder CreateIpSecConnectionTunnelDetailsOracleInitiationEnum = "INITIATOR_OR_RESPONDER"
	CreateIpSecConnectionTunnelDetailsOracleInitiationResponderOnly        CreateIpSecConnectionTunnelDetailsOracleInitiationEnum = "RESPONDER_ONLY"
)

var mappingCreateIpSecConnectionTunnelDetailsOracleInitiationEnum = map[string]CreateIpSecConnectionTunnelDetailsOracleInitiationEnum{
	"INITIATOR_OR_RESPONDER": CreateIpSecConnectionTunnelDetailsOracleInitiationInitiatorOrResponder,
	"RESPONDER_ONLY":         CreateIpSecConnectionTunnelDetailsOracleInitiationResponderOnly,
}

var mappingCreateIpSecConnectionTunnelDetailsOracleInitiationEnumLowerCase = map[string]CreateIpSecConnectionTunnelDetailsOracleInitiationEnum{
	"initiator_or_responder": CreateIpSecConnectionTunnelDetailsOracleInitiationInitiatorOrResponder,
	"responder_only":         CreateIpSecConnectionTunnelDetailsOracleInitiationResponderOnly,
}

// GetCreateIpSecConnectionTunnelDetailsOracleInitiationEnumValues Enumerates the set of values for CreateIpSecConnectionTunnelDetailsOracleInitiationEnum
func GetCreateIpSecConnectionTunnelDetailsOracleInitiationEnumValues() []CreateIpSecConnectionTunnelDetailsOracleInitiationEnum {
	values := make([]CreateIpSecConnectionTunnelDetailsOracleInitiationEnum, 0)
	for _, v := range mappingCreateIpSecConnectionTunnelDetailsOracleInitiationEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateIpSecConnectionTunnelDetailsOracleInitiationEnumStringValues Enumerates the set of values in String for CreateIpSecConnectionTunnelDetailsOracleInitiationEnum
func GetCreateIpSecConnectionTunnelDetailsOracleInitiationEnumStringValues() []string {
	return []string{
		"INITIATOR_OR_RESPONDER",
		"RESPONDER_ONLY",
	}
}

// GetMappingCreateIpSecConnectionTunnelDetailsOracleInitiationEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateIpSecConnectionTunnelDetailsOracleInitiationEnum(val string) (CreateIpSecConnectionTunnelDetailsOracleInitiationEnum, bool) {
	enum, ok := mappingCreateIpSecConnectionTunnelDetailsOracleInitiationEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum Enum with underlying type: string
type CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum string

// Set of constants representing the allowable values for CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum
const (
	CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnabled  CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum = "ENABLED"
	CreateIpSecConnectionTunnelDetailsNatTranslationEnabledDisabled CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum = "DISABLED"
	CreateIpSecConnectionTunnelDetailsNatTranslationEnabledAuto     CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum = "AUTO"
)

var mappingCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum = map[string]CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum{
	"ENABLED":  CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnabled,
	"DISABLED": CreateIpSecConnectionTunnelDetailsNatTranslationEnabledDisabled,
	"AUTO":     CreateIpSecConnectionTunnelDetailsNatTranslationEnabledAuto,
}

var mappingCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumLowerCase = map[string]CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum{
	"enabled":  CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnabled,
	"disabled": CreateIpSecConnectionTunnelDetailsNatTranslationEnabledDisabled,
	"auto":     CreateIpSecConnectionTunnelDetailsNatTranslationEnabledAuto,
}

// GetCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumValues Enumerates the set of values for CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum
func GetCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumValues() []CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum {
	values := make([]CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum, 0)
	for _, v := range mappingCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumStringValues Enumerates the set of values in String for CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum
func GetCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumStringValues() []string {
	return []string{
		"ENABLED",
		"DISABLED",
		"AUTO",
	}
}

// GetMappingCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum(val string) (CreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnum, bool) {
	enum, ok := mappingCreateIpSecConnectionTunnelDetailsNatTranslationEnabledEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
