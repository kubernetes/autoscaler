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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// CreateInternalDnsRecordDetails This structure is used when creating DnsRecord for internal clients.
type CreateInternalDnsRecordDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment to contain the DnsRecord.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the Internal Hosted Zone the DnsRecord belongs to.
	InternalHostedZoneId *string `mandatory:"true" json:"internalHostedZoneId"`

	// Name of the DnsRecord.
	// -*A:* Partially Qualified DNS Name that will be mapped to the IPv4 address
	Name *string `mandatory:"true" json:"name"`

	// Type of Dns Record according to RFC 1035 (https://tools.ietf.org/html/rfc1035).
	// Currently supported list of types are the following.
	// -*A:* Type 1, a hostname to IPv4 address
	Type CreateInternalDnsRecordDetailsTypeEnum `mandatory:"true" json:"type"`

	// Value for the DnsRecord.
	// -*A:* One or more IPv4 addresses. Comma separated.
	Value *string `mandatory:"true" json:"value"`

	// Time to live value in seconds for the DnsRecord, according to RFC 1035 (https://tools.ietf.org/html/rfc1035).
	// Defaults to 86400.
	Ttl *int `mandatory:"false" json:"ttl"`
}

func (m CreateInternalDnsRecordDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateInternalDnsRecordDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingCreateInternalDnsRecordDetailsTypeEnum[string(m.Type)]; !ok && m.Type != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Type: %s. Supported values are: %s.", m.Type, strings.Join(GetCreateInternalDnsRecordDetailsTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateInternalDnsRecordDetailsTypeEnum Enum with underlying type: string
type CreateInternalDnsRecordDetailsTypeEnum string

// Set of constants representing the allowable values for CreateInternalDnsRecordDetailsTypeEnum
const (
	CreateInternalDnsRecordDetailsTypeA CreateInternalDnsRecordDetailsTypeEnum = "A"
)

var mappingCreateInternalDnsRecordDetailsTypeEnum = map[string]CreateInternalDnsRecordDetailsTypeEnum{
	"A": CreateInternalDnsRecordDetailsTypeA,
}

// GetCreateInternalDnsRecordDetailsTypeEnumValues Enumerates the set of values for CreateInternalDnsRecordDetailsTypeEnum
func GetCreateInternalDnsRecordDetailsTypeEnumValues() []CreateInternalDnsRecordDetailsTypeEnum {
	values := make([]CreateInternalDnsRecordDetailsTypeEnum, 0)
	for _, v := range mappingCreateInternalDnsRecordDetailsTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateInternalDnsRecordDetailsTypeEnumStringValues Enumerates the set of values in String for CreateInternalDnsRecordDetailsTypeEnum
func GetCreateInternalDnsRecordDetailsTypeEnumStringValues() []string {
	return []string{
		"A",
	}
}
