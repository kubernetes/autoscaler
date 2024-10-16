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

// SoftResetActionDetails Parameters for the `softReset` InstanceAction. If omitted, default values are used.
type SoftResetActionDetails struct {

	// For instances that use a DenseIO shape, the flag denoting whether
	// reboot migration (https://docs.cloud.oracle.com/iaas/Content/Compute/References/infrastructure-maintenance.htm#reboot)
	// is performed for the instance. The default value is `false`.
	// If the instance has a date in the Maintenance reboot field and you do nothing (or set this flag to `false`), the instance
	// will be rebuilt at the scheduled maintenance time. The instance will experience 2-6 hours of downtime during the
	// maintenance process. The local NVMe-based SSD will be preserved.
	// If you want to minimize downtime and can delete the SSD, you can set this flag to `true` and proactively reboot the
	// instance before the scheduled maintenance time. The instance will be reboot migrated to a healthy host and the SSD will be
	// deleted. A short downtime occurs during the migration.
	// **Caution:** When `true`, the SSD is permanently deleted. We recommend that you create a backup of the SSD before proceeding.
	AllowDenseRebootMigration *bool `mandatory:"false" json:"allowDenseRebootMigration"`
}

func (m SoftResetActionDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m SoftResetActionDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m SoftResetActionDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeSoftResetActionDetails SoftResetActionDetails
	s := struct {
		DiscriminatorParam string `json:"actionType"`
		MarshalTypeSoftResetActionDetails
	}{
		"softreset",
		(MarshalTypeSoftResetActionDetails)(m),
	}

	return json.Marshal(&s)
}
