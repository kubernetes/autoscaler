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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// UpdateVtapDetails These details can be included in a request to update a virtual test access point (VTAP).
type UpdateVtapDetails struct {

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

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the source point where packets are captured.
	SourceId *string `mandatory:"false" json:"sourceId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the destination resource where mirrored packets are sent.
	TargetId *string `mandatory:"false" json:"targetId"`

	// The IP address of the destination resource where mirrored packets are sent.
	TargetIp *string `mandatory:"false" json:"targetIp"`

	// The capture filter's Oracle ID (OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm)).
	CaptureFilterId *string `mandatory:"false" json:"captureFilterId"`

	// Defines an encapsulation header type for the VTAP's mirrored traffic.
	EncapsulationProtocol UpdateVtapDetailsEncapsulationProtocolEnum `mandatory:"false" json:"encapsulationProtocol,omitempty"`

	// The virtual extensible LAN (VXLAN) network identifier (or VXLAN segment ID) that uniquely identifies the VXLAN.
	VxlanNetworkIdentifier *int64 `mandatory:"false" json:"vxlanNetworkIdentifier"`

	// Used to start or stop a `Vtap` resource.
	// * `TRUE` directs the VTAP to start mirroring traffic.
	// * `FALSE` (Default) directs the VTAP to stop mirroring traffic.
	IsVtapEnabled *bool `mandatory:"false" json:"isVtapEnabled"`

	// Used to control the priority of traffic. It is an optional field. If it not passed, the value is DEFAULT
	TrafficMode UpdateVtapDetailsTrafficModeEnum `mandatory:"false" json:"trafficMode,omitempty"`

	// The maximum size of the packets to be included in the filter.
	MaxPacketSize *int `mandatory:"false" json:"maxPacketSize"`

	// The IP Address of the source private endpoint.
	SourcePrivateEndpointIp *string `mandatory:"false" json:"sourcePrivateEndpointIp"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the subnet that source private endpoint belongs to.
	SourcePrivateEndpointSubnetId *string `mandatory:"false" json:"sourcePrivateEndpointSubnetId"`
}

func (m UpdateVtapDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateVtapDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingUpdateVtapDetailsEncapsulationProtocolEnum[string(m.EncapsulationProtocol)]; !ok && m.EncapsulationProtocol != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for EncapsulationProtocol: %s. Supported values are: %s.", m.EncapsulationProtocol, strings.Join(GetUpdateVtapDetailsEncapsulationProtocolEnumStringValues(), ",")))
	}
	if _, ok := mappingUpdateVtapDetailsTrafficModeEnum[string(m.TrafficMode)]; !ok && m.TrafficMode != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TrafficMode: %s. Supported values are: %s.", m.TrafficMode, strings.Join(GetUpdateVtapDetailsTrafficModeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UpdateVtapDetailsEncapsulationProtocolEnum Enum with underlying type: string
type UpdateVtapDetailsEncapsulationProtocolEnum string

// Set of constants representing the allowable values for UpdateVtapDetailsEncapsulationProtocolEnum
const (
	UpdateVtapDetailsEncapsulationProtocolVxlan UpdateVtapDetailsEncapsulationProtocolEnum = "VXLAN"
)

var mappingUpdateVtapDetailsEncapsulationProtocolEnum = map[string]UpdateVtapDetailsEncapsulationProtocolEnum{
	"VXLAN": UpdateVtapDetailsEncapsulationProtocolVxlan,
}

// GetUpdateVtapDetailsEncapsulationProtocolEnumValues Enumerates the set of values for UpdateVtapDetailsEncapsulationProtocolEnum
func GetUpdateVtapDetailsEncapsulationProtocolEnumValues() []UpdateVtapDetailsEncapsulationProtocolEnum {
	values := make([]UpdateVtapDetailsEncapsulationProtocolEnum, 0)
	for _, v := range mappingUpdateVtapDetailsEncapsulationProtocolEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateVtapDetailsEncapsulationProtocolEnumStringValues Enumerates the set of values in String for UpdateVtapDetailsEncapsulationProtocolEnum
func GetUpdateVtapDetailsEncapsulationProtocolEnumStringValues() []string {
	return []string{
		"VXLAN",
	}
}

// UpdateVtapDetailsTrafficModeEnum Enum with underlying type: string
type UpdateVtapDetailsTrafficModeEnum string

// Set of constants representing the allowable values for UpdateVtapDetailsTrafficModeEnum
const (
	UpdateVtapDetailsTrafficModeDefault  UpdateVtapDetailsTrafficModeEnum = "DEFAULT"
	UpdateVtapDetailsTrafficModePriority UpdateVtapDetailsTrafficModeEnum = "PRIORITY"
)

var mappingUpdateVtapDetailsTrafficModeEnum = map[string]UpdateVtapDetailsTrafficModeEnum{
	"DEFAULT":  UpdateVtapDetailsTrafficModeDefault,
	"PRIORITY": UpdateVtapDetailsTrafficModePriority,
}

// GetUpdateVtapDetailsTrafficModeEnumValues Enumerates the set of values for UpdateVtapDetailsTrafficModeEnum
func GetUpdateVtapDetailsTrafficModeEnumValues() []UpdateVtapDetailsTrafficModeEnum {
	values := make([]UpdateVtapDetailsTrafficModeEnum, 0)
	for _, v := range mappingUpdateVtapDetailsTrafficModeEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateVtapDetailsTrafficModeEnumStringValues Enumerates the set of values in String for UpdateVtapDetailsTrafficModeEnum
func GetUpdateVtapDetailsTrafficModeEnumStringValues() []string {
	return []string{
		"DEFAULT",
		"PRIORITY",
	}
}
