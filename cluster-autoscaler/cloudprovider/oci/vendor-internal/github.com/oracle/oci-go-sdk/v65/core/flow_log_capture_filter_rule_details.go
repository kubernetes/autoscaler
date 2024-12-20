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

// FlowLogCaptureFilterRuleDetails The set of rules governing what traffic the VCN flow log collects.
type FlowLogCaptureFilterRuleDetails struct {

	// Indicates whether a VCN flow log capture filter rule is enabled.
	IsEnabled *bool `mandatory:"false" json:"isEnabled"`

	// A lower number indicates a higher priority, range 0-9. Each rule must have a distinct priority.
	Priority *int `mandatory:"false" json:"priority"`

	// Sampling interval as `1` of `X`, where `X` is an integer not greater than `100000`.
	SamplingRate *int `mandatory:"false" json:"samplingRate"`

	// Traffic from this CIDR will be captured in the VCN flow log.
	SourceCidr *string `mandatory:"false" json:"sourceCidr"`

	// Traffic to this CIDR will be captured in the VCN flow log.
	DestinationCidr *string `mandatory:"false" json:"destinationCidr"`

	// The transport protocol the filter uses.
	Protocol *string `mandatory:"false" json:"protocol"`

	IcmpOptions *IcmpOptions `mandatory:"false" json:"icmpOptions"`

	TcpOptions *TcpOptions `mandatory:"false" json:"tcpOptions"`

	UdpOptions *UdpOptions `mandatory:"false" json:"udpOptions"`

	// Type or types of VCN flow logs to store. `ALL` includes records for both accepted traffic and
	// rejected traffic.
	FlowLogType FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum `mandatory:"false" json:"flowLogType,omitempty"`

	// Include or exclude a `ruleAction` object.
	RuleAction FlowLogCaptureFilterRuleDetailsRuleActionEnum `mandatory:"false" json:"ruleAction,omitempty"`
}

func (m FlowLogCaptureFilterRuleDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m FlowLogCaptureFilterRuleDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingFlowLogCaptureFilterRuleDetailsFlowLogTypeEnum(string(m.FlowLogType)); !ok && m.FlowLogType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for FlowLogType: %s. Supported values are: %s.", m.FlowLogType, strings.Join(GetFlowLogCaptureFilterRuleDetailsFlowLogTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingFlowLogCaptureFilterRuleDetailsRuleActionEnum(string(m.RuleAction)); !ok && m.RuleAction != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for RuleAction: %s. Supported values are: %s.", m.RuleAction, strings.Join(GetFlowLogCaptureFilterRuleDetailsRuleActionEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum Enum with underlying type: string
type FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum string

// Set of constants representing the allowable values for FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum
const (
	FlowLogCaptureFilterRuleDetailsFlowLogTypeAll    FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum = "ALL"
	FlowLogCaptureFilterRuleDetailsFlowLogTypeReject FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum = "REJECT"
	FlowLogCaptureFilterRuleDetailsFlowLogTypeAccept FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum = "ACCEPT"
)

var mappingFlowLogCaptureFilterRuleDetailsFlowLogTypeEnum = map[string]FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum{
	"ALL":    FlowLogCaptureFilterRuleDetailsFlowLogTypeAll,
	"REJECT": FlowLogCaptureFilterRuleDetailsFlowLogTypeReject,
	"ACCEPT": FlowLogCaptureFilterRuleDetailsFlowLogTypeAccept,
}

var mappingFlowLogCaptureFilterRuleDetailsFlowLogTypeEnumLowerCase = map[string]FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum{
	"all":    FlowLogCaptureFilterRuleDetailsFlowLogTypeAll,
	"reject": FlowLogCaptureFilterRuleDetailsFlowLogTypeReject,
	"accept": FlowLogCaptureFilterRuleDetailsFlowLogTypeAccept,
}

// GetFlowLogCaptureFilterRuleDetailsFlowLogTypeEnumValues Enumerates the set of values for FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum
func GetFlowLogCaptureFilterRuleDetailsFlowLogTypeEnumValues() []FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum {
	values := make([]FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum, 0)
	for _, v := range mappingFlowLogCaptureFilterRuleDetailsFlowLogTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetFlowLogCaptureFilterRuleDetailsFlowLogTypeEnumStringValues Enumerates the set of values in String for FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum
func GetFlowLogCaptureFilterRuleDetailsFlowLogTypeEnumStringValues() []string {
	return []string{
		"ALL",
		"REJECT",
		"ACCEPT",
	}
}

// GetMappingFlowLogCaptureFilterRuleDetailsFlowLogTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingFlowLogCaptureFilterRuleDetailsFlowLogTypeEnum(val string) (FlowLogCaptureFilterRuleDetailsFlowLogTypeEnum, bool) {
	enum, ok := mappingFlowLogCaptureFilterRuleDetailsFlowLogTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// FlowLogCaptureFilterRuleDetailsRuleActionEnum Enum with underlying type: string
type FlowLogCaptureFilterRuleDetailsRuleActionEnum string

// Set of constants representing the allowable values for FlowLogCaptureFilterRuleDetailsRuleActionEnum
const (
	FlowLogCaptureFilterRuleDetailsRuleActionInclude FlowLogCaptureFilterRuleDetailsRuleActionEnum = "INCLUDE"
	FlowLogCaptureFilterRuleDetailsRuleActionExclude FlowLogCaptureFilterRuleDetailsRuleActionEnum = "EXCLUDE"
)

var mappingFlowLogCaptureFilterRuleDetailsRuleActionEnum = map[string]FlowLogCaptureFilterRuleDetailsRuleActionEnum{
	"INCLUDE": FlowLogCaptureFilterRuleDetailsRuleActionInclude,
	"EXCLUDE": FlowLogCaptureFilterRuleDetailsRuleActionExclude,
}

var mappingFlowLogCaptureFilterRuleDetailsRuleActionEnumLowerCase = map[string]FlowLogCaptureFilterRuleDetailsRuleActionEnum{
	"include": FlowLogCaptureFilterRuleDetailsRuleActionInclude,
	"exclude": FlowLogCaptureFilterRuleDetailsRuleActionExclude,
}

// GetFlowLogCaptureFilterRuleDetailsRuleActionEnumValues Enumerates the set of values for FlowLogCaptureFilterRuleDetailsRuleActionEnum
func GetFlowLogCaptureFilterRuleDetailsRuleActionEnumValues() []FlowLogCaptureFilterRuleDetailsRuleActionEnum {
	values := make([]FlowLogCaptureFilterRuleDetailsRuleActionEnum, 0)
	for _, v := range mappingFlowLogCaptureFilterRuleDetailsRuleActionEnum {
		values = append(values, v)
	}
	return values
}

// GetFlowLogCaptureFilterRuleDetailsRuleActionEnumStringValues Enumerates the set of values in String for FlowLogCaptureFilterRuleDetailsRuleActionEnum
func GetFlowLogCaptureFilterRuleDetailsRuleActionEnumStringValues() []string {
	return []string{
		"INCLUDE",
		"EXCLUDE",
	}
}

// GetMappingFlowLogCaptureFilterRuleDetailsRuleActionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingFlowLogCaptureFilterRuleDetailsRuleActionEnum(val string) (FlowLogCaptureFilterRuleDetailsRuleActionEnum, bool) {
	enum, ok := mappingFlowLogCaptureFilterRuleDetailsRuleActionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
