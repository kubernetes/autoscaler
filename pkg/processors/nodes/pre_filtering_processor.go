/*
Copyright 2019 The Kubernetes Authors.

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

	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"

	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
)

// PreFilteringScaleDownNodeProcessor filters out scale down candidates from nodegroup with
// size <= minimum number of nodes for that nodegroup and filters out node from non-autoscaled
// nodegroups
type PreFilteringScaleDownNodeProcessor struct {
}

// GetPodDestinationCandidates returns nodes that potentially could act as destinations for pods
// that would become unscheduled after a scale down.
func (n *PreFilteringScaleDownNodeProcessor) GetPodDestinationCandidates(autoscalingCtx *ca_context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	return nodes, nil
}

// GetScaleDownCandidates returns nodes that potentially could be scaled down and
func (n *PreFilteringScaleDownNodeProcessor) GetScaleDownCandidates(ctx context.Context, autoscalingCtx *ca_context.AutoscalingContext,
	nodes []*apiv1.Node) ([]*apiv1.Node, errors.AutoscalerError) {
	logger := klog.FromContext(ctx)
	result := make([]*apiv1.Node, 0, len(nodes))

	nodeGroupSize := utils.GetNodeGroupSizeMap(ctx, autoscalingCtx.CloudProvider)

	for _, node := range nodes {
		nodeGroup, err := autoscalingCtx.CloudProvider.NodeGroupForNode(ctx, node)
		if err != nil {
			logger.Info("Error while checking node group for node", "node", klog.KObj(node), "err", err)
			continue
		}
		if nodeGroup == nil {
			logger.V(5).Info("Node should not be processed by cluster autoscaler (no node group config)", "node", klog.KObj(node))
			continue
		}
		size, found := nodeGroupSize[nodeGroup.Id()]
		if !found {
			logger.Error(nil, "Error while checking node group size: group size not found", "nodeGroupId", nodeGroup.Id())
			continue
		}
		minSize := nodeGroup.MinSize(ctx)
		if size <= minSize {
			logger.V(1).Info("Skipping node - node group min size reached", "node", klog.KObj(node), "currentSize", size, "minSize", minSize)
			continue
		}
		result = append(result, node)
	}
	return result, nil
}

// CleanUp is called at CA termination.
func (n *PreFilteringScaleDownNodeProcessor) CleanUp() {
}

// NewPreFilteringScaleDownNodeProcessor returns a new PreFilteringScaleDownNodeProcessor.
func NewPreFilteringScaleDownNodeProcessor() *PreFilteringScaleDownNodeProcessor {
	return &PreFilteringScaleDownNodeProcessor{}
}
