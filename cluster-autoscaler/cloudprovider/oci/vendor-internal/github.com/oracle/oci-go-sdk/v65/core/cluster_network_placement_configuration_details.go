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

// ClusterNetworkPlacementConfigurationDetails The location for where the instance pools in a cluster network will place instances.
type ClusterNetworkPlacementConfigurationDetails struct {

	// The availability domain to place instances.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The placement constraint when reserving hosts.
	PlacementConstraint ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum `mandatory:"false" json:"placementConstraint,omitempty"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the primary subnet to place instances. This field is deprecated.
	// Use `primaryVnicSubnets` instead to set VNIC data for instances in the pool.
	PrimarySubnetId *string `mandatory:"false" json:"primarySubnetId"`

	PrimaryVnicSubnets *InstancePoolPlacementPrimarySubnet `mandatory:"false" json:"primaryVnicSubnets"`

	// The set of secondary VNIC data for instances in the pool.
	SecondaryVnicSubnets []InstancePoolPlacementSecondaryVnicSubnet `mandatory:"false" json:"secondaryVnicSubnets"`
}

func (m ClusterNetworkPlacementConfigurationDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ClusterNetworkPlacementConfigurationDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum(string(m.PlacementConstraint)); !ok && m.PlacementConstraint != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for PlacementConstraint: %s. Supported values are: %s.", m.PlacementConstraint, strings.Join(GetClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum Enum with underlying type: string
type ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum string

// Set of constants representing the allowable values for ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum
const (
	ClusterNetworkPlacementConfigurationDetailsPlacementConstraintSingleTier                   ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum = "SINGLE_TIER"
	ClusterNetworkPlacementConfigurationDetailsPlacementConstraintSingleBlock                  ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum = "SINGLE_BLOCK"
	ClusterNetworkPlacementConfigurationDetailsPlacementConstraintPackedDistributionMultiBlock ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum = "PACKED_DISTRIBUTION_MULTI_BLOCK"
)

var mappingClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum = map[string]ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum{
	"SINGLE_TIER":                     ClusterNetworkPlacementConfigurationDetailsPlacementConstraintSingleTier,
	"SINGLE_BLOCK":                    ClusterNetworkPlacementConfigurationDetailsPlacementConstraintSingleBlock,
	"PACKED_DISTRIBUTION_MULTI_BLOCK": ClusterNetworkPlacementConfigurationDetailsPlacementConstraintPackedDistributionMultiBlock,
}

var mappingClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnumLowerCase = map[string]ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum{
	"single_tier":                     ClusterNetworkPlacementConfigurationDetailsPlacementConstraintSingleTier,
	"single_block":                    ClusterNetworkPlacementConfigurationDetailsPlacementConstraintSingleBlock,
	"packed_distribution_multi_block": ClusterNetworkPlacementConfigurationDetailsPlacementConstraintPackedDistributionMultiBlock,
}

// GetClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnumValues Enumerates the set of values for ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum
func GetClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnumValues() []ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum {
	values := make([]ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum, 0)
	for _, v := range mappingClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum {
		values = append(values, v)
	}
	return values
}

// GetClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnumStringValues Enumerates the set of values in String for ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum
func GetClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnumStringValues() []string {
	return []string{
		"SINGLE_TIER",
		"SINGLE_BLOCK",
		"PACKED_DISTRIBUTION_MULTI_BLOCK",
	}
}

// GetMappingClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum(val string) (ClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnum, bool) {
	enum, ok := mappingClusterNetworkPlacementConfigurationDetailsPlacementConstraintEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
