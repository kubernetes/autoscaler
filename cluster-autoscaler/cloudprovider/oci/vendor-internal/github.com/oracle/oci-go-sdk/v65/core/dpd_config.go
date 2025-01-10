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

// DpdConfig These configuration details are used for dead peer detection (DPD). DPD periodically checks the stability of the connection to the customer premises (CPE), and may be used to detect that the link to the CPE has gone down.
type DpdConfig struct {

	// This option defines whether DPD can be initiated from the Oracle side of the connection.
	DpdMode DpdConfigDpdModeEnum `mandatory:"false" json:"dpdMode,omitempty"`

	// DPD timeout in seconds. This sets the longest interval between CPE device health messages before the IPSec connection indicates it has lost contact with the CPE. The default is 20 seconds.
	DpdTimeoutInSec *int `mandatory:"false" json:"dpdTimeoutInSec"`
}

func (m DpdConfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DpdConfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingDpdConfigDpdModeEnum(string(m.DpdMode)); !ok && m.DpdMode != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DpdMode: %s. Supported values are: %s.", m.DpdMode, strings.Join(GetDpdConfigDpdModeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// DpdConfigDpdModeEnum Enum with underlying type: string
type DpdConfigDpdModeEnum string

// Set of constants representing the allowable values for DpdConfigDpdModeEnum
const (
	DpdConfigDpdModeInitiateAndRespond DpdConfigDpdModeEnum = "INITIATE_AND_RESPOND"
	DpdConfigDpdModeRespondOnly        DpdConfigDpdModeEnum = "RESPOND_ONLY"
)

var mappingDpdConfigDpdModeEnum = map[string]DpdConfigDpdModeEnum{
	"INITIATE_AND_RESPOND": DpdConfigDpdModeInitiateAndRespond,
	"RESPOND_ONLY":         DpdConfigDpdModeRespondOnly,
}

var mappingDpdConfigDpdModeEnumLowerCase = map[string]DpdConfigDpdModeEnum{
	"initiate_and_respond": DpdConfigDpdModeInitiateAndRespond,
	"respond_only":         DpdConfigDpdModeRespondOnly,
}

// GetDpdConfigDpdModeEnumValues Enumerates the set of values for DpdConfigDpdModeEnum
func GetDpdConfigDpdModeEnumValues() []DpdConfigDpdModeEnum {
	values := make([]DpdConfigDpdModeEnum, 0)
	for _, v := range mappingDpdConfigDpdModeEnum {
		values = append(values, v)
	}
	return values
}

// GetDpdConfigDpdModeEnumStringValues Enumerates the set of values in String for DpdConfigDpdModeEnum
func GetDpdConfigDpdModeEnumStringValues() []string {
	return []string{
		"INITIATE_AND_RESPOND",
		"RESPOND_ONLY",
	}
}

// GetMappingDpdConfigDpdModeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDpdConfigDpdModeEnum(val string) (DpdConfigDpdModeEnum, bool) {
	enum, ok := mappingDpdConfigDpdModeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
