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

// Volume A detachable block volume device that allows you to dynamically expand
// the storage capacity of an instance. For more information, see
// Overview of Cloud Volume Storage (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
type Volume struct {

	// The availability domain of the volume.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment that contains the volume.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The OCID of the volume.
	Id *string `mandatory:"true" json:"id"`

	// The current state of a volume.
	LifecycleState VolumeLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The size of the volume in MBs. This field is deprecated. Use
	// sizeInGBs instead.
	SizeInMBs *int64 `mandatory:"true" json:"sizeInMBs"`

	// The date and time the volume was created. Format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// System tags for this resource. Each key is predefined and scoped to a namespace.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	SystemTags map[string]map[string]interface{} `mandatory:"false" json:"systemTags"`

	// Specifies whether the cloned volume's data has finished copying from the source volume or backup.
	IsHydrated *bool `mandatory:"false" json:"isHydrated"`

	// The OCID of the Vault service key which is the master encryption key for the volume.
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`

	// The number of volume performance units (VPUs) that will be applied to this volume per GB,
	// representing the Block Volume service's elastic performance options.
	// See Block Volume Performance Levels (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/blockvolumeperformance.htm#perf_levels) for more information.
	// Allowed values:
	//   * `0`: Represents Lower Cost option.
	//   * `10`: Represents Balanced option.
	//   * `20`: Represents Higher Performance option.
	//   * `30`-`120`: Represents the Ultra High Performance option.
	// For performance autotune enabled volumes, It would be the Default(Minimum) VPUs/GB.
	VpusPerGB *int64 `mandatory:"false" json:"vpusPerGB"`

	// The clusterPlacementGroup Id of the volume for volume placement.
	ClusterPlacementGroupId *string `mandatory:"false" json:"clusterPlacementGroupId"`

	// The size of the volume in GBs.
	SizeInGBs *int64 `mandatory:"false" json:"sizeInGBs"`

	SourceDetails VolumeSourceDetails `mandatory:"false" json:"sourceDetails"`

	// The OCID of the source volume group.
	VolumeGroupId *string `mandatory:"false" json:"volumeGroupId"`

	// Specifies whether the auto-tune performance is enabled for this volume. This field is deprecated.
	// Use the `DetachedVolumeAutotunePolicy` instead to enable the volume for detached autotune.
	IsAutoTuneEnabled *bool `mandatory:"false" json:"isAutoTuneEnabled"`

	// The number of Volume Performance Units per GB that this volume is effectively tuned to.
	AutoTunedVpusPerGB *int64 `mandatory:"false" json:"autoTunedVpusPerGB"`

	// The list of block volume replicas of this volume.
	BlockVolumeReplicas []BlockVolumeReplicaInfo `mandatory:"false" json:"blockVolumeReplicas"`

	// The list of autotune policies enabled for this volume.
	AutotunePolicies []AutotunePolicy `mandatory:"false" json:"autotunePolicies"`
}

func (m Volume) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m Volume) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingVolumeLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetVolumeLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *Volume) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DefinedTags             map[string]map[string]interface{} `json:"definedTags"`
		FreeformTags            map[string]string                 `json:"freeformTags"`
		SystemTags              map[string]map[string]interface{} `json:"systemTags"`
		IsHydrated              *bool                             `json:"isHydrated"`
		KmsKeyId                *string                           `json:"kmsKeyId"`
		VpusPerGB               *int64                            `json:"vpusPerGB"`
		ClusterPlacementGroupId *string                           `json:"clusterPlacementGroupId"`
		SizeInGBs               *int64                            `json:"sizeInGBs"`
		SourceDetails           volumesourcedetails               `json:"sourceDetails"`
		VolumeGroupId           *string                           `json:"volumeGroupId"`
		IsAutoTuneEnabled       *bool                             `json:"isAutoTuneEnabled"`
		AutoTunedVpusPerGB      *int64                            `json:"autoTunedVpusPerGB"`
		BlockVolumeReplicas     []BlockVolumeReplicaInfo          `json:"blockVolumeReplicas"`
		AutotunePolicies        []autotunepolicy                  `json:"autotunePolicies"`
		AvailabilityDomain      *string                           `json:"availabilityDomain"`
		CompartmentId           *string                           `json:"compartmentId"`
		DisplayName             *string                           `json:"displayName"`
		Id                      *string                           `json:"id"`
		LifecycleState          VolumeLifecycleStateEnum          `json:"lifecycleState"`
		SizeInMBs               *int64                            `json:"sizeInMBs"`
		TimeCreated             *common.SDKTime                   `json:"timeCreated"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DefinedTags = model.DefinedTags

	m.FreeformTags = model.FreeformTags

	m.SystemTags = model.SystemTags

	m.IsHydrated = model.IsHydrated

	m.KmsKeyId = model.KmsKeyId

	m.VpusPerGB = model.VpusPerGB

	m.ClusterPlacementGroupId = model.ClusterPlacementGroupId

	m.SizeInGBs = model.SizeInGBs

	nn, e = model.SourceDetails.UnmarshalPolymorphicJSON(model.SourceDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.SourceDetails = nn.(VolumeSourceDetails)
	} else {
		m.SourceDetails = nil
	}

	m.VolumeGroupId = model.VolumeGroupId

	m.IsAutoTuneEnabled = model.IsAutoTuneEnabled

	m.AutoTunedVpusPerGB = model.AutoTunedVpusPerGB

	m.BlockVolumeReplicas = make([]BlockVolumeReplicaInfo, len(model.BlockVolumeReplicas))
	copy(m.BlockVolumeReplicas, model.BlockVolumeReplicas)
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
	m.AvailabilityDomain = model.AvailabilityDomain

	m.CompartmentId = model.CompartmentId

	m.DisplayName = model.DisplayName

	m.Id = model.Id

	m.LifecycleState = model.LifecycleState

	m.SizeInMBs = model.SizeInMBs

	m.TimeCreated = model.TimeCreated

	return
}

// VolumeLifecycleStateEnum Enum with underlying type: string
type VolumeLifecycleStateEnum string

// Set of constants representing the allowable values for VolumeLifecycleStateEnum
const (
	VolumeLifecycleStateProvisioning VolumeLifecycleStateEnum = "PROVISIONING"
	VolumeLifecycleStateRestoring    VolumeLifecycleStateEnum = "RESTORING"
	VolumeLifecycleStateAvailable    VolumeLifecycleStateEnum = "AVAILABLE"
	VolumeLifecycleStateTerminating  VolumeLifecycleStateEnum = "TERMINATING"
	VolumeLifecycleStateTerminated   VolumeLifecycleStateEnum = "TERMINATED"
	VolumeLifecycleStateFaulty       VolumeLifecycleStateEnum = "FAULTY"
)

var mappingVolumeLifecycleStateEnum = map[string]VolumeLifecycleStateEnum{
	"PROVISIONING": VolumeLifecycleStateProvisioning,
	"RESTORING":    VolumeLifecycleStateRestoring,
	"AVAILABLE":    VolumeLifecycleStateAvailable,
	"TERMINATING":  VolumeLifecycleStateTerminating,
	"TERMINATED":   VolumeLifecycleStateTerminated,
	"FAULTY":       VolumeLifecycleStateFaulty,
}

var mappingVolumeLifecycleStateEnumLowerCase = map[string]VolumeLifecycleStateEnum{
	"provisioning": VolumeLifecycleStateProvisioning,
	"restoring":    VolumeLifecycleStateRestoring,
	"available":    VolumeLifecycleStateAvailable,
	"terminating":  VolumeLifecycleStateTerminating,
	"terminated":   VolumeLifecycleStateTerminated,
	"faulty":       VolumeLifecycleStateFaulty,
}

// GetVolumeLifecycleStateEnumValues Enumerates the set of values for VolumeLifecycleStateEnum
func GetVolumeLifecycleStateEnumValues() []VolumeLifecycleStateEnum {
	values := make([]VolumeLifecycleStateEnum, 0)
	for _, v := range mappingVolumeLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVolumeLifecycleStateEnumStringValues Enumerates the set of values in String for VolumeLifecycleStateEnum
func GetVolumeLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"RESTORING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
		"FAULTY",
	}
}

// GetMappingVolumeLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVolumeLifecycleStateEnum(val string) (VolumeLifecycleStateEnum, bool) {
	enum, ok := mappingVolumeLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
