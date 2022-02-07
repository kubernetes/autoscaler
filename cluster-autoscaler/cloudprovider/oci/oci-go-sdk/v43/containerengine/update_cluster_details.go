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

// UpdateClusterDetails The properties that define a request to update a cluster.
type UpdateClusterDetails struct {

	// The new name for the cluster. Avoid entering confidential information.
	Name *string `mandatory:"false" json:"name"`

	// The version of Kubernetes to which the cluster masters should be upgraded.
	KubernetesVersion *string `mandatory:"false" json:"kubernetesVersion"`

	Options *UpdateClusterOptionsDetails `mandatory:"false" json:"options"`

	// The image verification policy for signature validation. Once a policy is created and enabled with
	// one or more kms keys, the policy will ensure all images deployed has been signed with the key(s)
	// attached to the policy.
	ImagePolicyConfig *UpdateImagePolicyConfigDetails `mandatory:"false" json:"imagePolicyConfig"`
}

func (m UpdateClusterDetails) String() string {
	return common.PointerString(m)
}
