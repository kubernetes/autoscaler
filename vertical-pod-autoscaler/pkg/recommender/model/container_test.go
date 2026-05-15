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

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

var (
	timeLayout       = "2006-01-02 15:04:05"
	testTimestamp, _ = time.Parse(timeLayout, "2017-04-18 17:35:05")

	TestRequest = Resources{
		ResourceCPU:    CPUAmountFromCores(2.3),
		ResourceMemory: MemoryAmountFromBytes(5e8),
	}
)

const (
	kb = 1024
	mb = 1024 * kb
)

func newUsageSample(timestamp time.Time, usage int64, resource ResourceName) *ContainerUsageSample {
	return &ContainerUsageSample{
		MeasureStart: timestamp,
		Usage:        ResourceAmount(usage),
		Resource:     resource,
	}
}

type ContainerTest struct {
	mockCPUHistogram        *util.MockHistogram
	mockMemoryHistogram     *util.MockHistogram
	aggregateContainerState *AggregateContainerState
	container               *ContainerState
}

func newContainerTest() ContainerTest {
	mockCPUHistogram := new(util.MockHistogram)
	mockMemoryHistogram := new(util.MockHistogram)
	aggregateContainerState := &AggregateContainerState{
		AggregateCPUUsage:                 mockCPUHistogram,
		AggregateMemoryPeaks:              mockMemoryHistogram,
		OOMBumpUpRatio:                    1.2,                                                       // Default value, can be adjusted as needed
		OOMMinBumpUp:                      1.048576e+08,                                              // Default value (100Mi), can be adjusted as needed
		MemoryAggregationIntervalDuration: GetAggregationsConfig().MemoryAggregationIntervalDuration, // Default value, can be adjusted as needed
	}
	container := &ContainerState{
		Request:    TestRequest,
		aggregator: aggregateContainerState,
	}
	return ContainerTest{
		mockCPUHistogram:        mockCPUHistogram,
		mockMemoryHistogram:     mockMemoryHistogram,
		aggregateContainerState: aggregateContainerState,
		container:               container,
	}
}

func newContainerTestWithCustomMemoryInterval(interval time.Duration) ContainerTest {
	test := newContainerTest()
	test.aggregateContainerState.MemoryAggregationIntervalDuration = interval
	return test
}

func newContainerTestWithCustomMemoryIntervalCount(count int64) ContainerTest {
	test := newContainerTest()
	test.aggregateContainerState.MemoryAggregationIntervalCount = count
	return test
}

// Add 6 usage samples (3 valid, 3 invalid) to a container. Verifies that for
// valid samples the CPU measures are aggregated in the CPU histogram and
// the memory measures are aggregated in the memory peaks sliding window.
// Verifies that invalid samples (out-of-order or negative usage) are ignored.
func TestAggregateContainerUsageSamples(t *testing.T) {
	test := newContainerTest()
	c := test.container
	memoryAggregationInterval := GetAggregationsConfig().MemoryAggregationIntervalDuration
	// Verify that CPU measures are added to the CPU histogram.
	// The weight should be equal to the current request.
	timeStep := memoryAggregationInterval / 2
	test.mockCPUHistogram.On("AddSample", 3.14, minSampleWeight, testTimestamp)
	test.mockCPUHistogram.On("AddSample", 6.28, minSampleWeight, testTimestamp.Add(timeStep))
	test.mockCPUHistogram.On("AddSample", 1.57, minSampleWeight, testTimestamp.Add(2*timeStep))
	// Verify that memory peaks are added to the memory peaks histogram.
	memoryAggregationWindowEnd := testTimestamp.Add(memoryAggregationInterval)
	test.mockMemoryHistogram.On("AddSample", 5.0, 1.0, memoryAggregationWindowEnd)
	test.mockMemoryHistogram.On("SubtractSample", 5.0, 1.0, memoryAggregationWindowEnd)
	test.mockMemoryHistogram.On("AddSample", 10.0, 1.0, memoryAggregationWindowEnd)
	memoryAggregationWindowEnd = memoryAggregationWindowEnd.Add(memoryAggregationInterval)
	test.mockMemoryHistogram.On("AddSample", 2.0, 1.0, memoryAggregationWindowEnd)

	// Add three CPU and memory usage samples.
	assert.True(t, c.AddSample(newUsageSample(
		testTimestamp, 3140, ResourceCPU)))
	assert.True(t, c.AddSample(newUsageSample(
		testTimestamp, 5, ResourceMemory)))

	assert.True(t, c.AddSample(newUsageSample(
		testTimestamp.Add(timeStep), 6280, ResourceCPU)))
	assert.True(t, c.AddSample(newUsageSample(
		testTimestamp.Add(timeStep), 10, ResourceMemory)))

	assert.True(t, c.AddSample(newUsageSample(
		testTimestamp.Add(2*timeStep), 1570, ResourceCPU)))
	assert.True(t, c.AddSample(newUsageSample(
		testTimestamp.Add(2*timeStep), 2, ResourceMemory)))

	// Discard invalid samples.
	assert.False(t, c.AddSample(newUsageSample( // Out of order sample.
		testTimestamp.Add(2*timeStep), 1000, ResourceCPU)))
	assert.False(t, c.AddSample(newUsageSample( // Negative CPU usage.
		testTimestamp.Add(4*timeStep), -1000, ResourceCPU)))
	assert.False(t, c.AddSample(newUsageSample( // Negative memory usage.
		testTimestamp.Add(4*timeStep), -1000, ResourceMemory)))
}

func TestRecordOOMIncreasedByBumpUp(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationIntervalDuration)
	// Bump Up factor is 20%.
	test.mockMemoryHistogram.On("AddSample", 1200.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb)))
}

func TestRecordOOMDontRunAway(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationIntervalDuration)

	// Bump Up factor is 20%.
	test.mockMemoryHistogram.On("AddSample", 1200.0*mb, 1.0, memoryAggregationWindowEnd)
	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb)))

	// new smaller OOMs don't influence the sample value (oomPeak)
	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(999*mb)))
	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(999*mb)))

	test.mockMemoryHistogram.On("SubtractSample", 1200.0*mb, 1.0, memoryAggregationWindowEnd)
	test.mockMemoryHistogram.On("AddSample", 2400.0*mb, 1.0, memoryAggregationWindowEnd)
	// a larger OOM should increase the sample value
	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(2000*mb)))
}

func TestRecordOOMIncreasedByMin(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationIntervalDuration)
	// Min grow by 100Mb.
	test.mockMemoryHistogram.On("AddSample", 101.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1*mb)))
}

func TestRecordOOMMaxedWithKnownSample(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationIntervalDuration)

	test.mockMemoryHistogram.On("AddSample", 3000.0*mb, 1.0, memoryAggregationWindowEnd)
	assert.True(t, test.container.AddSample(newUsageSample(testTimestamp, 3000*mb, ResourceMemory)))

	// Last known sample is higher than request, so it is taken.
	test.mockMemoryHistogram.On("SubtractSample", 3000.0*mb, 1.0, memoryAggregationWindowEnd)
	test.mockMemoryHistogram.On("AddSample", 3600.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb)))
}

func TestRecordOOMDiscardsOldSample(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationIntervalDuration)

	test.mockMemoryHistogram.On("AddSample", 1000.0*mb, 1.0, memoryAggregationWindowEnd)
	assert.True(t, test.container.AddSample(newUsageSample(testTimestamp, 1000*mb, ResourceMemory)))

	// OOM is stale, mem not changed.
	assert.Error(t, test.container.RecordOOM(testTimestamp.Add(-30*time.Hour), ResourceAmount(1000*mb)))
}

func TestRecordOOMInNewWindow(t *testing.T) {
	test := newContainerTest()
	memoryAggregationInterval := GetAggregationsConfig().MemoryAggregationIntervalDuration
	memoryAggregationWindowEnd := testTimestamp.Add(memoryAggregationInterval)

	test.mockMemoryHistogram.On("AddSample", 2000.0*mb, 1.0, memoryAggregationWindowEnd)
	assert.True(t, test.container.AddSample(newUsageSample(testTimestamp, 2000*mb, ResourceMemory)))

	memoryAggregationWindowEnd = memoryAggregationWindowEnd.Add(2 * memoryAggregationInterval)
	test.mockMemoryHistogram.On("AddSample", 2400.0*mb, 1.0, memoryAggregationWindowEnd)
	assert.NoError(t, test.container.RecordOOM(testTimestamp.Add(2*memoryAggregationInterval), ResourceAmount(1000*mb)))
}

// Tests with custom MemoryAggregationInterval to verify the per-VPA
// MemoryAggregationIntervalSeconds setting affects container behavior.

// Verifies that with a custom 1h interval, memory samples 2h apart land in
// different aggregation windows (each gets its own peak), unlike the default
// 24h interval where they would share the same window.
func TestMemorySamplesWithCustomAggregationInterval(t *testing.T) {
	customInterval := 1 * time.Hour
	test := newContainerTestWithCustomMemoryInterval(customInterval)
	c := test.container

	windowEnd1 := testTimestamp.Add(customInterval)
	// First sample opens the first window.
	test.mockMemoryHistogram.On("AddSample", 5.0, 1.0, windowEnd1)
	assert.True(t, c.AddSample(newUsageSample(testTimestamp, 5, ResourceMemory)))

	// Second sample 2h later falls outside the 1h window, opening a new one.
	windowEnd2 := windowEnd1.Add(2 * customInterval)
	test.mockMemoryHistogram.On("AddSample", 10.0, 1.0, windowEnd2)
	assert.True(t, c.AddSample(newUsageSample(testTimestamp.Add(2*customInterval), 10, ResourceMemory)))

	test.mockCPUHistogram.AssertExpectations(t)
	test.mockMemoryHistogram.AssertExpectations(t)
}

// Verifies that with a custom 1h interval, a memory sample at half-interval
// (30min) stays within the same window and updates the peak via subtract+add.
func TestMemoryPeakUpdateWithinCustomInterval(t *testing.T) {
	customInterval := 1 * time.Hour
	test := newContainerTestWithCustomMemoryInterval(customInterval)
	c := test.container

	windowEnd := testTimestamp.Add(customInterval)

	// First sample sets the initial peak.
	test.mockMemoryHistogram.On("AddSample", 5.0, 1.0, windowEnd)
	assert.True(t, c.AddSample(newUsageSample(testTimestamp, 5, ResourceMemory)))

	// Second sample 30min later is still within the 1h window and has a higher
	// value, so the old peak is subtracted and the new one is added.
	test.mockMemoryHistogram.On("SubtractSample", 5.0, 1.0, windowEnd)
	test.mockMemoryHistogram.On("AddSample", 10.0, 1.0, windowEnd)
	assert.True(t, c.AddSample(newUsageSample(testTimestamp.Add(30*time.Minute), 10, ResourceMemory)))

	test.mockMemoryHistogram.AssertExpectations(t)
}

// Verifies that with a custom 1h interval, an OOM event 2h in the past is
// discarded as stale. With the default 24h interval the same OOM would be accepted.
func TestRecordOOMDiscardsOldSampleWithCustomInterval(t *testing.T) {
	customInterval := 1 * time.Hour
	test := newContainerTestWithCustomMemoryInterval(customInterval)
	c := test.container

	windowEnd := testTimestamp.Add(customInterval)
	test.mockMemoryHistogram.On("AddSample", 1000.0*mb, 1.0, windowEnd)
	assert.True(t, c.AddSample(newUsageSample(testTimestamp, 1000*mb, ResourceMemory)))

	// OOM 2h before testTimestamp is older than the 1h interval, so it should be discarded.
	assert.Error(t, c.RecordOOM(testTimestamp.Add(-2*time.Hour), ResourceAmount(1000*mb)))

	test.mockMemoryHistogram.AssertExpectations(t)
}

// Verifies that with a custom 1h interval, an OOM in a new window shifts
// WindowEnd by the custom interval rather than the default 24h.
func TestRecordOOMInNewWindowWithCustomInterval(t *testing.T) {
	customInterval := 1 * time.Hour
	test := newContainerTestWithCustomMemoryInterval(customInterval)
	c := test.container

	windowEnd := testTimestamp.Add(customInterval)
	test.mockMemoryHistogram.On("AddSample", 2000.0*mb, 1.0, windowEnd)
	assert.True(t, c.AddSample(newUsageSample(testTimestamp, 2000*mb, ResourceMemory)))

	// OOM 2 intervals later triggers a window shift by the custom interval.
	newWindowEnd := windowEnd.Add(2 * customInterval)
	test.mockMemoryHistogram.On("AddSample", 2400.0*mb, 1.0, newWindowEnd)
	assert.NoError(t, c.RecordOOM(testTimestamp.Add(2*customInterval), ResourceAmount(1000*mb)))

	test.mockMemoryHistogram.AssertExpectations(t)
}

// Tests with custom MemoryAggregationIntervalCount to verify the per-VPA
// MemoryAggregationIntervalCount setting affects container behavior.

// Verifies that with a custom count of 2 (instead of default 8), the OOM
// staleness window shrinks. An OOM event at 3 intervals ago is discarded
// with count=2 (window=2*interval) but would be accepted with default count=8.
func TestRecordOOMDiscardsOldSampleWithCustomIntervalCount(t *testing.T) {
	test := newContainerTestWithCustomMemoryIntervalCount(2)
	c := test.container

	interval := GetAggregationsConfig().MemoryAggregationIntervalDuration
	windowEnd := testTimestamp.Add(interval)
	test.mockMemoryHistogram.On("AddSample", 1000.0*mb, 1.0, windowEnd)
	assert.True(t, c.AddSample(newUsageSample(testTimestamp, 1000*mb, ResourceMemory)))

	// OOM 3 intervals before testTimestamp: with count=2 the staleness check
	// (WindowEnd - interval) is only 1 interval before WindowEnd, so the OOM
	// at 3 intervals ago is stale and discarded.
	assert.Error(t, c.RecordOOM(testTimestamp.Add(-3*interval), ResourceAmount(1000*mb)))

	test.mockMemoryHistogram.AssertExpectations(t)
}

// Verifies that with a large custom count, memory samples spread across many
// intervals still produce separate peaks. This confirms the count doesn't
// interfere with per-interval windowing behavior.
func TestMemorySamplesWithCustomAggregationIntervalCount(t *testing.T) {
	test := newContainerTestWithCustomMemoryIntervalCount(16)
	c := test.container

	interval := GetAggregationsConfig().MemoryAggregationIntervalDuration

	windowEnd1 := testTimestamp.Add(interval)
	// First sample opens the first window.
	test.mockMemoryHistogram.On("AddSample", 5.0, 1.0, windowEnd1)
	assert.True(t, c.AddSample(newUsageSample(testTimestamp, 5, ResourceMemory)))

	// Second sample 2 intervals later falls outside the current window.
	windowEnd2 := windowEnd1.Add(2 * interval)
	test.mockMemoryHistogram.On("AddSample", 10.0, 1.0, windowEnd2)
	assert.True(t, c.AddSample(newUsageSample(testTimestamp.Add(2*interval), 10, ResourceMemory)))

	test.mockCPUHistogram.AssertExpectations(t)
	test.mockMemoryHistogram.AssertExpectations(t)
}
