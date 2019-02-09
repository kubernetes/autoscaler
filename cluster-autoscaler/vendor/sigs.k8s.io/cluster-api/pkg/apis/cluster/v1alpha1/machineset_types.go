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
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

/// [MachineSet]
// MachineSet ensures that a specified number of machines replicas are running at any given time.
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.labelSelector
type MachineSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineSetSpec   `json:"spec,omitempty"`
	Status MachineSetStatus `json:"status,omitempty"`
}

/// [MachineSet]

/// [MachineSetSpec]
// MachineSetSpec defines the desired state of MachineSet
type MachineSetSpec struct {
	// Replicas is the number of desired replicas.
	// This is a pointer to distinguish between explicit zero and unspecified.
	// Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// MinReadySeconds is the minimum number of seconds for which a newly created machine should be ready.
	// Defaults to 0 (machine will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`

	// Selector is a label query over machines that should match the replica count.
	// Label keys and values that must match in order to be controlled by this MachineSet.
	// It must match the machine template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector metav1.LabelSelector `json:"selector"`

	// Template is the object that describes the machine that will be created if
	// insufficient replicas are detected.
	// +optional
	Template MachineTemplateSpec `json:"template,omitempty"`
}

/// [MachineSetSpec] // doxygen marker

/// [MachineTemplateSpec] // doxygen marker
// MachineTemplateSpec describes the data needed to create a Machine from a template
type MachineTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the machine.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec MachineSpec `json:"spec,omitempty"`
}

/// [MachineTemplateSpec]

/// [MachineSetStatus]
// MachineSetStatus defines the observed state of MachineSet
type MachineSetStatus struct {
	// Replicas is the most recently observed number of replicas.
	Replicas int32 `json:"replicas"`

	// The number of replicas that have labels matching the labels of the machine template of the MachineSet.
	// +optional
	FullyLabeledReplicas int32 `json:"fullyLabeledReplicas,omitempty"`

	// The number of ready replicas for this MachineSet. A machine is considered ready when the node has been created and is "Ready".
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// The number of available replicas (ready for at least minReadySeconds) for this MachineSet.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed MachineSet.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// In the event that there is a terminal problem reconciling the
	// replicas, both ErrorReason and ErrorMessage will be set. ErrorReason
	// will be populated with a succinct value suitable for machine
	// interpretation, while ErrorMessage will contain a more verbose
	// string suitable for logging and human consumption.
	//
	// These fields should not be set for transitive errors that a
	// controller faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the MachineTemplate's spec or the configuration of
	// the machine controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the machine controller, or the
	// responsible machine controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the MachineSet object and/or logged in the
	// controller's output.
	// +optional
	ErrorReason *common.MachineSetStatusError `json:"errorReason,omitempty"`
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

/// [MachineSetStatus]

func (m *MachineSet) Validate() field.ErrorList {
	errors := field.ErrorList{}

	// validate spec.selector and spec.template.labels
	fldPath := field.NewPath("spec")
	errors = append(errors, metav1validation.ValidateLabelSelector(&m.Spec.Selector, fldPath.Child("selector"))...)
	if len(m.Spec.Selector.MatchLabels)+len(m.Spec.Selector.MatchExpressions) == 0 {
		errors = append(errors, field.Invalid(fldPath.Child("selector"), m.Spec.Selector, "empty selector is not valid for MachineSet."))
	}
	selector, err := metav1.LabelSelectorAsSelector(&m.Spec.Selector)
	if err != nil {
		errors = append(errors, field.Invalid(fldPath.Child("selector"), m.Spec.Selector, "invalid label selector."))
	} else {
		labels := labels.Set(m.Spec.Template.Labels)
		if !selector.Matches(labels) {
			errors = append(errors, field.Invalid(fldPath.Child("template", "metadata", "labels"), m.Spec.Template.Labels, "`selector` does not match template `labels`"))
		}
	}

	return errors
}

// DefaultingFunction sets default MachineSet field values
func (m *MachineSet) Default() {
	log.Printf("Defaulting fields for MachineSet %s\n", m.Name)

	if m.Spec.Replicas == nil {
		m.Spec.Replicas = new(int32)
		*m.Spec.Replicas = 1
	}

	if len(m.Namespace) == 0 {
		m.Namespace = metav1.NamespaceDefault
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineSetList contains a list of MachineSet
type MachineSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MachineSet{}, &MachineSetList{})
}
