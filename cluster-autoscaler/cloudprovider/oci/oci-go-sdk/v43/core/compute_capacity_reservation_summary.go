// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// ComputeCapacityReservationSummary Summary information for a compute capacity reservation.
type ComputeCapacityReservationSummary struct {

	// The OCID of the instance reservation configuration.
	Id *string `mandatory:"true" json:"id"`

	// The availability domain of the capacity reservation.
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The date and time the capacity reservation was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// A user-friendly name for the capacity reservation. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	// Example: `My Reservation`
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The current state of the capacity reservation.
	LifecycleState ComputeCapacityReservationLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The number of instances for which capacity will be held in this
	// compute capacity reservation. This number is the sum of the values of the `reservedCount` fields
	// for all of the instance reservation configurations under this reservation.
	// The purpose of this field is to calculate the percentage usage of the reservation.
	ReservedInstanceCount *int64 `mandatory:"false" json:"reservedInstanceCount"`

	// The total number of instances currently consuming space in
	// this compute capacity reservation. This number is the sum of the values of the `usedCount` fields
	// for all of the instance reservation configurations under this reservation.
	// The purpose of this field is to calculate the percentage usage of the reservation.
	UsedInstanceCount *int64 `mandatory:"false" json:"usedInstanceCount"`

	// Whether this capacity reservation is the default.
	// For more information, see Capacity Reservations (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/reserve-capacity.htm#default).
	IsDefaultReservation *bool `mandatory:"false" json:"isDefaultReservation"`
}

func (m ComputeCapacityReservationSummary) String() string {
	return common.PointerString(m)
}
