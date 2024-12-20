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

// UpdateInstanceSourceViaBootVolumeDetails The details for updating the instance source from an existing boot volume.
type UpdateInstanceSourceViaBootVolumeDetails struct {

	// The OCID of the boot volume used to boot the instance.
	BootVolumeId *string `mandatory:"true" json:"bootVolumeId"`

	// Whether to preserve the boot volume that was previously attached to the instance after a successful replacement of that boot volume.
	IsPreserveBootVolumeEnabled *bool `mandatory:"false" json:"isPreserveBootVolumeEnabled"`
}

// GetIsPreserveBootVolumeEnabled returns IsPreserveBootVolumeEnabled
func (m UpdateInstanceSourceViaBootVolumeDetails) GetIsPreserveBootVolumeEnabled() *bool {
	return m.IsPreserveBootVolumeEnabled
}

func (m UpdateInstanceSourceViaBootVolumeDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateInstanceSourceViaBootVolumeDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m UpdateInstanceSourceViaBootVolumeDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeUpdateInstanceSourceViaBootVolumeDetails UpdateInstanceSourceViaBootVolumeDetails
	s := struct {
		DiscriminatorParam string `json:"sourceType"`
		MarshalTypeUpdateInstanceSourceViaBootVolumeDetails
	}{
		"bootVolume",
		(MarshalTypeUpdateInstanceSourceViaBootVolumeDetails)(m),
	}

	return json.Marshal(&s)
}
