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

// Vtap A virtual test access point (VTAP) provides a way to mirror all traffic from a designated source to a selected target in order to facilitate troubleshooting, security analysis, and data monitoring.
// A VTAP is functionally similar to a test access point (TAP) you might deploy in your on-premises network.
// A *CaptureFilter* contains a set of *CaptureFilterRuleDetails* governing what traffic a VTAP mirrors.
type Vtap struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the `Vtap` resource.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN containing the `Vtap` resource.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// The VTAP's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The VTAP's administrative lifecycle state.
	LifecycleState VtapLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the source point where packets are captured.
	SourceId *string `mandatory:"true" json:"sourceId"`

	// The capture filter's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	CaptureFilterId *string `mandatory:"true" json:"captureFilterId"`

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

	// The VTAP's current running state.
	LifecycleStateDetails VtapLifecycleStateDetailsEnum `mandatory:"false" json:"lifecycleStateDetails,omitempty"`

	// The date and time the VTAP was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2020-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the destination resource where mirrored packets are sent.
	TargetId *string `mandatory:"false" json:"targetId"`

	// The IP address of the destination resource where mirrored packets are sent.
	TargetIp *string `mandatory:"false" json:"targetIp"`

	// Defines an encapsulation header type for the VTAP's mirrored traffic.
	EncapsulationProtocol VtapEncapsulationProtocolEnum `mandatory:"false" json:"encapsulationProtocol,omitempty"`

	// The virtual extensible LAN (VXLAN) network identifier (or VXLAN segment ID) that uniquely identifies the VXLAN.
	VxlanNetworkIdentifier *int64 `mandatory:"false" json:"vxlanNetworkIdentifier"`

	// Used to start or stop a `Vtap` resource.
	// * `TRUE` directs the VTAP to start mirroring traffic.
	// * `FALSE` (Default) directs the VTAP to stop mirroring traffic.
	IsVtapEnabled *bool `mandatory:"false" json:"isVtapEnabled"`

	// The source type for the VTAP.
	SourceType VtapSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`

	// Used to control the priority of traffic. It is an optional field. If it not passed, the value is DEFAULT
	TrafficMode VtapTrafficModeEnum `mandatory:"false" json:"trafficMode,omitempty"`

	// The maximum size of the packets to be included in the filter.
	MaxPacketSize *int `mandatory:"false" json:"maxPacketSize"`

	// The target type for the VTAP.
	TargetType VtapTargetTypeEnum `mandatory:"false" json:"targetType,omitempty"`

	// The IP Address of the source private endpoint.
	SourcePrivateEndpointIp *string `mandatory:"false" json:"sourcePrivateEndpointIp"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet that source private endpoint belongs to.
	SourcePrivateEndpointSubnetId *string `mandatory:"false" json:"sourcePrivateEndpointSubnetId"`
}

func (m Vtap) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m Vtap) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingVtapLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetVtapLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := GetMappingVtapLifecycleStateDetailsEnum(string(m.LifecycleStateDetails)); !ok && m.LifecycleStateDetails != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleStateDetails: %s. Supported values are: %s.", m.LifecycleStateDetails, strings.Join(GetVtapLifecycleStateDetailsEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVtapEncapsulationProtocolEnum(string(m.EncapsulationProtocol)); !ok && m.EncapsulationProtocol != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for EncapsulationProtocol: %s. Supported values are: %s.", m.EncapsulationProtocol, strings.Join(GetVtapEncapsulationProtocolEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVtapSourceTypeEnum(string(m.SourceType)); !ok && m.SourceType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SourceType: %s. Supported values are: %s.", m.SourceType, strings.Join(GetVtapSourceTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVtapTrafficModeEnum(string(m.TrafficMode)); !ok && m.TrafficMode != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TrafficMode: %s. Supported values are: %s.", m.TrafficMode, strings.Join(GetVtapTrafficModeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVtapTargetTypeEnum(string(m.TargetType)); !ok && m.TargetType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TargetType: %s. Supported values are: %s.", m.TargetType, strings.Join(GetVtapTargetTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// VtapLifecycleStateEnum Enum with underlying type: string
type VtapLifecycleStateEnum string

// Set of constants representing the allowable values for VtapLifecycleStateEnum
const (
	VtapLifecycleStateProvisioning VtapLifecycleStateEnum = "PROVISIONING"
	VtapLifecycleStateAvailable    VtapLifecycleStateEnum = "AVAILABLE"
	VtapLifecycleStateUpdating     VtapLifecycleStateEnum = "UPDATING"
	VtapLifecycleStateTerminating  VtapLifecycleStateEnum = "TERMINATING"
	VtapLifecycleStateTerminated   VtapLifecycleStateEnum = "TERMINATED"
)

var mappingVtapLifecycleStateEnum = map[string]VtapLifecycleStateEnum{
	"PROVISIONING": VtapLifecycleStateProvisioning,
	"AVAILABLE":    VtapLifecycleStateAvailable,
	"UPDATING":     VtapLifecycleStateUpdating,
	"TERMINATING":  VtapLifecycleStateTerminating,
	"TERMINATED":   VtapLifecycleStateTerminated,
}

var mappingVtapLifecycleStateEnumLowerCase = map[string]VtapLifecycleStateEnum{
	"provisioning": VtapLifecycleStateProvisioning,
	"available":    VtapLifecycleStateAvailable,
	"updating":     VtapLifecycleStateUpdating,
	"terminating":  VtapLifecycleStateTerminating,
	"terminated":   VtapLifecycleStateTerminated,
}

// GetVtapLifecycleStateEnumValues Enumerates the set of values for VtapLifecycleStateEnum
func GetVtapLifecycleStateEnumValues() []VtapLifecycleStateEnum {
	values := make([]VtapLifecycleStateEnum, 0)
	for _, v := range mappingVtapLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVtapLifecycleStateEnumStringValues Enumerates the set of values in String for VtapLifecycleStateEnum
func GetVtapLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"UPDATING",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingVtapLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVtapLifecycleStateEnum(val string) (VtapLifecycleStateEnum, bool) {
	enum, ok := mappingVtapLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VtapLifecycleStateDetailsEnum Enum with underlying type: string
type VtapLifecycleStateDetailsEnum string

// Set of constants representing the allowable values for VtapLifecycleStateDetailsEnum
const (
	VtapLifecycleStateDetailsRunning VtapLifecycleStateDetailsEnum = "RUNNING"
	VtapLifecycleStateDetailsStopped VtapLifecycleStateDetailsEnum = "STOPPED"
)

var mappingVtapLifecycleStateDetailsEnum = map[string]VtapLifecycleStateDetailsEnum{
	"RUNNING": VtapLifecycleStateDetailsRunning,
	"STOPPED": VtapLifecycleStateDetailsStopped,
}

var mappingVtapLifecycleStateDetailsEnumLowerCase = map[string]VtapLifecycleStateDetailsEnum{
	"running": VtapLifecycleStateDetailsRunning,
	"stopped": VtapLifecycleStateDetailsStopped,
}

// GetVtapLifecycleStateDetailsEnumValues Enumerates the set of values for VtapLifecycleStateDetailsEnum
func GetVtapLifecycleStateDetailsEnumValues() []VtapLifecycleStateDetailsEnum {
	values := make([]VtapLifecycleStateDetailsEnum, 0)
	for _, v := range mappingVtapLifecycleStateDetailsEnum {
		values = append(values, v)
	}
	return values
}

// GetVtapLifecycleStateDetailsEnumStringValues Enumerates the set of values in String for VtapLifecycleStateDetailsEnum
func GetVtapLifecycleStateDetailsEnumStringValues() []string {
	return []string{
		"RUNNING",
		"STOPPED",
	}
}

// GetMappingVtapLifecycleStateDetailsEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVtapLifecycleStateDetailsEnum(val string) (VtapLifecycleStateDetailsEnum, bool) {
	enum, ok := mappingVtapLifecycleStateDetailsEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VtapEncapsulationProtocolEnum Enum with underlying type: string
type VtapEncapsulationProtocolEnum string

// Set of constants representing the allowable values for VtapEncapsulationProtocolEnum
const (
	VtapEncapsulationProtocolVxlan VtapEncapsulationProtocolEnum = "VXLAN"
)

var mappingVtapEncapsulationProtocolEnum = map[string]VtapEncapsulationProtocolEnum{
	"VXLAN": VtapEncapsulationProtocolVxlan,
}

var mappingVtapEncapsulationProtocolEnumLowerCase = map[string]VtapEncapsulationProtocolEnum{
	"vxlan": VtapEncapsulationProtocolVxlan,
}

// GetVtapEncapsulationProtocolEnumValues Enumerates the set of values for VtapEncapsulationProtocolEnum
func GetVtapEncapsulationProtocolEnumValues() []VtapEncapsulationProtocolEnum {
	values := make([]VtapEncapsulationProtocolEnum, 0)
	for _, v := range mappingVtapEncapsulationProtocolEnum {
		values = append(values, v)
	}
	return values
}

// GetVtapEncapsulationProtocolEnumStringValues Enumerates the set of values in String for VtapEncapsulationProtocolEnum
func GetVtapEncapsulationProtocolEnumStringValues() []string {
	return []string{
		"VXLAN",
	}
}

// GetMappingVtapEncapsulationProtocolEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVtapEncapsulationProtocolEnum(val string) (VtapEncapsulationProtocolEnum, bool) {
	enum, ok := mappingVtapEncapsulationProtocolEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VtapSourceTypeEnum Enum with underlying type: string
type VtapSourceTypeEnum string

// Set of constants representing the allowable values for VtapSourceTypeEnum
const (
	VtapSourceTypeVnic                    VtapSourceTypeEnum = "VNIC"
	VtapSourceTypeSubnet                  VtapSourceTypeEnum = "SUBNET"
	VtapSourceTypeLoadBalancer            VtapSourceTypeEnum = "LOAD_BALANCER"
	VtapSourceTypeDbSystem                VtapSourceTypeEnum = "DB_SYSTEM"
	VtapSourceTypeExadataVmCluster        VtapSourceTypeEnum = "EXADATA_VM_CLUSTER"
	VtapSourceTypeAutonomousDataWarehouse VtapSourceTypeEnum = "AUTONOMOUS_DATA_WAREHOUSE"
)

var mappingVtapSourceTypeEnum = map[string]VtapSourceTypeEnum{
	"VNIC":                      VtapSourceTypeVnic,
	"SUBNET":                    VtapSourceTypeSubnet,
	"LOAD_BALANCER":             VtapSourceTypeLoadBalancer,
	"DB_SYSTEM":                 VtapSourceTypeDbSystem,
	"EXADATA_VM_CLUSTER":        VtapSourceTypeExadataVmCluster,
	"AUTONOMOUS_DATA_WAREHOUSE": VtapSourceTypeAutonomousDataWarehouse,
}

var mappingVtapSourceTypeEnumLowerCase = map[string]VtapSourceTypeEnum{
	"vnic":                      VtapSourceTypeVnic,
	"subnet":                    VtapSourceTypeSubnet,
	"load_balancer":             VtapSourceTypeLoadBalancer,
	"db_system":                 VtapSourceTypeDbSystem,
	"exadata_vm_cluster":        VtapSourceTypeExadataVmCluster,
	"autonomous_data_warehouse": VtapSourceTypeAutonomousDataWarehouse,
}

// GetVtapSourceTypeEnumValues Enumerates the set of values for VtapSourceTypeEnum
func GetVtapSourceTypeEnumValues() []VtapSourceTypeEnum {
	values := make([]VtapSourceTypeEnum, 0)
	for _, v := range mappingVtapSourceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetVtapSourceTypeEnumStringValues Enumerates the set of values in String for VtapSourceTypeEnum
func GetVtapSourceTypeEnumStringValues() []string {
	return []string{
		"VNIC",
		"SUBNET",
		"LOAD_BALANCER",
		"DB_SYSTEM",
		"EXADATA_VM_CLUSTER",
		"AUTONOMOUS_DATA_WAREHOUSE",
	}
}

// GetMappingVtapSourceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVtapSourceTypeEnum(val string) (VtapSourceTypeEnum, bool) {
	enum, ok := mappingVtapSourceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VtapTrafficModeEnum Enum with underlying type: string
type VtapTrafficModeEnum string

// Set of constants representing the allowable values for VtapTrafficModeEnum
const (
	VtapTrafficModeDefault  VtapTrafficModeEnum = "DEFAULT"
	VtapTrafficModePriority VtapTrafficModeEnum = "PRIORITY"
)

var mappingVtapTrafficModeEnum = map[string]VtapTrafficModeEnum{
	"DEFAULT":  VtapTrafficModeDefault,
	"PRIORITY": VtapTrafficModePriority,
}

var mappingVtapTrafficModeEnumLowerCase = map[string]VtapTrafficModeEnum{
	"default":  VtapTrafficModeDefault,
	"priority": VtapTrafficModePriority,
}

// GetVtapTrafficModeEnumValues Enumerates the set of values for VtapTrafficModeEnum
func GetVtapTrafficModeEnumValues() []VtapTrafficModeEnum {
	values := make([]VtapTrafficModeEnum, 0)
	for _, v := range mappingVtapTrafficModeEnum {
		values = append(values, v)
	}
	return values
}

// GetVtapTrafficModeEnumStringValues Enumerates the set of values in String for VtapTrafficModeEnum
func GetVtapTrafficModeEnumStringValues() []string {
	return []string{
		"DEFAULT",
		"PRIORITY",
	}
}

// GetMappingVtapTrafficModeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVtapTrafficModeEnum(val string) (VtapTrafficModeEnum, bool) {
	enum, ok := mappingVtapTrafficModeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VtapTargetTypeEnum Enum with underlying type: string
type VtapTargetTypeEnum string

// Set of constants representing the allowable values for VtapTargetTypeEnum
const (
	VtapTargetTypeVnic                VtapTargetTypeEnum = "VNIC"
	VtapTargetTypeNetworkLoadBalancer VtapTargetTypeEnum = "NETWORK_LOAD_BALANCER"
	VtapTargetTypeIpAddress           VtapTargetTypeEnum = "IP_ADDRESS"
)

var mappingVtapTargetTypeEnum = map[string]VtapTargetTypeEnum{
	"VNIC":                  VtapTargetTypeVnic,
	"NETWORK_LOAD_BALANCER": VtapTargetTypeNetworkLoadBalancer,
	"IP_ADDRESS":            VtapTargetTypeIpAddress,
}

var mappingVtapTargetTypeEnumLowerCase = map[string]VtapTargetTypeEnum{
	"vnic":                  VtapTargetTypeVnic,
	"network_load_balancer": VtapTargetTypeNetworkLoadBalancer,
	"ip_address":            VtapTargetTypeIpAddress,
}

// GetVtapTargetTypeEnumValues Enumerates the set of values for VtapTargetTypeEnum
func GetVtapTargetTypeEnumValues() []VtapTargetTypeEnum {
	values := make([]VtapTargetTypeEnum, 0)
	for _, v := range mappingVtapTargetTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetVtapTargetTypeEnumStringValues Enumerates the set of values in String for VtapTargetTypeEnum
func GetVtapTargetTypeEnumStringValues() []string {
	return []string{
		"VNIC",
		"NETWORK_LOAD_BALANCER",
		"IP_ADDRESS",
	}
}

// GetMappingVtapTargetTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVtapTargetTypeEnum(val string) (VtapTargetTypeEnum, bool) {
	enum, ok := mappingVtapTargetTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
