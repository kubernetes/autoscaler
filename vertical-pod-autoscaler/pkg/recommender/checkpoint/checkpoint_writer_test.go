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

package checkpoint

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	core "k8s.io/client-go/testing"
	"k8s.io/klog/v2"
	klogtest "k8s.io/klog/v2/test"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	fakeautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1/fake"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

// TODO: Extract these constants to a common test module.
var (
	testPodID1 = model.PodID{
		Namespace: "namespace-1",
		PodName:   "pod-1",
	}
	testContainerID1 = model.ContainerID{
		PodID:         testPodID1,
		ContainerName: "container-1",
	}
	testVpaID1 = model.VpaID{
		Namespace: "namespace-1",
		VpaName:   "vpa-1",
	}
	testLabels      = map[string]string{"label-1": "value-1"}
	testSelectorStr = "label-1 = value-1"
	testRequest     = model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(3.14),
		model.ResourceMemory: model.MemoryAmountFromBytes(3.14e9),
	}
)

const testGcPeriod = time.Minute

func addVpa(t *testing.T, cluster model.ClusterState, vpaID model.VpaID, selector string) *model.Vpa {
	var apiObject vpa_types.VerticalPodAutoscaler
	apiObject.Namespace = vpaID.Namespace
	apiObject.Name = vpaID.VpaName
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	err := cluster.AddOrUpdateVpa(&apiObject, parsedSelector)
	if err != nil {
		t.Fatalf("AddOrUpdateVpa() failed: %v", err)
	}
	return cluster.VPAs()[vpaID]
}

func TestMergeContainerStateForCheckpointDropsRecentMemoryPeak(t *testing.T) {
	for _, tc := range []struct {
		name      string
		replicas  int
		unusedKey bool
	}{
		{name: "single container", replicas: 1},
		{name: "replicas sharing an aggregate key", replicas: 2},
		{name: "unused key with the same container name", replicas: 2, unusedKey: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cluster := model.NewClusterState(testGcPeriod)
			interval := model.GetAggregationsConfig().MemoryAggregationIntervalDuration
			firstSampleTime := time.Unix(1700000000, 0)
			recentSampleTime := firstSampleTime.Add(interval)
			oldPeak := model.MemoryAmountFromBytes(1024 * 1024 * 1024)
			recentPeak := model.MemoryAmountFromBytes(2 * 1024 * 1024 * 1024)

			addMemorySample := func(id model.ContainerID, timestamp time.Time, usage model.ResourceAmount) {
				t.Helper()
				if err := cluster.AddSample(&model.ContainerUsageSampleWithKey{
					Container: id,
					ContainerUsageSample: model.ContainerUsageSample{
						MeasureStart: timestamp,
						Usage:        usage,
						Resource:     model.ResourceMemory,
					},
				}); err != nil {
					t.Fatal(err)
				}
			}
			for i := range tc.replicas {
				podID := model.PodID{Namespace: testPodID1.Namespace, PodName: fmt.Sprintf("replica-%d", i)}
				cluster.AddOrUpdatePod(podID, testLabels, corev1.PodRunning)
				id := model.ContainerID{PodID: podID, ContainerName: testContainerID1.ContainerName}
				if err := cluster.AddOrUpdateContainer(id, testRequest); err != nil {
					t.Fatal(err)
				}
				addMemorySample(id, firstSampleTime, oldPeak)
				addMemorySample(id, recentSampleTime, recentPeak)
			}
			if tc.unusedKey {
				podID := model.PodID{Namespace: testPodID1.Namespace, PodName: "unmatched"}
				cluster.AddOrUpdatePod(podID, labels.Set{"label-1": "other-value"}, corev1.PodRunning)
				id := model.ContainerID{PodID: podID, ContainerName: testContainerID1.ContainerName}
				if err := cluster.AddOrUpdateContainer(id, testRequest); err != nil {
					t.Fatal(err)
				}
				// Reuse the historical peaks bucket. Subtracting this unrelated
				// current peak would incorrectly remove the VPAs retained history.
				addMemorySample(id, recentSampleTime, oldPeak)
			}
			vpa := addVpa(t, cluster, testVpaID1, testSelectorStr)
			groups := groupContainersByAggregate(cluster)
			if tc.unusedKey {
				assert.Len(t, groups, 2)
			} else {
				assert.Len(t, groups, 1)
			}
			for key, group := range groups {
				if vpa.UsesAggregation(key) {
					assert.Len(t, group, tc.replicas)
				}
			}

			for _, completed := range []bool{false, true} {
				name := "current peaks excluded"
				now := recentSampleTime
				if completed {
					name = "completed peaks retained"
					now = now.Add(interval)
				}
				t.Run(name, func(t *testing.T) {
					expected := model.NewAggregateContainerState()
					for range tc.replicas {
						// Memory peaks enter the histogram at their windows end.
						expected.AggregateMemoryPeaks.AddSample(model.BytesFromMemoryAmount(oldPeak), 1, recentSampleTime)
						if completed {
							expected.AggregateMemoryPeaks.AddSample(model.BytesFromMemoryAmount(recentPeak), 1, recentSampleTime.Add(interval))
						}
					}
					states := buildAggregateContainerStateMap(vpa, groups, now)
					state, exists := states[testContainerID1.ContainerName]
					if !assert.True(t, exists, "missing container history") || !assert.NotNil(t, state) {
						return
					}
					assert.True(t, expected.AggregateMemoryPeaks.Equals(state.AggregateMemoryPeaks),
						"unexpected checkpoint peaks: got %s, want %s", state.AggregateMemoryPeaks, expected.AggregateMemoryPeaks)
				})
			}
		})
	}
}

func TestIsFetchingHistory(t *testing.T) {
	testCases := []struct {
		vpa               *model.Vpa
		isFetchingHistory bool
	}{
		{
			vpa:               &model.Vpa{},
			isFetchingHistory: false,
		},
		{
			vpa: func() *model.Vpa {
				vpa := model.NewVpa(model.VpaID{}, nil, time.Time{})
				vpa.SetConditionsMap(map[vpa_types.VerticalPodAutoscalerConditionType]vpa_types.VerticalPodAutoscalerCondition{
					vpa_types.FetchingHistory: {
						Type:   vpa_types.FetchingHistory,
						Status: corev1.ConditionFalse,
					},
				})
				return vpa
			}(),
			isFetchingHistory: false,
		},
		{
			vpa: func() *model.Vpa {
				vpa := model.NewVpa(model.VpaID{}, nil, time.Time{})
				vpa.SetConditionsMap(map[vpa_types.VerticalPodAutoscalerConditionType]vpa_types.VerticalPodAutoscalerCondition{
					vpa_types.FetchingHistory: {
						Type:   vpa_types.FetchingHistory,
						Status: corev1.ConditionTrue,
					},
				})
				return vpa
			}(),
			isFetchingHistory: true,
		},
	}

	for _, tc := range testCases {
		assert.Equalf(t, tc.isFetchingHistory, isFetchingHistory(tc.vpa), "%+v should have %v as isFetchingHistoryResult", tc.vpa, tc.isFetchingHistory)
	}
}

func TestGetVpasToCheckpointSorts(t *testing.T) {
	time1 := time.Unix(10000, 0)
	time2 := time.Unix(20000, 0)

	genVpaID := func(index int) model.VpaID {
		return model.VpaID{
			VpaName: fmt.Sprintf("vpa-%d", index),
		}
	}
	vpa0 := &model.Vpa{
		ID: genVpaID(0),
	}
	vpa1 := &model.Vpa{
		ID:                genVpaID(1),
		CheckpointWritten: time1,
	}
	vpa2 := &model.Vpa{
		ID:                genVpaID(2),
		CheckpointWritten: time2,
	}
	vpas := make(map[model.VpaID]*model.Vpa)
	addVpa := func(vpa *model.Vpa) {
		vpas[vpa.ID] = vpa
	}
	addVpa(vpa2)
	addVpa(vpa0)
	addVpa(vpa1)
	result := getVpasToCheckpoint(vpas)
	assert.Equal(t, genVpaID(0), result[0].ID)
	assert.Equal(t, genVpaID(1), result[1].ID)
	assert.Equal(t, genVpaID(2), result[2].ID)
}

func TestStoreCheckpointsMakesProgressEvenForCancelledContext(t *testing.T) {
	klogtest.InitKlog(t)
	tmpLogBuffer := bytes.NewBuffer(nil)
	klog.SetOutput(tmpLogBuffer)

	concurrentWorkers := 2

	// immediately cancel the context to check if at least checkpoints for `concurrentWorkers` number of VPAs get written
	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	clusterState := model.NewClusterState(testGcPeriod)

	// prepare ClusterState with 5 VPAs referencing Pods
	vpaBuilder := test.VerticalPodAutoscaler().WithContainer("container-1").WithContainer("container-2").WithNamespace("test-namespace")

	for i := range 5 {
		targetRef := &autoscalingv1.CrossVersionObjectReference{
			Kind:       "Pod",
			Name:       fmt.Sprintf("pod-%d", i),
			APIVersion: "apps/v1",
		}
		labelSelector, _ := labels.Parse(fmt.Sprintf("app=pod-%d", i))
		vpa := vpaBuilder.WithName(fmt.Sprintf("vpa-%d", i)).WithTargetRef(targetRef).Get()
		err := clusterState.AddOrUpdateVpa(vpa, labelSelector)
		assert.NoError(t, err)
	}

	// prepare ClusterState with 5 pods that have 2 containers each
	for i := range 5 {
		podID := model.PodID{
			Namespace: "test-namespace",
			PodName:   fmt.Sprintf("pod-%d", i),
		}
		podLabels := map[string]string{"app": fmt.Sprintf("pod-%d", i)}
		clusterState.AddOrUpdatePod(podID, podLabels, corev1.PodRunning)
		for j := range 2 {
			containerID := model.ContainerID{
				PodID:         podID,
				ContainerName: fmt.Sprintf("container-%d", j),
			}
			err := clusterState.AddOrUpdateContainer(containerID, testRequest)
			assert.NoError(t, err)
		}
	}

	patchedCheckpoints := []string{}
	checkpointClient := &fakeautoscalingv1.FakeAutoscalingV1{Fake: &core.Fake{}}
	checkpointClient.AddReactor("patch", "verticalpodautoscalercheckpoints", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		patchAction := action.(core.PatchAction)
		name := patchAction.GetName()
		time.Sleep(2 * time.Millisecond) // Simulate some delay in patching, such that we can test the timeout
		patchedCheckpoints = append(patchedCheckpoints, name)

		return true, nil, nil
	})

	writer := NewCheckpointWriter(clusterState, checkpointClient)
	writer.StoreCheckpoints(ctx, concurrentWorkers)

	// Because we have 2 concurrent workers, expect 2 VPAs to get processed. Each worker picks a VPA to process before checking if the context has been cancelled.
	// Each VPA has 2 Containers, therefore we expect 4 Checkpoints to be written
	assert.Equal(t, 4, len(patchedCheckpoints), "Expected 4 checkpoints to be written, but got %d", len(patchedCheckpoints))

	assert.Contains(t, tmpLogBuffer.String(), "context canceled")
}
