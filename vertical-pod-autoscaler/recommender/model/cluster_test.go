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
)

var (
	testTimestamp, _ = time.Parse(TimeLayout, "2017-04-18 17:35:05")
	testPodID        = PodID{"namespace-1", "pod-1"}
	testContainerID  = ContainerID{testPodID, "container-1"}
	testVpaID        = VpaID{"vpa-1"}
	testLabels       = map[string]string{"label-1": "value-1"}
	emptyLabels      = map[string]string{}
	testSelectorStr  = "label-1 = value-1"
)

func makeTestUsageSample() *ContainerUsageSampleWithKey {
	return &ContainerUsageSampleWithKey{ContainerUsageSample{testTimestamp, 1.0, 1.0}, testContainerID}
}

func TestClusterAddSample(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState()
	cluster.AddOrUpdatePod(testPodID, testLabels)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID))

	// Add a usage sample to the container.
	cluster.AddSample(makeTestUsageSample())

	// Verify that the sample was aggregated into the container stats.
	containerStats := cluster.Pods[testPodID].Containers["container-1"]
	assert.Equal(t, testTimestamp, containerStats.lastSampleStart)
}

// Verifies that AddSample and AddOrUpdateContainer methods return a proper
// KeyError when referring to a non-existent pod.
func TestMissingKeys(t *testing.T) {
	cluster := NewClusterState()
	err := cluster.AddSample(makeTestUsageSample())
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")

	err = cluster.AddOrUpdateContainer(testContainerID)
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")
}

func addTestPod(cluster *ClusterState) *PodState {
	cluster.AddOrUpdatePod(testPodID, testLabels)
	return cluster.Pods[testPodID]
}

func addTestVpa(cluster *ClusterState) *Vpa {
	cluster.AddOrUpdateVpa(testVpaID, testSelectorStr)
	return cluster.Vpas[testVpaID]
}

// Creates a VPA followed by a matching pod. Verifies that the links between the
// VPA and the pod are set correctly.
func TestAddVpaThenAddPod(t *testing.T) {
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	pod := addTestPod(cluster)
	assert.Equal(t, pod, vpa.Pods[testPodID])
	assert.Equal(t, vpa, pod.Vpa)
}

// Creates a pod followed by a matching VPA. Verifies that the links between the
// VPA and the pod are set correctly.
func TestAddPodThenAddVpa(t *testing.T) {
	cluster := NewClusterState()
	pod := addTestPod(cluster)
	vpa := addTestVpa(cluster)
	assert.Equal(t, pod, vpa.Pods[testPodID])
	assert.Equal(t, vpa, pod.Vpa)
}

// Creates a VPA and a matching pod. Verifies that after deleting the VPA the
// pod does not link to any Vpa.
func TestDeleteVpa(t *testing.T) {
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	pod := addTestPod(cluster)
	cluster.DeleteVpa(vpa.ID)
	assert.Nil(t, pod.Vpa)
}

// Creates a VPA and a matching pod. Verifies that after deleting the pod the
// VPA does not control any pods.
func TestDeletePod(t *testing.T) {
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	pod := addTestPod(cluster)
	assert.NoError(t, cluster.DeletePod(pod.ID))
	assert.Empty(t, vpa.Pods)
}

// Creates a VPA and a matching pod, then change the pod labels such that it is
// no longer matched by the VPA. Verifies that the links between the pod and the
// VPA are removed.
func TestChangePodLabels(t *testing.T) {
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	pod := addTestPod(cluster)
	// Update Pod labels to no longer match the VPA.
	cluster.AddOrUpdatePod(testPodID, emptyLabels)
	assert.Empty(t, vpa.Pods)
	assert.Nil(t, pod.Vpa)
}

// Creates a VPA and a matching pod, then change the VPA pod selector 3 times:
// first such that it still matches the pod, then such that it no longer matches
// the pod, finally such that it matches the pod again. Verifies that the links
// between the pod and the VPA are updated correctly each time.
func TestUpdatePodSelector(t *testing.T) {
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	pod := addTestPod(cluster)

	// Update the VPA selector such that it still matches the Pod.
	assert.NoError(t, cluster.AddOrUpdateVpa(testVpaID, "label-1 in (value-1,value-2)"))
	vpa = cluster.Vpas[testVpaID]
	assert.Equal(t, pod, vpa.Pods[testPodID])
	assert.Equal(t, vpa, pod.Vpa)

	// Update the VPA selector to no longer match the Pod.
	assert.NoError(t, cluster.AddOrUpdateVpa(testVpaID, "label-1 = value-2"))
	vpa = cluster.Vpas[testVpaID]
	assert.Empty(t, vpa.Pods)
	assert.Nil(t, pod.Vpa)

	// Update the VPA selector to match the Pod again.
	assert.NoError(t, cluster.AddOrUpdateVpa(testVpaID, "label-1 = value-1"))
	vpa = cluster.Vpas[testVpaID]
	assert.Equal(t, pod, vpa.Pods[testPodID])
	assert.Equal(t, vpa, pod.Vpa)
}

// Creates a VPA and a matching pod, then add another VPA matching the same pod.
// Verifies that the pod is controlled by the first VPA.
// Next deletes the first VPA and verfies that the pod is controlled by the
// remaning VPA. Finally deletes the other VPA and verifies that the pod is no
// longer controlled by any VPA.
func TestTwoVpasForPod(t *testing.T) {
	cluster := NewClusterState()
	cluster.AddOrUpdateVpa(VpaID{"vpa-1"}, "label-1 = value-1")
	pod := addTestPod(cluster)
	cluster.AddOrUpdateVpa(VpaID{"vpa-2"}, "label-1 in (value-1,value-2)")
	assert.Equal(t, cluster.Vpas[VpaID{"vpa-1"}], pod.Vpa)
	// Delete the VPA that currently controls the Pod. Expect that it will
	// switch to the remaining one.
	assert.NoError(t, cluster.DeleteVpa(VpaID{"vpa-1"}))
	assert.Equal(t, cluster.Vpas[VpaID{"vpa-2"}], pod.Vpa)
	// Delete the other VPA. The Pod is no longer vertically-scaled by anyone.
	assert.NoError(t, cluster.DeleteVpa(VpaID{"vpa-2"}))
	assert.Nil(t, pod.Vpa)
}

// Verifies that a VPA with an empty selector (matching all pods) matches a pod
// with labels as well as a pod with no labels.
func TestEmptySelector(t *testing.T) {
	cluster := NewClusterState()
	// Create a VPA with an empty selector (matching all pods).
	assert.NoError(t, cluster.AddOrUpdateVpa(testVpaID, ""))
	// Create a pod with labels.
	cluster.AddOrUpdatePod(testPodID, testLabels)
	// Create a pod without labels.
	anotherPodID := PodID{"namespace-1", "pod-2"}
	cluster.AddOrUpdatePod(anotherPodID, emptyLabels)
	// Both pods should be matched by the VPA.
	assert.Equal(t, cluster.Vpas[testVpaID], cluster.Pods[testPodID].Vpa)
	assert.Equal(t, cluster.Vpas[testVpaID], cluster.Pods[anotherPodID].Vpa)
}
