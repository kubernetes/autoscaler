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

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the source point where packets are captured.
	SourceId *string `mandatory:"false" json:"sourceId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the destination resource where mirrored packets are sent.
	TargetId *string `mandatory:"false" json:"targetId"`

	// The IP address of the destination resource where mirrored packets are sent.
	TargetIp *string `mandatory:"false" json:"targetIp"`

	// The capture filter's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
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

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet that source private endpoint belongs to.
	SourcePrivateEndpointSubnetId *string `mandatory:"false" json:"sourcePrivateEndpointSubnetId"`

	// The target type for the VTAP.
	TargetType UpdateVtapDetailsTargetTypeEnum `mandatory:"false" json:"targetType,omitempty"`

	// The source type for the VTAP.
	SourceType UpdateVtapDetailsSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`
}

func (m UpdateVtapDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateVtapDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingUpdateVtapDetailsEncapsulationProtocolEnum(string(m.EncapsulationProtocol)); !ok && m.EncapsulationProtocol != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for EncapsulationProtocol: %s. Supported values are: %s.", m.EncapsulationProtocol, strings.Join(GetUpdateVtapDetailsEncapsulationProtocolEnumStringValues(), ",")))
	}
	if _, ok := GetMappingUpdateVtapDetailsTrafficModeEnum(string(m.TrafficMode)); !ok && m.TrafficMode != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TrafficMode: %s. Supported values are: %s.", m.TrafficMode, strings.Join(GetUpdateVtapDetailsTrafficModeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingUpdateVtapDetailsTargetTypeEnum(string(m.TargetType)); !ok && m.TargetType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TargetType: %s. Supported values are: %s.", m.TargetType, strings.Join(GetUpdateVtapDetailsTargetTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingUpdateVtapDetailsSourceTypeEnum(string(m.SourceType)); !ok && m.SourceType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SourceType: %s. Supported values are: %s.", m.SourceType, strings.Join(GetUpdateVtapDetailsSourceTypeEnumStringValues(), ",")))
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

var mappingUpdateVtapDetailsEncapsulationProtocolEnumLowerCase = map[string]UpdateVtapDetailsEncapsulationProtocolEnum{
	"vxlan": UpdateVtapDetailsEncapsulationProtocolVxlan,
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

// GetMappingUpdateVtapDetailsEncapsulationProtocolEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateVtapDetailsEncapsulationProtocolEnum(val string) (UpdateVtapDetailsEncapsulationProtocolEnum, bool) {
	enum, ok := mappingUpdateVtapDetailsEncapsulationProtocolEnumLowerCase[strings.ToLower(val)]
	return enum, ok
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

var mappingUpdateVtapDetailsTrafficModeEnumLowerCase = map[string]UpdateVtapDetailsTrafficModeEnum{
	"default":  UpdateVtapDetailsTrafficModeDefault,
	"priority": UpdateVtapDetailsTrafficModePriority,
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

// GetMappingUpdateVtapDetailsTrafficModeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateVtapDetailsTrafficModeEnum(val string) (UpdateVtapDetailsTrafficModeEnum, bool) {
	enum, ok := mappingUpdateVtapDetailsTrafficModeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// UpdateVtapDetailsTargetTypeEnum Enum with underlying type: string
type UpdateVtapDetailsTargetTypeEnum string

// Set of constants representing the allowable values for UpdateVtapDetailsTargetTypeEnum
const (
	UpdateVtapDetailsTargetTypeVnic                UpdateVtapDetailsTargetTypeEnum = "VNIC"
	UpdateVtapDetailsTargetTypeNetworkLoadBalancer UpdateVtapDetailsTargetTypeEnum = "NETWORK_LOAD_BALANCER"
	UpdateVtapDetailsTargetTypeIpAddress           UpdateVtapDetailsTargetTypeEnum = "IP_ADDRESS"
)

var mappingUpdateVtapDetailsTargetTypeEnum = map[string]UpdateVtapDetailsTargetTypeEnum{
	"VNIC":                  UpdateVtapDetailsTargetTypeVnic,
	"NETWORK_LOAD_BALANCER": UpdateVtapDetailsTargetTypeNetworkLoadBalancer,
	"IP_ADDRESS":            UpdateVtapDetailsTargetTypeIpAddress,
}

var mappingUpdateVtapDetailsTargetTypeEnumLowerCase = map[string]UpdateVtapDetailsTargetTypeEnum{
	"vnic":                  UpdateVtapDetailsTargetTypeVnic,
	"network_load_balancer": UpdateVtapDetailsTargetTypeNetworkLoadBalancer,
	"ip_address":            UpdateVtapDetailsTargetTypeIpAddress,
}

// GetUpdateVtapDetailsTargetTypeEnumValues Enumerates the set of values for UpdateVtapDetailsTargetTypeEnum
func GetUpdateVtapDetailsTargetTypeEnumValues() []UpdateVtapDetailsTargetTypeEnum {
	values := make([]UpdateVtapDetailsTargetTypeEnum, 0)
	for _, v := range mappingUpdateVtapDetailsTargetTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateVtapDetailsTargetTypeEnumStringValues Enumerates the set of values in String for UpdateVtapDetailsTargetTypeEnum
func GetUpdateVtapDetailsTargetTypeEnumStringValues() []string {
	return []string{
		"VNIC",
		"NETWORK_LOAD_BALANCER",
		"IP_ADDRESS",
	}
}

// GetMappingUpdateVtapDetailsTargetTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateVtapDetailsTargetTypeEnum(val string) (UpdateVtapDetailsTargetTypeEnum, bool) {
	enum, ok := mappingUpdateVtapDetailsTargetTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// UpdateVtapDetailsSourceTypeEnum Enum with underlying type: string
type UpdateVtapDetailsSourceTypeEnum string

// Set of constants representing the allowable values for UpdateVtapDetailsSourceTypeEnum
const (
	UpdateVtapDetailsSourceTypeVnic                    UpdateVtapDetailsSourceTypeEnum = "VNIC"
	UpdateVtapDetailsSourceTypeSubnet                  UpdateVtapDetailsSourceTypeEnum = "SUBNET"
	UpdateVtapDetailsSourceTypeLoadBalancer            UpdateVtapDetailsSourceTypeEnum = "LOAD_BALANCER"
	UpdateVtapDetailsSourceTypeDbSystem                UpdateVtapDetailsSourceTypeEnum = "DB_SYSTEM"
	UpdateVtapDetailsSourceTypeExadataVmCluster        UpdateVtapDetailsSourceTypeEnum = "EXADATA_VM_CLUSTER"
	UpdateVtapDetailsSourceTypeAutonomousDataWarehouse UpdateVtapDetailsSourceTypeEnum = "AUTONOMOUS_DATA_WAREHOUSE"
)

var mappingUpdateVtapDetailsSourceTypeEnum = map[string]UpdateVtapDetailsSourceTypeEnum{
	"VNIC":                      UpdateVtapDetailsSourceTypeVnic,
	"SUBNET":                    UpdateVtapDetailsSourceTypeSubnet,
	"LOAD_BALANCER":             UpdateVtapDetailsSourceTypeLoadBalancer,
	"DB_SYSTEM":                 UpdateVtapDetailsSourceTypeDbSystem,
	"EXADATA_VM_CLUSTER":        UpdateVtapDetailsSourceTypeExadataVmCluster,
	"AUTONOMOUS_DATA_WAREHOUSE": UpdateVtapDetailsSourceTypeAutonomousDataWarehouse,
}

var mappingUpdateVtapDetailsSourceTypeEnumLowerCase = map[string]UpdateVtapDetailsSourceTypeEnum{
	"vnic":                      UpdateVtapDetailsSourceTypeVnic,
	"subnet":                    UpdateVtapDetailsSourceTypeSubnet,
	"load_balancer":             UpdateVtapDetailsSourceTypeLoadBalancer,
	"db_system":                 UpdateVtapDetailsSourceTypeDbSystem,
	"exadata_vm_cluster":        UpdateVtapDetailsSourceTypeExadataVmCluster,
	"autonomous_data_warehouse": UpdateVtapDetailsSourceTypeAutonomousDataWarehouse,
}

// GetUpdateVtapDetailsSourceTypeEnumValues Enumerates the set of values for UpdateVtapDetailsSourceTypeEnum
func GetUpdateVtapDetailsSourceTypeEnumValues() []UpdateVtapDetailsSourceTypeEnum {
	values := make([]UpdateVtapDetailsSourceTypeEnum, 0)
	for _, v := range mappingUpdateVtapDetailsSourceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateVtapDetailsSourceTypeEnumStringValues Enumerates the set of values in String for UpdateVtapDetailsSourceTypeEnum
func GetUpdateVtapDetailsSourceTypeEnumStringValues() []string {
	return []string{
		"VNIC",
		"SUBNET",
		"LOAD_BALANCER",
		"DB_SYSTEM",
		"EXADATA_VM_CLUSTER",
		"AUTONOMOUS_DATA_WAREHOUSE",
	}
}

// GetMappingUpdateVtapDetailsSourceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingUpdateVtapDetailsSourceTypeEnum(val string) (UpdateVtapDetailsSourceTypeEnum, bool) {
	enum, ok := mappingUpdateVtapDetailsSourceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
