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

// RebootMigrateActionDetails Parameters for the `rebootMigrate` InstanceAction.
type RebootMigrateActionDetails struct {

	// For bare metal instances that have local storage, this must be set to true to verify that the local storage
	// will be deleted during the migration.  For instances without, this parameter has no effect.
	DeleteLocalStorage *bool `mandatory:"false" json:"deleteLocalStorage"`

	// If present, this parameter will set (or reset) the scheduled time that the instance will be reboot
	// migrated in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).  This will also change
	// the `timeMaintenanceRebootDue` field on the instance.
	// If not present, the reboot migration will be triggered immediately.
	TimeScheduled *common.SDKTime `mandatory:"false" json:"timeScheduled"`
}

func (m RebootMigrateActionDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m RebootMigrateActionDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m RebootMigrateActionDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeRebootMigrateActionDetails RebootMigrateActionDetails
	s := struct {
		DiscriminatorParam string `json:"actionType"`
		MarshalTypeRebootMigrateActionDetails
	}{
		"rebootMigrate",
		(MarshalTypeRebootMigrateActionDetails)(m),
	}

	return json.Marshal(&s)
}
