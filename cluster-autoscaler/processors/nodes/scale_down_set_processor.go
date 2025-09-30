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
	"slices"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/annotations"
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
func (p *CompositeScaleDownSetProcessor) FilterUnremovableNodes(ctx *context.AutoscalingContext, scaleDownCtx *ScaleDownContext, candidates []simulator.NodeToBeRemoved) ([]simulator.NodeToBeRemoved, []simulator.UnremovableNode) {
	unremovableNodes := []simulator.UnremovableNode{}
	nodesToBeRemoved := []simulator.NodeToBeRemoved{}
	nodesToBeRemoved = append(nodesToBeRemoved, candidates...)

	for indx, p := range p.orderedProcessorList {
		processorRemovableNodes, processorUnremovableNodes := p.FilterUnremovableNodes(ctx, scaleDownCtx, nodesToBeRemoved)

		if len(processorRemovableNodes)+len(processorUnremovableNodes) != len(candidates) {
			klog.Errorf("Scale down set composite processor failed with processor at index %d: removable nodes (%d) + unremovable nodes (%d) != candidates nodes (%d)",
				indx, len(processorRemovableNodes), len(processorUnremovableNodes), len(candidates))
		}

		nodesToBeRemoved = processorRemovableNodes
		unremovableNodes = append(unremovableNodes, processorUnremovableNodes...)
	}
	return nodesToBeRemoved, unremovableNodes
}

// CleanUp is called at CA termination
func (p *CompositeScaleDownSetProcessor) CleanUp() {
	for _, p := range p.orderedProcessorList {
		p.CleanUp()
	}
}

// AtomicResizeFilteringProcessor removes node groups which should be scaled down as one unit
// if only part of these nodes were scheduled for scale down.
// NOTE! When chaining with other processors, AtomicResizeFilteringProcessors should be always used last.
// Otherwise, it's possible that another processor will break the property that this processor aims to restore:
// no partial scale-downs for node groups that should be resized atomically.
type AtomicResizeFilteringProcessor struct {
}

// FilterUnremovableNodes marks all candidate nodes as unremovable if ZeroOrMaxNodeScaling is enabled and number of nodes to remove are not equal to target or current size
func (p *AtomicResizeFilteringProcessor) FilterUnremovableNodes(ctx *context.AutoscalingContext, scaleDownCtx *ScaleDownContext, candidates []simulator.NodeToBeRemoved) ([]simulator.NodeToBeRemoved, []simulator.UnremovableNode) {
	nodesToBeRemoved := []simulator.NodeToBeRemoved{}
	unremovableNodes := []simulator.UnremovableNode{}

	atomicQuota := klogx.NodesLoggingQuota()
	standardQuota := klogx.NodesLoggingQuota()
	nodesByGroup := map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{}
	allNodes, err := allNodes(ctx.ClusterSnapshot)
	if err != nil {
		klog.Errorf("failed to read all nodes from the cluster snapshot for filtering unremovable nodes, err: %s", err)
	}

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
	for nodeGroup, consideredNodes := range nodesByGroup {
		ngSize, err := nodeGroup.TargetSize()
		if err != nil {
			klog.Errorf("Nodes from group %s will not scale down, failed to get target size: %s", nodeGroup.Id(), err)
			for _, node := range consideredNodes {
				unremovableNodes = append(unremovableNodes, simulator.UnremovableNode{Node: node.Node, Reason: simulator.UnexpectedError})
			}
			continue
		}
		if ngSize == len(consideredNodes) {
			klog.V(2).Infof("Scheduling atomic scale down for all %v nodes from node group %s", len(consideredNodes), nodeGroup.Id())
			nodesToBeRemoved = append(nodesToBeRemoved, consideredNodes...)
		} else {
			registeredNodes, err := p.getAllRegisteredNodesForNodeGroup(nodeGroup, allNodes)
			if err != nil {
				klog.Errorf("Failed to get registered nodes for group %s: %v", nodeGroup.Id(), err)
				unremovableNodes = p.atomicScaleDownFailed(consideredNodes, ngSize, unremovableNodes, nodeGroup)
			} else if len(registeredNodes) == len(consideredNodes) {
				klog.V(2).Infof("Scheduling atomic scale down for all %v registered nodes from node group %s", len(consideredNodes), nodeGroup.Id())
				nodesToBeRemoved = append(nodesToBeRemoved, consideredNodes...)
			} else {
				unremovableNodes = p.atomicScaleDownFailed(consideredNodes, len(registeredNodes), unremovableNodes, nodeGroup)
			}
		}
	}
	return nodesToBeRemoved, unremovableNodes
}

func (p *AtomicResizeFilteringProcessor) atomicScaleDownFailed(nodes []simulator.NodeToBeRemoved, ngSize int, unremovableNodes []simulator.UnremovableNode, nodeGroup cloudprovider.NodeGroup) []simulator.UnremovableNode {
	klog.V(2).Infof("Skipping scale down for %v nodes from node group %s, all %v nodes have to be scaled down atomically", len(nodes), nodeGroup.Id(), ngSize)
	unremovableNodes = slices.Grow(unremovableNodes, len(nodes))
	for _, node := range nodes {
		unremovableNodes = append(unremovableNodes, simulator.UnremovableNode{Node: node.Node, Reason: simulator.AtomicScaleDownFailed})
	}
	return unremovableNodes
}

func allNodes(s clustersnapshot.ClusterSnapshot) ([]*v1.Node, error) {
	nodeInfos, err := s.ListNodeInfos()
	if err != nil {
		// This should never happen, List() returns err only because scheduler interface requires it.
		return nil, err
	}
	nodes := make([]*v1.Node, len(nodeInfos))
	for i, ni := range nodeInfos {
		nodes[i] = ni.Node()
	}
	return nodes, nil
}

func (p *AtomicResizeFilteringProcessor) getAllRegisteredNodesForNodeGroup(nodeGroup cloudprovider.NodeGroup, allNodes []*v1.Node) ([]*v1.Node, error) {
	allNodesInNodeGroup, err := nodeGroup.Nodes()
	if err != nil {
		return nil, err
	}
	nodeByNodeName := map[string]cloudprovider.Instance{}
	for _, node := range allNodesInNodeGroup {
		nodeByNodeName[node.Id] = node
	}
	var registeredNodesForNodeGroup []*v1.Node
	for _, node := range allNodes {
		if val, ok := node.Annotations[annotations.NodeUpcomingAnnotation]; ok {
			if res, ok := strconv.ParseBool(val); ok == nil && res {
				continue
			}
		}
		if _, ok := nodeByNodeName[node.Spec.ProviderID]; ok {
			registeredNodesForNodeGroup = append(registeredNodesForNodeGroup, node)
		}
	}
	return registeredNodesForNodeGroup, nil
}

// CleanUp is called at CA termination
func (p *AtomicResizeFilteringProcessor) CleanUp() {
}

// NewAtomicResizeFilteringProcessor returns a new AtomicResizeFilteringProcessor
func NewAtomicResizeFilteringProcessor() *AtomicResizeFilteringProcessor {
	return &AtomicResizeFilteringProcessor{}
}
