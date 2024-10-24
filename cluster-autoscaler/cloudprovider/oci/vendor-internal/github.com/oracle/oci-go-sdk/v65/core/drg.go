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

// Drg A dynamic routing gateway (DRG) is a virtual router that provides a path for private
// network traffic between networks. You use it with other Networking
// Service components to create a connection to your on-premises network using Site-to-Site VPN (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingIPsec.htm) or a connection that uses
// FastConnect (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm). For more information, see
// Networking Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type Drg struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the DRG.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The DRG's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The DRG's current state.
	LifecycleState DrgLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

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

	// The date and time the DRG was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	DefaultDrgRouteTables *DefaultDrgRouteTables `mandatory:"false" json:"defaultDrgRouteTables"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of this DRG's default export route distribution for the DRG attachments.
	DefaultExportDrgRouteDistributionId *string `mandatory:"false" json:"defaultExportDrgRouteDistributionId"`
}

func (m Drg) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m Drg) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingDrgLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetDrgLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// DrgLifecycleStateEnum Enum with underlying type: string
type DrgLifecycleStateEnum string

// Set of constants representing the allowable values for DrgLifecycleStateEnum
const (
	DrgLifecycleStateProvisioning DrgLifecycleStateEnum = "PROVISIONING"
	DrgLifecycleStateAvailable    DrgLifecycleStateEnum = "AVAILABLE"
	DrgLifecycleStateTerminating  DrgLifecycleStateEnum = "TERMINATING"
	DrgLifecycleStateTerminated   DrgLifecycleStateEnum = "TERMINATED"
)

var mappingDrgLifecycleStateEnum = map[string]DrgLifecycleStateEnum{
	"PROVISIONING": DrgLifecycleStateProvisioning,
	"AVAILABLE":    DrgLifecycleStateAvailable,
	"TERMINATING":  DrgLifecycleStateTerminating,
	"TERMINATED":   DrgLifecycleStateTerminated,
}

var mappingDrgLifecycleStateEnumLowerCase = map[string]DrgLifecycleStateEnum{
	"provisioning": DrgLifecycleStateProvisioning,
	"available":    DrgLifecycleStateAvailable,
	"terminating":  DrgLifecycleStateTerminating,
	"terminated":   DrgLifecycleStateTerminated,
}

// GetDrgLifecycleStateEnumValues Enumerates the set of values for DrgLifecycleStateEnum
func GetDrgLifecycleStateEnumValues() []DrgLifecycleStateEnum {
	values := make([]DrgLifecycleStateEnum, 0)
	for _, v := range mappingDrgLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgLifecycleStateEnumStringValues Enumerates the set of values in String for DrgLifecycleStateEnum
func GetDrgLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingDrgLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgLifecycleStateEnum(val string) (DrgLifecycleStateEnum, bool) {
	enum, ok := mappingDrgLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
