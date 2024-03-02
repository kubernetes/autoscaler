/*
Copyright 2021 The Kubernetes Authors.

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

package nodes

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	klog "k8s.io/klog/v2"
)

// CompositeScaleDownSetProcessor is a ScaleDownSetProcessor composed of multiple sub-processors passed as an argument.
type CompositeScaleDownSetProcessor struct {
	orderedProcessorList []ScaleDownSetProcessor
}

// NewCompositeScaleDownSetProcessor creates new CompositeScaleDownSetProcessor. The order on a list defines order in witch
// sub-processors are invoked.
func NewCompositeScaleDownSetProcessor(orderedProcessorList []ScaleDownSetProcessor) *CompositeScaleDownSetProcessor {
	return &CompositeScaleDownSetProcessor{
		orderedProcessorList: orderedProcessorList,
	}
}

// GetNodesToRemove selects nodes to remove.
func (p *CompositeScaleDownSetProcessor) GetNodesToRemove(ctx *context.AutoscalingContext, candidates []simulator.NodeToBeRemoved, maxCount int) []simulator.NodeToBeRemoved {
	for _, p := range p.orderedProcessorList {
		candidates = p.GetNodesToRemove(ctx, candidates, maxCount)
	}
	return candidates
}

// CleanUp is called at CA termination
func (p *CompositeScaleDownSetProcessor) CleanUp() {
	for _, p := range p.orderedProcessorList {
		p.CleanUp()
	}
}

// MaxNodesProcessor selects first maxCount nodes (if possible) to be removed
type MaxNodesProcessor struct {
}

// GetNodesToRemove selects up to maxCount nodes for deletion, by selecting a first maxCount candidates
func (p *MaxNodesProcessor) GetNodesToRemove(ctx *context.AutoscalingContext, candidates []simulator.NodeToBeRemoved, maxCount int) []simulator.NodeToBeRemoved {
	end := len(candidates)
	if len(candidates) > maxCount {
		end = maxCount
	}
	return candidates[:end]
}

// CleanUp is called at CA termination
func (p *MaxNodesProcessor) CleanUp() {
}

// NewMaxNodesProcessor returns a new MaxNodesProcessor
func NewMaxNodesProcessor() *MaxNodesProcessor {
	return &MaxNodesProcessor{}
}

// AtomicResizeFilteringProcessor removes node groups which should be scaled down as one unit
// if only part of these nodes were scheduled for scale down.
// NOTE! When chaining with other processors, AtomicResizeFilteringProcessors should be always used last.
// Otherwise, it's possible that another processor will break the property that this processor aims to restore:
// no partial scale-downs for node groups that should be resized atomically.
type AtomicResizeFilteringProcessor struct {
}

// GetNodesToRemove selects up to maxCount nodes for deletion, by selecting a first maxCount candidates
func (p *AtomicResizeFilteringProcessor) GetNodesToRemove(ctx *context.AutoscalingContext, candidates []simulator.NodeToBeRemoved, maxCount int) []simulator.NodeToBeRemoved {
	atomicQuota := klogx.NodesLoggingQuota()
	standardQuota := klogx.NodesLoggingQuota()
	nodesByGroup := map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{}
	result := []simulator.NodeToBeRemoved{}
	for _, node := range candidates {
		nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(node.Node)
		if err != nil {
			klog.Errorf("Node %v will not scale down, failed to get node info: %s", node.Node.Name, err)
			continue
		}
		autoscalingOptions, err := nodeGroup.GetOptions(ctx.NodeGroupDefaults)
		if err != nil && err != cloudprovider.ErrNotImplemented {
			klog.Errorf("Failed to get autoscaling options for node group %s: %v", nodeGroup.Id(), err)
			continue
		}
		if autoscalingOptions != nil && autoscalingOptions.ZeroOrMaxNodeScaling {
			klogx.V(2).UpTo(atomicQuota).Infof("Considering node %s for atomic scale down", node.Node.Name)
			nodesByGroup[nodeGroup] = append(nodesByGroup[nodeGroup], node)
		} else {
			klogx.V(2).UpTo(standardQuota).Infof("Considering node %s for standard scale down", node.Node.Name)
			result = append(result, node)
		}
	}
	klogx.V(2).Over(atomicQuota).Infof("Considering %d other nodes for atomic scale down", -atomicQuota.Left())
	klogx.V(2).Over(standardQuota).Infof("Considering %d other nodes for standard scale down", -atomicQuota.Left())
	for nodeGroup, nodes := range nodesByGroup {
		ngSize, err := nodeGroup.TargetSize()
		if err != nil {
			klog.Errorf("Nodes from group %s will not scale down, failed to get target size: %s", nodeGroup.Id(), err)
			continue
		}
		if ngSize == len(nodes) {
			klog.V(2).Infof("Scheduling atomic scale down for all %v nodes from node group %s", len(nodes), nodeGroup.Id())
			result = append(result, nodes...)
		} else {
			klog.V(2).Infof("Skipping scale down for %v nodes from node group %s, all %v nodes have to be scaled down atomically", len(nodes), nodeGroup.Id(), ngSize)
		}
	}
	return result
}

// CleanUp is called at CA termination
func (p *AtomicResizeFilteringProcessor) CleanUp() {
}

// NewAtomicResizeFilteringProcessor returns a new AtomicResizeFilteringProcessor
func NewAtomicResizeFilteringProcessor() *AtomicResizeFilteringProcessor {
	return &AtomicResizeFilteringProcessor{}
}
