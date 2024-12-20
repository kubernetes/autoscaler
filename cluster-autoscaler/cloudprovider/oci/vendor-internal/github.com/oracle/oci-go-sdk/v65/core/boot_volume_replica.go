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

// BootVolumeReplica An asynchronous replica of a boot volume that can then be used to create
// a new boot volume or recover a boot volume. For more information, see Overview
// of Cross-Region Volume Replication (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/volumereplication.htm)
// To use any of the API operations, you must be authorized in an IAM policy.
// If you're not authorized, talk to an administrator. If you're an administrator
// who needs to write policies to give users access, see Getting Started with
// Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
type BootVolumeReplica struct {

	// The availability domain of the boot volume replica.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment that contains the boot volume replica.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The boot volume replica's Oracle ID (OCID).
	Id *string `mandatory:"true" json:"id"`

	// The current state of a boot volume replica.
	LifecycleState BootVolumeReplicaLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The size of the source boot volume, in GBs.
	SizeInGBs *int64 `mandatory:"true" json:"sizeInGBs"`

	// The date and time the boot volume replica was created. Format defined
	// by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The date and time the boot volume replica was last synced from the source boot volume.
	// Format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeLastSynced *common.SDKTime `mandatory:"true" json:"timeLastSynced"`

	// The OCID of the source boot volume.
	BootVolumeId *string `mandatory:"true" json:"bootVolumeId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The image OCID used to create the boot volume the replica is replicated from.
	ImageId *string `mandatory:"false" json:"imageId"`

	// The total size of the data transferred from the source boot volume to the boot volume replica, in GBs.
	TotalDataTransferredInGBs *int64 `mandatory:"false" json:"totalDataTransferredInGBs"`

	// The OCID of the volume group replica.
	VolumeGroupReplicaId *string `mandatory:"false" json:"volumeGroupReplicaId"`
}

func (m BootVolumeReplica) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m BootVolumeReplica) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingBootVolumeReplicaLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetBootVolumeReplicaLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// BootVolumeReplicaLifecycleStateEnum Enum with underlying type: string
type BootVolumeReplicaLifecycleStateEnum string

// Set of constants representing the allowable values for BootVolumeReplicaLifecycleStateEnum
const (
	BootVolumeReplicaLifecycleStateProvisioning BootVolumeReplicaLifecycleStateEnum = "PROVISIONING"
	BootVolumeReplicaLifecycleStateAvailable    BootVolumeReplicaLifecycleStateEnum = "AVAILABLE"
	BootVolumeReplicaLifecycleStateActivating   BootVolumeReplicaLifecycleStateEnum = "ACTIVATING"
	BootVolumeReplicaLifecycleStateTerminating  BootVolumeReplicaLifecycleStateEnum = "TERMINATING"
	BootVolumeReplicaLifecycleStateTerminated   BootVolumeReplicaLifecycleStateEnum = "TERMINATED"
	BootVolumeReplicaLifecycleStateFaulty       BootVolumeReplicaLifecycleStateEnum = "FAULTY"
)

var mappingBootVolumeReplicaLifecycleStateEnum = map[string]BootVolumeReplicaLifecycleStateEnum{
	"PROVISIONING": BootVolumeReplicaLifecycleStateProvisioning,
	"AVAILABLE":    BootVolumeReplicaLifecycleStateAvailable,
	"ACTIVATING":   BootVolumeReplicaLifecycleStateActivating,
	"TERMINATING":  BootVolumeReplicaLifecycleStateTerminating,
	"TERMINATED":   BootVolumeReplicaLifecycleStateTerminated,
	"FAULTY":       BootVolumeReplicaLifecycleStateFaulty,
}

var mappingBootVolumeReplicaLifecycleStateEnumLowerCase = map[string]BootVolumeReplicaLifecycleStateEnum{
	"provisioning": BootVolumeReplicaLifecycleStateProvisioning,
	"available":    BootVolumeReplicaLifecycleStateAvailable,
	"activating":   BootVolumeReplicaLifecycleStateActivating,
	"terminating":  BootVolumeReplicaLifecycleStateTerminating,
	"terminated":   BootVolumeReplicaLifecycleStateTerminated,
	"faulty":       BootVolumeReplicaLifecycleStateFaulty,
}

// GetBootVolumeReplicaLifecycleStateEnumValues Enumerates the set of values for BootVolumeReplicaLifecycleStateEnum
func GetBootVolumeReplicaLifecycleStateEnumValues() []BootVolumeReplicaLifecycleStateEnum {
	values := make([]BootVolumeReplicaLifecycleStateEnum, 0)
	for _, v := range mappingBootVolumeReplicaLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetBootVolumeReplicaLifecycleStateEnumStringValues Enumerates the set of values in String for BootVolumeReplicaLifecycleStateEnum
func GetBootVolumeReplicaLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"ACTIVATING",
		"TERMINATING",
		"TERMINATED",
		"FAULTY",
	}
}

// GetMappingBootVolumeReplicaLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingBootVolumeReplicaLifecycleStateEnum(val string) (BootVolumeReplicaLifecycleStateEnum, bool) {
	enum, ok := mappingBootVolumeReplicaLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
