package resourcequotas

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
)

// NodeFilter is an interface for filtering nodes which should be included in usage calculations.
type NodeFilter interface {
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

func (u *usageCalculator) calculateUsages(ctx *context.AutoscalingContext, nodes []*corev1.Node, limiters []Quota) (map[string]resourceList, error) {
	usages := make(map[string]resourceList)
	for _, rl := range limiters {
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
		for _, rl := range limiters {
			if rl.AppliesTo(node) {
				for resourceType, resourceCount := range delta {
					usages[rl.ID()][resourceType] += resourceCount
				}
			}
		}
	}
	return usages, nil
}
