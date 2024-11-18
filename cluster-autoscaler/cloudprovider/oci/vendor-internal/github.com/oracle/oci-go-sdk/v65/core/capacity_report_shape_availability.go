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

// CapacityReportShapeAvailability Information about the available capacity for a shape.
type CapacityReportShapeAvailability struct {

	// The fault domain for the capacity report.
	// If you do not specify the fault domain, the capacity report includes information about all fault domains.
	FaultDomain *string `mandatory:"false" json:"faultDomain"`

	// The shape that the capacity report was requested for.
	InstanceShape *string `mandatory:"false" json:"instanceShape"`

	InstanceShapeConfig *CapacityReportInstanceShapeConfig `mandatory:"false" json:"instanceShapeConfig"`

	// The total number of new instances that can be created with the specified shape configuration.
	AvailableCount *int64 `mandatory:"false" json:"availableCount"`

	// A flag denoting whether capacity is available.
	AvailabilityStatus CapacityReportShapeAvailabilityAvailabilityStatusEnum `mandatory:"false" json:"availabilityStatus,omitempty"`
}

func (m CapacityReportShapeAvailability) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CapacityReportShapeAvailability) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCapacityReportShapeAvailabilityAvailabilityStatusEnum(string(m.AvailabilityStatus)); !ok && m.AvailabilityStatus != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AvailabilityStatus: %s. Supported values are: %s.", m.AvailabilityStatus, strings.Join(GetCapacityReportShapeAvailabilityAvailabilityStatusEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CapacityReportShapeAvailabilityAvailabilityStatusEnum Enum with underlying type: string
type CapacityReportShapeAvailabilityAvailabilityStatusEnum string

// Set of constants representing the allowable values for CapacityReportShapeAvailabilityAvailabilityStatusEnum
const (
	CapacityReportShapeAvailabilityAvailabilityStatusOutOfHostCapacity    CapacityReportShapeAvailabilityAvailabilityStatusEnum = "OUT_OF_HOST_CAPACITY"
	CapacityReportShapeAvailabilityAvailabilityStatusHardwareNotSupported CapacityReportShapeAvailabilityAvailabilityStatusEnum = "HARDWARE_NOT_SUPPORTED"
	CapacityReportShapeAvailabilityAvailabilityStatusAvailable            CapacityReportShapeAvailabilityAvailabilityStatusEnum = "AVAILABLE"
)

var mappingCapacityReportShapeAvailabilityAvailabilityStatusEnum = map[string]CapacityReportShapeAvailabilityAvailabilityStatusEnum{
	"OUT_OF_HOST_CAPACITY":   CapacityReportShapeAvailabilityAvailabilityStatusOutOfHostCapacity,
	"HARDWARE_NOT_SUPPORTED": CapacityReportShapeAvailabilityAvailabilityStatusHardwareNotSupported,
	"AVAILABLE":              CapacityReportShapeAvailabilityAvailabilityStatusAvailable,
}

var mappingCapacityReportShapeAvailabilityAvailabilityStatusEnumLowerCase = map[string]CapacityReportShapeAvailabilityAvailabilityStatusEnum{
	"out_of_host_capacity":   CapacityReportShapeAvailabilityAvailabilityStatusOutOfHostCapacity,
	"hardware_not_supported": CapacityReportShapeAvailabilityAvailabilityStatusHardwareNotSupported,
	"available":              CapacityReportShapeAvailabilityAvailabilityStatusAvailable,
}

// GetCapacityReportShapeAvailabilityAvailabilityStatusEnumValues Enumerates the set of values for CapacityReportShapeAvailabilityAvailabilityStatusEnum
func GetCapacityReportShapeAvailabilityAvailabilityStatusEnumValues() []CapacityReportShapeAvailabilityAvailabilityStatusEnum {
	values := make([]CapacityReportShapeAvailabilityAvailabilityStatusEnum, 0)
	for _, v := range mappingCapacityReportShapeAvailabilityAvailabilityStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetCapacityReportShapeAvailabilityAvailabilityStatusEnumStringValues Enumerates the set of values in String for CapacityReportShapeAvailabilityAvailabilityStatusEnum
func GetCapacityReportShapeAvailabilityAvailabilityStatusEnumStringValues() []string {
	return []string{
		"OUT_OF_HOST_CAPACITY",
		"HARDWARE_NOT_SUPPORTED",
		"AVAILABLE",
	}
}

// GetMappingCapacityReportShapeAvailabilityAvailabilityStatusEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCapacityReportShapeAvailabilityAvailabilityStatusEnum(val string) (CapacityReportShapeAvailabilityAvailabilityStatusEnum, bool) {
	enum, ok := mappingCapacityReportShapeAvailabilityAvailabilityStatusEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
