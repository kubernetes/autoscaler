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

// CreateInternalVnicDetails This structure is used when creating vnic for internal clients.
// For more information about VNICs, see
// Virtual Network Interface Cards (VNICs) (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingVNICs.htm).
type CreateInternalVnicDetails struct {

	// Whether the VNIC should be assigned a public IP address. Defaults to whether
	// the subnet is public or private. If not set and the VNIC is being created
	// in a private subnet (that is, where `prohibitPublicIpOnVnic` = true in the
	// Subnet), then no public IP address is assigned.
	// If not set and the subnet is public (`prohibitPublicIpOnVnic` = false), then
	// a public IP address is assigned. If set to true and
	// `prohibitPublicIpOnVnic` = true, an error is returned.
	// **Note:** This public IP address is associated with the primary private IP
	// on the VNIC. For more information, see
	// IP Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingIPaddresses.htm).
	// **Note:** There's a limit to the number of PublicIp
	// a VNIC or instance can have. If you try to create a secondary VNIC
	// with an assigned public IP for an instance that has already
	// reached its public IP limit, an error is returned. For information
	// about the public IP limits, see
	// Public IP Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingpublicIPs.htm).
	// Example: `false`
	AssignPublicIp *bool `mandatory:"false" json:"assignPublicIp"`

	// The availability domain of the instance.
	// Availability domain can not be provided if isServiceVnic is true, it is required otherwise.
	//   Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// Not for general use!
	// Contact sic_vcn_us_grp@oracle.com before setting this flag.
	// Indicates that the Cavium should not enforce Internet ingress/egress throttling.
	// Defaults to `false`, in which case we do enforce that throttling.
	// At least one of subnetId OR the vlanId are required
	BypassInternetThrottle *bool `mandatory:"false" json:"bypassInternetThrottle"`

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
	// The value appears in the Vnic object and also the
	// PrivateIp object returned by
	// ListPrivateIps and
	// GetPrivateIp.
	// For more information, see
	// DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm).
	// When launching an instance, use this `hostnameLabel` instead
	// of the deprecated `hostnameLabel` in
	// LaunchInstanceDetails.
	// If you provide both, the values must match.
	// Example: `bminstance-1`
	HostnameLabel *string `mandatory:"false" json:"hostnameLabel"`

	// Indicates if the VNIC is associated with (and will be attached to) a BM instance.
	IsBmVnic *bool `mandatory:"false" json:"isBmVnic"`

	// Indicates if the VNIC is bridge which means cavium will ARP for its MAC address.
	IsBridgeVnic *bool `mandatory:"false" json:"isBridgeVnic"`

	// Indicates if this VNIC can issue GARP requests. False by default.
	IsGarpEnabled *bool `mandatory:"false" json:"isGarpEnabled"`

	// Indicates if MAC learning is enabled for the VNIC. The default is `false`.
	// When this flag is enabled, then VCN CP does not allocate MAC address,
	// hence MAC address will be set as null as part of the VNIC that is returned.
	IsMacLearningEnabled *bool `mandatory:"false" json:"isMacLearningEnabled"`

	// Indicates if the VNIC is managed by a internal partner team. And customer is not allowed
	// the perform update/delete operations on it directly.
	// Defaults to `False`
	IsManaged *bool `mandatory:"false" json:"isManaged"`

	// Indicates if the VNIC is primary which means it cannot be detached.
	IsPrimary *bool `mandatory:"false" json:"isPrimary"`

	// Indicates if the VNIC is a service vnic.
	IsServiceVnic *bool `mandatory:"false" json:"isServiceVnic"`

	// Only provided when no publicIpPoolId is specified.
	InternalPoolName CreateInternalVnicDetailsInternalPoolNameEnum `mandatory:"false" json:"internalPoolName,omitempty"`

	// The overlay MAC address of the instance
	MacAddress *string `mandatory:"false" json:"macAddress"`

	// A list of the OCIDs of the network security groups (NSGs) to add the VNIC to. For more
	// information about NSGs, see
	// NetworkSecurityGroup.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// ID of the entity owning the VNIC. This is passed in the create vnic call.
	// If none is passed and if there is an attachment then the attached instanceId is the ownerId.
	OwnerId *string `mandatory:"false" json:"ownerId"`

	// A private IP address of your choice to assign to the VNIC. Must be an
	// available IP address within the subnet's CIDR. If you don't specify a
	// value, Oracle automatically assigns a private IP address from the subnet.
	// This is the VNIC's *primary* private IP address. The value appears in
	// the Vnic object and also the
	// PrivateIp object returned by
	// ListPrivateIps and
	// GetPrivateIp.
	// Example: `10.0.3.3`
	PrivateIp *string `mandatory:"false" json:"privateIp"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the pool object created by the current tenancy
	PublicIpPoolId *string `mandatory:"false" json:"publicIpPoolId"`

	// ID of the customer visible upstream resource that the VNIC is associated with. This property is
	// exposed to customers as part of API to list members of a network security group.
	// For example, if the VNIC is associated with a loadbalancer or dbsystem instance, then it needs
	// to be set to corresponding customer visible loadbalancer or dbsystem instance OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
	// Note that the partner team creating/managing the VNIC is owner of this metadata.
	ResourceId *string `mandatory:"false" json:"resourceId"`

	// Type of the customer visible upstream resource that the VNIC is associated with. This property can be
	// exposed to customers as part of API to list members of a network security group.
	// For example, it can be set as,
	//  - `loadbalancer` if corresponding resourceId is a loadbalancer instance's OCID
	//  - `dbsystem` if corresponding resourceId is a dbsystem instance's OCID
	// Note that the partner team creating/managing the VNIC is owner of this metadata.
	ResourceType *string `mandatory:"false" json:"resourceType"`

	// Whether the source/destination check is disabled on the VNIC.
	// Defaults to `false`, which means the check is performed. For information
	// about why you would skip the source/destination check, see
	// Using a Private IP as a Route Target (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm#privateip).
	// Example: `true`
	SkipSourceDestCheck *bool `mandatory:"false" json:"skipSourceDestCheck"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the subnet to create the VNIC in. When launching an instance,
	// use this `subnetId` instead of the deprecated `subnetId` in
	// LaunchInstanceDetails.
	// At least one of them is required; if you provide both, the values must match.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// Usage of system tag keys. These predefined keys are scoped to namespaces.
	// Example: `{"orcl-cloud": {"free-tier-retained": "true"}}`
	SystemTags map[string]map[string]interface{} `mandatory:"false" json:"systemTags"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VLAN that the VNIC belongs to
	VlanId *string `mandatory:"false" json:"vlanId"`

	// ID of the compartment
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// Indicates if generation of NAT IP should be skipped if the associated VNIC has this flag set to true.
	IsNatIpAllocationDisabled *bool `mandatory:"false" json:"isNatIpAllocationDisabled"`

	// Indicates if private IP creation should be blocked for external customers. Default to false.
	// For example, Exadata team can create private IP through internal api. External customers who call public api
	// are prohibited to add private IP to Exadata node.
	IsPrivateIpCreationBlocked *bool `mandatory:"false" json:"isPrivateIpCreationBlocked"`

	// Indicates if the VNIC should get a public IP.
	HasPublicIp *bool `mandatory:"false" json:"hasPublicIp"`

	// Indicates if the VNIC is connected to VNIC as a Service. Defaults to `False`.
	IsVnicServiceVnic *bool `mandatory:"false" json:"isVnicServiceVnic"`

	// MPLS label to be used with a VNIC connected to VNIC as a Service. Required if isVnicServiceVnic is `True`.
	ServiceMplsLabel *int `mandatory:"false" json:"serviceMplsLabel"`

	// Type of service VNIC. Feature or use case that is creating this service VNIC. Used for forecasting, resource limits enforcement, and capacity management.
	ServiceVnicType CreateInternalVnicDetailsServiceVnicTypeEnum `mandatory:"false" json:"serviceVnicType,omitempty"`

	// Shape of VNIC that will be used to allocate resource in the data plane once the VNIC is attached
	VnicShape CreateInternalVnicDetailsVnicShapeEnum `mandatory:"false" json:"vnicShape,omitempty"`
}

func (m CreateInternalVnicDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateInternalVnicDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingCreateInternalVnicDetailsInternalPoolNameEnum[string(m.InternalPoolName)]; !ok && m.InternalPoolName != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for InternalPoolName: %s. Supported values are: %s.", m.InternalPoolName, strings.Join(GetCreateInternalVnicDetailsInternalPoolNameEnumStringValues(), ",")))
	}
	if _, ok := mappingCreateInternalVnicDetailsServiceVnicTypeEnum[string(m.ServiceVnicType)]; !ok && m.ServiceVnicType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ServiceVnicType: %s. Supported values are: %s.", m.ServiceVnicType, strings.Join(GetCreateInternalVnicDetailsServiceVnicTypeEnumStringValues(), ",")))
	}
	if _, ok := mappingCreateInternalVnicDetailsVnicShapeEnum[string(m.VnicShape)]; !ok && m.VnicShape != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for VnicShape: %s. Supported values are: %s.", m.VnicShape, strings.Join(GetCreateInternalVnicDetailsVnicShapeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CreateInternalVnicDetailsInternalPoolNameEnum Enum with underlying type: string
type CreateInternalVnicDetailsInternalPoolNameEnum string

// Set of constants representing the allowable values for CreateInternalVnicDetailsInternalPoolNameEnum
const (
	CreateInternalVnicDetailsInternalPoolNameExternal   CreateInternalVnicDetailsInternalPoolNameEnum = "EXTERNAL"
	CreateInternalVnicDetailsInternalPoolNameSociEgress CreateInternalVnicDetailsInternalPoolNameEnum = "SOCI_EGRESS"
)

var mappingCreateInternalVnicDetailsInternalPoolNameEnum = map[string]CreateInternalVnicDetailsInternalPoolNameEnum{
	"EXTERNAL":    CreateInternalVnicDetailsInternalPoolNameExternal,
	"SOCI_EGRESS": CreateInternalVnicDetailsInternalPoolNameSociEgress,
}

// GetCreateInternalVnicDetailsInternalPoolNameEnumValues Enumerates the set of values for CreateInternalVnicDetailsInternalPoolNameEnum
func GetCreateInternalVnicDetailsInternalPoolNameEnumValues() []CreateInternalVnicDetailsInternalPoolNameEnum {
	values := make([]CreateInternalVnicDetailsInternalPoolNameEnum, 0)
	for _, v := range mappingCreateInternalVnicDetailsInternalPoolNameEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateInternalVnicDetailsInternalPoolNameEnumStringValues Enumerates the set of values in String for CreateInternalVnicDetailsInternalPoolNameEnum
func GetCreateInternalVnicDetailsInternalPoolNameEnumStringValues() []string {
	return []string{
		"EXTERNAL",
		"SOCI_EGRESS",
	}
}

// CreateInternalVnicDetailsServiceVnicTypeEnum Enum with underlying type: string
type CreateInternalVnicDetailsServiceVnicTypeEnum string

// Set of constants representing the allowable values for CreateInternalVnicDetailsServiceVnicTypeEnum
const (
	CreateInternalVnicDetailsServiceVnicTypePrivateEndpoint           CreateInternalVnicDetailsServiceVnicTypeEnum = "PRIVATE_ENDPOINT"
	CreateInternalVnicDetailsServiceVnicTypeReverseConnectionEndpoint CreateInternalVnicDetailsServiceVnicTypeEnum = "REVERSE_CONNECTION_ENDPOINT"
	CreateInternalVnicDetailsServiceVnicTypeRealVirtualRouter         CreateInternalVnicDetailsServiceVnicTypeEnum = "REAL_VIRTUAL_ROUTER"
	CreateInternalVnicDetailsServiceVnicTypePrivateDnsEndpoint        CreateInternalVnicDetailsServiceVnicTypeEnum = "PRIVATE_DNS_ENDPOINT"
)

var mappingCreateInternalVnicDetailsServiceVnicTypeEnum = map[string]CreateInternalVnicDetailsServiceVnicTypeEnum{
	"PRIVATE_ENDPOINT":            CreateInternalVnicDetailsServiceVnicTypePrivateEndpoint,
	"REVERSE_CONNECTION_ENDPOINT": CreateInternalVnicDetailsServiceVnicTypeReverseConnectionEndpoint,
	"REAL_VIRTUAL_ROUTER":         CreateInternalVnicDetailsServiceVnicTypeRealVirtualRouter,
	"PRIVATE_DNS_ENDPOINT":        CreateInternalVnicDetailsServiceVnicTypePrivateDnsEndpoint,
}

// GetCreateInternalVnicDetailsServiceVnicTypeEnumValues Enumerates the set of values for CreateInternalVnicDetailsServiceVnicTypeEnum
func GetCreateInternalVnicDetailsServiceVnicTypeEnumValues() []CreateInternalVnicDetailsServiceVnicTypeEnum {
	values := make([]CreateInternalVnicDetailsServiceVnicTypeEnum, 0)
	for _, v := range mappingCreateInternalVnicDetailsServiceVnicTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateInternalVnicDetailsServiceVnicTypeEnumStringValues Enumerates the set of values in String for CreateInternalVnicDetailsServiceVnicTypeEnum
func GetCreateInternalVnicDetailsServiceVnicTypeEnumStringValues() []string {
	return []string{
		"PRIVATE_ENDPOINT",
		"REVERSE_CONNECTION_ENDPOINT",
		"REAL_VIRTUAL_ROUTER",
		"PRIVATE_DNS_ENDPOINT",
	}
}

// CreateInternalVnicDetailsVnicShapeEnum Enum with underlying type: string
type CreateInternalVnicDetailsVnicShapeEnum string

// Set of constants representing the allowable values for CreateInternalVnicDetailsVnicShapeEnum
const (
	CreateInternalVnicDetailsVnicShapeDynamic                    CreateInternalVnicDetailsVnicShapeEnum = "DYNAMIC"
	CreateInternalVnicDetailsVnicShapeFixed0040                  CreateInternalVnicDetailsVnicShapeEnum = "FIXED0040"
	CreateInternalVnicDetailsVnicShapeFixed0060                  CreateInternalVnicDetailsVnicShapeEnum = "FIXED0060"
	CreateInternalVnicDetailsVnicShapeFixed0060Psm               CreateInternalVnicDetailsVnicShapeEnum = "FIXED0060_PSM"
	CreateInternalVnicDetailsVnicShapeFixed0100                  CreateInternalVnicDetailsVnicShapeEnum = "FIXED0100"
	CreateInternalVnicDetailsVnicShapeFixed0120                  CreateInternalVnicDetailsVnicShapeEnum = "FIXED0120"
	CreateInternalVnicDetailsVnicShapeFixed01202x                CreateInternalVnicDetailsVnicShapeEnum = "FIXED0120_2X"
	CreateInternalVnicDetailsVnicShapeFixed0200                  CreateInternalVnicDetailsVnicShapeEnum = "FIXED0200"
	CreateInternalVnicDetailsVnicShapeFixed0240                  CreateInternalVnicDetailsVnicShapeEnum = "FIXED0240"
	CreateInternalVnicDetailsVnicShapeFixed0480                  CreateInternalVnicDetailsVnicShapeEnum = "FIXED0480"
	CreateInternalVnicDetailsVnicShapeEntirehost                 CreateInternalVnicDetailsVnicShapeEnum = "ENTIREHOST"
	CreateInternalVnicDetailsVnicShapeDynamic25g                 CreateInternalVnicDetailsVnicShapeEnum = "DYNAMIC_25G"
	CreateInternalVnicDetailsVnicShapeFixed004025g               CreateInternalVnicDetailsVnicShapeEnum = "FIXED0040_25G"
	CreateInternalVnicDetailsVnicShapeFixed010025g               CreateInternalVnicDetailsVnicShapeEnum = "FIXED0100_25G"
	CreateInternalVnicDetailsVnicShapeFixed020025g               CreateInternalVnicDetailsVnicShapeEnum = "FIXED0200_25G"
	CreateInternalVnicDetailsVnicShapeFixed040025g               CreateInternalVnicDetailsVnicShapeEnum = "FIXED0400_25G"
	CreateInternalVnicDetailsVnicShapeFixed080025g               CreateInternalVnicDetailsVnicShapeEnum = "FIXED0800_25G"
	CreateInternalVnicDetailsVnicShapeFixed160025g               CreateInternalVnicDetailsVnicShapeEnum = "FIXED1600_25G"
	CreateInternalVnicDetailsVnicShapeFixed240025g               CreateInternalVnicDetailsVnicShapeEnum = "FIXED2400_25G"
	CreateInternalVnicDetailsVnicShapeEntirehost25g              CreateInternalVnicDetailsVnicShapeEnum = "ENTIREHOST_25G"
	CreateInternalVnicDetailsVnicShapeDynamicE125g               CreateInternalVnicDetailsVnicShapeEnum = "DYNAMIC_E1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0040E125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0040_E1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0070E125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0070_E1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0140E125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0140_E1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0280E125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0280_E1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0560E125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0560_E1_25G"
	CreateInternalVnicDetailsVnicShapeFixed1120E125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1120_E1_25G"
	CreateInternalVnicDetailsVnicShapeFixed1680E125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1680_E1_25G"
	CreateInternalVnicDetailsVnicShapeEntirehostE125g            CreateInternalVnicDetailsVnicShapeEnum = "ENTIREHOST_E1_25G"
	CreateInternalVnicDetailsVnicShapeDynamicB125g               CreateInternalVnicDetailsVnicShapeEnum = "DYNAMIC_B1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0040B125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0040_B1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0060B125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0060_B1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0120B125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0120_B1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0240B125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0240_B1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0480B125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0480_B1_25G"
	CreateInternalVnicDetailsVnicShapeFixed0960B125g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0960_B1_25G"
	CreateInternalVnicDetailsVnicShapeEntirehostB125g            CreateInternalVnicDetailsVnicShapeEnum = "ENTIREHOST_B1_25G"
	CreateInternalVnicDetailsVnicShapeMicroVmFixed0048E125g      CreateInternalVnicDetailsVnicShapeEnum = "MICRO_VM_FIXED0048_E1_25G"
	CreateInternalVnicDetailsVnicShapeMicroLbFixed0001E125g      CreateInternalVnicDetailsVnicShapeEnum = "MICRO_LB_FIXED0001_E1_25G"
	CreateInternalVnicDetailsVnicShapeVnicaasFixed0200           CreateInternalVnicDetailsVnicShapeEnum = "VNICAAS_FIXED0200"
	CreateInternalVnicDetailsVnicShapeVnicaasFixed0400           CreateInternalVnicDetailsVnicShapeEnum = "VNICAAS_FIXED0400"
	CreateInternalVnicDetailsVnicShapeVnicaasFixed0700           CreateInternalVnicDetailsVnicShapeEnum = "VNICAAS_FIXED0700"
	CreateInternalVnicDetailsVnicShapeVnicaasNlbApproved10g      CreateInternalVnicDetailsVnicShapeEnum = "VNICAAS_NLB_APPROVED_10G"
	CreateInternalVnicDetailsVnicShapeVnicaasNlbApproved25g      CreateInternalVnicDetailsVnicShapeEnum = "VNICAAS_NLB_APPROVED_25G"
	CreateInternalVnicDetailsVnicShapeVnicaasTelesis25g          CreateInternalVnicDetailsVnicShapeEnum = "VNICAAS_TELESIS_25G"
	CreateInternalVnicDetailsVnicShapeVnicaasTelesis10g          CreateInternalVnicDetailsVnicShapeEnum = "VNICAAS_TELESIS_10G"
	CreateInternalVnicDetailsVnicShapeVnicaasAmbassadorFixed0100 CreateInternalVnicDetailsVnicShapeEnum = "VNICAAS_AMBASSADOR_FIXED0100"
	CreateInternalVnicDetailsVnicShapeVnicaasPrivatedns          CreateInternalVnicDetailsVnicShapeEnum = "VNICAAS_PRIVATEDNS"
	CreateInternalVnicDetailsVnicShapeVnicaasFwaas               CreateInternalVnicDetailsVnicShapeEnum = "VNICAAS_FWAAS"
	CreateInternalVnicDetailsVnicShapeDynamicE350g               CreateInternalVnicDetailsVnicShapeEnum = "DYNAMIC_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed0040E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0040_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed0100E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0100_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed0200E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0200_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed0300E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0300_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed0400E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0400_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed0500E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0500_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed0600E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0600_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed0700E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0700_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed0800E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0800_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed0900E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0900_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed1000E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1000_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed1100E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1100_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed1200E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1200_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed1300E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1300_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed1400E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1400_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed1500E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1500_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed1600E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1600_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed1700E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1700_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed1800E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1800_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed1900E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1900_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed2000E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2000_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed2100E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2100_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed2200E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2200_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed2300E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2300_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed2400E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2400_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed2500E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2500_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed2600E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2600_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed2700E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2700_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed2800E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2800_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed2900E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2900_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed3000E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3000_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed3100E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3100_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed3200E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3200_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed3300E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3300_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed3400E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3400_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed3500E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3500_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed3600E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3600_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed3700E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3700_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed3800E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3800_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed3900E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3900_E3_50G"
	CreateInternalVnicDetailsVnicShapeFixed4000E350g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED4000_E3_50G"
	CreateInternalVnicDetailsVnicShapeEntirehostE350g            CreateInternalVnicDetailsVnicShapeEnum = "ENTIREHOST_E3_50G"
	CreateInternalVnicDetailsVnicShapeDynamicE450g               CreateInternalVnicDetailsVnicShapeEnum = "DYNAMIC_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed0040E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0040_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed0100E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0100_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed0200E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0200_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed0300E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0300_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed0400E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0400_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed0500E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0500_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed0600E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0600_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed0700E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0700_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed0800E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0800_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed0900E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0900_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed1000E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1000_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed1100E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1100_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed1200E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1200_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed1300E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1300_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed1400E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1400_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed1500E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1500_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed1600E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1600_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed1700E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1700_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed1800E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1800_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed1900E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1900_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed2000E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2000_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed2100E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2100_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed2200E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2200_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed2300E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2300_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed2400E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2400_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed2500E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2500_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed2600E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2600_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed2700E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2700_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed2800E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2800_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed2900E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2900_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed3000E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3000_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed3100E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3100_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed3200E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3200_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed3300E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3300_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed3400E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3400_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed3500E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3500_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed3600E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3600_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed3700E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3700_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed3800E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3800_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed3900E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3900_E4_50G"
	CreateInternalVnicDetailsVnicShapeFixed4000E450g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED4000_E4_50G"
	CreateInternalVnicDetailsVnicShapeEntirehostE450g            CreateInternalVnicDetailsVnicShapeEnum = "ENTIREHOST_E4_50G"
	CreateInternalVnicDetailsVnicShapeMicroVmFixed0050E350g      CreateInternalVnicDetailsVnicShapeEnum = "MICRO_VM_FIXED0050_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0025E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0025_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0050E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0050_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0075E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0075_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0100E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0100_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0125E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0125_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0150E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0150_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0175E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0175_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0200E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0200_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0225E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0225_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0250E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0250_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0275E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0275_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0300E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0300_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0325E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0325_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0350E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0350_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0375E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0375_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0400E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0400_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0425E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0425_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0450E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0450_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0475E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0475_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0500E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0500_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0525E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0525_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0550E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0550_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0575E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0575_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0600E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0600_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0625E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0625_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0650E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0650_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0675E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0675_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0700E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0700_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0725E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0725_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0750E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0750_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0775E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0775_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0800E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0800_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0825E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0825_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0850E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0850_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0875E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0875_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0900E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0925E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0925_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0950E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0950_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0975E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0975_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1000E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1000_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1025E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1025_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1050E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1050_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1075E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1075_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1100E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1100_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1125E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1125_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1150E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1150_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1175E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1175_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1200E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1200_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1225E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1225_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1250E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1250_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1275E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1275_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1300E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1300_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1325E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1325_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1350E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1350_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1375E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1375_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1400E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1400_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1425E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1425_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1450E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1450_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1475E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1475_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1500E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1500_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1525E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1525_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1550E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1550_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1575E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1575_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1600E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1600_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1625E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1625_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1650E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1650_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1700E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1700_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1725E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1725_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1750E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1750_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1800E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1850E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1850_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1875E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1875_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1900E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1900_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1925E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1925_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1950E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1950_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2000E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2000_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2025E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2025_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2050E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2050_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2100E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2100_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2125E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2125_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2150E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2150_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2175E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2175_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2200E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2200_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2250E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2250_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2275E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2275_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2300E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2300_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2325E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2325_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2350E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2350_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2375E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2375_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2400E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2400_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2450E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2450_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2475E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2475_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2500E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2500_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2550E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2550_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2600E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2600_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2625E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2625_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2650E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2650_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2700E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2750E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2750_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2775E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2775_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2800E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2800_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2850E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2850_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2875E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2875_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2900E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2900_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2925E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2925_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2950E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2950_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2975E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2975_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3000E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3000_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3025E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3025_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3050E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3050_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3075E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3075_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3100E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3100_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3125E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3125_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3150E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3150_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3200E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3200_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3225E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3225_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3250E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3250_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3300E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3300_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3325E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3325_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3375E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3375_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3400E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3400_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3450E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3450_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3500E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3500_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3525E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3525_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3575E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3575_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3600E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3625E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3625_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3675E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3675_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3700E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3700_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3750E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3750_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3800E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3800_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3825E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3825_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3850E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3850_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3875E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3875_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3900E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3900_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3975E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3975_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4000E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4000_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4025E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4025_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4050E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4050_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4100E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4100_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4125E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4125_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4200E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4200_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4225E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4225_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4250E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4250_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4275E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4275_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4300E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4300_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4350E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4350_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4375E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4375_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4400E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4400_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4425E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4425_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4500E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4550E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4550_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4575E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4575_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4600E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4600_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4625E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4625_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4650E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4650_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4675E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4675_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4700E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4700_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4725E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4725_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4750E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4750_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4800E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4800_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4875E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4875_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4900E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4900_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4950E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4950_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed5000E350g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED5000_E3_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0025E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0025_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0050E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0050_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0075E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0075_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0100E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0100_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0125E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0125_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0150E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0150_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0175E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0175_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0200E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0200_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0225E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0225_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0250E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0250_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0275E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0275_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0300E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0300_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0325E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0325_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0350E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0350_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0375E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0375_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0400E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0400_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0425E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0425_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0450E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0450_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0475E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0475_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0500E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0500_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0525E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0525_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0550E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0550_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0575E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0575_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0600E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0600_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0625E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0625_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0650E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0650_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0675E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0675_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0700E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0700_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0725E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0725_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0750E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0750_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0775E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0775_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0800E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0800_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0825E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0825_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0850E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0850_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0875E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0875_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0900E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0925E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0925_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0950E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0950_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0975E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0975_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1000E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1000_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1025E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1025_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1050E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1050_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1075E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1075_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1100E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1100_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1125E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1125_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1150E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1150_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1175E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1175_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1200E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1200_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1225E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1225_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1250E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1250_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1275E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1275_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1300E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1300_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1325E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1325_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1350E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1350_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1375E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1375_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1400E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1400_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1425E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1425_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1450E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1450_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1475E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1475_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1500E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1500_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1525E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1525_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1550E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1550_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1575E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1575_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1600E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1600_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1625E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1625_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1650E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1650_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1700E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1700_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1725E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1725_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1750E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1750_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1800E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1850E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1850_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1875E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1875_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1900E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1900_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1925E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1925_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1950E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1950_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2000E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2000_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2025E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2025_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2050E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2050_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2100E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2100_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2125E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2125_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2150E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2150_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2175E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2175_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2200E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2200_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2250E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2250_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2275E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2275_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2300E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2300_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2325E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2325_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2350E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2350_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2375E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2375_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2400E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2400_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2450E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2450_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2475E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2475_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2500E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2500_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2550E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2550_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2600E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2600_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2625E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2625_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2650E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2650_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2700E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2750E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2750_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2775E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2775_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2800E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2800_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2850E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2850_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2875E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2875_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2900E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2900_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2925E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2925_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2950E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2950_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2975E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2975_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3000E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3000_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3025E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3025_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3050E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3050_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3075E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3075_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3100E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3100_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3125E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3125_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3150E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3150_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3200E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3200_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3225E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3225_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3250E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3250_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3300E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3300_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3325E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3325_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3375E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3375_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3400E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3400_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3450E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3450_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3500E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3500_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3525E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3525_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3575E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3575_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3600E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3625E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3625_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3675E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3675_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3700E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3700_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3750E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3750_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3800E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3800_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3825E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3825_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3850E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3850_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3875E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3875_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3900E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3900_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3975E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3975_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4000E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4000_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4025E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4025_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4050E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4050_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4100E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4100_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4125E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4125_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4200E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4200_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4225E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4225_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4250E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4250_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4275E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4275_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4300E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4300_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4350E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4350_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4375E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4375_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4400E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4400_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4425E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4425_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4500E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4550E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4550_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4575E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4575_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4600E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4600_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4625E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4625_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4650E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4650_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4675E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4675_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4700E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4700_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4725E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4725_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4750E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4750_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4800E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4800_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4875E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4875_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4900E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4900_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4950E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4950_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed5000E450g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED5000_E4_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0020A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0020_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0040A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0040_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0060A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0060_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0080A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0080_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0100A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0100_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0120A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0120_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0140A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0140_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0160A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0160_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0180A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0180_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0200A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0200_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0220A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0220_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0240A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0240_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0260A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0260_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0280A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0280_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0300A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0300_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0320A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0320_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0340A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0340_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0360A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0360_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0380A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0380_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0400A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0400_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0420A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0420_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0440A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0440_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0460A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0460_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0480A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0480_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0500A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0500_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0520A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0520_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0540A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0540_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0560A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0560_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0580A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0580_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0600A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0600_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0620A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0620_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0640A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0640_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0660A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0660_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0680A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0680_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0700A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0700_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0720A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0720_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0740A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0740_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0760A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0760_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0780A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0780_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0800A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0800_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0820A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0820_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0840A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0840_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0860A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0860_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0880A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0880_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0900A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0920A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0920_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0940A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0940_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0960A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0960_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0980A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0980_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1000A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1000_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1020A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1020_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1040A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1040_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1060A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1060_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1080A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1080_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1100A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1100_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1120A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1120_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1140A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1140_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1160A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1160_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1180A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1180_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1200A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1200_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1220A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1220_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1240A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1240_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1260A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1260_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1280A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1280_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1300A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1300_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1320A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1320_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1340A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1340_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1360A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1360_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1380A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1380_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1400A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1400_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1420A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1420_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1440A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1440_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1460A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1460_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1480A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1480_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1500A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1500_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1520A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1520_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1540A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1540_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1560A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1560_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1580A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1580_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1600A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1600_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1620A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1620_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1640A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1640_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1660A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1660_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1680A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1680_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1700A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1700_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1720A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1720_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1740A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1740_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1760A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1760_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1780A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1780_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1800A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1820A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1820_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1840A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1840_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1860A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1860_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1880A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1880_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1900A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1900_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1920A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1920_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1940A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1940_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1960A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1960_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1980A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1980_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2000A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2000_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2020A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2020_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2040A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2040_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2060A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2060_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2080A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2080_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2100A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2100_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2120A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2120_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2140A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2140_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2160A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2160_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2180A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2180_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2200A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2200_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2220A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2220_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2240A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2240_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2260A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2260_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2280A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2280_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2300A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2300_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2320A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2320_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2340A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2340_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2360A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2360_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2380A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2380_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2400A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2400_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2420A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2420_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2440A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2440_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2460A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2460_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2480A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2480_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2500A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2500_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2520A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2520_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2540A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2540_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2560A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2560_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2580A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2580_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2600A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2600_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2620A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2620_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2640A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2640_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2660A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2660_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2680A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2680_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2700A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2720A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2720_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2740A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2740_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2760A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2760_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2780A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2780_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2800A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2800_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2820A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2820_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2840A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2840_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2860A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2860_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2880A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2880_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2900A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2900_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2920A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2920_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2940A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2940_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2960A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2960_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2980A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2980_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3000A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3000_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3020A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3020_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3040A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3040_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3060A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3060_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3080A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3080_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3100A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3100_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3120A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3120_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3140A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3140_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3160A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3160_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3180A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3180_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3200A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3200_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3220A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3220_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3240A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3240_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3260A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3260_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3280A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3280_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3300A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3300_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3320A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3320_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3340A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3340_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3360A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3360_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3380A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3380_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3400A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3400_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3420A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3420_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3440A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3440_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3460A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3460_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3480A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3480_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3500A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3500_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3520A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3520_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3540A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3540_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3560A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3560_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3580A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3580_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3600A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3620A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3620_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3640A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3640_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3660A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3660_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3680A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3680_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3700A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3700_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3720A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3720_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3740A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3740_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3760A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3760_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3780A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3780_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3800A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3800_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3820A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3820_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3840A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3840_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3860A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3860_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3880A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3880_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3900A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3900_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3920A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3920_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3940A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3940_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3960A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3960_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3980A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3980_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4000A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4000_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4020A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4020_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4040A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4040_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4060A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4060_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4080A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4080_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4100A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4100_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4120A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4120_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4140A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4140_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4160A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4160_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4180A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4180_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4200A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4200_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4220A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4220_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4240A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4240_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4260A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4260_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4280A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4280_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4300A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4300_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4320A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4320_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4340A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4340_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4360A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4360_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4380A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4380_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4400A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4400_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4420A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4420_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4440A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4440_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4460A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4460_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4480A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4480_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4500A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4520A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4520_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4540A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4540_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4560A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4560_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4580A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4580_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4600A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4600_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4620A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4620_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4640A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4640_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4660A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4660_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4680A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4680_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4700A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4700_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4720A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4720_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4740A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4740_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4760A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4760_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4780A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4780_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4800A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4800_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4820A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4820_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4840A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4840_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4860A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4860_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4880A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4880_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4900A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4900_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4920A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4920_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4940A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4940_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4960A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4960_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4980A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4980_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed5000A150g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED5000_A1_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0090X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0090_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0180X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0180_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0270X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0270_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0360X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0360_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0450X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0450_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0540X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0540_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0630X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0630_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0720X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0720_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0810X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0810_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0900X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0900_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0990X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED0990_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1080X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1080_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1170X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1170_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1260X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1260_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1350X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1350_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1440X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1440_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1530X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1530_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1620X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1620_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1710X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1710_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1800X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1800_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1890X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1890_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1980X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED1980_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2070X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2070_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2160X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2160_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2250X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2250_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2340X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2340_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2430X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2430_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2520X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2520_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2610X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2610_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2700X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2700_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2790X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2790_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2880X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2880_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2970X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED2970_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3060X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3060_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3150X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3150_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3240X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3240_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3330X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3330_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3420X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3420_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3510X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3510_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3600X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3600_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3690X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3690_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3780X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3780_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3870X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3870_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3960X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED3960_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4050X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4050_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4140X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4140_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4230X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4230_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4320X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4320_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4410X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4410_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4500X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4500_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4590X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4590_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4680X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4680_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4770X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4770_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4860X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4860_X9_50G"
	CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4950X950g    CreateInternalVnicDetailsVnicShapeEnum = "SUBCORE_VM_FIXED4950_X9_50G"
	CreateInternalVnicDetailsVnicShapeDynamicA150g               CreateInternalVnicDetailsVnicShapeEnum = "DYNAMIC_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed0040A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0040_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed0100A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0100_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed0200A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0200_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed0300A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0300_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed0400A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0400_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed0500A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0500_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed0600A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0600_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed0700A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0700_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed0800A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0800_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed0900A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0900_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed1000A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1000_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed1100A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1100_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed1200A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1200_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed1300A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1300_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed1400A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1400_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed1500A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1500_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed1600A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1600_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed1700A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1700_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed1800A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1800_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed1900A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1900_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed2000A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2000_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed2100A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2100_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed2200A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2200_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed2300A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2300_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed2400A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2400_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed2500A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2500_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed2600A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2600_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed2700A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2700_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed2800A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2800_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed2900A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2900_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed3000A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3000_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed3100A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3100_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed3200A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3200_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed3300A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3300_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed3400A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3400_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed3500A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3500_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed3600A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3600_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed3700A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3700_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed3800A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3800_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed3900A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3900_A1_50G"
	CreateInternalVnicDetailsVnicShapeFixed4000A150g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED4000_A1_50G"
	CreateInternalVnicDetailsVnicShapeEntirehostA150g            CreateInternalVnicDetailsVnicShapeEnum = "ENTIREHOST_A1_50G"
	CreateInternalVnicDetailsVnicShapeDynamicX950g               CreateInternalVnicDetailsVnicShapeEnum = "DYNAMIC_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed0040X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0040_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed0400X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0400_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed0800X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED0800_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed1200X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1200_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed1600X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED1600_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed2000X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2000_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed2400X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2400_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed2800X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED2800_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed3200X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3200_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed3600X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED3600_X9_50G"
	CreateInternalVnicDetailsVnicShapeFixed4000X950g             CreateInternalVnicDetailsVnicShapeEnum = "FIXED4000_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed0100X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED0100_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed0200X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED0200_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed0300X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED0300_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed0400X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED0400_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed0500X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED0500_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed0600X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED0600_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed0700X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED0700_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed0800X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED0800_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed0900X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED0900_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed1000X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED1000_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed1100X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED1100_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed1200X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED1200_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed1300X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED1300_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed1400X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED1400_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed1500X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED1500_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed1600X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED1600_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed1700X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED1700_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed1800X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED1800_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed1900X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED1900_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed2000X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED2000_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed2100X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED2100_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed2200X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED2200_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed2300X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED2300_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed2400X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED2400_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed2500X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED2500_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed2600X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED2600_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed2700X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED2700_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed2800X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED2800_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed2900X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED2900_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed3000X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED3000_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed3100X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED3100_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed3200X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED3200_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed3300X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED3300_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed3400X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED3400_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed3500X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED3500_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed3600X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED3600_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed3700X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED3700_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed3800X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED3800_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed3900X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED3900_X9_50G"
	CreateInternalVnicDetailsVnicShapeStandardVmFixed4000X950g   CreateInternalVnicDetailsVnicShapeEnum = "STANDARD_VM_FIXED4000_X9_50G"
	CreateInternalVnicDetailsVnicShapeEntirehostX950g            CreateInternalVnicDetailsVnicShapeEnum = "ENTIREHOST_X9_50G"
)

var mappingCreateInternalVnicDetailsVnicShapeEnum = map[string]CreateInternalVnicDetailsVnicShapeEnum{
	"DYNAMIC":                      CreateInternalVnicDetailsVnicShapeDynamic,
	"FIXED0040":                    CreateInternalVnicDetailsVnicShapeFixed0040,
	"FIXED0060":                    CreateInternalVnicDetailsVnicShapeFixed0060,
	"FIXED0060_PSM":                CreateInternalVnicDetailsVnicShapeFixed0060Psm,
	"FIXED0100":                    CreateInternalVnicDetailsVnicShapeFixed0100,
	"FIXED0120":                    CreateInternalVnicDetailsVnicShapeFixed0120,
	"FIXED0120_2X":                 CreateInternalVnicDetailsVnicShapeFixed01202x,
	"FIXED0200":                    CreateInternalVnicDetailsVnicShapeFixed0200,
	"FIXED0240":                    CreateInternalVnicDetailsVnicShapeFixed0240,
	"FIXED0480":                    CreateInternalVnicDetailsVnicShapeFixed0480,
	"ENTIREHOST":                   CreateInternalVnicDetailsVnicShapeEntirehost,
	"DYNAMIC_25G":                  CreateInternalVnicDetailsVnicShapeDynamic25g,
	"FIXED0040_25G":                CreateInternalVnicDetailsVnicShapeFixed004025g,
	"FIXED0100_25G":                CreateInternalVnicDetailsVnicShapeFixed010025g,
	"FIXED0200_25G":                CreateInternalVnicDetailsVnicShapeFixed020025g,
	"FIXED0400_25G":                CreateInternalVnicDetailsVnicShapeFixed040025g,
	"FIXED0800_25G":                CreateInternalVnicDetailsVnicShapeFixed080025g,
	"FIXED1600_25G":                CreateInternalVnicDetailsVnicShapeFixed160025g,
	"FIXED2400_25G":                CreateInternalVnicDetailsVnicShapeFixed240025g,
	"ENTIREHOST_25G":               CreateInternalVnicDetailsVnicShapeEntirehost25g,
	"DYNAMIC_E1_25G":               CreateInternalVnicDetailsVnicShapeDynamicE125g,
	"FIXED0040_E1_25G":             CreateInternalVnicDetailsVnicShapeFixed0040E125g,
	"FIXED0070_E1_25G":             CreateInternalVnicDetailsVnicShapeFixed0070E125g,
	"FIXED0140_E1_25G":             CreateInternalVnicDetailsVnicShapeFixed0140E125g,
	"FIXED0280_E1_25G":             CreateInternalVnicDetailsVnicShapeFixed0280E125g,
	"FIXED0560_E1_25G":             CreateInternalVnicDetailsVnicShapeFixed0560E125g,
	"FIXED1120_E1_25G":             CreateInternalVnicDetailsVnicShapeFixed1120E125g,
	"FIXED1680_E1_25G":             CreateInternalVnicDetailsVnicShapeFixed1680E125g,
	"ENTIREHOST_E1_25G":            CreateInternalVnicDetailsVnicShapeEntirehostE125g,
	"DYNAMIC_B1_25G":               CreateInternalVnicDetailsVnicShapeDynamicB125g,
	"FIXED0040_B1_25G":             CreateInternalVnicDetailsVnicShapeFixed0040B125g,
	"FIXED0060_B1_25G":             CreateInternalVnicDetailsVnicShapeFixed0060B125g,
	"FIXED0120_B1_25G":             CreateInternalVnicDetailsVnicShapeFixed0120B125g,
	"FIXED0240_B1_25G":             CreateInternalVnicDetailsVnicShapeFixed0240B125g,
	"FIXED0480_B1_25G":             CreateInternalVnicDetailsVnicShapeFixed0480B125g,
	"FIXED0960_B1_25G":             CreateInternalVnicDetailsVnicShapeFixed0960B125g,
	"ENTIREHOST_B1_25G":            CreateInternalVnicDetailsVnicShapeEntirehostB125g,
	"MICRO_VM_FIXED0048_E1_25G":    CreateInternalVnicDetailsVnicShapeMicroVmFixed0048E125g,
	"MICRO_LB_FIXED0001_E1_25G":    CreateInternalVnicDetailsVnicShapeMicroLbFixed0001E125g,
	"VNICAAS_FIXED0200":            CreateInternalVnicDetailsVnicShapeVnicaasFixed0200,
	"VNICAAS_FIXED0400":            CreateInternalVnicDetailsVnicShapeVnicaasFixed0400,
	"VNICAAS_FIXED0700":            CreateInternalVnicDetailsVnicShapeVnicaasFixed0700,
	"VNICAAS_NLB_APPROVED_10G":     CreateInternalVnicDetailsVnicShapeVnicaasNlbApproved10g,
	"VNICAAS_NLB_APPROVED_25G":     CreateInternalVnicDetailsVnicShapeVnicaasNlbApproved25g,
	"VNICAAS_TELESIS_25G":          CreateInternalVnicDetailsVnicShapeVnicaasTelesis25g,
	"VNICAAS_TELESIS_10G":          CreateInternalVnicDetailsVnicShapeVnicaasTelesis10g,
	"VNICAAS_AMBASSADOR_FIXED0100": CreateInternalVnicDetailsVnicShapeVnicaasAmbassadorFixed0100,
	"VNICAAS_PRIVATEDNS":           CreateInternalVnicDetailsVnicShapeVnicaasPrivatedns,
	"VNICAAS_FWAAS":                CreateInternalVnicDetailsVnicShapeVnicaasFwaas,
	"DYNAMIC_E3_50G":               CreateInternalVnicDetailsVnicShapeDynamicE350g,
	"FIXED0040_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed0040E350g,
	"FIXED0100_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed0100E350g,
	"FIXED0200_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed0200E350g,
	"FIXED0300_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed0300E350g,
	"FIXED0400_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed0400E350g,
	"FIXED0500_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed0500E350g,
	"FIXED0600_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed0600E350g,
	"FIXED0700_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed0700E350g,
	"FIXED0800_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed0800E350g,
	"FIXED0900_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed0900E350g,
	"FIXED1000_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed1000E350g,
	"FIXED1100_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed1100E350g,
	"FIXED1200_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed1200E350g,
	"FIXED1300_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed1300E350g,
	"FIXED1400_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed1400E350g,
	"FIXED1500_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed1500E350g,
	"FIXED1600_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed1600E350g,
	"FIXED1700_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed1700E350g,
	"FIXED1800_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed1800E350g,
	"FIXED1900_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed1900E350g,
	"FIXED2000_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed2000E350g,
	"FIXED2100_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed2100E350g,
	"FIXED2200_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed2200E350g,
	"FIXED2300_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed2300E350g,
	"FIXED2400_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed2400E350g,
	"FIXED2500_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed2500E350g,
	"FIXED2600_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed2600E350g,
	"FIXED2700_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed2700E350g,
	"FIXED2800_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed2800E350g,
	"FIXED2900_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed2900E350g,
	"FIXED3000_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed3000E350g,
	"FIXED3100_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed3100E350g,
	"FIXED3200_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed3200E350g,
	"FIXED3300_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed3300E350g,
	"FIXED3400_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed3400E350g,
	"FIXED3500_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed3500E350g,
	"FIXED3600_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed3600E350g,
	"FIXED3700_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed3700E350g,
	"FIXED3800_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed3800E350g,
	"FIXED3900_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed3900E350g,
	"FIXED4000_E3_50G":             CreateInternalVnicDetailsVnicShapeFixed4000E350g,
	"ENTIREHOST_E3_50G":            CreateInternalVnicDetailsVnicShapeEntirehostE350g,
	"DYNAMIC_E4_50G":               CreateInternalVnicDetailsVnicShapeDynamicE450g,
	"FIXED0040_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed0040E450g,
	"FIXED0100_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed0100E450g,
	"FIXED0200_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed0200E450g,
	"FIXED0300_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed0300E450g,
	"FIXED0400_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed0400E450g,
	"FIXED0500_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed0500E450g,
	"FIXED0600_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed0600E450g,
	"FIXED0700_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed0700E450g,
	"FIXED0800_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed0800E450g,
	"FIXED0900_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed0900E450g,
	"FIXED1000_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed1000E450g,
	"FIXED1100_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed1100E450g,
	"FIXED1200_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed1200E450g,
	"FIXED1300_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed1300E450g,
	"FIXED1400_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed1400E450g,
	"FIXED1500_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed1500E450g,
	"FIXED1600_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed1600E450g,
	"FIXED1700_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed1700E450g,
	"FIXED1800_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed1800E450g,
	"FIXED1900_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed1900E450g,
	"FIXED2000_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed2000E450g,
	"FIXED2100_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed2100E450g,
	"FIXED2200_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed2200E450g,
	"FIXED2300_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed2300E450g,
	"FIXED2400_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed2400E450g,
	"FIXED2500_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed2500E450g,
	"FIXED2600_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed2600E450g,
	"FIXED2700_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed2700E450g,
	"FIXED2800_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed2800E450g,
	"FIXED2900_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed2900E450g,
	"FIXED3000_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed3000E450g,
	"FIXED3100_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed3100E450g,
	"FIXED3200_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed3200E450g,
	"FIXED3300_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed3300E450g,
	"FIXED3400_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed3400E450g,
	"FIXED3500_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed3500E450g,
	"FIXED3600_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed3600E450g,
	"FIXED3700_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed3700E450g,
	"FIXED3800_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed3800E450g,
	"FIXED3900_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed3900E450g,
	"FIXED4000_E4_50G":             CreateInternalVnicDetailsVnicShapeFixed4000E450g,
	"ENTIREHOST_E4_50G":            CreateInternalVnicDetailsVnicShapeEntirehostE450g,
	"MICRO_VM_FIXED0050_E3_50G":    CreateInternalVnicDetailsVnicShapeMicroVmFixed0050E350g,
	"SUBCORE_VM_FIXED0025_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0025E350g,
	"SUBCORE_VM_FIXED0050_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0050E350g,
	"SUBCORE_VM_FIXED0075_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0075E350g,
	"SUBCORE_VM_FIXED0100_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0100E350g,
	"SUBCORE_VM_FIXED0125_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0125E350g,
	"SUBCORE_VM_FIXED0150_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0150E350g,
	"SUBCORE_VM_FIXED0175_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0175E350g,
	"SUBCORE_VM_FIXED0200_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0200E350g,
	"SUBCORE_VM_FIXED0225_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0225E350g,
	"SUBCORE_VM_FIXED0250_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0250E350g,
	"SUBCORE_VM_FIXED0275_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0275E350g,
	"SUBCORE_VM_FIXED0300_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0300E350g,
	"SUBCORE_VM_FIXED0325_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0325E350g,
	"SUBCORE_VM_FIXED0350_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0350E350g,
	"SUBCORE_VM_FIXED0375_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0375E350g,
	"SUBCORE_VM_FIXED0400_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0400E350g,
	"SUBCORE_VM_FIXED0425_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0425E350g,
	"SUBCORE_VM_FIXED0450_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0450E350g,
	"SUBCORE_VM_FIXED0475_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0475E350g,
	"SUBCORE_VM_FIXED0500_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0500E350g,
	"SUBCORE_VM_FIXED0525_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0525E350g,
	"SUBCORE_VM_FIXED0550_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0550E350g,
	"SUBCORE_VM_FIXED0575_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0575E350g,
	"SUBCORE_VM_FIXED0600_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0600E350g,
	"SUBCORE_VM_FIXED0625_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0625E350g,
	"SUBCORE_VM_FIXED0650_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0650E350g,
	"SUBCORE_VM_FIXED0675_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0675E350g,
	"SUBCORE_VM_FIXED0700_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0700E350g,
	"SUBCORE_VM_FIXED0725_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0725E350g,
	"SUBCORE_VM_FIXED0750_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0750E350g,
	"SUBCORE_VM_FIXED0775_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0775E350g,
	"SUBCORE_VM_FIXED0800_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0800E350g,
	"SUBCORE_VM_FIXED0825_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0825E350g,
	"SUBCORE_VM_FIXED0850_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0850E350g,
	"SUBCORE_VM_FIXED0875_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0875E350g,
	"SUBCORE_VM_FIXED0900_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0900E350g,
	"SUBCORE_VM_FIXED0925_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0925E350g,
	"SUBCORE_VM_FIXED0950_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0950E350g,
	"SUBCORE_VM_FIXED0975_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0975E350g,
	"SUBCORE_VM_FIXED1000_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1000E350g,
	"SUBCORE_VM_FIXED1025_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1025E350g,
	"SUBCORE_VM_FIXED1050_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1050E350g,
	"SUBCORE_VM_FIXED1075_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1075E350g,
	"SUBCORE_VM_FIXED1100_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1100E350g,
	"SUBCORE_VM_FIXED1125_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1125E350g,
	"SUBCORE_VM_FIXED1150_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1150E350g,
	"SUBCORE_VM_FIXED1175_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1175E350g,
	"SUBCORE_VM_FIXED1200_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1200E350g,
	"SUBCORE_VM_FIXED1225_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1225E350g,
	"SUBCORE_VM_FIXED1250_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1250E350g,
	"SUBCORE_VM_FIXED1275_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1275E350g,
	"SUBCORE_VM_FIXED1300_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1300E350g,
	"SUBCORE_VM_FIXED1325_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1325E350g,
	"SUBCORE_VM_FIXED1350_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1350E350g,
	"SUBCORE_VM_FIXED1375_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1375E350g,
	"SUBCORE_VM_FIXED1400_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1400E350g,
	"SUBCORE_VM_FIXED1425_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1425E350g,
	"SUBCORE_VM_FIXED1450_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1450E350g,
	"SUBCORE_VM_FIXED1475_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1475E350g,
	"SUBCORE_VM_FIXED1500_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1500E350g,
	"SUBCORE_VM_FIXED1525_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1525E350g,
	"SUBCORE_VM_FIXED1550_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1550E350g,
	"SUBCORE_VM_FIXED1575_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1575E350g,
	"SUBCORE_VM_FIXED1600_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1600E350g,
	"SUBCORE_VM_FIXED1625_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1625E350g,
	"SUBCORE_VM_FIXED1650_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1650E350g,
	"SUBCORE_VM_FIXED1700_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1700E350g,
	"SUBCORE_VM_FIXED1725_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1725E350g,
	"SUBCORE_VM_FIXED1750_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1750E350g,
	"SUBCORE_VM_FIXED1800_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1800E350g,
	"SUBCORE_VM_FIXED1850_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1850E350g,
	"SUBCORE_VM_FIXED1875_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1875E350g,
	"SUBCORE_VM_FIXED1900_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1900E350g,
	"SUBCORE_VM_FIXED1925_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1925E350g,
	"SUBCORE_VM_FIXED1950_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1950E350g,
	"SUBCORE_VM_FIXED2000_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2000E350g,
	"SUBCORE_VM_FIXED2025_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2025E350g,
	"SUBCORE_VM_FIXED2050_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2050E350g,
	"SUBCORE_VM_FIXED2100_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2100E350g,
	"SUBCORE_VM_FIXED2125_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2125E350g,
	"SUBCORE_VM_FIXED2150_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2150E350g,
	"SUBCORE_VM_FIXED2175_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2175E350g,
	"SUBCORE_VM_FIXED2200_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2200E350g,
	"SUBCORE_VM_FIXED2250_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2250E350g,
	"SUBCORE_VM_FIXED2275_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2275E350g,
	"SUBCORE_VM_FIXED2300_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2300E350g,
	"SUBCORE_VM_FIXED2325_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2325E350g,
	"SUBCORE_VM_FIXED2350_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2350E350g,
	"SUBCORE_VM_FIXED2375_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2375E350g,
	"SUBCORE_VM_FIXED2400_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2400E350g,
	"SUBCORE_VM_FIXED2450_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2450E350g,
	"SUBCORE_VM_FIXED2475_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2475E350g,
	"SUBCORE_VM_FIXED2500_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2500E350g,
	"SUBCORE_VM_FIXED2550_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2550E350g,
	"SUBCORE_VM_FIXED2600_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2600E350g,
	"SUBCORE_VM_FIXED2625_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2625E350g,
	"SUBCORE_VM_FIXED2650_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2650E350g,
	"SUBCORE_VM_FIXED2700_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2700E350g,
	"SUBCORE_VM_FIXED2750_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2750E350g,
	"SUBCORE_VM_FIXED2775_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2775E350g,
	"SUBCORE_VM_FIXED2800_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2800E350g,
	"SUBCORE_VM_FIXED2850_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2850E350g,
	"SUBCORE_VM_FIXED2875_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2875E350g,
	"SUBCORE_VM_FIXED2900_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2900E350g,
	"SUBCORE_VM_FIXED2925_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2925E350g,
	"SUBCORE_VM_FIXED2950_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2950E350g,
	"SUBCORE_VM_FIXED2975_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2975E350g,
	"SUBCORE_VM_FIXED3000_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3000E350g,
	"SUBCORE_VM_FIXED3025_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3025E350g,
	"SUBCORE_VM_FIXED3050_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3050E350g,
	"SUBCORE_VM_FIXED3075_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3075E350g,
	"SUBCORE_VM_FIXED3100_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3100E350g,
	"SUBCORE_VM_FIXED3125_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3125E350g,
	"SUBCORE_VM_FIXED3150_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3150E350g,
	"SUBCORE_VM_FIXED3200_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3200E350g,
	"SUBCORE_VM_FIXED3225_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3225E350g,
	"SUBCORE_VM_FIXED3250_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3250E350g,
	"SUBCORE_VM_FIXED3300_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3300E350g,
	"SUBCORE_VM_FIXED3325_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3325E350g,
	"SUBCORE_VM_FIXED3375_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3375E350g,
	"SUBCORE_VM_FIXED3400_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3400E350g,
	"SUBCORE_VM_FIXED3450_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3450E350g,
	"SUBCORE_VM_FIXED3500_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3500E350g,
	"SUBCORE_VM_FIXED3525_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3525E350g,
	"SUBCORE_VM_FIXED3575_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3575E350g,
	"SUBCORE_VM_FIXED3600_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3600E350g,
	"SUBCORE_VM_FIXED3625_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3625E350g,
	"SUBCORE_VM_FIXED3675_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3675E350g,
	"SUBCORE_VM_FIXED3700_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3700E350g,
	"SUBCORE_VM_FIXED3750_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3750E350g,
	"SUBCORE_VM_FIXED3800_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3800E350g,
	"SUBCORE_VM_FIXED3825_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3825E350g,
	"SUBCORE_VM_FIXED3850_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3850E350g,
	"SUBCORE_VM_FIXED3875_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3875E350g,
	"SUBCORE_VM_FIXED3900_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3900E350g,
	"SUBCORE_VM_FIXED3975_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3975E350g,
	"SUBCORE_VM_FIXED4000_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4000E350g,
	"SUBCORE_VM_FIXED4025_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4025E350g,
	"SUBCORE_VM_FIXED4050_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4050E350g,
	"SUBCORE_VM_FIXED4100_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4100E350g,
	"SUBCORE_VM_FIXED4125_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4125E350g,
	"SUBCORE_VM_FIXED4200_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4200E350g,
	"SUBCORE_VM_FIXED4225_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4225E350g,
	"SUBCORE_VM_FIXED4250_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4250E350g,
	"SUBCORE_VM_FIXED4275_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4275E350g,
	"SUBCORE_VM_FIXED4300_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4300E350g,
	"SUBCORE_VM_FIXED4350_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4350E350g,
	"SUBCORE_VM_FIXED4375_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4375E350g,
	"SUBCORE_VM_FIXED4400_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4400E350g,
	"SUBCORE_VM_FIXED4425_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4425E350g,
	"SUBCORE_VM_FIXED4500_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4500E350g,
	"SUBCORE_VM_FIXED4550_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4550E350g,
	"SUBCORE_VM_FIXED4575_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4575E350g,
	"SUBCORE_VM_FIXED4600_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4600E350g,
	"SUBCORE_VM_FIXED4625_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4625E350g,
	"SUBCORE_VM_FIXED4650_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4650E350g,
	"SUBCORE_VM_FIXED4675_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4675E350g,
	"SUBCORE_VM_FIXED4700_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4700E350g,
	"SUBCORE_VM_FIXED4725_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4725E350g,
	"SUBCORE_VM_FIXED4750_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4750E350g,
	"SUBCORE_VM_FIXED4800_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4800E350g,
	"SUBCORE_VM_FIXED4875_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4875E350g,
	"SUBCORE_VM_FIXED4900_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4900E350g,
	"SUBCORE_VM_FIXED4950_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4950E350g,
	"SUBCORE_VM_FIXED5000_E3_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed5000E350g,
	"SUBCORE_VM_FIXED0025_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0025E450g,
	"SUBCORE_VM_FIXED0050_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0050E450g,
	"SUBCORE_VM_FIXED0075_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0075E450g,
	"SUBCORE_VM_FIXED0100_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0100E450g,
	"SUBCORE_VM_FIXED0125_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0125E450g,
	"SUBCORE_VM_FIXED0150_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0150E450g,
	"SUBCORE_VM_FIXED0175_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0175E450g,
	"SUBCORE_VM_FIXED0200_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0200E450g,
	"SUBCORE_VM_FIXED0225_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0225E450g,
	"SUBCORE_VM_FIXED0250_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0250E450g,
	"SUBCORE_VM_FIXED0275_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0275E450g,
	"SUBCORE_VM_FIXED0300_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0300E450g,
	"SUBCORE_VM_FIXED0325_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0325E450g,
	"SUBCORE_VM_FIXED0350_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0350E450g,
	"SUBCORE_VM_FIXED0375_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0375E450g,
	"SUBCORE_VM_FIXED0400_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0400E450g,
	"SUBCORE_VM_FIXED0425_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0425E450g,
	"SUBCORE_VM_FIXED0450_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0450E450g,
	"SUBCORE_VM_FIXED0475_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0475E450g,
	"SUBCORE_VM_FIXED0500_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0500E450g,
	"SUBCORE_VM_FIXED0525_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0525E450g,
	"SUBCORE_VM_FIXED0550_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0550E450g,
	"SUBCORE_VM_FIXED0575_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0575E450g,
	"SUBCORE_VM_FIXED0600_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0600E450g,
	"SUBCORE_VM_FIXED0625_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0625E450g,
	"SUBCORE_VM_FIXED0650_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0650E450g,
	"SUBCORE_VM_FIXED0675_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0675E450g,
	"SUBCORE_VM_FIXED0700_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0700E450g,
	"SUBCORE_VM_FIXED0725_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0725E450g,
	"SUBCORE_VM_FIXED0750_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0750E450g,
	"SUBCORE_VM_FIXED0775_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0775E450g,
	"SUBCORE_VM_FIXED0800_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0800E450g,
	"SUBCORE_VM_FIXED0825_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0825E450g,
	"SUBCORE_VM_FIXED0850_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0850E450g,
	"SUBCORE_VM_FIXED0875_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0875E450g,
	"SUBCORE_VM_FIXED0900_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0900E450g,
	"SUBCORE_VM_FIXED0925_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0925E450g,
	"SUBCORE_VM_FIXED0950_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0950E450g,
	"SUBCORE_VM_FIXED0975_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0975E450g,
	"SUBCORE_VM_FIXED1000_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1000E450g,
	"SUBCORE_VM_FIXED1025_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1025E450g,
	"SUBCORE_VM_FIXED1050_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1050E450g,
	"SUBCORE_VM_FIXED1075_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1075E450g,
	"SUBCORE_VM_FIXED1100_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1100E450g,
	"SUBCORE_VM_FIXED1125_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1125E450g,
	"SUBCORE_VM_FIXED1150_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1150E450g,
	"SUBCORE_VM_FIXED1175_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1175E450g,
	"SUBCORE_VM_FIXED1200_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1200E450g,
	"SUBCORE_VM_FIXED1225_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1225E450g,
	"SUBCORE_VM_FIXED1250_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1250E450g,
	"SUBCORE_VM_FIXED1275_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1275E450g,
	"SUBCORE_VM_FIXED1300_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1300E450g,
	"SUBCORE_VM_FIXED1325_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1325E450g,
	"SUBCORE_VM_FIXED1350_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1350E450g,
	"SUBCORE_VM_FIXED1375_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1375E450g,
	"SUBCORE_VM_FIXED1400_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1400E450g,
	"SUBCORE_VM_FIXED1425_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1425E450g,
	"SUBCORE_VM_FIXED1450_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1450E450g,
	"SUBCORE_VM_FIXED1475_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1475E450g,
	"SUBCORE_VM_FIXED1500_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1500E450g,
	"SUBCORE_VM_FIXED1525_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1525E450g,
	"SUBCORE_VM_FIXED1550_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1550E450g,
	"SUBCORE_VM_FIXED1575_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1575E450g,
	"SUBCORE_VM_FIXED1600_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1600E450g,
	"SUBCORE_VM_FIXED1625_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1625E450g,
	"SUBCORE_VM_FIXED1650_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1650E450g,
	"SUBCORE_VM_FIXED1700_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1700E450g,
	"SUBCORE_VM_FIXED1725_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1725E450g,
	"SUBCORE_VM_FIXED1750_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1750E450g,
	"SUBCORE_VM_FIXED1800_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1800E450g,
	"SUBCORE_VM_FIXED1850_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1850E450g,
	"SUBCORE_VM_FIXED1875_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1875E450g,
	"SUBCORE_VM_FIXED1900_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1900E450g,
	"SUBCORE_VM_FIXED1925_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1925E450g,
	"SUBCORE_VM_FIXED1950_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1950E450g,
	"SUBCORE_VM_FIXED2000_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2000E450g,
	"SUBCORE_VM_FIXED2025_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2025E450g,
	"SUBCORE_VM_FIXED2050_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2050E450g,
	"SUBCORE_VM_FIXED2100_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2100E450g,
	"SUBCORE_VM_FIXED2125_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2125E450g,
	"SUBCORE_VM_FIXED2150_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2150E450g,
	"SUBCORE_VM_FIXED2175_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2175E450g,
	"SUBCORE_VM_FIXED2200_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2200E450g,
	"SUBCORE_VM_FIXED2250_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2250E450g,
	"SUBCORE_VM_FIXED2275_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2275E450g,
	"SUBCORE_VM_FIXED2300_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2300E450g,
	"SUBCORE_VM_FIXED2325_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2325E450g,
	"SUBCORE_VM_FIXED2350_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2350E450g,
	"SUBCORE_VM_FIXED2375_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2375E450g,
	"SUBCORE_VM_FIXED2400_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2400E450g,
	"SUBCORE_VM_FIXED2450_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2450E450g,
	"SUBCORE_VM_FIXED2475_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2475E450g,
	"SUBCORE_VM_FIXED2500_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2500E450g,
	"SUBCORE_VM_FIXED2550_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2550E450g,
	"SUBCORE_VM_FIXED2600_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2600E450g,
	"SUBCORE_VM_FIXED2625_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2625E450g,
	"SUBCORE_VM_FIXED2650_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2650E450g,
	"SUBCORE_VM_FIXED2700_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2700E450g,
	"SUBCORE_VM_FIXED2750_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2750E450g,
	"SUBCORE_VM_FIXED2775_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2775E450g,
	"SUBCORE_VM_FIXED2800_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2800E450g,
	"SUBCORE_VM_FIXED2850_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2850E450g,
	"SUBCORE_VM_FIXED2875_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2875E450g,
	"SUBCORE_VM_FIXED2900_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2900E450g,
	"SUBCORE_VM_FIXED2925_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2925E450g,
	"SUBCORE_VM_FIXED2950_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2950E450g,
	"SUBCORE_VM_FIXED2975_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2975E450g,
	"SUBCORE_VM_FIXED3000_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3000E450g,
	"SUBCORE_VM_FIXED3025_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3025E450g,
	"SUBCORE_VM_FIXED3050_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3050E450g,
	"SUBCORE_VM_FIXED3075_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3075E450g,
	"SUBCORE_VM_FIXED3100_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3100E450g,
	"SUBCORE_VM_FIXED3125_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3125E450g,
	"SUBCORE_VM_FIXED3150_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3150E450g,
	"SUBCORE_VM_FIXED3200_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3200E450g,
	"SUBCORE_VM_FIXED3225_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3225E450g,
	"SUBCORE_VM_FIXED3250_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3250E450g,
	"SUBCORE_VM_FIXED3300_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3300E450g,
	"SUBCORE_VM_FIXED3325_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3325E450g,
	"SUBCORE_VM_FIXED3375_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3375E450g,
	"SUBCORE_VM_FIXED3400_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3400E450g,
	"SUBCORE_VM_FIXED3450_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3450E450g,
	"SUBCORE_VM_FIXED3500_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3500E450g,
	"SUBCORE_VM_FIXED3525_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3525E450g,
	"SUBCORE_VM_FIXED3575_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3575E450g,
	"SUBCORE_VM_FIXED3600_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3600E450g,
	"SUBCORE_VM_FIXED3625_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3625E450g,
	"SUBCORE_VM_FIXED3675_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3675E450g,
	"SUBCORE_VM_FIXED3700_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3700E450g,
	"SUBCORE_VM_FIXED3750_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3750E450g,
	"SUBCORE_VM_FIXED3800_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3800E450g,
	"SUBCORE_VM_FIXED3825_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3825E450g,
	"SUBCORE_VM_FIXED3850_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3850E450g,
	"SUBCORE_VM_FIXED3875_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3875E450g,
	"SUBCORE_VM_FIXED3900_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3900E450g,
	"SUBCORE_VM_FIXED3975_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3975E450g,
	"SUBCORE_VM_FIXED4000_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4000E450g,
	"SUBCORE_VM_FIXED4025_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4025E450g,
	"SUBCORE_VM_FIXED4050_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4050E450g,
	"SUBCORE_VM_FIXED4100_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4100E450g,
	"SUBCORE_VM_FIXED4125_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4125E450g,
	"SUBCORE_VM_FIXED4200_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4200E450g,
	"SUBCORE_VM_FIXED4225_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4225E450g,
	"SUBCORE_VM_FIXED4250_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4250E450g,
	"SUBCORE_VM_FIXED4275_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4275E450g,
	"SUBCORE_VM_FIXED4300_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4300E450g,
	"SUBCORE_VM_FIXED4350_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4350E450g,
	"SUBCORE_VM_FIXED4375_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4375E450g,
	"SUBCORE_VM_FIXED4400_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4400E450g,
	"SUBCORE_VM_FIXED4425_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4425E450g,
	"SUBCORE_VM_FIXED4500_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4500E450g,
	"SUBCORE_VM_FIXED4550_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4550E450g,
	"SUBCORE_VM_FIXED4575_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4575E450g,
	"SUBCORE_VM_FIXED4600_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4600E450g,
	"SUBCORE_VM_FIXED4625_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4625E450g,
	"SUBCORE_VM_FIXED4650_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4650E450g,
	"SUBCORE_VM_FIXED4675_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4675E450g,
	"SUBCORE_VM_FIXED4700_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4700E450g,
	"SUBCORE_VM_FIXED4725_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4725E450g,
	"SUBCORE_VM_FIXED4750_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4750E450g,
	"SUBCORE_VM_FIXED4800_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4800E450g,
	"SUBCORE_VM_FIXED4875_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4875E450g,
	"SUBCORE_VM_FIXED4900_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4900E450g,
	"SUBCORE_VM_FIXED4950_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4950E450g,
	"SUBCORE_VM_FIXED5000_E4_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed5000E450g,
	"SUBCORE_VM_FIXED0020_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0020A150g,
	"SUBCORE_VM_FIXED0040_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0040A150g,
	"SUBCORE_VM_FIXED0060_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0060A150g,
	"SUBCORE_VM_FIXED0080_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0080A150g,
	"SUBCORE_VM_FIXED0100_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0100A150g,
	"SUBCORE_VM_FIXED0120_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0120A150g,
	"SUBCORE_VM_FIXED0140_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0140A150g,
	"SUBCORE_VM_FIXED0160_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0160A150g,
	"SUBCORE_VM_FIXED0180_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0180A150g,
	"SUBCORE_VM_FIXED0200_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0200A150g,
	"SUBCORE_VM_FIXED0220_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0220A150g,
	"SUBCORE_VM_FIXED0240_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0240A150g,
	"SUBCORE_VM_FIXED0260_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0260A150g,
	"SUBCORE_VM_FIXED0280_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0280A150g,
	"SUBCORE_VM_FIXED0300_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0300A150g,
	"SUBCORE_VM_FIXED0320_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0320A150g,
	"SUBCORE_VM_FIXED0340_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0340A150g,
	"SUBCORE_VM_FIXED0360_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0360A150g,
	"SUBCORE_VM_FIXED0380_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0380A150g,
	"SUBCORE_VM_FIXED0400_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0400A150g,
	"SUBCORE_VM_FIXED0420_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0420A150g,
	"SUBCORE_VM_FIXED0440_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0440A150g,
	"SUBCORE_VM_FIXED0460_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0460A150g,
	"SUBCORE_VM_FIXED0480_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0480A150g,
	"SUBCORE_VM_FIXED0500_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0500A150g,
	"SUBCORE_VM_FIXED0520_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0520A150g,
	"SUBCORE_VM_FIXED0540_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0540A150g,
	"SUBCORE_VM_FIXED0560_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0560A150g,
	"SUBCORE_VM_FIXED0580_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0580A150g,
	"SUBCORE_VM_FIXED0600_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0600A150g,
	"SUBCORE_VM_FIXED0620_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0620A150g,
	"SUBCORE_VM_FIXED0640_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0640A150g,
	"SUBCORE_VM_FIXED0660_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0660A150g,
	"SUBCORE_VM_FIXED0680_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0680A150g,
	"SUBCORE_VM_FIXED0700_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0700A150g,
	"SUBCORE_VM_FIXED0720_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0720A150g,
	"SUBCORE_VM_FIXED0740_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0740A150g,
	"SUBCORE_VM_FIXED0760_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0760A150g,
	"SUBCORE_VM_FIXED0780_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0780A150g,
	"SUBCORE_VM_FIXED0800_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0800A150g,
	"SUBCORE_VM_FIXED0820_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0820A150g,
	"SUBCORE_VM_FIXED0840_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0840A150g,
	"SUBCORE_VM_FIXED0860_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0860A150g,
	"SUBCORE_VM_FIXED0880_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0880A150g,
	"SUBCORE_VM_FIXED0900_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0900A150g,
	"SUBCORE_VM_FIXED0920_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0920A150g,
	"SUBCORE_VM_FIXED0940_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0940A150g,
	"SUBCORE_VM_FIXED0960_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0960A150g,
	"SUBCORE_VM_FIXED0980_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0980A150g,
	"SUBCORE_VM_FIXED1000_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1000A150g,
	"SUBCORE_VM_FIXED1020_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1020A150g,
	"SUBCORE_VM_FIXED1040_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1040A150g,
	"SUBCORE_VM_FIXED1060_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1060A150g,
	"SUBCORE_VM_FIXED1080_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1080A150g,
	"SUBCORE_VM_FIXED1100_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1100A150g,
	"SUBCORE_VM_FIXED1120_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1120A150g,
	"SUBCORE_VM_FIXED1140_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1140A150g,
	"SUBCORE_VM_FIXED1160_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1160A150g,
	"SUBCORE_VM_FIXED1180_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1180A150g,
	"SUBCORE_VM_FIXED1200_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1200A150g,
	"SUBCORE_VM_FIXED1220_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1220A150g,
	"SUBCORE_VM_FIXED1240_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1240A150g,
	"SUBCORE_VM_FIXED1260_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1260A150g,
	"SUBCORE_VM_FIXED1280_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1280A150g,
	"SUBCORE_VM_FIXED1300_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1300A150g,
	"SUBCORE_VM_FIXED1320_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1320A150g,
	"SUBCORE_VM_FIXED1340_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1340A150g,
	"SUBCORE_VM_FIXED1360_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1360A150g,
	"SUBCORE_VM_FIXED1380_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1380A150g,
	"SUBCORE_VM_FIXED1400_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1400A150g,
	"SUBCORE_VM_FIXED1420_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1420A150g,
	"SUBCORE_VM_FIXED1440_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1440A150g,
	"SUBCORE_VM_FIXED1460_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1460A150g,
	"SUBCORE_VM_FIXED1480_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1480A150g,
	"SUBCORE_VM_FIXED1500_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1500A150g,
	"SUBCORE_VM_FIXED1520_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1520A150g,
	"SUBCORE_VM_FIXED1540_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1540A150g,
	"SUBCORE_VM_FIXED1560_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1560A150g,
	"SUBCORE_VM_FIXED1580_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1580A150g,
	"SUBCORE_VM_FIXED1600_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1600A150g,
	"SUBCORE_VM_FIXED1620_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1620A150g,
	"SUBCORE_VM_FIXED1640_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1640A150g,
	"SUBCORE_VM_FIXED1660_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1660A150g,
	"SUBCORE_VM_FIXED1680_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1680A150g,
	"SUBCORE_VM_FIXED1700_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1700A150g,
	"SUBCORE_VM_FIXED1720_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1720A150g,
	"SUBCORE_VM_FIXED1740_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1740A150g,
	"SUBCORE_VM_FIXED1760_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1760A150g,
	"SUBCORE_VM_FIXED1780_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1780A150g,
	"SUBCORE_VM_FIXED1800_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1800A150g,
	"SUBCORE_VM_FIXED1820_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1820A150g,
	"SUBCORE_VM_FIXED1840_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1840A150g,
	"SUBCORE_VM_FIXED1860_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1860A150g,
	"SUBCORE_VM_FIXED1880_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1880A150g,
	"SUBCORE_VM_FIXED1900_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1900A150g,
	"SUBCORE_VM_FIXED1920_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1920A150g,
	"SUBCORE_VM_FIXED1940_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1940A150g,
	"SUBCORE_VM_FIXED1960_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1960A150g,
	"SUBCORE_VM_FIXED1980_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1980A150g,
	"SUBCORE_VM_FIXED2000_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2000A150g,
	"SUBCORE_VM_FIXED2020_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2020A150g,
	"SUBCORE_VM_FIXED2040_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2040A150g,
	"SUBCORE_VM_FIXED2060_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2060A150g,
	"SUBCORE_VM_FIXED2080_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2080A150g,
	"SUBCORE_VM_FIXED2100_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2100A150g,
	"SUBCORE_VM_FIXED2120_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2120A150g,
	"SUBCORE_VM_FIXED2140_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2140A150g,
	"SUBCORE_VM_FIXED2160_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2160A150g,
	"SUBCORE_VM_FIXED2180_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2180A150g,
	"SUBCORE_VM_FIXED2200_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2200A150g,
	"SUBCORE_VM_FIXED2220_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2220A150g,
	"SUBCORE_VM_FIXED2240_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2240A150g,
	"SUBCORE_VM_FIXED2260_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2260A150g,
	"SUBCORE_VM_FIXED2280_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2280A150g,
	"SUBCORE_VM_FIXED2300_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2300A150g,
	"SUBCORE_VM_FIXED2320_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2320A150g,
	"SUBCORE_VM_FIXED2340_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2340A150g,
	"SUBCORE_VM_FIXED2360_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2360A150g,
	"SUBCORE_VM_FIXED2380_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2380A150g,
	"SUBCORE_VM_FIXED2400_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2400A150g,
	"SUBCORE_VM_FIXED2420_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2420A150g,
	"SUBCORE_VM_FIXED2440_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2440A150g,
	"SUBCORE_VM_FIXED2460_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2460A150g,
	"SUBCORE_VM_FIXED2480_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2480A150g,
	"SUBCORE_VM_FIXED2500_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2500A150g,
	"SUBCORE_VM_FIXED2520_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2520A150g,
	"SUBCORE_VM_FIXED2540_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2540A150g,
	"SUBCORE_VM_FIXED2560_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2560A150g,
	"SUBCORE_VM_FIXED2580_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2580A150g,
	"SUBCORE_VM_FIXED2600_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2600A150g,
	"SUBCORE_VM_FIXED2620_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2620A150g,
	"SUBCORE_VM_FIXED2640_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2640A150g,
	"SUBCORE_VM_FIXED2660_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2660A150g,
	"SUBCORE_VM_FIXED2680_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2680A150g,
	"SUBCORE_VM_FIXED2700_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2700A150g,
	"SUBCORE_VM_FIXED2720_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2720A150g,
	"SUBCORE_VM_FIXED2740_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2740A150g,
	"SUBCORE_VM_FIXED2760_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2760A150g,
	"SUBCORE_VM_FIXED2780_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2780A150g,
	"SUBCORE_VM_FIXED2800_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2800A150g,
	"SUBCORE_VM_FIXED2820_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2820A150g,
	"SUBCORE_VM_FIXED2840_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2840A150g,
	"SUBCORE_VM_FIXED2860_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2860A150g,
	"SUBCORE_VM_FIXED2880_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2880A150g,
	"SUBCORE_VM_FIXED2900_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2900A150g,
	"SUBCORE_VM_FIXED2920_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2920A150g,
	"SUBCORE_VM_FIXED2940_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2940A150g,
	"SUBCORE_VM_FIXED2960_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2960A150g,
	"SUBCORE_VM_FIXED2980_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2980A150g,
	"SUBCORE_VM_FIXED3000_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3000A150g,
	"SUBCORE_VM_FIXED3020_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3020A150g,
	"SUBCORE_VM_FIXED3040_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3040A150g,
	"SUBCORE_VM_FIXED3060_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3060A150g,
	"SUBCORE_VM_FIXED3080_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3080A150g,
	"SUBCORE_VM_FIXED3100_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3100A150g,
	"SUBCORE_VM_FIXED3120_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3120A150g,
	"SUBCORE_VM_FIXED3140_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3140A150g,
	"SUBCORE_VM_FIXED3160_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3160A150g,
	"SUBCORE_VM_FIXED3180_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3180A150g,
	"SUBCORE_VM_FIXED3200_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3200A150g,
	"SUBCORE_VM_FIXED3220_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3220A150g,
	"SUBCORE_VM_FIXED3240_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3240A150g,
	"SUBCORE_VM_FIXED3260_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3260A150g,
	"SUBCORE_VM_FIXED3280_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3280A150g,
	"SUBCORE_VM_FIXED3300_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3300A150g,
	"SUBCORE_VM_FIXED3320_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3320A150g,
	"SUBCORE_VM_FIXED3340_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3340A150g,
	"SUBCORE_VM_FIXED3360_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3360A150g,
	"SUBCORE_VM_FIXED3380_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3380A150g,
	"SUBCORE_VM_FIXED3400_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3400A150g,
	"SUBCORE_VM_FIXED3420_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3420A150g,
	"SUBCORE_VM_FIXED3440_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3440A150g,
	"SUBCORE_VM_FIXED3460_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3460A150g,
	"SUBCORE_VM_FIXED3480_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3480A150g,
	"SUBCORE_VM_FIXED3500_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3500A150g,
	"SUBCORE_VM_FIXED3520_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3520A150g,
	"SUBCORE_VM_FIXED3540_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3540A150g,
	"SUBCORE_VM_FIXED3560_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3560A150g,
	"SUBCORE_VM_FIXED3580_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3580A150g,
	"SUBCORE_VM_FIXED3600_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3600A150g,
	"SUBCORE_VM_FIXED3620_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3620A150g,
	"SUBCORE_VM_FIXED3640_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3640A150g,
	"SUBCORE_VM_FIXED3660_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3660A150g,
	"SUBCORE_VM_FIXED3680_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3680A150g,
	"SUBCORE_VM_FIXED3700_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3700A150g,
	"SUBCORE_VM_FIXED3720_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3720A150g,
	"SUBCORE_VM_FIXED3740_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3740A150g,
	"SUBCORE_VM_FIXED3760_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3760A150g,
	"SUBCORE_VM_FIXED3780_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3780A150g,
	"SUBCORE_VM_FIXED3800_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3800A150g,
	"SUBCORE_VM_FIXED3820_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3820A150g,
	"SUBCORE_VM_FIXED3840_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3840A150g,
	"SUBCORE_VM_FIXED3860_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3860A150g,
	"SUBCORE_VM_FIXED3880_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3880A150g,
	"SUBCORE_VM_FIXED3900_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3900A150g,
	"SUBCORE_VM_FIXED3920_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3920A150g,
	"SUBCORE_VM_FIXED3940_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3940A150g,
	"SUBCORE_VM_FIXED3960_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3960A150g,
	"SUBCORE_VM_FIXED3980_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3980A150g,
	"SUBCORE_VM_FIXED4000_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4000A150g,
	"SUBCORE_VM_FIXED4020_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4020A150g,
	"SUBCORE_VM_FIXED4040_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4040A150g,
	"SUBCORE_VM_FIXED4060_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4060A150g,
	"SUBCORE_VM_FIXED4080_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4080A150g,
	"SUBCORE_VM_FIXED4100_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4100A150g,
	"SUBCORE_VM_FIXED4120_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4120A150g,
	"SUBCORE_VM_FIXED4140_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4140A150g,
	"SUBCORE_VM_FIXED4160_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4160A150g,
	"SUBCORE_VM_FIXED4180_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4180A150g,
	"SUBCORE_VM_FIXED4200_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4200A150g,
	"SUBCORE_VM_FIXED4220_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4220A150g,
	"SUBCORE_VM_FIXED4240_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4240A150g,
	"SUBCORE_VM_FIXED4260_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4260A150g,
	"SUBCORE_VM_FIXED4280_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4280A150g,
	"SUBCORE_VM_FIXED4300_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4300A150g,
	"SUBCORE_VM_FIXED4320_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4320A150g,
	"SUBCORE_VM_FIXED4340_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4340A150g,
	"SUBCORE_VM_FIXED4360_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4360A150g,
	"SUBCORE_VM_FIXED4380_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4380A150g,
	"SUBCORE_VM_FIXED4400_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4400A150g,
	"SUBCORE_VM_FIXED4420_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4420A150g,
	"SUBCORE_VM_FIXED4440_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4440A150g,
	"SUBCORE_VM_FIXED4460_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4460A150g,
	"SUBCORE_VM_FIXED4480_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4480A150g,
	"SUBCORE_VM_FIXED4500_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4500A150g,
	"SUBCORE_VM_FIXED4520_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4520A150g,
	"SUBCORE_VM_FIXED4540_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4540A150g,
	"SUBCORE_VM_FIXED4560_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4560A150g,
	"SUBCORE_VM_FIXED4580_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4580A150g,
	"SUBCORE_VM_FIXED4600_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4600A150g,
	"SUBCORE_VM_FIXED4620_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4620A150g,
	"SUBCORE_VM_FIXED4640_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4640A150g,
	"SUBCORE_VM_FIXED4660_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4660A150g,
	"SUBCORE_VM_FIXED4680_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4680A150g,
	"SUBCORE_VM_FIXED4700_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4700A150g,
	"SUBCORE_VM_FIXED4720_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4720A150g,
	"SUBCORE_VM_FIXED4740_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4740A150g,
	"SUBCORE_VM_FIXED4760_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4760A150g,
	"SUBCORE_VM_FIXED4780_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4780A150g,
	"SUBCORE_VM_FIXED4800_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4800A150g,
	"SUBCORE_VM_FIXED4820_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4820A150g,
	"SUBCORE_VM_FIXED4840_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4840A150g,
	"SUBCORE_VM_FIXED4860_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4860A150g,
	"SUBCORE_VM_FIXED4880_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4880A150g,
	"SUBCORE_VM_FIXED4900_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4900A150g,
	"SUBCORE_VM_FIXED4920_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4920A150g,
	"SUBCORE_VM_FIXED4940_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4940A150g,
	"SUBCORE_VM_FIXED4960_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4960A150g,
	"SUBCORE_VM_FIXED4980_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4980A150g,
	"SUBCORE_VM_FIXED5000_A1_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed5000A150g,
	"SUBCORE_VM_FIXED0090_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0090X950g,
	"SUBCORE_VM_FIXED0180_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0180X950g,
	"SUBCORE_VM_FIXED0270_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0270X950g,
	"SUBCORE_VM_FIXED0360_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0360X950g,
	"SUBCORE_VM_FIXED0450_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0450X950g,
	"SUBCORE_VM_FIXED0540_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0540X950g,
	"SUBCORE_VM_FIXED0630_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0630X950g,
	"SUBCORE_VM_FIXED0720_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0720X950g,
	"SUBCORE_VM_FIXED0810_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0810X950g,
	"SUBCORE_VM_FIXED0900_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0900X950g,
	"SUBCORE_VM_FIXED0990_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed0990X950g,
	"SUBCORE_VM_FIXED1080_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1080X950g,
	"SUBCORE_VM_FIXED1170_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1170X950g,
	"SUBCORE_VM_FIXED1260_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1260X950g,
	"SUBCORE_VM_FIXED1350_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1350X950g,
	"SUBCORE_VM_FIXED1440_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1440X950g,
	"SUBCORE_VM_FIXED1530_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1530X950g,
	"SUBCORE_VM_FIXED1620_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1620X950g,
	"SUBCORE_VM_FIXED1710_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1710X950g,
	"SUBCORE_VM_FIXED1800_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1800X950g,
	"SUBCORE_VM_FIXED1890_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1890X950g,
	"SUBCORE_VM_FIXED1980_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed1980X950g,
	"SUBCORE_VM_FIXED2070_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2070X950g,
	"SUBCORE_VM_FIXED2160_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2160X950g,
	"SUBCORE_VM_FIXED2250_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2250X950g,
	"SUBCORE_VM_FIXED2340_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2340X950g,
	"SUBCORE_VM_FIXED2430_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2430X950g,
	"SUBCORE_VM_FIXED2520_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2520X950g,
	"SUBCORE_VM_FIXED2610_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2610X950g,
	"SUBCORE_VM_FIXED2700_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2700X950g,
	"SUBCORE_VM_FIXED2790_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2790X950g,
	"SUBCORE_VM_FIXED2880_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2880X950g,
	"SUBCORE_VM_FIXED2970_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed2970X950g,
	"SUBCORE_VM_FIXED3060_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3060X950g,
	"SUBCORE_VM_FIXED3150_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3150X950g,
	"SUBCORE_VM_FIXED3240_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3240X950g,
	"SUBCORE_VM_FIXED3330_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3330X950g,
	"SUBCORE_VM_FIXED3420_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3420X950g,
	"SUBCORE_VM_FIXED3510_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3510X950g,
	"SUBCORE_VM_FIXED3600_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3600X950g,
	"SUBCORE_VM_FIXED3690_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3690X950g,
	"SUBCORE_VM_FIXED3780_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3780X950g,
	"SUBCORE_VM_FIXED3870_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3870X950g,
	"SUBCORE_VM_FIXED3960_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed3960X950g,
	"SUBCORE_VM_FIXED4050_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4050X950g,
	"SUBCORE_VM_FIXED4140_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4140X950g,
	"SUBCORE_VM_FIXED4230_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4230X950g,
	"SUBCORE_VM_FIXED4320_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4320X950g,
	"SUBCORE_VM_FIXED4410_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4410X950g,
	"SUBCORE_VM_FIXED4500_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4500X950g,
	"SUBCORE_VM_FIXED4590_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4590X950g,
	"SUBCORE_VM_FIXED4680_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4680X950g,
	"SUBCORE_VM_FIXED4770_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4770X950g,
	"SUBCORE_VM_FIXED4860_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4860X950g,
	"SUBCORE_VM_FIXED4950_X9_50G":  CreateInternalVnicDetailsVnicShapeSubcoreVmFixed4950X950g,
	"DYNAMIC_A1_50G":               CreateInternalVnicDetailsVnicShapeDynamicA150g,
	"FIXED0040_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed0040A150g,
	"FIXED0100_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed0100A150g,
	"FIXED0200_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed0200A150g,
	"FIXED0300_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed0300A150g,
	"FIXED0400_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed0400A150g,
	"FIXED0500_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed0500A150g,
	"FIXED0600_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed0600A150g,
	"FIXED0700_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed0700A150g,
	"FIXED0800_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed0800A150g,
	"FIXED0900_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed0900A150g,
	"FIXED1000_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed1000A150g,
	"FIXED1100_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed1100A150g,
	"FIXED1200_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed1200A150g,
	"FIXED1300_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed1300A150g,
	"FIXED1400_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed1400A150g,
	"FIXED1500_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed1500A150g,
	"FIXED1600_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed1600A150g,
	"FIXED1700_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed1700A150g,
	"FIXED1800_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed1800A150g,
	"FIXED1900_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed1900A150g,
	"FIXED2000_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed2000A150g,
	"FIXED2100_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed2100A150g,
	"FIXED2200_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed2200A150g,
	"FIXED2300_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed2300A150g,
	"FIXED2400_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed2400A150g,
	"FIXED2500_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed2500A150g,
	"FIXED2600_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed2600A150g,
	"FIXED2700_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed2700A150g,
	"FIXED2800_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed2800A150g,
	"FIXED2900_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed2900A150g,
	"FIXED3000_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed3000A150g,
	"FIXED3100_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed3100A150g,
	"FIXED3200_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed3200A150g,
	"FIXED3300_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed3300A150g,
	"FIXED3400_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed3400A150g,
	"FIXED3500_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed3500A150g,
	"FIXED3600_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed3600A150g,
	"FIXED3700_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed3700A150g,
	"FIXED3800_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed3800A150g,
	"FIXED3900_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed3900A150g,
	"FIXED4000_A1_50G":             CreateInternalVnicDetailsVnicShapeFixed4000A150g,
	"ENTIREHOST_A1_50G":            CreateInternalVnicDetailsVnicShapeEntirehostA150g,
	"DYNAMIC_X9_50G":               CreateInternalVnicDetailsVnicShapeDynamicX950g,
	"FIXED0040_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed0040X950g,
	"FIXED0400_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed0400X950g,
	"FIXED0800_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed0800X950g,
	"FIXED1200_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed1200X950g,
	"FIXED1600_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed1600X950g,
	"FIXED2000_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed2000X950g,
	"FIXED2400_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed2400X950g,
	"FIXED2800_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed2800X950g,
	"FIXED3200_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed3200X950g,
	"FIXED3600_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed3600X950g,
	"FIXED4000_X9_50G":             CreateInternalVnicDetailsVnicShapeFixed4000X950g,
	"STANDARD_VM_FIXED0100_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed0100X950g,
	"STANDARD_VM_FIXED0200_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed0200X950g,
	"STANDARD_VM_FIXED0300_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed0300X950g,
	"STANDARD_VM_FIXED0400_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed0400X950g,
	"STANDARD_VM_FIXED0500_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed0500X950g,
	"STANDARD_VM_FIXED0600_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed0600X950g,
	"STANDARD_VM_FIXED0700_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed0700X950g,
	"STANDARD_VM_FIXED0800_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed0800X950g,
	"STANDARD_VM_FIXED0900_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed0900X950g,
	"STANDARD_VM_FIXED1000_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed1000X950g,
	"STANDARD_VM_FIXED1100_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed1100X950g,
	"STANDARD_VM_FIXED1200_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed1200X950g,
	"STANDARD_VM_FIXED1300_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed1300X950g,
	"STANDARD_VM_FIXED1400_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed1400X950g,
	"STANDARD_VM_FIXED1500_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed1500X950g,
	"STANDARD_VM_FIXED1600_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed1600X950g,
	"STANDARD_VM_FIXED1700_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed1700X950g,
	"STANDARD_VM_FIXED1800_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed1800X950g,
	"STANDARD_VM_FIXED1900_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed1900X950g,
	"STANDARD_VM_FIXED2000_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed2000X950g,
	"STANDARD_VM_FIXED2100_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed2100X950g,
	"STANDARD_VM_FIXED2200_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed2200X950g,
	"STANDARD_VM_FIXED2300_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed2300X950g,
	"STANDARD_VM_FIXED2400_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed2400X950g,
	"STANDARD_VM_FIXED2500_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed2500X950g,
	"STANDARD_VM_FIXED2600_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed2600X950g,
	"STANDARD_VM_FIXED2700_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed2700X950g,
	"STANDARD_VM_FIXED2800_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed2800X950g,
	"STANDARD_VM_FIXED2900_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed2900X950g,
	"STANDARD_VM_FIXED3000_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed3000X950g,
	"STANDARD_VM_FIXED3100_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed3100X950g,
	"STANDARD_VM_FIXED3200_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed3200X950g,
	"STANDARD_VM_FIXED3300_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed3300X950g,
	"STANDARD_VM_FIXED3400_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed3400X950g,
	"STANDARD_VM_FIXED3500_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed3500X950g,
	"STANDARD_VM_FIXED3600_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed3600X950g,
	"STANDARD_VM_FIXED3700_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed3700X950g,
	"STANDARD_VM_FIXED3800_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed3800X950g,
	"STANDARD_VM_FIXED3900_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed3900X950g,
	"STANDARD_VM_FIXED4000_X9_50G": CreateInternalVnicDetailsVnicShapeStandardVmFixed4000X950g,
	"ENTIREHOST_X9_50G":            CreateInternalVnicDetailsVnicShapeEntirehostX950g,
}

// GetCreateInternalVnicDetailsVnicShapeEnumValues Enumerates the set of values for CreateInternalVnicDetailsVnicShapeEnum
func GetCreateInternalVnicDetailsVnicShapeEnumValues() []CreateInternalVnicDetailsVnicShapeEnum {
	values := make([]CreateInternalVnicDetailsVnicShapeEnum, 0)
	for _, v := range mappingCreateInternalVnicDetailsVnicShapeEnum {
		values = append(values, v)
	}
	return values
}

// GetCreateInternalVnicDetailsVnicShapeEnumStringValues Enumerates the set of values in String for CreateInternalVnicDetailsVnicShapeEnum
func GetCreateInternalVnicDetailsVnicShapeEnumStringValues() []string {
	return []string{
		"DYNAMIC",
		"FIXED0040",
		"FIXED0060",
		"FIXED0060_PSM",
		"FIXED0100",
		"FIXED0120",
		"FIXED0120_2X",
		"FIXED0200",
		"FIXED0240",
		"FIXED0480",
		"ENTIREHOST",
		"DYNAMIC_25G",
		"FIXED0040_25G",
		"FIXED0100_25G",
		"FIXED0200_25G",
		"FIXED0400_25G",
		"FIXED0800_25G",
		"FIXED1600_25G",
		"FIXED2400_25G",
		"ENTIREHOST_25G",
		"DYNAMIC_E1_25G",
		"FIXED0040_E1_25G",
		"FIXED0070_E1_25G",
		"FIXED0140_E1_25G",
		"FIXED0280_E1_25G",
		"FIXED0560_E1_25G",
		"FIXED1120_E1_25G",
		"FIXED1680_E1_25G",
		"ENTIREHOST_E1_25G",
		"DYNAMIC_B1_25G",
		"FIXED0040_B1_25G",
		"FIXED0060_B1_25G",
		"FIXED0120_B1_25G",
		"FIXED0240_B1_25G",
		"FIXED0480_B1_25G",
		"FIXED0960_B1_25G",
		"ENTIREHOST_B1_25G",
		"MICRO_VM_FIXED0048_E1_25G",
		"MICRO_LB_FIXED0001_E1_25G",
		"VNICAAS_FIXED0200",
		"VNICAAS_FIXED0400",
		"VNICAAS_FIXED0700",
		"VNICAAS_NLB_APPROVED_10G",
		"VNICAAS_NLB_APPROVED_25G",
		"VNICAAS_TELESIS_25G",
		"VNICAAS_TELESIS_10G",
		"VNICAAS_AMBASSADOR_FIXED0100",
		"VNICAAS_PRIVATEDNS",
		"VNICAAS_FWAAS",
		"DYNAMIC_E3_50G",
		"FIXED0040_E3_50G",
		"FIXED0100_E3_50G",
		"FIXED0200_E3_50G",
		"FIXED0300_E3_50G",
		"FIXED0400_E3_50G",
		"FIXED0500_E3_50G",
		"FIXED0600_E3_50G",
		"FIXED0700_E3_50G",
		"FIXED0800_E3_50G",
		"FIXED0900_E3_50G",
		"FIXED1000_E3_50G",
		"FIXED1100_E3_50G",
		"FIXED1200_E3_50G",
		"FIXED1300_E3_50G",
		"FIXED1400_E3_50G",
		"FIXED1500_E3_50G",
		"FIXED1600_E3_50G",
		"FIXED1700_E3_50G",
		"FIXED1800_E3_50G",
		"FIXED1900_E3_50G",
		"FIXED2000_E3_50G",
		"FIXED2100_E3_50G",
		"FIXED2200_E3_50G",
		"FIXED2300_E3_50G",
		"FIXED2400_E3_50G",
		"FIXED2500_E3_50G",
		"FIXED2600_E3_50G",
		"FIXED2700_E3_50G",
		"FIXED2800_E3_50G",
		"FIXED2900_E3_50G",
		"FIXED3000_E3_50G",
		"FIXED3100_E3_50G",
		"FIXED3200_E3_50G",
		"FIXED3300_E3_50G",
		"FIXED3400_E3_50G",
		"FIXED3500_E3_50G",
		"FIXED3600_E3_50G",
		"FIXED3700_E3_50G",
		"FIXED3800_E3_50G",
		"FIXED3900_E3_50G",
		"FIXED4000_E3_50G",
		"ENTIREHOST_E3_50G",
		"DYNAMIC_E4_50G",
		"FIXED0040_E4_50G",
		"FIXED0100_E4_50G",
		"FIXED0200_E4_50G",
		"FIXED0300_E4_50G",
		"FIXED0400_E4_50G",
		"FIXED0500_E4_50G",
		"FIXED0600_E4_50G",
		"FIXED0700_E4_50G",
		"FIXED0800_E4_50G",
		"FIXED0900_E4_50G",
		"FIXED1000_E4_50G",
		"FIXED1100_E4_50G",
		"FIXED1200_E4_50G",
		"FIXED1300_E4_50G",
		"FIXED1400_E4_50G",
		"FIXED1500_E4_50G",
		"FIXED1600_E4_50G",
		"FIXED1700_E4_50G",
		"FIXED1800_E4_50G",
		"FIXED1900_E4_50G",
		"FIXED2000_E4_50G",
		"FIXED2100_E4_50G",
		"FIXED2200_E4_50G",
		"FIXED2300_E4_50G",
		"FIXED2400_E4_50G",
		"FIXED2500_E4_50G",
		"FIXED2600_E4_50G",
		"FIXED2700_E4_50G",
		"FIXED2800_E4_50G",
		"FIXED2900_E4_50G",
		"FIXED3000_E4_50G",
		"FIXED3100_E4_50G",
		"FIXED3200_E4_50G",
		"FIXED3300_E4_50G",
		"FIXED3400_E4_50G",
		"FIXED3500_E4_50G",
		"FIXED3600_E4_50G",
		"FIXED3700_E4_50G",
		"FIXED3800_E4_50G",
		"FIXED3900_E4_50G",
		"FIXED4000_E4_50G",
		"ENTIREHOST_E4_50G",
		"MICRO_VM_FIXED0050_E3_50G",
		"SUBCORE_VM_FIXED0025_E3_50G",
		"SUBCORE_VM_FIXED0050_E3_50G",
		"SUBCORE_VM_FIXED0075_E3_50G",
		"SUBCORE_VM_FIXED0100_E3_50G",
		"SUBCORE_VM_FIXED0125_E3_50G",
		"SUBCORE_VM_FIXED0150_E3_50G",
		"SUBCORE_VM_FIXED0175_E3_50G",
		"SUBCORE_VM_FIXED0200_E3_50G",
		"SUBCORE_VM_FIXED0225_E3_50G",
		"SUBCORE_VM_FIXED0250_E3_50G",
		"SUBCORE_VM_FIXED0275_E3_50G",
		"SUBCORE_VM_FIXED0300_E3_50G",
		"SUBCORE_VM_FIXED0325_E3_50G",
		"SUBCORE_VM_FIXED0350_E3_50G",
		"SUBCORE_VM_FIXED0375_E3_50G",
		"SUBCORE_VM_FIXED0400_E3_50G",
		"SUBCORE_VM_FIXED0425_E3_50G",
		"SUBCORE_VM_FIXED0450_E3_50G",
		"SUBCORE_VM_FIXED0475_E3_50G",
		"SUBCORE_VM_FIXED0500_E3_50G",
		"SUBCORE_VM_FIXED0525_E3_50G",
		"SUBCORE_VM_FIXED0550_E3_50G",
		"SUBCORE_VM_FIXED0575_E3_50G",
		"SUBCORE_VM_FIXED0600_E3_50G",
		"SUBCORE_VM_FIXED0625_E3_50G",
		"SUBCORE_VM_FIXED0650_E3_50G",
		"SUBCORE_VM_FIXED0675_E3_50G",
		"SUBCORE_VM_FIXED0700_E3_50G",
		"SUBCORE_VM_FIXED0725_E3_50G",
		"SUBCORE_VM_FIXED0750_E3_50G",
		"SUBCORE_VM_FIXED0775_E3_50G",
		"SUBCORE_VM_FIXED0800_E3_50G",
		"SUBCORE_VM_FIXED0825_E3_50G",
		"SUBCORE_VM_FIXED0850_E3_50G",
		"SUBCORE_VM_FIXED0875_E3_50G",
		"SUBCORE_VM_FIXED0900_E3_50G",
		"SUBCORE_VM_FIXED0925_E3_50G",
		"SUBCORE_VM_FIXED0950_E3_50G",
		"SUBCORE_VM_FIXED0975_E3_50G",
		"SUBCORE_VM_FIXED1000_E3_50G",
		"SUBCORE_VM_FIXED1025_E3_50G",
		"SUBCORE_VM_FIXED1050_E3_50G",
		"SUBCORE_VM_FIXED1075_E3_50G",
		"SUBCORE_VM_FIXED1100_E3_50G",
		"SUBCORE_VM_FIXED1125_E3_50G",
		"SUBCORE_VM_FIXED1150_E3_50G",
		"SUBCORE_VM_FIXED1175_E3_50G",
		"SUBCORE_VM_FIXED1200_E3_50G",
		"SUBCORE_VM_FIXED1225_E3_50G",
		"SUBCORE_VM_FIXED1250_E3_50G",
		"SUBCORE_VM_FIXED1275_E3_50G",
		"SUBCORE_VM_FIXED1300_E3_50G",
		"SUBCORE_VM_FIXED1325_E3_50G",
		"SUBCORE_VM_FIXED1350_E3_50G",
		"SUBCORE_VM_FIXED1375_E3_50G",
		"SUBCORE_VM_FIXED1400_E3_50G",
		"SUBCORE_VM_FIXED1425_E3_50G",
		"SUBCORE_VM_FIXED1450_E3_50G",
		"SUBCORE_VM_FIXED1475_E3_50G",
		"SUBCORE_VM_FIXED1500_E3_50G",
		"SUBCORE_VM_FIXED1525_E3_50G",
		"SUBCORE_VM_FIXED1550_E3_50G",
		"SUBCORE_VM_FIXED1575_E3_50G",
		"SUBCORE_VM_FIXED1600_E3_50G",
		"SUBCORE_VM_FIXED1625_E3_50G",
		"SUBCORE_VM_FIXED1650_E3_50G",
		"SUBCORE_VM_FIXED1700_E3_50G",
		"SUBCORE_VM_FIXED1725_E3_50G",
		"SUBCORE_VM_FIXED1750_E3_50G",
		"SUBCORE_VM_FIXED1800_E3_50G",
		"SUBCORE_VM_FIXED1850_E3_50G",
		"SUBCORE_VM_FIXED1875_E3_50G",
		"SUBCORE_VM_FIXED1900_E3_50G",
		"SUBCORE_VM_FIXED1925_E3_50G",
		"SUBCORE_VM_FIXED1950_E3_50G",
		"SUBCORE_VM_FIXED2000_E3_50G",
		"SUBCORE_VM_FIXED2025_E3_50G",
		"SUBCORE_VM_FIXED2050_E3_50G",
		"SUBCORE_VM_FIXED2100_E3_50G",
		"SUBCORE_VM_FIXED2125_E3_50G",
		"SUBCORE_VM_FIXED2150_E3_50G",
		"SUBCORE_VM_FIXED2175_E3_50G",
		"SUBCORE_VM_FIXED2200_E3_50G",
		"SUBCORE_VM_FIXED2250_E3_50G",
		"SUBCORE_VM_FIXED2275_E3_50G",
		"SUBCORE_VM_FIXED2300_E3_50G",
		"SUBCORE_VM_FIXED2325_E3_50G",
		"SUBCORE_VM_FIXED2350_E3_50G",
		"SUBCORE_VM_FIXED2375_E3_50G",
		"SUBCORE_VM_FIXED2400_E3_50G",
		"SUBCORE_VM_FIXED2450_E3_50G",
		"SUBCORE_VM_FIXED2475_E3_50G",
		"SUBCORE_VM_FIXED2500_E3_50G",
		"SUBCORE_VM_FIXED2550_E3_50G",
		"SUBCORE_VM_FIXED2600_E3_50G",
		"SUBCORE_VM_FIXED2625_E3_50G",
		"SUBCORE_VM_FIXED2650_E3_50G",
		"SUBCORE_VM_FIXED2700_E3_50G",
		"SUBCORE_VM_FIXED2750_E3_50G",
		"SUBCORE_VM_FIXED2775_E3_50G",
		"SUBCORE_VM_FIXED2800_E3_50G",
		"SUBCORE_VM_FIXED2850_E3_50G",
		"SUBCORE_VM_FIXED2875_E3_50G",
		"SUBCORE_VM_FIXED2900_E3_50G",
		"SUBCORE_VM_FIXED2925_E3_50G",
		"SUBCORE_VM_FIXED2950_E3_50G",
		"SUBCORE_VM_FIXED2975_E3_50G",
		"SUBCORE_VM_FIXED3000_E3_50G",
		"SUBCORE_VM_FIXED3025_E3_50G",
		"SUBCORE_VM_FIXED3050_E3_50G",
		"SUBCORE_VM_FIXED3075_E3_50G",
		"SUBCORE_VM_FIXED3100_E3_50G",
		"SUBCORE_VM_FIXED3125_E3_50G",
		"SUBCORE_VM_FIXED3150_E3_50G",
		"SUBCORE_VM_FIXED3200_E3_50G",
		"SUBCORE_VM_FIXED3225_E3_50G",
		"SUBCORE_VM_FIXED3250_E3_50G",
		"SUBCORE_VM_FIXED3300_E3_50G",
		"SUBCORE_VM_FIXED3325_E3_50G",
		"SUBCORE_VM_FIXED3375_E3_50G",
		"SUBCORE_VM_FIXED3400_E3_50G",
		"SUBCORE_VM_FIXED3450_E3_50G",
		"SUBCORE_VM_FIXED3500_E3_50G",
		"SUBCORE_VM_FIXED3525_E3_50G",
		"SUBCORE_VM_FIXED3575_E3_50G",
		"SUBCORE_VM_FIXED3600_E3_50G",
		"SUBCORE_VM_FIXED3625_E3_50G",
		"SUBCORE_VM_FIXED3675_E3_50G",
		"SUBCORE_VM_FIXED3700_E3_50G",
		"SUBCORE_VM_FIXED3750_E3_50G",
		"SUBCORE_VM_FIXED3800_E3_50G",
		"SUBCORE_VM_FIXED3825_E3_50G",
		"SUBCORE_VM_FIXED3850_E3_50G",
		"SUBCORE_VM_FIXED3875_E3_50G",
		"SUBCORE_VM_FIXED3900_E3_50G",
		"SUBCORE_VM_FIXED3975_E3_50G",
		"SUBCORE_VM_FIXED4000_E3_50G",
		"SUBCORE_VM_FIXED4025_E3_50G",
		"SUBCORE_VM_FIXED4050_E3_50G",
		"SUBCORE_VM_FIXED4100_E3_50G",
		"SUBCORE_VM_FIXED4125_E3_50G",
		"SUBCORE_VM_FIXED4200_E3_50G",
		"SUBCORE_VM_FIXED4225_E3_50G",
		"SUBCORE_VM_FIXED4250_E3_50G",
		"SUBCORE_VM_FIXED4275_E3_50G",
		"SUBCORE_VM_FIXED4300_E3_50G",
		"SUBCORE_VM_FIXED4350_E3_50G",
		"SUBCORE_VM_FIXED4375_E3_50G",
		"SUBCORE_VM_FIXED4400_E3_50G",
		"SUBCORE_VM_FIXED4425_E3_50G",
		"SUBCORE_VM_FIXED4500_E3_50G",
		"SUBCORE_VM_FIXED4550_E3_50G",
		"SUBCORE_VM_FIXED4575_E3_50G",
		"SUBCORE_VM_FIXED4600_E3_50G",
		"SUBCORE_VM_FIXED4625_E3_50G",
		"SUBCORE_VM_FIXED4650_E3_50G",
		"SUBCORE_VM_FIXED4675_E3_50G",
		"SUBCORE_VM_FIXED4700_E3_50G",
		"SUBCORE_VM_FIXED4725_E3_50G",
		"SUBCORE_VM_FIXED4750_E3_50G",
		"SUBCORE_VM_FIXED4800_E3_50G",
		"SUBCORE_VM_FIXED4875_E3_50G",
		"SUBCORE_VM_FIXED4900_E3_50G",
		"SUBCORE_VM_FIXED4950_E3_50G",
		"SUBCORE_VM_FIXED5000_E3_50G",
		"SUBCORE_VM_FIXED0025_E4_50G",
		"SUBCORE_VM_FIXED0050_E4_50G",
		"SUBCORE_VM_FIXED0075_E4_50G",
		"SUBCORE_VM_FIXED0100_E4_50G",
		"SUBCORE_VM_FIXED0125_E4_50G",
		"SUBCORE_VM_FIXED0150_E4_50G",
		"SUBCORE_VM_FIXED0175_E4_50G",
		"SUBCORE_VM_FIXED0200_E4_50G",
		"SUBCORE_VM_FIXED0225_E4_50G",
		"SUBCORE_VM_FIXED0250_E4_50G",
		"SUBCORE_VM_FIXED0275_E4_50G",
		"SUBCORE_VM_FIXED0300_E4_50G",
		"SUBCORE_VM_FIXED0325_E4_50G",
		"SUBCORE_VM_FIXED0350_E4_50G",
		"SUBCORE_VM_FIXED0375_E4_50G",
		"SUBCORE_VM_FIXED0400_E4_50G",
		"SUBCORE_VM_FIXED0425_E4_50G",
		"SUBCORE_VM_FIXED0450_E4_50G",
		"SUBCORE_VM_FIXED0475_E4_50G",
		"SUBCORE_VM_FIXED0500_E4_50G",
		"SUBCORE_VM_FIXED0525_E4_50G",
		"SUBCORE_VM_FIXED0550_E4_50G",
		"SUBCORE_VM_FIXED0575_E4_50G",
		"SUBCORE_VM_FIXED0600_E4_50G",
		"SUBCORE_VM_FIXED0625_E4_50G",
		"SUBCORE_VM_FIXED0650_E4_50G",
		"SUBCORE_VM_FIXED0675_E4_50G",
		"SUBCORE_VM_FIXED0700_E4_50G",
		"SUBCORE_VM_FIXED0725_E4_50G",
		"SUBCORE_VM_FIXED0750_E4_50G",
		"SUBCORE_VM_FIXED0775_E4_50G",
		"SUBCORE_VM_FIXED0800_E4_50G",
		"SUBCORE_VM_FIXED0825_E4_50G",
		"SUBCORE_VM_FIXED0850_E4_50G",
		"SUBCORE_VM_FIXED0875_E4_50G",
		"SUBCORE_VM_FIXED0900_E4_50G",
		"SUBCORE_VM_FIXED0925_E4_50G",
		"SUBCORE_VM_FIXED0950_E4_50G",
		"SUBCORE_VM_FIXED0975_E4_50G",
		"SUBCORE_VM_FIXED1000_E4_50G",
		"SUBCORE_VM_FIXED1025_E4_50G",
		"SUBCORE_VM_FIXED1050_E4_50G",
		"SUBCORE_VM_FIXED1075_E4_50G",
		"SUBCORE_VM_FIXED1100_E4_50G",
		"SUBCORE_VM_FIXED1125_E4_50G",
		"SUBCORE_VM_FIXED1150_E4_50G",
		"SUBCORE_VM_FIXED1175_E4_50G",
		"SUBCORE_VM_FIXED1200_E4_50G",
		"SUBCORE_VM_FIXED1225_E4_50G",
		"SUBCORE_VM_FIXED1250_E4_50G",
		"SUBCORE_VM_FIXED1275_E4_50G",
		"SUBCORE_VM_FIXED1300_E4_50G",
		"SUBCORE_VM_FIXED1325_E4_50G",
		"SUBCORE_VM_FIXED1350_E4_50G",
		"SUBCORE_VM_FIXED1375_E4_50G",
		"SUBCORE_VM_FIXED1400_E4_50G",
		"SUBCORE_VM_FIXED1425_E4_50G",
		"SUBCORE_VM_FIXED1450_E4_50G",
		"SUBCORE_VM_FIXED1475_E4_50G",
		"SUBCORE_VM_FIXED1500_E4_50G",
		"SUBCORE_VM_FIXED1525_E4_50G",
		"SUBCORE_VM_FIXED1550_E4_50G",
		"SUBCORE_VM_FIXED1575_E4_50G",
		"SUBCORE_VM_FIXED1600_E4_50G",
		"SUBCORE_VM_FIXED1625_E4_50G",
		"SUBCORE_VM_FIXED1650_E4_50G",
		"SUBCORE_VM_FIXED1700_E4_50G",
		"SUBCORE_VM_FIXED1725_E4_50G",
		"SUBCORE_VM_FIXED1750_E4_50G",
		"SUBCORE_VM_FIXED1800_E4_50G",
		"SUBCORE_VM_FIXED1850_E4_50G",
		"SUBCORE_VM_FIXED1875_E4_50G",
		"SUBCORE_VM_FIXED1900_E4_50G",
		"SUBCORE_VM_FIXED1925_E4_50G",
		"SUBCORE_VM_FIXED1950_E4_50G",
		"SUBCORE_VM_FIXED2000_E4_50G",
		"SUBCORE_VM_FIXED2025_E4_50G",
		"SUBCORE_VM_FIXED2050_E4_50G",
		"SUBCORE_VM_FIXED2100_E4_50G",
		"SUBCORE_VM_FIXED2125_E4_50G",
		"SUBCORE_VM_FIXED2150_E4_50G",
		"SUBCORE_VM_FIXED2175_E4_50G",
		"SUBCORE_VM_FIXED2200_E4_50G",
		"SUBCORE_VM_FIXED2250_E4_50G",
		"SUBCORE_VM_FIXED2275_E4_50G",
		"SUBCORE_VM_FIXED2300_E4_50G",
		"SUBCORE_VM_FIXED2325_E4_50G",
		"SUBCORE_VM_FIXED2350_E4_50G",
		"SUBCORE_VM_FIXED2375_E4_50G",
		"SUBCORE_VM_FIXED2400_E4_50G",
		"SUBCORE_VM_FIXED2450_E4_50G",
		"SUBCORE_VM_FIXED2475_E4_50G",
		"SUBCORE_VM_FIXED2500_E4_50G",
		"SUBCORE_VM_FIXED2550_E4_50G",
		"SUBCORE_VM_FIXED2600_E4_50G",
		"SUBCORE_VM_FIXED2625_E4_50G",
		"SUBCORE_VM_FIXED2650_E4_50G",
		"SUBCORE_VM_FIXED2700_E4_50G",
		"SUBCORE_VM_FIXED2750_E4_50G",
		"SUBCORE_VM_FIXED2775_E4_50G",
		"SUBCORE_VM_FIXED2800_E4_50G",
		"SUBCORE_VM_FIXED2850_E4_50G",
		"SUBCORE_VM_FIXED2875_E4_50G",
		"SUBCORE_VM_FIXED2900_E4_50G",
		"SUBCORE_VM_FIXED2925_E4_50G",
		"SUBCORE_VM_FIXED2950_E4_50G",
		"SUBCORE_VM_FIXED2975_E4_50G",
		"SUBCORE_VM_FIXED3000_E4_50G",
		"SUBCORE_VM_FIXED3025_E4_50G",
		"SUBCORE_VM_FIXED3050_E4_50G",
		"SUBCORE_VM_FIXED3075_E4_50G",
		"SUBCORE_VM_FIXED3100_E4_50G",
		"SUBCORE_VM_FIXED3125_E4_50G",
		"SUBCORE_VM_FIXED3150_E4_50G",
		"SUBCORE_VM_FIXED3200_E4_50G",
		"SUBCORE_VM_FIXED3225_E4_50G",
		"SUBCORE_VM_FIXED3250_E4_50G",
		"SUBCORE_VM_FIXED3300_E4_50G",
		"SUBCORE_VM_FIXED3325_E4_50G",
		"SUBCORE_VM_FIXED3375_E4_50G",
		"SUBCORE_VM_FIXED3400_E4_50G",
		"SUBCORE_VM_FIXED3450_E4_50G",
		"SUBCORE_VM_FIXED3500_E4_50G",
		"SUBCORE_VM_FIXED3525_E4_50G",
		"SUBCORE_VM_FIXED3575_E4_50G",
		"SUBCORE_VM_FIXED3600_E4_50G",
		"SUBCORE_VM_FIXED3625_E4_50G",
		"SUBCORE_VM_FIXED3675_E4_50G",
		"SUBCORE_VM_FIXED3700_E4_50G",
		"SUBCORE_VM_FIXED3750_E4_50G",
		"SUBCORE_VM_FIXED3800_E4_50G",
		"SUBCORE_VM_FIXED3825_E4_50G",
		"SUBCORE_VM_FIXED3850_E4_50G",
		"SUBCORE_VM_FIXED3875_E4_50G",
		"SUBCORE_VM_FIXED3900_E4_50G",
		"SUBCORE_VM_FIXED3975_E4_50G",
		"SUBCORE_VM_FIXED4000_E4_50G",
		"SUBCORE_VM_FIXED4025_E4_50G",
		"SUBCORE_VM_FIXED4050_E4_50G",
		"SUBCORE_VM_FIXED4100_E4_50G",
		"SUBCORE_VM_FIXED4125_E4_50G",
		"SUBCORE_VM_FIXED4200_E4_50G",
		"SUBCORE_VM_FIXED4225_E4_50G",
		"SUBCORE_VM_FIXED4250_E4_50G",
		"SUBCORE_VM_FIXED4275_E4_50G",
		"SUBCORE_VM_FIXED4300_E4_50G",
		"SUBCORE_VM_FIXED4350_E4_50G",
		"SUBCORE_VM_FIXED4375_E4_50G",
		"SUBCORE_VM_FIXED4400_E4_50G",
		"SUBCORE_VM_FIXED4425_E4_50G",
		"SUBCORE_VM_FIXED4500_E4_50G",
		"SUBCORE_VM_FIXED4550_E4_50G",
		"SUBCORE_VM_FIXED4575_E4_50G",
		"SUBCORE_VM_FIXED4600_E4_50G",
		"SUBCORE_VM_FIXED4625_E4_50G",
		"SUBCORE_VM_FIXED4650_E4_50G",
		"SUBCORE_VM_FIXED4675_E4_50G",
		"SUBCORE_VM_FIXED4700_E4_50G",
		"SUBCORE_VM_FIXED4725_E4_50G",
		"SUBCORE_VM_FIXED4750_E4_50G",
		"SUBCORE_VM_FIXED4800_E4_50G",
		"SUBCORE_VM_FIXED4875_E4_50G",
		"SUBCORE_VM_FIXED4900_E4_50G",
		"SUBCORE_VM_FIXED4950_E4_50G",
		"SUBCORE_VM_FIXED5000_E4_50G",
		"SUBCORE_VM_FIXED0020_A1_50G",
		"SUBCORE_VM_FIXED0040_A1_50G",
		"SUBCORE_VM_FIXED0060_A1_50G",
		"SUBCORE_VM_FIXED0080_A1_50G",
		"SUBCORE_VM_FIXED0100_A1_50G",
		"SUBCORE_VM_FIXED0120_A1_50G",
		"SUBCORE_VM_FIXED0140_A1_50G",
		"SUBCORE_VM_FIXED0160_A1_50G",
		"SUBCORE_VM_FIXED0180_A1_50G",
		"SUBCORE_VM_FIXED0200_A1_50G",
		"SUBCORE_VM_FIXED0220_A1_50G",
		"SUBCORE_VM_FIXED0240_A1_50G",
		"SUBCORE_VM_FIXED0260_A1_50G",
		"SUBCORE_VM_FIXED0280_A1_50G",
		"SUBCORE_VM_FIXED0300_A1_50G",
		"SUBCORE_VM_FIXED0320_A1_50G",
		"SUBCORE_VM_FIXED0340_A1_50G",
		"SUBCORE_VM_FIXED0360_A1_50G",
		"SUBCORE_VM_FIXED0380_A1_50G",
		"SUBCORE_VM_FIXED0400_A1_50G",
		"SUBCORE_VM_FIXED0420_A1_50G",
		"SUBCORE_VM_FIXED0440_A1_50G",
		"SUBCORE_VM_FIXED0460_A1_50G",
		"SUBCORE_VM_FIXED0480_A1_50G",
		"SUBCORE_VM_FIXED0500_A1_50G",
		"SUBCORE_VM_FIXED0520_A1_50G",
		"SUBCORE_VM_FIXED0540_A1_50G",
		"SUBCORE_VM_FIXED0560_A1_50G",
		"SUBCORE_VM_FIXED0580_A1_50G",
		"SUBCORE_VM_FIXED0600_A1_50G",
		"SUBCORE_VM_FIXED0620_A1_50G",
		"SUBCORE_VM_FIXED0640_A1_50G",
		"SUBCORE_VM_FIXED0660_A1_50G",
		"SUBCORE_VM_FIXED0680_A1_50G",
		"SUBCORE_VM_FIXED0700_A1_50G",
		"SUBCORE_VM_FIXED0720_A1_50G",
		"SUBCORE_VM_FIXED0740_A1_50G",
		"SUBCORE_VM_FIXED0760_A1_50G",
		"SUBCORE_VM_FIXED0780_A1_50G",
		"SUBCORE_VM_FIXED0800_A1_50G",
		"SUBCORE_VM_FIXED0820_A1_50G",
		"SUBCORE_VM_FIXED0840_A1_50G",
		"SUBCORE_VM_FIXED0860_A1_50G",
		"SUBCORE_VM_FIXED0880_A1_50G",
		"SUBCORE_VM_FIXED0900_A1_50G",
		"SUBCORE_VM_FIXED0920_A1_50G",
		"SUBCORE_VM_FIXED0940_A1_50G",
		"SUBCORE_VM_FIXED0960_A1_50G",
		"SUBCORE_VM_FIXED0980_A1_50G",
		"SUBCORE_VM_FIXED1000_A1_50G",
		"SUBCORE_VM_FIXED1020_A1_50G",
		"SUBCORE_VM_FIXED1040_A1_50G",
		"SUBCORE_VM_FIXED1060_A1_50G",
		"SUBCORE_VM_FIXED1080_A1_50G",
		"SUBCORE_VM_FIXED1100_A1_50G",
		"SUBCORE_VM_FIXED1120_A1_50G",
		"SUBCORE_VM_FIXED1140_A1_50G",
		"SUBCORE_VM_FIXED1160_A1_50G",
		"SUBCORE_VM_FIXED1180_A1_50G",
		"SUBCORE_VM_FIXED1200_A1_50G",
		"SUBCORE_VM_FIXED1220_A1_50G",
		"SUBCORE_VM_FIXED1240_A1_50G",
		"SUBCORE_VM_FIXED1260_A1_50G",
		"SUBCORE_VM_FIXED1280_A1_50G",
		"SUBCORE_VM_FIXED1300_A1_50G",
		"SUBCORE_VM_FIXED1320_A1_50G",
		"SUBCORE_VM_FIXED1340_A1_50G",
		"SUBCORE_VM_FIXED1360_A1_50G",
		"SUBCORE_VM_FIXED1380_A1_50G",
		"SUBCORE_VM_FIXED1400_A1_50G",
		"SUBCORE_VM_FIXED1420_A1_50G",
		"SUBCORE_VM_FIXED1440_A1_50G",
		"SUBCORE_VM_FIXED1460_A1_50G",
		"SUBCORE_VM_FIXED1480_A1_50G",
		"SUBCORE_VM_FIXED1500_A1_50G",
		"SUBCORE_VM_FIXED1520_A1_50G",
		"SUBCORE_VM_FIXED1540_A1_50G",
		"SUBCORE_VM_FIXED1560_A1_50G",
		"SUBCORE_VM_FIXED1580_A1_50G",
		"SUBCORE_VM_FIXED1600_A1_50G",
		"SUBCORE_VM_FIXED1620_A1_50G",
		"SUBCORE_VM_FIXED1640_A1_50G",
		"SUBCORE_VM_FIXED1660_A1_50G",
		"SUBCORE_VM_FIXED1680_A1_50G",
		"SUBCORE_VM_FIXED1700_A1_50G",
		"SUBCORE_VM_FIXED1720_A1_50G",
		"SUBCORE_VM_FIXED1740_A1_50G",
		"SUBCORE_VM_FIXED1760_A1_50G",
		"SUBCORE_VM_FIXED1780_A1_50G",
		"SUBCORE_VM_FIXED1800_A1_50G",
		"SUBCORE_VM_FIXED1820_A1_50G",
		"SUBCORE_VM_FIXED1840_A1_50G",
		"SUBCORE_VM_FIXED1860_A1_50G",
		"SUBCORE_VM_FIXED1880_A1_50G",
		"SUBCORE_VM_FIXED1900_A1_50G",
		"SUBCORE_VM_FIXED1920_A1_50G",
		"SUBCORE_VM_FIXED1940_A1_50G",
		"SUBCORE_VM_FIXED1960_A1_50G",
		"SUBCORE_VM_FIXED1980_A1_50G",
		"SUBCORE_VM_FIXED2000_A1_50G",
		"SUBCORE_VM_FIXED2020_A1_50G",
		"SUBCORE_VM_FIXED2040_A1_50G",
		"SUBCORE_VM_FIXED2060_A1_50G",
		"SUBCORE_VM_FIXED2080_A1_50G",
		"SUBCORE_VM_FIXED2100_A1_50G",
		"SUBCORE_VM_FIXED2120_A1_50G",
		"SUBCORE_VM_FIXED2140_A1_50G",
		"SUBCORE_VM_FIXED2160_A1_50G",
		"SUBCORE_VM_FIXED2180_A1_50G",
		"SUBCORE_VM_FIXED2200_A1_50G",
		"SUBCORE_VM_FIXED2220_A1_50G",
		"SUBCORE_VM_FIXED2240_A1_50G",
		"SUBCORE_VM_FIXED2260_A1_50G",
		"SUBCORE_VM_FIXED2280_A1_50G",
		"SUBCORE_VM_FIXED2300_A1_50G",
		"SUBCORE_VM_FIXED2320_A1_50G",
		"SUBCORE_VM_FIXED2340_A1_50G",
		"SUBCORE_VM_FIXED2360_A1_50G",
		"SUBCORE_VM_FIXED2380_A1_50G",
		"SUBCORE_VM_FIXED2400_A1_50G",
		"SUBCORE_VM_FIXED2420_A1_50G",
		"SUBCORE_VM_FIXED2440_A1_50G",
		"SUBCORE_VM_FIXED2460_A1_50G",
		"SUBCORE_VM_FIXED2480_A1_50G",
		"SUBCORE_VM_FIXED2500_A1_50G",
		"SUBCORE_VM_FIXED2520_A1_50G",
		"SUBCORE_VM_FIXED2540_A1_50G",
		"SUBCORE_VM_FIXED2560_A1_50G",
		"SUBCORE_VM_FIXED2580_A1_50G",
		"SUBCORE_VM_FIXED2600_A1_50G",
		"SUBCORE_VM_FIXED2620_A1_50G",
		"SUBCORE_VM_FIXED2640_A1_50G",
		"SUBCORE_VM_FIXED2660_A1_50G",
		"SUBCORE_VM_FIXED2680_A1_50G",
		"SUBCORE_VM_FIXED2700_A1_50G",
		"SUBCORE_VM_FIXED2720_A1_50G",
		"SUBCORE_VM_FIXED2740_A1_50G",
		"SUBCORE_VM_FIXED2760_A1_50G",
		"SUBCORE_VM_FIXED2780_A1_50G",
		"SUBCORE_VM_FIXED2800_A1_50G",
		"SUBCORE_VM_FIXED2820_A1_50G",
		"SUBCORE_VM_FIXED2840_A1_50G",
		"SUBCORE_VM_FIXED2860_A1_50G",
		"SUBCORE_VM_FIXED2880_A1_50G",
		"SUBCORE_VM_FIXED2900_A1_50G",
		"SUBCORE_VM_FIXED2920_A1_50G",
		"SUBCORE_VM_FIXED2940_A1_50G",
		"SUBCORE_VM_FIXED2960_A1_50G",
		"SUBCORE_VM_FIXED2980_A1_50G",
		"SUBCORE_VM_FIXED3000_A1_50G",
		"SUBCORE_VM_FIXED3020_A1_50G",
		"SUBCORE_VM_FIXED3040_A1_50G",
		"SUBCORE_VM_FIXED3060_A1_50G",
		"SUBCORE_VM_FIXED3080_A1_50G",
		"SUBCORE_VM_FIXED3100_A1_50G",
		"SUBCORE_VM_FIXED3120_A1_50G",
		"SUBCORE_VM_FIXED3140_A1_50G",
		"SUBCORE_VM_FIXED3160_A1_50G",
		"SUBCORE_VM_FIXED3180_A1_50G",
		"SUBCORE_VM_FIXED3200_A1_50G",
		"SUBCORE_VM_FIXED3220_A1_50G",
		"SUBCORE_VM_FIXED3240_A1_50G",
		"SUBCORE_VM_FIXED3260_A1_50G",
		"SUBCORE_VM_FIXED3280_A1_50G",
		"SUBCORE_VM_FIXED3300_A1_50G",
		"SUBCORE_VM_FIXED3320_A1_50G",
		"SUBCORE_VM_FIXED3340_A1_50G",
		"SUBCORE_VM_FIXED3360_A1_50G",
		"SUBCORE_VM_FIXED3380_A1_50G",
		"SUBCORE_VM_FIXED3400_A1_50G",
		"SUBCORE_VM_FIXED3420_A1_50G",
		"SUBCORE_VM_FIXED3440_A1_50G",
		"SUBCORE_VM_FIXED3460_A1_50G",
		"SUBCORE_VM_FIXED3480_A1_50G",
		"SUBCORE_VM_FIXED3500_A1_50G",
		"SUBCORE_VM_FIXED3520_A1_50G",
		"SUBCORE_VM_FIXED3540_A1_50G",
		"SUBCORE_VM_FIXED3560_A1_50G",
		"SUBCORE_VM_FIXED3580_A1_50G",
		"SUBCORE_VM_FIXED3600_A1_50G",
		"SUBCORE_VM_FIXED3620_A1_50G",
		"SUBCORE_VM_FIXED3640_A1_50G",
		"SUBCORE_VM_FIXED3660_A1_50G",
		"SUBCORE_VM_FIXED3680_A1_50G",
		"SUBCORE_VM_FIXED3700_A1_50G",
		"SUBCORE_VM_FIXED3720_A1_50G",
		"SUBCORE_VM_FIXED3740_A1_50G",
		"SUBCORE_VM_FIXED3760_A1_50G",
		"SUBCORE_VM_FIXED3780_A1_50G",
		"SUBCORE_VM_FIXED3800_A1_50G",
		"SUBCORE_VM_FIXED3820_A1_50G",
		"SUBCORE_VM_FIXED3840_A1_50G",
		"SUBCORE_VM_FIXED3860_A1_50G",
		"SUBCORE_VM_FIXED3880_A1_50G",
		"SUBCORE_VM_FIXED3900_A1_50G",
		"SUBCORE_VM_FIXED3920_A1_50G",
		"SUBCORE_VM_FIXED3940_A1_50G",
		"SUBCORE_VM_FIXED3960_A1_50G",
		"SUBCORE_VM_FIXED3980_A1_50G",
		"SUBCORE_VM_FIXED4000_A1_50G",
		"SUBCORE_VM_FIXED4020_A1_50G",
		"SUBCORE_VM_FIXED4040_A1_50G",
		"SUBCORE_VM_FIXED4060_A1_50G",
		"SUBCORE_VM_FIXED4080_A1_50G",
		"SUBCORE_VM_FIXED4100_A1_50G",
		"SUBCORE_VM_FIXED4120_A1_50G",
		"SUBCORE_VM_FIXED4140_A1_50G",
		"SUBCORE_VM_FIXED4160_A1_50G",
		"SUBCORE_VM_FIXED4180_A1_50G",
		"SUBCORE_VM_FIXED4200_A1_50G",
		"SUBCORE_VM_FIXED4220_A1_50G",
		"SUBCORE_VM_FIXED4240_A1_50G",
		"SUBCORE_VM_FIXED4260_A1_50G",
		"SUBCORE_VM_FIXED4280_A1_50G",
		"SUBCORE_VM_FIXED4300_A1_50G",
		"SUBCORE_VM_FIXED4320_A1_50G",
		"SUBCORE_VM_FIXED4340_A1_50G",
		"SUBCORE_VM_FIXED4360_A1_50G",
		"SUBCORE_VM_FIXED4380_A1_50G",
		"SUBCORE_VM_FIXED4400_A1_50G",
		"SUBCORE_VM_FIXED4420_A1_50G",
		"SUBCORE_VM_FIXED4440_A1_50G",
		"SUBCORE_VM_FIXED4460_A1_50G",
		"SUBCORE_VM_FIXED4480_A1_50G",
		"SUBCORE_VM_FIXED4500_A1_50G",
		"SUBCORE_VM_FIXED4520_A1_50G",
		"SUBCORE_VM_FIXED4540_A1_50G",
		"SUBCORE_VM_FIXED4560_A1_50G",
		"SUBCORE_VM_FIXED4580_A1_50G",
		"SUBCORE_VM_FIXED4600_A1_50G",
		"SUBCORE_VM_FIXED4620_A1_50G",
		"SUBCORE_VM_FIXED4640_A1_50G",
		"SUBCORE_VM_FIXED4660_A1_50G",
		"SUBCORE_VM_FIXED4680_A1_50G",
		"SUBCORE_VM_FIXED4700_A1_50G",
		"SUBCORE_VM_FIXED4720_A1_50G",
		"SUBCORE_VM_FIXED4740_A1_50G",
		"SUBCORE_VM_FIXED4760_A1_50G",
		"SUBCORE_VM_FIXED4780_A1_50G",
		"SUBCORE_VM_FIXED4800_A1_50G",
		"SUBCORE_VM_FIXED4820_A1_50G",
		"SUBCORE_VM_FIXED4840_A1_50G",
		"SUBCORE_VM_FIXED4860_A1_50G",
		"SUBCORE_VM_FIXED4880_A1_50G",
		"SUBCORE_VM_FIXED4900_A1_50G",
		"SUBCORE_VM_FIXED4920_A1_50G",
		"SUBCORE_VM_FIXED4940_A1_50G",
		"SUBCORE_VM_FIXED4960_A1_50G",
		"SUBCORE_VM_FIXED4980_A1_50G",
		"SUBCORE_VM_FIXED5000_A1_50G",
		"SUBCORE_VM_FIXED0090_X9_50G",
		"SUBCORE_VM_FIXED0180_X9_50G",
		"SUBCORE_VM_FIXED0270_X9_50G",
		"SUBCORE_VM_FIXED0360_X9_50G",
		"SUBCORE_VM_FIXED0450_X9_50G",
		"SUBCORE_VM_FIXED0540_X9_50G",
		"SUBCORE_VM_FIXED0630_X9_50G",
		"SUBCORE_VM_FIXED0720_X9_50G",
		"SUBCORE_VM_FIXED0810_X9_50G",
		"SUBCORE_VM_FIXED0900_X9_50G",
		"SUBCORE_VM_FIXED0990_X9_50G",
		"SUBCORE_VM_FIXED1080_X9_50G",
		"SUBCORE_VM_FIXED1170_X9_50G",
		"SUBCORE_VM_FIXED1260_X9_50G",
		"SUBCORE_VM_FIXED1350_X9_50G",
		"SUBCORE_VM_FIXED1440_X9_50G",
		"SUBCORE_VM_FIXED1530_X9_50G",
		"SUBCORE_VM_FIXED1620_X9_50G",
		"SUBCORE_VM_FIXED1710_X9_50G",
		"SUBCORE_VM_FIXED1800_X9_50G",
		"SUBCORE_VM_FIXED1890_X9_50G",
		"SUBCORE_VM_FIXED1980_X9_50G",
		"SUBCORE_VM_FIXED2070_X9_50G",
		"SUBCORE_VM_FIXED2160_X9_50G",
		"SUBCORE_VM_FIXED2250_X9_50G",
		"SUBCORE_VM_FIXED2340_X9_50G",
		"SUBCORE_VM_FIXED2430_X9_50G",
		"SUBCORE_VM_FIXED2520_X9_50G",
		"SUBCORE_VM_FIXED2610_X9_50G",
		"SUBCORE_VM_FIXED2700_X9_50G",
		"SUBCORE_VM_FIXED2790_X9_50G",
		"SUBCORE_VM_FIXED2880_X9_50G",
		"SUBCORE_VM_FIXED2970_X9_50G",
		"SUBCORE_VM_FIXED3060_X9_50G",
		"SUBCORE_VM_FIXED3150_X9_50G",
		"SUBCORE_VM_FIXED3240_X9_50G",
		"SUBCORE_VM_FIXED3330_X9_50G",
		"SUBCORE_VM_FIXED3420_X9_50G",
		"SUBCORE_VM_FIXED3510_X9_50G",
		"SUBCORE_VM_FIXED3600_X9_50G",
		"SUBCORE_VM_FIXED3690_X9_50G",
		"SUBCORE_VM_FIXED3780_X9_50G",
		"SUBCORE_VM_FIXED3870_X9_50G",
		"SUBCORE_VM_FIXED3960_X9_50G",
		"SUBCORE_VM_FIXED4050_X9_50G",
		"SUBCORE_VM_FIXED4140_X9_50G",
		"SUBCORE_VM_FIXED4230_X9_50G",
		"SUBCORE_VM_FIXED4320_X9_50G",
		"SUBCORE_VM_FIXED4410_X9_50G",
		"SUBCORE_VM_FIXED4500_X9_50G",
		"SUBCORE_VM_FIXED4590_X9_50G",
		"SUBCORE_VM_FIXED4680_X9_50G",
		"SUBCORE_VM_FIXED4770_X9_50G",
		"SUBCORE_VM_FIXED4860_X9_50G",
		"SUBCORE_VM_FIXED4950_X9_50G",
		"DYNAMIC_A1_50G",
		"FIXED0040_A1_50G",
		"FIXED0100_A1_50G",
		"FIXED0200_A1_50G",
		"FIXED0300_A1_50G",
		"FIXED0400_A1_50G",
		"FIXED0500_A1_50G",
		"FIXED0600_A1_50G",
		"FIXED0700_A1_50G",
		"FIXED0800_A1_50G",
		"FIXED0900_A1_50G",
		"FIXED1000_A1_50G",
		"FIXED1100_A1_50G",
		"FIXED1200_A1_50G",
		"FIXED1300_A1_50G",
		"FIXED1400_A1_50G",
		"FIXED1500_A1_50G",
		"FIXED1600_A1_50G",
		"FIXED1700_A1_50G",
		"FIXED1800_A1_50G",
		"FIXED1900_A1_50G",
		"FIXED2000_A1_50G",
		"FIXED2100_A1_50G",
		"FIXED2200_A1_50G",
		"FIXED2300_A1_50G",
		"FIXED2400_A1_50G",
		"FIXED2500_A1_50G",
		"FIXED2600_A1_50G",
		"FIXED2700_A1_50G",
		"FIXED2800_A1_50G",
		"FIXED2900_A1_50G",
		"FIXED3000_A1_50G",
		"FIXED3100_A1_50G",
		"FIXED3200_A1_50G",
		"FIXED3300_A1_50G",
		"FIXED3400_A1_50G",
		"FIXED3500_A1_50G",
		"FIXED3600_A1_50G",
		"FIXED3700_A1_50G",
		"FIXED3800_A1_50G",
		"FIXED3900_A1_50G",
		"FIXED4000_A1_50G",
		"ENTIREHOST_A1_50G",
		"DYNAMIC_X9_50G",
		"FIXED0040_X9_50G",
		"FIXED0400_X9_50G",
		"FIXED0800_X9_50G",
		"FIXED1200_X9_50G",
		"FIXED1600_X9_50G",
		"FIXED2000_X9_50G",
		"FIXED2400_X9_50G",
		"FIXED2800_X9_50G",
		"FIXED3200_X9_50G",
		"FIXED3600_X9_50G",
		"FIXED4000_X9_50G",
		"STANDARD_VM_FIXED0100_X9_50G",
		"STANDARD_VM_FIXED0200_X9_50G",
		"STANDARD_VM_FIXED0300_X9_50G",
		"STANDARD_VM_FIXED0400_X9_50G",
		"STANDARD_VM_FIXED0500_X9_50G",
		"STANDARD_VM_FIXED0600_X9_50G",
		"STANDARD_VM_FIXED0700_X9_50G",
		"STANDARD_VM_FIXED0800_X9_50G",
		"STANDARD_VM_FIXED0900_X9_50G",
		"STANDARD_VM_FIXED1000_X9_50G",
		"STANDARD_VM_FIXED1100_X9_50G",
		"STANDARD_VM_FIXED1200_X9_50G",
		"STANDARD_VM_FIXED1300_X9_50G",
		"STANDARD_VM_FIXED1400_X9_50G",
		"STANDARD_VM_FIXED1500_X9_50G",
		"STANDARD_VM_FIXED1600_X9_50G",
		"STANDARD_VM_FIXED1700_X9_50G",
		"STANDARD_VM_FIXED1800_X9_50G",
		"STANDARD_VM_FIXED1900_X9_50G",
		"STANDARD_VM_FIXED2000_X9_50G",
		"STANDARD_VM_FIXED2100_X9_50G",
		"STANDARD_VM_FIXED2200_X9_50G",
		"STANDARD_VM_FIXED2300_X9_50G",
		"STANDARD_VM_FIXED2400_X9_50G",
		"STANDARD_VM_FIXED2500_X9_50G",
		"STANDARD_VM_FIXED2600_X9_50G",
		"STANDARD_VM_FIXED2700_X9_50G",
		"STANDARD_VM_FIXED2800_X9_50G",
		"STANDARD_VM_FIXED2900_X9_50G",
		"STANDARD_VM_FIXED3000_X9_50G",
		"STANDARD_VM_FIXED3100_X9_50G",
		"STANDARD_VM_FIXED3200_X9_50G",
		"STANDARD_VM_FIXED3300_X9_50G",
		"STANDARD_VM_FIXED3400_X9_50G",
		"STANDARD_VM_FIXED3500_X9_50G",
		"STANDARD_VM_FIXED3600_X9_50G",
		"STANDARD_VM_FIXED3700_X9_50G",
		"STANDARD_VM_FIXED3800_X9_50G",
		"STANDARD_VM_FIXED3900_X9_50G",
		"STANDARD_VM_FIXED4000_X9_50G",
		"ENTIREHOST_X9_50G",
	}
}
