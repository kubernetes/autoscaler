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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/test"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/controller_fetcher"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	"k8s.io/klog/v2"
)

var (
	testPodID       = vpa_model.PodID{Namespace: "namespace-1", PodName: "pod-1"}
	testPodID3      = vpa_model.PodID{Namespace: "namespace-1", PodName: "pod-3"}
	testPodID4      = vpa_model.PodID{Namespace: "namespace-1", PodName: "pod-4"}
	testContainerID = vpa_model.ContainerID{PodID: testPodID, ContainerName: "container-1"}
	testMpaID       = MpaID{"namespace-1", "mpa-1"}
	testAnnotations = mpaAnnotationsMap{"key-1": "value-1"}
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
	mapper restmapper.DeferredDiscoveryRESTMapper
	scaleNamespacer scale.ScalesGetter
}

func (f *fakeControllerFetcher) GetRESTMappings(groupKind schema.GroupKind) ([]*apimeta.RESTMapping, error) {
	return f.mapper.RESTMappings(groupKind)
}

func (f *fakeControllerFetcher) Scales(namespace string) (scale.ScaleInterface) {
	return f.scaleNamespacer.Scales(namespace)
}

func (f *fakeControllerFetcher) FindTopMostWellKnownOrScalable(controller *controllerfetcher.ControllerKeyWithAPIVersion) (*controllerfetcher.ControllerKeyWithAPIVersion, error) {
	return f.key, f.err
}

const testGcPeriod = time.Minute

func makeTestUsageSample() *ContainerUsageSampleWithKey {
	return &ContainerUsageSampleWithKey{vpa_model.ContainerUsageSample{
		MeasureStart: testTimestamp,
		Usage:        1.0,
		Request:      testRequest[vpa_model.ResourceCPU],
		Resource:     vpa_model.ResourceCPU},
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
	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	mpa := addTestMpa(cluster)
	addTestPod(cluster)

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	usageSample := makeTestUsageSample()

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(usageSample))

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	// AggegateContainerState are valid for 8 days since last sample
	cluster.garbageCollectAggregateCollectionStates(usageSample.MeasureStart.Add(9*24*time.Hour), testControllerFetcher)

	// AggegateContainerState should be deleted from both cluster and mpa
	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, mpa.aggregateContainerStates)
}

func TestClusterGCAggregateContainerStateDeletesOldEmpty(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	mpa := addTestMpa(cluster)
	addTestPod(cluster)

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	// No usage samples added.

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	assert.Len(t, cluster.aggregateStateMap, 1)
	var creationTime time.Time
	for _, aggregateState := range cluster.aggregateStateMap {
		creationTime = aggregateState.CreationTime
	}

	// Verify empty aggregate states are not removed right away.
	cluster.garbageCollectAggregateCollectionStates(creationTime.Add(1*time.Minute), testControllerFetcher) // AggegateContainerState should be deleted from both cluster and mpa
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	// AggegateContainerState are valid for 8 days since creation
	cluster.garbageCollectAggregateCollectionStates(creationTime.Add(9*24*time.Hour), testControllerFetcher)

	// AggegateContainerState should be deleted from both cluster and mpa
	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, mpa.aggregateContainerStates)
}

func TestClusterGCAggregateContainerStateDeletesEmptyInactiveWithoutController(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	mpa := addTestMpa(cluster)
	pod := addTestPod(cluster)
	// Controller Fetcher returns nil, meaning that there is no corresponding controller alive.
	controller := &fakeControllerFetcher{
		key: nil,
		err: nil,
	}

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	// No usage samples added.

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	cluster.garbageCollectAggregateCollectionStates(testTimestamp, controller)

	// AggegateContainerState should not be deleted as the pod is still active.
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	cluster.Pods[pod.ID].Phase = apiv1.PodSucceeded
	cluster.garbageCollectAggregateCollectionStates(testTimestamp, controller)

	// AggegateContainerState should be empty as the pod is no longer active, controller is not alive
	// and there are no usage samples.
	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, mpa.aggregateContainerStates)
}

func TestClusterGCAggregateContainerStateLeavesEmptyInactiveWithController(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	mpa := addTestMpa(cluster)
	pod := addTestPod(cluster)
	// Controller Fetcher returns existing controller, meaning that there is a corresponding controller alive.
	controller := testControllerFetcher

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	// No usage samples added.

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	cluster.garbageCollectAggregateCollectionStates(testTimestamp, controller)

	// AggegateContainerState should not be deleted as the pod is still active.
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	cluster.Pods[pod.ID].Phase = apiv1.PodSucceeded
	cluster.garbageCollectAggregateCollectionStates(testTimestamp, controller)

	// AggegateContainerState should not be delated as the controller is still alive.
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)
}

func TestClusterGCAggregateContainerStateLeavesValid(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	mpa := addTestMpa(cluster)
	addTestPod(cluster)

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	usageSample := makeTestUsageSample()

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(usageSample))

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	// AggegateContainerState are valid for 8 days since last sample
	cluster.garbageCollectAggregateCollectionStates(usageSample.MeasureStart.Add(7*24*time.Hour), testControllerFetcher)

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)
}

func TestAddSampleAfterAggregateContainerStateGCed(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	mpa := addTestMpa(cluster)
	pod := addTestPod(cluster)
	addTestContainer(t, cluster)

	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))
	usageSample := makeTestUsageSample()

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(usageSample))

	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.Contains(t, mpa.aggregateContainerStates, aggregateStateKey)

	// AggegateContainerState are invalid after 8 days since last sample
	gcTimestamp := usageSample.MeasureStart.Add(10 * 24 * time.Hour)
	cluster.garbageCollectAggregateCollectionStates(gcTimestamp, testControllerFetcher)

	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, mpa.aggregateContainerStates)
	assert.Contains(t, pod.Containers, testContainerID.ContainerName)

	newUsageSample := &ContainerUsageSampleWithKey{vpa_model.ContainerUsageSample{
		MeasureStart: gcTimestamp.Add(1 * time.Hour),
		Usage:        usageSample.Usage,
		Request:      usageSample.Request,
		Resource:     usageSample.Resource},
		testContainerID}
	// Add usage sample to the container again.
	assert.NoError(t, cluster.AddSample(newUsageSample))

	assert.Contains(t, mpa.aggregateContainerStates, aggregateStateKey)
}

func TestClusterGCRateLimiting(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	usageSample := makeTestUsageSample()
	sampleExpireTime := usageSample.MeasureStart.Add(9 * 24 * time.Hour)
	// AggegateContainerState are valid for 8 days since last sample but this run
	// doesn't remove the sample, because we didn't add it yet.
	cluster.RateLimitedGarbageCollectAggregateCollectionStates(sampleExpireTime, testControllerFetcher)
	mpa := addTestMpa(cluster)
	addTestPod(cluster)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))

	// Add a usage sample to the container.
	assert.NoError(t, cluster.AddSample(usageSample))
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	// Sample is expired but this run doesn't remove it yet, because less than testGcPeriod
	// elapsed since the previous run.
	cluster.RateLimitedGarbageCollectAggregateCollectionStates(sampleExpireTime.Add(testGcPeriod/2), testControllerFetcher)
	assert.NotEmpty(t, cluster.aggregateStateMap)
	assert.NotEmpty(t, mpa.aggregateContainerStates)

	// AggegateContainerState should be deleted from both cluster and mpa
	cluster.RateLimitedGarbageCollectAggregateCollectionStates(sampleExpireTime.Add(2*testGcPeriod), testControllerFetcher)
	assert.Empty(t, cluster.aggregateStateMap)
	assert.Empty(t, mpa.aggregateContainerStates)
}

func TestClusterRecordOOM(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewClusterState(testGcPeriod)
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID, testRequest))

	// RecordOOM
	assert.NoError(t, cluster.RecordOOM(testContainerID, time.Unix(0, 0), vpa_model.ResourceAmount(10)))

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

	err = cluster.RecordOOM(testContainerID, time.Unix(0, 0), vpa_model.ResourceAmount(10))
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")

	err = cluster.AddOrUpdateContainer(testContainerID, testRequest)
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")
}

func addMpa(cluster *ClusterState, id MpaID, annotations mpaAnnotationsMap, selector string, scaleTargetRef *autoscaling.CrossVersionObjectReference) *Mpa {
	apiObject := test.MultidimPodAutoscaler().WithNamespace(id.Namespace).
		WithName(id.MpaName).WithContainer(testContainerID.ContainerName).WithAnnotations(annotations).WithScaleTargetRef(scaleTargetRef).Get()
	return addMpaObject(cluster, id, apiObject, selector)
}

func addMpaObject(cluster *ClusterState, id MpaID, mpa *mpa_types.MultidimPodAutoscaler, selector string) *Mpa {
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	err := cluster.AddOrUpdateMpa(mpa, parsedSelector)
	if err != nil {
		klog.Fatalf("AddOrUpdateMpa() failed: %v", err)
	}
	return cluster.Mpas[id]
}

func addTestMpa(cluster *ClusterState) *Mpa {
	return addMpa(cluster, testMpaID, testAnnotations, testSelectorStr, testTargetRef)
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

// Creates a MPA followed by a matching pod. Verifies that the links between
// MPA, the container and the aggregation are set correctly.
func TestAddMpaThenAddPod(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	mpa := addTestMpa(cluster)
	assert.Empty(t, mpa.aggregateContainerStates)
	addTestPod(cluster)
	addTestContainer(t, cluster)
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.Contains(t, mpa.aggregateContainerStates, aggregateStateKey)
}

// Creates a pod followed by a matching MPA. Verifies that the links between
// MPA, the container and the aggregation are set correctly.
func TestAddPodThenAddMpa(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	addTestPod(cluster)
	addTestContainer(t, cluster)
	mpa := addTestMpa(cluster)
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.Contains(t, mpa.aggregateContainerStates, aggregateStateKey)
}

// Creates a MPA and a matching pod, then change the pod labels such that it is
// no longer matched by the MPA. Verifies that the links between the pod and the
// MPA are removed.
func TestChangePodLabels(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	mpa := addTestMpa(cluster)
	addTestPod(cluster)
	addTestContainer(t, cluster)
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.Contains(t, mpa.aggregateContainerStates, aggregateStateKey)
	// Update Pod labels to no longer match the MPA.
	cluster.AddOrUpdatePod(testPodID, emptyLabels, apiv1.PodRunning)
	aggregateStateKey = cluster.aggregateStateKeyForContainerID(testContainerID)
	assert.NotContains(t, mpa.aggregateContainerStates, aggregateStateKey)
}

// Creates a MPA and verifies that annotation updates work properly.
func TestUpdateAnnotations(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	mpa := addTestMpa(cluster)
	// Verify that the annotations match the test annotations.
	assert.Equal(t, mpa.Annotations, testAnnotations)
	// Update the annotations (non-empty).
	annotations := mpaAnnotationsMap{"key-2": "value-2"}
	mpa = addMpa(cluster, testMpaID, annotations, testSelectorStr, testTargetRef)
	assert.Equal(t, mpa.Annotations, annotations)
	// Update the annotations (empty).
	annotations = mpaAnnotationsMap{}
	mpa = addMpa(cluster, testMpaID, annotations, testSelectorStr, testTargetRef)
	assert.Equal(t, mpa.Annotations, annotations)
}

// Creates a MPA and a matching pod, then change the MPA pod selector 3 times:
// first such that it still matches the pod, then such that it no longer matches
// the pod, finally such that it matches the pod again. Verifies that the links
// between the pod and the MPA are updated correctly each time.
func TestUpdatePodSelector(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	addTestPod(cluster)
	addTestContainer(t, cluster)

	// Update the MPA selector such that it still matches the Pod.
	mpa := addMpa(cluster, testMpaID, testAnnotations, "label-1 in (value-1,value-2)", testTargetRef)
	assert.Contains(t, mpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(testContainerID))

	// Update the MPA selector to no longer match the Pod.
	mpa = addMpa(cluster, testMpaID, testAnnotations, "label-1 = value-2", testTargetRef)
	assert.NotContains(t, mpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(testContainerID))

	// Update the MPA selector to match the Pod again.
	mpa = addMpa(cluster, testMpaID, testAnnotations, "label-1 = value-1", testTargetRef)
	assert.Contains(t, mpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(testContainerID))
}

// Test setting ResourcePolicy and UpdatePolicy on adding or updating MPA object
func TestAddOrUpdateMPAPolicies(t *testing.T) {
	testMpaBuilder := test.MultidimPodAutoscaler().WithName(testMpaID.MpaName).
		WithNamespace(testMpaID.Namespace).WithContainer(testContainerID.ContainerName)
	updateModeAuto := vpa_types.UpdateModeAuto
	updateModeOff := vpa_types.UpdateModeOff
	scalingModeAuto := vpa_types.ContainerScalingModeAuto
	scalingModeOff := vpa_types.ContainerScalingModeOff
	cases := []struct {
		name                string
		oldMpa              *mpa_types.MultidimPodAutoscaler
		newMpa              *mpa_types.MultidimPodAutoscaler
		resourcePolicy      *vpa_types.PodResourcePolicy
		expectedScalingMode *vpa_types.ContainerScalingMode
		expectedUpdateMode  *vpa_types.UpdateMode
	}{
		{
			name:   "Defaults to auto",
			oldMpa: nil,
			newMpa: testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).Get(),
			// Container scaling mode is a separate concept from update mode in the MPA object,
			// hence the UpdateModeOff does not influence container scaling mode here.
			expectedScalingMode: &scalingModeAuto,
			expectedUpdateMode:  &updateModeOff,
		}, {
			name:   "Default scaling mode set to Off",
			oldMpa: nil,
			newMpa: testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
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
		}, {
			name:   "Explicit scaling mode set to Off",
			oldMpa: nil,
			newMpa: testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
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
		}, {
			name:   "Other container has explicit scaling mode Off",
			oldMpa: nil,
			newMpa: testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
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
		}, {
			name:   "Scaling mode to default Off",
			oldMpa: testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			newMpa: testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
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
		}, {
			name:   "Scaling mode to explicit Off",
			oldMpa: testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			newMpa: testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
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
		},
		// Tests checking changes to UpdateMode.
		{
			name:                "UpdateMode from Off to Auto",
			oldMpa:              testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).Get(),
			newMpa:              testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			expectedScalingMode: &scalingModeAuto,
			expectedUpdateMode:  &updateModeAuto,
		}, {
			name:                "UpdateMode from Auto to Off",
			oldMpa:              testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).Get(),
			newMpa:              testMpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).Get(),
			expectedScalingMode: &scalingModeAuto,
			expectedUpdateMode:  &updateModeOff,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cluster := NewClusterState(testGcPeriod)
			addTestPod(cluster)
			addTestContainer(t, cluster)
			if tc.oldMpa != nil {
				oldMpa := addMpaObject(cluster, testMpaID, tc.oldMpa, testSelectorStr)
				if !assert.Contains(t, cluster.Mpas, testMpaID) {
					t.FailNow()
				}
				assert.Len(t, oldMpa.aggregateContainerStates, 1, "Expected one container aggregation in MPA %v", testMpaID)
				for containerName, aggregation := range oldMpa.aggregateContainerStates {
					assert.Equal(t, &scalingModeAuto, aggregation.GetScalingMode(), "Unexpected scaling mode for container %s", containerName)
				}
			}
			tc.newMpa.Spec.ResourcePolicy = tc.resourcePolicy
			addMpaObject(cluster, testMpaID, tc.newMpa, testSelectorStr)
			mpa, found := cluster.Mpas[testMpaID]
			if !assert.True(t, found, "MPA %+v not found in cluster state.", testMpaID) {
				t.FailNow()
			}
			assert.Equal(t, tc.expectedUpdateMode, mpa.UpdateMode)
			assert.Len(t, mpa.aggregateContainerStates, 1, "Expected one container aggregation in MPA %v", testMpaID)
			for containerName, aggregation := range mpa.aggregateContainerStates {
				assert.Equal(t, tc.expectedUpdateMode, aggregation.UpdateMode, "Unexpected update mode for container %s", containerName)
				assert.Equal(t, tc.expectedScalingMode, aggregation.GetScalingMode(), "Unexpected scaling mode for container %s", containerName)
			}
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
	podID1 := vpa_model.PodID{Namespace: "namespace-1", PodName: "pod-1"}
	podID2 := vpa_model.PodID{Namespace: "namespace-1", PodName: "pod-2"}
	containerID1 := vpa_model.ContainerID{PodID: podID1, ContainerName: "foo-container"}
	containerID2 := vpa_model.ContainerID{PodID: podID2, ContainerName: "foo-container"}

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
	podID1 := vpa_model.PodID{Namespace: "namespace-1", PodName: "foo-pod"}
	podID2 := vpa_model.PodID{Namespace: "namespace-2", PodName: "foo-pod"}
	containerID1 := vpa_model.ContainerID{PodID: podID1, ContainerName: "foo-container"}
	containerID2 := vpa_model.ContainerID{PodID: podID2, ContainerName: "foo-container"}

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

// Verifies that a MPA with an empty selector (matching all pods) matches a pod
// with labels as well as a pod with no labels.
func TestEmptySelector(t *testing.T) {
	cluster := NewClusterState(testGcPeriod)
	// Create a MPA with an empty selector (matching all pods).
	mpa := addMpa(cluster, testMpaID, testAnnotations, "", testTargetRef)
	// Create a pod with labels. Add a container.
	cluster.AddOrUpdatePod(testPodID, testLabels, apiv1.PodRunning)
	containerID1 := vpa_model.ContainerID{PodID: testPodID, ContainerName: "foo"}
	assert.NoError(t, cluster.AddOrUpdateContainer(containerID1, testRequest))

	// Create a pod without labels. Add a container.
	anotherPodID := vpa_model.PodID{Namespace: "namespace-1", PodName: "pod-2"}
	cluster.AddOrUpdatePod(anotherPodID, emptyLabels, apiv1.PodRunning)
	containerID2 := vpa_model.ContainerID{PodID: anotherPodID, ContainerName: "foo"}
	assert.NoError(t, cluster.AddOrUpdateContainer(containerID2, testRequest))

	// Both pods should be matched by the MPA.
	assert.Contains(t, mpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(containerID1))
	assert.Contains(t, mpa.aggregateContainerStates, cluster.aggregateStateKeyForContainerID(containerID2))
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
			name:           "MPA has recommendation",
			recommendation: test.Recommendation().WithContainer("test").WithTarget("100m", "200G").Get(),
			now:            testTimestamp,
			expectedEmpty:  false,
			expectedError:  nil,
		}, {
			name:           "MPA recommendation appears",
			recommendation: test.Recommendation().WithContainer("test").WithTarget("100m", "200G").Get(),
			lastLogged:     testTimestamp.Add(-10 * time.Minute),
			now:            testTimestamp,
			expectedEmpty:  false,
			expectedError:  nil,
		}, {
			name:               "MPA recommendation missing",
			recommendation:     &vpa_types.RecommendedPodResources{},
			lastLogged:         testTimestamp.Add(-10 * time.Minute),
			now:                testTimestamp,
			expectedEmpty:      true,
			expectedLastLogged: testTimestamp.Add(-10 * time.Minute),
			expectedError:      nil,
		}, {
			name:               "MPA recommendation missing and needs logging",
			recommendation:     &vpa_types.RecommendedPodResources{},
			lastLogged:         testTimestamp.Add(-40 * time.Minute),
			now:                testTimestamp,
			expectedEmpty:      true,
			expectedLastLogged: testTimestamp,
			expectedError:      fmt.Errorf("MPA namespace-1/mpa-1 is missing recommendation for more than %v", RecommendationMissingMaxDuration),
		}, {
			name:               "MPA recommendation disappears",
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
			mpa := addMpa(cluster, testMpaID, testAnnotations, testSelectorStr, testTargetRef)
			cluster.Mpas[testMpaID].Recommendation = tc.recommendation
			if !tc.lastLogged.IsZero() {
				cluster.EmptyMPAs[testMpaID] = tc.lastLogged
			}

			err := cluster.RecordRecommendation(mpa, tc.now)
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
				if tc.expectedEmpty {
					assert.Contains(t, cluster.EmptyMPAs, testMpaID)
					assert.Equal(t, cluster.EmptyMPAs[testMpaID], tc.expectedLastLogged)
				} else {
					assert.NotContains(t, cluster.EmptyMPAs, testMpaID)
				}
			}
		})
	}
}

type podDesc struct {
	id     vpa_model.PodID
	labels labels.Set
	phase  apiv1.PodPhase
}

func TestGetActiveMatchingPods(t *testing.T) {
	cases := []struct {
		name         string
		mpaSelector  string
		pods         []podDesc
		expectedPods []vpa_model.PodID
	}{
		{
			name:         "No pods",
			mpaSelector:  testSelectorStr,
			pods:         []podDesc{},
			expectedPods: []vpa_model.PodID{},
		}, {
			name:        "Matching pod",
			mpaSelector: testSelectorStr,
			pods: []podDesc{
				{
					id:     testPodID,
					labels: testLabels,
					phase:  apiv1.PodRunning,
				},
			},
			expectedPods: []vpa_model.PodID{testPodID},
		}, {
			name:        "Matching pod is inactive",
			mpaSelector: testSelectorStr,
			pods: []podDesc{
				{
					id:     testPodID,
					labels: testLabels,
					phase:  apiv1.PodFailed,
				},
			},
			expectedPods: []vpa_model.PodID{testPodID},
		}, {
			name:        "No matching pods",
			mpaSelector: testSelectorStr,
			pods: []podDesc{
				{
					id:     testPodID,
					labels: emptyLabels,
					phase:  apiv1.PodRunning,
				}, {
					id:     vpa_model.PodID{Namespace: "different-than-mpa", PodName: "pod-1"},
					labels: testLabels,
					phase:  apiv1.PodRunning,
				},
			},
			expectedPods: []vpa_model.PodID{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cluster := NewClusterState(testGcPeriod)
			mpa := addMpa(cluster, testMpaID, testAnnotations, tc.mpaSelector, testTargetRef)
			for _, pod := range tc.pods {
				cluster.AddOrUpdatePod(pod.id, pod.labels, pod.phase)
			}
			pods := cluster.GetMatchingPods(mpa)
			assert.ElementsMatch(t, tc.expectedPods, pods)
		})
	}
}

func TestMPAWithMatchingPods(t *testing.T) {
	cases := []struct {
		name          string
		mpaSelector   string
		pods          []podDesc
		expectedMatch int
	}{
		{
			name:          "No pods",
			mpaSelector:   testSelectorStr,
			pods:          []podDesc{},
			expectedMatch: 0,
		},
		{
			name:        "MPA with matching pod",
			mpaSelector: testSelectorStr,
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
			mpaSelector: testSelectorStr,
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
			name:        "MPA with 2 matching pods, 1 not matching",
			mpaSelector: testSelectorStr,
			pods: []podDesc{
				{
					testPodID,
					emptyLabels, // does not match MPA
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
	// Run with adding MPA first
	for _, tc := range cases {
		t.Run(tc.name+", MPA first", func(t *testing.T) {
			cluster := NewClusterState(testGcPeriod)
			mpa := addMpa(cluster, testMpaID, testAnnotations, tc.mpaSelector, testTargetRef)
			for _, podDesc := range tc.pods {
				cluster.AddOrUpdatePod(podDesc.id, podDesc.labels, podDesc.phase)
				containerID := vpa_model.ContainerID{PodID: testPodID, ContainerName: "foo"}
				assert.NoError(t, cluster.AddOrUpdateContainer(containerID, testRequest))
			}
			assert.Equal(t, tc.expectedMatch, cluster.Mpas[mpa.ID].PodCount)
		})
	}
	// Run with adding Pods first
	for _, tc := range cases {
		t.Run(tc.name+", Pods first", func(t *testing.T) {
			cluster := NewClusterState(testGcPeriod)
			for _, podDesc := range tc.pods {
				cluster.AddOrUpdatePod(podDesc.id, podDesc.labels, podDesc.phase)
				containerID := vpa_model.ContainerID{PodID: testPodID, ContainerName: "foo"}
				assert.NoError(t, cluster.AddOrUpdateContainer(containerID, testRequest))
			}
			mpa := addMpa(cluster, testMpaID, testAnnotations, tc.mpaSelector, testTargetRef)
			assert.Equal(t, tc.expectedMatch, cluster.Mpas[mpa.ID].PodCount)
		})
	}
}
