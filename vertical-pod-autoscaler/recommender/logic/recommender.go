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
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/util"
)

const (
	// SupportedCheckpointVersion is the tag of the supported version of serialized checkpoints.
	SupportedCheckpointVersion = "v1"
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
	aggregateContainerStateMap := buildAggregateContainerStateMap(&vpa.Pods)
	var recommendation RecommendedPodResources = make(RecommendedPodResources)
	for containerName, aggregatedContainerState := range aggregateContainerStateMap {
		if aggregatedContainerState.totalSamplesCount > 0 {
			recommendation[containerName] = r.getRecommendedContainerResources(aggregatedContainerState)
		}
	}
	return recommendation
}

// AggregateContainerState holds input signals aggregated from a set of containers.
// It can be used as an input to compute the recommendation.
type AggregateContainerState struct {
	aggregateCPUUsage    util.Histogram
	aggregateMemoryPeaks util.Histogram
	firstSampleStart     time.Time
	lastSampleStart      time.Time
	totalSamplesCount    int
}

// Merges the state of an individual container into AggregateContainerState.
func (a *AggregateContainerState) mergeContainerState(container *model.ContainerState) {
	a.aggregateCPUUsage.Merge(container.CPUUsage)
	memoryPeaks := container.MemoryUsagePeaks.Contents()
	peakTime := container.WindowEnd
	for i := len(memoryPeaks) - 1; i >= 0; i-- {
		a.aggregateMemoryPeaks.AddSample(float64(memoryPeaks[i]), 1.0, peakTime)
		peakTime = peakTime.Add(-model.MemoryAggregationInterval)
	}
	// Note: we look at CPU samples to calculate the total lifespan and sample count.
	if a.firstSampleStart.IsZero() || (!container.FirstCPUSampleStart.IsZero() && container.FirstCPUSampleStart.Before(a.firstSampleStart)) {
		a.firstSampleStart = container.FirstCPUSampleStart
	}
	if container.LastCPUSampleStart.After(a.lastSampleStart) {
		a.lastSampleStart = container.LastCPUSampleStart
	}
	a.totalSamplesCount += container.CPUSamplesCount
}

// Returns a new, empty AggregateContainerState.
func newAggregateContainerState() *AggregateContainerState {
	return &AggregateContainerState{
		aggregateCPUUsage:    util.NewDecayingHistogram(model.CPUHistogramOptions, model.CPUHistogramDecayHalfLife),
		aggregateMemoryPeaks: util.NewDecayingHistogram(model.MemoryHistogramOptions, model.MemoryHistogramDecayHalfLife),
	}
}

// Takes a set of pods and groups their containers by name.
// Returns a map from the container name to the AggregateContainerState.
func buildAggregateContainerStateMap(pods *map[model.PodID]*model.PodState) map[string]*AggregateContainerState {
	aggregateContainerStateMap := make(map[string]*AggregateContainerState)
	for _, pod := range *pods {
		for containerName, container := range pod.Containers {
			aggregateContainerState, isInitialized := aggregateContainerStateMap[containerName]
			if !isInitialized {
				aggregateContainerState = newAggregateContainerState()
				aggregateContainerStateMap[containerName] = aggregateContainerState
			}
			aggregateContainerState.mergeContainerState(container)
		}
	}
	return aggregateContainerStateMap
}

// Takes AggregateContainerState and returns a container recommendation.
func (r *podResourceRecommender) getRecommendedContainerResources(s *AggregateContainerState) RecommendedContainerResources {
	return RecommendedContainerResources{
		r.targetEstimator.GetResourceEstimation(s),
		r.lowerBoundEstimator.GetResourceEstimation(s),
		r.upperBoundEstimator.GetResourceEstimation(s),
	}
}

// SaveToCheckpoint serializes AggregateContainerState as VerticalPodAutoscalerCheckpointStatus.
// The serialization may result in loss of precission of the histograms.
func (a *AggregateContainerState) SaveToCheckpoint() (*vpa_types.VerticalPodAutoscalerCheckpointStatus, error) {
	memory, err := a.aggregateMemoryPeaks.SaveToChekpoint()
	if err != nil {
		return nil, err
	}
	cpu, err := a.aggregateCPUUsage.SaveToChekpoint()
	if err != nil {
		return nil, err
	}
	return &vpa_types.VerticalPodAutoscalerCheckpointStatus{
		FirstSampleStart:  metav1.NewTime(a.firstSampleStart),
		LastSampleStart:   metav1.NewTime(a.lastSampleStart),
		TotalSamplesCount: a.totalSamplesCount,
		MemoryHistogram:   *memory,
		CPUHistogram:      *cpu,
		Version:           SupportedCheckpointVersion,
	}, nil
}

// LoadFromCheckpoint deserializes data from VerticalPodAutoscalerCheckpointStatus
// into the AggregateContainerState.
func (a *AggregateContainerState) LoadFromCheckpoint(checkpoint *vpa_types.VerticalPodAutoscalerCheckpointStatus) error {
	if checkpoint.Version != SupportedCheckpointVersion {
		return fmt.Errorf("Unssuported checkpoint version %s", checkpoint.Version)
	}
	a.totalSamplesCount = checkpoint.TotalSamplesCount
	a.firstSampleStart = checkpoint.FirstSampleStart.Time
	a.lastSampleStart = checkpoint.LastSampleStart.Time
	err := a.aggregateMemoryPeaks.LoadFromCheckpoint(&checkpoint.MemoryHistogram)
	if err != nil {
		return err
	}
	err = a.aggregateCPUUsage.LoadFromCheckpoint(&checkpoint.CPUHistogram)
	if err != nil {
		return err
	}
	return nil
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

	// Use 10% safety margin on top of the recommended resources.
	safetyMarginFraction := 0.1
	// Minimum safety margin is 0.2 core and 300MB memory.
	minSafetyMargin := model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.2),
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
