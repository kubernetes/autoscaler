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

	if _, ok := mappingPhaseTwoConfigDetailsAuthenticationAlgorithmEnum[string(m.AuthenticationAlgorithm)]; !ok && m.AuthenticationAlgorithm != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AuthenticationAlgorithm: %s. Supported values are: %s.", m.AuthenticationAlgorithm, strings.Join(GetPhaseTwoConfigDetailsAuthenticationAlgorithmEnumStringValues(), ",")))
	}
	if _, ok := mappingPhaseTwoConfigDetailsEncryptionAlgorithmEnum[string(m.EncryptionAlgorithm)]; !ok && m.EncryptionAlgorithm != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for EncryptionAlgorithm: %s. Supported values are: %s.", m.EncryptionAlgorithm, strings.Join(GetPhaseTwoConfigDetailsEncryptionAlgorithmEnumStringValues(), ",")))
	}
	if _, ok := mappingPhaseTwoConfigDetailsPfsDhGroupEnum[string(m.PfsDhGroup)]; !ok && m.PfsDhGroup != "" {
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
