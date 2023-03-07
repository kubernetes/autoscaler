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

// CreateInternalPublicIpDetails This structure is used when creating publicIps for internal clients.
type CreateInternalPublicIpDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment to contain the public IP. For ephemeral public IPs,
	// you must set this to the private IP's
	//  compartment OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Defines when the public IP is deleted and released back to the Oracle Cloud
	// Infrastructure public IP pool. For more information, see
	// Public IP Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingpublicIPs.htm).
	Lifetime CreateInternalPublicIpDetailsLifetimeEnum `mandatory:"true" json:"lifetime"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the entity to assign the public IP to.
	AssignedEntityId *string `mandatory:"false" json:"assignedEntityId"`

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

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the pool object created by the current tenancy
	PublicIpPoolId *string `mandatory:"false" json:"publicIpPoolId"`

	// Only provided when no publicIpPoolId is specified.
	InternalPoolName CreateInternalPublicIpDetailsInternalPoolNameEnum `mandatory:"false" json:"internalPoolName,omitempty"`
}

func (m CreateInternalPublicIpDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateInternalPublicIpDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingCreateInternalPublicIpDetailsLifetimeEnum[string(m.Lifetime)]; !ok && m.Lifetime != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Lifetime: %s. Supported values are: %s.", m.Lifetime, strings.Join(GetCreateInternalPublicIpDetailsLifetimeEnumStringValues(), ",")))
	}

	if _, ok := mappingCreateInternalPublicIpDetailsInternalPoolNameEnum[string(m.InternalPoolName)]; !ok && m.InternalPoolName != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for InternalPoolName: %s. Supported values are: %s.", m.InternalPoolName, strings.Join(GetCreateInternalPublicIpDetailsInternalPoolNameEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateInternalPublicIpDetailsLifetimeEnum Enum with underlying type: string
type CreateInternalPublicIpDetailsLifetimeEnum string

// Set of constants representing the allowable values for CreateInternalPublicIpDetailsLifetimeEnum
const (
	CreateInternalPublicIpDetailsLifetimeEphemeral CreateInternalPublicIpDetailsLifetimeEnum = "EPHEMERAL"
	CreateInternalPublicIpDetailsLifetimeReserved  CreateInternalPublicIpDetailsLifetimeEnum = "RESERVED"
)

var mappingCreateInternalPublicIpDetailsLifetimeEnum = map[string]CreateInternalPublicIpDetailsLifetimeEnum{
	"EPHEMERAL": CreateInternalPublicIpDetailsLifetimeEphemeral,
	"RESERVED":  CreateInternalPublicIpDetailsLifetimeReserved,
}

// GetCreateInternalPublicIpDetailsLifetimeEnumValues Enumerates the set of values for CreateInternalPublicIpDetailsLifetimeEnum
func GetCreateInternalPublicIpDetailsLifetimeEnumValues() []CreateInternalPublicIpDetailsLifetimeEnum {
	values := make([]CreateInternalPublicIpDetailsLifetimeEnum, 0)
	for _, v := range mappingCreateInternalPublicIpDetailsLifetimeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateInternalPublicIpDetailsLifetimeEnumStringValues Enumerates the set of values in String for CreateInternalPublicIpDetailsLifetimeEnum
func GetCreateInternalPublicIpDetailsLifetimeEnumStringValues() []string {
	return []string{
		"EPHEMERAL",
		"RESERVED",
	}
}

// CreateInternalPublicIpDetailsInternalPoolNameEnum Enum with underlying type: string
type CreateInternalPublicIpDetailsInternalPoolNameEnum string

// Set of constants representing the allowable values for CreateInternalPublicIpDetailsInternalPoolNameEnum
const (
	CreateInternalPublicIpDetailsInternalPoolNameExternal   CreateInternalPublicIpDetailsInternalPoolNameEnum = "EXTERNAL"
	CreateInternalPublicIpDetailsInternalPoolNameSociEgress CreateInternalPublicIpDetailsInternalPoolNameEnum = "SOCI_EGRESS"
)

var mappingCreateInternalPublicIpDetailsInternalPoolNameEnum = map[string]CreateInternalPublicIpDetailsInternalPoolNameEnum{
	"EXTERNAL":    CreateInternalPublicIpDetailsInternalPoolNameExternal,
	"SOCI_EGRESS": CreateInternalPublicIpDetailsInternalPoolNameSociEgress,
}

// GetCreateInternalPublicIpDetailsInternalPoolNameEnumValues Enumerates the set of values for CreateInternalPublicIpDetailsInternalPoolNameEnum
func GetCreateInternalPublicIpDetailsInternalPoolNameEnumValues() []CreateInternalPublicIpDetailsInternalPoolNameEnum {
	values := make([]CreateInternalPublicIpDetailsInternalPoolNameEnum, 0)
	for _, v := range mappingCreateInternalPublicIpDetailsInternalPoolNameEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateInternalPublicIpDetailsInternalPoolNameEnumStringValues Enumerates the set of values in String for CreateInternalPublicIpDetailsInternalPoolNameEnum
func GetCreateInternalPublicIpDetailsInternalPoolNameEnumStringValues() []string {
	return []string{
		"EXTERNAL",
		"SOCI_EGRESS",
	}
}
