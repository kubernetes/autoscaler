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

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var (
	safetyMarginFraction = flag.Float64("recommendation-margin-fraction", 0.15, `Fraction of usage added as the safety margin to the recommended request`)
	podMinCPUMillicores  = flag.Float64("pod-recommendation-min-cpu-millicores", 25, `Minimum CPU recommendation for a pod`)
	podMinMemoryMb       = flag.Float64("pod-recommendation-min-memory-mb", 250, `Minimum memory recommendation for a pod`)
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
	targetEstimator     ResourceEstimator
	lowerBoundEstimator ResourceEstimator
	upperBoundEstimator ResourceEstimator
}

func (r *podResourceRecommender) GetRecommendedPodResources(containerNameToAggregateStateMap model.ContainerNameToAggregateStateMap) RecommendedPodResources {
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
	targetCPUPercentile := 0.9
	lowerBoundCPUPercentile := 0.5
	upperBoundCPUPercentile := 0.95

	targetMemoryPeaksPercentile := 0.9
	lowerBoundMemoryPeaksPercentile := 0.5
	upperBoundMemoryPeaksPercentile := 0.95

	targetEstimator := NewPercentileEstimator(targetCPUPercentile, targetMemoryPeaksPercentile)
	lowerBoundEstimator := NewPercentileEstimator(lowerBoundCPUPercentile, lowerBoundMemoryPeaksPercentile)
	upperBoundEstimator := NewPercentileEstimator(upperBoundCPUPercentile, upperBoundMemoryPeaksPercentile)

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
