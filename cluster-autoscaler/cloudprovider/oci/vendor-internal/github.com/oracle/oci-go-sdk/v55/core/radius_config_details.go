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

// RadiusConfigDetails The detail of RADIUS's authentication configuration.
type RadiusConfigDetails struct {

	// Allowed values:
	//   * `PAP`: An authentication option used by PPP to validate user.
	//   * `CHAP`: An authentication option used by PPP to validate user but performed only at the initial link establishment.
	//   * `MS-CHAP-v2`: An authentication option used by PPP to periodically validate the user.
	AuthenticationType RadiusConfigDetailsAuthenticationTypeEnum `mandatory:"false" json:"authenticationType,omitempty"`

	// A list of usable RADIUS server.
	Servers []RadiusServerDetails `mandatory:"false" json:"servers"`

	// Whether to enable RADIUS accounting. When this enabled, accouting port becomes required filed.
	IsRadiusAccountingEnabled *bool `mandatory:"false" json:"isRadiusAccountingEnabled"`

	// Enable case-sensitivity or not in RADIUS authentication.
	IsCaseSensitive *bool `mandatory:"false" json:"isCaseSensitive"`
}

func (m RadiusConfigDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m RadiusConfigDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingRadiusConfigDetailsAuthenticationTypeEnum[string(m.AuthenticationType)]; !ok && m.AuthenticationType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AuthenticationType: %s. Supported values are: %s.", m.AuthenticationType, strings.Join(GetRadiusConfigDetailsAuthenticationTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// RadiusConfigDetailsAuthenticationTypeEnum Enum with underlying type: string
type RadiusConfigDetailsAuthenticationTypeEnum string

// Set of constants representing the allowable values for RadiusConfigDetailsAuthenticationTypeEnum
const (
	RadiusConfigDetailsAuthenticationTypePap      RadiusConfigDetailsAuthenticationTypeEnum = "PAP"
	RadiusConfigDetailsAuthenticationTypeChap     RadiusConfigDetailsAuthenticationTypeEnum = "CHAP"
	RadiusConfigDetailsAuthenticationTypeMsChapV2 RadiusConfigDetailsAuthenticationTypeEnum = "MS_CHAP_V2"
)

var mappingRadiusConfigDetailsAuthenticationTypeEnum = map[string]RadiusConfigDetailsAuthenticationTypeEnum{
	"PAP":        RadiusConfigDetailsAuthenticationTypePap,
	"CHAP":       RadiusConfigDetailsAuthenticationTypeChap,
	"MS_CHAP_V2": RadiusConfigDetailsAuthenticationTypeMsChapV2,
}

// GetRadiusConfigDetailsAuthenticationTypeEnumValues Enumerates the set of values for RadiusConfigDetailsAuthenticationTypeEnum
func GetRadiusConfigDetailsAuthenticationTypeEnumValues() []RadiusConfigDetailsAuthenticationTypeEnum {
	values := make([]RadiusConfigDetailsAuthenticationTypeEnum, 0)
	for _, v := range mappingRadiusConfigDetailsAuthenticationTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetRadiusConfigDetailsAuthenticationTypeEnumStringValues Enumerates the set of values in String for RadiusConfigDetailsAuthenticationTypeEnum
func GetRadiusConfigDetailsAuthenticationTypeEnumStringValues() []string {
	return []string{
		"PAP",
		"CHAP",
		"MS_CHAP_V2",
	}
}
