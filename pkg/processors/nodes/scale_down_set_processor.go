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
	"context"
	"slices"
	"strconv"

	v1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/clustersnapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/annotations"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/klogx"
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
func (p *CompositeScaleDownSetProcessor) FilterUnremovableNodes(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, scaleDownCtx *ScaleDownContext, candidates []simulator.NodeToBeRemoved) ([]simulator.NodeToBeRemoved, []simulator.UnremovableNode) {
	logger := klog.FromContext(ctx)
	unremovableNodes := []simulator.UnremovableNode{}
	nodesToBeRemoved := []simulator.NodeToBeRemoved{}
	nodesToBeRemoved = append(nodesToBeRemoved, candidates...)

	for indx, p := range p.orderedProcessorList {
		processorRemovableNodes, processorUnremovableNodes := p.FilterUnremovableNodes(ctx, autoscalingCtx, scaleDownCtx, nodesToBeRemoved)

		if len(processorRemovableNodes)+len(processorUnremovableNodes) != len(candidates) {
			logger.Error(nil, "Scale down set composite processor failed with processor at index: removable nodes + unremovable nodes != candidates nodes", "index", indx, "removableNodesCount", len(processorRemovableNodes), "unremovableNodesCount", len(processorUnremovableNodes), "candidatesCount", len(candidates))
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
func (p *AtomicResizeFilteringProcessor) FilterUnremovableNodes(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext, scaleDownCtx *ScaleDownContext, candidates []simulator.NodeToBeRemoved) ([]simulator.NodeToBeRemoved, []simulator.UnremovableNode) {
	logger := klog.FromContext(ctx)
	nodesToBeRemoved := []simulator.NodeToBeRemoved{}
	unremovableNodes := []simulator.UnremovableNode{}

	atomicQuota := klogx.NodesLoggingQuota()
	standardQuota := klogx.NodesLoggingQuota()
	nodesByGroup := map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{}
	allNodes, err := allNodes(autoscalingCtx.ClusterSnapshot)
	if err != nil {
		logger.Error(err, "Failed to read all nodes from the cluster snapshot for filtering unremovable nodes")
	}

	for _, node := range candidates {
		nodeGroup, err := autoscalingCtx.CloudProvider.NodeGroupForNode(ctx, node.Node)
		if err != nil {
			logger.Error(err, "Node will not scale down, failed to get node info", "node", klog.KObj(node.Node))
			unremovableNodes = append(unremovableNodes, simulator.UnremovableNode{Node: node.Node, Reason: simulator.UnexpectedError})
			continue
		}
		if nodeGroup == nil {
			logger.V(4).Info("Node will not scale down, has no node group config.", "node", klog.KObj(node.Node))
			unremovableNodes = append(unremovableNodes, simulator.UnremovableNode{Node: node.Node, Reason: simulator.NotAutoscaled})
			continue
		}
		autoscalingOptions, err := nodeGroup.GetOptions(ctx, autoscalingCtx.NodeGroupDefaults)
		if err != nil && err != cloudprovider.ErrNotImplemented {
			logger.Error(err, "Failed to get autoscaling options for node group", "nodeGroupId", nodeGroup.Id())
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
		ngSize, err := nodeGroup.TargetSize(ctx)
		if err != nil {
			logger.Error(err, "Nodes from group will not scale down, failed to get target size", "nodeGroupId", nodeGroup.Id())
			for _, node := range consideredNodes {
				unremovableNodes = append(unremovableNodes, simulator.UnremovableNode{Node: node.Node, Reason: simulator.UnexpectedError})
			}
			continue
		}
		if ngSize == len(consideredNodes) {
			logger.V(2).Info("Scheduling atomic scale down for all nodes from node group", "nodesCount", len(consideredNodes), "nodeGroupId", nodeGroup.Id())
			nodesToBeRemoved = append(nodesToBeRemoved, consideredNodes...)
		} else {
			registeredNodes, err := p.getAllRegisteredNodesForNodeGroup(ctx, nodeGroup, allNodes)
			if err != nil {
				logger.Error(err, "Failed to get registered nodes for node group", "nodeGroupId", nodeGroup.Id())
				unremovableNodes = p.atomicScaleDownFailed(ctx, consideredNodes, ngSize, unremovableNodes, nodeGroup)
			} else if len(registeredNodes) == len(consideredNodes) {
				logger.V(2).Info("Scheduling atomic scale down for all registered nodes from node group", "nodesCount", len(consideredNodes), "nodeGroupId", nodeGroup.Id())
				nodesToBeRemoved = append(nodesToBeRemoved, consideredNodes...)
			} else {
				unremovableNodes = p.atomicScaleDownFailed(ctx, consideredNodes, len(registeredNodes), unremovableNodes, nodeGroup)
			}
		}
	}
	return nodesToBeRemoved, unremovableNodes
}

func (p *AtomicResizeFilteringProcessor) atomicScaleDownFailed(ctx context.Context, nodes []simulator.NodeToBeRemoved, ngSize int, unremovableNodes []simulator.UnremovableNode, nodeGroup cloudprovider.NodeGroup) []simulator.UnremovableNode {
	logger := klog.FromContext(ctx)
	logger.V(2).Info("Skipping scale down for nodes from node group, all nodes have to be scaled down atomically", "nodesCount", len(nodes), "nodeGroupId", nodeGroup.Id(), "nodeGroupSize", ngSize)
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

func (p *AtomicResizeFilteringProcessor) getAllRegisteredNodesForNodeGroup(ctx context.Context, nodeGroup cloudprovider.NodeGroup, allNodes []*v1.Node) ([]*v1.Node, error) {
	allNodesInNodeGroup, err := nodeGroup.Nodes(ctx)
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
