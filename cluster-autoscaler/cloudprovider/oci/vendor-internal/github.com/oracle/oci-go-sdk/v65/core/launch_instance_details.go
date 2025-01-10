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

// LaunchInstanceDetails Instance launch details.
// Use the `sourceDetails` parameter to specify whether a boot volume or an image should be used to launch a new instance.
type LaunchInstanceDetails struct {

	// The availability domain of the instance.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the compute capacity reservation this instance is launched under.
	// You can opt out of all default reservations by specifying an empty string as input for this field.
	// For more information, see Capacity Reservations (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/reserve-capacity.htm#default).
	CapacityReservationId *string `mandatory:"false" json:"capacityReservationId"`

	CreateVnicDetails *CreateVnicDetails `mandatory:"false" json:"createVnicDetails"`

	// The OCID of the dedicated virtual machine host to place the instance on.
	DedicatedVmHostId *string `mandatory:"false" json:"dedicatedVmHostId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Security Attributes for this resource. This is unique to ZPR, and helps identify which resources are allowed to be accessed by what permission controls.
	// Example: `{"Oracle-DataSecurity-ZPR": {"MaxEgressCount": {"value":"42","mode":"audit"}}}`
	SecurityAttributes map[string]map[string]interface{} `mandatory:"false" json:"securityAttributes"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Additional metadata key/value pairs that you provide. They serve the same purpose and
	// functionality as fields in the `metadata` object.
	// They are distinguished from `metadata` fields in that these can be nested JSON objects
	// (whereas `metadata` fields are string/string maps only).
	// The combined size of the `metadata` and `extendedMetadata` objects can be a maximum of
	// 32,000 bytes.
	ExtendedMetadata map[string]interface{} `mandatory:"false" json:"extendedMetadata"`

	// A fault domain is a grouping of hardware and infrastructure within an availability domain.
	// Each availability domain contains three fault domains. Fault domains let you distribute your
	// instances so that they are not on the same physical hardware within a single availability domain.
	// A hardware failure or Compute hardware maintenance that affects one fault domain does not affect
	// instances in other fault domains.
	// If you do not specify the fault domain, the system selects one for you.
	//
	// To get a list of fault domains, use the
	// ListFaultDomains operation in the
	// Identity and Access Management Service API.
	// Example: `FAULT-DOMAIN-1`
	FaultDomain *string `mandatory:"false" json:"faultDomain"`

	// The OCID of the cluster placement group of the instance.
	ClusterPlacementGroupId *string `mandatory:"false" json:"clusterPlacementGroupId"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the
	// compute cluster (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/compute-clusters.htm) that the instance will be created in.
	ComputeClusterId *string `mandatory:"false" json:"computeClusterId"`

	// Deprecated. Instead use `hostnameLabel` in
	// CreateVnicDetails.
	// If you provide both, the values must match.
	HostnameLabel *string `mandatory:"false" json:"hostnameLabel"`

	// Deprecated. Use `sourceDetails` with InstanceSourceViaImageDetails
	// source type instead. If you specify values for both, the values must match.
	ImageId *string `mandatory:"false" json:"imageId"`

	// This is an advanced option.
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

	LaunchOptions *LaunchOptions `mandatory:"false" json:"launchOptions"`

	InstanceOptions *InstanceOptions `mandatory:"false" json:"instanceOptions"`

	AvailabilityConfig *LaunchInstanceAvailabilityConfigDetails `mandatory:"false" json:"availabilityConfig"`

	PreemptibleInstanceConfig *PreemptibleInstanceConfigDetails `mandatory:"false" json:"preemptibleInstanceConfig"`

	// Custom metadata key/value pairs that you provide, such as the SSH public key
	// required to connect to the instance.
	// A metadata service runs on every launched instance. The service is an HTTP
	// endpoint listening on 169.254.169.254. You can use the service to:
	// * Provide information to Cloud-Init (https://cloudinit.readthedocs.org/en/latest/)
	//   to be used for various system initialization tasks.
	// * Get information about the instance, including the custom metadata that you
	//   provide when you launch the instance.
	//  **Providing Cloud-Init Metadata**
	//  You can use the following metadata key names to provide information to
	//  Cloud-Init:
	//  **"ssh_authorized_keys"** - Provide one or more public SSH keys to be
	//  included in the `~/.ssh/authorized_keys` file for the default user on the
	//  instance. Use a newline character to separate multiple keys. The SSH
	//  keys must be in the format necessary for the `authorized_keys` file, as shown
	//  in the example below.
	//  **"user_data"** - Provide your own base64-encoded data to be used by
	//  Cloud-Init to run custom scripts or provide custom Cloud-Init configuration. For
	//  information about how to take advantage of user data, see the
	//  Cloud-Init Documentation (http://cloudinit.readthedocs.org/en/latest/topics/format.html).
	//  **Metadata Example**
	//       "metadata" : {
	//          "quake_bot_level" : "Severe",
	//          "ssh_authorized_keys" : "ssh-rsa <your_public_SSH_key>== rsa-key-20160227",
	//          "user_data" : "<your_public_SSH_key>=="
	//       }
	//  **Getting Metadata on the Instance**
	//  To get information about your instance, connect to the instance using SSH and issue any of the
	//  following GET requests:
	//      curl -H "Authorization: Bearer Oracle" http://169.254.169.254/opc/v2/instance/
	//      curl -H "Authorization: Bearer Oracle" http://169.254.169.254/opc/v2/instance/metadata/
	//      curl -H "Authorization: Bearer Oracle" http://169.254.169.254/opc/v2/instance/metadata/<any-key-name>
	//  You'll get back a response that includes all the instance information; only the metadata information; or
	//  the metadata information for the specified key name, respectively.
	//  The combined size of the `metadata` and `extendedMetadata` objects can be a maximum of 32,000 bytes.
	Metadata map[string]string `mandatory:"false" json:"metadata"`

	AgentConfig *LaunchInstanceAgentConfigDetails `mandatory:"false" json:"agentConfig"`

	// The shape of an instance. The shape determines the number of CPUs, amount of memory,
	// and other resources allocated to the instance.
	// You can enumerate all available shapes by calling ListShapes.
	Shape *string `mandatory:"false" json:"shape"`

	ShapeConfig *LaunchInstanceShapeConfigDetails `mandatory:"false" json:"shapeConfig"`

	SourceDetails InstanceSourceDetails `mandatory:"false" json:"sourceDetails"`

	// Deprecated. Instead use `subnetId` in
	// CreateVnicDetails.
	// At least one of them is required; if you provide both, the values must match.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// Volume attachments to create as part of the launch instance operation.
	LaunchVolumeAttachments []LaunchAttachVolumeDetails `mandatory:"false" json:"launchVolumeAttachments"`

	// Whether to enable in-transit encryption for the data volume's paravirtualized attachment. This field applies to both block volumes and boot volumes. The default value is false.
	IsPvEncryptionInTransitEnabled *bool `mandatory:"false" json:"isPvEncryptionInTransitEnabled"`

	PlatformConfig LaunchInstancePlatformConfig `mandatory:"false" json:"platformConfig"`

	// The OCID of the Instance Configuration containing instance launch details. Any other fields supplied in this instance launch request will override the details stored in the Instance Configuration for this instance launch.
	InstanceConfigurationId *string `mandatory:"false" json:"instanceConfigurationId"`
}

func (m LaunchInstanceDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m LaunchInstanceDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *LaunchInstanceDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		CapacityReservationId          *string                                  `json:"capacityReservationId"`
		CreateVnicDetails              *CreateVnicDetails                       `json:"createVnicDetails"`
		DedicatedVmHostId              *string                                  `json:"dedicatedVmHostId"`
		DefinedTags                    map[string]map[string]interface{}        `json:"definedTags"`
		SecurityAttributes             map[string]map[string]interface{}        `json:"securityAttributes"`
		DisplayName                    *string                                  `json:"displayName"`
		ExtendedMetadata               map[string]interface{}                   `json:"extendedMetadata"`
		FaultDomain                    *string                                  `json:"faultDomain"`
		ClusterPlacementGroupId        *string                                  `json:"clusterPlacementGroupId"`
		FreeformTags                   map[string]string                        `json:"freeformTags"`
		ComputeClusterId               *string                                  `json:"computeClusterId"`
		HostnameLabel                  *string                                  `json:"hostnameLabel"`
		ImageId                        *string                                  `json:"imageId"`
		IpxeScript                     *string                                  `json:"ipxeScript"`
		LaunchOptions                  *LaunchOptions                           `json:"launchOptions"`
		InstanceOptions                *InstanceOptions                         `json:"instanceOptions"`
		AvailabilityConfig             *LaunchInstanceAvailabilityConfigDetails `json:"availabilityConfig"`
		PreemptibleInstanceConfig      *PreemptibleInstanceConfigDetails        `json:"preemptibleInstanceConfig"`
		Metadata                       map[string]string                        `json:"metadata"`
		AgentConfig                    *LaunchInstanceAgentConfigDetails        `json:"agentConfig"`
		Shape                          *string                                  `json:"shape"`
		ShapeConfig                    *LaunchInstanceShapeConfigDetails        `json:"shapeConfig"`
		SourceDetails                  instancesourcedetails                    `json:"sourceDetails"`
		SubnetId                       *string                                  `json:"subnetId"`
		LaunchVolumeAttachments        []launchattachvolumedetails              `json:"launchVolumeAttachments"`
		IsPvEncryptionInTransitEnabled *bool                                    `json:"isPvEncryptionInTransitEnabled"`
		PlatformConfig                 launchinstanceplatformconfig             `json:"platformConfig"`
		InstanceConfigurationId        *string                                  `json:"instanceConfigurationId"`
		AvailabilityDomain             *string                                  `json:"availabilityDomain"`
		CompartmentId                  *string                                  `json:"compartmentId"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.CapacityReservationId = model.CapacityReservationId

	m.CreateVnicDetails = model.CreateVnicDetails

	m.DedicatedVmHostId = model.DedicatedVmHostId

	m.DefinedTags = model.DefinedTags

	m.SecurityAttributes = model.SecurityAttributes

	m.DisplayName = model.DisplayName

	m.ExtendedMetadata = model.ExtendedMetadata

	m.FaultDomain = model.FaultDomain

	m.ClusterPlacementGroupId = model.ClusterPlacementGroupId

	m.FreeformTags = model.FreeformTags

	m.ComputeClusterId = model.ComputeClusterId

	m.HostnameLabel = model.HostnameLabel

	m.ImageId = model.ImageId

	m.IpxeScript = model.IpxeScript

	m.LaunchOptions = model.LaunchOptions

	m.InstanceOptions = model.InstanceOptions

	m.AvailabilityConfig = model.AvailabilityConfig

	m.PreemptibleInstanceConfig = model.PreemptibleInstanceConfig

	m.Metadata = model.Metadata

	m.AgentConfig = model.AgentConfig

	m.Shape = model.Shape

	m.ShapeConfig = model.ShapeConfig

	nn, e = model.SourceDetails.UnmarshalPolymorphicJSON(model.SourceDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.SourceDetails = nn.(InstanceSourceDetails)
	} else {
		m.SourceDetails = nil
	}

	m.SubnetId = model.SubnetId

	m.LaunchVolumeAttachments = make([]LaunchAttachVolumeDetails, len(model.LaunchVolumeAttachments))
	for i, n := range model.LaunchVolumeAttachments {
		nn, e = n.UnmarshalPolymorphicJSON(n.JsonData)
		if e != nil {
			return e
		}
		if nn != nil {
			m.LaunchVolumeAttachments[i] = nn.(LaunchAttachVolumeDetails)
		} else {
			m.LaunchVolumeAttachments[i] = nil
		}
	}
	m.IsPvEncryptionInTransitEnabled = model.IsPvEncryptionInTransitEnabled

	nn, e = model.PlatformConfig.UnmarshalPolymorphicJSON(model.PlatformConfig.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.PlatformConfig = nn.(LaunchInstancePlatformConfig)
	} else {
		m.PlatformConfig = nil
	}

	m.InstanceConfigurationId = model.InstanceConfigurationId

	m.AvailabilityDomain = model.AvailabilityDomain

	m.CompartmentId = model.CompartmentId

	return
}
