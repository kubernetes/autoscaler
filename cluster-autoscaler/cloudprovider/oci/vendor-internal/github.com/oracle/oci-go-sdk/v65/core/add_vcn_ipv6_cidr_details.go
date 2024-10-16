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

// AddVcnIpv6CidrDetails Details used when adding a ULA or private IPv6 prefix or an IPv6 GUA assigned by Oracle or a BYOIPv6 prefix. You can add only one of these per request.
type AddVcnIpv6CidrDetails struct {

	// This field is not required and should only be specified if a ULA or private IPv6 prefix is desired for VCN's private IP address space.
	// SeeIPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	// Example: `2001:0db8:0123::/48` or `fd00:1000:0:1::/64`
	Ipv6PrivateCidrBlock *string `mandatory:"false" json:"ipv6PrivateCidrBlock"`

	// Indicates whether Oracle will allocate an IPv6 GUA. Only one prefix of /56 size can be allocated by Oracle as a GUA.
	IsOracleGuaAllocationEnabled *bool `mandatory:"false" json:"isOracleGuaAllocationEnabled"`

	Byoipv6CidrDetail *Byoipv6CidrDetails `mandatory:"false" json:"byoipv6CidrDetail"`
}

func (m AddVcnIpv6CidrDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m AddVcnIpv6CidrDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
