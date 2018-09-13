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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
)

// ScaleDownStatus represents the state of scale down.
type ScaleDownStatus struct {
	Result            ScaleDownResult
	ScaledDownNodes   []*ScaleDownNode
	NodeDeleteResults map[string]error
}

// ScaleDownNode represents the state of a node that's being scaled down.
type ScaleDownNode struct {
	Node        *apiv1.Node
	NodeGroup   cloudprovider.NodeGroup
	EvictedPods []*apiv1.Pod
	UtilInfo    simulator.UtilizationInfo
}

// ScaleDownResult represents the result of scale down.
type ScaleDownResult int

const (
	// ScaleDownError - scale down finished with error.
	ScaleDownError ScaleDownResult = iota
	// ScaleDownNoUnneeded - no unneeded nodes and no errors.
	ScaleDownNoUnneeded
	// ScaleDownNoNodeDeleted - unneeded nodes present but not available for deletion.
	ScaleDownNoNodeDeleted
	// ScaleDownNodeDeleted - a node was deleted.
	ScaleDownNodeDeleted
	// ScaleDownNodeDeleteStarted - a node deletion process was started.
	ScaleDownNodeDeleteStarted
)

// ScaleDownStatusProcessor processes the status of the cluster after a scale-down.
type ScaleDownStatusProcessor interface {
	Process(context *context.AutoscalingContext, status *ScaleDownStatus)
	CleanUp()
}

// NewDefaultScaleDownStatusProcessor creates a default instance of ScaleUpStatusProcessor.
func NewDefaultScaleDownStatusProcessor() ScaleDownStatusProcessor {
	return &NoOpScaleDownStatusProcessor{}
}

// NoOpScaleDownStatusProcessor is a ScaleDownStatusProcessor implementations useful for testing.
type NoOpScaleDownStatusProcessor struct{}

// Process processes the status of the cluster after a scale-down.
func (p *NoOpScaleDownStatusProcessor) Process(context *context.AutoscalingContext, status *ScaleDownStatus) {
}

// CleanUp cleans up the processor's internal structures.
func (p *NoOpScaleDownStatusProcessor) CleanUp() {
}
