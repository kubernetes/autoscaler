/*
Copyright 2018 The Kubernetes Authors.

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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

var (
	testPodID1  = vpa_model.PodID{Namespace: "namespace-1", PodName: "pod-1"}
	testPodID2  = vpa_model.PodID{Namespace: "namespace-1", PodName: "pod-2"}
	testRequest = vpa_model.Resources{
		vpa_model.ResourceCPU:    vpa_model.CPUAmountFromCores(3.14),
		vpa_model.ResourceMemory: vpa_model.MemoryAmountFromBytes(3.14e9),
	}
)

func addTestCPUSample(cluster *ClusterState, container vpa_model.ContainerID, cpuCores float64) error {
	sample := ContainerUsageSampleWithKey{
		Container: container,
		ContainerUsageSample: vpa_model.ContainerUsageSample{
			MeasureStart: testTimestamp,
			Usage:        vpa_model.CPUAmountFromCores(cpuCores),
			Request:      testRequest[vpa_model.ResourceCPU],
			Resource:     vpa_model.ResourceCPU,
		},
	}
	return cluster.AddSample(&sample)
}

func addTestMemorySample(cluster *ClusterState, container vpa_model.ContainerID, memoryBytes float64) error {
	sample := ContainerUsageSampleWithKey{
		Container: container,
		ContainerUsageSample: vpa_model.ContainerUsageSample{
			MeasureStart: testTimestamp,
			Usage:        vpa_model.MemoryAmountFromBytes(memoryBytes),
			Request:      testRequest[vpa_model.ResourceMemory],
			Resource:     vpa_model.ResourceMemory,
		},
	}
	return cluster.AddSample(&sample)
}

// Creates two pods, each having two containers:
//
//	testPodID1: { 'app-A', 'app-B' }
//	testPodID2: { 'app-A', 'app-C' }
//
// Adds a few usage samples to the containers.
// Verifies that AggregateStateByContainerName() properly aggregates
// container CPU and memory peak histograms, grouping the two containers
// with the same name ('app-A') together.
func TestAggregateStateByContainerName(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	cluster.AddOrUpdatePod(testPodID1, testLabels, apiv1.PodRunning)
	otherLabels := labels.Set{"label-2": "value-2"}
	cluster.AddOrUpdatePod(testPodID2, otherLabels, apiv1.PodRunning)

	// Create 4 containers: 2 with the same name and 2 with different names.
	containers := []vpa_model.ContainerID{
		{PodID: testPodID1, ContainerName: "app-A"},
		{PodID: testPodID1, ContainerName: "app-B"},
		{PodID: testPodID2, ContainerName: "app-A"},
		{PodID: testPodID2, ContainerName: "app-C"},
	}
	for _, c := range containers {
		assert.NoError(t, cluster.AddOrUpdateContainer(c, testRequest))
	}

	// Add CPU usage samples to all containers.
	assert.NoError(t, addTestCPUSample(cluster, containers[0], 1.0)) // app-A
	assert.NoError(t, addTestCPUSample(cluster, containers[1], 5.0)) // app-B
	assert.NoError(t, addTestCPUSample(cluster, containers[2], 3.0)) // app-A
	assert.NoError(t, addTestCPUSample(cluster, containers[3], 5.0)) // app-C
	// Add Memory usage samples to all containers.
	assert.NoError(t, addTestMemorySample(cluster, containers[0], 2e9))  // app-A
	assert.NoError(t, addTestMemorySample(cluster, containers[1], 10e9)) // app-B
	assert.NoError(t, addTestMemorySample(cluster, containers[2], 4e9))  // app-A
	assert.NoError(t, addTestMemorySample(cluster, containers[3], 10e9)) // app-C

	// Build the AggregateContainerStateMap.
	aggregateResources := AggregateStateByContainerName(cluster.aggregateStateMap)
	assert.Contains(t, aggregateResources, "app-A")
	assert.Contains(t, aggregateResources, "app-B")
	assert.Contains(t, aggregateResources, "app-C")

	// Expect samples from all containers to be grouped by the container name.
	assert.Equal(t, 2, aggregateResources["app-A"].TotalSamplesCount)
	assert.Equal(t, 1, aggregateResources["app-B"].TotalSamplesCount)
	assert.Equal(t, 1, aggregateResources["app-C"].TotalSamplesCount)

	config := vpa_model.GetAggregationsConfig()
	// Compute the expected histograms for the "app-A" containers.
	expectedCPUHistogram := util.NewDecayingHistogram(config.CPUHistogramOptions, config.CPUHistogramDecayHalfLife)
	expectedCPUHistogram.Merge(cluster.findOrCreateAggregateContainerState(containers[0]).AggregateCPUUsage)
	expectedCPUHistogram.Merge(cluster.findOrCreateAggregateContainerState(containers[2]).AggregateCPUUsage)
	actualCPUHistogram := aggregateResources["app-A"].AggregateCPUUsage

	expectedMemoryHistogram := util.NewDecayingHistogram(config.MemoryHistogramOptions, config.MemoryHistogramDecayHalfLife)
	expectedMemoryHistogram.AddSample(2e9, 1.0, cluster.GetContainer(containers[0]).WindowEnd)
	expectedMemoryHistogram.AddSample(4e9, 1.0, cluster.GetContainer(containers[2]).WindowEnd)
	actualMemoryHistogram := aggregateResources["app-A"].AggregateMemoryPeaks

	assert.True(t, expectedCPUHistogram.Equals(actualCPUHistogram), "Expected:\n%s\nActual:\n%s", expectedCPUHistogram, actualCPUHistogram)
	assert.True(t, expectedMemoryHistogram.Equals(actualMemoryHistogram), "Expected:\n%s\nActual:\n%s", expectedMemoryHistogram, actualMemoryHistogram)
}

func TestAggregateContainerStateSaveToCheckpoint(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	cs := vpa_model.NewAggregateContainerState()
	t1, t2 := time.Date(2018, time.January, 1, 2, 3, 4, 0, location), time.Date(2018, time.February, 1, 2, 3, 4, 0, location)
	cs.FirstSampleStart = t1
	cs.LastSampleStart = t2
	cs.TotalSamplesCount = 10

	cs.AggregateCPUUsage.AddSample(1, 33, t2)
	cs.AggregateMemoryPeaks.AddSample(1, 55, t1)
	cs.AggregateMemoryPeaks.AddSample(10000000, 55, t1)
	checkpoint, err := cs.SaveToCheckpoint()

	assert.NoError(t, err)

	assert.True(t, time.Since(checkpoint.LastUpdateTime.Time) < 10*time.Second)
	assert.Equal(t, t1, checkpoint.FirstSampleStart.Time)
	assert.Equal(t, t2, checkpoint.LastSampleStart.Time)
	assert.Equal(t, 10, checkpoint.TotalSamplesCount)

	assert.Equal(t, vpa_model.SupportedCheckpointVersion, checkpoint.Version)

	// Basic check that serialization of histograms happened.
	// Full tests are part of the Histogram.
	assert.Len(t, checkpoint.CPUHistogram.BucketWeights, 1)
	assert.Len(t, checkpoint.MemoryHistogram.BucketWeights, 2)
}

func TestAggregateContainerStateLoadFromCheckpointFailsForVersionMismatch(t *testing.T) {
	checkpoint := vpa_types.VerticalPodAutoscalerCheckpointStatus{
		Version: "foo",
	}
	cs := vpa_model.NewAggregateContainerState()
	err := cs.LoadFromCheckpoint(&checkpoint)
	assert.Error(t, err)
}

func TestAggregateContainerStateLoadFromCheckpoint(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	t1, t2 := time.Date(2018, time.January, 1, 2, 3, 4, 0, location), time.Date(2018, time.February, 1, 2, 3, 4, 0, location)

	checkpoint := vpa_types.VerticalPodAutoscalerCheckpointStatus{
		Version:           vpa_model.SupportedCheckpointVersion,
		LastUpdateTime:    metav1.NewTime(time.Now()),
		FirstSampleStart:  metav1.NewTime(t1),
		LastSampleStart:   metav1.NewTime(t2),
		TotalSamplesCount: 20,
		MemoryHistogram: vpa_types.HistogramCheckpoint{
			BucketWeights: map[int]uint32{
				0: 10,
			},
			TotalWeight: 33.0,
		},
		CPUHistogram: vpa_types.HistogramCheckpoint{
			BucketWeights: map[int]uint32{
				0: 10,
			},
			TotalWeight: 44.0,
		},
	}

	cs := vpa_model.NewAggregateContainerState()
	err := cs.LoadFromCheckpoint(&checkpoint)
	assert.NoError(t, err)

	assert.Equal(t, t1, cs.FirstSampleStart)
	assert.Equal(t, t2, cs.LastSampleStart)
	assert.Equal(t, 20, cs.TotalSamplesCount)
	assert.False(t, cs.AggregateCPUUsage.IsEmpty())
	assert.False(t, cs.AggregateMemoryPeaks.IsEmpty())
}

func TestAggregateContainerStateIsExpired(t *testing.T) {
	cs := vpa_model.NewAggregateContainerState()
	cs.LastSampleStart = testTimestamp
	cs.TotalSamplesCount = 1
	assert.False(t, isStateExpired(cs, testTimestamp.Add(7*24*time.Hour)))
	assert.True(t, isStateExpired(cs, testTimestamp.Add(8*24*time.Hour)))

	csEmpty := vpa_model.NewAggregateContainerState()
	csEmpty.TotalSamplesCount = 0
	csEmpty.CreationTime = testTimestamp
	assert.False(t, isStateExpired(csEmpty, testTimestamp.Add(7*24*time.Hour)))
	assert.True(t, isStateExpired(csEmpty, testTimestamp.Add(8*24*time.Hour)))
}

func TestUpdateFromPolicyScalingMode(t *testing.T) {
	scalingModeAuto := vpa_types.ContainerScalingModeAuto
	scalingModeOff := vpa_types.ContainerScalingModeOff
	testCases := []struct {
		name     string
		policy   *vpa_types.ContainerResourcePolicy
		expected *vpa_types.ContainerScalingMode
	}{
		{
			name: "Explicit auto scaling mode",
			policy: &vpa_types.ContainerResourcePolicy{
				Mode: &scalingModeAuto,
			},
			expected: &scalingModeAuto,
		}, {
			name: "Off scaling mode",
			policy: &vpa_types.ContainerResourcePolicy{
				Mode: &scalingModeOff,
			},
			expected: &scalingModeOff,
		}, {
			name:     "No mode specified - default to Auto",
			policy:   &vpa_types.ContainerResourcePolicy{},
			expected: &scalingModeAuto,
		}, {
			name:     "Nil policy - default to Auto",
			policy:   nil,
			expected: &scalingModeAuto,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := vpa_model.NewAggregateContainerState()
			cs.UpdateFromPolicy(tc.policy)
			assert.Equal(t, tc.expected, cs.GetScalingMode())
		})
	}
}

func TestUpdateFromPolicyControlledResources(t *testing.T) {
	testCases := []struct {
		name     string
		policy   *vpa_types.ContainerResourcePolicy
		expected []vpa_model.ResourceName
	}{
		{
			name: "Explicit ControlledResources",
			policy: &vpa_types.ContainerResourcePolicy{
				ControlledResources: &[]apiv1.ResourceName{apiv1.ResourceCPU, apiv1.ResourceMemory},
			},
			expected: []vpa_model.ResourceName{vpa_model.ResourceCPU, vpa_model.ResourceMemory},
		}, {
			name: "Empty ControlledResources",
			policy: &vpa_types.ContainerResourcePolicy{
				ControlledResources: &[]apiv1.ResourceName{},
			},
			expected: []vpa_model.ResourceName{},
		}, {
			name: "ControlledResources with one resource",
			policy: &vpa_types.ContainerResourcePolicy{
				ControlledResources: &[]apiv1.ResourceName{apiv1.ResourceMemory},
			},
			expected: []vpa_model.ResourceName{vpa_model.ResourceMemory},
		}, {
			name:     "No ControlledResources specified - used default",
			policy:   &vpa_types.ContainerResourcePolicy{},
			expected: []vpa_model.ResourceName{vpa_model.ResourceCPU, vpa_model.ResourceMemory},
		}, {
			name:     "Nil policy - use default",
			policy:   nil,
			expected: []vpa_model.ResourceName{vpa_model.ResourceCPU, vpa_model.ResourceMemory},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := vpa_model.NewAggregateContainerState()
			cs.UpdateFromPolicy(tc.policy)
			assert.Equal(t, tc.expected, cs.GetControlledResources())
		})
	}
}
