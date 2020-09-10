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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +genclient:method=UpdateScale,verb=update,subresource=scale,input=k8s.io/api/autoscaling/v1.Scale,result=k8s.io/api/autoscaling/v1.Scale
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`,description="Total number of ready machines targeted by this machine deployment."
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.replicas`,description="Number of desired machines."
// +kubebuilder:printcolumn:name="Up-to-date",type=integer,JSONPath=`.status.updatedReplicas`,description="Total number of non-terminated machines targeted by this machine deployment that have the desired template spec."
// +kubebuilder:printcolumn:name="Available",type=integer,JSONPath=`.status.availableReplicas`,description="Total number of available machines (ready for at least minReadySeconds) targeted by this machine deployment."
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata"
// Deployment enables declarative updates for machines and MachineSets.
type MachineDeployment struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the MachineDeployment.
	// +optional
	Spec MachineDeploymentSpec `json:"spec,omitempty"`

	// Most recently observed status of the MachineDeployment.
	// +optional
	Status MachineDeploymentStatus `json:"status,omitempty"`
}

// MachineDeploymentSpec is the specification of the desired behavior of the MachineDeployment.
type MachineDeploymentSpec struct {
	// Number of desired machines. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 0.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Label selector for machines. Existing MachineSets whose machines are
	// selected by this will be the ones affected by this MachineDeployment.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Template describes the machines that will be created.
	Template MachineTemplateSpec `json:"template"`

	// The MachineDeployment strategy to use to replace existing machines with new ones.
	// +optional
	// +patchStrategy=retainKeys
	Strategy MachineDeploymentStrategy `json:"strategy,omitempty" patchStrategy:"retainKeys"`

	// Minimum number of seconds for which a newly created machine should be ready
	// without any of its container crashing, for it to be considered available.
	// Defaults to 0 (machine will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`

	// The number of old MachineSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`

	// Indicates that the MachineDeployment is paused and will not be processed by the
	// MachineDeployment controller.
	// +optional
	Paused bool `json:"paused,omitempty"`

	// DEPRECATED.
	// The config this MachineDeployment is rolling back to. Will be cleared after rollback is done.
	// +optional
	RollbackTo *RollbackConfig `json:"rollbackTo,omitempty"`

	// The maximum time in seconds for a MachineDeployment to make progress before it
	// is considered to be failed. The MachineDeployment controller will continue to
	// process failed MachineDeployments and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the MachineDeployment status. Note that progress will
	// not be estimated during the time a MachineDeployment is paused. This is not set
	// by default.
	// +optional
	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty"`
}

const (
	// DefaultDeploymentUniqueLabelKey is the default key of the selector that is added
	// to existing MCs (and label key that is added to its machines) to prevent the existing MCs
	// to select new machines (and old machines being select by new MC).
	DefaultMachineDeploymentUniqueLabelKey string = "machine-template-hash"
)

// MachineDeploymentStrategy describes how to replace existing machines with new ones.
type MachineDeploymentStrategy struct {
	// Type of MachineDeployment. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.
	// +optional
	Type MachineDeploymentStrategyType `json:"type,omitempty"`

	// Rolling update config params. Present only if MachineDeploymentStrategyType =
	// RollingUpdate.
	//---
	// TODO: Update this to follow our convention for oneOf, whatever we decide it
	// to be.
	// +optional
	RollingUpdate *RollingUpdateMachineDeployment `json:"rollingUpdate,omitempty"`
}

type MachineDeploymentStrategyType string

const (
	// Kill all existing machines before creating new ones.
	RecreateMachineDeploymentStrategyType MachineDeploymentStrategyType = "Recreate"

	// Replace the old MCs by new one using rolling update i.e gradually scale down the old MCs and scale up the new one.
	RollingUpdateMachineDeploymentStrategyType MachineDeploymentStrategyType = "RollingUpdate"
)

// Spec to control the desired behavior of rolling update.
type RollingUpdateMachineDeployment struct {
	// The maximum number of machines that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired machines (ex: 10%).
	// Absolute number is calculated from percentage by rounding down.
	// This can not be 0 if MaxSurge is 0.
	// By default, a fixed value of 1 is used.
	// Example: when this is set to 30%, the old MC can be scaled down to 70% of desired machines
	// immediately when the rolling update starts. Once new machines are ready, old MC
	// can be scaled down further, followed by scaling up the new MC, ensuring
	// that the total number of machines available at all times during the update is at
	// least 70% of desired machines.
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`

	// The maximum number of machines that can be scheduled above the desired number of
	// machines.
	// Value can be an absolute number (ex: 5) or a percentage of desired machines (ex: 10%).
	// This can not be 0 if MaxUnavailable is 0.
	// Absolute number is calculated from percentage by rounding up.
	// By default, a value of 1 is used.
	// Example: when this is set to 30%, the new MC can be scaled up immediately when
	// the rolling update starts, such that the total number of old and new machines do not exceed
	// 130% of desired machines. Once old machines have been killed,
	// new MC can be scaled up further, ensuring that total number of machines running
	// at any time during the update is atmost 130% of desired machines.
	// +optional
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty"`
}

// MachineDeploymentStatus is the most recently observed status of the MachineDeployment.
type MachineDeploymentStatus struct {
	// The generation observed by the MachineDeployment controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Total number of non-terminated machines targeted by this MachineDeployment (their labels match the selector).
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Total number of non-terminated machines targeted by this MachineDeployment that have the desired template spec.
	// +optional
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`

	// Total number of ready machines targeted by this MachineDeployment.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Total number of available machines (ready for at least minReadySeconds) targeted by this MachineDeployment.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// Total number of unavailable machines targeted by this MachineDeployment. This is the total number of
	// machines that are still required for the MachineDeployment to have 100% available capacity. They may
	// either be machines that are running but not yet available or machines that still have not been created.
	// +optional
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`

	// Represents the latest available observations of a MachineDeployment's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []MachineDeploymentCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Count of hash collisions for the MachineDeployment. The MachineDeployment controller uses this
	// field as a collision avoidance mechanism when it needs to create the name for the
	// newest MachineSet.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty"`

	// FailedMachines has summary of machines on which lastOperation Failed
	// +optional
	FailedMachines []*MachineSummary `json:"failedMachines,omitempty"`
}

type MachineDeploymentConditionType string

// These are valid conditions of a MachineDeployment.
const (
	// Available means the MachineDeployment is available, ie. at least the minimum available
	// replicas required are up and running for at least minReadySeconds.
	MachineDeploymentAvailable MachineDeploymentConditionType = "Available"

	// Progressing means the MachineDeployment is progressing. Progress for a MachineDeployment is
	// considered when a new machine set is created or adopted, and when new machines scale
	// up or old machines scale down. Progress is not estimated for paused MachineDeployments or
	// when progressDeadlineSeconds is not specified.
	MachineDeploymentProgressing MachineDeploymentConditionType = "Progressing"

	// ReplicaFailure is added in a MachineDeployment when one of its machines fails to be created
	// or deleted.
	MachineDeploymentReplicaFailure MachineDeploymentConditionType = "ReplicaFailure"

	// MachineDeploymentFrozen is added in a MachineDeployment when one of its machines fails to be created
	// or deleted.
	MachineDeploymentFrozen MachineDeploymentConditionType = "Frozen"
)

type RollbackConfig struct {
	// The revision to rollback to. If set to 0, rollback to the last revision.
	// +optional
	Revision int64 `json:"revision,omitempty"`
}

// MachineDeploymentCondition describes the state of a MachineDeployment at a certain point.
type MachineDeploymentCondition struct {
	// Type of MachineDeployment condition.
	Type MachineDeploymentConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status"`

	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// MachineDeploymentList is a list of MachineDeployments.
type MachineDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of MachineDeployments.
	Items []MachineDeployment `json:"items"`
}
