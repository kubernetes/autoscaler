/*
Copyright 2023 The Kubernetes Authors.

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

package drainability

import (
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
)

// OutcomeType identifies the action that should be taken when it comes to
// draining a pod.
type OutcomeType int

const (
	// UndefinedOutcome means the Rule did not match and that another one
	// has to be applied.
	UndefinedOutcome OutcomeType = iota
	// DrainOk means that the pod can be drained.
	DrainOk
	// BlockDrain means that the pod should block drain for its entire node.
	BlockDrain
	// SkipDrain means that the pod doesn't block drain of other pods, but
	// should not be drained itself.
	SkipDrain
)

// Status contains all information about drainability of a single pod.
// TODO(x13n): Move values from drain.BlockingPodReason to some typed string.
type Status struct {
	// Outcome indicates what can happen when it comes to draining a
	// specific pod.
	Outcome OutcomeType
	// Overrides specifies Outcomes that should be trumped by this Status.
	// If Overrides is empty, this Status is returned immediately.
	// If Overrides is non-empty, we continue running the remaining Rules. If a
	// Rule is encountered that matches one of the Outcomes specified by this
	// field, this Status will will be returned instead.
	Overrides []OutcomeType
	// Reason contains the reason why a pod is blocking node drain. It is
	// set only when Outcome is BlockDrain.
	BlockingReason drain.BlockingPodReason
	// Error contains an optional error message.
	Error error
}

// NewDrainableStatus returns a new Status indicating that a pod can be drained.
func NewDrainableStatus() Status {
	return Status{
		Outcome: DrainOk,
	}
}

// NewBlockedStatus returns a new Status indicating that a pod is blocked and cannot be drained.
func NewBlockedStatus(reason drain.BlockingPodReason, err error) Status {
	return Status{
		Outcome:        BlockDrain,
		BlockingReason: reason,
		Error:          err,
	}
}

// NewSkipStatus returns a new Status indicating that a pod should be skipped when draining a node.
func NewSkipStatus() Status {
	return Status{
		Outcome: SkipDrain,
	}
}

// NewUndefinedStatus returns a new Status that doesn't contain a decision.
func NewUndefinedStatus() Status {
	return Status{}
}
