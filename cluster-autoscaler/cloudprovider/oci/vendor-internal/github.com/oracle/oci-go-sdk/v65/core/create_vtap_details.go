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

// CreateVtapDetails These details are included in a request to create a virtual test access point (VTAP).
type CreateVtapDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the `Vtap` resource.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN containing the `Vtap` resource.
	VcnId *string `mandatory:"true" json:"vcnId"`

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

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the destination resource where mirrored packets are sent.
	TargetId *string `mandatory:"false" json:"targetId"`

	// The IP address of the destination resource where mirrored packets are sent.
	TargetIp *string `mandatory:"false" json:"targetIp"`

	// Defines an encapsulation header type for the VTAP's mirrored traffic.
	EncapsulationProtocol CreateVtapDetailsEncapsulationProtocolEnum `mandatory:"false" json:"encapsulationProtocol,omitempty"`

	// The virtual extensible LAN (VXLAN) network identifier (or VXLAN segment ID) that uniquely identifies the VXLAN.
	VxlanNetworkIdentifier *int64 `mandatory:"false" json:"vxlanNetworkIdentifier"`

	// Used to start or stop a `Vtap` resource.
	// * `TRUE` directs the VTAP to start mirroring traffic.
	// * `FALSE` (Default) directs the VTAP to stop mirroring traffic.
	IsVtapEnabled *bool `mandatory:"false" json:"isVtapEnabled"`

	// The source type for the VTAP.
	SourceType CreateVtapDetailsSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`

	// Used to control the priority of traffic. It is an optional field. If it not passed, the value is DEFAULT
	TrafficMode CreateVtapDetailsTrafficModeEnum `mandatory:"false" json:"trafficMode,omitempty"`

	// The maximum size of the packets to be included in the filter.
	MaxPacketSize *int `mandatory:"false" json:"maxPacketSize"`

	// The target type for the VTAP.
	TargetType CreateVtapDetailsTargetTypeEnum `mandatory:"false" json:"targetType,omitempty"`

	// The IP Address of the source private endpoint.
	SourcePrivateEndpointIp *string `mandatory:"false" json:"sourcePrivateEndpointIp"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet that source private endpoint belongs to.
	SourcePrivateEndpointSubnetId *string `mandatory:"false" json:"sourcePrivateEndpointSubnetId"`
}

func (m CreateVtapDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateVtapDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCreateVtapDetailsEncapsulationProtocolEnum(string(m.EncapsulationProtocol)); !ok && m.EncapsulationProtocol != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for EncapsulationProtocol: %s. Supported values are: %s.", m.EncapsulationProtocol, strings.Join(GetCreateVtapDetailsEncapsulationProtocolEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCreateVtapDetailsSourceTypeEnum(string(m.SourceType)); !ok && m.SourceType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SourceType: %s. Supported values are: %s.", m.SourceType, strings.Join(GetCreateVtapDetailsSourceTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCreateVtapDetailsTrafficModeEnum(string(m.TrafficMode)); !ok && m.TrafficMode != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TrafficMode: %s. Supported values are: %s.", m.TrafficMode, strings.Join(GetCreateVtapDetailsTrafficModeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCreateVtapDetailsTargetTypeEnum(string(m.TargetType)); !ok && m.TargetType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TargetType: %s. Supported values are: %s.", m.TargetType, strings.Join(GetCreateVtapDetailsTargetTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateVtapDetailsEncapsulationProtocolEnum Enum with underlying type: string
type CreateVtapDetailsEncapsulationProtocolEnum string

// Set of constants representing the allowable values for CreateVtapDetailsEncapsulationProtocolEnum
const (
	CreateVtapDetailsEncapsulationProtocolVxlan CreateVtapDetailsEncapsulationProtocolEnum = "VXLAN"
)

var mappingCreateVtapDetailsEncapsulationProtocolEnum = map[string]CreateVtapDetailsEncapsulationProtocolEnum{
	"VXLAN": CreateVtapDetailsEncapsulationProtocolVxlan,
}

var mappingCreateVtapDetailsEncapsulationProtocolEnumLowerCase = map[string]CreateVtapDetailsEncapsulationProtocolEnum{
	"vxlan": CreateVtapDetailsEncapsulationProtocolVxlan,
}

// GetCreateVtapDetailsEncapsulationProtocolEnumValues Enumerates the set of values for CreateVtapDetailsEncapsulationProtocolEnum
func GetCreateVtapDetailsEncapsulationProtocolEnumValues() []CreateVtapDetailsEncapsulationProtocolEnum {
	values := make([]CreateVtapDetailsEncapsulationProtocolEnum, 0)
	for _, v := range mappingCreateVtapDetailsEncapsulationProtocolEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateVtapDetailsEncapsulationProtocolEnumStringValues Enumerates the set of values in String for CreateVtapDetailsEncapsulationProtocolEnum
func GetCreateVtapDetailsEncapsulationProtocolEnumStringValues() []string {
	return []string{
		"VXLAN",
	}
}

// GetMappingCreateVtapDetailsEncapsulationProtocolEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateVtapDetailsEncapsulationProtocolEnum(val string) (CreateVtapDetailsEncapsulationProtocolEnum, bool) {
	enum, ok := mappingCreateVtapDetailsEncapsulationProtocolEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CreateVtapDetailsSourceTypeEnum Enum with underlying type: string
type CreateVtapDetailsSourceTypeEnum string

// Set of constants representing the allowable values for CreateVtapDetailsSourceTypeEnum
const (
	CreateVtapDetailsSourceTypeVnic                    CreateVtapDetailsSourceTypeEnum = "VNIC"
	CreateVtapDetailsSourceTypeSubnet                  CreateVtapDetailsSourceTypeEnum = "SUBNET"
	CreateVtapDetailsSourceTypeLoadBalancer            CreateVtapDetailsSourceTypeEnum = "LOAD_BALANCER"
	CreateVtapDetailsSourceTypeDbSystem                CreateVtapDetailsSourceTypeEnum = "DB_SYSTEM"
	CreateVtapDetailsSourceTypeExadataVmCluster        CreateVtapDetailsSourceTypeEnum = "EXADATA_VM_CLUSTER"
	CreateVtapDetailsSourceTypeAutonomousDataWarehouse CreateVtapDetailsSourceTypeEnum = "AUTONOMOUS_DATA_WAREHOUSE"
)

var mappingCreateVtapDetailsSourceTypeEnum = map[string]CreateVtapDetailsSourceTypeEnum{
	"VNIC":                      CreateVtapDetailsSourceTypeVnic,
	"SUBNET":                    CreateVtapDetailsSourceTypeSubnet,
	"LOAD_BALANCER":             CreateVtapDetailsSourceTypeLoadBalancer,
	"DB_SYSTEM":                 CreateVtapDetailsSourceTypeDbSystem,
	"EXADATA_VM_CLUSTER":        CreateVtapDetailsSourceTypeExadataVmCluster,
	"AUTONOMOUS_DATA_WAREHOUSE": CreateVtapDetailsSourceTypeAutonomousDataWarehouse,
}

var mappingCreateVtapDetailsSourceTypeEnumLowerCase = map[string]CreateVtapDetailsSourceTypeEnum{
	"vnic":                      CreateVtapDetailsSourceTypeVnic,
	"subnet":                    CreateVtapDetailsSourceTypeSubnet,
	"load_balancer":             CreateVtapDetailsSourceTypeLoadBalancer,
	"db_system":                 CreateVtapDetailsSourceTypeDbSystem,
	"exadata_vm_cluster":        CreateVtapDetailsSourceTypeExadataVmCluster,
	"autonomous_data_warehouse": CreateVtapDetailsSourceTypeAutonomousDataWarehouse,
}

// GetCreateVtapDetailsSourceTypeEnumValues Enumerates the set of values for CreateVtapDetailsSourceTypeEnum
func GetCreateVtapDetailsSourceTypeEnumValues() []CreateVtapDetailsSourceTypeEnum {
	values := make([]CreateVtapDetailsSourceTypeEnum, 0)
	for _, v := range mappingCreateVtapDetailsSourceTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateVtapDetailsSourceTypeEnumStringValues Enumerates the set of values in String for CreateVtapDetailsSourceTypeEnum
func GetCreateVtapDetailsSourceTypeEnumStringValues() []string {
	return []string{
		"VNIC",
		"SUBNET",
		"LOAD_BALANCER",
		"DB_SYSTEM",
		"EXADATA_VM_CLUSTER",
		"AUTONOMOUS_DATA_WAREHOUSE",
	}
}

// GetMappingCreateVtapDetailsSourceTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateVtapDetailsSourceTypeEnum(val string) (CreateVtapDetailsSourceTypeEnum, bool) {
	enum, ok := mappingCreateVtapDetailsSourceTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CreateVtapDetailsTrafficModeEnum Enum with underlying type: string
type CreateVtapDetailsTrafficModeEnum string

// Set of constants representing the allowable values for CreateVtapDetailsTrafficModeEnum
const (
	CreateVtapDetailsTrafficModeDefault  CreateVtapDetailsTrafficModeEnum = "DEFAULT"
	CreateVtapDetailsTrafficModePriority CreateVtapDetailsTrafficModeEnum = "PRIORITY"
)

var mappingCreateVtapDetailsTrafficModeEnum = map[string]CreateVtapDetailsTrafficModeEnum{
	"DEFAULT":  CreateVtapDetailsTrafficModeDefault,
	"PRIORITY": CreateVtapDetailsTrafficModePriority,
}

var mappingCreateVtapDetailsTrafficModeEnumLowerCase = map[string]CreateVtapDetailsTrafficModeEnum{
	"default":  CreateVtapDetailsTrafficModeDefault,
	"priority": CreateVtapDetailsTrafficModePriority,
}

// GetCreateVtapDetailsTrafficModeEnumValues Enumerates the set of values for CreateVtapDetailsTrafficModeEnum
func GetCreateVtapDetailsTrafficModeEnumValues() []CreateVtapDetailsTrafficModeEnum {
	values := make([]CreateVtapDetailsTrafficModeEnum, 0)
	for _, v := range mappingCreateVtapDetailsTrafficModeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateVtapDetailsTrafficModeEnumStringValues Enumerates the set of values in String for CreateVtapDetailsTrafficModeEnum
func GetCreateVtapDetailsTrafficModeEnumStringValues() []string {
	return []string{
		"DEFAULT",
		"PRIORITY",
	}
}

// GetMappingCreateVtapDetailsTrafficModeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateVtapDetailsTrafficModeEnum(val string) (CreateVtapDetailsTrafficModeEnum, bool) {
	enum, ok := mappingCreateVtapDetailsTrafficModeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CreateVtapDetailsTargetTypeEnum Enum with underlying type: string
type CreateVtapDetailsTargetTypeEnum string

// Set of constants representing the allowable values for CreateVtapDetailsTargetTypeEnum
const (
	CreateVtapDetailsTargetTypeVnic                CreateVtapDetailsTargetTypeEnum = "VNIC"
	CreateVtapDetailsTargetTypeNetworkLoadBalancer CreateVtapDetailsTargetTypeEnum = "NETWORK_LOAD_BALANCER"
	CreateVtapDetailsTargetTypeIpAddress           CreateVtapDetailsTargetTypeEnum = "IP_ADDRESS"
)

var mappingCreateVtapDetailsTargetTypeEnum = map[string]CreateVtapDetailsTargetTypeEnum{
	"VNIC":                  CreateVtapDetailsTargetTypeVnic,
	"NETWORK_LOAD_BALANCER": CreateVtapDetailsTargetTypeNetworkLoadBalancer,
	"IP_ADDRESS":            CreateVtapDetailsTargetTypeIpAddress,
}

var mappingCreateVtapDetailsTargetTypeEnumLowerCase = map[string]CreateVtapDetailsTargetTypeEnum{
	"vnic":                  CreateVtapDetailsTargetTypeVnic,
	"network_load_balancer": CreateVtapDetailsTargetTypeNetworkLoadBalancer,
	"ip_address":            CreateVtapDetailsTargetTypeIpAddress,
}

// GetCreateVtapDetailsTargetTypeEnumValues Enumerates the set of values for CreateVtapDetailsTargetTypeEnum
func GetCreateVtapDetailsTargetTypeEnumValues() []CreateVtapDetailsTargetTypeEnum {
	values := make([]CreateVtapDetailsTargetTypeEnum, 0)
	for _, v := range mappingCreateVtapDetailsTargetTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateVtapDetailsTargetTypeEnumStringValues Enumerates the set of values in String for CreateVtapDetailsTargetTypeEnum
func GetCreateVtapDetailsTargetTypeEnumStringValues() []string {
	return []string{
		"VNIC",
		"NETWORK_LOAD_BALANCER",
		"IP_ADDRESS",
	}
}

// GetMappingCreateVtapDetailsTargetTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCreateVtapDetailsTargetTypeEnum(val string) (CreateVtapDetailsTargetTypeEnum, bool) {
	enum, ok := mappingCreateVtapDetailsTargetTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
