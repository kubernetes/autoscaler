package base

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

// HistogramOptions describing a histogram with a given number of fixed-size
// buckets, with the first bucket starting at 0.0.
// Requires numBuckets > 0, bucketSize > 0.
type LinearHistogramOptions struct {
	numBuckets int
	bucketSize float64
}

func NewLinearHistogramOptions(
	maxValue float64, bucketSize float64) (*LinearHistogramOptions, error) {
	if maxValue <= 0.0 || bucketSize <= 0.0 {
		return nil, errors.New("maxValue and bucketSize must both be positive")
	} else {
		numBuckets := int(math.Ceil(maxValue / bucketSize))
		return &LinearHistogramOptions{numBuckets, bucketSize}, nil
	}
}

// HistogramOptions describing a histogram with exponentially growing bucket
// boundaries. The first bucket covers the range [0..firstBucketSize).
// Consecutive buckets are of the form [x(n)..x(n) * ratio),
// for n = 1 .. numBuckets - 1.
// Requires ratio > 1, numBuckets > 0, firstBucketSize > 0.
type ExponentialHistogramOptions struct {
	numBuckets int
	firstBucketSize float64
	ratio float64
}

func NewExponentialHistogramOptions(
	maxValue float64, firstBucketSize float64, ratio float64) (*ExponentialHistogramOptions, error) {
	if maxValue <= 0.0 || firstBucketSize <= 0.0 || ratio <= 1.0 {
		return nil, errors.New(
			"maxValue and firstBucketSize must be > 0.0, ratio must be > 1.0")
	} else {
		numBuckets := int(math.Ceil(math.Log(maxValue / firstBucketSize) / math.Log(ratio))) + 1
		return &ExponentialHistogramOptions{numBuckets, firstBucketSize, ratio}, nil
	}
}

func (o *LinearHistogramOptions) NumBuckets() int {
	return o.numBuckets
}

func (o *LinearHistogramOptions) FindBucket(value float64) int {
	bucket := int(value / o.bucketSize)
	if bucket < 0 {
		return 0
	}
	if bucket >= o.numBuckets {
		return o.numBuckets - 1
	}
	return bucket
}

func (o *LinearHistogramOptions) GetBucketStart(bucket int) float64 {
	return float64(bucket) * o.bucketSize
}

func (o *ExponentialHistogramOptions) NumBuckets() int {
	return o.numBuckets
}

func (o *ExponentialHistogramOptions) FindBucket(value float64) int {
	if value < o.firstBucketSize {
		return 0
	}
	bucket := int(math.Log(value / o.firstBucketSize) / math.Log(o.ratio)) + 1
	if bucket >= o.numBuckets {
		return o.numBuckets - 1
	}
	return bucket
}

func (o *ExponentialHistogramOptions) GetBucketStart(bucket int) float64 {
	if bucket == 0 {
		return 0.0
	} else {
		return o.firstBucketSize * math.Pow(o.ratio, float64(bucket - 1))
	}
}

