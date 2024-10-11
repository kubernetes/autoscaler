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

// InstanceMaintenanceEvent It is the event in which the maintenance action will be be performed on the customer instance on the scheduled date and time.
type InstanceMaintenanceEvent struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the maintenance event.
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the instance.
	InstanceId *string `mandatory:"true" json:"instanceId"`

	// The OCID of the compartment that contains the instance.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// This indicates the priority and allowed actions for this Maintenance. Higher priority forms of Maintenance have
	// tighter restrictions and may not be rescheduled, while lower priority/severity Maintenance can be rescheduled,
	// deferred, or even cancelled. Please see the
	// Instance Maintenance (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/placeholder.htm) documentation for details.
	MaintenanceCategory InstanceMaintenanceEventMaintenanceCategoryEnum `mandatory:"true" json:"maintenanceCategory"`

	// This is the reason that Maintenance is being performed. See
	// Instance Maintenance (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/placeholder.htm) documentation for details.
	MaintenanceReason InstanceMaintenanceEventMaintenanceReasonEnum `mandatory:"true" json:"maintenanceReason"`

	// This is the action that will be performed on the Instance by OCI when the Maintenance begins.
	InstanceAction InstanceMaintenanceEventInstanceActionEnum `mandatory:"true" json:"instanceAction"`

	// These are alternative actions to the requested instanceAction that can be taken to resolve the Maintenance.
	AlternativeResolutionActions []InstanceMaintenanceAlternativeResolutionActionsEnum `mandatory:"true" json:"alternativeResolutionActions"`

	// The beginning of the time window when Maintenance is scheduled to begin. The Maintenance will not begin before
	// this time.
	TimeWindowStart *common.SDKTime `mandatory:"true" json:"timeWindowStart"`

	// Indicates if this MaintenanceEvent is capable of being rescheduled up to the timeHardDueDate.
	CanReschedule *bool `mandatory:"true" json:"canReschedule"`

	// The date and time the maintenance event was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The current state of the maintenance event.
	LifecycleState InstanceMaintenanceEventLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The creator of the maintenance event.
	CreatedBy InstanceMaintenanceEventCreatedByEnum `mandatory:"true" json:"createdBy"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The time at which the Maintenance actually started.
	TimeStarted *common.SDKTime `mandatory:"false" json:"timeStarted"`

	// The time at which the Maintenance actually finished.
	TimeFinished *common.SDKTime `mandatory:"false" json:"timeFinished"`

	// The duration of the time window Maintenance is scheduled to begin within.
	StartWindowDuration *string `mandatory:"false" json:"startWindowDuration"`

	// This is the estimated duration of the Maintenance, once the Maintenance has entered the STARTED state.
	EstimatedDuration *string `mandatory:"false" json:"estimatedDuration"`

	// It is the scheduled hard due date and time of the maintenance event.
	// The maintenance event will happen at this time and the due date will not be extended.
	TimeHardDueDate *common.SDKTime `mandatory:"false" json:"timeHardDueDate"`

	// Provides more details about the state of the maintenance event.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// It is the descriptive information about the maintenance taking place on the customer instance.
	Description *string `mandatory:"false" json:"description"`

	// A unique identifier that will group Instances that have a relationship with one another and must be scheduled
	// together for the Maintenance to proceed. Any Instances that have a relationship with one another from a Maintenance
	// perspective will have a matching correlationToken.
	CorrelationToken *string `mandatory:"false" json:"correlationToken"`

	// For Instances that have local storage, this field is set to true when local storage
	// will be deleted as a result of the Maintenance.
	CanDeleteLocalStorage *bool `mandatory:"false" json:"canDeleteLocalStorage"`

	// Additional details of the maintenance in the form of json.
	AdditionalDetails map[string]string `mandatory:"false" json:"additionalDetails"`
}

func (m InstanceMaintenanceEvent) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstanceMaintenanceEvent) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingInstanceMaintenanceEventMaintenanceCategoryEnum(string(m.MaintenanceCategory)); !ok && m.MaintenanceCategory != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for MaintenanceCategory: %s. Supported values are: %s.", m.MaintenanceCategory, strings.Join(GetInstanceMaintenanceEventMaintenanceCategoryEnumStringValues(), ",")))
	}
	if _, ok := GetMappingInstanceMaintenanceEventMaintenanceReasonEnum(string(m.MaintenanceReason)); !ok && m.MaintenanceReason != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for MaintenanceReason: %s. Supported values are: %s.", m.MaintenanceReason, strings.Join(GetInstanceMaintenanceEventMaintenanceReasonEnumStringValues(), ",")))
	}
	if _, ok := GetMappingInstanceMaintenanceEventInstanceActionEnum(string(m.InstanceAction)); !ok && m.InstanceAction != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for InstanceAction: %s. Supported values are: %s.", m.InstanceAction, strings.Join(GetInstanceMaintenanceEventInstanceActionEnumStringValues(), ",")))
	}
	for _, val := range m.AlternativeResolutionActions {
		if _, ok := GetMappingInstanceMaintenanceAlternativeResolutionActionsEnum(string(val)); !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AlternativeResolutionActions: %s. Supported values are: %s.", val, strings.Join(GetInstanceMaintenanceAlternativeResolutionActionsEnumStringValues(), ",")))
		}
	}

	if _, ok := GetMappingInstanceMaintenanceEventLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInstanceMaintenanceEventLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingInstanceMaintenanceEventCreatedByEnum(string(m.CreatedBy)); !ok && m.CreatedBy != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for CreatedBy: %s. Supported values are: %s.", m.CreatedBy, strings.Join(GetInstanceMaintenanceEventCreatedByEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstanceMaintenanceEventMaintenanceCategoryEnum Enum with underlying type: string
type InstanceMaintenanceEventMaintenanceCategoryEnum string

// Set of constants representing the allowable values for InstanceMaintenanceEventMaintenanceCategoryEnum
const (
	InstanceMaintenanceEventMaintenanceCategoryEmergency    InstanceMaintenanceEventMaintenanceCategoryEnum = "EMERGENCY"
	InstanceMaintenanceEventMaintenanceCategoryMandatory    InstanceMaintenanceEventMaintenanceCategoryEnum = "MANDATORY"
	InstanceMaintenanceEventMaintenanceCategoryFlexible     InstanceMaintenanceEventMaintenanceCategoryEnum = "FLEXIBLE"
	InstanceMaintenanceEventMaintenanceCategoryOptional     InstanceMaintenanceEventMaintenanceCategoryEnum = "OPTIONAL"
	InstanceMaintenanceEventMaintenanceCategoryNotification InstanceMaintenanceEventMaintenanceCategoryEnum = "NOTIFICATION"
)

var mappingInstanceMaintenanceEventMaintenanceCategoryEnum = map[string]InstanceMaintenanceEventMaintenanceCategoryEnum{
	"EMERGENCY":    InstanceMaintenanceEventMaintenanceCategoryEmergency,
	"MANDATORY":    InstanceMaintenanceEventMaintenanceCategoryMandatory,
	"FLEXIBLE":     InstanceMaintenanceEventMaintenanceCategoryFlexible,
	"OPTIONAL":     InstanceMaintenanceEventMaintenanceCategoryOptional,
	"NOTIFICATION": InstanceMaintenanceEventMaintenanceCategoryNotification,
}

var mappingInstanceMaintenanceEventMaintenanceCategoryEnumLowerCase = map[string]InstanceMaintenanceEventMaintenanceCategoryEnum{
	"emergency":    InstanceMaintenanceEventMaintenanceCategoryEmergency,
	"mandatory":    InstanceMaintenanceEventMaintenanceCategoryMandatory,
	"flexible":     InstanceMaintenanceEventMaintenanceCategoryFlexible,
	"optional":     InstanceMaintenanceEventMaintenanceCategoryOptional,
	"notification": InstanceMaintenanceEventMaintenanceCategoryNotification,
}

// GetInstanceMaintenanceEventMaintenanceCategoryEnumValues Enumerates the set of values for InstanceMaintenanceEventMaintenanceCategoryEnum
func GetInstanceMaintenanceEventMaintenanceCategoryEnumValues() []InstanceMaintenanceEventMaintenanceCategoryEnum {
	values := make([]InstanceMaintenanceEventMaintenanceCategoryEnum, 0)
	for _, v := range mappingInstanceMaintenanceEventMaintenanceCategoryEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceMaintenanceEventMaintenanceCategoryEnumStringValues Enumerates the set of values in String for InstanceMaintenanceEventMaintenanceCategoryEnum
func GetInstanceMaintenanceEventMaintenanceCategoryEnumStringValues() []string {
	return []string{
		"EMERGENCY",
		"MANDATORY",
		"FLEXIBLE",
		"OPTIONAL",
		"NOTIFICATION",
	}
}

// GetMappingInstanceMaintenanceEventMaintenanceCategoryEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceMaintenanceEventMaintenanceCategoryEnum(val string) (InstanceMaintenanceEventMaintenanceCategoryEnum, bool) {
	enum, ok := mappingInstanceMaintenanceEventMaintenanceCategoryEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// InstanceMaintenanceEventMaintenanceReasonEnum Enum with underlying type: string
type InstanceMaintenanceEventMaintenanceReasonEnum string

// Set of constants representing the allowable values for InstanceMaintenanceEventMaintenanceReasonEnum
const (
	InstanceMaintenanceEventMaintenanceReasonEvacuation           InstanceMaintenanceEventMaintenanceReasonEnum = "EVACUATION"
	InstanceMaintenanceEventMaintenanceReasonEnvironmentalFactors InstanceMaintenanceEventMaintenanceReasonEnum = "ENVIRONMENTAL_FACTORS"
	InstanceMaintenanceEventMaintenanceReasonDecommission         InstanceMaintenanceEventMaintenanceReasonEnum = "DECOMMISSION"
	InstanceMaintenanceEventMaintenanceReasonHardwareReplacement  InstanceMaintenanceEventMaintenanceReasonEnum = "HARDWARE_REPLACEMENT"
	InstanceMaintenanceEventMaintenanceReasonFirmwareUpdate       InstanceMaintenanceEventMaintenanceReasonEnum = "FIRMWARE_UPDATE"
	InstanceMaintenanceEventMaintenanceReasonSecurityUpdate       InstanceMaintenanceEventMaintenanceReasonEnum = "SECURITY_UPDATE"
)

var mappingInstanceMaintenanceEventMaintenanceReasonEnum = map[string]InstanceMaintenanceEventMaintenanceReasonEnum{
	"EVACUATION":            InstanceMaintenanceEventMaintenanceReasonEvacuation,
	"ENVIRONMENTAL_FACTORS": InstanceMaintenanceEventMaintenanceReasonEnvironmentalFactors,
	"DECOMMISSION":          InstanceMaintenanceEventMaintenanceReasonDecommission,
	"HARDWARE_REPLACEMENT":  InstanceMaintenanceEventMaintenanceReasonHardwareReplacement,
	"FIRMWARE_UPDATE":       InstanceMaintenanceEventMaintenanceReasonFirmwareUpdate,
	"SECURITY_UPDATE":       InstanceMaintenanceEventMaintenanceReasonSecurityUpdate,
}

var mappingInstanceMaintenanceEventMaintenanceReasonEnumLowerCase = map[string]InstanceMaintenanceEventMaintenanceReasonEnum{
	"evacuation":            InstanceMaintenanceEventMaintenanceReasonEvacuation,
	"environmental_factors": InstanceMaintenanceEventMaintenanceReasonEnvironmentalFactors,
	"decommission":          InstanceMaintenanceEventMaintenanceReasonDecommission,
	"hardware_replacement":  InstanceMaintenanceEventMaintenanceReasonHardwareReplacement,
	"firmware_update":       InstanceMaintenanceEventMaintenanceReasonFirmwareUpdate,
	"security_update":       InstanceMaintenanceEventMaintenanceReasonSecurityUpdate,
}

// GetInstanceMaintenanceEventMaintenanceReasonEnumValues Enumerates the set of values for InstanceMaintenanceEventMaintenanceReasonEnum
func GetInstanceMaintenanceEventMaintenanceReasonEnumValues() []InstanceMaintenanceEventMaintenanceReasonEnum {
	values := make([]InstanceMaintenanceEventMaintenanceReasonEnum, 0)
	for _, v := range mappingInstanceMaintenanceEventMaintenanceReasonEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceMaintenanceEventMaintenanceReasonEnumStringValues Enumerates the set of values in String for InstanceMaintenanceEventMaintenanceReasonEnum
func GetInstanceMaintenanceEventMaintenanceReasonEnumStringValues() []string {
	return []string{
		"EVACUATION",
		"ENVIRONMENTAL_FACTORS",
		"DECOMMISSION",
		"HARDWARE_REPLACEMENT",
		"FIRMWARE_UPDATE",
		"SECURITY_UPDATE",
	}
}

// GetMappingInstanceMaintenanceEventMaintenanceReasonEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceMaintenanceEventMaintenanceReasonEnum(val string) (InstanceMaintenanceEventMaintenanceReasonEnum, bool) {
	enum, ok := mappingInstanceMaintenanceEventMaintenanceReasonEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// InstanceMaintenanceEventInstanceActionEnum Enum with underlying type: string
type InstanceMaintenanceEventInstanceActionEnum string

// Set of constants representing the allowable values for InstanceMaintenanceEventInstanceActionEnum
const (
	InstanceMaintenanceEventInstanceActionRebootMigration InstanceMaintenanceEventInstanceActionEnum = "REBOOT_MIGRATION"
	InstanceMaintenanceEventInstanceActionTerminate       InstanceMaintenanceEventInstanceActionEnum = "TERMINATE"
	InstanceMaintenanceEventInstanceActionStop            InstanceMaintenanceEventInstanceActionEnum = "STOP"
	InstanceMaintenanceEventInstanceActionNone            InstanceMaintenanceEventInstanceActionEnum = "NONE"
)

var mappingInstanceMaintenanceEventInstanceActionEnum = map[string]InstanceMaintenanceEventInstanceActionEnum{
	"REBOOT_MIGRATION": InstanceMaintenanceEventInstanceActionRebootMigration,
	"TERMINATE":        InstanceMaintenanceEventInstanceActionTerminate,
	"STOP":             InstanceMaintenanceEventInstanceActionStop,
	"NONE":             InstanceMaintenanceEventInstanceActionNone,
}

var mappingInstanceMaintenanceEventInstanceActionEnumLowerCase = map[string]InstanceMaintenanceEventInstanceActionEnum{
	"reboot_migration": InstanceMaintenanceEventInstanceActionRebootMigration,
	"terminate":        InstanceMaintenanceEventInstanceActionTerminate,
	"stop":             InstanceMaintenanceEventInstanceActionStop,
	"none":             InstanceMaintenanceEventInstanceActionNone,
}

// GetInstanceMaintenanceEventInstanceActionEnumValues Enumerates the set of values for InstanceMaintenanceEventInstanceActionEnum
func GetInstanceMaintenanceEventInstanceActionEnumValues() []InstanceMaintenanceEventInstanceActionEnum {
	values := make([]InstanceMaintenanceEventInstanceActionEnum, 0)
	for _, v := range mappingInstanceMaintenanceEventInstanceActionEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceMaintenanceEventInstanceActionEnumStringValues Enumerates the set of values in String for InstanceMaintenanceEventInstanceActionEnum
func GetInstanceMaintenanceEventInstanceActionEnumStringValues() []string {
	return []string{
		"REBOOT_MIGRATION",
		"TERMINATE",
		"STOP",
		"NONE",
	}
}

// GetMappingInstanceMaintenanceEventInstanceActionEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceMaintenanceEventInstanceActionEnum(val string) (InstanceMaintenanceEventInstanceActionEnum, bool) {
	enum, ok := mappingInstanceMaintenanceEventInstanceActionEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// InstanceMaintenanceEventLifecycleStateEnum Enum with underlying type: string
type InstanceMaintenanceEventLifecycleStateEnum string

// Set of constants representing the allowable values for InstanceMaintenanceEventLifecycleStateEnum
const (
	InstanceMaintenanceEventLifecycleStateScheduled  InstanceMaintenanceEventLifecycleStateEnum = "SCHEDULED"
	InstanceMaintenanceEventLifecycleStateStarted    InstanceMaintenanceEventLifecycleStateEnum = "STARTED"
	InstanceMaintenanceEventLifecycleStateProcessing InstanceMaintenanceEventLifecycleStateEnum = "PROCESSING"
	InstanceMaintenanceEventLifecycleStateSucceeded  InstanceMaintenanceEventLifecycleStateEnum = "SUCCEEDED"
	InstanceMaintenanceEventLifecycleStateFailed     InstanceMaintenanceEventLifecycleStateEnum = "FAILED"
	InstanceMaintenanceEventLifecycleStateCanceled   InstanceMaintenanceEventLifecycleStateEnum = "CANCELED"
)

var mappingInstanceMaintenanceEventLifecycleStateEnum = map[string]InstanceMaintenanceEventLifecycleStateEnum{
	"SCHEDULED":  InstanceMaintenanceEventLifecycleStateScheduled,
	"STARTED":    InstanceMaintenanceEventLifecycleStateStarted,
	"PROCESSING": InstanceMaintenanceEventLifecycleStateProcessing,
	"SUCCEEDED":  InstanceMaintenanceEventLifecycleStateSucceeded,
	"FAILED":     InstanceMaintenanceEventLifecycleStateFailed,
	"CANCELED":   InstanceMaintenanceEventLifecycleStateCanceled,
}

var mappingInstanceMaintenanceEventLifecycleStateEnumLowerCase = map[string]InstanceMaintenanceEventLifecycleStateEnum{
	"scheduled":  InstanceMaintenanceEventLifecycleStateScheduled,
	"started":    InstanceMaintenanceEventLifecycleStateStarted,
	"processing": InstanceMaintenanceEventLifecycleStateProcessing,
	"succeeded":  InstanceMaintenanceEventLifecycleStateSucceeded,
	"failed":     InstanceMaintenanceEventLifecycleStateFailed,
	"canceled":   InstanceMaintenanceEventLifecycleStateCanceled,
}

// GetInstanceMaintenanceEventLifecycleStateEnumValues Enumerates the set of values for InstanceMaintenanceEventLifecycleStateEnum
func GetInstanceMaintenanceEventLifecycleStateEnumValues() []InstanceMaintenanceEventLifecycleStateEnum {
	values := make([]InstanceMaintenanceEventLifecycleStateEnum, 0)
	for _, v := range mappingInstanceMaintenanceEventLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceMaintenanceEventLifecycleStateEnumStringValues Enumerates the set of values in String for InstanceMaintenanceEventLifecycleStateEnum
func GetInstanceMaintenanceEventLifecycleStateEnumStringValues() []string {
	return []string{
		"SCHEDULED",
		"STARTED",
		"PROCESSING",
		"SUCCEEDED",
		"FAILED",
		"CANCELED",
	}
}

// GetMappingInstanceMaintenanceEventLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceMaintenanceEventLifecycleStateEnum(val string) (InstanceMaintenanceEventLifecycleStateEnum, bool) {
	enum, ok := mappingInstanceMaintenanceEventLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// InstanceMaintenanceEventCreatedByEnum Enum with underlying type: string
type InstanceMaintenanceEventCreatedByEnum string

// Set of constants representing the allowable values for InstanceMaintenanceEventCreatedByEnum
const (
	InstanceMaintenanceEventCreatedByCustomer InstanceMaintenanceEventCreatedByEnum = "CUSTOMER"
	InstanceMaintenanceEventCreatedBySystem   InstanceMaintenanceEventCreatedByEnum = "SYSTEM"
)

var mappingInstanceMaintenanceEventCreatedByEnum = map[string]InstanceMaintenanceEventCreatedByEnum{
	"CUSTOMER": InstanceMaintenanceEventCreatedByCustomer,
	"SYSTEM":   InstanceMaintenanceEventCreatedBySystem,
}

var mappingInstanceMaintenanceEventCreatedByEnumLowerCase = map[string]InstanceMaintenanceEventCreatedByEnum{
	"customer": InstanceMaintenanceEventCreatedByCustomer,
	"system":   InstanceMaintenanceEventCreatedBySystem,
}

// GetInstanceMaintenanceEventCreatedByEnumValues Enumerates the set of values for InstanceMaintenanceEventCreatedByEnum
func GetInstanceMaintenanceEventCreatedByEnumValues() []InstanceMaintenanceEventCreatedByEnum {
	values := make([]InstanceMaintenanceEventCreatedByEnum, 0)
	for _, v := range mappingInstanceMaintenanceEventCreatedByEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceMaintenanceEventCreatedByEnumStringValues Enumerates the set of values in String for InstanceMaintenanceEventCreatedByEnum
func GetInstanceMaintenanceEventCreatedByEnumStringValues() []string {
	return []string{
		"CUSTOMER",
		"SYSTEM",
	}
}

// GetMappingInstanceMaintenanceEventCreatedByEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingInstanceMaintenanceEventCreatedByEnum(val string) (InstanceMaintenanceEventCreatedByEnum, bool) {
	enum, ok := mappingInstanceMaintenanceEventCreatedByEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
