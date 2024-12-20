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

// ShapeMemoryOptions Memory properties.
type ShapeMemoryOptions struct {

	// The minimum amount of memory, in gigabytes.
	MinInGBs *float32 `mandatory:"false" json:"minInGBs"`

	// The maximum amount of memory, in gigabytes.
	MaxInGBs *float32 `mandatory:"false" json:"maxInGBs"`

	// The default amount of memory per OCPU available for this shape, in gigabytes.
	DefaultPerOcpuInGBs *float32 `mandatory:"false" json:"defaultPerOcpuInGBs"`

	// The minimum amount of memory per OCPU available for this shape, in gigabytes.
	MinPerOcpuInGBs *float32 `mandatory:"false" json:"minPerOcpuInGBs"`

	// The maximum amount of memory per OCPU available for this shape, in gigabytes.
	MaxPerOcpuInGBs *float32 `mandatory:"false" json:"maxPerOcpuInGBs"`
}

func (m ShapeMemoryOptions) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ShapeMemoryOptions) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
