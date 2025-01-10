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
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// DhcpOptions A set of DHCP options. Used by the VCN to automatically provide configuration
// information to the instances when they boot up. There are two options you can set:
// - DhcpDnsOption: Lets you specify how DNS (hostname resolution) is
// handled in the subnets in your VCN.
// - DhcpSearchDomainOption: Lets you specify
// a search domain name to use for DNS queries.
// For more information, see  DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm)
// and DHCP Options (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingDHCP.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type DhcpOptions struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the set of DHCP options.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)) for the set of DHCP options.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the set of DHCP options.
	LifecycleState DhcpOptionsLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The collection of individual DHCP options.
	Options []DhcpOption `mandatory:"true" json:"options"`

	// Date and time the set of DHCP options was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN the set of DHCP options belongs to.
	VcnId *string `mandatory:"true" json:"vcnId"`

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

	// The search domain name type of DHCP options
	DomainNameType DhcpOptionsDomainNameTypeEnum `mandatory:"false" json:"domainNameType,omitempty"`
}

func (m DhcpOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DhcpOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingDhcpOptionsLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetDhcpOptionsLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := GetMappingDhcpOptionsDomainNameTypeEnum(string(m.DomainNameType)); !ok && m.DomainNameType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DomainNameType: %s. Supported values are: %s.", m.DomainNameType, strings.Join(GetDhcpOptionsDomainNameTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *DhcpOptions) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DefinedTags    map[string]map[string]interface{} `json:"definedTags"`
		DisplayName    *string                           `json:"displayName"`
		FreeformTags   map[string]string                 `json:"freeformTags"`
		DomainNameType DhcpOptionsDomainNameTypeEnum     `json:"domainNameType"`
		CompartmentId  *string                           `json:"compartmentId"`
		Id             *string                           `json:"id"`
		LifecycleState DhcpOptionsLifecycleStateEnum     `json:"lifecycleState"`
		Options        []dhcpoption                      `json:"options"`
		TimeCreated    *common.SDKTime                   `json:"timeCreated"`
		VcnId          *string                           `json:"vcnId"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DefinedTags = model.DefinedTags

	m.DisplayName = model.DisplayName

	m.FreeformTags = model.FreeformTags

	m.DomainNameType = model.DomainNameType

	m.CompartmentId = model.CompartmentId

	m.Id = model.Id

	m.LifecycleState = model.LifecycleState

	m.Options = make([]DhcpOption, len(model.Options))
	for i, n := range model.Options {
		nn, e = n.UnmarshalPolymorphicJSON(n.JsonData)
		if e != nil {
			return e
		}
		if nn != nil {
			m.Options[i] = nn.(DhcpOption)
		} else {
			m.Options[i] = nil
		}
	}
	m.TimeCreated = model.TimeCreated

	m.VcnId = model.VcnId

	return
}

// DhcpOptionsLifecycleStateEnum Enum with underlying type: string
type DhcpOptionsLifecycleStateEnum string

// Set of constants representing the allowable values for DhcpOptionsLifecycleStateEnum
const (
	DhcpOptionsLifecycleStateProvisioning DhcpOptionsLifecycleStateEnum = "PROVISIONING"
	DhcpOptionsLifecycleStateAvailable    DhcpOptionsLifecycleStateEnum = "AVAILABLE"
	DhcpOptionsLifecycleStateTerminating  DhcpOptionsLifecycleStateEnum = "TERMINATING"
	DhcpOptionsLifecycleStateTerminated   DhcpOptionsLifecycleStateEnum = "TERMINATED"
)

var mappingDhcpOptionsLifecycleStateEnum = map[string]DhcpOptionsLifecycleStateEnum{
	"PROVISIONING": DhcpOptionsLifecycleStateProvisioning,
	"AVAILABLE":    DhcpOptionsLifecycleStateAvailable,
	"TERMINATING":  DhcpOptionsLifecycleStateTerminating,
	"TERMINATED":   DhcpOptionsLifecycleStateTerminated,
}

var mappingDhcpOptionsLifecycleStateEnumLowerCase = map[string]DhcpOptionsLifecycleStateEnum{
	"provisioning": DhcpOptionsLifecycleStateProvisioning,
	"available":    DhcpOptionsLifecycleStateAvailable,
	"terminating":  DhcpOptionsLifecycleStateTerminating,
	"terminated":   DhcpOptionsLifecycleStateTerminated,
}

// GetDhcpOptionsLifecycleStateEnumValues Enumerates the set of values for DhcpOptionsLifecycleStateEnum
func GetDhcpOptionsLifecycleStateEnumValues() []DhcpOptionsLifecycleStateEnum {
	values := make([]DhcpOptionsLifecycleStateEnum, 0)
	for _, v := range mappingDhcpOptionsLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetDhcpOptionsLifecycleStateEnumStringValues Enumerates the set of values in String for DhcpOptionsLifecycleStateEnum
func GetDhcpOptionsLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingDhcpOptionsLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDhcpOptionsLifecycleStateEnum(val string) (DhcpOptionsLifecycleStateEnum, bool) {
	enum, ok := mappingDhcpOptionsLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// DhcpOptionsDomainNameTypeEnum Enum with underlying type: string
type DhcpOptionsDomainNameTypeEnum string

// Set of constants representing the allowable values for DhcpOptionsDomainNameTypeEnum
const (
	DhcpOptionsDomainNameTypeSubnetDomain DhcpOptionsDomainNameTypeEnum = "SUBNET_DOMAIN"
	DhcpOptionsDomainNameTypeVcnDomain    DhcpOptionsDomainNameTypeEnum = "VCN_DOMAIN"
	DhcpOptionsDomainNameTypeCustomDomain DhcpOptionsDomainNameTypeEnum = "CUSTOM_DOMAIN"
)

var mappingDhcpOptionsDomainNameTypeEnum = map[string]DhcpOptionsDomainNameTypeEnum{
	"SUBNET_DOMAIN": DhcpOptionsDomainNameTypeSubnetDomain,
	"VCN_DOMAIN":    DhcpOptionsDomainNameTypeVcnDomain,
	"CUSTOM_DOMAIN": DhcpOptionsDomainNameTypeCustomDomain,
}

var mappingDhcpOptionsDomainNameTypeEnumLowerCase = map[string]DhcpOptionsDomainNameTypeEnum{
	"subnet_domain": DhcpOptionsDomainNameTypeSubnetDomain,
	"vcn_domain":    DhcpOptionsDomainNameTypeVcnDomain,
	"custom_domain": DhcpOptionsDomainNameTypeCustomDomain,
}

// GetDhcpOptionsDomainNameTypeEnumValues Enumerates the set of values for DhcpOptionsDomainNameTypeEnum
func GetDhcpOptionsDomainNameTypeEnumValues() []DhcpOptionsDomainNameTypeEnum {
	values := make([]DhcpOptionsDomainNameTypeEnum, 0)
	for _, v := range mappingDhcpOptionsDomainNameTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetDhcpOptionsDomainNameTypeEnumStringValues Enumerates the set of values in String for DhcpOptionsDomainNameTypeEnum
func GetDhcpOptionsDomainNameTypeEnumStringValues() []string {
	return []string{
		"SUBNET_DOMAIN",
		"VCN_DOMAIN",
		"CUSTOM_DOMAIN",
	}
}

// GetMappingDhcpOptionsDomainNameTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDhcpOptionsDomainNameTypeEnum(val string) (DhcpOptionsDomainNameTypeEnum, bool) {
	enum, ok := mappingDhcpOptionsDomainNameTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
