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

// LetterOfAuthority The Letter of Authority for the cross-connect. You must submit this letter when
// requesting cabling for the cross-connect at the FastConnect location.
type LetterOfAuthority struct {

	// The name of the entity authorized by this Letter of Authority.
	AuthorizedEntityName *string `mandatory:"false" json:"authorizedEntityName"`

	// The type of cross-connect fiber, termination, and optical specification.
	CircuitType LetterOfAuthorityCircuitTypeEnum `mandatory:"false" json:"circuitType,omitempty"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the cross-connect.
	CrossConnectId *string `mandatory:"false" json:"crossConnectId"`

	// The address of the FastConnect location.
	FacilityLocation *string `mandatory:"false" json:"facilityLocation"`

	// The meet-me room port for this cross-connect.
	PortName *string `mandatory:"false" json:"portName"`

	// The date and time when the Letter of Authority expires, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeExpires *common.SDKTime `mandatory:"false" json:"timeExpires"`

	// The date and time the Letter of Authority was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeIssued *common.SDKTime `mandatory:"false" json:"timeIssued"`
}

func (m LetterOfAuthority) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m LetterOfAuthority) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingLetterOfAuthorityCircuitTypeEnum(string(m.CircuitType)); !ok && m.CircuitType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for CircuitType: %s. Supported values are: %s.", m.CircuitType, strings.Join(GetLetterOfAuthorityCircuitTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// LetterOfAuthorityCircuitTypeEnum Enum with underlying type: string
type LetterOfAuthorityCircuitTypeEnum string

// Set of constants representing the allowable values for LetterOfAuthorityCircuitTypeEnum
const (
	LetterOfAuthorityCircuitTypeLc LetterOfAuthorityCircuitTypeEnum = "Single_mode_LC"
	LetterOfAuthorityCircuitTypeSc LetterOfAuthorityCircuitTypeEnum = "Single_mode_SC"
)

var mappingLetterOfAuthorityCircuitTypeEnum = map[string]LetterOfAuthorityCircuitTypeEnum{
	"Single_mode_LC": LetterOfAuthorityCircuitTypeLc,
	"Single_mode_SC": LetterOfAuthorityCircuitTypeSc,
}

var mappingLetterOfAuthorityCircuitTypeEnumLowerCase = map[string]LetterOfAuthorityCircuitTypeEnum{
	"single_mode_lc": LetterOfAuthorityCircuitTypeLc,
	"single_mode_sc": LetterOfAuthorityCircuitTypeSc,
}

// GetLetterOfAuthorityCircuitTypeEnumValues Enumerates the set of values for LetterOfAuthorityCircuitTypeEnum
func GetLetterOfAuthorityCircuitTypeEnumValues() []LetterOfAuthorityCircuitTypeEnum {
	values := make([]LetterOfAuthorityCircuitTypeEnum, 0)
	for _, v := range mappingLetterOfAuthorityCircuitTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetLetterOfAuthorityCircuitTypeEnumStringValues Enumerates the set of values in String for LetterOfAuthorityCircuitTypeEnum
func GetLetterOfAuthorityCircuitTypeEnumStringValues() []string {
	return []string{
		"Single_mode_LC",
		"Single_mode_SC",
	}
}

// GetMappingLetterOfAuthorityCircuitTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingLetterOfAuthorityCircuitTypeEnum(val string) (LetterOfAuthorityCircuitTypeEnum, bool) {
	enum, ok := mappingLetterOfAuthorityCircuitTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
