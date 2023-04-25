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

// PrivateIpNextHopTarget Details of a private IP nextHop target.
type PrivateIpNextHopTarget struct {

	// NextHop target's MPLS label.
	MplsLabel *int `mandatory:"false" json:"mplsLabel"`

	PortRange *PortRange `mandatory:"false" json:"portRange"`

	// NextHop target's substrate IP.
	SubstrateIp *string `mandatory:"false" json:"substrateIp"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the nextHop target entity.
	TargetId *string `mandatory:"false" json:"targetId"`

	// Type of nextHop target.
	TargetType PrivateIpNextHopTargetTargetTypeEnum `mandatory:"false" json:"targetType,omitempty"`
}

func (m PrivateIpNextHopTarget) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m PrivateIpNextHopTarget) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingPrivateIpNextHopTargetTargetTypeEnum[string(m.TargetType)]; !ok && m.TargetType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TargetType: %s. Supported values are: %s.", m.TargetType, strings.Join(GetPrivateIpNextHopTargetTargetTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// PrivateIpNextHopTargetTargetTypeEnum Enum with underlying type: string
type PrivateIpNextHopTargetTargetTypeEnum string

// Set of constants representing the allowable values for PrivateIpNextHopTargetTargetTypeEnum
const (
	PrivateIpNextHopTargetTargetTypePadp       PrivateIpNextHopTargetTargetTypeEnum = "PADP"
	PrivateIpNextHopTargetTargetTypeVnicWorker PrivateIpNextHopTargetTargetTypeEnum = "VNIC_WORKER"
)

var mappingPrivateIpNextHopTargetTargetTypeEnum = map[string]PrivateIpNextHopTargetTargetTypeEnum{
	"PADP":        PrivateIpNextHopTargetTargetTypePadp,
	"VNIC_WORKER": PrivateIpNextHopTargetTargetTypeVnicWorker,
}

// GetPrivateIpNextHopTargetTargetTypeEnumValues Enumerates the set of values for PrivateIpNextHopTargetTargetTypeEnum
func GetPrivateIpNextHopTargetTargetTypeEnumValues() []PrivateIpNextHopTargetTargetTypeEnum {
	values := make([]PrivateIpNextHopTargetTargetTypeEnum, 0)
	for _, v := range mappingPrivateIpNextHopTargetTargetTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetPrivateIpNextHopTargetTargetTypeEnumStringValues Enumerates the set of values in String for PrivateIpNextHopTargetTargetTypeEnum
func GetPrivateIpNextHopTargetTargetTypeEnumStringValues() []string {
	return []string{
		"PADP",
		"VNIC_WORKER",
	}
}
