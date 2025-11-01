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

package resourcequotas

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
)

// NodeFilter customizes what nodes should be included in usage calculations.
type NodeFilter interface {
	// ExcludeFromTracking returns true if the node should be excluded from usage calculations.
	ExcludeFromTracking(node *corev1.Node) bool
}

type usageCalculator struct {
	crp        customresources.CustomResourcesProcessor
	nodeFilter NodeFilter
}

func newUsageCalculator(crp customresources.CustomResourcesProcessor, nodeFilter NodeFilter) *usageCalculator {
	return &usageCalculator{
		crp:        crp,
		nodeFilter: nodeFilter,
	}
}

// calculateUsages calculates resources used by nodes for every quota.
// Returns a map with quota ID as a key and resources used in the corresponding quota as a value.
func (u *usageCalculator) calculateUsages(ctx *context.AutoscalingContext, nodes []*corev1.Node, quotas []Quota) (map[string]resourceList, error) {
	usages := make(map[string]resourceList)
	for _, rl := range quotas {
		usages[rl.ID()] = make(resourceList)
	}

	for _, node := range nodes {
		if u.nodeFilter != nil && u.nodeFilter.ExcludeFromTracking(node) {
			continue
		}

		ng, err := ctx.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			return nil, fmt.Errorf("failed to get node group for node %q: %w", node.Name, err)
		}
		delta, err := deltaForNode(ctx, u.crp, node, ng)
		if err != nil {
			return nil, err
		}
		for _, rq := range quotas {
			if rq.AppliesTo(node) {
				for resourceType, resourceCount := range delta {
					usages[rq.ID()][resourceType] += resourceCount
				}
			}
		}
	}
	return usages, nil
}
