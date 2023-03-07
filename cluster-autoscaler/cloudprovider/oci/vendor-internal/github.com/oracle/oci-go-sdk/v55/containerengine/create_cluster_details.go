// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// CreateClusterDetails The properties that define a request to create a cluster.
type CreateClusterDetails struct {

	// The name of the cluster. Avoid entering confidential information.
	Name *string `mandatory:"true" json:"name"`

	// The OCID of the compartment in which to create the cluster.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the virtual cloud network (VCN) in which to create the cluster.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// The version of Kubernetes to install into the cluster masters.
	KubernetesVersion *string `mandatory:"true" json:"kubernetesVersion"`

	// The network configuration for access to the Cluster control plane.
	EndpointConfig *CreateClusterEndpointConfigDetails `mandatory:"false" json:"endpointConfig"`

	// The OCID of the KMS key to be used as the master encryption key for Kubernetes secret encryption.
	// When used, `kubernetesVersion` must be at least `v1.13.0`.
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Optional attributes for the cluster.
	Options *ClusterCreateOptions `mandatory:"false" json:"options"`

	// The image verification policy for signature validation. Once a policy is created and enabled with
	// one or more kms keys, the policy will ensure all images deployed has been signed with the key(s)
	// attached to the policy.
	ImagePolicyConfig *CreateImagePolicyConfigDetails `mandatory:"false" json:"imagePolicyConfig"`
}

func (m CreateClusterDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateClusterDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
