/*
Copyright 2018 The Kubernetes Authors.

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

package model

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/util"
)

// ContainerMergePolicy contorls how MergeContainerState cobines samples.
type ContainerMergePolicy bool

const (
	// SupportedCheckpointVersion is the tag of the supported version of serialized checkpoints.
	SupportedCheckpointVersion = "v1"

	// MergeForRecommendation means that all samples are combined during MergeContainerState call.
	MergeForRecommendation ContainerMergePolicy = true
	// MergeForCheckpoint controls that  MergeContainerState call will omit last peak memory
	// sample if it would result in positive feedback loop during crash loop.
	MergeForCheckpoint ContainerMergePolicy = false
)

// AggregateContainerState holds input signals aggregated from a set of containers.
// It can be used as an input to compute the recommendation.
type AggregateContainerState struct {
	AggregateCPUUsage    util.Histogram
	AggregateMemoryPeaks util.Histogram
	FirstSampleStart     time.Time
	LastSampleStart      time.Time
	TotalSamplesCount    int
}

// MergeContainerState merges the state of an individual container into AggregateContainerState.
func (a *AggregateContainerState) MergeContainerState(container *ContainerState,
	mergePolicy ContainerMergePolicy, now time.Time) {
	a.AggregateCPUUsage.Merge(container.CPUUsage)
	memoryPeaks := container.MemoryUsagePeaks.Contents()
	peakTime := container.WindowEnd
	for i := len(memoryPeaks) - 1; i >= 0; i-- {
		if mergePolicy == MergeForRecommendation || peakTime.Before(now) {
			a.AggregateMemoryPeaks.AddSample(float64(memoryPeaks[i]), 1.0, peakTime)
		}
		peakTime = peakTime.Add(-MemoryAggregationInterval)
	}
	// Note: we look at CPU samples to calculate the total lifespan and sample count.
	if a.FirstSampleStart.IsZero() || (!container.FirstCPUSampleStart.IsZero() && container.FirstCPUSampleStart.Before(a.FirstSampleStart)) {
		a.FirstSampleStart = container.FirstCPUSampleStart
	}
	if container.LastCPUSampleStart.After(a.LastSampleStart) {
		a.LastSampleStart = container.LastCPUSampleStart
	}
	a.TotalSamplesCount += container.CPUSamplesCount
}

// NewAggregateContainerState returns a new, empty AggregateContainerState.
func NewAggregateContainerState() *AggregateContainerState {
	return &AggregateContainerState{
		AggregateCPUUsage:    util.NewDecayingHistogram(CPUHistogramOptions, CPUHistogramDecayHalfLife),
		AggregateMemoryPeaks: util.NewDecayingHistogram(MemoryHistogramOptions, MemoryHistogramDecayHalfLife),
	}
}

// SaveToCheckpoint serializes AggregateContainerState as VerticalPodAutoscalerCheckpointStatus.
// The serialization may result in loss of precission of the histograms.
func (a *AggregateContainerState) SaveToCheckpoint() (*vpa_types.VerticalPodAutoscalerCheckpointStatus, error) {
	memory, err := a.AggregateMemoryPeaks.SaveToChekpoint()
	if err != nil {
		return nil, err
	}
	cpu, err := a.AggregateCPUUsage.SaveToChekpoint()
	if err != nil {
		return nil, err
	}
	return &vpa_types.VerticalPodAutoscalerCheckpointStatus{
		FirstSampleStart:  metav1.NewTime(a.FirstSampleStart),
		LastSampleStart:   metav1.NewTime(a.LastSampleStart),
		TotalSamplesCount: a.TotalSamplesCount,
		MemoryHistogram:   *memory,
		CPUHistogram:      *cpu,
		Version:           SupportedCheckpointVersion,
	}, nil
}

// LoadFromCheckpoint deserializes data from VerticalPodAutoscalerCheckpointStatus
// into the AggregateContainerState.
func (a *AggregateContainerState) LoadFromCheckpoint(checkpoint *vpa_types.VerticalPodAutoscalerCheckpointStatus) error {
	if checkpoint.Version != SupportedCheckpointVersion {
		return fmt.Errorf("Unsuported checkpoint version %s", checkpoint.Version)
	}
	a.TotalSamplesCount = checkpoint.TotalSamplesCount
	a.FirstSampleStart = checkpoint.FirstSampleStart.Time
	a.LastSampleStart = checkpoint.LastSampleStart.Time
	err := a.AggregateMemoryPeaks.LoadFromCheckpoint(&checkpoint.MemoryHistogram)
	if err != nil {
		return err
	}
	err = a.AggregateCPUUsage.LoadFromCheckpoint(&checkpoint.CPUHistogram)
	if err != nil {
		return err
	}
	return nil
}

// DeepCopy returns a copy of the AggregateContainerState
func (a *AggregateContainerState) DeepCopy() *AggregateContainerState {
	copy := NewAggregateContainerState()
	copy.TotalSamplesCount = a.TotalSamplesCount
	copy.FirstSampleStart = a.FirstSampleStart
	copy.LastSampleStart = a.FirstSampleStart
	copy.AggregateCPUUsage.Merge(a.AggregateCPUUsage)
	copy.AggregateMemoryPeaks.Merge(a.AggregateMemoryPeaks)
	return copy
}

// BuildAggregateContainerStateMap takes a set of pods and groups their containers by name.
// If checkpoint data is available it is incorporated into AggregateContainerState
func BuildAggregateContainerStateMap(vpa *Vpa, mergePolicy ContainerMergePolicy, now time.Time) map[string]*AggregateContainerState {
	aggregateContainerStateMap := make(map[string]*AggregateContainerState)
	for k, v := range vpa.ContainerCheckpoints {
		aggregateContainerStateMap[k] = v.DeepCopy()
	}
	for _, pod := range vpa.Pods {
		for containerName, container := range pod.Containers {
			aggregateContainerState, isInitialized := aggregateContainerStateMap[containerName]
			if !isInitialized {
				aggregateContainerState = NewAggregateContainerState()
				aggregateContainerStateMap[containerName] = aggregateContainerState
			}
			aggregateContainerState.MergeContainerState(container, mergePolicy, now)
		}
	}
	return aggregateContainerStateMap
}
