package resourcelimits

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

// Limiter is an interface for a single resource limit.
type Limiter interface {
	ID() string
	AppliesTo(node *corev1.Node) bool
	MaxLimits() map[string]int64
	MinLimits() map[string]int64
}

// resourceList is a map of resource names to their quantities.
type resourceList map[string]int64

// Tracker tracks resource limits. A single tracker should track either max or min limits.
type Tracker struct {
	crp        customresources.CustomResourcesProcessor
	limiters   []Limiter
	limitsLeft map[string]resourceList
}

// newTracker creates a new Tracker.
func newTracker(crp customresources.CustomResourcesProcessor, limiters []Limiter, limitsLeft map[string]resourceList) *Tracker {
	return &Tracker{
		crp:        crp,
		limiters:   limiters,
		limitsLeft: limitsLeft,
	}
}

type TrackerOptions struct {
	CRP        customresources.CustomResourcesProcessor
	Providers  []Provider
	NodeFilter NodeFilter
}

// ApplyDelta checks if a delta is within limits and applies it.
func (t *Tracker) ApplyDelta(
	ctx *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup, node *corev1.Node, nodeDelta int,
) (*CheckDeltaResult, error) {
	delta, err := deltaForNode(ctx, t.crp, node, nodeGroup)
	if err != nil {
		return nil, err
	}
	matchingLimiters := t.matchingLimiters(node)

	result := t.checkDelta(delta, matchingLimiters, nodeDelta)

	if result.AllowedDelta != nodeDelta {
		return result, nil
	}

	for _, rl := range matchingLimiters {
		if t.limitsLeft[rl.ID()] == nil {
			continue
		}
		for resource, resourceDelta := range delta {
			if limit, ok := t.limitsLeft[rl.ID()][resource]; ok {
				t.limitsLeft[rl.ID()][resource] = max(limit-resourceDelta*int64(result.AllowedDelta), 0)
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
	matchingLimiters := t.matchingLimiters(node)
	return t.checkDelta(delta, matchingLimiters, nodeDelta), nil
}

func (t *Tracker) checkDelta(delta resourceList, matchingLimiters []Limiter, nodeDelta int) *CheckDeltaResult {
	result := &CheckDeltaResult{
		AllowedDelta: nodeDelta,
	}

	exceededResources := make(map[string][]string)
	for _, rl := range matchingLimiters {
		limiterLimitsLeft := t.limitsLeft[rl.ID()]
		for resource, resourceDelta := range delta {
			if resourceDelta <= 0 {
				continue
			}

			limitsLeft, ok := limiterLimitsLeft[resource]
			if !ok {
				continue
			}

			if limitsLeft < resourceDelta*int64(nodeDelta) {
				allowedNodes := limitsLeft / resourceDelta
				if allowedNodes < int64(result.AllowedDelta) {
					result.AllowedDelta = int(allowedNodes)
				}
				exceededResources[rl.ID()] = append(exceededResources[rl.ID()], resource)
			}
		}
	}
	result.ExceededResources = exceededResources
	return result
}

func (t *Tracker) matchingLimiters(node *corev1.Node) []Limiter {
	var matchingLimiters []Limiter
	for _, rl := range t.limiters {
		if rl.AppliesTo(node) {
			matchingLimiters = append(matchingLimiters, rl)
		}
	}
	return matchingLimiters
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
