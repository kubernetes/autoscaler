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
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"

	"github.com/stretchr/testify/assert"
)

// TODO: Extract these constants to a common test module.
var (
	testPodID1       = model.PodID{"namespace-1", "pod-1"}
	testContainerID1 = model.ContainerID{testPodID1, "container-1"}
	testVpaID1       = model.VpaID{"namespace-1", "vpa-1"}
	testLabels       = map[string]string{"label-1": "value-1"}
	testSelectorStr  = "label-1 = value-1"
	testRequest      = model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(3.14),
		model.ResourceMemory: model.MemoryAmountFromBytes(3.14e9),
	}
)

func addVpa(t *testing.T, cluster *model.ClusterState, vpaID model.VpaID, selector string) *model.Vpa {
	var apiObject vpa_types.VerticalPodAutoscaler
	apiObject.Namespace = vpaID.Namespace
	apiObject.Name = vpaID.VpaName
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	err := cluster.AddOrUpdateVpa(&apiObject, parsedSelector)
	if err != nil {
		t.Fatalf("AddOrUpdateVpa() failed: %v", err)
	}
	return cluster.Vpas[vpaID]
}

func TestMergeContainerStateForCheckpointDropsRecentMemoryPeak(t *testing.T) {
	cluster := model.NewClusterState()
	cluster.AddOrUpdatePod(testPodID1, testLabels, v1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID1, testRequest))
	container := cluster.GetContainer(testContainerID1)

	timeNow := time.Unix(1, 0)
	container.AddSample(&model.ContainerUsageSample{
		timeNow, model.MemoryAmountFromBytes(1024 * 1024 * 1024), testRequest[model.ResourceMemory], model.ResourceMemory})
	vpa := addVpa(t, cluster, testVpaID1, testSelectorStr)

	// Verify that the current peak is excluded from the aggregation.
	aggregateContainerStateMap := buildAggregateContainerStateMap(vpa, cluster, timeNow)
	if assert.Contains(t, aggregateContainerStateMap, "container-1") {
		assert.True(t, aggregateContainerStateMap["container-1"].AggregateMemoryPeaks.IsEmpty(),
			"Current peak was not excluded from the aggregation.")
	}
	// Verify that an old peak is not excluded from the aggregation.
	timeNow = timeNow.Add(model.MemoryAggregationInterval)
	aggregateContainerStateMap = buildAggregateContainerStateMap(vpa, cluster, timeNow)
	if assert.Contains(t, aggregateContainerStateMap, "container-1") {
		assert.False(t, aggregateContainerStateMap["container-1"].AggregateMemoryPeaks.IsEmpty(),
			"Old peak should not be excluded from the aggregation.")
	}
}

func TestIsFetchingHistory(t *testing.T) {

	testCases := []struct {
		vpa               model.Vpa
		isFetchingHistory bool
	}{
		{
			vpa:               model.Vpa{},
			isFetchingHistory: false,
		},
		{
			vpa: model.Vpa{
				PodSelector: nil,
				Conditions: map[vpa_types.VerticalPodAutoscalerConditionType]vpa_types.VerticalPodAutoscalerCondition{
					vpa_types.FetchingHistory: {
						Type:   vpa_types.FetchingHistory,
						Status: v1.ConditionFalse,
					},
				},
			},
			isFetchingHistory: false,
		},
		{
			vpa: model.Vpa{
				PodSelector: nil,
				Conditions: map[vpa_types.VerticalPodAutoscalerConditionType]vpa_types.VerticalPodAutoscalerCondition{
					vpa_types.FetchingHistory: {
						Type:   vpa_types.FetchingHistory,
						Status: v1.ConditionTrue,
					},
				},
			},
			isFetchingHistory: true,
		},
	}

	for _, tc := range testCases {
		assert.Equalf(t, tc.isFetchingHistory, isFetchingHistory(&tc.vpa), "%+v should have %v as isFetchingHistoryResult", tc.vpa, tc.isFetchingHistory)
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
