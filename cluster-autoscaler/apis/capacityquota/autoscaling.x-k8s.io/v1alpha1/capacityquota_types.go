/*
Copyright 2025 The Kubernetes Authors.

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

// +groupName=autoscaling.x-k8s.io
// +k8s:protobuf-gen=package

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceName is the name identifying a resource mirroring k8s.io/api/core/v1.ResourceName.
type ResourceName string

const (
	// ResourceCPU - CPU, in cores. (500m = .5 cores)
	ResourceCPU ResourceName = "cpu"
	// ResourceMemory - memory in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceMemory ResourceName = "memory"
	// ResourceNodes - number of nodes, in units.
	ResourceNodes ResourceName = "nodes"
)

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[ResourceName]resource.Quantity

// CapacityQuotaSpec defines the desired state of CapacityQuota
type CapacityQuotaSpec struct {
	// Selector is a label selector selecting the nodes to which the quota applies.
	// Empty or nil selector matches all nodes.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Limits define quota limits.
	// +required
	Limits CapacityQuotaLimits `json:"limits"`
}

// CapacityQuotaStatus defines the observed state of CapacityQuota.
type CapacityQuotaStatus struct {
	// TODO: status should report resources currently in use
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=cq
// +genclient

// CapacityQuota limits the amount of resources that can be provisioned in the cluster
// by the node autoscaler. Resources used are calculated by summing up resources
// reported in the status.capacity field of each node passing the configured
// label selector. When making a provisioning decision, node autoscaler will
// take all CapacityQuota objects that match the labels of the upcoming node.
// If provisioning that node would exceed any of the matching quotas, node
// autoscaler will not provision it. Quotas are best-effort, and it is possible
// that in rare circumstances node autoscaler will exceed them, for example
// due to stale caches.
// More info: https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/granular-resource-limits.md
type CapacityQuota struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of CapacityQuota
	// +required
	Spec CapacityQuotaSpec `json:"spec"`

	// status defines the observed state of CapacityQuota
	// +optional
	Status CapacityQuotaStatus `json:"status,omitempty,omitzero"`
}

// CapacityQuotaLimits define quota limits.
type CapacityQuotaLimits struct {
	// Resources define resource limits of this quota.
	//
	// Currently supported built-in resources: cpu, memory. Additionally,
	// nodes key can be used to limit the number of existing nodes.
	// Node autoscaler implementations and cloud providers can support custom
	// resources, such as GPU.
	// +required
	Resources ResourceList `json:"resources"`
}

// +kubebuilder:object:root=true

// CapacityQuotaList contains a list of CapacityQuota
type CapacityQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CapacityQuota `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CapacityQuota{}, &CapacityQuotaList{})
}
