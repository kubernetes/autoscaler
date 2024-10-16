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
	"strings"
)

// EncryptionInTransitTypeEnum Enum with underlying type: string
type EncryptionInTransitTypeEnum string

// Set of constants representing the allowable values for EncryptionInTransitTypeEnum
const (
	EncryptionInTransitTypeNone                  EncryptionInTransitTypeEnum = "NONE"
	EncryptionInTransitTypeBmEncryptionInTransit EncryptionInTransitTypeEnum = "BM_ENCRYPTION_IN_TRANSIT"
)

var mappingEncryptionInTransitTypeEnum = map[string]EncryptionInTransitTypeEnum{
	"NONE":                     EncryptionInTransitTypeNone,
	"BM_ENCRYPTION_IN_TRANSIT": EncryptionInTransitTypeBmEncryptionInTransit,
}

var mappingEncryptionInTransitTypeEnumLowerCase = map[string]EncryptionInTransitTypeEnum{
	"none":                     EncryptionInTransitTypeNone,
	"bm_encryption_in_transit": EncryptionInTransitTypeBmEncryptionInTransit,
}

// GetEncryptionInTransitTypeEnumValues Enumerates the set of values for EncryptionInTransitTypeEnum
func GetEncryptionInTransitTypeEnumValues() []EncryptionInTransitTypeEnum {
	values := make([]EncryptionInTransitTypeEnum, 0)
	for _, v := range mappingEncryptionInTransitTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetEncryptionInTransitTypeEnumStringValues Enumerates the set of values in String for EncryptionInTransitTypeEnum
func GetEncryptionInTransitTypeEnumStringValues() []string {
	return []string{
		"NONE",
		"BM_ENCRYPTION_IN_TRANSIT",
	}
}

// GetMappingEncryptionInTransitTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingEncryptionInTransitTypeEnum(val string) (EncryptionInTransitTypeEnum, bool) {
	enum, ok := mappingEncryptionInTransitTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
