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

// MigrateDrgDetails The request to update a `DrgAttachment` and migrate the `Drg` to the destination DRG type.
type MigrateDrgDetails struct {

	// Type of the `Drg` to be migrated to.
	DestinationDrgType MigrateDrgDetailsDestinationDrgTypeEnum `mandatory:"true" json:"destinationDrgType"`

	// The DRG attachment's Oracle ID (OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm)).
	DrgAttachmentId *string `mandatory:"false" json:"drgAttachmentId"`

	// NextHop target's MPLS label.
	MplsLabel *string `mandatory:"false" json:"mplsLabel"`

	// The string in the form ASN:mplsLabel.
	RouteTarget *string `mandatory:"false" json:"routeTarget"`
}

func (m MigrateDrgDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m MigrateDrgDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingMigrateDrgDetailsDestinationDrgTypeEnum[string(m.DestinationDrgType)]; !ok && m.DestinationDrgType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DestinationDrgType: %s. Supported values are: %s.", m.DestinationDrgType, strings.Join(GetMigrateDrgDetailsDestinationDrgTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MigrateDrgDetailsDestinationDrgTypeEnum Enum with underlying type: string
type MigrateDrgDetailsDestinationDrgTypeEnum string

// Set of constants representing the allowable values for MigrateDrgDetailsDestinationDrgTypeEnum
const (
	MigrateDrgDetailsDestinationDrgTypeClassical  MigrateDrgDetailsDestinationDrgTypeEnum = "DRG_CLASSICAL"
	MigrateDrgDetailsDestinationDrgTypeTransitHub MigrateDrgDetailsDestinationDrgTypeEnum = "DRG_TRANSIT_HUB"
)

var mappingMigrateDrgDetailsDestinationDrgTypeEnum = map[string]MigrateDrgDetailsDestinationDrgTypeEnum{
	"DRG_CLASSICAL":   MigrateDrgDetailsDestinationDrgTypeClassical,
	"DRG_TRANSIT_HUB": MigrateDrgDetailsDestinationDrgTypeTransitHub,
}

// GetMigrateDrgDetailsDestinationDrgTypeEnumValues Enumerates the set of values for MigrateDrgDetailsDestinationDrgTypeEnum
func GetMigrateDrgDetailsDestinationDrgTypeEnumValues() []MigrateDrgDetailsDestinationDrgTypeEnum {
	values := make([]MigrateDrgDetailsDestinationDrgTypeEnum, 0)
	for _, v := range mappingMigrateDrgDetailsDestinationDrgTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetMigrateDrgDetailsDestinationDrgTypeEnumStringValues Enumerates the set of values in String for MigrateDrgDetailsDestinationDrgTypeEnum
func GetMigrateDrgDetailsDestinationDrgTypeEnumStringValues() []string {
	return []string{
		"DRG_CLASSICAL",
		"DRG_TRANSIT_HUB",
	}
}
