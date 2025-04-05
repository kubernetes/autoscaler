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
		AggregateCPUUsage:    mockCPUHistogram,
		AggregateMemoryPeaks: mockMemoryHistogram,
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

// Add 6 usage samples (3 valid, 3 invalid) to a container. Verifies that for
// valid samples the CPU measures are aggregated in the CPU histogram and
// the memory measures are aggregated in the memory peaks sliding window.
// Verifies that invalid samples (out-of-order or negative usage) are ignored.
func TestAggregateContainerUsageSamples(t *testing.T) {
	test := newContainerTest()
	c := test.container
	memoryAggregationInterval := GetAggregationsConfig().MemoryAggregationInterval
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
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationInterval)
	// Bump Up factor is 20%.
	test.mockMemoryHistogram.On("AddSample", 1200.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb), nil, nil))
}

func TestRecordOOMDontRunAway(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationInterval)

	// Bump Up factor is 20%.
	test.mockMemoryHistogram.On("AddSample", 1200.0*mb, 1.0, memoryAggregationWindowEnd)
	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb), nil, nil))

	// new smaller OOMs don't influence the sample value (oomPeak)
	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(999*mb), nil, nil))
	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(999*mb), nil, nil))

	test.mockMemoryHistogram.On("SubtractSample", 1200.0*mb, 1.0, memoryAggregationWindowEnd)
	test.mockMemoryHistogram.On("AddSample", 2400.0*mb, 1.0, memoryAggregationWindowEnd)
	// a larger OOM should increase the sample value
	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(2000*mb), nil, nil))
}

func TestRecordOOMIncreasedByMin(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationInterval)
	// Min grow by 100Mb.
	test.mockMemoryHistogram.On("AddSample", 101.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1*mb), nil, nil))
}

func TestRecordOOMMaxedWithKnownSample(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationInterval)

	test.mockMemoryHistogram.On("AddSample", 3000.0*mb, 1.0, memoryAggregationWindowEnd)
	assert.True(t, test.container.AddSample(newUsageSample(testTimestamp, 3000*mb, ResourceMemory)))

	// Last known sample is higher than request, so it is taken.
	test.mockMemoryHistogram.On("SubtractSample", 3000.0*mb, 1.0, memoryAggregationWindowEnd)
	test.mockMemoryHistogram.On("AddSample", 3600.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb), nil, nil))
}

func TestRecordOOMDiscardsOldSample(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationInterval)

	test.mockMemoryHistogram.On("AddSample", 1000.0*mb, 1.0, memoryAggregationWindowEnd)
	assert.True(t, test.container.AddSample(newUsageSample(testTimestamp, 1000*mb, ResourceMemory)))

	// OOM is stale, mem not changed.
	assert.Error(t, test.container.RecordOOM(testTimestamp.Add(-30*time.Hour), ResourceAmount(1000*mb), nil, nil))
}

func TestRecordOOMInNewWindow(t *testing.T) {
	test := newContainerTest()
	memoryAggregationInterval := GetAggregationsConfig().MemoryAggregationInterval
	memoryAggregationWindowEnd := testTimestamp.Add(memoryAggregationInterval)

	test.mockMemoryHistogram.On("AddSample", 2000.0*mb, 1.0, memoryAggregationWindowEnd)
	assert.True(t, test.container.AddSample(newUsageSample(testTimestamp, 2000*mb, ResourceMemory)))

	memoryAggregationWindowEnd = memoryAggregationWindowEnd.Add(2 * memoryAggregationInterval)
	test.mockMemoryHistogram.On("AddSample", 2400.0*mb, 1.0, memoryAggregationWindowEnd)
	assert.NoError(t, test.container.RecordOOM(testTimestamp.Add(2*memoryAggregationInterval), ResourceAmount(1000*mb), nil, nil))
}

func TestRecordOOMWithCustomBumpUpRatio(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationInterval)

	customBumpUpRatio := 1.5 // 50% increase
	test.mockMemoryHistogram.On("AddSample", 1500.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb), &customBumpUpRatio, nil))
}

func TestRecordOOMWithCustomMinBumpUp(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationInterval)

	customMinBumpUp := float64(200 * mb)
	test.mockMemoryHistogram.On("AddSample", 1200.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb), nil, &customMinBumpUp))
}

func TestRecordOOMWithCustomBumpUpRatioAndMinBumpUp(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationInterval)

	customBumpUpRatio := 1.3 // 30% increase
	customMinBumpUp := float64(500 * mb)
	test.mockMemoryHistogram.On("AddSample", 1500.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb), &customBumpUpRatio, &customMinBumpUp))
}

func TestRecordOOMWithZeroBumpUpRatio(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationInterval)

	zeroBumpUpRatio := 0.0
	test.mockMemoryHistogram.On("AddSample", 1100.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb), &zeroBumpUpRatio, nil))
}

func TestRecordOOMWithNegativeBumpUpRatio(t *testing.T) {
	test := newContainerTest()
	memoryAggregationWindowEnd := testTimestamp.Add(GetAggregationsConfig().MemoryAggregationInterval)

	negativeBumpUpRatio := -0.5
	test.mockMemoryHistogram.On("AddSample", 1200.0*mb, 1.0, memoryAggregationWindowEnd)

	assert.NoError(t, test.container.RecordOOM(testTimestamp, ResourceAmount(1000*mb), &negativeBumpUpRatio, nil))
}
