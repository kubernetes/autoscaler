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

// UpdateBootVolumeDetails The representation of UpdateBootVolumeDetails
type UpdateBootVolumeDetails struct {

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

	// The size to resize the volume to in GBs. Has to be larger than the current size.
	SizeInGBs *int64 `mandatory:"false" json:"sizeInGBs"`

	// The number of volume performance units (VPUs) that will be applied to this volume per GB,
	// representing the Block Volume service's elastic performance options.
	// See Block Volume Performance Levels (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/blockvolumeperformance.htm#perf_levels) for more information.
	// Allowed values:
	//   * `10`: Represents Balanced option.
	//   * `20`: Represents Higher Performance option.
	//   * `30`-`120`: Represents the Ultra High Performance option.
	// For performance autotune enabled volumes, it would be the Default(Minimum) VPUs/GB.
	VpusPerGB *int64 `mandatory:"false" json:"vpusPerGB"`

	// Specifies whether the auto-tune performance is enabled for this boot volume. This field is deprecated.
	// Use the `DetachedVolumeAutotunePolicy` instead to enable the volume for detached autotune.
	IsAutoTuneEnabled *bool `mandatory:"false" json:"isAutoTuneEnabled"`

	// The list of boot volume replicas that this boot volume will be updated to have
	// in the specified destination availability domains.
	BootVolumeReplicas []BootVolumeReplicaDetails `mandatory:"false" json:"bootVolumeReplicas"`

	// The list of autotune policies to be enabled for this volume.
	AutotunePolicies []AutotunePolicy `mandatory:"false" json:"autotunePolicies"`
}

func (m UpdateBootVolumeDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateBootVolumeDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *UpdateBootVolumeDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DefinedTags        map[string]map[string]interface{} `json:"definedTags"`
		DisplayName        *string                           `json:"displayName"`
		FreeformTags       map[string]string                 `json:"freeformTags"`
		SizeInGBs          *int64                            `json:"sizeInGBs"`
		VpusPerGB          *int64                            `json:"vpusPerGB"`
		IsAutoTuneEnabled  *bool                             `json:"isAutoTuneEnabled"`
		BootVolumeReplicas []BootVolumeReplicaDetails        `json:"bootVolumeReplicas"`
		AutotunePolicies   []autotunepolicy                  `json:"autotunePolicies"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DefinedTags = model.DefinedTags

	m.DisplayName = model.DisplayName

	m.FreeformTags = model.FreeformTags

	m.SizeInGBs = model.SizeInGBs

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
	return
}
