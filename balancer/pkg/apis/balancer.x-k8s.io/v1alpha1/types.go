/*
Copyright 2023 The Kubernetes Authors.

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

// Package v1alpha1 contains definitions of Balancer related objects.
package v1alpha1

import (
	hpa "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BalancerList is a list of Balancer objects.
type BalancerList struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	// items is the list of Balancer objects.
	Items []Balancer `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
// +kubebuilder:subresource:status

// Balancer is an object used to automatically keep the desired number of
// replicas (pods) distributed among the specified set of targets (deployments
// or other objects that expose the Scale subresource).
type Balancer struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the Balancer behavior.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	// +kubebuilder:validation:Required
	Spec BalancerSpec `json:"spec" protobuf:"bytes,2,name=spec"`

	// Current information about the Balancer.
	// +optional
	Status BalancerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// BalancerSpec is the specification of the Balancer behavior.
type BalancerSpec struct {
	// Targets is a list of targets between which Balancer tries to distribute
	// replicas.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=2
	Targets []BalancerTarget `json:"targets" protobuf:"bytes,5,rep,name=targets"`

	// Replicas is the number of pods that should be distributed among the
	// declared targets according to the specified policy.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas" protobuf:"varint,2,name=replicas"`

	// Selector that groups the pods from all targets together (and only those).
	// Ideally it should match the selector used by the Service built on top of the
	// Balancer. All pods selectable by targets' selector must match to this selector,
	// however target's selector don't have to be a superset of this one (although
	// it is recommended).
	// +kubebuilder:validation:Required
	Selector metav1.LabelSelector `json:"selector" protobuf:"bytes,3,rep,name=selector"`

	// Policy defines how the balancer should distribute replicas among targets.
	// +kubebuilder:validation:Required
	Policy BalancerPolicy `json:"policy" protobuf:"bytes,4,rep,name=policy"`
}

// BalancerTarget is the declaration of one of the targets between which the balancer
// tries to distribute replicas.
type BalancerTarget struct {
	// Name of the target.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name" protobuf:"bytes,1,name=name"`

	// ScaleTargetRef is a reference that points to a target resource to balance.
	// The target needs to expose the Scale subresource.
	// +kubebuilder:validation:Required
	ScaleTargetRef hpa.CrossVersionObjectReference `json:"scaleTargetRef" protobuf:"bytes,2,opt,name=scaleTargetRef"`

	// MinReplicas is the minimum number of replicas inside of this target.
	// Balancer will set at least this amount on the target, even if the total
	// desired number of replicas for Balancer is lower. 0 will be used (no min) if not
	// provided.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"varint,3,opt,name=minReplicas"`

	// MaxReplicas is the maximum number of replicas inside of this target.
	// Balancer will set at most this amount on the target, even if the total
	// desired number of replicas for the Balancer is higher. There will be no
	// limit if not provided.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxReplicas *int32 `json:"maxReplicas,omitempty" protobuf:"varint,4,opt,name=maxReplicas"`
}

// BalancerPolicyName is the name of the balancer Policy.
type BalancerPolicyName string

const (
	// PriorityPolicyName is the name used in Balancer Spec for priority policy.
	PriorityPolicyName BalancerPolicyName = "priority"
	// ProportionalPolicyName is the name used in Balancer Spec for proportional policy
	ProportionalPolicyName BalancerPolicyName = "proportional"
)

// BalancerPolicy defines Balancer policy for replica distribution.
type BalancerPolicy struct {
	// PolicyName decides how to balance replicas across the targets.
	// Depending on the name one of the fields Priorities or Proportions must be set.
	// +kubebuilder:validation:Required
	PolicyName BalancerPolicyName `json:"policyName" protobuf:"bytes,1,name=policyName"`

	// Priorities contains detailed specification of how to balance when balancer
	// policy name is set to Priority.
	// +optional
	Priorities *PriorityPolicy `json:"priorities,omitempty" protobuf:"bytes,2,opt,name=priorities"`

	// Proportions contains detailed specification of how to balance when
	// balancer policy name is set to Proportional.
	// +optional
	Proportions *ProportionalPolicy `json:"proportions,omitempty" protobuf:"bytes,3,opt,name=proportions"`

	// Fallback contains specification of how to recognize and what to do if some
	// replicas fail to start in one or more targets. No fallback happens if not-set.
	// +optional
	Fallback *FallbackPolicy `json:"fallback,omitempty" protobuf:"bytes,4,opt,name=fallback"`
}

// PriorityPolicy contains details for Priority-based policy for Balancer.
type PriorityPolicy struct {
	// TargetOrder is the priority-based list of Balancer targets names. The first target
	// on the list gets the replicas until its maxReplicas is reached (or replicas
	// fail to start). Then the replicas go to the second target and so on. MinReplicas
	// is guaranteed to be fulfilled, irrespective of the order, presence on the
	// list, and/or total Balancer's replica count.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=2
	TargetOrder []string `json:"targetOrder" protobuf:"bytes,1,rep,name=targetOrder"`
}

// ProportionalPolicy contains details for Proportion-based policy for Balancer.
type ProportionalPolicy struct {
	// TargetProportions is a map from Balancer targets names to rates. Replicas are
	// distributed so that the max difference between the current replica share
	// and the desired replica share is minimized. Once a target reaches maxReplicas
	// it is removed from the calculations and replicas are distributed with
	// the updated proportions. MinReplicas is guaranteed for a target, irrespective
	// of the total Balancer's replica count, proportions or the presence in the map.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinProperties=2
	TargetProportions map[string]int32 `json:"targetProportions" protobuf:"bytes,1,opt,name=targetProportions"`
}

// FallbackPolicy contains information how to recognize and handle replicas
// that failed to start within the specified time period.
type FallbackPolicy struct {
	// StartupTimeoutSeconds defines how long will the Balancer wait before
	// considering a pending/not-started pod as blocked and starting another
	// replica in some other target. Once the replica is finally started,
	// replicas in other targets may be stopped.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	StartupTimeoutSeconds int32 `json:"startupTimeoutSeconds" protobuf:"variant,1,opt,name=startupTimeoutSeconds"`
}

// BalancerStatus describes the Balancer runtime state.
type BalancerStatus struct {

	// Replicas is an actual number of observed pods matching Balancer selector.
	Replicas int32 `json:"replicas" protobuf:"varint,1,opt,name=replicas"`

	// Selector is a query over pods that should match the replicas count. This is same
	// as the label selector but in the string format to avoid introspection
	// by clients. The string will be in the same format as the query-param syntax.
	// More info about label selectors: http://kubernetes.io/docs/user-guide/labels#label-selectors
	Selector string `json:"selector" protobuf:"bytes,2,opt,name=selector"`

	// Conditions is the set of conditions required for this Balancer to work properly,
	// and indicates whether or not those conditions are met.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,3,rep,name=conditions"`
}

const (
	// BalancerConditionRunning is the name of the condition used to indicate the state of the
	// Balancer.
	BalancerConditionRunning = "Balancing"
)
