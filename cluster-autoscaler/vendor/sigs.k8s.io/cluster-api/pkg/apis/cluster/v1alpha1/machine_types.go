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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
)

// Finalizer is set on PrepareForCreate callback
const MachineFinalizer = "machine.cluster.k8s.io"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

/// [Machine]
// Machine is the Schema for the machines API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Machine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineSpec   `json:"spec,omitempty"`
	Status MachineStatus `json:"status,omitempty"`
}

/// [Machine]

/// [MachineSpec]
// MachineSpec defines the desired state of Machine
type MachineSpec struct {
	// ObjectMeta will autopopulate the Node created. Use this to
	// indicate what labels, annotations, name prefix, etc., should be used
	// when creating the Node.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Taints is the full, authoritative list of taints to apply to the corresponding
	// Node. This list will overwrite any modifications made to the Node on
	// an ongoing basis.
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty"`

	// ProviderSpec details Provider-specific configuration to use during node creation.
	// +optional
	ProviderSpec ProviderSpec `json:"providerSpec"`

	// Versions of key software to use. This field is optional at cluster
	// creation time, and omitting the field indicates that the cluster
	// installation tool should select defaults for the user. These
	// defaults may differ based on the cluster installer, but the tool
	// should populate the values it uses when persisting Machine objects.
	// A Machine spec missing this field at runtime is invalid.
	// +optional
	Versions MachineVersionInfo `json:"versions,omitempty"`

	// ConfigSource is used to populate in the associated Node for dynamic kubelet config. This
	// field already exists in Node, so any updates to it in the Machine
	// spec will be automatically copied to the linked NodeRef from the
	// status. The rest of dynamic kubelet config support should then work
	// as-is.
	// +optional
	ConfigSource *corev1.NodeConfigSource `json:"configSource,omitempty"`

	// ProviderID is the identification ID of the machine provided by the provider.
	// This field must match the provider ID as seen on the node object corresponding to this machine.
	// This field is required by higher level consumers of cluster-api. Example use case is cluster autoscaler
	// with cluster-api as provider. Clean-up login in the autoscaler compares machines v/s nodes to find out
	// machines at provider which could not get registered as Kubernetes nodes. With cluster-api as a
	// generic out-of-tree provider for autoscaler, this field is required by autoscaler to be
	// able to have a provider view of the list of machines. Another list of nodes is queries from the k8s apiserver
	// and then comparison is done to find out unregistered machines and are marked for delete.
	// This field will be set by the actuators and consumed by higher level entities like autoscaler  who will
	// be interfacing with cluster-api as generic provider.
	// +optional
	ProviderID *string `json:"providerID,omitempty"`
}

/// [MachineSpec]

/// [MachineStatus]
// MachineStatus defines the observed state of Machine
type MachineStatus struct {
	// NodeRef will point to the corresponding Node if it exists.
	// +optional
	NodeRef *corev1.ObjectReference `json:"nodeRef,omitempty"`

	// LastUpdated identifies when this status was last observed.
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// Versions specifies the current versions of software on the corresponding Node (if it
	// exists). This is provided for a few reasons:
	//
	// 1) It is more convenient than checking the NodeRef, traversing it to
	//    the Node, and finding the appropriate field in Node.Status.NodeInfo
	//    (which uses different field names and formatting).
	// 2) It removes some of the dependency on the structure of the Node,
	//    so that if the structure of Node.Status.NodeInfo changes, only
	//    machine controllers need to be updated, rather than every client
	//    of the Machines API.
	// 3) There is no other simple way to check the control plane
	//    version. A client would have to connect directly to the apiserver
	//    running on the target node in order to find out its version.
	// +optional
	Versions *MachineVersionInfo `json:"versions,omitempty"`

	// ErrorReason will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a succinct value suitable
	// for machine interpretation.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	ErrorReason *common.MachineStatusError `json:"errorReason,omitempty"`

	// ErrorMessage will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`

	// ProviderStatus details a Provider-specific status.
	// It is recommended that providers maintain their
	// own versioned API types that should be
	// serialized/deserialized from this field.
	// +optional
	ProviderStatus *runtime.RawExtension `json:"providerStatus,omitempty"`

	// Addresses is a list of addresses assigned to the machine. Queried from cloud provider, if available.
	// +optional
	Addresses []corev1.NodeAddress `json:"addresses,omitempty"`

	// Conditions lists the conditions synced from the node conditions of the corresponding node-object.
	// Machine-controller is responsible for keeping conditions up-to-date.
	// MachineSet controller will be taking these conditions as a signal to decide if
	// machine is healthy or needs to be replaced.
	// Refer: https://kubernetes.io/docs/concepts/architecture/nodes/#condition
	// +optional
	Conditions []corev1.NodeCondition `json:"conditions,omitempty"`

	// LastOperation describes the last-operation performed by the machine-controller.
	// This API should be useful as a history in terms of the latest operation performed on the
	// specific machine. It should also convey the state of the latest-operation for example if
	// it is still on-going, failed or completed successfully.
	// +optional
	LastOperation *LastOperation `json:"lastOperation,omitempty"`

	// Phase represents the current phase of machine actuation.
	// E.g. Pending, Running, Terminating, Failed etc.
	// +optional
	Phase *string `json:"phase,omitempty"`
}

// LastOperation represents the detail of the last performed operation on the MachineObject.
type LastOperation struct {
	// Description is the human-readable description of the last operation.
	Description *string `json:"description,omitempty"`

	// LastUpdated is the timestamp at which LastOperation API was last-updated.
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// State is the current status of the last performed operation.
	// E.g. Processing, Failed, Successful etc
	State *string `json:"state,omitempty"`

	// Type is the type of operation which was last performed.
	// E.g. Create, Delete, Update etc
	Type *string `json:"type,omitempty"`
}

/// [MachineStatus]

/// [MachineVersionInfo]
type MachineVersionInfo struct {
	// Kubelet is the semantic version of kubelet to run
	Kubelet string `json:"kubelet"`

	// ControlPlane is the semantic version of the Kubernetes control plane to
	// run. This should only be populated when the machine is a
	// control plane.
	// +optional
	ControlPlane string `json:"controlPlane,omitempty"`
}

/// [MachineVersionInfo]

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineList contains a list of Machine
type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Machine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Machine{}, &MachineList{})
}
