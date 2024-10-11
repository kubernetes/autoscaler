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

// ByoipRange Oracle offers the ability to Bring Your Own IP (BYOIP), importing public IP addresses or IPv6 addresses that you currently own to Oracle Cloud Infrastructure. A `ByoipRange` resource is a record of the imported address block (a BYOIP CIDR block) and also some associated metadata.
// The process used to Bring Your Own IP (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/BYOIP.htm) is explained in the documentation.
type ByoipRange struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the BYOIP CIDR block.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the `ByoipRange` resource.
	Id *string `mandatory:"true" json:"id"`

	// The `ByoipRange` resource's current state.
	LifecycleState ByoipRangeLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the `ByoipRange` resource was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The validation token is an internally-generated ASCII string used in the validation process. See Importing a CIDR block (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/BYOIP.htm#import_cidr) for details.
	ValidationToken *string `mandatory:"true" json:"validationToken"`

	// A list of `ByoipRangeVcnIpv6AllocationSummary` objects.
	ByoipRangeVcnIpv6Allocations []ByoipRangeVcnIpv6AllocationSummary `mandatory:"false" json:"byoipRangeVcnIpv6Allocations"`

	// The public IPv4 CIDR block being imported from on-premises to the Oracle cloud.
	CidrBlock *string `mandatory:"false" json:"cidrBlock"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The IPv6 prefix being imported to the Oracle cloud. This prefix must be /48 or larger, and can be subdivided into sub-ranges used
	// across multiple VCNs. A BYOIPv6 prefix can be also assigned across multiple VCNs, and each VCN must be /64 or larger. You may specify
	// a ULA or private IPv6 prefix of /64 or larger to use in the VCN. IPv6-enabled subnets will remain a fixed /64 in size.
	Ipv6CidrBlock *string `mandatory:"false" json:"ipv6CidrBlock"`

	// The `ByoipRange` resource's current status.
	LifecycleDetails ByoipRangeLifecycleDetailsEnum `mandatory:"false" json:"lifecycleDetails,omitempty"`

	// The date and time the `ByoipRange` resource was validated, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeValidated *common.SDKTime `mandatory:"false" json:"timeValidated"`

	// The date and time the `ByoipRange` resource was advertised to the internet by BGP, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeAdvertised *common.SDKTime `mandatory:"false" json:"timeAdvertised"`

	// The date and time the `ByoipRange` resource was withdrawn from advertisement by BGP to the internet, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeWithdrawn *common.SDKTime `mandatory:"false" json:"timeWithdrawn"`
}

func (m ByoipRange) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ByoipRange) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingByoipRangeLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetByoipRangeLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := GetMappingByoipRangeLifecycleDetailsEnum(string(m.LifecycleDetails)); !ok && m.LifecycleDetails != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleDetails: %s. Supported values are: %s.", m.LifecycleDetails, strings.Join(GetByoipRangeLifecycleDetailsEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ByoipRangeLifecycleDetailsEnum Enum with underlying type: string
type ByoipRangeLifecycleDetailsEnum string

// Set of constants representing the allowable values for ByoipRangeLifecycleDetailsEnum
const (
	ByoipRangeLifecycleDetailsCreating    ByoipRangeLifecycleDetailsEnum = "CREATING"
	ByoipRangeLifecycleDetailsValidating  ByoipRangeLifecycleDetailsEnum = "VALIDATING"
	ByoipRangeLifecycleDetailsProvisioned ByoipRangeLifecycleDetailsEnum = "PROVISIONED"
	ByoipRangeLifecycleDetailsActive      ByoipRangeLifecycleDetailsEnum = "ACTIVE"
	ByoipRangeLifecycleDetailsFailed      ByoipRangeLifecycleDetailsEnum = "FAILED"
	ByoipRangeLifecycleDetailsDeleting    ByoipRangeLifecycleDetailsEnum = "DELETING"
	ByoipRangeLifecycleDetailsDeleted     ByoipRangeLifecycleDetailsEnum = "DELETED"
	ByoipRangeLifecycleDetailsAdvertising ByoipRangeLifecycleDetailsEnum = "ADVERTISING"
	ByoipRangeLifecycleDetailsWithdrawing ByoipRangeLifecycleDetailsEnum = "WITHDRAWING"
)

var mappingByoipRangeLifecycleDetailsEnum = map[string]ByoipRangeLifecycleDetailsEnum{
	"CREATING":    ByoipRangeLifecycleDetailsCreating,
	"VALIDATING":  ByoipRangeLifecycleDetailsValidating,
	"PROVISIONED": ByoipRangeLifecycleDetailsProvisioned,
	"ACTIVE":      ByoipRangeLifecycleDetailsActive,
	"FAILED":      ByoipRangeLifecycleDetailsFailed,
	"DELETING":    ByoipRangeLifecycleDetailsDeleting,
	"DELETED":     ByoipRangeLifecycleDetailsDeleted,
	"ADVERTISING": ByoipRangeLifecycleDetailsAdvertising,
	"WITHDRAWING": ByoipRangeLifecycleDetailsWithdrawing,
}

var mappingByoipRangeLifecycleDetailsEnumLowerCase = map[string]ByoipRangeLifecycleDetailsEnum{
	"creating":    ByoipRangeLifecycleDetailsCreating,
	"validating":  ByoipRangeLifecycleDetailsValidating,
	"provisioned": ByoipRangeLifecycleDetailsProvisioned,
	"active":      ByoipRangeLifecycleDetailsActive,
	"failed":      ByoipRangeLifecycleDetailsFailed,
	"deleting":    ByoipRangeLifecycleDetailsDeleting,
	"deleted":     ByoipRangeLifecycleDetailsDeleted,
	"advertising": ByoipRangeLifecycleDetailsAdvertising,
	"withdrawing": ByoipRangeLifecycleDetailsWithdrawing,
}

// GetByoipRangeLifecycleDetailsEnumValues Enumerates the set of values for ByoipRangeLifecycleDetailsEnum
func GetByoipRangeLifecycleDetailsEnumValues() []ByoipRangeLifecycleDetailsEnum {
	values := make([]ByoipRangeLifecycleDetailsEnum, 0)
	for _, v := range mappingByoipRangeLifecycleDetailsEnum {
		values = append(values, v)
	}
	return values
}

// GetByoipRangeLifecycleDetailsEnumStringValues Enumerates the set of values in String for ByoipRangeLifecycleDetailsEnum
func GetByoipRangeLifecycleDetailsEnumStringValues() []string {
	return []string{
		"CREATING",
		"VALIDATING",
		"PROVISIONED",
		"ACTIVE",
		"FAILED",
		"DELETING",
		"DELETED",
		"ADVERTISING",
		"WITHDRAWING",
	}
}

// GetMappingByoipRangeLifecycleDetailsEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingByoipRangeLifecycleDetailsEnum(val string) (ByoipRangeLifecycleDetailsEnum, bool) {
	enum, ok := mappingByoipRangeLifecycleDetailsEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// ByoipRangeLifecycleStateEnum Enum with underlying type: string
type ByoipRangeLifecycleStateEnum string

// Set of constants representing the allowable values for ByoipRangeLifecycleStateEnum
const (
	ByoipRangeLifecycleStateInactive ByoipRangeLifecycleStateEnum = "INACTIVE"
	ByoipRangeLifecycleStateUpdating ByoipRangeLifecycleStateEnum = "UPDATING"
	ByoipRangeLifecycleStateActive   ByoipRangeLifecycleStateEnum = "ACTIVE"
	ByoipRangeLifecycleStateDeleting ByoipRangeLifecycleStateEnum = "DELETING"
	ByoipRangeLifecycleStateDeleted  ByoipRangeLifecycleStateEnum = "DELETED"
)

var mappingByoipRangeLifecycleStateEnum = map[string]ByoipRangeLifecycleStateEnum{
	"INACTIVE": ByoipRangeLifecycleStateInactive,
	"UPDATING": ByoipRangeLifecycleStateUpdating,
	"ACTIVE":   ByoipRangeLifecycleStateActive,
	"DELETING": ByoipRangeLifecycleStateDeleting,
	"DELETED":  ByoipRangeLifecycleStateDeleted,
}

var mappingByoipRangeLifecycleStateEnumLowerCase = map[string]ByoipRangeLifecycleStateEnum{
	"inactive": ByoipRangeLifecycleStateInactive,
	"updating": ByoipRangeLifecycleStateUpdating,
	"active":   ByoipRangeLifecycleStateActive,
	"deleting": ByoipRangeLifecycleStateDeleting,
	"deleted":  ByoipRangeLifecycleStateDeleted,
}

// GetByoipRangeLifecycleStateEnumValues Enumerates the set of values for ByoipRangeLifecycleStateEnum
func GetByoipRangeLifecycleStateEnumValues() []ByoipRangeLifecycleStateEnum {
	values := make([]ByoipRangeLifecycleStateEnum, 0)
	for _, v := range mappingByoipRangeLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetByoipRangeLifecycleStateEnumStringValues Enumerates the set of values in String for ByoipRangeLifecycleStateEnum
func GetByoipRangeLifecycleStateEnumStringValues() []string {
	return []string{
		"INACTIVE",
		"UPDATING",
		"ACTIVE",
		"DELETING",
		"DELETED",
	}
}

// GetMappingByoipRangeLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingByoipRangeLifecycleStateEnum(val string) (ByoipRangeLifecycleStateEnum, bool) {
	enum, ok := mappingByoipRangeLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
