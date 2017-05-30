package base

// Histogram represents an approximate distribution of some variable.
type Histogram interface {
	// Returns an approximation of the given percentile of the distribution.
	// Note: the argument passed to Percentile() is a number between 0 and 1.
	// For example 0.5 corresponds to the median and 0.9 to the 90th percentile.
	// If the histogram is empty, Percentile() returns 0.0.
	Percentile(percentile float64) float64

	// Add a sample with a given value and weight. A sample can have negative
	// weight, as long as the total weight of samples with the given value is
	// not negative.
	AddSample(value float64, weight float64)

	// Returns true if the histogram is empty.
	Empty() bool
}

// Simple bucket-based implementation of the histogram. Resolution of the histogram
// depends on the options. Samples are rounded down to the bottom of the bucket.
// There's no interpolation within buckets.
type SimpleHistogram struct {
	options *HistogramOptions	// Bucketing scheme.
	buckets []float64	// Weight of samples in each bucket.
	totalWeight float64	// Weight of samples in all buckets.
	// Index of the first non-empty bucket if there's any. Otherwise index of
	// the last bucket.
	minBucket int
	// Index of the last non-empty bucket if there's any. Otherwise 0.
	maxBucket int
}

func NewHistogram(options HistogramOptions) Histogram {
	return &SimpleHistogram{
		&options, make([]float64, options.NumBuckets()), 0.0, options.NumBuckets() - 1, 0}
}

func (h *SimpleHistogram) AddSample(value float64, weight float64) {
	bucket := (*h.options).FindBucket(value)
	if h.buckets[bucket] + weight <= 0.0 {
		h.clearBucket(bucket)
	} else {
		h.buckets[bucket] += weight
		h.totalWeight += weight
		if bucket < h.minBucket { h.minBucket = bucket }
		if bucket > h.maxBucket { h.maxBucket = bucket }
	}
}

func (h *SimpleHistogram) Percentile(percentile float64) float64 {
	if h.Empty() {
		return 0.0
	}
	partialSum := 0.0
	threshold := percentile * h.totalWeight
	bucket := h.minBucket;
	for ; bucket < h.maxBucket; bucket++ {
		partialSum += h.buckets[bucket]
		if partialSum >= threshold {
			break
		}
	}
	return (*h.options).GetBucketStart(bucket)
}

func (h *SimpleHistogram) Empty() bool {
	return h.totalWeight == 0.0
}

func (h *SimpleHistogram) clearBucket(bucket int) {
	h.totalWeight -= h.buckets[bucket]
	h.buckets[bucket] = 0.0
	lastBucket := (*h.options).NumBuckets() - 1
	for h.buckets[h.minBucket] == 0.0 && h.minBucket < lastBucket {
		h.minBucket += 1
	}
	for h.buckets[h.maxBucket] == 0.0 && h.maxBucket > 0 {
		h.maxBucket -= 1
	}
}

