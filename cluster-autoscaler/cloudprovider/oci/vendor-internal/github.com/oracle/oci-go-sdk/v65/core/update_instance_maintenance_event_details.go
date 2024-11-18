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

// UpdateInstanceMaintenanceEventDetails Specifies the properties for updating maintenance due date.
type UpdateInstanceMaintenanceEventDetails struct {

	// The beginning of the time window when Maintenance is scheduled to begin. The Maintenance will not begin before
	// this time.
	// The timeWindowEnd is automatically calculated based on the maintenanceReason and the instanceAction.
	TimeWindowStart *common.SDKTime `mandatory:"false" json:"timeWindowStart"`

	// One of the alternativeResolutionActions that was provided in the InstanceMaintenanceEvent.
	AlternativeResolutionAction InstanceMaintenanceAlternativeResolutionActionsEnum `mandatory:"false" json:"alternativeResolutionAction,omitempty"`

	// This field is only applicable when setting the alternativeResolutionAction.
	// For Instances that have local storage, this must be set to true to verify that the local storage
	// will be deleted during the migration. For instances without, this parameter has no effect.
	// In cases where the local storage will be lost, this parameter must be set or the request will fail.
	CanDeleteLocalStorage *bool `mandatory:"false" json:"canDeleteLocalStorage"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m UpdateInstanceMaintenanceEventDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateInstanceMaintenanceEventDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingInstanceMaintenanceAlternativeResolutionActionsEnum(string(m.AlternativeResolutionAction)); !ok && m.AlternativeResolutionAction != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AlternativeResolutionAction: %s. Supported values are: %s.", m.AlternativeResolutionAction, strings.Join(GetInstanceMaintenanceAlternativeResolutionActionsEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
