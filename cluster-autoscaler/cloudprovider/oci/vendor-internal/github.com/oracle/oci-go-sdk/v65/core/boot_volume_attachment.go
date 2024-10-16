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

// BootVolumeAttachment Represents an attachment between a boot volume and an instance.
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
type BootVolumeAttachment struct {

	// The availability domain of an instance.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the boot volume.
	BootVolumeId *string `mandatory:"true" json:"bootVolumeId"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the boot volume attachment.
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the instance the boot volume is attached to.
	InstanceId *string `mandatory:"true" json:"instanceId"`

	// The current state of the boot volume attachment.
	LifecycleState BootVolumeAttachmentLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the boot volume was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The date and time the boot volume attachment was updated, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// Whether in-transit encryption for the boot volume's paravirtualized attachment is enabled or not.
	IsPvEncryptionInTransitEnabled *bool `mandatory:"false" json:"isPvEncryptionInTransitEnabled"`

	// Refer the top-level definition of encryptionInTransitType.
	// The default value is NONE.
	EncryptionInTransitType EncryptionInTransitTypeEnum `mandatory:"false" json:"encryptionInTransitType,omitempty"`
}

func (m BootVolumeAttachment) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m BootVolumeAttachment) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingBootVolumeAttachmentLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetBootVolumeAttachmentLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := GetMappingEncryptionInTransitTypeEnum(string(m.EncryptionInTransitType)); !ok && m.EncryptionInTransitType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for EncryptionInTransitType: %s. Supported values are: %s.", m.EncryptionInTransitType, strings.Join(GetEncryptionInTransitTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// BootVolumeAttachmentLifecycleStateEnum Enum with underlying type: string
type BootVolumeAttachmentLifecycleStateEnum string

// Set of constants representing the allowable values for BootVolumeAttachmentLifecycleStateEnum
const (
	BootVolumeAttachmentLifecycleStateAttaching BootVolumeAttachmentLifecycleStateEnum = "ATTACHING"
	BootVolumeAttachmentLifecycleStateAttached  BootVolumeAttachmentLifecycleStateEnum = "ATTACHED"
	BootVolumeAttachmentLifecycleStateDetaching BootVolumeAttachmentLifecycleStateEnum = "DETACHING"
	BootVolumeAttachmentLifecycleStateDetached  BootVolumeAttachmentLifecycleStateEnum = "DETACHED"
)

var mappingBootVolumeAttachmentLifecycleStateEnum = map[string]BootVolumeAttachmentLifecycleStateEnum{
	"ATTACHING": BootVolumeAttachmentLifecycleStateAttaching,
	"ATTACHED":  BootVolumeAttachmentLifecycleStateAttached,
	"DETACHING": BootVolumeAttachmentLifecycleStateDetaching,
	"DETACHED":  BootVolumeAttachmentLifecycleStateDetached,
}

var mappingBootVolumeAttachmentLifecycleStateEnumLowerCase = map[string]BootVolumeAttachmentLifecycleStateEnum{
	"attaching": BootVolumeAttachmentLifecycleStateAttaching,
	"attached":  BootVolumeAttachmentLifecycleStateAttached,
	"detaching": BootVolumeAttachmentLifecycleStateDetaching,
	"detached":  BootVolumeAttachmentLifecycleStateDetached,
}

// GetBootVolumeAttachmentLifecycleStateEnumValues Enumerates the set of values for BootVolumeAttachmentLifecycleStateEnum
func GetBootVolumeAttachmentLifecycleStateEnumValues() []BootVolumeAttachmentLifecycleStateEnum {
	values := make([]BootVolumeAttachmentLifecycleStateEnum, 0)
	for _, v := range mappingBootVolumeAttachmentLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetBootVolumeAttachmentLifecycleStateEnumStringValues Enumerates the set of values in String for BootVolumeAttachmentLifecycleStateEnum
func GetBootVolumeAttachmentLifecycleStateEnumStringValues() []string {
	return []string{
		"ATTACHING",
		"ATTACHED",
		"DETACHING",
		"DETACHED",
	}
}

// GetMappingBootVolumeAttachmentLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingBootVolumeAttachmentLifecycleStateEnum(val string) (BootVolumeAttachmentLifecycleStateEnum, bool) {
	enum, ok := mappingBootVolumeAttachmentLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
