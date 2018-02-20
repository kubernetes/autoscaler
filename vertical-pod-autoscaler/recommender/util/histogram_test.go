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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	// Minimum precision of histogram values (relative).
	valueEpsilon = 1e-15
	// Minimum precision of histogram weights (absolute).
	weightEpsilon = 1e-15
	// Arbitrary timestamp.
	anyTime = time.Time{}
	// Test histogram options.
	testHistogramOptions, _ = NewLinearHistogramOptions(10.0, 1.0, weightEpsilon)
)

// Verifies that Percentile() returns 0.0 when called on an empty histogram for
// any percentile.
func TestPercentilesEmptyHistogram(t *testing.T) {
	h := NewHistogram(testHistogramOptions)
	for p := -0.5; p <= 1.5; p += 0.5 {
		assert.Equal(t, 0.0, h.Percentile(p))
	}
}

// Verifies that Percentile() returns the correct values of selected
// percentiles on the following histogram: { 1: 1, 2: 2, 3: 3, 4: 4 }.
func TestPercentiles(t *testing.T) {
	h := NewHistogram(testHistogramOptions)
	for i := 1; i <= 4; i++ {
		h.AddSample(float64(i), float64(i), anyTime)
	}
	assert.InEpsilon(t, 1.5, h.Percentile(0.0), valueEpsilon)
	assert.InEpsilon(t, 1.5, h.Percentile(0.1), valueEpsilon)
	assert.InEpsilon(t, 2.5, h.Percentile(0.2), valueEpsilon)
	assert.InEpsilon(t, 2.5, h.Percentile(0.3), valueEpsilon)
	assert.InEpsilon(t, 3.5, h.Percentile(0.4), valueEpsilon)
	assert.InEpsilon(t, 3.5, h.Percentile(0.5), valueEpsilon)
	assert.InEpsilon(t, 3.5, h.Percentile(0.6), valueEpsilon)
	assert.InEpsilon(t, 4.5, h.Percentile(0.7), valueEpsilon)
	assert.InEpsilon(t, 4.5, h.Percentile(0.8), valueEpsilon)
	assert.InEpsilon(t, 4.5, h.Percentile(0.9), valueEpsilon)
	assert.InEpsilon(t, 4.5, h.Percentile(1.0), valueEpsilon)
}

// Verifies that querying percentile < 0.0 returns the minimum value in the
// histogram, while querying percentile > 1.0 returns the maximum of the
// histogram.
func TestPercentileOutOfBounds(t *testing.T) {
	options, err := NewLinearHistogramOptions(1.0, 0.1, weightEpsilon)
	assert.Nil(t, err)
	h := NewHistogram(options)
	assert.Nil(t, err)
	h.AddSample(0.1, 0.1, anyTime)
	h.AddSample(0.2, 0.2, anyTime)

	assert.InEpsilon(t, 0.15, h.Percentile(-0.1), valueEpsilon)
	assert.InEpsilon(t, 0.25, h.Percentile(1.1), valueEpsilon)

	// Fill the boundary buckets.
	h.AddSample(0.0, 0.1, anyTime)
	h.AddSample(1.0, 0.2, anyTime)
	assert.InEpsilon(t, 0.05, h.Percentile(-0.1), valueEpsilon)
	assert.InEpsilon(t, 1.0, h.Percentile(1.1), valueEpsilon)
}

// Verifies that IsEmpty() returns true on an empty histogram and false otherwise.
func TestEmptyHistogram(t *testing.T) {
	options, err := NewLinearHistogramOptions(1.0, 0.1, weightEpsilon)
	assert.Nil(t, err)
	h := NewHistogram(options)
	assert.Nil(t, err)
	assert.True(t, h.IsEmpty())
	h.AddSample(0.1, weightEpsilon*2.5, anyTime) // Sample weight = epsilon * 2.5.
	assert.False(t, h.IsEmpty())
	h.SubtractSample(0.1, weightEpsilon, anyTime) // Sample weight = epsilon * 1.5.
	assert.False(t, h.IsEmpty())
	h.SubtractSample(0.1, weightEpsilon, anyTime) // Sample weight = epsilon * 0.5.
	assert.True(t, h.IsEmpty())
}

// Verifies that Merge() works as expected on two sample histograms.
func TestHistogramMerge(t *testing.T) {
	h1 := NewHistogram(testHistogramOptions)
	h1.AddSample(1, 1, anyTime)
	h1.AddSample(2, 1, anyTime)

	h2 := NewHistogram(testHistogramOptions)
	h2.AddSample(2, 1, anyTime)
	h2.AddSample(3, 1, anyTime)

	expected := NewHistogram(testHistogramOptions)
	expected.AddSample(1, 1, anyTime)
	expected.AddSample(2, 1, anyTime)
	expected.AddSample(2, 1, anyTime)
	expected.AddSample(3, 1, anyTime)

	h1.Merge(h2)
	assert.True(t, h1.Equals(expected))
}
