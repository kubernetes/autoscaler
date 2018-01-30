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
	cluster.AddOrUpdatePod(testPodID1, testLabels)
	cluster.AddOrUpdatePod(testPodID2, testLabels)

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
	expectedCPUHistogram.Merge(&cluster.GetContainer(containers[2]).CPUUsage)
	actualCPUHistogram := aggregateResources["app-A"].aggregateCPUUsage

	expectedMemoryHistogram := util.NewHistogram(model.MemoryHistogramOptions)
	expectedMemoryHistogram.AddSample(2e9, 1.0)
	expectedMemoryHistogram.AddSample(4e9, 1.0)
	actualMemoryHistogram := aggregateResources["app-A"].aggregateMemoryPeaks

	assert.True(t, util.HistogramsEqual(expectedCPUHistogram, actualCPUHistogram), "Expected:\n%s\nActual:\n%s", expectedCPUHistogram, actualCPUHistogram)
	assert.True(t, util.HistogramsEqual(expectedMemoryHistogram, actualMemoryHistogram), "Expected:\n%s\nActual:\n%s", expectedMemoryHistogram, actualMemoryHistogram)
}
