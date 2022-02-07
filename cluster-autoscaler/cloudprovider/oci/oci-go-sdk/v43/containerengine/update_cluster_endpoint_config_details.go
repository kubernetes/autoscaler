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

// UpdateClusterEndpointConfigDetails The properties that define a request to update a cluster endpoint config.
type UpdateClusterEndpointConfigDetails struct {

	// A list of the OCIDs of the network security groups (NSGs) to apply to the cluster endpoint. For more information about NSGs, see NetworkSecurityGroup.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// Whether the cluster should be assigned a public IP address. Defaults to false. If set to true on a private subnet, the cluster update will fail.
	IsPublicIpEnabled *bool `mandatory:"false" json:"isPublicIpEnabled"`
}

func (m UpdateClusterEndpointConfigDetails) String() string {
	return common.PointerString(m)
}
