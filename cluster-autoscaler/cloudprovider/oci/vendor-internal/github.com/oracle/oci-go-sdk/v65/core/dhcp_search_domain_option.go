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
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// DhcpSearchDomainOption DHCP option for specifying a search domain name for DNS queries. For more information, see
// DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm).
type DhcpSearchDomainOption struct {

	// A single search domain name according to RFC 952 (https://tools.ietf.org/html/rfc952)
	// and RFC 1123 (https://tools.ietf.org/html/rfc1123). During a DNS query,
	// the OS will append this search domain name to the value being queried.
	// If you set DhcpDnsOption to `VcnLocalPlusInternet`,
	// and you assign a DNS label to the VCN during creation, the search domain name in the
	// VCN's default set of DHCP options is automatically set to the VCN domain
	// (for example, `vcn1.oraclevcn.com`).
	// If you don't want to use a search domain name, omit this option from the
	// set of DHCP options. Do not include this option with an empty list
	// of search domain names, or with an empty string as the value for any search
	// domain name.
	SearchDomainNames []string `mandatory:"true" json:"searchDomainNames"`
}

func (m DhcpSearchDomainOption) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DhcpSearchDomainOption) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m DhcpSearchDomainOption) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeDhcpSearchDomainOption DhcpSearchDomainOption
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeDhcpSearchDomainOption
	}{
		"SearchDomain",
		(MarshalTypeDhcpSearchDomainOption)(m),
	}

	return json.Marshal(&s)
}
