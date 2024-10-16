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

// IcmpOptions Optional and valid only for ICMP and ICMPv6. Use to specify a particular ICMP type and code
// as defined in:
// - ICMP Parameters (http://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml)
// - ICMPv6 Parameters (https://www.iana.org/assignments/icmpv6-parameters/icmpv6-parameters.xhtml)
// If you specify ICMP or ICMPv6 as the protocol but omit this object, then all ICMP types and
// codes are allowed. If you do provide this object, the type is required and the code is optional.
// To enable MTU negotiation for ingress internet traffic via IPv4, make sure to allow type 3 ("Destination
// Unreachable") code 4 ("Fragmentation Needed and Don't Fragment was Set"). If you need to specify
// multiple codes for a single type, create a separate security list rule for each.
type IcmpOptions struct {

	// The ICMP type.
	Type *int `mandatory:"true" json:"type"`

	// The ICMP code (optional).
	Code *int `mandatory:"false" json:"code"`
}

func (m IcmpOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m IcmpOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
