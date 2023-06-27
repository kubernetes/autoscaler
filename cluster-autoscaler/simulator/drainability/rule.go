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

	apiv1 "k8s.io/api/core/v1"
)

// Status indicates whether a pod can be drained, with an optional error message when not.
// TODO(x13n): Move values from drain.BlockingPodReason to some typed string.
type Status struct {
	// Matched indicates whether the Rule can be applied to a given pod.
	// `false` indicates that the Rule doesn't match and that another one
	// has to be applied.
	Matched bool
	// Reason contains the decision whether to drain the pod or not.
	Reason drain.BlockingPodReason
	// Error contains an optional error message.
	Error error
}

// NewDrainableStatus returns a new Status indicating that a pod can be drained.
func NewDrainableStatus() Status {
	return Status{
		Matched: true,
		Reason:  drain.NoReason,
	}
}

// NewBlockedStatus returns a new Status indicating that a pod is blocked and cannot be drained.
func NewBlockedStatus(reason drain.BlockingPodReason, err error) Status {
	return Status{
		Matched: true,
		Reason:  reason,
		Error:   err,
	}
}

// NewUnmatchedStatus returns a new Status that doesn't contain a decision.
func NewUnmatchedStatus() Status {
	return Status{}
}

// Rule determines whether a given pod can be drained or not.
type Rule interface {
	// Drainable determines whether a given pod is drainable according to
	// the specific Rule.
	Drainable(*apiv1.Pod) Status
}
