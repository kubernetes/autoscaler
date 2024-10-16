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

// NetworkSecurityGroup A *network security group* (NSG) provides virtual firewall rules for a specific set of
// Vnic in a VCN. Compare NSGs with SecurityList,
// which provide virtual firewall rules to all the VNICs in a *subnet*.
// A network security group consists of two items:
//   - The set of Vnic that all have the same security rule needs (for
//     example, a group of Compute instances all running the same application)
//   - A set of NSG SecurityRule that apply to the VNICs in the group
//
// After creating an NSG, you can add VNICs and security rules to it. For example, when you create
// an instance, you can specify one or more NSGs to add the instance to (see
// `CreateVnicDetails)`. Or you can add an existing
// instance to an NSG with `UpdateVnic`.
// To add security rules to an NSG, see
// `AddNetworkSecurityGroupSecurityRules`.
// To list the VNICs in an NSG, see
// `ListNetworkSecurityGroupVnics`.
// To list the security rules in an NSG, see
// `ListNetworkSecurityGroupSecurityRules`.
// For more information about network security groups, see
// `Network Security Groups (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/networksecuritygroups.htm)`.
// **Important:** Oracle Cloud Infrastructure Compute service images automatically include firewall rules (for example,
// Linux iptables, Windows firewall). If there are issues with some type of access to an instance,
// make sure all of the following are set correctly:
//   - Any security rules in any NSGs the instance's VNIC belongs to
//   - Any `SecurityList` associated with the instance's subnet
//   - The instance's OS firewall rules
//
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type NetworkSecurityGroup struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment the network security group is in.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the network security group.
	Id *string `mandatory:"true" json:"id"`

	// The network security group's current state.
	LifecycleState NetworkSecurityGroupLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the network security group was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the network security group's VCN.
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
}

func (m NetworkSecurityGroup) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m NetworkSecurityGroup) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingNetworkSecurityGroupLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetNetworkSecurityGroupLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// NetworkSecurityGroupLifecycleStateEnum Enum with underlying type: string
type NetworkSecurityGroupLifecycleStateEnum string

// Set of constants representing the allowable values for NetworkSecurityGroupLifecycleStateEnum
const (
	NetworkSecurityGroupLifecycleStateProvisioning NetworkSecurityGroupLifecycleStateEnum = "PROVISIONING"
	NetworkSecurityGroupLifecycleStateAvailable    NetworkSecurityGroupLifecycleStateEnum = "AVAILABLE"
	NetworkSecurityGroupLifecycleStateTerminating  NetworkSecurityGroupLifecycleStateEnum = "TERMINATING"
	NetworkSecurityGroupLifecycleStateTerminated   NetworkSecurityGroupLifecycleStateEnum = "TERMINATED"
)

var mappingNetworkSecurityGroupLifecycleStateEnum = map[string]NetworkSecurityGroupLifecycleStateEnum{
	"PROVISIONING": NetworkSecurityGroupLifecycleStateProvisioning,
	"AVAILABLE":    NetworkSecurityGroupLifecycleStateAvailable,
	"TERMINATING":  NetworkSecurityGroupLifecycleStateTerminating,
	"TERMINATED":   NetworkSecurityGroupLifecycleStateTerminated,
}

var mappingNetworkSecurityGroupLifecycleStateEnumLowerCase = map[string]NetworkSecurityGroupLifecycleStateEnum{
	"provisioning": NetworkSecurityGroupLifecycleStateProvisioning,
	"available":    NetworkSecurityGroupLifecycleStateAvailable,
	"terminating":  NetworkSecurityGroupLifecycleStateTerminating,
	"terminated":   NetworkSecurityGroupLifecycleStateTerminated,
}

// GetNetworkSecurityGroupLifecycleStateEnumValues Enumerates the set of values for NetworkSecurityGroupLifecycleStateEnum
func GetNetworkSecurityGroupLifecycleStateEnumValues() []NetworkSecurityGroupLifecycleStateEnum {
	values := make([]NetworkSecurityGroupLifecycleStateEnum, 0)
	for _, v := range mappingNetworkSecurityGroupLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetNetworkSecurityGroupLifecycleStateEnumStringValues Enumerates the set of values in String for NetworkSecurityGroupLifecycleStateEnum
func GetNetworkSecurityGroupLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingNetworkSecurityGroupLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingNetworkSecurityGroupLifecycleStateEnum(val string) (NetworkSecurityGroupLifecycleStateEnum, bool) {
	enum, ok := mappingNetworkSecurityGroupLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
