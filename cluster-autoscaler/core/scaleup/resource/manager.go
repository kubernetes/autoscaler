/*
Copyright 2022 The Kubernetes Authors.

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

package resource

import (
	"fmt"
	"math"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
)

// LimitUnknown is used as a value in ResourcesLimits if actual limit could not be obtained due to errors talking to cloud provider.
const LimitUnknown = math.MaxInt64

// Manager provides resource checks before scaling up the cluster.
type Manager struct {
	crp customresources.CustomResourcesProcessor
}

// LimitsCheckResult contains the limit check result and the exceeded resources if any.
type LimitsCheckResult struct {
	Exceeded          bool
	ExceededResources []string
}

// Limits is a map: the key is resource type and the value is resource limit.
type Limits map[string]int64

// Delta is a map: the key is resource type and the value is resource delta.
type Delta map[string]int64

// NewManager creates an instance of scale up resource manager with provided parameters.
func NewManager(crp customresources.CustomResourcesProcessor) *Manager {
	return &Manager{
		crp: crp,
	}
}

// DeltaForNode calculates the amount of resources that will be used from the cluster when creating a node.
func (m *Manager) DeltaForNode(ctx *context.AutoscalingContext, nodeInfo *framework.NodeInfo, nodeGroup cloudprovider.NodeGroup) (Delta, errors.AutoscalerError) {
	resultScaleUpDelta := make(Delta)
	nodeCPU, nodeMemory := utils.GetNodeCoresAndMemory(nodeInfo.Node())
	resultScaleUpDelta[cloudprovider.ResourceNameCores] = nodeCPU
	resultScaleUpDelta[cloudprovider.ResourceNameMemory] = nodeMemory

	resourceLimiter, err := ctx.CloudProvider.GetResourceLimiter()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}

	if cloudprovider.ContainsCustomResources(resourceLimiter.GetResources()) {
		resourceTargets, err := m.crp.GetNodeResourceTargets(ctx, nodeInfo.Node(), nodeGroup)
		if err != nil {
			return Delta{}, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("failed to get target custom resources for node group %v: ", nodeGroup.Id())
		}

		for _, resourceTarget := range resourceTargets {
			resultScaleUpDelta[resourceTarget.ResourceType] = resourceTarget.ResourceCount
		}
	}

	return resultScaleUpDelta, nil
}

// ResourcesLeft calculates the amount of resources left in the cluster.
func (m *Manager) ResourcesLeft(ctx *context.AutoscalingContext, nodeInfos map[string]*framework.NodeInfo, nodes []*corev1.Node) (Limits, errors.AutoscalerError) {
	nodesFromNotAutoscaledGroups, err := utils.FilterOutNodesFromNotAutoscaledGroups(nodes, ctx.CloudProvider)
	if err != nil {
		return nil, err.AddPrefix("failed to filter out nodes which are from not autoscaled groups: ")
	}

	totalCores, totalMem, errCoresMem := m.coresMemoryTotal(ctx, nodeInfos, nodesFromNotAutoscaledGroups)

	resourceLimiter, errgo := ctx.CloudProvider.GetResourceLimiter()
	if errgo != nil {
		return nil, errors.ToAutoscalerError(errors.CloudProviderError, errgo)
	}

	var totalResources map[string]int64
	var totalResourcesErr error
	if cloudprovider.ContainsCustomResources(resourceLimiter.GetResources()) {
		totalResources, totalResourcesErr = m.customResourcesTotal(ctx, nodeInfos, nodesFromNotAutoscaledGroups)
	}

	resultScaleUpLimits := make(Limits)
	for _, resource := range resourceLimiter.GetResources() {
		max := resourceLimiter.GetMax(resource)
		// we put only actual limits into final map. No entry means no limit.
		if max > 0 {
			if (resource == cloudprovider.ResourceNameCores || resource == cloudprovider.ResourceNameMemory) && errCoresMem != nil {
				// core resource info missing - no reason to proceed with scale up
				return Limits{}, errCoresMem
			}

			switch {
			case resource == cloudprovider.ResourceNameCores:
				if errCoresMem != nil {
					resultScaleUpLimits[resource] = LimitUnknown
				} else {
					resultScaleUpLimits[resource] = computeBelowMax(totalCores, max)
				}
			case resource == cloudprovider.ResourceNameMemory:
				if errCoresMem != nil {
					resultScaleUpLimits[resource] = LimitUnknown
				} else {
					resultScaleUpLimits[resource] = computeBelowMax(totalMem, max)
				}
			case cloudprovider.IsCustomResource(resource):
				if totalResourcesErr != nil {
					resultScaleUpLimits[resource] = LimitUnknown
				} else {
					resultScaleUpLimits[resource] = computeBelowMax(totalResources[resource], max)
				}
			default:
				klog.Errorf("Scale up limits defined for unsupported resource '%s'", resource)
			}
		}
	}

	return resultScaleUpLimits, nil
}

// ApplyLimits calculates the new node count by applying the left resource limits of the cluster.
func (m *Manager) ApplyLimits(ctx *context.AutoscalingContext, newCount int, resourceLeft Limits, nodeInfo *framework.NodeInfo, nodeGroup cloudprovider.NodeGroup) (int, errors.AutoscalerError) {
	delta, err := m.DeltaForNode(ctx, nodeInfo, nodeGroup)
	if err != nil {
		return 0, err
	}

	for resource, resourceDelta := range delta {
		limit, limitFound := resourceLeft[resource]
		if !limitFound {
			continue
		}

		if limit == LimitUnknown {
			// should never happen - checked before
			return 0, errors.NewAutoscalerError(
				errors.InternalError,
				fmt.Sprintf("limit unknown for resource %s", resource))
		}

		if int64(newCount)*resourceDelta <= limit {
			// no capping required
			continue
		}

		newCount = int(limit / resourceDelta)
		klog.V(1).Infof("Capping scale-up size due to limit for resource %s", resource)
		if newCount < 1 {
			// should never happen - checked before
			return 0, errors.NewAutoscalerError(
				errors.InternalError,
				fmt.Sprintf("cannot create any node; max limit for resource %s reached", resource))
		}
	}

	return newCount, nil
}

// CheckDeltaWithinLimits compares the resource limit and resource delta, and returns the limit check result.
func CheckDeltaWithinLimits(left Limits, delta Delta) LimitsCheckResult {
	exceededResources := sets.NewString()
	for resource, resourceDelta := range delta {
		resourceLeft, found := left[resource]
		if found {
			if (resourceDelta > 0) && (resourceLeft == LimitUnknown || resourceDelta > resourceLeft) {
				exceededResources.Insert(resource)
			}
		}
	}

	if len(exceededResources) > 0 {
		return LimitsCheckResult{true, exceededResources.List()}
	}
	return LimitsNotExceeded()
}

// LimitsNotExceeded returns a not exceeded limit check result.
func LimitsNotExceeded() LimitsCheckResult {
	return LimitsCheckResult{false, []string{}}
}

func (m *Manager) coresMemoryTotal(ctx *context.AutoscalingContext, nodeInfos map[string]*framework.NodeInfo, nodesFromNotAutoscaledGroups []*corev1.Node) (int64, int64, errors.AutoscalerError) {
	var coresTotal int64
	var memoryTotal int64
	for _, nodeGroup := range ctx.CloudProvider.NodeGroups() {
		currentSize, err := nodeGroup.TargetSize()
		if err != nil {
			return 0, 0, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("failed to get node group size of %v: ", nodeGroup.Id())
		}

		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			return 0, 0, errors.NewAutoscalerError(errors.CloudProviderError, "No node info for: %s", nodeGroup.Id())
		}

		if currentSize > 0 {
			nodeCPU, nodeMemory := utils.GetNodeCoresAndMemory(nodeInfo.Node())
			coresTotal = coresTotal + int64(currentSize)*nodeCPU
			memoryTotal = memoryTotal + int64(currentSize)*nodeMemory
		}
	}

	for _, node := range nodesFromNotAutoscaledGroups {
		cores, memory := utils.GetNodeCoresAndMemory(node)
		coresTotal += cores
		memoryTotal += memory
	}

	return coresTotal, memoryTotal, nil
}

func (m *Manager) customResourcesTotal(ctx *context.AutoscalingContext, nodeInfos map[string]*framework.NodeInfo, nodesFromNotAutoscaledGroups []*corev1.Node) (map[string]int64, errors.AutoscalerError) {
	result := make(map[string]int64)
	for _, nodeGroup := range ctx.CloudProvider.NodeGroups() {
		currentSize, err := nodeGroup.TargetSize()
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("failed to get node group size of %v: ", nodeGroup.Id())
		}

		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			return nil, errors.NewAutoscalerError(errors.CloudProviderError, "No node info for: %s", nodeGroup.Id())
		}

		if currentSize > 0 {
			resourceTargets, err := m.crp.GetNodeResourceTargets(ctx, nodeInfo.Node(), nodeGroup)
			if err != nil {
				return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("failed to get custom resource target for node group %v: ", nodeGroup.Id())
			}

			for _, resourceTarget := range resourceTargets {
				if resourceTarget.ResourceType == "" || resourceTarget.ResourceCount == 0 {
					continue
				}
				result[resourceTarget.ResourceType] += resourceTarget.ResourceCount * int64(currentSize)
			}
		}
	}

	for _, node := range nodesFromNotAutoscaledGroups {
		resourceTargets, err := m.crp.GetNodeResourceTargets(ctx, node, nil)
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("failed to get custom resource target for node %v: ", node.Name)
		}

		for _, resourceTarget := range resourceTargets {
			if resourceTarget.ResourceType == "" || resourceTarget.ResourceCount == 0 {
				continue
			}
			result[resourceTarget.ResourceType] += resourceTarget.ResourceCount
		}
	}

	return result, nil
}

func computeBelowMax(total int64, max int64) int64 {
	if total < max {
		return max - total
	}
	return 0
}
