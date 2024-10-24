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

// CapacityReservationInstanceSummary Condensed instance data when listing instances in a compute capacity reservation.
type CapacityReservationInstanceSummary struct {

	// The OCID of the instance.
	Id *string `mandatory:"true" json:"id"`

	// The availability domain the instance is running in.
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment that contains the instance.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The shape of the instance. The shape determines the number of CPUs, amount of memory,
	// and other resources allocated to the instance.
	// You can enumerate all available shapes by calling ListComputeCapacityReservationInstanceShapes.
	Shape *string `mandatory:"true" json:"shape"`

	// The fault domain the instance is running in.
	FaultDomain *string `mandatory:"false" json:"faultDomain"`

	// The OCID of the cluster placement group of the instance.
	ClusterPlacementGroupId *string `mandatory:"false" json:"clusterPlacementGroupId"`

	ShapeConfig *InstanceReservationShapeConfigDetails `mandatory:"false" json:"shapeConfig"`
}

func (m CapacityReservationInstanceSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CapacityReservationInstanceSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
