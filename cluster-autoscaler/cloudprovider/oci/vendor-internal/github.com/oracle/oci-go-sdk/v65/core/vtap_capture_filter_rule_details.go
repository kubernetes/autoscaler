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

// VtapCaptureFilterRuleDetails This resource contains the rules governing what traffic a VTAP mirrors.
type VtapCaptureFilterRuleDetails struct {

	// The traffic direction the VTAP is configured to mirror.
	TrafficDirection VtapCaptureFilterRuleDetailsTrafficDirectionEnum `mandatory:"true" json:"trafficDirection"`

	// Include or exclude packets meeting this definition from mirrored traffic.
	RuleAction VtapCaptureFilterRuleDetailsRuleActionEnum `mandatory:"false" json:"ruleAction,omitempty"`

	// Traffic from this CIDR block to the VTAP source will be mirrored to the VTAP target.
	SourceCidr *string `mandatory:"false" json:"sourceCidr"`

	// Traffic sent to this CIDR block through the VTAP source will be mirrored to the VTAP target.
	DestinationCidr *string `mandatory:"false" json:"destinationCidr"`

	// The transport protocol used in the filter. If do not choose a protocol, all protocols will be used in the filter.
	// Supported options are:
	//   * 1 = ICMP
	//   * 6 = TCP
	//   * 17 = UDP
	Protocol *string `mandatory:"false" json:"protocol"`

	IcmpOptions *IcmpOptions `mandatory:"false" json:"icmpOptions"`

	TcpOptions *TcpOptions `mandatory:"false" json:"tcpOptions"`

	UdpOptions *UdpOptions `mandatory:"false" json:"udpOptions"`
}

func (m VtapCaptureFilterRuleDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VtapCaptureFilterRuleDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingVtapCaptureFilterRuleDetailsTrafficDirectionEnum(string(m.TrafficDirection)); !ok && m.TrafficDirection != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TrafficDirection: %s. Supported values are: %s.", m.TrafficDirection, strings.Join(GetVtapCaptureFilterRuleDetailsTrafficDirectionEnumStringValues(), ",")))
	}

	if _, ok := GetMappingVtapCaptureFilterRuleDetailsRuleActionEnum(string(m.RuleAction)); !ok && m.RuleAction != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RuleAction: %s. Supported values are: %s.", m.RuleAction, strings.Join(GetVtapCaptureFilterRuleDetailsRuleActionEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// VtapCaptureFilterRuleDetailsTrafficDirectionEnum Enum with underlying type: string
type VtapCaptureFilterRuleDetailsTrafficDirectionEnum string

// Set of constants representing the allowable values for VtapCaptureFilterRuleDetailsTrafficDirectionEnum
const (
	VtapCaptureFilterRuleDetailsTrafficDirectionIngress VtapCaptureFilterRuleDetailsTrafficDirectionEnum = "INGRESS"
	VtapCaptureFilterRuleDetailsTrafficDirectionEgress  VtapCaptureFilterRuleDetailsTrafficDirectionEnum = "EGRESS"
)

var mappingVtapCaptureFilterRuleDetailsTrafficDirectionEnum = map[string]VtapCaptureFilterRuleDetailsTrafficDirectionEnum{
	"INGRESS": VtapCaptureFilterRuleDetailsTrafficDirectionIngress,
	"EGRESS":  VtapCaptureFilterRuleDetailsTrafficDirectionEgress,
}

var mappingVtapCaptureFilterRuleDetailsTrafficDirectionEnumLowerCase = map[string]VtapCaptureFilterRuleDetailsTrafficDirectionEnum{
	"ingress": VtapCaptureFilterRuleDetailsTrafficDirectionIngress,
	"egress":  VtapCaptureFilterRuleDetailsTrafficDirectionEgress,
}

// GetVtapCaptureFilterRuleDetailsTrafficDirectionEnumValues Enumerates the set of values for VtapCaptureFilterRuleDetailsTrafficDirectionEnum
func GetVtapCaptureFilterRuleDetailsTrafficDirectionEnumValues() []VtapCaptureFilterRuleDetailsTrafficDirectionEnum {
	values := make([]VtapCaptureFilterRuleDetailsTrafficDirectionEnum, 0)
	for _, v := range mappingVtapCaptureFilterRuleDetailsTrafficDirectionEnum {
		values = append(values, v)
	}
	return values
}

// GetVtapCaptureFilterRuleDetailsTrafficDirectionEnumStringValues Enumerates the set of values in String for VtapCaptureFilterRuleDetailsTrafficDirectionEnum
func GetVtapCaptureFilterRuleDetailsTrafficDirectionEnumStringValues() []string {
	return []string{
		"INGRESS",
		"EGRESS",
	}
}

// GetMappingVtapCaptureFilterRuleDetailsTrafficDirectionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVtapCaptureFilterRuleDetailsTrafficDirectionEnum(val string) (VtapCaptureFilterRuleDetailsTrafficDirectionEnum, bool) {
	enum, ok := mappingVtapCaptureFilterRuleDetailsTrafficDirectionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// VtapCaptureFilterRuleDetailsRuleActionEnum Enum with underlying type: string
type VtapCaptureFilterRuleDetailsRuleActionEnum string

// Set of constants representing the allowable values for VtapCaptureFilterRuleDetailsRuleActionEnum
const (
	VtapCaptureFilterRuleDetailsRuleActionInclude VtapCaptureFilterRuleDetailsRuleActionEnum = "INCLUDE"
	VtapCaptureFilterRuleDetailsRuleActionExclude VtapCaptureFilterRuleDetailsRuleActionEnum = "EXCLUDE"
)

var mappingVtapCaptureFilterRuleDetailsRuleActionEnum = map[string]VtapCaptureFilterRuleDetailsRuleActionEnum{
	"INCLUDE": VtapCaptureFilterRuleDetailsRuleActionInclude,
	"EXCLUDE": VtapCaptureFilterRuleDetailsRuleActionExclude,
}

var mappingVtapCaptureFilterRuleDetailsRuleActionEnumLowerCase = map[string]VtapCaptureFilterRuleDetailsRuleActionEnum{
	"include": VtapCaptureFilterRuleDetailsRuleActionInclude,
	"exclude": VtapCaptureFilterRuleDetailsRuleActionExclude,
}

// GetVtapCaptureFilterRuleDetailsRuleActionEnumValues Enumerates the set of values for VtapCaptureFilterRuleDetailsRuleActionEnum
func GetVtapCaptureFilterRuleDetailsRuleActionEnumValues() []VtapCaptureFilterRuleDetailsRuleActionEnum {
	values := make([]VtapCaptureFilterRuleDetailsRuleActionEnum, 0)
	for _, v := range mappingVtapCaptureFilterRuleDetailsRuleActionEnum {
		values = append(values, v)
	}
	return values
}

// GetVtapCaptureFilterRuleDetailsRuleActionEnumStringValues Enumerates the set of values in String for VtapCaptureFilterRuleDetailsRuleActionEnum
func GetVtapCaptureFilterRuleDetailsRuleActionEnumStringValues() []string {
	return []string{
		"INCLUDE",
		"EXCLUDE",
	}
}

// GetMappingVtapCaptureFilterRuleDetailsRuleActionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVtapCaptureFilterRuleDetailsRuleActionEnum(val string) (VtapCaptureFilterRuleDetailsRuleActionEnum, bool) {
	enum, ok := mappingVtapCaptureFilterRuleDetailsRuleActionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
