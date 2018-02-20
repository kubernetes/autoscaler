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

package model

import (
	"math"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/util"
)

// ContainerUsageSample is a measure of resource usage of a container over some
// interval.
type ContainerUsageSample struct {
	// Start of the measurement interval.
	MeasureStart time.Time
	// Average CPU usage in cores or memory usage in bytes.
	Usage ResourceAmount
	// Which resource is this sample for.
	Resource ResourceName
}

// ContainerState stores information about a single container instance.
// It holds the recent history of CPU and memory utilization.
// * CPU is stored in form of a distribution (histogram).
//   Currently we're using fixed weight samples in the CPU histogram (i.e. old
//   and fresh samples are equally important). Old samples are never deleted.
//   TODO: Add exponential decaying of weights over time to address this.
// * Memory is stored for the period of length MemoryAggregationWindowLength in
//   the form of usage peaks, one value per MemoryAggregationInterval.
//   For example if window legth is one week and aggregation interval is one day
//   it will store 7 peaks, one per day, for the last week.
//   Note: samples are added to intervals based on their start timestamps.
type ContainerState struct {
	// Distribution of CPU usage. The measurement unit is 1 CPU core.
	CPUUsage util.Histogram
	// Start of the latest CPU usage sample that was aggregated.
	lastCPUSampleStart time.Time

	// Memory peaks stored in the intervals belonging to the aggregation window
	// (one value per interval). The measurement unit is a byte.
	MemoryUsagePeaks util.FloatSlidingWindow
	// End time of the most recent interval covered by the aggregation window.
	WindowEnd time.Time
	// Start of the latest memory usage sample that was aggregated.
	lastMemorySampleStart time.Time
}

// NewContainerState returns a new, empty ContainerState.
func NewContainerState() *ContainerState {
	return &ContainerState{
		CPUUsage:           util.NewHistogram(CPUHistogramOptions),
		lastCPUSampleStart: time.Unix(0, 0),
		MemoryUsagePeaks: util.NewFloatSlidingWindow(
			int(MemoryAggregationWindowLength / MemoryAggregationInterval)),
		WindowEnd:             time.Unix(0, 0),
		lastMemorySampleStart: time.Unix(0, 0)}
}

func (sample *ContainerUsageSample) isValid(expectedResource ResourceName) bool {
	return sample.Usage >= 0 && sample.Resource == expectedResource
}

func (container *ContainerState) addCPUSample(sample *ContainerUsageSample) bool {
	// Order should not matter for the histogram, other than deduplication.
	// TODO: Timestamp should be used to properly weigh the samples.
	if !sample.isValid(ResourceCPU) || !sample.MeasureStart.After(container.lastCPUSampleStart) {
		return false // Discard invalid, duplicate or out-of-order samples.
	}
	container.CPUUsage.AddSample(CoresFromCPUAmount(sample.Usage), 1.0, sample.MeasureStart)
	container.lastCPUSampleStart = sample.MeasureStart
	return true
}

func (container *ContainerState) addMemorySample(sample *ContainerUsageSample) bool {
	ts := sample.MeasureStart
	if !sample.isValid(ResourceMemory) || !ts.After(container.lastMemorySampleStart) {
		return false // Discard invalid, duplicate or out-of-order samples.
	}
	if !ts.Before(container.WindowEnd.Add(MemoryAggregationWindowLength)) {
		// The gap between this sample and the previous interval is so
		// large that the whole sliding window gets reset.
		// This also happens on the first memory usage sample.
		container.MemoryUsagePeaks.Clear()
		container.WindowEnd = ts.Add(MemoryAggregationInterval)
	} else {
		for !ts.Before(container.WindowEnd) {
			// Shift the memory aggregation window to the next interval.
			container.MemoryUsagePeaks.Push(0.0)
			container.WindowEnd =
				container.WindowEnd.Add(MemoryAggregationInterval)
		}
	}
	// Update the memory peak for the current interval.
	if container.MemoryUsagePeaks.Head() == nil {
		// Window is empty.
		container.MemoryUsagePeaks.Push(0.0)
	}
	*container.MemoryUsagePeaks.Head() = math.Max(
		*container.MemoryUsagePeaks.Head(), BytesFromMemoryAmount(sample.Usage))
	container.lastMemorySampleStart = ts
	return true
}

// AddSample adds a usage sample to the given ContainerState. Requires samples
// for a single resource to be passed in chronological order (i.e. in order of
// growing MeasureStart). Invalid samples (out of order or measure out of legal
// range) are discarded. Returns true if the sample was aggregated, false if it
// was discarded.
// Note: usage samples don't hold their end timestamp / duration. They are
// implicitly assumed to be disjoint when aggregating.
func (container *ContainerState) AddSample(sample *ContainerUsageSample) bool {
	switch sample.Resource {
	case ResourceCPU:
		return container.addCPUSample(sample)
	case ResourceMemory:
		return container.addMemorySample(sample)
	default:
		return false
	}
}
