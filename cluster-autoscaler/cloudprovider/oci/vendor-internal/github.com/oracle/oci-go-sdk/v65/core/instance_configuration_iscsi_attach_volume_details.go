// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// InstanceConfigurationIscsiAttachVolumeDetails The representation of InstanceConfigurationIscsiAttachVolumeDetails
type InstanceConfigurationIscsiAttachVolumeDetails struct {

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Whether the attachment should be created in read-only mode.
	IsReadOnly *bool `mandatory:"false" json:"isReadOnly"`

	// The device name.
	Device *string `mandatory:"false" json:"device"`

	// Whether the attachment should be created in shareable mode. If an attachment
	// is created in shareable mode, then other instances can attach the same volume, provided
	// that they also create their attachments in shareable mode. Only certain volume types can
	// be attached in shareable mode. Defaults to false if not specified.
	IsShareable *bool `mandatory:"false" json:"isShareable"`

	// Whether to use CHAP authentication for the volume attachment. Defaults to false.
	UseChap *bool `mandatory:"false" json:"useChap"`
}

// GetDisplayName returns DisplayName
func (m InstanceConfigurationIscsiAttachVolumeDetails) GetDisplayName() *string {
	return m.DisplayName
}

// GetIsReadOnly returns IsReadOnly
func (m InstanceConfigurationIscsiAttachVolumeDetails) GetIsReadOnly() *bool {
	return m.IsReadOnly
}

// GetDevice returns Device
func (m InstanceConfigurationIscsiAttachVolumeDetails) GetDevice() *string {
	return m.Device
}

// GetIsShareable returns IsShareable
func (m InstanceConfigurationIscsiAttachVolumeDetails) GetIsShareable() *bool {
	return m.IsShareable
}

func (m InstanceConfigurationIscsiAttachVolumeDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstanceConfigurationIscsiAttachVolumeDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m InstanceConfigurationIscsiAttachVolumeDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeInstanceConfigurationIscsiAttachVolumeDetails InstanceConfigurationIscsiAttachVolumeDetails
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeInstanceConfigurationIscsiAttachVolumeDetails
	}{
		"iscsi",
		(MarshalTypeInstanceConfigurationIscsiAttachVolumeDetails)(m),
	}

	return json.Marshal(&s)
}
