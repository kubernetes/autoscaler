/*
Copyright 2017 The Kubernetes Authors.

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

package logic

import (
	"flag"
	"sort"

	apiv1 "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/klog/v2"
)

var (
	safetyMarginFraction             = flag.Float64("recommendation-margin-fraction", 0.15, `Fraction of usage added as the safety margin to the recommended request`)
	podMinCPUMillicores              = flag.Float64("pod-recommendation-min-cpu-millicores", 25, `Minimum CPU recommendation for a pod`)
	podMinMemoryMb                   = flag.Float64("pod-recommendation-min-memory-mb", 250, `Minimum memory recommendation for a pod`)
	podAutoDetectedMaxNodeAllocation = flag.Bool("auto-detected-max-node-allocation", true, "If true, the recommender will use the maximum node allocation as the upper bound for the recommendation.")
	targetCPUPercentile              = flag.Float64("target-cpu-percentile", 0.9, "CPU usage percentile that will be used as a base for CPU target recommendation. Doesn't affect CPU lower bound, CPU upper bound nor memory recommendations.")
	lowerBoundCPUPercentile          = flag.Float64("recommendation-lower-bound-cpu-percentile", 0.5, `CPU usage percentile that will be used for the lower bound on CPU recommendation.`)
	upperBoundCPUPercentile          = flag.Float64("recommendation-upper-bound-cpu-percentile", 0.95, `CPU usage percentile that will be used for the upper bound on CPU recommendation.`)
	targetMemoryPercentile           = flag.Float64("target-memory-percentile", 0.9, "Memory usage percentile that will be used as a base for memory target recommendation. Doesn't affect memory lower bound nor memory upper bound.")
	lowerBoundMemoryPercentile       = flag.Float64("recommendation-lower-bound-memory-percentile", 0.5, `Memory usage percentile that will be used for the lower bound on memory recommendation.`)
	upperBoundMemoryPercentile       = flag.Float64("recommendation-upper-bound-memory-percentile", 0.95, `Memory usage percentile that will be used for the upper bound on memory recommendation.`)
)

// PodResourceRecommender computes resource recommendation for a Vpa object.
type PodResourceRecommender interface {
	GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap, nodes map[string]apiv1.NodeStatus) RecommendedPodResources
}

// RecommendedPodResources is a Map from container name to recommended resources.
type RecommendedPodResources map[string]RecommendedContainerResources

// RecommendedContainerResources is the recommendation of resources for a
// container.
type RecommendedContainerResources struct {
	// Recommended optimal amount of resources.
	Target model.Resources
	// Recommended minimum amount of resources.
	LowerBound model.Resources
	// Recommended maximum amount of resources.
	UpperBound model.Resources
}

type podResourceRecommender struct {
	targetEstimator     ResourceEstimator
	lowerBoundEstimator ResourceEstimator
	upperBoundEstimator ResourceEstimator
}

func (r *podResourceRecommender) GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap, nodes map[string]apiv1.NodeStatus) RecommendedPodResources {
	var recommendation = make(RecommendedPodResources)
	if len(containerNameToAggregateStateMap) == 0 {
		return recommendation
	}

	fraction := 1.0 / float64(len(containerNameToAggregateStateMap))
	minResources := model.Resources{
		model.ResourceCPU:    model.ScaleResource(model.CPUAmountFromCores(*podMinCPUMillicores*0.001), fraction),
		model.ResourceMemory: model.ScaleResource(model.MemoryAmountFromBytes(*podMinMemoryMb*1024*1024), fraction),
	}

	recommender := &podResourceRecommender{
		WithMinResources(minResources, r.targetEstimator),
		WithMinResources(minResources, r.lowerBoundEstimator),
		WithMinResources(minResources, r.upperBoundEstimator),
	}

	if *podAutoDetectedMaxNodeAllocation {
		// Get the maximum node allocatable resources
		klog.InfoS("Using the maximum node allocatable resources as the upper bound for the recommendation")
		maxNodeStatus := r.GetMaxNodeStatus(nodes, r.upperBoundEstimator)
		maxAllocatableCPU := model.CPUAmountFromCores(float64(maxNodeStatus.Allocatable.Cpu().MilliValue()) / 1000)
		maxAllocatableMemory := model.MemoryAmountFromBytes(float64(maxNodeStatus.Allocatable.Memory().Value()))
		for containerName, aggregatedContainerState := range containerNameToAggregateStateMap {
			containerRecommendation := recommender.estimateContainerResources(aggregatedContainerState)
			// Cap the upper bound to the maximum node allocatable resources
			containerRecommendation.UpperBound[model.ResourceCPU] = model.ResourceAmountMax(
				containerRecommendation.UpperBound[model.ResourceCPU],
				model.CPUAmountFromCores(model.CoresFromCPUAmount(maxAllocatableCPU)*0.95), // Use 95% of max allocatable as a safety margin
			)
			containerRecommendation.UpperBound[model.ResourceMemory] = model.ResourceAmountMax(
				containerRecommendation.UpperBound[model.ResourceMemory],
				model.MemoryAmountFromBytes(model.BytesFromMemoryAmount(maxAllocatableMemory)*0.95), // Use 95% of max allocatable as a safety margin
			)
			klog.InfoS("Recommendation for container", "container", containerName, "target", containerRecommendation.Target, "lowerBound", containerRecommendation.LowerBound, "upperBound", containerRecommendation.UpperBound)
			recommendation[containerName] = containerRecommendation
		}
		return recommendation
	}

	for containerName, aggregatedContainerState := range containerNameToAggregateStateMap {
		recommendation[containerName] = recommender.estimateContainerResources(aggregatedContainerState)
	}
	return recommendation
}

// Takes AggregateContainerState and returns a container recommendation.
func (r *podResourceRecommender) estimateContainerResources(s *model.AggregateContainerState) RecommendedContainerResources {
	return RecommendedContainerResources{
		FilterControlledResources(r.targetEstimator.GetResourceEstimation(s), s.GetControlledResources()),
		FilterControlledResources(r.lowerBoundEstimator.GetResourceEstimation(s), s.GetControlledResources()),
		FilterControlledResources(r.upperBoundEstimator.GetResourceEstimation(s), s.GetControlledResources()),
	}
}

// FilterControlledResources returns estimations from 'estimation' only for resources present in 'controlledResources'.
func FilterControlledResources(estimation model.Resources, controlledResources []model.ResourceName) model.Resources {
	result := make(model.Resources)
	for _, resource := range controlledResources {
		if value, ok := estimation[resource]; ok {
			result[resource] = value
		}
	}
	return result
}

// CreatePodResourceRecommender returns the primary recommender.
func CreatePodResourceRecommender() PodResourceRecommender {
	targetEstimator := NewPercentileEstimator(*targetCPUPercentile, *targetMemoryPercentile)
	lowerBoundEstimator := NewPercentileEstimator(*lowerBoundCPUPercentile, *lowerBoundMemoryPercentile)
	upperBoundEstimator := NewPercentileEstimator(*upperBoundCPUPercentile, *upperBoundMemoryPercentile)

	targetEstimator = WithMargin(*safetyMarginFraction, targetEstimator)
	lowerBoundEstimator = WithMargin(*safetyMarginFraction, lowerBoundEstimator)
	upperBoundEstimator = WithMargin(*safetyMarginFraction, upperBoundEstimator)

	// Apply confidence multiplier to the upper bound estimator. This means
	// that the updater will be less eager to evict pods with short history
	// in order to reclaim unused resources.
	// Using the confidence multiplier 1 with exponent +1 means that
	// the upper bound is multiplied by (1 + 1/history-length-in-days).
	// See estimator.go to see how the history length and the confidence
	// multiplier are determined. The formula yields the following multipliers:
	// No history     : *INF  (do not force pod eviction)
	// 12h history    : *3    (force pod eviction if the request is > 3 * upper bound)
	// 24h history    : *2
	// 1 week history : *1.14
	upperBoundEstimator = WithConfidenceMultiplier(1.0, 1.0, upperBoundEstimator)

	// Apply confidence multiplier to the lower bound estimator. This means
	// that the updater will be less eager to evict pods with short history
	// in order to provision them with more resources.
	// Using the confidence multiplier 0.001 with exponent -2 means that
	// the lower bound is multiplied by the factor (1 + 0.001/history-length-in-days)^-2
	// (which is very rapidly converging to 1.0).
	// See estimator.go to see how the history length and the confidence
	// multiplier are determined. The formula yields the following multipliers:
	// No history   : *0   (do not force pod eviction)
	// 5m history   : *0.6 (force pod eviction if the request is < 0.6 * lower bound)
	// 30m history  : *0.9
	// 60m history  : *0.95
	lowerBoundEstimator = WithConfidenceMultiplier(0.001, -2.0, lowerBoundEstimator)
	return &podResourceRecommender{
		targetEstimator,
		lowerBoundEstimator,
		upperBoundEstimator}
}

// MapToListOfRecommendedContainerResources converts the map of RecommendedContainerResources into a stable sorted list
// This can be used to get a stable sequence while ranging on the data
func MapToListOfRecommendedContainerResources(resources RecommendedPodResources) *vpa_types.RecommendedPodResources {
	containerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
	// Sort the container names from the map. This is because maps are an
	// unordered data structure, and iterating through the map will return
	// a different order on every call.
	containerNames := make([]string, 0, len(resources))
	for containerName := range resources {
		containerNames = append(containerNames, containerName)
	}
	sort.Strings(containerNames)
	// Create the list of recommendations for each container.
	for _, name := range containerNames {
		containerResources = append(containerResources, vpa_types.RecommendedContainerResources{
			ContainerName:  name,
			Target:         model.ResourcesAsResourceList(resources[name].Target),
			LowerBound:     model.ResourcesAsResourceList(resources[name].LowerBound),
			UpperBound:     model.ResourcesAsResourceList(resources[name].UpperBound),
			UncappedTarget: model.ResourcesAsResourceList(resources[name].Target),
		})
	}
	recommendation := &vpa_types.RecommendedPodResources{
		ContainerRecommendations: containerResources,
	}
	return recommendation
}

func (r *podResourceRecommender) GetMaxNodeStatus(nodes map[string]apiv1.NodeStatus, estimator ResourceEstimator) apiv1.NodeStatus {
	var maxNodeStatus apiv1.NodeStatus
	var selectedNode string
	maxNodeCpu := model.CPUAmountFromCores(0)
	maxNodeMemory := model.MemoryAmountFromBytes(0)
	for nodeName, nodeStatus := range nodes {
		nodeCpu := model.CPUAmountFromCores(float64(nodeStatus.Allocatable.Cpu().MilliValue()) / 1000)
		nodeMemory := model.MemoryAmountFromBytes(float64(nodeStatus.Allocatable.Memory().Value()))
		if nodeCpu > maxNodeCpu && nodeMemory > maxNodeMemory {
			selectedNode = nodeName
			maxNodeCpu = nodeCpu
			maxNodeMemory = nodeMemory
			maxNodeStatus = nodeStatus
		}
	}
	klog.InfoS("Selected node for max node status", "node", selectedNode, "cpu", maxNodeCpu, "memory", maxNodeMemory)
	return maxNodeStatus
}
