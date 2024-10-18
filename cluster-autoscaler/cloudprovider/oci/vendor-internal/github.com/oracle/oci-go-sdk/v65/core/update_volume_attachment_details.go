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

// UpdateVolumeAttachmentDetails details for updating a volume attachment.
type UpdateVolumeAttachmentDetails struct {

	// The iscsi login state of the volume attachment. For a multipath volume attachment,
	// all iscsi sessions need to be all logged-in or logged-out to be in logged-in or logged-out state.
	IscsiLoginState UpdateVolumeAttachmentDetailsIscsiLoginStateEnum `mandatory:"false" json:"iscsiLoginState,omitempty"`
}

func (m UpdateVolumeAttachmentDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateVolumeAttachmentDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingUpdateVolumeAttachmentDetailsIscsiLoginStateEnum(string(m.IscsiLoginState)); !ok && m.IscsiLoginState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for IscsiLoginState: %s. Supported values are: %s.", m.IscsiLoginState, strings.Join(GetUpdateVolumeAttachmentDetailsIscsiLoginStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
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

var mappingUpdateVolumeAttachmentDetailsIscsiLoginStateEnum = map[string]UpdateVolumeAttachmentDetailsIscsiLoginStateEnum{
	"UNKNOWN":          UpdateVolumeAttachmentDetailsIscsiLoginStateUnknown,
	"LOGGING_IN":       UpdateVolumeAttachmentDetailsIscsiLoginStateLoggingIn,
	"LOGIN_SUCCEEDED":  UpdateVolumeAttachmentDetailsIscsiLoginStateLoginSucceeded,
	"LOGIN_FAILED":     UpdateVolumeAttachmentDetailsIscsiLoginStateLoginFailed,
	"LOGGING_OUT":      UpdateVolumeAttachmentDetailsIscsiLoginStateLoggingOut,
	"LOGOUT_SUCCEEDED": UpdateVolumeAttachmentDetailsIscsiLoginStateLogoutSucceeded,
	"LOGOUT_FAILED":    UpdateVolumeAttachmentDetailsIscsiLoginStateLogoutFailed,
}

var mappingUpdateVolumeAttachmentDetailsIscsiLoginStateEnumLowerCase = map[string]UpdateVolumeAttachmentDetailsIscsiLoginStateEnum{
	"unknown":          UpdateVolumeAttachmentDetailsIscsiLoginStateUnknown,
	"logging_in":       UpdateVolumeAttachmentDetailsIscsiLoginStateLoggingIn,
	"login_succeeded":  UpdateVolumeAttachmentDetailsIscsiLoginStateLoginSucceeded,
	"login_failed":     UpdateVolumeAttachmentDetailsIscsiLoginStateLoginFailed,
	"logging_out":      UpdateVolumeAttachmentDetailsIscsiLoginStateLoggingOut,
	"logout_succeeded": UpdateVolumeAttachmentDetailsIscsiLoginStateLogoutSucceeded,
	"logout_failed":    UpdateVolumeAttachmentDetailsIscsiLoginStateLogoutFailed,
}

// GetUpdateVolumeAttachmentDetailsIscsiLoginStateEnumValues Enumerates the set of values for UpdateVolumeAttachmentDetailsIscsiLoginStateEnum
func GetUpdateVolumeAttachmentDetailsIscsiLoginStateEnumValues() []UpdateVolumeAttachmentDetailsIscsiLoginStateEnum {
	values := make([]UpdateVolumeAttachmentDetailsIscsiLoginStateEnum, 0)
	for _, v := range mappingUpdateVolumeAttachmentDetailsIscsiLoginStateEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateVolumeAttachmentDetailsIscsiLoginStateEnumStringValues Enumerates the set of values in String for UpdateVolumeAttachmentDetailsIscsiLoginStateEnum
func GetUpdateVolumeAttachmentDetailsIscsiLoginStateEnumStringValues() []string {
	return []string{
		"UNKNOWN",
		"LOGGING_IN",
		"LOGIN_SUCCEEDED",
		"LOGIN_FAILED",
		"LOGGING_OUT",
		"LOGOUT_SUCCEEDED",
		"LOGOUT_FAILED",
	}
}

// GetMappingUpdateVolumeAttachmentDetailsIscsiLoginStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateVolumeAttachmentDetailsIscsiLoginStateEnum(val string) (UpdateVolumeAttachmentDetailsIscsiLoginStateEnum, bool) {
	enum, ok := mappingUpdateVolumeAttachmentDetailsIscsiLoginStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
