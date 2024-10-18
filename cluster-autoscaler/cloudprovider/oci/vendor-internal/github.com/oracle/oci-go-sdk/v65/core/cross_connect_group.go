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

// CrossConnectGroup For use with Oracle Cloud Infrastructure FastConnect. A cross-connect group
// is a link aggregation group (LAG), which can contain one or more
// CrossConnect. Customers who are colocated with
// Oracle in a FastConnect location create and use cross-connect groups. For more
// information, see FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
// **Note:** If you're a provider who is setting up a physical connection to Oracle so customers
// can use FastConnect over the connection, be aware that your connection is modeled the
// same way as a colocated customer's (with `CrossConnect` and `CrossConnectGroup` objects, and so on).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type CrossConnectGroup struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the cross-connect group.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

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

	// The cross-connect group's Oracle ID (OCID).
	Id *string `mandatory:"false" json:"id"`

	// The cross-connect group's current state.
	LifecycleState CrossConnectGroupLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// A reference name or identifier for the physical fiber connection that this cross-connect
	// group uses.
	CustomerReferenceName *string `mandatory:"false" json:"customerReferenceName"`

	// The date and time the cross-connect group was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	MacsecProperties *MacsecProperties `mandatory:"false" json:"macsecProperties"`

	// The FastConnect device that terminates the physical connection.
	OciPhysicalDeviceName *string `mandatory:"false" json:"ociPhysicalDeviceName"`

	// The FastConnect device that terminates the logical connection.
	// This device might be different than the device that terminates the physical connection.
	OciLogicalDeviceName *string `mandatory:"false" json:"ociLogicalDeviceName"`
}

func (m CrossConnectGroup) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CrossConnectGroup) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingCrossConnectGroupLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetCrossConnectGroupLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CrossConnectGroupLifecycleStateEnum Enum with underlying type: string
type CrossConnectGroupLifecycleStateEnum string

// Set of constants representing the allowable values for CrossConnectGroupLifecycleStateEnum
const (
	CrossConnectGroupLifecycleStateProvisioning CrossConnectGroupLifecycleStateEnum = "PROVISIONING"
	CrossConnectGroupLifecycleStateProvisioned  CrossConnectGroupLifecycleStateEnum = "PROVISIONED"
	CrossConnectGroupLifecycleStateInactive     CrossConnectGroupLifecycleStateEnum = "INACTIVE"
	CrossConnectGroupLifecycleStateTerminating  CrossConnectGroupLifecycleStateEnum = "TERMINATING"
	CrossConnectGroupLifecycleStateTerminated   CrossConnectGroupLifecycleStateEnum = "TERMINATED"
)

var mappingCrossConnectGroupLifecycleStateEnum = map[string]CrossConnectGroupLifecycleStateEnum{
	"PROVISIONING": CrossConnectGroupLifecycleStateProvisioning,
	"PROVISIONED":  CrossConnectGroupLifecycleStateProvisioned,
	"INACTIVE":     CrossConnectGroupLifecycleStateInactive,
	"TERMINATING":  CrossConnectGroupLifecycleStateTerminating,
	"TERMINATED":   CrossConnectGroupLifecycleStateTerminated,
}

var mappingCrossConnectGroupLifecycleStateEnumLowerCase = map[string]CrossConnectGroupLifecycleStateEnum{
	"provisioning": CrossConnectGroupLifecycleStateProvisioning,
	"provisioned":  CrossConnectGroupLifecycleStateProvisioned,
	"inactive":     CrossConnectGroupLifecycleStateInactive,
	"terminating":  CrossConnectGroupLifecycleStateTerminating,
	"terminated":   CrossConnectGroupLifecycleStateTerminated,
}

// GetCrossConnectGroupLifecycleStateEnumValues Enumerates the set of values for CrossConnectGroupLifecycleStateEnum
func GetCrossConnectGroupLifecycleStateEnumValues() []CrossConnectGroupLifecycleStateEnum {
	values := make([]CrossConnectGroupLifecycleStateEnum, 0)
	for _, v := range mappingCrossConnectGroupLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetCrossConnectGroupLifecycleStateEnumStringValues Enumerates the set of values in String for CrossConnectGroupLifecycleStateEnum
func GetCrossConnectGroupLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"PROVISIONED",
		"INACTIVE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingCrossConnectGroupLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCrossConnectGroupLifecycleStateEnum(val string) (CrossConnectGroupLifecycleStateEnum, bool) {
	enum, ok := mappingCrossConnectGroupLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
