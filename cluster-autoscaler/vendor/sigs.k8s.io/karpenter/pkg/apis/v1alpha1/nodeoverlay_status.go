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

package v1alpha1

import (
	"github.com/awslabs/operatorpkg/status"
)

const (
	// ConditionTypeValidationSucceeded = "ValidationSucceeded" condition indicates that the
	// runtime-based configuration is valid and conflict for this NodeOverlay
	ConditionTypeValidationSucceeded = "ValidationSucceeded"
)

// NodeOverlayStatus defines the observed state of NodeOverlay
type NodeOverlayStatus struct {
	// Conditions contains signals for health and readiness
	// +optional
	// +listType=map
	// +listMapKey=type
	//nolint:kubeapilinter
	Conditions []status.Condition `json:"conditions,omitempty"`
}

func (in *NodeOverlay) StatusConditions(opts ...status.ForOption) status.ConditionSet {
	return status.NewReadyConditions(
		ConditionTypeValidationSucceeded,
	).For(in, opts...)
}

func (in *NodeOverlay) GetConditions() []status.Condition {
	return in.Status.Conditions
}

func (in *NodeOverlay) SetConditions(conditions []status.Condition) {
	in.Status.Conditions = conditions
}
