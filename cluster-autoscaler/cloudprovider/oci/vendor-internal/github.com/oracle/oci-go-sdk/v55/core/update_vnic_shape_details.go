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

// UpdateVnicShapeDetails This structure is used when updating the shape of VNIC in VNIC attachment.
type UpdateVnicShapeDetails struct {

	// VNIC whose attachments need to be updated to the destination vnic shape.
	VnicId *string `mandatory:"true" json:"vnicId"`

	// Shape of VNIC that will be used to update VNIC attachment.
	VnicShape UpdateVnicShapeDetailsVnicShapeEnum `mandatory:"true" json:"vnicShape"`
}

func (m UpdateVnicShapeDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateVnicShapeDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingUpdateVnicShapeDetailsVnicShapeEnum[string(m.VnicShape)]; !ok && m.VnicShape != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for VnicShape: %s. Supported values are: %s.", m.VnicShape, strings.Join(GetUpdateVnicShapeDetailsVnicShapeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UpdateVnicShapeDetailsVnicShapeEnum Enum with underlying type: string
type UpdateVnicShapeDetailsVnicShapeEnum string

// Set of constants representing the allowable values for UpdateVnicShapeDetailsVnicShapeEnum
const (
	UpdateVnicShapeDetailsVnicShapeDynamic                    UpdateVnicShapeDetailsVnicShapeEnum = "DYNAMIC"
	UpdateVnicShapeDetailsVnicShapeFixed0040                  UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0040"
	UpdateVnicShapeDetailsVnicShapeFixed0060                  UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0060"
	UpdateVnicShapeDetailsVnicShapeFixed0060Psm               UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0060_PSM"
	UpdateVnicShapeDetailsVnicShapeFixed0100                  UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0100"
	UpdateVnicShapeDetailsVnicShapeFixed0120                  UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0120"
	UpdateVnicShapeDetailsVnicShapeFixed01202x                UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0120_2X"
	UpdateVnicShapeDetailsVnicShapeFixed0200                  UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0200"
	UpdateVnicShapeDetailsVnicShapeFixed0240                  UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0240"
	UpdateVnicShapeDetailsVnicShapeFixed0480                  UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0480"
	UpdateVnicShapeDetailsVnicShapeEntirehost                 UpdateVnicShapeDetailsVnicShapeEnum = "ENTIREHOST"
	UpdateVnicShapeDetailsVnicShapeDynamic25g                 UpdateVnicShapeDetailsVnicShapeEnum = "DYNAMIC_25G"
	UpdateVnicShapeDetailsVnicShapeFixed004025g               UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0040_25G"
	UpdateVnicShapeDetailsVnicShapeFixed010025g               UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0100_25G"
	UpdateVnicShapeDetailsVnicShapeFixed020025g               UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0200_25G"
	UpdateVnicShapeDetailsVnicShapeFixed040025g               UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0400_25G"
	UpdateVnicShapeDetailsVnicShapeFixed080025g               UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0800_25G"
	UpdateVnicShapeDetailsVnicShapeFixed160025g               UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1600_25G"
	UpdateVnicShapeDetailsVnicShapeFixed240025g               UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2400_25G"
	UpdateVnicShapeDetailsVnicShapeEntirehost25g              UpdateVnicShapeDetailsVnicShapeEnum = "ENTIREHOST_25G"
	UpdateVnicShapeDetailsVnicShapeDynamicE125g               UpdateVnicShapeDetailsVnicShapeEnum = "DYNAMIC_E1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0040E125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0040_E1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0070E125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0070_E1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0140E125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0140_E1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0280E125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0280_E1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0560E125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0560_E1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed1120E125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1120_E1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed1680E125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1680_E1_25G"
	UpdateVnicShapeDetailsVnicShapeEntirehostE125g            UpdateVnicShapeDetailsVnicShapeEnum = "ENTIREHOST_E1_25G"
	UpdateVnicShapeDetailsVnicShapeDynamicB125g               UpdateVnicShapeDetailsVnicShapeEnum = "DYNAMIC_B1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0040B125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0040_B1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0060B125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0060_B1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0120B125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0120_B1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0240B125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0240_B1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0480B125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0480_B1_25G"
	UpdateVnicShapeDetailsVnicShapeFixed0960B125g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0960_B1_25G"
	UpdateVnicShapeDetailsVnicShapeEntirehostB125g            UpdateVnicShapeDetailsVnicShapeEnum = "ENTIREHOST_B1_25G"
	UpdateVnicShapeDetailsVnicShapeMicroVmFixed0048E125g      UpdateVnicShapeDetailsVnicShapeEnum = "MICRO_VM_FIXED0048_E1_25G"
	UpdateVnicShapeDetailsVnicShapeMicroLbFixed0001E125g      UpdateVnicShapeDetailsVnicShapeEnum = "MICRO_LB_FIXED0001_E1_25G"
	UpdateVnicShapeDetailsVnicShapeVnicaasFixed0200           UpdateVnicShapeDetailsVnicShapeEnum = "VNICAAS_FIXED0200"
	UpdateVnicShapeDetailsVnicShapeVnicaasFixed0400           UpdateVnicShapeDetailsVnicShapeEnum = "VNICAAS_FIXED0400"
	UpdateVnicShapeDetailsVnicShapeVnicaasFixed0700           UpdateVnicShapeDetailsVnicShapeEnum = "VNICAAS_FIXED0700"
	UpdateVnicShapeDetailsVnicShapeVnicaasNlbApproved10g      UpdateVnicShapeDetailsVnicShapeEnum = "VNICAAS_NLB_APPROVED_10G"
	UpdateVnicShapeDetailsVnicShapeVnicaasNlbApproved25g      UpdateVnicShapeDetailsVnicShapeEnum = "VNICAAS_NLB_APPROVED_25G"
	UpdateVnicShapeDetailsVnicShapeVnicaasTelesis25g          UpdateVnicShapeDetailsVnicShapeEnum = "VNICAAS_TELESIS_25G"
	UpdateVnicShapeDetailsVnicShapeVnicaasTelesis10g          UpdateVnicShapeDetailsVnicShapeEnum = "VNICAAS_TELESIS_10G"
	UpdateVnicShapeDetailsVnicShapeVnicaasAmbassadorFixed0100 UpdateVnicShapeDetailsVnicShapeEnum = "VNICAAS_AMBASSADOR_FIXED0100"
	UpdateVnicShapeDetailsVnicShapeVnicaasPrivatedns          UpdateVnicShapeDetailsVnicShapeEnum = "VNICAAS_PRIVATEDNS"
	UpdateVnicShapeDetailsVnicShapeVnicaasFwaas               UpdateVnicShapeDetailsVnicShapeEnum = "VNICAAS_FWAAS"
	UpdateVnicShapeDetailsVnicShapeDynamicE350g               UpdateVnicShapeDetailsVnicShapeEnum = "DYNAMIC_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0040E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0040_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0100E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0100_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0200E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0200_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0300E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0300_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0400E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0400_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0500E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0500_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0600E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0600_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0700E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0700_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0800E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0800_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0900E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0900_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1000E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1000_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1100E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1100_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1200E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1200_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1300E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1300_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1400E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1400_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1500E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1500_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1600E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1600_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1700E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1700_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1800E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1800_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1900E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1900_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2000E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2000_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2100E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2100_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2200E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2200_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2300E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2300_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2400E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2400_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2500E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2500_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2600E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2600_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2700E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2700_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2800E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2800_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2900E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2900_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3000E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3000_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3100E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3100_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3200E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3200_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3300E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3300_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3400E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3400_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3500E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3500_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3600E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3600_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3700E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3700_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3800E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3800_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3900E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3900_E3_50G"
	UpdateVnicShapeDetailsVnicShapeFixed4000E350g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED4000_E3_50G"
	UpdateVnicShapeDetailsVnicShapeEntirehostE350g            UpdateVnicShapeDetailsVnicShapeEnum = "ENTIREHOST_E3_50G"
	UpdateVnicShapeDetailsVnicShapeDynamicE450g               UpdateVnicShapeDetailsVnicShapeEnum = "DYNAMIC_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0040E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0040_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0100E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0100_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0200E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0200_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0300E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0300_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0400E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0400_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0500E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0500_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0600E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0600_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0700E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0700_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0800E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0800_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0900E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0900_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1000E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1000_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1100E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1100_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1200E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1200_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1300E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1300_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1400E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1400_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1500E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1500_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1600E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1600_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1700E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1700_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1800E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1800_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1900E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1900_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2000E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2000_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2100E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2100_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2200E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2200_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2300E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2300_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2400E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2400_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2500E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2500_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2600E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2600_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2700E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2700_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2800E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2800_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2900E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2900_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3000E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3000_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3100E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3100_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3200E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3200_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3300E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3300_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3400E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3400_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3500E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3500_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3600E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3600_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3700E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3700_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3800E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3800_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3900E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3900_E4_50G"
	UpdateVnicShapeDetailsVnicShapeFixed4000E450g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED4000_E4_50G"
	UpdateVnicShapeDetailsVnicShapeEntirehostE450g            UpdateVnicShapeDetailsVnicShapeEnum = "ENTIREHOST_E4_50G"
	UpdateVnicShapeDetailsVnicShapeMicroVmFixed0050E350g      UpdateVnicShapeDetailsVnicShapeEnum = "MICRO_VM_FIXED0050_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0025E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0025_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0050E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0050_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0075E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0075_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0100E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0100_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0125E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0125_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0150E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0150_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0175E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0175_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0200E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0200_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0225E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0225_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0250E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0250_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0275E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0275_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0300E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0300_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0325E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0325_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0350E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0350_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0375E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0375_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0400E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0400_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0425E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0425_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0450E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0450_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0475E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0475_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0500E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0500_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0525E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0525_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0550E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0550_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0575E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0575_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0600E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0600_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0625E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0625_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0650E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0650_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0675E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0675_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0700E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0700_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0725E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0725_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0750E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0750_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0775E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0775_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0800E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0800_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0825E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0825_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0850E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0850_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0875E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0875_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0900E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0925E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0925_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0950E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0950_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0975E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0975_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1000E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1000_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1025E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1025_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1050E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1050_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1075E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1075_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1100E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1100_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1125E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1125_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1150E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1150_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1175E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1175_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1200E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1200_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1225E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1225_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1250E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1250_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1275E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1275_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1300E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1300_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1325E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1325_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1350E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1350_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1375E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1375_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1400E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1400_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1425E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1425_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1450E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1450_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1475E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1475_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1500E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1500_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1525E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1525_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1550E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1550_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1575E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1575_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1600E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1600_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1625E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1625_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1650E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1650_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1700E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1700_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1725E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1725_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1750E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1750_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1800E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1850E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1850_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1875E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1875_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1900E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1900_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1925E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1925_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1950E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1950_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2000E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2000_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2025E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2025_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2050E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2050_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2100E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2100_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2125E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2125_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2150E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2150_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2175E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2175_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2200E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2200_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2250E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2250_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2275E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2275_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2300E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2300_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2325E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2325_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2350E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2350_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2375E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2375_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2400E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2400_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2450E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2450_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2475E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2475_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2500E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2500_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2550E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2550_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2600E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2600_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2625E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2625_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2650E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2650_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2700E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2750E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2750_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2775E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2775_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2800E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2800_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2850E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2850_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2875E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2875_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2900E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2900_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2925E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2925_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2950E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2950_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2975E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2975_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3000E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3000_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3025E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3025_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3050E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3050_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3075E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3075_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3100E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3100_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3125E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3125_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3150E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3150_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3200E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3200_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3225E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3225_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3250E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3250_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3300E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3300_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3325E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3325_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3375E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3375_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3400E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3400_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3450E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3450_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3500E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3500_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3525E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3525_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3575E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3575_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3600E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3625E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3625_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3675E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3675_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3700E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3700_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3750E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3750_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3800E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3800_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3825E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3825_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3850E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3850_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3875E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3875_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3900E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3900_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3975E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3975_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4000E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4000_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4025E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4025_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4050E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4050_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4100E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4100_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4125E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4125_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4200E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4200_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4225E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4225_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4250E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4250_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4275E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4275_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4300E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4300_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4350E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4350_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4375E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4375_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4400E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4400_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4425E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4425_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4500E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4550E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4550_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4575E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4575_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4600E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4600_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4625E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4625_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4650E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4650_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4675E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4675_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4700E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4700_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4725E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4725_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4750E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4750_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4800E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4800_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4875E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4875_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4900E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4900_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4950E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4950_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed5000E350g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED5000_E3_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0025E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0025_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0050E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0050_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0075E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0075_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0100E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0100_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0125E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0125_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0150E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0150_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0175E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0175_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0200E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0200_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0225E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0225_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0250E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0250_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0275E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0275_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0300E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0300_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0325E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0325_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0350E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0350_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0375E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0375_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0400E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0400_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0425E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0425_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0450E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0450_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0475E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0475_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0500E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0500_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0525E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0525_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0550E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0550_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0575E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0575_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0600E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0600_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0625E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0625_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0650E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0650_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0675E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0675_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0700E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0700_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0725E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0725_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0750E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0750_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0775E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0775_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0800E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0800_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0825E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0825_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0850E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0850_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0875E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0875_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0900E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0925E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0925_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0950E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0950_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0975E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0975_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1000E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1000_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1025E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1025_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1050E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1050_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1075E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1075_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1100E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1100_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1125E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1125_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1150E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1150_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1175E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1175_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1200E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1200_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1225E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1225_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1250E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1250_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1275E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1275_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1300E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1300_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1325E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1325_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1350E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1350_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1375E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1375_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1400E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1400_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1425E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1425_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1450E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1450_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1475E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1475_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1500E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1500_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1525E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1525_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1550E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1550_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1575E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1575_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1600E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1600_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1625E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1625_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1650E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1650_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1700E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1700_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1725E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1725_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1750E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1750_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1800E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1850E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1850_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1875E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1875_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1900E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1900_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1925E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1925_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1950E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1950_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2000E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2000_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2025E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2025_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2050E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2050_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2100E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2100_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2125E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2125_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2150E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2150_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2175E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2175_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2200E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2200_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2250E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2250_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2275E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2275_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2300E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2300_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2325E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2325_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2350E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2350_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2375E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2375_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2400E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2400_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2450E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2450_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2475E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2475_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2500E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2500_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2550E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2550_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2600E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2600_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2625E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2625_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2650E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2650_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2700E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2750E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2750_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2775E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2775_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2800E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2800_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2850E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2850_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2875E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2875_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2900E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2900_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2925E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2925_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2950E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2950_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2975E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2975_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3000E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3000_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3025E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3025_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3050E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3050_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3075E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3075_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3100E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3100_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3125E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3125_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3150E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3150_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3200E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3200_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3225E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3225_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3250E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3250_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3300E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3300_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3325E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3325_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3375E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3375_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3400E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3400_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3450E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3450_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3500E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3500_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3525E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3525_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3575E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3575_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3600E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3625E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3625_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3675E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3675_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3700E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3700_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3750E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3750_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3800E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3800_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3825E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3825_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3850E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3850_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3875E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3875_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3900E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3900_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3975E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3975_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4000E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4000_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4025E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4025_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4050E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4050_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4100E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4100_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4125E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4125_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4200E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4200_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4225E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4225_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4250E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4250_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4275E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4275_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4300E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4300_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4350E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4350_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4375E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4375_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4400E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4400_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4425E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4425_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4500E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4550E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4550_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4575E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4575_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4600E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4600_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4625E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4625_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4650E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4650_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4675E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4675_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4700E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4700_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4725E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4725_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4750E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4750_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4800E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4800_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4875E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4875_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4900E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4900_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4950E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4950_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed5000E450g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED5000_E4_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0020A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0020_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0040A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0040_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0060A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0060_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0080A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0080_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0100A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0100_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0120A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0120_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0140A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0140_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0160A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0160_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0180A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0180_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0200A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0200_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0220A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0220_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0240A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0240_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0260A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0260_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0280A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0280_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0300A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0300_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0320A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0320_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0340A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0340_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0360A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0360_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0380A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0380_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0400A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0400_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0420A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0420_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0440A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0440_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0460A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0460_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0480A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0480_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0500A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0500_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0520A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0520_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0540A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0540_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0560A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0560_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0580A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0580_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0600A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0600_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0620A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0620_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0640A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0640_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0660A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0660_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0680A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0680_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0700A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0700_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0720A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0720_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0740A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0740_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0760A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0760_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0780A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0780_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0800A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0800_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0820A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0820_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0840A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0840_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0860A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0860_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0880A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0880_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0900A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0920A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0920_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0940A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0940_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0960A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0960_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0980A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0980_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1000A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1000_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1020A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1020_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1040A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1040_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1060A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1060_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1080A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1080_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1100A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1100_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1120A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1120_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1140A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1140_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1160A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1160_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1180A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1180_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1200A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1200_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1220A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1220_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1240A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1240_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1260A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1260_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1280A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1280_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1300A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1300_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1320A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1320_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1340A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1340_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1360A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1360_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1380A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1380_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1400A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1400_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1420A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1420_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1440A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1440_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1460A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1460_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1480A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1480_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1500A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1500_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1520A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1520_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1540A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1540_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1560A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1560_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1580A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1580_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1600A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1600_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1620A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1620_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1640A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1640_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1660A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1660_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1680A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1680_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1700A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1700_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1720A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1720_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1740A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1740_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1760A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1760_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1780A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1780_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1800A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1820A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1820_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1840A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1840_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1860A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1860_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1880A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1880_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1900A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1900_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1920A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1920_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1940A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1940_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1960A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1960_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1980A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1980_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2000A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2000_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2020A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2020_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2040A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2040_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2060A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2060_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2080A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2080_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2100A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2100_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2120A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2120_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2140A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2140_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2160A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2160_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2180A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2180_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2200A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2200_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2220A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2220_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2240A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2240_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2260A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2260_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2280A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2280_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2300A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2300_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2320A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2320_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2340A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2340_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2360A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2360_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2380A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2380_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2400A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2400_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2420A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2420_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2440A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2440_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2460A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2460_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2480A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2480_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2500A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2500_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2520A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2520_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2540A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2540_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2560A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2560_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2580A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2580_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2600A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2600_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2620A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2620_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2640A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2640_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2660A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2660_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2680A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2680_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2700A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2720A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2720_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2740A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2740_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2760A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2760_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2780A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2780_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2800A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2800_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2820A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2820_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2840A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2840_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2860A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2860_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2880A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2880_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2900A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2900_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2920A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2920_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2940A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2940_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2960A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2960_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2980A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2980_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3000A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3000_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3020A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3020_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3040A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3040_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3060A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3060_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3080A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3080_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3100A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3100_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3120A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3120_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3140A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3140_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3160A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3160_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3180A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3180_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3200A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3200_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3220A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3220_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3240A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3240_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3260A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3260_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3280A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3280_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3300A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3300_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3320A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3320_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3340A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3340_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3360A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3360_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3380A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3380_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3400A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3400_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3420A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3420_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3440A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3440_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3460A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3460_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3480A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3480_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3500A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3500_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3520A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3520_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3540A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3540_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3560A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3560_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3580A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3580_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3600A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3620A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3620_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3640A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3640_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3660A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3660_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3680A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3680_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3700A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3700_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3720A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3720_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3740A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3740_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3760A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3760_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3780A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3780_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3800A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3800_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3820A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3820_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3840A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3840_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3860A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3860_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3880A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3880_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3900A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3900_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3920A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3920_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3940A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3940_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3960A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3960_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3980A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3980_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4000A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4000_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4020A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4020_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4040A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4040_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4060A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4060_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4080A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4080_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4100A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4100_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4120A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4120_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4140A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4140_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4160A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4160_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4180A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4180_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4200A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4200_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4220A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4220_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4240A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4240_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4260A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4260_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4280A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4280_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4300A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4300_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4320A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4320_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4340A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4340_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4360A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4360_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4380A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4380_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4400A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4400_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4420A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4420_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4440A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4440_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4460A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4460_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4480A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4480_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4500A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4520A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4520_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4540A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4540_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4560A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4560_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4580A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4580_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4600A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4600_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4620A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4620_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4640A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4640_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4660A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4660_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4680A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4680_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4700A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4700_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4720A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4720_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4740A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4740_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4760A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4760_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4780A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4780_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4800A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4800_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4820A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4820_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4840A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4840_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4860A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4860_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4880A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4880_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4900A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4900_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4920A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4920_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4940A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4940_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4960A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4960_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4980A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4980_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed5000A150g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED5000_A1_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0090X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0090_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0180X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0180_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0270X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0270_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0360X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0360_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0450X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0450_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0540X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0540_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0630X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0630_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0720X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0720_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0810X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0810_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0900X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0990X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0990_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1080X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1080_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1170X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1170_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1260X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1260_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1350X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1350_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1440X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1440_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1530X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1530_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1620X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1620_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1710X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1710_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1800X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1890X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1890_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1980X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1980_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2070X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2070_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2160X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2160_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2250X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2250_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2340X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2340_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2430X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2430_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2520X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2520_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2610X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2610_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2700X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2790X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2790_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2880X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2880_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2970X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2970_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3060X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3060_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3150X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3150_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3240X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3240_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3330X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3330_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3420X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3420_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3510X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3510_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3600X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3690X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3690_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3780X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3780_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3870X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3870_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3960X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3960_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4050X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4050_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4140X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4140_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4230X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4230_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4320X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4320_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4410X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4410_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4500X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4590X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4590_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4680X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4680_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4770X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4770_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4860X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4860_X9_50G"
	UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4950X950g    UpdateVnicShapeDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4950_X9_50G"
	UpdateVnicShapeDetailsVnicShapeDynamicA150g               UpdateVnicShapeDetailsVnicShapeEnum = "DYNAMIC_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0040A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0040_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0100A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0100_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0200A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0200_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0300A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0300_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0400A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0400_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0500A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0500_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0600A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0600_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0700A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0700_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0800A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0800_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0900A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0900_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1000A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1000_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1100A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1100_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1200A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1200_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1300A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1300_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1400A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1400_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1500A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1500_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1600A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1600_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1700A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1700_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1800A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1800_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1900A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1900_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2000A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2000_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2100A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2100_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2200A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2200_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2300A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2300_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2400A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2400_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2500A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2500_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2600A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2600_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2700A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2700_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2800A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2800_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2900A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2900_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3000A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3000_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3100A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3100_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3200A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3200_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3300A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3300_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3400A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3400_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3500A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3500_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3600A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3600_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3700A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3700_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3800A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3800_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3900A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3900_A1_50G"
	UpdateVnicShapeDetailsVnicShapeFixed4000A150g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED4000_A1_50G"
	UpdateVnicShapeDetailsVnicShapeEntirehostA150g            UpdateVnicShapeDetailsVnicShapeEnum = "ENTIREHOST_A1_50G"
	UpdateVnicShapeDetailsVnicShapeDynamicX950g               UpdateVnicShapeDetailsVnicShapeEnum = "DYNAMIC_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0040X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0040_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0400X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0400_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed0800X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED0800_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1200X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1200_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed1600X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED1600_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2000X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2000_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2400X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2400_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed2800X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED2800_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3200X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3200_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed3600X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED3600_X9_50G"
	UpdateVnicShapeDetailsVnicShapeFixed4000X950g             UpdateVnicShapeDetailsVnicShapeEnum = "FIXED4000_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed0100X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED0100_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed0200X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED0200_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed0300X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED0300_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed0400X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED0400_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed0500X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED0500_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed0600X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED0600_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed0700X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED0700_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed0800X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED0800_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed0900X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED0900_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed1000X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED1000_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed1100X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED1100_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed1200X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED1200_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed1300X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED1300_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed1400X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED1400_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed1500X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED1500_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed1600X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED1600_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed1700X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED1700_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed1800X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED1800_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed1900X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED1900_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed2000X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED2000_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed2100X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED2100_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed2200X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED2200_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed2300X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED2300_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed2400X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED2400_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed2500X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED2500_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed2600X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED2600_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed2700X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED2700_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed2800X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED2800_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed2900X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED2900_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed3000X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED3000_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed3100X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED3100_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed3200X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED3200_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed3300X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED3300_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed3400X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED3400_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed3500X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED3500_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed3600X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED3600_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed3700X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED3700_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed3800X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED3800_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed3900X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED3900_X9_50G"
	UpdateVnicShapeDetailsVnicShapeStandardVmFixed4000X950g   UpdateVnicShapeDetailsVnicShapeEnum = "STANDARD_VM_FIXED4000_X9_50G"
	UpdateVnicShapeDetailsVnicShapeEntirehostX950g            UpdateVnicShapeDetailsVnicShapeEnum = "ENTIREHOST_X9_50G"
)

var mappingUpdateVnicShapeDetailsVnicShapeEnum = map[string]UpdateVnicShapeDetailsVnicShapeEnum{
	"DYNAMIC":                      UpdateVnicShapeDetailsVnicShapeDynamic,
	"FIXED0040":                    UpdateVnicShapeDetailsVnicShapeFixed0040,
	"FIXED0060":                    UpdateVnicShapeDetailsVnicShapeFixed0060,
	"FIXED0060_PSM":                UpdateVnicShapeDetailsVnicShapeFixed0060Psm,
	"FIXED0100":                    UpdateVnicShapeDetailsVnicShapeFixed0100,
	"FIXED0120":                    UpdateVnicShapeDetailsVnicShapeFixed0120,
	"FIXED0120_2X":                 UpdateVnicShapeDetailsVnicShapeFixed01202x,
	"FIXED0200":                    UpdateVnicShapeDetailsVnicShapeFixed0200,
	"FIXED0240":                    UpdateVnicShapeDetailsVnicShapeFixed0240,
	"FIXED0480":                    UpdateVnicShapeDetailsVnicShapeFixed0480,
	"ENTIREHOST":                   UpdateVnicShapeDetailsVnicShapeEntirehost,
	"DYNAMIC_25G":                  UpdateVnicShapeDetailsVnicShapeDynamic25g,
	"FIXED0040_25G":                UpdateVnicShapeDetailsVnicShapeFixed004025g,
	"FIXED0100_25G":                UpdateVnicShapeDetailsVnicShapeFixed010025g,
	"FIXED0200_25G":                UpdateVnicShapeDetailsVnicShapeFixed020025g,
	"FIXED0400_25G":                UpdateVnicShapeDetailsVnicShapeFixed040025g,
	"FIXED0800_25G":                UpdateVnicShapeDetailsVnicShapeFixed080025g,
	"FIXED1600_25G":                UpdateVnicShapeDetailsVnicShapeFixed160025g,
	"FIXED2400_25G":                UpdateVnicShapeDetailsVnicShapeFixed240025g,
	"ENTIREHOST_25G":               UpdateVnicShapeDetailsVnicShapeEntirehost25g,
	"DYNAMIC_E1_25G":               UpdateVnicShapeDetailsVnicShapeDynamicE125g,
	"FIXED0040_E1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0040E125g,
	"FIXED0070_E1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0070E125g,
	"FIXED0140_E1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0140E125g,
	"FIXED0280_E1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0280E125g,
	"FIXED0560_E1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0560E125g,
	"FIXED1120_E1_25G":             UpdateVnicShapeDetailsVnicShapeFixed1120E125g,
	"FIXED1680_E1_25G":             UpdateVnicShapeDetailsVnicShapeFixed1680E125g,
	"ENTIREHOST_E1_25G":            UpdateVnicShapeDetailsVnicShapeEntirehostE125g,
	"DYNAMIC_B1_25G":               UpdateVnicShapeDetailsVnicShapeDynamicB125g,
	"FIXED0040_B1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0040B125g,
	"FIXED0060_B1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0060B125g,
	"FIXED0120_B1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0120B125g,
	"FIXED0240_B1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0240B125g,
	"FIXED0480_B1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0480B125g,
	"FIXED0960_B1_25G":             UpdateVnicShapeDetailsVnicShapeFixed0960B125g,
	"ENTIREHOST_B1_25G":            UpdateVnicShapeDetailsVnicShapeEntirehostB125g,
	"MICRO_VM_FIXED0048_E1_25G":    UpdateVnicShapeDetailsVnicShapeMicroVmFixed0048E125g,
	"MICRO_LB_FIXED0001_E1_25G":    UpdateVnicShapeDetailsVnicShapeMicroLbFixed0001E125g,
	"VNICAAS_FIXED0200":            UpdateVnicShapeDetailsVnicShapeVnicaasFixed0200,
	"VNICAAS_FIXED0400":            UpdateVnicShapeDetailsVnicShapeVnicaasFixed0400,
	"VNICAAS_FIXED0700":            UpdateVnicShapeDetailsVnicShapeVnicaasFixed0700,
	"VNICAAS_NLB_APPROVED_10G":     UpdateVnicShapeDetailsVnicShapeVnicaasNlbApproved10g,
	"VNICAAS_NLB_APPROVED_25G":     UpdateVnicShapeDetailsVnicShapeVnicaasNlbApproved25g,
	"VNICAAS_TELESIS_25G":          UpdateVnicShapeDetailsVnicShapeVnicaasTelesis25g,
	"VNICAAS_TELESIS_10G":          UpdateVnicShapeDetailsVnicShapeVnicaasTelesis10g,
	"VNICAAS_AMBASSADOR_FIXED0100": UpdateVnicShapeDetailsVnicShapeVnicaasAmbassadorFixed0100,
	"VNICAAS_PRIVATEDNS":           UpdateVnicShapeDetailsVnicShapeVnicaasPrivatedns,
	"VNICAAS_FWAAS":                UpdateVnicShapeDetailsVnicShapeVnicaasFwaas,
	"DYNAMIC_E3_50G":               UpdateVnicShapeDetailsVnicShapeDynamicE350g,
	"FIXED0040_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed0040E350g,
	"FIXED0100_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed0100E350g,
	"FIXED0200_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed0200E350g,
	"FIXED0300_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed0300E350g,
	"FIXED0400_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed0400E350g,
	"FIXED0500_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed0500E350g,
	"FIXED0600_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed0600E350g,
	"FIXED0700_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed0700E350g,
	"FIXED0800_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed0800E350g,
	"FIXED0900_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed0900E350g,
	"FIXED1000_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed1000E350g,
	"FIXED1100_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed1100E350g,
	"FIXED1200_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed1200E350g,
	"FIXED1300_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed1300E350g,
	"FIXED1400_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed1400E350g,
	"FIXED1500_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed1500E350g,
	"FIXED1600_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed1600E350g,
	"FIXED1700_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed1700E350g,
	"FIXED1800_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed1800E350g,
	"FIXED1900_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed1900E350g,
	"FIXED2000_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed2000E350g,
	"FIXED2100_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed2100E350g,
	"FIXED2200_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed2200E350g,
	"FIXED2300_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed2300E350g,
	"FIXED2400_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed2400E350g,
	"FIXED2500_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed2500E350g,
	"FIXED2600_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed2600E350g,
	"FIXED2700_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed2700E350g,
	"FIXED2800_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed2800E350g,
	"FIXED2900_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed2900E350g,
	"FIXED3000_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed3000E350g,
	"FIXED3100_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed3100E350g,
	"FIXED3200_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed3200E350g,
	"FIXED3300_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed3300E350g,
	"FIXED3400_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed3400E350g,
	"FIXED3500_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed3500E350g,
	"FIXED3600_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed3600E350g,
	"FIXED3700_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed3700E350g,
	"FIXED3800_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed3800E350g,
	"FIXED3900_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed3900E350g,
	"FIXED4000_E3_50G":             UpdateVnicShapeDetailsVnicShapeFixed4000E350g,
	"ENTIREHOST_E3_50G":            UpdateVnicShapeDetailsVnicShapeEntirehostE350g,
	"DYNAMIC_E4_50G":               UpdateVnicShapeDetailsVnicShapeDynamicE450g,
	"FIXED0040_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed0040E450g,
	"FIXED0100_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed0100E450g,
	"FIXED0200_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed0200E450g,
	"FIXED0300_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed0300E450g,
	"FIXED0400_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed0400E450g,
	"FIXED0500_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed0500E450g,
	"FIXED0600_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed0600E450g,
	"FIXED0700_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed0700E450g,
	"FIXED0800_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed0800E450g,
	"FIXED0900_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed0900E450g,
	"FIXED1000_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed1000E450g,
	"FIXED1100_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed1100E450g,
	"FIXED1200_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed1200E450g,
	"FIXED1300_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed1300E450g,
	"FIXED1400_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed1400E450g,
	"FIXED1500_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed1500E450g,
	"FIXED1600_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed1600E450g,
	"FIXED1700_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed1700E450g,
	"FIXED1800_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed1800E450g,
	"FIXED1900_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed1900E450g,
	"FIXED2000_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed2000E450g,
	"FIXED2100_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed2100E450g,
	"FIXED2200_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed2200E450g,
	"FIXED2300_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed2300E450g,
	"FIXED2400_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed2400E450g,
	"FIXED2500_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed2500E450g,
	"FIXED2600_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed2600E450g,
	"FIXED2700_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed2700E450g,
	"FIXED2800_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed2800E450g,
	"FIXED2900_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed2900E450g,
	"FIXED3000_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed3000E450g,
	"FIXED3100_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed3100E450g,
	"FIXED3200_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed3200E450g,
	"FIXED3300_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed3300E450g,
	"FIXED3400_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed3400E450g,
	"FIXED3500_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed3500E450g,
	"FIXED3600_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed3600E450g,
	"FIXED3700_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed3700E450g,
	"FIXED3800_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed3800E450g,
	"FIXED3900_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed3900E450g,
	"FIXED4000_E4_50G":             UpdateVnicShapeDetailsVnicShapeFixed4000E450g,
	"ENTIREHOST_E4_50G":            UpdateVnicShapeDetailsVnicShapeEntirehostE450g,
	"MICRO_VM_FIXED0050_E3_50G":    UpdateVnicShapeDetailsVnicShapeMicroVmFixed0050E350g,
	"SUBCORE_VM_FIXED0025_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0025E350g,
	"SUBCORE_VM_FIXED0050_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0050E350g,
	"SUBCORE_VM_FIXED0075_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0075E350g,
	"SUBCORE_VM_FIXED0100_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0100E350g,
	"SUBCORE_VM_FIXED0125_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0125E350g,
	"SUBCORE_VM_FIXED0150_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0150E350g,
	"SUBCORE_VM_FIXED0175_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0175E350g,
	"SUBCORE_VM_FIXED0200_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0200E350g,
	"SUBCORE_VM_FIXED0225_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0225E350g,
	"SUBCORE_VM_FIXED0250_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0250E350g,
	"SUBCORE_VM_FIXED0275_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0275E350g,
	"SUBCORE_VM_FIXED0300_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0300E350g,
	"SUBCORE_VM_FIXED0325_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0325E350g,
	"SUBCORE_VM_FIXED0350_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0350E350g,
	"SUBCORE_VM_FIXED0375_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0375E350g,
	"SUBCORE_VM_FIXED0400_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0400E350g,
	"SUBCORE_VM_FIXED0425_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0425E350g,
	"SUBCORE_VM_FIXED0450_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0450E350g,
	"SUBCORE_VM_FIXED0475_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0475E350g,
	"SUBCORE_VM_FIXED0500_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0500E350g,
	"SUBCORE_VM_FIXED0525_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0525E350g,
	"SUBCORE_VM_FIXED0550_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0550E350g,
	"SUBCORE_VM_FIXED0575_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0575E350g,
	"SUBCORE_VM_FIXED0600_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0600E350g,
	"SUBCORE_VM_FIXED0625_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0625E350g,
	"SUBCORE_VM_FIXED0650_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0650E350g,
	"SUBCORE_VM_FIXED0675_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0675E350g,
	"SUBCORE_VM_FIXED0700_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0700E350g,
	"SUBCORE_VM_FIXED0725_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0725E350g,
	"SUBCORE_VM_FIXED0750_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0750E350g,
	"SUBCORE_VM_FIXED0775_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0775E350g,
	"SUBCORE_VM_FIXED0800_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0800E350g,
	"SUBCORE_VM_FIXED0825_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0825E350g,
	"SUBCORE_VM_FIXED0850_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0850E350g,
	"SUBCORE_VM_FIXED0875_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0875E350g,
	"SUBCORE_VM_FIXED0900_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0900E350g,
	"SUBCORE_VM_FIXED0925_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0925E350g,
	"SUBCORE_VM_FIXED0950_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0950E350g,
	"SUBCORE_VM_FIXED0975_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0975E350g,
	"SUBCORE_VM_FIXED1000_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1000E350g,
	"SUBCORE_VM_FIXED1025_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1025E350g,
	"SUBCORE_VM_FIXED1050_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1050E350g,
	"SUBCORE_VM_FIXED1075_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1075E350g,
	"SUBCORE_VM_FIXED1100_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1100E350g,
	"SUBCORE_VM_FIXED1125_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1125E350g,
	"SUBCORE_VM_FIXED1150_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1150E350g,
	"SUBCORE_VM_FIXED1175_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1175E350g,
	"SUBCORE_VM_FIXED1200_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1200E350g,
	"SUBCORE_VM_FIXED1225_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1225E350g,
	"SUBCORE_VM_FIXED1250_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1250E350g,
	"SUBCORE_VM_FIXED1275_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1275E350g,
	"SUBCORE_VM_FIXED1300_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1300E350g,
	"SUBCORE_VM_FIXED1325_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1325E350g,
	"SUBCORE_VM_FIXED1350_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1350E350g,
	"SUBCORE_VM_FIXED1375_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1375E350g,
	"SUBCORE_VM_FIXED1400_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1400E350g,
	"SUBCORE_VM_FIXED1425_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1425E350g,
	"SUBCORE_VM_FIXED1450_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1450E350g,
	"SUBCORE_VM_FIXED1475_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1475E350g,
	"SUBCORE_VM_FIXED1500_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1500E350g,
	"SUBCORE_VM_FIXED1525_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1525E350g,
	"SUBCORE_VM_FIXED1550_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1550E350g,
	"SUBCORE_VM_FIXED1575_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1575E350g,
	"SUBCORE_VM_FIXED1600_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1600E350g,
	"SUBCORE_VM_FIXED1625_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1625E350g,
	"SUBCORE_VM_FIXED1650_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1650E350g,
	"SUBCORE_VM_FIXED1700_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1700E350g,
	"SUBCORE_VM_FIXED1725_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1725E350g,
	"SUBCORE_VM_FIXED1750_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1750E350g,
	"SUBCORE_VM_FIXED1800_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1800E350g,
	"SUBCORE_VM_FIXED1850_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1850E350g,
	"SUBCORE_VM_FIXED1875_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1875E350g,
	"SUBCORE_VM_FIXED1900_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1900E350g,
	"SUBCORE_VM_FIXED1925_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1925E350g,
	"SUBCORE_VM_FIXED1950_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1950E350g,
	"SUBCORE_VM_FIXED2000_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2000E350g,
	"SUBCORE_VM_FIXED2025_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2025E350g,
	"SUBCORE_VM_FIXED2050_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2050E350g,
	"SUBCORE_VM_FIXED2100_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2100E350g,
	"SUBCORE_VM_FIXED2125_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2125E350g,
	"SUBCORE_VM_FIXED2150_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2150E350g,
	"SUBCORE_VM_FIXED2175_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2175E350g,
	"SUBCORE_VM_FIXED2200_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2200E350g,
	"SUBCORE_VM_FIXED2250_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2250E350g,
	"SUBCORE_VM_FIXED2275_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2275E350g,
	"SUBCORE_VM_FIXED2300_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2300E350g,
	"SUBCORE_VM_FIXED2325_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2325E350g,
	"SUBCORE_VM_FIXED2350_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2350E350g,
	"SUBCORE_VM_FIXED2375_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2375E350g,
	"SUBCORE_VM_FIXED2400_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2400E350g,
	"SUBCORE_VM_FIXED2450_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2450E350g,
	"SUBCORE_VM_FIXED2475_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2475E350g,
	"SUBCORE_VM_FIXED2500_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2500E350g,
	"SUBCORE_VM_FIXED2550_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2550E350g,
	"SUBCORE_VM_FIXED2600_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2600E350g,
	"SUBCORE_VM_FIXED2625_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2625E350g,
	"SUBCORE_VM_FIXED2650_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2650E350g,
	"SUBCORE_VM_FIXED2700_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2700E350g,
	"SUBCORE_VM_FIXED2750_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2750E350g,
	"SUBCORE_VM_FIXED2775_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2775E350g,
	"SUBCORE_VM_FIXED2800_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2800E350g,
	"SUBCORE_VM_FIXED2850_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2850E350g,
	"SUBCORE_VM_FIXED2875_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2875E350g,
	"SUBCORE_VM_FIXED2900_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2900E350g,
	"SUBCORE_VM_FIXED2925_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2925E350g,
	"SUBCORE_VM_FIXED2950_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2950E350g,
	"SUBCORE_VM_FIXED2975_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2975E350g,
	"SUBCORE_VM_FIXED3000_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3000E350g,
	"SUBCORE_VM_FIXED3025_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3025E350g,
	"SUBCORE_VM_FIXED3050_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3050E350g,
	"SUBCORE_VM_FIXED3075_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3075E350g,
	"SUBCORE_VM_FIXED3100_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3100E350g,
	"SUBCORE_VM_FIXED3125_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3125E350g,
	"SUBCORE_VM_FIXED3150_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3150E350g,
	"SUBCORE_VM_FIXED3200_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3200E350g,
	"SUBCORE_VM_FIXED3225_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3225E350g,
	"SUBCORE_VM_FIXED3250_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3250E350g,
	"SUBCORE_VM_FIXED3300_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3300E350g,
	"SUBCORE_VM_FIXED3325_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3325E350g,
	"SUBCORE_VM_FIXED3375_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3375E350g,
	"SUBCORE_VM_FIXED3400_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3400E350g,
	"SUBCORE_VM_FIXED3450_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3450E350g,
	"SUBCORE_VM_FIXED3500_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3500E350g,
	"SUBCORE_VM_FIXED3525_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3525E350g,
	"SUBCORE_VM_FIXED3575_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3575E350g,
	"SUBCORE_VM_FIXED3600_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3600E350g,
	"SUBCORE_VM_FIXED3625_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3625E350g,
	"SUBCORE_VM_FIXED3675_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3675E350g,
	"SUBCORE_VM_FIXED3700_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3700E350g,
	"SUBCORE_VM_FIXED3750_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3750E350g,
	"SUBCORE_VM_FIXED3800_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3800E350g,
	"SUBCORE_VM_FIXED3825_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3825E350g,
	"SUBCORE_VM_FIXED3850_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3850E350g,
	"SUBCORE_VM_FIXED3875_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3875E350g,
	"SUBCORE_VM_FIXED3900_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3900E350g,
	"SUBCORE_VM_FIXED3975_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3975E350g,
	"SUBCORE_VM_FIXED4000_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4000E350g,
	"SUBCORE_VM_FIXED4025_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4025E350g,
	"SUBCORE_VM_FIXED4050_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4050E350g,
	"SUBCORE_VM_FIXED4100_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4100E350g,
	"SUBCORE_VM_FIXED4125_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4125E350g,
	"SUBCORE_VM_FIXED4200_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4200E350g,
	"SUBCORE_VM_FIXED4225_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4225E350g,
	"SUBCORE_VM_FIXED4250_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4250E350g,
	"SUBCORE_VM_FIXED4275_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4275E350g,
	"SUBCORE_VM_FIXED4300_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4300E350g,
	"SUBCORE_VM_FIXED4350_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4350E350g,
	"SUBCORE_VM_FIXED4375_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4375E350g,
	"SUBCORE_VM_FIXED4400_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4400E350g,
	"SUBCORE_VM_FIXED4425_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4425E350g,
	"SUBCORE_VM_FIXED4500_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4500E350g,
	"SUBCORE_VM_FIXED4550_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4550E350g,
	"SUBCORE_VM_FIXED4575_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4575E350g,
	"SUBCORE_VM_FIXED4600_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4600E350g,
	"SUBCORE_VM_FIXED4625_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4625E350g,
	"SUBCORE_VM_FIXED4650_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4650E350g,
	"SUBCORE_VM_FIXED4675_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4675E350g,
	"SUBCORE_VM_FIXED4700_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4700E350g,
	"SUBCORE_VM_FIXED4725_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4725E350g,
	"SUBCORE_VM_FIXED4750_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4750E350g,
	"SUBCORE_VM_FIXED4800_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4800E350g,
	"SUBCORE_VM_FIXED4875_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4875E350g,
	"SUBCORE_VM_FIXED4900_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4900E350g,
	"SUBCORE_VM_FIXED4950_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4950E350g,
	"SUBCORE_VM_FIXED5000_E3_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed5000E350g,
	"SUBCORE_VM_FIXED0025_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0025E450g,
	"SUBCORE_VM_FIXED0050_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0050E450g,
	"SUBCORE_VM_FIXED0075_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0075E450g,
	"SUBCORE_VM_FIXED0100_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0100E450g,
	"SUBCORE_VM_FIXED0125_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0125E450g,
	"SUBCORE_VM_FIXED0150_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0150E450g,
	"SUBCORE_VM_FIXED0175_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0175E450g,
	"SUBCORE_VM_FIXED0200_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0200E450g,
	"SUBCORE_VM_FIXED0225_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0225E450g,
	"SUBCORE_VM_FIXED0250_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0250E450g,
	"SUBCORE_VM_FIXED0275_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0275E450g,
	"SUBCORE_VM_FIXED0300_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0300E450g,
	"SUBCORE_VM_FIXED0325_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0325E450g,
	"SUBCORE_VM_FIXED0350_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0350E450g,
	"SUBCORE_VM_FIXED0375_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0375E450g,
	"SUBCORE_VM_FIXED0400_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0400E450g,
	"SUBCORE_VM_FIXED0425_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0425E450g,
	"SUBCORE_VM_FIXED0450_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0450E450g,
	"SUBCORE_VM_FIXED0475_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0475E450g,
	"SUBCORE_VM_FIXED0500_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0500E450g,
	"SUBCORE_VM_FIXED0525_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0525E450g,
	"SUBCORE_VM_FIXED0550_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0550E450g,
	"SUBCORE_VM_FIXED0575_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0575E450g,
	"SUBCORE_VM_FIXED0600_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0600E450g,
	"SUBCORE_VM_FIXED0625_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0625E450g,
	"SUBCORE_VM_FIXED0650_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0650E450g,
	"SUBCORE_VM_FIXED0675_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0675E450g,
	"SUBCORE_VM_FIXED0700_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0700E450g,
	"SUBCORE_VM_FIXED0725_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0725E450g,
	"SUBCORE_VM_FIXED0750_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0750E450g,
	"SUBCORE_VM_FIXED0775_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0775E450g,
	"SUBCORE_VM_FIXED0800_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0800E450g,
	"SUBCORE_VM_FIXED0825_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0825E450g,
	"SUBCORE_VM_FIXED0850_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0850E450g,
	"SUBCORE_VM_FIXED0875_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0875E450g,
	"SUBCORE_VM_FIXED0900_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0900E450g,
	"SUBCORE_VM_FIXED0925_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0925E450g,
	"SUBCORE_VM_FIXED0950_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0950E450g,
	"SUBCORE_VM_FIXED0975_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0975E450g,
	"SUBCORE_VM_FIXED1000_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1000E450g,
	"SUBCORE_VM_FIXED1025_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1025E450g,
	"SUBCORE_VM_FIXED1050_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1050E450g,
	"SUBCORE_VM_FIXED1075_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1075E450g,
	"SUBCORE_VM_FIXED1100_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1100E450g,
	"SUBCORE_VM_FIXED1125_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1125E450g,
	"SUBCORE_VM_FIXED1150_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1150E450g,
	"SUBCORE_VM_FIXED1175_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1175E450g,
	"SUBCORE_VM_FIXED1200_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1200E450g,
	"SUBCORE_VM_FIXED1225_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1225E450g,
	"SUBCORE_VM_FIXED1250_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1250E450g,
	"SUBCORE_VM_FIXED1275_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1275E450g,
	"SUBCORE_VM_FIXED1300_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1300E450g,
	"SUBCORE_VM_FIXED1325_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1325E450g,
	"SUBCORE_VM_FIXED1350_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1350E450g,
	"SUBCORE_VM_FIXED1375_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1375E450g,
	"SUBCORE_VM_FIXED1400_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1400E450g,
	"SUBCORE_VM_FIXED1425_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1425E450g,
	"SUBCORE_VM_FIXED1450_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1450E450g,
	"SUBCORE_VM_FIXED1475_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1475E450g,
	"SUBCORE_VM_FIXED1500_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1500E450g,
	"SUBCORE_VM_FIXED1525_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1525E450g,
	"SUBCORE_VM_FIXED1550_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1550E450g,
	"SUBCORE_VM_FIXED1575_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1575E450g,
	"SUBCORE_VM_FIXED1600_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1600E450g,
	"SUBCORE_VM_FIXED1625_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1625E450g,
	"SUBCORE_VM_FIXED1650_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1650E450g,
	"SUBCORE_VM_FIXED1700_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1700E450g,
	"SUBCORE_VM_FIXED1725_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1725E450g,
	"SUBCORE_VM_FIXED1750_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1750E450g,
	"SUBCORE_VM_FIXED1800_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1800E450g,
	"SUBCORE_VM_FIXED1850_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1850E450g,
	"SUBCORE_VM_FIXED1875_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1875E450g,
	"SUBCORE_VM_FIXED1900_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1900E450g,
	"SUBCORE_VM_FIXED1925_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1925E450g,
	"SUBCORE_VM_FIXED1950_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1950E450g,
	"SUBCORE_VM_FIXED2000_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2000E450g,
	"SUBCORE_VM_FIXED2025_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2025E450g,
	"SUBCORE_VM_FIXED2050_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2050E450g,
	"SUBCORE_VM_FIXED2100_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2100E450g,
	"SUBCORE_VM_FIXED2125_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2125E450g,
	"SUBCORE_VM_FIXED2150_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2150E450g,
	"SUBCORE_VM_FIXED2175_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2175E450g,
	"SUBCORE_VM_FIXED2200_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2200E450g,
	"SUBCORE_VM_FIXED2250_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2250E450g,
	"SUBCORE_VM_FIXED2275_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2275E450g,
	"SUBCORE_VM_FIXED2300_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2300E450g,
	"SUBCORE_VM_FIXED2325_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2325E450g,
	"SUBCORE_VM_FIXED2350_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2350E450g,
	"SUBCORE_VM_FIXED2375_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2375E450g,
	"SUBCORE_VM_FIXED2400_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2400E450g,
	"SUBCORE_VM_FIXED2450_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2450E450g,
	"SUBCORE_VM_FIXED2475_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2475E450g,
	"SUBCORE_VM_FIXED2500_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2500E450g,
	"SUBCORE_VM_FIXED2550_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2550E450g,
	"SUBCORE_VM_FIXED2600_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2600E450g,
	"SUBCORE_VM_FIXED2625_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2625E450g,
	"SUBCORE_VM_FIXED2650_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2650E450g,
	"SUBCORE_VM_FIXED2700_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2700E450g,
	"SUBCORE_VM_FIXED2750_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2750E450g,
	"SUBCORE_VM_FIXED2775_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2775E450g,
	"SUBCORE_VM_FIXED2800_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2800E450g,
	"SUBCORE_VM_FIXED2850_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2850E450g,
	"SUBCORE_VM_FIXED2875_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2875E450g,
	"SUBCORE_VM_FIXED2900_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2900E450g,
	"SUBCORE_VM_FIXED2925_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2925E450g,
	"SUBCORE_VM_FIXED2950_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2950E450g,
	"SUBCORE_VM_FIXED2975_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2975E450g,
	"SUBCORE_VM_FIXED3000_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3000E450g,
	"SUBCORE_VM_FIXED3025_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3025E450g,
	"SUBCORE_VM_FIXED3050_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3050E450g,
	"SUBCORE_VM_FIXED3075_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3075E450g,
	"SUBCORE_VM_FIXED3100_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3100E450g,
	"SUBCORE_VM_FIXED3125_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3125E450g,
	"SUBCORE_VM_FIXED3150_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3150E450g,
	"SUBCORE_VM_FIXED3200_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3200E450g,
	"SUBCORE_VM_FIXED3225_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3225E450g,
	"SUBCORE_VM_FIXED3250_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3250E450g,
	"SUBCORE_VM_FIXED3300_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3300E450g,
	"SUBCORE_VM_FIXED3325_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3325E450g,
	"SUBCORE_VM_FIXED3375_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3375E450g,
	"SUBCORE_VM_FIXED3400_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3400E450g,
	"SUBCORE_VM_FIXED3450_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3450E450g,
	"SUBCORE_VM_FIXED3500_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3500E450g,
	"SUBCORE_VM_FIXED3525_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3525E450g,
	"SUBCORE_VM_FIXED3575_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3575E450g,
	"SUBCORE_VM_FIXED3600_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3600E450g,
	"SUBCORE_VM_FIXED3625_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3625E450g,
	"SUBCORE_VM_FIXED3675_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3675E450g,
	"SUBCORE_VM_FIXED3700_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3700E450g,
	"SUBCORE_VM_FIXED3750_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3750E450g,
	"SUBCORE_VM_FIXED3800_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3800E450g,
	"SUBCORE_VM_FIXED3825_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3825E450g,
	"SUBCORE_VM_FIXED3850_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3850E450g,
	"SUBCORE_VM_FIXED3875_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3875E450g,
	"SUBCORE_VM_FIXED3900_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3900E450g,
	"SUBCORE_VM_FIXED3975_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3975E450g,
	"SUBCORE_VM_FIXED4000_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4000E450g,
	"SUBCORE_VM_FIXED4025_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4025E450g,
	"SUBCORE_VM_FIXED4050_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4050E450g,
	"SUBCORE_VM_FIXED4100_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4100E450g,
	"SUBCORE_VM_FIXED4125_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4125E450g,
	"SUBCORE_VM_FIXED4200_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4200E450g,
	"SUBCORE_VM_FIXED4225_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4225E450g,
	"SUBCORE_VM_FIXED4250_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4250E450g,
	"SUBCORE_VM_FIXED4275_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4275E450g,
	"SUBCORE_VM_FIXED4300_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4300E450g,
	"SUBCORE_VM_FIXED4350_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4350E450g,
	"SUBCORE_VM_FIXED4375_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4375E450g,
	"SUBCORE_VM_FIXED4400_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4400E450g,
	"SUBCORE_VM_FIXED4425_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4425E450g,
	"SUBCORE_VM_FIXED4500_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4500E450g,
	"SUBCORE_VM_FIXED4550_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4550E450g,
	"SUBCORE_VM_FIXED4575_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4575E450g,
	"SUBCORE_VM_FIXED4600_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4600E450g,
	"SUBCORE_VM_FIXED4625_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4625E450g,
	"SUBCORE_VM_FIXED4650_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4650E450g,
	"SUBCORE_VM_FIXED4675_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4675E450g,
	"SUBCORE_VM_FIXED4700_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4700E450g,
	"SUBCORE_VM_FIXED4725_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4725E450g,
	"SUBCORE_VM_FIXED4750_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4750E450g,
	"SUBCORE_VM_FIXED4800_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4800E450g,
	"SUBCORE_VM_FIXED4875_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4875E450g,
	"SUBCORE_VM_FIXED4900_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4900E450g,
	"SUBCORE_VM_FIXED4950_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4950E450g,
	"SUBCORE_VM_FIXED5000_E4_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed5000E450g,
	"SUBCORE_VM_FIXED0020_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0020A150g,
	"SUBCORE_VM_FIXED0040_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0040A150g,
	"SUBCORE_VM_FIXED0060_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0060A150g,
	"SUBCORE_VM_FIXED0080_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0080A150g,
	"SUBCORE_VM_FIXED0100_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0100A150g,
	"SUBCORE_VM_FIXED0120_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0120A150g,
	"SUBCORE_VM_FIXED0140_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0140A150g,
	"SUBCORE_VM_FIXED0160_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0160A150g,
	"SUBCORE_VM_FIXED0180_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0180A150g,
	"SUBCORE_VM_FIXED0200_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0200A150g,
	"SUBCORE_VM_FIXED0220_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0220A150g,
	"SUBCORE_VM_FIXED0240_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0240A150g,
	"SUBCORE_VM_FIXED0260_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0260A150g,
	"SUBCORE_VM_FIXED0280_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0280A150g,
	"SUBCORE_VM_FIXED0300_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0300A150g,
	"SUBCORE_VM_FIXED0320_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0320A150g,
	"SUBCORE_VM_FIXED0340_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0340A150g,
	"SUBCORE_VM_FIXED0360_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0360A150g,
	"SUBCORE_VM_FIXED0380_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0380A150g,
	"SUBCORE_VM_FIXED0400_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0400A150g,
	"SUBCORE_VM_FIXED0420_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0420A150g,
	"SUBCORE_VM_FIXED0440_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0440A150g,
	"SUBCORE_VM_FIXED0460_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0460A150g,
	"SUBCORE_VM_FIXED0480_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0480A150g,
	"SUBCORE_VM_FIXED0500_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0500A150g,
	"SUBCORE_VM_FIXED0520_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0520A150g,
	"SUBCORE_VM_FIXED0540_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0540A150g,
	"SUBCORE_VM_FIXED0560_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0560A150g,
	"SUBCORE_VM_FIXED0580_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0580A150g,
	"SUBCORE_VM_FIXED0600_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0600A150g,
	"SUBCORE_VM_FIXED0620_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0620A150g,
	"SUBCORE_VM_FIXED0640_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0640A150g,
	"SUBCORE_VM_FIXED0660_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0660A150g,
	"SUBCORE_VM_FIXED0680_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0680A150g,
	"SUBCORE_VM_FIXED0700_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0700A150g,
	"SUBCORE_VM_FIXED0720_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0720A150g,
	"SUBCORE_VM_FIXED0740_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0740A150g,
	"SUBCORE_VM_FIXED0760_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0760A150g,
	"SUBCORE_VM_FIXED0780_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0780A150g,
	"SUBCORE_VM_FIXED0800_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0800A150g,
	"SUBCORE_VM_FIXED0820_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0820A150g,
	"SUBCORE_VM_FIXED0840_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0840A150g,
	"SUBCORE_VM_FIXED0860_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0860A150g,
	"SUBCORE_VM_FIXED0880_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0880A150g,
	"SUBCORE_VM_FIXED0900_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0900A150g,
	"SUBCORE_VM_FIXED0920_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0920A150g,
	"SUBCORE_VM_FIXED0940_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0940A150g,
	"SUBCORE_VM_FIXED0960_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0960A150g,
	"SUBCORE_VM_FIXED0980_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0980A150g,
	"SUBCORE_VM_FIXED1000_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1000A150g,
	"SUBCORE_VM_FIXED1020_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1020A150g,
	"SUBCORE_VM_FIXED1040_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1040A150g,
	"SUBCORE_VM_FIXED1060_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1060A150g,
	"SUBCORE_VM_FIXED1080_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1080A150g,
	"SUBCORE_VM_FIXED1100_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1100A150g,
	"SUBCORE_VM_FIXED1120_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1120A150g,
	"SUBCORE_VM_FIXED1140_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1140A150g,
	"SUBCORE_VM_FIXED1160_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1160A150g,
	"SUBCORE_VM_FIXED1180_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1180A150g,
	"SUBCORE_VM_FIXED1200_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1200A150g,
	"SUBCORE_VM_FIXED1220_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1220A150g,
	"SUBCORE_VM_FIXED1240_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1240A150g,
	"SUBCORE_VM_FIXED1260_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1260A150g,
	"SUBCORE_VM_FIXED1280_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1280A150g,
	"SUBCORE_VM_FIXED1300_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1300A150g,
	"SUBCORE_VM_FIXED1320_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1320A150g,
	"SUBCORE_VM_FIXED1340_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1340A150g,
	"SUBCORE_VM_FIXED1360_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1360A150g,
	"SUBCORE_VM_FIXED1380_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1380A150g,
	"SUBCORE_VM_FIXED1400_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1400A150g,
	"SUBCORE_VM_FIXED1420_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1420A150g,
	"SUBCORE_VM_FIXED1440_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1440A150g,
	"SUBCORE_VM_FIXED1460_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1460A150g,
	"SUBCORE_VM_FIXED1480_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1480A150g,
	"SUBCORE_VM_FIXED1500_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1500A150g,
	"SUBCORE_VM_FIXED1520_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1520A150g,
	"SUBCORE_VM_FIXED1540_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1540A150g,
	"SUBCORE_VM_FIXED1560_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1560A150g,
	"SUBCORE_VM_FIXED1580_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1580A150g,
	"SUBCORE_VM_FIXED1600_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1600A150g,
	"SUBCORE_VM_FIXED1620_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1620A150g,
	"SUBCORE_VM_FIXED1640_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1640A150g,
	"SUBCORE_VM_FIXED1660_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1660A150g,
	"SUBCORE_VM_FIXED1680_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1680A150g,
	"SUBCORE_VM_FIXED1700_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1700A150g,
	"SUBCORE_VM_FIXED1720_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1720A150g,
	"SUBCORE_VM_FIXED1740_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1740A150g,
	"SUBCORE_VM_FIXED1760_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1760A150g,
	"SUBCORE_VM_FIXED1780_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1780A150g,
	"SUBCORE_VM_FIXED1800_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1800A150g,
	"SUBCORE_VM_FIXED1820_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1820A150g,
	"SUBCORE_VM_FIXED1840_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1840A150g,
	"SUBCORE_VM_FIXED1860_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1860A150g,
	"SUBCORE_VM_FIXED1880_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1880A150g,
	"SUBCORE_VM_FIXED1900_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1900A150g,
	"SUBCORE_VM_FIXED1920_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1920A150g,
	"SUBCORE_VM_FIXED1940_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1940A150g,
	"SUBCORE_VM_FIXED1960_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1960A150g,
	"SUBCORE_VM_FIXED1980_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1980A150g,
	"SUBCORE_VM_FIXED2000_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2000A150g,
	"SUBCORE_VM_FIXED2020_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2020A150g,
	"SUBCORE_VM_FIXED2040_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2040A150g,
	"SUBCORE_VM_FIXED2060_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2060A150g,
	"SUBCORE_VM_FIXED2080_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2080A150g,
	"SUBCORE_VM_FIXED2100_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2100A150g,
	"SUBCORE_VM_FIXED2120_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2120A150g,
	"SUBCORE_VM_FIXED2140_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2140A150g,
	"SUBCORE_VM_FIXED2160_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2160A150g,
	"SUBCORE_VM_FIXED2180_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2180A150g,
	"SUBCORE_VM_FIXED2200_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2200A150g,
	"SUBCORE_VM_FIXED2220_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2220A150g,
	"SUBCORE_VM_FIXED2240_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2240A150g,
	"SUBCORE_VM_FIXED2260_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2260A150g,
	"SUBCORE_VM_FIXED2280_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2280A150g,
	"SUBCORE_VM_FIXED2300_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2300A150g,
	"SUBCORE_VM_FIXED2320_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2320A150g,
	"SUBCORE_VM_FIXED2340_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2340A150g,
	"SUBCORE_VM_FIXED2360_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2360A150g,
	"SUBCORE_VM_FIXED2380_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2380A150g,
	"SUBCORE_VM_FIXED2400_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2400A150g,
	"SUBCORE_VM_FIXED2420_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2420A150g,
	"SUBCORE_VM_FIXED2440_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2440A150g,
	"SUBCORE_VM_FIXED2460_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2460A150g,
	"SUBCORE_VM_FIXED2480_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2480A150g,
	"SUBCORE_VM_FIXED2500_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2500A150g,
	"SUBCORE_VM_FIXED2520_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2520A150g,
	"SUBCORE_VM_FIXED2540_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2540A150g,
	"SUBCORE_VM_FIXED2560_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2560A150g,
	"SUBCORE_VM_FIXED2580_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2580A150g,
	"SUBCORE_VM_FIXED2600_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2600A150g,
	"SUBCORE_VM_FIXED2620_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2620A150g,
	"SUBCORE_VM_FIXED2640_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2640A150g,
	"SUBCORE_VM_FIXED2660_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2660A150g,
	"SUBCORE_VM_FIXED2680_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2680A150g,
	"SUBCORE_VM_FIXED2700_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2700A150g,
	"SUBCORE_VM_FIXED2720_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2720A150g,
	"SUBCORE_VM_FIXED2740_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2740A150g,
	"SUBCORE_VM_FIXED2760_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2760A150g,
	"SUBCORE_VM_FIXED2780_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2780A150g,
	"SUBCORE_VM_FIXED2800_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2800A150g,
	"SUBCORE_VM_FIXED2820_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2820A150g,
	"SUBCORE_VM_FIXED2840_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2840A150g,
	"SUBCORE_VM_FIXED2860_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2860A150g,
	"SUBCORE_VM_FIXED2880_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2880A150g,
	"SUBCORE_VM_FIXED2900_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2900A150g,
	"SUBCORE_VM_FIXED2920_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2920A150g,
	"SUBCORE_VM_FIXED2940_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2940A150g,
	"SUBCORE_VM_FIXED2960_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2960A150g,
	"SUBCORE_VM_FIXED2980_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2980A150g,
	"SUBCORE_VM_FIXED3000_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3000A150g,
	"SUBCORE_VM_FIXED3020_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3020A150g,
	"SUBCORE_VM_FIXED3040_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3040A150g,
	"SUBCORE_VM_FIXED3060_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3060A150g,
	"SUBCORE_VM_FIXED3080_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3080A150g,
	"SUBCORE_VM_FIXED3100_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3100A150g,
	"SUBCORE_VM_FIXED3120_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3120A150g,
	"SUBCORE_VM_FIXED3140_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3140A150g,
	"SUBCORE_VM_FIXED3160_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3160A150g,
	"SUBCORE_VM_FIXED3180_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3180A150g,
	"SUBCORE_VM_FIXED3200_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3200A150g,
	"SUBCORE_VM_FIXED3220_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3220A150g,
	"SUBCORE_VM_FIXED3240_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3240A150g,
	"SUBCORE_VM_FIXED3260_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3260A150g,
	"SUBCORE_VM_FIXED3280_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3280A150g,
	"SUBCORE_VM_FIXED3300_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3300A150g,
	"SUBCORE_VM_FIXED3320_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3320A150g,
	"SUBCORE_VM_FIXED3340_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3340A150g,
	"SUBCORE_VM_FIXED3360_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3360A150g,
	"SUBCORE_VM_FIXED3380_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3380A150g,
	"SUBCORE_VM_FIXED3400_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3400A150g,
	"SUBCORE_VM_FIXED3420_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3420A150g,
	"SUBCORE_VM_FIXED3440_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3440A150g,
	"SUBCORE_VM_FIXED3460_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3460A150g,
	"SUBCORE_VM_FIXED3480_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3480A150g,
	"SUBCORE_VM_FIXED3500_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3500A150g,
	"SUBCORE_VM_FIXED3520_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3520A150g,
	"SUBCORE_VM_FIXED3540_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3540A150g,
	"SUBCORE_VM_FIXED3560_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3560A150g,
	"SUBCORE_VM_FIXED3580_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3580A150g,
	"SUBCORE_VM_FIXED3600_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3600A150g,
	"SUBCORE_VM_FIXED3620_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3620A150g,
	"SUBCORE_VM_FIXED3640_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3640A150g,
	"SUBCORE_VM_FIXED3660_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3660A150g,
	"SUBCORE_VM_FIXED3680_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3680A150g,
	"SUBCORE_VM_FIXED3700_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3700A150g,
	"SUBCORE_VM_FIXED3720_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3720A150g,
	"SUBCORE_VM_FIXED3740_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3740A150g,
	"SUBCORE_VM_FIXED3760_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3760A150g,
	"SUBCORE_VM_FIXED3780_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3780A150g,
	"SUBCORE_VM_FIXED3800_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3800A150g,
	"SUBCORE_VM_FIXED3820_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3820A150g,
	"SUBCORE_VM_FIXED3840_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3840A150g,
	"SUBCORE_VM_FIXED3860_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3860A150g,
	"SUBCORE_VM_FIXED3880_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3880A150g,
	"SUBCORE_VM_FIXED3900_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3900A150g,
	"SUBCORE_VM_FIXED3920_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3920A150g,
	"SUBCORE_VM_FIXED3940_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3940A150g,
	"SUBCORE_VM_FIXED3960_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3960A150g,
	"SUBCORE_VM_FIXED3980_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3980A150g,
	"SUBCORE_VM_FIXED4000_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4000A150g,
	"SUBCORE_VM_FIXED4020_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4020A150g,
	"SUBCORE_VM_FIXED4040_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4040A150g,
	"SUBCORE_VM_FIXED4060_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4060A150g,
	"SUBCORE_VM_FIXED4080_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4080A150g,
	"SUBCORE_VM_FIXED4100_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4100A150g,
	"SUBCORE_VM_FIXED4120_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4120A150g,
	"SUBCORE_VM_FIXED4140_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4140A150g,
	"SUBCORE_VM_FIXED4160_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4160A150g,
	"SUBCORE_VM_FIXED4180_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4180A150g,
	"SUBCORE_VM_FIXED4200_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4200A150g,
	"SUBCORE_VM_FIXED4220_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4220A150g,
	"SUBCORE_VM_FIXED4240_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4240A150g,
	"SUBCORE_VM_FIXED4260_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4260A150g,
	"SUBCORE_VM_FIXED4280_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4280A150g,
	"SUBCORE_VM_FIXED4300_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4300A150g,
	"SUBCORE_VM_FIXED4320_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4320A150g,
	"SUBCORE_VM_FIXED4340_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4340A150g,
	"SUBCORE_VM_FIXED4360_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4360A150g,
	"SUBCORE_VM_FIXED4380_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4380A150g,
	"SUBCORE_VM_FIXED4400_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4400A150g,
	"SUBCORE_VM_FIXED4420_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4420A150g,
	"SUBCORE_VM_FIXED4440_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4440A150g,
	"SUBCORE_VM_FIXED4460_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4460A150g,
	"SUBCORE_VM_FIXED4480_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4480A150g,
	"SUBCORE_VM_FIXED4500_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4500A150g,
	"SUBCORE_VM_FIXED4520_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4520A150g,
	"SUBCORE_VM_FIXED4540_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4540A150g,
	"SUBCORE_VM_FIXED4560_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4560A150g,
	"SUBCORE_VM_FIXED4580_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4580A150g,
	"SUBCORE_VM_FIXED4600_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4600A150g,
	"SUBCORE_VM_FIXED4620_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4620A150g,
	"SUBCORE_VM_FIXED4640_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4640A150g,
	"SUBCORE_VM_FIXED4660_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4660A150g,
	"SUBCORE_VM_FIXED4680_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4680A150g,
	"SUBCORE_VM_FIXED4700_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4700A150g,
	"SUBCORE_VM_FIXED4720_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4720A150g,
	"SUBCORE_VM_FIXED4740_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4740A150g,
	"SUBCORE_VM_FIXED4760_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4760A150g,
	"SUBCORE_VM_FIXED4780_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4780A150g,
	"SUBCORE_VM_FIXED4800_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4800A150g,
	"SUBCORE_VM_FIXED4820_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4820A150g,
	"SUBCORE_VM_FIXED4840_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4840A150g,
	"SUBCORE_VM_FIXED4860_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4860A150g,
	"SUBCORE_VM_FIXED4880_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4880A150g,
	"SUBCORE_VM_FIXED4900_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4900A150g,
	"SUBCORE_VM_FIXED4920_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4920A150g,
	"SUBCORE_VM_FIXED4940_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4940A150g,
	"SUBCORE_VM_FIXED4960_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4960A150g,
	"SUBCORE_VM_FIXED4980_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4980A150g,
	"SUBCORE_VM_FIXED5000_A1_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed5000A150g,
	"SUBCORE_VM_FIXED0090_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0090X950g,
	"SUBCORE_VM_FIXED0180_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0180X950g,
	"SUBCORE_VM_FIXED0270_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0270X950g,
	"SUBCORE_VM_FIXED0360_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0360X950g,
	"SUBCORE_VM_FIXED0450_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0450X950g,
	"SUBCORE_VM_FIXED0540_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0540X950g,
	"SUBCORE_VM_FIXED0630_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0630X950g,
	"SUBCORE_VM_FIXED0720_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0720X950g,
	"SUBCORE_VM_FIXED0810_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0810X950g,
	"SUBCORE_VM_FIXED0900_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0900X950g,
	"SUBCORE_VM_FIXED0990_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed0990X950g,
	"SUBCORE_VM_FIXED1080_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1080X950g,
	"SUBCORE_VM_FIXED1170_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1170X950g,
	"SUBCORE_VM_FIXED1260_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1260X950g,
	"SUBCORE_VM_FIXED1350_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1350X950g,
	"SUBCORE_VM_FIXED1440_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1440X950g,
	"SUBCORE_VM_FIXED1530_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1530X950g,
	"SUBCORE_VM_FIXED1620_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1620X950g,
	"SUBCORE_VM_FIXED1710_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1710X950g,
	"SUBCORE_VM_FIXED1800_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1800X950g,
	"SUBCORE_VM_FIXED1890_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1890X950g,
	"SUBCORE_VM_FIXED1980_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed1980X950g,
	"SUBCORE_VM_FIXED2070_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2070X950g,
	"SUBCORE_VM_FIXED2160_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2160X950g,
	"SUBCORE_VM_FIXED2250_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2250X950g,
	"SUBCORE_VM_FIXED2340_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2340X950g,
	"SUBCORE_VM_FIXED2430_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2430X950g,
	"SUBCORE_VM_FIXED2520_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2520X950g,
	"SUBCORE_VM_FIXED2610_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2610X950g,
	"SUBCORE_VM_FIXED2700_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2700X950g,
	"SUBCORE_VM_FIXED2790_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2790X950g,
	"SUBCORE_VM_FIXED2880_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2880X950g,
	"SUBCORE_VM_FIXED2970_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed2970X950g,
	"SUBCORE_VM_FIXED3060_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3060X950g,
	"SUBCORE_VM_FIXED3150_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3150X950g,
	"SUBCORE_VM_FIXED3240_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3240X950g,
	"SUBCORE_VM_FIXED3330_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3330X950g,
	"SUBCORE_VM_FIXED3420_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3420X950g,
	"SUBCORE_VM_FIXED3510_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3510X950g,
	"SUBCORE_VM_FIXED3600_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3600X950g,
	"SUBCORE_VM_FIXED3690_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3690X950g,
	"SUBCORE_VM_FIXED3780_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3780X950g,
	"SUBCORE_VM_FIXED3870_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3870X950g,
	"SUBCORE_VM_FIXED3960_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed3960X950g,
	"SUBCORE_VM_FIXED4050_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4050X950g,
	"SUBCORE_VM_FIXED4140_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4140X950g,
	"SUBCORE_VM_FIXED4230_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4230X950g,
	"SUBCORE_VM_FIXED4320_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4320X950g,
	"SUBCORE_VM_FIXED4410_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4410X950g,
	"SUBCORE_VM_FIXED4500_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4500X950g,
	"SUBCORE_VM_FIXED4590_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4590X950g,
	"SUBCORE_VM_FIXED4680_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4680X950g,
	"SUBCORE_VM_FIXED4770_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4770X950g,
	"SUBCORE_VM_FIXED4860_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4860X950g,
	"SUBCORE_VM_FIXED4950_X9_50G":  UpdateVnicShapeDetailsVnicShapeSubcoreVmFixed4950X950g,
	"DYNAMIC_A1_50G":               UpdateVnicShapeDetailsVnicShapeDynamicA150g,
	"FIXED0040_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed0040A150g,
	"FIXED0100_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed0100A150g,
	"FIXED0200_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed0200A150g,
	"FIXED0300_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed0300A150g,
	"FIXED0400_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed0400A150g,
	"FIXED0500_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed0500A150g,
	"FIXED0600_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed0600A150g,
	"FIXED0700_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed0700A150g,
	"FIXED0800_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed0800A150g,
	"FIXED0900_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed0900A150g,
	"FIXED1000_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed1000A150g,
	"FIXED1100_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed1100A150g,
	"FIXED1200_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed1200A150g,
	"FIXED1300_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed1300A150g,
	"FIXED1400_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed1400A150g,
	"FIXED1500_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed1500A150g,
	"FIXED1600_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed1600A150g,
	"FIXED1700_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed1700A150g,
	"FIXED1800_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed1800A150g,
	"FIXED1900_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed1900A150g,
	"FIXED2000_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed2000A150g,
	"FIXED2100_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed2100A150g,
	"FIXED2200_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed2200A150g,
	"FIXED2300_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed2300A150g,
	"FIXED2400_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed2400A150g,
	"FIXED2500_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed2500A150g,
	"FIXED2600_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed2600A150g,
	"FIXED2700_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed2700A150g,
	"FIXED2800_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed2800A150g,
	"FIXED2900_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed2900A150g,
	"FIXED3000_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed3000A150g,
	"FIXED3100_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed3100A150g,
	"FIXED3200_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed3200A150g,
	"FIXED3300_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed3300A150g,
	"FIXED3400_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed3400A150g,
	"FIXED3500_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed3500A150g,
	"FIXED3600_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed3600A150g,
	"FIXED3700_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed3700A150g,
	"FIXED3800_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed3800A150g,
	"FIXED3900_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed3900A150g,
	"FIXED4000_A1_50G":             UpdateVnicShapeDetailsVnicShapeFixed4000A150g,
	"ENTIREHOST_A1_50G":            UpdateVnicShapeDetailsVnicShapeEntirehostA150g,
	"DYNAMIC_X9_50G":               UpdateVnicShapeDetailsVnicShapeDynamicX950g,
	"FIXED0040_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed0040X950g,
	"FIXED0400_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed0400X950g,
	"FIXED0800_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed0800X950g,
	"FIXED1200_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed1200X950g,
	"FIXED1600_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed1600X950g,
	"FIXED2000_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed2000X950g,
	"FIXED2400_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed2400X950g,
	"FIXED2800_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed2800X950g,
	"FIXED3200_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed3200X950g,
	"FIXED3600_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed3600X950g,
	"FIXED4000_X9_50G":             UpdateVnicShapeDetailsVnicShapeFixed4000X950g,
	"STANDARD_VM_FIXED0100_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed0100X950g,
	"STANDARD_VM_FIXED0200_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed0200X950g,
	"STANDARD_VM_FIXED0300_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed0300X950g,
	"STANDARD_VM_FIXED0400_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed0400X950g,
	"STANDARD_VM_FIXED0500_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed0500X950g,
	"STANDARD_VM_FIXED0600_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed0600X950g,
	"STANDARD_VM_FIXED0700_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed0700X950g,
	"STANDARD_VM_FIXED0800_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed0800X950g,
	"STANDARD_VM_FIXED0900_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed0900X950g,
	"STANDARD_VM_FIXED1000_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed1000X950g,
	"STANDARD_VM_FIXED1100_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed1100X950g,
	"STANDARD_VM_FIXED1200_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed1200X950g,
	"STANDARD_VM_FIXED1300_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed1300X950g,
	"STANDARD_VM_FIXED1400_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed1400X950g,
	"STANDARD_VM_FIXED1500_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed1500X950g,
	"STANDARD_VM_FIXED1600_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed1600X950g,
	"STANDARD_VM_FIXED1700_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed1700X950g,
	"STANDARD_VM_FIXED1800_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed1800X950g,
	"STANDARD_VM_FIXED1900_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed1900X950g,
	"STANDARD_VM_FIXED2000_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed2000X950g,
	"STANDARD_VM_FIXED2100_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed2100X950g,
	"STANDARD_VM_FIXED2200_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed2200X950g,
	"STANDARD_VM_FIXED2300_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed2300X950g,
	"STANDARD_VM_FIXED2400_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed2400X950g,
	"STANDARD_VM_FIXED2500_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed2500X950g,
	"STANDARD_VM_FIXED2600_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed2600X950g,
	"STANDARD_VM_FIXED2700_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed2700X950g,
	"STANDARD_VM_FIXED2800_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed2800X950g,
	"STANDARD_VM_FIXED2900_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed2900X950g,
	"STANDARD_VM_FIXED3000_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed3000X950g,
	"STANDARD_VM_FIXED3100_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed3100X950g,
	"STANDARD_VM_FIXED3200_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed3200X950g,
	"STANDARD_VM_FIXED3300_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed3300X950g,
	"STANDARD_VM_FIXED3400_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed3400X950g,
	"STANDARD_VM_FIXED3500_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed3500X950g,
	"STANDARD_VM_FIXED3600_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed3600X950g,
	"STANDARD_VM_FIXED3700_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed3700X950g,
	"STANDARD_VM_FIXED3800_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed3800X950g,
	"STANDARD_VM_FIXED3900_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed3900X950g,
	"STANDARD_VM_FIXED4000_X9_50G": UpdateVnicShapeDetailsVnicShapeStandardVmFixed4000X950g,
	"ENTIREHOST_X9_50G":            UpdateVnicShapeDetailsVnicShapeEntirehostX950g,
}

// GetUpdateVnicShapeDetailsVnicShapeEnumValues Enumerates the set of values for UpdateVnicShapeDetailsVnicShapeEnum
func GetUpdateVnicShapeDetailsVnicShapeEnumValues() []UpdateVnicShapeDetailsVnicShapeEnum {
	values := make([]UpdateVnicShapeDetailsVnicShapeEnum, 0)
	for _, v := range mappingUpdateVnicShapeDetailsVnicShapeEnum {
		values = append(values, v)
	}
	return values
}

// GetUpdateVnicShapeDetailsVnicShapeEnumStringValues Enumerates the set of values in String for UpdateVnicShapeDetailsVnicShapeEnum
func GetUpdateVnicShapeDetailsVnicShapeEnumStringValues() []string {
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
