// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0


// Package machine is the internal version of the API.
package machine

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// WARNING!
// IF YOU MODIFY ANY OF THE TYPES HERE COPY THEM TO ./v1alpha1/types.go
// AND RUN  ./hack/generate-code

/********************** Machine APIs ***************/

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Machine TODO
type Machine struct {
	// ObjectMeta for machine object
	metav1.ObjectMeta

	// TypeMeta for machine object
	metav1.TypeMeta

	// Spec contains the specification of the machine
	Spec MachineSpec

	// Status contains fields depicting the status
	Status MachineStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineList is a collection of Machines.
type MachineList struct {
	// ObjectMeta for MachineList object
	metav1.TypeMeta

	// TypeMeta for MachineList object
	metav1.ListMeta

	// Items contains the list of machines
	Items []Machine
}

// MachineSpec is the specification of a Machine.
type MachineSpec struct {

	// Class contains the machineclass attributes of a machine
	Class ClassSpec

	// ProviderID represents the provider's unique ID given to a machine
	ProviderID string

	NodeTemplateSpec NodeTemplateSpec

	// Configuration for the machine-controller.
	*MachineConfiguration
}

// NodeTemplateSpec describes the data a node should have when created from a template
type NodeTemplateSpec struct {
	metav1.ObjectMeta

	Spec corev1.NodeSpec
}

// MachineTemplateSpec describes the data a machine should have when created from a template
type MachineTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta

	// Specification of the desired behavior of the machine.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec MachineSpec
}

// MachineConfiguration describes the configurations useful for the machine-controller.
type MachineConfiguration struct {
	// MachineDrainTimeout is the time out after which machine is deleted force-fully.
	MachineDrainTimeout *metav1.Duration

	// MachineHealthTimeout is the timeout after which machine is declared unhealthy/failed.
	MachineHealthTimeout *metav1.Duration

	// MachineCreationTimeout is the timeout after which machinie creation is declared failed.
	MachineCreationTimeout *metav1.Duration

	// MaxEvictRetries is the number of retries that will be attempted while draining the node.
	MaxEvictRetries *int32

	// NodeConditions are the set of conditions if set to true for MachineHealthTimeOut, machine will be declared failed.
	NodeConditions *string
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineTemplate describes a template for creating copies of a predefined machine.
type MachineTemplate struct {
	metav1.TypeMeta

	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta

	// Template defines the machines that will be created from this machine template.
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Template MachineTemplateSpec
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineTemplateList is a list of MachineTemplates.
type MachineTemplateList struct {
	metav1.TypeMeta

	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	metav1.ListMeta

	// List of machine templates
	Items []MachineTemplate
}

// ClassSpec is the class specification of machine
type ClassSpec struct {
	// API group to which it belongs
	APIGroup string

	// Kind for machine class
	Kind string

	// Name of machine class
	Name string
}

// CurrentStatus contains information about the current status of Machine.
type CurrentStatus struct {
	Phase MachinePhase

	TimeoutActive bool

	// Last update time of current status
	LastUpdateTime metav1.Time
}

// MachineStatus holds the most recently observed status of Machine.
type MachineStatus struct {
	// Conditions of this machine, same as node
	Conditions []corev1.NodeCondition

	// Last operation refers to the status of the last operation performed
	LastOperation LastOperation

	// Current status of the machine object
	CurrentStatus CurrentStatus

	// LastKnownState can store details of the last known state of the VM by the plugins.
	// It can be used by future operation calls to determine current infrastucture state
	// +optional
	LastKnownState string
}

// LastOperation suggests the last operation performed on the object
type LastOperation struct {
	// Description of the current operation
	Description string

	// ErrorCode of the current operation if any
	// +optional
	ErrorCode string

	// Last update time of current operation
	LastUpdateTime metav1.Time

	// State of operation
	State MachineState

	// Type of operation
	Type MachineOperationType
}

// MachinePhase is a label for the condition of a machines at the current time.
type MachinePhase string

// These are the valid statuses of machines.
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

	// MachineFailed means operation failed leading to machine status failure
	MachineFailed MachinePhase = "Failed"

	// MachineCrashLoopBackOff means creation or deletion of the machine is failing. It means that machine object is present but there is no corresponding VM.
	MachineCrashLoopBackOff MachinePhase = "CrashLoopBackOff"
)

// MachinePhase is a label for the condition of a machines at the current time.
type MachineState string

// These are the valid statuses of machines.
const (
	// MachineStatePending means there are operations pending on this machine state
	MachineStateProcessing MachineState = "Processing"

	// MachineStateFailed means operation failed leading to machine status failure
	MachineStateFailed MachineState = "Failed"

	// MachineStateSuccessful indicates that the node is not ready at the moment
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

	// MachineOperationHealthCheck indicates that the operation was a create
	MachineOperationHealthCheck MachineOperationType = "HealthCheck"

	// MachineOperationDelete indicates that the operation was a create
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

/********************** MachineSet APIs ***************/

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineSet TODO
type MachineSet struct {
	metav1.ObjectMeta

	metav1.TypeMeta

	Spec MachineSetSpec

	Status MachineSetStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineSetList is a collection of MachineSet.
type MachineSetList struct {
	metav1.TypeMeta

	metav1.ListMeta

	Items []MachineSet
}

// MachineSetSpec is the specification of a MachineSet.
type MachineSetSpec struct {
	Replicas int32

	Selector *metav1.LabelSelector

	MachineClass ClassSpec

	Template MachineTemplateSpec

	MinReadySeconds int32
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
	Type MachineSetConditionType

	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus

	// The last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time

	// The reason for the condition's last transition.
	Reason string

	// A human readable message indicating details about the transition.
	Message string
}

// MachineSetStatus holds the most recently observed status of MachineSet.
type MachineSetStatus struct {
	// Replicas is the number of actual replicas.
	Replicas int32

	// The number of pods that have labels matching the labels of the pod template of the replicaset.
	FullyLabeledReplicas int32

	// The number of ready replicas for this replica set.
	ReadyReplicas int32

	// The number of available replicas (ready for at least minReadySeconds) for this replica set.
	AvailableReplicas int32

	// ObservedGeneration is the most recent generation observed by the controller.
	ObservedGeneration int64

	// Represents the latest available observations of a replica set's current state.
	Conditions []MachineSetCondition

	// LastOperation performed
	LastOperation LastOperation

	// FailedMachines has summary of machines on which lastOperation Failed
	FailedMachines *[]MachineSummary
}

// MachineSummary store the summary of machine.
type MachineSummary struct {
	// Name of the machine object
	Name string

	// ProviderID represents the provider's unique ID given to a machine
	ProviderID string

	// Last operation refers to the status of the last operation performed
	LastOperation LastOperation

	// OwnerRef
	OwnerRef string
}

/********************** MachineDeployment APIs ***************/

// +genclient
// +genclient:method=GetScale,verb=get,subresource=scale,result=Scale
// +genclient:method=UpdateScale,verb=update,subresource=scale,input=Scale,result=Scale
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Deployment enables declarative updates for machines and MachineSets.
type MachineDeployment struct {
	metav1.TypeMeta
	// Standard object metadata.
	metav1.ObjectMeta

	// Specification of the desired behavior of the MachineDeployment.
	Spec MachineDeploymentSpec

	// Most recently observed status of the MachineDeployment.
	Status MachineDeploymentStatus
}

// MachineDeploymentSpec is the specification of the desired behavior of the MachineDeployment.
type MachineDeploymentSpec struct {
	// Number of desired machines. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	Replicas int32

	// Label selector for machines. Existing MachineSets whose machines are
	// selected by this will be the ones affected by this MachineDeployment.
	Selector *metav1.LabelSelector

	// Template describes the machines that will be created.
	Template MachineTemplateSpec

	// The MachineDeployment strategy to use to replace existing machines with new ones.
	// +patchStrategy=retainKeys
	Strategy MachineDeploymentStrategy

	// Minimum number of seconds for which a newly created machine should be ready
	// without any of its container crashing, for it to be considered available.
	// Defaults to 0 (machine will be considered available as soon as it is ready)
	MinReadySeconds int32

	// The number of old MachineSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	RevisionHistoryLimit *int32

	// Indicates that the MachineDeployment is paused and will not be processed by the
	// MachineDeployment controller.
	Paused bool

	// DEPRECATED.
	// The config this MachineDeployment is rolling back to. Will be cleared after rollback is done.
	RollbackTo *RollbackConfig

	// The maximum time in seconds for a MachineDeployment to make progress before it
	// is considered to be failed. The MachineDeployment controller will continue to
	// process failed MachineDeployments and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the MachineDeployment status. Note that progress will
	// not be estimated during the time a MachineDeployment is paused. This is not set
	// by default.
	ProgressDeadlineSeconds *int32
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DEPRECATED.
// MachineDeploymentRollback stores the information required to rollback a MachineDeployment.
type MachineDeploymentRollback struct {
	metav1.TypeMeta

	// Required: This must match the Name of a MachineDeployment.
	Name string

	// The annotations to be updated to a MachineDeployment
	UpdatedAnnotations map[string]string

	// The config of this MachineDeployment rollback.
	RollbackTo RollbackConfig
}

type RollbackConfig struct {
	// The revision to rollback to. If set to 0, rollback to the last revision.
	Revision int64
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
	Type MachineDeploymentStrategyType

	// Rolling update config params. Present only if MachineDeploymentStrategyType =
	// RollingUpdate.
	//---
	// TODO: Update this to follow our convention for oneOf, whatever we decide it
	// to be.
	RollingUpdate *RollingUpdateMachineDeployment
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
	MaxUnavailable *intstr.IntOrString

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
	MaxSurge *intstr.IntOrString
}

// MachineDeploymentStatus is the most recently observed status of the MachineDeployment.
type MachineDeploymentStatus struct {
	// The generation observed by the MachineDeployment controller.
	ObservedGeneration int64

	// Total number of non-terminated machines targeted by this MachineDeployment (their labels match the selector).
	Replicas int32

	// Total number of non-terminated machines targeted by this MachineDeployment that have the desired template spec.
	UpdatedReplicas int32

	// Total number of ready machines targeted by this MachineDeployment.
	ReadyReplicas int32

	// Total number of available machines (ready for at least minReadySeconds) targeted by this MachineDeployment.
	AvailableReplicas int32

	// Total number of unavailable machines targeted by this MachineDeployment. This is the total number of
	// machines that are still required for the MachineDeployment to have 100% available capacity. They may
	// either be machines that are running but not yet available or machines that still have not been created.
	UnavailableReplicas int32

	// Represents the latest available observations of a MachineDeployment's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []MachineDeploymentCondition

	// Count of hash collisions for the MachineDeployment. The MachineDeployment controller uses this
	// field as a collision avoidance mechanism when it needs to create the name for the
	// newest MachineSet.
	CollisionCount *int32

	// FailedMachines has summary of machines on which lastOperation Failed
	FailedMachines []*MachineSummary
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

// MachineDeploymentCondition describes the state of a MachineDeployment at a certain point.
type MachineDeploymentCondition struct {
	// Type of MachineDeployment condition.
	Type MachineDeploymentConditionType

	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus

	// The last time this condition was updated.
	LastUpdateTime metav1.Time

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time

	// The reason for the condition's last transition.
	Reason string

	// A human readable message indicating details about the transition.
	Message string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineDeploymentList is a list of MachineDeployments.
type MachineDeploymentList struct {
	metav1.TypeMeta
	// Standard list metadata.
	metav1.ListMeta

	// Items is the list of MachineDeployments.
	Items []MachineDeployment
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineClass can be used to templatize and re-use provider configuration
// across multiple Machines / MachineSets / MachineDeployments.
// +k8s:openapi-gen=true
// +resource:path=machineclasses
type MachineClass struct {
	metav1.TypeMeta

	// +optional
	metav1.ObjectMeta

	// +optional
	// NodeTemplate contains subfields to track all node resources and other node info required to scale nodegroup from zero
	NodeTemplate *NodeTemplate

	// CredentialsSecretRef can optionally store the credentials (in this case the SecretRef does not need to store them).
	// This might be useful if multiple machine classes with the same credentials but different user-datas are used.
	CredentialsSecretRef *corev1.SecretReference

	// Provider is the combination of name and location of cloud-specific drivers.
	// eg. awsdriver//127.0.0.1:8080
	Provider string

	// Provider-specific configuration to use during node creation.
	ProviderSpec runtime.RawExtension

	// SecretRef stores the necessary secrets such as credentials or userdata.
	SecretRef *corev1.SecretReference
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineClassList contains a list of MachineClasses
type MachineClassList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []MachineClass
}

// NodeTemplate contains subfields to track all node resources and other node info required to scale nodegroup from zero
type NodeTemplate struct {

	// Capacity contains subfields to track all node resources required to scale nodegroup from zero
	Capacity corev1.ResourceList

	// Instance type of the node belonging to nodeGroup
	InstanceType string

	// Region of the node belonging to nodeGroup
	Region string

	// Zone of the node belonging to nodeGroup
	Zone string

	// +optional
	// CPU Architecture of the node belonging to nodeGroup
	Architecture *string
}
