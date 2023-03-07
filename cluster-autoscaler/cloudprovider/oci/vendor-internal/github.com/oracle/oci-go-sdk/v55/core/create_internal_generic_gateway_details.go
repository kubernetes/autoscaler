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

// CreateInternalGenericGatewayDetails Details to create an internal generic gateway.
type CreateInternalGenericGatewayDetails struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the gateway's compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Information required to fill headers of packets to be sent to the gateway.
	GatewayHeaderData *int64 `mandatory:"true" json:"gatewayHeaderData"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the real gateway that this generic gateway represents.
	GatewayId *string `mandatory:"true" json:"gatewayId"`

	// The type of the gateway.
	GatewayType CreateInternalGenericGatewayDetailsGatewayTypeEnum `mandatory:"true" json:"gatewayType"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN the generic gateway belongs to.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// IP address of the gateway.
	GatewayIpAddresses []string `mandatory:"false" json:"gatewayIpAddresses"`

	// Tuples, mapping AD and regional identifiers to the corresponding routing data
	GatewayRouteMap []GatewayRouteData `mandatory:"false" json:"gatewayRouteMap"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the route table associated with the gateway
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see
	// Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m CreateInternalGenericGatewayDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateInternalGenericGatewayDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingCreateInternalGenericGatewayDetailsGatewayTypeEnum[string(m.GatewayType)]; !ok && m.GatewayType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for GatewayType: %s. Supported values are: %s.", m.GatewayType, strings.Join(GetCreateInternalGenericGatewayDetailsGatewayTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateInternalGenericGatewayDetailsGatewayTypeEnum Enum with underlying type: string
type CreateInternalGenericGatewayDetailsGatewayTypeEnum string

// Set of constants representing the allowable values for CreateInternalGenericGatewayDetailsGatewayTypeEnum
const (
	CreateInternalGenericGatewayDetailsGatewayTypeServicegateway       CreateInternalGenericGatewayDetailsGatewayTypeEnum = "SERVICEGATEWAY"
	CreateInternalGenericGatewayDetailsGatewayTypeNatgateway           CreateInternalGenericGatewayDetailsGatewayTypeEnum = "NATGATEWAY"
	CreateInternalGenericGatewayDetailsGatewayTypePrivateaccessgateway CreateInternalGenericGatewayDetailsGatewayTypeEnum = "PRIVATEACCESSGATEWAY"
)

var mappingCreateInternalGenericGatewayDetailsGatewayTypeEnum = map[string]CreateInternalGenericGatewayDetailsGatewayTypeEnum{
	"SERVICEGATEWAY":       CreateInternalGenericGatewayDetailsGatewayTypeServicegateway,
	"NATGATEWAY":           CreateInternalGenericGatewayDetailsGatewayTypeNatgateway,
	"PRIVATEACCESSGATEWAY": CreateInternalGenericGatewayDetailsGatewayTypePrivateaccessgateway,
}

// GetCreateInternalGenericGatewayDetailsGatewayTypeEnumValues Enumerates the set of values for CreateInternalGenericGatewayDetailsGatewayTypeEnum
func GetCreateInternalGenericGatewayDetailsGatewayTypeEnumValues() []CreateInternalGenericGatewayDetailsGatewayTypeEnum {
	values := make([]CreateInternalGenericGatewayDetailsGatewayTypeEnum, 0)
	for _, v := range mappingCreateInternalGenericGatewayDetailsGatewayTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateInternalGenericGatewayDetailsGatewayTypeEnumStringValues Enumerates the set of values in String for CreateInternalGenericGatewayDetailsGatewayTypeEnum
func GetCreateInternalGenericGatewayDetailsGatewayTypeEnumStringValues() []string {
	return []string{
		"SERVICEGATEWAY",
		"NATGATEWAY",
		"PRIVATEACCESSGATEWAY",
	}
}
