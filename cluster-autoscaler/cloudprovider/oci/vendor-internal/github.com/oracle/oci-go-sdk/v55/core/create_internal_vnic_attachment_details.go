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

// CreateInternalVnicAttachmentDetails Details for attaching a service VNIC to VNICaaS fleet.
type CreateInternalVnicAttachmentDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment to contain the VNIC attachment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// A private IP address of your choice to assign to the VNIC. Must be an
	// available IP address within the subnet's CIDR. If you don't specify a
	// value, Oracle automatically assigns a private IP address from the subnet.
	// This is the VNIC's *primary* private IP address. The value appears in
	// the Vnic object and also the
	// PrivateIp object returned by
	// ListPrivateIps and
	// GetPrivateIp.
	// Example: `10.0.3.3`
	PrivateIp *string `mandatory:"false" json:"privateIp"`

	// The compute instance id
	InstanceId *string `mandatory:"false" json:"instanceId"`

	// Id for a group of vnics sharing same resource pool, e.g. a group id could be a site/gateway id for
	// Ambassador SVNICs under the same site/gateway.
	GroupId *string `mandatory:"false" json:"groupId"`

	// The compute instance id of the resource pool to be used for the vnic
	InstanceIdForResourcePool *string `mandatory:"false" json:"instanceIdForResourcePool"`

	// The availability domain of the VNIC attachment
	InternalAvailabilityDomain *string `mandatory:"false" json:"internalAvailabilityDomain"`

	// The overlay MAC address of the instance
	MacAddress *string `mandatory:"false" json:"macAddress"`

	// index of NIC that VNIC is attaching to (OS boot order)
	NicIndex *int `mandatory:"false" json:"nicIndex"`

	// The tag used internally to identify the sending VNIC. It can be specified in scenarios where a specific
	// tag needs to be assigned. Examples of such scenarios include reboot migration and VMware support.
	VlanTag *int `mandatory:"false" json:"vlanTag"`

	// Shape of VNIC that is used to allocate resource in the data plane.
	VnicShape CreateInternalVnicAttachmentDetailsVnicShapeEnum `mandatory:"false" json:"vnicShape,omitempty"`

	// The substrate IP address of the instance
	SubstrateIp *string `mandatory:"false" json:"substrateIp"`

	// Indicates if vlanTag 0 can be assigned to this vnic or not.
	IsSkipVlanTag0 *bool `mandatory:"false" json:"isSkipVlanTag0"`

	// Specifies the shard to attach the VNIC to
	ShardId *string `mandatory:"false" json:"shardId"`

	// Property describing customer facing metrics
	MetadataList []CfmMetadata `mandatory:"false" json:"metadataList"`

	// Identifier of how the target instance is going be launched. For example, launch from marketplace.
	// STANDARD is the default type if not specified.
	LaunchType CreateInternalVnicAttachmentDetailsLaunchTypeEnum `mandatory:"false" json:"launchType,omitempty"`
}

func (m CreateInternalVnicAttachmentDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateInternalVnicAttachmentDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingCreateInternalVnicAttachmentDetailsVnicShapeEnum[string(m.VnicShape)]; !ok && m.VnicShape != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for VnicShape: %s. Supported values are: %s.", m.VnicShape, strings.Join(GetCreateInternalVnicAttachmentDetailsVnicShapeEnumStringValues(), ",")))
	}
	if _, ok := mappingCreateInternalVnicAttachmentDetailsLaunchTypeEnum[string(m.LaunchType)]; !ok && m.LaunchType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LaunchType: %s. Supported values are: %s.", m.LaunchType, strings.Join(GetCreateInternalVnicAttachmentDetailsLaunchTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateInternalVnicAttachmentDetailsVnicShapeEnum Enum with underlying type: string
type CreateInternalVnicAttachmentDetailsVnicShapeEnum string

// Set of constants representing the allowable values for CreateInternalVnicAttachmentDetailsVnicShapeEnum
const (
	CreateInternalVnicAttachmentDetailsVnicShapeDynamic                    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "DYNAMIC"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0040                  CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0040"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0060                  CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0060"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0060Psm               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0060_PSM"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0100                  CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0100"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0120                  CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0120"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed01202x                CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0120_2X"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0200                  CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0200"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0240                  CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0240"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0480                  CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0480"
	CreateInternalVnicAttachmentDetailsVnicShapeEntirehost                 CreateInternalVnicAttachmentDetailsVnicShapeEnum = "ENTIREHOST"
	CreateInternalVnicAttachmentDetailsVnicShapeDynamic25g                 CreateInternalVnicAttachmentDetailsVnicShapeEnum = "DYNAMIC_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed004025g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0040_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed010025g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0100_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed020025g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0200_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed040025g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0400_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed080025g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0800_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed160025g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1600_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed240025g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2400_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeEntirehost25g              CreateInternalVnicAttachmentDetailsVnicShapeEnum = "ENTIREHOST_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeDynamicE125g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "DYNAMIC_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0040E125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0040_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0070E125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0070_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0140E125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0140_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0280E125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0280_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0560E125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0560_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1120E125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1120_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1680E125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1680_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeEntirehostE125g            CreateInternalVnicAttachmentDetailsVnicShapeEnum = "ENTIREHOST_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeDynamicB125g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "DYNAMIC_B1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0040B125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0040_B1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0060B125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0060_B1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0120B125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0120_B1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0240B125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0240_B1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0480B125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0480_B1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0960B125g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0960_B1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeEntirehostB125g            CreateInternalVnicAttachmentDetailsVnicShapeEnum = "ENTIREHOST_B1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeMicroVmFixed0048E125g      CreateInternalVnicAttachmentDetailsVnicShapeEnum = "MICRO_VM_FIXED0048_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeMicroLbFixed0001E125g      CreateInternalVnicAttachmentDetailsVnicShapeEnum = "MICRO_LB_FIXED0001_E1_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeVnicaasFixed0200           CreateInternalVnicAttachmentDetailsVnicShapeEnum = "VNICAAS_FIXED0200"
	CreateInternalVnicAttachmentDetailsVnicShapeVnicaasFixed0400           CreateInternalVnicAttachmentDetailsVnicShapeEnum = "VNICAAS_FIXED0400"
	CreateInternalVnicAttachmentDetailsVnicShapeVnicaasFixed0700           CreateInternalVnicAttachmentDetailsVnicShapeEnum = "VNICAAS_FIXED0700"
	CreateInternalVnicAttachmentDetailsVnicShapeVnicaasNlbApproved10g      CreateInternalVnicAttachmentDetailsVnicShapeEnum = "VNICAAS_NLB_APPROVED_10G"
	CreateInternalVnicAttachmentDetailsVnicShapeVnicaasNlbApproved25g      CreateInternalVnicAttachmentDetailsVnicShapeEnum = "VNICAAS_NLB_APPROVED_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeVnicaasTelesis25g          CreateInternalVnicAttachmentDetailsVnicShapeEnum = "VNICAAS_TELESIS_25G"
	CreateInternalVnicAttachmentDetailsVnicShapeVnicaasTelesis10g          CreateInternalVnicAttachmentDetailsVnicShapeEnum = "VNICAAS_TELESIS_10G"
	CreateInternalVnicAttachmentDetailsVnicShapeVnicaasAmbassadorFixed0100 CreateInternalVnicAttachmentDetailsVnicShapeEnum = "VNICAAS_AMBASSADOR_FIXED0100"
	CreateInternalVnicAttachmentDetailsVnicShapeVnicaasPrivatedns          CreateInternalVnicAttachmentDetailsVnicShapeEnum = "VNICAAS_PRIVATEDNS"
	CreateInternalVnicAttachmentDetailsVnicShapeVnicaasFwaas               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "VNICAAS_FWAAS"
	CreateInternalVnicAttachmentDetailsVnicShapeDynamicE350g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "DYNAMIC_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0040E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0040_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0100E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0100_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0200E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0200_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0300E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0300_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0400E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0400_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0500E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0500_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0600E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0600_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0700E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0700_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0800E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0800_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0900E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0900_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1000E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1000_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1100E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1100_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1200E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1200_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1300E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1300_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1400E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1400_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1500E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1500_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1600E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1600_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1700E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1700_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1800E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1800_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1900E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1900_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2000E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2000_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2100E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2100_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2200E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2200_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2300E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2300_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2400E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2400_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2500E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2500_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2600E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2600_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2700E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2700_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2800E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2800_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2900E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2900_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3000E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3000_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3100E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3100_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3200E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3200_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3300E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3300_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3400E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3400_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3500E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3500_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3600E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3600_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3700E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3700_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3800E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3800_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3900E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3900_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed4000E350g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED4000_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeEntirehostE350g            CreateInternalVnicAttachmentDetailsVnicShapeEnum = "ENTIREHOST_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeDynamicE450g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "DYNAMIC_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0040E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0040_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0100E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0100_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0200E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0200_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0300E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0300_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0400E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0400_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0500E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0500_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0600E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0600_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0700E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0700_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0800E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0800_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0900E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0900_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1000E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1000_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1100E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1100_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1200E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1200_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1300E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1300_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1400E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1400_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1500E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1500_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1600E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1600_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1700E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1700_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1800E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1800_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1900E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1900_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2000E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2000_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2100E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2100_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2200E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2200_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2300E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2300_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2400E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2400_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2500E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2500_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2600E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2600_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2700E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2700_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2800E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2800_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2900E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2900_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3000E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3000_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3100E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3100_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3200E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3200_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3300E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3300_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3400E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3400_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3500E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3500_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3600E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3600_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3700E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3700_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3800E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3800_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3900E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3900_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed4000E450g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED4000_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeEntirehostE450g            CreateInternalVnicAttachmentDetailsVnicShapeEnum = "ENTIREHOST_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeMicroVmFixed0050E350g      CreateInternalVnicAttachmentDetailsVnicShapeEnum = "MICRO_VM_FIXED0050_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0025E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0025_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0050E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0050_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0075E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0075_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0100E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0100_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0125E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0125_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0150E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0150_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0175E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0175_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0200E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0200_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0225E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0225_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0250E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0250_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0275E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0275_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0300E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0300_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0325E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0325_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0350E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0350_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0375E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0375_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0400E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0400_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0425E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0425_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0450E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0450_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0475E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0475_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0500E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0500_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0525E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0525_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0550E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0550_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0575E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0575_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0600E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0600_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0625E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0625_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0650E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0650_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0675E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0675_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0700E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0700_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0725E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0725_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0750E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0750_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0775E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0775_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0800E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0800_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0825E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0825_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0850E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0850_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0875E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0875_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0900E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0925E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0925_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0950E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0950_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0975E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0975_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1000E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1000_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1025E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1025_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1050E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1050_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1075E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1075_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1100E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1100_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1125E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1125_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1150E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1150_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1175E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1175_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1200E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1200_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1225E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1225_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1250E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1250_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1275E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1275_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1300E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1300_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1325E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1325_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1350E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1350_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1375E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1375_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1400E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1400_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1425E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1425_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1450E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1450_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1475E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1475_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1500E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1500_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1525E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1525_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1550E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1550_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1575E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1575_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1600E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1600_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1625E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1625_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1650E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1650_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1700E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1700_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1725E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1725_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1750E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1750_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1800E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1850E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1850_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1875E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1875_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1900E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1900_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1925E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1925_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1950E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1950_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2000E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2000_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2025E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2025_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2050E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2050_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2100E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2100_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2125E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2125_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2150E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2150_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2175E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2175_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2200E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2200_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2250E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2250_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2275E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2275_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2300E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2300_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2325E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2325_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2350E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2350_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2375E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2375_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2400E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2400_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2450E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2450_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2475E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2475_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2500E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2500_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2550E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2550_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2600E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2600_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2625E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2625_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2650E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2650_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2700E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2750E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2750_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2775E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2775_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2800E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2800_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2850E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2850_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2875E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2875_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2900E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2900_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2925E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2925_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2950E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2950_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2975E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2975_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3000E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3000_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3025E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3025_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3050E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3050_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3075E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3075_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3100E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3100_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3125E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3125_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3150E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3150_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3200E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3200_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3225E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3225_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3250E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3250_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3300E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3300_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3325E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3325_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3375E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3375_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3400E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3400_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3450E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3450_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3500E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3500_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3525E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3525_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3575E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3575_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3600E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3625E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3625_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3675E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3675_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3700E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3700_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3750E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3750_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3800E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3800_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3825E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3825_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3850E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3850_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3875E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3875_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3900E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3900_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3975E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3975_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4000E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4000_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4025E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4025_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4050E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4050_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4100E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4100_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4125E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4125_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4200E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4200_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4225E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4225_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4250E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4250_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4275E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4275_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4300E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4300_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4350E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4350_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4375E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4375_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4400E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4400_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4425E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4425_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4500E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4550E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4550_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4575E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4575_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4600E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4600_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4625E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4625_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4650E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4650_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4675E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4675_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4700E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4700_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4725E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4725_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4750E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4750_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4800E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4800_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4875E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4875_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4900E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4900_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4950E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4950_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed5000E350g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED5000_E3_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0025E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0025_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0050E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0050_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0075E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0075_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0100E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0100_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0125E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0125_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0150E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0150_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0175E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0175_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0200E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0200_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0225E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0225_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0250E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0250_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0275E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0275_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0300E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0300_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0325E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0325_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0350E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0350_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0375E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0375_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0400E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0400_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0425E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0425_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0450E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0450_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0475E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0475_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0500E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0500_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0525E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0525_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0550E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0550_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0575E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0575_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0600E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0600_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0625E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0625_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0650E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0650_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0675E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0675_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0700E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0700_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0725E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0725_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0750E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0750_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0775E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0775_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0800E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0800_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0825E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0825_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0850E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0850_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0875E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0875_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0900E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0925E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0925_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0950E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0950_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0975E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0975_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1000E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1000_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1025E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1025_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1050E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1050_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1075E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1075_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1100E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1100_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1125E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1125_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1150E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1150_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1175E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1175_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1200E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1200_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1225E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1225_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1250E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1250_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1275E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1275_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1300E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1300_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1325E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1325_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1350E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1350_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1375E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1375_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1400E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1400_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1425E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1425_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1450E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1450_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1475E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1475_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1500E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1500_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1525E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1525_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1550E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1550_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1575E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1575_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1600E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1600_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1625E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1625_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1650E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1650_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1700E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1700_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1725E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1725_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1750E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1750_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1800E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1850E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1850_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1875E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1875_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1900E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1900_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1925E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1925_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1950E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1950_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2000E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2000_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2025E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2025_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2050E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2050_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2100E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2100_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2125E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2125_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2150E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2150_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2175E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2175_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2200E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2200_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2250E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2250_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2275E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2275_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2300E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2300_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2325E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2325_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2350E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2350_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2375E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2375_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2400E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2400_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2450E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2450_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2475E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2475_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2500E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2500_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2550E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2550_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2600E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2600_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2625E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2625_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2650E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2650_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2700E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2750E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2750_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2775E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2775_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2800E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2800_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2850E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2850_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2875E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2875_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2900E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2900_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2925E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2925_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2950E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2950_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2975E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2975_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3000E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3000_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3025E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3025_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3050E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3050_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3075E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3075_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3100E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3100_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3125E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3125_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3150E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3150_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3200E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3200_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3225E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3225_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3250E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3250_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3300E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3300_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3325E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3325_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3375E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3375_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3400E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3400_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3450E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3450_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3500E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3500_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3525E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3525_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3575E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3575_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3600E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3625E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3625_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3675E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3675_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3700E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3700_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3750E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3750_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3800E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3800_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3825E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3825_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3850E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3850_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3875E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3875_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3900E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3900_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3975E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3975_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4000E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4000_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4025E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4025_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4050E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4050_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4100E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4100_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4125E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4125_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4200E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4200_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4225E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4225_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4250E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4250_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4275E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4275_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4300E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4300_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4350E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4350_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4375E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4375_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4400E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4400_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4425E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4425_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4500E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4550E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4550_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4575E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4575_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4600E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4600_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4625E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4625_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4650E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4650_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4675E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4675_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4700E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4700_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4725E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4725_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4750E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4750_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4800E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4800_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4875E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4875_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4900E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4900_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4950E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4950_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed5000E450g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED5000_E4_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0020A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0020_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0040A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0040_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0060A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0060_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0080A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0080_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0100A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0100_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0120A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0120_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0140A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0140_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0160A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0160_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0180A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0180_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0200A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0200_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0220A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0220_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0240A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0240_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0260A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0260_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0280A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0280_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0300A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0300_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0320A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0320_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0340A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0340_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0360A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0360_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0380A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0380_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0400A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0400_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0420A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0420_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0440A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0440_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0460A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0460_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0480A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0480_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0500A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0500_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0520A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0520_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0540A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0540_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0560A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0560_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0580A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0580_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0600A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0600_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0620A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0620_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0640A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0640_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0660A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0660_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0680A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0680_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0700A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0700_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0720A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0720_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0740A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0740_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0760A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0760_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0780A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0780_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0800A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0800_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0820A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0820_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0840A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0840_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0860A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0860_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0880A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0880_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0900A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0920A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0920_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0940A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0940_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0960A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0960_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0980A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0980_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1000A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1000_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1020A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1020_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1040A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1040_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1060A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1060_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1080A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1080_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1100A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1100_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1120A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1120_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1140A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1140_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1160A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1160_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1180A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1180_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1200A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1200_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1220A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1220_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1240A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1240_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1260A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1260_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1280A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1280_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1300A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1300_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1320A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1320_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1340A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1340_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1360A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1360_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1380A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1380_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1400A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1400_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1420A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1420_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1440A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1440_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1460A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1460_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1480A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1480_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1500A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1500_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1520A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1520_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1540A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1540_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1560A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1560_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1580A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1580_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1600A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1600_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1620A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1620_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1640A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1640_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1660A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1660_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1680A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1680_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1700A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1700_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1720A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1720_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1740A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1740_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1760A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1760_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1780A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1780_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1800A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1820A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1820_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1840A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1840_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1860A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1860_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1880A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1880_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1900A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1900_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1920A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1920_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1940A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1940_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1960A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1960_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1980A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1980_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2000A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2000_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2020A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2020_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2040A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2040_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2060A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2060_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2080A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2080_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2100A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2100_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2120A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2120_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2140A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2140_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2160A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2160_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2180A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2180_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2200A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2200_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2220A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2220_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2240A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2240_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2260A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2260_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2280A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2280_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2300A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2300_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2320A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2320_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2340A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2340_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2360A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2360_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2380A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2380_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2400A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2400_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2420A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2420_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2440A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2440_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2460A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2460_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2480A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2480_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2500A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2500_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2520A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2520_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2540A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2540_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2560A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2560_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2580A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2580_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2600A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2600_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2620A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2620_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2640A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2640_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2660A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2660_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2680A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2680_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2700A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2720A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2720_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2740A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2740_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2760A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2760_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2780A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2780_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2800A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2800_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2820A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2820_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2840A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2840_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2860A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2860_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2880A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2880_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2900A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2900_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2920A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2920_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2940A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2940_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2960A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2960_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2980A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2980_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3000A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3000_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3020A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3020_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3040A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3040_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3060A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3060_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3080A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3080_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3100A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3100_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3120A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3120_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3140A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3140_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3160A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3160_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3180A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3180_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3200A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3200_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3220A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3220_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3240A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3240_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3260A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3260_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3280A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3280_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3300A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3300_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3320A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3320_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3340A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3340_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3360A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3360_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3380A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3380_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3400A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3400_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3420A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3420_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3440A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3440_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3460A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3460_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3480A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3480_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3500A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3500_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3520A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3520_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3540A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3540_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3560A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3560_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3580A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3580_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3600A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3620A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3620_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3640A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3640_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3660A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3660_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3680A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3680_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3700A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3700_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3720A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3720_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3740A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3740_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3760A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3760_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3780A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3780_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3800A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3800_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3820A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3820_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3840A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3840_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3860A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3860_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3880A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3880_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3900A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3900_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3920A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3920_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3940A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3940_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3960A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3960_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3980A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3980_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4000A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4000_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4020A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4020_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4040A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4040_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4060A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4060_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4080A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4080_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4100A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4100_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4120A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4120_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4140A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4140_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4160A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4160_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4180A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4180_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4200A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4200_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4220A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4220_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4240A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4240_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4260A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4260_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4280A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4280_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4300A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4300_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4320A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4320_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4340A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4340_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4360A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4360_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4380A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4380_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4400A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4400_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4420A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4420_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4440A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4440_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4460A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4460_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4480A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4480_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4500A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4520A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4520_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4540A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4540_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4560A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4560_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4580A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4580_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4600A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4600_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4620A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4620_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4640A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4640_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4660A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4660_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4680A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4680_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4700A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4700_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4720A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4720_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4740A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4740_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4760A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4760_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4780A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4780_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4800A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4800_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4820A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4820_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4840A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4840_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4860A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4860_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4880A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4880_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4900A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4900_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4920A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4920_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4940A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4940_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4960A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4960_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4980A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4980_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed5000A150g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED5000_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0090X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0090_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0180X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0180_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0270X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0270_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0360X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0360_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0450X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0450_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0540X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0540_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0630X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0630_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0720X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0720_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0810X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0810_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0900X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0990X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0990_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1080X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1080_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1170X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1170_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1260X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1260_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1350X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1350_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1440X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1440_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1530X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1530_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1620X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1620_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1710X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1710_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1800X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1890X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1890_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1980X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1980_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2070X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2070_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2160X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2160_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2250X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2250_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2340X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2340_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2430X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2430_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2520X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2520_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2610X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2610_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2700X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2790X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2790_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2880X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2880_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2970X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2970_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3060X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3060_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3150X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3150_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3240X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3240_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3330X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3330_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3420X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3420_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3510X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3510_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3600X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3690X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3690_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3780X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3780_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3870X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3870_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3960X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3960_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4050X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4050_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4140X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4140_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4230X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4230_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4320X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4320_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4410X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4410_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4500X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4590X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4590_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4680X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4680_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4770X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4770_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4860X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4860_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4950X950g    CreateInternalVnicAttachmentDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4950_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeDynamicA150g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "DYNAMIC_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0040A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0040_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0100A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0100_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0200A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0200_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0300A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0300_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0400A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0400_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0500A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0500_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0600A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0600_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0700A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0700_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0800A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0800_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0900A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0900_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1000A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1000_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1100A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1100_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1200A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1200_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1300A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1300_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1400A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1400_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1500A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1500_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1600A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1600_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1700A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1700_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1800A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1800_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1900A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1900_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2000A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2000_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2100A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2100_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2200A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2200_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2300A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2300_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2400A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2400_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2500A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2500_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2600A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2600_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2700A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2700_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2800A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2800_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2900A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2900_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3000A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3000_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3100A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3100_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3200A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3200_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3300A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3300_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3400A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3400_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3500A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3500_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3600A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3600_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3700A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3700_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3800A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3800_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3900A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3900_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed4000A150g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED4000_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeEntirehostA150g            CreateInternalVnicAttachmentDetailsVnicShapeEnum = "ENTIREHOST_A1_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeDynamicX950g               CreateInternalVnicAttachmentDetailsVnicShapeEnum = "DYNAMIC_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0040X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0040_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0400X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0400_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed0800X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED0800_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1200X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1200_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed1600X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED1600_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2000X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2000_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2400X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2400_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed2800X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED2800_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3200X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3200_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed3600X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED3600_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeFixed4000X950g             CreateInternalVnicAttachmentDetailsVnicShapeEnum = "FIXED4000_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0100X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED0100_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0200X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED0200_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0300X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED0300_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0400X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED0400_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0500X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED0500_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0600X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED0600_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0700X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED0700_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0800X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED0800_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0900X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED0900_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1000X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED1000_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1100X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED1100_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1200X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED1200_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1300X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED1300_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1400X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED1400_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1500X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED1500_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1600X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED1600_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1700X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED1700_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1800X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED1800_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1900X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED1900_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2000X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED2000_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2100X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED2100_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2200X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED2200_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2300X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED2300_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2400X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED2400_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2500X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED2500_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2600X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED2600_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2700X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED2700_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2800X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED2800_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2900X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED2900_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3000X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED3000_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3100X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED3100_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3200X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED3200_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3300X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED3300_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3400X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED3400_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3500X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED3500_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3600X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED3600_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3700X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED3700_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3800X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED3800_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3900X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED3900_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed4000X950g   CreateInternalVnicAttachmentDetailsVnicShapeEnum = "STANDARD_VM_FIXED4000_X9_50G"
	CreateInternalVnicAttachmentDetailsVnicShapeEntirehostX950g            CreateInternalVnicAttachmentDetailsVnicShapeEnum = "ENTIREHOST_X9_50G"
)

var mappingCreateInternalVnicAttachmentDetailsVnicShapeEnum = map[string]CreateInternalVnicAttachmentDetailsVnicShapeEnum{
	"DYNAMIC":                      CreateInternalVnicAttachmentDetailsVnicShapeDynamic,
	"FIXED0040":                    CreateInternalVnicAttachmentDetailsVnicShapeFixed0040,
	"FIXED0060":                    CreateInternalVnicAttachmentDetailsVnicShapeFixed0060,
	"FIXED0060_PSM":                CreateInternalVnicAttachmentDetailsVnicShapeFixed0060Psm,
	"FIXED0100":                    CreateInternalVnicAttachmentDetailsVnicShapeFixed0100,
	"FIXED0120":                    CreateInternalVnicAttachmentDetailsVnicShapeFixed0120,
	"FIXED0120_2X":                 CreateInternalVnicAttachmentDetailsVnicShapeFixed01202x,
	"FIXED0200":                    CreateInternalVnicAttachmentDetailsVnicShapeFixed0200,
	"FIXED0240":                    CreateInternalVnicAttachmentDetailsVnicShapeFixed0240,
	"FIXED0480":                    CreateInternalVnicAttachmentDetailsVnicShapeFixed0480,
	"ENTIREHOST":                   CreateInternalVnicAttachmentDetailsVnicShapeEntirehost,
	"DYNAMIC_25G":                  CreateInternalVnicAttachmentDetailsVnicShapeDynamic25g,
	"FIXED0040_25G":                CreateInternalVnicAttachmentDetailsVnicShapeFixed004025g,
	"FIXED0100_25G":                CreateInternalVnicAttachmentDetailsVnicShapeFixed010025g,
	"FIXED0200_25G":                CreateInternalVnicAttachmentDetailsVnicShapeFixed020025g,
	"FIXED0400_25G":                CreateInternalVnicAttachmentDetailsVnicShapeFixed040025g,
	"FIXED0800_25G":                CreateInternalVnicAttachmentDetailsVnicShapeFixed080025g,
	"FIXED1600_25G":                CreateInternalVnicAttachmentDetailsVnicShapeFixed160025g,
	"FIXED2400_25G":                CreateInternalVnicAttachmentDetailsVnicShapeFixed240025g,
	"ENTIREHOST_25G":               CreateInternalVnicAttachmentDetailsVnicShapeEntirehost25g,
	"DYNAMIC_E1_25G":               CreateInternalVnicAttachmentDetailsVnicShapeDynamicE125g,
	"FIXED0040_E1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0040E125g,
	"FIXED0070_E1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0070E125g,
	"FIXED0140_E1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0140E125g,
	"FIXED0280_E1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0280E125g,
	"FIXED0560_E1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0560E125g,
	"FIXED1120_E1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1120E125g,
	"FIXED1680_E1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1680E125g,
	"ENTIREHOST_E1_25G":            CreateInternalVnicAttachmentDetailsVnicShapeEntirehostE125g,
	"DYNAMIC_B1_25G":               CreateInternalVnicAttachmentDetailsVnicShapeDynamicB125g,
	"FIXED0040_B1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0040B125g,
	"FIXED0060_B1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0060B125g,
	"FIXED0120_B1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0120B125g,
	"FIXED0240_B1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0240B125g,
	"FIXED0480_B1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0480B125g,
	"FIXED0960_B1_25G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0960B125g,
	"ENTIREHOST_B1_25G":            CreateInternalVnicAttachmentDetailsVnicShapeEntirehostB125g,
	"MICRO_VM_FIXED0048_E1_25G":    CreateInternalVnicAttachmentDetailsVnicShapeMicroVmFixed0048E125g,
	"MICRO_LB_FIXED0001_E1_25G":    CreateInternalVnicAttachmentDetailsVnicShapeMicroLbFixed0001E125g,
	"VNICAAS_FIXED0200":            CreateInternalVnicAttachmentDetailsVnicShapeVnicaasFixed0200,
	"VNICAAS_FIXED0400":            CreateInternalVnicAttachmentDetailsVnicShapeVnicaasFixed0400,
	"VNICAAS_FIXED0700":            CreateInternalVnicAttachmentDetailsVnicShapeVnicaasFixed0700,
	"VNICAAS_NLB_APPROVED_10G":     CreateInternalVnicAttachmentDetailsVnicShapeVnicaasNlbApproved10g,
	"VNICAAS_NLB_APPROVED_25G":     CreateInternalVnicAttachmentDetailsVnicShapeVnicaasNlbApproved25g,
	"VNICAAS_TELESIS_25G":          CreateInternalVnicAttachmentDetailsVnicShapeVnicaasTelesis25g,
	"VNICAAS_TELESIS_10G":          CreateInternalVnicAttachmentDetailsVnicShapeVnicaasTelesis10g,
	"VNICAAS_AMBASSADOR_FIXED0100": CreateInternalVnicAttachmentDetailsVnicShapeVnicaasAmbassadorFixed0100,
	"VNICAAS_PRIVATEDNS":           CreateInternalVnicAttachmentDetailsVnicShapeVnicaasPrivatedns,
	"VNICAAS_FWAAS":                CreateInternalVnicAttachmentDetailsVnicShapeVnicaasFwaas,
	"DYNAMIC_E3_50G":               CreateInternalVnicAttachmentDetailsVnicShapeDynamicE350g,
	"FIXED0040_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0040E350g,
	"FIXED0100_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0100E350g,
	"FIXED0200_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0200E350g,
	"FIXED0300_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0300E350g,
	"FIXED0400_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0400E350g,
	"FIXED0500_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0500E350g,
	"FIXED0600_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0600E350g,
	"FIXED0700_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0700E350g,
	"FIXED0800_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0800E350g,
	"FIXED0900_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0900E350g,
	"FIXED1000_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1000E350g,
	"FIXED1100_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1100E350g,
	"FIXED1200_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1200E350g,
	"FIXED1300_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1300E350g,
	"FIXED1400_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1400E350g,
	"FIXED1500_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1500E350g,
	"FIXED1600_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1600E350g,
	"FIXED1700_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1700E350g,
	"FIXED1800_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1800E350g,
	"FIXED1900_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1900E350g,
	"FIXED2000_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2000E350g,
	"FIXED2100_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2100E350g,
	"FIXED2200_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2200E350g,
	"FIXED2300_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2300E350g,
	"FIXED2400_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2400E350g,
	"FIXED2500_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2500E350g,
	"FIXED2600_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2600E350g,
	"FIXED2700_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2700E350g,
	"FIXED2800_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2800E350g,
	"FIXED2900_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2900E350g,
	"FIXED3000_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3000E350g,
	"FIXED3100_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3100E350g,
	"FIXED3200_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3200E350g,
	"FIXED3300_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3300E350g,
	"FIXED3400_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3400E350g,
	"FIXED3500_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3500E350g,
	"FIXED3600_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3600E350g,
	"FIXED3700_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3700E350g,
	"FIXED3800_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3800E350g,
	"FIXED3900_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3900E350g,
	"FIXED4000_E3_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed4000E350g,
	"ENTIREHOST_E3_50G":            CreateInternalVnicAttachmentDetailsVnicShapeEntirehostE350g,
	"DYNAMIC_E4_50G":               CreateInternalVnicAttachmentDetailsVnicShapeDynamicE450g,
	"FIXED0040_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0040E450g,
	"FIXED0100_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0100E450g,
	"FIXED0200_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0200E450g,
	"FIXED0300_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0300E450g,
	"FIXED0400_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0400E450g,
	"FIXED0500_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0500E450g,
	"FIXED0600_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0600E450g,
	"FIXED0700_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0700E450g,
	"FIXED0800_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0800E450g,
	"FIXED0900_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0900E450g,
	"FIXED1000_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1000E450g,
	"FIXED1100_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1100E450g,
	"FIXED1200_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1200E450g,
	"FIXED1300_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1300E450g,
	"FIXED1400_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1400E450g,
	"FIXED1500_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1500E450g,
	"FIXED1600_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1600E450g,
	"FIXED1700_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1700E450g,
	"FIXED1800_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1800E450g,
	"FIXED1900_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1900E450g,
	"FIXED2000_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2000E450g,
	"FIXED2100_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2100E450g,
	"FIXED2200_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2200E450g,
	"FIXED2300_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2300E450g,
	"FIXED2400_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2400E450g,
	"FIXED2500_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2500E450g,
	"FIXED2600_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2600E450g,
	"FIXED2700_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2700E450g,
	"FIXED2800_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2800E450g,
	"FIXED2900_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2900E450g,
	"FIXED3000_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3000E450g,
	"FIXED3100_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3100E450g,
	"FIXED3200_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3200E450g,
	"FIXED3300_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3300E450g,
	"FIXED3400_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3400E450g,
	"FIXED3500_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3500E450g,
	"FIXED3600_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3600E450g,
	"FIXED3700_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3700E450g,
	"FIXED3800_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3800E450g,
	"FIXED3900_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3900E450g,
	"FIXED4000_E4_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed4000E450g,
	"ENTIREHOST_E4_50G":            CreateInternalVnicAttachmentDetailsVnicShapeEntirehostE450g,
	"MICRO_VM_FIXED0050_E3_50G":    CreateInternalVnicAttachmentDetailsVnicShapeMicroVmFixed0050E350g,
	"SUBCORE_VM_FIXED0025_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0025E350g,
	"SUBCORE_VM_FIXED0050_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0050E350g,
	"SUBCORE_VM_FIXED0075_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0075E350g,
	"SUBCORE_VM_FIXED0100_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0100E350g,
	"SUBCORE_VM_FIXED0125_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0125E350g,
	"SUBCORE_VM_FIXED0150_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0150E350g,
	"SUBCORE_VM_FIXED0175_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0175E350g,
	"SUBCORE_VM_FIXED0200_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0200E350g,
	"SUBCORE_VM_FIXED0225_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0225E350g,
	"SUBCORE_VM_FIXED0250_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0250E350g,
	"SUBCORE_VM_FIXED0275_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0275E350g,
	"SUBCORE_VM_FIXED0300_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0300E350g,
	"SUBCORE_VM_FIXED0325_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0325E350g,
	"SUBCORE_VM_FIXED0350_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0350E350g,
	"SUBCORE_VM_FIXED0375_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0375E350g,
	"SUBCORE_VM_FIXED0400_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0400E350g,
	"SUBCORE_VM_FIXED0425_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0425E350g,
	"SUBCORE_VM_FIXED0450_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0450E350g,
	"SUBCORE_VM_FIXED0475_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0475E350g,
	"SUBCORE_VM_FIXED0500_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0500E350g,
	"SUBCORE_VM_FIXED0525_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0525E350g,
	"SUBCORE_VM_FIXED0550_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0550E350g,
	"SUBCORE_VM_FIXED0575_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0575E350g,
	"SUBCORE_VM_FIXED0600_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0600E350g,
	"SUBCORE_VM_FIXED0625_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0625E350g,
	"SUBCORE_VM_FIXED0650_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0650E350g,
	"SUBCORE_VM_FIXED0675_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0675E350g,
	"SUBCORE_VM_FIXED0700_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0700E350g,
	"SUBCORE_VM_FIXED0725_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0725E350g,
	"SUBCORE_VM_FIXED0750_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0750E350g,
	"SUBCORE_VM_FIXED0775_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0775E350g,
	"SUBCORE_VM_FIXED0800_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0800E350g,
	"SUBCORE_VM_FIXED0825_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0825E350g,
	"SUBCORE_VM_FIXED0850_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0850E350g,
	"SUBCORE_VM_FIXED0875_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0875E350g,
	"SUBCORE_VM_FIXED0900_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0900E350g,
	"SUBCORE_VM_FIXED0925_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0925E350g,
	"SUBCORE_VM_FIXED0950_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0950E350g,
	"SUBCORE_VM_FIXED0975_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0975E350g,
	"SUBCORE_VM_FIXED1000_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1000E350g,
	"SUBCORE_VM_FIXED1025_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1025E350g,
	"SUBCORE_VM_FIXED1050_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1050E350g,
	"SUBCORE_VM_FIXED1075_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1075E350g,
	"SUBCORE_VM_FIXED1100_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1100E350g,
	"SUBCORE_VM_FIXED1125_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1125E350g,
	"SUBCORE_VM_FIXED1150_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1150E350g,
	"SUBCORE_VM_FIXED1175_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1175E350g,
	"SUBCORE_VM_FIXED1200_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1200E350g,
	"SUBCORE_VM_FIXED1225_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1225E350g,
	"SUBCORE_VM_FIXED1250_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1250E350g,
	"SUBCORE_VM_FIXED1275_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1275E350g,
	"SUBCORE_VM_FIXED1300_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1300E350g,
	"SUBCORE_VM_FIXED1325_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1325E350g,
	"SUBCORE_VM_FIXED1350_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1350E350g,
	"SUBCORE_VM_FIXED1375_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1375E350g,
	"SUBCORE_VM_FIXED1400_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1400E350g,
	"SUBCORE_VM_FIXED1425_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1425E350g,
	"SUBCORE_VM_FIXED1450_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1450E350g,
	"SUBCORE_VM_FIXED1475_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1475E350g,
	"SUBCORE_VM_FIXED1500_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1500E350g,
	"SUBCORE_VM_FIXED1525_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1525E350g,
	"SUBCORE_VM_FIXED1550_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1550E350g,
	"SUBCORE_VM_FIXED1575_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1575E350g,
	"SUBCORE_VM_FIXED1600_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1600E350g,
	"SUBCORE_VM_FIXED1625_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1625E350g,
	"SUBCORE_VM_FIXED1650_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1650E350g,
	"SUBCORE_VM_FIXED1700_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1700E350g,
	"SUBCORE_VM_FIXED1725_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1725E350g,
	"SUBCORE_VM_FIXED1750_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1750E350g,
	"SUBCORE_VM_FIXED1800_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1800E350g,
	"SUBCORE_VM_FIXED1850_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1850E350g,
	"SUBCORE_VM_FIXED1875_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1875E350g,
	"SUBCORE_VM_FIXED1900_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1900E350g,
	"SUBCORE_VM_FIXED1925_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1925E350g,
	"SUBCORE_VM_FIXED1950_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1950E350g,
	"SUBCORE_VM_FIXED2000_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2000E350g,
	"SUBCORE_VM_FIXED2025_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2025E350g,
	"SUBCORE_VM_FIXED2050_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2050E350g,
	"SUBCORE_VM_FIXED2100_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2100E350g,
	"SUBCORE_VM_FIXED2125_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2125E350g,
	"SUBCORE_VM_FIXED2150_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2150E350g,
	"SUBCORE_VM_FIXED2175_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2175E350g,
	"SUBCORE_VM_FIXED2200_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2200E350g,
	"SUBCORE_VM_FIXED2250_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2250E350g,
	"SUBCORE_VM_FIXED2275_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2275E350g,
	"SUBCORE_VM_FIXED2300_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2300E350g,
	"SUBCORE_VM_FIXED2325_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2325E350g,
	"SUBCORE_VM_FIXED2350_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2350E350g,
	"SUBCORE_VM_FIXED2375_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2375E350g,
	"SUBCORE_VM_FIXED2400_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2400E350g,
	"SUBCORE_VM_FIXED2450_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2450E350g,
	"SUBCORE_VM_FIXED2475_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2475E350g,
	"SUBCORE_VM_FIXED2500_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2500E350g,
	"SUBCORE_VM_FIXED2550_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2550E350g,
	"SUBCORE_VM_FIXED2600_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2600E350g,
	"SUBCORE_VM_FIXED2625_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2625E350g,
	"SUBCORE_VM_FIXED2650_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2650E350g,
	"SUBCORE_VM_FIXED2700_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2700E350g,
	"SUBCORE_VM_FIXED2750_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2750E350g,
	"SUBCORE_VM_FIXED2775_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2775E350g,
	"SUBCORE_VM_FIXED2800_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2800E350g,
	"SUBCORE_VM_FIXED2850_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2850E350g,
	"SUBCORE_VM_FIXED2875_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2875E350g,
	"SUBCORE_VM_FIXED2900_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2900E350g,
	"SUBCORE_VM_FIXED2925_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2925E350g,
	"SUBCORE_VM_FIXED2950_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2950E350g,
	"SUBCORE_VM_FIXED2975_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2975E350g,
	"SUBCORE_VM_FIXED3000_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3000E350g,
	"SUBCORE_VM_FIXED3025_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3025E350g,
	"SUBCORE_VM_FIXED3050_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3050E350g,
	"SUBCORE_VM_FIXED3075_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3075E350g,
	"SUBCORE_VM_FIXED3100_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3100E350g,
	"SUBCORE_VM_FIXED3125_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3125E350g,
	"SUBCORE_VM_FIXED3150_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3150E350g,
	"SUBCORE_VM_FIXED3200_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3200E350g,
	"SUBCORE_VM_FIXED3225_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3225E350g,
	"SUBCORE_VM_FIXED3250_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3250E350g,
	"SUBCORE_VM_FIXED3300_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3300E350g,
	"SUBCORE_VM_FIXED3325_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3325E350g,
	"SUBCORE_VM_FIXED3375_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3375E350g,
	"SUBCORE_VM_FIXED3400_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3400E350g,
	"SUBCORE_VM_FIXED3450_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3450E350g,
	"SUBCORE_VM_FIXED3500_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3500E350g,
	"SUBCORE_VM_FIXED3525_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3525E350g,
	"SUBCORE_VM_FIXED3575_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3575E350g,
	"SUBCORE_VM_FIXED3600_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3600E350g,
	"SUBCORE_VM_FIXED3625_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3625E350g,
	"SUBCORE_VM_FIXED3675_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3675E350g,
	"SUBCORE_VM_FIXED3700_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3700E350g,
	"SUBCORE_VM_FIXED3750_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3750E350g,
	"SUBCORE_VM_FIXED3800_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3800E350g,
	"SUBCORE_VM_FIXED3825_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3825E350g,
	"SUBCORE_VM_FIXED3850_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3850E350g,
	"SUBCORE_VM_FIXED3875_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3875E350g,
	"SUBCORE_VM_FIXED3900_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3900E350g,
	"SUBCORE_VM_FIXED3975_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3975E350g,
	"SUBCORE_VM_FIXED4000_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4000E350g,
	"SUBCORE_VM_FIXED4025_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4025E350g,
	"SUBCORE_VM_FIXED4050_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4050E350g,
	"SUBCORE_VM_FIXED4100_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4100E350g,
	"SUBCORE_VM_FIXED4125_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4125E350g,
	"SUBCORE_VM_FIXED4200_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4200E350g,
	"SUBCORE_VM_FIXED4225_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4225E350g,
	"SUBCORE_VM_FIXED4250_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4250E350g,
	"SUBCORE_VM_FIXED4275_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4275E350g,
	"SUBCORE_VM_FIXED4300_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4300E350g,
	"SUBCORE_VM_FIXED4350_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4350E350g,
	"SUBCORE_VM_FIXED4375_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4375E350g,
	"SUBCORE_VM_FIXED4400_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4400E350g,
	"SUBCORE_VM_FIXED4425_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4425E350g,
	"SUBCORE_VM_FIXED4500_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4500E350g,
	"SUBCORE_VM_FIXED4550_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4550E350g,
	"SUBCORE_VM_FIXED4575_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4575E350g,
	"SUBCORE_VM_FIXED4600_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4600E350g,
	"SUBCORE_VM_FIXED4625_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4625E350g,
	"SUBCORE_VM_FIXED4650_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4650E350g,
	"SUBCORE_VM_FIXED4675_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4675E350g,
	"SUBCORE_VM_FIXED4700_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4700E350g,
	"SUBCORE_VM_FIXED4725_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4725E350g,
	"SUBCORE_VM_FIXED4750_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4750E350g,
	"SUBCORE_VM_FIXED4800_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4800E350g,
	"SUBCORE_VM_FIXED4875_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4875E350g,
	"SUBCORE_VM_FIXED4900_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4900E350g,
	"SUBCORE_VM_FIXED4950_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4950E350g,
	"SUBCORE_VM_FIXED5000_E3_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed5000E350g,
	"SUBCORE_VM_FIXED0025_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0025E450g,
	"SUBCORE_VM_FIXED0050_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0050E450g,
	"SUBCORE_VM_FIXED0075_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0075E450g,
	"SUBCORE_VM_FIXED0100_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0100E450g,
	"SUBCORE_VM_FIXED0125_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0125E450g,
	"SUBCORE_VM_FIXED0150_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0150E450g,
	"SUBCORE_VM_FIXED0175_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0175E450g,
	"SUBCORE_VM_FIXED0200_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0200E450g,
	"SUBCORE_VM_FIXED0225_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0225E450g,
	"SUBCORE_VM_FIXED0250_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0250E450g,
	"SUBCORE_VM_FIXED0275_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0275E450g,
	"SUBCORE_VM_FIXED0300_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0300E450g,
	"SUBCORE_VM_FIXED0325_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0325E450g,
	"SUBCORE_VM_FIXED0350_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0350E450g,
	"SUBCORE_VM_FIXED0375_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0375E450g,
	"SUBCORE_VM_FIXED0400_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0400E450g,
	"SUBCORE_VM_FIXED0425_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0425E450g,
	"SUBCORE_VM_FIXED0450_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0450E450g,
	"SUBCORE_VM_FIXED0475_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0475E450g,
	"SUBCORE_VM_FIXED0500_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0500E450g,
	"SUBCORE_VM_FIXED0525_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0525E450g,
	"SUBCORE_VM_FIXED0550_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0550E450g,
	"SUBCORE_VM_FIXED0575_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0575E450g,
	"SUBCORE_VM_FIXED0600_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0600E450g,
	"SUBCORE_VM_FIXED0625_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0625E450g,
	"SUBCORE_VM_FIXED0650_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0650E450g,
	"SUBCORE_VM_FIXED0675_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0675E450g,
	"SUBCORE_VM_FIXED0700_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0700E450g,
	"SUBCORE_VM_FIXED0725_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0725E450g,
	"SUBCORE_VM_FIXED0750_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0750E450g,
	"SUBCORE_VM_FIXED0775_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0775E450g,
	"SUBCORE_VM_FIXED0800_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0800E450g,
	"SUBCORE_VM_FIXED0825_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0825E450g,
	"SUBCORE_VM_FIXED0850_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0850E450g,
	"SUBCORE_VM_FIXED0875_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0875E450g,
	"SUBCORE_VM_FIXED0900_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0900E450g,
	"SUBCORE_VM_FIXED0925_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0925E450g,
	"SUBCORE_VM_FIXED0950_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0950E450g,
	"SUBCORE_VM_FIXED0975_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0975E450g,
	"SUBCORE_VM_FIXED1000_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1000E450g,
	"SUBCORE_VM_FIXED1025_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1025E450g,
	"SUBCORE_VM_FIXED1050_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1050E450g,
	"SUBCORE_VM_FIXED1075_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1075E450g,
	"SUBCORE_VM_FIXED1100_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1100E450g,
	"SUBCORE_VM_FIXED1125_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1125E450g,
	"SUBCORE_VM_FIXED1150_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1150E450g,
	"SUBCORE_VM_FIXED1175_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1175E450g,
	"SUBCORE_VM_FIXED1200_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1200E450g,
	"SUBCORE_VM_FIXED1225_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1225E450g,
	"SUBCORE_VM_FIXED1250_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1250E450g,
	"SUBCORE_VM_FIXED1275_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1275E450g,
	"SUBCORE_VM_FIXED1300_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1300E450g,
	"SUBCORE_VM_FIXED1325_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1325E450g,
	"SUBCORE_VM_FIXED1350_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1350E450g,
	"SUBCORE_VM_FIXED1375_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1375E450g,
	"SUBCORE_VM_FIXED1400_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1400E450g,
	"SUBCORE_VM_FIXED1425_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1425E450g,
	"SUBCORE_VM_FIXED1450_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1450E450g,
	"SUBCORE_VM_FIXED1475_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1475E450g,
	"SUBCORE_VM_FIXED1500_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1500E450g,
	"SUBCORE_VM_FIXED1525_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1525E450g,
	"SUBCORE_VM_FIXED1550_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1550E450g,
	"SUBCORE_VM_FIXED1575_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1575E450g,
	"SUBCORE_VM_FIXED1600_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1600E450g,
	"SUBCORE_VM_FIXED1625_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1625E450g,
	"SUBCORE_VM_FIXED1650_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1650E450g,
	"SUBCORE_VM_FIXED1700_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1700E450g,
	"SUBCORE_VM_FIXED1725_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1725E450g,
	"SUBCORE_VM_FIXED1750_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1750E450g,
	"SUBCORE_VM_FIXED1800_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1800E450g,
	"SUBCORE_VM_FIXED1850_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1850E450g,
	"SUBCORE_VM_FIXED1875_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1875E450g,
	"SUBCORE_VM_FIXED1900_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1900E450g,
	"SUBCORE_VM_FIXED1925_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1925E450g,
	"SUBCORE_VM_FIXED1950_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1950E450g,
	"SUBCORE_VM_FIXED2000_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2000E450g,
	"SUBCORE_VM_FIXED2025_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2025E450g,
	"SUBCORE_VM_FIXED2050_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2050E450g,
	"SUBCORE_VM_FIXED2100_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2100E450g,
	"SUBCORE_VM_FIXED2125_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2125E450g,
	"SUBCORE_VM_FIXED2150_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2150E450g,
	"SUBCORE_VM_FIXED2175_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2175E450g,
	"SUBCORE_VM_FIXED2200_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2200E450g,
	"SUBCORE_VM_FIXED2250_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2250E450g,
	"SUBCORE_VM_FIXED2275_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2275E450g,
	"SUBCORE_VM_FIXED2300_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2300E450g,
	"SUBCORE_VM_FIXED2325_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2325E450g,
	"SUBCORE_VM_FIXED2350_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2350E450g,
	"SUBCORE_VM_FIXED2375_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2375E450g,
	"SUBCORE_VM_FIXED2400_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2400E450g,
	"SUBCORE_VM_FIXED2450_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2450E450g,
	"SUBCORE_VM_FIXED2475_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2475E450g,
	"SUBCORE_VM_FIXED2500_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2500E450g,
	"SUBCORE_VM_FIXED2550_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2550E450g,
	"SUBCORE_VM_FIXED2600_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2600E450g,
	"SUBCORE_VM_FIXED2625_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2625E450g,
	"SUBCORE_VM_FIXED2650_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2650E450g,
	"SUBCORE_VM_FIXED2700_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2700E450g,
	"SUBCORE_VM_FIXED2750_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2750E450g,
	"SUBCORE_VM_FIXED2775_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2775E450g,
	"SUBCORE_VM_FIXED2800_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2800E450g,
	"SUBCORE_VM_FIXED2850_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2850E450g,
	"SUBCORE_VM_FIXED2875_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2875E450g,
	"SUBCORE_VM_FIXED2900_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2900E450g,
	"SUBCORE_VM_FIXED2925_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2925E450g,
	"SUBCORE_VM_FIXED2950_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2950E450g,
	"SUBCORE_VM_FIXED2975_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2975E450g,
	"SUBCORE_VM_FIXED3000_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3000E450g,
	"SUBCORE_VM_FIXED3025_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3025E450g,
	"SUBCORE_VM_FIXED3050_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3050E450g,
	"SUBCORE_VM_FIXED3075_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3075E450g,
	"SUBCORE_VM_FIXED3100_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3100E450g,
	"SUBCORE_VM_FIXED3125_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3125E450g,
	"SUBCORE_VM_FIXED3150_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3150E450g,
	"SUBCORE_VM_FIXED3200_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3200E450g,
	"SUBCORE_VM_FIXED3225_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3225E450g,
	"SUBCORE_VM_FIXED3250_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3250E450g,
	"SUBCORE_VM_FIXED3300_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3300E450g,
	"SUBCORE_VM_FIXED3325_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3325E450g,
	"SUBCORE_VM_FIXED3375_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3375E450g,
	"SUBCORE_VM_FIXED3400_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3400E450g,
	"SUBCORE_VM_FIXED3450_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3450E450g,
	"SUBCORE_VM_FIXED3500_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3500E450g,
	"SUBCORE_VM_FIXED3525_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3525E450g,
	"SUBCORE_VM_FIXED3575_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3575E450g,
	"SUBCORE_VM_FIXED3600_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3600E450g,
	"SUBCORE_VM_FIXED3625_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3625E450g,
	"SUBCORE_VM_FIXED3675_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3675E450g,
	"SUBCORE_VM_FIXED3700_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3700E450g,
	"SUBCORE_VM_FIXED3750_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3750E450g,
	"SUBCORE_VM_FIXED3800_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3800E450g,
	"SUBCORE_VM_FIXED3825_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3825E450g,
	"SUBCORE_VM_FIXED3850_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3850E450g,
	"SUBCORE_VM_FIXED3875_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3875E450g,
	"SUBCORE_VM_FIXED3900_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3900E450g,
	"SUBCORE_VM_FIXED3975_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3975E450g,
	"SUBCORE_VM_FIXED4000_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4000E450g,
	"SUBCORE_VM_FIXED4025_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4025E450g,
	"SUBCORE_VM_FIXED4050_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4050E450g,
	"SUBCORE_VM_FIXED4100_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4100E450g,
	"SUBCORE_VM_FIXED4125_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4125E450g,
	"SUBCORE_VM_FIXED4200_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4200E450g,
	"SUBCORE_VM_FIXED4225_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4225E450g,
	"SUBCORE_VM_FIXED4250_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4250E450g,
	"SUBCORE_VM_FIXED4275_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4275E450g,
	"SUBCORE_VM_FIXED4300_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4300E450g,
	"SUBCORE_VM_FIXED4350_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4350E450g,
	"SUBCORE_VM_FIXED4375_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4375E450g,
	"SUBCORE_VM_FIXED4400_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4400E450g,
	"SUBCORE_VM_FIXED4425_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4425E450g,
	"SUBCORE_VM_FIXED4500_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4500E450g,
	"SUBCORE_VM_FIXED4550_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4550E450g,
	"SUBCORE_VM_FIXED4575_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4575E450g,
	"SUBCORE_VM_FIXED4600_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4600E450g,
	"SUBCORE_VM_FIXED4625_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4625E450g,
	"SUBCORE_VM_FIXED4650_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4650E450g,
	"SUBCORE_VM_FIXED4675_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4675E450g,
	"SUBCORE_VM_FIXED4700_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4700E450g,
	"SUBCORE_VM_FIXED4725_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4725E450g,
	"SUBCORE_VM_FIXED4750_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4750E450g,
	"SUBCORE_VM_FIXED4800_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4800E450g,
	"SUBCORE_VM_FIXED4875_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4875E450g,
	"SUBCORE_VM_FIXED4900_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4900E450g,
	"SUBCORE_VM_FIXED4950_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4950E450g,
	"SUBCORE_VM_FIXED5000_E4_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed5000E450g,
	"SUBCORE_VM_FIXED0020_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0020A150g,
	"SUBCORE_VM_FIXED0040_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0040A150g,
	"SUBCORE_VM_FIXED0060_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0060A150g,
	"SUBCORE_VM_FIXED0080_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0080A150g,
	"SUBCORE_VM_FIXED0100_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0100A150g,
	"SUBCORE_VM_FIXED0120_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0120A150g,
	"SUBCORE_VM_FIXED0140_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0140A150g,
	"SUBCORE_VM_FIXED0160_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0160A150g,
	"SUBCORE_VM_FIXED0180_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0180A150g,
	"SUBCORE_VM_FIXED0200_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0200A150g,
	"SUBCORE_VM_FIXED0220_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0220A150g,
	"SUBCORE_VM_FIXED0240_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0240A150g,
	"SUBCORE_VM_FIXED0260_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0260A150g,
	"SUBCORE_VM_FIXED0280_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0280A150g,
	"SUBCORE_VM_FIXED0300_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0300A150g,
	"SUBCORE_VM_FIXED0320_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0320A150g,
	"SUBCORE_VM_FIXED0340_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0340A150g,
	"SUBCORE_VM_FIXED0360_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0360A150g,
	"SUBCORE_VM_FIXED0380_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0380A150g,
	"SUBCORE_VM_FIXED0400_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0400A150g,
	"SUBCORE_VM_FIXED0420_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0420A150g,
	"SUBCORE_VM_FIXED0440_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0440A150g,
	"SUBCORE_VM_FIXED0460_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0460A150g,
	"SUBCORE_VM_FIXED0480_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0480A150g,
	"SUBCORE_VM_FIXED0500_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0500A150g,
	"SUBCORE_VM_FIXED0520_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0520A150g,
	"SUBCORE_VM_FIXED0540_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0540A150g,
	"SUBCORE_VM_FIXED0560_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0560A150g,
	"SUBCORE_VM_FIXED0580_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0580A150g,
	"SUBCORE_VM_FIXED0600_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0600A150g,
	"SUBCORE_VM_FIXED0620_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0620A150g,
	"SUBCORE_VM_FIXED0640_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0640A150g,
	"SUBCORE_VM_FIXED0660_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0660A150g,
	"SUBCORE_VM_FIXED0680_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0680A150g,
	"SUBCORE_VM_FIXED0700_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0700A150g,
	"SUBCORE_VM_FIXED0720_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0720A150g,
	"SUBCORE_VM_FIXED0740_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0740A150g,
	"SUBCORE_VM_FIXED0760_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0760A150g,
	"SUBCORE_VM_FIXED0780_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0780A150g,
	"SUBCORE_VM_FIXED0800_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0800A150g,
	"SUBCORE_VM_FIXED0820_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0820A150g,
	"SUBCORE_VM_FIXED0840_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0840A150g,
	"SUBCORE_VM_FIXED0860_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0860A150g,
	"SUBCORE_VM_FIXED0880_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0880A150g,
	"SUBCORE_VM_FIXED0900_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0900A150g,
	"SUBCORE_VM_FIXED0920_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0920A150g,
	"SUBCORE_VM_FIXED0940_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0940A150g,
	"SUBCORE_VM_FIXED0960_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0960A150g,
	"SUBCORE_VM_FIXED0980_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0980A150g,
	"SUBCORE_VM_FIXED1000_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1000A150g,
	"SUBCORE_VM_FIXED1020_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1020A150g,
	"SUBCORE_VM_FIXED1040_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1040A150g,
	"SUBCORE_VM_FIXED1060_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1060A150g,
	"SUBCORE_VM_FIXED1080_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1080A150g,
	"SUBCORE_VM_FIXED1100_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1100A150g,
	"SUBCORE_VM_FIXED1120_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1120A150g,
	"SUBCORE_VM_FIXED1140_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1140A150g,
	"SUBCORE_VM_FIXED1160_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1160A150g,
	"SUBCORE_VM_FIXED1180_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1180A150g,
	"SUBCORE_VM_FIXED1200_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1200A150g,
	"SUBCORE_VM_FIXED1220_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1220A150g,
	"SUBCORE_VM_FIXED1240_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1240A150g,
	"SUBCORE_VM_FIXED1260_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1260A150g,
	"SUBCORE_VM_FIXED1280_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1280A150g,
	"SUBCORE_VM_FIXED1300_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1300A150g,
	"SUBCORE_VM_FIXED1320_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1320A150g,
	"SUBCORE_VM_FIXED1340_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1340A150g,
	"SUBCORE_VM_FIXED1360_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1360A150g,
	"SUBCORE_VM_FIXED1380_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1380A150g,
	"SUBCORE_VM_FIXED1400_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1400A150g,
	"SUBCORE_VM_FIXED1420_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1420A150g,
	"SUBCORE_VM_FIXED1440_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1440A150g,
	"SUBCORE_VM_FIXED1460_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1460A150g,
	"SUBCORE_VM_FIXED1480_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1480A150g,
	"SUBCORE_VM_FIXED1500_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1500A150g,
	"SUBCORE_VM_FIXED1520_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1520A150g,
	"SUBCORE_VM_FIXED1540_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1540A150g,
	"SUBCORE_VM_FIXED1560_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1560A150g,
	"SUBCORE_VM_FIXED1580_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1580A150g,
	"SUBCORE_VM_FIXED1600_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1600A150g,
	"SUBCORE_VM_FIXED1620_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1620A150g,
	"SUBCORE_VM_FIXED1640_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1640A150g,
	"SUBCORE_VM_FIXED1660_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1660A150g,
	"SUBCORE_VM_FIXED1680_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1680A150g,
	"SUBCORE_VM_FIXED1700_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1700A150g,
	"SUBCORE_VM_FIXED1720_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1720A150g,
	"SUBCORE_VM_FIXED1740_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1740A150g,
	"SUBCORE_VM_FIXED1760_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1760A150g,
	"SUBCORE_VM_FIXED1780_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1780A150g,
	"SUBCORE_VM_FIXED1800_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1800A150g,
	"SUBCORE_VM_FIXED1820_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1820A150g,
	"SUBCORE_VM_FIXED1840_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1840A150g,
	"SUBCORE_VM_FIXED1860_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1860A150g,
	"SUBCORE_VM_FIXED1880_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1880A150g,
	"SUBCORE_VM_FIXED1900_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1900A150g,
	"SUBCORE_VM_FIXED1920_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1920A150g,
	"SUBCORE_VM_FIXED1940_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1940A150g,
	"SUBCORE_VM_FIXED1960_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1960A150g,
	"SUBCORE_VM_FIXED1980_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1980A150g,
	"SUBCORE_VM_FIXED2000_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2000A150g,
	"SUBCORE_VM_FIXED2020_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2020A150g,
	"SUBCORE_VM_FIXED2040_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2040A150g,
	"SUBCORE_VM_FIXED2060_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2060A150g,
	"SUBCORE_VM_FIXED2080_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2080A150g,
	"SUBCORE_VM_FIXED2100_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2100A150g,
	"SUBCORE_VM_FIXED2120_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2120A150g,
	"SUBCORE_VM_FIXED2140_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2140A150g,
	"SUBCORE_VM_FIXED2160_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2160A150g,
	"SUBCORE_VM_FIXED2180_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2180A150g,
	"SUBCORE_VM_FIXED2200_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2200A150g,
	"SUBCORE_VM_FIXED2220_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2220A150g,
	"SUBCORE_VM_FIXED2240_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2240A150g,
	"SUBCORE_VM_FIXED2260_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2260A150g,
	"SUBCORE_VM_FIXED2280_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2280A150g,
	"SUBCORE_VM_FIXED2300_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2300A150g,
	"SUBCORE_VM_FIXED2320_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2320A150g,
	"SUBCORE_VM_FIXED2340_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2340A150g,
	"SUBCORE_VM_FIXED2360_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2360A150g,
	"SUBCORE_VM_FIXED2380_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2380A150g,
	"SUBCORE_VM_FIXED2400_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2400A150g,
	"SUBCORE_VM_FIXED2420_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2420A150g,
	"SUBCORE_VM_FIXED2440_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2440A150g,
	"SUBCORE_VM_FIXED2460_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2460A150g,
	"SUBCORE_VM_FIXED2480_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2480A150g,
	"SUBCORE_VM_FIXED2500_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2500A150g,
	"SUBCORE_VM_FIXED2520_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2520A150g,
	"SUBCORE_VM_FIXED2540_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2540A150g,
	"SUBCORE_VM_FIXED2560_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2560A150g,
	"SUBCORE_VM_FIXED2580_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2580A150g,
	"SUBCORE_VM_FIXED2600_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2600A150g,
	"SUBCORE_VM_FIXED2620_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2620A150g,
	"SUBCORE_VM_FIXED2640_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2640A150g,
	"SUBCORE_VM_FIXED2660_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2660A150g,
	"SUBCORE_VM_FIXED2680_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2680A150g,
	"SUBCORE_VM_FIXED2700_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2700A150g,
	"SUBCORE_VM_FIXED2720_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2720A150g,
	"SUBCORE_VM_FIXED2740_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2740A150g,
	"SUBCORE_VM_FIXED2760_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2760A150g,
	"SUBCORE_VM_FIXED2780_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2780A150g,
	"SUBCORE_VM_FIXED2800_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2800A150g,
	"SUBCORE_VM_FIXED2820_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2820A150g,
	"SUBCORE_VM_FIXED2840_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2840A150g,
	"SUBCORE_VM_FIXED2860_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2860A150g,
	"SUBCORE_VM_FIXED2880_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2880A150g,
	"SUBCORE_VM_FIXED2900_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2900A150g,
	"SUBCORE_VM_FIXED2920_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2920A150g,
	"SUBCORE_VM_FIXED2940_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2940A150g,
	"SUBCORE_VM_FIXED2960_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2960A150g,
	"SUBCORE_VM_FIXED2980_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2980A150g,
	"SUBCORE_VM_FIXED3000_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3000A150g,
	"SUBCORE_VM_FIXED3020_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3020A150g,
	"SUBCORE_VM_FIXED3040_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3040A150g,
	"SUBCORE_VM_FIXED3060_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3060A150g,
	"SUBCORE_VM_FIXED3080_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3080A150g,
	"SUBCORE_VM_FIXED3100_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3100A150g,
	"SUBCORE_VM_FIXED3120_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3120A150g,
	"SUBCORE_VM_FIXED3140_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3140A150g,
	"SUBCORE_VM_FIXED3160_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3160A150g,
	"SUBCORE_VM_FIXED3180_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3180A150g,
	"SUBCORE_VM_FIXED3200_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3200A150g,
	"SUBCORE_VM_FIXED3220_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3220A150g,
	"SUBCORE_VM_FIXED3240_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3240A150g,
	"SUBCORE_VM_FIXED3260_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3260A150g,
	"SUBCORE_VM_FIXED3280_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3280A150g,
	"SUBCORE_VM_FIXED3300_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3300A150g,
	"SUBCORE_VM_FIXED3320_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3320A150g,
	"SUBCORE_VM_FIXED3340_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3340A150g,
	"SUBCORE_VM_FIXED3360_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3360A150g,
	"SUBCORE_VM_FIXED3380_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3380A150g,
	"SUBCORE_VM_FIXED3400_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3400A150g,
	"SUBCORE_VM_FIXED3420_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3420A150g,
	"SUBCORE_VM_FIXED3440_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3440A150g,
	"SUBCORE_VM_FIXED3460_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3460A150g,
	"SUBCORE_VM_FIXED3480_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3480A150g,
	"SUBCORE_VM_FIXED3500_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3500A150g,
	"SUBCORE_VM_FIXED3520_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3520A150g,
	"SUBCORE_VM_FIXED3540_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3540A150g,
	"SUBCORE_VM_FIXED3560_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3560A150g,
	"SUBCORE_VM_FIXED3580_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3580A150g,
	"SUBCORE_VM_FIXED3600_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3600A150g,
	"SUBCORE_VM_FIXED3620_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3620A150g,
	"SUBCORE_VM_FIXED3640_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3640A150g,
	"SUBCORE_VM_FIXED3660_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3660A150g,
	"SUBCORE_VM_FIXED3680_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3680A150g,
	"SUBCORE_VM_FIXED3700_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3700A150g,
	"SUBCORE_VM_FIXED3720_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3720A150g,
	"SUBCORE_VM_FIXED3740_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3740A150g,
	"SUBCORE_VM_FIXED3760_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3760A150g,
	"SUBCORE_VM_FIXED3780_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3780A150g,
	"SUBCORE_VM_FIXED3800_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3800A150g,
	"SUBCORE_VM_FIXED3820_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3820A150g,
	"SUBCORE_VM_FIXED3840_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3840A150g,
	"SUBCORE_VM_FIXED3860_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3860A150g,
	"SUBCORE_VM_FIXED3880_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3880A150g,
	"SUBCORE_VM_FIXED3900_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3900A150g,
	"SUBCORE_VM_FIXED3920_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3920A150g,
	"SUBCORE_VM_FIXED3940_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3940A150g,
	"SUBCORE_VM_FIXED3960_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3960A150g,
	"SUBCORE_VM_FIXED3980_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3980A150g,
	"SUBCORE_VM_FIXED4000_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4000A150g,
	"SUBCORE_VM_FIXED4020_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4020A150g,
	"SUBCORE_VM_FIXED4040_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4040A150g,
	"SUBCORE_VM_FIXED4060_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4060A150g,
	"SUBCORE_VM_FIXED4080_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4080A150g,
	"SUBCORE_VM_FIXED4100_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4100A150g,
	"SUBCORE_VM_FIXED4120_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4120A150g,
	"SUBCORE_VM_FIXED4140_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4140A150g,
	"SUBCORE_VM_FIXED4160_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4160A150g,
	"SUBCORE_VM_FIXED4180_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4180A150g,
	"SUBCORE_VM_FIXED4200_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4200A150g,
	"SUBCORE_VM_FIXED4220_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4220A150g,
	"SUBCORE_VM_FIXED4240_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4240A150g,
	"SUBCORE_VM_FIXED4260_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4260A150g,
	"SUBCORE_VM_FIXED4280_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4280A150g,
	"SUBCORE_VM_FIXED4300_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4300A150g,
	"SUBCORE_VM_FIXED4320_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4320A150g,
	"SUBCORE_VM_FIXED4340_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4340A150g,
	"SUBCORE_VM_FIXED4360_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4360A150g,
	"SUBCORE_VM_FIXED4380_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4380A150g,
	"SUBCORE_VM_FIXED4400_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4400A150g,
	"SUBCORE_VM_FIXED4420_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4420A150g,
	"SUBCORE_VM_FIXED4440_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4440A150g,
	"SUBCORE_VM_FIXED4460_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4460A150g,
	"SUBCORE_VM_FIXED4480_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4480A150g,
	"SUBCORE_VM_FIXED4500_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4500A150g,
	"SUBCORE_VM_FIXED4520_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4520A150g,
	"SUBCORE_VM_FIXED4540_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4540A150g,
	"SUBCORE_VM_FIXED4560_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4560A150g,
	"SUBCORE_VM_FIXED4580_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4580A150g,
	"SUBCORE_VM_FIXED4600_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4600A150g,
	"SUBCORE_VM_FIXED4620_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4620A150g,
	"SUBCORE_VM_FIXED4640_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4640A150g,
	"SUBCORE_VM_FIXED4660_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4660A150g,
	"SUBCORE_VM_FIXED4680_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4680A150g,
	"SUBCORE_VM_FIXED4700_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4700A150g,
	"SUBCORE_VM_FIXED4720_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4720A150g,
	"SUBCORE_VM_FIXED4740_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4740A150g,
	"SUBCORE_VM_FIXED4760_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4760A150g,
	"SUBCORE_VM_FIXED4780_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4780A150g,
	"SUBCORE_VM_FIXED4800_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4800A150g,
	"SUBCORE_VM_FIXED4820_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4820A150g,
	"SUBCORE_VM_FIXED4840_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4840A150g,
	"SUBCORE_VM_FIXED4860_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4860A150g,
	"SUBCORE_VM_FIXED4880_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4880A150g,
	"SUBCORE_VM_FIXED4900_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4900A150g,
	"SUBCORE_VM_FIXED4920_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4920A150g,
	"SUBCORE_VM_FIXED4940_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4940A150g,
	"SUBCORE_VM_FIXED4960_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4960A150g,
	"SUBCORE_VM_FIXED4980_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4980A150g,
	"SUBCORE_VM_FIXED5000_A1_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed5000A150g,
	"SUBCORE_VM_FIXED0090_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0090X950g,
	"SUBCORE_VM_FIXED0180_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0180X950g,
	"SUBCORE_VM_FIXED0270_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0270X950g,
	"SUBCORE_VM_FIXED0360_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0360X950g,
	"SUBCORE_VM_FIXED0450_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0450X950g,
	"SUBCORE_VM_FIXED0540_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0540X950g,
	"SUBCORE_VM_FIXED0630_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0630X950g,
	"SUBCORE_VM_FIXED0720_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0720X950g,
	"SUBCORE_VM_FIXED0810_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0810X950g,
	"SUBCORE_VM_FIXED0900_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0900X950g,
	"SUBCORE_VM_FIXED0990_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed0990X950g,
	"SUBCORE_VM_FIXED1080_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1080X950g,
	"SUBCORE_VM_FIXED1170_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1170X950g,
	"SUBCORE_VM_FIXED1260_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1260X950g,
	"SUBCORE_VM_FIXED1350_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1350X950g,
	"SUBCORE_VM_FIXED1440_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1440X950g,
	"SUBCORE_VM_FIXED1530_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1530X950g,
	"SUBCORE_VM_FIXED1620_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1620X950g,
	"SUBCORE_VM_FIXED1710_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1710X950g,
	"SUBCORE_VM_FIXED1800_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1800X950g,
	"SUBCORE_VM_FIXED1890_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1890X950g,
	"SUBCORE_VM_FIXED1980_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed1980X950g,
	"SUBCORE_VM_FIXED2070_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2070X950g,
	"SUBCORE_VM_FIXED2160_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2160X950g,
	"SUBCORE_VM_FIXED2250_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2250X950g,
	"SUBCORE_VM_FIXED2340_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2340X950g,
	"SUBCORE_VM_FIXED2430_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2430X950g,
	"SUBCORE_VM_FIXED2520_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2520X950g,
	"SUBCORE_VM_FIXED2610_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2610X950g,
	"SUBCORE_VM_FIXED2700_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2700X950g,
	"SUBCORE_VM_FIXED2790_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2790X950g,
	"SUBCORE_VM_FIXED2880_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2880X950g,
	"SUBCORE_VM_FIXED2970_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed2970X950g,
	"SUBCORE_VM_FIXED3060_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3060X950g,
	"SUBCORE_VM_FIXED3150_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3150X950g,
	"SUBCORE_VM_FIXED3240_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3240X950g,
	"SUBCORE_VM_FIXED3330_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3330X950g,
	"SUBCORE_VM_FIXED3420_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3420X950g,
	"SUBCORE_VM_FIXED3510_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3510X950g,
	"SUBCORE_VM_FIXED3600_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3600X950g,
	"SUBCORE_VM_FIXED3690_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3690X950g,
	"SUBCORE_VM_FIXED3780_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3780X950g,
	"SUBCORE_VM_FIXED3870_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3870X950g,
	"SUBCORE_VM_FIXED3960_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed3960X950g,
	"SUBCORE_VM_FIXED4050_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4050X950g,
	"SUBCORE_VM_FIXED4140_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4140X950g,
	"SUBCORE_VM_FIXED4230_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4230X950g,
	"SUBCORE_VM_FIXED4320_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4320X950g,
	"SUBCORE_VM_FIXED4410_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4410X950g,
	"SUBCORE_VM_FIXED4500_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4500X950g,
	"SUBCORE_VM_FIXED4590_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4590X950g,
	"SUBCORE_VM_FIXED4680_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4680X950g,
	"SUBCORE_VM_FIXED4770_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4770X950g,
	"SUBCORE_VM_FIXED4860_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4860X950g,
	"SUBCORE_VM_FIXED4950_X9_50G":  CreateInternalVnicAttachmentDetailsVnicShapeSubcoreVmFixed4950X950g,
	"DYNAMIC_A1_50G":               CreateInternalVnicAttachmentDetailsVnicShapeDynamicA150g,
	"FIXED0040_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0040A150g,
	"FIXED0100_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0100A150g,
	"FIXED0200_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0200A150g,
	"FIXED0300_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0300A150g,
	"FIXED0400_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0400A150g,
	"FIXED0500_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0500A150g,
	"FIXED0600_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0600A150g,
	"FIXED0700_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0700A150g,
	"FIXED0800_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0800A150g,
	"FIXED0900_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0900A150g,
	"FIXED1000_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1000A150g,
	"FIXED1100_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1100A150g,
	"FIXED1200_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1200A150g,
	"FIXED1300_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1300A150g,
	"FIXED1400_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1400A150g,
	"FIXED1500_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1500A150g,
	"FIXED1600_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1600A150g,
	"FIXED1700_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1700A150g,
	"FIXED1800_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1800A150g,
	"FIXED1900_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1900A150g,
	"FIXED2000_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2000A150g,
	"FIXED2100_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2100A150g,
	"FIXED2200_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2200A150g,
	"FIXED2300_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2300A150g,
	"FIXED2400_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2400A150g,
	"FIXED2500_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2500A150g,
	"FIXED2600_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2600A150g,
	"FIXED2700_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2700A150g,
	"FIXED2800_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2800A150g,
	"FIXED2900_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2900A150g,
	"FIXED3000_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3000A150g,
	"FIXED3100_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3100A150g,
	"FIXED3200_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3200A150g,
	"FIXED3300_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3300A150g,
	"FIXED3400_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3400A150g,
	"FIXED3500_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3500A150g,
	"FIXED3600_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3600A150g,
	"FIXED3700_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3700A150g,
	"FIXED3800_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3800A150g,
	"FIXED3900_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3900A150g,
	"FIXED4000_A1_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed4000A150g,
	"ENTIREHOST_A1_50G":            CreateInternalVnicAttachmentDetailsVnicShapeEntirehostA150g,
	"DYNAMIC_X9_50G":               CreateInternalVnicAttachmentDetailsVnicShapeDynamicX950g,
	"FIXED0040_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0040X950g,
	"FIXED0400_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0400X950g,
	"FIXED0800_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed0800X950g,
	"FIXED1200_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1200X950g,
	"FIXED1600_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed1600X950g,
	"FIXED2000_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2000X950g,
	"FIXED2400_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2400X950g,
	"FIXED2800_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed2800X950g,
	"FIXED3200_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3200X950g,
	"FIXED3600_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed3600X950g,
	"FIXED4000_X9_50G":             CreateInternalVnicAttachmentDetailsVnicShapeFixed4000X950g,
	"STANDARD_VM_FIXED0100_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0100X950g,
	"STANDARD_VM_FIXED0200_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0200X950g,
	"STANDARD_VM_FIXED0300_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0300X950g,
	"STANDARD_VM_FIXED0400_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0400X950g,
	"STANDARD_VM_FIXED0500_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0500X950g,
	"STANDARD_VM_FIXED0600_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0600X950g,
	"STANDARD_VM_FIXED0700_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0700X950g,
	"STANDARD_VM_FIXED0800_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0800X950g,
	"STANDARD_VM_FIXED0900_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed0900X950g,
	"STANDARD_VM_FIXED1000_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1000X950g,
	"STANDARD_VM_FIXED1100_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1100X950g,
	"STANDARD_VM_FIXED1200_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1200X950g,
	"STANDARD_VM_FIXED1300_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1300X950g,
	"STANDARD_VM_FIXED1400_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1400X950g,
	"STANDARD_VM_FIXED1500_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1500X950g,
	"STANDARD_VM_FIXED1600_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1600X950g,
	"STANDARD_VM_FIXED1700_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1700X950g,
	"STANDARD_VM_FIXED1800_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1800X950g,
	"STANDARD_VM_FIXED1900_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed1900X950g,
	"STANDARD_VM_FIXED2000_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2000X950g,
	"STANDARD_VM_FIXED2100_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2100X950g,
	"STANDARD_VM_FIXED2200_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2200X950g,
	"STANDARD_VM_FIXED2300_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2300X950g,
	"STANDARD_VM_FIXED2400_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2400X950g,
	"STANDARD_VM_FIXED2500_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2500X950g,
	"STANDARD_VM_FIXED2600_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2600X950g,
	"STANDARD_VM_FIXED2700_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2700X950g,
	"STANDARD_VM_FIXED2800_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2800X950g,
	"STANDARD_VM_FIXED2900_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed2900X950g,
	"STANDARD_VM_FIXED3000_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3000X950g,
	"STANDARD_VM_FIXED3100_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3100X950g,
	"STANDARD_VM_FIXED3200_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3200X950g,
	"STANDARD_VM_FIXED3300_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3300X950g,
	"STANDARD_VM_FIXED3400_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3400X950g,
	"STANDARD_VM_FIXED3500_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3500X950g,
	"STANDARD_VM_FIXED3600_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3600X950g,
	"STANDARD_VM_FIXED3700_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3700X950g,
	"STANDARD_VM_FIXED3800_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3800X950g,
	"STANDARD_VM_FIXED3900_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed3900X950g,
	"STANDARD_VM_FIXED4000_X9_50G": CreateInternalVnicAttachmentDetailsVnicShapeStandardVmFixed4000X950g,
	"ENTIREHOST_X9_50G":            CreateInternalVnicAttachmentDetailsVnicShapeEntirehostX950g,
}

// GetCreateInternalVnicAttachmentDetailsVnicShapeEnumValues Enumerates the set of values for CreateInternalVnicAttachmentDetailsVnicShapeEnum
func GetCreateInternalVnicAttachmentDetailsVnicShapeEnumValues() []CreateInternalVnicAttachmentDetailsVnicShapeEnum {
	values := make([]CreateInternalVnicAttachmentDetailsVnicShapeEnum, 0)
	for _, v := range mappingCreateInternalVnicAttachmentDetailsVnicShapeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateInternalVnicAttachmentDetailsVnicShapeEnumStringValues Enumerates the set of values in String for CreateInternalVnicAttachmentDetailsVnicShapeEnum
func GetCreateInternalVnicAttachmentDetailsVnicShapeEnumStringValues() []string {
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

// CreateInternalVnicAttachmentDetailsLaunchTypeEnum Enum with underlying type: string
type CreateInternalVnicAttachmentDetailsLaunchTypeEnum string

// Set of constants representing the allowable values for CreateInternalVnicAttachmentDetailsLaunchTypeEnum
const (
	CreateInternalVnicAttachmentDetailsLaunchTypeMarketplace CreateInternalVnicAttachmentDetailsLaunchTypeEnum = "MARKETPLACE"
	CreateInternalVnicAttachmentDetailsLaunchTypeStandard    CreateInternalVnicAttachmentDetailsLaunchTypeEnum = "STANDARD"
)

var mappingCreateInternalVnicAttachmentDetailsLaunchTypeEnum = map[string]CreateInternalVnicAttachmentDetailsLaunchTypeEnum{
	"MARKETPLACE": CreateInternalVnicAttachmentDetailsLaunchTypeMarketplace,
	"STANDARD":    CreateInternalVnicAttachmentDetailsLaunchTypeStandard,
}

// GetCreateInternalVnicAttachmentDetailsLaunchTypeEnumValues Enumerates the set of values for CreateInternalVnicAttachmentDetailsLaunchTypeEnum
func GetCreateInternalVnicAttachmentDetailsLaunchTypeEnumValues() []CreateInternalVnicAttachmentDetailsLaunchTypeEnum {
	values := make([]CreateInternalVnicAttachmentDetailsLaunchTypeEnum, 0)
	for _, v := range mappingCreateInternalVnicAttachmentDetailsLaunchTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateInternalVnicAttachmentDetailsLaunchTypeEnumStringValues Enumerates the set of values in String for CreateInternalVnicAttachmentDetailsLaunchTypeEnum
func GetCreateInternalVnicAttachmentDetailsLaunchTypeEnumStringValues() []string {
	return []string{
		"MARKETPLACE",
		"STANDARD",
	}
}
