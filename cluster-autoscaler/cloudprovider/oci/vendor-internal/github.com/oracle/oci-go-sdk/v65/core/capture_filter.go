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

// CaptureFilter A capture filter contains a set of *CaptureFilterRuleDetails* governing what traffic is
// mirrored for a *Vtap* or captured for a *VCN Flow Log (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/vcn-flow-logs.htm)*.
// The capture filter is created with no rules defined, and it must have at least one rule to mirror traffic for the VTAP or collect VCN flow logs.
type CaptureFilter struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the capture filter.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The capture filter's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The capture filter's current administrative state.
	LifecycleState CaptureFilterLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

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

	// Indicates which service will use this capture filter
	FilterType CaptureFilterFilterTypeEnum `mandatory:"false" json:"filterType,omitempty"`

	// The date and time the capture filter was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2021-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The set of rules governing what traffic a VTAP mirrors.
	VtapCaptureFilterRules []VtapCaptureFilterRuleDetails `mandatory:"false" json:"vtapCaptureFilterRules"`

	// The set of rules governing what traffic the VCN flow log collects.
	FlowLogCaptureFilterRules []FlowLogCaptureFilterRuleDetails `mandatory:"false" json:"flowLogCaptureFilterRules"`
}

func (m CaptureFilter) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CaptureFilter) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingCaptureFilterLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetCaptureFilterLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := GetMappingCaptureFilterFilterTypeEnum(string(m.FilterType)); !ok && m.FilterType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for FilterType: %s. Supported values are: %s.", m.FilterType, strings.Join(GetCaptureFilterFilterTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CaptureFilterLifecycleStateEnum Enum with underlying type: string
type CaptureFilterLifecycleStateEnum string

// Set of constants representing the allowable values for CaptureFilterLifecycleStateEnum
const (
	CaptureFilterLifecycleStateProvisioning CaptureFilterLifecycleStateEnum = "PROVISIONING"
	CaptureFilterLifecycleStateAvailable    CaptureFilterLifecycleStateEnum = "AVAILABLE"
	CaptureFilterLifecycleStateUpdating     CaptureFilterLifecycleStateEnum = "UPDATING"
	CaptureFilterLifecycleStateTerminating  CaptureFilterLifecycleStateEnum = "TERMINATING"
	CaptureFilterLifecycleStateTerminated   CaptureFilterLifecycleStateEnum = "TERMINATED"
)

var mappingCaptureFilterLifecycleStateEnum = map[string]CaptureFilterLifecycleStateEnum{
	"PROVISIONING": CaptureFilterLifecycleStateProvisioning,
	"AVAILABLE":    CaptureFilterLifecycleStateAvailable,
	"UPDATING":     CaptureFilterLifecycleStateUpdating,
	"TERMINATING":  CaptureFilterLifecycleStateTerminating,
	"TERMINATED":   CaptureFilterLifecycleStateTerminated,
}

var mappingCaptureFilterLifecycleStateEnumLowerCase = map[string]CaptureFilterLifecycleStateEnum{
	"provisioning": CaptureFilterLifecycleStateProvisioning,
	"available":    CaptureFilterLifecycleStateAvailable,
	"updating":     CaptureFilterLifecycleStateUpdating,
	"terminating":  CaptureFilterLifecycleStateTerminating,
	"terminated":   CaptureFilterLifecycleStateTerminated,
}

// GetCaptureFilterLifecycleStateEnumValues Enumerates the set of values for CaptureFilterLifecycleStateEnum
func GetCaptureFilterLifecycleStateEnumValues() []CaptureFilterLifecycleStateEnum {
	values := make([]CaptureFilterLifecycleStateEnum, 0)
	for _, v := range mappingCaptureFilterLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetCaptureFilterLifecycleStateEnumStringValues Enumerates the set of values in String for CaptureFilterLifecycleStateEnum
func GetCaptureFilterLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"UPDATING",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingCaptureFilterLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCaptureFilterLifecycleStateEnum(val string) (CaptureFilterLifecycleStateEnum, bool) {
	enum, ok := mappingCaptureFilterLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CaptureFilterFilterTypeEnum Enum with underlying type: string
type CaptureFilterFilterTypeEnum string

// Set of constants representing the allowable values for CaptureFilterFilterTypeEnum
const (
	CaptureFilterFilterTypeVtap    CaptureFilterFilterTypeEnum = "VTAP"
	CaptureFilterFilterTypeFlowlog CaptureFilterFilterTypeEnum = "FLOWLOG"
)

var mappingCaptureFilterFilterTypeEnum = map[string]CaptureFilterFilterTypeEnum{
	"VTAP":    CaptureFilterFilterTypeVtap,
	"FLOWLOG": CaptureFilterFilterTypeFlowlog,
}

var mappingCaptureFilterFilterTypeEnumLowerCase = map[string]CaptureFilterFilterTypeEnum{
	"vtap":    CaptureFilterFilterTypeVtap,
	"flowlog": CaptureFilterFilterTypeFlowlog,
}

// GetCaptureFilterFilterTypeEnumValues Enumerates the set of values for CaptureFilterFilterTypeEnum
func GetCaptureFilterFilterTypeEnumValues() []CaptureFilterFilterTypeEnum {
	values := make([]CaptureFilterFilterTypeEnum, 0)
	for _, v := range mappingCaptureFilterFilterTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCaptureFilterFilterTypeEnumStringValues Enumerates the set of values in String for CaptureFilterFilterTypeEnum
func GetCaptureFilterFilterTypeEnumStringValues() []string {
	return []string{
		"VTAP",
		"FLOWLOG",
	}
}

// GetMappingCaptureFilterFilterTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCaptureFilterFilterTypeEnum(val string) (CaptureFilterFilterTypeEnum, bool) {
	enum, ok := mappingCaptureFilterFilterTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
