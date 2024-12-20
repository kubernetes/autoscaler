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

// IpSecConnection A connection between a DRG and CPE. This connection consists of multiple IPSec
// tunnels. Creating this connection is one of the steps required when setting up
// a Site-to-Site VPN.
// **Important:**  Each tunnel in an IPSec connection can use either static routing or BGP dynamic
// routing (see the IPSecConnectionTunnel object's
// `routing` attribute). Originally only static routing was supported and
// every IPSec connection was required to have at least one static route configured.
// To maintain backward compatibility in the API when support for BPG dynamic routing was introduced,
// the API accepts an empty list of static routes if you configure both of the IPSec tunnels to use
// BGP dynamic routing. If you switch a tunnel's routing from `BGP` to `STATIC`, you must first
// ensure that the IPSec connection is configured with at least one valid CIDR block static route.
// Oracle uses the IPSec connection's static routes when routing a tunnel's traffic *only*
// if that tunnel's `routing` attribute = `STATIC`. Otherwise the static routes are ignored.
// For more information about the workflow for setting up an IPSec connection, see
// Site-to-Site VPN Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/overviewIPsec.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type IpSecConnection struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the IPSec connection.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the Cpe object.
	CpeId *string `mandatory:"true" json:"cpeId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DRG.
	DrgId *string `mandatory:"true" json:"drgId"`

	// The IPSec connection's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The IPSec connection's current state.
	LifecycleState IpSecConnectionLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Static routes to the CPE. The CIDR must not be a
	// multicast address or class E address.
	// Used for routing a given IPSec tunnel's traffic only if the tunnel
	// is using static routing. If you configure at least one tunnel to use static routing, then
	// you must provide at least one valid static route. If you configure both
	// tunnels to use BGP dynamic routing, you can provide an empty list for the static routes.
	// The CIDR can be either IPv4 or IPv6. IPv6 addressing is supported for all commercial and government regions.
	// See IPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	// Example: `10.0.1.0/24`
	// Example: `2001:db8::/32`
	StaticRoutes []string `mandatory:"true" json:"staticRoutes"`

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

	// Your identifier for your CPE device. Can be either an IP address or a hostname (specifically,
	// the fully qualified domain name (FQDN)). The type of identifier here must correspond
	// to the value for `cpeLocalIdentifierType`.
	// If you don't provide a value when creating the IPSec connection, the `ipAddress` attribute
	// for the Cpe object specified by `cpeId` is used as the `cpeLocalIdentifier`.
	// For information about why you'd provide this value, see
	// If Your CPE Is Behind a NAT Device (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/overviewIPsec.htm#nat).
	// Example IP address: `10.0.3.3`
	// Example hostname: `cpe.example.com`
	CpeLocalIdentifier *string `mandatory:"false" json:"cpeLocalIdentifier"`

	// The type of identifier for your CPE device. The value here must correspond to the value
	// for `cpeLocalIdentifier`.
	CpeLocalIdentifierType IpSecConnectionCpeLocalIdentifierTypeEnum `mandatory:"false" json:"cpeLocalIdentifierType,omitempty"`

	// The date and time the IPSec connection was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The transport type used for the IPSec connection.
	TransportType IpSecConnectionTransportTypeEnum `mandatory:"false" json:"transportType,omitempty"`
}

func (m IpSecConnection) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m IpSecConnection) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingIpSecConnectionLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetIpSecConnectionLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := GetMappingIpSecConnectionCpeLocalIdentifierTypeEnum(string(m.CpeLocalIdentifierType)); !ok && m.CpeLocalIdentifierType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for CpeLocalIdentifierType: %s. Supported values are: %s.", m.CpeLocalIdentifierType, strings.Join(GetIpSecConnectionCpeLocalIdentifierTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingIpSecConnectionTransportTypeEnum(string(m.TransportType)); !ok && m.TransportType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TransportType: %s. Supported values are: %s.", m.TransportType, strings.Join(GetIpSecConnectionTransportTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// IpSecConnectionLifecycleStateEnum Enum with underlying type: string
type IpSecConnectionLifecycleStateEnum string

// Set of constants representing the allowable values for IpSecConnectionLifecycleStateEnum
const (
	IpSecConnectionLifecycleStateProvisioning IpSecConnectionLifecycleStateEnum = "PROVISIONING"
	IpSecConnectionLifecycleStateAvailable    IpSecConnectionLifecycleStateEnum = "AVAILABLE"
	IpSecConnectionLifecycleStateTerminating  IpSecConnectionLifecycleStateEnum = "TERMINATING"
	IpSecConnectionLifecycleStateTerminated   IpSecConnectionLifecycleStateEnum = "TERMINATED"
)

var mappingIpSecConnectionLifecycleStateEnum = map[string]IpSecConnectionLifecycleStateEnum{
	"PROVISIONING": IpSecConnectionLifecycleStateProvisioning,
	"AVAILABLE":    IpSecConnectionLifecycleStateAvailable,
	"TERMINATING":  IpSecConnectionLifecycleStateTerminating,
	"TERMINATED":   IpSecConnectionLifecycleStateTerminated,
}

var mappingIpSecConnectionLifecycleStateEnumLowerCase = map[string]IpSecConnectionLifecycleStateEnum{
	"provisioning": IpSecConnectionLifecycleStateProvisioning,
	"available":    IpSecConnectionLifecycleStateAvailable,
	"terminating":  IpSecConnectionLifecycleStateTerminating,
	"terminated":   IpSecConnectionLifecycleStateTerminated,
}

// GetIpSecConnectionLifecycleStateEnumValues Enumerates the set of values for IpSecConnectionLifecycleStateEnum
func GetIpSecConnectionLifecycleStateEnumValues() []IpSecConnectionLifecycleStateEnum {
	values := make([]IpSecConnectionLifecycleStateEnum, 0)
	for _, v := range mappingIpSecConnectionLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetIpSecConnectionLifecycleStateEnumStringValues Enumerates the set of values in String for IpSecConnectionLifecycleStateEnum
func GetIpSecConnectionLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingIpSecConnectionLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpSecConnectionLifecycleStateEnum(val string) (IpSecConnectionLifecycleStateEnum, bool) {
	enum, ok := mappingIpSecConnectionLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpSecConnectionCpeLocalIdentifierTypeEnum Enum with underlying type: string
type IpSecConnectionCpeLocalIdentifierTypeEnum string

// Set of constants representing the allowable values for IpSecConnectionCpeLocalIdentifierTypeEnum
const (
	IpSecConnectionCpeLocalIdentifierTypeIpAddress IpSecConnectionCpeLocalIdentifierTypeEnum = "IP_ADDRESS"
	IpSecConnectionCpeLocalIdentifierTypeHostname  IpSecConnectionCpeLocalIdentifierTypeEnum = "HOSTNAME"
)

var mappingIpSecConnectionCpeLocalIdentifierTypeEnum = map[string]IpSecConnectionCpeLocalIdentifierTypeEnum{
	"IP_ADDRESS": IpSecConnectionCpeLocalIdentifierTypeIpAddress,
	"HOSTNAME":   IpSecConnectionCpeLocalIdentifierTypeHostname,
}

var mappingIpSecConnectionCpeLocalIdentifierTypeEnumLowerCase = map[string]IpSecConnectionCpeLocalIdentifierTypeEnum{
	"ip_address": IpSecConnectionCpeLocalIdentifierTypeIpAddress,
	"hostname":   IpSecConnectionCpeLocalIdentifierTypeHostname,
}

// GetIpSecConnectionCpeLocalIdentifierTypeEnumValues Enumerates the set of values for IpSecConnectionCpeLocalIdentifierTypeEnum
func GetIpSecConnectionCpeLocalIdentifierTypeEnumValues() []IpSecConnectionCpeLocalIdentifierTypeEnum {
	values := make([]IpSecConnectionCpeLocalIdentifierTypeEnum, 0)
	for _, v := range mappingIpSecConnectionCpeLocalIdentifierTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetIpSecConnectionCpeLocalIdentifierTypeEnumStringValues Enumerates the set of values in String for IpSecConnectionCpeLocalIdentifierTypeEnum
func GetIpSecConnectionCpeLocalIdentifierTypeEnumStringValues() []string {
	return []string{
		"IP_ADDRESS",
		"HOSTNAME",
	}
}

// GetMappingIpSecConnectionCpeLocalIdentifierTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpSecConnectionCpeLocalIdentifierTypeEnum(val string) (IpSecConnectionCpeLocalIdentifierTypeEnum, bool) {
	enum, ok := mappingIpSecConnectionCpeLocalIdentifierTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// IpSecConnectionTransportTypeEnum Enum with underlying type: string
type IpSecConnectionTransportTypeEnum string

// Set of constants representing the allowable values for IpSecConnectionTransportTypeEnum
const (
	IpSecConnectionTransportTypeInternet    IpSecConnectionTransportTypeEnum = "INTERNET"
	IpSecConnectionTransportTypeFastconnect IpSecConnectionTransportTypeEnum = "FASTCONNECT"
)

var mappingIpSecConnectionTransportTypeEnum = map[string]IpSecConnectionTransportTypeEnum{
	"INTERNET":    IpSecConnectionTransportTypeInternet,
	"FASTCONNECT": IpSecConnectionTransportTypeFastconnect,
}

var mappingIpSecConnectionTransportTypeEnumLowerCase = map[string]IpSecConnectionTransportTypeEnum{
	"internet":    IpSecConnectionTransportTypeInternet,
	"fastconnect": IpSecConnectionTransportTypeFastconnect,
}

// GetIpSecConnectionTransportTypeEnumValues Enumerates the set of values for IpSecConnectionTransportTypeEnum
func GetIpSecConnectionTransportTypeEnumValues() []IpSecConnectionTransportTypeEnum {
	values := make([]IpSecConnectionTransportTypeEnum, 0)
	for _, v := range mappingIpSecConnectionTransportTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetIpSecConnectionTransportTypeEnumStringValues Enumerates the set of values in String for IpSecConnectionTransportTypeEnum
func GetIpSecConnectionTransportTypeEnumStringValues() []string {
	return []string{
		"INTERNET",
		"FASTCONNECT",
	}
}

// GetMappingIpSecConnectionTransportTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingIpSecConnectionTransportTypeEnum(val string) (IpSecConnectionTransportTypeEnum, bool) {
	enum, ok := mappingIpSecConnectionTransportTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
