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
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// FlowLogConfig *Flow logs* record information about traffic that is either allowed or rejected by the
// SecurityList that control traffic in and out of a
// Vnic.
// A *flow log configuration* (`FlowLogConfig`) contains information about where to store flow
// logs (an Object Storage bucket in your tenancy), and the type of logs to store.
// **Important:** For logs to be placed in the Object Storage bucket listed in the configuration,
// an administrator must create an IAM policy in your tenancy that lets the Networking service
// put objects in that bucket. Otherwise, no flow logs can be written to the bucket.
// Here's the required policy (which consists of three separate statements):
// `define tenancy VcnFlowLogs as ocid1.tenancy.oc1..<var>&lt;unique_ID&gt;</var>`
// `define dynamic-group FlowLogsDynamicGroup as ocid1.dynamicgroup.oc1..<var>&lt;unique_ID&gt;</var>`
// `admit dynamic-group FlowLogsDynamicGroup of tenancy VcnFlowLogs to manage objects in tenancy where target.bucket.name='yourbucketname'`
// To enable flow logs for a subnet: after creating a flow
// log configuration, attach the flow log configuration to that subnet. See
// FlowLogConfigAttachment and
// CreateFlowLogConfigAttachment.
type FlowLogConfig struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the flow
	// log configuration.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The flow log configuration's OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
	Id *string `mandatory:"true" json:"id"`

	// The flow log configuration's current state.
	LifecycleState FlowLogConfigLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Type or types of flow logs to store. `ALL` includes records for both accepted traffic and
	// rejected traffic.
	FlowLogType FlowLogConfigFlowLogTypeEnum `mandatory:"true" json:"flowLogType"`

	Destination FlowLogDestination `mandatory:"true" json:"destination"`

	// The date and time the flow log configuration was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m FlowLogConfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m FlowLogConfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingFlowLogConfigLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetFlowLogConfigLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := mappingFlowLogConfigFlowLogTypeEnum[string(m.FlowLogType)]; !ok && m.FlowLogType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for FlowLogType: %s. Supported values are: %s.", m.FlowLogType, strings.Join(GetFlowLogConfigFlowLogTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *FlowLogConfig) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DefinedTags    map[string]map[string]interface{} `json:"definedTags"`
		FreeformTags   map[string]string                 `json:"freeformTags"`
		CompartmentId  *string                           `json:"compartmentId"`
		DisplayName    *string                           `json:"displayName"`
		Id             *string                           `json:"id"`
		LifecycleState FlowLogConfigLifecycleStateEnum   `json:"lifecycleState"`
		FlowLogType    FlowLogConfigFlowLogTypeEnum      `json:"flowLogType"`
		Destination    flowlogdestination                `json:"destination"`
		TimeCreated    *common.SDKTime                   `json:"timeCreated"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DefinedTags = model.DefinedTags

	m.FreeformTags = model.FreeformTags

	m.CompartmentId = model.CompartmentId

	m.DisplayName = model.DisplayName

	m.Id = model.Id

	m.LifecycleState = model.LifecycleState

	m.FlowLogType = model.FlowLogType

	nn, e = model.Destination.UnmarshalPolymorphicJSON(model.Destination.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.Destination = nn.(FlowLogDestination)
	} else {
		m.Destination = nil
	}

	m.TimeCreated = model.TimeCreated

	return
}

// FlowLogConfigLifecycleStateEnum Enum with underlying type: string
type FlowLogConfigLifecycleStateEnum string

// Set of constants representing the allowable values for FlowLogConfigLifecycleStateEnum
const (
	FlowLogConfigLifecycleStateProvisioning FlowLogConfigLifecycleStateEnum = "PROVISIONING"
	FlowLogConfigLifecycleStateAvailable    FlowLogConfigLifecycleStateEnum = "AVAILABLE"
	FlowLogConfigLifecycleStateTerminating  FlowLogConfigLifecycleStateEnum = "TERMINATING"
	FlowLogConfigLifecycleStateTerminated   FlowLogConfigLifecycleStateEnum = "TERMINATED"
)

var mappingFlowLogConfigLifecycleStateEnum = map[string]FlowLogConfigLifecycleStateEnum{
	"PROVISIONING": FlowLogConfigLifecycleStateProvisioning,
	"AVAILABLE":    FlowLogConfigLifecycleStateAvailable,
	"TERMINATING":  FlowLogConfigLifecycleStateTerminating,
	"TERMINATED":   FlowLogConfigLifecycleStateTerminated,
}

// GetFlowLogConfigLifecycleStateEnumValues Enumerates the set of values for FlowLogConfigLifecycleStateEnum
func GetFlowLogConfigLifecycleStateEnumValues() []FlowLogConfigLifecycleStateEnum {
	values := make([]FlowLogConfigLifecycleStateEnum, 0)
	for _, v := range mappingFlowLogConfigLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetFlowLogConfigLifecycleStateEnumStringValues Enumerates the set of values in String for FlowLogConfigLifecycleStateEnum
func GetFlowLogConfigLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// FlowLogConfigFlowLogTypeEnum Enum with underlying type: string
type FlowLogConfigFlowLogTypeEnum string

// Set of constants representing the allowable values for FlowLogConfigFlowLogTypeEnum
const (
	FlowLogConfigFlowLogTypeAll FlowLogConfigFlowLogTypeEnum = "ALL"
)

var mappingFlowLogConfigFlowLogTypeEnum = map[string]FlowLogConfigFlowLogTypeEnum{
	"ALL": FlowLogConfigFlowLogTypeAll,
}

// GetFlowLogConfigFlowLogTypeEnumValues Enumerates the set of values for FlowLogConfigFlowLogTypeEnum
func GetFlowLogConfigFlowLogTypeEnumValues() []FlowLogConfigFlowLogTypeEnum {
	values := make([]FlowLogConfigFlowLogTypeEnum, 0)
	for _, v := range mappingFlowLogConfigFlowLogTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetFlowLogConfigFlowLogTypeEnumStringValues Enumerates the set of values in String for FlowLogConfigFlowLogTypeEnum
func GetFlowLogConfigFlowLogTypeEnumStringValues() []string {
	return []string{
		"ALL",
	}
}
