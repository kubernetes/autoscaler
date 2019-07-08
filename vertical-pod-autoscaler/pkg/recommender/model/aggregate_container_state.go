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

// VPA collects CPU and memory usage measurements from all containers running in
// the cluster and aggregates them in memory in structures called
// AggregateContainerState.
// During aggregation the usage samples are grouped together by the key called
// AggregateStateKey and stored in structures such as histograms of CPU and
// memory usage, that are parts of the AggregateContainerState.
//
// The AggregateStateKey consists of the container name, the namespace and the
// set of labels on the pod the container belongs to. In other words, whenever
// two samples come from containers with the same name, in the same namespace
// and with the same pod labels, they end up in the same histogram.
//
// Recall that VPA produces one recommendation for all containers with a given
// name and namespace, having pod labels that match a given selector. Therefore
// for each VPA object and container name the recommender has to take all
// matching AggregateContainerStates and further aggregate them together, in
// order to obtain the final aggregation that is the input to the recommender
// function.

package model

import (
	"fmt"
	"math"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

// ContainerNameToAggregateStateMap maps a container name to AggregateContainerState
// that aggregates state of containers with that name.
type ContainerNameToAggregateStateMap map[string]*AggregateContainerState

const (
	// SupportedCheckpointVersion is the tag of the supported version of serialized checkpoints.
	// Version id should be incremented on every non incompatible change, i.e. if the new
	// version of the recommender binary can't initialize from the old checkpoint format or the
	// previous version of the recommender binary can't initialize from the new checkpoint format.
	SupportedCheckpointVersion = "v3"
)

// ContainerStateAggregator is an interface for objects that consume and
// aggregate container usage samples.
type ContainerStateAggregator interface {
	// AddSample aggregates a single usage sample.
	AddSample(sample *ContainerUsageSample)
	// SubtractSample removes a single usage sample. The subtracted sample
	// should be equal to some sample that was aggregated with AddSample()
	// in the past.
	SubtractSample(sample *ContainerUsageSample)
	// GetLastRecommendation returns last recommendation calculated for this
	// aggregator.
	GetLastRecommendation() corev1.ResourceList
	// NeedsRecommendation returns true if this aggregator should have
	// a recommendation calculated.
	NeedsRecommendation() bool
	// GetUpdateMode returns the update mode of VPA controlling this aggregator,
	// nil if aggregator is not autoscaled.
	GetUpdateMode() *vpa_types.UpdateMode
}

// AggregateContainerState holds input signals aggregated from a set of containers.
// It can be used as an input to compute the recommendation.
// The CPU and memory distributions use decaying histograms by default
// (see NewAggregateContainerState()).
// Implements ContainerStateAggregator interface.
type AggregateContainerState struct {
	// AggregateCPUUsage is a distribution of all CPU samples.
	AggregateCPUUsage util.Histogram
	// AggregateMemoryPeaks is a distribution of memory peaks from all containers:
	// each container should add one peak per memory aggregation interval (e.g. once every 24h).
	AggregateMemoryPeaks util.Histogram
	// Note: first/last sample timestamps as well as the sample count are based only on CPU samples.
	FirstSampleStart   time.Time
	LastSampleStart    time.Time
	TotalSamplesCount  int
	CreationTime       time.Time
	LastRecommendation corev1.ResourceList
	IsUnderVPA         bool
	UpdateMode         *vpa_types.UpdateMode
}

// GetLastRecommendation returns last recorded recommendation.
func (a *AggregateContainerState) GetLastRecommendation() corev1.ResourceList {
	return a.LastRecommendation
}

// NeedsRecommendation returns true if the state should have recommendation calculated.
func (a *AggregateContainerState) NeedsRecommendation() bool {
	return a.IsUnderVPA
}

// GetUpdateMode returns the update mode of VPA controlling this aggregator,
// nil if aggregator is not autoscaled.
func (a *AggregateContainerState) GetUpdateMode() *vpa_types.UpdateMode {
	return a.UpdateMode
}

// MarkNotAutoscaled registers that this container state is not contorled by
// a VPA object.
func (a *AggregateContainerState) MarkNotAutoscaled() {
	a.IsUnderVPA = false
	a.LastRecommendation = nil
	a.UpdateMode = nil
}

// MergeContainerState merges two AggregateContainerStates.
func (a *AggregateContainerState) MergeContainerState(other *AggregateContainerState) {
	a.AggregateCPUUsage.Merge(other.AggregateCPUUsage)
	a.AggregateMemoryPeaks.Merge(other.AggregateMemoryPeaks)

	if !other.FirstSampleStart.IsZero() && other.FirstSampleStart.Before(a.FirstSampleStart) {
		a.FirstSampleStart = other.FirstSampleStart
	}
	if other.LastSampleStart.After(a.LastSampleStart) {
		a.LastSampleStart = other.LastSampleStart
	}
	a.TotalSamplesCount += other.TotalSamplesCount
}

// NewAggregateContainerState returns a new, empty AggregateContainerState.
func NewAggregateContainerState() *AggregateContainerState {
	return &AggregateContainerState{
		AggregateCPUUsage:    util.NewDecayingHistogram(CPUHistogramOptions, CPUHistogramDecayHalfLife),
		AggregateMemoryPeaks: util.NewDecayingHistogram(MemoryHistogramOptions, MemoryHistogramDecayHalfLife),
		CreationTime:         time.Now(),
	}
}

// AddSample aggregates a single usage sample.
func (a *AggregateContainerState) AddSample(sample *ContainerUsageSample) {
	switch sample.Resource {
	case ResourceCPU:
		a.addCPUSample(sample)
	case ResourceMemory:
		a.AggregateMemoryPeaks.AddSample(BytesFromMemoryAmount(sample.Usage), 1.0, sample.MeasureStart)
	default:
		panic(fmt.Sprintf("AddSample doesn't support resource '%s'", sample.Resource))
	}
}

// SubtractSample removes a single usage sample from an aggregation.
// The subtracted sample should be equal to some sample that was aggregated with
// AddSample() in the past.
// Only memory samples can be subtracted at the moment. Support for CPU could be
// added if necessary.
func (a *AggregateContainerState) SubtractSample(sample *ContainerUsageSample) {
	switch sample.Resource {
	case ResourceMemory:
		a.AggregateMemoryPeaks.SubtractSample(BytesFromMemoryAmount(sample.Usage), 1.0, sample.MeasureStart)
	default:
		panic(fmt.Sprintf("SubtractSample doesn't support resource '%s'", sample.Resource))
	}
}

func (a *AggregateContainerState) addCPUSample(sample *ContainerUsageSample) {
	cpuUsageCores := CoresFromCPUAmount(sample.Usage)
	cpuRequestCores := CoresFromCPUAmount(sample.Request)
	// Samples are added with the weight equal to the current request. This means that
	// whenever the request is increased, the history accumulated so far effectively decays,
	// which helps react quickly to CPU starvation.
	a.AggregateCPUUsage.AddSample(
		cpuUsageCores, math.Max(cpuRequestCores, minSampleWeight), sample.MeasureStart)
	if sample.MeasureStart.After(a.LastSampleStart) {
		a.LastSampleStart = sample.MeasureStart
	}
	if a.FirstSampleStart.IsZero() || sample.MeasureStart.Before(a.FirstSampleStart) {
		a.FirstSampleStart = sample.MeasureStart
	}
	a.TotalSamplesCount++
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
		return fmt.Errorf("unsuported checkpoint version %s", checkpoint.Version)
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

func (a *AggregateContainerState) isExpired(now time.Time) bool {
	if a.isEmpty() {
		return now.Sub(a.CreationTime) >= MemoryAggregationWindowLength
	}
	return now.Sub(a.LastSampleStart) >= MemoryAggregationWindowLength
}

func (a *AggregateContainerState) isEmpty() bool {
	return a.TotalSamplesCount == 0
}

// AggregateStateByContainerName takes a set of AggregateContainerStates and merge them
// grouping by the container name. The result is a map from the container name to the aggregation
// from all input containers with the given name.
func AggregateStateByContainerName(aggregateContainerStateMap aggregateContainerStatesMap) ContainerNameToAggregateStateMap {
	containerNameToAggregateStateMap := make(ContainerNameToAggregateStateMap)
	for aggregationKey, aggregation := range aggregateContainerStateMap {
		containerName := aggregationKey.ContainerName()
		aggregateContainerState, isInitialized := containerNameToAggregateStateMap[containerName]
		if !isInitialized {
			aggregateContainerState = NewAggregateContainerState()
			containerNameToAggregateStateMap[containerName] = aggregateContainerState
		}
		aggregateContainerState.MergeContainerState(aggregation)
	}
	return containerNameToAggregateStateMap
}

// ContainerStateAggregatorProxy is a wrapper for ContainerStateAggregator
// that creates CnontainerStateAgregator for container if it is no longer
// present in the cluster state.
type ContainerStateAggregatorProxy struct {
	containerID ContainerID
	cluster     *ClusterState
}

// NewContainerStateAggregatorProxy creates a ContainerStateAggregatorProxy
// pointing to the cluster state.
func NewContainerStateAggregatorProxy(cluster *ClusterState, containerID ContainerID) ContainerStateAggregator {
	return &ContainerStateAggregatorProxy{containerID, cluster}
}

// AddSample adds a container sample to the aggregator.
func (p *ContainerStateAggregatorProxy) AddSample(sample *ContainerUsageSample) {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	aggregator.AddSample(sample)
}

// SubtractSample subtracts a container sample from the aggregator.
func (p *ContainerStateAggregatorProxy) SubtractSample(sample *ContainerUsageSample) {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	aggregator.SubtractSample(sample)
}

// GetLastRecommendation returns last recorded recommendation.
func (p *ContainerStateAggregatorProxy) GetLastRecommendation() corev1.ResourceList {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	return aggregator.GetLastRecommendation()
}

// NeedsRecommendation returns true if the aggregator should have recommendation calculated.
func (p *ContainerStateAggregatorProxy) NeedsRecommendation() bool {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	return aggregator.NeedsRecommendation()
}

// GetUpdateMode returns update mode of VPA controlling the aggregator.
func (p *ContainerStateAggregatorProxy) GetUpdateMode() *vpa_types.UpdateMode {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	return aggregator.GetUpdateMode()
}
