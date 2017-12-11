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

package util

import (
	"fmt"
	"strings"
)

// Histogram represents an approximate distribution of some variable.
type Histogram interface {
	// Returns an approximation of the given percentile of the distribution.
	// Note: the argument passed to Percentile() is a number between
	// 0 and 1. For example 0.5 corresponds to the median and 0.9 to the
	// 90th percentile.
	// If the histogram is empty, Percentile() returns 0.0.
	Percentile(percentile float64) float64

	// Add a sample with a given value and weight.
	AddSample(value float64, weight float64)

	// Remove a sample with a given value and weight. Note that the total
	// weight of samples with a given value cannot be negative.
	SubtractSample(value float64, weight float64)

	// Add all samples from another histogram. Requires the histograms to be
	// of the exact same type.
	Merge(other *Histogram)

	// Returns true if the histogram is empty.
	IsEmpty() bool

	// Returns a human-readable text description of the histogram.
	String() string
}

// NewHistogram returns a new Histogram instance using given options.
func NewHistogram(options *HistogramOptions) Histogram {
	return &histogram{
		options, make([]float64, (*options).NumBuckets()), 0.0,
		(*options).NumBuckets() - 1, 0}
}

// Simple bucket-based implementation of the Histogram interface. Each bucket
// holds the total weight of samples that belong to it.
// Percentile() returns the middle of the correspodning bucket.
// Resolution (bucket boundaries) of the histogram depends on the options.
// There's no interpolation within buckets (i.e. one sample falls to exactly one
// bucket).
// A bucket is considered empty if its weight is smaller than options.Epsilon().
type histogram struct {
	// Bucketing scheme.
	options *HistogramOptions
	// Cumulative weight of samples in each bucket.
	bucketWeight []float64
	// Total cumulative weight of samples in all buckets.
	totalWeight float64
	// Index of the first non-empty bucket if there's any. Otherwise index
	// of the last bucket.
	minBucket int
	// Index of the last non-empty bucket if there's any. Otherwise 0.
	maxBucket int
}

func (h *histogram) AddSample(value float64, weight float64) {
	if weight < 0.0 {
		panic("sample weight must be non-negative")
	}
	bucket := (*h.options).FindBucket(value)
	h.bucketWeight[bucket] += weight
	h.totalWeight += weight
	if bucket < h.minBucket {
		h.minBucket = bucket
	}
	if bucket > h.maxBucket {
		h.maxBucket = bucket
	}
}

func (h *histogram) SubtractSample(value float64, weight float64) {
	if weight < 0.0 {
		panic("sample weight must be non-negative")
	}
	bucket := (*h.options).FindBucket(value)
	epsilon := (*h.options).Epsilon()
	if weight > h.bucketWeight[bucket]-epsilon {
		weight = h.bucketWeight[bucket]
	}
	h.totalWeight -= weight
	h.bucketWeight[bucket] -= weight
	lastBucket := (*h.options).NumBuckets() - 1
	for h.bucketWeight[h.minBucket] < epsilon && h.minBucket < lastBucket {
		h.minBucket++
	}
	for h.bucketWeight[h.maxBucket] < epsilon && h.maxBucket > 0 {
		h.maxBucket--
	}
}

func (h *histogram) Merge(other *Histogram) {
	o := (*other).(*histogram)
	if h.options != o.options {
		panic("can't merge histograms with different options")
	}
	for bucket := o.minBucket; bucket <= o.maxBucket; bucket++ {
		h.bucketWeight[bucket] += o.bucketWeight[bucket]
	}
	h.totalWeight += o.totalWeight
	if o.minBucket < h.minBucket {
		h.minBucket = o.minBucket
	}
	if o.maxBucket > h.maxBucket {
		h.maxBucket = o.maxBucket
	}
}

func (h *histogram) Percentile(percentile float64) float64 {
	if h.IsEmpty() {
		return 0.0
	}
	partialSum := 0.0
	threshold := percentile * h.totalWeight
	bucket := h.minBucket
	for ; bucket < h.maxBucket; bucket++ {
		partialSum += h.bucketWeight[bucket]
		if partialSum >= threshold {
			break
		}
	}
	bucketStart := (*h.options).GetBucketStart(bucket)
	if bucket < (*h.options).NumBuckets()-1 {
		// Return the middle point between the bucket boundaries.
		bucketEnd := (*h.options).GetBucketStart(bucket + 1)
		return (bucketStart + bucketEnd) / 2.0
	}
	// Return the start of the last bucket (note that the last bucket
	// doesn't have an upper bound).
	return bucketStart
}

func (h *histogram) IsEmpty() bool {
	return h.bucketWeight[h.minBucket] < (*h.options).Epsilon()
}

func (h *histogram) String() string {
	lines := []string{"%-tile\tvalue"}
	for i := 0; i <= 100; i += 5 {
		lines = append(lines, fmt.Sprintf("%d\t%.3f", i, h.Percentile(0.01*float64(i))))
	}
	return strings.Join(lines, "\n")
}

// HistogramsEqual is a helper function for comparing 2 histograms.
func HistogramsEqual(histogram1 Histogram, histogram2 Histogram) bool {
	h1 := (histogram1).(*histogram)
	h2 := (histogram2).(*histogram)
	if h1.options != h2.options || h1.minBucket != h2.minBucket || h1.maxBucket != h2.maxBucket {
		return false
	}
	for bucket := h1.minBucket; bucket <= h1.maxBucket; bucket++ {
		diff := h1.bucketWeight[bucket] - h2.bucketWeight[bucket]
		if diff > 1e-15 || diff < -1e-15 {
			return false
		}
	}
	return true
}
