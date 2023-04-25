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

// InternalDnsRecord DnsRecord representing a single RRSet.
type InternalDnsRecord struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the DNS record.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DNS record.
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the Internal Hosted Zone the `DnsRecord` belongs to.
	InternalHostedZoneId *string `mandatory:"true" json:"internalHostedZoneId"`

	// The DnsRecord's current state.
	LifecycleState InternalDnsRecordLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Name of the DnsRecord.
	// -*A:* Partially Qualified DNS Name that will be mapped to the IPv4 address
	Name *string `mandatory:"true" json:"name"`

	// Type of Dns Record according to RFC 1035 (https://tools.ietf.org/html/rfc1035).
	// Currently supported list of types are the following.
	// -*A:* Type 1, a host name to IPv4 address
	Type InternalDnsRecordTypeEnum `mandatory:"true" json:"type"`

	// Value for the DnsRecord.
	// -*A:* One or more IPv4 addresses. Comma separated.
	Value *string `mandatory:"true" json:"value"`

	// Time to live value for the DnsRecord, according to RFC 1035 (https://tools.ietf.org/html/rfc1035).
	// Defaults to 86400.
	Ttl *int `mandatory:"false" json:"ttl"`
}

func (m InternalDnsRecord) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InternalDnsRecord) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingInternalDnsRecordLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInternalDnsRecordLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := mappingInternalDnsRecordTypeEnum[string(m.Type)]; !ok && m.Type != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Type: %s. Supported values are: %s.", m.Type, strings.Join(GetInternalDnsRecordTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InternalDnsRecordLifecycleStateEnum Enum with underlying type: string
type InternalDnsRecordLifecycleStateEnum string

// Set of constants representing the allowable values for InternalDnsRecordLifecycleStateEnum
const (
	InternalDnsRecordLifecycleStateProvisioning InternalDnsRecordLifecycleStateEnum = "PROVISIONING"
	InternalDnsRecordLifecycleStateAvailable    InternalDnsRecordLifecycleStateEnum = "AVAILABLE"
	InternalDnsRecordLifecycleStateTerminating  InternalDnsRecordLifecycleStateEnum = "TERMINATING"
	InternalDnsRecordLifecycleStateTerminated   InternalDnsRecordLifecycleStateEnum = "TERMINATED"
)

var mappingInternalDnsRecordLifecycleStateEnum = map[string]InternalDnsRecordLifecycleStateEnum{
	"PROVISIONING": InternalDnsRecordLifecycleStateProvisioning,
	"AVAILABLE":    InternalDnsRecordLifecycleStateAvailable,
	"TERMINATING":  InternalDnsRecordLifecycleStateTerminating,
	"TERMINATED":   InternalDnsRecordLifecycleStateTerminated,
}

// GetInternalDnsRecordLifecycleStateEnumValues Enumerates the set of values for InternalDnsRecordLifecycleStateEnum
func GetInternalDnsRecordLifecycleStateEnumValues() []InternalDnsRecordLifecycleStateEnum {
	values := make([]InternalDnsRecordLifecycleStateEnum, 0)
	for _, v := range mappingInternalDnsRecordLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalDnsRecordLifecycleStateEnumStringValues Enumerates the set of values in String for InternalDnsRecordLifecycleStateEnum
func GetInternalDnsRecordLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// InternalDnsRecordTypeEnum Enum with underlying type: string
type InternalDnsRecordTypeEnum string

// Set of constants representing the allowable values for InternalDnsRecordTypeEnum
const (
	InternalDnsRecordTypeA InternalDnsRecordTypeEnum = "A"
)

var mappingInternalDnsRecordTypeEnum = map[string]InternalDnsRecordTypeEnum{
	"A": InternalDnsRecordTypeA,
}

// GetInternalDnsRecordTypeEnumValues Enumerates the set of values for InternalDnsRecordTypeEnum
func GetInternalDnsRecordTypeEnumValues() []InternalDnsRecordTypeEnum {
	values := make([]InternalDnsRecordTypeEnum, 0)
	for _, v := range mappingInternalDnsRecordTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalDnsRecordTypeEnumStringValues Enumerates the set of values in String for InternalDnsRecordTypeEnum
func GetInternalDnsRecordTypeEnumStringValues() []string {
	return []string{
		"A",
	}
}
