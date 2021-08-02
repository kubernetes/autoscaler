// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// BgpSessionInfo Information for establishing a BGP session for the IPSec tunnel.
type BgpSessionInfo struct {

	// The IP address for the Oracle end of the inside tunnel interface.
	// If the tunnel's `routing` attribute is set to `BGP`
	// (see IPSecConnectionTunnel), this IP address
	// is required and used for the tunnel's BGP session.
	// If `routing` is instead set to `STATIC`, this IP address is optional. You can set this IP
	// address so you can troubleshoot or monitor the tunnel.
	// The value must be a /30 or /31.
	// Example: `10.0.0.4/31`
	OracleInterfaceIp *string `mandatory:"false" json:"oracleInterfaceIp"`

	// The IP address for the CPE end of the inside tunnel interface.
	// If the tunnel's `routing` attribute is set to `BGP`
	// (see IPSecConnectionTunnel), this IP address
	// is required and used for the tunnel's BGP session.
	// If `routing` is instead set to `STATIC`, this IP address is optional. You can set this IP
	// address so you can troubleshoot or monitor the tunnel.
	// The value must be a /30 or /31.
	// Example: `10.0.0.5/31`
	CustomerInterfaceIp *string `mandatory:"false" json:"customerInterfaceIp"`

	// The IPv6 address for the Oracle end of the inside tunnel interface. This IP address is optional.
	// If the tunnel's `routing` attribute is set to `BGP`
	// (see IPSecConnectionTunnel), this IP address
	// is used for the tunnel's BGP session.
	// If `routing` is instead set to `STATIC`, you can set this IP
	// address to troubleshoot or monitor the tunnel.
	// Only subnet masks from /64 up to /127 are allowed.
	// Example: `2001:db8::1/64`
	OracleInterfaceIpv6 *string `mandatory:"false" json:"oracleInterfaceIpv6"`

	// The IPv6 address for the CPE end of the inside tunnel interface. This IP address is optional.
	// If the tunnel's `routing` attribute is set to `BGP`
	// (see IPSecConnectionTunnel), this IP address
	// is used for the tunnel's BGP session.
	// If `routing` is instead set to `STATIC`, you can set this IP
	// address to troubleshoot or monitor the tunnel.
	// Only subnet masks from /64 up to /127 are allowed.
	// Example: `2001:db8::1/64`
	CustomerInterfaceIpv6 *string `mandatory:"false" json:"customerInterfaceIpv6"`

	// The Oracle BGP ASN.
	OracleBgpAsn *string `mandatory:"false" json:"oracleBgpAsn"`

	// If the tunnel's `routing` attribute is set to `BGP`
	// (see IPSecConnectionTunnel), this ASN
	// is required and used for the tunnel's BGP session. This is the ASN of the network on the
	// CPE end of the BGP session. Can be a 2-byte or 4-byte ASN. Uses "asplain" format.
	// If the tunnel uses static routing, the `customerBgpAsn` must be null.
	// Example: `12345` (2-byte) or `1587232876` (4-byte)
	CustomerBgpAsn *string `mandatory:"false" json:"customerBgpAsn"`

	// The state of the BGP session.
	BgpState BgpSessionInfoBgpStateEnum `mandatory:"false" json:"bgpState,omitempty"`

	// The state of the BGP IPv6 session.
	BgpIpv6State BgpSessionInfoBgpIpv6StateEnum `mandatory:"false" json:"bgpIpv6State,omitempty"`
}

func (m BgpSessionInfo) String() string {
	return common.PointerString(m)
}

// BgpSessionInfoBgpStateEnum Enum with underlying type: string
type BgpSessionInfoBgpStateEnum string

// Set of constants representing the allowable values for BgpSessionInfoBgpStateEnum
const (
	BgpSessionInfoBgpStateUp   BgpSessionInfoBgpStateEnum = "UP"
	BgpSessionInfoBgpStateDown BgpSessionInfoBgpStateEnum = "DOWN"
)

var mappingBgpSessionInfoBgpState = map[string]BgpSessionInfoBgpStateEnum{
	"UP":   BgpSessionInfoBgpStateUp,
	"DOWN": BgpSessionInfoBgpStateDown,
}

// GetBgpSessionInfoBgpStateEnumValues Enumerates the set of values for BgpSessionInfoBgpStateEnum
func GetBgpSessionInfoBgpStateEnumValues() []BgpSessionInfoBgpStateEnum {
	values := make([]BgpSessionInfoBgpStateEnum, 0)
	for _, v := range mappingBgpSessionInfoBgpState {
		values = append(values, v)
	}
	return values
}

// BgpSessionInfoBgpIpv6StateEnum Enum with underlying type: string
type BgpSessionInfoBgpIpv6StateEnum string

// Set of constants representing the allowable values for BgpSessionInfoBgpIpv6StateEnum
const (
	BgpSessionInfoBgpIpv6StateUp   BgpSessionInfoBgpIpv6StateEnum = "UP"
	BgpSessionInfoBgpIpv6StateDown BgpSessionInfoBgpIpv6StateEnum = "DOWN"
)

var mappingBgpSessionInfoBgpIpv6State = map[string]BgpSessionInfoBgpIpv6StateEnum{
	"UP":   BgpSessionInfoBgpIpv6StateUp,
	"DOWN": BgpSessionInfoBgpIpv6StateDown,
}

// GetBgpSessionInfoBgpIpv6StateEnumValues Enumerates the set of values for BgpSessionInfoBgpIpv6StateEnum
func GetBgpSessionInfoBgpIpv6StateEnumValues() []BgpSessionInfoBgpIpv6StateEnum {
	values := make([]BgpSessionInfoBgpIpv6StateEnum, 0)
	for _, v := range mappingBgpSessionInfoBgpIpv6State {
		values = append(values, v)
	}
	return values
}
