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
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// ParavirtualizedVolumeAttachment A paravirtualized volume attachment.
type ParavirtualizedVolumeAttachment struct {

	// The availability domain of an instance.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the volume attachment.
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the instance the volume is attached to.
	InstanceId *string `mandatory:"true" json:"instanceId"`

	// The date and time the volume was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID of the volume.
	VolumeId *string `mandatory:"true" json:"volumeId"`

	// The device name.
	Device *string `mandatory:"false" json:"device"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Whether the attachment was created in read-only mode.
	IsReadOnly *bool `mandatory:"false" json:"isReadOnly"`

	// Whether the attachment should be created in shareable mode. If an attachment
	// is created in shareable mode, then other instances can attach the same volume, provided
	// that they also create their attachments in shareable mode. Only certain volume types can
	// be attached in shareable mode. Defaults to false if not specified.
	IsShareable *bool `mandatory:"false" json:"isShareable"`

	// Whether in-transit encryption for the data volume's paravirtualized attachment is enabled or not.
	IsPvEncryptionInTransitEnabled *bool `mandatory:"false" json:"isPvEncryptionInTransitEnabled"`

	// Whether the Iscsi or Paravirtualized attachment is multipath or not, it is not applicable to NVMe attachment.
	IsMultipath *bool `mandatory:"false" json:"isMultipath"`

	// Flag indicating if this volume was created for the customer as part of a simplified launch.
	// Used to determine whether the volume requires deletion on instance termination.
	IsVolumeCreatedDuringLaunch *bool `mandatory:"false" json:"isVolumeCreatedDuringLaunch"`

	// The current state of the volume attachment.
	LifecycleState VolumeAttachmentLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The iscsi login state of the volume attachment. For a Iscsi volume attachment,
	// all iscsi sessions need to be all logged-in or logged-out to be in logged-in or logged-out state.
	IscsiLoginState VolumeAttachmentIscsiLoginStateEnum `mandatory:"false" json:"iscsiLoginState,omitempty"`
}

// GetAvailabilityDomain returns AvailabilityDomain
func (m ParavirtualizedVolumeAttachment) GetAvailabilityDomain() *string {
	return m.AvailabilityDomain
}

// GetCompartmentId returns CompartmentId
func (m ParavirtualizedVolumeAttachment) GetCompartmentId() *string {
	return m.CompartmentId
}

// GetDevice returns Device
func (m ParavirtualizedVolumeAttachment) GetDevice() *string {
	return m.Device
}

// GetDisplayName returns DisplayName
func (m ParavirtualizedVolumeAttachment) GetDisplayName() *string {
	return m.DisplayName
}

// GetId returns Id
func (m ParavirtualizedVolumeAttachment) GetId() *string {
	return m.Id
}

// GetInstanceId returns InstanceId
func (m ParavirtualizedVolumeAttachment) GetInstanceId() *string {
	return m.InstanceId
}

// GetIsReadOnly returns IsReadOnly
func (m ParavirtualizedVolumeAttachment) GetIsReadOnly() *bool {
	return m.IsReadOnly
}

// GetIsShareable returns IsShareable
func (m ParavirtualizedVolumeAttachment) GetIsShareable() *bool {
	return m.IsShareable
}

// GetLifecycleState returns LifecycleState
func (m ParavirtualizedVolumeAttachment) GetLifecycleState() VolumeAttachmentLifecycleStateEnum {
	return m.LifecycleState
}

// GetTimeCreated returns TimeCreated
func (m ParavirtualizedVolumeAttachment) GetTimeCreated() *common.SDKTime {
	return m.TimeCreated
}

// GetVolumeId returns VolumeId
func (m ParavirtualizedVolumeAttachment) GetVolumeId() *string {
	return m.VolumeId
}

// GetIsPvEncryptionInTransitEnabled returns IsPvEncryptionInTransitEnabled
func (m ParavirtualizedVolumeAttachment) GetIsPvEncryptionInTransitEnabled() *bool {
	return m.IsPvEncryptionInTransitEnabled
}

// GetIsMultipath returns IsMultipath
func (m ParavirtualizedVolumeAttachment) GetIsMultipath() *bool {
	return m.IsMultipath
}

// GetIscsiLoginState returns IscsiLoginState
func (m ParavirtualizedVolumeAttachment) GetIscsiLoginState() VolumeAttachmentIscsiLoginStateEnum {
	return m.IscsiLoginState
}

// GetIsVolumeCreatedDuringLaunch returns IsVolumeCreatedDuringLaunch
func (m ParavirtualizedVolumeAttachment) GetIsVolumeCreatedDuringLaunch() *bool {
	return m.IsVolumeCreatedDuringLaunch
}

func (m ParavirtualizedVolumeAttachment) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ParavirtualizedVolumeAttachment) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingVolumeAttachmentLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetVolumeAttachmentLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVolumeAttachmentIscsiLoginStateEnum(string(m.IscsiLoginState)); !ok && m.IscsiLoginState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for IscsiLoginState: %s. Supported values are: %s.", m.IscsiLoginState, strings.Join(GetVolumeAttachmentIscsiLoginStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m ParavirtualizedVolumeAttachment) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeParavirtualizedVolumeAttachment ParavirtualizedVolumeAttachment
	s := struct {
		DiscriminatorParam string `json:"attachmentType"`
		MarshalTypeParavirtualizedVolumeAttachment
	}{
		"paravirtualized",
		(MarshalTypeParavirtualizedVolumeAttachment)(m),
	}

	return json.Marshal(&s)
}
