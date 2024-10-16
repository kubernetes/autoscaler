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

// NatGateway A NAT (Network Address Translation) gateway, which represents a router that lets instances
// without public IPs contact the public internet without exposing the instance to inbound
// internet traffic. For more information, see
// NAT Gateway (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/NATgateway.htm).
// To use any of the API operations, you must be authorized in an
// IAM policy. If you are not authorized, talk to an
// administrator. If you are an administrator who needs to write
// policies to give users access, see Getting Started with
// Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type NatGateway struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment that contains
	// the NAT gateway.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the
	// NAT gateway.
	Id *string `mandatory:"true" json:"id"`

	// Whether the NAT gateway blocks traffic through it. The default is `false`.
	// Example: `true`
	BlockTraffic *bool `mandatory:"true" json:"blockTraffic"`

	// The NAT gateway's current state.
	LifecycleState NatGatewayLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The IP address associated with the NAT gateway.
	NatIp *string `mandatory:"true" json:"natIp"`

	// The date and time the NAT gateway was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN the NAT gateway
	// belongs to.
	VcnId *string `mandatory:"true" json:"vcnId"`

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

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the public IP address associated with the NAT gateway.
	PublicIpId *string `mandatory:"false" json:"publicIpId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the route table used by the NAT gateway.
	// If you don't specify a route table here, the NAT gateway is created without an associated route
	// table. The Networking service does NOT automatically associate the attached VCN's default route table
	// with the NAT gateway.
	RouteTableId *string `mandatory:"false" json:"routeTableId"`
}

func (m NatGateway) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m NatGateway) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingNatGatewayLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetNatGatewayLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// NatGatewayLifecycleStateEnum Enum with underlying type: string
type NatGatewayLifecycleStateEnum string

// Set of constants representing the allowable values for NatGatewayLifecycleStateEnum
const (
	NatGatewayLifecycleStateProvisioning NatGatewayLifecycleStateEnum = "PROVISIONING"
	NatGatewayLifecycleStateAvailable    NatGatewayLifecycleStateEnum = "AVAILABLE"
	NatGatewayLifecycleStateTerminating  NatGatewayLifecycleStateEnum = "TERMINATING"
	NatGatewayLifecycleStateTerminated   NatGatewayLifecycleStateEnum = "TERMINATED"
)

var mappingNatGatewayLifecycleStateEnum = map[string]NatGatewayLifecycleStateEnum{
	"PROVISIONING": NatGatewayLifecycleStateProvisioning,
	"AVAILABLE":    NatGatewayLifecycleStateAvailable,
	"TERMINATING":  NatGatewayLifecycleStateTerminating,
	"TERMINATED":   NatGatewayLifecycleStateTerminated,
}

var mappingNatGatewayLifecycleStateEnumLowerCase = map[string]NatGatewayLifecycleStateEnum{
	"provisioning": NatGatewayLifecycleStateProvisioning,
	"available":    NatGatewayLifecycleStateAvailable,
	"terminating":  NatGatewayLifecycleStateTerminating,
	"terminated":   NatGatewayLifecycleStateTerminated,
}

// GetNatGatewayLifecycleStateEnumValues Enumerates the set of values for NatGatewayLifecycleStateEnum
func GetNatGatewayLifecycleStateEnumValues() []NatGatewayLifecycleStateEnum {
	values := make([]NatGatewayLifecycleStateEnum, 0)
	for _, v := range mappingNatGatewayLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetNatGatewayLifecycleStateEnumStringValues Enumerates the set of values in String for NatGatewayLifecycleStateEnum
func GetNatGatewayLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingNatGatewayLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingNatGatewayLifecycleStateEnum(val string) (NatGatewayLifecycleStateEnum, bool) {
	enum, ok := mappingNatGatewayLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
