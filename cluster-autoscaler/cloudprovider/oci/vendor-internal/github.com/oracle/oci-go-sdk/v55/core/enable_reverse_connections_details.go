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

// EnableReverseConnectionsDetails Details for enabling reverse connections for a private endpoint.
type EnableReverseConnectionsDetails struct {

	// A list of IP addresses in the customer VCN to be used as the source IPs for reverse connection packets
	// traveling from the service's VCN to the customer's VCN. If no list is specified or
	// an empty list is provided, an IP address will be chosen from the customer subnet's CIDR.
	ReverseConnectionsSourceIps []ReverseConnectionsSourceIpDetails `mandatory:"false" json:"reverseConnectionsSourceIps"`

	// Whether a proxy should be configured for the reverse connection. If the service
	// does not intend to use any proxy, set this to `false`.
	// Example: `false`
	IsProxyEnabled *bool `mandatory:"false" json:"isProxyEnabled"`

	// List of proxies to be spawned for this reverse connection. All proxy types specified in
	// this list will be spawned and they will use the proxyIp(below) as the address.
	// If not specified, the field will default back to a list with DNS proxy only.
	ProxyType []EnableReverseConnectionsDetailsProxyTypeEnum `mandatory:"false" json:"proxyType,omitempty"`

	// The IP address in the service VCN to be used to reach the reverse connection proxy
	// services DNS & SCAN proxy. If no value is provided, an available IP address will
	// be chosen from the service subnet's CIDR.
	ProxyIp *string `mandatory:"false" json:"proxyIp"`

	// The IP address in the service VCN to be used to reach the DNS proxy that resolves the
	// customer FQDN for reverse connections. If no value is provided, an available IP address will
	// be chosen from the service subnet's CIDR.
	// This field will be deprecated in favor of proxyIp in future.
	DnsProxyIp *string `mandatory:"false" json:"dnsProxyIp"`

	// The context in which the DNS proxy will resolve the DNS queries. The default is `SERVICE`.
	// For example: if the service does not know the specific DNS zones for the customer VCNs, set
	// this to `CUSTOMER`, and set `excludedDnsZones` to the list of DNS zones in your service
	// provider VCN.
	// Allowed values:
	//  * `SERVICE`: All DNS queries will be resolved within the service VCN's DNS context,
	//    unless the FQDN belongs to one of zones in the `excludedDnsZones` list.
	//  * `CUSTOMER`: All DNS queries will be resolved within the customer VCN's DNS context,
	//    unless the FQDN belongs to one of zones in the `excludedDnsZones` list.
	DefaultDnsResolutionContext EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum `mandatory:"false" json:"defaultDnsResolutionContext,omitempty"`

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

	// Layer 4 transport protocol to be used when resolving DNS queries within the default DNS resolution context.
	DefaultDnsContextTransport EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum `mandatory:"false" json:"defaultDnsContextTransport,omitempty"`

	// List of CIDRs that this reverse connection configuration will allocate the NAT IP addresses from.
	// CIDRs on this list should not be shared by other reverse connection enabled private endpoints.
	// When not specified, if the customerEndpointSize is non null, reverse connection configuration will get
	// NAT IP addresses from the dedicated pool of CIDRs, else will get specified from the common pool of CIDRs.
	// This field cannot be specified if the customerEndpointsSize field is non null and vice versa.
	ReverseConnectionNatIpCidrs []string `mandatory:"false" json:"reverseConnectionNatIpCidrs"`

	// Whether the reverse connection should be configured with single Ip. If the service
	// does not intend to use single Ip for both forward and reverse connection, set this to `false`.
	// Example: `false`
	IsSingleIpEnabled *bool `mandatory:"false" json:"isSingleIpEnabled"`
}

func (m EnableReverseConnectionsDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m EnableReverseConnectionsDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	for _, val := range m.ProxyType {
		if _, ok := mappingEnableReverseConnectionsDetailsProxyTypeEnum[string(val)]; !ok && val != "" {
			errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ProxyType: %s. Supported values are: %s.", val, strings.Join(GetEnableReverseConnectionsDetailsProxyTypeEnumStringValues(), ",")))
		}
	}

	if _, ok := mappingEnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum[string(m.DefaultDnsResolutionContext)]; !ok && m.DefaultDnsResolutionContext != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DefaultDnsResolutionContext: %s. Supported values are: %s.", m.DefaultDnsResolutionContext, strings.Join(GetEnableReverseConnectionsDetailsDefaultDnsResolutionContextEnumStringValues(), ",")))
	}
	if _, ok := mappingEnableReverseConnectionsDetailsDefaultDnsContextTransportEnum[string(m.DefaultDnsContextTransport)]; !ok && m.DefaultDnsContextTransport != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DefaultDnsContextTransport: %s. Supported values are: %s.", m.DefaultDnsContextTransport, strings.Join(GetEnableReverseConnectionsDetailsDefaultDnsContextTransportEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// EnableReverseConnectionsDetailsProxyTypeEnum Enum with underlying type: string
type EnableReverseConnectionsDetailsProxyTypeEnum string

// Set of constants representing the allowable values for EnableReverseConnectionsDetailsProxyTypeEnum
const (
	EnableReverseConnectionsDetailsProxyTypeDns  EnableReverseConnectionsDetailsProxyTypeEnum = "DNS"
	EnableReverseConnectionsDetailsProxyTypeScan EnableReverseConnectionsDetailsProxyTypeEnum = "SCAN"
)

var mappingEnableReverseConnectionsDetailsProxyTypeEnum = map[string]EnableReverseConnectionsDetailsProxyTypeEnum{
	"DNS":  EnableReverseConnectionsDetailsProxyTypeDns,
	"SCAN": EnableReverseConnectionsDetailsProxyTypeScan,
}

// GetEnableReverseConnectionsDetailsProxyTypeEnumValues Enumerates the set of values for EnableReverseConnectionsDetailsProxyTypeEnum
func GetEnableReverseConnectionsDetailsProxyTypeEnumValues() []EnableReverseConnectionsDetailsProxyTypeEnum {
	values := make([]EnableReverseConnectionsDetailsProxyTypeEnum, 0)
	for _, v := range mappingEnableReverseConnectionsDetailsProxyTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetEnableReverseConnectionsDetailsProxyTypeEnumStringValues Enumerates the set of values in String for EnableReverseConnectionsDetailsProxyTypeEnum
func GetEnableReverseConnectionsDetailsProxyTypeEnumStringValues() []string {
	return []string{
		"DNS",
		"SCAN",
	}
}

// EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum Enum with underlying type: string
type EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum string

// Set of constants representing the allowable values for EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum
const (
	EnableReverseConnectionsDetailsDefaultDnsResolutionContextService  EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum = "SERVICE"
	EnableReverseConnectionsDetailsDefaultDnsResolutionContextCustomer EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum = "CUSTOMER"
)

var mappingEnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum = map[string]EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum{
	"SERVICE":  EnableReverseConnectionsDetailsDefaultDnsResolutionContextService,
	"CUSTOMER": EnableReverseConnectionsDetailsDefaultDnsResolutionContextCustomer,
}

// GetEnableReverseConnectionsDetailsDefaultDnsResolutionContextEnumValues Enumerates the set of values for EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum
func GetEnableReverseConnectionsDetailsDefaultDnsResolutionContextEnumValues() []EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum {
	values := make([]EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum, 0)
	for _, v := range mappingEnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum {
		values = append(values, v)
	}
	return values
}

// GetEnableReverseConnectionsDetailsDefaultDnsResolutionContextEnumStringValues Enumerates the set of values in String for EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum
func GetEnableReverseConnectionsDetailsDefaultDnsResolutionContextEnumStringValues() []string {
	return []string{
		"SERVICE",
		"CUSTOMER",
	}
}

// EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum Enum with underlying type: string
type EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum string

// Set of constants representing the allowable values for EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum
const (
	EnableReverseConnectionsDetailsDefaultDnsContextTransportTcp EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum = "TCP"
	EnableReverseConnectionsDetailsDefaultDnsContextTransportUdp EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum = "UDP"
)

var mappingEnableReverseConnectionsDetailsDefaultDnsContextTransportEnum = map[string]EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum{
	"TCP": EnableReverseConnectionsDetailsDefaultDnsContextTransportTcp,
	"UDP": EnableReverseConnectionsDetailsDefaultDnsContextTransportUdp,
}

// GetEnableReverseConnectionsDetailsDefaultDnsContextTransportEnumValues Enumerates the set of values for EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum
func GetEnableReverseConnectionsDetailsDefaultDnsContextTransportEnumValues() []EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum {
	values := make([]EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum, 0)
	for _, v := range mappingEnableReverseConnectionsDetailsDefaultDnsContextTransportEnum {
		values = append(values, v)
	}
	return values
}

// GetEnableReverseConnectionsDetailsDefaultDnsContextTransportEnumStringValues Enumerates the set of values in String for EnableReverseConnectionsDetailsDefaultDnsContextTransportEnum
func GetEnableReverseConnectionsDetailsDefaultDnsContextTransportEnumStringValues() []string {
	return []string{
		"TCP",
		"UDP",
	}
}
