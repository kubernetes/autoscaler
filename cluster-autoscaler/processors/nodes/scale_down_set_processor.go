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
	"fmt"

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

// FilterUnremovableNodes filters the passed removable candidates from unremovable nodes by calling orderedProcessorList in order
func (p *CompositeScaleDownSetProcessor) FilterUnremovableNodes(ctx *context.AutoscalingContext, scaleDownCtx ScaleDownContext, candidates []simulator.NodeToBeRemoved) ([]simulator.NodeToBeRemoved, []simulator.UnremovableNode) {

	unremovableNodes := []simulator.UnremovableNode{}
	nodesToBeRemoved := []simulator.NodeToBeRemoved{}
	nodesToBeRemoved = append(nodesToBeRemoved, candidates...)

	for indx, p := range p.orderedProcessorList {
		processorRemovableNodes, processorUnremovableNodes := p.FilterUnremovableNodes(ctx, scaleDownCtx, nodesToBeRemoved)

		if !isValidCandidatesFiltering(candidates, processorRemovableNodes, processorUnremovableNodes) {
			klog.Errorf("Scale deown set composite processor failed with processor at index %d", indx)
		}

		nodesToBeRemoved = processorRemovableNodes
		unremovableNodes = append(unremovableNodes, processorUnremovableNodes...)
	}
	return nodesToBeRemoved, unremovableNodes
}

// isValidCandidatesFeltering checks that exactly all candidates are splitted into removableNodes and unremovableNodes and loggign additional/missing nodes
func isValidCandidatesFiltering(candidates []simulator.NodeToBeRemoved, removableNodes []simulator.NodeToBeRemoved, unremovableNodes []simulator.UnremovableNode) bool {
	isValidCandidatesSplit := true

	candidatesDictionary := map[string]bool{}
	for _, c := range candidates {
		if candidatesDictionary[c.Node.Name] {
			isValidCandidatesSplit = false
			klog.Errorf("Scale deown set composite processor error: node %s is repeated in removable candidates nodes", c.Node.Name)
		}
		candidatesDictionary[c.Node.Name] = true
	}

	for _, c := range removableNodes {
		if !candidatesDictionary[c.Node.Name] {
			isValidCandidatesSplit = false
			klog.Errorf("Scale deown set composite processor error: removable node %s not found in original removable candidates nodes", c.Node.Name)
		} else {
			delete(candidatesDictionary, c.Node.Name)
		}
	}
	for _, c := range unremovableNodes {
		if !candidatesDictionary[c.Node.Name] {
			isValidCandidatesSplit = false
			klog.Errorf("Scale deown set composite processor error: unremovable node %s not found in original removable candidates nodes", c.Node.Name)
		} else {
			delete(candidatesDictionary, c.Node.Name)
		}
	}

	if len(candidatesDictionary) != 0 {
		isValidCandidatesSplit = false
		remainingNodes := ""
		for k := range candidatesDictionary {
			remainingNodes += fmt.Sprintf(" %s ", k)
		}
		klog.Errorf("Scale deown set composite processor error: removable candidate nodes %s are not in the returned removable/unremovable lists", remainingNodes)
	}
	return isValidCandidatesSplit
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

// FilterUnremovableNodes selects first maxCount candidates for deletion and marks the rest as unremovable
func (p *MaxNodesProcessor) FilterUnremovableNodes(ctx *context.AutoscalingContext, scaleDownCtx ScaleDownContext, candidates []simulator.NodeToBeRemoved) ([]simulator.NodeToBeRemoved, []simulator.UnremovableNode) {

	nodesToBeRemoved := []simulator.NodeToBeRemoved{}
	unremovableNodes := []simulator.UnremovableNode{}

	for idx, node := range candidates {
		if idx < scaleDownCtx.MaxNodeCountToRemove {
			nodesToBeRemoved = append(nodesToBeRemoved, node)
		} else {
			unremovableNodes = append(unremovableNodes, simulator.UnremovableNode{Node: node.Node, Reason: simulator.NodeGroupMaxDeletionCountReached})
		}
	}

	return nodesToBeRemoved, unremovableNodes
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

// FilterUnremovableNodes marks all candidate nodes as unremovable if ZeroOrMaxNodeScaling is enabled and number of nodes to remove are not equal to target size
func (p *AtomicResizeFilteringProcessor) FilterUnremovableNodes(ctx *context.AutoscalingContext, scaleDownCtx ScaleDownContext, candidates []simulator.NodeToBeRemoved) ([]simulator.NodeToBeRemoved, []simulator.UnremovableNode) {

	nodesToBeRemoved := []simulator.NodeToBeRemoved{}
	unremovableNodes := []simulator.UnremovableNode{}

	atomicQuota := klogx.NodesLoggingQuota()
	standardQuota := klogx.NodesLoggingQuota()
	nodesByGroup := map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{}
	for _, node := range candidates {
		nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(node.Node)
		if err != nil {
			klog.Errorf("Node %v will not scale down, failed to get node info: %s", node.Node.Name, err)
			unremovableNodes = append(unremovableNodes, simulator.UnremovableNode{Node: node.Node, Reason: simulator.UnexpectedError})
			continue
		}
		autoscalingOptions, err := nodeGroup.GetOptions(ctx.NodeGroupDefaults)
		if err != nil && err != cloudprovider.ErrNotImplemented {
			klog.Errorf("Failed to get autoscaling options for node group %s: %v", nodeGroup.Id(), err)
			unremovableNodes = append(unremovableNodes, simulator.UnremovableNode{Node: node.Node, Reason: simulator.UnexpectedError})
			continue
		}
		if autoscalingOptions != nil && autoscalingOptions.ZeroOrMaxNodeScaling {
			klogx.V(2).UpTo(atomicQuota).Infof("Considering node %s for atomic scale down", node.Node.Name)
			nodesByGroup[nodeGroup] = append(nodesByGroup[nodeGroup], node)
		} else {
			klogx.V(2).UpTo(standardQuota).Infof("Considering node %s for standard scale down", node.Node.Name)
			nodesToBeRemoved = append(nodesToBeRemoved, node)
		}
	}
	klogx.V(2).Over(atomicQuota).Infof("Considering %d other nodes for atomic scale down", -atomicQuota.Left())
	klogx.V(2).Over(standardQuota).Infof("Considering %d other nodes for standard scale down", -atomicQuota.Left())
	for nodeGroup, nodes := range nodesByGroup {
		ngSize, err := nodeGroup.TargetSize()
		if err != nil {
			klog.Errorf("Nodes from group %s will not scale down, failed to get target size: %s", nodeGroup.Id(), err)
			for _, node := range nodes {
				unremovableNodes = append(unremovableNodes, simulator.UnremovableNode{Node: node.Node, Reason: simulator.UnexpectedError})
			}
			continue
		}
		if ngSize == len(nodes) {
			klog.V(2).Infof("Scheduling atomic scale down for all %v nodes from node group %s", len(nodes), nodeGroup.Id())
			nodesToBeRemoved = append(nodesToBeRemoved, nodes...)
		} else {
			klog.V(2).Infof("Skipping scale down for %v nodes from node group %s, all %v nodes have to be scaled down atomically", len(nodes), nodeGroup.Id(), ngSize)
			for _, node := range nodes {
				unremovableNodes = append(unremovableNodes, simulator.UnremovableNode{Node: node.Node, Reason: simulator.AtomicScaleDownFailed})
			}
		}
	}
	return nodesToBeRemoved, unremovableNodes
}

// CleanUp is called at CA termination
func (p *AtomicResizeFilteringProcessor) CleanUp() {
}

// NewAtomicResizeFilteringProcessor returns a new AtomicResizeFilteringProcessor
func NewAtomicResizeFilteringProcessor() *AtomicResizeFilteringProcessor {
	return &AtomicResizeFilteringProcessor{}
}
