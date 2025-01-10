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

// Instance A compute host. The image used to launch the instance determines its operating system and other
// software. The shape specified during the launch process determines the number of CPUs and memory
// allocated to the instance.
// When you launch an instance, it is automatically attached to a virtual
// network interface card (VNIC), called the *primary VNIC*. The VNIC
// has a private IP address from the subnet's CIDR. You can either assign a
// private IP address of your choice or let Oracle automatically assign one.
// You can choose whether the instance has a public IP address. To retrieve the
// addresses, use the ListVnicAttachments
// operation to get the VNIC ID for the instance, and then call
// GetVnic with the VNIC ID.
// For more information, see
// Overview of the Compute Service (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
type Instance struct {

	// The availability domain the instance is running in.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment that contains the instance.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the instance.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the instance.
	LifecycleState InstanceLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The region that contains the availability domain the instance is running in.
	// For the us-phoenix-1 and us-ashburn-1 regions, `phx` and `iad` are returned, respectively.
	// For all other regions, the full region name is returned.
	// Examples: `phx`, `eu-frankfurt-1`
	Region *string `mandatory:"true" json:"region"`

	// The shape of the instance. The shape determines the number of CPUs and the amount of memory
	// allocated to the instance. You can enumerate all available shapes by calling
	// ListShapes.
	Shape *string `mandatory:"true" json:"shape"`

	// The date and time the instance was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID of the compute capacity reservation this instance is launched under.
	// When this field contains an empty string or is null, the instance is not currently in a capacity reservation.
	// For more information, see Capacity Reservations (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/reserve-capacity.htm#default).
	CapacityReservationId *string `mandatory:"false" json:"capacityReservationId"`

	// The OCID of the cluster placement group of the instance.
	ClusterPlacementGroupId *string `mandatory:"false" json:"clusterPlacementGroupId"`

	// The OCID of the dedicated virtual machine host that the instance is placed on.
	DedicatedVmHostId *string `mandatory:"false" json:"dedicatedVmHostId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Security Attributes for this resource. This is unique to ZPR, and helps identify which resources are allowed to be accessed by what permission controls.
	// Example: `{"Oracle-DataSecurity-ZPR": {"MaxEgressCount": {"value":"42","mode":"audit"}}}`
	SecurityAttributes map[string]map[string]interface{} `mandatory:"false" json:"securityAttributes"`

	// The lifecycle state of the `securityAttributes`
	SecurityAttributesState InstanceSecurityAttributesStateEnum `mandatory:"false" json:"securityAttributesState,omitempty"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Additional metadata key/value pairs that you provide. They serve the same purpose and functionality
	// as fields in the `metadata` object.
	// They are distinguished from `metadata` fields in that these can be nested JSON objects (whereas `metadata`
	// fields are string/string maps only).
	ExtendedMetadata map[string]interface{} `mandatory:"false" json:"extendedMetadata"`

	// The name of the fault domain the instance is running in.
	// A fault domain is a grouping of hardware and infrastructure within an availability domain.
	// Each availability domain contains three fault domains. Fault domains let you distribute your
	// instances so that they are not on the same physical hardware within a single availability domain.
	// A hardware failure or Compute hardware maintenance that affects one fault domain does not affect
	// instances in other fault domains.
	// If you do not specify the fault domain, the system selects one for you.
	// Example: `FAULT-DOMAIN-1`
	FaultDomain *string `mandatory:"false" json:"faultDomain"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Deprecated. Use `sourceDetails` instead.
	ImageId *string `mandatory:"false" json:"imageId"`

	// When a bare metal or virtual machine
	// instance boots, the iPXE firmware that runs on the instance is
	// configured to run an iPXE script to continue the boot process.
	// If you want more control over the boot process, you can provide
	// your own custom iPXE script that will run when the instance boots.
	// Be aware that the same iPXE script will run
	// every time an instance boots, not only after the initial
	// LaunchInstance call.
	// The default iPXE script connects to the instance's local boot
	// volume over iSCSI and performs a network boot. If you use a custom iPXE
	// script and want to network-boot from the instance's local boot volume
	// over iSCSI the same way as the default iPXE script, use the
	// following iSCSI IP address: 169.254.0.2, and boot volume IQN:
	// iqn.2015-02.oracle.boot.
	// If your instance boot volume attachment type is paravirtualized,
	// the boot volume is attached to the instance through virtio-scsi and no iPXE script is used.
	// If your instance boot volume attachment type is paravirtualized
	// and you use custom iPXE to network boot into your instance,
	// the primary boot volume is attached as a data volume through virtio-scsi drive.
	// For more information about the Bring Your Own Image feature of
	// Oracle Cloud Infrastructure, see
	// Bring Your Own Image (https://docs.cloud.oracle.com/iaas/Content/Compute/References/bringyourownimage.htm).
	// For more information about iPXE, see http://ipxe.org.
	IpxeScript *string `mandatory:"false" json:"ipxeScript"`

	// Specifies the configuration mode for launching virtual machine (VM) instances. The configuration modes are:
	// * `NATIVE` - VM instances launch with iSCSI boot and VFIO devices. The default value for platform images.
	// * `EMULATED` - VM instances launch with emulated devices, such as the E1000 network driver and emulated SCSI disk controller.
	// * `PARAVIRTUALIZED` - VM instances launch with paravirtualized devices using VirtIO drivers.
	// * `CUSTOM` - VM instances launch with custom configuration settings specified in the `LaunchOptions` parameter.
	LaunchMode InstanceLaunchModeEnum `mandatory:"false" json:"launchMode,omitempty"`

	LaunchOptions *LaunchOptions `mandatory:"false" json:"launchOptions"`

	InstanceOptions *InstanceOptions `mandatory:"false" json:"instanceOptions"`

	AvailabilityConfig *InstanceAvailabilityConfig `mandatory:"false" json:"availabilityConfig"`

	PreemptibleInstanceConfig *PreemptibleInstanceConfigDetails `mandatory:"false" json:"preemptibleInstanceConfig"`

	// Custom metadata that you provide.
	Metadata map[string]string `mandatory:"false" json:"metadata"`

	ShapeConfig *InstanceShapeConfig `mandatory:"false" json:"shapeConfig"`

	// Whether the instance’s OCPUs and memory are distributed across multiple NUMA nodes.
	IsCrossNumaNode *bool `mandatory:"false" json:"isCrossNumaNode"`

	SourceDetails InstanceSourceDetails `mandatory:"false" json:"sourceDetails"`

	// System tags for this resource. Each key is predefined and scoped to a namespace.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	SystemTags map[string]map[string]interface{} `mandatory:"false" json:"systemTags"`

	AgentConfig *InstanceAgentConfig `mandatory:"false" json:"agentConfig"`

	// The date and time the instance is expected to be stopped / started,  in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// After that time if instance hasn't been rebooted, Oracle will reboot the instance within 24 hours of the due time.
	// Regardless of how the instance was stopped, the flag will be reset to empty as soon as instance reaches Stopped state.
	// Example: `2018-05-25T21:10:29.600Z`
	TimeMaintenanceRebootDue *common.SDKTime `mandatory:"false" json:"timeMaintenanceRebootDue"`

	PlatformConfig PlatformConfig `mandatory:"false" json:"platformConfig"`

	// The OCID of the Instance Configuration used to source launch details for this instance. Any other fields supplied in the instance launch request override the details stored in the Instance Configuration for this instance launch.
	InstanceConfigurationId *string `mandatory:"false" json:"instanceConfigurationId"`
}

func (m Instance) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m Instance) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingInstanceLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInstanceLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := GetMappingInstanceSecurityAttributesStateEnum(string(m.SecurityAttributesState)); !ok && m.SecurityAttributesState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for SecurityAttributesState: %s. Supported values are: %s.", m.SecurityAttributesState, strings.Join(GetInstanceSecurityAttributesStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingInstanceLaunchModeEnum(string(m.LaunchMode)); !ok && m.LaunchMode != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LaunchMode: %s. Supported values are: %s.", m.LaunchMode, strings.Join(GetInstanceLaunchModeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *Instance) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		CapacityReservationId     *string                             `json:"capacityReservationId"`
		ClusterPlacementGroupId   *string                             `json:"clusterPlacementGroupId"`
		DedicatedVmHostId         *string                             `json:"dedicatedVmHostId"`
		DefinedTags               map[string]map[string]interface{}   `json:"definedTags"`
		SecurityAttributes        map[string]map[string]interface{}   `json:"securityAttributes"`
		SecurityAttributesState   InstanceSecurityAttributesStateEnum `json:"securityAttributesState"`
		DisplayName               *string                             `json:"displayName"`
		ExtendedMetadata          map[string]interface{}              `json:"extendedMetadata"`
		FaultDomain               *string                             `json:"faultDomain"`
		FreeformTags              map[string]string                   `json:"freeformTags"`
		ImageId                   *string                             `json:"imageId"`
		IpxeScript                *string                             `json:"ipxeScript"`
		LaunchMode                InstanceLaunchModeEnum              `json:"launchMode"`
		LaunchOptions             *LaunchOptions                      `json:"launchOptions"`
		InstanceOptions           *InstanceOptions                    `json:"instanceOptions"`
		AvailabilityConfig        *InstanceAvailabilityConfig         `json:"availabilityConfig"`
		PreemptibleInstanceConfig *PreemptibleInstanceConfigDetails   `json:"preemptibleInstanceConfig"`
		Metadata                  map[string]string                   `json:"metadata"`
		ShapeConfig               *InstanceShapeConfig                `json:"shapeConfig"`
		IsCrossNumaNode           *bool                               `json:"isCrossNumaNode"`
		SourceDetails             instancesourcedetails               `json:"sourceDetails"`
		SystemTags                map[string]map[string]interface{}   `json:"systemTags"`
		AgentConfig               *InstanceAgentConfig                `json:"agentConfig"`
		TimeMaintenanceRebootDue  *common.SDKTime                     `json:"timeMaintenanceRebootDue"`
		PlatformConfig            platformconfig                      `json:"platformConfig"`
		InstanceConfigurationId   *string                             `json:"instanceConfigurationId"`
		AvailabilityDomain        *string                             `json:"availabilityDomain"`
		CompartmentId             *string                             `json:"compartmentId"`
		Id                        *string                             `json:"id"`
		LifecycleState            InstanceLifecycleStateEnum          `json:"lifecycleState"`
		Region                    *string                             `json:"region"`
		Shape                     *string                             `json:"shape"`
		TimeCreated               *common.SDKTime                     `json:"timeCreated"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.CapacityReservationId = model.CapacityReservationId

	m.ClusterPlacementGroupId = model.ClusterPlacementGroupId

	m.DedicatedVmHostId = model.DedicatedVmHostId

	m.DefinedTags = model.DefinedTags

	m.SecurityAttributes = model.SecurityAttributes

	m.SecurityAttributesState = model.SecurityAttributesState

	m.DisplayName = model.DisplayName

	m.ExtendedMetadata = model.ExtendedMetadata

	m.FaultDomain = model.FaultDomain

	m.FreeformTags = model.FreeformTags

	m.ImageId = model.ImageId

	m.IpxeScript = model.IpxeScript

	m.LaunchMode = model.LaunchMode

	m.LaunchOptions = model.LaunchOptions

	m.InstanceOptions = model.InstanceOptions

	m.AvailabilityConfig = model.AvailabilityConfig

	m.PreemptibleInstanceConfig = model.PreemptibleInstanceConfig

	m.Metadata = model.Metadata

	m.ShapeConfig = model.ShapeConfig

	m.IsCrossNumaNode = model.IsCrossNumaNode

	nn, e = model.SourceDetails.UnmarshalPolymorphicJSON(model.SourceDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.SourceDetails = nn.(InstanceSourceDetails)
	} else {
		m.SourceDetails = nil
	}

	m.SystemTags = model.SystemTags

	m.AgentConfig = model.AgentConfig

	m.TimeMaintenanceRebootDue = model.TimeMaintenanceRebootDue

	nn, e = model.PlatformConfig.UnmarshalPolymorphicJSON(model.PlatformConfig.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.PlatformConfig = nn.(PlatformConfig)
	} else {
		m.PlatformConfig = nil
	}

	m.InstanceConfigurationId = model.InstanceConfigurationId

	m.AvailabilityDomain = model.AvailabilityDomain

	m.CompartmentId = model.CompartmentId

	m.Id = model.Id

	m.LifecycleState = model.LifecycleState

	m.Region = model.Region

	m.Shape = model.Shape

	m.TimeCreated = model.TimeCreated

	return
}

// InstanceSecurityAttributesStateEnum Enum with underlying type: string
type InstanceSecurityAttributesStateEnum string

// Set of constants representing the allowable values for InstanceSecurityAttributesStateEnum
const (
	InstanceSecurityAttributesStateStable   InstanceSecurityAttributesStateEnum = "STABLE"
	InstanceSecurityAttributesStateUpdating InstanceSecurityAttributesStateEnum = "UPDATING"
)

var mappingInstanceSecurityAttributesStateEnum = map[string]InstanceSecurityAttributesStateEnum{
	"STABLE":   InstanceSecurityAttributesStateStable,
	"UPDATING": InstanceSecurityAttributesStateUpdating,
}

var mappingInstanceSecurityAttributesStateEnumLowerCase = map[string]InstanceSecurityAttributesStateEnum{
	"stable":   InstanceSecurityAttributesStateStable,
	"updating": InstanceSecurityAttributesStateUpdating,
}

// GetInstanceSecurityAttributesStateEnumValues Enumerates the set of values for InstanceSecurityAttributesStateEnum
func GetInstanceSecurityAttributesStateEnumValues() []InstanceSecurityAttributesStateEnum {
	values := make([]InstanceSecurityAttributesStateEnum, 0)
	for _, v := range mappingInstanceSecurityAttributesStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceSecurityAttributesStateEnumStringValues Enumerates the set of values in String for InstanceSecurityAttributesStateEnum
func GetInstanceSecurityAttributesStateEnumStringValues() []string {
	return []string{
		"STABLE",
		"UPDATING",
	}
}

// GetMappingInstanceSecurityAttributesStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceSecurityAttributesStateEnum(val string) (InstanceSecurityAttributesStateEnum, bool) {
	enum, ok := mappingInstanceSecurityAttributesStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// InstanceLaunchModeEnum Enum with underlying type: string
type InstanceLaunchModeEnum string

// Set of constants representing the allowable values for InstanceLaunchModeEnum
const (
	InstanceLaunchModeNative          InstanceLaunchModeEnum = "NATIVE"
	InstanceLaunchModeEmulated        InstanceLaunchModeEnum = "EMULATED"
	InstanceLaunchModeParavirtualized InstanceLaunchModeEnum = "PARAVIRTUALIZED"
	InstanceLaunchModeCustom          InstanceLaunchModeEnum = "CUSTOM"
)

var mappingInstanceLaunchModeEnum = map[string]InstanceLaunchModeEnum{
	"NATIVE":          InstanceLaunchModeNative,
	"EMULATED":        InstanceLaunchModeEmulated,
	"PARAVIRTUALIZED": InstanceLaunchModeParavirtualized,
	"CUSTOM":          InstanceLaunchModeCustom,
}

var mappingInstanceLaunchModeEnumLowerCase = map[string]InstanceLaunchModeEnum{
	"native":          InstanceLaunchModeNative,
	"emulated":        InstanceLaunchModeEmulated,
	"paravirtualized": InstanceLaunchModeParavirtualized,
	"custom":          InstanceLaunchModeCustom,
}

// GetInstanceLaunchModeEnumValues Enumerates the set of values for InstanceLaunchModeEnum
func GetInstanceLaunchModeEnumValues() []InstanceLaunchModeEnum {
	values := make([]InstanceLaunchModeEnum, 0)
	for _, v := range mappingInstanceLaunchModeEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceLaunchModeEnumStringValues Enumerates the set of values in String for InstanceLaunchModeEnum
func GetInstanceLaunchModeEnumStringValues() []string {
	return []string{
		"NATIVE",
		"EMULATED",
		"PARAVIRTUALIZED",
		"CUSTOM",
	}
}

// GetMappingInstanceLaunchModeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceLaunchModeEnum(val string) (InstanceLaunchModeEnum, bool) {
	enum, ok := mappingInstanceLaunchModeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// InstanceLifecycleStateEnum Enum with underlying type: string
type InstanceLifecycleStateEnum string

// Set of constants representing the allowable values for InstanceLifecycleStateEnum
const (
	InstanceLifecycleStateMoving        InstanceLifecycleStateEnum = "MOVING"
	InstanceLifecycleStateProvisioning  InstanceLifecycleStateEnum = "PROVISIONING"
	InstanceLifecycleStateRunning       InstanceLifecycleStateEnum = "RUNNING"
	InstanceLifecycleStateStarting      InstanceLifecycleStateEnum = "STARTING"
	InstanceLifecycleStateStopping      InstanceLifecycleStateEnum = "STOPPING"
	InstanceLifecycleStateStopped       InstanceLifecycleStateEnum = "STOPPED"
	InstanceLifecycleStateCreatingImage InstanceLifecycleStateEnum = "CREATING_IMAGE"
	InstanceLifecycleStateTerminating   InstanceLifecycleStateEnum = "TERMINATING"
	InstanceLifecycleStateTerminated    InstanceLifecycleStateEnum = "TERMINATED"
)

var mappingInstanceLifecycleStateEnum = map[string]InstanceLifecycleStateEnum{
	"MOVING":         InstanceLifecycleStateMoving,
	"PROVISIONING":   InstanceLifecycleStateProvisioning,
	"RUNNING":        InstanceLifecycleStateRunning,
	"STARTING":       InstanceLifecycleStateStarting,
	"STOPPING":       InstanceLifecycleStateStopping,
	"STOPPED":        InstanceLifecycleStateStopped,
	"CREATING_IMAGE": InstanceLifecycleStateCreatingImage,
	"TERMINATING":    InstanceLifecycleStateTerminating,
	"TERMINATED":     InstanceLifecycleStateTerminated,
}

var mappingInstanceLifecycleStateEnumLowerCase = map[string]InstanceLifecycleStateEnum{
	"moving":         InstanceLifecycleStateMoving,
	"provisioning":   InstanceLifecycleStateProvisioning,
	"running":        InstanceLifecycleStateRunning,
	"starting":       InstanceLifecycleStateStarting,
	"stopping":       InstanceLifecycleStateStopping,
	"stopped":        InstanceLifecycleStateStopped,
	"creating_image": InstanceLifecycleStateCreatingImage,
	"terminating":    InstanceLifecycleStateTerminating,
	"terminated":     InstanceLifecycleStateTerminated,
}

// GetInstanceLifecycleStateEnumValues Enumerates the set of values for InstanceLifecycleStateEnum
func GetInstanceLifecycleStateEnumValues() []InstanceLifecycleStateEnum {
	values := make([]InstanceLifecycleStateEnum, 0)
	for _, v := range mappingInstanceLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceLifecycleStateEnumStringValues Enumerates the set of values in String for InstanceLifecycleStateEnum
func GetInstanceLifecycleStateEnumStringValues() []string {
	return []string{
		"MOVING",
		"PROVISIONING",
		"RUNNING",
		"STARTING",
		"STOPPING",
		"STOPPED",
		"CREATING_IMAGE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingInstanceLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceLifecycleStateEnum(val string) (InstanceLifecycleStateEnum, bool) {
	enum, ok := mappingInstanceLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
