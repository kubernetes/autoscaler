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

// CapacityQuotaSpec defines the desired state of CapacityQuota
type CapacityQuotaSpec struct {
	// Selector is a label selector selecting the nodes to which the quota applies.
	// Empty or nil selector matches all nodes.
	// +optional
	Selector *LabelSelector `json:"selector,omitempty"`

	// Limits define quota limits.
	// +required
	Limits CapacityQuotaLimits `json:"limits"`
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

// CapacityQuotaStatus defines the observed state of CapacityQuota.
type CapacityQuotaStatus struct {
	// Used shows the current usage of the quota.
	// +optional
	Used *CapacityQuotaUsage `json:"used,omitempty"`

	// Conditions provide a standard mechanism for reporting the quota's state.
	// +listType=map
	// +listMapKey=type
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// CapacityQuotaUsage shows the current usage of the quota.
type CapacityQuotaUsage struct {
	// Resources shows the current usage of the resources defined in the quota limits.
	Resources ResourceList `json:"resources"`
}

// LabelSelector mirrors k8s.io/apimachinery/pkg/apis/meta/v1.LabelSelector with CRD-level validation rules.
type LabelSelector struct {
	// matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
	// map is equivalent to an element of matchExpressions, whose key field is "key", the
	// operator is "In", and the values array contains only "value". The requirements are ANDed.
	// +kubebuilder:validation:MaxProperties=64
	// +kubebuilder:validation:XValidation:rule="self.all(x, size(x) < 317 && !format.qualifiedName().validate(x).hasValue())",message="key must be a valid qualified name"
	// +optional
	MatchLabels map[string]labelValue `json:"matchLabels,omitempty"`
	// matchExpressions is a list of label selector requirements. The requirements are ANDed.
	// +kubebuilder:validation:MaxItems=16
	// +optional
	// +listType=atomic
	MatchExpressions []LabelSelectorRequirement `json:"matchExpressions,omitempty"`
}

// +kubebuilder:validation:MaxLength=63
// +kubebuilder:validation:XValidation:rule="!format.labelValue().validate(self).hasValue()",message="label value must be a valid label value"
type labelValue = string

// LabelSelectorRequirement is a selector that contains values, a key, and an operator that
// relates the key and values.
// +kubebuilder:validation:XValidation:rule="self.operator in ['In', 'NotIn'] ? (has(self.values) && size(self.values) > 0) : true",message="values set must be non-empty when operator is In or NotIn"
// +kubebuilder:validation:XValidation:rule="self.operator in ['Exists', 'DoesNotExist'] ? (!has(self.values) || size(self.values) == 0) : true",message="values set must be empty when operator is Exists or DoesNotExist"
type LabelSelectorRequirement struct {
	// key is the label key that the selector applies to.
	// +kubebuilder:validation:MaxLength=316
	// +kubebuilder:validation:XValidation:rule="!format.qualifiedName().validate(self).hasValue()",message="key must be a valid qualified name"
	Key string `json:"key"`
	// operator represents a key's relationship to a set of values.
	// Valid operators are In, NotIn, Exists and DoesNotExist.
	// +kubebuilder:validation:Enum=In;NotIn;Exists;DoesNotExist
	Operator LabelSelectorOperator `json:"operator"`
	// values is an array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty. This array is replaced during a strategic
	// merge patch.
	// +kubebuilder:validation:MaxItems=64
	// +kubebuilder:validation:items:MaxLength=63
	// +kubebuilder:validation:XValidation:rule="self.all(v, !format.labelValue().validate(v).hasValue())",message="values must be valid label values"
	// +optional
	// +listType=atomic
	Values []string `json:"values,omitempty"`
}

// LabelSelectorOperator is the set of operators that can be used in a selector requirement.
type LabelSelectorOperator string

const (
	// LabelSelectorOpIn checks if a label value is in a list of values.
	LabelSelectorOpIn LabelSelectorOperator = "In"
	// LabelSelectorOpNotIn checks if a label value is not in a list of values.
	LabelSelectorOpNotIn LabelSelectorOperator = "NotIn"
	// LabelSelectorOpExists checks if a label exists.
	LabelSelectorOpExists LabelSelectorOperator = "Exists"
	// LabelSelectorOpDoesNotExist checks if a label does not exist.
	LabelSelectorOpDoesNotExist LabelSelectorOperator = "DoesNotExist"
)

// +kubebuilder:object:root=true

// CapacityQuotaList contains a list of CapacityQuota
type CapacityQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CapacityQuota `json:"items"`
}

// AsMetaLabelSelector converts a LabelSelector to a metav1.LabelSelector.
func (s *LabelSelector) AsMetaLabelSelector() *metav1.LabelSelector {
	if s == nil {
		return nil
	}
	matchLabels := make(map[string]string, len(s.MatchLabels))
	for k, v := range s.MatchLabels {
		matchLabels[k] = v
	}
	matchExpressions := make([]metav1.LabelSelectorRequirement, len(s.MatchExpressions))
	for i, expr := range s.MatchExpressions {
		values := make([]string, len(expr.Values))
		copy(values, expr.Values)
		matchExpressions[i] = metav1.LabelSelectorRequirement{
			Key:      expr.Key,
			Operator: metav1.LabelSelectorOperator(expr.Operator),
			Values:   values,
		}
	}

	return &metav1.LabelSelector{
		MatchLabels:      s.MatchLabels,
		MatchExpressions: matchExpressions,
	}
}

func init() {
	SchemeBuilder.Register(&CapacityQuota{}, &CapacityQuotaList{})
}
