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

// VcnDnsResolverAssociation The information about the VCN and the DNS resolver in the association.
type VcnDnsResolverAssociation struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN in the association.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// The current state of the association.
	LifecycleState VcnDnsResolverAssociationLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DNS resolver in the association.
	DnsResolverId *string `mandatory:"false" json:"dnsResolverId"`
}

func (m VcnDnsResolverAssociation) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VcnDnsResolverAssociation) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingVcnDnsResolverAssociationLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetVcnDnsResolverAssociationLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// VcnDnsResolverAssociationLifecycleStateEnum Enum with underlying type: string
type VcnDnsResolverAssociationLifecycleStateEnum string

// Set of constants representing the allowable values for VcnDnsResolverAssociationLifecycleStateEnum
const (
	VcnDnsResolverAssociationLifecycleStateProvisioning VcnDnsResolverAssociationLifecycleStateEnum = "PROVISIONING"
	VcnDnsResolverAssociationLifecycleStateAvailable    VcnDnsResolverAssociationLifecycleStateEnum = "AVAILABLE"
	VcnDnsResolverAssociationLifecycleStateTerminating  VcnDnsResolverAssociationLifecycleStateEnum = "TERMINATING"
	VcnDnsResolverAssociationLifecycleStateTerminated   VcnDnsResolverAssociationLifecycleStateEnum = "TERMINATED"
)

var mappingVcnDnsResolverAssociationLifecycleStateEnum = map[string]VcnDnsResolverAssociationLifecycleStateEnum{
	"PROVISIONING": VcnDnsResolverAssociationLifecycleStateProvisioning,
	"AVAILABLE":    VcnDnsResolverAssociationLifecycleStateAvailable,
	"TERMINATING":  VcnDnsResolverAssociationLifecycleStateTerminating,
	"TERMINATED":   VcnDnsResolverAssociationLifecycleStateTerminated,
}

var mappingVcnDnsResolverAssociationLifecycleStateEnumLowerCase = map[string]VcnDnsResolverAssociationLifecycleStateEnum{
	"provisioning": VcnDnsResolverAssociationLifecycleStateProvisioning,
	"available":    VcnDnsResolverAssociationLifecycleStateAvailable,
	"terminating":  VcnDnsResolverAssociationLifecycleStateTerminating,
	"terminated":   VcnDnsResolverAssociationLifecycleStateTerminated,
}

// GetVcnDnsResolverAssociationLifecycleStateEnumValues Enumerates the set of values for VcnDnsResolverAssociationLifecycleStateEnum
func GetVcnDnsResolverAssociationLifecycleStateEnumValues() []VcnDnsResolverAssociationLifecycleStateEnum {
	values := make([]VcnDnsResolverAssociationLifecycleStateEnum, 0)
	for _, v := range mappingVcnDnsResolverAssociationLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVcnDnsResolverAssociationLifecycleStateEnumStringValues Enumerates the set of values in String for VcnDnsResolverAssociationLifecycleStateEnum
func GetVcnDnsResolverAssociationLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingVcnDnsResolverAssociationLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVcnDnsResolverAssociationLifecycleStateEnum(val string) (VcnDnsResolverAssociationLifecycleStateEnum, bool) {
	enum, ok := mappingVcnDnsResolverAssociationLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
