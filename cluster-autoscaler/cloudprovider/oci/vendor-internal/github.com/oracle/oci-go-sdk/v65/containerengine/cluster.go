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
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// Cluster A Kubernetes cluster. Avoid entering confidential information.
type Cluster struct {

	// The OCID of the cluster.
	Id *string `mandatory:"false" json:"id"`

	// The name of the cluster.
	Name *string `mandatory:"false" json:"name"`

	// The OCID of the compartment in which the cluster exists.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The network configuration for access to the Cluster control plane.
	EndpointConfig *ClusterEndpointConfig `mandatory:"false" json:"endpointConfig"`

	// The OCID of the virtual cloud network (VCN) in which the cluster exists.
	VcnId *string `mandatory:"false" json:"vcnId"`

	// The version of Kubernetes running on the cluster masters.
	KubernetesVersion *string `mandatory:"false" json:"kubernetesVersion"`

	// The OCID of the KMS key to be used as the master encryption key for Kubernetes secret encryption.
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`

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

	// The state of the cluster masters. For more information, see Monitoring Clusters (https://docs.cloud.oracle.com/Content/ContEng/Tasks/contengmonitoringclusters.htm)
	LifecycleState ClusterLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Details about the state of the cluster masters.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// Endpoints served up by the cluster masters.
	Endpoints *ClusterEndpoints `mandatory:"false" json:"endpoints"`

	// Available Kubernetes versions to which the clusters masters may be upgraded.
	AvailableKubernetesUpgrades []string `mandatory:"false" json:"availableKubernetesUpgrades"`

	// The image verification policy for signature validation.
	ImagePolicyConfig *ImagePolicyConfig `mandatory:"false" json:"imagePolicyConfig"`

	// Available CNIs and network options for existing and new node pools of the cluster
	ClusterPodNetworkOptions []ClusterPodNetworkOptionDetails `mandatory:"false" json:"clusterPodNetworkOptions"`

	// Type of cluster
	Type ClusterTypeEnum `mandatory:"false" json:"type,omitempty"`
}

func (m Cluster) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m Cluster) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingClusterLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetClusterLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingClusterTypeEnum(string(m.Type)); !ok && m.Type != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Type: %s. Supported values are: %s.", m.Type, strings.Join(GetClusterTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *Cluster) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		Id                          *string                           `json:"id"`
		Name                        *string                           `json:"name"`
		CompartmentId               *string                           `json:"compartmentId"`
		EndpointConfig              *ClusterEndpointConfig            `json:"endpointConfig"`
		VcnId                       *string                           `json:"vcnId"`
		KubernetesVersion           *string                           `json:"kubernetesVersion"`
		KmsKeyId                    *string                           `json:"kmsKeyId"`
		FreeformTags                map[string]string                 `json:"freeformTags"`
		DefinedTags                 map[string]map[string]interface{} `json:"definedTags"`
		SystemTags                  map[string]map[string]interface{} `json:"systemTags"`
		Options                     *ClusterCreateOptions             `json:"options"`
		Metadata                    *ClusterMetadata                  `json:"metadata"`
		LifecycleState              ClusterLifecycleStateEnum         `json:"lifecycleState"`
		LifecycleDetails            *string                           `json:"lifecycleDetails"`
		Endpoints                   *ClusterEndpoints                 `json:"endpoints"`
		AvailableKubernetesUpgrades []string                          `json:"availableKubernetesUpgrades"`
		ImagePolicyConfig           *ImagePolicyConfig                `json:"imagePolicyConfig"`
		ClusterPodNetworkOptions    []clusterpodnetworkoptiondetails  `json:"clusterPodNetworkOptions"`
		Type                        ClusterTypeEnum                   `json:"type"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.Id = model.Id

	m.Name = model.Name

	m.CompartmentId = model.CompartmentId

	m.EndpointConfig = model.EndpointConfig

	m.VcnId = model.VcnId

	m.KubernetesVersion = model.KubernetesVersion

	m.KmsKeyId = model.KmsKeyId

	m.FreeformTags = model.FreeformTags

	m.DefinedTags = model.DefinedTags

	m.SystemTags = model.SystemTags

	m.Options = model.Options

	m.Metadata = model.Metadata

	m.LifecycleState = model.LifecycleState

	m.LifecycleDetails = model.LifecycleDetails

	m.Endpoints = model.Endpoints

	m.AvailableKubernetesUpgrades = make([]string, len(model.AvailableKubernetesUpgrades))
	copy(m.AvailableKubernetesUpgrades, model.AvailableKubernetesUpgrades)
	m.ImagePolicyConfig = model.ImagePolicyConfig

	m.ClusterPodNetworkOptions = make([]ClusterPodNetworkOptionDetails, len(model.ClusterPodNetworkOptions))
	for i, n := range model.ClusterPodNetworkOptions {
		nn, e = n.UnmarshalPolymorphicJSON(n.JsonData)
		if e != nil {
			return e
		}
		if nn != nil {
			m.ClusterPodNetworkOptions[i] = nn.(ClusterPodNetworkOptionDetails)
		} else {
			m.ClusterPodNetworkOptions[i] = nil
		}
	}
	m.Type = model.Type

	return
}
