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

// InternalVnic This is a vnic type only used in operations with overlay customers and RCE.
// It defines additonal properties: isManaged, resourceType, resourceId, isBMVnic, isGarpEnabled and isServiceVnic
type InternalVnic struct {

	// The VNIC's availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the VNIC.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VNIC.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the VNIC.
	LifecycleState InternalVnicLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Indicates if the VNIC is managed by a internal partner team. And customer is not allowed
	// the perform update/delete operations on it directly.
	// Defaults to `False`
	IsManaged *bool `mandatory:"false" json:"isManaged"`

	// Type of the customer visible upstream resource that the VNIC is associated with. This property can be
	// exposed to customers as part of API to list members of a network security group.
	// For example, it can be set as,
	//  - `loadbalancer` if corresponding resourceId is a loadbalancer instance's OCID
	//  - `dbsystem` if corresponding resourceId is a dbsystem instance's OCID
	// Note that the partner team creating/managing the VNIC is owner of this metadata.
	// type:
	ResourceType *string `mandatory:"false" json:"resourceType"`

	// ID of the customer visible upstream resource that the VNIC is associated with. This property is
	// exposed to customers as part of API to list members of a network security group.
	// For example, if the VNIC is associated with a loadbalancer or dbsystem instance, then it needs
	// to be set to corresponding customer visible loadbalancer or dbsystem instance OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
	// Note that the partner team creating/managing the VNIC is owner of this metadata.
	ResourceId *string `mandatory:"false" json:"resourceId"`

	// Indicates if the VNIC is associated with (and will be attached to) a BM instance.
	IsBmVnic *bool `mandatory:"false" json:"isBmVnic"`

	// Indicates if the VNIC is a service vnic.
	IsServiceVnic *bool `mandatory:"false" json:"isServiceVnic"`

	// Indicates if this VNIC can issue GARP requests. False by default.
	IsGarpEnabled *bool `mandatory:"false" json:"isGarpEnabled"`

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

	// The hostname for the VNIC's primary private IP. Used for DNS. The value is the hostname
	// portion of the primary private IP's fully qualified domain name (FQDN)
	// (for example, `bminstance-1` in FQDN `bminstance-1.subnet123.vcn1.oraclevcn.com`).
	// Must be unique across all VNICs in the subnet and comply with
	// RFC 952 (https://tools.ietf.org/html/rfc952) and
	// RFC 1123 (https://tools.ietf.org/html/rfc1123).
	// For more information, see
	// DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm).
	// Example: `bminstance-1`
	HostnameLabel *string `mandatory:"false" json:"hostnameLabel"`

	// Whether the VNIC is the primary VNIC (the VNIC that is automatically created
	// and attached during instance launch).
	IsPrimary *bool `mandatory:"false" json:"isPrimary"`

	// The MAC address of the VNIC.
	// Example: `00:00:00:00:00:01`
	MacAddress *string `mandatory:"false" json:"macAddress"`

	// A list of the OCIDs of the network security groups that the VNIC belongs to. For more
	// information about NSGs, see
	// NetworkSecurityGroup.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// The private IP address of the primary `privateIp` object on the VNIC.
	// The address is within the CIDR of the VNIC's subnet.
	// **Note: ** This is null if the VNIC is in a subnet that has `isLearningEnabled` = `true`.
	// Example: `10.0.3.3`
	PrivateIp *string `mandatory:"false" json:"privateIp"`

	// The public IP address of the VNIC, if one is assigned.
	PublicIp *string `mandatory:"false" json:"publicIp"`

	// Whether the source/destination check is disabled on the VNIC.
	// Defaults to `false`, which means the check is performed. For information
	// about why you would skip the source/destination check, see
	// Using a Private IP as a Route Target (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm#privateip).
	// Example: `true`
	SkipSourceDestCheck *bool `mandatory:"false" json:"skipSourceDestCheck"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet the VNIC is in.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// The date and time the VNIC was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Indicates that the Cavium does not enforce Internet ingress/egress throttling.
	// Only present if explicitly assigned upon VNIC creation.
	BypassInternetThrottle *bool `mandatory:"false" json:"bypassInternetThrottle"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VLAN that the VNIC belongs to.
	VlanId *string `mandatory:"false" json:"vlanId"`

	// Indicates if a floating IP is mapped to the VNIC.
	HasFloatingIp *bool `mandatory:"false" json:"hasFloatingIp"`

	// Indicates if the VNIC is bridge which means cavium will ARP for its MAC address.
	IsBridgeVnic *bool `mandatory:"false" json:"isBridgeVnic"`

	// Indicates if MAC learning is enabled for the VNIC.
	IsMacLearningEnabled *bool `mandatory:"false" json:"isMacLearningEnabled"`

	// Indicates if generation of NAT IP should be skipped if the associated VNIC has this flag set to true.
	IsNatIpAllocationDisabled *bool `mandatory:"false" json:"isNatIpAllocationDisabled"`

	// Indicates if private IP creation should be blocked for external customers. Default to false.
	// For example, Exadata team can create private IP through internal api. External customers who call public api
	// are prohibited to add private IP to Exadata node.
	IsPrivateIpCreationBlocked *bool `mandatory:"false" json:"isPrivateIpCreationBlocked"`

	// ID of the entity owning the VNIC. This is passed in the create vnic call.
	// If none is passed and if there is an attachment then the attached instanceId is the ownerId.
	OwnerId *string `mandatory:"false" json:"ownerId"`

	// floating private IPs attached to this VNIC
	FloatingPrivateIPs []FloatingIpInfo `mandatory:"false" json:"floatingPrivateIPs"`

	// The CIDR of the subnet the VNIC belongs to
	SubnetCidr *string `mandatory:"false" json:"subnetCidr"`

	// The IP address of the VNIC's subnet's virtual router
	VirtualRouterIp *string `mandatory:"false" json:"virtualRouterIp"`

	// The CIDR IPv6 address block of the Subnet. The CIDR length is always /64.
	// Example: `2001:0db8:0123:4567::/64`
	Ipv6CidrBlock *string `mandatory:"false" json:"ipv6CidrBlock"`

	// The IPv6 address of the virtual router.
	// Example: `2001:0db8:0123:4567:89ab:cdef:1234:5678`
	Ipv6VirtualRouterIp *string `mandatory:"false" json:"ipv6VirtualRouterIp"`
}

func (m InternalVnic) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InternalVnic) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingInternalVnicLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInternalVnicLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InternalVnicLifecycleStateEnum Enum with underlying type: string
type InternalVnicLifecycleStateEnum string

// Set of constants representing the allowable values for InternalVnicLifecycleStateEnum
const (
	InternalVnicLifecycleStateProvisioning InternalVnicLifecycleStateEnum = "PROVISIONING"
	InternalVnicLifecycleStateAvailable    InternalVnicLifecycleStateEnum = "AVAILABLE"
	InternalVnicLifecycleStateTerminating  InternalVnicLifecycleStateEnum = "TERMINATING"
	InternalVnicLifecycleStateTerminated   InternalVnicLifecycleStateEnum = "TERMINATED"
)

var mappingInternalVnicLifecycleStateEnum = map[string]InternalVnicLifecycleStateEnum{
	"PROVISIONING": InternalVnicLifecycleStateProvisioning,
	"AVAILABLE":    InternalVnicLifecycleStateAvailable,
	"TERMINATING":  InternalVnicLifecycleStateTerminating,
	"TERMINATED":   InternalVnicLifecycleStateTerminated,
}

// GetInternalVnicLifecycleStateEnumValues Enumerates the set of values for InternalVnicLifecycleStateEnum
func GetInternalVnicLifecycleStateEnumValues() []InternalVnicLifecycleStateEnum {
	values := make([]InternalVnicLifecycleStateEnum, 0)
	for _, v := range mappingInternalVnicLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalVnicLifecycleStateEnumStringValues Enumerates the set of values in String for InternalVnicLifecycleStateEnum
func GetInternalVnicLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}
