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

// CreateBootVolumeDetails The representation of CreateBootVolumeDetails
type CreateBootVolumeDetails struct {

	// The OCID of the compartment that contains the boot volume.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	SourceDetails BootVolumeSourceDetails `mandatory:"true" json:"sourceDetails"`

	// The availability domain of the volume. Omissible for cloning a volume. The new volume will be created in the availability domain of the source volume.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// If provided, specifies the ID of the boot volume backup policy to assign to the newly
	// created boot volume. If omitted, no policy will be assigned.
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

	// The OCID of the Vault service key to assign as the master encryption key
	// for the boot volume.
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`

	// The size of the volume in GBs.
	SizeInGBs *int64 `mandatory:"false" json:"sizeInGBs"`

	// The clusterPlacementGroup Id of the volume for volume placement.
	ClusterPlacementGroupId *string `mandatory:"false" json:"clusterPlacementGroupId"`

	// The number of volume performance units (VPUs) that will be applied to this volume per GB,
	// representing the Block Volume service's elastic performance options.
	// See Block Volume Performance Levels (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/blockvolumeperformance.htm#perf_levels) for more information.
	// Allowed values:
	//   * `10`: Represents the Balanced option.
	//   * `20`: Represents the Higher Performance option.
	//   * `30`-`120`: Represents the Ultra High Performance option.
	// For performance autotune enabled volumes, it would be the Default(Minimum) VPUs/GB.
	VpusPerGB *int64 `mandatory:"false" json:"vpusPerGB"`

	// Specifies whether the auto-tune performance is enabled for this boot volume. This field is deprecated.
	// Use the `DetachedVolumeAutotunePolicy` instead to enable the volume for detached autotune.
	IsAutoTuneEnabled *bool `mandatory:"false" json:"isAutoTuneEnabled"`

	// The list of boot volume replicas to be enabled for this boot volume
	// in the specified destination availability domains.
	BootVolumeReplicas []BootVolumeReplicaDetails `mandatory:"false" json:"bootVolumeReplicas"`

	// The list of autotune policies to be enabled for this volume.
	AutotunePolicies []AutotunePolicy `mandatory:"false" json:"autotunePolicies"`
}

func (m CreateBootVolumeDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateBootVolumeDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *CreateBootVolumeDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		AvailabilityDomain      *string                           `json:"availabilityDomain"`
		BackupPolicyId          *string                           `json:"backupPolicyId"`
		DefinedTags             map[string]map[string]interface{} `json:"definedTags"`
		DisplayName             *string                           `json:"displayName"`
		FreeformTags            map[string]string                 `json:"freeformTags"`
		KmsKeyId                *string                           `json:"kmsKeyId"`
		SizeInGBs               *int64                            `json:"sizeInGBs"`
		ClusterPlacementGroupId *string                           `json:"clusterPlacementGroupId"`
		VpusPerGB               *int64                            `json:"vpusPerGB"`
		IsAutoTuneEnabled       *bool                             `json:"isAutoTuneEnabled"`
		BootVolumeReplicas      []BootVolumeReplicaDetails        `json:"bootVolumeReplicas"`
		AutotunePolicies        []autotunepolicy                  `json:"autotunePolicies"`
		CompartmentId           *string                           `json:"compartmentId"`
		SourceDetails           bootvolumesourcedetails           `json:"sourceDetails"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.AvailabilityDomain = model.AvailabilityDomain

	m.BackupPolicyId = model.BackupPolicyId

	m.DefinedTags = model.DefinedTags

	m.DisplayName = model.DisplayName

	m.FreeformTags = model.FreeformTags

	m.KmsKeyId = model.KmsKeyId

	m.SizeInGBs = model.SizeInGBs

	m.ClusterPlacementGroupId = model.ClusterPlacementGroupId

	m.VpusPerGB = model.VpusPerGB

	m.IsAutoTuneEnabled = model.IsAutoTuneEnabled

	m.BootVolumeReplicas = make([]BootVolumeReplicaDetails, len(model.BootVolumeReplicas))
	copy(m.BootVolumeReplicas, model.BootVolumeReplicas)
	m.AutotunePolicies = make([]AutotunePolicy, len(model.AutotunePolicies))
	for i, n := range model.AutotunePolicies {
		nn, e = n.UnmarshalPolymorphicJSON(n.JsonData)
		if e != nil {
			return e
		}
		if nn != nil {
			m.AutotunePolicies[i] = nn.(AutotunePolicy)
		} else {
			m.AutotunePolicies[i] = nil
		}
	}
	m.CompartmentId = model.CompartmentId

	nn, e = model.SourceDetails.UnmarshalPolymorphicJSON(model.SourceDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.SourceDetails = nn.(BootVolumeSourceDetails)
	} else {
		m.SourceDetails = nil
	}

	return
}
