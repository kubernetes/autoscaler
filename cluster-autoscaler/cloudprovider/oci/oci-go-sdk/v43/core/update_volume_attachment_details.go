// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// UpdateVolumeAttachmentDetails details for updating a volume attachment.
type UpdateVolumeAttachmentDetails struct {

	// The iscsi login state of the volume attachment. For a multipath volume attachment,
	// all iscsi sessions need to be all logged-in or logged-out to be in logged-in or logged-out state.
	IscsiLoginState UpdateVolumeAttachmentDetailsIscsiLoginStateEnum `mandatory:"false" json:"iscsiLoginState,omitempty"`
}

func (m UpdateVolumeAttachmentDetails) String() string {
	return common.PointerString(m)
}

// UpdateVolumeAttachmentDetailsIscsiLoginStateEnum Enum with underlying type: string
type UpdateVolumeAttachmentDetailsIscsiLoginStateEnum string

// Set of constants representing the allowable values for UpdateVolumeAttachmentDetailsIscsiLoginStateEnum
const (
	UpdateVolumeAttachmentDetailsIscsiLoginStateUnknown         UpdateVolumeAttachmentDetailsIscsiLoginStateEnum = "UNKNOWN"
	UpdateVolumeAttachmentDetailsIscsiLoginStateLoggingIn       UpdateVolumeAttachmentDetailsIscsiLoginStateEnum = "LOGGING_IN"
	UpdateVolumeAttachmentDetailsIscsiLoginStateLoginSucceeded  UpdateVolumeAttachmentDetailsIscsiLoginStateEnum = "LOGIN_SUCCEEDED"
	UpdateVolumeAttachmentDetailsIscsiLoginStateLoginFailed     UpdateVolumeAttachmentDetailsIscsiLoginStateEnum = "LOGIN_FAILED"
	UpdateVolumeAttachmentDetailsIscsiLoginStateLoggingOut      UpdateVolumeAttachmentDetailsIscsiLoginStateEnum = "LOGGING_OUT"
	UpdateVolumeAttachmentDetailsIscsiLoginStateLogoutSucceeded UpdateVolumeAttachmentDetailsIscsiLoginStateEnum = "LOGOUT_SUCCEEDED"
	UpdateVolumeAttachmentDetailsIscsiLoginStateLogoutFailed    UpdateVolumeAttachmentDetailsIscsiLoginStateEnum = "LOGOUT_FAILED"
)

var mappingUpdateVolumeAttachmentDetailsIscsiLoginState = map[string]UpdateVolumeAttachmentDetailsIscsiLoginStateEnum{
	"UNKNOWN":          UpdateVolumeAttachmentDetailsIscsiLoginStateUnknown,
	"LOGGING_IN":       UpdateVolumeAttachmentDetailsIscsiLoginStateLoggingIn,
	"LOGIN_SUCCEEDED":  UpdateVolumeAttachmentDetailsIscsiLoginStateLoginSucceeded,
	"LOGIN_FAILED":     UpdateVolumeAttachmentDetailsIscsiLoginStateLoginFailed,
	"LOGGING_OUT":      UpdateVolumeAttachmentDetailsIscsiLoginStateLoggingOut,
	"LOGOUT_SUCCEEDED": UpdateVolumeAttachmentDetailsIscsiLoginStateLogoutSucceeded,
	"LOGOUT_FAILED":    UpdateVolumeAttachmentDetailsIscsiLoginStateLogoutFailed,
}

// GetUpdateVolumeAttachmentDetailsIscsiLoginStateEnumValues Enumerates the set of values for UpdateVolumeAttachmentDetailsIscsiLoginStateEnum
func GetUpdateVolumeAttachmentDetailsIscsiLoginStateEnumValues() []UpdateVolumeAttachmentDetailsIscsiLoginStateEnum {
	values := make([]UpdateVolumeAttachmentDetailsIscsiLoginStateEnum, 0)
	for _, v := range mappingUpdateVolumeAttachmentDetailsIscsiLoginState {
		values = append(values, v)
	}
	return values
}
