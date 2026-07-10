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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NodeClaimSpec describes the desired state of the NodeClaim
type NodeClaimSpec struct {
	// taints will be applied to the NodeClaim's node.
	// +optional
	// +listType=atomic
	Taints []v1.Taint `json:"taints,omitempty"`
	// startupTaints are taints that are applied to nodes upon startup which are expected to be removed automatically
	// within a short period of time, typically by a DaemonSet that tolerates the taint. These are commonly used by
	// daemonsets to allow initialization and enforce startup ordering.  StartupTaints are ignored for provisioning
	// purposes in that pods are not required to tolerate a StartupTaint in order to have nodes provisioned for them.
	// +optional
	// +listType=atomic
	StartupTaints []v1.Taint `json:"startupTaints,omitempty"`
	// Requirements are layered with GetLabels and applied to every node.
	// +kubebuilder:validation:XValidation:message="requirements with operator 'In' must have a value defined",rule="self.all(x, x.operator == 'In' ? x.values.size() != 0 : true)"
	// +kubebuilder:validation:XValidation:message="requirements operator 'Gt', 'Lt', 'Gte', or 'Lte' must have a single positive integer value",rule="self.all(x, (x.operator == 'Gt' || x.operator == 'Lt' || x.operator == 'Gte' || x.operator == 'Lte') ? (x.values.size() == 1 && int(x.values[0]) >= 0) : true)"
	// +kubebuilder:validation:XValidation:message="requirements with 'minValues' must have at least that many values specified in the 'values' field",rule="self.all(x, (x.operator == 'In' && has(x.minValues)) ? x.values.size() >= x.minValues : true)"
	// +kubebuilder:validation:MaxItems:=100
	// +required
	// +listType=atomic
	//nolint:kubeapilinter
	Requirements []NodeSelectorRequirementWithMinValues `json:"requirements" hash:"ignore"`
	//nolint:kubeapilinter
	// Resources models the resource requirements for the NodeClaim to launch
	// +optional
	Resources ResourceRequirements `json:"resources,omitempty" hash:"ignore"`
	//nolint:kubeapilinter
	// NodeClassRef is a reference to an object that defines provider specific configuration
	// +required
	NodeClassRef *NodeClassReference `json:"nodeClassRef"`
	//nolint:kubeapilinter
	// TerminationGracePeriod is the maximum duration the controller will wait before forcefully deleting the pods on a node, measured from when deletion is first initiated.
	//
	// Warning: this feature takes precedence over a Pod's terminationGracePeriodSeconds value, and bypasses any blocked PDBs or the karpenter.sh/do-not-disrupt annotation.
	//
	// This field is intended to be used by cluster administrators to enforce that nodes can be cycled within a given time period.
	// When set, drifted nodes will begin draining even if there are pods blocking eviction. Draining will respect PDBs and the do-not-disrupt annotation until the TGP is reached.
	//
	// Karpenter will preemptively delete pods so their terminationGracePeriodSeconds align with the node's terminationGracePeriod.
	// If a pod would be terminated without being granted its full terminationGracePeriodSeconds prior to the node timeout,
	// that pod will be deleted at T = node timeout - pod terminationGracePeriodSeconds.
	//
	// The feature can also be used to allow maximum time limits for long-running jobs which can delay node termination with preStop hooks.
	// If left undefined, the controller will wait indefinitely for pods to be drained.
	// +kubebuilder:validation:Pattern=`^([0-9]+(s|m|h))+$`
	// +kubebuilder:validation:Type="string"
	// +optional
	TerminationGracePeriod *metav1.Duration `json:"terminationGracePeriod,omitempty"`
	// ExpireAfter is the duration the controller will wait
	// before terminating a node, measured from when the node is created. This
	// is useful to implement features like eventually consistent node upgrade,
	// memory leak protection, and disruption testing.
	// +kubebuilder:default:="720h"
	// +kubebuilder:validation:Pattern=`^(([0-9]+(s|m|h))+|Never)$`
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Schemaless
	// +optional
	ExpireAfter NillableDuration `json:"expireAfter,omitempty"`
}

const (
	NodeSelectorOpGte v1.NodeSelectorOperator = "Gte"
	NodeSelectorOpLte v1.NodeSelectorOperator = "Lte"
)

// A node selector requirement is a selector that contains values, a key, an operator that relates the key and values
// and minValues that represent the requirement to have at least that many values.
type NodeSelectorRequirementWithMinValues struct {
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
	Operator v1.NodeSelectorOperator `json:"operator,omitempty"`
	//nolint:kubeapilinter
	// An array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty. If the operator is Gt, Lt, Gte, or Lte, the values
	// array must have a single element, which will be interpreted as an integer.
	// This array is replaced during a strategic merge patch.
	// +optional
	// +listType=atomic
	Values []string `json:"values,omitempty"`
	//nolint:kubeapilinter
	// This field is ALPHA and can be dropped or replaced at any time
	// MinValues is the minimum number of unique values required to define the flexibility of the specific requirement.
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=50
	// +optional
	MinValues *int `json:"minValues,omitempty"`
}

// ResourceRequirements models the required resources for the NodeClaim to launch
// Ths will eventually be transformed into v1.ResourceRequirements when we support resources.limits
type ResourceRequirements struct {
	//nolint:kubeapilinter
	// Requests describes the minimum required resources for the NodeClaim to launch
	// +optional
	Requests v1.ResourceList `json:"requests,omitempty"`
}

type NodeClassReference struct {
	//nolint:kubeapilinter
	// Kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
	// +kubebuilder:validation:XValidation:rule="self != ''",message="kind may not be empty"
	// +required
	Kind string `json:"kind"`
	//nolint:kubeapilinter
	// Name of the referent; More info: http://kubernetes.io/docs/user-guide/identifiers#names
	// +kubebuilder:validation:XValidation:rule="self != ''",message="name may not be empty"
	// +required
	Name string `json:"name"`
	//nolint:kubeapilinter
	// API version of the referent
	// +kubebuilder:validation:XValidation:rule="self != ''",message="group may not be empty"
	// +kubebuilder:validation:Pattern=`^[^/]*$`
	// +required
	Group string `json:"group"`
}

func (ncr *NodeClassReference) GroupKind() schema.GroupKind {
	return schema.GroupKind{
		Group: ncr.Group,
		Kind:  ncr.Kind,
	}
}

// +kubebuilder:object:generate=false
type Provider = runtime.RawExtension

// NodeClaim is the Schema for the NodeClaims API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=nodeclaims,scope=Cluster,categories=karpenter
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".metadata.labels.node\\.kubernetes\\.io/instance-type",description=""
// +kubebuilder:printcolumn:name="Capacity",type="string",JSONPath=".metadata.labels.karpenter\\.sh/capacity-type",description=""
// +kubebuilder:printcolumn:name="Zone",type="string",JSONPath=".metadata.labels.topology\\.kubernetes\\.io/zone",description=""
// +kubebuilder:printcolumn:name="Node",type="string",JSONPath=".status.nodeName",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="ImageID",type="string",JSONPath=".status.imageID",priority=1,description=""
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.providerID",priority=1,description=""
// +kubebuilder:printcolumn:name="NodePool",type="string",JSONPath=".metadata.labels.karpenter\\.sh/nodepool",priority=1,description=""
// +kubebuilder:printcolumn:name="NodeClass",type="string",JSONPath=".spec.nodeClassRef.name",priority=1,description=""
// +kubebuilder:printcolumn:name="Drifted",type="string",JSONPath=".status.conditions[?(@.type==\"Drifted\")].status",priority=1,description=""
type NodeClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"` //nolint:kubeapilinter

	//nolint:kubeapilinter
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec is immutable"
	// +required
	Spec   NodeClaimSpec   `json:"spec"`
	Status NodeClaimStatus `json:"status,omitempty"` //nolint:kubeapilinter
}

// NodeClaimList contains a list of NodeClaims
// +kubebuilder:object:root=true
type NodeClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeClaim `json:"items"`
}
