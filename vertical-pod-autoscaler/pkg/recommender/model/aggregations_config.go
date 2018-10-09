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
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

var (
	// MemoryAggregationWindowLength is the length of the memory usage history
	// aggregated by VPA, which is 8 days.
	MemoryAggregationWindowLength = time.Hour * 8 * 24
	// MemoryAggregationInterval is the length of a single interval, for
	// which the peak memory usage is computed.
	// Memory usage peaks are aggregated in daily intervals. In other words
	// there is one memory usage sample per day (the maximum usage over that
	// day).
	// Note: AggregationWindowLength must be integrally divisible by this value.
	MemoryAggregationInterval = time.Hour * 24
	// CPUHistogramOptions are options to be used by histograms that store
	// CPU measures expressed in cores.
	CPUHistogramOptions = cpuHistogramOptions()
	// MemoryHistogramOptions are options to be used by histograms that
	// store memory measures expressed in bytes.
	MemoryHistogramOptions = memoryHistogramOptions()
	// HistogramBucketSizeGrowth defines the growth rate of the histogram buckets.
	// Each bucket is wider than the previous one by this fraction.
	HistogramBucketSizeGrowth = 0.05 // Make each bucket 5% larger than the previous one.
	// MemoryHistogramDecayHalfLife is the amount of time it takes a historical
	// memory usage sample to lose half of its weight. In other words, a fresh
	// usage sample is twice as 'important' as one with age equal to the half
	// life period.
	MemoryHistogramDecayHalfLife = time.Hour * 24
	// CPUHistogramDecayHalfLife is the amount of time it takes a historical
	// CPU usage sample to lose half of its weight.
	CPUHistogramDecayHalfLife = time.Hour * 24
)

const (
	// minSampleWeight is the minimal weight of any sample (prior to including decaying factor)
	minSampleWeight = 0.1
	// epsilon is the minimal weight kept in histograms, it should be small enough that old samples
	// (just inside MemoryAggregationWindowLength) added with minSampleWeight are still kept
	epsilon = 0.001 * minSampleWeight
)

func cpuHistogramOptions() util.HistogramOptions {
	// CPU histograms use exponential bucketing scheme with the smallest bucket
	// size of 0.01 core, max of 1000.0 cores and the relative error of HistogramRelativeError.
	//
	// When parameters below are changed SupportedCheckpointVersion has to be bumped.
	options, err := util.NewExponentialHistogramOptions(1000.0, 0.01, 1.+HistogramBucketSizeGrowth, epsilon)
	if err != nil {
		panic("Invalid CPU histogram options") // Should not happen.
	}
	return options
}

func memoryHistogramOptions() util.HistogramOptions {
	// Memory histograms use exponential bucketing scheme with the smallest
	// bucket size of 10MB, max of 1TB and the relative error of HistogramRelativeError.
	//
	// When parameters below are changed SupportedCheckpointVersion has to be bumped.
	options, err := util.NewExponentialHistogramOptions(1e12, 1e7, 1.+HistogramBucketSizeGrowth, epsilon)
	if err != nil {
		panic("Invalid memory histogram options") // Should not happen.
	}
	return options
}
