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
	klog "k8s.io/klog/v2"
)

// PostFilteringScaleDownNodeProcessor selects first maxCount nodes (if possible) to be removed
type PostFilteringScaleDownNodeProcessor struct {
}

// GetNodesToRemove selects up to maxCount nodes for deletion, by selecting a first maxCount candidates
func (p *PostFilteringScaleDownNodeProcessor) GetNodesToRemove(ctx *context.AutoscalingContext, candidates []simulator.NodeToBeRemoved, maxCount int) []simulator.NodeToBeRemoved {
	end := len(candidates)
	if len(candidates) > maxCount {
		end = maxCount
	}
	return p.filterOutIncompleteAtomicNodeGroups(ctx, candidates[:end])
}

// CleanUp is called at CA termination
func (p *PostFilteringScaleDownNodeProcessor) CleanUp() {
}

// NewPostFilteringScaleDownNodeProcessor returns a new PostFilteringScaleDownNodeProcessor
func NewPostFilteringScaleDownNodeProcessor() *PostFilteringScaleDownNodeProcessor {
	return &PostFilteringScaleDownNodeProcessor{}
}

func (p *PostFilteringScaleDownNodeProcessor) filterOutIncompleteAtomicNodeGroups(ctx *context.AutoscalingContext, nodes []simulator.NodeToBeRemoved) []simulator.NodeToBeRemoved {
	nodesByGroup := map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{}
	result := []simulator.NodeToBeRemoved{}
	for _, node := range nodes {
		nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(node.Node)
		if err != nil {
			klog.Errorf("Node %v will not scale down, failed to get node info: %s", node.Node.Name, err)
			continue
		}
		autoscalingOptions, err := nodeGroup.GetOptions(ctx.NodeGroupDefaults)
		if err != nil {
			klog.Errorf("Failed to get autoscaling options for node group %s: %v", nodeGroup.Id(), err)
			continue
		}
		if autoscalingOptions != nil && autoscalingOptions.ZeroOrMaxNodeScaling {
			klog.V(2).Infof("Considering node %s for atomic scale down", node.Node.Name)
			nodesByGroup[nodeGroup] = append(nodesByGroup[nodeGroup], node)
		} else {
			klog.V(2).Infof("Considering node %s for standard scale down", node.Node.Name)
			result = append(result, node)
		}
	}
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
