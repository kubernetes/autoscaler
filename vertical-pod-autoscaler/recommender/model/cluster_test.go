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
	testTimestamp, _ = time.Parse(TimeLayout, "2017-04-18 17:35:05")
	testPodID        = PodID{"namespace-1", "pod-1"}
	testContainerID  = ContainerID{testPodID, "container-1"}
	testVpaID        = VpaID{"namespace-1", "vpa-1"}
	testLabels       = map[string]string{"label-1": "value-1"}
	emptyLabels      = map[string]string{}
	testSelectorStr  = "label-1 = value-1"
)

func makeTestUsageSample() *ContainerUsageSampleWithKey {
	return &ContainerUsageSampleWithKey{ContainerUsageSample{testTimestamp, 1.0, ResourceCPU}, testContainerID}
}

func TestClusterAddSample(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState()
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))

	// Add a usage sample to the container.
	cluster.AddSample(makeTestUsageSample())

	// Verify that the sample was aggregated into the container stats.
	containerStats := cluster.Pods[testPodID].Containers["container-1"]
	assert.Equal(t, testTimestamp, containerStats.LastCPUSampleStart)
}

func TestClusterRecordOOM(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState()
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))

	// RecordOOM
	assert.NoError(t, cluster.RecordOOM(testContainerID, time.Unix(0, 0), ResourceAmount(10)))

	// Verify that OOM was aggregated into the container stats.
	containerStats := cluster.Pods[testPodID].Containers["container-1"]
	assert.NotEmpty(t, containerStats.MemoryUsagePeaks.Contents)
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

// Creates a VPA followed by a matching pod. Verifies that the links between the
// VPA and the pod are set correctly.
func TestAddVpaThenAddPod(t *testing.T) {
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	pod := addTestPod(cluster)
	assert.Equal(t, pod, vpa.Pods[testPodID])
	assert.Equal(t, map[VpaID]*Vpa{testVpaID: vpa}, pod.MatchingVpas)
}

// Creates a pod followed by a matching VPA. Verifies that the links between the
// VPA and the pod are set correctly.
func TestAddPodThenAddVpa(t *testing.T) {
	cluster := NewClusterState()
	pod := addTestPod(cluster)
	vpa := addTestVpa(cluster)
	assert.Equal(t, pod, vpa.Pods[testPodID])
	assert.Equal(t, map[VpaID]*Vpa{testVpaID: vpa}, pod.MatchingVpas)
}

// Creates a VPA and a matching pod. Verifies that after deleting the VPA the
// pod does not link to any Vpa.
func TestDeleteVpa(t *testing.T) {
	cluster := NewClusterState()
	vpa := addTestVpa(cluster)
	pod := addTestPod(cluster)
	cluster.DeleteVpa(vpa.ID)
	assert.Empty(t, pod.MatchingVpas)
	assert.NotContains(t, cluster.Vpas, vpa.ID)
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
	cluster.AddOrUpdatePod(testPodID, emptyLabels, apiv1.PodRunning)
	assert.Empty(t, vpa.Pods)
	assert.Empty(t, pod.MatchingVpas)
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
	vpa = addVpa(cluster, testVpaID, "label-1 in (value-1,value-2)")
	assert.Equal(t, pod, vpa.Pods[testPodID])
	assert.Equal(t, map[VpaID]*Vpa{testVpaID: vpa}, pod.MatchingVpas)

	// Update the VPA selector to no longer match the Pod.
	vpa = addVpa(cluster, testVpaID, "label-1 = value-2")
	assert.Empty(t, vpa.Pods)
	assert.Empty(t, pod.MatchingVpas)

	// Update the VPA selector to match the Pod again.
	vpa = addVpa(cluster, testVpaID, "label-1 = value-1")
	assert.Equal(t, pod, vpa.Pods[testPodID])
	assert.Equal(t, map[VpaID]*Vpa{testVpaID: vpa}, pod.MatchingVpas)
}

// Creates a VPA and a matching pod, then add another VPA matching the same pod.
// Verifies that the pod knows about both of them.
// Next deletes one of them and verfies that the pod is controlled by the
// remaning VPA. Finally deletes the other VPA and verifies that the pod is no
// longer controlled by any VPA.
func TestTwoVpasForPod(t *testing.T) {
	cluster := NewClusterState()
	addVpa(cluster, VpaID{"namespace-1", "vpa-1"}, "label-1 = value-1")
	pod := addTestPod(cluster)
	addVpa(cluster, VpaID{"namespace-1", "vpa-2"}, "label-1 in (value-1,value-2)")
	assert.Equal(t, map[VpaID]*Vpa{
		{"namespace-1", "vpa-1"}: cluster.Vpas[VpaID{"namespace-1", "vpa-1"}],
		{"namespace-1", "vpa-2"}: cluster.Vpas[VpaID{"namespace-1", "vpa-2"}]},
		pod.MatchingVpas)
	// Delete the first VPA from the Pod. Expect that it will still
	// have the other one.
	assert.NoError(t, cluster.DeleteVpa(VpaID{"namespace-1", "vpa-1"}))
	assert.Equal(t, map[VpaID]*Vpa{
		{"namespace-1", "vpa-2"}: cluster.Vpas[VpaID{"namespace-1", "vpa-2"}]},
		pod.MatchingVpas)
	// Delete the other VPA. The Pod is no longer vertically-scaled by anyone.
	assert.NoError(t, cluster.DeleteVpa(VpaID{"namespace-1", "vpa-2"}))
	assert.Empty(t, pod.MatchingVpas)
}

// Verifies that a VPA with an empty selector (matching all pods) matches a pod
// with labels as well as a pod with no labels.
func TestEmptySelector(t *testing.T) {
	cluster := NewClusterState()
	// Create a VPA with an empty selector (matching all pods).
	addVpa(cluster, testVpaID, "")
	// Create a pod with labels.
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	// Create a pod without labels.
	anotherPodID := PodID{"namespace-1", "pod-2"}
	cluster.AddOrUpdatePod(anotherPodID, emptyLabels, apiv1.PodRunning)
	// Both pods should be matched by the VPA.
	assert.Equal(t, map[VpaID]*Vpa{testVpaID: cluster.Vpas[testVpaID]},
		cluster.Pods[testPodID].MatchingVpas)
	assert.Equal(t, map[VpaID]*Vpa{testVpaID: cluster.Vpas[testVpaID]},
		cluster.Pods[anotherPodID].MatchingVpas)
}
