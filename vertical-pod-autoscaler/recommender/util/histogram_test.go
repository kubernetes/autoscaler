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

	"github.com/stretchr/testify/assert"
)

// Verifies that Percentile() returns 0.0 when called on an empty histogram for
// any percentile.
func TestPercentilesEmptyHistogram(t *testing.T) {
	options, err := NewLinearHistogramOptions(1.0, 0.1)
	assert.Nil(t, err)
	h := NewHistogram(options)
	for p := -0.5; p <= 1.5; p += 0.5 {
		assert.Equal(t, 0.0, h.Percentile(p))
	}
}

// Verifies that Percentile() returns the correct values of selected
// percentiles on the following histogram: { 1: 1, 2: 2, 3: 3, 4: 4 }.
func TestPercentiles(t *testing.T) {
	options, err := NewLinearHistogramOptions(10.0, 1.0)
	assert.Nil(t, err)
	h := NewHistogram(options)
	for i := 1; i <= 4; i++ {
		h.AddSample(float64(i), float64(i))
	}
	assert.Equal(t, 1.0, h.Percentile(0.0))
	assert.Equal(t, 1.0, h.Percentile(0.1))
	assert.Equal(t, 2.0, h.Percentile(0.2))
	assert.Equal(t, 2.0, h.Percentile(0.3))
	assert.Equal(t, 3.0, h.Percentile(0.4))
	assert.Equal(t, 3.0, h.Percentile(0.5))
	assert.Equal(t, 3.0, h.Percentile(0.6))
	assert.Equal(t, 4.0, h.Percentile(0.7))
	assert.Equal(t, 4.0, h.Percentile(0.8))
	assert.Equal(t, 4.0, h.Percentile(0.9))
	assert.Equal(t, 4.0, h.Percentile(1.0))
}

// Verifies that querying percentile < 0.0 returns the minimum value in the
// histogram, while querying percentile > 1.0 returns the maximum of the
// histogram.
func TestPercentileOutOfBounds(t *testing.T) {
	options, err := NewLinearHistogramOptions(1.0, 0.1)
	assert.Nil(t, err)
	h := NewHistogram(options)
	assert.Nil(t, err)
	h.AddSample(0.1, 0.1)
	h.AddSample(0.2, 0.2)

	assert.Equal(t, 0.1, h.Percentile(-0.1))
	assert.Equal(t, 0.2, h.Percentile(1.1))
}

// Verifies that Empty() returns true on an empty histogram and false otherwise.
func TestEmptyHistogram(t *testing.T) {
	options, err := NewLinearHistogramOptions(1.0, 0.1)
	assert.Nil(t, err)
	h := NewHistogram(options)
	assert.Nil(t, err)
	assert.True(t, h.Empty())
	h.AddSample(0.1, 1.0) // Add a sample.
	assert.False(t, h.Empty())
	h.AddSample(0.1, -0.5) // Remove part of a sample.
	assert.False(t, h.Empty())
	h.AddSample(0.1, -0.5) // Remove the remaining part of the sample.
	assert.True(t, h.Empty())
}
