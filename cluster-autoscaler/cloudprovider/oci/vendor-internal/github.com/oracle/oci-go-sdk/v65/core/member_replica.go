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

// MemberReplica OCIDs for the volume replicas in this volume group replica.
type MemberReplica struct {

	// The volume replica ID.
	VolumeReplicaId *string `mandatory:"true" json:"volumeReplicaId"`

	// Membership state of the volume replica in relation to the volume group replica.
	MembershipState MemberReplicaMembershipStateEnum `mandatory:"false" json:"membershipState,omitempty"`
}

func (m MemberReplica) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m MemberReplica) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingMemberReplicaMembershipStateEnum(string(m.MembershipState)); !ok && m.MembershipState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for MembershipState: %s. Supported values are: %s.", m.MembershipState, strings.Join(GetMemberReplicaMembershipStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MemberReplicaMembershipStateEnum Enum with underlying type: string
type MemberReplicaMembershipStateEnum string

// Set of constants representing the allowable values for MemberReplicaMembershipStateEnum
const (
	MemberReplicaMembershipStateAddPending    MemberReplicaMembershipStateEnum = "ADD_PENDING"
	MemberReplicaMembershipStateStable        MemberReplicaMembershipStateEnum = "STABLE"
	MemberReplicaMembershipStateRemovePending MemberReplicaMembershipStateEnum = "REMOVE_PENDING"
)

var mappingMemberReplicaMembershipStateEnum = map[string]MemberReplicaMembershipStateEnum{
	"ADD_PENDING":    MemberReplicaMembershipStateAddPending,
	"STABLE":         MemberReplicaMembershipStateStable,
	"REMOVE_PENDING": MemberReplicaMembershipStateRemovePending,
}

var mappingMemberReplicaMembershipStateEnumLowerCase = map[string]MemberReplicaMembershipStateEnum{
	"add_pending":    MemberReplicaMembershipStateAddPending,
	"stable":         MemberReplicaMembershipStateStable,
	"remove_pending": MemberReplicaMembershipStateRemovePending,
}

// GetMemberReplicaMembershipStateEnumValues Enumerates the set of values for MemberReplicaMembershipStateEnum
func GetMemberReplicaMembershipStateEnumValues() []MemberReplicaMembershipStateEnum {
	values := make([]MemberReplicaMembershipStateEnum, 0)
	for _, v := range mappingMemberReplicaMembershipStateEnum {
		values = append(values, v)
	}
	return values
}

// GetMemberReplicaMembershipStateEnumStringValues Enumerates the set of values in String for MemberReplicaMembershipStateEnum
func GetMemberReplicaMembershipStateEnumStringValues() []string {
	return []string{
		"ADD_PENDING",
		"STABLE",
		"REMOVE_PENDING",
	}
}

// GetMappingMemberReplicaMembershipStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingMemberReplicaMembershipStateEnum(val string) (MemberReplicaMembershipStateEnum, bool) {
	enum, ok := mappingMemberReplicaMembershipStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
