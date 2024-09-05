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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

var (
	testPodID       = PodID{"namespace-1", "pod-1"}
	testPodID3      = PodID{"namespace-1", "pod-3"}
	testPodID4      = PodID{"namespace-1", "pod-4"}
	testContainerID = ContainerID{testPodID, "container-1"}
	testVpaID       = VpaID{"namespace-1", "vpa-1"}
	testAnnotations = vpaAnnotationsMap{"key-1": "value-1"}
	testLabels      = map[string]string{"label-1": "value-1"}
	emptyLabels     = map[string]string{}
	testSelectorStr = "label-1 = value-1"
	testTargetRef   = &autoscaling.CrossVersionObjectReference{
		Kind:       "kind-1",
		Name:       "name-1",
		APIVersion: "apiVersion-1",
	}
	testControllerKey = &controllerfetcher.ControllerKeyWithAPIVersion{
		ControllerKey: controllerfetcher.ControllerKey{
			Kind:      "kind-1",
			Name:      "name-1",
			Namespace: "namespace-1",
		},
		ApiVersion: "apiVersion-1",
	}
	testControllerFetcher = &fakeControllerFetcher{
		key: testControllerKey,
		err: nil,
	}
)

type fakeControllerFetcher struct {
	key *controllerfetcher.ControllerKeyWithAPIVersion
	err error
}

func (f *fakeControllerFetcher) FindTopMostWellKnownOrScalable(_ context.Context, _ *controllerfetcher.ControllerKeyWithAPIVersion) (*controllerfetcher.ControllerKeyWithAPIVersion, error) {
	return f.key, f.err
}

const testGcPeriod = time.Minute

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
	cluster := NewClusterState(testGcPeriod)
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(makeTestUsageSample()))

	// Verify that the sample was aggregated into the container stats.
	containerStats := cluster.Pods[testPodID].Containers["container-1"]
	assert.Equal(t, testTimestamp, containerStats.LastCPUSampleStart)
}

func TestClusterGCAggregateContainerStateDeletesOld(t *testing.T) {
	ctx := context.Background()

	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	vpa := addTestVpa(cluster)
	addTestPod(cluster)

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	usageSample := makeTestUsageSample()

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(usageSample))

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	// AggregateContainerState are valid for 8 days since last sample
	cluster.garbageCollectAggregateCollectionStates(ctx, usageSample.MeasureStart.Add(9*24*time.Hour), testControllerFetcher)

	// AggregateContainerState should be deleted from both cluster and vpa
	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, vpa.aggregateContainerStates)
}

func TestClusterGCAggregateContainerStateDeletesOldEmpty(t *testing.T) {
	ctx := context.Background()

	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	vpa := addTestVpa(cluster)
	addTestPod(cluster)

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	// No usage samples added.

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	assert.Len(t, cluster.aggregateStateMap, 1)
	var creationTime time.Time
	for _, aggregateState := range cluster.aggregateStateMap {
		creationTime = aggregateState.CreationTime
	}

	// Verify empty aggregate states are not removed right away.
	cluster.garbageCollectAggregateCollectionStates(ctx, creationTime.Add(1*time.Minute), testControllerFetcher) // AggregateContainerState should be deleted from both cluster and vpa
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	// AggregateContainerState are valid for 8 days since creation
	cluster.garbageCollectAggregateCollectionStates(ctx, creationTime.Add(9*24*time.Hour), testControllerFetcher)

	// AggregateContainerState should be deleted from both cluster and vpa
	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, vpa.aggregateContainerStates)
}

func TestClusterGCAggregateContainerStateDeletesEmptyInactiveWithoutController(t *testing.T) {
	ctx := context.Background()

	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	vpa := addTestVpa(cluster)
	pod := addTestPod(cluster)
	// Controller Fetcher returns nil, meaning that there is no corresponding controller alive.
	controller := &fakeControllerFetcher{
		key: nil,
		err: nil,
	}

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	// No usage samples added.

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	cluster.garbageCollectAggregateCollectionStates(ctx, testTimestamp, controller)

	// AggregateContainerState should not be deleted as the pod is still active.
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	cluster.Pods[pod.ID].Phase = apiv1.PodSucceeded
	cluster.garbageCollectAggregateCollectionStates(ctx, testTimestamp, controller)

	// AggregateContainerState should be empty as the pod is no longer active, controller is not alive
	// and there are no usage samples.
	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, vpa.aggregateContainerStates)
}

func TestClusterGCAggregateContainerStateLeavesEmptyInactiveWithController(t *testing.T) {
	ctx := context.Background()

	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	vpa := addTestVpa(cluster)
	pod := addTestPod(cluster)
	// Controller Fetcher returns existing controller, meaning that there is a corresponding controller alive.
	controller := testControllerFetcher

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	// No usage samples added.

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	cluster.garbageCollectAggregateCollectionStates(ctx, testTimestamp, controller)

	// AggregateContainerState should not be deleted as the pod is still active.
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	cluster.Pods[pod.ID].Phase = apiv1.PodSucceeded
	cluster.garbageCollectAggregateCollectionStates(ctx, testTimestamp, controller)

	// AggregateContainerState should not be deleted as the controller is still alive.
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)
}

func TestClusterGCAggregateContainerStateLeavesValid(t *testing.T) {
	ctx := context.Background()

	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	vpa := addTestVpa(cluster)
	addTestPod(cluster)

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	usageSample := makeTestUsageSample()

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(usageSample))

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	// AggregateContainerState are valid for 8 days since last sample
	cluster.garbageCollectAggregateCollectionStates(ctx, usageSample.MeasureStart.Add(7*24*time.Hour), testControllerFetcher)

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)
}

func TestAddSampleAfterAggregateContainerStateGCed(t *testing.T) {
	ctx := context.Background()

	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	vpa := addTestVpa(cluster)
	pod := addTestPod(cluster)
	addTestContainer(t, cluster)

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	usageSample := makeTestUsageSample()

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(usageSample))

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.Contains(t, vpa.aggregateContainerStates, aggregateStateKey)

	// AggregateContainerState are invalid after 8 days since last sample
	gcTimestamp := usageSample.MeasureStart.Add(10 * 24 * time.Hour)
	cluster.garbageCollectAggregateCollectionStates(ctx, gcTimestamp, testControllerFetcher)

	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, vpa.aggregateContainerStates)
	assert.Contains(t, pod.Containers, testContainerID.ContainerName)

	newUsageSample := &ContainerUsageSampleWithKey{ContainerUsageSample{
		MeasureStart: gcTimestamp.Add(1 * time.Hour),
		Usage:        usageSample.Usage,
		Request:      usageSample.Request,
		Resource:     usageSample.Resource},
		testContainerID}
	// Add usage sample to the container again.
	assert.NoError(t, cluster.AddSample(newUsageSample))

	assert.Contains(t, vpa.aggregateContainerStates, aggregateStateKey)
}

func TestClusterGCRateLimiting(t *testing.T) {
	ctx := context.Background()

	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	usageSample := makeTestUsageSample()
	sampleExpireTime := usageSample.MeasureStart.Add(9 * 24 * time.Hour)
	// AggregateContainerState are valid for 8 days since last sample but this run
	// doesn't remove the sample, because we didn't add it yet.
	cluster.RateLimitedGarbageCollectAggregateCollectionStates(ctx, sampleExpireTime, testControllerFetcher)
	vpa := addTestVpa(cluster)
	addTestPod(cluster)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(usageSample))
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	// Sample is expired but this run doesn't remove it yet, because less than testGcPeriod
	// elapsed since the previous run.
	cluster.RateLimitedGarbageCollectAggregateCollectionStates(ctx, sampleExpireTime.Add(testGcPeriod/2), testControllerFetcher)
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, vpa.aggregateContainerStates)

	// AggregateContainerState should be deleted from both cluster and vpa
	cluster.RateLimitedGarbageCollectAggregateCollectionStates(ctx, sampleExpireTime.Add(2*testGcPeriod), testControllerFetcher)
	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, vpa.aggregateContainerStates)
}

func TestClusterRecordOOM(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
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
	cluster := NewClusterState(testGcPeriod)
	err := cluster.AddSample(makeTestUsageSample())
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")

	err = cluster.RecordOOM(testContainerID, time.Unix(0, 0), ResourceAmount(10))
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")

	err = cluster.AddOrUpdateContainer(testContainerID, testRequest)
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")
}

func addVpa(cluster *ClusterState, id VpaID, annotations vpaAnnotationsMap, selector string, targetRef *autoscaling.CrossVersionObjectReference) *Vpa {
	apiObject := test.VerticalPodAutoscaler().WithNamespace(id.Namespace).
		WithName(id.VpaName).WithContainer(testContainerID.ContainerName).WithAnnotations(annotations).WithTargetRef(targetRef).Get()
	return addVpaObject(cluster, id, apiObject, selector)
}

func addVpaObject(cluster *ClusterState, id VpaID, vpa *vpa_types.VerticalPodAutoscaler, selector string) *Vpa {
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	err := cluster.AddOrUpdateVpa(vpa, parsedSelector)
	if err != nil {
		klog.Fatalf("AddOrUpdateVpa() failed: %v", err)
	}
	return cluster.Vpas[id]
}

func addTestVpa(cluster *ClusterState) *Vpa {
	return addVpa(cluster, testVpaID, testAnnotations, testSelectorStr, testTargetRef)
}

func addTestPod(cluster *ClusterState) *PodState {
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	return cluster.Pods[testPodID]
}

func addTestContainer(t *testing.T, cluster *ClusterState) *ContainerState {
	err := cluster.AddOrUpdateContainer(testContainerID, testRequest)
	assert.NoError(t, err)
	return cluster.GetContainer(testContainerID)
}

// Creates a VPA followed by a matching pod. Verifies that the links between
// VPA, the container and the aggregation are set correctly.
func TestAddVpaThenAddPod(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	vpa := addTestVpa(cluster)
	assert.Empty(t, vpa.aggregateContainerStates)
	addTestPod(cluster)
	addTestContainer(t, cluster)
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.Contains(t, vpa.aggregateContainerStates, aggregateStateKey)
}

// Creates a pod followed by a matching VPA. Verifies that the links between
// VPA, the container and the aggregation are set correctly.
func TestAddPodThenAddVpa(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	addTestPod(cluster)
	addTestContainer(t, cluster)
	vpa := addTestVpa(cluster)
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.Contains(t, vpa.aggregateContainerStates, aggregateStateKey)
}

// Creates a VPA and a matching pod, then change the pod labels such that it is
// no longer matched by the VPA. Verifies that the links between the pod and the
// VPA are removed.
func TestChangePodLabels(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	vpa := addTestVpa(cluster)
	addTestPod(cluster)
	addTestContainer(t, cluster)
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.Contains(t, vpa.aggregateContainerStates, aggregateStateKey)
	// Update Pod labels to no longer match the VPA.
	cluster.AddOrUpdatePod(testPodID, emptyLabels, apiv1.PodRunning)
	aggregateStateKey = cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.NotContains(t, vpa.aggregateContainerStates, aggregateStateKey)
}

// Creates a VPA and verifies that annotation updates work properly.
func TestUpdateAnnotations(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	vpa := addTestVpa(cluster)
	// Verify that the annotations match the test annotations.
	assert.Equal(t, vpa.Annotations, testAnnotations)
	// Update the annotations (non-empty).
	annotations := vpaAnnotationsMap{"key-2": "value-2"}
	vpa = addVpa(cluster, testVpaID, annotations, testSelectorStr, testTargetRef)
	assert.Equal(t, vpa.Annotations, annotations)
	// Update the annotations (empty).
	annotations = vpaAnnotationsMap{}
	vpa = addVpa(cluster, testVpaID, annotations, testSelectorStr, testTargetRef)
	assert.Equal(t, vpa.Annotations, annotations)
}

// Creates a VPA and a matching pod, then change the VPA pod selector 3 times:
// first such that it still matches the pod, then such that it no longer matches
// the pod, finally such that it matches the pod again. Verifies that the links
// between the pod and the VPA are updated correctly each time.
func TestUpdatePodSelector(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	addTestPod(cluster)
	addTestContainer(t, cluster)

	// Update the VPA selector such that it still matches the Pod.
	vpa := addVpa(cluster, testVpaID, testAnnotations, "label-1 in (value-1,value-2)", testTargetRef)
	assert.Contains(t, vpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(testContainerID))

	// Update the VPA selector to no longer match the Pod.
	vpa = addVpa(cluster, testVpaID, testAnnotations, "label-1 = value-2", testTargetRef)
	assert.NotContains(t, vpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(testContainerID))

	// Update the VPA selector to match the Pod again.
	vpa = addVpa(cluster, testVpaID, testAnnotations, "label-1 = value-1", testTargetRef)
	assert.Contains(t, vpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(testContainerID))
}

// Test setting ResourcePolicy and UpdatePolicy on adding or updating VPA object
func TestAddOrUpdateVPAPolicies(t *testing.T) {
	testVpaBuilder := test.VerticalPodAutoscaler().WithName(testVpaID.VpaName).
		WithNamespace(testVpaID.Namespace).WithContainer(testContainerID.ContainerName)
	updateModeAuto := vpa_types.UpdateModeAuto
	updateModeOff := vpa_types.UpdateModeOff
	scalingModeAuto := vpa_types.ContainerScalingModeAuto
	scalingModeOff := vpa_types.ContainerScalingModeOff
	cases := []struct {
		name                string
		oldVpa              *vpa_types.VerticalPodAutoscaler
		newVpa              *vpa_types.VerticalPodAutoscaler
		resourcePolicy      *vpa_types.PodResourcePolicy
		expectedScalingMode *vpa_types.ContainerScalingMode
		expectedUpdateMode  *vpa_types.UpdateMode
		expectedAPIVersion  string
	}{
		{
			name:   "Defaults to auto",
			oldVpa: nil,
			newVpa: testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).Get(),
			// Container scaling mode is a separate concept from update mode in the VPA object,
			// hence the UpdateModeOff does not influence container scaling mode here.
			expectedScalingMode: &scalingModeAuto,
			expectedUpdateMode:  &updateModeOff,
			expectedAPIVersion:  "v1",
		}, {
			name:   "Default scaling mode set to Off",
			oldVpa: nil,
			newVpa: testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "*",
						Mode:          &scalingModeOff,
					},
				},
			},
			expectedScalingMode: &scalingModeOff,
			expectedUpdateMode:  &updateModeAuto,
			expectedAPIVersion:  "v1",
		}, {
			name:   "Explicit scaling mode set to Off",
			oldVpa: nil,
			newVpa: testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: testContainerID.ContainerName,
						Mode:          &scalingModeOff,
					},
				},
			},
			expectedScalingMode: &scalingModeOff,
			expectedUpdateMode:  &updateModeAuto,
			expectedAPIVersion:  "v1",
		}, {
			name:   "Other container has explicit scaling mode Off",
			oldVpa: nil,
			newVpa: testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "other-container",
						Mode:          &scalingModeOff,
					},
				},
			},
			expectedScalingMode: &scalingModeAuto,
			expectedUpdateMode:  &updateModeAuto,
			expectedAPIVersion:  "v1",
		}, {
			name:   "Scaling mode to default Off",
			oldVpa: testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			newVpa: testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "*",
						Mode:          &scalingModeOff,
					},
				},
			},
			expectedScalingMode: &scalingModeOff,
			expectedUpdateMode:  &updateModeAuto,
			expectedAPIVersion:  "v1",
		}, {
			name:   "Scaling mode to explicit Off",
			oldVpa: testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			newVpa: testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: testContainerID.ContainerName,
						Mode:          &scalingModeOff,
					},
				},
			},
			expectedScalingMode: &scalingModeOff,
			expectedUpdateMode:  &updateModeAuto,
			expectedAPIVersion:  "v1",
		},
		// Tests checking changes to UpdateMode.
		{
			name:                "UpdateMode from Off to Auto",
			oldVpa:              testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).Get(),
			newVpa:              testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			expectedScalingMode: &scalingModeAuto,
			expectedUpdateMode:  &updateModeAuto,
			expectedAPIVersion:  "v1",
		}, {
			name:                "UpdateMode from Auto to Off",
			oldVpa:              testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			newVpa:              testVpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).Get(),
			expectedScalingMode: &scalingModeAuto,
			expectedUpdateMode:  &updateModeOff,
			expectedAPIVersion:  "v1",
		},
		// Test different API versions being recorded.
		// Note that this path for testing the apiVersions is not actively exercised
		// in a running recommender. The GroupVersion is cleared before it reaches
		// the recommenders code. These tests only test the propagation of version
		// changes. When introducing new api versions that need to be differentiated
		// in logic and/or metrics a dedicated detection mechanism is needed for
		// those new versions. We can not get this information from the api request:
		// https://github.com/kubernetes/kubernetes/pull/59264#issuecomment-362579495
		{
			name:                "Record APIVersion v1",
			oldVpa:              nil,
			newVpa:              testVpaBuilder.WithGroupVersion(metav1.GroupVersion(vpa_types.SchemeGroupVersion)).Get(),
			expectedScalingMode: &scalingModeAuto,
			expectedAPIVersion:  "v1",
		},
		{
			name:   "Record APIVersion v1beta2",
			oldVpa: nil,
			newVpa: testVpaBuilder.WithGroupVersion(metav1.GroupVersion{
				Group:   vpa_types.SchemeGroupVersion.Group,
				Version: "v1beta2",
			}).Get(),
			expectedScalingMode: &scalingModeAuto,
			expectedAPIVersion:  "v1beta2",
		},
		{
			name:   "Record APIVersion v1beta1",
			oldVpa: nil,
			newVpa: testVpaBuilder.WithGroupVersion(metav1.GroupVersion{
				Group:   vpa_types.SchemeGroupVersion.Group,
				Version: "v1beta1",
			}).Get(),
			expectedScalingMode: &scalingModeAuto,
			expectedAPIVersion:  "v1beta1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cluster := NewClusterState(testGcPeriod)
			addTestPod(cluster)
			addTestContainer(t, cluster)
			if tc.oldVpa != nil {
				oldVpa := addVpaObject(cluster, testVpaID, tc.oldVpa, testSelectorStr)
				if !assert.Contains(t, cluster.Vpas, testVpaID) {
					t.FailNow()
				}
				assert.Len(t, oldVpa.aggregateContainerStates, 1, "Expected one container aggregation in VPA %v", testVpaID)
				for containerName, aggregation := range oldVpa.aggregateContainerStates {
					assert.Equal(t, &scalingModeAuto, aggregation.GetScalingMode(), "Unexpected scaling mode for container %s", containerName)
				}
			}
			tc.newVpa.Spec.ResourcePolicy = tc.resourcePolicy
			addVpaObject(cluster, testVpaID, tc.newVpa, testSelectorStr)
			vpa, found := cluster.Vpas[testVpaID]
			if !assert.True(t, found, "VPA %+v not found in cluster state.", testVpaID) {
				t.FailNow()
			}
			assert.Equal(t, tc.expectedUpdateMode, vpa.UpdateMode)
			assert.Len(t, vpa.aggregateContainerStates, 1, "Expected one container aggregation in VPA %v", testVpaID)
			for containerName, aggregation := range vpa.aggregateContainerStates {
				assert.Equal(t, tc.expectedUpdateMode, aggregation.UpdateMode, "Unexpected update mode for container %s", containerName)
				assert.Equal(t, tc.expectedScalingMode, aggregation.GetScalingMode(), "Unexpected scaling mode for container %s", containerName)
			}
			assert.Equal(t, tc.expectedAPIVersion, vpa.APIVersion)
		})
	}
}

// Verify that two copies of the same AggregateStateKey are equal.
func TestEqualAggregateStateKey(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
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

	cluster := NewClusterState(testGcPeriod)
	cluster.AddOrUpdatePod(podID1, testLabels, apiv1.PodRunning)
	cluster.AddOrUpdatePod(podID2, testLabels, apiv1.PodRunning)
	err := cluster.AddOrUpdateContainer(containerID1, testRequest)
	assert.NoError(t, err)
	err = cluster.AddOrUpdateContainer(containerID2, testRequest)
	assert.NoError(t, err)

	// Expect only one aggregation to be created.
	assert.Equal(t, 1, len(cluster.aggregateStateMap))
}

// Verify that two identical containers in different namespaces are not aggregated together.
func TestTwoPodsWithDifferentNamespaces(t *testing.T) {
	podID1 := PodID{"namespace-1", "foo-pod"}
	podID2 := PodID{"namespace-2", "foo-pod"}
	containerID1 := ContainerID{podID1, "foo-container"}
	containerID2 := ContainerID{podID2, "foo-container"}

	cluster := NewClusterState(testGcPeriod)
	cluster.AddOrUpdatePod(podID1, testLabels, apiv1.PodRunning)
	cluster.AddOrUpdatePod(podID2, testLabels, apiv1.PodRunning)
	err := cluster.AddOrUpdateContainer(containerID1, testRequest)
	assert.NoError(t, err)
	err = cluster.AddOrUpdateContainer(containerID2, testRequest)
	assert.NoError(t, err)

	// Expect two separate aggregations to be created.
	assert.Equal(t, 2, len(cluster.aggregateStateMap))
	// Expect only one entry to be present in the labels set map.
	assert.Equal(t, 1, len(cluster.labelSetMap))
}

// Verifies that a VPA with an empty selector (matching all pods) matches a pod
// with labels as well as a pod with no labels.
func TestEmptySelector(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	// Create a VPA with an empty selector (matching all pods).
	vpa := addVpa(cluster, testVpaID, testAnnotations, "", testTargetRef)
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

func TestRecordRecommendation(t *testing.T) {
	cases := []struct {
		name               string
		recommendation     *vpa_types.RecommendedPodResources
		lastLogged         time.Time
		now                time.Time
		expectedEmpty      bool
		expectedLastLogged time.Time
		expectedError      error
	}{
		{
			name:           "VPA has recommendation",
			recommendation: test.Recommendation().WithContainer("test").WithTarget("100m", "200G").Get(),
			now:            testTimestamp,
			expectedEmpty:  false,
			expectedError:  nil,
		}, {
			name:           "VPA recommendation appears",
			recommendation: test.Recommendation().WithContainer("test").WithTarget("100m", "200G").Get(),
			lastLogged:     testTimestamp.Add(-10 * time.Minute),
			now:            testTimestamp,
			expectedEmpty:  false,
			expectedError:  nil,
		}, {
			name:               "VPA recommendation missing",
			recommendation:     &vpa_types.RecommendedPodResources{},
			lastLogged:         testTimestamp.Add(-10 * time.Minute),
			now:                testTimestamp,
			expectedEmpty:      true,
			expectedLastLogged: testTimestamp.Add(-10 * time.Minute),
			expectedError:      nil,
		}, {
			name:               "VPA recommendation missing and needs logging",
			recommendation:     &vpa_types.RecommendedPodResources{},
			lastLogged:         testTimestamp.Add(-40 * time.Minute),
			now:                testTimestamp,
			expectedEmpty:      true,
			expectedLastLogged: testTimestamp,
			expectedError:      fmt.Errorf("VPA namespace-1/vpa-1 is missing recommendation for more than %v", RecommendationMissingMaxDuration),
		}, {
			name:               "VPA recommendation disappears",
			recommendation:     &vpa_types.RecommendedPodResources{},
			now:                testTimestamp,
			expectedEmpty:      true,
			expectedLastLogged: testTimestamp,
			expectedError:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cluster := NewClusterState(testGcPeriod)
			vpa := addVpa(cluster, testVpaID, testAnnotations, testSelectorStr, testTargetRef)
			cluster.Vpas[testVpaID].Recommendation = tc.recommendation
			if !tc.lastLogged.IsZero() {
				cluster.EmptyVPAs[testVpaID] = tc.lastLogged
			}

			err := cluster.RecordRecommendation(vpa, tc.now)
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
				if tc.expectedEmpty {
					assert.Contains(t, cluster.EmptyVPAs, testVpaID)
					assert.Equal(t, cluster.EmptyVPAs[testVpaID], tc.expectedLastLogged)
				} else {
					assert.NotContains(t, cluster.EmptyVPAs, testVpaID)
				}
			}
		})
	}
}

type podDesc struct {
	id     PodID
	labels labels.Set
	phase  apiv1.PodPhase
}

func TestGetActiveMatchingPods(t *testing.T) {
	cases := []struct {
		name         string
		vpaSelector  string
		pods         []podDesc
		expectedPods []PodID
	}{
		{
			name:         "No pods",
			vpaSelector:  testSelectorStr,
			pods:         []podDesc{},
			expectedPods: []PodID{},
		}, {
			name:        "Matching pod",
			vpaSelector: testSelectorStr,
			pods: []podDesc{
				{
					id:     testPodID,
					labels: testLabels,
					phase:  apiv1.PodRunning,
				},
			},
			expectedPods: []PodID{testPodID},
		}, {
			name:        "Matching pod is inactive",
			vpaSelector: testSelectorStr,
			pods: []podDesc{
				{
					id:     testPodID,
					labels: testLabels,
					phase:  apiv1.PodFailed,
				},
			},
			expectedPods: []PodID{testPodID},
		}, {
			name:        "No matching pods",
			vpaSelector: testSelectorStr,
			pods: []podDesc{
				{
					id:     testPodID,
					labels: emptyLabels,
					phase:  apiv1.PodRunning,
				}, {
					id:     PodID{Namespace: "different-than-vpa", PodName: "pod-1"},
					labels: testLabels,
					phase:  apiv1.PodRunning,
				},
			},
			expectedPods: []PodID{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cluster := NewClusterState(testGcPeriod)
			vpa := addVpa(cluster, testVpaID, testAnnotations, tc.vpaSelector, testTargetRef)
			for _, pod := range tc.pods {
				cluster.AddOrUpdatePod(pod.id, pod.labels, pod.phase)
			}
			pods := cluster.GetMatchingPods(vpa)
			assert.ElementsMatch(t, tc.expectedPods, pods)
		})
	}
}

func TestVPAWithMatchingPods(t *testing.T) {
	cases := []struct {
		name          string
		vpaSelector   string
		pods          []podDesc
		expectedMatch int
	}{
		{
			name:          "No pods",
			vpaSelector:   testSelectorStr,
			pods:          []podDesc{},
			expectedMatch: 0,
		},
		{
			name:        "VPA with matching pod",
			vpaSelector: testSelectorStr,
			pods: []podDesc{
				{
					testPodID,
					testLabels,
					apiv1.PodRunning,
				},
			},
			expectedMatch: 1,
		},
		{
			name:        "No matching pod",
			vpaSelector: testSelectorStr,
			pods: []podDesc{
				{
					testPodID,
					emptyLabels,
					apiv1.PodRunning,
				},
			},
			expectedMatch: 0,
		},
		{
			name:        "VPA with 2 matching pods, 1 not matching",
			vpaSelector: testSelectorStr,
			pods: []podDesc{
				{
					testPodID,
					emptyLabels, // does not match VPA
					apiv1.PodRunning,
				},
				{
					testPodID3,
					testLabels,
					apiv1.PodRunning,
				},
				{
					testPodID4,
					testLabels,
					apiv1.PodRunning,
				},
			},
			expectedMatch: 2,
		},
	}
	// Run with adding VPA first
	for _, tc := range cases {
		t.Run(tc.name+", VPA first", func(t *testing.T) {
			cluster := NewClusterState(testGcPeriod)
			vpa := addVpa(cluster, testVpaID, testAnnotations, tc.vpaSelector, testTargetRef)
			for _, podDesc := range tc.pods {
				cluster.AddOrUpdatePod(podDesc.id, podDesc.labels, podDesc.phase)
				containerID := ContainerID{testPodID, "foo"}
				assert.NoError(t, cluster.AddOrUpdateContainer(containerID, testRequest))
			}
			assert.Equal(t, tc.expectedMatch, cluster.Vpas[vpa.ID].PodCount)
		})
	}
	// Run with adding Pods first
	for _, tc := range cases {
		t.Run(tc.name+", Pods first", func(t *testing.T) {
			cluster := NewClusterState(testGcPeriod)
			for _, podDesc := range tc.pods {
				cluster.AddOrUpdatePod(podDesc.id, podDesc.labels, podDesc.phase)
				containerID := ContainerID{testPodID, "foo"}
				assert.NoError(t, cluster.AddOrUpdateContainer(containerID, testRequest))
			}
			vpa := addVpa(cluster, testVpaID, testAnnotations, tc.vpaSelector, testTargetRef)
			assert.Equal(t, tc.expectedMatch, cluster.Vpas[vpa.ID].PodCount)
		})
	}
}
