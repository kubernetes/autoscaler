// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Kubernetes Engine API
//
// API for the Kubernetes Engine service (also known as the Container Engine for Kubernetes service). Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Kubernetes Engine (https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// NodePoolCyclingDetails Node Pool Cycling Details
type NodePoolCyclingDetails struct {

	// Maximum active nodes that would be terminated from nodepool during the cycling nodepool process.
	// OKE supports both integer and percentage input.
	// Defaults to 0, Ranges from 0 to Nodepool size or 0% to 100%
	MaximumUnavailable *string `mandatory:"false" json:"maximumUnavailable"`

	// Maximum additional new compute instances that would be temporarily created and added to nodepool during the cycling nodepool process.
	// OKE supports both integer and percentage input.
	// Defaults to 1, Ranges from 0 to Nodepool size or 0% to 100%
	MaximumSurge *string `mandatory:"false" json:"maximumSurge"`

	// If nodes in the nodepool will be cycled to have new changes.
	IsNodeCyclingEnabled *bool `mandatory:"false" json:"isNodeCyclingEnabled"`
}

func (m NodePoolCyclingDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m NodePoolCyclingDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
