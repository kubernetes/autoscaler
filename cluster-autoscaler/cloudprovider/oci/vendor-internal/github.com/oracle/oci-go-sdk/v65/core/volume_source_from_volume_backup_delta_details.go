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

// VolumeSourceFromVolumeBackupDeltaDetails Specifies the volume backups (first & second) and block size in bytes.
type VolumeSourceFromVolumeBackupDeltaDetails struct {

	// The OCID of the first volume backup.
	FirstBackupId *string `mandatory:"true" json:"firstBackupId"`

	// The OCID of the second volume backup.
	SecondBackupId *string `mandatory:"true" json:"secondBackupId"`

	// Block size in bytes to be considered while performing volume restore. The value must be a power of 2; ranging from 4KB (4096 bytes) to 1MB (1048576 bytes). If omitted, defaults to 4,096 bytes (4 KiB).
	ChangeBlockSizeInBytes *int64 `mandatory:"false" json:"changeBlockSizeInBytes"`
}

func (m VolumeSourceFromVolumeBackupDeltaDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VolumeSourceFromVolumeBackupDeltaDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m VolumeSourceFromVolumeBackupDeltaDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeVolumeSourceFromVolumeBackupDeltaDetails VolumeSourceFromVolumeBackupDeltaDetails
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeVolumeSourceFromVolumeBackupDeltaDetails
	}{
		"volumeBackupDelta",
		(MarshalTypeVolumeSourceFromVolumeBackupDeltaDetails)(m),
	}

	return json.Marshal(&s)
}
