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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// CreateVolumeBackupPolicyAssignmentDetails The representation of CreateVolumeBackupPolicyAssignmentDetails
type CreateVolumeBackupPolicyAssignmentDetails struct {

	// The OCID of the volume to assign the policy to.
	AssetId *string `mandatory:"true" json:"assetId"`

	// The OCID of the volume backup policy to assign to the volume.
	PolicyId *string `mandatory:"true" json:"policyId"`

	// The OCID of the Vault service key which is the master encryption key for the block / boot volume cross region backups, which will be used in the destination region to encrypt the backup's encryption keys.
	// For more information about the Vault service and encryption keys, see
	// Overview of Vault service (https://docs.oracle.com/iaas/Content/KeyManagement/Concepts/keyoverview.htm) and
	// Using Keys (https://docs.oracle.com/iaas/Content/KeyManagement/Tasks/usingkeys.htm).
	XrcKmsKeyId *string `mandatory:"false" json:"xrcKmsKeyId"`
}

func (m CreateVolumeBackupPolicyAssignmentDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateVolumeBackupPolicyAssignmentDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
