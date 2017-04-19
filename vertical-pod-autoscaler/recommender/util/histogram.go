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

// Histogram represents an approximate distribution of some variable.
type Histogram interface {
	// Returns an approximation of the given percentile of the distribution.
	// Note: the argument passed to Percentile() is a number between
	// 0 and 1. For example 0.5 corresponds to the median and 0.9 to the
	// 90th percentile.
	// If the histogram is empty, Percentile() returns 0.0.
	Percentile(percentile float64) float64

	// Add a sample with a given value and weight. A sample can have
	// negative weight, as long as the total weight of samples with the
	// given value is not negative.
	AddSample(value float64, weight float64)

	// Returns true if the histogram is empty.
	Empty() bool
}

// NewHistogram returns a new Histogram instance using given options.
func NewHistogram(options HistogramOptions) Histogram {
	return &histogram{
		&options, make([]float64, options.NumBuckets()), 0.0,
		options.NumBuckets() - 1, 0}
}

// Simple bucket-based implementation of the Histogram interface. Samples added
// to the histogram are rounded down to the bucket boundary. Each bucket holds
// the total weight of samples that belong to it.
// Resolution (bucket boundaries) of the histogram depends on the options.
// There's no interpolation within buckets (i.e. one sample falls to exactly one
// bucket).
type histogram struct {
	options     *HistogramOptions // Bucketing scheme.
	buckets     []float64         // Weight of samples in each bucket.
	totalWeight float64           // Weight of samples in all buckets.
	// Index of the first non-empty bucket if there's any. Otherwise index
	// of the last bucket.
	minBucket int
	// Index of the last non-empty bucket if there's any. Otherwise 0.
	maxBucket int
}

func (h *histogram) AddSample(value float64, weight float64) {
	bucket := (*h.options).FindBucket(value)
	if h.buckets[bucket]+weight <= 0.0 {
		h.clearBucket(bucket)
	} else {
		h.buckets[bucket] += weight
		h.totalWeight += weight
		if bucket < h.minBucket {
			h.minBucket = bucket
		}
		if bucket > h.maxBucket {
			h.maxBucket = bucket
		}
	}
}

func (h *histogram) Percentile(percentile float64) float64 {
	if h.Empty() {
		return 0.0
	}
	partialSum := 0.0
	threshold := percentile * h.totalWeight
	bucket := h.minBucket
	for ; bucket < h.maxBucket; bucket++ {
		partialSum += h.buckets[bucket]
		if partialSum >= threshold {
			break
		}
	}
	return (*h.options).GetBucketStart(bucket)
}

func (h *histogram) Empty() bool {
	return h.totalWeight == 0.0
}

func (h *histogram) clearBucket(bucket int) {
	h.totalWeight -= h.buckets[bucket]
	h.buckets[bucket] = 0.0
	lastBucket := (*h.options).NumBuckets() - 1
	for h.buckets[h.minBucket] == 0.0 && h.minBucket < lastBucket {
		h.minBucket++
	}
	for h.buckets[h.maxBucket] == 0.0 && h.maxBucket > 0 {
		h.maxBucket--
	}
}
