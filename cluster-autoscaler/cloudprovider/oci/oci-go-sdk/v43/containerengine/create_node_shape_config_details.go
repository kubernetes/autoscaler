// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Container Engine for Kubernetes API
//
// API for the Container Engine for Kubernetes service. Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Container Engine for Kubernetes (https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// CreateNodeShapeConfigDetails The shape configuration of the nodes.
type CreateNodeShapeConfigDetails struct {

	// The total number of OCPUs available to each node in the node pool.
	// See here (https://docs.cloud.oracle.com/en-us/iaas/api/#/en/iaas/20160918/Shape/) for details.
	Ocpus *float32 `mandatory:"false" json:"ocpus"`

	// The total amount of memory available to each node, in gigabytes.
	MemoryInGBs *float32 `mandatory:"false" json:"memoryInGBs"`
}

func (m CreateNodeShapeConfigDetails) String() string {
	return common.PointerString(m)
}
