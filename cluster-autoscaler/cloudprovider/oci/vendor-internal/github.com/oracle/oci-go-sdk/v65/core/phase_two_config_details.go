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

// PhaseTwoConfigDetails Configuration details for IPSec phase two configuration parameters.
type PhaseTwoConfigDetails struct {

	// Indicates whether custom configuration is enabled for phase two options.
	IsCustomPhaseTwoConfig *bool `mandatory:"false" json:"isCustomPhaseTwoConfig"`

	// The authentication algorithm proposed during phase two tunnel negotiation.
	AuthenticationAlgorithm PhaseTwoConfigDetailsAuthenticationAlgorithmEnum `mandatory:"false" json:"authenticationAlgorithm,omitempty"`

	// The encryption algorithm proposed during phase two tunnel negotiation.
	EncryptionAlgorithm PhaseTwoConfigDetailsEncryptionAlgorithmEnum `mandatory:"false" json:"encryptionAlgorithm,omitempty"`

	// Lifetime in seconds for the IPSec session key set in phase two. The default is 3600 which is equivalent to 1 hour.
	LifetimeInSeconds *int `mandatory:"false" json:"lifetimeInSeconds"`

	// Indicates whether perfect forward secrecy (PFS) is enabled.
	IsPfsEnabled *bool `mandatory:"false" json:"isPfsEnabled"`

	// The Diffie-Hellman group used for PFS, if PFS is enabled.
	PfsDhGroup PhaseTwoConfigDetailsPfsDhGroupEnum `mandatory:"false" json:"pfsDhGroup,omitempty"`
}

func (m PhaseTwoConfigDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m PhaseTwoConfigDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingPhaseTwoConfigDetailsAuthenticationAlgorithmEnum(string(m.AuthenticationAlgorithm)); !ok && m.AuthenticationAlgorithm != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AuthenticationAlgorithm: %s. Supported values are: %s.", m.AuthenticationAlgorithm, strings.Join(GetPhaseTwoConfigDetailsAuthenticationAlgorithmEnumStringValues(), ",")))
	}
	if _, ok := GetMappingPhaseTwoConfigDetailsEncryptionAlgorithmEnum(string(m.EncryptionAlgorithm)); !ok && m.EncryptionAlgorithm != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for EncryptionAlgorithm: %s. Supported values are: %s.", m.EncryptionAlgorithm, strings.Join(GetPhaseTwoConfigDetailsEncryptionAlgorithmEnumStringValues(), ",")))
	}
	if _, ok := GetMappingPhaseTwoConfigDetailsPfsDhGroupEnum(string(m.PfsDhGroup)); !ok && m.PfsDhGroup != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for PfsDhGroup: %s. Supported values are: %s.", m.PfsDhGroup, strings.Join(GetPhaseTwoConfigDetailsPfsDhGroupEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// PhaseTwoConfigDetailsAuthenticationAlgorithmEnum Enum with underlying type: string
type PhaseTwoConfigDetailsAuthenticationAlgorithmEnum string

// Set of constants representing the allowable values for PhaseTwoConfigDetailsAuthenticationAlgorithmEnum
const (
	PhaseTwoConfigDetailsAuthenticationAlgorithmSha2256128 PhaseTwoConfigDetailsAuthenticationAlgorithmEnum = "HMAC_SHA2_256_128"
	PhaseTwoConfigDetailsAuthenticationAlgorithmSha1128    PhaseTwoConfigDetailsAuthenticationAlgorithmEnum = "HMAC_SHA1_128"
)

var mappingPhaseTwoConfigDetailsAuthenticationAlgorithmEnum = map[string]PhaseTwoConfigDetailsAuthenticationAlgorithmEnum{
	"HMAC_SHA2_256_128": PhaseTwoConfigDetailsAuthenticationAlgorithmSha2256128,
	"HMAC_SHA1_128":     PhaseTwoConfigDetailsAuthenticationAlgorithmSha1128,
}

var mappingPhaseTwoConfigDetailsAuthenticationAlgorithmEnumLowerCase = map[string]PhaseTwoConfigDetailsAuthenticationAlgorithmEnum{
	"hmac_sha2_256_128": PhaseTwoConfigDetailsAuthenticationAlgorithmSha2256128,
	"hmac_sha1_128":     PhaseTwoConfigDetailsAuthenticationAlgorithmSha1128,
}

// GetPhaseTwoConfigDetailsAuthenticationAlgorithmEnumValues Enumerates the set of values for PhaseTwoConfigDetailsAuthenticationAlgorithmEnum
func GetPhaseTwoConfigDetailsAuthenticationAlgorithmEnumValues() []PhaseTwoConfigDetailsAuthenticationAlgorithmEnum {
	values := make([]PhaseTwoConfigDetailsAuthenticationAlgorithmEnum, 0)
	for _, v := range mappingPhaseTwoConfigDetailsAuthenticationAlgorithmEnum {
		values = append(values, v)
	}
	return values
}

// GetPhaseTwoConfigDetailsAuthenticationAlgorithmEnumStringValues Enumerates the set of values in String for PhaseTwoConfigDetailsAuthenticationAlgorithmEnum
func GetPhaseTwoConfigDetailsAuthenticationAlgorithmEnumStringValues() []string {
	return []string{
		"HMAC_SHA2_256_128",
		"HMAC_SHA1_128",
	}
}

// GetMappingPhaseTwoConfigDetailsAuthenticationAlgorithmEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingPhaseTwoConfigDetailsAuthenticationAlgorithmEnum(val string) (PhaseTwoConfigDetailsAuthenticationAlgorithmEnum, bool) {
	enum, ok := mappingPhaseTwoConfigDetailsAuthenticationAlgorithmEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// PhaseTwoConfigDetailsEncryptionAlgorithmEnum Enum with underlying type: string
type PhaseTwoConfigDetailsEncryptionAlgorithmEnum string

// Set of constants representing the allowable values for PhaseTwoConfigDetailsEncryptionAlgorithmEnum
const (
	PhaseTwoConfigDetailsEncryptionAlgorithm256Gcm PhaseTwoConfigDetailsEncryptionAlgorithmEnum = "AES_256_GCM"
	PhaseTwoConfigDetailsEncryptionAlgorithm192Gcm PhaseTwoConfigDetailsEncryptionAlgorithmEnum = "AES_192_GCM"
	PhaseTwoConfigDetailsEncryptionAlgorithm128Gcm PhaseTwoConfigDetailsEncryptionAlgorithmEnum = "AES_128_GCM"
	PhaseTwoConfigDetailsEncryptionAlgorithm256Cbc PhaseTwoConfigDetailsEncryptionAlgorithmEnum = "AES_256_CBC"
	PhaseTwoConfigDetailsEncryptionAlgorithm192Cbc PhaseTwoConfigDetailsEncryptionAlgorithmEnum = "AES_192_CBC"
	PhaseTwoConfigDetailsEncryptionAlgorithm128Cbc PhaseTwoConfigDetailsEncryptionAlgorithmEnum = "AES_128_CBC"
)

var mappingPhaseTwoConfigDetailsEncryptionAlgorithmEnum = map[string]PhaseTwoConfigDetailsEncryptionAlgorithmEnum{
	"AES_256_GCM": PhaseTwoConfigDetailsEncryptionAlgorithm256Gcm,
	"AES_192_GCM": PhaseTwoConfigDetailsEncryptionAlgorithm192Gcm,
	"AES_128_GCM": PhaseTwoConfigDetailsEncryptionAlgorithm128Gcm,
	"AES_256_CBC": PhaseTwoConfigDetailsEncryptionAlgorithm256Cbc,
	"AES_192_CBC": PhaseTwoConfigDetailsEncryptionAlgorithm192Cbc,
	"AES_128_CBC": PhaseTwoConfigDetailsEncryptionAlgorithm128Cbc,
}

var mappingPhaseTwoConfigDetailsEncryptionAlgorithmEnumLowerCase = map[string]PhaseTwoConfigDetailsEncryptionAlgorithmEnum{
	"aes_256_gcm": PhaseTwoConfigDetailsEncryptionAlgorithm256Gcm,
	"aes_192_gcm": PhaseTwoConfigDetailsEncryptionAlgorithm192Gcm,
	"aes_128_gcm": PhaseTwoConfigDetailsEncryptionAlgorithm128Gcm,
	"aes_256_cbc": PhaseTwoConfigDetailsEncryptionAlgorithm256Cbc,
	"aes_192_cbc": PhaseTwoConfigDetailsEncryptionAlgorithm192Cbc,
	"aes_128_cbc": PhaseTwoConfigDetailsEncryptionAlgorithm128Cbc,
}

// GetPhaseTwoConfigDetailsEncryptionAlgorithmEnumValues Enumerates the set of values for PhaseTwoConfigDetailsEncryptionAlgorithmEnum
func GetPhaseTwoConfigDetailsEncryptionAlgorithmEnumValues() []PhaseTwoConfigDetailsEncryptionAlgorithmEnum {
	values := make([]PhaseTwoConfigDetailsEncryptionAlgorithmEnum, 0)
	for _, v := range mappingPhaseTwoConfigDetailsEncryptionAlgorithmEnum {
		values = append(values, v)
	}
	return values
}

// GetPhaseTwoConfigDetailsEncryptionAlgorithmEnumStringValues Enumerates the set of values in String for PhaseTwoConfigDetailsEncryptionAlgorithmEnum
func GetPhaseTwoConfigDetailsEncryptionAlgorithmEnumStringValues() []string {
	return []string{
		"AES_256_GCM",
		"AES_192_GCM",
		"AES_128_GCM",
		"AES_256_CBC",
		"AES_192_CBC",
		"AES_128_CBC",
	}
}

// GetMappingPhaseTwoConfigDetailsEncryptionAlgorithmEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingPhaseTwoConfigDetailsEncryptionAlgorithmEnum(val string) (PhaseTwoConfigDetailsEncryptionAlgorithmEnum, bool) {
	enum, ok := mappingPhaseTwoConfigDetailsEncryptionAlgorithmEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// PhaseTwoConfigDetailsPfsDhGroupEnum Enum with underlying type: string
type PhaseTwoConfigDetailsPfsDhGroupEnum string

// Set of constants representing the allowable values for PhaseTwoConfigDetailsPfsDhGroupEnum
const (
	PhaseTwoConfigDetailsPfsDhGroupGroup2  PhaseTwoConfigDetailsPfsDhGroupEnum = "GROUP2"
	PhaseTwoConfigDetailsPfsDhGroupGroup5  PhaseTwoConfigDetailsPfsDhGroupEnum = "GROUP5"
	PhaseTwoConfigDetailsPfsDhGroupGroup14 PhaseTwoConfigDetailsPfsDhGroupEnum = "GROUP14"
	PhaseTwoConfigDetailsPfsDhGroupGroup19 PhaseTwoConfigDetailsPfsDhGroupEnum = "GROUP19"
	PhaseTwoConfigDetailsPfsDhGroupGroup20 PhaseTwoConfigDetailsPfsDhGroupEnum = "GROUP20"
	PhaseTwoConfigDetailsPfsDhGroupGroup24 PhaseTwoConfigDetailsPfsDhGroupEnum = "GROUP24"
)

var mappingPhaseTwoConfigDetailsPfsDhGroupEnum = map[string]PhaseTwoConfigDetailsPfsDhGroupEnum{
	"GROUP2":  PhaseTwoConfigDetailsPfsDhGroupGroup2,
	"GROUP5":  PhaseTwoConfigDetailsPfsDhGroupGroup5,
	"GROUP14": PhaseTwoConfigDetailsPfsDhGroupGroup14,
	"GROUP19": PhaseTwoConfigDetailsPfsDhGroupGroup19,
	"GROUP20": PhaseTwoConfigDetailsPfsDhGroupGroup20,
	"GROUP24": PhaseTwoConfigDetailsPfsDhGroupGroup24,
}

var mappingPhaseTwoConfigDetailsPfsDhGroupEnumLowerCase = map[string]PhaseTwoConfigDetailsPfsDhGroupEnum{
	"group2":  PhaseTwoConfigDetailsPfsDhGroupGroup2,
	"group5":  PhaseTwoConfigDetailsPfsDhGroupGroup5,
	"group14": PhaseTwoConfigDetailsPfsDhGroupGroup14,
	"group19": PhaseTwoConfigDetailsPfsDhGroupGroup19,
	"group20": PhaseTwoConfigDetailsPfsDhGroupGroup20,
	"group24": PhaseTwoConfigDetailsPfsDhGroupGroup24,
}

// GetPhaseTwoConfigDetailsPfsDhGroupEnumValues Enumerates the set of values for PhaseTwoConfigDetailsPfsDhGroupEnum
func GetPhaseTwoConfigDetailsPfsDhGroupEnumValues() []PhaseTwoConfigDetailsPfsDhGroupEnum {
	values := make([]PhaseTwoConfigDetailsPfsDhGroupEnum, 0)
	for _, v := range mappingPhaseTwoConfigDetailsPfsDhGroupEnum {
		values = append(values, v)
	}
	return values
}

// GetPhaseTwoConfigDetailsPfsDhGroupEnumStringValues Enumerates the set of values in String for PhaseTwoConfigDetailsPfsDhGroupEnum
func GetPhaseTwoConfigDetailsPfsDhGroupEnumStringValues() []string {
	return []string{
		"GROUP2",
		"GROUP5",
		"GROUP14",
		"GROUP19",
		"GROUP20",
		"GROUP24",
	}
}

// GetMappingPhaseTwoConfigDetailsPfsDhGroupEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingPhaseTwoConfigDetailsPfsDhGroupEnum(val string) (PhaseTwoConfigDetailsPfsDhGroupEnum, bool) {
	enum, ok := mappingPhaseTwoConfigDetailsPfsDhGroupEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
