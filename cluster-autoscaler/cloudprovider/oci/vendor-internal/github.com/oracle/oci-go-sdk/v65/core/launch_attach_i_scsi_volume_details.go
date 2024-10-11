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

// LaunchAttachIScsiVolumeDetails Details specific to ISCSI type volume attachments.
type LaunchAttachIScsiVolumeDetails struct {

	// The device name. To retrieve a list of devices for a given instance, see ListInstanceDevices.
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

	// The OCID of the volume. If CreateVolumeDetails is specified, this field must be omitted from the request.
	VolumeId *string `mandatory:"false" json:"volumeId"`

	LaunchCreateVolumeDetails LaunchCreateVolumeDetails `mandatory:"false" json:"launchCreateVolumeDetails"`

	// Whether to use CHAP authentication for the volume attachment. Defaults to false.
	UseChap *bool `mandatory:"false" json:"useChap"`

	// Whether to enable Oracle Cloud Agent to perform the iSCSI login and logout commands after the volume attach or detach operations for non multipath-enabled iSCSI attachments.
	IsAgentAutoIscsiLoginEnabled *bool `mandatory:"false" json:"isAgentAutoIscsiLoginEnabled"`

	// Refer the top-level definition of encryptionInTransitType.
	// The default value is NONE.
	EncryptionInTransitType EncryptionInTransitTypeEnum `mandatory:"false" json:"encryptionInTransitType,omitempty"`
}

// GetDevice returns Device
func (m LaunchAttachIScsiVolumeDetails) GetDevice() *string {
	return m.Device
}

// GetDisplayName returns DisplayName
func (m LaunchAttachIScsiVolumeDetails) GetDisplayName() *string {
	return m.DisplayName
}

// GetIsReadOnly returns IsReadOnly
func (m LaunchAttachIScsiVolumeDetails) GetIsReadOnly() *bool {
	return m.IsReadOnly
}

// GetIsShareable returns IsShareable
func (m LaunchAttachIScsiVolumeDetails) GetIsShareable() *bool {
	return m.IsShareable
}

// GetVolumeId returns VolumeId
func (m LaunchAttachIScsiVolumeDetails) GetVolumeId() *string {
	return m.VolumeId
}

// GetLaunchCreateVolumeDetails returns LaunchCreateVolumeDetails
func (m LaunchAttachIScsiVolumeDetails) GetLaunchCreateVolumeDetails() LaunchCreateVolumeDetails {
	return m.LaunchCreateVolumeDetails
}

func (m LaunchAttachIScsiVolumeDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m LaunchAttachIScsiVolumeDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingEncryptionInTransitTypeEnum(string(m.EncryptionInTransitType)); !ok && m.EncryptionInTransitType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for EncryptionInTransitType: %s. Supported values are: %s.", m.EncryptionInTransitType, strings.Join(GetEncryptionInTransitTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m LaunchAttachIScsiVolumeDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeLaunchAttachIScsiVolumeDetails LaunchAttachIScsiVolumeDetails
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeLaunchAttachIScsiVolumeDetails
	}{
		"iscsi",
		(MarshalTypeLaunchAttachIScsiVolumeDetails)(m),
	}

	return json.Marshal(&s)
}

// UnmarshalJSON unmarshals from json
func (m *LaunchAttachIScsiVolumeDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		Device                       *string                     `json:"device"`
		DisplayName                  *string                     `json:"displayName"`
		IsReadOnly                   *bool                       `json:"isReadOnly"`
		IsShareable                  *bool                       `json:"isShareable"`
		VolumeId                     *string                     `json:"volumeId"`
		LaunchCreateVolumeDetails    launchcreatevolumedetails   `json:"launchCreateVolumeDetails"`
		UseChap                      *bool                       `json:"useChap"`
		EncryptionInTransitType      EncryptionInTransitTypeEnum `json:"encryptionInTransitType"`
		IsAgentAutoIscsiLoginEnabled *bool                       `json:"isAgentAutoIscsiLoginEnabled"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.Device = model.Device

	m.DisplayName = model.DisplayName

	m.IsReadOnly = model.IsReadOnly

	m.IsShareable = model.IsShareable

	m.VolumeId = model.VolumeId

	nn, e = model.LaunchCreateVolumeDetails.UnmarshalPolymorphicJSON(model.LaunchCreateVolumeDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.LaunchCreateVolumeDetails = nn.(LaunchCreateVolumeDetails)
	} else {
		m.LaunchCreateVolumeDetails = nil
	}

	m.UseChap = model.UseChap

	m.EncryptionInTransitType = model.EncryptionInTransitType

	m.IsAgentAutoIscsiLoginEnabled = model.IsAgentAutoIscsiLoginEnabled

	return
}
