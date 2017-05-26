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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/util"
)

var (
	TimeLayout = "2006-01-02 15:04:05"
)

func newUsageSample(timestamp time.Time, cpuUsage float64, memoryUsage float64) *ContainerUsageSample {
	return &ContainerUsageSample{timestamp, cpuUsage, memoryUsage}
}

// Add 6 usage samples (3 valid, 3 invalid) to a container. Verifies that for
// valid samples the CPU measures are aggregated in the CPU histogram and
// the memory measures are aggregated in the memory peaks sliding window.
// Verifies that invalid samples (out-of-order or negative usage) are ignored.
func TestAggregateContainerUsageSamples(t *testing.T) {
	testTimestamp, err := time.Parse(TimeLayout, "2017-04-18 17:35:05")
	assert.Nil(t, err)
	mockCPUHistogram := new(util.MockHistogram)
	memoryUsagePeaks := util.NewFloatSlidingWindow(
		int(MemoryAggregationWindowLength / MemoryAggregationInterval))
	c := &ContainerStats{
		mockCPUHistogram,
		memoryUsagePeaks,
		time.Unix(0, 0),
		time.Unix(0, 0)}

	// Verify that a CPU measures are added to the CPU histogram.
	mockCPUHistogram.On("AddSample", 3.14, 1.0)
	mockCPUHistogram.On("AddSample", 6.28, 1.0)
	mockCPUHistogram.On("AddSample", 1.57, 1.0)

	// Add three usage samples.
	assert.True(t, c.AddSample(newUsageSample(
		testTimestamp, 3.14, 5.0)))
	assert.True(t, c.AddSample(newUsageSample(
		testTimestamp.Add(MemoryAggregationInterval/2), 6.28, 10.0)))
	assert.True(t, c.AddSample(newUsageSample(
		testTimestamp.Add(MemoryAggregationInterval), 1.57, 2.5)))

	// Discard invalid samples.
	assert.False(t, c.AddSample(newUsageSample( // Out of order sample.
		testTimestamp.Add(MemoryAggregationInterval), 1.0, 1.0)))
	assert.False(t, c.AddSample(newUsageSample( // Negative CPU usage.
		testTimestamp.Add(MemoryAggregationInterval*2), -1.0, 1.0)))
	assert.False(t, c.AddSample(newUsageSample( // Negative memory usage.
		testTimestamp.Add(MemoryAggregationInterval*2), 1.0, -1.0)))

	// Verify that memory peak samples were aggregated properly.
	assert.Equal(t, []float64{10.0, 2.5}, memoryUsagePeaks.Contents())
}
