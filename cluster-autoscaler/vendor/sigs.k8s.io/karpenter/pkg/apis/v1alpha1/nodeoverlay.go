/*
Copyright The Kubernetes Authors.

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
	"sort"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A node selector requirement is a selector that contains values, a key, an operator that relates the key and values
// to have at least that many values.
type NodeSelectorRequirement struct {
	//nolint:kubeapilinter
	// The label key that the selector applies to.
	// +required
	Key string `json:"key"`

	// NOTE below: The code generator works strangely, and will union these
	// enum values into the ones already defined on v1.NodeSelectorOperator

	//nolint:kubeapilinter
	// Represents a key's relationship to a set of values.
	// Valid operators are In, NotIn, Exists, DoesNotExist. Gt, Lt, Gte, and Lte.
	// +kubebuilder:validation:Enum:=Gte;Lte
	// +required
	Operator corev1.NodeSelectorOperator `json:"operator,omitempty"`
	//nolint:kubeapilinter
	// An array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty. If the operator is Gt, Lt, Gte, or Lte, the values
	// array must have a single element, which will be interpreted as an integer.
	// +optional
	// +listType=atomic
	Values []string `json:"values,omitempty"`
}

// AsNodeSelectorRequirement converts to corev1.NodeSelectorRequirement
func (r NodeSelectorRequirement) AsNodeSelectorRequirement() corev1.NodeSelectorRequirement {
	return corev1.NodeSelectorRequirement{Key: r.Key, Operator: r.Operator, Values: r.Values}
}

type NodeOverlaySpec struct {
	// requirements constrain when this NodeOverlay is applied during scheduling simulations.
	// These requirements can match:
	// - Well-known labels (e.g., node.kubernetes.io/instance-type, karpenter.sh/nodepool)
	// - Custom labels from NodePool's spec.template.labels
	// +kubebuilder:validation:XValidation:message="requirements with operator 'NotIn' must have a value defined",rule="self.all(x, x.operator == 'NotIn' ? x.values.size() != 0 : true)"
	// +kubebuilder:validation:XValidation:message="requirements with operator 'In' must have a value defined",rule="self.all(x, x.operator == 'In' ? x.values.size() != 0 : true)"
	// +kubebuilder:validation:XValidation:message="requirements operator 'Gt', 'Lt', 'Gte' or 'Lte' must have a single positive integer value",rule="self.all(x, (x.operator == 'Gt' || x.operator == 'Lt' || x.operator == 'Gte' || x.operator == 'Lte') ? (x.values.size() == 1 && int(x.values[0]) >= 0) : true)"
	// +kubebuilder:validation:MaxItems:=100
	// +required
	// +listType=atomic
	Requirements []NodeSelectorRequirement `json:"requirements,omitempty"`
	//nolint:kubeapilinter
	// PriceAdjustment specifies the price change for matching instance types. Accepts either:
	// - A fixed price modifier (e.g., -0.5, 1.2)
	// - A percentage modifier (e.g., +10% for increase, -15% for decrease)
	// +kubebuilder:validation:Pattern=`^(([+-]{1}(\d*\.?\d+))|(\+{1}\d*\.?\d+%)|(^(-\d{1,2}(\.\d+)?%)$)|(-100%))$`
	// +optional
	PriceAdjustment *string `json:"priceAdjustment,omitempty"`
	//nolint:kubeapilinter
	// Price specifies amount for an instance types that match the specified labels. Users can override prices using a signed float representing the price override
	// +kubebuilder:validation:Pattern=`^\d+(\.\d+)?$`
	// +optional
	Price *string `json:"price,omitempty"`
	//nolint:kubeapilinter
	// Capacity adds extended resources only, and does not replace any existing resources.
	// These extended resources are appended to the node's existing resource list.
	// Note: This field does not modify or override standard resources like cpu, memory, ephemeral-storage, or pods.
	// +kubebuilder:validation:XValidation:message="invalid resource restricted",rule="self.all(x, !(x in ['cpu', 'memory', 'ephemeral-storage', 'pods']))"
	// +optional
	Capacity corev1.ResourceList `json:"capacity,omitempty"`
	//nolint:kubeapilinter
	// Weight defines the priority of this NodeOverlay when overriding node attributes.
	// NodeOverlays with higher numerical weights take precedence over those with lower weights.
	// If no weight is specified, the NodeOverlay is treated as having a weight of 0.
	// When multiple NodeOverlays have identical weights, they are merged in alphabetical order.
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=10000
	// +optional
	Weight *int32 `json:"weight,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:path=nodeoverlays,scope=Cluster,categories=karpenter,shortName=overlays
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="Weight",type="integer",JSONPath=".spec.weight",priority=1,description=""
// +kubebuilder:subresource:status
type NodeOverlay struct {
	metav1.TypeMeta `json:",inline"`
	//nolint:kubeapilinter
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//nolint:kubeapilinter
	// +kubebuilder:validation:XValidation:message="cannot set both 'price' and 'priceAdjustment'",rule="!has(self.price) || !has(self.priceAdjustment)"
	Spec   NodeOverlaySpec   `json:"spec"`
	Status NodeOverlayStatus `json:"status,omitempty"` //nolint:kubeapilinter
}

// +kubebuilder:object:root=true
type NodeOverlayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeOverlay `json:"items"`
}

// OrderByWeight orders the NodeOverlays in the provided slice by their priority weight in-place. This priority evaluates
// the following things in precedence order:
//  1. NodeOverlays that have a larger weight are ordered first
//  2. If two NodeOverlays have the same weight, then the NodeOverlay with the name later in the alphabet will come first
func (nol *NodeOverlayList) OrderByWeight() {
	sort.Slice(nol.Items, func(a, b int) bool {
		weightA := lo.FromPtr(nol.Items[a].Spec.Weight)
		weightB := lo.FromPtr(nol.Items[b].Spec.Weight)
		if weightA == weightB {
			// Order Node Overlay by name for a consistent ordering when sorting equal weight
			return nol.Items[a].Name > nol.Items[b].Name
		}
		return weightA > weightB
	})
}
