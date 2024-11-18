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

// ComputeCapacityTopology A compute capacity topology that allows you to query your bare metal hosts and their RDMA network topology.
type ComputeCapacityTopology struct {

	// The availability domain of the compute capacity topology.
	// Example: `Uocm:US-CHICAGO-1-AD-2`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	CapacitySource CapacitySource `mandatory:"true" json:"capacitySource"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment that contains the compute capacity topology.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compute capacity topology.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the compute capacity topology.
	LifecycleState ComputeCapacityTopologyLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time that the compute capacity topology was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The date and time that the compute capacity topology was updated, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeUpdated *common.SDKTime `mandatory:"true" json:"timeUpdated"`

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
}

func (m ComputeCapacityTopology) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ComputeCapacityTopology) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingComputeCapacityTopologyLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetComputeCapacityTopologyLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *ComputeCapacityTopology) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DefinedTags        map[string]map[string]interface{}         `json:"definedTags"`
		DisplayName        *string                                   `json:"displayName"`
		FreeformTags       map[string]string                         `json:"freeformTags"`
		AvailabilityDomain *string                                   `json:"availabilityDomain"`
		CapacitySource     capacitysource                            `json:"capacitySource"`
		CompartmentId      *string                                   `json:"compartmentId"`
		Id                 *string                                   `json:"id"`
		LifecycleState     ComputeCapacityTopologyLifecycleStateEnum `json:"lifecycleState"`
		TimeCreated        *common.SDKTime                           `json:"timeCreated"`
		TimeUpdated        *common.SDKTime                           `json:"timeUpdated"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DefinedTags = model.DefinedTags

	m.DisplayName = model.DisplayName

	m.FreeformTags = model.FreeformTags

	m.AvailabilityDomain = model.AvailabilityDomain

	nn, e = model.CapacitySource.UnmarshalPolymorphicJSON(model.CapacitySource.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.CapacitySource = nn.(CapacitySource)
	} else {
		m.CapacitySource = nil
	}

	m.CompartmentId = model.CompartmentId

	m.Id = model.Id

	m.LifecycleState = model.LifecycleState

	m.TimeCreated = model.TimeCreated

	m.TimeUpdated = model.TimeUpdated

	return
}

// ComputeCapacityTopologyLifecycleStateEnum Enum with underlying type: string
type ComputeCapacityTopologyLifecycleStateEnum string

// Set of constants representing the allowable values for ComputeCapacityTopologyLifecycleStateEnum
const (
	ComputeCapacityTopologyLifecycleStateActive   ComputeCapacityTopologyLifecycleStateEnum = "ACTIVE"
	ComputeCapacityTopologyLifecycleStateCreating ComputeCapacityTopologyLifecycleStateEnum = "CREATING"
	ComputeCapacityTopologyLifecycleStateUpdating ComputeCapacityTopologyLifecycleStateEnum = "UPDATING"
	ComputeCapacityTopologyLifecycleStateDeleted  ComputeCapacityTopologyLifecycleStateEnum = "DELETED"
	ComputeCapacityTopologyLifecycleStateDeleting ComputeCapacityTopologyLifecycleStateEnum = "DELETING"
)

var mappingComputeCapacityTopologyLifecycleStateEnum = map[string]ComputeCapacityTopologyLifecycleStateEnum{
	"ACTIVE":   ComputeCapacityTopologyLifecycleStateActive,
	"CREATING": ComputeCapacityTopologyLifecycleStateCreating,
	"UPDATING": ComputeCapacityTopologyLifecycleStateUpdating,
	"DELETED":  ComputeCapacityTopologyLifecycleStateDeleted,
	"DELETING": ComputeCapacityTopologyLifecycleStateDeleting,
}

var mappingComputeCapacityTopologyLifecycleStateEnumLowerCase = map[string]ComputeCapacityTopologyLifecycleStateEnum{
	"active":   ComputeCapacityTopologyLifecycleStateActive,
	"creating": ComputeCapacityTopologyLifecycleStateCreating,
	"updating": ComputeCapacityTopologyLifecycleStateUpdating,
	"deleted":  ComputeCapacityTopologyLifecycleStateDeleted,
	"deleting": ComputeCapacityTopologyLifecycleStateDeleting,
}

// GetComputeCapacityTopologyLifecycleStateEnumValues Enumerates the set of values for ComputeCapacityTopologyLifecycleStateEnum
func GetComputeCapacityTopologyLifecycleStateEnumValues() []ComputeCapacityTopologyLifecycleStateEnum {
	values := make([]ComputeCapacityTopologyLifecycleStateEnum, 0)
	for _, v := range mappingComputeCapacityTopologyLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetComputeCapacityTopologyLifecycleStateEnumStringValues Enumerates the set of values in String for ComputeCapacityTopologyLifecycleStateEnum
func GetComputeCapacityTopologyLifecycleStateEnumStringValues() []string {
	return []string{
		"ACTIVE",
		"CREATING",
		"UPDATING",
		"DELETED",
		"DELETING",
	}
}

// GetMappingComputeCapacityTopologyLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingComputeCapacityTopologyLifecycleStateEnum(val string) (ComputeCapacityTopologyLifecycleStateEnum, bool) {
	enum, ok := mappingComputeCapacityTopologyLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
