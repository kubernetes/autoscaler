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

// LocalPeeringConnection Details regarding a local peering connection, which is an entity that allows two VCNs to communicate
// without traversing the Internet.
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type LocalPeeringConnection struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the local peering connection.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The local peering connection's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The local peering connection's current lifecycle state.
	LifecycleState LocalPeeringConnectionLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Indicates whether the local peering connection is peered with another local peering connection.
	PeeringStatus LocalPeeringConnectionPeeringStatusEnum `mandatory:"true" json:"peeringStatus"`

	// The date and time the local peering connection was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN the local peering connection belongs to.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// Indicates whether the peer local peering connection is contained within another tenancy.
	IsCrossTenancyPeering *bool `mandatory:"false" json:"isCrossTenancyPeering"`

	// Indicates the range of IPs available on the peer. `null` if not peered.
	PeerAdvertisedCidr *string `mandatory:"false" json:"peerAdvertisedCidr"`

	// Additional information regarding the peering status if applicable.
	PeeringStatusDetails *string `mandatory:"false" json:"peeringStatusDetails"`
}

func (m LocalPeeringConnection) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m LocalPeeringConnection) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingLocalPeeringConnectionLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetLocalPeeringConnectionLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := mappingLocalPeeringConnectionPeeringStatusEnum[string(m.PeeringStatus)]; !ok && m.PeeringStatus != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for PeeringStatus: %s. Supported values are: %s.", m.PeeringStatus, strings.Join(GetLocalPeeringConnectionPeeringStatusEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// LocalPeeringConnectionLifecycleStateEnum Enum with underlying type: string
type LocalPeeringConnectionLifecycleStateEnum string

// Set of constants representing the allowable values for LocalPeeringConnectionLifecycleStateEnum
const (
	LocalPeeringConnectionLifecycleStateProvisioning LocalPeeringConnectionLifecycleStateEnum = "PROVISIONING"
	LocalPeeringConnectionLifecycleStateAvailable    LocalPeeringConnectionLifecycleStateEnum = "AVAILABLE"
	LocalPeeringConnectionLifecycleStateTerminating  LocalPeeringConnectionLifecycleStateEnum = "TERMINATING"
	LocalPeeringConnectionLifecycleStateTerminated   LocalPeeringConnectionLifecycleStateEnum = "TERMINATED"
)

var mappingLocalPeeringConnectionLifecycleStateEnum = map[string]LocalPeeringConnectionLifecycleStateEnum{
	"PROVISIONING": LocalPeeringConnectionLifecycleStateProvisioning,
	"AVAILABLE":    LocalPeeringConnectionLifecycleStateAvailable,
	"TERMINATING":  LocalPeeringConnectionLifecycleStateTerminating,
	"TERMINATED":   LocalPeeringConnectionLifecycleStateTerminated,
}

// GetLocalPeeringConnectionLifecycleStateEnumValues Enumerates the set of values for LocalPeeringConnectionLifecycleStateEnum
func GetLocalPeeringConnectionLifecycleStateEnumValues() []LocalPeeringConnectionLifecycleStateEnum {
	values := make([]LocalPeeringConnectionLifecycleStateEnum, 0)
	for _, v := range mappingLocalPeeringConnectionLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetLocalPeeringConnectionLifecycleStateEnumStringValues Enumerates the set of values in String for LocalPeeringConnectionLifecycleStateEnum
func GetLocalPeeringConnectionLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// LocalPeeringConnectionPeeringStatusEnum Enum with underlying type: string
type LocalPeeringConnectionPeeringStatusEnum string

// Set of constants representing the allowable values for LocalPeeringConnectionPeeringStatusEnum
const (
	LocalPeeringConnectionPeeringStatusInvalid LocalPeeringConnectionPeeringStatusEnum = "INVALID"
	LocalPeeringConnectionPeeringStatusNew     LocalPeeringConnectionPeeringStatusEnum = "NEW"
	LocalPeeringConnectionPeeringStatusPeered  LocalPeeringConnectionPeeringStatusEnum = "PEERED"
	LocalPeeringConnectionPeeringStatusPending LocalPeeringConnectionPeeringStatusEnum = "PENDING"
	LocalPeeringConnectionPeeringStatusRevoked LocalPeeringConnectionPeeringStatusEnum = "REVOKED"
)

var mappingLocalPeeringConnectionPeeringStatusEnum = map[string]LocalPeeringConnectionPeeringStatusEnum{
	"INVALID": LocalPeeringConnectionPeeringStatusInvalid,
	"NEW":     LocalPeeringConnectionPeeringStatusNew,
	"PEERED":  LocalPeeringConnectionPeeringStatusPeered,
	"PENDING": LocalPeeringConnectionPeeringStatusPending,
	"REVOKED": LocalPeeringConnectionPeeringStatusRevoked,
}

// GetLocalPeeringConnectionPeeringStatusEnumValues Enumerates the set of values for LocalPeeringConnectionPeeringStatusEnum
func GetLocalPeeringConnectionPeeringStatusEnumValues() []LocalPeeringConnectionPeeringStatusEnum {
	values := make([]LocalPeeringConnectionPeeringStatusEnum, 0)
	for _, v := range mappingLocalPeeringConnectionPeeringStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetLocalPeeringConnectionPeeringStatusEnumStringValues Enumerates the set of values in String for LocalPeeringConnectionPeeringStatusEnum
func GetLocalPeeringConnectionPeeringStatusEnumStringValues() []string {
	return []string{
		"INVALID",
		"NEW",
		"PEERED",
		"PENDING",
		"REVOKED",
	}
}
