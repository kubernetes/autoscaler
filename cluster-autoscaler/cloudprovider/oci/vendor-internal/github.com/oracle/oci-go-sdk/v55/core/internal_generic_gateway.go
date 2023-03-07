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

// InternalGenericGateway An internal generic gateway.
type InternalGenericGateway struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the generic gateway.
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the gateway's compartment.
	GatewayCompartmentId *string `mandatory:"true" json:"gatewayCompartmentId"`

	// Information required to fill headers of packets to be sent to the gateway.
	GatewayHeaderData *int64 `mandatory:"true" json:"gatewayHeaderData"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the real gateway that this generic gateway stands for.
	GatewayId *string `mandatory:"true" json:"gatewayId"`

	// They type of the gateway.
	GatewayType InternalGenericGatewayGatewayTypeEnum `mandatory:"true" json:"gatewayType"`

	// The current state of the generic gateway.
	LifecycleState InternalGenericGatewayLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// IP addresses of the gateway.
	GatewayIpAddresses []string `mandatory:"false" json:"gatewayIpAddresses"`

	// Tuples, mapping AD and regional identifiers to the corresponding routing data
	GatewayRouteMap []GatewayRouteData `mandatory:"false" json:"gatewayRouteMap"`

	// Creation time of the entity.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN the generic gateway belongs to.
	VcnId *string `mandatory:"false" json:"vcnId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the route table associated with the gateway
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see
	// Resource T
	// ags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m InternalGenericGateway) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InternalGenericGateway) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingInternalGenericGatewayGatewayTypeEnum[string(m.GatewayType)]; !ok && m.GatewayType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for GatewayType: %s. Supported values are: %s.", m.GatewayType, strings.Join(GetInternalGenericGatewayGatewayTypeEnumStringValues(), ",")))
	}
	if _, ok := mappingInternalGenericGatewayLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInternalGenericGatewayLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InternalGenericGatewayGatewayTypeEnum Enum with underlying type: string
type InternalGenericGatewayGatewayTypeEnum string

// Set of constants representing the allowable values for InternalGenericGatewayGatewayTypeEnum
const (
	InternalGenericGatewayGatewayTypeServicegateway       InternalGenericGatewayGatewayTypeEnum = "SERVICEGATEWAY"
	InternalGenericGatewayGatewayTypeNatgateway           InternalGenericGatewayGatewayTypeEnum = "NATGATEWAY"
	InternalGenericGatewayGatewayTypePrivateaccessgateway InternalGenericGatewayGatewayTypeEnum = "PRIVATEACCESSGATEWAY"
)

var mappingInternalGenericGatewayGatewayTypeEnum = map[string]InternalGenericGatewayGatewayTypeEnum{
	"SERVICEGATEWAY":       InternalGenericGatewayGatewayTypeServicegateway,
	"NATGATEWAY":           InternalGenericGatewayGatewayTypeNatgateway,
	"PRIVATEACCESSGATEWAY": InternalGenericGatewayGatewayTypePrivateaccessgateway,
}

// GetInternalGenericGatewayGatewayTypeEnumValues Enumerates the set of values for InternalGenericGatewayGatewayTypeEnum
func GetInternalGenericGatewayGatewayTypeEnumValues() []InternalGenericGatewayGatewayTypeEnum {
	values := make([]InternalGenericGatewayGatewayTypeEnum, 0)
	for _, v := range mappingInternalGenericGatewayGatewayTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalGenericGatewayGatewayTypeEnumStringValues Enumerates the set of values in String for InternalGenericGatewayGatewayTypeEnum
func GetInternalGenericGatewayGatewayTypeEnumStringValues() []string {
	return []string{
		"SERVICEGATEWAY",
		"NATGATEWAY",
		"PRIVATEACCESSGATEWAY",
	}
}

// InternalGenericGatewayLifecycleStateEnum Enum with underlying type: string
type InternalGenericGatewayLifecycleStateEnum string

// Set of constants representing the allowable values for InternalGenericGatewayLifecycleStateEnum
const (
	InternalGenericGatewayLifecycleStateProvisioning InternalGenericGatewayLifecycleStateEnum = "PROVISIONING"
	InternalGenericGatewayLifecycleStateAvailable    InternalGenericGatewayLifecycleStateEnum = "AVAILABLE"
	InternalGenericGatewayLifecycleStateAttaching    InternalGenericGatewayLifecycleStateEnum = "ATTACHING"
	InternalGenericGatewayLifecycleStateAttached     InternalGenericGatewayLifecycleStateEnum = "ATTACHED"
	InternalGenericGatewayLifecycleStateDetaching    InternalGenericGatewayLifecycleStateEnum = "DETACHING"
	InternalGenericGatewayLifecycleStateDetached     InternalGenericGatewayLifecycleStateEnum = "DETACHED"
	InternalGenericGatewayLifecycleStateTerminating  InternalGenericGatewayLifecycleStateEnum = "TERMINATING"
	InternalGenericGatewayLifecycleStateTerminated   InternalGenericGatewayLifecycleStateEnum = "TERMINATED"
)

var mappingInternalGenericGatewayLifecycleStateEnum = map[string]InternalGenericGatewayLifecycleStateEnum{
	"PROVISIONING": InternalGenericGatewayLifecycleStateProvisioning,
	"AVAILABLE":    InternalGenericGatewayLifecycleStateAvailable,
	"ATTACHING":    InternalGenericGatewayLifecycleStateAttaching,
	"ATTACHED":     InternalGenericGatewayLifecycleStateAttached,
	"DETACHING":    InternalGenericGatewayLifecycleStateDetaching,
	"DETACHED":     InternalGenericGatewayLifecycleStateDetached,
	"TERMINATING":  InternalGenericGatewayLifecycleStateTerminating,
	"TERMINATED":   InternalGenericGatewayLifecycleStateTerminated,
}

// GetInternalGenericGatewayLifecycleStateEnumValues Enumerates the set of values for InternalGenericGatewayLifecycleStateEnum
func GetInternalGenericGatewayLifecycleStateEnumValues() []InternalGenericGatewayLifecycleStateEnum {
	values := make([]InternalGenericGatewayLifecycleStateEnum, 0)
	for _, v := range mappingInternalGenericGatewayLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalGenericGatewayLifecycleStateEnumStringValues Enumerates the set of values in String for InternalGenericGatewayLifecycleStateEnum
func GetInternalGenericGatewayLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"ATTACHING",
		"ATTACHED",
		"DETACHING",
		"DETACHED",
		"TERMINATING",
		"TERMINATED",
	}
}
