/*
Copyright 2025 The Kubernetes Authors.

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

package customresources

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// DefaultCustomResourcesProcessor handles multiple custom resource processors and
// executes them in order.
type DefaultCustomResourcesProcessor struct {
	customResourcesProcessors []CustomResourcesProcessor
}

// NewDefaultCustomResourcesProcessor returns an instance of DefaultCustomResourcesProcessor.
func NewDefaultCustomResourcesProcessor(draEnabled bool) CustomResourcesProcessor {
	customProcessors := []CustomResourcesProcessor{&GpuCustomResourcesProcessor{}}
	if draEnabled {
		customProcessors = append(customProcessors, &DraCustomResourcesProcessor{})
	}
	return &DefaultCustomResourcesProcessor{customProcessors}
}

// FilterOutNodesWithUnreadyResources calls the corresponding method for internal custom resources processors in order.
func (p *DefaultCustomResourcesProcessor) FilterOutNodesWithUnreadyResources(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node, draSnapshot *drasnapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := allNodes
	newReadyNodes := readyNodes
	for _, processor := range p.customResourcesProcessors {
		newAllNodes, newReadyNodes = processor.FilterOutNodesWithUnreadyResources(context, newAllNodes, newReadyNodes, draSnapshot)
	}
	return newAllNodes, newReadyNodes
}

// GetNodeResourceTargets calls the corresponding method for internal custom resources processors in order.
func (p *DefaultCustomResourcesProcessor) GetNodeResourceTargets(context *context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError) {
	customResourcesTargets := []CustomResourceTarget{}
	for _, processor := range p.customResourcesProcessors {
		targets, err := processor.GetNodeResourceTargets(context, node, nodeGroup)
		if err != nil {
			return nil, err
		}
		customResourcesTargets = append(customResourcesTargets, targets...)
	}
	return customResourcesTargets, nil
}

// CleanUp cleans up all internal custom resources processors.
func (p *DefaultCustomResourcesProcessor) CleanUp() {
	for _, processor := range p.customResourcesProcessors {
		processor.CleanUp()
	}
}
