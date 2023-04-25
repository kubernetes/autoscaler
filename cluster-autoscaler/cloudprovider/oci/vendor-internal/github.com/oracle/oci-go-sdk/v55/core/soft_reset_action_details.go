// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
//

package core

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// SoftResetActionDetails Parameters for the softReset instance action (https://docs.cloud.oracle.com/iaas/latest/Instance/InstanceAction). If omitted, default values are used.
type SoftResetActionDetails struct {

	// For instances with a date in the Maintenance reboot field, the flag denoting whether reboot migration is enabled for instances that use the DenseIO shape. The default value is 'false'.
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
