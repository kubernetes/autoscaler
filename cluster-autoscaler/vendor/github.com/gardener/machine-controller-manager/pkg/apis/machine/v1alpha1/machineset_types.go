// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// WARNING!
// IF YOU MODIFY ANY OF THE TYPES HERE COPY THEM TO ../types.go
// AND RUN `make generate`

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:method=UpdateScale,verb=update,subresource=scale,input=k8s.io/api/autoscaling/v1.Scale,result=k8s.io/api/autoscaling/v1.Scale
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.replicas`,description="Number of desired replicas."
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.replicas`,description="Number of actual replicas."
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`,description="Number of ready replicas for this machine set."
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata"

// MachineSet TODO
type MachineSet struct {
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	metav1.TypeMeta `json:",inline"`

	// +optional
	Spec MachineSetSpec `json:"spec,omitempty"`

	// +optional
	Status MachineSetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// MachineSetList is a collection of MachineSet.
type MachineSetList struct {
	// +optional
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// +optional
	Items []MachineSet `json:"items"`
}

// MachineSetSpec is the specification of a MachineSet.
type MachineSetSpec struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// +optional
	MachineClass ClassSpec `json:"machineClass,omitempty"`

	// +optional
	Template MachineTemplateSpec `json:"template,omitempty"`

	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`
}

// MachineSetConditionType is the condition on machineset object
type MachineSetConditionType string

// These are valid conditions of a machine set.
const (
	// MachineSetReplicaFailure is added in a machine set when one of its machines fails to be created
	// due to insufficient quota, limit ranges, machine security policy, node selectors, etc. or deleted
	// due to kubelet being down or finalizers are failing.
	MachineSetReplicaFailure MachineSetConditionType = "ReplicaFailure"
	// MachineSetFrozen is set when the machineset has exceeded its replica threshold at the safety controller
	MachineSetFrozen MachineSetConditionType = "Frozen"
)

// MachineSetCondition describes the state of a machine set at a certain point.
type MachineSetCondition struct {
	// Type of machine set condition.
	Type MachineSetConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status"`

	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// MachineSetStatus holds the most recently observed status of MachineSet.
type MachineSetStatus struct {
	// Replicas is the number of actual replicas.
	Replicas int32 `json:"replicas,omitempty"`

	// The number of pods that have labels matching the labels of the pod template of the replicaset.
	// +optional
	FullyLabeledReplicas int32 `json:"fullyLabeledReplicas,omitempty"`

	// The number of ready replicas for this replica set.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// The number of available replicas (ready for at least minReadySeconds) for this replica set.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// ObservedGeneration is the most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents the latest available observations of a replica set's current state.
	// +optional
	Conditions []MachineSetCondition `json:"machineSetCondition,omitempty"`

	// LastOperation performed
	LastOperation LastOperation `json:"lastOperation,omitempty"`

	// FailedMachines has summary of machines on which lastOperation Failed
	// +optional
	FailedMachines *[]MachineSummary `json:"failedMachines,omitempty"`
}
