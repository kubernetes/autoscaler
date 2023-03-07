// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
//

package core

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// CreateClientVpnDetails A request to create clientVpn.
type CreateClientVpnDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment that user sent request.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the attachedSubnet (VNIC) in customer tenancy.
	SubnetId *string `mandatory:"true" json:"subnetId"`

	// A list of accessible subnets from this clientVpnEnpoint.
	AccessibleSubnetCidrs []string `mandatory:"true" json:"accessibleSubnetCidrs"`

	DnsConfig *DnsConfigDetails `mandatory:"true" json:"dnsConfig"`

	// Whether re-route Internet traffic or not.
	IsRerouteEnabled *bool `mandatory:"true" json:"isRerouteEnabled"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// A limit that allows the maximum number of VPN concurrent connections.
	MaxConnections *int `mandatory:"false" json:"maxConnections"`

	// A subnet for openVPN clients to access servers. Default is 172.27.224.0/20
	ClientSubnetCidr *string `mandatory:"false" json:"clientSubnetCidr"`

	// Allowed values:
	//   * `NAT`: NAT mode supports the one-way access. In NAT mode, client can access the Internet from server endpoint
	//   but server endpoint cannot access the Internet from client.
	//   * `ROUTING`: ROUTING mode supports the two-way access. In ROUTING mode, client and server endpoint can access the
	//   Internet to each other.
	AddressingMode CreateClientVpnDetailsAddressingModeEnum `mandatory:"false" json:"addressingMode,omitempty"`

	// Allowed values:
	//   * `LOCAL`: Local authentication mode that applies users and password to get authentication through the server.
	//   * `RADIUS`: RADIUS authentication mode applies users and password to get authentication through the RADIUS server.
	//   * `LDAP`: LDAP authentication mode that applies users and passwords to get authentication through the LDAP server.
	AuthenticationMode CreateClientVpnDetailsAuthenticationModeEnum `mandatory:"false" json:"authenticationMode,omitempty"`

	RadiusConfig *RadiusConfigDetails `mandatory:"false" json:"radiusConfig"`

	LdapConfig *LdapConfigDetails `mandatory:"false" json:"ldapConfig"`

	SslCert SslCertDetails `mandatory:"false" json:"sslCert"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m CreateClientVpnDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateClientVpnDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingCreateClientVpnDetailsAddressingModeEnum[string(m.AddressingMode)]; !ok && m.AddressingMode != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AddressingMode: %s. Supported values are: %s.", m.AddressingMode, strings.Join(GetCreateClientVpnDetailsAddressingModeEnumStringValues(), ",")))
	}
	if _, ok := mappingCreateClientVpnDetailsAuthenticationModeEnum[string(m.AuthenticationMode)]; !ok && m.AuthenticationMode != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AuthenticationMode: %s. Supported values are: %s.", m.AuthenticationMode, strings.Join(GetCreateClientVpnDetailsAuthenticationModeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *CreateClientVpnDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DisplayName           *string                                      `json:"displayName"`
		MaxConnections        *int                                         `json:"maxConnections"`
		ClientSubnetCidr      *string                                      `json:"clientSubnetCidr"`
		AddressingMode        CreateClientVpnDetailsAddressingModeEnum     `json:"addressingMode"`
		AuthenticationMode    CreateClientVpnDetailsAuthenticationModeEnum `json:"authenticationMode"`
		RadiusConfig          *RadiusConfigDetails                         `json:"radiusConfig"`
		LdapConfig            *LdapConfigDetails                           `json:"ldapConfig"`
		SslCert               sslcertdetails                               `json:"sslCert"`
		DefinedTags           map[string]map[string]interface{}            `json:"definedTags"`
		FreeformTags          map[string]string                            `json:"freeformTags"`
		CompartmentId         *string                                      `json:"compartmentId"`
		SubnetId              *string                                      `json:"subnetId"`
		AccessibleSubnetCidrs []string                                     `json:"accessibleSubnetCidrs"`
		DnsConfig             *DnsConfigDetails                            `json:"dnsConfig"`
		IsRerouteEnabled      *bool                                        `json:"isRerouteEnabled"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DisplayName = model.DisplayName

	m.MaxConnections = model.MaxConnections

	m.ClientSubnetCidr = model.ClientSubnetCidr

	m.AddressingMode = model.AddressingMode

	m.AuthenticationMode = model.AuthenticationMode

	m.RadiusConfig = model.RadiusConfig

	m.LdapConfig = model.LdapConfig

	nn, e = model.SslCert.UnmarshalPolymorphicJSON(model.SslCert.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.SslCert = nn.(SslCertDetails)
	} else {
		m.SslCert = nil
	}

	m.DefinedTags = model.DefinedTags

	m.FreeformTags = model.FreeformTags

	m.CompartmentId = model.CompartmentId

	m.SubnetId = model.SubnetId

	m.AccessibleSubnetCidrs = make([]string, len(model.AccessibleSubnetCidrs))
	for i, n := range model.AccessibleSubnetCidrs {
		m.AccessibleSubnetCidrs[i] = n
	}

	m.DnsConfig = model.DnsConfig

	m.IsRerouteEnabled = model.IsRerouteEnabled

	return
}

// CreateClientVpnDetailsAddressingModeEnum Enum with underlying type: string
type CreateClientVpnDetailsAddressingModeEnum string

// Set of constants representing the allowable values for CreateClientVpnDetailsAddressingModeEnum
const (
	CreateClientVpnDetailsAddressingModeNat     CreateClientVpnDetailsAddressingModeEnum = "NAT"
	CreateClientVpnDetailsAddressingModeRouting CreateClientVpnDetailsAddressingModeEnum = "ROUTING"
)

var mappingCreateClientVpnDetailsAddressingModeEnum = map[string]CreateClientVpnDetailsAddressingModeEnum{
	"NAT":     CreateClientVpnDetailsAddressingModeNat,
	"ROUTING": CreateClientVpnDetailsAddressingModeRouting,
}

// GetCreateClientVpnDetailsAddressingModeEnumValues Enumerates the set of values for CreateClientVpnDetailsAddressingModeEnum
func GetCreateClientVpnDetailsAddressingModeEnumValues() []CreateClientVpnDetailsAddressingModeEnum {
	values := make([]CreateClientVpnDetailsAddressingModeEnum, 0)
	for _, v := range mappingCreateClientVpnDetailsAddressingModeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateClientVpnDetailsAddressingModeEnumStringValues Enumerates the set of values in String for CreateClientVpnDetailsAddressingModeEnum
func GetCreateClientVpnDetailsAddressingModeEnumStringValues() []string {
	return []string{
		"NAT",
		"ROUTING",
	}
}

// CreateClientVpnDetailsAuthenticationModeEnum Enum with underlying type: string
type CreateClientVpnDetailsAuthenticationModeEnum string

// Set of constants representing the allowable values for CreateClientVpnDetailsAuthenticationModeEnum
const (
	CreateClientVpnDetailsAuthenticationModeLocal  CreateClientVpnDetailsAuthenticationModeEnum = "LOCAL"
	CreateClientVpnDetailsAuthenticationModeRadius CreateClientVpnDetailsAuthenticationModeEnum = "RADIUS"
	CreateClientVpnDetailsAuthenticationModeLdap   CreateClientVpnDetailsAuthenticationModeEnum = "LDAP"
)

var mappingCreateClientVpnDetailsAuthenticationModeEnum = map[string]CreateClientVpnDetailsAuthenticationModeEnum{
	"LOCAL":  CreateClientVpnDetailsAuthenticationModeLocal,
	"RADIUS": CreateClientVpnDetailsAuthenticationModeRadius,
	"LDAP":   CreateClientVpnDetailsAuthenticationModeLdap,
}

// GetCreateClientVpnDetailsAuthenticationModeEnumValues Enumerates the set of values for CreateClientVpnDetailsAuthenticationModeEnum
func GetCreateClientVpnDetailsAuthenticationModeEnumValues() []CreateClientVpnDetailsAuthenticationModeEnum {
	values := make([]CreateClientVpnDetailsAuthenticationModeEnum, 0)
	for _, v := range mappingCreateClientVpnDetailsAuthenticationModeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateClientVpnDetailsAuthenticationModeEnumStringValues Enumerates the set of values in String for CreateClientVpnDetailsAuthenticationModeEnum
func GetCreateClientVpnDetailsAuthenticationModeEnumStringValues() []string {
	return []string{
		"LOCAL",
		"RADIUS",
		"LDAP",
	}
}
