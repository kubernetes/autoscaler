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
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
)

// PodResourceRecommender computes resource recommendation for a Vpa object.
type PodResourceRecommender interface {
	GetRecommendedPodResources(vpa *model.Vpa) RecommendedPodResources
}

// RecommendedPodResources is a Map from container name to recommended resources.
type RecommendedPodResources map[string]RecommendedContainerResources

// RecommendedContainerResources is the recommendation of resources for a
// container.
type RecommendedContainerResources struct {
	// Recommended optimal amount of resources.
	Target model.Resources
	// Recommended minimum amount of resources.
	MinRecommended model.Resources
	// Recommended maximum amount of resources.
	MaxRecommended model.Resources
}

type podResourceRecommender struct {
	targetEstimator     ResourceEstimator
	lowerBoundEstimator ResourceEstimator
	upperBoundEstimator ResourceEstimator
}

// NewPodResourceRecommender returns a new PodResourceRecommender which is an
// aggregation of three ResourceEstimators, one for each of the target, min
// and max recommended resources.
func NewPodResourceRecommender(
	targetEstimator ResourceEstimator,
	lowerBoundEstimator ResourceEstimator,
	upperBoundEstimator ResourceEstimator) PodResourceRecommender {
	return &podResourceRecommender{targetEstimator, lowerBoundEstimator, upperBoundEstimator}
}

// Returns recommended resources for a given Vpa object.
func (r *podResourceRecommender) GetRecommendedPodResources(vpa *model.Vpa) RecommendedPodResources {
	aggregateContainerStateMap := model.BuildAggregateContainerStateMap(vpa, model.MergeForRecommendation, time.Unix(0, 0))
	var recommendation = make(RecommendedPodResources)
	for containerName, aggregatedContainerState := range aggregateContainerStateMap {
		if aggregatedContainerState.TotalSamplesCount > 0 {
			recommendation[containerName] = r.getRecommendedContainerResources(aggregatedContainerState)
		}
	}
	return recommendation
}

// Takes AggregateContainerState and returns a container recommendation.
func (r *podResourceRecommender) getRecommendedContainerResources(s *model.AggregateContainerState) RecommendedContainerResources {
	return RecommendedContainerResources{
		r.targetEstimator.GetResourceEstimation(s),
		r.lowerBoundEstimator.GetResourceEstimation(s),
		r.upperBoundEstimator.GetResourceEstimation(s),
	}
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

	// Use 15% safety margin on top of the recommended resources.
	safetyMarginFraction := 0.15
	// Minimum safety margin is 0.3 core and 300MB memory.
	minSafetyMargin := model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.3),
		model.ResourceMemory: model.MemoryAmountFromBytes(300 * 1024 * 1024),
	}
	targetEstimator = WithSafetyMargin(safetyMarginFraction, minSafetyMargin, targetEstimator)
	lowerBoundEstimator = WithSafetyMargin(safetyMarginFraction, minSafetyMargin, lowerBoundEstimator)
	upperBoundEstimator = WithSafetyMargin(safetyMarginFraction, minSafetyMargin, upperBoundEstimator)

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

	return NewPodResourceRecommender(
		targetEstimator,
		lowerBoundEstimator,
		upperBoundEstimator)
}
