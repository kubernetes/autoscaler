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

// InternalVnicAttachment Details of a service VNIC attachment or an attachment of a non-service VNIC to a compute instance.
type InternalVnicAttachment struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the VNIC attachment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VNIC.
	Id *string `mandatory:"true" json:"id"`

	// The current state of a VNIC attachment.
	LifecycleState InternalVnicAttachmentLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The substrate or anycast IP address of the VNICaaS fleet that the VNIC is attached to.
	SubstrateIp *string `mandatory:"true" json:"substrateIp"`

	// The date and time the VNIC attachment was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The slot number of the VNIC.
	SlotId *int `mandatory:"false" json:"slotId"`

	// Shape of VNIC that is used to allocate resource in the data plane.
	VnicShape InternalVnicAttachmentVnicShapeEnum `mandatory:"false" json:"vnicShape,omitempty"`

	// The instance that a VNIC is attached to
	InstanceId *string `mandatory:"false" json:"instanceId"`

	// Composite key created from SubstrateIp, and data plane IDs of VCN and VNIC
	DataPlaneId *string `mandatory:"false" json:"dataPlaneId"`

	// The availability domain of a VNIC attachment
	InternalAvailabilityDomain *string `mandatory:"false" json:"internalAvailabilityDomain"`

	// The Network Address Translated IP to communicate with internal services
	NatIp *string `mandatory:"false" json:"natIp"`

	// The MAC address of the compute instance
	OverlayMac *string `mandatory:"false" json:"overlayMac"`

	// The tag used internally to identify sending VNIC
	VlanTag *int `mandatory:"false" json:"vlanTag"`

	// Index of NIC that VNIC is attached to (OS boot order)
	NicIndex *int `mandatory:"false" json:"nicIndex"`

	MigrationInfo *MigrationInfo `mandatory:"false" json:"migrationInfo"`

	// Property describing customer facing metrics
	MetadataList []CfmMetadata `mandatory:"false" json:"metadataList"`
}

func (m InternalVnicAttachment) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InternalVnicAttachment) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingInternalVnicAttachmentLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInternalVnicAttachmentLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := mappingInternalVnicAttachmentVnicShapeEnum[string(m.VnicShape)]; !ok && m.VnicShape != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for VnicShape: %s. Supported values are: %s.", m.VnicShape, strings.Join(GetInternalVnicAttachmentVnicShapeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InternalVnicAttachmentLifecycleStateEnum Enum with underlying type: string
type InternalVnicAttachmentLifecycleStateEnum string

// Set of constants representing the allowable values for InternalVnicAttachmentLifecycleStateEnum
const (
	InternalVnicAttachmentLifecycleStateAttaching InternalVnicAttachmentLifecycleStateEnum = "ATTACHING"
	InternalVnicAttachmentLifecycleStateAttached  InternalVnicAttachmentLifecycleStateEnum = "ATTACHED"
	InternalVnicAttachmentLifecycleStateDetaching InternalVnicAttachmentLifecycleStateEnum = "DETACHING"
	InternalVnicAttachmentLifecycleStateDetached  InternalVnicAttachmentLifecycleStateEnum = "DETACHED"
)

var mappingInternalVnicAttachmentLifecycleStateEnum = map[string]InternalVnicAttachmentLifecycleStateEnum{
	"ATTACHING": InternalVnicAttachmentLifecycleStateAttaching,
	"ATTACHED":  InternalVnicAttachmentLifecycleStateAttached,
	"DETACHING": InternalVnicAttachmentLifecycleStateDetaching,
	"DETACHED":  InternalVnicAttachmentLifecycleStateDetached,
}

// GetInternalVnicAttachmentLifecycleStateEnumValues Enumerates the set of values for InternalVnicAttachmentLifecycleStateEnum
func GetInternalVnicAttachmentLifecycleStateEnumValues() []InternalVnicAttachmentLifecycleStateEnum {
	values := make([]InternalVnicAttachmentLifecycleStateEnum, 0)
	for _, v := range mappingInternalVnicAttachmentLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalVnicAttachmentLifecycleStateEnumStringValues Enumerates the set of values in String for InternalVnicAttachmentLifecycleStateEnum
func GetInternalVnicAttachmentLifecycleStateEnumStringValues() []string {
	return []string{
		"ATTACHING",
		"ATTACHED",
		"DETACHING",
		"DETACHED",
	}
}

// InternalVnicAttachmentVnicShapeEnum Enum with underlying type: string
type InternalVnicAttachmentVnicShapeEnum string

// Set of constants representing the allowable values for InternalVnicAttachmentVnicShapeEnum
const (
	InternalVnicAttachmentVnicShapeDynamic                    InternalVnicAttachmentVnicShapeEnum = "DYNAMIC"
	InternalVnicAttachmentVnicShapeFixed0040                  InternalVnicAttachmentVnicShapeEnum = "FIXED0040"
	InternalVnicAttachmentVnicShapeFixed0060                  InternalVnicAttachmentVnicShapeEnum = "FIXED0060"
	InternalVnicAttachmentVnicShapeFixed0060Psm               InternalVnicAttachmentVnicShapeEnum = "FIXED0060_PSM"
	InternalVnicAttachmentVnicShapeFixed0100                  InternalVnicAttachmentVnicShapeEnum = "FIXED0100"
	InternalVnicAttachmentVnicShapeFixed0120                  InternalVnicAttachmentVnicShapeEnum = "FIXED0120"
	InternalVnicAttachmentVnicShapeFixed01202x                InternalVnicAttachmentVnicShapeEnum = "FIXED0120_2X"
	InternalVnicAttachmentVnicShapeFixed0200                  InternalVnicAttachmentVnicShapeEnum = "FIXED0200"
	InternalVnicAttachmentVnicShapeFixed0240                  InternalVnicAttachmentVnicShapeEnum = "FIXED0240"
	InternalVnicAttachmentVnicShapeFixed0480                  InternalVnicAttachmentVnicShapeEnum = "FIXED0480"
	InternalVnicAttachmentVnicShapeEntirehost                 InternalVnicAttachmentVnicShapeEnum = "ENTIREHOST"
	InternalVnicAttachmentVnicShapeDynamic25g                 InternalVnicAttachmentVnicShapeEnum = "DYNAMIC_25G"
	InternalVnicAttachmentVnicShapeFixed004025g               InternalVnicAttachmentVnicShapeEnum = "FIXED0040_25G"
	InternalVnicAttachmentVnicShapeFixed010025g               InternalVnicAttachmentVnicShapeEnum = "FIXED0100_25G"
	InternalVnicAttachmentVnicShapeFixed020025g               InternalVnicAttachmentVnicShapeEnum = "FIXED0200_25G"
	InternalVnicAttachmentVnicShapeFixed040025g               InternalVnicAttachmentVnicShapeEnum = "FIXED0400_25G"
	InternalVnicAttachmentVnicShapeFixed080025g               InternalVnicAttachmentVnicShapeEnum = "FIXED0800_25G"
	InternalVnicAttachmentVnicShapeFixed160025g               InternalVnicAttachmentVnicShapeEnum = "FIXED1600_25G"
	InternalVnicAttachmentVnicShapeFixed240025g               InternalVnicAttachmentVnicShapeEnum = "FIXED2400_25G"
	InternalVnicAttachmentVnicShapeEntirehost25g              InternalVnicAttachmentVnicShapeEnum = "ENTIREHOST_25G"
	InternalVnicAttachmentVnicShapeDynamicE125g               InternalVnicAttachmentVnicShapeEnum = "DYNAMIC_E1_25G"
	InternalVnicAttachmentVnicShapeFixed0040E125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0040_E1_25G"
	InternalVnicAttachmentVnicShapeFixed0070E125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0070_E1_25G"
	InternalVnicAttachmentVnicShapeFixed0140E125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0140_E1_25G"
	InternalVnicAttachmentVnicShapeFixed0280E125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0280_E1_25G"
	InternalVnicAttachmentVnicShapeFixed0560E125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0560_E1_25G"
	InternalVnicAttachmentVnicShapeFixed1120E125g             InternalVnicAttachmentVnicShapeEnum = "FIXED1120_E1_25G"
	InternalVnicAttachmentVnicShapeFixed1680E125g             InternalVnicAttachmentVnicShapeEnum = "FIXED1680_E1_25G"
	InternalVnicAttachmentVnicShapeEntirehostE125g            InternalVnicAttachmentVnicShapeEnum = "ENTIREHOST_E1_25G"
	InternalVnicAttachmentVnicShapeDynamicB125g               InternalVnicAttachmentVnicShapeEnum = "DYNAMIC_B1_25G"
	InternalVnicAttachmentVnicShapeFixed0040B125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0040_B1_25G"
	InternalVnicAttachmentVnicShapeFixed0060B125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0060_B1_25G"
	InternalVnicAttachmentVnicShapeFixed0120B125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0120_B1_25G"
	InternalVnicAttachmentVnicShapeFixed0240B125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0240_B1_25G"
	InternalVnicAttachmentVnicShapeFixed0480B125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0480_B1_25G"
	InternalVnicAttachmentVnicShapeFixed0960B125g             InternalVnicAttachmentVnicShapeEnum = "FIXED0960_B1_25G"
	InternalVnicAttachmentVnicShapeEntirehostB125g            InternalVnicAttachmentVnicShapeEnum = "ENTIREHOST_B1_25G"
	InternalVnicAttachmentVnicShapeMicroVmFixed0048E125g      InternalVnicAttachmentVnicShapeEnum = "MICRO_VM_FIXED0048_E1_25G"
	InternalVnicAttachmentVnicShapeMicroLbFixed0001E125g      InternalVnicAttachmentVnicShapeEnum = "MICRO_LB_FIXED0001_E1_25G"
	InternalVnicAttachmentVnicShapeVnicaasFixed0200           InternalVnicAttachmentVnicShapeEnum = "VNICAAS_FIXED0200"
	InternalVnicAttachmentVnicShapeVnicaasFixed0400           InternalVnicAttachmentVnicShapeEnum = "VNICAAS_FIXED0400"
	InternalVnicAttachmentVnicShapeVnicaasFixed0700           InternalVnicAttachmentVnicShapeEnum = "VNICAAS_FIXED0700"
	InternalVnicAttachmentVnicShapeVnicaasNlbApproved10g      InternalVnicAttachmentVnicShapeEnum = "VNICAAS_NLB_APPROVED_10G"
	InternalVnicAttachmentVnicShapeVnicaasNlbApproved25g      InternalVnicAttachmentVnicShapeEnum = "VNICAAS_NLB_APPROVED_25G"
	InternalVnicAttachmentVnicShapeVnicaasTelesis25g          InternalVnicAttachmentVnicShapeEnum = "VNICAAS_TELESIS_25G"
	InternalVnicAttachmentVnicShapeVnicaasTelesis10g          InternalVnicAttachmentVnicShapeEnum = "VNICAAS_TELESIS_10G"
	InternalVnicAttachmentVnicShapeVnicaasAmbassadorFixed0100 InternalVnicAttachmentVnicShapeEnum = "VNICAAS_AMBASSADOR_FIXED0100"
	InternalVnicAttachmentVnicShapeVnicaasPrivatedns          InternalVnicAttachmentVnicShapeEnum = "VNICAAS_PRIVATEDNS"
	InternalVnicAttachmentVnicShapeVnicaasFwaas               InternalVnicAttachmentVnicShapeEnum = "VNICAAS_FWAAS"
	InternalVnicAttachmentVnicShapeDynamicE350g               InternalVnicAttachmentVnicShapeEnum = "DYNAMIC_E3_50G"
	InternalVnicAttachmentVnicShapeFixed0040E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED0040_E3_50G"
	InternalVnicAttachmentVnicShapeFixed0100E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED0100_E3_50G"
	InternalVnicAttachmentVnicShapeFixed0200E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED0200_E3_50G"
	InternalVnicAttachmentVnicShapeFixed0300E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED0300_E3_50G"
	InternalVnicAttachmentVnicShapeFixed0400E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED0400_E3_50G"
	InternalVnicAttachmentVnicShapeFixed0500E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED0500_E3_50G"
	InternalVnicAttachmentVnicShapeFixed0600E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED0600_E3_50G"
	InternalVnicAttachmentVnicShapeFixed0700E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED0700_E3_50G"
	InternalVnicAttachmentVnicShapeFixed0800E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED0800_E3_50G"
	InternalVnicAttachmentVnicShapeFixed0900E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED0900_E3_50G"
	InternalVnicAttachmentVnicShapeFixed1000E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED1000_E3_50G"
	InternalVnicAttachmentVnicShapeFixed1100E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED1100_E3_50G"
	InternalVnicAttachmentVnicShapeFixed1200E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED1200_E3_50G"
	InternalVnicAttachmentVnicShapeFixed1300E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED1300_E3_50G"
	InternalVnicAttachmentVnicShapeFixed1400E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED1400_E3_50G"
	InternalVnicAttachmentVnicShapeFixed1500E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED1500_E3_50G"
	InternalVnicAttachmentVnicShapeFixed1600E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED1600_E3_50G"
	InternalVnicAttachmentVnicShapeFixed1700E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED1700_E3_50G"
	InternalVnicAttachmentVnicShapeFixed1800E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED1800_E3_50G"
	InternalVnicAttachmentVnicShapeFixed1900E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED1900_E3_50G"
	InternalVnicAttachmentVnicShapeFixed2000E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED2000_E3_50G"
	InternalVnicAttachmentVnicShapeFixed2100E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED2100_E3_50G"
	InternalVnicAttachmentVnicShapeFixed2200E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED2200_E3_50G"
	InternalVnicAttachmentVnicShapeFixed2300E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED2300_E3_50G"
	InternalVnicAttachmentVnicShapeFixed2400E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED2400_E3_50G"
	InternalVnicAttachmentVnicShapeFixed2500E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED2500_E3_50G"
	InternalVnicAttachmentVnicShapeFixed2600E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED2600_E3_50G"
	InternalVnicAttachmentVnicShapeFixed2700E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED2700_E3_50G"
	InternalVnicAttachmentVnicShapeFixed2800E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED2800_E3_50G"
	InternalVnicAttachmentVnicShapeFixed2900E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED2900_E3_50G"
	InternalVnicAttachmentVnicShapeFixed3000E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED3000_E3_50G"
	InternalVnicAttachmentVnicShapeFixed3100E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED3100_E3_50G"
	InternalVnicAttachmentVnicShapeFixed3200E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED3200_E3_50G"
	InternalVnicAttachmentVnicShapeFixed3300E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED3300_E3_50G"
	InternalVnicAttachmentVnicShapeFixed3400E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED3400_E3_50G"
	InternalVnicAttachmentVnicShapeFixed3500E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED3500_E3_50G"
	InternalVnicAttachmentVnicShapeFixed3600E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED3600_E3_50G"
	InternalVnicAttachmentVnicShapeFixed3700E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED3700_E3_50G"
	InternalVnicAttachmentVnicShapeFixed3800E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED3800_E3_50G"
	InternalVnicAttachmentVnicShapeFixed3900E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED3900_E3_50G"
	InternalVnicAttachmentVnicShapeFixed4000E350g             InternalVnicAttachmentVnicShapeEnum = "FIXED4000_E3_50G"
	InternalVnicAttachmentVnicShapeEntirehostE350g            InternalVnicAttachmentVnicShapeEnum = "ENTIREHOST_E3_50G"
	InternalVnicAttachmentVnicShapeDynamicE450g               InternalVnicAttachmentVnicShapeEnum = "DYNAMIC_E4_50G"
	InternalVnicAttachmentVnicShapeFixed0040E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED0040_E4_50G"
	InternalVnicAttachmentVnicShapeFixed0100E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED0100_E4_50G"
	InternalVnicAttachmentVnicShapeFixed0200E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED0200_E4_50G"
	InternalVnicAttachmentVnicShapeFixed0300E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED0300_E4_50G"
	InternalVnicAttachmentVnicShapeFixed0400E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED0400_E4_50G"
	InternalVnicAttachmentVnicShapeFixed0500E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED0500_E4_50G"
	InternalVnicAttachmentVnicShapeFixed0600E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED0600_E4_50G"
	InternalVnicAttachmentVnicShapeFixed0700E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED0700_E4_50G"
	InternalVnicAttachmentVnicShapeFixed0800E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED0800_E4_50G"
	InternalVnicAttachmentVnicShapeFixed0900E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED0900_E4_50G"
	InternalVnicAttachmentVnicShapeFixed1000E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED1000_E4_50G"
	InternalVnicAttachmentVnicShapeFixed1100E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED1100_E4_50G"
	InternalVnicAttachmentVnicShapeFixed1200E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED1200_E4_50G"
	InternalVnicAttachmentVnicShapeFixed1300E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED1300_E4_50G"
	InternalVnicAttachmentVnicShapeFixed1400E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED1400_E4_50G"
	InternalVnicAttachmentVnicShapeFixed1500E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED1500_E4_50G"
	InternalVnicAttachmentVnicShapeFixed1600E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED1600_E4_50G"
	InternalVnicAttachmentVnicShapeFixed1700E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED1700_E4_50G"
	InternalVnicAttachmentVnicShapeFixed1800E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED1800_E4_50G"
	InternalVnicAttachmentVnicShapeFixed1900E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED1900_E4_50G"
	InternalVnicAttachmentVnicShapeFixed2000E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED2000_E4_50G"
	InternalVnicAttachmentVnicShapeFixed2100E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED2100_E4_50G"
	InternalVnicAttachmentVnicShapeFixed2200E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED2200_E4_50G"
	InternalVnicAttachmentVnicShapeFixed2300E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED2300_E4_50G"
	InternalVnicAttachmentVnicShapeFixed2400E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED2400_E4_50G"
	InternalVnicAttachmentVnicShapeFixed2500E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED2500_E4_50G"
	InternalVnicAttachmentVnicShapeFixed2600E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED2600_E4_50G"
	InternalVnicAttachmentVnicShapeFixed2700E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED2700_E4_50G"
	InternalVnicAttachmentVnicShapeFixed2800E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED2800_E4_50G"
	InternalVnicAttachmentVnicShapeFixed2900E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED2900_E4_50G"
	InternalVnicAttachmentVnicShapeFixed3000E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED3000_E4_50G"
	InternalVnicAttachmentVnicShapeFixed3100E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED3100_E4_50G"
	InternalVnicAttachmentVnicShapeFixed3200E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED3200_E4_50G"
	InternalVnicAttachmentVnicShapeFixed3300E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED3300_E4_50G"
	InternalVnicAttachmentVnicShapeFixed3400E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED3400_E4_50G"
	InternalVnicAttachmentVnicShapeFixed3500E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED3500_E4_50G"
	InternalVnicAttachmentVnicShapeFixed3600E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED3600_E4_50G"
	InternalVnicAttachmentVnicShapeFixed3700E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED3700_E4_50G"
	InternalVnicAttachmentVnicShapeFixed3800E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED3800_E4_50G"
	InternalVnicAttachmentVnicShapeFixed3900E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED3900_E4_50G"
	InternalVnicAttachmentVnicShapeFixed4000E450g             InternalVnicAttachmentVnicShapeEnum = "FIXED4000_E4_50G"
	InternalVnicAttachmentVnicShapeEntirehostE450g            InternalVnicAttachmentVnicShapeEnum = "ENTIREHOST_E4_50G"
	InternalVnicAttachmentVnicShapeMicroVmFixed0050E350g      InternalVnicAttachmentVnicShapeEnum = "MICRO_VM_FIXED0050_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0025E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0025_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0050E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0050_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0075E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0075_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0100E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0100_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0125E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0125_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0150E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0150_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0175E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0175_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0200E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0200_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0225E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0225_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0250E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0250_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0275E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0275_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0300E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0300_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0325E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0325_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0350E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0350_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0375E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0375_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0400E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0400_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0425E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0425_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0450E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0450_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0475E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0475_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0500E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0500_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0525E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0525_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0550E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0550_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0575E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0575_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0600E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0600_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0625E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0625_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0650E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0650_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0675E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0675_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0700E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0700_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0725E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0725_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0750E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0750_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0775E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0775_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0800E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0800_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0825E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0825_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0850E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0850_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0875E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0875_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0900E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0900_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0925E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0925_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0950E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0950_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0975E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0975_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1000E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1000_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1025E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1025_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1050E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1050_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1075E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1075_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1100E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1100_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1125E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1125_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1150E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1150_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1175E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1175_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1200E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1200_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1225E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1225_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1250E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1250_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1275E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1275_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1300E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1300_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1325E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1325_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1350E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1350_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1375E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1375_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1400E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1400_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1425E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1425_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1450E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1450_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1475E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1475_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1500E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1500_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1525E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1525_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1550E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1550_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1575E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1575_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1600E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1600_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1625E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1625_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1650E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1650_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1700E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1700_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1725E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1725_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1750E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1750_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1800E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1800_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1850E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1850_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1875E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1875_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1900E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1900_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1925E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1925_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1950E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1950_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2000E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2000_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2025E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2025_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2050E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2050_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2100E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2100_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2125E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2125_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2150E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2150_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2175E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2175_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2200E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2200_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2250E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2250_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2275E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2275_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2300E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2300_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2325E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2325_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2350E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2350_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2375E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2375_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2400E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2400_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2450E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2450_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2475E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2475_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2500E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2500_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2550E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2550_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2600E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2600_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2625E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2625_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2650E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2650_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2700E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2700_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2750E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2750_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2775E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2775_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2800E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2800_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2850E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2850_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2875E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2875_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2900E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2900_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2925E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2925_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2950E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2950_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2975E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2975_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3000E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3000_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3025E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3025_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3050E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3050_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3075E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3075_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3100E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3100_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3125E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3125_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3150E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3150_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3200E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3200_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3225E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3225_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3250E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3250_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3300E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3300_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3325E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3325_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3375E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3375_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3400E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3400_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3450E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3450_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3500E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3500_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3525E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3525_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3575E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3575_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3600E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3600_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3625E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3625_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3675E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3675_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3700E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3700_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3750E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3750_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3800E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3800_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3825E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3825_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3850E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3850_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3875E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3875_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3900E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3900_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3975E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3975_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4000E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4000_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4025E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4025_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4050E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4050_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4100E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4100_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4125E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4125_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4200E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4200_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4225E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4225_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4250E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4250_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4275E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4275_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4300E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4300_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4350E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4350_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4375E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4375_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4400E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4400_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4425E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4425_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4500E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4500_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4550E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4550_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4575E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4575_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4600E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4600_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4625E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4625_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4650E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4650_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4675E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4675_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4700E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4700_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4725E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4725_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4750E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4750_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4800E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4800_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4875E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4875_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4900E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4900_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4950E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4950_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed5000E350g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED5000_E3_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0025E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0025_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0050E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0050_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0075E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0075_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0100E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0100_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0125E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0125_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0150E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0150_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0175E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0175_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0200E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0200_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0225E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0225_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0250E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0250_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0275E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0275_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0300E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0300_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0325E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0325_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0350E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0350_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0375E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0375_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0400E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0400_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0425E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0425_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0450E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0450_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0475E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0475_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0500E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0500_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0525E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0525_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0550E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0550_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0575E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0575_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0600E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0600_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0625E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0625_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0650E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0650_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0675E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0675_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0700E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0700_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0725E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0725_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0750E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0750_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0775E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0775_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0800E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0800_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0825E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0825_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0850E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0850_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0875E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0875_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0900E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0900_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0925E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0925_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0950E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0950_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0975E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0975_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1000E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1000_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1025E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1025_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1050E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1050_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1075E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1075_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1100E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1100_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1125E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1125_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1150E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1150_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1175E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1175_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1200E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1200_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1225E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1225_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1250E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1250_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1275E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1275_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1300E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1300_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1325E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1325_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1350E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1350_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1375E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1375_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1400E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1400_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1425E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1425_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1450E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1450_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1475E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1475_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1500E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1500_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1525E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1525_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1550E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1550_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1575E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1575_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1600E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1600_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1625E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1625_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1650E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1650_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1700E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1700_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1725E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1725_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1750E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1750_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1800E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1800_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1850E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1850_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1875E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1875_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1900E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1900_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1925E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1925_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1950E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1950_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2000E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2000_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2025E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2025_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2050E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2050_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2100E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2100_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2125E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2125_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2150E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2150_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2175E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2175_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2200E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2200_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2250E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2250_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2275E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2275_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2300E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2300_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2325E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2325_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2350E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2350_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2375E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2375_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2400E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2400_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2450E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2450_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2475E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2475_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2500E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2500_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2550E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2550_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2600E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2600_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2625E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2625_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2650E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2650_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2700E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2700_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2750E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2750_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2775E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2775_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2800E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2800_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2850E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2850_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2875E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2875_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2900E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2900_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2925E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2925_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2950E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2950_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2975E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2975_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3000E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3000_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3025E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3025_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3050E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3050_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3075E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3075_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3100E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3100_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3125E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3125_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3150E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3150_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3200E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3200_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3225E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3225_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3250E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3250_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3300E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3300_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3325E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3325_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3375E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3375_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3400E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3400_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3450E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3450_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3500E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3500_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3525E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3525_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3575E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3575_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3600E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3600_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3625E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3625_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3675E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3675_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3700E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3700_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3750E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3750_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3800E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3800_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3825E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3825_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3850E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3850_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3875E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3875_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3900E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3900_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3975E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3975_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4000E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4000_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4025E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4025_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4050E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4050_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4100E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4100_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4125E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4125_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4200E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4200_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4225E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4225_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4250E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4250_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4275E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4275_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4300E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4300_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4350E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4350_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4375E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4375_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4400E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4400_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4425E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4425_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4500E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4500_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4550E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4550_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4575E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4575_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4600E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4600_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4625E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4625_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4650E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4650_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4675E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4675_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4700E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4700_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4725E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4725_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4750E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4750_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4800E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4800_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4875E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4875_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4900E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4900_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4950E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4950_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed5000E450g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED5000_E4_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0020A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0020_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0040A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0040_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0060A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0060_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0080A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0080_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0100A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0100_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0120A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0120_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0140A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0140_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0160A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0160_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0180A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0180_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0200A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0200_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0220A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0220_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0240A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0240_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0260A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0260_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0280A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0280_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0300A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0300_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0320A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0320_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0340A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0340_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0360A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0360_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0380A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0380_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0400A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0400_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0420A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0420_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0440A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0440_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0460A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0460_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0480A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0480_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0500A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0500_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0520A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0520_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0540A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0540_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0560A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0560_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0580A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0580_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0600A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0600_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0620A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0620_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0640A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0640_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0660A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0660_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0680A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0680_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0700A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0700_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0720A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0720_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0740A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0740_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0760A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0760_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0780A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0780_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0800A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0800_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0820A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0820_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0840A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0840_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0860A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0860_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0880A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0880_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0900A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0900_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0920A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0920_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0940A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0940_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0960A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0960_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0980A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0980_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1000A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1000_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1020A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1020_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1040A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1040_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1060A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1060_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1080A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1080_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1100A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1100_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1120A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1120_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1140A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1140_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1160A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1160_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1180A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1180_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1200A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1200_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1220A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1220_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1240A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1240_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1260A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1260_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1280A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1280_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1300A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1300_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1320A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1320_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1340A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1340_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1360A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1360_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1380A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1380_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1400A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1400_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1420A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1420_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1440A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1440_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1460A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1460_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1480A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1480_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1500A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1500_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1520A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1520_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1540A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1540_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1560A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1560_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1580A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1580_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1600A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1600_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1620A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1620_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1640A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1640_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1660A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1660_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1680A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1680_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1700A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1700_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1720A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1720_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1740A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1740_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1760A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1760_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1780A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1780_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1800A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1800_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1820A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1820_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1840A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1840_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1860A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1860_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1880A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1880_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1900A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1900_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1920A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1920_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1940A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1940_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1960A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1960_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1980A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1980_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2000A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2000_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2020A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2020_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2040A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2040_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2060A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2060_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2080A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2080_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2100A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2100_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2120A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2120_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2140A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2140_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2160A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2160_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2180A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2180_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2200A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2200_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2220A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2220_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2240A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2240_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2260A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2260_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2280A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2280_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2300A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2300_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2320A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2320_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2340A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2340_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2360A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2360_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2380A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2380_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2400A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2400_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2420A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2420_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2440A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2440_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2460A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2460_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2480A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2480_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2500A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2500_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2520A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2520_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2540A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2540_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2560A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2560_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2580A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2580_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2600A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2600_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2620A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2620_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2640A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2640_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2660A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2660_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2680A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2680_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2700A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2700_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2720A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2720_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2740A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2740_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2760A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2760_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2780A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2780_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2800A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2800_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2820A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2820_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2840A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2840_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2860A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2860_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2880A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2880_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2900A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2900_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2920A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2920_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2940A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2940_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2960A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2960_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2980A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2980_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3000A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3000_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3020A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3020_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3040A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3040_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3060A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3060_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3080A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3080_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3100A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3100_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3120A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3120_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3140A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3140_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3160A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3160_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3180A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3180_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3200A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3200_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3220A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3220_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3240A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3240_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3260A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3260_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3280A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3280_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3300A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3300_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3320A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3320_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3340A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3340_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3360A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3360_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3380A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3380_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3400A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3400_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3420A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3420_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3440A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3440_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3460A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3460_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3480A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3480_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3500A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3500_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3520A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3520_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3540A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3540_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3560A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3560_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3580A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3580_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3600A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3600_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3620A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3620_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3640A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3640_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3660A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3660_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3680A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3680_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3700A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3700_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3720A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3720_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3740A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3740_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3760A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3760_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3780A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3780_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3800A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3800_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3820A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3820_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3840A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3840_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3860A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3860_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3880A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3880_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3900A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3900_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3920A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3920_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3940A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3940_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3960A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3960_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3980A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3980_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4000A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4000_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4020A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4020_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4040A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4040_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4060A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4060_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4080A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4080_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4100A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4100_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4120A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4120_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4140A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4140_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4160A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4160_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4180A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4180_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4200A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4200_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4220A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4220_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4240A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4240_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4260A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4260_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4280A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4280_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4300A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4300_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4320A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4320_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4340A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4340_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4360A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4360_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4380A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4380_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4400A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4400_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4420A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4420_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4440A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4440_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4460A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4460_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4480A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4480_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4500A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4500_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4520A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4520_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4540A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4540_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4560A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4560_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4580A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4580_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4600A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4600_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4620A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4620_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4640A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4640_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4660A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4660_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4680A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4680_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4700A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4700_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4720A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4720_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4740A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4740_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4760A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4760_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4780A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4780_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4800A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4800_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4820A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4820_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4840A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4840_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4860A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4860_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4880A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4880_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4900A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4900_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4920A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4920_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4940A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4940_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4960A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4960_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4980A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4980_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed5000A150g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED5000_A1_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0090X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0090_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0180X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0180_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0270X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0270_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0360X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0360_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0450X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0450_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0540X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0540_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0630X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0630_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0720X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0720_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0810X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0810_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0900X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0900_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed0990X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED0990_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1080X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1080_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1170X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1170_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1260X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1260_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1350X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1350_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1440X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1440_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1530X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1530_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1620X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1620_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1710X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1710_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1800X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1800_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1890X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1890_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed1980X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED1980_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2070X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2070_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2160X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2160_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2250X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2250_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2340X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2340_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2430X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2430_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2520X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2520_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2610X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2610_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2700X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2700_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2790X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2790_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2880X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2880_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed2970X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED2970_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3060X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3060_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3150X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3150_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3240X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3240_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3330X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3330_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3420X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3420_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3510X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3510_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3600X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3600_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3690X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3690_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3780X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3780_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3870X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3870_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed3960X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED3960_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4050X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4050_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4140X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4140_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4230X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4230_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4320X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4320_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4410X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4410_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4500X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4500_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4590X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4590_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4680X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4680_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4770X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4770_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4860X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4860_X9_50G"
	InternalVnicAttachmentVnicShapeSubcoreVmFixed4950X950g    InternalVnicAttachmentVnicShapeEnum = "SUBCORE_VM_FIXED4950_X9_50G"
	InternalVnicAttachmentVnicShapeDynamicA150g               InternalVnicAttachmentVnicShapeEnum = "DYNAMIC_A1_50G"
	InternalVnicAttachmentVnicShapeFixed0040A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED0040_A1_50G"
	InternalVnicAttachmentVnicShapeFixed0100A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED0100_A1_50G"
	InternalVnicAttachmentVnicShapeFixed0200A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED0200_A1_50G"
	InternalVnicAttachmentVnicShapeFixed0300A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED0300_A1_50G"
	InternalVnicAttachmentVnicShapeFixed0400A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED0400_A1_50G"
	InternalVnicAttachmentVnicShapeFixed0500A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED0500_A1_50G"
	InternalVnicAttachmentVnicShapeFixed0600A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED0600_A1_50G"
	InternalVnicAttachmentVnicShapeFixed0700A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED0700_A1_50G"
	InternalVnicAttachmentVnicShapeFixed0800A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED0800_A1_50G"
	InternalVnicAttachmentVnicShapeFixed0900A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED0900_A1_50G"
	InternalVnicAttachmentVnicShapeFixed1000A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED1000_A1_50G"
	InternalVnicAttachmentVnicShapeFixed1100A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED1100_A1_50G"
	InternalVnicAttachmentVnicShapeFixed1200A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED1200_A1_50G"
	InternalVnicAttachmentVnicShapeFixed1300A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED1300_A1_50G"
	InternalVnicAttachmentVnicShapeFixed1400A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED1400_A1_50G"
	InternalVnicAttachmentVnicShapeFixed1500A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED1500_A1_50G"
	InternalVnicAttachmentVnicShapeFixed1600A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED1600_A1_50G"
	InternalVnicAttachmentVnicShapeFixed1700A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED1700_A1_50G"
	InternalVnicAttachmentVnicShapeFixed1800A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED1800_A1_50G"
	InternalVnicAttachmentVnicShapeFixed1900A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED1900_A1_50G"
	InternalVnicAttachmentVnicShapeFixed2000A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED2000_A1_50G"
	InternalVnicAttachmentVnicShapeFixed2100A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED2100_A1_50G"
	InternalVnicAttachmentVnicShapeFixed2200A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED2200_A1_50G"
	InternalVnicAttachmentVnicShapeFixed2300A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED2300_A1_50G"
	InternalVnicAttachmentVnicShapeFixed2400A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED2400_A1_50G"
	InternalVnicAttachmentVnicShapeFixed2500A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED2500_A1_50G"
	InternalVnicAttachmentVnicShapeFixed2600A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED2600_A1_50G"
	InternalVnicAttachmentVnicShapeFixed2700A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED2700_A1_50G"
	InternalVnicAttachmentVnicShapeFixed2800A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED2800_A1_50G"
	InternalVnicAttachmentVnicShapeFixed2900A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED2900_A1_50G"
	InternalVnicAttachmentVnicShapeFixed3000A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED3000_A1_50G"
	InternalVnicAttachmentVnicShapeFixed3100A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED3100_A1_50G"
	InternalVnicAttachmentVnicShapeFixed3200A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED3200_A1_50G"
	InternalVnicAttachmentVnicShapeFixed3300A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED3300_A1_50G"
	InternalVnicAttachmentVnicShapeFixed3400A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED3400_A1_50G"
	InternalVnicAttachmentVnicShapeFixed3500A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED3500_A1_50G"
	InternalVnicAttachmentVnicShapeFixed3600A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED3600_A1_50G"
	InternalVnicAttachmentVnicShapeFixed3700A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED3700_A1_50G"
	InternalVnicAttachmentVnicShapeFixed3800A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED3800_A1_50G"
	InternalVnicAttachmentVnicShapeFixed3900A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED3900_A1_50G"
	InternalVnicAttachmentVnicShapeFixed4000A150g             InternalVnicAttachmentVnicShapeEnum = "FIXED4000_A1_50G"
	InternalVnicAttachmentVnicShapeEntirehostA150g            InternalVnicAttachmentVnicShapeEnum = "ENTIREHOST_A1_50G"
	InternalVnicAttachmentVnicShapeDynamicX950g               InternalVnicAttachmentVnicShapeEnum = "DYNAMIC_X9_50G"
	InternalVnicAttachmentVnicShapeFixed0040X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED0040_X9_50G"
	InternalVnicAttachmentVnicShapeFixed0400X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED0400_X9_50G"
	InternalVnicAttachmentVnicShapeFixed0800X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED0800_X9_50G"
	InternalVnicAttachmentVnicShapeFixed1200X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED1200_X9_50G"
	InternalVnicAttachmentVnicShapeFixed1600X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED1600_X9_50G"
	InternalVnicAttachmentVnicShapeFixed2000X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED2000_X9_50G"
	InternalVnicAttachmentVnicShapeFixed2400X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED2400_X9_50G"
	InternalVnicAttachmentVnicShapeFixed2800X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED2800_X9_50G"
	InternalVnicAttachmentVnicShapeFixed3200X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED3200_X9_50G"
	InternalVnicAttachmentVnicShapeFixed3600X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED3600_X9_50G"
	InternalVnicAttachmentVnicShapeFixed4000X950g             InternalVnicAttachmentVnicShapeEnum = "FIXED4000_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed0100X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED0100_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed0200X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED0200_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed0300X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED0300_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed0400X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED0400_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed0500X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED0500_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed0600X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED0600_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed0700X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED0700_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed0800X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED0800_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed0900X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED0900_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed1000X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED1000_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed1100X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED1100_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed1200X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED1200_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed1300X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED1300_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed1400X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED1400_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed1500X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED1500_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed1600X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED1600_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed1700X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED1700_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed1800X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED1800_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed1900X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED1900_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed2000X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED2000_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed2100X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED2100_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed2200X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED2200_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed2300X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED2300_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed2400X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED2400_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed2500X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED2500_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed2600X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED2600_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed2700X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED2700_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed2800X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED2800_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed2900X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED2900_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed3000X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED3000_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed3100X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED3100_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed3200X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED3200_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed3300X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED3300_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed3400X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED3400_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed3500X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED3500_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed3600X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED3600_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed3700X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED3700_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed3800X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED3800_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed3900X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED3900_X9_50G"
	InternalVnicAttachmentVnicShapeStandardVmFixed4000X950g   InternalVnicAttachmentVnicShapeEnum = "STANDARD_VM_FIXED4000_X9_50G"
	InternalVnicAttachmentVnicShapeEntirehostX950g            InternalVnicAttachmentVnicShapeEnum = "ENTIREHOST_X9_50G"
)

var mappingInternalVnicAttachmentVnicShapeEnum = map[string]InternalVnicAttachmentVnicShapeEnum{
	"DYNAMIC":                      InternalVnicAttachmentVnicShapeDynamic,
	"FIXED0040":                    InternalVnicAttachmentVnicShapeFixed0040,
	"FIXED0060":                    InternalVnicAttachmentVnicShapeFixed0060,
	"FIXED0060_PSM":                InternalVnicAttachmentVnicShapeFixed0060Psm,
	"FIXED0100":                    InternalVnicAttachmentVnicShapeFixed0100,
	"FIXED0120":                    InternalVnicAttachmentVnicShapeFixed0120,
	"FIXED0120_2X":                 InternalVnicAttachmentVnicShapeFixed01202x,
	"FIXED0200":                    InternalVnicAttachmentVnicShapeFixed0200,
	"FIXED0240":                    InternalVnicAttachmentVnicShapeFixed0240,
	"FIXED0480":                    InternalVnicAttachmentVnicShapeFixed0480,
	"ENTIREHOST":                   InternalVnicAttachmentVnicShapeEntirehost,
	"DYNAMIC_25G":                  InternalVnicAttachmentVnicShapeDynamic25g,
	"FIXED0040_25G":                InternalVnicAttachmentVnicShapeFixed004025g,
	"FIXED0100_25G":                InternalVnicAttachmentVnicShapeFixed010025g,
	"FIXED0200_25G":                InternalVnicAttachmentVnicShapeFixed020025g,
	"FIXED0400_25G":                InternalVnicAttachmentVnicShapeFixed040025g,
	"FIXED0800_25G":                InternalVnicAttachmentVnicShapeFixed080025g,
	"FIXED1600_25G":                InternalVnicAttachmentVnicShapeFixed160025g,
	"FIXED2400_25G":                InternalVnicAttachmentVnicShapeFixed240025g,
	"ENTIREHOST_25G":               InternalVnicAttachmentVnicShapeEntirehost25g,
	"DYNAMIC_E1_25G":               InternalVnicAttachmentVnicShapeDynamicE125g,
	"FIXED0040_E1_25G":             InternalVnicAttachmentVnicShapeFixed0040E125g,
	"FIXED0070_E1_25G":             InternalVnicAttachmentVnicShapeFixed0070E125g,
	"FIXED0140_E1_25G":             InternalVnicAttachmentVnicShapeFixed0140E125g,
	"FIXED0280_E1_25G":             InternalVnicAttachmentVnicShapeFixed0280E125g,
	"FIXED0560_E1_25G":             InternalVnicAttachmentVnicShapeFixed0560E125g,
	"FIXED1120_E1_25G":             InternalVnicAttachmentVnicShapeFixed1120E125g,
	"FIXED1680_E1_25G":             InternalVnicAttachmentVnicShapeFixed1680E125g,
	"ENTIREHOST_E1_25G":            InternalVnicAttachmentVnicShapeEntirehostE125g,
	"DYNAMIC_B1_25G":               InternalVnicAttachmentVnicShapeDynamicB125g,
	"FIXED0040_B1_25G":             InternalVnicAttachmentVnicShapeFixed0040B125g,
	"FIXED0060_B1_25G":             InternalVnicAttachmentVnicShapeFixed0060B125g,
	"FIXED0120_B1_25G":             InternalVnicAttachmentVnicShapeFixed0120B125g,
	"FIXED0240_B1_25G":             InternalVnicAttachmentVnicShapeFixed0240B125g,
	"FIXED0480_B1_25G":             InternalVnicAttachmentVnicShapeFixed0480B125g,
	"FIXED0960_B1_25G":             InternalVnicAttachmentVnicShapeFixed0960B125g,
	"ENTIREHOST_B1_25G":            InternalVnicAttachmentVnicShapeEntirehostB125g,
	"MICRO_VM_FIXED0048_E1_25G":    InternalVnicAttachmentVnicShapeMicroVmFixed0048E125g,
	"MICRO_LB_FIXED0001_E1_25G":    InternalVnicAttachmentVnicShapeMicroLbFixed0001E125g,
	"VNICAAS_FIXED0200":            InternalVnicAttachmentVnicShapeVnicaasFixed0200,
	"VNICAAS_FIXED0400":            InternalVnicAttachmentVnicShapeVnicaasFixed0400,
	"VNICAAS_FIXED0700":            InternalVnicAttachmentVnicShapeVnicaasFixed0700,
	"VNICAAS_NLB_APPROVED_10G":     InternalVnicAttachmentVnicShapeVnicaasNlbApproved10g,
	"VNICAAS_NLB_APPROVED_25G":     InternalVnicAttachmentVnicShapeVnicaasNlbApproved25g,
	"VNICAAS_TELESIS_25G":          InternalVnicAttachmentVnicShapeVnicaasTelesis25g,
	"VNICAAS_TELESIS_10G":          InternalVnicAttachmentVnicShapeVnicaasTelesis10g,
	"VNICAAS_AMBASSADOR_FIXED0100": InternalVnicAttachmentVnicShapeVnicaasAmbassadorFixed0100,
	"VNICAAS_PRIVATEDNS":           InternalVnicAttachmentVnicShapeVnicaasPrivatedns,
	"VNICAAS_FWAAS":                InternalVnicAttachmentVnicShapeVnicaasFwaas,
	"DYNAMIC_E3_50G":               InternalVnicAttachmentVnicShapeDynamicE350g,
	"FIXED0040_E3_50G":             InternalVnicAttachmentVnicShapeFixed0040E350g,
	"FIXED0100_E3_50G":             InternalVnicAttachmentVnicShapeFixed0100E350g,
	"FIXED0200_E3_50G":             InternalVnicAttachmentVnicShapeFixed0200E350g,
	"FIXED0300_E3_50G":             InternalVnicAttachmentVnicShapeFixed0300E350g,
	"FIXED0400_E3_50G":             InternalVnicAttachmentVnicShapeFixed0400E350g,
	"FIXED0500_E3_50G":             InternalVnicAttachmentVnicShapeFixed0500E350g,
	"FIXED0600_E3_50G":             InternalVnicAttachmentVnicShapeFixed0600E350g,
	"FIXED0700_E3_50G":             InternalVnicAttachmentVnicShapeFixed0700E350g,
	"FIXED0800_E3_50G":             InternalVnicAttachmentVnicShapeFixed0800E350g,
	"FIXED0900_E3_50G":             InternalVnicAttachmentVnicShapeFixed0900E350g,
	"FIXED1000_E3_50G":             InternalVnicAttachmentVnicShapeFixed1000E350g,
	"FIXED1100_E3_50G":             InternalVnicAttachmentVnicShapeFixed1100E350g,
	"FIXED1200_E3_50G":             InternalVnicAttachmentVnicShapeFixed1200E350g,
	"FIXED1300_E3_50G":             InternalVnicAttachmentVnicShapeFixed1300E350g,
	"FIXED1400_E3_50G":             InternalVnicAttachmentVnicShapeFixed1400E350g,
	"FIXED1500_E3_50G":             InternalVnicAttachmentVnicShapeFixed1500E350g,
	"FIXED1600_E3_50G":             InternalVnicAttachmentVnicShapeFixed1600E350g,
	"FIXED1700_E3_50G":             InternalVnicAttachmentVnicShapeFixed1700E350g,
	"FIXED1800_E3_50G":             InternalVnicAttachmentVnicShapeFixed1800E350g,
	"FIXED1900_E3_50G":             InternalVnicAttachmentVnicShapeFixed1900E350g,
	"FIXED2000_E3_50G":             InternalVnicAttachmentVnicShapeFixed2000E350g,
	"FIXED2100_E3_50G":             InternalVnicAttachmentVnicShapeFixed2100E350g,
	"FIXED2200_E3_50G":             InternalVnicAttachmentVnicShapeFixed2200E350g,
	"FIXED2300_E3_50G":             InternalVnicAttachmentVnicShapeFixed2300E350g,
	"FIXED2400_E3_50G":             InternalVnicAttachmentVnicShapeFixed2400E350g,
	"FIXED2500_E3_50G":             InternalVnicAttachmentVnicShapeFixed2500E350g,
	"FIXED2600_E3_50G":             InternalVnicAttachmentVnicShapeFixed2600E350g,
	"FIXED2700_E3_50G":             InternalVnicAttachmentVnicShapeFixed2700E350g,
	"FIXED2800_E3_50G":             InternalVnicAttachmentVnicShapeFixed2800E350g,
	"FIXED2900_E3_50G":             InternalVnicAttachmentVnicShapeFixed2900E350g,
	"FIXED3000_E3_50G":             InternalVnicAttachmentVnicShapeFixed3000E350g,
	"FIXED3100_E3_50G":             InternalVnicAttachmentVnicShapeFixed3100E350g,
	"FIXED3200_E3_50G":             InternalVnicAttachmentVnicShapeFixed3200E350g,
	"FIXED3300_E3_50G":             InternalVnicAttachmentVnicShapeFixed3300E350g,
	"FIXED3400_E3_50G":             InternalVnicAttachmentVnicShapeFixed3400E350g,
	"FIXED3500_E3_50G":             InternalVnicAttachmentVnicShapeFixed3500E350g,
	"FIXED3600_E3_50G":             InternalVnicAttachmentVnicShapeFixed3600E350g,
	"FIXED3700_E3_50G":             InternalVnicAttachmentVnicShapeFixed3700E350g,
	"FIXED3800_E3_50G":             InternalVnicAttachmentVnicShapeFixed3800E350g,
	"FIXED3900_E3_50G":             InternalVnicAttachmentVnicShapeFixed3900E350g,
	"FIXED4000_E3_50G":             InternalVnicAttachmentVnicShapeFixed4000E350g,
	"ENTIREHOST_E3_50G":            InternalVnicAttachmentVnicShapeEntirehostE350g,
	"DYNAMIC_E4_50G":               InternalVnicAttachmentVnicShapeDynamicE450g,
	"FIXED0040_E4_50G":             InternalVnicAttachmentVnicShapeFixed0040E450g,
	"FIXED0100_E4_50G":             InternalVnicAttachmentVnicShapeFixed0100E450g,
	"FIXED0200_E4_50G":             InternalVnicAttachmentVnicShapeFixed0200E450g,
	"FIXED0300_E4_50G":             InternalVnicAttachmentVnicShapeFixed0300E450g,
	"FIXED0400_E4_50G":             InternalVnicAttachmentVnicShapeFixed0400E450g,
	"FIXED0500_E4_50G":             InternalVnicAttachmentVnicShapeFixed0500E450g,
	"FIXED0600_E4_50G":             InternalVnicAttachmentVnicShapeFixed0600E450g,
	"FIXED0700_E4_50G":             InternalVnicAttachmentVnicShapeFixed0700E450g,
	"FIXED0800_E4_50G":             InternalVnicAttachmentVnicShapeFixed0800E450g,
	"FIXED0900_E4_50G":             InternalVnicAttachmentVnicShapeFixed0900E450g,
	"FIXED1000_E4_50G":             InternalVnicAttachmentVnicShapeFixed1000E450g,
	"FIXED1100_E4_50G":             InternalVnicAttachmentVnicShapeFixed1100E450g,
	"FIXED1200_E4_50G":             InternalVnicAttachmentVnicShapeFixed1200E450g,
	"FIXED1300_E4_50G":             InternalVnicAttachmentVnicShapeFixed1300E450g,
	"FIXED1400_E4_50G":             InternalVnicAttachmentVnicShapeFixed1400E450g,
	"FIXED1500_E4_50G":             InternalVnicAttachmentVnicShapeFixed1500E450g,
	"FIXED1600_E4_50G":             InternalVnicAttachmentVnicShapeFixed1600E450g,
	"FIXED1700_E4_50G":             InternalVnicAttachmentVnicShapeFixed1700E450g,
	"FIXED1800_E4_50G":             InternalVnicAttachmentVnicShapeFixed1800E450g,
	"FIXED1900_E4_50G":             InternalVnicAttachmentVnicShapeFixed1900E450g,
	"FIXED2000_E4_50G":             InternalVnicAttachmentVnicShapeFixed2000E450g,
	"FIXED2100_E4_50G":             InternalVnicAttachmentVnicShapeFixed2100E450g,
	"FIXED2200_E4_50G":             InternalVnicAttachmentVnicShapeFixed2200E450g,
	"FIXED2300_E4_50G":             InternalVnicAttachmentVnicShapeFixed2300E450g,
	"FIXED2400_E4_50G":             InternalVnicAttachmentVnicShapeFixed2400E450g,
	"FIXED2500_E4_50G":             InternalVnicAttachmentVnicShapeFixed2500E450g,
	"FIXED2600_E4_50G":             InternalVnicAttachmentVnicShapeFixed2600E450g,
	"FIXED2700_E4_50G":             InternalVnicAttachmentVnicShapeFixed2700E450g,
	"FIXED2800_E4_50G":             InternalVnicAttachmentVnicShapeFixed2800E450g,
	"FIXED2900_E4_50G":             InternalVnicAttachmentVnicShapeFixed2900E450g,
	"FIXED3000_E4_50G":             InternalVnicAttachmentVnicShapeFixed3000E450g,
	"FIXED3100_E4_50G":             InternalVnicAttachmentVnicShapeFixed3100E450g,
	"FIXED3200_E4_50G":             InternalVnicAttachmentVnicShapeFixed3200E450g,
	"FIXED3300_E4_50G":             InternalVnicAttachmentVnicShapeFixed3300E450g,
	"FIXED3400_E4_50G":             InternalVnicAttachmentVnicShapeFixed3400E450g,
	"FIXED3500_E4_50G":             InternalVnicAttachmentVnicShapeFixed3500E450g,
	"FIXED3600_E4_50G":             InternalVnicAttachmentVnicShapeFixed3600E450g,
	"FIXED3700_E4_50G":             InternalVnicAttachmentVnicShapeFixed3700E450g,
	"FIXED3800_E4_50G":             InternalVnicAttachmentVnicShapeFixed3800E450g,
	"FIXED3900_E4_50G":             InternalVnicAttachmentVnicShapeFixed3900E450g,
	"FIXED4000_E4_50G":             InternalVnicAttachmentVnicShapeFixed4000E450g,
	"ENTIREHOST_E4_50G":            InternalVnicAttachmentVnicShapeEntirehostE450g,
	"MICRO_VM_FIXED0050_E3_50G":    InternalVnicAttachmentVnicShapeMicroVmFixed0050E350g,
	"SUBCORE_VM_FIXED0025_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0025E350g,
	"SUBCORE_VM_FIXED0050_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0050E350g,
	"SUBCORE_VM_FIXED0075_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0075E350g,
	"SUBCORE_VM_FIXED0100_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0100E350g,
	"SUBCORE_VM_FIXED0125_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0125E350g,
	"SUBCORE_VM_FIXED0150_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0150E350g,
	"SUBCORE_VM_FIXED0175_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0175E350g,
	"SUBCORE_VM_FIXED0200_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0200E350g,
	"SUBCORE_VM_FIXED0225_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0225E350g,
	"SUBCORE_VM_FIXED0250_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0250E350g,
	"SUBCORE_VM_FIXED0275_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0275E350g,
	"SUBCORE_VM_FIXED0300_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0300E350g,
	"SUBCORE_VM_FIXED0325_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0325E350g,
	"SUBCORE_VM_FIXED0350_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0350E350g,
	"SUBCORE_VM_FIXED0375_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0375E350g,
	"SUBCORE_VM_FIXED0400_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0400E350g,
	"SUBCORE_VM_FIXED0425_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0425E350g,
	"SUBCORE_VM_FIXED0450_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0450E350g,
	"SUBCORE_VM_FIXED0475_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0475E350g,
	"SUBCORE_VM_FIXED0500_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0500E350g,
	"SUBCORE_VM_FIXED0525_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0525E350g,
	"SUBCORE_VM_FIXED0550_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0550E350g,
	"SUBCORE_VM_FIXED0575_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0575E350g,
	"SUBCORE_VM_FIXED0600_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0600E350g,
	"SUBCORE_VM_FIXED0625_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0625E350g,
	"SUBCORE_VM_FIXED0650_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0650E350g,
	"SUBCORE_VM_FIXED0675_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0675E350g,
	"SUBCORE_VM_FIXED0700_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0700E350g,
	"SUBCORE_VM_FIXED0725_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0725E350g,
	"SUBCORE_VM_FIXED0750_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0750E350g,
	"SUBCORE_VM_FIXED0775_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0775E350g,
	"SUBCORE_VM_FIXED0800_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0800E350g,
	"SUBCORE_VM_FIXED0825_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0825E350g,
	"SUBCORE_VM_FIXED0850_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0850E350g,
	"SUBCORE_VM_FIXED0875_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0875E350g,
	"SUBCORE_VM_FIXED0900_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0900E350g,
	"SUBCORE_VM_FIXED0925_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0925E350g,
	"SUBCORE_VM_FIXED0950_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0950E350g,
	"SUBCORE_VM_FIXED0975_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0975E350g,
	"SUBCORE_VM_FIXED1000_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1000E350g,
	"SUBCORE_VM_FIXED1025_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1025E350g,
	"SUBCORE_VM_FIXED1050_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1050E350g,
	"SUBCORE_VM_FIXED1075_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1075E350g,
	"SUBCORE_VM_FIXED1100_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1100E350g,
	"SUBCORE_VM_FIXED1125_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1125E350g,
	"SUBCORE_VM_FIXED1150_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1150E350g,
	"SUBCORE_VM_FIXED1175_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1175E350g,
	"SUBCORE_VM_FIXED1200_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1200E350g,
	"SUBCORE_VM_FIXED1225_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1225E350g,
	"SUBCORE_VM_FIXED1250_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1250E350g,
	"SUBCORE_VM_FIXED1275_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1275E350g,
	"SUBCORE_VM_FIXED1300_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1300E350g,
	"SUBCORE_VM_FIXED1325_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1325E350g,
	"SUBCORE_VM_FIXED1350_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1350E350g,
	"SUBCORE_VM_FIXED1375_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1375E350g,
	"SUBCORE_VM_FIXED1400_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1400E350g,
	"SUBCORE_VM_FIXED1425_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1425E350g,
	"SUBCORE_VM_FIXED1450_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1450E350g,
	"SUBCORE_VM_FIXED1475_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1475E350g,
	"SUBCORE_VM_FIXED1500_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1500E350g,
	"SUBCORE_VM_FIXED1525_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1525E350g,
	"SUBCORE_VM_FIXED1550_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1550E350g,
	"SUBCORE_VM_FIXED1575_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1575E350g,
	"SUBCORE_VM_FIXED1600_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1600E350g,
	"SUBCORE_VM_FIXED1625_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1625E350g,
	"SUBCORE_VM_FIXED1650_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1650E350g,
	"SUBCORE_VM_FIXED1700_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1700E350g,
	"SUBCORE_VM_FIXED1725_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1725E350g,
	"SUBCORE_VM_FIXED1750_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1750E350g,
	"SUBCORE_VM_FIXED1800_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1800E350g,
	"SUBCORE_VM_FIXED1850_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1850E350g,
	"SUBCORE_VM_FIXED1875_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1875E350g,
	"SUBCORE_VM_FIXED1900_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1900E350g,
	"SUBCORE_VM_FIXED1925_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1925E350g,
	"SUBCORE_VM_FIXED1950_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1950E350g,
	"SUBCORE_VM_FIXED2000_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2000E350g,
	"SUBCORE_VM_FIXED2025_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2025E350g,
	"SUBCORE_VM_FIXED2050_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2050E350g,
	"SUBCORE_VM_FIXED2100_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2100E350g,
	"SUBCORE_VM_FIXED2125_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2125E350g,
	"SUBCORE_VM_FIXED2150_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2150E350g,
	"SUBCORE_VM_FIXED2175_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2175E350g,
	"SUBCORE_VM_FIXED2200_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2200E350g,
	"SUBCORE_VM_FIXED2250_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2250E350g,
	"SUBCORE_VM_FIXED2275_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2275E350g,
	"SUBCORE_VM_FIXED2300_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2300E350g,
	"SUBCORE_VM_FIXED2325_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2325E350g,
	"SUBCORE_VM_FIXED2350_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2350E350g,
	"SUBCORE_VM_FIXED2375_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2375E350g,
	"SUBCORE_VM_FIXED2400_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2400E350g,
	"SUBCORE_VM_FIXED2450_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2450E350g,
	"SUBCORE_VM_FIXED2475_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2475E350g,
	"SUBCORE_VM_FIXED2500_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2500E350g,
	"SUBCORE_VM_FIXED2550_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2550E350g,
	"SUBCORE_VM_FIXED2600_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2600E350g,
	"SUBCORE_VM_FIXED2625_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2625E350g,
	"SUBCORE_VM_FIXED2650_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2650E350g,
	"SUBCORE_VM_FIXED2700_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2700E350g,
	"SUBCORE_VM_FIXED2750_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2750E350g,
	"SUBCORE_VM_FIXED2775_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2775E350g,
	"SUBCORE_VM_FIXED2800_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2800E350g,
	"SUBCORE_VM_FIXED2850_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2850E350g,
	"SUBCORE_VM_FIXED2875_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2875E350g,
	"SUBCORE_VM_FIXED2900_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2900E350g,
	"SUBCORE_VM_FIXED2925_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2925E350g,
	"SUBCORE_VM_FIXED2950_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2950E350g,
	"SUBCORE_VM_FIXED2975_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2975E350g,
	"SUBCORE_VM_FIXED3000_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3000E350g,
	"SUBCORE_VM_FIXED3025_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3025E350g,
	"SUBCORE_VM_FIXED3050_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3050E350g,
	"SUBCORE_VM_FIXED3075_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3075E350g,
	"SUBCORE_VM_FIXED3100_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3100E350g,
	"SUBCORE_VM_FIXED3125_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3125E350g,
	"SUBCORE_VM_FIXED3150_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3150E350g,
	"SUBCORE_VM_FIXED3200_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3200E350g,
	"SUBCORE_VM_FIXED3225_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3225E350g,
	"SUBCORE_VM_FIXED3250_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3250E350g,
	"SUBCORE_VM_FIXED3300_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3300E350g,
	"SUBCORE_VM_FIXED3325_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3325E350g,
	"SUBCORE_VM_FIXED3375_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3375E350g,
	"SUBCORE_VM_FIXED3400_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3400E350g,
	"SUBCORE_VM_FIXED3450_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3450E350g,
	"SUBCORE_VM_FIXED3500_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3500E350g,
	"SUBCORE_VM_FIXED3525_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3525E350g,
	"SUBCORE_VM_FIXED3575_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3575E350g,
	"SUBCORE_VM_FIXED3600_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3600E350g,
	"SUBCORE_VM_FIXED3625_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3625E350g,
	"SUBCORE_VM_FIXED3675_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3675E350g,
	"SUBCORE_VM_FIXED3700_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3700E350g,
	"SUBCORE_VM_FIXED3750_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3750E350g,
	"SUBCORE_VM_FIXED3800_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3800E350g,
	"SUBCORE_VM_FIXED3825_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3825E350g,
	"SUBCORE_VM_FIXED3850_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3850E350g,
	"SUBCORE_VM_FIXED3875_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3875E350g,
	"SUBCORE_VM_FIXED3900_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3900E350g,
	"SUBCORE_VM_FIXED3975_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3975E350g,
	"SUBCORE_VM_FIXED4000_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4000E350g,
	"SUBCORE_VM_FIXED4025_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4025E350g,
	"SUBCORE_VM_FIXED4050_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4050E350g,
	"SUBCORE_VM_FIXED4100_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4100E350g,
	"SUBCORE_VM_FIXED4125_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4125E350g,
	"SUBCORE_VM_FIXED4200_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4200E350g,
	"SUBCORE_VM_FIXED4225_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4225E350g,
	"SUBCORE_VM_FIXED4250_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4250E350g,
	"SUBCORE_VM_FIXED4275_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4275E350g,
	"SUBCORE_VM_FIXED4300_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4300E350g,
	"SUBCORE_VM_FIXED4350_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4350E350g,
	"SUBCORE_VM_FIXED4375_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4375E350g,
	"SUBCORE_VM_FIXED4400_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4400E350g,
	"SUBCORE_VM_FIXED4425_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4425E350g,
	"SUBCORE_VM_FIXED4500_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4500E350g,
	"SUBCORE_VM_FIXED4550_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4550E350g,
	"SUBCORE_VM_FIXED4575_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4575E350g,
	"SUBCORE_VM_FIXED4600_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4600E350g,
	"SUBCORE_VM_FIXED4625_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4625E350g,
	"SUBCORE_VM_FIXED4650_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4650E350g,
	"SUBCORE_VM_FIXED4675_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4675E350g,
	"SUBCORE_VM_FIXED4700_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4700E350g,
	"SUBCORE_VM_FIXED4725_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4725E350g,
	"SUBCORE_VM_FIXED4750_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4750E350g,
	"SUBCORE_VM_FIXED4800_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4800E350g,
	"SUBCORE_VM_FIXED4875_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4875E350g,
	"SUBCORE_VM_FIXED4900_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4900E350g,
	"SUBCORE_VM_FIXED4950_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4950E350g,
	"SUBCORE_VM_FIXED5000_E3_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed5000E350g,
	"SUBCORE_VM_FIXED0025_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0025E450g,
	"SUBCORE_VM_FIXED0050_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0050E450g,
	"SUBCORE_VM_FIXED0075_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0075E450g,
	"SUBCORE_VM_FIXED0100_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0100E450g,
	"SUBCORE_VM_FIXED0125_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0125E450g,
	"SUBCORE_VM_FIXED0150_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0150E450g,
	"SUBCORE_VM_FIXED0175_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0175E450g,
	"SUBCORE_VM_FIXED0200_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0200E450g,
	"SUBCORE_VM_FIXED0225_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0225E450g,
	"SUBCORE_VM_FIXED0250_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0250E450g,
	"SUBCORE_VM_FIXED0275_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0275E450g,
	"SUBCORE_VM_FIXED0300_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0300E450g,
	"SUBCORE_VM_FIXED0325_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0325E450g,
	"SUBCORE_VM_FIXED0350_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0350E450g,
	"SUBCORE_VM_FIXED0375_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0375E450g,
	"SUBCORE_VM_FIXED0400_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0400E450g,
	"SUBCORE_VM_FIXED0425_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0425E450g,
	"SUBCORE_VM_FIXED0450_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0450E450g,
	"SUBCORE_VM_FIXED0475_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0475E450g,
	"SUBCORE_VM_FIXED0500_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0500E450g,
	"SUBCORE_VM_FIXED0525_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0525E450g,
	"SUBCORE_VM_FIXED0550_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0550E450g,
	"SUBCORE_VM_FIXED0575_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0575E450g,
	"SUBCORE_VM_FIXED0600_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0600E450g,
	"SUBCORE_VM_FIXED0625_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0625E450g,
	"SUBCORE_VM_FIXED0650_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0650E450g,
	"SUBCORE_VM_FIXED0675_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0675E450g,
	"SUBCORE_VM_FIXED0700_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0700E450g,
	"SUBCORE_VM_FIXED0725_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0725E450g,
	"SUBCORE_VM_FIXED0750_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0750E450g,
	"SUBCORE_VM_FIXED0775_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0775E450g,
	"SUBCORE_VM_FIXED0800_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0800E450g,
	"SUBCORE_VM_FIXED0825_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0825E450g,
	"SUBCORE_VM_FIXED0850_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0850E450g,
	"SUBCORE_VM_FIXED0875_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0875E450g,
	"SUBCORE_VM_FIXED0900_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0900E450g,
	"SUBCORE_VM_FIXED0925_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0925E450g,
	"SUBCORE_VM_FIXED0950_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0950E450g,
	"SUBCORE_VM_FIXED0975_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0975E450g,
	"SUBCORE_VM_FIXED1000_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1000E450g,
	"SUBCORE_VM_FIXED1025_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1025E450g,
	"SUBCORE_VM_FIXED1050_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1050E450g,
	"SUBCORE_VM_FIXED1075_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1075E450g,
	"SUBCORE_VM_FIXED1100_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1100E450g,
	"SUBCORE_VM_FIXED1125_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1125E450g,
	"SUBCORE_VM_FIXED1150_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1150E450g,
	"SUBCORE_VM_FIXED1175_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1175E450g,
	"SUBCORE_VM_FIXED1200_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1200E450g,
	"SUBCORE_VM_FIXED1225_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1225E450g,
	"SUBCORE_VM_FIXED1250_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1250E450g,
	"SUBCORE_VM_FIXED1275_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1275E450g,
	"SUBCORE_VM_FIXED1300_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1300E450g,
	"SUBCORE_VM_FIXED1325_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1325E450g,
	"SUBCORE_VM_FIXED1350_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1350E450g,
	"SUBCORE_VM_FIXED1375_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1375E450g,
	"SUBCORE_VM_FIXED1400_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1400E450g,
	"SUBCORE_VM_FIXED1425_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1425E450g,
	"SUBCORE_VM_FIXED1450_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1450E450g,
	"SUBCORE_VM_FIXED1475_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1475E450g,
	"SUBCORE_VM_FIXED1500_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1500E450g,
	"SUBCORE_VM_FIXED1525_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1525E450g,
	"SUBCORE_VM_FIXED1550_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1550E450g,
	"SUBCORE_VM_FIXED1575_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1575E450g,
	"SUBCORE_VM_FIXED1600_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1600E450g,
	"SUBCORE_VM_FIXED1625_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1625E450g,
	"SUBCORE_VM_FIXED1650_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1650E450g,
	"SUBCORE_VM_FIXED1700_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1700E450g,
	"SUBCORE_VM_FIXED1725_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1725E450g,
	"SUBCORE_VM_FIXED1750_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1750E450g,
	"SUBCORE_VM_FIXED1800_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1800E450g,
	"SUBCORE_VM_FIXED1850_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1850E450g,
	"SUBCORE_VM_FIXED1875_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1875E450g,
	"SUBCORE_VM_FIXED1900_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1900E450g,
	"SUBCORE_VM_FIXED1925_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1925E450g,
	"SUBCORE_VM_FIXED1950_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1950E450g,
	"SUBCORE_VM_FIXED2000_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2000E450g,
	"SUBCORE_VM_FIXED2025_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2025E450g,
	"SUBCORE_VM_FIXED2050_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2050E450g,
	"SUBCORE_VM_FIXED2100_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2100E450g,
	"SUBCORE_VM_FIXED2125_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2125E450g,
	"SUBCORE_VM_FIXED2150_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2150E450g,
	"SUBCORE_VM_FIXED2175_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2175E450g,
	"SUBCORE_VM_FIXED2200_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2200E450g,
	"SUBCORE_VM_FIXED2250_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2250E450g,
	"SUBCORE_VM_FIXED2275_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2275E450g,
	"SUBCORE_VM_FIXED2300_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2300E450g,
	"SUBCORE_VM_FIXED2325_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2325E450g,
	"SUBCORE_VM_FIXED2350_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2350E450g,
	"SUBCORE_VM_FIXED2375_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2375E450g,
	"SUBCORE_VM_FIXED2400_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2400E450g,
	"SUBCORE_VM_FIXED2450_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2450E450g,
	"SUBCORE_VM_FIXED2475_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2475E450g,
	"SUBCORE_VM_FIXED2500_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2500E450g,
	"SUBCORE_VM_FIXED2550_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2550E450g,
	"SUBCORE_VM_FIXED2600_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2600E450g,
	"SUBCORE_VM_FIXED2625_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2625E450g,
	"SUBCORE_VM_FIXED2650_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2650E450g,
	"SUBCORE_VM_FIXED2700_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2700E450g,
	"SUBCORE_VM_FIXED2750_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2750E450g,
	"SUBCORE_VM_FIXED2775_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2775E450g,
	"SUBCORE_VM_FIXED2800_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2800E450g,
	"SUBCORE_VM_FIXED2850_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2850E450g,
	"SUBCORE_VM_FIXED2875_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2875E450g,
	"SUBCORE_VM_FIXED2900_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2900E450g,
	"SUBCORE_VM_FIXED2925_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2925E450g,
	"SUBCORE_VM_FIXED2950_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2950E450g,
	"SUBCORE_VM_FIXED2975_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2975E450g,
	"SUBCORE_VM_FIXED3000_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3000E450g,
	"SUBCORE_VM_FIXED3025_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3025E450g,
	"SUBCORE_VM_FIXED3050_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3050E450g,
	"SUBCORE_VM_FIXED3075_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3075E450g,
	"SUBCORE_VM_FIXED3100_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3100E450g,
	"SUBCORE_VM_FIXED3125_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3125E450g,
	"SUBCORE_VM_FIXED3150_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3150E450g,
	"SUBCORE_VM_FIXED3200_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3200E450g,
	"SUBCORE_VM_FIXED3225_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3225E450g,
	"SUBCORE_VM_FIXED3250_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3250E450g,
	"SUBCORE_VM_FIXED3300_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3300E450g,
	"SUBCORE_VM_FIXED3325_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3325E450g,
	"SUBCORE_VM_FIXED3375_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3375E450g,
	"SUBCORE_VM_FIXED3400_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3400E450g,
	"SUBCORE_VM_FIXED3450_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3450E450g,
	"SUBCORE_VM_FIXED3500_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3500E450g,
	"SUBCORE_VM_FIXED3525_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3525E450g,
	"SUBCORE_VM_FIXED3575_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3575E450g,
	"SUBCORE_VM_FIXED3600_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3600E450g,
	"SUBCORE_VM_FIXED3625_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3625E450g,
	"SUBCORE_VM_FIXED3675_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3675E450g,
	"SUBCORE_VM_FIXED3700_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3700E450g,
	"SUBCORE_VM_FIXED3750_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3750E450g,
	"SUBCORE_VM_FIXED3800_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3800E450g,
	"SUBCORE_VM_FIXED3825_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3825E450g,
	"SUBCORE_VM_FIXED3850_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3850E450g,
	"SUBCORE_VM_FIXED3875_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3875E450g,
	"SUBCORE_VM_FIXED3900_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3900E450g,
	"SUBCORE_VM_FIXED3975_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3975E450g,
	"SUBCORE_VM_FIXED4000_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4000E450g,
	"SUBCORE_VM_FIXED4025_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4025E450g,
	"SUBCORE_VM_FIXED4050_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4050E450g,
	"SUBCORE_VM_FIXED4100_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4100E450g,
	"SUBCORE_VM_FIXED4125_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4125E450g,
	"SUBCORE_VM_FIXED4200_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4200E450g,
	"SUBCORE_VM_FIXED4225_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4225E450g,
	"SUBCORE_VM_FIXED4250_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4250E450g,
	"SUBCORE_VM_FIXED4275_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4275E450g,
	"SUBCORE_VM_FIXED4300_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4300E450g,
	"SUBCORE_VM_FIXED4350_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4350E450g,
	"SUBCORE_VM_FIXED4375_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4375E450g,
	"SUBCORE_VM_FIXED4400_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4400E450g,
	"SUBCORE_VM_FIXED4425_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4425E450g,
	"SUBCORE_VM_FIXED4500_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4500E450g,
	"SUBCORE_VM_FIXED4550_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4550E450g,
	"SUBCORE_VM_FIXED4575_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4575E450g,
	"SUBCORE_VM_FIXED4600_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4600E450g,
	"SUBCORE_VM_FIXED4625_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4625E450g,
	"SUBCORE_VM_FIXED4650_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4650E450g,
	"SUBCORE_VM_FIXED4675_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4675E450g,
	"SUBCORE_VM_FIXED4700_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4700E450g,
	"SUBCORE_VM_FIXED4725_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4725E450g,
	"SUBCORE_VM_FIXED4750_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4750E450g,
	"SUBCORE_VM_FIXED4800_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4800E450g,
	"SUBCORE_VM_FIXED4875_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4875E450g,
	"SUBCORE_VM_FIXED4900_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4900E450g,
	"SUBCORE_VM_FIXED4950_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4950E450g,
	"SUBCORE_VM_FIXED5000_E4_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed5000E450g,
	"SUBCORE_VM_FIXED0020_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0020A150g,
	"SUBCORE_VM_FIXED0040_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0040A150g,
	"SUBCORE_VM_FIXED0060_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0060A150g,
	"SUBCORE_VM_FIXED0080_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0080A150g,
	"SUBCORE_VM_FIXED0100_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0100A150g,
	"SUBCORE_VM_FIXED0120_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0120A150g,
	"SUBCORE_VM_FIXED0140_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0140A150g,
	"SUBCORE_VM_FIXED0160_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0160A150g,
	"SUBCORE_VM_FIXED0180_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0180A150g,
	"SUBCORE_VM_FIXED0200_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0200A150g,
	"SUBCORE_VM_FIXED0220_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0220A150g,
	"SUBCORE_VM_FIXED0240_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0240A150g,
	"SUBCORE_VM_FIXED0260_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0260A150g,
	"SUBCORE_VM_FIXED0280_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0280A150g,
	"SUBCORE_VM_FIXED0300_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0300A150g,
	"SUBCORE_VM_FIXED0320_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0320A150g,
	"SUBCORE_VM_FIXED0340_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0340A150g,
	"SUBCORE_VM_FIXED0360_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0360A150g,
	"SUBCORE_VM_FIXED0380_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0380A150g,
	"SUBCORE_VM_FIXED0400_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0400A150g,
	"SUBCORE_VM_FIXED0420_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0420A150g,
	"SUBCORE_VM_FIXED0440_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0440A150g,
	"SUBCORE_VM_FIXED0460_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0460A150g,
	"SUBCORE_VM_FIXED0480_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0480A150g,
	"SUBCORE_VM_FIXED0500_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0500A150g,
	"SUBCORE_VM_FIXED0520_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0520A150g,
	"SUBCORE_VM_FIXED0540_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0540A150g,
	"SUBCORE_VM_FIXED0560_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0560A150g,
	"SUBCORE_VM_FIXED0580_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0580A150g,
	"SUBCORE_VM_FIXED0600_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0600A150g,
	"SUBCORE_VM_FIXED0620_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0620A150g,
	"SUBCORE_VM_FIXED0640_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0640A150g,
	"SUBCORE_VM_FIXED0660_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0660A150g,
	"SUBCORE_VM_FIXED0680_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0680A150g,
	"SUBCORE_VM_FIXED0700_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0700A150g,
	"SUBCORE_VM_FIXED0720_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0720A150g,
	"SUBCORE_VM_FIXED0740_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0740A150g,
	"SUBCORE_VM_FIXED0760_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0760A150g,
	"SUBCORE_VM_FIXED0780_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0780A150g,
	"SUBCORE_VM_FIXED0800_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0800A150g,
	"SUBCORE_VM_FIXED0820_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0820A150g,
	"SUBCORE_VM_FIXED0840_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0840A150g,
	"SUBCORE_VM_FIXED0860_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0860A150g,
	"SUBCORE_VM_FIXED0880_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0880A150g,
	"SUBCORE_VM_FIXED0900_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0900A150g,
	"SUBCORE_VM_FIXED0920_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0920A150g,
	"SUBCORE_VM_FIXED0940_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0940A150g,
	"SUBCORE_VM_FIXED0960_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0960A150g,
	"SUBCORE_VM_FIXED0980_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0980A150g,
	"SUBCORE_VM_FIXED1000_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1000A150g,
	"SUBCORE_VM_FIXED1020_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1020A150g,
	"SUBCORE_VM_FIXED1040_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1040A150g,
	"SUBCORE_VM_FIXED1060_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1060A150g,
	"SUBCORE_VM_FIXED1080_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1080A150g,
	"SUBCORE_VM_FIXED1100_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1100A150g,
	"SUBCORE_VM_FIXED1120_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1120A150g,
	"SUBCORE_VM_FIXED1140_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1140A150g,
	"SUBCORE_VM_FIXED1160_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1160A150g,
	"SUBCORE_VM_FIXED1180_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1180A150g,
	"SUBCORE_VM_FIXED1200_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1200A150g,
	"SUBCORE_VM_FIXED1220_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1220A150g,
	"SUBCORE_VM_FIXED1240_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1240A150g,
	"SUBCORE_VM_FIXED1260_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1260A150g,
	"SUBCORE_VM_FIXED1280_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1280A150g,
	"SUBCORE_VM_FIXED1300_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1300A150g,
	"SUBCORE_VM_FIXED1320_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1320A150g,
	"SUBCORE_VM_FIXED1340_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1340A150g,
	"SUBCORE_VM_FIXED1360_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1360A150g,
	"SUBCORE_VM_FIXED1380_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1380A150g,
	"SUBCORE_VM_FIXED1400_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1400A150g,
	"SUBCORE_VM_FIXED1420_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1420A150g,
	"SUBCORE_VM_FIXED1440_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1440A150g,
	"SUBCORE_VM_FIXED1460_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1460A150g,
	"SUBCORE_VM_FIXED1480_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1480A150g,
	"SUBCORE_VM_FIXED1500_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1500A150g,
	"SUBCORE_VM_FIXED1520_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1520A150g,
	"SUBCORE_VM_FIXED1540_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1540A150g,
	"SUBCORE_VM_FIXED1560_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1560A150g,
	"SUBCORE_VM_FIXED1580_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1580A150g,
	"SUBCORE_VM_FIXED1600_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1600A150g,
	"SUBCORE_VM_FIXED1620_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1620A150g,
	"SUBCORE_VM_FIXED1640_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1640A150g,
	"SUBCORE_VM_FIXED1660_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1660A150g,
	"SUBCORE_VM_FIXED1680_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1680A150g,
	"SUBCORE_VM_FIXED1700_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1700A150g,
	"SUBCORE_VM_FIXED1720_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1720A150g,
	"SUBCORE_VM_FIXED1740_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1740A150g,
	"SUBCORE_VM_FIXED1760_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1760A150g,
	"SUBCORE_VM_FIXED1780_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1780A150g,
	"SUBCORE_VM_FIXED1800_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1800A150g,
	"SUBCORE_VM_FIXED1820_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1820A150g,
	"SUBCORE_VM_FIXED1840_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1840A150g,
	"SUBCORE_VM_FIXED1860_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1860A150g,
	"SUBCORE_VM_FIXED1880_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1880A150g,
	"SUBCORE_VM_FIXED1900_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1900A150g,
	"SUBCORE_VM_FIXED1920_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1920A150g,
	"SUBCORE_VM_FIXED1940_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1940A150g,
	"SUBCORE_VM_FIXED1960_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1960A150g,
	"SUBCORE_VM_FIXED1980_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1980A150g,
	"SUBCORE_VM_FIXED2000_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2000A150g,
	"SUBCORE_VM_FIXED2020_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2020A150g,
	"SUBCORE_VM_FIXED2040_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2040A150g,
	"SUBCORE_VM_FIXED2060_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2060A150g,
	"SUBCORE_VM_FIXED2080_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2080A150g,
	"SUBCORE_VM_FIXED2100_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2100A150g,
	"SUBCORE_VM_FIXED2120_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2120A150g,
	"SUBCORE_VM_FIXED2140_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2140A150g,
	"SUBCORE_VM_FIXED2160_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2160A150g,
	"SUBCORE_VM_FIXED2180_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2180A150g,
	"SUBCORE_VM_FIXED2200_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2200A150g,
	"SUBCORE_VM_FIXED2220_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2220A150g,
	"SUBCORE_VM_FIXED2240_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2240A150g,
	"SUBCORE_VM_FIXED2260_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2260A150g,
	"SUBCORE_VM_FIXED2280_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2280A150g,
	"SUBCORE_VM_FIXED2300_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2300A150g,
	"SUBCORE_VM_FIXED2320_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2320A150g,
	"SUBCORE_VM_FIXED2340_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2340A150g,
	"SUBCORE_VM_FIXED2360_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2360A150g,
	"SUBCORE_VM_FIXED2380_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2380A150g,
	"SUBCORE_VM_FIXED2400_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2400A150g,
	"SUBCORE_VM_FIXED2420_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2420A150g,
	"SUBCORE_VM_FIXED2440_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2440A150g,
	"SUBCORE_VM_FIXED2460_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2460A150g,
	"SUBCORE_VM_FIXED2480_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2480A150g,
	"SUBCORE_VM_FIXED2500_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2500A150g,
	"SUBCORE_VM_FIXED2520_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2520A150g,
	"SUBCORE_VM_FIXED2540_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2540A150g,
	"SUBCORE_VM_FIXED2560_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2560A150g,
	"SUBCORE_VM_FIXED2580_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2580A150g,
	"SUBCORE_VM_FIXED2600_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2600A150g,
	"SUBCORE_VM_FIXED2620_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2620A150g,
	"SUBCORE_VM_FIXED2640_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2640A150g,
	"SUBCORE_VM_FIXED2660_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2660A150g,
	"SUBCORE_VM_FIXED2680_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2680A150g,
	"SUBCORE_VM_FIXED2700_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2700A150g,
	"SUBCORE_VM_FIXED2720_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2720A150g,
	"SUBCORE_VM_FIXED2740_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2740A150g,
	"SUBCORE_VM_FIXED2760_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2760A150g,
	"SUBCORE_VM_FIXED2780_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2780A150g,
	"SUBCORE_VM_FIXED2800_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2800A150g,
	"SUBCORE_VM_FIXED2820_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2820A150g,
	"SUBCORE_VM_FIXED2840_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2840A150g,
	"SUBCORE_VM_FIXED2860_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2860A150g,
	"SUBCORE_VM_FIXED2880_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2880A150g,
	"SUBCORE_VM_FIXED2900_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2900A150g,
	"SUBCORE_VM_FIXED2920_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2920A150g,
	"SUBCORE_VM_FIXED2940_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2940A150g,
	"SUBCORE_VM_FIXED2960_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2960A150g,
	"SUBCORE_VM_FIXED2980_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2980A150g,
	"SUBCORE_VM_FIXED3000_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3000A150g,
	"SUBCORE_VM_FIXED3020_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3020A150g,
	"SUBCORE_VM_FIXED3040_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3040A150g,
	"SUBCORE_VM_FIXED3060_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3060A150g,
	"SUBCORE_VM_FIXED3080_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3080A150g,
	"SUBCORE_VM_FIXED3100_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3100A150g,
	"SUBCORE_VM_FIXED3120_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3120A150g,
	"SUBCORE_VM_FIXED3140_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3140A150g,
	"SUBCORE_VM_FIXED3160_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3160A150g,
	"SUBCORE_VM_FIXED3180_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3180A150g,
	"SUBCORE_VM_FIXED3200_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3200A150g,
	"SUBCORE_VM_FIXED3220_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3220A150g,
	"SUBCORE_VM_FIXED3240_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3240A150g,
	"SUBCORE_VM_FIXED3260_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3260A150g,
	"SUBCORE_VM_FIXED3280_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3280A150g,
	"SUBCORE_VM_FIXED3300_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3300A150g,
	"SUBCORE_VM_FIXED3320_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3320A150g,
	"SUBCORE_VM_FIXED3340_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3340A150g,
	"SUBCORE_VM_FIXED3360_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3360A150g,
	"SUBCORE_VM_FIXED3380_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3380A150g,
	"SUBCORE_VM_FIXED3400_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3400A150g,
	"SUBCORE_VM_FIXED3420_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3420A150g,
	"SUBCORE_VM_FIXED3440_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3440A150g,
	"SUBCORE_VM_FIXED3460_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3460A150g,
	"SUBCORE_VM_FIXED3480_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3480A150g,
	"SUBCORE_VM_FIXED3500_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3500A150g,
	"SUBCORE_VM_FIXED3520_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3520A150g,
	"SUBCORE_VM_FIXED3540_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3540A150g,
	"SUBCORE_VM_FIXED3560_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3560A150g,
	"SUBCORE_VM_FIXED3580_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3580A150g,
	"SUBCORE_VM_FIXED3600_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3600A150g,
	"SUBCORE_VM_FIXED3620_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3620A150g,
	"SUBCORE_VM_FIXED3640_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3640A150g,
	"SUBCORE_VM_FIXED3660_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3660A150g,
	"SUBCORE_VM_FIXED3680_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3680A150g,
	"SUBCORE_VM_FIXED3700_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3700A150g,
	"SUBCORE_VM_FIXED3720_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3720A150g,
	"SUBCORE_VM_FIXED3740_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3740A150g,
	"SUBCORE_VM_FIXED3760_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3760A150g,
	"SUBCORE_VM_FIXED3780_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3780A150g,
	"SUBCORE_VM_FIXED3800_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3800A150g,
	"SUBCORE_VM_FIXED3820_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3820A150g,
	"SUBCORE_VM_FIXED3840_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3840A150g,
	"SUBCORE_VM_FIXED3860_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3860A150g,
	"SUBCORE_VM_FIXED3880_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3880A150g,
	"SUBCORE_VM_FIXED3900_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3900A150g,
	"SUBCORE_VM_FIXED3920_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3920A150g,
	"SUBCORE_VM_FIXED3940_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3940A150g,
	"SUBCORE_VM_FIXED3960_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3960A150g,
	"SUBCORE_VM_FIXED3980_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3980A150g,
	"SUBCORE_VM_FIXED4000_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4000A150g,
	"SUBCORE_VM_FIXED4020_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4020A150g,
	"SUBCORE_VM_FIXED4040_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4040A150g,
	"SUBCORE_VM_FIXED4060_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4060A150g,
	"SUBCORE_VM_FIXED4080_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4080A150g,
	"SUBCORE_VM_FIXED4100_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4100A150g,
	"SUBCORE_VM_FIXED4120_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4120A150g,
	"SUBCORE_VM_FIXED4140_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4140A150g,
	"SUBCORE_VM_FIXED4160_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4160A150g,
	"SUBCORE_VM_FIXED4180_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4180A150g,
	"SUBCORE_VM_FIXED4200_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4200A150g,
	"SUBCORE_VM_FIXED4220_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4220A150g,
	"SUBCORE_VM_FIXED4240_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4240A150g,
	"SUBCORE_VM_FIXED4260_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4260A150g,
	"SUBCORE_VM_FIXED4280_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4280A150g,
	"SUBCORE_VM_FIXED4300_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4300A150g,
	"SUBCORE_VM_FIXED4320_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4320A150g,
	"SUBCORE_VM_FIXED4340_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4340A150g,
	"SUBCORE_VM_FIXED4360_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4360A150g,
	"SUBCORE_VM_FIXED4380_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4380A150g,
	"SUBCORE_VM_FIXED4400_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4400A150g,
	"SUBCORE_VM_FIXED4420_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4420A150g,
	"SUBCORE_VM_FIXED4440_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4440A150g,
	"SUBCORE_VM_FIXED4460_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4460A150g,
	"SUBCORE_VM_FIXED4480_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4480A150g,
	"SUBCORE_VM_FIXED4500_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4500A150g,
	"SUBCORE_VM_FIXED4520_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4520A150g,
	"SUBCORE_VM_FIXED4540_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4540A150g,
	"SUBCORE_VM_FIXED4560_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4560A150g,
	"SUBCORE_VM_FIXED4580_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4580A150g,
	"SUBCORE_VM_FIXED4600_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4600A150g,
	"SUBCORE_VM_FIXED4620_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4620A150g,
	"SUBCORE_VM_FIXED4640_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4640A150g,
	"SUBCORE_VM_FIXED4660_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4660A150g,
	"SUBCORE_VM_FIXED4680_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4680A150g,
	"SUBCORE_VM_FIXED4700_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4700A150g,
	"SUBCORE_VM_FIXED4720_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4720A150g,
	"SUBCORE_VM_FIXED4740_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4740A150g,
	"SUBCORE_VM_FIXED4760_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4760A150g,
	"SUBCORE_VM_FIXED4780_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4780A150g,
	"SUBCORE_VM_FIXED4800_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4800A150g,
	"SUBCORE_VM_FIXED4820_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4820A150g,
	"SUBCORE_VM_FIXED4840_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4840A150g,
	"SUBCORE_VM_FIXED4860_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4860A150g,
	"SUBCORE_VM_FIXED4880_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4880A150g,
	"SUBCORE_VM_FIXED4900_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4900A150g,
	"SUBCORE_VM_FIXED4920_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4920A150g,
	"SUBCORE_VM_FIXED4940_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4940A150g,
	"SUBCORE_VM_FIXED4960_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4960A150g,
	"SUBCORE_VM_FIXED4980_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4980A150g,
	"SUBCORE_VM_FIXED5000_A1_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed5000A150g,
	"SUBCORE_VM_FIXED0090_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0090X950g,
	"SUBCORE_VM_FIXED0180_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0180X950g,
	"SUBCORE_VM_FIXED0270_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0270X950g,
	"SUBCORE_VM_FIXED0360_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0360X950g,
	"SUBCORE_VM_FIXED0450_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0450X950g,
	"SUBCORE_VM_FIXED0540_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0540X950g,
	"SUBCORE_VM_FIXED0630_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0630X950g,
	"SUBCORE_VM_FIXED0720_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0720X950g,
	"SUBCORE_VM_FIXED0810_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0810X950g,
	"SUBCORE_VM_FIXED0900_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0900X950g,
	"SUBCORE_VM_FIXED0990_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed0990X950g,
	"SUBCORE_VM_FIXED1080_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1080X950g,
	"SUBCORE_VM_FIXED1170_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1170X950g,
	"SUBCORE_VM_FIXED1260_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1260X950g,
	"SUBCORE_VM_FIXED1350_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1350X950g,
	"SUBCORE_VM_FIXED1440_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1440X950g,
	"SUBCORE_VM_FIXED1530_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1530X950g,
	"SUBCORE_VM_FIXED1620_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1620X950g,
	"SUBCORE_VM_FIXED1710_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1710X950g,
	"SUBCORE_VM_FIXED1800_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1800X950g,
	"SUBCORE_VM_FIXED1890_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1890X950g,
	"SUBCORE_VM_FIXED1980_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed1980X950g,
	"SUBCORE_VM_FIXED2070_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2070X950g,
	"SUBCORE_VM_FIXED2160_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2160X950g,
	"SUBCORE_VM_FIXED2250_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2250X950g,
	"SUBCORE_VM_FIXED2340_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2340X950g,
	"SUBCORE_VM_FIXED2430_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2430X950g,
	"SUBCORE_VM_FIXED2520_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2520X950g,
	"SUBCORE_VM_FIXED2610_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2610X950g,
	"SUBCORE_VM_FIXED2700_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2700X950g,
	"SUBCORE_VM_FIXED2790_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2790X950g,
	"SUBCORE_VM_FIXED2880_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2880X950g,
	"SUBCORE_VM_FIXED2970_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed2970X950g,
	"SUBCORE_VM_FIXED3060_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3060X950g,
	"SUBCORE_VM_FIXED3150_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3150X950g,
	"SUBCORE_VM_FIXED3240_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3240X950g,
	"SUBCORE_VM_FIXED3330_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3330X950g,
	"SUBCORE_VM_FIXED3420_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3420X950g,
	"SUBCORE_VM_FIXED3510_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3510X950g,
	"SUBCORE_VM_FIXED3600_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3600X950g,
	"SUBCORE_VM_FIXED3690_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3690X950g,
	"SUBCORE_VM_FIXED3780_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3780X950g,
	"SUBCORE_VM_FIXED3870_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3870X950g,
	"SUBCORE_VM_FIXED3960_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed3960X950g,
	"SUBCORE_VM_FIXED4050_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4050X950g,
	"SUBCORE_VM_FIXED4140_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4140X950g,
	"SUBCORE_VM_FIXED4230_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4230X950g,
	"SUBCORE_VM_FIXED4320_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4320X950g,
	"SUBCORE_VM_FIXED4410_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4410X950g,
	"SUBCORE_VM_FIXED4500_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4500X950g,
	"SUBCORE_VM_FIXED4590_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4590X950g,
	"SUBCORE_VM_FIXED4680_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4680X950g,
	"SUBCORE_VM_FIXED4770_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4770X950g,
	"SUBCORE_VM_FIXED4860_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4860X950g,
	"SUBCORE_VM_FIXED4950_X9_50G":  InternalVnicAttachmentVnicShapeSubcoreVmFixed4950X950g,
	"DYNAMIC_A1_50G":               InternalVnicAttachmentVnicShapeDynamicA150g,
	"FIXED0040_A1_50G":             InternalVnicAttachmentVnicShapeFixed0040A150g,
	"FIXED0100_A1_50G":             InternalVnicAttachmentVnicShapeFixed0100A150g,
	"FIXED0200_A1_50G":             InternalVnicAttachmentVnicShapeFixed0200A150g,
	"FIXED0300_A1_50G":             InternalVnicAttachmentVnicShapeFixed0300A150g,
	"FIXED0400_A1_50G":             InternalVnicAttachmentVnicShapeFixed0400A150g,
	"FIXED0500_A1_50G":             InternalVnicAttachmentVnicShapeFixed0500A150g,
	"FIXED0600_A1_50G":             InternalVnicAttachmentVnicShapeFixed0600A150g,
	"FIXED0700_A1_50G":             InternalVnicAttachmentVnicShapeFixed0700A150g,
	"FIXED0800_A1_50G":             InternalVnicAttachmentVnicShapeFixed0800A150g,
	"FIXED0900_A1_50G":             InternalVnicAttachmentVnicShapeFixed0900A150g,
	"FIXED1000_A1_50G":             InternalVnicAttachmentVnicShapeFixed1000A150g,
	"FIXED1100_A1_50G":             InternalVnicAttachmentVnicShapeFixed1100A150g,
	"FIXED1200_A1_50G":             InternalVnicAttachmentVnicShapeFixed1200A150g,
	"FIXED1300_A1_50G":             InternalVnicAttachmentVnicShapeFixed1300A150g,
	"FIXED1400_A1_50G":             InternalVnicAttachmentVnicShapeFixed1400A150g,
	"FIXED1500_A1_50G":             InternalVnicAttachmentVnicShapeFixed1500A150g,
	"FIXED1600_A1_50G":             InternalVnicAttachmentVnicShapeFixed1600A150g,
	"FIXED1700_A1_50G":             InternalVnicAttachmentVnicShapeFixed1700A150g,
	"FIXED1800_A1_50G":             InternalVnicAttachmentVnicShapeFixed1800A150g,
	"FIXED1900_A1_50G":             InternalVnicAttachmentVnicShapeFixed1900A150g,
	"FIXED2000_A1_50G":             InternalVnicAttachmentVnicShapeFixed2000A150g,
	"FIXED2100_A1_50G":             InternalVnicAttachmentVnicShapeFixed2100A150g,
	"FIXED2200_A1_50G":             InternalVnicAttachmentVnicShapeFixed2200A150g,
	"FIXED2300_A1_50G":             InternalVnicAttachmentVnicShapeFixed2300A150g,
	"FIXED2400_A1_50G":             InternalVnicAttachmentVnicShapeFixed2400A150g,
	"FIXED2500_A1_50G":             InternalVnicAttachmentVnicShapeFixed2500A150g,
	"FIXED2600_A1_50G":             InternalVnicAttachmentVnicShapeFixed2600A150g,
	"FIXED2700_A1_50G":             InternalVnicAttachmentVnicShapeFixed2700A150g,
	"FIXED2800_A1_50G":             InternalVnicAttachmentVnicShapeFixed2800A150g,
	"FIXED2900_A1_50G":             InternalVnicAttachmentVnicShapeFixed2900A150g,
	"FIXED3000_A1_50G":             InternalVnicAttachmentVnicShapeFixed3000A150g,
	"FIXED3100_A1_50G":             InternalVnicAttachmentVnicShapeFixed3100A150g,
	"FIXED3200_A1_50G":             InternalVnicAttachmentVnicShapeFixed3200A150g,
	"FIXED3300_A1_50G":             InternalVnicAttachmentVnicShapeFixed3300A150g,
	"FIXED3400_A1_50G":             InternalVnicAttachmentVnicShapeFixed3400A150g,
	"FIXED3500_A1_50G":             InternalVnicAttachmentVnicShapeFixed3500A150g,
	"FIXED3600_A1_50G":             InternalVnicAttachmentVnicShapeFixed3600A150g,
	"FIXED3700_A1_50G":             InternalVnicAttachmentVnicShapeFixed3700A150g,
	"FIXED3800_A1_50G":             InternalVnicAttachmentVnicShapeFixed3800A150g,
	"FIXED3900_A1_50G":             InternalVnicAttachmentVnicShapeFixed3900A150g,
	"FIXED4000_A1_50G":             InternalVnicAttachmentVnicShapeFixed4000A150g,
	"ENTIREHOST_A1_50G":            InternalVnicAttachmentVnicShapeEntirehostA150g,
	"DYNAMIC_X9_50G":               InternalVnicAttachmentVnicShapeDynamicX950g,
	"FIXED0040_X9_50G":             InternalVnicAttachmentVnicShapeFixed0040X950g,
	"FIXED0400_X9_50G":             InternalVnicAttachmentVnicShapeFixed0400X950g,
	"FIXED0800_X9_50G":             InternalVnicAttachmentVnicShapeFixed0800X950g,
	"FIXED1200_X9_50G":             InternalVnicAttachmentVnicShapeFixed1200X950g,
	"FIXED1600_X9_50G":             InternalVnicAttachmentVnicShapeFixed1600X950g,
	"FIXED2000_X9_50G":             InternalVnicAttachmentVnicShapeFixed2000X950g,
	"FIXED2400_X9_50G":             InternalVnicAttachmentVnicShapeFixed2400X950g,
	"FIXED2800_X9_50G":             InternalVnicAttachmentVnicShapeFixed2800X950g,
	"FIXED3200_X9_50G":             InternalVnicAttachmentVnicShapeFixed3200X950g,
	"FIXED3600_X9_50G":             InternalVnicAttachmentVnicShapeFixed3600X950g,
	"FIXED4000_X9_50G":             InternalVnicAttachmentVnicShapeFixed4000X950g,
	"STANDARD_VM_FIXED0100_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed0100X950g,
	"STANDARD_VM_FIXED0200_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed0200X950g,
	"STANDARD_VM_FIXED0300_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed0300X950g,
	"STANDARD_VM_FIXED0400_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed0400X950g,
	"STANDARD_VM_FIXED0500_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed0500X950g,
	"STANDARD_VM_FIXED0600_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed0600X950g,
	"STANDARD_VM_FIXED0700_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed0700X950g,
	"STANDARD_VM_FIXED0800_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed0800X950g,
	"STANDARD_VM_FIXED0900_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed0900X950g,
	"STANDARD_VM_FIXED1000_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed1000X950g,
	"STANDARD_VM_FIXED1100_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed1100X950g,
	"STANDARD_VM_FIXED1200_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed1200X950g,
	"STANDARD_VM_FIXED1300_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed1300X950g,
	"STANDARD_VM_FIXED1400_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed1400X950g,
	"STANDARD_VM_FIXED1500_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed1500X950g,
	"STANDARD_VM_FIXED1600_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed1600X950g,
	"STANDARD_VM_FIXED1700_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed1700X950g,
	"STANDARD_VM_FIXED1800_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed1800X950g,
	"STANDARD_VM_FIXED1900_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed1900X950g,
	"STANDARD_VM_FIXED2000_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed2000X950g,
	"STANDARD_VM_FIXED2100_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed2100X950g,
	"STANDARD_VM_FIXED2200_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed2200X950g,
	"STANDARD_VM_FIXED2300_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed2300X950g,
	"STANDARD_VM_FIXED2400_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed2400X950g,
	"STANDARD_VM_FIXED2500_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed2500X950g,
	"STANDARD_VM_FIXED2600_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed2600X950g,
	"STANDARD_VM_FIXED2700_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed2700X950g,
	"STANDARD_VM_FIXED2800_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed2800X950g,
	"STANDARD_VM_FIXED2900_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed2900X950g,
	"STANDARD_VM_FIXED3000_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed3000X950g,
	"STANDARD_VM_FIXED3100_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed3100X950g,
	"STANDARD_VM_FIXED3200_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed3200X950g,
	"STANDARD_VM_FIXED3300_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed3300X950g,
	"STANDARD_VM_FIXED3400_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed3400X950g,
	"STANDARD_VM_FIXED3500_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed3500X950g,
	"STANDARD_VM_FIXED3600_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed3600X950g,
	"STANDARD_VM_FIXED3700_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed3700X950g,
	"STANDARD_VM_FIXED3800_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed3800X950g,
	"STANDARD_VM_FIXED3900_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed3900X950g,
	"STANDARD_VM_FIXED4000_X9_50G": InternalVnicAttachmentVnicShapeStandardVmFixed4000X950g,
	"ENTIREHOST_X9_50G":            InternalVnicAttachmentVnicShapeEntirehostX950g,
}

// GetInternalVnicAttachmentVnicShapeEnumValues Enumerates the set of values for InternalVnicAttachmentVnicShapeEnum
func GetInternalVnicAttachmentVnicShapeEnumValues() []InternalVnicAttachmentVnicShapeEnum {
	values := make([]InternalVnicAttachmentVnicShapeEnum, 0)
	for _, v := range mappingInternalVnicAttachmentVnicShapeEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalVnicAttachmentVnicShapeEnumStringValues Enumerates the set of values in String for InternalVnicAttachmentVnicShapeEnum
func GetInternalVnicAttachmentVnicShapeEnumStringValues() []string {
	return []string{
		"DYNAMIC",
		"FIXED0040",
		"FIXED0060",
		"FIXED0060_PSM",
		"FIXED0100",
		"FIXED0120",
		"FIXED0120_2X",
		"FIXED0200",
		"FIXED0240",
		"FIXED0480",
		"ENTIREHOST",
		"DYNAMIC_25G",
		"FIXED0040_25G",
		"FIXED0100_25G",
		"FIXED0200_25G",
		"FIXED0400_25G",
		"FIXED0800_25G",
		"FIXED1600_25G",
		"FIXED2400_25G",
		"ENTIREHOST_25G",
		"DYNAMIC_E1_25G",
		"FIXED0040_E1_25G",
		"FIXED0070_E1_25G",
		"FIXED0140_E1_25G",
		"FIXED0280_E1_25G",
		"FIXED0560_E1_25G",
		"FIXED1120_E1_25G",
		"FIXED1680_E1_25G",
		"ENTIREHOST_E1_25G",
		"DYNAMIC_B1_25G",
		"FIXED0040_B1_25G",
		"FIXED0060_B1_25G",
		"FIXED0120_B1_25G",
		"FIXED0240_B1_25G",
		"FIXED0480_B1_25G",
		"FIXED0960_B1_25G",
		"ENTIREHOST_B1_25G",
		"MICRO_VM_FIXED0048_E1_25G",
		"MICRO_LB_FIXED0001_E1_25G",
		"VNICAAS_FIXED0200",
		"VNICAAS_FIXED0400",
		"VNICAAS_FIXED0700",
		"VNICAAS_NLB_APPROVED_10G",
		"VNICAAS_NLB_APPROVED_25G",
		"VNICAAS_TELESIS_25G",
		"VNICAAS_TELESIS_10G",
		"VNICAAS_AMBASSADOR_FIXED0100",
		"VNICAAS_PRIVATEDNS",
		"VNICAAS_FWAAS",
		"DYNAMIC_E3_50G",
		"FIXED0040_E3_50G",
		"FIXED0100_E3_50G",
		"FIXED0200_E3_50G",
		"FIXED0300_E3_50G",
		"FIXED0400_E3_50G",
		"FIXED0500_E3_50G",
		"FIXED0600_E3_50G",
		"FIXED0700_E3_50G",
		"FIXED0800_E3_50G",
		"FIXED0900_E3_50G",
		"FIXED1000_E3_50G",
		"FIXED1100_E3_50G",
		"FIXED1200_E3_50G",
		"FIXED1300_E3_50G",
		"FIXED1400_E3_50G",
		"FIXED1500_E3_50G",
		"FIXED1600_E3_50G",
		"FIXED1700_E3_50G",
		"FIXED1800_E3_50G",
		"FIXED1900_E3_50G",
		"FIXED2000_E3_50G",
		"FIXED2100_E3_50G",
		"FIXED2200_E3_50G",
		"FIXED2300_E3_50G",
		"FIXED2400_E3_50G",
		"FIXED2500_E3_50G",
		"FIXED2600_E3_50G",
		"FIXED2700_E3_50G",
		"FIXED2800_E3_50G",
		"FIXED2900_E3_50G",
		"FIXED3000_E3_50G",
		"FIXED3100_E3_50G",
		"FIXED3200_E3_50G",
		"FIXED3300_E3_50G",
		"FIXED3400_E3_50G",
		"FIXED3500_E3_50G",
		"FIXED3600_E3_50G",
		"FIXED3700_E3_50G",
		"FIXED3800_E3_50G",
		"FIXED3900_E3_50G",
		"FIXED4000_E3_50G",
		"ENTIREHOST_E3_50G",
		"DYNAMIC_E4_50G",
		"FIXED0040_E4_50G",
		"FIXED0100_E4_50G",
		"FIXED0200_E4_50G",
		"FIXED0300_E4_50G",
		"FIXED0400_E4_50G",
		"FIXED0500_E4_50G",
		"FIXED0600_E4_50G",
		"FIXED0700_E4_50G",
		"FIXED0800_E4_50G",
		"FIXED0900_E4_50G",
		"FIXED1000_E4_50G",
		"FIXED1100_E4_50G",
		"FIXED1200_E4_50G",
		"FIXED1300_E4_50G",
		"FIXED1400_E4_50G",
		"FIXED1500_E4_50G",
		"FIXED1600_E4_50G",
		"FIXED1700_E4_50G",
		"FIXED1800_E4_50G",
		"FIXED1900_E4_50G",
		"FIXED2000_E4_50G",
		"FIXED2100_E4_50G",
		"FIXED2200_E4_50G",
		"FIXED2300_E4_50G",
		"FIXED2400_E4_50G",
		"FIXED2500_E4_50G",
		"FIXED2600_E4_50G",
		"FIXED2700_E4_50G",
		"FIXED2800_E4_50G",
		"FIXED2900_E4_50G",
		"FIXED3000_E4_50G",
		"FIXED3100_E4_50G",
		"FIXED3200_E4_50G",
		"FIXED3300_E4_50G",
		"FIXED3400_E4_50G",
		"FIXED3500_E4_50G",
		"FIXED3600_E4_50G",
		"FIXED3700_E4_50G",
		"FIXED3800_E4_50G",
		"FIXED3900_E4_50G",
		"FIXED4000_E4_50G",
		"ENTIREHOST_E4_50G",
		"MICRO_VM_FIXED0050_E3_50G",
		"SUBCORE_VM_FIXED0025_E3_50G",
		"SUBCORE_VM_FIXED0050_E3_50G",
		"SUBCORE_VM_FIXED0075_E3_50G",
		"SUBCORE_VM_FIXED0100_E3_50G",
		"SUBCORE_VM_FIXED0125_E3_50G",
		"SUBCORE_VM_FIXED0150_E3_50G",
		"SUBCORE_VM_FIXED0175_E3_50G",
		"SUBCORE_VM_FIXED0200_E3_50G",
		"SUBCORE_VM_FIXED0225_E3_50G",
		"SUBCORE_VM_FIXED0250_E3_50G",
		"SUBCORE_VM_FIXED0275_E3_50G",
		"SUBCORE_VM_FIXED0300_E3_50G",
		"SUBCORE_VM_FIXED0325_E3_50G",
		"SUBCORE_VM_FIXED0350_E3_50G",
		"SUBCORE_VM_FIXED0375_E3_50G",
		"SUBCORE_VM_FIXED0400_E3_50G",
		"SUBCORE_VM_FIXED0425_E3_50G",
		"SUBCORE_VM_FIXED0450_E3_50G",
		"SUBCORE_VM_FIXED0475_E3_50G",
		"SUBCORE_VM_FIXED0500_E3_50G",
		"SUBCORE_VM_FIXED0525_E3_50G",
		"SUBCORE_VM_FIXED0550_E3_50G",
		"SUBCORE_VM_FIXED0575_E3_50G",
		"SUBCORE_VM_FIXED0600_E3_50G",
		"SUBCORE_VM_FIXED0625_E3_50G",
		"SUBCORE_VM_FIXED0650_E3_50G",
		"SUBCORE_VM_FIXED0675_E3_50G",
		"SUBCORE_VM_FIXED0700_E3_50G",
		"SUBCORE_VM_FIXED0725_E3_50G",
		"SUBCORE_VM_FIXED0750_E3_50G",
		"SUBCORE_VM_FIXED0775_E3_50G",
		"SUBCORE_VM_FIXED0800_E3_50G",
		"SUBCORE_VM_FIXED0825_E3_50G",
		"SUBCORE_VM_FIXED0850_E3_50G",
		"SUBCORE_VM_FIXED0875_E3_50G",
		"SUBCORE_VM_FIXED0900_E3_50G",
		"SUBCORE_VM_FIXED0925_E3_50G",
		"SUBCORE_VM_FIXED0950_E3_50G",
		"SUBCORE_VM_FIXED0975_E3_50G",
		"SUBCORE_VM_FIXED1000_E3_50G",
		"SUBCORE_VM_FIXED1025_E3_50G",
		"SUBCORE_VM_FIXED1050_E3_50G",
		"SUBCORE_VM_FIXED1075_E3_50G",
		"SUBCORE_VM_FIXED1100_E3_50G",
		"SUBCORE_VM_FIXED1125_E3_50G",
		"SUBCORE_VM_FIXED1150_E3_50G",
		"SUBCORE_VM_FIXED1175_E3_50G",
		"SUBCORE_VM_FIXED1200_E3_50G",
		"SUBCORE_VM_FIXED1225_E3_50G",
		"SUBCORE_VM_FIXED1250_E3_50G",
		"SUBCORE_VM_FIXED1275_E3_50G",
		"SUBCORE_VM_FIXED1300_E3_50G",
		"SUBCORE_VM_FIXED1325_E3_50G",
		"SUBCORE_VM_FIXED1350_E3_50G",
		"SUBCORE_VM_FIXED1375_E3_50G",
		"SUBCORE_VM_FIXED1400_E3_50G",
		"SUBCORE_VM_FIXED1425_E3_50G",
		"SUBCORE_VM_FIXED1450_E3_50G",
		"SUBCORE_VM_FIXED1475_E3_50G",
		"SUBCORE_VM_FIXED1500_E3_50G",
		"SUBCORE_VM_FIXED1525_E3_50G",
		"SUBCORE_VM_FIXED1550_E3_50G",
		"SUBCORE_VM_FIXED1575_E3_50G",
		"SUBCORE_VM_FIXED1600_E3_50G",
		"SUBCORE_VM_FIXED1625_E3_50G",
		"SUBCORE_VM_FIXED1650_E3_50G",
		"SUBCORE_VM_FIXED1700_E3_50G",
		"SUBCORE_VM_FIXED1725_E3_50G",
		"SUBCORE_VM_FIXED1750_E3_50G",
		"SUBCORE_VM_FIXED1800_E3_50G",
		"SUBCORE_VM_FIXED1850_E3_50G",
		"SUBCORE_VM_FIXED1875_E3_50G",
		"SUBCORE_VM_FIXED1900_E3_50G",
		"SUBCORE_VM_FIXED1925_E3_50G",
		"SUBCORE_VM_FIXED1950_E3_50G",
		"SUBCORE_VM_FIXED2000_E3_50G",
		"SUBCORE_VM_FIXED2025_E3_50G",
		"SUBCORE_VM_FIXED2050_E3_50G",
		"SUBCORE_VM_FIXED2100_E3_50G",
		"SUBCORE_VM_FIXED2125_E3_50G",
		"SUBCORE_VM_FIXED2150_E3_50G",
		"SUBCORE_VM_FIXED2175_E3_50G",
		"SUBCORE_VM_FIXED2200_E3_50G",
		"SUBCORE_VM_FIXED2250_E3_50G",
		"SUBCORE_VM_FIXED2275_E3_50G",
		"SUBCORE_VM_FIXED2300_E3_50G",
		"SUBCORE_VM_FIXED2325_E3_50G",
		"SUBCORE_VM_FIXED2350_E3_50G",
		"SUBCORE_VM_FIXED2375_E3_50G",
		"SUBCORE_VM_FIXED2400_E3_50G",
		"SUBCORE_VM_FIXED2450_E3_50G",
		"SUBCORE_VM_FIXED2475_E3_50G",
		"SUBCORE_VM_FIXED2500_E3_50G",
		"SUBCORE_VM_FIXED2550_E3_50G",
		"SUBCORE_VM_FIXED2600_E3_50G",
		"SUBCORE_VM_FIXED2625_E3_50G",
		"SUBCORE_VM_FIXED2650_E3_50G",
		"SUBCORE_VM_FIXED2700_E3_50G",
		"SUBCORE_VM_FIXED2750_E3_50G",
		"SUBCORE_VM_FIXED2775_E3_50G",
		"SUBCORE_VM_FIXED2800_E3_50G",
		"SUBCORE_VM_FIXED2850_E3_50G",
		"SUBCORE_VM_FIXED2875_E3_50G",
		"SUBCORE_VM_FIXED2900_E3_50G",
		"SUBCORE_VM_FIXED2925_E3_50G",
		"SUBCORE_VM_FIXED2950_E3_50G",
		"SUBCORE_VM_FIXED2975_E3_50G",
		"SUBCORE_VM_FIXED3000_E3_50G",
		"SUBCORE_VM_FIXED3025_E3_50G",
		"SUBCORE_VM_FIXED3050_E3_50G",
		"SUBCORE_VM_FIXED3075_E3_50G",
		"SUBCORE_VM_FIXED3100_E3_50G",
		"SUBCORE_VM_FIXED3125_E3_50G",
		"SUBCORE_VM_FIXED3150_E3_50G",
		"SUBCORE_VM_FIXED3200_E3_50G",
		"SUBCORE_VM_FIXED3225_E3_50G",
		"SUBCORE_VM_FIXED3250_E3_50G",
		"SUBCORE_VM_FIXED3300_E3_50G",
		"SUBCORE_VM_FIXED3325_E3_50G",
		"SUBCORE_VM_FIXED3375_E3_50G",
		"SUBCORE_VM_FIXED3400_E3_50G",
		"SUBCORE_VM_FIXED3450_E3_50G",
		"SUBCORE_VM_FIXED3500_E3_50G",
		"SUBCORE_VM_FIXED3525_E3_50G",
		"SUBCORE_VM_FIXED3575_E3_50G",
		"SUBCORE_VM_FIXED3600_E3_50G",
		"SUBCORE_VM_FIXED3625_E3_50G",
		"SUBCORE_VM_FIXED3675_E3_50G",
		"SUBCORE_VM_FIXED3700_E3_50G",
		"SUBCORE_VM_FIXED3750_E3_50G",
		"SUBCORE_VM_FIXED3800_E3_50G",
		"SUBCORE_VM_FIXED3825_E3_50G",
		"SUBCORE_VM_FIXED3850_E3_50G",
		"SUBCORE_VM_FIXED3875_E3_50G",
		"SUBCORE_VM_FIXED3900_E3_50G",
		"SUBCORE_VM_FIXED3975_E3_50G",
		"SUBCORE_VM_FIXED4000_E3_50G",
		"SUBCORE_VM_FIXED4025_E3_50G",
		"SUBCORE_VM_FIXED4050_E3_50G",
		"SUBCORE_VM_FIXED4100_E3_50G",
		"SUBCORE_VM_FIXED4125_E3_50G",
		"SUBCORE_VM_FIXED4200_E3_50G",
		"SUBCORE_VM_FIXED4225_E3_50G",
		"SUBCORE_VM_FIXED4250_E3_50G",
		"SUBCORE_VM_FIXED4275_E3_50G",
		"SUBCORE_VM_FIXED4300_E3_50G",
		"SUBCORE_VM_FIXED4350_E3_50G",
		"SUBCORE_VM_FIXED4375_E3_50G",
		"SUBCORE_VM_FIXED4400_E3_50G",
		"SUBCORE_VM_FIXED4425_E3_50G",
		"SUBCORE_VM_FIXED4500_E3_50G",
		"SUBCORE_VM_FIXED4550_E3_50G",
		"SUBCORE_VM_FIXED4575_E3_50G",
		"SUBCORE_VM_FIXED4600_E3_50G",
		"SUBCORE_VM_FIXED4625_E3_50G",
		"SUBCORE_VM_FIXED4650_E3_50G",
		"SUBCORE_VM_FIXED4675_E3_50G",
		"SUBCORE_VM_FIXED4700_E3_50G",
		"SUBCORE_VM_FIXED4725_E3_50G",
		"SUBCORE_VM_FIXED4750_E3_50G",
		"SUBCORE_VM_FIXED4800_E3_50G",
		"SUBCORE_VM_FIXED4875_E3_50G",
		"SUBCORE_VM_FIXED4900_E3_50G",
		"SUBCORE_VM_FIXED4950_E3_50G",
		"SUBCORE_VM_FIXED5000_E3_50G",
		"SUBCORE_VM_FIXED0025_E4_50G",
		"SUBCORE_VM_FIXED0050_E4_50G",
		"SUBCORE_VM_FIXED0075_E4_50G",
		"SUBCORE_VM_FIXED0100_E4_50G",
		"SUBCORE_VM_FIXED0125_E4_50G",
		"SUBCORE_VM_FIXED0150_E4_50G",
		"SUBCORE_VM_FIXED0175_E4_50G",
		"SUBCORE_VM_FIXED0200_E4_50G",
		"SUBCORE_VM_FIXED0225_E4_50G",
		"SUBCORE_VM_FIXED0250_E4_50G",
		"SUBCORE_VM_FIXED0275_E4_50G",
		"SUBCORE_VM_FIXED0300_E4_50G",
		"SUBCORE_VM_FIXED0325_E4_50G",
		"SUBCORE_VM_FIXED0350_E4_50G",
		"SUBCORE_VM_FIXED0375_E4_50G",
		"SUBCORE_VM_FIXED0400_E4_50G",
		"SUBCORE_VM_FIXED0425_E4_50G",
		"SUBCORE_VM_FIXED0450_E4_50G",
		"SUBCORE_VM_FIXED0475_E4_50G",
		"SUBCORE_VM_FIXED0500_E4_50G",
		"SUBCORE_VM_FIXED0525_E4_50G",
		"SUBCORE_VM_FIXED0550_E4_50G",
		"SUBCORE_VM_FIXED0575_E4_50G",
		"SUBCORE_VM_FIXED0600_E4_50G",
		"SUBCORE_VM_FIXED0625_E4_50G",
		"SUBCORE_VM_FIXED0650_E4_50G",
		"SUBCORE_VM_FIXED0675_E4_50G",
		"SUBCORE_VM_FIXED0700_E4_50G",
		"SUBCORE_VM_FIXED0725_E4_50G",
		"SUBCORE_VM_FIXED0750_E4_50G",
		"SUBCORE_VM_FIXED0775_E4_50G",
		"SUBCORE_VM_FIXED0800_E4_50G",
		"SUBCORE_VM_FIXED0825_E4_50G",
		"SUBCORE_VM_FIXED0850_E4_50G",
		"SUBCORE_VM_FIXED0875_E4_50G",
		"SUBCORE_VM_FIXED0900_E4_50G",
		"SUBCORE_VM_FIXED0925_E4_50G",
		"SUBCORE_VM_FIXED0950_E4_50G",
		"SUBCORE_VM_FIXED0975_E4_50G",
		"SUBCORE_VM_FIXED1000_E4_50G",
		"SUBCORE_VM_FIXED1025_E4_50G",
		"SUBCORE_VM_FIXED1050_E4_50G",
		"SUBCORE_VM_FIXED1075_E4_50G",
		"SUBCORE_VM_FIXED1100_E4_50G",
		"SUBCORE_VM_FIXED1125_E4_50G",
		"SUBCORE_VM_FIXED1150_E4_50G",
		"SUBCORE_VM_FIXED1175_E4_50G",
		"SUBCORE_VM_FIXED1200_E4_50G",
		"SUBCORE_VM_FIXED1225_E4_50G",
		"SUBCORE_VM_FIXED1250_E4_50G",
		"SUBCORE_VM_FIXED1275_E4_50G",
		"SUBCORE_VM_FIXED1300_E4_50G",
		"SUBCORE_VM_FIXED1325_E4_50G",
		"SUBCORE_VM_FIXED1350_E4_50G",
		"SUBCORE_VM_FIXED1375_E4_50G",
		"SUBCORE_VM_FIXED1400_E4_50G",
		"SUBCORE_VM_FIXED1425_E4_50G",
		"SUBCORE_VM_FIXED1450_E4_50G",
		"SUBCORE_VM_FIXED1475_E4_50G",
		"SUBCORE_VM_FIXED1500_E4_50G",
		"SUBCORE_VM_FIXED1525_E4_50G",
		"SUBCORE_VM_FIXED1550_E4_50G",
		"SUBCORE_VM_FIXED1575_E4_50G",
		"SUBCORE_VM_FIXED1600_E4_50G",
		"SUBCORE_VM_FIXED1625_E4_50G",
		"SUBCORE_VM_FIXED1650_E4_50G",
		"SUBCORE_VM_FIXED1700_E4_50G",
		"SUBCORE_VM_FIXED1725_E4_50G",
		"SUBCORE_VM_FIXED1750_E4_50G",
		"SUBCORE_VM_FIXED1800_E4_50G",
		"SUBCORE_VM_FIXED1850_E4_50G",
		"SUBCORE_VM_FIXED1875_E4_50G",
		"SUBCORE_VM_FIXED1900_E4_50G",
		"SUBCORE_VM_FIXED1925_E4_50G",
		"SUBCORE_VM_FIXED1950_E4_50G",
		"SUBCORE_VM_FIXED2000_E4_50G",
		"SUBCORE_VM_FIXED2025_E4_50G",
		"SUBCORE_VM_FIXED2050_E4_50G",
		"SUBCORE_VM_FIXED2100_E4_50G",
		"SUBCORE_VM_FIXED2125_E4_50G",
		"SUBCORE_VM_FIXED2150_E4_50G",
		"SUBCORE_VM_FIXED2175_E4_50G",
		"SUBCORE_VM_FIXED2200_E4_50G",
		"SUBCORE_VM_FIXED2250_E4_50G",
		"SUBCORE_VM_FIXED2275_E4_50G",
		"SUBCORE_VM_FIXED2300_E4_50G",
		"SUBCORE_VM_FIXED2325_E4_50G",
		"SUBCORE_VM_FIXED2350_E4_50G",
		"SUBCORE_VM_FIXED2375_E4_50G",
		"SUBCORE_VM_FIXED2400_E4_50G",
		"SUBCORE_VM_FIXED2450_E4_50G",
		"SUBCORE_VM_FIXED2475_E4_50G",
		"SUBCORE_VM_FIXED2500_E4_50G",
		"SUBCORE_VM_FIXED2550_E4_50G",
		"SUBCORE_VM_FIXED2600_E4_50G",
		"SUBCORE_VM_FIXED2625_E4_50G",
		"SUBCORE_VM_FIXED2650_E4_50G",
		"SUBCORE_VM_FIXED2700_E4_50G",
		"SUBCORE_VM_FIXED2750_E4_50G",
		"SUBCORE_VM_FIXED2775_E4_50G",
		"SUBCORE_VM_FIXED2800_E4_50G",
		"SUBCORE_VM_FIXED2850_E4_50G",
		"SUBCORE_VM_FIXED2875_E4_50G",
		"SUBCORE_VM_FIXED2900_E4_50G",
		"SUBCORE_VM_FIXED2925_E4_50G",
		"SUBCORE_VM_FIXED2950_E4_50G",
		"SUBCORE_VM_FIXED2975_E4_50G",
		"SUBCORE_VM_FIXED3000_E4_50G",
		"SUBCORE_VM_FIXED3025_E4_50G",
		"SUBCORE_VM_FIXED3050_E4_50G",
		"SUBCORE_VM_FIXED3075_E4_50G",
		"SUBCORE_VM_FIXED3100_E4_50G",
		"SUBCORE_VM_FIXED3125_E4_50G",
		"SUBCORE_VM_FIXED3150_E4_50G",
		"SUBCORE_VM_FIXED3200_E4_50G",
		"SUBCORE_VM_FIXED3225_E4_50G",
		"SUBCORE_VM_FIXED3250_E4_50G",
		"SUBCORE_VM_FIXED3300_E4_50G",
		"SUBCORE_VM_FIXED3325_E4_50G",
		"SUBCORE_VM_FIXED3375_E4_50G",
		"SUBCORE_VM_FIXED3400_E4_50G",
		"SUBCORE_VM_FIXED3450_E4_50G",
		"SUBCORE_VM_FIXED3500_E4_50G",
		"SUBCORE_VM_FIXED3525_E4_50G",
		"SUBCORE_VM_FIXED3575_E4_50G",
		"SUBCORE_VM_FIXED3600_E4_50G",
		"SUBCORE_VM_FIXED3625_E4_50G",
		"SUBCORE_VM_FIXED3675_E4_50G",
		"SUBCORE_VM_FIXED3700_E4_50G",
		"SUBCORE_VM_FIXED3750_E4_50G",
		"SUBCORE_VM_FIXED3800_E4_50G",
		"SUBCORE_VM_FIXED3825_E4_50G",
		"SUBCORE_VM_FIXED3850_E4_50G",
		"SUBCORE_VM_FIXED3875_E4_50G",
		"SUBCORE_VM_FIXED3900_E4_50G",
		"SUBCORE_VM_FIXED3975_E4_50G",
		"SUBCORE_VM_FIXED4000_E4_50G",
		"SUBCORE_VM_FIXED4025_E4_50G",
		"SUBCORE_VM_FIXED4050_E4_50G",
		"SUBCORE_VM_FIXED4100_E4_50G",
		"SUBCORE_VM_FIXED4125_E4_50G",
		"SUBCORE_VM_FIXED4200_E4_50G",
		"SUBCORE_VM_FIXED4225_E4_50G",
		"SUBCORE_VM_FIXED4250_E4_50G",
		"SUBCORE_VM_FIXED4275_E4_50G",
		"SUBCORE_VM_FIXED4300_E4_50G",
		"SUBCORE_VM_FIXED4350_E4_50G",
		"SUBCORE_VM_FIXED4375_E4_50G",
		"SUBCORE_VM_FIXED4400_E4_50G",
		"SUBCORE_VM_FIXED4425_E4_50G",
		"SUBCORE_VM_FIXED4500_E4_50G",
		"SUBCORE_VM_FIXED4550_E4_50G",
		"SUBCORE_VM_FIXED4575_E4_50G",
		"SUBCORE_VM_FIXED4600_E4_50G",
		"SUBCORE_VM_FIXED4625_E4_50G",
		"SUBCORE_VM_FIXED4650_E4_50G",
		"SUBCORE_VM_FIXED4675_E4_50G",
		"SUBCORE_VM_FIXED4700_E4_50G",
		"SUBCORE_VM_FIXED4725_E4_50G",
		"SUBCORE_VM_FIXED4750_E4_50G",
		"SUBCORE_VM_FIXED4800_E4_50G",
		"SUBCORE_VM_FIXED4875_E4_50G",
		"SUBCORE_VM_FIXED4900_E4_50G",
		"SUBCORE_VM_FIXED4950_E4_50G",
		"SUBCORE_VM_FIXED5000_E4_50G",
		"SUBCORE_VM_FIXED0020_A1_50G",
		"SUBCORE_VM_FIXED0040_A1_50G",
		"SUBCORE_VM_FIXED0060_A1_50G",
		"SUBCORE_VM_FIXED0080_A1_50G",
		"SUBCORE_VM_FIXED0100_A1_50G",
		"SUBCORE_VM_FIXED0120_A1_50G",
		"SUBCORE_VM_FIXED0140_A1_50G",
		"SUBCORE_VM_FIXED0160_A1_50G",
		"SUBCORE_VM_FIXED0180_A1_50G",
		"SUBCORE_VM_FIXED0200_A1_50G",
		"SUBCORE_VM_FIXED0220_A1_50G",
		"SUBCORE_VM_FIXED0240_A1_50G",
		"SUBCORE_VM_FIXED0260_A1_50G",
		"SUBCORE_VM_FIXED0280_A1_50G",
		"SUBCORE_VM_FIXED0300_A1_50G",
		"SUBCORE_VM_FIXED0320_A1_50G",
		"SUBCORE_VM_FIXED0340_A1_50G",
		"SUBCORE_VM_FIXED0360_A1_50G",
		"SUBCORE_VM_FIXED0380_A1_50G",
		"SUBCORE_VM_FIXED0400_A1_50G",
		"SUBCORE_VM_FIXED0420_A1_50G",
		"SUBCORE_VM_FIXED0440_A1_50G",
		"SUBCORE_VM_FIXED0460_A1_50G",
		"SUBCORE_VM_FIXED0480_A1_50G",
		"SUBCORE_VM_FIXED0500_A1_50G",
		"SUBCORE_VM_FIXED0520_A1_50G",
		"SUBCORE_VM_FIXED0540_A1_50G",
		"SUBCORE_VM_FIXED0560_A1_50G",
		"SUBCORE_VM_FIXED0580_A1_50G",
		"SUBCORE_VM_FIXED0600_A1_50G",
		"SUBCORE_VM_FIXED0620_A1_50G",
		"SUBCORE_VM_FIXED0640_A1_50G",
		"SUBCORE_VM_FIXED0660_A1_50G",
		"SUBCORE_VM_FIXED0680_A1_50G",
		"SUBCORE_VM_FIXED0700_A1_50G",
		"SUBCORE_VM_FIXED0720_A1_50G",
		"SUBCORE_VM_FIXED0740_A1_50G",
		"SUBCORE_VM_FIXED0760_A1_50G",
		"SUBCORE_VM_FIXED0780_A1_50G",
		"SUBCORE_VM_FIXED0800_A1_50G",
		"SUBCORE_VM_FIXED0820_A1_50G",
		"SUBCORE_VM_FIXED0840_A1_50G",
		"SUBCORE_VM_FIXED0860_A1_50G",
		"SUBCORE_VM_FIXED0880_A1_50G",
		"SUBCORE_VM_FIXED0900_A1_50G",
		"SUBCORE_VM_FIXED0920_A1_50G",
		"SUBCORE_VM_FIXED0940_A1_50G",
		"SUBCORE_VM_FIXED0960_A1_50G",
		"SUBCORE_VM_FIXED0980_A1_50G",
		"SUBCORE_VM_FIXED1000_A1_50G",
		"SUBCORE_VM_FIXED1020_A1_50G",
		"SUBCORE_VM_FIXED1040_A1_50G",
		"SUBCORE_VM_FIXED1060_A1_50G",
		"SUBCORE_VM_FIXED1080_A1_50G",
		"SUBCORE_VM_FIXED1100_A1_50G",
		"SUBCORE_VM_FIXED1120_A1_50G",
		"SUBCORE_VM_FIXED1140_A1_50G",
		"SUBCORE_VM_FIXED1160_A1_50G",
		"SUBCORE_VM_FIXED1180_A1_50G",
		"SUBCORE_VM_FIXED1200_A1_50G",
		"SUBCORE_VM_FIXED1220_A1_50G",
		"SUBCORE_VM_FIXED1240_A1_50G",
		"SUBCORE_VM_FIXED1260_A1_50G",
		"SUBCORE_VM_FIXED1280_A1_50G",
		"SUBCORE_VM_FIXED1300_A1_50G",
		"SUBCORE_VM_FIXED1320_A1_50G",
		"SUBCORE_VM_FIXED1340_A1_50G",
		"SUBCORE_VM_FIXED1360_A1_50G",
		"SUBCORE_VM_FIXED1380_A1_50G",
		"SUBCORE_VM_FIXED1400_A1_50G",
		"SUBCORE_VM_FIXED1420_A1_50G",
		"SUBCORE_VM_FIXED1440_A1_50G",
		"SUBCORE_VM_FIXED1460_A1_50G",
		"SUBCORE_VM_FIXED1480_A1_50G",
		"SUBCORE_VM_FIXED1500_A1_50G",
		"SUBCORE_VM_FIXED1520_A1_50G",
		"SUBCORE_VM_FIXED1540_A1_50G",
		"SUBCORE_VM_FIXED1560_A1_50G",
		"SUBCORE_VM_FIXED1580_A1_50G",
		"SUBCORE_VM_FIXED1600_A1_50G",
		"SUBCORE_VM_FIXED1620_A1_50G",
		"SUBCORE_VM_FIXED1640_A1_50G",
		"SUBCORE_VM_FIXED1660_A1_50G",
		"SUBCORE_VM_FIXED1680_A1_50G",
		"SUBCORE_VM_FIXED1700_A1_50G",
		"SUBCORE_VM_FIXED1720_A1_50G",
		"SUBCORE_VM_FIXED1740_A1_50G",
		"SUBCORE_VM_FIXED1760_A1_50G",
		"SUBCORE_VM_FIXED1780_A1_50G",
		"SUBCORE_VM_FIXED1800_A1_50G",
		"SUBCORE_VM_FIXED1820_A1_50G",
		"SUBCORE_VM_FIXED1840_A1_50G",
		"SUBCORE_VM_FIXED1860_A1_50G",
		"SUBCORE_VM_FIXED1880_A1_50G",
		"SUBCORE_VM_FIXED1900_A1_50G",
		"SUBCORE_VM_FIXED1920_A1_50G",
		"SUBCORE_VM_FIXED1940_A1_50G",
		"SUBCORE_VM_FIXED1960_A1_50G",
		"SUBCORE_VM_FIXED1980_A1_50G",
		"SUBCORE_VM_FIXED2000_A1_50G",
		"SUBCORE_VM_FIXED2020_A1_50G",
		"SUBCORE_VM_FIXED2040_A1_50G",
		"SUBCORE_VM_FIXED2060_A1_50G",
		"SUBCORE_VM_FIXED2080_A1_50G",
		"SUBCORE_VM_FIXED2100_A1_50G",
		"SUBCORE_VM_FIXED2120_A1_50G",
		"SUBCORE_VM_FIXED2140_A1_50G",
		"SUBCORE_VM_FIXED2160_A1_50G",
		"SUBCORE_VM_FIXED2180_A1_50G",
		"SUBCORE_VM_FIXED2200_A1_50G",
		"SUBCORE_VM_FIXED2220_A1_50G",
		"SUBCORE_VM_FIXED2240_A1_50G",
		"SUBCORE_VM_FIXED2260_A1_50G",
		"SUBCORE_VM_FIXED2280_A1_50G",
		"SUBCORE_VM_FIXED2300_A1_50G",
		"SUBCORE_VM_FIXED2320_A1_50G",
		"SUBCORE_VM_FIXED2340_A1_50G",
		"SUBCORE_VM_FIXED2360_A1_50G",
		"SUBCORE_VM_FIXED2380_A1_50G",
		"SUBCORE_VM_FIXED2400_A1_50G",
		"SUBCORE_VM_FIXED2420_A1_50G",
		"SUBCORE_VM_FIXED2440_A1_50G",
		"SUBCORE_VM_FIXED2460_A1_50G",
		"SUBCORE_VM_FIXED2480_A1_50G",
		"SUBCORE_VM_FIXED2500_A1_50G",
		"SUBCORE_VM_FIXED2520_A1_50G",
		"SUBCORE_VM_FIXED2540_A1_50G",
		"SUBCORE_VM_FIXED2560_A1_50G",
		"SUBCORE_VM_FIXED2580_A1_50G",
		"SUBCORE_VM_FIXED2600_A1_50G",
		"SUBCORE_VM_FIXED2620_A1_50G",
		"SUBCORE_VM_FIXED2640_A1_50G",
		"SUBCORE_VM_FIXED2660_A1_50G",
		"SUBCORE_VM_FIXED2680_A1_50G",
		"SUBCORE_VM_FIXED2700_A1_50G",
		"SUBCORE_VM_FIXED2720_A1_50G",
		"SUBCORE_VM_FIXED2740_A1_50G",
		"SUBCORE_VM_FIXED2760_A1_50G",
		"SUBCORE_VM_FIXED2780_A1_50G",
		"SUBCORE_VM_FIXED2800_A1_50G",
		"SUBCORE_VM_FIXED2820_A1_50G",
		"SUBCORE_VM_FIXED2840_A1_50G",
		"SUBCORE_VM_FIXED2860_A1_50G",
		"SUBCORE_VM_FIXED2880_A1_50G",
		"SUBCORE_VM_FIXED2900_A1_50G",
		"SUBCORE_VM_FIXED2920_A1_50G",
		"SUBCORE_VM_FIXED2940_A1_50G",
		"SUBCORE_VM_FIXED2960_A1_50G",
		"SUBCORE_VM_FIXED2980_A1_50G",
		"SUBCORE_VM_FIXED3000_A1_50G",
		"SUBCORE_VM_FIXED3020_A1_50G",
		"SUBCORE_VM_FIXED3040_A1_50G",
		"SUBCORE_VM_FIXED3060_A1_50G",
		"SUBCORE_VM_FIXED3080_A1_50G",
		"SUBCORE_VM_FIXED3100_A1_50G",
		"SUBCORE_VM_FIXED3120_A1_50G",
		"SUBCORE_VM_FIXED3140_A1_50G",
		"SUBCORE_VM_FIXED3160_A1_50G",
		"SUBCORE_VM_FIXED3180_A1_50G",
		"SUBCORE_VM_FIXED3200_A1_50G",
		"SUBCORE_VM_FIXED3220_A1_50G",
		"SUBCORE_VM_FIXED3240_A1_50G",
		"SUBCORE_VM_FIXED3260_A1_50G",
		"SUBCORE_VM_FIXED3280_A1_50G",
		"SUBCORE_VM_FIXED3300_A1_50G",
		"SUBCORE_VM_FIXED3320_A1_50G",
		"SUBCORE_VM_FIXED3340_A1_50G",
		"SUBCORE_VM_FIXED3360_A1_50G",
		"SUBCORE_VM_FIXED3380_A1_50G",
		"SUBCORE_VM_FIXED3400_A1_50G",
		"SUBCORE_VM_FIXED3420_A1_50G",
		"SUBCORE_VM_FIXED3440_A1_50G",
		"SUBCORE_VM_FIXED3460_A1_50G",
		"SUBCORE_VM_FIXED3480_A1_50G",
		"SUBCORE_VM_FIXED3500_A1_50G",
		"SUBCORE_VM_FIXED3520_A1_50G",
		"SUBCORE_VM_FIXED3540_A1_50G",
		"SUBCORE_VM_FIXED3560_A1_50G",
		"SUBCORE_VM_FIXED3580_A1_50G",
		"SUBCORE_VM_FIXED3600_A1_50G",
		"SUBCORE_VM_FIXED3620_A1_50G",
		"SUBCORE_VM_FIXED3640_A1_50G",
		"SUBCORE_VM_FIXED3660_A1_50G",
		"SUBCORE_VM_FIXED3680_A1_50G",
		"SUBCORE_VM_FIXED3700_A1_50G",
		"SUBCORE_VM_FIXED3720_A1_50G",
		"SUBCORE_VM_FIXED3740_A1_50G",
		"SUBCORE_VM_FIXED3760_A1_50G",
		"SUBCORE_VM_FIXED3780_A1_50G",
		"SUBCORE_VM_FIXED3800_A1_50G",
		"SUBCORE_VM_FIXED3820_A1_50G",
		"SUBCORE_VM_FIXED3840_A1_50G",
		"SUBCORE_VM_FIXED3860_A1_50G",
		"SUBCORE_VM_FIXED3880_A1_50G",
		"SUBCORE_VM_FIXED3900_A1_50G",
		"SUBCORE_VM_FIXED3920_A1_50G",
		"SUBCORE_VM_FIXED3940_A1_50G",
		"SUBCORE_VM_FIXED3960_A1_50G",
		"SUBCORE_VM_FIXED3980_A1_50G",
		"SUBCORE_VM_FIXED4000_A1_50G",
		"SUBCORE_VM_FIXED4020_A1_50G",
		"SUBCORE_VM_FIXED4040_A1_50G",
		"SUBCORE_VM_FIXED4060_A1_50G",
		"SUBCORE_VM_FIXED4080_A1_50G",
		"SUBCORE_VM_FIXED4100_A1_50G",
		"SUBCORE_VM_FIXED4120_A1_50G",
		"SUBCORE_VM_FIXED4140_A1_50G",
		"SUBCORE_VM_FIXED4160_A1_50G",
		"SUBCORE_VM_FIXED4180_A1_50G",
		"SUBCORE_VM_FIXED4200_A1_50G",
		"SUBCORE_VM_FIXED4220_A1_50G",
		"SUBCORE_VM_FIXED4240_A1_50G",
		"SUBCORE_VM_FIXED4260_A1_50G",
		"SUBCORE_VM_FIXED4280_A1_50G",
		"SUBCORE_VM_FIXED4300_A1_50G",
		"SUBCORE_VM_FIXED4320_A1_50G",
		"SUBCORE_VM_FIXED4340_A1_50G",
		"SUBCORE_VM_FIXED4360_A1_50G",
		"SUBCORE_VM_FIXED4380_A1_50G",
		"SUBCORE_VM_FIXED4400_A1_50G",
		"SUBCORE_VM_FIXED4420_A1_50G",
		"SUBCORE_VM_FIXED4440_A1_50G",
		"SUBCORE_VM_FIXED4460_A1_50G",
		"SUBCORE_VM_FIXED4480_A1_50G",
		"SUBCORE_VM_FIXED4500_A1_50G",
		"SUBCORE_VM_FIXED4520_A1_50G",
		"SUBCORE_VM_FIXED4540_A1_50G",
		"SUBCORE_VM_FIXED4560_A1_50G",
		"SUBCORE_VM_FIXED4580_A1_50G",
		"SUBCORE_VM_FIXED4600_A1_50G",
		"SUBCORE_VM_FIXED4620_A1_50G",
		"SUBCORE_VM_FIXED4640_A1_50G",
		"SUBCORE_VM_FIXED4660_A1_50G",
		"SUBCORE_VM_FIXED4680_A1_50G",
		"SUBCORE_VM_FIXED4700_A1_50G",
		"SUBCORE_VM_FIXED4720_A1_50G",
		"SUBCORE_VM_FIXED4740_A1_50G",
		"SUBCORE_VM_FIXED4760_A1_50G",
		"SUBCORE_VM_FIXED4780_A1_50G",
		"SUBCORE_VM_FIXED4800_A1_50G",
		"SUBCORE_VM_FIXED4820_A1_50G",
		"SUBCORE_VM_FIXED4840_A1_50G",
		"SUBCORE_VM_FIXED4860_A1_50G",
		"SUBCORE_VM_FIXED4880_A1_50G",
		"SUBCORE_VM_FIXED4900_A1_50G",
		"SUBCORE_VM_FIXED4920_A1_50G",
		"SUBCORE_VM_FIXED4940_A1_50G",
		"SUBCORE_VM_FIXED4960_A1_50G",
		"SUBCORE_VM_FIXED4980_A1_50G",
		"SUBCORE_VM_FIXED5000_A1_50G",
		"SUBCORE_VM_FIXED0090_X9_50G",
		"SUBCORE_VM_FIXED0180_X9_50G",
		"SUBCORE_VM_FIXED0270_X9_50G",
		"SUBCORE_VM_FIXED0360_X9_50G",
		"SUBCORE_VM_FIXED0450_X9_50G",
		"SUBCORE_VM_FIXED0540_X9_50G",
		"SUBCORE_VM_FIXED0630_X9_50G",
		"SUBCORE_VM_FIXED0720_X9_50G",
		"SUBCORE_VM_FIXED0810_X9_50G",
		"SUBCORE_VM_FIXED0900_X9_50G",
		"SUBCORE_VM_FIXED0990_X9_50G",
		"SUBCORE_VM_FIXED1080_X9_50G",
		"SUBCORE_VM_FIXED1170_X9_50G",
		"SUBCORE_VM_FIXED1260_X9_50G",
		"SUBCORE_VM_FIXED1350_X9_50G",
		"SUBCORE_VM_FIXED1440_X9_50G",
		"SUBCORE_VM_FIXED1530_X9_50G",
		"SUBCORE_VM_FIXED1620_X9_50G",
		"SUBCORE_VM_FIXED1710_X9_50G",
		"SUBCORE_VM_FIXED1800_X9_50G",
		"SUBCORE_VM_FIXED1890_X9_50G",
		"SUBCORE_VM_FIXED1980_X9_50G",
		"SUBCORE_VM_FIXED2070_X9_50G",
		"SUBCORE_VM_FIXED2160_X9_50G",
		"SUBCORE_VM_FIXED2250_X9_50G",
		"SUBCORE_VM_FIXED2340_X9_50G",
		"SUBCORE_VM_FIXED2430_X9_50G",
		"SUBCORE_VM_FIXED2520_X9_50G",
		"SUBCORE_VM_FIXED2610_X9_50G",
		"SUBCORE_VM_FIXED2700_X9_50G",
		"SUBCORE_VM_FIXED2790_X9_50G",
		"SUBCORE_VM_FIXED2880_X9_50G",
		"SUBCORE_VM_FIXED2970_X9_50G",
		"SUBCORE_VM_FIXED3060_X9_50G",
		"SUBCORE_VM_FIXED3150_X9_50G",
		"SUBCORE_VM_FIXED3240_X9_50G",
		"SUBCORE_VM_FIXED3330_X9_50G",
		"SUBCORE_VM_FIXED3420_X9_50G",
		"SUBCORE_VM_FIXED3510_X9_50G",
		"SUBCORE_VM_FIXED3600_X9_50G",
		"SUBCORE_VM_FIXED3690_X9_50G",
		"SUBCORE_VM_FIXED3780_X9_50G",
		"SUBCORE_VM_FIXED3870_X9_50G",
		"SUBCORE_VM_FIXED3960_X9_50G",
		"SUBCORE_VM_FIXED4050_X9_50G",
		"SUBCORE_VM_FIXED4140_X9_50G",
		"SUBCORE_VM_FIXED4230_X9_50G",
		"SUBCORE_VM_FIXED4320_X9_50G",
		"SUBCORE_VM_FIXED4410_X9_50G",
		"SUBCORE_VM_FIXED4500_X9_50G",
		"SUBCORE_VM_FIXED4590_X9_50G",
		"SUBCORE_VM_FIXED4680_X9_50G",
		"SUBCORE_VM_FIXED4770_X9_50G",
		"SUBCORE_VM_FIXED4860_X9_50G",
		"SUBCORE_VM_FIXED4950_X9_50G",
		"DYNAMIC_A1_50G",
		"FIXED0040_A1_50G",
		"FIXED0100_A1_50G",
		"FIXED0200_A1_50G",
		"FIXED0300_A1_50G",
		"FIXED0400_A1_50G",
		"FIXED0500_A1_50G",
		"FIXED0600_A1_50G",
		"FIXED0700_A1_50G",
		"FIXED0800_A1_50G",
		"FIXED0900_A1_50G",
		"FIXED1000_A1_50G",
		"FIXED1100_A1_50G",
		"FIXED1200_A1_50G",
		"FIXED1300_A1_50G",
		"FIXED1400_A1_50G",
		"FIXED1500_A1_50G",
		"FIXED1600_A1_50G",
		"FIXED1700_A1_50G",
		"FIXED1800_A1_50G",
		"FIXED1900_A1_50G",
		"FIXED2000_A1_50G",
		"FIXED2100_A1_50G",
		"FIXED2200_A1_50G",
		"FIXED2300_A1_50G",
		"FIXED2400_A1_50G",
		"FIXED2500_A1_50G",
		"FIXED2600_A1_50G",
		"FIXED2700_A1_50G",
		"FIXED2800_A1_50G",
		"FIXED2900_A1_50G",
		"FIXED3000_A1_50G",
		"FIXED3100_A1_50G",
		"FIXED3200_A1_50G",
		"FIXED3300_A1_50G",
		"FIXED3400_A1_50G",
		"FIXED3500_A1_50G",
		"FIXED3600_A1_50G",
		"FIXED3700_A1_50G",
		"FIXED3800_A1_50G",
		"FIXED3900_A1_50G",
		"FIXED4000_A1_50G",
		"ENTIREHOST_A1_50G",
		"DYNAMIC_X9_50G",
		"FIXED0040_X9_50G",
		"FIXED0400_X9_50G",
		"FIXED0800_X9_50G",
		"FIXED1200_X9_50G",
		"FIXED1600_X9_50G",
		"FIXED2000_X9_50G",
		"FIXED2400_X9_50G",
		"FIXED2800_X9_50G",
		"FIXED3200_X9_50G",
		"FIXED3600_X9_50G",
		"FIXED4000_X9_50G",
		"STANDARD_VM_FIXED0100_X9_50G",
		"STANDARD_VM_FIXED0200_X9_50G",
		"STANDARD_VM_FIXED0300_X9_50G",
		"STANDARD_VM_FIXED0400_X9_50G",
		"STANDARD_VM_FIXED0500_X9_50G",
		"STANDARD_VM_FIXED0600_X9_50G",
		"STANDARD_VM_FIXED0700_X9_50G",
		"STANDARD_VM_FIXED0800_X9_50G",
		"STANDARD_VM_FIXED0900_X9_50G",
		"STANDARD_VM_FIXED1000_X9_50G",
		"STANDARD_VM_FIXED1100_X9_50G",
		"STANDARD_VM_FIXED1200_X9_50G",
		"STANDARD_VM_FIXED1300_X9_50G",
		"STANDARD_VM_FIXED1400_X9_50G",
		"STANDARD_VM_FIXED1500_X9_50G",
		"STANDARD_VM_FIXED1600_X9_50G",
		"STANDARD_VM_FIXED1700_X9_50G",
		"STANDARD_VM_FIXED1800_X9_50G",
		"STANDARD_VM_FIXED1900_X9_50G",
		"STANDARD_VM_FIXED2000_X9_50G",
		"STANDARD_VM_FIXED2100_X9_50G",
		"STANDARD_VM_FIXED2200_X9_50G",
		"STANDARD_VM_FIXED2300_X9_50G",
		"STANDARD_VM_FIXED2400_X9_50G",
		"STANDARD_VM_FIXED2500_X9_50G",
		"STANDARD_VM_FIXED2600_X9_50G",
		"STANDARD_VM_FIXED2700_X9_50G",
		"STANDARD_VM_FIXED2800_X9_50G",
		"STANDARD_VM_FIXED2900_X9_50G",
		"STANDARD_VM_FIXED3000_X9_50G",
		"STANDARD_VM_FIXED3100_X9_50G",
		"STANDARD_VM_FIXED3200_X9_50G",
		"STANDARD_VM_FIXED3300_X9_50G",
		"STANDARD_VM_FIXED3400_X9_50G",
		"STANDARD_VM_FIXED3500_X9_50G",
		"STANDARD_VM_FIXED3600_X9_50G",
		"STANDARD_VM_FIXED3700_X9_50G",
		"STANDARD_VM_FIXED3800_X9_50G",
		"STANDARD_VM_FIXED3900_X9_50G",
		"STANDARD_VM_FIXED4000_X9_50G",
		"ENTIREHOST_X9_50G",
	}
}
