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

// CreateVolumeGroupDetails The representation of CreateVolumeGroupDetails
type CreateVolumeGroupDetails struct {

	// The availability domain of the volume group.
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment that contains the volume group.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	SourceDetails VolumeGroupSourceDetails `mandatory:"true" json:"sourceDetails"`

	// If provided, specifies the ID of the volume backup policy to assign to the newly
	// created volume group. If omitted, no policy will be assigned.
	BackupPolicyId *string `mandatory:"false" json:"backupPolicyId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The list of volume group replicas that this volume group will be enabled to have
	// in the specified destination availability domains.
	VolumeGroupReplicas []VolumeGroupReplicaDetails `mandatory:"false" json:"volumeGroupReplicas"`

	// The clusterPlacementGroup Id of the volume group for volume group placement.
	ClusterPlacementGroupId *string `mandatory:"false" json:"clusterPlacementGroupId"`
}

func (m CreateVolumeGroupDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateVolumeGroupDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *CreateVolumeGroupDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		BackupPolicyId          *string                           `json:"backupPolicyId"`
		DefinedTags             map[string]map[string]interface{} `json:"definedTags"`
		DisplayName             *string                           `json:"displayName"`
		FreeformTags            map[string]string                 `json:"freeformTags"`
		VolumeGroupReplicas     []VolumeGroupReplicaDetails       `json:"volumeGroupReplicas"`
		ClusterPlacementGroupId *string                           `json:"clusterPlacementGroupId"`
		AvailabilityDomain      *string                           `json:"availabilityDomain"`
		CompartmentId           *string                           `json:"compartmentId"`
		SourceDetails           volumegroupsourcedetails          `json:"sourceDetails"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.BackupPolicyId = model.BackupPolicyId

	m.DefinedTags = model.DefinedTags

	m.DisplayName = model.DisplayName

	m.FreeformTags = model.FreeformTags

	m.VolumeGroupReplicas = make([]VolumeGroupReplicaDetails, len(model.VolumeGroupReplicas))
	copy(m.VolumeGroupReplicas, model.VolumeGroupReplicas)
	m.ClusterPlacementGroupId = model.ClusterPlacementGroupId

	m.AvailabilityDomain = model.AvailabilityDomain

	m.CompartmentId = model.CompartmentId

	nn, e = model.SourceDetails.UnmarshalPolymorphicJSON(model.SourceDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.SourceDetails = nn.(VolumeGroupSourceDetails)
	} else {
		m.SourceDetails = nil
	}

	return
}
