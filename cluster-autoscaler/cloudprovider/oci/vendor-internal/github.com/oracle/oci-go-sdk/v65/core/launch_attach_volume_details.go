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

// LaunchAttachVolumeDetails The details of the volume to attach.
type LaunchAttachVolumeDetails interface {

	// The device name. To retrieve a list of devices for a given instance, see ListInstanceDevices.
	GetDevice() *string

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	GetDisplayName() *string

	// Whether the attachment was created in read-only mode.
	GetIsReadOnly() *bool

	// Whether the attachment should be created in shareable mode. If an attachment
	// is created in shareable mode, then other instances can attach the same volume, provided
	// that they also create their attachments in shareable mode. Only certain volume types can
	// be attached in shareable mode. Defaults to false if not specified.
	GetIsShareable() *bool

	// The OCID of the volume. If CreateVolumeDetails is specified, this field must be omitted from the request.
	GetVolumeId() *string

	GetLaunchCreateVolumeDetails() LaunchCreateVolumeDetails
}

type launchattachvolumedetails struct {
	JsonData                  []byte
	Device                    *string                   `mandatory:"false" json:"device"`
	DisplayName               *string                   `mandatory:"false" json:"displayName"`
	IsReadOnly                *bool                     `mandatory:"false" json:"isReadOnly"`
	IsShareable               *bool                     `mandatory:"false" json:"isShareable"`
	VolumeId                  *string                   `mandatory:"false" json:"volumeId"`
	LaunchCreateVolumeDetails launchcreatevolumedetails `mandatory:"false" json:"launchCreateVolumeDetails"`
	Type                      string                    `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *launchattachvolumedetails) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerlaunchattachvolumedetails launchattachvolumedetails
	s := struct {
		Model Unmarshalerlaunchattachvolumedetails
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Device = s.Model.Device
	m.DisplayName = s.Model.DisplayName
	m.IsReadOnly = s.Model.IsReadOnly
	m.IsShareable = s.Model.IsShareable
	m.VolumeId = s.Model.VolumeId
	m.LaunchCreateVolumeDetails = s.Model.LaunchCreateVolumeDetails
	m.Type = s.Model.Type

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *launchattachvolumedetails) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "paravirtualized":
		mm := LaunchAttachParavirtualizedVolumeDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "iscsi":
		mm := LaunchAttachIScsiVolumeDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for LaunchAttachVolumeDetails: %s.", m.Type)
		return *m, nil
	}
}

// GetDevice returns Device
func (m launchattachvolumedetails) GetDevice() *string {
	return m.Device
}

// GetDisplayName returns DisplayName
func (m launchattachvolumedetails) GetDisplayName() *string {
	return m.DisplayName
}

// GetIsReadOnly returns IsReadOnly
func (m launchattachvolumedetails) GetIsReadOnly() *bool {
	return m.IsReadOnly
}

// GetIsShareable returns IsShareable
func (m launchattachvolumedetails) GetIsShareable() *bool {
	return m.IsShareable
}

// GetVolumeId returns VolumeId
func (m launchattachvolumedetails) GetVolumeId() *string {
	return m.VolumeId
}

// GetLaunchCreateVolumeDetails returns LaunchCreateVolumeDetails
func (m launchattachvolumedetails) GetLaunchCreateVolumeDetails() launchcreatevolumedetails {
	return m.LaunchCreateVolumeDetails
}

func (m launchattachvolumedetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m launchattachvolumedetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
