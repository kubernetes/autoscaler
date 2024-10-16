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

// AttachBootVolumeDetails The representation of AttachBootVolumeDetails
type AttachBootVolumeDetails struct {

	// The OCID of the  boot volume.
	BootVolumeId *string `mandatory:"true" json:"bootVolumeId"`

	// The OCID of the instance.
	InstanceId *string `mandatory:"true" json:"instanceId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Refer the top-level definition of encryptionInTransitType.
	// The default value is NONE.
	EncryptionInTransitType EncryptionInTransitTypeEnum `mandatory:"false" json:"encryptionInTransitType,omitempty"`
}

func (m AttachBootVolumeDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m AttachBootVolumeDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingEncryptionInTransitTypeEnum(string(m.EncryptionInTransitType)); !ok && m.EncryptionInTransitType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for EncryptionInTransitType: %s. Supported values are: %s.", m.EncryptionInTransitType, strings.Join(GetEncryptionInTransitTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
