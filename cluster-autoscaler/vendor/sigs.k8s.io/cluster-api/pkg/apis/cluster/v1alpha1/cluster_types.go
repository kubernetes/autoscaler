/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
)

const ClusterFinalizer = "cluster.cluster.k8s.io"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

/// [Cluster]
// Cluster is the Schema for the clusters API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

/// [Cluster]

/// [ClusterSpec]
// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// Cluster network configuration
	ClusterNetwork ClusterNetworkingConfig `json:"clusterNetwork"`

	// Provider-specific serialized configuration to use during
	// cluster creation. It is recommended that providers maintain
	// their own versioned API types that should be
	// serialized/deserialized from this field.
	// +optional
	ProviderSpec ProviderSpec `json:"providerSpec,omitempty"`
}

/// [ClusterSpec]

/// [ClusterNetworkingConfig]
// ClusterNetworkingConfig specifies the different networking
// parameters for a cluster.
type ClusterNetworkingConfig struct {
	// The network ranges from which service VIPs are allocated.
	Services NetworkRanges `json:"services"`

	// The network ranges from which Pod networks are allocated.
	Pods NetworkRanges `json:"pods"`

	// Domain name for services.
	ServiceDomain string `json:"serviceDomain"`
}

/// [ClusterNetworkingConfig]

/// [NetworkRanges]
// NetworkRanges represents ranges of network addresses.
type NetworkRanges struct {
	CIDRBlocks []string `json:"cidrBlocks"`
}

/// [NetworkRanges]

/// [ClusterStatus]
// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// APIEndpoint represents the endpoint to communicate with the IP.
	// +optional
	APIEndpoints []APIEndpoint `json:"apiEndpoints,omitempty"`

	// NB: Eventually we will redefine ErrorReason as ClusterStatusError once the
	// following issue is fixed.
	// https://github.com/kubernetes-incubator/apiserver-builder/issues/176

	// If set, indicates that there is a problem reconciling the
	// state, and will be set to a token value suitable for
	// programmatic interpretation.
	// +optional
	ErrorReason common.ClusterStatusError `json:"errorReason,omitempty"`

	// If set, indicates that there is a problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`

	// Provider-specific status.
	// It is recommended that providers maintain their
	// own versioned API types that should be
	// serialized/deserialized from this field.
	// +optional
	ProviderStatus *runtime.RawExtension `json:"providerStatus,omitempty"`
}

/// [ClusterStatus]

/// [APIEndpoint]
// APIEndpoint represents a reachable Kubernetes API endpoint.
type APIEndpoint struct {
	// The hostname on which the API server is serving.
	Host string `json:"host"`

	// The port on which the API server is serving.
	Port int `json:"port"`
}

/// [APIEndpoint]

func (o *Cluster) Validate() field.ErrorList {
	errors := field.ErrorList{}
	// perform validation here and add to errors using field.Invalid
	if o.Spec.ClusterNetwork.ServiceDomain == "" {
		errors = append(errors, field.Invalid(
			field.NewPath("Spec", "ClusterNetwork", "ServiceDomain"),
			o.Spec.ClusterNetwork.ServiceDomain,
			"invalid cluster configuration: missing Cluster.Spec.ClusterNetwork.ServiceDomain"))
	}
	if len(o.Spec.ClusterNetwork.Pods.CIDRBlocks) == 0 {
		errors = append(errors, field.Invalid(
			field.NewPath("Spec", "ClusterNetwork", "Pods"),
			o.Spec.ClusterNetwork.Pods,
			"invalid cluster configuration: missing Cluster.Spec.ClusterNetwork.Pods"))
	}
	if len(o.Spec.ClusterNetwork.Services.CIDRBlocks) == 0 {
		errors = append(errors, field.Invalid(
			field.NewPath("Spec", "ClusterNetwork", "Services"),
			o.Spec.ClusterNetwork.Services,
			"invalid cluster configuration: missing Cluster.Spec.ClusterNetwork.Services"))
	}
	return errors
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
