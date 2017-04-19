package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test all methods of LinearHistogramOptions using a sample bucketing scheme.
func TestLinearHistogramOptions(t *testing.T) {
	o, err := NewLinearHistogramOptions(5.0, 0.3)
	assert.Nil(t, err)
	assert.Equal(t, 17, o.NumBuckets())

	assert.Equal(t, 0.0, o.GetBucketStart(0))
	assert.Equal(t, 4.8, o.GetBucketStart(16))

	assert.Equal(t, 0, o.FindBucket(-1.0))
	assert.Equal(t, 0, o.FindBucket(0.0))
	assert.Equal(t, 4, o.FindBucket(1.3))
	assert.Equal(t, 16, o.FindBucket(100.0))
}

// Test all methods of ExponentialHistogramOptions using a sample bucketing scheme.
func TestExponentialHistogramOptions(t *testing.T) {
	o, err := NewExponentialHistogramOptions(100.0, 10.0, 2.0)
	assert.Nil(t, err)
	assert.Equal(t, 5, o.NumBuckets())

	assert.Equal(t, 0.0, o.GetBucketStart(0))
	assert.Equal(t, 10.0, o.GetBucketStart(1))
	assert.Equal(t, 20.0, o.GetBucketStart(2))
	assert.Equal(t, 40.0, o.GetBucketStart(3))
	assert.Equal(t, 80.0, o.GetBucketStart(4))

	assert.Equal(t, 0, o.FindBucket(-1.0))
	assert.Equal(t, 0, o.FindBucket(9.99))
	assert.Equal(t, 1, o.FindBucket(10.0))
	assert.Equal(t, 2, o.FindBucket(20.0))
	assert.Equal(t, 4, o.FindBucket(100.0))
}

