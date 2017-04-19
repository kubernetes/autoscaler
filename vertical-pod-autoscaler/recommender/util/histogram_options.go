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
	"errors"
	"math"
)

// HistogramOptions define the number and size of buckets of a histogram.
type HistogramOptions interface {
	NumBuckets() int
	// Returns the index of the bucket to which the given value falls.
	// If the value is outside of the range covered by the histogram, it
	// returns the closest bucket (either the first or the last one).
	FindBucket(value float64) int
	// Returns the start of the bucket with a given index. If the index is
	// outside the [0..NumBuckets() - 1] range, the result is undefined.
	GetBucketStart(bucket int) float64
}

// NewLinearHistogramOptions returns HistogramOptions describing a histogram
// with a given number of fixed-size buckets, with the first bucket starting
// at 0.0. Requires maxValue > 0, bucketSize > 0.
func NewLinearHistogramOptions(
	maxValue float64, bucketSize float64) (HistogramOptions, error) {
	if maxValue <= 0.0 || bucketSize <= 0.0 {
		return nil, errors.New("maxValue and bucketSize must both be positive")
	}
	numBuckets := int(math.Ceil(maxValue / bucketSize))
	return &linearHistogramOptions{numBuckets, bucketSize}, nil
}

// NewExponentialHistogramOptions returns HistogramOptions describing a
// histogram with exponentially growing bucket boundaries. The first bucket
// covers the range [0..firstBucketSize). Consecutive buckets are of the form
// [x(n)..x(n) * ratio) for n = 1 .. numBuckets - 1.
// Requires maxValue > 0, firstBucketSize > 0, ratio > 1.
func NewExponentialHistogramOptions(
	maxValue float64, firstBucketSize float64, ratio float64) (HistogramOptions, error) {
	if maxValue <= 0.0 || firstBucketSize <= 0.0 || ratio <= 1.0 {
		return nil, errors.New(
			"maxValue and firstBucketSize must be > 0.0, ratio must be > 1.0")
	}
	numBuckets := int(math.Ceil(math.Log(maxValue/firstBucketSize)/math.Log(ratio))) + 1
	return &exponentialHistogramOptions{numBuckets, firstBucketSize, ratio}, nil
}

type linearHistogramOptions struct {
	numBuckets int
	bucketSize float64
}

type exponentialHistogramOptions struct {
	numBuckets      int
	firstBucketSize float64
	ratio           float64
}

func (o *linearHistogramOptions) NumBuckets() int {
	return o.numBuckets
}

func (o *linearHistogramOptions) FindBucket(value float64) int {
	bucket := int(value / o.bucketSize)
	if bucket < 0 {
		return 0
	}
	if bucket >= o.numBuckets {
		return o.numBuckets - 1
	}
	return bucket
}

func (o *linearHistogramOptions) GetBucketStart(bucket int) float64 {
	return float64(bucket) * o.bucketSize
}

func (o *exponentialHistogramOptions) NumBuckets() int {
	return o.numBuckets
}

func (o *exponentialHistogramOptions) FindBucket(value float64) int {
	if value < o.firstBucketSize {
		return 0
	}
	bucket := int(math.Log(value/o.firstBucketSize)/math.Log(o.ratio)) + 1
	if bucket >= o.numBuckets {
		return o.numBuckets - 1
	}
	return bucket
}

func (o *exponentialHistogramOptions) GetBucketStart(bucket int) float64 {
	if bucket == 0 {
		return 0.0
	}
	return o.firstBucketSize * math.Pow(o.ratio, float64(bucket-1))
}
