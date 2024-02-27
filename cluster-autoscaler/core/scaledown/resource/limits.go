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
	"math"
	"reflect"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/actuation"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	klog "k8s.io/klog/v2"
)

// Limits represents the total amount of resources that can be deleted from the cluster.
type Limits map[string]int64

// Delta represents the amount of resources that can be removed from the cluster by deleting a single node.
type Delta map[string]int64

// LimitsFinder decides what node removals would violate cluster resource limits.
type LimitsFinder struct {
	crp customresources.CustomResourcesProcessor
}

// NewLimitsFinder returns a new LimitsFinder
func NewLimitsFinder(crp customresources.CustomResourcesProcessor) *LimitsFinder {
	return &LimitsFinder{
		crp: crp,
	}
}

// used as a value in Limits if actual limit could not be obtained due to errors talking to cloud provider
const limitUnknown = math.MinInt64

// NoLimits returns empty Limits
func NoLimits() Limits {
	return nil
}

// LimitsLeft returns the amount of each resource that can be deleted from the
// cluster without violating any constraints.
func (lf *LimitsFinder) LimitsLeft(context *context.AutoscalingContext, nodes []*apiv1.Node, resourceLimiter *cloudprovider.ResourceLimiter, timestamp time.Time) Limits {
	totalCores, totalMem := coresMemoryTotal(context, nodes, timestamp)

	var totalResources map[string]int64
	var totalResourcesErr error
	if cloudprovider.ContainsCustomResources(resourceLimiter.GetResources()) {
		totalResources, totalResourcesErr = lf.customResourcesTotal(context, nodes, timestamp)
	}

	resultScaleDownLimits := make(Limits)
	for _, resource := range resourceLimiter.GetResources() {
		min := resourceLimiter.GetMin(resource)

		// we put only actual limits into final map. No entry means no limit.
		if min > 0 {
			switch {
			case resource == cloudprovider.ResourceNameCores:
				resultScaleDownLimits[resource] = computeAboveMin(totalCores, min)
			case resource == cloudprovider.ResourceNameMemory:
				resultScaleDownLimits[resource] = computeAboveMin(totalMem, min)
			case cloudprovider.IsCustomResource(resource):
				if totalResourcesErr != nil {
					resultScaleDownLimits[resource] = limitUnknown
				} else {
					resultScaleDownLimits[resource] = computeAboveMin(totalResources[resource], min)
				}
			default:
				klog.Errorf("Scale down limits defined for unsupported resource '%s'", resource)
			}
		}
	}
	return resultScaleDownLimits
}

func computeAboveMin(total int64, min int64) int64 {
	if total > min {
		return total - min
	}
	return 0
}

func coresMemoryTotal(ctx *context.AutoscalingContext, nodes []*apiv1.Node, timestamp time.Time) (int64, int64) {
	var coresTotal, memoryTotal int64
	for _, node := range nodes {
		if actuation.IsNodeBeingDeleted(ctx, node, timestamp) {
			// Nodes being deleted do not count towards total cluster resources
			continue
		}
		cores, memory := core_utils.GetNodeCoresAndMemory(node)

		coresTotal += cores
		memoryTotal += memory
	}

	return coresTotal, memoryTotal
}

func (lf *LimitsFinder) customResourcesTotal(context *context.AutoscalingContext, nodes []*apiv1.Node, timestamp time.Time) (map[string]int64, error) {
	result := make(map[string]int64)
	ngCache := make(map[string][]customresources.CustomResourceTarget)
	for _, node := range nodes {
		if actuation.IsNodeBeingDeleted(context, node, timestamp) {
			// Nodes being deleted do not count towards total cluster resources
			continue
		}
		nodeGroup, err := context.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("cannot get node group for node %v when calculating cluster custom resource usage", node.Name)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			// We do not trust cloud providers to return properly constructed nil for interface type - hence the reflection check.
			// See https://golang.org/doc/faq#nil_error
			// TODO[lukaszos] consider creating cloud_provider sanitizer which will wrap cloud provider and ensure sane behaviour.
			nodeGroup = nil
		}

		var resourceTargets []customresources.CustomResourceTarget
		var cacheHit bool

		if nodeGroup != nil {
			resourceTargets, cacheHit = ngCache[nodeGroup.Id()]
		}
		if !cacheHit {
			resourceTargets, err = lf.crp.GetNodeResourceTargets(context, node, nodeGroup)
			if err != nil {
				return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("cannot get custom resource count for node %v when calculating cluster custom resource usage")
			}
			if nodeGroup != nil {
				ngCache[nodeGroup.Id()] = resourceTargets
			}
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

// DeepCopy returns a copy of the original limits.
func (l Limits) DeepCopy() Limits {
	copy := Limits{}
	for k, v := range l {
		copy[k] = v
	}
	return copy
}

// DeltaForNode calculates the amount of resources that will disappear from
// the cluster if a given node is deleted.
func (lf *LimitsFinder) DeltaForNode(context *context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup, resourcesWithLimits []string) (Delta, errors.AutoscalerError) {
	resultScaleDownDelta := make(Delta)

	nodeCPU, nodeMemory := core_utils.GetNodeCoresAndMemory(node)
	resultScaleDownDelta[cloudprovider.ResourceNameCores] = nodeCPU
	resultScaleDownDelta[cloudprovider.ResourceNameMemory] = nodeMemory

	if cloudprovider.ContainsCustomResources(resourcesWithLimits) {
		resourceTargets, err := lf.crp.GetNodeResourceTargets(context, node, nodeGroup)
		if err != nil {
			return Delta{}, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("failed to get node %v custom resources: %v", node.Name)
		}
		for _, resourceTarget := range resourceTargets {
			resultScaleDownDelta[resourceTarget.ResourceType] = resourceTarget.ResourceCount
		}
	}
	return resultScaleDownDelta, nil
}

// LimitsCheckResult contains information about resources exceeded during a check.
type LimitsCheckResult struct {
	ExceededResources []string
}

// Exceeded returns true iff at least one resource limit was exceeded during
// the check.
func (l LimitsCheckResult) Exceeded() bool {
	return len(l.ExceededResources) > 0
}

// CheckDeltaWithinLimits checks if the resource delta is within limits.
func (l *Limits) CheckDeltaWithinLimits(delta Delta) LimitsCheckResult {
	exceededResources := sets.NewString()
	for resource, resourceDelta := range delta {
		resourceLeft, found := (*l)[resource]
		if found {
			if (resourceDelta > 0) && (resourceLeft == limitUnknown || resourceDelta > resourceLeft) {
				exceededResources.Insert(resource)
			}
		}
	}
	return LimitsCheckResult{exceededResources.List()}
}

// TryDecrementBy makes an attempt to decrement resource limits by the provided
// delta. If the limits would be exceeded, they are not modified.
func (l *Limits) TryDecrementBy(delta Delta) LimitsCheckResult {
	result := l.CheckDeltaWithinLimits(delta)
	if result.Exceeded() {
		return result
	}
	for resource, resourceDelta := range delta {
		resourceLeft, found := (*l)[resource]
		if found {
			(*l)[resource] = resourceLeft - resourceDelta
		}
	}
	return LimitsCheckResult{}
}
