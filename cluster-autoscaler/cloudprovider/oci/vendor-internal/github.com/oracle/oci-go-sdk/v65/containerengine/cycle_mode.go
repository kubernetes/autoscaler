// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Kubernetes Engine API
//
// API for the Kubernetes Engine service (also known as the Container Engine for Kubernetes service). Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Kubernetes Engine (https://docs.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"strings"
)

// CycleModeEnum Enum with underlying type: string
type CycleModeEnum string

// Set of constants representing the allowable values for CycleModeEnum
const (
	CycleModeBootVolumeReplace CycleModeEnum = "BOOT_VOLUME_REPLACE"
	CycleModeInstanceReplace   CycleModeEnum = "INSTANCE_REPLACE"
)

var mappingCycleModeEnum = map[string]CycleModeEnum{
	"BOOT_VOLUME_REPLACE": CycleModeBootVolumeReplace,
	"INSTANCE_REPLACE":    CycleModeInstanceReplace,
}

var mappingCycleModeEnumLowerCase = map[string]CycleModeEnum{
	"boot_volume_replace": CycleModeBootVolumeReplace,
	"instance_replace":    CycleModeInstanceReplace,
}

// GetCycleModeEnumValues Enumerates the set of values for CycleModeEnum
func GetCycleModeEnumValues() []CycleModeEnum {
	values := make([]CycleModeEnum, 0)
	for _, v := range mappingCycleModeEnum {
		values = append(values, v)
	}
	return values
}

// GetCycleModeEnumStringValues Enumerates the set of values in String for CycleModeEnum
func GetCycleModeEnumStringValues() []string {
	return []string{
		"BOOT_VOLUME_REPLACE",
		"INSTANCE_REPLACE",
	}
}

// GetMappingCycleModeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCycleModeEnum(val string) (CycleModeEnum, bool) {
	enum, ok := mappingCycleModeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
