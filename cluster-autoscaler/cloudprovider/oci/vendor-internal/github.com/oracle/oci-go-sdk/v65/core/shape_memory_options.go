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

// ShapeMemoryOptions For a flexible shape, the amount of memory available for instances that use this shape.
// If this field is null, then this shape has a fixed amount of memory equivalent to `memoryInGBs`.
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

	// The maximum amount of memory per NUMA node, in gigabytes.
	MaxPerNumaNodeInGBs *float32 `mandatory:"false" json:"maxPerNumaNodeInGBs"`
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
