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

// FlowLogConfigAttachment Represents an attachment between a flow log configuration and a resource such as a subnet. By
// creating a `FlowLogConfigAttachment`, you turn on flow logs for the attached resource. See
// CreateFlowLogConfigAttachment.
type FlowLogConfigAttachment struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the
	// flow log configuration attachment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The flow log configuration attachment's OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
	Id *string `mandatory:"true" json:"id"`

	// The flow log configuration attachment's current state.
	LifecycleState FlowLogConfigAttachmentLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the resource that the flow log
	// configuration is attached to.
	TargetEntityId *string `mandatory:"true" json:"targetEntityId"`

	// The type of resource that the flow log configuration is attached to.
	TargetEntityType FlowLogConfigAttachmentTargetEntityTypeEnum `mandatory:"true" json:"targetEntityType"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the flow log configuration that
	// is attached to the resource.
	FlowLogConfigId *string `mandatory:"true" json:"flowLogConfigId"`

	// The date and time the flow log configuration attachment was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`
}

func (m FlowLogConfigAttachment) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m FlowLogConfigAttachment) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingFlowLogConfigAttachmentLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetFlowLogConfigAttachmentLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := mappingFlowLogConfigAttachmentTargetEntityTypeEnum[string(m.TargetEntityType)]; !ok && m.TargetEntityType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TargetEntityType: %s. Supported values are: %s.", m.TargetEntityType, strings.Join(GetFlowLogConfigAttachmentTargetEntityTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// FlowLogConfigAttachmentLifecycleStateEnum Enum with underlying type: string
type FlowLogConfigAttachmentLifecycleStateEnum string

// Set of constants representing the allowable values for FlowLogConfigAttachmentLifecycleStateEnum
const (
	FlowLogConfigAttachmentLifecycleStateProvisioning FlowLogConfigAttachmentLifecycleStateEnum = "PROVISIONING"
	FlowLogConfigAttachmentLifecycleStateAvailable    FlowLogConfigAttachmentLifecycleStateEnum = "AVAILABLE"
	FlowLogConfigAttachmentLifecycleStateTerminating  FlowLogConfigAttachmentLifecycleStateEnum = "TERMINATING"
	FlowLogConfigAttachmentLifecycleStateTerminated   FlowLogConfigAttachmentLifecycleStateEnum = "TERMINATED"
)

var mappingFlowLogConfigAttachmentLifecycleStateEnum = map[string]FlowLogConfigAttachmentLifecycleStateEnum{
	"PROVISIONING": FlowLogConfigAttachmentLifecycleStateProvisioning,
	"AVAILABLE":    FlowLogConfigAttachmentLifecycleStateAvailable,
	"TERMINATING":  FlowLogConfigAttachmentLifecycleStateTerminating,
	"TERMINATED":   FlowLogConfigAttachmentLifecycleStateTerminated,
}

// GetFlowLogConfigAttachmentLifecycleStateEnumValues Enumerates the set of values for FlowLogConfigAttachmentLifecycleStateEnum
func GetFlowLogConfigAttachmentLifecycleStateEnumValues() []FlowLogConfigAttachmentLifecycleStateEnum {
	values := make([]FlowLogConfigAttachmentLifecycleStateEnum, 0)
	for _, v := range mappingFlowLogConfigAttachmentLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetFlowLogConfigAttachmentLifecycleStateEnumStringValues Enumerates the set of values in String for FlowLogConfigAttachmentLifecycleStateEnum
func GetFlowLogConfigAttachmentLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// FlowLogConfigAttachmentTargetEntityTypeEnum Enum with underlying type: string
type FlowLogConfigAttachmentTargetEntityTypeEnum string

// Set of constants representing the allowable values for FlowLogConfigAttachmentTargetEntityTypeEnum
const (
	FlowLogConfigAttachmentTargetEntityTypeSubnet FlowLogConfigAttachmentTargetEntityTypeEnum = "SUBNET"
)

var mappingFlowLogConfigAttachmentTargetEntityTypeEnum = map[string]FlowLogConfigAttachmentTargetEntityTypeEnum{
	"SUBNET": FlowLogConfigAttachmentTargetEntityTypeSubnet,
}

// GetFlowLogConfigAttachmentTargetEntityTypeEnumValues Enumerates the set of values for FlowLogConfigAttachmentTargetEntityTypeEnum
func GetFlowLogConfigAttachmentTargetEntityTypeEnumValues() []FlowLogConfigAttachmentTargetEntityTypeEnum {
	values := make([]FlowLogConfigAttachmentTargetEntityTypeEnum, 0)
	for _, v := range mappingFlowLogConfigAttachmentTargetEntityTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetFlowLogConfigAttachmentTargetEntityTypeEnumStringValues Enumerates the set of values in String for FlowLogConfigAttachmentTargetEntityTypeEnum
func GetFlowLogConfigAttachmentTargetEntityTypeEnumStringValues() []string {
	return []string{
		"SUBNET",
	}
}
