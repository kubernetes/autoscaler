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
	measureStart time.Time
	// Average CPU usage in cores.
	cpuUsage float64
	// Randomly sampled instant memory usage in bytes.
	memoryUsage float64
}

// ContainerStats stores information about a single container instance.
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
type ContainerStats struct {
	// Distribution of CPU usage. The measurement unit is 1 CPU core.
	cpuUsage util.Histogram
	// Memory peaks stored in the intervals belonging to the aggregation window
	// (one value per interval). The measurement unit is a byte.
	memoryUsagePeaks util.FloatSlidingWindow
	// End time of the most recent interval covered by the aggregation window.
	windowEnd time.Time
	// Start of the latest usage sample that was aggregated.
	lastSampleStart time.Time
}

// NewContainerStats returns a new, empty ContainerStats.
func NewContainerStats() *ContainerStats {
	return &ContainerStats{
		util.NewHistogram(cpuHistogramOptions()), // cpuUsage
		util.NewFloatSlidingWindow( // memoryUsagePeaks
			int(MemoryAggregationWindowLength / MemoryAggregationInterval)),
		time.Unix(0, 0),
		time.Unix(0, 0)}
}

func (sample *ContainerUsageSample) isValid() bool {
	return sample.cpuUsage >= 0.0 && sample.memoryUsage >= 0.0
}

// AddSample adds a usage sample to the given ContainerStats. Requires samples
// to be passed in chronological order (i.e. in order of growing measureStart).
// Invalid samples (out of order or measure out of legal range) are discarded.
// Returns true if the sample was aggregated, false if it was discarded.
// Note: usage samples don't hold their end timestamp / duration. They are
// implicitly assumed to be disjoint when aggregating.
func (container *ContainerStats) AddSample(sample *ContainerUsageSample) bool {
	ts := sample.measureStart
	if !sample.isValid() || !ts.After(container.lastSampleStart) {
		return false // Discard invalid or out-of-order samples.
	}
	if !ts.Before(container.windowEnd.Add(MemoryAggregationWindowLength)) {
		// The gap between this sample and the previous interval is so
		// large that the whole sliding window gets reset.
		// This also happens on the first memory usage sample.
		container.memoryUsagePeaks.Clear()
		container.windowEnd = ts.Add(MemoryAggregationInterval)
	} else {
		for !ts.Before(container.windowEnd) {
			// Shift the memory aggregation window to the next interval.
			container.memoryUsagePeaks.Push(0.0)
			container.windowEnd =
				container.windowEnd.Add(MemoryAggregationInterval)
		}
	}
	// Update the memory peak for the current interval.
	if container.memoryUsagePeaks.Head() == nil {
		// Window is empty.
		container.memoryUsagePeaks.Push(0.0)
	}
	*container.memoryUsagePeaks.Head() = math.Max(
		*container.memoryUsagePeaks.Head(), sample.memoryUsage)
	// Update the CPU usage distribution.
	container.cpuUsage.AddSample(sample.cpuUsage, 1.0)
	container.lastSampleStart = ts
	return true
}
