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

// CreateDhcpDetails The representation of CreateDhcpDetails
type CreateDhcpDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment to contain the set of DHCP options.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A set of DHCP options.
	Options []DhcpOption `mandatory:"true" json:"options"`

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
	DomainNameType CreateDhcpDetailsDomainNameTypeEnum `mandatory:"false" json:"domainNameType,omitempty"`
}

func (m CreateDhcpDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateDhcpDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCreateDhcpDetailsDomainNameTypeEnum(string(m.DomainNameType)); !ok && m.DomainNameType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DomainNameType: %s. Supported values are: %s.", m.DomainNameType, strings.Join(GetCreateDhcpDetailsDomainNameTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *CreateDhcpDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DefinedTags    map[string]map[string]interface{}   `json:"definedTags"`
		DisplayName    *string                             `json:"displayName"`
		FreeformTags   map[string]string                   `json:"freeformTags"`
		DomainNameType CreateDhcpDetailsDomainNameTypeEnum `json:"domainNameType"`
		CompartmentId  *string                             `json:"compartmentId"`
		Options        []dhcpoption                        `json:"options"`
		VcnId          *string                             `json:"vcnId"`
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
	m.VcnId = model.VcnId

	return
}

// CreateDhcpDetailsDomainNameTypeEnum Enum with underlying type: string
type CreateDhcpDetailsDomainNameTypeEnum string

// Set of constants representing the allowable values for CreateDhcpDetailsDomainNameTypeEnum
const (
	CreateDhcpDetailsDomainNameTypeSubnetDomain CreateDhcpDetailsDomainNameTypeEnum = "SUBNET_DOMAIN"
	CreateDhcpDetailsDomainNameTypeVcnDomain    CreateDhcpDetailsDomainNameTypeEnum = "VCN_DOMAIN"
	CreateDhcpDetailsDomainNameTypeCustomDomain CreateDhcpDetailsDomainNameTypeEnum = "CUSTOM_DOMAIN"
)

var mappingCreateDhcpDetailsDomainNameTypeEnum = map[string]CreateDhcpDetailsDomainNameTypeEnum{
	"SUBNET_DOMAIN": CreateDhcpDetailsDomainNameTypeSubnetDomain,
	"VCN_DOMAIN":    CreateDhcpDetailsDomainNameTypeVcnDomain,
	"CUSTOM_DOMAIN": CreateDhcpDetailsDomainNameTypeCustomDomain,
}

var mappingCreateDhcpDetailsDomainNameTypeEnumLowerCase = map[string]CreateDhcpDetailsDomainNameTypeEnum{
	"subnet_domain": CreateDhcpDetailsDomainNameTypeSubnetDomain,
	"vcn_domain":    CreateDhcpDetailsDomainNameTypeVcnDomain,
	"custom_domain": CreateDhcpDetailsDomainNameTypeCustomDomain,
}

// GetCreateDhcpDetailsDomainNameTypeEnumValues Enumerates the set of values for CreateDhcpDetailsDomainNameTypeEnum
func GetCreateDhcpDetailsDomainNameTypeEnumValues() []CreateDhcpDetailsDomainNameTypeEnum {
	values := make([]CreateDhcpDetailsDomainNameTypeEnum, 0)
	for _, v := range mappingCreateDhcpDetailsDomainNameTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateDhcpDetailsDomainNameTypeEnumStringValues Enumerates the set of values in String for CreateDhcpDetailsDomainNameTypeEnum
func GetCreateDhcpDetailsDomainNameTypeEnumStringValues() []string {
	return []string{
		"SUBNET_DOMAIN",
		"VCN_DOMAIN",
		"CUSTOM_DOMAIN",
	}
}

// GetMappingCreateDhcpDetailsDomainNameTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateDhcpDetailsDomainNameTypeEnum(val string) (CreateDhcpDetailsDomainNameTypeEnum, bool) {
	enum, ok := mappingCreateDhcpDetailsDomainNameTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
