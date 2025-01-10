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

// FastConnectProviderService A service offering from a supported provider. For more information,
// see FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
type FastConnectProviderService struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the service offered by the provider.
	Id *string `mandatory:"true" json:"id"`

	// Who is responsible for managing the private peering BGP information.
	PrivatePeeringBgpManagement FastConnectProviderServicePrivatePeeringBgpManagementEnum `mandatory:"true" json:"privatePeeringBgpManagement"`

	// The name of the provider.
	ProviderName *string `mandatory:"true" json:"providerName"`

	// The name of the service offered by the provider.
	ProviderServiceName *string `mandatory:"true" json:"providerServiceName"`

	// Who is responsible for managing the public peering BGP information.
	PublicPeeringBgpManagement FastConnectProviderServicePublicPeeringBgpManagementEnum `mandatory:"true" json:"publicPeeringBgpManagement"`

	// Who is responsible for managing the ASN information for the network at the other end
	// of the connection from Oracle.
	CustomerAsnManagement FastConnectProviderServiceCustomerAsnManagementEnum `mandatory:"true" json:"customerAsnManagement"`

	// Who is responsible for managing the provider service key.
	ProviderServiceKeyManagement FastConnectProviderServiceProviderServiceKeyManagementEnum `mandatory:"true" json:"providerServiceKeyManagement"`

	// Who is responsible for managing the virtual circuit bandwidth.
	BandwithShapeManagement FastConnectProviderServiceBandwithShapeManagementEnum `mandatory:"true" json:"bandwithShapeManagement"`

	// Total number of cross-connect or cross-connect groups required for the virtual circuit.
	RequiredTotalCrossConnects *int `mandatory:"true" json:"requiredTotalCrossConnects"`

	// Provider service type.
	Type FastConnectProviderServiceTypeEnum `mandatory:"true" json:"type"`

	// The location of the provider's website or portal. This portal is where you can get information
	// about the provider service, create a virtual circuit connection from the provider to Oracle
	// Cloud Infrastructure, and retrieve your provider service key for that virtual circuit connection.
	// Example: `https://example.com`
	Description *string `mandatory:"false" json:"description"`

	// An array of virtual circuit types supported by this service.
	SupportedVirtualCircuitTypes []FastConnectProviderServiceSupportedVirtualCircuitTypesEnum `mandatory:"false" json:"supportedVirtualCircuitTypes,omitempty"`
}

func (m FastConnectProviderService) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m FastConnectProviderService) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingFastConnectProviderServicePrivatePeeringBgpManagementEnum(string(m.PrivatePeeringBgpManagement)); !ok && m.PrivatePeeringBgpManagement != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for PrivatePeeringBgpManagement: %s. Supported values are: %s.", m.PrivatePeeringBgpManagement, strings.Join(GetFastConnectProviderServicePrivatePeeringBgpManagementEnumStringValues(), ",")))
	}
	if _, ok := GetMappingFastConnectProviderServicePublicPeeringBgpManagementEnum(string(m.PublicPeeringBgpManagement)); !ok && m.PublicPeeringBgpManagement != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for PublicPeeringBgpManagement: %s. Supported values are: %s.", m.PublicPeeringBgpManagement, strings.Join(GetFastConnectProviderServicePublicPeeringBgpManagementEnumStringValues(), ",")))
	}
	if _, ok := GetMappingFastConnectProviderServiceCustomerAsnManagementEnum(string(m.CustomerAsnManagement)); !ok && m.CustomerAsnManagement != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for CustomerAsnManagement: %s. Supported values are: %s.", m.CustomerAsnManagement, strings.Join(GetFastConnectProviderServiceCustomerAsnManagementEnumStringValues(), ",")))
	}
	if _, ok := GetMappingFastConnectProviderServiceProviderServiceKeyManagementEnum(string(m.ProviderServiceKeyManagement)); !ok && m.ProviderServiceKeyManagement != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ProviderServiceKeyManagement: %s. Supported values are: %s.", m.ProviderServiceKeyManagement, strings.Join(GetFastConnectProviderServiceProviderServiceKeyManagementEnumStringValues(), ",")))
	}
	if _, ok := GetMappingFastConnectProviderServiceBandwithShapeManagementEnum(string(m.BandwithShapeManagement)); !ok && m.BandwithShapeManagement != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BandwithShapeManagement: %s. Supported values are: %s.", m.BandwithShapeManagement, strings.Join(GetFastConnectProviderServiceBandwithShapeManagementEnumStringValues(), ",")))
	}
	if _, ok := GetMappingFastConnectProviderServiceTypeEnum(string(m.Type)); !ok && m.Type != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Type: %s. Supported values are: %s.", m.Type, strings.Join(GetFastConnectProviderServiceTypeEnumStringValues(), ",")))
	}

	for _, val := range m.SupportedVirtualCircuitTypes {
		if _, ok := GetMappingFastConnectProviderServiceSupportedVirtualCircuitTypesEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SupportedVirtualCircuitTypes: %s. Supported values are: %s.", val, strings.Join(GetFastConnectProviderServiceSupportedVirtualCircuitTypesEnumStringValues(), ",")))
		}
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// FastConnectProviderServicePrivatePeeringBgpManagementEnum Enum with underlying type: string
type FastConnectProviderServicePrivatePeeringBgpManagementEnum string

// Set of constants representing the allowable values for FastConnectProviderServicePrivatePeeringBgpManagementEnum
const (
	FastConnectProviderServicePrivatePeeringBgpManagementCustomerManaged FastConnectProviderServicePrivatePeeringBgpManagementEnum = "CUSTOMER_MANAGED"
	FastConnectProviderServicePrivatePeeringBgpManagementProviderManaged FastConnectProviderServicePrivatePeeringBgpManagementEnum = "PROVIDER_MANAGED"
	FastConnectProviderServicePrivatePeeringBgpManagementOracleManaged   FastConnectProviderServicePrivatePeeringBgpManagementEnum = "ORACLE_MANAGED"
)

var mappingFastConnectProviderServicePrivatePeeringBgpManagementEnum = map[string]FastConnectProviderServicePrivatePeeringBgpManagementEnum{
	"CUSTOMER_MANAGED": FastConnectProviderServicePrivatePeeringBgpManagementCustomerManaged,
	"PROVIDER_MANAGED": FastConnectProviderServicePrivatePeeringBgpManagementProviderManaged,
	"ORACLE_MANAGED":   FastConnectProviderServicePrivatePeeringBgpManagementOracleManaged,
}

var mappingFastConnectProviderServicePrivatePeeringBgpManagementEnumLowerCase = map[string]FastConnectProviderServicePrivatePeeringBgpManagementEnum{
	"customer_managed": FastConnectProviderServicePrivatePeeringBgpManagementCustomerManaged,
	"provider_managed": FastConnectProviderServicePrivatePeeringBgpManagementProviderManaged,
	"oracle_managed":   FastConnectProviderServicePrivatePeeringBgpManagementOracleManaged,
}

// GetFastConnectProviderServicePrivatePeeringBgpManagementEnumValues Enumerates the set of values for FastConnectProviderServicePrivatePeeringBgpManagementEnum
func GetFastConnectProviderServicePrivatePeeringBgpManagementEnumValues() []FastConnectProviderServicePrivatePeeringBgpManagementEnum {
	values := make([]FastConnectProviderServicePrivatePeeringBgpManagementEnum, 0)
	for _, v := range mappingFastConnectProviderServicePrivatePeeringBgpManagementEnum {
		values = append(values, v)
	}
	return values
}

// GetFastConnectProviderServicePrivatePeeringBgpManagementEnumStringValues Enumerates the set of values in String for FastConnectProviderServicePrivatePeeringBgpManagementEnum
func GetFastConnectProviderServicePrivatePeeringBgpManagementEnumStringValues() []string {
	return []string{
		"CUSTOMER_MANAGED",
		"PROVIDER_MANAGED",
		"ORACLE_MANAGED",
	}
}

// GetMappingFastConnectProviderServicePrivatePeeringBgpManagementEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingFastConnectProviderServicePrivatePeeringBgpManagementEnum(val string) (FastConnectProviderServicePrivatePeeringBgpManagementEnum, bool) {
	enum, ok := mappingFastConnectProviderServicePrivatePeeringBgpManagementEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// FastConnectProviderServicePublicPeeringBgpManagementEnum Enum with underlying type: string
type FastConnectProviderServicePublicPeeringBgpManagementEnum string

// Set of constants representing the allowable values for FastConnectProviderServicePublicPeeringBgpManagementEnum
const (
	FastConnectProviderServicePublicPeeringBgpManagementCustomerManaged FastConnectProviderServicePublicPeeringBgpManagementEnum = "CUSTOMER_MANAGED"
	FastConnectProviderServicePublicPeeringBgpManagementProviderManaged FastConnectProviderServicePublicPeeringBgpManagementEnum = "PROVIDER_MANAGED"
	FastConnectProviderServicePublicPeeringBgpManagementOracleManaged   FastConnectProviderServicePublicPeeringBgpManagementEnum = "ORACLE_MANAGED"
)

var mappingFastConnectProviderServicePublicPeeringBgpManagementEnum = map[string]FastConnectProviderServicePublicPeeringBgpManagementEnum{
	"CUSTOMER_MANAGED": FastConnectProviderServicePublicPeeringBgpManagementCustomerManaged,
	"PROVIDER_MANAGED": FastConnectProviderServicePublicPeeringBgpManagementProviderManaged,
	"ORACLE_MANAGED":   FastConnectProviderServicePublicPeeringBgpManagementOracleManaged,
}

var mappingFastConnectProviderServicePublicPeeringBgpManagementEnumLowerCase = map[string]FastConnectProviderServicePublicPeeringBgpManagementEnum{
	"customer_managed": FastConnectProviderServicePublicPeeringBgpManagementCustomerManaged,
	"provider_managed": FastConnectProviderServicePublicPeeringBgpManagementProviderManaged,
	"oracle_managed":   FastConnectProviderServicePublicPeeringBgpManagementOracleManaged,
}

// GetFastConnectProviderServicePublicPeeringBgpManagementEnumValues Enumerates the set of values for FastConnectProviderServicePublicPeeringBgpManagementEnum
func GetFastConnectProviderServicePublicPeeringBgpManagementEnumValues() []FastConnectProviderServicePublicPeeringBgpManagementEnum {
	values := make([]FastConnectProviderServicePublicPeeringBgpManagementEnum, 0)
	for _, v := range mappingFastConnectProviderServicePublicPeeringBgpManagementEnum {
		values = append(values, v)
	}
	return values
}

// GetFastConnectProviderServicePublicPeeringBgpManagementEnumStringValues Enumerates the set of values in String for FastConnectProviderServicePublicPeeringBgpManagementEnum
func GetFastConnectProviderServicePublicPeeringBgpManagementEnumStringValues() []string {
	return []string{
		"CUSTOMER_MANAGED",
		"PROVIDER_MANAGED",
		"ORACLE_MANAGED",
	}
}

// GetMappingFastConnectProviderServicePublicPeeringBgpManagementEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingFastConnectProviderServicePublicPeeringBgpManagementEnum(val string) (FastConnectProviderServicePublicPeeringBgpManagementEnum, bool) {
	enum, ok := mappingFastConnectProviderServicePublicPeeringBgpManagementEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// FastConnectProviderServiceSupportedVirtualCircuitTypesEnum Enum with underlying type: string
type FastConnectProviderServiceSupportedVirtualCircuitTypesEnum string

// Set of constants representing the allowable values for FastConnectProviderServiceSupportedVirtualCircuitTypesEnum
const (
	FastConnectProviderServiceSupportedVirtualCircuitTypesPublic  FastConnectProviderServiceSupportedVirtualCircuitTypesEnum = "PUBLIC"
	FastConnectProviderServiceSupportedVirtualCircuitTypesPrivate FastConnectProviderServiceSupportedVirtualCircuitTypesEnum = "PRIVATE"
)

var mappingFastConnectProviderServiceSupportedVirtualCircuitTypesEnum = map[string]FastConnectProviderServiceSupportedVirtualCircuitTypesEnum{
	"PUBLIC":  FastConnectProviderServiceSupportedVirtualCircuitTypesPublic,
	"PRIVATE": FastConnectProviderServiceSupportedVirtualCircuitTypesPrivate,
}

var mappingFastConnectProviderServiceSupportedVirtualCircuitTypesEnumLowerCase = map[string]FastConnectProviderServiceSupportedVirtualCircuitTypesEnum{
	"public":  FastConnectProviderServiceSupportedVirtualCircuitTypesPublic,
	"private": FastConnectProviderServiceSupportedVirtualCircuitTypesPrivate,
}

// GetFastConnectProviderServiceSupportedVirtualCircuitTypesEnumValues Enumerates the set of values for FastConnectProviderServiceSupportedVirtualCircuitTypesEnum
func GetFastConnectProviderServiceSupportedVirtualCircuitTypesEnumValues() []FastConnectProviderServiceSupportedVirtualCircuitTypesEnum {
	values := make([]FastConnectProviderServiceSupportedVirtualCircuitTypesEnum, 0)
	for _, v := range mappingFastConnectProviderServiceSupportedVirtualCircuitTypesEnum {
		values = append(values, v)
	}
	return values
}

// GetFastConnectProviderServiceSupportedVirtualCircuitTypesEnumStringValues Enumerates the set of values in String for FastConnectProviderServiceSupportedVirtualCircuitTypesEnum
func GetFastConnectProviderServiceSupportedVirtualCircuitTypesEnumStringValues() []string {
	return []string{
		"PUBLIC",
		"PRIVATE",
	}
}

// GetMappingFastConnectProviderServiceSupportedVirtualCircuitTypesEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingFastConnectProviderServiceSupportedVirtualCircuitTypesEnum(val string) (FastConnectProviderServiceSupportedVirtualCircuitTypesEnum, bool) {
	enum, ok := mappingFastConnectProviderServiceSupportedVirtualCircuitTypesEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// FastConnectProviderServiceCustomerAsnManagementEnum Enum with underlying type: string
type FastConnectProviderServiceCustomerAsnManagementEnum string

// Set of constants representing the allowable values for FastConnectProviderServiceCustomerAsnManagementEnum
const (
	FastConnectProviderServiceCustomerAsnManagementCustomerManaged FastConnectProviderServiceCustomerAsnManagementEnum = "CUSTOMER_MANAGED"
	FastConnectProviderServiceCustomerAsnManagementProviderManaged FastConnectProviderServiceCustomerAsnManagementEnum = "PROVIDER_MANAGED"
	FastConnectProviderServiceCustomerAsnManagementOracleManaged   FastConnectProviderServiceCustomerAsnManagementEnum = "ORACLE_MANAGED"
)

var mappingFastConnectProviderServiceCustomerAsnManagementEnum = map[string]FastConnectProviderServiceCustomerAsnManagementEnum{
	"CUSTOMER_MANAGED": FastConnectProviderServiceCustomerAsnManagementCustomerManaged,
	"PROVIDER_MANAGED": FastConnectProviderServiceCustomerAsnManagementProviderManaged,
	"ORACLE_MANAGED":   FastConnectProviderServiceCustomerAsnManagementOracleManaged,
}

var mappingFastConnectProviderServiceCustomerAsnManagementEnumLowerCase = map[string]FastConnectProviderServiceCustomerAsnManagementEnum{
	"customer_managed": FastConnectProviderServiceCustomerAsnManagementCustomerManaged,
	"provider_managed": FastConnectProviderServiceCustomerAsnManagementProviderManaged,
	"oracle_managed":   FastConnectProviderServiceCustomerAsnManagementOracleManaged,
}

// GetFastConnectProviderServiceCustomerAsnManagementEnumValues Enumerates the set of values for FastConnectProviderServiceCustomerAsnManagementEnum
func GetFastConnectProviderServiceCustomerAsnManagementEnumValues() []FastConnectProviderServiceCustomerAsnManagementEnum {
	values := make([]FastConnectProviderServiceCustomerAsnManagementEnum, 0)
	for _, v := range mappingFastConnectProviderServiceCustomerAsnManagementEnum {
		values = append(values, v)
	}
	return values
}

// GetFastConnectProviderServiceCustomerAsnManagementEnumStringValues Enumerates the set of values in String for FastConnectProviderServiceCustomerAsnManagementEnum
func GetFastConnectProviderServiceCustomerAsnManagementEnumStringValues() []string {
	return []string{
		"CUSTOMER_MANAGED",
		"PROVIDER_MANAGED",
		"ORACLE_MANAGED",
	}
}

// GetMappingFastConnectProviderServiceCustomerAsnManagementEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingFastConnectProviderServiceCustomerAsnManagementEnum(val string) (FastConnectProviderServiceCustomerAsnManagementEnum, bool) {
	enum, ok := mappingFastConnectProviderServiceCustomerAsnManagementEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// FastConnectProviderServiceProviderServiceKeyManagementEnum Enum with underlying type: string
type FastConnectProviderServiceProviderServiceKeyManagementEnum string

// Set of constants representing the allowable values for FastConnectProviderServiceProviderServiceKeyManagementEnum
const (
	FastConnectProviderServiceProviderServiceKeyManagementCustomerManaged FastConnectProviderServiceProviderServiceKeyManagementEnum = "CUSTOMER_MANAGED"
	FastConnectProviderServiceProviderServiceKeyManagementProviderManaged FastConnectProviderServiceProviderServiceKeyManagementEnum = "PROVIDER_MANAGED"
	FastConnectProviderServiceProviderServiceKeyManagementOracleManaged   FastConnectProviderServiceProviderServiceKeyManagementEnum = "ORACLE_MANAGED"
)

var mappingFastConnectProviderServiceProviderServiceKeyManagementEnum = map[string]FastConnectProviderServiceProviderServiceKeyManagementEnum{
	"CUSTOMER_MANAGED": FastConnectProviderServiceProviderServiceKeyManagementCustomerManaged,
	"PROVIDER_MANAGED": FastConnectProviderServiceProviderServiceKeyManagementProviderManaged,
	"ORACLE_MANAGED":   FastConnectProviderServiceProviderServiceKeyManagementOracleManaged,
}

var mappingFastConnectProviderServiceProviderServiceKeyManagementEnumLowerCase = map[string]FastConnectProviderServiceProviderServiceKeyManagementEnum{
	"customer_managed": FastConnectProviderServiceProviderServiceKeyManagementCustomerManaged,
	"provider_managed": FastConnectProviderServiceProviderServiceKeyManagementProviderManaged,
	"oracle_managed":   FastConnectProviderServiceProviderServiceKeyManagementOracleManaged,
}

// GetFastConnectProviderServiceProviderServiceKeyManagementEnumValues Enumerates the set of values for FastConnectProviderServiceProviderServiceKeyManagementEnum
func GetFastConnectProviderServiceProviderServiceKeyManagementEnumValues() []FastConnectProviderServiceProviderServiceKeyManagementEnum {
	values := make([]FastConnectProviderServiceProviderServiceKeyManagementEnum, 0)
	for _, v := range mappingFastConnectProviderServiceProviderServiceKeyManagementEnum {
		values = append(values, v)
	}
	return values
}

// GetFastConnectProviderServiceProviderServiceKeyManagementEnumStringValues Enumerates the set of values in String for FastConnectProviderServiceProviderServiceKeyManagementEnum
func GetFastConnectProviderServiceProviderServiceKeyManagementEnumStringValues() []string {
	return []string{
		"CUSTOMER_MANAGED",
		"PROVIDER_MANAGED",
		"ORACLE_MANAGED",
	}
}

// GetMappingFastConnectProviderServiceProviderServiceKeyManagementEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingFastConnectProviderServiceProviderServiceKeyManagementEnum(val string) (FastConnectProviderServiceProviderServiceKeyManagementEnum, bool) {
	enum, ok := mappingFastConnectProviderServiceProviderServiceKeyManagementEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// FastConnectProviderServiceBandwithShapeManagementEnum Enum with underlying type: string
type FastConnectProviderServiceBandwithShapeManagementEnum string

// Set of constants representing the allowable values for FastConnectProviderServiceBandwithShapeManagementEnum
const (
	FastConnectProviderServiceBandwithShapeManagementCustomerManaged FastConnectProviderServiceBandwithShapeManagementEnum = "CUSTOMER_MANAGED"
	FastConnectProviderServiceBandwithShapeManagementProviderManaged FastConnectProviderServiceBandwithShapeManagementEnum = "PROVIDER_MANAGED"
	FastConnectProviderServiceBandwithShapeManagementOracleManaged   FastConnectProviderServiceBandwithShapeManagementEnum = "ORACLE_MANAGED"
)

var mappingFastConnectProviderServiceBandwithShapeManagementEnum = map[string]FastConnectProviderServiceBandwithShapeManagementEnum{
	"CUSTOMER_MANAGED": FastConnectProviderServiceBandwithShapeManagementCustomerManaged,
	"PROVIDER_MANAGED": FastConnectProviderServiceBandwithShapeManagementProviderManaged,
	"ORACLE_MANAGED":   FastConnectProviderServiceBandwithShapeManagementOracleManaged,
}

var mappingFastConnectProviderServiceBandwithShapeManagementEnumLowerCase = map[string]FastConnectProviderServiceBandwithShapeManagementEnum{
	"customer_managed": FastConnectProviderServiceBandwithShapeManagementCustomerManaged,
	"provider_managed": FastConnectProviderServiceBandwithShapeManagementProviderManaged,
	"oracle_managed":   FastConnectProviderServiceBandwithShapeManagementOracleManaged,
}

// GetFastConnectProviderServiceBandwithShapeManagementEnumValues Enumerates the set of values for FastConnectProviderServiceBandwithShapeManagementEnum
func GetFastConnectProviderServiceBandwithShapeManagementEnumValues() []FastConnectProviderServiceBandwithShapeManagementEnum {
	values := make([]FastConnectProviderServiceBandwithShapeManagementEnum, 0)
	for _, v := range mappingFastConnectProviderServiceBandwithShapeManagementEnum {
		values = append(values, v)
	}
	return values
}

// GetFastConnectProviderServiceBandwithShapeManagementEnumStringValues Enumerates the set of values in String for FastConnectProviderServiceBandwithShapeManagementEnum
func GetFastConnectProviderServiceBandwithShapeManagementEnumStringValues() []string {
	return []string{
		"CUSTOMER_MANAGED",
		"PROVIDER_MANAGED",
		"ORACLE_MANAGED",
	}
}

// GetMappingFastConnectProviderServiceBandwithShapeManagementEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingFastConnectProviderServiceBandwithShapeManagementEnum(val string) (FastConnectProviderServiceBandwithShapeManagementEnum, bool) {
	enum, ok := mappingFastConnectProviderServiceBandwithShapeManagementEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// FastConnectProviderServiceTypeEnum Enum with underlying type: string
type FastConnectProviderServiceTypeEnum string

// Set of constants representing the allowable values for FastConnectProviderServiceTypeEnum
const (
	FastConnectProviderServiceTypeLayer2 FastConnectProviderServiceTypeEnum = "LAYER2"
	FastConnectProviderServiceTypeLayer3 FastConnectProviderServiceTypeEnum = "LAYER3"
)

var mappingFastConnectProviderServiceTypeEnum = map[string]FastConnectProviderServiceTypeEnum{
	"LAYER2": FastConnectProviderServiceTypeLayer2,
	"LAYER3": FastConnectProviderServiceTypeLayer3,
}

var mappingFastConnectProviderServiceTypeEnumLowerCase = map[string]FastConnectProviderServiceTypeEnum{
	"layer2": FastConnectProviderServiceTypeLayer2,
	"layer3": FastConnectProviderServiceTypeLayer3,
}

// GetFastConnectProviderServiceTypeEnumValues Enumerates the set of values for FastConnectProviderServiceTypeEnum
func GetFastConnectProviderServiceTypeEnumValues() []FastConnectProviderServiceTypeEnum {
	values := make([]FastConnectProviderServiceTypeEnum, 0)
	for _, v := range mappingFastConnectProviderServiceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetFastConnectProviderServiceTypeEnumStringValues Enumerates the set of values in String for FastConnectProviderServiceTypeEnum
func GetFastConnectProviderServiceTypeEnumStringValues() []string {
	return []string{
		"LAYER2",
		"LAYER3",
	}
}

// GetMappingFastConnectProviderServiceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingFastConnectProviderServiceTypeEnum(val string) (FastConnectProviderServiceTypeEnum, bool) {
	enum, ok := mappingFastConnectProviderServiceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
