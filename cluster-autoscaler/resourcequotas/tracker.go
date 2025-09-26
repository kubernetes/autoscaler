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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
)

const (
	// ResourceNodes is a resource name for number of nodes.
	ResourceNodes = "nodes"
)

// Quota is an interface for a single quota.
type Quota interface {
	ID() string
	// AppliesTo returns true if the quota applies to the given node.
	AppliesTo(node *corev1.Node) bool
	// Limits returns the resource limits defined by the quota.
	Limits() map[string]int64
}

// resourceList is a map of resource names to their quantities.
type resourceList map[string]int64

// Tracker tracks resource quotas.
type Tracker struct {
	crp           customresources.CustomResourcesProcessor
	quotaStatuses []*quotaStatus
}

type quotaStatus struct {
	quota      Quota
	limitsLeft resourceList
}

// newTracker creates a new Tracker.
func newTracker(crp customresources.CustomResourcesProcessor, quotaStatuses []*quotaStatus) *Tracker {
	return &Tracker{
		crp:           crp,
		quotaStatuses: quotaStatuses,
	}
}

// ApplyDelta checks if a delta is within limits and applies it. Delta is applied only if it can be applied entirely.
// See CheckDelta documentation for more information.
func (t *Tracker) ApplyDelta(
	autoscalingCtx *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup, node *corev1.Node, nodeDelta int,
) (*CheckDeltaResult, error) {
	delta, err := nodeResources(autoscalingCtx, t.crp, node, nodeGroup)
	if err != nil {
		return nil, err
	}
	matchingQuotas := t.matchingQuotaStatuses(node)

	result := t.checkDelta(delta, matchingQuotas, nodeDelta)

	if result.AllowedDelta != nodeDelta {
		return result, nil
	}

	for _, qs := range matchingQuotas {
		for resource, resourceDelta := range delta {
			if limit, ok := qs.limitsLeft[resource]; ok {
				qs.limitsLeft[resource] = max(limit-resourceDelta*int64(result.AllowedDelta), 0)
			}
		}
	}

	return result, nil
}

// CheckDelta checks if a delta is within limits and returns a struct containing information
// about exceeded quotas, if any, and how many nodes could be added without violating the quotas,
// which is less than or equal to nodeDelta.
//
// nodeDelta is the number of nodes that we try to add to the cluster. Resources used by each node
// are taken from the template node passed via the node parameter. nodeGroup is required to fetch
// the custom resources, such as GPU.
func (t *Tracker) CheckDelta(
	autoscalingCtx *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup, node *corev1.Node, nodeDelta int,
) (*CheckDeltaResult, error) {
	// TODO: cache deltas
	delta, err := nodeResources(autoscalingCtx, t.crp, node, nodeGroup)
	if err != nil {
		return nil, err
	}
	matchingQuotas := t.matchingQuotaStatuses(node)
	return t.checkDelta(delta, matchingQuotas, nodeDelta), nil
}

func (t *Tracker) checkDelta(delta resourceList, matchingQuotas []*quotaStatus, nodeDelta int) *CheckDeltaResult {
	result := &CheckDeltaResult{
		AllowedDelta: nodeDelta,
	}

	for _, qs := range matchingQuotas {
		var exceededResources []string
		for resource, resourceDelta := range delta {
			if resourceDelta <= 0 {
				continue
			}

			limitsLeft, ok := qs.limitsLeft[resource]
			if !ok {
				continue
			}

			// node has resourceDelta units of resource, and we try to add nodeDelta nodes.
			// Therefore, we need to check if resourceDelta*nodeDelta is within the limitsLeft.
			if resourcesNeeded := resourceDelta * int64(nodeDelta); limitsLeft < resourcesNeeded {
				allowedNodes := limitsLeft / resourceDelta
				if allowedNodes < int64(result.AllowedDelta) {
					result.AllowedDelta = int(allowedNodes)
				}
				exceededResources = append(exceededResources, resource)
			}
		}
		if len(exceededResources) > 0 {
			result.ExceededQuotas = append(result.ExceededQuotas, ExceededQuota{
				ID: qs.quota.ID(), ExceededResources: exceededResources,
			})
		}
	}

	return result
}

func (t *Tracker) matchingQuotaStatuses(node *corev1.Node) []*quotaStatus {
	var statuses []*quotaStatus
	for _, ls := range t.quotaStatuses {
		if ls.quota.AppliesTo(node) {
			statuses = append(statuses, ls)
		}
	}
	return statuses
}

// CheckDeltaResult is a result of checking a delta.
type CheckDeltaResult struct {
	// ExceededQuotas contains information about quotas that were exceeded.
	ExceededQuotas []ExceededQuota
	// AllowedDelta specifies how many nodes could be added without violating the quotas.
	AllowedDelta int
}

// Exceeded returns true if any resource limit was exceeded.
func (r *CheckDeltaResult) Exceeded() bool {
	return len(r.ExceededQuotas) > 0
}

// ExceededQuota contains information about quota that was exceeded.
type ExceededQuota struct {
	ID                string
	ExceededResources []string
}

// nodeResources calculates the amount of resources that will be used from the cluster when creating a node.
func nodeResources(autoscalingCtx *context.AutoscalingContext, crp customresources.CustomResourcesProcessor, node *corev1.Node, nodeGroup cloudprovider.NodeGroup) (resourceList, error) {
	// TODO: storage?
	nodeCPU, nodeMemory := utils.GetNodeCoresAndMemory(node)
	nodeResources := resourceList{
		string(corev1.ResourceCPU):    nodeCPU,
		string(corev1.ResourceMemory): nodeMemory,
		ResourceNodes:                 1,
	}

	resourceTargets, err := crp.GetNodeResourceTargets(autoscalingCtx, node, nodeGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom resources: %w", err)
	}

	for _, resourceTarget := range resourceTargets {
		nodeResources[resourceTarget.ResourceType] = resourceTarget.ResourceCount
	}

	return nodeResources, nil
}
