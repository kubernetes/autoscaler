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

// InstanceReservationConfigDetails A template that contains the settings to use when defining the instance reservation configuration.
type InstanceReservationConfigDetails struct {

	// The shape requested when launching instances using reserved capacity.
	// The shape determines the number of CPUs, amount of memory,
	// and other resources allocated to the instance.
	// You can list all available shapes by calling ListComputeCapacityReservationInstanceShapes.
	InstanceShape *string `mandatory:"true" json:"instanceShape"`

	// The amount of capacity to reserve in this reservation configuration.
	ReservedCount *int64 `mandatory:"true" json:"reservedCount"`

	InstanceShapeConfig *InstanceReservationShapeConfigDetails `mandatory:"false" json:"instanceShapeConfig"`

	// The fault domain to use for instances created using this reservation configuration.
	// For more information, see Fault Domains (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/regions.htm#fault).
	// If you do not specify the fault domain, the capacity is available for an instance
	// that does not specify a fault domain. To change the fault domain for a reservation,
	// delete the reservation and create a new one in the preferred fault domain.
	// To retrieve a list of fault domains, use the `ListFaultDomains` operation in
	// the Identity and Access Management Service API (https://docs.cloud.oracle.com/iaas/api/#/en/identity/20160918/).
	// Example: `FAULT-DOMAIN-1`
	FaultDomain *string `mandatory:"false" json:"faultDomain"`
}

func (m InstanceReservationConfigDetails) String() string {
	return common.PointerString(m)
}
