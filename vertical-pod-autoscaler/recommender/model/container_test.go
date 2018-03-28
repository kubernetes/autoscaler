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
	TimeLayout  = "2006-01-02 15:04:05"
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
	return &ContainerUsageSample{timestamp, ResourceAmount(usage), resource}
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
	c := &ContainerState{
		Request:               TestRequest,
		CPUUsage:              mockCPUHistogram,
		LastCPUSampleStart:    time.Unix(0, 0),
		MemoryUsagePeaks:      memoryUsagePeaks,
		WindowEnd:             time.Unix(0, 0),
		lastMemorySampleStart: time.Unix(0, 0)}

	// Verify that CPU measures are added to the CPU histogram.
	// The weight should be equal to the current request.
	timeStep := MemoryAggregationInterval / 2
	mockCPUHistogram.On("AddSample", 3.14, 2.3, testTimestamp)
	mockCPUHistogram.On("AddSample", 6.28, 2.3, testTimestamp.Add(timeStep))
	mockCPUHistogram.On("AddSample", 1.57, 2.3, testTimestamp.Add(2*timeStep))

	// Add three usage samples.
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

	// Verify that memory peak samples were aggregated properly.
	assert.Equal(t, []float64{10, 2}, memoryUsagePeaks.Contents())
}

func TestRecordOOMIncreasedByBumpUp(t *testing.T) {
	testTimestamp, err := time.Parse(TimeLayout, "2017-04-18 17:35:05")
	assert.Nil(t, err)
	mockCPUHistogram := new(util.MockHistogram)
	memoryUsagePeaks := util.NewFloatSlidingWindow(
		int(MemoryAggregationWindowLength / MemoryAggregationInterval))
	c := &ContainerState{
		Request:               TestRequest,
		CPUUsage:              mockCPUHistogram,
		LastCPUSampleStart:    time.Unix(0, 0),
		MemoryUsagePeaks:      memoryUsagePeaks,
		WindowEnd:             time.Unix(0, 0),
		lastMemorySampleStart: time.Unix(0, 0)}

	assert.NoError(t, c.RecordOOM(testTimestamp, ResourceAmount(1000*mb)))
	// Bump Up factor is 20%
	assert.Equal(t, []float64{1200 * mb}, memoryUsagePeaks.Contents())
}

func TestRecordOOMIncreasedByMin(t *testing.T) {
	testTimestamp, err := time.Parse(TimeLayout, "2017-04-18 17:35:05")
	assert.Nil(t, err)
	mockCPUHistogram := new(util.MockHistogram)
	memoryUsagePeaks := util.NewFloatSlidingWindow(
		int(MemoryAggregationWindowLength / MemoryAggregationInterval))
	c := &ContainerState{
		Request:               TestRequest,
		CPUUsage:              mockCPUHistogram,
		LastCPUSampleStart:    time.Unix(0, 0),
		MemoryUsagePeaks:      memoryUsagePeaks,
		WindowEnd:             time.Unix(0, 0),
		lastMemorySampleStart: time.Unix(0, 0)}

	assert.NoError(t, c.RecordOOM(testTimestamp, ResourceAmount(1*mb)))
	// Min grow by 100Mb
	assert.Equal(t, []float64{101 * mb}, memoryUsagePeaks.Contents())
}

func TestRecordOOMMaxedWithKnownSample(t *testing.T) {
	testTimestamp, err := time.Parse(TimeLayout, "2017-04-18 17:35:05")
	assert.Nil(t, err)
	mockCPUHistogram := new(util.MockHistogram)
	memoryUsagePeaks := util.NewFloatSlidingWindow(
		int(MemoryAggregationWindowLength / MemoryAggregationInterval))
	c := &ContainerState{
		Request:               TestRequest,
		CPUUsage:              mockCPUHistogram,
		LastCPUSampleStart:    time.Unix(0, 0),
		MemoryUsagePeaks:      memoryUsagePeaks,
		WindowEnd:             time.Unix(0, 0),
		lastMemorySampleStart: time.Unix(0, 0)}

	assert.True(t, c.AddSample(newUsageSample(testTimestamp, 3000*mb, ResourceMemory)))
	assert.NoError(t, c.RecordOOM(testTimestamp, ResourceAmount(1000*mb)))
	// Last known sample is higher then request, so it is taken.
	assert.Equal(t, []float64{3600 * mb}, memoryUsagePeaks.Contents())
}

func TestRecordOOMDiscardsOldSample(t *testing.T) {
	testTimestamp, err := time.Parse(TimeLayout, "2017-04-18 17:35:05")
	assert.Nil(t, err)
	mockCPUHistogram := new(util.MockHistogram)
	memoryUsagePeaks := util.NewFloatSlidingWindow(
		int(MemoryAggregationWindowLength / MemoryAggregationInterval))
	c := &ContainerState{
		Request:               TestRequest,
		CPUUsage:              mockCPUHistogram,
		LastCPUSampleStart:    time.Unix(0, 0),
		MemoryUsagePeaks:      memoryUsagePeaks,
		WindowEnd:             time.Unix(0, 0),
		lastMemorySampleStart: time.Unix(0, 0)}

	assert.True(t, c.AddSample(newUsageSample(testTimestamp, 1000*mb, ResourceMemory)))
	assert.Error(t, c.RecordOOM(testTimestamp.Add(-30*time.Hour), ResourceAmount(1000*mb)))
	// OOM is obsolete, mem not changed
	assert.Equal(t, []float64{1000 * mb}, memoryUsagePeaks.Contents())
}

func TestRecordOOMInNewWindow(t *testing.T) {
	testTimestamp, err := time.Parse(TimeLayout, "2017-04-18 17:35:05")
	assert.Nil(t, err)
	mockCPUHistogram := new(util.MockHistogram)
	memoryUsagePeaks := util.NewFloatSlidingWindow(
		int(MemoryAggregationWindowLength / MemoryAggregationInterval))
	c := &ContainerState{
		Request:               TestRequest,
		CPUUsage:              mockCPUHistogram,
		LastCPUSampleStart:    time.Unix(0, 0),
		MemoryUsagePeaks:      memoryUsagePeaks,
		WindowEnd:             time.Unix(0, 0),
		lastMemorySampleStart: time.Unix(0, 0)}

	assert.True(t, c.AddSample(newUsageSample(testTimestamp, 2000*mb, ResourceMemory)))
	assert.NoError(t, c.RecordOOM(testTimestamp.Add(2*MemoryAggregationInterval), ResourceAmount(1000*mb)))
	assert.Equal(t, []float64{2000 * mb, 0, 1200 * mb}, memoryUsagePeaks.Contents())
}
