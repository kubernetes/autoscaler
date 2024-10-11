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

// VolumeGroupReplica An asynchronous replica of a volume group that can then be used to create a new volume group
// or recover a volume group. For more information, see Volume Group Replication (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/volumegroupreplication.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
type VolumeGroupReplica struct {

	// The availability domain of the volume group replica.
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment that contains the volume group replica.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The OCID for the volume group replica.
	Id *string `mandatory:"true" json:"id"`

	// The current state of a volume group.
	LifecycleState VolumeGroupReplicaLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The aggregate size of the volume group replica in GBs.
	SizeInGBs *int64 `mandatory:"true" json:"sizeInGBs"`

	// The OCID of the source volume group.
	VolumeGroupId *string `mandatory:"true" json:"volumeGroupId"`

	// The date and time the volume group replica was created. Format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// Volume replicas within this volume group replica.
	MemberReplicas []MemberReplica `mandatory:"true" json:"memberReplicas"`

	// The date and time the volume group replica was last synced from the source volume group.
	// Format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeLastSynced *common.SDKTime `mandatory:"true" json:"timeLastSynced"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m VolumeGroupReplica) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VolumeGroupReplica) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingVolumeGroupReplicaLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetVolumeGroupReplicaLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// VolumeGroupReplicaLifecycleStateEnum Enum with underlying type: string
type VolumeGroupReplicaLifecycleStateEnum string

// Set of constants representing the allowable values for VolumeGroupReplicaLifecycleStateEnum
const (
	VolumeGroupReplicaLifecycleStateProvisioning VolumeGroupReplicaLifecycleStateEnum = "PROVISIONING"
	VolumeGroupReplicaLifecycleStateAvailable    VolumeGroupReplicaLifecycleStateEnum = "AVAILABLE"
	VolumeGroupReplicaLifecycleStateActivating   VolumeGroupReplicaLifecycleStateEnum = "ACTIVATING"
	VolumeGroupReplicaLifecycleStateTerminating  VolumeGroupReplicaLifecycleStateEnum = "TERMINATING"
	VolumeGroupReplicaLifecycleStateTerminated   VolumeGroupReplicaLifecycleStateEnum = "TERMINATED"
	VolumeGroupReplicaLifecycleStateFaulty       VolumeGroupReplicaLifecycleStateEnum = "FAULTY"
)

var mappingVolumeGroupReplicaLifecycleStateEnum = map[string]VolumeGroupReplicaLifecycleStateEnum{
	"PROVISIONING": VolumeGroupReplicaLifecycleStateProvisioning,
	"AVAILABLE":    VolumeGroupReplicaLifecycleStateAvailable,
	"ACTIVATING":   VolumeGroupReplicaLifecycleStateActivating,
	"TERMINATING":  VolumeGroupReplicaLifecycleStateTerminating,
	"TERMINATED":   VolumeGroupReplicaLifecycleStateTerminated,
	"FAULTY":       VolumeGroupReplicaLifecycleStateFaulty,
}

var mappingVolumeGroupReplicaLifecycleStateEnumLowerCase = map[string]VolumeGroupReplicaLifecycleStateEnum{
	"provisioning": VolumeGroupReplicaLifecycleStateProvisioning,
	"available":    VolumeGroupReplicaLifecycleStateAvailable,
	"activating":   VolumeGroupReplicaLifecycleStateActivating,
	"terminating":  VolumeGroupReplicaLifecycleStateTerminating,
	"terminated":   VolumeGroupReplicaLifecycleStateTerminated,
	"faulty":       VolumeGroupReplicaLifecycleStateFaulty,
}

// GetVolumeGroupReplicaLifecycleStateEnumValues Enumerates the set of values for VolumeGroupReplicaLifecycleStateEnum
func GetVolumeGroupReplicaLifecycleStateEnumValues() []VolumeGroupReplicaLifecycleStateEnum {
	values := make([]VolumeGroupReplicaLifecycleStateEnum, 0)
	for _, v := range mappingVolumeGroupReplicaLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVolumeGroupReplicaLifecycleStateEnumStringValues Enumerates the set of values in String for VolumeGroupReplicaLifecycleStateEnum
func GetVolumeGroupReplicaLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"ACTIVATING",
		"TERMINATING",
		"TERMINATED",
		"FAULTY",
	}
}

// GetMappingVolumeGroupReplicaLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVolumeGroupReplicaLifecycleStateEnum(val string) (VolumeGroupReplicaLifecycleStateEnum, bool) {
	enum, ok := mappingVolumeGroupReplicaLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
