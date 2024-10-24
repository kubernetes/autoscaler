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

// TunnelRouteSummary A summary of the routes advertised to and received from the on-premises network.
type TunnelRouteSummary struct {

	// The BGP network layer reachability information.
	Prefix *string `mandatory:"false" json:"prefix"`

	// The age of the route.
	Age *int64 `mandatory:"false" json:"age"`

	// Indicates this is the best route.
	IsBestPath *bool `mandatory:"false" json:"isBestPath"`

	// A list of ASNs in AS_Path.
	AsPath []int `mandatory:"false" json:"asPath"`

	// The source of the route advertisement.
	Advertiser TunnelRouteSummaryAdvertiserEnum `mandatory:"false" json:"advertiser,omitempty"`
}

func (m TunnelRouteSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m TunnelRouteSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingTunnelRouteSummaryAdvertiserEnum(string(m.Advertiser)); !ok && m.Advertiser != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Advertiser: %s. Supported values are: %s.", m.Advertiser, strings.Join(GetTunnelRouteSummaryAdvertiserEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// TunnelRouteSummaryAdvertiserEnum Enum with underlying type: string
type TunnelRouteSummaryAdvertiserEnum string

// Set of constants representing the allowable values for TunnelRouteSummaryAdvertiserEnum
const (
	TunnelRouteSummaryAdvertiserCustomer TunnelRouteSummaryAdvertiserEnum = "CUSTOMER"
	TunnelRouteSummaryAdvertiserOracle   TunnelRouteSummaryAdvertiserEnum = "ORACLE"
)

var mappingTunnelRouteSummaryAdvertiserEnum = map[string]TunnelRouteSummaryAdvertiserEnum{
	"CUSTOMER": TunnelRouteSummaryAdvertiserCustomer,
	"ORACLE":   TunnelRouteSummaryAdvertiserOracle,
}

var mappingTunnelRouteSummaryAdvertiserEnumLowerCase = map[string]TunnelRouteSummaryAdvertiserEnum{
	"customer": TunnelRouteSummaryAdvertiserCustomer,
	"oracle":   TunnelRouteSummaryAdvertiserOracle,
}

// GetTunnelRouteSummaryAdvertiserEnumValues Enumerates the set of values for TunnelRouteSummaryAdvertiserEnum
func GetTunnelRouteSummaryAdvertiserEnumValues() []TunnelRouteSummaryAdvertiserEnum {
	values := make([]TunnelRouteSummaryAdvertiserEnum, 0)
	for _, v := range mappingTunnelRouteSummaryAdvertiserEnum {
		values = append(values, v)
	}
	return values
}

// GetTunnelRouteSummaryAdvertiserEnumStringValues Enumerates the set of values in String for TunnelRouteSummaryAdvertiserEnum
func GetTunnelRouteSummaryAdvertiserEnumStringValues() []string {
	return []string{
		"CUSTOMER",
		"ORACLE",
	}
}

// GetMappingTunnelRouteSummaryAdvertiserEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingTunnelRouteSummaryAdvertiserEnum(val string) (TunnelRouteSummaryAdvertiserEnum, bool) {
	enum, ok := mappingTunnelRouteSummaryAdvertiserEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
