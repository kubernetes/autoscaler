// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// WARNING!
// IF YOU MODIFY ANY OF THE TYPES HERE COPY THEM TO ../types.go
// AND RUN `make generate`

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WARNING!
// IF YOU MODIFY ANY OF THE TYPES HERE COPY THEM TO ../types.go
// AND RUN  ./hack/generate-code

// NodeLabelKey is the key for node label on machine object
const (
	NodeLabelKey string = "node"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:shortName="mc"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.currentStatus.phase`,description="Current status of the machine."
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata"
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.metadata.labels.node`,description="Node backing the machine object"
// +kubebuilder:printcolumn:name="ProviderID",type=string,JSONPath=`.spec.providerID`,description="ProviderID of the infra instance backing the machine object",priority=1

// Machine is the representation of a physical or virtual machine.
type Machine struct {
	// ObjectMeta for machine object
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// TypeMeta for machine object
	metav1.TypeMeta `json:",inline"`

	// Spec contains the specification of the machine
	Spec MachineSpec `json:"spec,omitempty"`

	// Status contains fields depicting the status
	Status MachineStatus `json:"status,omitempty"`
}

// MachineSpec is the specification of a Machine.
type MachineSpec struct {

	// Class contains the machineclass attributes of a machine
	// +optional
	Class ClassSpec `json:"class,omitempty"`

	// ProviderID represents the provider's unique ID given to a machine
	// +optional
	ProviderID string `json:"providerID,omitempty"`

	// NodeTemplateSpec describes the data a node should have when created from a template
	// +optional
	NodeTemplateSpec NodeTemplateSpec `json:"nodeTemplate,omitempty"`

	// Configuration for the machine-controller.
	// +optional
	*MachineConfiguration `json:",inline"`
}

// ClassSpec is the class specification of machine
type ClassSpec struct {
	// API group to which it belongs
	APIGroup string `json:"apiGroup,omitempty"`

	// Kind for machine class
	Kind string `json:"kind,omitempty"`

	// Name of machine class
	Name string `json:"name,omitempty"`
}

// NodeTemplateSpec describes the data a node should have when created from a template
type NodeTemplateSpec struct {
	// +kubebuilder:validation:XPreserveUnknownFields
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// NodeSpec describes the attributes that a node is created with.
	// +optional
	Spec corev1.NodeSpec `json:"spec,omitempty"`
}

// MachineStatus holds the most recently observed status of Machine.
type MachineStatus struct {
	// Conditions of this machine, same as node
	Conditions []corev1.NodeCondition `json:"conditions,omitempty"`

	// Last operation refers to the status of the last operation performed
	LastOperation LastOperation `json:"lastOperation,omitempty"`

	// Current status of the machine object
	CurrentStatus CurrentStatus `json:"currentStatus,omitempty"`

	// LastKnownState can store details of the last known state of the VM by the plugins.
	// It can be used by future operation calls to determine current infrastucture state
	// +optional
	LastKnownState string `json:"lastKnownState,omitempty"`
}

// LastOperation suggests the last operation performed on the object
type LastOperation struct {
	// Description of the current operation
	Description string `json:"description,omitempty"`

	// ErrorCode of the current operation if any
	// +optional
	ErrorCode string `json:"errorCode,omitempty"`

	// Last update time of current operation
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// State of operation
	State MachineState `json:"state,omitempty"`

	// Type of operation
	Type MachineOperationType `json:"type,omitempty"`
}

// MachinePhase is a label for the condition of a machine at the current time.
type MachinePhase string

// These are the valid phases of machines.
const (
	// MachinePending means that the machine is being created
	MachinePending MachinePhase = "Pending"

	// MachineAvailable means that machine is present on provider but hasn't joined cluster yet
	MachineAvailable MachinePhase = "Available"

	// MachineRunning means node is ready and running successfully
	MachineRunning MachinePhase = "Running"

	// MachineTerminating means node is terminating
	MachineTerminating MachinePhase = "Terminating"

	// MachineUnknown indicates that the node is not ready at the movement
	MachineUnknown MachinePhase = "Unknown"

	// MachineFailed means operation timed out
	MachineFailed MachinePhase = "Failed"

	// MachineCrashLoopBackOff means machine creation is failing. It means that machine object is present but there is no corresponding VM.
	MachineCrashLoopBackOff MachinePhase = "CrashLoopBackOff"
)

// MachineState is a current state of the operation.
type MachineState string

// These are the valid states of operations performed on the machine.
const (
	// MachineStateProcessing means operation is not yet complete
	MachineStateProcessing MachineState = "Processing"

	// MachineStateFailed means operation failed 
	MachineStateFailed MachineState = "Failed"

	// MachineStateSuccessful means operation completed successfully 
	MachineStateSuccessful MachineState = "Successful"
)

// MachineOperationType is a label for the operation performed on a machine object.
type MachineOperationType string

// These are the valid statuses of machines.
const (
	// MachineOperationCreate indicates that the operation was a create
	MachineOperationCreate MachineOperationType = "Create"

	// MachineOperationUpdate indicates that the operation was an update
	MachineOperationUpdate MachineOperationType = "Update"

	// MachineOperationHealthCheck indicates that the operation was a health check of node object
	MachineOperationHealthCheck MachineOperationType = "HealthCheck"

	// MachineOperationDelete indicates that the operation was a delete
	MachineOperationDelete MachineOperationType = "Delete"
)

// The below types are used by kube_client and api_server.

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition;
// "ConditionFalse" means a resource is not in the condition; "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// CurrentStatus contains information about the current status of Machine.
type CurrentStatus struct {
	Phase MachinePhase `json:"phase,omitempty"`

	TimeoutActive bool `json:"timeoutActive,omitempty"`

	// Last update time of current status
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// MachineList is a collection of Machines.
type MachineList struct {
	// ObjectMeta for MachineList object
	metav1.TypeMeta `json:",inline"`

	// TypeMeta for MachineList object
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items contains the list of machines
	Items []Machine `json:"items"`
}
