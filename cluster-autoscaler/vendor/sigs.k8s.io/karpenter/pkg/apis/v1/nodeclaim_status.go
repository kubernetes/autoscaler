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
	"github.com/awslabs/operatorpkg/status"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionTypeLaunched             = "Launched"
	ConditionTypeRegistered           = "Registered"
	ConditionTypeInitialized          = "Initialized"
	ConditionTypeConsolidatable       = "Consolidatable"
	ConditionTypeDrifted              = "Drifted"
	ConditionTypeDrained              = "Drained"
	ConditionTypeVolumesDetached      = "VolumesDetached"
	ConditionTypeInstanceTerminating  = "InstanceTerminating"
	ConditionTypeConsistentStateFound = "ConsistentStateFound"
	ConditionTypeDisruptionReason     = "DisruptionReason"
)

// NodeClaimStatus defines the observed state of NodeClaim
type NodeClaimStatus struct {
	//nolint:kubeapilinter
	// NodeName is the name of the corresponding node object
	// +optional
	NodeName string `json:"nodeName,omitempty"`
	//nolint:kubeapilinter
	// ProviderID of the corresponding node object
	// +optional
	ProviderID string `json:"providerID,omitempty"`
	//nolint:kubeapilinter
	// ImageID is an identifier for the image that runs on the node
	// +optional
	ImageID string `json:"imageID,omitempty"`
	//nolint:kubeapilinter
	// Capacity is the estimated full capacity of the node
	// +optional
	Capacity v1.ResourceList `json:"capacity,omitempty"`
	//nolint:kubeapilinter
	// Allocatable is the estimated allocatable capacity of the node
	// +optional
	Allocatable v1.ResourceList `json:"allocatable,omitempty"`
	// Conditions contains signals for health and readiness
	// +optional
	// +listType=map
	// +listMapKey=type
	//nolint:kubeapilinter
	Conditions []status.Condition `json:"conditions,omitempty"`
	//nolint:kubeapilinter
	// LastPodEventTime is updated with the last time a pod was scheduled
	// or removed from the node. A pod going terminal or terminating
	// is also considered as removed.
	// +optional
	LastPodEventTime metav1.Time `json:"lastPodEventTime,omitempty"`
}

func (in *NodeClaim) StatusConditions(opts ...status.ForOption) status.ConditionSet {
	return status.NewReadyConditions(
		ConditionTypeLaunched,
		ConditionTypeRegistered,
		ConditionTypeInitialized,
	).For(in, opts...)
}

func (in *NodeClaim) GetConditions() []status.Condition {
	return in.Status.Conditions
}

func (in *NodeClaim) SetConditions(conditions []status.Condition) {
	in.Status.Conditions = conditions
}
