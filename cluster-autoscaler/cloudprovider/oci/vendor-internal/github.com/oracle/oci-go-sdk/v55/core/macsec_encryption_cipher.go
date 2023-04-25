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

// MacsecEncryptionCipherEnum Enum with underlying type: string
type MacsecEncryptionCipherEnum string

// Set of constants representing the allowable values for MacsecEncryptionCipherEnum
const (
	MacsecEncryptionCipherAes128Gcm    MacsecEncryptionCipherEnum = "AES128_GCM"
	MacsecEncryptionCipherAes128GcmXpn MacsecEncryptionCipherEnum = "AES128_GCM_XPN"
	MacsecEncryptionCipherAes256Gcm    MacsecEncryptionCipherEnum = "AES256_GCM"
	MacsecEncryptionCipherAes256GcmXpn MacsecEncryptionCipherEnum = "AES256_GCM_XPN"
)

var mappingMacsecEncryptionCipherEnum = map[string]MacsecEncryptionCipherEnum{
	"AES128_GCM":     MacsecEncryptionCipherAes128Gcm,
	"AES128_GCM_XPN": MacsecEncryptionCipherAes128GcmXpn,
	"AES256_GCM":     MacsecEncryptionCipherAes256Gcm,
	"AES256_GCM_XPN": MacsecEncryptionCipherAes256GcmXpn,
}

// GetMacsecEncryptionCipherEnumValues Enumerates the set of values for MacsecEncryptionCipherEnum
func GetMacsecEncryptionCipherEnumValues() []MacsecEncryptionCipherEnum {
	values := make([]MacsecEncryptionCipherEnum, 0)
	for _, v := range mappingMacsecEncryptionCipherEnum {
		values = append(values, v)
	}
	return values
}

// GetMacsecEncryptionCipherEnumStringValues Enumerates the set of values in String for MacsecEncryptionCipherEnum
func GetMacsecEncryptionCipherEnumStringValues() []string {
	return []string{
		"AES128_GCM",
		"AES128_GCM_XPN",
		"AES256_GCM",
		"AES256_GCM_XPN",
	}
}
