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

// InternalDrg A dynamic routing gateway (DRG), which is a virtual router that provides a path for private
// network traffic between your VCN and your existing network. You use it with other Networking
// Service components to create a Site-to-Site VPN or a connection that uses
// Oracle Cloud Infrastructure FastConnect. For more information, see
// Overview of the Networking Service (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type InternalDrg struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the DRG.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The DRG's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The DRG's current state.
	LifecycleState InternalDrgLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Anycast IP of the El Paso fleet handling the ingress traffic.
	IngressIP *string `mandatory:"true" json:"ingressIP"`

	// Anycast IP of the El Paso fleet handling the egress traffic.
	EgressIP *string `mandatory:"true" json:"egressIP"`

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

	// Route data for the Drg.
	RouteData *string `mandatory:"false" json:"routeData"`

	// NextHop target's MPLS label.
	MplsLabel *string `mandatory:"false" json:"mplsLabel"`

	// The string in the form ASN:mplsLabel.
	RouteTarget *string `mandatory:"false" json:"routeTarget"`

	// The type of the DRG.
	DrgType InternalDrgDrgTypeEnum `mandatory:"false" json:"drgType,omitempty"`
}

func (m InternalDrg) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InternalDrg) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingInternalDrgLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInternalDrgLifecycleStateEnumStringValues(), ",")))
	}

	if _, ok := mappingInternalDrgDrgTypeEnum[string(m.DrgType)]; !ok && m.DrgType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DrgType: %s. Supported values are: %s.", m.DrgType, strings.Join(GetInternalDrgDrgTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InternalDrgLifecycleStateEnum Enum with underlying type: string
type InternalDrgLifecycleStateEnum string

// Set of constants representing the allowable values for InternalDrgLifecycleStateEnum
const (
	InternalDrgLifecycleStateProvisioning InternalDrgLifecycleStateEnum = "PROVISIONING"
	InternalDrgLifecycleStateAvailable    InternalDrgLifecycleStateEnum = "AVAILABLE"
	InternalDrgLifecycleStateTerminating  InternalDrgLifecycleStateEnum = "TERMINATING"
	InternalDrgLifecycleStateTerminated   InternalDrgLifecycleStateEnum = "TERMINATED"
)

var mappingInternalDrgLifecycleStateEnum = map[string]InternalDrgLifecycleStateEnum{
	"PROVISIONING": InternalDrgLifecycleStateProvisioning,
	"AVAILABLE":    InternalDrgLifecycleStateAvailable,
	"TERMINATING":  InternalDrgLifecycleStateTerminating,
	"TERMINATED":   InternalDrgLifecycleStateTerminated,
}

// GetInternalDrgLifecycleStateEnumValues Enumerates the set of values for InternalDrgLifecycleStateEnum
func GetInternalDrgLifecycleStateEnumValues() []InternalDrgLifecycleStateEnum {
	values := make([]InternalDrgLifecycleStateEnum, 0)
	for _, v := range mappingInternalDrgLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalDrgLifecycleStateEnumStringValues Enumerates the set of values in String for InternalDrgLifecycleStateEnum
func GetInternalDrgLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// InternalDrgDrgTypeEnum Enum with underlying type: string
type InternalDrgDrgTypeEnum string

// Set of constants representing the allowable values for InternalDrgDrgTypeEnum
const (
	InternalDrgDrgTypeClassical  InternalDrgDrgTypeEnum = "DRG_CLASSICAL"
	InternalDrgDrgTypeTransitHub InternalDrgDrgTypeEnum = "DRG_TRANSIT_HUB"
)

var mappingInternalDrgDrgTypeEnum = map[string]InternalDrgDrgTypeEnum{
	"DRG_CLASSICAL":   InternalDrgDrgTypeClassical,
	"DRG_TRANSIT_HUB": InternalDrgDrgTypeTransitHub,
}

// GetInternalDrgDrgTypeEnumValues Enumerates the set of values for InternalDrgDrgTypeEnum
func GetInternalDrgDrgTypeEnumValues() []InternalDrgDrgTypeEnum {
	values := make([]InternalDrgDrgTypeEnum, 0)
	for _, v := range mappingInternalDrgDrgTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalDrgDrgTypeEnumStringValues Enumerates the set of values in String for InternalDrgDrgTypeEnum
func GetInternalDrgDrgTypeEnumStringValues() []string {
	return []string{
		"DRG_CLASSICAL",
		"DRG_TRANSIT_HUB",
	}
}
