package resourcequotas

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
)

const (
	ResourceNameCores  = "cpu"
	ResourceNameMemory = "memory"
	ResourceNameNodes  = "nodes"
)

// Quota is an interface for a single quota.
type Quota interface {
	ID() string
	AppliesTo(node *corev1.Node) bool
	Limits() map[string]int64
}

// resourceList is a map of resource names to their quantities.
type resourceList map[string]int64

// Tracker tracks resource limits. A single tracker should track either max or min limits.
type Tracker struct {
	crp        customresources.CustomResourcesProcessor
	quotas     []Quota
	limitsLeft map[string]resourceList
}

// newTracker creates a new Tracker.
func newTracker(crp customresources.CustomResourcesProcessor, quotas []Quota, limitsLeft map[string]resourceList) *Tracker {
	return &Tracker{
		crp:        crp,
		quotas:     quotas,
		limitsLeft: limitsLeft,
	}
}

// ApplyDelta checks if a delta is within limits and applies it.
func (t *Tracker) ApplyDelta(
	ctx *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup, node *corev1.Node, nodeDelta int,
) (*CheckDeltaResult, error) {
	delta, err := deltaForNode(ctx, t.crp, node, nodeGroup)
	if err != nil {
		return nil, err
	}
	matchingQuotas := t.matchingQuotas(node)

	result := t.checkDelta(delta, matchingQuotas, nodeDelta)

	if result.AllowedDelta != nodeDelta {
		return result, nil
	}

	for _, rq := range matchingQuotas {
		if t.limitsLeft[rq.ID()] == nil {
			continue
		}
		for resource, resourceDelta := range delta {
			if limit, ok := t.limitsLeft[rq.ID()][resource]; ok {
				t.limitsLeft[rq.ID()][resource] = max(limit-resourceDelta*int64(result.AllowedDelta), 0)
			}
		}
	}

	return result, nil
}

// CheckDelta checks if a delta is within limits, without applying it.
func (t *Tracker) CheckDelta(
	ctx *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup, node *corev1.Node, nodeDelta int,
) (*CheckDeltaResult, error) {
	// TODO: cache deltas
	delta, err := deltaForNode(ctx, t.crp, node, nodeGroup)
	if err != nil {
		return nil, err
	}
	matchingQuotas := t.matchingQuotas(node)
	return t.checkDelta(delta, matchingQuotas, nodeDelta), nil
}

func (t *Tracker) checkDelta(delta resourceList, matchingQuotas []Quota, nodeDelta int) *CheckDeltaResult {
	result := &CheckDeltaResult{
		AllowedDelta: nodeDelta,
	}

	exceededResources := make(map[string][]string)
	for _, rq := range matchingQuotas {
		quotaLimitsLeft := t.limitsLeft[rq.ID()]
		for resource, resourceDelta := range delta {
			if resourceDelta <= 0 {
				continue
			}

			limitsLeft, ok := quotaLimitsLeft[resource]
			if !ok {
				continue
			}

			if limitsLeft < resourceDelta*int64(nodeDelta) {
				allowedNodes := limitsLeft / resourceDelta
				if allowedNodes < int64(result.AllowedDelta) {
					result.AllowedDelta = int(allowedNodes)
				}
				exceededResources[rq.ID()] = append(exceededResources[rq.ID()], resource)
			}
		}
	}
	result.ExceededResources = exceededResources
	return result
}

func (t *Tracker) matchingQuotas(node *corev1.Node) []Quota {
	var quotas []Quota
	for _, rl := range t.quotas {
		if rl.AppliesTo(node) {
			quotas = append(quotas, rl)
		}
	}
	return quotas
}

// CheckDeltaResult is a result of checking a delta.
type CheckDeltaResult struct {
	ExceededResources map[string][]string
	AllowedDelta      int
}

// Exceeded returns true if any resource limit was exceeded.
func (r *CheckDeltaResult) Exceeded() bool {
	return len(r.ExceededResources) > 0
}

// deltaForNode calculates the amount of resources that will be used from the cluster when creating a node.
func deltaForNode(ctx *context.AutoscalingContext, crp customresources.CustomResourcesProcessor, node *corev1.Node, nodeGroup cloudprovider.NodeGroup) (resourceList, error) {
	// TODO: storage?
	nodeCPU, nodeMemory := utils.GetNodeCoresAndMemory(node)
	nodeResources := resourceList{
		ResourceNameCores:  nodeCPU,
		ResourceNameMemory: nodeMemory,
		ResourceNameNodes:  1,
	}

	resourceTargets, err := crp.GetNodeResourceTargets(ctx, node, nodeGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom resources: %w", err)
	}

	for _, resourceTarget := range resourceTargets {
		nodeResources[resourceTarget.ResourceType] = resourceTarget.ResourceCount
	}

	return nodeResources, nil
}
