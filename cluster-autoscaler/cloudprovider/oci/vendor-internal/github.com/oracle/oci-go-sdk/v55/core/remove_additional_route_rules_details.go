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

// RemoveAdditionalRouteRulesDetails The configuration details for the remove rules operation. Only one of the three properties should be
// supplied to call the removeAdditionalRouteRules API.
type RemoveAdditionalRouteRulesDetails struct {

	// The list of route rule identifiers used for removing route rules from route table.
	AdditionalRouteRuleIds []string `mandatory:"false" json:"additionalRouteRuleIds"`

	// The list of destinations used for removing route rules from route table. This is only supplied
	// when additionalRouteRuleIds are not provided.
	Destinations []string `mandatory:"false" json:"destinations"`

	// The pairs of <destination, target> used for removing route rules from route table. This is only
	// supplied when additionalRouteRules and destinations are not provided.
	RouteDestinationRouteTargets []RouteDestinationRouteTargetDetails `mandatory:"false" json:"routeDestinationRouteTargets"`
}

func (m RemoveAdditionalRouteRulesDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m RemoveAdditionalRouteRulesDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
