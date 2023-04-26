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

// ReverseConnectionConfiguration Reverse connection configuration details for the private endpoint.
// When reverse connection support is enabled, the private endpoint allows reverse connections to be
// established to the customer VCN. The packets traveling from the service's VCN to the customer's
// VCN in a reverse connection use a different source IP address than the private endpoint's IP address.
type ReverseConnectionConfiguration struct {

	// The reverse connection configuration's current state.
	LifecycleState ReverseConnectionConfigurationLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The list of IP addresses in the customer VCN to be used as a source IP for reverse connection packets
	// traveling from the service's VCN to the customer's VCN.
	ReverseConnectionsSourceIps []ReverseConnectionsSourceIpDetails `mandatory:"false" json:"reverseConnectionsSourceIps"`

	// Whether a DNS proxy is configured for the reverse connection. If the service
	// does not intend to use DNS FQDN to communicate to customer endpoints, set this to `false`.
	// Example: `false`
	IsProxyEnabled *bool `mandatory:"false" json:"isProxyEnabled"`

	// List of proxies to be spawned for this reverse connection. All proxy types specified in
	// this list will be spawned and they will use the proxy IP address.
	// If not specified, the field will default back to a list with DNS proxy only.
	ProxyType []ReverseConnectionConfigurationProxyTypeEnum `mandatory:"false" json:"proxyType,omitempty"`

	// The IP address in the service VCN to be used to reach the reverse connection
	// proxies - DNS/SCAN.
	ProxyIp *string `mandatory:"false" json:"proxyIp"`

	// The IP address in the service VCN to be used to reach the reverse connection
	// proxies - DNS/SCAN.
	// This field will be deprecated in favor of proxyIp in future.
	DnsProxyIp *string `mandatory:"false" json:"dnsProxyIp"`

	// The context in which the DNS proxy will resolve the DNS queries. The default is `SERVICE`.
	// For example: if the service does not know the specific DNS zones for the customer VCNs, set
	// this to `CUSTOMER`, and set `excludedDnsZones` to the list of DNS zones in your service
	// provider VCN.
	// Allowed values:
	//  * `SERVICE` : All DNS queries will be resolved within the service VCN's DNS context,
	//    unless the FQDN belongs to one of zones in the `excludedDnsZones` list.
	//  * `CUSTOMER` : All DNS queries will be resolved within the customer VCN's DNS context,
	//    unless the FQDN belongs to one of zones in the `excludedDnsZones` list.
	DefaultDnsResolutionContext ReverseConnectionConfigurationDefaultDnsResolutionContextEnum `mandatory:"false" json:"defaultDnsResolutionContext,omitempty"`

	// List of DNS zones to exclude from the default DNS resolution context.
	ExcludedDnsZones []string `mandatory:"false" json:"excludedDnsZones"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the service's subnet where
	// the DNS proxy endpoint will be created.
	ServiceSubnetId *string `mandatory:"false" json:"serviceSubnetId"`

	// A list of the OCIDs of the network security groups that the reverse connection's VNIC belongs to.
	// For more information about NSGs, see
	// NetworkSecurityGroup.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// Number of customer endpoints that the service provider expects to establish connections to using this RCE. The default is 0.
	// When non-zero value is specified, reverse connection configuration will be allocated with a list of CIDRs, from
	// which NAT IP addresses will be allocated. These list of CIDRs will not be shared by other reverse
	// connection enabled private endpoints.
	// When zero is specified, reverse connection configuration will get NAT IP addresses from common pool of CIDRs,
	// which will be shared with other reverse connection enabled private endpoints.
	// If the private endpoint was enabled with reverse connection with 0 already, the field is not updatable.
	// The size may not be updated with smaller number than previously specified value, but may be increased.
	CustomerEndpointsSize *int `mandatory:"false" json:"customerEndpointsSize"`

	// List of CIDRs that this reverse connection configuration will allocate the NAT IP addresses from.
	// CIDRs on this list is guaranteed to be not shared by other reverse connection enabled private endpoints.
	ReverseConnectionNatIpCidrs []string `mandatory:"false" json:"reverseConnectionNatIpCidrs"`

	// Layer 4 transport protocol to be used when resolving DNS queries within the default DNS resolution context.
	DefaultDnsContextTransport ReverseConnectionConfigurationDefaultDnsContextTransportEnum `mandatory:"false" json:"defaultDnsContextTransport,omitempty"`

	// Whether the reverse connection should be configured with single IP. If the service
	// does not intend to use single IP for both forward and reverse connection, set this to `false`.
	// Example: `false`
	IsSingleIpEnabled *bool `mandatory:"false" json:"isSingleIpEnabled"`
}

func (m ReverseConnectionConfiguration) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ReverseConnectionConfiguration) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingReverseConnectionConfigurationLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetReverseConnectionConfigurationLifecycleStateEnumStringValues(), ",")))
	}
	for _, val := range m.ProxyType {
		if _, ok := mappingReverseConnectionConfigurationProxyTypeEnum[string(val)]; !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ProxyType: %s. Supported values are: %s.", val, strings.Join(GetReverseConnectionConfigurationProxyTypeEnumStringValues(), ",")))
		}
	}

	if _, ok := mappingReverseConnectionConfigurationDefaultDnsResolutionContextEnum[string(m.DefaultDnsResolutionContext)]; !ok && m.DefaultDnsResolutionContext != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DefaultDnsResolutionContext: %s. Supported values are: %s.", m.DefaultDnsResolutionContext, strings.Join(GetReverseConnectionConfigurationDefaultDnsResolutionContextEnumStringValues(), ",")))
	}
	if _, ok := mappingReverseConnectionConfigurationDefaultDnsContextTransportEnum[string(m.DefaultDnsContextTransport)]; !ok && m.DefaultDnsContextTransport != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DefaultDnsContextTransport: %s. Supported values are: %s.", m.DefaultDnsContextTransport, strings.Join(GetReverseConnectionConfigurationDefaultDnsContextTransportEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ReverseConnectionConfigurationLifecycleStateEnum Enum with underlying type: string
type ReverseConnectionConfigurationLifecycleStateEnum string

// Set of constants representing the allowable values for ReverseConnectionConfigurationLifecycleStateEnum
const (
	ReverseConnectionConfigurationLifecycleStateProvisioning ReverseConnectionConfigurationLifecycleStateEnum = "PROVISIONING"
	ReverseConnectionConfigurationLifecycleStateAvailable    ReverseConnectionConfigurationLifecycleStateEnum = "AVAILABLE"
	ReverseConnectionConfigurationLifecycleStateUpdating     ReverseConnectionConfigurationLifecycleStateEnum = "UPDATING"
	ReverseConnectionConfigurationLifecycleStateTerminating  ReverseConnectionConfigurationLifecycleStateEnum = "TERMINATING"
	ReverseConnectionConfigurationLifecycleStateTerminated   ReverseConnectionConfigurationLifecycleStateEnum = "TERMINATED"
	ReverseConnectionConfigurationLifecycleStateFailed       ReverseConnectionConfigurationLifecycleStateEnum = "FAILED"
)

var mappingReverseConnectionConfigurationLifecycleStateEnum = map[string]ReverseConnectionConfigurationLifecycleStateEnum{
	"PROVISIONING": ReverseConnectionConfigurationLifecycleStateProvisioning,
	"AVAILABLE":    ReverseConnectionConfigurationLifecycleStateAvailable,
	"UPDATING":     ReverseConnectionConfigurationLifecycleStateUpdating,
	"TERMINATING":  ReverseConnectionConfigurationLifecycleStateTerminating,
	"TERMINATED":   ReverseConnectionConfigurationLifecycleStateTerminated,
	"FAILED":       ReverseConnectionConfigurationLifecycleStateFailed,
}

// GetReverseConnectionConfigurationLifecycleStateEnumValues Enumerates the set of values for ReverseConnectionConfigurationLifecycleStateEnum
func GetReverseConnectionConfigurationLifecycleStateEnumValues() []ReverseConnectionConfigurationLifecycleStateEnum {
	values := make([]ReverseConnectionConfigurationLifecycleStateEnum, 0)
	for _, v := range mappingReverseConnectionConfigurationLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetReverseConnectionConfigurationLifecycleStateEnumStringValues Enumerates the set of values in String for ReverseConnectionConfigurationLifecycleStateEnum
func GetReverseConnectionConfigurationLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"UPDATING",
		"TERMINATING",
		"TERMINATED",
		"FAILED",
	}
}

// ReverseConnectionConfigurationProxyTypeEnum Enum with underlying type: string
type ReverseConnectionConfigurationProxyTypeEnum string

// Set of constants representing the allowable values for ReverseConnectionConfigurationProxyTypeEnum
const (
	ReverseConnectionConfigurationProxyTypeDns  ReverseConnectionConfigurationProxyTypeEnum = "DNS"
	ReverseConnectionConfigurationProxyTypeScan ReverseConnectionConfigurationProxyTypeEnum = "SCAN"
)

var mappingReverseConnectionConfigurationProxyTypeEnum = map[string]ReverseConnectionConfigurationProxyTypeEnum{
	"DNS":  ReverseConnectionConfigurationProxyTypeDns,
	"SCAN": ReverseConnectionConfigurationProxyTypeScan,
}

// GetReverseConnectionConfigurationProxyTypeEnumValues Enumerates the set of values for ReverseConnectionConfigurationProxyTypeEnum
func GetReverseConnectionConfigurationProxyTypeEnumValues() []ReverseConnectionConfigurationProxyTypeEnum {
	values := make([]ReverseConnectionConfigurationProxyTypeEnum, 0)
	for _, v := range mappingReverseConnectionConfigurationProxyTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetReverseConnectionConfigurationProxyTypeEnumStringValues Enumerates the set of values in String for ReverseConnectionConfigurationProxyTypeEnum
func GetReverseConnectionConfigurationProxyTypeEnumStringValues() []string {
	return []string{
		"DNS",
		"SCAN",
	}
}

// ReverseConnectionConfigurationDefaultDnsResolutionContextEnum Enum with underlying type: string
type ReverseConnectionConfigurationDefaultDnsResolutionContextEnum string

// Set of constants representing the allowable values for ReverseConnectionConfigurationDefaultDnsResolutionContextEnum
const (
	ReverseConnectionConfigurationDefaultDnsResolutionContextService  ReverseConnectionConfigurationDefaultDnsResolutionContextEnum = "SERVICE"
	ReverseConnectionConfigurationDefaultDnsResolutionContextCustomer ReverseConnectionConfigurationDefaultDnsResolutionContextEnum = "CUSTOMER"
)

var mappingReverseConnectionConfigurationDefaultDnsResolutionContextEnum = map[string]ReverseConnectionConfigurationDefaultDnsResolutionContextEnum{
	"SERVICE":  ReverseConnectionConfigurationDefaultDnsResolutionContextService,
	"CUSTOMER": ReverseConnectionConfigurationDefaultDnsResolutionContextCustomer,
}

// GetReverseConnectionConfigurationDefaultDnsResolutionContextEnumValues Enumerates the set of values for ReverseConnectionConfigurationDefaultDnsResolutionContextEnum
func GetReverseConnectionConfigurationDefaultDnsResolutionContextEnumValues() []ReverseConnectionConfigurationDefaultDnsResolutionContextEnum {
	values := make([]ReverseConnectionConfigurationDefaultDnsResolutionContextEnum, 0)
	for _, v := range mappingReverseConnectionConfigurationDefaultDnsResolutionContextEnum {
		values = append(values, v)
	}
	return values
}

// GetReverseConnectionConfigurationDefaultDnsResolutionContextEnumStringValues Enumerates the set of values in String for ReverseConnectionConfigurationDefaultDnsResolutionContextEnum
func GetReverseConnectionConfigurationDefaultDnsResolutionContextEnumStringValues() []string {
	return []string{
		"SERVICE",
		"CUSTOMER",
	}
}

// ReverseConnectionConfigurationDefaultDnsContextTransportEnum Enum with underlying type: string
type ReverseConnectionConfigurationDefaultDnsContextTransportEnum string

// Set of constants representing the allowable values for ReverseConnectionConfigurationDefaultDnsContextTransportEnum
const (
	ReverseConnectionConfigurationDefaultDnsContextTransportTcp ReverseConnectionConfigurationDefaultDnsContextTransportEnum = "TCP"
	ReverseConnectionConfigurationDefaultDnsContextTransportUdp ReverseConnectionConfigurationDefaultDnsContextTransportEnum = "UDP"
)

var mappingReverseConnectionConfigurationDefaultDnsContextTransportEnum = map[string]ReverseConnectionConfigurationDefaultDnsContextTransportEnum{
	"TCP": ReverseConnectionConfigurationDefaultDnsContextTransportTcp,
	"UDP": ReverseConnectionConfigurationDefaultDnsContextTransportUdp,
}

// GetReverseConnectionConfigurationDefaultDnsContextTransportEnumValues Enumerates the set of values for ReverseConnectionConfigurationDefaultDnsContextTransportEnum
func GetReverseConnectionConfigurationDefaultDnsContextTransportEnumValues() []ReverseConnectionConfigurationDefaultDnsContextTransportEnum {
	values := make([]ReverseConnectionConfigurationDefaultDnsContextTransportEnum, 0)
	for _, v := range mappingReverseConnectionConfigurationDefaultDnsContextTransportEnum {
		values = append(values, v)
	}
	return values
}

// GetReverseConnectionConfigurationDefaultDnsContextTransportEnumStringValues Enumerates the set of values in String for ReverseConnectionConfigurationDefaultDnsContextTransportEnum
func GetReverseConnectionConfigurationDefaultDnsContextTransportEnumStringValues() []string {
	return []string{
		"TCP",
		"UDP",
	}
}
