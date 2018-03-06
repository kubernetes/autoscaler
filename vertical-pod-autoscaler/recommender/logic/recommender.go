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
