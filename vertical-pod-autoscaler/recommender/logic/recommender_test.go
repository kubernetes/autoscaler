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

package logic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/util"
)

var (
	TimeLayout       = "2006-01-02 15:04:05"
	testTimestamp, _ = time.Parse(TimeLayout, "2017-04-18 17:35:05")
	testPodID1       = model.PodID{"namespace-1", "pod-1"}
	testPodID2       = model.PodID{"namespace-1", "pod-2"}
	testContainerID1 = model.ContainerID{testPodID1, "container-1"}
	testVpaID        = model.VpaID{"namespace-1", "vpa-1"}
	testLabels       = map[string]string{"label-1": "value-1"}
	emptyLabels      = map[string]string{}
	testSelectorStr  = "label-1 = value-1"
)

func addTestSample(cluster *model.ClusterState, container model.ContainerID, cpu float64, memory float64) error {
	var sample model.ContainerUsageSampleWithKey
	sample.Container = container
	sample.MeasureStart = testTimestamp
	sample.Usage = model.CPUAmountFromCores(cpu)
	sample.Resource = model.ResourceCPU
	err := cluster.AddSample(&sample)
	if err != nil {
		return err
	}
	sample.Usage = model.MemoryAmountFromBytes(memory)
	sample.Resource = model.ResourceMemory
	return cluster.AddSample(&sample)
}

// Creates two pods, each having two containers:
//   testPodID1: { 'app-A', 'app-B' }
//   testPodID2: { 'app-A', 'app-C' }
// Adds a few usage samples to the containers.
// Verifies that buildAggregateContainerStateMap() properly aggregates
// container CPU and memory peak histograms, grouping the two containers
// with the same name ('app-A') together.
func TestBuildAggregateResourcesMap(t *testing.T) {
	cluster := model.NewClusterState()
	cluster.AddOrUpdatePod(testPodID1, testLabels, apiv1.PodRunning)
	cluster.AddOrUpdatePod(testPodID2, testLabels, apiv1.PodRunning)

	// Create 4 containers: 2 with the same name and 2 with different names.
	containers := []model.ContainerID{
		{testPodID1, "app-A"},
		{testPodID1, "app-B"},
		{testPodID2, "app-A"},
		{testPodID2, "app-C"},
	}
	for _, c := range containers {
		assert.NoError(t, cluster.AddOrUpdateContainer(c))
	}

	// Add usage samples to all containers.
	assert.NoError(t, addTestSample(cluster, containers[0], 1.0, 2e9))  // app-A
	assert.NoError(t, addTestSample(cluster, containers[1], 5.0, 10e9)) // app-B
	assert.NoError(t, addTestSample(cluster, containers[2], 3.0, 4e9))  // app-A
	assert.NoError(t, addTestSample(cluster, containers[3], 5.0, 10e9)) // app-C

	// Build the AggregateContainerStateMap.
	aggregateResources := buildAggregateContainerStateMap(&cluster.Pods)
	assert.Contains(t, aggregateResources, "app-A")
	assert.Contains(t, aggregateResources, "app-B")
	assert.Contains(t, aggregateResources, "app-C")

	// Compute the expected histograms for the "app-A" containers.
	expectedCPUHistogram := cluster.GetContainer(containers[0]).CPUUsage
	expectedCPUHistogram.Merge(cluster.GetContainer(containers[2]).CPUUsage)
	actualCPUHistogram := aggregateResources["app-A"].aggregateCPUUsage

	expectedMemoryHistogram := util.NewDecayingHistogram(model.MemoryHistogramOptions, model.MemoryHistogramDecayHalfLife)
	expectedMemoryHistogram.AddSample(2e9, 1.0, cluster.GetContainer(containers[0]).WindowEnd)
	expectedMemoryHistogram.AddSample(4e9, 1.0, cluster.GetContainer(containers[2]).WindowEnd)
	actualMemoryHistogram := aggregateResources["app-A"].aggregateMemoryPeaks

	assert.True(t, expectedCPUHistogram.Equals(actualCPUHistogram), "Expected:\n%s\nActual:\n%s", expectedCPUHistogram, actualCPUHistogram)
	assert.True(t, expectedMemoryHistogram.Equals(actualMemoryHistogram), "Expected:\n%s\nActual:\n%s", expectedMemoryHistogram, actualMemoryHistogram)
}

func TestAggregateContainerStateSaveToCheckpoint(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	cs := newAggregateContainerState()
	t1, t2 := time.Date(2018, time.January, 1, 2, 3, 4, 0, location), time.Date(2018, time.February, 1, 2, 3, 4, 0, location)
	cs.firstSampleStart = t1
	cs.lastSampleStart = t2
	cs.totalSamplesCount = 10

	cs.aggregateCPUUsage.AddSample(1, 33, t2)
	cs.aggregateMemoryPeaks.AddSample(1, 55, t1)
	cs.aggregateMemoryPeaks.AddSample(10000000, 55, t1)
	checkpoint, err := cs.SaveToCheckpoint()

	assert.NoError(t, err)
	assert.Equal(t, t1, checkpoint.FirstSampleStart.Time)
	assert.Equal(t, t2, checkpoint.LastSampleStart.Time)
	assert.Equal(t, 10, checkpoint.TotalSamplesCount)

	assert.Equal(t, SupportedCheckpointVersion, checkpoint.Version)

	// Basic check that serialization of histograms happened.
	// Full tests are part of the Histogram.
	assert.Len(t, checkpoint.CPUHistogram.BucketWeights, 1)
	assert.Len(t, checkpoint.MemoryHistogram.BucketWeights, 2)
}

func TestAggregateContainerStateLoadFromCheckpointFailsForVersionMismatch(t *testing.T) {
	checkpoint := vpa_types.VerticalPodAutoscalerCheckpointStatus{
		Version: "foo",
	}
	cs := newAggregateContainerState()
	err := cs.LoadFromCheckpoint(&checkpoint)
	assert.Error(t, err)
}

func TestAggregateContainerStateLoadFromCheckpoint(t *testing.T) {
	location, _ := time.LoadLocation("UTC")

	t1, t2 := time.Date(2018, time.January, 1, 2, 3, 4, 0, location), time.Date(2018, time.February, 1, 2, 3, 4, 0, location)

	checkpoint := vpa_types.VerticalPodAutoscalerCheckpointStatus{
		Version:           SupportedCheckpointVersion,
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

	cs := newAggregateContainerState()
	err := cs.LoadFromCheckpoint(&checkpoint)
	assert.NoError(t, err)

	assert.Equal(t, t1, cs.firstSampleStart)
	assert.Equal(t, t2, cs.lastSampleStart)
	assert.Equal(t, 20, cs.totalSamplesCount)
	assert.False(t, cs.aggregateCPUUsage.IsEmpty())
	assert.False(t, cs.aggregateMemoryPeaks.IsEmpty())
}
