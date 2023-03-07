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

// ClusterSummary The properties that define a cluster summary.
type ClusterSummary struct {

	// The OCID of the cluster.
	Id *string `mandatory:"false" json:"id"`

	// The name of the cluster.
	Name *string `mandatory:"false" json:"name"`

	// The OCID of the compartment in which the cluster exists.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The network configuration for access to the Cluster control plane.
	EndpointConfig *ClusterEndpointConfig `mandatory:"false" json:"endpointConfig"`

	// The OCID of the virtual cloud network (VCN) in which the cluster exists
	VcnId *string `mandatory:"false" json:"vcnId"`

	// The version of Kubernetes running on the cluster masters.
	KubernetesVersion *string `mandatory:"false" json:"kubernetesVersion"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Usage of system tag keys. These predefined keys are scoped to namespaces.
	// Example: `{"orcl-cloud": {"free-tier-retained": "true"}}`
	SystemTags map[string]map[string]interface{} `mandatory:"false" json:"systemTags"`

	// Optional attributes for the cluster.
	Options *ClusterCreateOptions `mandatory:"false" json:"options"`

	// Metadata about the cluster.
	Metadata *ClusterMetadata `mandatory:"false" json:"metadata"`

	// The state of the cluster masters.
	LifecycleState ClusterLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Details about the state of the cluster masters.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// Endpoints served up by the cluster masters.
	Endpoints *ClusterEndpoints `mandatory:"false" json:"endpoints"`

	// Available Kubernetes versions to which the clusters masters may be upgraded.
	AvailableKubernetesUpgrades []string `mandatory:"false" json:"availableKubernetesUpgrades"`

	// The image verification policy for signature validation.
	ImagePolicyConfig *ImagePolicyConfig `mandatory:"false" json:"imagePolicyConfig"`
}

func (m ClusterSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ClusterSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingClusterLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetClusterLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ClusterSummaryLifecycleStateEnum is an alias to type: ClusterLifecycleStateEnum
// Consider using ClusterLifecycleStateEnum instead
// Deprecated
type ClusterSummaryLifecycleStateEnum = ClusterLifecycleStateEnum

// Set of constants representing the allowable values for ClusterLifecycleStateEnum
// Deprecated
const (
	ClusterSummaryLifecycleStateCreating ClusterLifecycleStateEnum = "CREATING"
	ClusterSummaryLifecycleStateActive   ClusterLifecycleStateEnum = "ACTIVE"
	ClusterSummaryLifecycleStateFailed   ClusterLifecycleStateEnum = "FAILED"
	ClusterSummaryLifecycleStateDeleting ClusterLifecycleStateEnum = "DELETING"
	ClusterSummaryLifecycleStateDeleted  ClusterLifecycleStateEnum = "DELETED"
	ClusterSummaryLifecycleStateUpdating ClusterLifecycleStateEnum = "UPDATING"
)

// GetClusterSummaryLifecycleStateEnumValues Enumerates the set of values for ClusterLifecycleStateEnum
// Consider using GetClusterLifecycleStateEnumValue
// Deprecated
var GetClusterSummaryLifecycleStateEnumValues = GetClusterLifecycleStateEnumValues
