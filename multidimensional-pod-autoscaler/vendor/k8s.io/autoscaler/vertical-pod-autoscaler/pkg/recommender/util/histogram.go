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
	"time"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

const (
	// MaxCheckpointWeight is the maximum weight that can be stored in
	// HistogramCheckpoint in a single bucket
	MaxCheckpointWeight uint32 = 10000
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
	AddSample(value float64, weight float64, time time.Time)

	// Remove a sample with a given value and weight. Note that the total
	// weight of samples with a given value cannot be negative.
	SubtractSample(value float64, weight float64, time time.Time)

	// Add all samples from another histogram. Requires the histograms to be
	// of the exact same type.
	Merge(other Histogram)

	// Returns true if the histogram is empty.
	IsEmpty() bool

	// Returns true if the histogram is equal to another one. The two
	// histograms must use the same HistogramOptions object (not two
	// different copies).
	// If the two histograms are not of the same runtime type returns false.
	Equals(other Histogram) bool

	// Returns a human-readable text description of the histogram.
	String() string

	// SaveToChekpoint returns a representation of the histogram as a
	// HistogramCheckpoint. During conversion buckets with small weights
	// can be omitted.
	SaveToChekpoint() (*vpa_types.HistogramCheckpoint, error)

	// LoadFromCheckpoint loads data from the checkpoint into the histogram
	// by appending samples.
	LoadFromCheckpoint(*vpa_types.HistogramCheckpoint) error
}

// NewHistogram returns a new Histogram instance using given options.
func NewHistogram(options HistogramOptions) Histogram {
	return &histogram{
		options:      options,
		bucketWeight: make([]float64, options.NumBuckets()),
		totalWeight:  0.0,
		minBucket:    options.NumBuckets() - 1,
		maxBucket:    0}
}

// Simple bucket-based implementation of the Histogram interface. Each bucket
// holds the total weight of samples that belong to it.
// Percentile() returns the upper bound of the corresponding bucket.
// Resolution (bucket boundaries) of the histogram depends on the options.
// There's no interpolation within buckets (i.e. one sample falls to exactly one
// bucket).
// A bucket is considered empty if its weight is smaller than options.Epsilon().
type histogram struct {
	// Bucketing scheme.
	options HistogramOptions
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

func (h *histogram) AddSample(value float64, weight float64, time time.Time) {
	if weight < 0.0 {
		panic("sample weight must be non-negative")
	}
	bucket := h.options.FindBucket(value)
	h.bucketWeight[bucket] += weight
	h.totalWeight += weight
	if bucket < h.minBucket && h.bucketWeight[bucket] >= h.options.Epsilon() {
		h.minBucket = bucket
	}
	if bucket > h.maxBucket && h.bucketWeight[bucket] >= h.options.Epsilon() {
		h.maxBucket = bucket
	}
}

func safeSubtract(value, sub, epsilon float64) float64 {
	value -= sub
	if value < epsilon {
		return 0.0
	}
	return value
}

func (h *histogram) SubtractSample(value float64, weight float64, time time.Time) {
	if weight < 0.0 {
		panic("sample weight must be non-negative")
	}
	bucket := h.options.FindBucket(value)
	epsilon := h.options.Epsilon()

	h.totalWeight = safeSubtract(h.totalWeight, weight, epsilon)
	h.bucketWeight[bucket] = safeSubtract(h.bucketWeight[bucket], weight, epsilon)

	h.updateMinAndMaxBucket()
}

func (h *histogram) Merge(other Histogram) {
	o := other.(*histogram)
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
	if bucket < h.options.NumBuckets()-1 {
		// Return the end of the bucket.
		return h.options.GetBucketStart(bucket + 1)
	}
	// Return the start of the last bucket (note that the last bucket
	// doesn't have an upper bound).
	return h.options.GetBucketStart(bucket)
}

func (h *histogram) IsEmpty() bool {
	return h.bucketWeight[h.minBucket] < h.options.Epsilon()
}

func (h *histogram) String() string {
	lines := []string{
		fmt.Sprintf("minBucket: %d, maxBucket: %d, totalWeight: %.3f",
			h.minBucket, h.maxBucket, h.totalWeight),
		"%-tile\tvalue",
	}
	for i := 0; i <= 100; i += 5 {
		lines = append(lines, fmt.Sprintf("%d\t%.3f", i, h.Percentile(0.01*float64(i))))
	}
	return strings.Join(lines, "\n")
}

func (h *histogram) Equals(other Histogram) bool {
	h2, typesMatch := other.(*histogram)
	if !typesMatch || h.options != h2.options || h.minBucket != h2.minBucket || h.maxBucket != h2.maxBucket {
		return false
	}
	for bucket := h.minBucket; bucket <= h.maxBucket; bucket++ {
		diff := h.bucketWeight[bucket] - h2.bucketWeight[bucket]
		if diff > 1e-15 || diff < -1e-15 {
			return false
		}
	}
	return true
}

// Adjusts the value of minBucket and maxBucket after any operation that
// decreases weights.
func (h *histogram) updateMinAndMaxBucket() {
	epsilon := h.options.Epsilon()
	lastBucket := h.options.NumBuckets() - 1
	for h.bucketWeight[h.minBucket] < epsilon && h.minBucket < lastBucket {
		h.minBucket++
	}
	for h.bucketWeight[h.maxBucket] < epsilon && h.maxBucket > 0 {
		h.maxBucket--
	}
}

func (h *histogram) SaveToChekpoint() (*vpa_types.HistogramCheckpoint, error) {
	result := vpa_types.HistogramCheckpoint{
		BucketWeights: make(map[int]uint32),
	}
	result.TotalWeight = h.totalWeight
	// Find max
	max := 0.
	for bucket := h.minBucket; bucket <= h.maxBucket; bucket++ {
		if h.bucketWeight[bucket] > max {
			max = h.bucketWeight[bucket]
		}
	}
	// Compute ratio
	ratio := float64(MaxCheckpointWeight) / max
	// Convert weights and drop near-zero weights
	for bucket := h.minBucket; bucket <= h.maxBucket; bucket++ {
		newWeight := uint32(round(h.bucketWeight[bucket] * ratio))
		if newWeight > 0 {
			result.BucketWeights[bucket] = newWeight
		}
	}

	return &result, nil
}

func (h *histogram) LoadFromCheckpoint(checkpoint *vpa_types.HistogramCheckpoint) error {
	if checkpoint == nil {
		return fmt.Errorf("Cannot load from empty checkpoint")
	}
	if checkpoint.TotalWeight < 0.0 {
		return fmt.Errorf("Cannot load checkpoint with negative weight %v", checkpoint.TotalWeight)
	}
	sum := int64(0)
	for bucket, weight := range checkpoint.BucketWeights {
		sum += int64(weight)
		if bucket >= h.options.NumBuckets() {
			return fmt.Errorf("Checkpoint has bucket %v that is exceeding histogram buckets %v", bucket, h.options.NumBuckets())
		}
		if bucket < 0 {
			return fmt.Errorf("Checkpoint has a negative bucket %v", bucket)
		}
	}
	if sum == 0 {
		return nil
	}
	ratio := checkpoint.TotalWeight / float64(sum)
	for bucket, weight := range checkpoint.BucketWeights {
		if bucket < h.minBucket {
			h.minBucket = bucket
		}
		if bucket > h.maxBucket {
			h.maxBucket = bucket
		}
		h.bucketWeight[bucket] += float64(weight) * ratio
	}
	h.totalWeight += checkpoint.TotalWeight

	return nil
}

// Multiplies all weights by a given factor. The factor must be non-negative.
// (note: this operation does not affect the percentiles of the distribution)
func (h *histogram) scale(factor float64) {
	if factor < 0.0 {
		panic("scale factor must be non-negative")
	}
	for bucket := h.minBucket; bucket <= h.maxBucket; bucket++ {
		h.bucketWeight[bucket] *= factor
	}
	h.totalWeight *= factor
	// Some buckets might become empty (weight < epsilon), so adjust min and max buckets.
	h.updateMinAndMaxBucket()
}
