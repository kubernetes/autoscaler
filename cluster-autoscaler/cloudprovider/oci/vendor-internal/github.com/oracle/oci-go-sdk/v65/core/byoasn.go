// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// Byoasn Oracle offers the ability to Bring Your Own Autonomous System Number (BYOASN), importing AS Numbers you currently own to Oracle Cloud Infrastructure. A `Byoasn` resource is a record of the imported AS Number and also some associated metadata. The process used to Bring Your Own ASN (https://docs.oracle.com/iaas/Content/Network/Concepts/BYOASN.htm) is explained in the documentation.
type Byoasn struct {

	// The `Byoasn` resource's current state.
	LifecycleState ByoasnLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the `Byoasn` resource.
	Id *string `mandatory:"true" json:"id"`

	// The Autonomous System Number (ASN) you are importing to the Oracle cloud.
	Asn *int64 `mandatory:"true" json:"asn"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the `Byoasn` resource.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The validation token is an internally-generated ASCII string used in the validation process. See Importing a Byoasn (https://docs.oracle.com/iaas/Content/Network/Concepts/BYOASN.htm) for details.
	ValidationToken *string `mandatory:"true" json:"validationToken"`

	// The date and time the `Byoasn` resource was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The date and time the `Byoasn` resource was validated, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeValidated *common.SDKTime `mandatory:"false" json:"timeValidated"`

	// The date and time the `Byoasn` resource was last updated, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// The BYOIP Ranges that has the `Byoasn` as origin.
	ByoipRanges []ByoasnByoipRange `mandatory:"false" json:"byoipRanges"`
}

func (m Byoasn) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m Byoasn) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingByoasnLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetByoasnLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ByoasnLifecycleStateEnum Enum with underlying type: string
type ByoasnLifecycleStateEnum string

// Set of constants representing the allowable values for ByoasnLifecycleStateEnum
const (
	ByoasnLifecycleStateUpdating ByoasnLifecycleStateEnum = "UPDATING"
	ByoasnLifecycleStateActive   ByoasnLifecycleStateEnum = "ACTIVE"
	ByoasnLifecycleStateDeleted  ByoasnLifecycleStateEnum = "DELETED"
	ByoasnLifecycleStateFailed   ByoasnLifecycleStateEnum = "FAILED"
	ByoasnLifecycleStateCreating ByoasnLifecycleStateEnum = "CREATING"
)

var mappingByoasnLifecycleStateEnum = map[string]ByoasnLifecycleStateEnum{
	"UPDATING": ByoasnLifecycleStateUpdating,
	"ACTIVE":   ByoasnLifecycleStateActive,
	"DELETED":  ByoasnLifecycleStateDeleted,
	"FAILED":   ByoasnLifecycleStateFailed,
	"CREATING": ByoasnLifecycleStateCreating,
}

var mappingByoasnLifecycleStateEnumLowerCase = map[string]ByoasnLifecycleStateEnum{
	"updating": ByoasnLifecycleStateUpdating,
	"active":   ByoasnLifecycleStateActive,
	"deleted":  ByoasnLifecycleStateDeleted,
	"failed":   ByoasnLifecycleStateFailed,
	"creating": ByoasnLifecycleStateCreating,
}

// GetByoasnLifecycleStateEnumValues Enumerates the set of values for ByoasnLifecycleStateEnum
func GetByoasnLifecycleStateEnumValues() []ByoasnLifecycleStateEnum {
	values := make([]ByoasnLifecycleStateEnum, 0)
	for _, v := range mappingByoasnLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetByoasnLifecycleStateEnumStringValues Enumerates the set of values in String for ByoasnLifecycleStateEnum
func GetByoasnLifecycleStateEnumStringValues() []string {
	return []string{
		"UPDATING",
		"ACTIVE",
		"DELETED",
		"FAILED",
		"CREATING",
	}
}

// GetMappingByoasnLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingByoasnLifecycleStateEnum(val string) (ByoasnLifecycleStateEnum, bool) {
	enum, ok := mappingByoasnLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
