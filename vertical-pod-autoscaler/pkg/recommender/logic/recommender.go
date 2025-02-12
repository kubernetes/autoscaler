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

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var (
	safetyMarginFraction       = flag.Float64("recommendation-margin-fraction", 0.15, `Fraction of usage added as the safety margin to the recommended request`)
	podMinCPUMillicores        = flag.Float64("pod-recommendation-min-cpu-millicores", 25, `Minimum CPU recommendation for a pod`)
	podMinMemoryMb             = flag.Float64("pod-recommendation-min-memory-mb", 250, `Minimum memory recommendation for a pod`)
	targetCPUPercentile        = flag.Float64("target-cpu-percentile", 0.9, "CPU usage percentile that will be used as a base for CPU target recommendation. Doesn't affect CPU lower bound, CPU upper bound nor memory recommendations.")
	lowerBoundCPUPercentile    = flag.Float64("recommendation-lower-bound-cpu-percentile", 0.5, `CPU usage percentile that will be used for the lower bound on CPU recommendation.`)
	upperBoundCPUPercentile    = flag.Float64("recommendation-upper-bound-cpu-percentile", 0.95, `CPU usage percentile that will be used for the upper bound on CPU recommendation.`)
	targetMemoryPercentile     = flag.Float64("target-memory-percentile", 0.9, "Memory usage percentile that will be used as a base for memory target recommendation. Doesn't affect memory lower bound nor memory upper bound.")
	lowerBoundMemoryPercentile = flag.Float64("recommendation-lower-bound-memory-percentile", 0.5, `Memory usage percentile that will be used for the lower bound on memory recommendation.`)
	upperBoundMemoryPercentile = flag.Float64("recommendation-upper-bound-memory-percentile", 0.95, `Memory usage percentile that will be used for the upper bound on memory recommendation.`)
	humanizeMemory             = flag.Bool("humanize-memory", false, "Convert memory values in recommendations to the highest appropriate SI unit with up to 2 decimal places for better readability.")
	roundCPUMillicores         = flag.Int("round-cpu-millicores", 1, `CPU recommendation rounding factor in millicores. The CPU value will always be rounded up to the nearest multiple of this factor.`)
)

// PodResourceRecommender computes resource recommendation for a Vpa object.
type PodResourceRecommender interface {
	GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap) RecommendedPodResources
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
	targetCPU        CPUEstimator
	targetMemory     MemoryEstimator
	lowerBoundCPU    CPUEstimator
	lowerBoundMemory MemoryEstimator
	upperBoundCPU    CPUEstimator
	upperBoundMemory MemoryEstimator
}

func (r *podResourceRecommender) GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap) RecommendedPodResources {
	var recommendation = make(RecommendedPodResources)
	if len(containerNameToAggregateStateMap) == 0 {
		return recommendation
	}

	fraction := 1.0 / float64(len(containerNameToAggregateStateMap))
	minCPU := model.ScaleResource(model.CPUAmountFromCores(*podMinCPUMillicores*0.001), fraction)
	minMemory := model.ScaleResource(model.MemoryAmountFromBytes(*podMinMemoryMb*1024*1024), fraction)

	recommender := &podResourceRecommender{
		WithCPUMinResource(minCPU, r.targetCPU),
		WithMemoryMinResource(minMemory, r.targetMemory),
		WithCPUMinResource(minCPU, r.lowerBoundCPU),
		WithMemoryMinResource(minMemory, r.lowerBoundMemory),
		WithCPUMinResource(minCPU, r.upperBoundCPU),
		WithMemoryMinResource(minMemory, r.upperBoundMemory),
	}

	for containerName, aggregatedContainerState := range containerNameToAggregateStateMap {
		recommendation[containerName] = recommender.estimateContainerResources(aggregatedContainerState)
	}
	return recommendation
}

// Takes AggregateContainerState and returns a container recommendation.
func (r *podResourceRecommender) estimateContainerResources(s *model.AggregateContainerState) RecommendedContainerResources {
	resources := s.GetControlledResources()
	target := model.Resources{model.ResourceCPU: r.targetCPU.GetCPUEstimation(s), model.ResourceMemory: r.targetMemory.GetMemoryEstimation(s)}
	lowerBound := model.Resources{model.ResourceCPU: r.lowerBoundCPU.GetCPUEstimation(s), model.ResourceMemory: r.lowerBoundMemory.GetMemoryEstimation(s)}
	upperBound := model.Resources{model.ResourceCPU: r.upperBoundCPU.GetCPUEstimation(s), model.ResourceMemory: r.upperBoundMemory.GetMemoryEstimation(s)}
	return RecommendedContainerResources{
		FilterControlledResources(target, resources),
		FilterControlledResources(lowerBound, resources),
		FilterControlledResources(upperBound, resources),
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
	targetCPU := NewPercentileCPUEstimator(*targetCPUPercentile)
	lowerBoundCPU := NewPercentileCPUEstimator(*lowerBoundCPUPercentile)
	upperBoundCPU := NewPercentileCPUEstimator(*upperBoundCPUPercentile)

	// Create base memory estimators
	targetMemory := NewPercentileMemoryEstimator(*targetMemoryPercentile)
	lowerBoundMemory := NewPercentileMemoryEstimator(*lowerBoundMemoryPercentile)
	upperBoundMemory := NewPercentileMemoryEstimator(*upperBoundMemoryPercentile)

	// Apply safety margins
	targetCPU = WithCPUMargin(*safetyMarginFraction, targetCPU)
	lowerBoundCPU = WithCPUMargin(*safetyMarginFraction, lowerBoundCPU)
	upperBoundCPU = WithCPUMargin(*safetyMarginFraction, upperBoundCPU)

	targetMemory = WithMemoryMargin(*safetyMarginFraction, targetMemory)
	lowerBoundMemory = WithMemoryMargin(*safetyMarginFraction, lowerBoundMemory)
	upperBoundMemory = WithMemoryMargin(*safetyMarginFraction, upperBoundMemory)

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

	upperBoundCPU = WithCPUConfidenceMultiplier(1.0, 1.0, upperBoundCPU)
	upperBoundMemory = WithMemoryConfidenceMultiplier(1.0, 1.0, upperBoundMemory)

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
	lowerBoundCPU = WithCPUConfidenceMultiplier(0.001, -2.0, lowerBoundCPU)
	lowerBoundMemory = WithMemoryConfidenceMultiplier(0.001, -2.0, lowerBoundMemory)
	return &podResourceRecommender{
		targetCPU,
		targetMemory,
		lowerBoundCPU,
		lowerBoundMemory,
		upperBoundCPU,
		upperBoundMemory,
	}
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
			Target:         model.ResourcesAsResourceList(resources[name].Target, *humanizeMemory, *roundCPUMillicores),
			LowerBound:     model.ResourcesAsResourceList(resources[name].LowerBound, *humanizeMemory, *roundCPUMillicores),
			UpperBound:     model.ResourcesAsResourceList(resources[name].UpperBound, *humanizeMemory, *roundCPUMillicores),
			UncappedTarget: model.ResourcesAsResourceList(resources[name].Target, *humanizeMemory, *roundCPUMillicores),
		})
	}
	recommendation := &vpa_types.RecommendedPodResources{
		ContainerRecommendations: containerResources,
	}
	return recommendation
}
