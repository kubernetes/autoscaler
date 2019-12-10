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

	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
)

var (
	testPodID       = PodID{"namespace-1", "pod-1"}
	testContainerID = ContainerID{testPodID, "container-1"}
	testVpaID       = VpaID{"namespace-1", "vpa-1"}
	testLabels      = map[string]string{"label-1": "value-1"}
	emptyLabels     = map[string]string{}
	testSelectorStr = "label-1 = value-1"
)

func makeTestUsageSample() *ContainerUsageSampleWithKey {
	return &ContainerUsageSampleWithKey{ContainerUsageSample{
		MeasureStart: testTimestamp,
		Usage:        1.0,
		Request:      testRequest[ResourceCPU],
		Resource:     ResourceCPU},
		testContainerID}
}

func TestClusterAddSample(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState()
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(makeTestUsageSample()))

	// Verify that the sample was aggregated into the container stats.
	containerStats := cluster.Pods[testPodID].Containers["container-1"]
	assert.Equal(t, testTimestamp, containerStats.LastCPUSampleStart)
}

func TestClusterGCAggregateContainerStateDeletesOld(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	addTestPod(cluster)

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	usageSample := makeTestUsageSample()

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(usageSample))

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	// AggegateContainerState are valid for 8 days since last sample
	cluster.GrabageCollectAggregateCollectionStates(usageSample.MeasureStart.Add(9 * 24 * time.Hour))

	// AggegateContainerState should be deleted from both cluster and vpa
	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, vpa.aggregateContainerStates)
}

func TestClusterGCAggregateContainerStateLeavesValid(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	addTestPod(cluster)

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	usageSample := makeTestUsageSample()

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(usageSample))

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	// AggegateContainerState are valid for 8 days since last sample
	cluster.GrabageCollectAggregateCollectionStates(usageSample.MeasureStart.Add(7 * 24 * time.Hour))

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)
}

func TestClusterRecordOOM(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState()
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))

	// RecordOOM
	assert.NoError(t, cluster.RecordOOM(testContainerID, time.Unix(0, 0), ResourceAmount(10)))

	// Verify that OOM was aggregated into the aggregated stats.
	aggregation := cluster.findOrCreateAggregateContainerState(testContainerID)
	assert.NotEmpty(t, aggregation.AggregateMemoryPeaks)
}

// Verifies that AddSample and AddOrUpdateContainer methods return a proper
// KeyError when referring to a non-existent pod.
func TestMissingKeys(t *testing.T) {
	cluster := NewClusterState()
	err := cluster.AddSample(makeTestUsageSample())
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")

	err = cluster.RecordOOM(testContainerID, time.Unix(0, 0), ResourceAmount(10))
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")

	err = cluster.AddOrUpdateContainer(testContainerID, testRequest)
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")
}

func addVpa(cluster *ClusterState, id VpaID, selector string) *Vpa {
	var apiObject vpa_types.VerticalPodAutoscaler
	apiObject.Namespace = id.Namespace
	apiObject.Name = id.VpaName
	apiObject.Spec.Selector, _ = metav1.ParseToLabelSelector(selector)
	err := cluster.AddOrUpdateVpa(&apiObject)
	if err != nil {
		glog.Fatalf("AddOrUpdateVpa() failed: %v", err)
	}
	return cluster.Vpas[id]
}

func addTestVpa(cluster *ClusterState) *Vpa {
	return addVpa(cluster, testVpaID, testSelectorStr)
}

func addTestPod(cluster *ClusterState) *PodState {
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	return cluster.Pods[testPodID]
}

func addTestContainer(cluster *ClusterState) *ContainerState {
	cluster.AddOrUpdateContainer(testContainerID, testRequest)
	return cluster.GetContainer(testContainerID)
}

// Creates a VPA followed by a matching pod. Verifies that the links between
// VPA, the container and the aggregation are set correctly.
func TestAddVpaThenAddPod(t *testing.T) {
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	addTestPod(cluster)
	container := addTestContainer(cluster)
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.True(t, container.aggregator == vpa.aggregateContainerStates[aggregateStateKey])
}

// Creates a pod followed by a matching VPA. Verifies that the links between
// VPA, the container and the aggregation are set correctly.
func TestAddPodThenAddVpa(t *testing.T) {
	cluster := NewClusterState()
	addTestPod(cluster)
	container := addTestContainer(cluster)
	vpa := addTestVpa(cluster)
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.Contains(t, vpa.aggregateContainerStates, aggregateStateKey)
	assert.True(t, container.aggregator == vpa.aggregateContainerStates[aggregateStateKey])
}

// Creates a VPA and a matching pod, then change the pod labels such that it is
// no longer matched by the VPA. Verifies that the links between the pod and the
// VPA are removed.
func TestChangePodLabels(t *testing.T) {
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	addTestPod(cluster)
	container := addTestContainer(cluster)
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	// Update Pod labels to no longer match the VPA.
	cluster.AddOrUpdatePod(testPodID, emptyLabels, apiv1.PodRunning)
	assert.False(t, container.aggregator == vpa.aggregateContainerStates[aggregateStateKey])
}

// Creates a VPA and a matching pod, then change the VPA pod selector 3 times:
// first such that it still matches the pod, then such that it no longer matches
// the pod, finally such that it matches the pod again. Verifies that the links
// between the pod and the VPA are updated correctly each time.
func TestUpdatePodSelector(t *testing.T) {
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	addTestPod(cluster)
	addTestContainer(cluster)

	// Update the VPA selector such that it still matches the Pod.
	vpa = addVpa(cluster, testVpaID, "label-1 in (value-1,value-2)")
	assert.Contains(t, vpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(testContainerID))

	// Update the VPA selector to no longer match the Pod.
	vpa = addVpa(cluster, testVpaID, "label-1 = value-2")
	assert.NotContains(t, vpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(testContainerID))

	// Update the VPA selector to match the Pod again.
	vpa = addVpa(cluster, testVpaID, "label-1 = value-1")
	assert.Contains(t, vpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(testContainerID))
}

// Verify that two copies of the same AggregateStateKey are equal.
func TestEqualAggregateStateKey(t *testing.T) {
	cluster := NewClusterState()
	pod := addTestPod(cluster)
	key1 := cluster.MakeAggregateStateKey(pod, "container-1")
	key2 := cluster.MakeAggregateStateKey(pod, "container-1")
	assert.True(t, key1 == key2)
}

// Verify that two containers with the same name, living in two pods with the same namespace and labels
// (although different pod names) are aggregated together.
func TestTwoPodsWithSameLabels(t *testing.T) {
	podID1 := PodID{"namespace-1", "pod-1"}
	podID2 := PodID{"namespace-1", "pod-2"}
	containerID1 := ContainerID{podID1, "foo-container"}
	containerID2 := ContainerID{podID2, "foo-container"}

	cluster := NewClusterState()
	cluster.AddOrUpdatePod(podID1, testLabels, apiv1.PodRunning)
	cluster.AddOrUpdatePod(podID2, testLabels, apiv1.PodRunning)
	cluster.AddOrUpdateContainer(containerID1, testRequest)
	cluster.AddOrUpdateContainer(containerID2, testRequest)

	// Expect only one aggregation to be created.
	assert.Equal(t, 1, len(cluster.aggregateStateMap))
}

// Verify that two identical containers in different namespaces are not aggregated together.
func TestTwoPodsWithDifferentNamespaces(t *testing.T) {
	podID1 := PodID{"namespace-1", "foo-pod"}
	podID2 := PodID{"namespace-2", "foo-pod"}
	containerID1 := ContainerID{podID1, "foo-container"}
	containerID2 := ContainerID{podID2, "foo-container"}

	cluster := NewClusterState()
	cluster.AddOrUpdatePod(podID1, testLabels, apiv1.PodRunning)
	cluster.AddOrUpdatePod(podID2, testLabels, apiv1.PodRunning)
	cluster.AddOrUpdateContainer(containerID1, testRequest)
	cluster.AddOrUpdateContainer(containerID2, testRequest)

	// Expect two separate aggregations to be created.
	assert.Equal(t, 2, len(cluster.aggregateStateMap))
	// Expect only one entry to be present in the labels set map.
	assert.Equal(t, 1, len(cluster.labelSetMap))
}

// Verifies that a VPA with an empty selector (matching all pods) matches a pod
// with labels as well as a pod with no labels.
func TestEmptySelector(t *testing.T) {
	cluster := NewClusterState()
	// Create a VPA with an empty selector (matching all pods).
	vpa := addVpa(cluster, testVpaID, "")
	// Create a pod with labels. Add a container.
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	containerID1 := ContainerID{testPodID, "foo"}
	assert.NoError(t, cluster.AddOrUpdateContainer(containerID1, testRequest))

	// Create a pod without labels. Add a container.
	anotherPodID := PodID{"namespace-1", "pod-2"}
	cluster.AddOrUpdatePod(anotherPodID, emptyLabels, apiv1.PodRunning)
	containerID2 := ContainerID{anotherPodID, "foo"}
	assert.NoError(t, cluster.AddOrUpdateContainer(containerID2, testRequest))

	// Both pods should be matched by the VPA.
	assert.Contains(t, vpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(containerID1))
	assert.Contains(t, vpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(containerID2))
}
