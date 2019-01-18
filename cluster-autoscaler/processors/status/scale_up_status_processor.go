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

package status

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
)

// ScaleUpStatus is the status of a scale-up attempt. This includes information
// on if scale-up happened, description of scale-up operation performed and
// status of pods that took part in the scale-up evaluation.
type ScaleUpStatus struct {
	Result                  ScaleUpResult
	ScaleUpInfos            []nodegroupset.ScaleUpInfo
	PodsTriggeredScaleUp    []*apiv1.Pod
	PodsRemainUnschedulable []NoScaleUpInfo
	PodsAwaitEvaluation     []*apiv1.Pod
}

// NoScaleUpInfo contains information about a pod that didn't trigger scale-up.
type NoScaleUpInfo struct {
	Pod                *apiv1.Pod
	RejectedNodeGroups map[string]Reasons
	SkippedNodeGroups  map[string]Reasons
}

// ScaleUpResult represents the result of a scale up.
type ScaleUpResult int

const (
	// ScaleUpSuccessful - a scale-up successfully occurred.
	ScaleUpSuccessful ScaleUpResult = iota
	// ScaleUpError - an unexpected error occurred during the scale-up attempt.
	ScaleUpError
	// ScaleUpNoOptionsAvailable - there were no node groups that could be considered for the scale-up.
	ScaleUpNoOptionsAvailable
	// ScaleUpNotNeeded - there was no need for a scale-up e.g. because there were no unschedulable pods.
	ScaleUpNotNeeded
	// ScaleUpNotTried - the scale up wasn't even attempted, e.g. an autoscaling iteration was skipped, or
	// an error occurred before the scale up logic.
	ScaleUpNotTried
	// ScaleUpInCooldown - the scale up wasn't even attempted, because it's in a cooldown state (it's suspended for a scheduled period of time).
	ScaleUpInCooldown
)

// WasSuccessful returns true if the scale-up was successful.
func (s *ScaleUpStatus) WasSuccessful() bool {
	return s.Result == ScaleUpSuccessful
}

// Reasons interface provides a list of reasons for why something happened or didn't happen.
type Reasons interface {
	Reasons() []string
}

// ScaleUpStatusProcessor processes the status of the cluster after a scale-up.
type ScaleUpStatusProcessor interface {
	Process(context *context.AutoscalingContext, status *ScaleUpStatus)
	CleanUp()
}

// NewDefaultScaleUpStatusProcessor creates a default instance of ScaleUpStatusProcessor.
func NewDefaultScaleUpStatusProcessor() ScaleUpStatusProcessor {
	return &EventingScaleUpStatusProcessor{}
}

// NoOpScaleUpStatusProcessor is a ScaleUpStatusProcessor implementations useful for testing.
type NoOpScaleUpStatusProcessor struct{}

// Process processes the status of the cluster after a scale-up.
func (p *NoOpScaleUpStatusProcessor) Process(context *context.AutoscalingContext, status *ScaleUpStatus) {
}

// CleanUp cleans up the processor's internal structures.
func (p *NoOpScaleUpStatusProcessor) CleanUp() {
}
