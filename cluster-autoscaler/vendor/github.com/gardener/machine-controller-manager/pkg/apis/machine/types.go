/*
Copyright (c) 2017 SAP SE or an SAP affiliate company. All rights reserved.

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
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta

	// Specification of the desired behavior of the machine.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
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
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta

	// Template defines the machines that will be created from this machine template.
	// https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	Template MachineTemplateSpec
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineTemplateList is a list of MachineTemplates.
type MachineTemplateList struct {
	metav1.TypeMeta

	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
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
	// Node string
	Node string

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

	// MachineRunning means node is terminating
	MachineTerminating MachinePhase = "Terminating"

	// MachineUnknown indicates that the node is not ready at the movement
	MachineUnknown MachinePhase = "Unknown"

	// MachineFailed means operation failed leading to machine status failure
	MachineFailed MachinePhase = "Failed"

	// MachineCrashLoopBackOff means creation or deletion of the machine is failing.
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

/********************** OpenStackMachineClass APIs ***************/

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenStackMachineClass TODO
type OpenStackMachineClass struct {
	metav1.ObjectMeta

	metav1.TypeMeta

	Spec OpenStackMachineClassSpec
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenStackMachineClassList is a collection of OpenStackMachineClasses.
type OpenStackMachineClassList struct {
	metav1.TypeMeta

	metav1.ListMeta

	Items []OpenStackMachineClass
}

// OpenStackMachineClassSpec is the specification of a OpenStackMachineClass.
type OpenStackMachineClassSpec struct {
	ImageID              string
	ImageName            string
	Region               string
	AvailabilityZone     string
	FlavorName           string
	KeyName              string
	SecurityGroups       []string
	Tags                 map[string]string
	NetworkID            string
	Networks             []OpenStackNetwork
	SubnetID             *string
	SecretRef            *corev1.SecretReference
	CredentialsSecretRef *corev1.SecretReference
	PodNetworkCidr       string
	RootDiskSize         int // in GB
	UseConfigDrive       *bool
	ServerGroupID        *string
}

type OpenStackNetwork struct {
	Id         string
	Name       string
	PodNetwork bool
}

/********************** AWSMachineClass APIs ***************/

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSMachineClass TODO
type AWSMachineClass struct {
	metav1.ObjectMeta

	metav1.TypeMeta

	Spec AWSMachineClassSpec
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSMachineClassList is a collection of AWSMachineClasses.
type AWSMachineClassList struct {
	metav1.TypeMeta

	metav1.ListMeta

	Items []AWSMachineClass
}

// AWSMachineClassSpec is the specification of a AWSMachineClass.
type AWSMachineClassSpec struct {
	AMI                  string
	Region               string
	BlockDevices         []AWSBlockDeviceMappingSpec
	EbsOptimized         bool
	IAM                  AWSIAMProfileSpec
	MachineType          string
	KeyName              string
	Monitoring           bool
	NetworkInterfaces    []AWSNetworkInterfaceSpec
	Tags                 map[string]string
	SpotPrice            *string
	SecretRef            *corev1.SecretReference
	CredentialsSecretRef *corev1.SecretReference

	// TODO add more here
}

type AWSBlockDeviceMappingSpec struct {

	// The device name exposed to the machine (for example, /dev/sdh or xvdh).
	DeviceName string

	// Parameters used to automatically set up EBS volumes when the machine is
	// launched.
	Ebs AWSEbsBlockDeviceSpec

	// Suppresses the specified device included in the block device mapping of the
	// AMI.
	NoDevice string

	// The virtual device name (ephemeralN). Machine store volumes are numbered
	// starting from 0. An machine type with 2 available machine store volumes
	// can specify mappings for ephemeral0 and ephemeral1.The number of available
	// machine store volumes depends on the machine type. After you connect to
	// the machine, you must mount the volume.
	//
	// Constraints: For M3 machines, you must specify machine store volumes in
	// the block device mapping for the machine. When you launch an M3 machine,
	// we ignore any machine store volumes specified in the block device mapping
	// for the AMI.
	VirtualName string
}

// Describes a block device for an EBS volume.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/ec2-2016-11-15/EbsBlockDevice
type AWSEbsBlockDeviceSpec struct {

	// Indicates whether the EBS volume is deleted on machine termination.
	DeleteOnTermination *bool

	// Indicates whether the EBS volume is encrypted. Encrypted Amazon EBS volumes
	// may only be attached to machines that support Amazon EBS encryption.
	Encrypted bool

	// The number of I/O operations per second (IOPS) that the volume supports.
	// For io1, this represents the number of IOPS that are provisioned for the
	// volume. For gp2, this represents the baseline performance of the volume and
	// the rate at which the volume accumulates I/O credits for bursting. For more
	// information about General Purpose SSD baseline performance, I/O credits,
	// and bursting, see Amazon EBS Volume Types (http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumeTypes.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	//
	// Constraint: Range is 100-20000 IOPS for io1 volumes and 100-10000 IOPS for
	// gp2 volumes.
	//
	// Condition: This parameter is required for requests to create io1 volumes;
	// it is not used in requests to create gp2, st1, sc1, or standard volumes.
	Iops int64

	// Identifier (key ID, key alias, ID ARN, or alias ARN) for a customer managed
	// CMK under which the EBS volume is encrypted.
	//
	// This parameter is only supported on BlockDeviceMapping objects called by
	// RunInstances (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RunInstances.html),
	// RequestSpotFleet (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RequestSpotFleet.html),
	// and RequestSpotInstances (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RequestSpotInstances.html).
	KmsKeyID *string

	// The ID of the snapshot.
	SnapshotID *string

	// The size of the volume, in GiB.
	//
	// Constraints: 1-16384 for General Purpose SSD (gp2), 4-16384 for Provisioned
	// IOPS SSD (io1), 500-16384 for Throughput Optimized HDD (st1), 500-16384 for
	// Cold HDD (sc1), and 1-1024 for Magnetic (standard) volumes. If you specify
	// a snapshot, the volume size must be equal to or larger than the snapshot
	// size.
	//
	// Default: If you're creating the volume from a snapshot and don't specify
	// a volume size, the default is the snapshot size.
	VolumeSize int64

	// The volume type: gp2, io1, st1, sc1, or standard.
	//
	// Default: standard
	VolumeType string
}

// Describes an IAM machine profile.
type AWSIAMProfileSpec struct {
	// The Amazon Resource Name (ARN) of the machine profile.
	ARN string

	// The name of the machine profile.
	Name string
}

// Describes a network interface.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/ec2-2016-11-15/MachineAWSNetworkInterfaceSpecification
type AWSNetworkInterfaceSpec struct {

	// Indicates whether to assign a public IPv4 address to an machine you launch
	// in a VPC. The public IP address can only be assigned to a network interface
	// for eth0, and can only be assigned to a new network interface, not an existing
	// one. You cannot specify more than one network interface in the request. If
	// launching into a default subnet, the default value is true.
	AssociatePublicIPAddress *bool

	// If set to true, the interface is deleted when the machine is terminated.
	// You can specify true only if creating a new network interface when launching
	// an machine.
	DeleteOnTermination *bool

	// The description of the network interface. Applies only if creating a network
	// interface when launching an machine.
	Description *string

	// The IDs of the security groups for the network interface. Applies only if
	// creating a network interface when launching an machine.
	SecurityGroupIDs []string

	// The ID of the subnet associated with the network string. Applies only if
	// creating a network interface when launching an machine.
	SubnetID string
}

/********************** AzureMachineClass APIs ***************/

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AzureMachineClass TODO
type AzureMachineClass struct {
	metav1.ObjectMeta

	metav1.TypeMeta

	Spec AzureMachineClassSpec
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AzureMachineClassList is a collection of AzureMachineClasses.
type AzureMachineClassList struct {
	metav1.TypeMeta

	metav1.ListMeta

	Items []AzureMachineClass
}

// AzureMachineClassSpec is the specification of a AzureMachineClass.
type AzureMachineClassSpec struct {
	Location             string
	Tags                 map[string]string
	Properties           AzureVirtualMachineProperties
	ResourceGroup        string
	SubnetInfo           AzureSubnetInfo
	SecretRef            *corev1.SecretReference
	CredentialsSecretRef *corev1.SecretReference
}

// AzureVirtualMachineProperties is describes the properties of a Virtual Machine.
type AzureVirtualMachineProperties struct {
	HardwareProfile AzureHardwareProfile
	StorageProfile  AzureStorageProfile
	OsProfile       AzureOSProfile
	NetworkProfile  AzureNetworkProfile
	AvailabilitySet *AzureSubResource
	IdentityID      *string
	Zone            *int
	MachineSet      *AzureMachineSetConfig
}

// AzureHardwareProfile is specifies the hardware settings for the virtual machine.
// Refer github.com/Azure/azure-sdk-for-go/arm/compute/models.go for VMSizes
type AzureHardwareProfile struct {
	VMSize string
}

// AzureStorageProfile is specifies the storage settings for the virtual machine disks.
type AzureStorageProfile struct {
	ImageReference AzureImageReference
	OsDisk         AzureOSDisk
	DataDisks      []AzureDataDisk
}

// AzureImageReference is specifies information about the image to use. You can specify information about platform images,
// marketplace images, or virtual machine images. This element is required when you want to use a platform image,
// marketplace image, or virtual machine image, but is not used in other creation operations.
type AzureImageReference struct {
	ID  string
	URN *string
}

// AzureOSDisk is specifies information about the operating system disk used by the virtual machine. <br><br> For more
// information about disks, see [About disks and VHDs for Azure virtual
// machines](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-windows-about-disks-vhds?toc=%2fazure%2fvirtual-machines%2fwindows%2ftoc.json).
type AzureOSDisk struct {
	Name         string
	Caching      string
	ManagedDisk  AzureManagedDiskParameters
	DiskSizeGB   int32
	CreateOption string
}

type AzureDataDisk struct {
	Name               string
	Lun                *int32
	Caching            string
	StorageAccountType string
	DiskSizeGB         int32
}

// AzureManagedDiskParameters is the parameters of a managed disk.
type AzureManagedDiskParameters struct {
	ID                 string
	StorageAccountType string
}

// AzureOSProfile is specifies the operating system settings for the virtual machine.
type AzureOSProfile struct {
	ComputerName       string
	AdminUsername      string
	AdminPassword      string
	CustomData         string
	LinuxConfiguration AzureLinuxConfiguration
}

// AzureLinuxConfiguration is specifies the Linux operating system settings on the virtual machine. <br><br>For a list of
// supported Linux distributions, see [Linux on Azure-Endorsed
// Distributions](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-endorsed-distros?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json)
// <br><br> For running non-endorsed distributions, see [Information for Non-Endorsed
// Distributions](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-create-upload-generic?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json).
type AzureLinuxConfiguration struct {
	DisablePasswordAuthentication bool
	SSH                           AzureSSHConfiguration
}

// AzureSSHConfiguration is SSH configuration for Linux based VMs running on Azure
type AzureSSHConfiguration struct {
	PublicKeys AzureSSHPublicKey
}

// AzureSSHPublicKey is contains information about SSH certificate public key and the path on the Linux VM where the public
// key is placed.
type AzureSSHPublicKey struct {
	Path    string
	KeyData string
}

// AzureNetworkProfile is specifies the network interfaces of the virtual machine.
type AzureNetworkProfile struct {
	NetworkInterfaces     AzureNetworkInterfaceReference
	AcceleratedNetworking *bool
}

// AzureNetworkInterfaceReference is describes a network interface reference.
type AzureNetworkInterfaceReference struct {
	ID string
	*AzureNetworkInterfaceReferenceProperties
}

// AzureNetworkInterfaceReferenceProperties is describes a network interface reference properties.
type AzureNetworkInterfaceReferenceProperties struct {
	Primary bool
}

// AzureSubResource is the Sub Resource definition.
type AzureSubResource struct {
	ID string
}

// AzureSubnetInfo is the information containing the subnet details
type AzureSubnetInfo struct {
	VnetName          string
	VnetResourceGroup *string
	SubnetName        string
}

// AzureMachineSetConfig contains the information about the machine set
type AzureMachineSetConfig struct {
	ID   string
	Kind string
}

const (
	// MachineSetKindAvailabilitySet is the machine set kind for AvailabilitySet
	MachineSetKindAvailabilitySet string = "availabilityset"
	// MachineSetKindVMO is the machine set kind for VirtualMachineScaleSet Orchestration Mode VM (VMO)
	MachineSetKindVMO string = "vmo"
)

/********************** GCPMachineClass APIs ***************/

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCPMachineClass TODO
type GCPMachineClass struct {
	metav1.ObjectMeta

	metav1.TypeMeta

	Spec GCPMachineClassSpec
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCPMachineClassList is a collection of GCPMachineClasses.
type GCPMachineClassList struct {
	metav1.TypeMeta

	metav1.ListMeta

	Items []GCPMachineClass
}

// GCPMachineClassSpec is the specification of a GCPMachineClass.
type GCPMachineClassSpec struct {
	CanIpForward         bool
	DeletionProtection   bool
	Description          *string
	Disks                []*GCPDisk
	Labels               map[string]string
	MachineType          string
	Metadata             []*GCPMetadata
	NetworkInterfaces    []*GCPNetworkInterface
	Scheduling           GCPScheduling
	SecretRef            *corev1.SecretReference
	CredentialsSecretRef *corev1.SecretReference
	ServiceAccounts      []GCPServiceAccount
	Tags                 []string
	Region               string
	Zone                 string
}

// GCPDisk describes disks for GCP.
type GCPDisk struct {
	AutoDelete *bool
	Boot       bool
	SizeGb     int64
	Type       string
	Interface  string
	Image      string
	Labels     map[string]string
}

// GCPMetadata describes metadata for GCP.
type GCPMetadata struct {
	Key   string
	Value *string
}

// GCPNetworkInterface describes network interfaces for GCP
type GCPNetworkInterface struct {
	DisableExternalIP bool
	Network           string
	Subnetwork        string
}

// GCPScheduling describes scheduling configuration for GCP.
type GCPScheduling struct {
	AutomaticRestart  bool
	OnHostMaintenance string
	Preemptible       bool
}

// GCPServiceAccount describes service accounts for GCP.
type GCPServiceAccount struct {
	Email  string
	Scopes []string
}

/********************** AlicloudMachineClass APIs ***************/

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlicloudMachineClass
type AlicloudMachineClass struct {
	metav1.ObjectMeta

	metav1.TypeMeta

	Spec AlicloudMachineClassSpec
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlicloudMachineClassList is a collection of AlicloudMachineClasses.
type AlicloudMachineClassList struct {
	metav1.TypeMeta

	metav1.ListMeta

	Items []AlicloudMachineClass
}

// AlicloudMachineClassSpec is the specification of a AlicloudMachineClass.
type AlicloudMachineClassSpec struct {
	ImageID                 string
	InstanceType            string
	Region                  string
	ZoneID                  string
	SecurityGroupID         string
	VSwitchID               string
	PrivateIPAddress        string
	SystemDisk              *AlicloudSystemDisk
	DataDisks               []AlicloudDataDisk
	InstanceChargeType      string
	InternetChargeType      string
	InternetMaxBandwidthIn  *int
	InternetMaxBandwidthOut *int
	SpotStrategy            string
	IoOptimized             string
	Tags                    map[string]string
	KeyPairName             string
	SecretRef               *corev1.SecretReference
	CredentialsSecretRef    *corev1.SecretReference
}

// AlicloudSystemDisk describes SystemDisk for Alicloud.
type AlicloudSystemDisk struct {
	Category string
	Size     int
}

// AlicloudDataDisk describes DataDisk for Alicloud.
type AlicloudDataDisk struct {
	Name               string
	Category           string
	Description        string
	Encrypted          bool
	Size               int
	DeleteWithInstance *bool
}

/********************** PacketMachineClass APIs ***************/

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PacketMachineClass TODO
type PacketMachineClass struct {
	metav1.ObjectMeta

	metav1.TypeMeta

	Spec PacketMachineClassSpec
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PacketMachineClassList is a collection of PacketMachineClasses.
type PacketMachineClassList struct {
	metav1.TypeMeta

	metav1.ListMeta

	Items []PacketMachineClass
}

// PacketMachineClassSpec is the specification of a PacketMachineClass.
type PacketMachineClassSpec struct {
	Facility     []string // required
	MachineType  string   // required
	OS           string   // required
	ProjectID    string   // required
	BillingCycle string
	Tags         []string
	SSHKeys      []string
	UserData     string

	SecretRef            *corev1.SecretReference
	CredentialsSecretRef *corev1.SecretReference

	// TODO add more here
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
}
