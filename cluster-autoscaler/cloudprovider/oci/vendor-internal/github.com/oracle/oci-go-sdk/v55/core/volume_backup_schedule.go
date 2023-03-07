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

// VolumeBackupSchedule Defines the backup frequency and retention period for a volume backup policy. For more information,
// see Policy-Based Backups (https://docs.cloud.oracle.com/iaas/Content/Block/Tasks/schedulingvolumebackups.htm).
type VolumeBackupSchedule struct {

	// The type of volume backup to create.
	BackupType VolumeBackupScheduleBackupTypeEnum `mandatory:"true" json:"backupType"`

	// The volume backup frequency.
	Period VolumeBackupSchedulePeriodEnum `mandatory:"true" json:"period"`

	// How long, in seconds, to keep the volume backups created by this schedule.
	RetentionSeconds *int `mandatory:"true" json:"retentionSeconds"`

	// The number of seconds that the volume backup start
	// time should be shifted from the default interval boundaries specified by
	// the period. The volume backup start time is the frequency start time plus the offset.
	OffsetSeconds *int `mandatory:"false" json:"offsetSeconds"`

	// Indicates how the offset is defined. If value is `STRUCTURED`,
	// then `hourOfDay`, `dayOfWeek`, `dayOfMonth`, and `month` fields are used
	// and `offsetSeconds` will be ignored in requests and users should ignore its
	// value from the responses.
	// `hourOfDay` is applicable for periods `ONE_DAY`,
	// `ONE_WEEK`, `ONE_MONTH` and `ONE_YEAR`.
	// `dayOfWeek` is applicable for period
	// `ONE_WEEK`.
	// `dayOfMonth` is applicable for periods `ONE_MONTH` and `ONE_YEAR`.
	// 'month' is applicable for period 'ONE_YEAR'.
	// They will be ignored in the requests for inapplicable periods.
	// If value is `NUMERIC_SECONDS`, then `offsetSeconds`
	// will be used for both requests and responses and the structured fields will be
	// ignored in the requests and users should ignore their values from the responses.
	// For clients using older versions of Apis and not sending `offsetType` in their
	// requests, the behaviour is just like `NUMERIC_SECONDS`.
	OffsetType VolumeBackupScheduleOffsetTypeEnum `mandatory:"false" json:"offsetType,omitempty"`

	// The hour of the day to schedule the volume backup.
	HourOfDay *int `mandatory:"false" json:"hourOfDay"`

	// The day of the week to schedule the volume backup.
	DayOfWeek VolumeBackupScheduleDayOfWeekEnum `mandatory:"false" json:"dayOfWeek,omitempty"`

	// The day of the month to schedule the volume backup.
	DayOfMonth *int `mandatory:"false" json:"dayOfMonth"`

	// The month of the year to schedule the volume backup.
	Month VolumeBackupScheduleMonthEnum `mandatory:"false" json:"month,omitempty"`

	// Specifies what time zone is the schedule in
	TimeZone VolumeBackupScheduleTimeZoneEnum `mandatory:"false" json:"timeZone,omitempty"`
}

func (m VolumeBackupSchedule) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VolumeBackupSchedule) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingVolumeBackupScheduleBackupTypeEnum[string(m.BackupType)]; !ok && m.BackupType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BackupType: %s. Supported values are: %s.", m.BackupType, strings.Join(GetVolumeBackupScheduleBackupTypeEnumStringValues(), ",")))
	}
	if _, ok := mappingVolumeBackupSchedulePeriodEnum[string(m.Period)]; !ok && m.Period != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Period: %s. Supported values are: %s.", m.Period, strings.Join(GetVolumeBackupSchedulePeriodEnumStringValues(), ",")))
	}

	if _, ok := mappingVolumeBackupScheduleOffsetTypeEnum[string(m.OffsetType)]; !ok && m.OffsetType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for OffsetType: %s. Supported values are: %s.", m.OffsetType, strings.Join(GetVolumeBackupScheduleOffsetTypeEnumStringValues(), ",")))
	}
	if _, ok := mappingVolumeBackupScheduleDayOfWeekEnum[string(m.DayOfWeek)]; !ok && m.DayOfWeek != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DayOfWeek: %s. Supported values are: %s.", m.DayOfWeek, strings.Join(GetVolumeBackupScheduleDayOfWeekEnumStringValues(), ",")))
	}
	if _, ok := mappingVolumeBackupScheduleMonthEnum[string(m.Month)]; !ok && m.Month != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Month: %s. Supported values are: %s.", m.Month, strings.Join(GetVolumeBackupScheduleMonthEnumStringValues(), ",")))
	}
	if _, ok := mappingVolumeBackupScheduleTimeZoneEnum[string(m.TimeZone)]; !ok && m.TimeZone != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for TimeZone: %s. Supported values are: %s.", m.TimeZone, strings.Join(GetVolumeBackupScheduleTimeZoneEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// VolumeBackupScheduleBackupTypeEnum Enum with underlying type: string
type VolumeBackupScheduleBackupTypeEnum string

// Set of constants representing the allowable values for VolumeBackupScheduleBackupTypeEnum
const (
	VolumeBackupScheduleBackupTypeFull        VolumeBackupScheduleBackupTypeEnum = "FULL"
	VolumeBackupScheduleBackupTypeIncremental VolumeBackupScheduleBackupTypeEnum = "INCREMENTAL"
)

var mappingVolumeBackupScheduleBackupTypeEnum = map[string]VolumeBackupScheduleBackupTypeEnum{
	"FULL":        VolumeBackupScheduleBackupTypeFull,
	"INCREMENTAL": VolumeBackupScheduleBackupTypeIncremental,
}

// GetVolumeBackupScheduleBackupTypeEnumValues Enumerates the set of values for VolumeBackupScheduleBackupTypeEnum
func GetVolumeBackupScheduleBackupTypeEnumValues() []VolumeBackupScheduleBackupTypeEnum {
	values := make([]VolumeBackupScheduleBackupTypeEnum, 0)
	for _, v := range mappingVolumeBackupScheduleBackupTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetVolumeBackupScheduleBackupTypeEnumStringValues Enumerates the set of values in String for VolumeBackupScheduleBackupTypeEnum
func GetVolumeBackupScheduleBackupTypeEnumStringValues() []string {
	return []string{
		"FULL",
		"INCREMENTAL",
	}
}

// VolumeBackupSchedulePeriodEnum Enum with underlying type: string
type VolumeBackupSchedulePeriodEnum string

// Set of constants representing the allowable values for VolumeBackupSchedulePeriodEnum
const (
	VolumeBackupSchedulePeriodHour  VolumeBackupSchedulePeriodEnum = "ONE_HOUR"
	VolumeBackupSchedulePeriodDay   VolumeBackupSchedulePeriodEnum = "ONE_DAY"
	VolumeBackupSchedulePeriodWeek  VolumeBackupSchedulePeriodEnum = "ONE_WEEK"
	VolumeBackupSchedulePeriodMonth VolumeBackupSchedulePeriodEnum = "ONE_MONTH"
	VolumeBackupSchedulePeriodYear  VolumeBackupSchedulePeriodEnum = "ONE_YEAR"
)

var mappingVolumeBackupSchedulePeriodEnum = map[string]VolumeBackupSchedulePeriodEnum{
	"ONE_HOUR":  VolumeBackupSchedulePeriodHour,
	"ONE_DAY":   VolumeBackupSchedulePeriodDay,
	"ONE_WEEK":  VolumeBackupSchedulePeriodWeek,
	"ONE_MONTH": VolumeBackupSchedulePeriodMonth,
	"ONE_YEAR":  VolumeBackupSchedulePeriodYear,
}

// GetVolumeBackupSchedulePeriodEnumValues Enumerates the set of values for VolumeBackupSchedulePeriodEnum
func GetVolumeBackupSchedulePeriodEnumValues() []VolumeBackupSchedulePeriodEnum {
	values := make([]VolumeBackupSchedulePeriodEnum, 0)
	for _, v := range mappingVolumeBackupSchedulePeriodEnum {
		values = append(values, v)
	}
	return values
}

// GetVolumeBackupSchedulePeriodEnumStringValues Enumerates the set of values in String for VolumeBackupSchedulePeriodEnum
func GetVolumeBackupSchedulePeriodEnumStringValues() []string {
	return []string{
		"ONE_HOUR",
		"ONE_DAY",
		"ONE_WEEK",
		"ONE_MONTH",
		"ONE_YEAR",
	}
}

// VolumeBackupScheduleOffsetTypeEnum Enum with underlying type: string
type VolumeBackupScheduleOffsetTypeEnum string

// Set of constants representing the allowable values for VolumeBackupScheduleOffsetTypeEnum
const (
	VolumeBackupScheduleOffsetTypeStructured     VolumeBackupScheduleOffsetTypeEnum = "STRUCTURED"
	VolumeBackupScheduleOffsetTypeNumericSeconds VolumeBackupScheduleOffsetTypeEnum = "NUMERIC_SECONDS"
)

var mappingVolumeBackupScheduleOffsetTypeEnum = map[string]VolumeBackupScheduleOffsetTypeEnum{
	"STRUCTURED":      VolumeBackupScheduleOffsetTypeStructured,
	"NUMERIC_SECONDS": VolumeBackupScheduleOffsetTypeNumericSeconds,
}

// GetVolumeBackupScheduleOffsetTypeEnumValues Enumerates the set of values for VolumeBackupScheduleOffsetTypeEnum
func GetVolumeBackupScheduleOffsetTypeEnumValues() []VolumeBackupScheduleOffsetTypeEnum {
	values := make([]VolumeBackupScheduleOffsetTypeEnum, 0)
	for _, v := range mappingVolumeBackupScheduleOffsetTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetVolumeBackupScheduleOffsetTypeEnumStringValues Enumerates the set of values in String for VolumeBackupScheduleOffsetTypeEnum
func GetVolumeBackupScheduleOffsetTypeEnumStringValues() []string {
	return []string{
		"STRUCTURED",
		"NUMERIC_SECONDS",
	}
}

// VolumeBackupScheduleDayOfWeekEnum Enum with underlying type: string
type VolumeBackupScheduleDayOfWeekEnum string

// Set of constants representing the allowable values for VolumeBackupScheduleDayOfWeekEnum
const (
	VolumeBackupScheduleDayOfWeekMonday    VolumeBackupScheduleDayOfWeekEnum = "MONDAY"
	VolumeBackupScheduleDayOfWeekTuesday   VolumeBackupScheduleDayOfWeekEnum = "TUESDAY"
	VolumeBackupScheduleDayOfWeekWednesday VolumeBackupScheduleDayOfWeekEnum = "WEDNESDAY"
	VolumeBackupScheduleDayOfWeekThursday  VolumeBackupScheduleDayOfWeekEnum = "THURSDAY"
	VolumeBackupScheduleDayOfWeekFriday    VolumeBackupScheduleDayOfWeekEnum = "FRIDAY"
	VolumeBackupScheduleDayOfWeekSaturday  VolumeBackupScheduleDayOfWeekEnum = "SATURDAY"
	VolumeBackupScheduleDayOfWeekSunday    VolumeBackupScheduleDayOfWeekEnum = "SUNDAY"
)

var mappingVolumeBackupScheduleDayOfWeekEnum = map[string]VolumeBackupScheduleDayOfWeekEnum{
	"MONDAY":    VolumeBackupScheduleDayOfWeekMonday,
	"TUESDAY":   VolumeBackupScheduleDayOfWeekTuesday,
	"WEDNESDAY": VolumeBackupScheduleDayOfWeekWednesday,
	"THURSDAY":  VolumeBackupScheduleDayOfWeekThursday,
	"FRIDAY":    VolumeBackupScheduleDayOfWeekFriday,
	"SATURDAY":  VolumeBackupScheduleDayOfWeekSaturday,
	"SUNDAY":    VolumeBackupScheduleDayOfWeekSunday,
}

// GetVolumeBackupScheduleDayOfWeekEnumValues Enumerates the set of values for VolumeBackupScheduleDayOfWeekEnum
func GetVolumeBackupScheduleDayOfWeekEnumValues() []VolumeBackupScheduleDayOfWeekEnum {
	values := make([]VolumeBackupScheduleDayOfWeekEnum, 0)
	for _, v := range mappingVolumeBackupScheduleDayOfWeekEnum {
		values = append(values, v)
	}
	return values
}

// GetVolumeBackupScheduleDayOfWeekEnumStringValues Enumerates the set of values in String for VolumeBackupScheduleDayOfWeekEnum
func GetVolumeBackupScheduleDayOfWeekEnumStringValues() []string {
	return []string{
		"MONDAY",
		"TUESDAY",
		"WEDNESDAY",
		"THURSDAY",
		"FRIDAY",
		"SATURDAY",
		"SUNDAY",
	}
}

// VolumeBackupScheduleMonthEnum Enum with underlying type: string
type VolumeBackupScheduleMonthEnum string

// Set of constants representing the allowable values for VolumeBackupScheduleMonthEnum
const (
	VolumeBackupScheduleMonthJanuary   VolumeBackupScheduleMonthEnum = "JANUARY"
	VolumeBackupScheduleMonthFebruary  VolumeBackupScheduleMonthEnum = "FEBRUARY"
	VolumeBackupScheduleMonthMarch     VolumeBackupScheduleMonthEnum = "MARCH"
	VolumeBackupScheduleMonthApril     VolumeBackupScheduleMonthEnum = "APRIL"
	VolumeBackupScheduleMonthMay       VolumeBackupScheduleMonthEnum = "MAY"
	VolumeBackupScheduleMonthJune      VolumeBackupScheduleMonthEnum = "JUNE"
	VolumeBackupScheduleMonthJuly      VolumeBackupScheduleMonthEnum = "JULY"
	VolumeBackupScheduleMonthAugust    VolumeBackupScheduleMonthEnum = "AUGUST"
	VolumeBackupScheduleMonthSeptember VolumeBackupScheduleMonthEnum = "SEPTEMBER"
	VolumeBackupScheduleMonthOctober   VolumeBackupScheduleMonthEnum = "OCTOBER"
	VolumeBackupScheduleMonthNovember  VolumeBackupScheduleMonthEnum = "NOVEMBER"
	VolumeBackupScheduleMonthDecember  VolumeBackupScheduleMonthEnum = "DECEMBER"
)

var mappingVolumeBackupScheduleMonthEnum = map[string]VolumeBackupScheduleMonthEnum{
	"JANUARY":   VolumeBackupScheduleMonthJanuary,
	"FEBRUARY":  VolumeBackupScheduleMonthFebruary,
	"MARCH":     VolumeBackupScheduleMonthMarch,
	"APRIL":     VolumeBackupScheduleMonthApril,
	"MAY":       VolumeBackupScheduleMonthMay,
	"JUNE":      VolumeBackupScheduleMonthJune,
	"JULY":      VolumeBackupScheduleMonthJuly,
	"AUGUST":    VolumeBackupScheduleMonthAugust,
	"SEPTEMBER": VolumeBackupScheduleMonthSeptember,
	"OCTOBER":   VolumeBackupScheduleMonthOctober,
	"NOVEMBER":  VolumeBackupScheduleMonthNovember,
	"DECEMBER":  VolumeBackupScheduleMonthDecember,
}

// GetVolumeBackupScheduleMonthEnumValues Enumerates the set of values for VolumeBackupScheduleMonthEnum
func GetVolumeBackupScheduleMonthEnumValues() []VolumeBackupScheduleMonthEnum {
	values := make([]VolumeBackupScheduleMonthEnum, 0)
	for _, v := range mappingVolumeBackupScheduleMonthEnum {
		values = append(values, v)
	}
	return values
}

// GetVolumeBackupScheduleMonthEnumStringValues Enumerates the set of values in String for VolumeBackupScheduleMonthEnum
func GetVolumeBackupScheduleMonthEnumStringValues() []string {
	return []string{
		"JANUARY",
		"FEBRUARY",
		"MARCH",
		"APRIL",
		"MAY",
		"JUNE",
		"JULY",
		"AUGUST",
		"SEPTEMBER",
		"OCTOBER",
		"NOVEMBER",
		"DECEMBER",
	}
}

// VolumeBackupScheduleTimeZoneEnum Enum with underlying type: string
type VolumeBackupScheduleTimeZoneEnum string

// Set of constants representing the allowable values for VolumeBackupScheduleTimeZoneEnum
const (
	VolumeBackupScheduleTimeZoneUtc                    VolumeBackupScheduleTimeZoneEnum = "UTC"
	VolumeBackupScheduleTimeZoneRegionalDataCenterTime VolumeBackupScheduleTimeZoneEnum = "REGIONAL_DATA_CENTER_TIME"
)

var mappingVolumeBackupScheduleTimeZoneEnum = map[string]VolumeBackupScheduleTimeZoneEnum{
	"UTC":                       VolumeBackupScheduleTimeZoneUtc,
	"REGIONAL_DATA_CENTER_TIME": VolumeBackupScheduleTimeZoneRegionalDataCenterTime,
}

// GetVolumeBackupScheduleTimeZoneEnumValues Enumerates the set of values for VolumeBackupScheduleTimeZoneEnum
func GetVolumeBackupScheduleTimeZoneEnumValues() []VolumeBackupScheduleTimeZoneEnum {
	values := make([]VolumeBackupScheduleTimeZoneEnum, 0)
	for _, v := range mappingVolumeBackupScheduleTimeZoneEnum {
		values = append(values, v)
	}
	return values
}

// GetVolumeBackupScheduleTimeZoneEnumStringValues Enumerates the set of values in String for VolumeBackupScheduleTimeZoneEnum
func GetVolumeBackupScheduleTimeZoneEnumStringValues() []string {
	return []string{
		"UTC",
		"REGIONAL_DATA_CENTER_TIME",
	}
}
