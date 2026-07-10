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
)

const (
	// ConditionTypeValidationSucceeded = "ValidationSucceeded" condition indicates that the
	// runtime-based configuration is valid for this NodePool
	ConditionTypeValidationSucceeded = "ValidationSucceeded"
	// ConditionTypeNodeClassReady = "NodeClassReady" condition indicates that underlying nodeClass was resolved and is reporting as Ready
	ConditionTypeNodeClassReady = "NodeClassReady"
	// ConditionTypeNodeRegistrationHealthy = "NodeRegistrationHealthy" condition indicates if a misconfiguration exists that is preventing successful node launch/registrations that requires manual investigation
	ConditionTypeNodeRegistrationHealthy = "NodeRegistrationHealthy"
)

// NodePoolStatus defines the observed state of NodePool
type NodePoolStatus struct {
	//nolint:kubeapilinter
	// Resources is the list of resources that have been provisioned.
	// +optional
	Resources v1.ResourceList `json:"resources,omitempty"`
	//nolint:kubeapilinter
	// Nodes is the count of nodes associated with this NodePool
	// +kubebuilder:default:=0
	// +optional
	Nodes *int64 `json:"nodes"`
	//nolint:kubeapilinter
	// NodeClassObservedGeneration represents the observed nodeClass generation for referenced nodeClass. If this does not match
	// the actual NodeClass Generation, NodeRegistrationHealthy status condition on the NodePool will be reset
	// +optional
	NodeClassObservedGeneration int64 `json:"nodeClassObservedGeneration,omitempty"`
	// Conditions contains signals for health and readiness
	// +optional
	// +listType=map
	// +listMapKey=type
	//nolint:kubeapilinter
	Conditions []status.Condition `json:"conditions,omitempty"`
}

func (in *NodePool) StatusConditions(opts ...status.ForOption) status.ConditionSet {
	return status.NewReadyConditions(
		ConditionTypeValidationSucceeded,
		ConditionTypeNodeClassReady,
	).For(in, opts...)
}

func (in *NodePool) GetConditions() []status.Condition {
	return in.Status.Conditions
}

func (in *NodePool) SetConditions(conditions []status.Condition) {
	in.Status.Conditions = conditions
}
