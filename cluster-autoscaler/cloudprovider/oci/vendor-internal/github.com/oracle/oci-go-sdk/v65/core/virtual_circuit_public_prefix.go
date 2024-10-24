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

// VirtualCircuitPublicPrefix A public IP prefix and its details. With a public virtual circuit, the customer
// specifies the customer-owned public IP prefixes to advertise across the connection.
// For more information, see FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
type VirtualCircuitPublicPrefix struct {

	// Publix IP prefix (CIDR) that the customer specified.
	CidrBlock *string `mandatory:"true" json:"cidrBlock"`

	// Oracle must verify that the customer owns the public IP prefix before traffic
	// for that prefix can flow across the virtual circuit. Verification can take a
	// few business days. `IN_PROGRESS` means Oracle is verifying the prefix. `COMPLETED`
	// means verification succeeded. `FAILED` means verification failed and traffic for
	// this prefix will not flow across the connection.
	VerificationState VirtualCircuitPublicPrefixVerificationStateEnum `mandatory:"true" json:"verificationState"`
}

func (m VirtualCircuitPublicPrefix) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VirtualCircuitPublicPrefix) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingVirtualCircuitPublicPrefixVerificationStateEnum(string(m.VerificationState)); !ok && m.VerificationState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for VerificationState: %s. Supported values are: %s.", m.VerificationState, strings.Join(GetVirtualCircuitPublicPrefixVerificationStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// VirtualCircuitPublicPrefixVerificationStateEnum Enum with underlying type: string
type VirtualCircuitPublicPrefixVerificationStateEnum string

// Set of constants representing the allowable values for VirtualCircuitPublicPrefixVerificationStateEnum
const (
	VirtualCircuitPublicPrefixVerificationStateInProgress VirtualCircuitPublicPrefixVerificationStateEnum = "IN_PROGRESS"
	VirtualCircuitPublicPrefixVerificationStateCompleted  VirtualCircuitPublicPrefixVerificationStateEnum = "COMPLETED"
	VirtualCircuitPublicPrefixVerificationStateFailed     VirtualCircuitPublicPrefixVerificationStateEnum = "FAILED"
)

var mappingVirtualCircuitPublicPrefixVerificationStateEnum = map[string]VirtualCircuitPublicPrefixVerificationStateEnum{
	"IN_PROGRESS": VirtualCircuitPublicPrefixVerificationStateInProgress,
	"COMPLETED":   VirtualCircuitPublicPrefixVerificationStateCompleted,
	"FAILED":      VirtualCircuitPublicPrefixVerificationStateFailed,
}

var mappingVirtualCircuitPublicPrefixVerificationStateEnumLowerCase = map[string]VirtualCircuitPublicPrefixVerificationStateEnum{
	"in_progress": VirtualCircuitPublicPrefixVerificationStateInProgress,
	"completed":   VirtualCircuitPublicPrefixVerificationStateCompleted,
	"failed":      VirtualCircuitPublicPrefixVerificationStateFailed,
}

// GetVirtualCircuitPublicPrefixVerificationStateEnumValues Enumerates the set of values for VirtualCircuitPublicPrefixVerificationStateEnum
func GetVirtualCircuitPublicPrefixVerificationStateEnumValues() []VirtualCircuitPublicPrefixVerificationStateEnum {
	values := make([]VirtualCircuitPublicPrefixVerificationStateEnum, 0)
	for _, v := range mappingVirtualCircuitPublicPrefixVerificationStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVirtualCircuitPublicPrefixVerificationStateEnumStringValues Enumerates the set of values in String for VirtualCircuitPublicPrefixVerificationStateEnum
func GetVirtualCircuitPublicPrefixVerificationStateEnumStringValues() []string {
	return []string{
		"IN_PROGRESS",
		"COMPLETED",
		"FAILED",
	}
}

// GetMappingVirtualCircuitPublicPrefixVerificationStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVirtualCircuitPublicPrefixVerificationStateEnum(val string) (VirtualCircuitPublicPrefixVerificationStateEnum, bool) {
	enum, ok := mappingVirtualCircuitPublicPrefixVerificationStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
