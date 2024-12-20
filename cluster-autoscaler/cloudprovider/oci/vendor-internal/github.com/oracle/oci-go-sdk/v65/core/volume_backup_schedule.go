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
	if _, ok := GetMappingVolumeBackupScheduleBackupTypeEnum(string(m.BackupType)); !ok && m.BackupType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for BackupType: %s. Supported values are: %s.", m.BackupType, strings.Join(GetVolumeBackupScheduleBackupTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVolumeBackupSchedulePeriodEnum(string(m.Period)); !ok && m.Period != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Period: %s. Supported values are: %s.", m.Period, strings.Join(GetVolumeBackupSchedulePeriodEnumStringValues(), ",")))
	}

	if _, ok := GetMappingVolumeBackupScheduleOffsetTypeEnum(string(m.OffsetType)); !ok && m.OffsetType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for OffsetType: %s. Supported values are: %s.", m.OffsetType, strings.Join(GetVolumeBackupScheduleOffsetTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVolumeBackupScheduleDayOfWeekEnum(string(m.DayOfWeek)); !ok && m.DayOfWeek != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DayOfWeek: %s. Supported values are: %s.", m.DayOfWeek, strings.Join(GetVolumeBackupScheduleDayOfWeekEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVolumeBackupScheduleMonthEnum(string(m.Month)); !ok && m.Month != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Month: %s. Supported values are: %s.", m.Month, strings.Join(GetVolumeBackupScheduleMonthEnumStringValues(), ",")))
	}
	if _, ok := GetMappingVolumeBackupScheduleTimeZoneEnum(string(m.TimeZone)); !ok && m.TimeZone != "" {
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

var mappingVolumeBackupScheduleBackupTypeEnumLowerCase = map[string]VolumeBackupScheduleBackupTypeEnum{
	"full":        VolumeBackupScheduleBackupTypeFull,
	"incremental": VolumeBackupScheduleBackupTypeIncremental,
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

// GetMappingVolumeBackupScheduleBackupTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVolumeBackupScheduleBackupTypeEnum(val string) (VolumeBackupScheduleBackupTypeEnum, bool) {
	enum, ok := mappingVolumeBackupScheduleBackupTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
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

var mappingVolumeBackupSchedulePeriodEnumLowerCase = map[string]VolumeBackupSchedulePeriodEnum{
	"one_hour":  VolumeBackupSchedulePeriodHour,
	"one_day":   VolumeBackupSchedulePeriodDay,
	"one_week":  VolumeBackupSchedulePeriodWeek,
	"one_month": VolumeBackupSchedulePeriodMonth,
	"one_year":  VolumeBackupSchedulePeriodYear,
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

// GetMappingVolumeBackupSchedulePeriodEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVolumeBackupSchedulePeriodEnum(val string) (VolumeBackupSchedulePeriodEnum, bool) {
	enum, ok := mappingVolumeBackupSchedulePeriodEnumLowerCase[strings.ToLower(val)]
	return enum, ok
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

var mappingVolumeBackupScheduleOffsetTypeEnumLowerCase = map[string]VolumeBackupScheduleOffsetTypeEnum{
	"structured":      VolumeBackupScheduleOffsetTypeStructured,
	"numeric_seconds": VolumeBackupScheduleOffsetTypeNumericSeconds,
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

// GetMappingVolumeBackupScheduleOffsetTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVolumeBackupScheduleOffsetTypeEnum(val string) (VolumeBackupScheduleOffsetTypeEnum, bool) {
	enum, ok := mappingVolumeBackupScheduleOffsetTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
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

var mappingVolumeBackupScheduleDayOfWeekEnumLowerCase = map[string]VolumeBackupScheduleDayOfWeekEnum{
	"monday":    VolumeBackupScheduleDayOfWeekMonday,
	"tuesday":   VolumeBackupScheduleDayOfWeekTuesday,
	"wednesday": VolumeBackupScheduleDayOfWeekWednesday,
	"thursday":  VolumeBackupScheduleDayOfWeekThursday,
	"friday":    VolumeBackupScheduleDayOfWeekFriday,
	"saturday":  VolumeBackupScheduleDayOfWeekSaturday,
	"sunday":    VolumeBackupScheduleDayOfWeekSunday,
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

// GetMappingVolumeBackupScheduleDayOfWeekEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVolumeBackupScheduleDayOfWeekEnum(val string) (VolumeBackupScheduleDayOfWeekEnum, bool) {
	enum, ok := mappingVolumeBackupScheduleDayOfWeekEnumLowerCase[strings.ToLower(val)]
	return enum, ok
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

var mappingVolumeBackupScheduleMonthEnumLowerCase = map[string]VolumeBackupScheduleMonthEnum{
	"january":   VolumeBackupScheduleMonthJanuary,
	"february":  VolumeBackupScheduleMonthFebruary,
	"march":     VolumeBackupScheduleMonthMarch,
	"april":     VolumeBackupScheduleMonthApril,
	"may":       VolumeBackupScheduleMonthMay,
	"june":      VolumeBackupScheduleMonthJune,
	"july":      VolumeBackupScheduleMonthJuly,
	"august":    VolumeBackupScheduleMonthAugust,
	"september": VolumeBackupScheduleMonthSeptember,
	"october":   VolumeBackupScheduleMonthOctober,
	"november":  VolumeBackupScheduleMonthNovember,
	"december":  VolumeBackupScheduleMonthDecember,
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

// GetMappingVolumeBackupScheduleMonthEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVolumeBackupScheduleMonthEnum(val string) (VolumeBackupScheduleMonthEnum, bool) {
	enum, ok := mappingVolumeBackupScheduleMonthEnumLowerCase[strings.ToLower(val)]
	return enum, ok
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

var mappingVolumeBackupScheduleTimeZoneEnumLowerCase = map[string]VolumeBackupScheduleTimeZoneEnum{
	"utc":                       VolumeBackupScheduleTimeZoneUtc,
	"regional_data_center_time": VolumeBackupScheduleTimeZoneRegionalDataCenterTime,
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

// GetMappingVolumeBackupScheduleTimeZoneEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVolumeBackupScheduleTimeZoneEnum(val string) (VolumeBackupScheduleTimeZoneEnum, bool) {
	enum, ok := mappingVolumeBackupScheduleTimeZoneEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
