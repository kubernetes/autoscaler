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

// ShapeNetworkBandwidthOptions Properties of network bandwidth.
type ShapeNetworkBandwidthOptions struct {

	// The minimum amount of networking bandwidth, in gigabits per second.
	MinInGbps *float32 `mandatory:"false" json:"minInGbps"`

	// The maximum amount of networking bandwidth, in gigabits per second.
	MaxInGbps *float32 `mandatory:"false" json:"maxInGbps"`

	// The default amount of networking bandwidth per OCPU, in gigabits per second.
	DefaultPerOcpuInGbps *float32 `mandatory:"false" json:"defaultPerOcpuInGbps"`
}

func (m ShapeNetworkBandwidthOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ShapeNetworkBandwidthOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
