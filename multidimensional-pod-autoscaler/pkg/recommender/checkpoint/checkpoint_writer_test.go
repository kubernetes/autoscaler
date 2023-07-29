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
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/model"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"

	"github.com/stretchr/testify/assert"
)

// TODO: Extract these constants to a common test module.
var (
	testPodID1 = vpa_model.PodID{
		Namespace: "namespace-1",
		PodName:   "pod-1",
	}
	testContainerID1 = vpa_model.ContainerID{
		PodID:         testPodID1,
		ContainerName: "container-1",
	}
	testMpaID1 = model.MpaID{
		Namespace: "namespace-1",
		MpaName:   "mpa-1",
	}
	testLabels      = map[string]string{"label-1": "value-1"}
	testSelectorStr = "label-1 = value-1"
	testRequest     = vpa_model.Resources{
		vpa_model.ResourceCPU:    vpa_model.CPUAmountFromCores(3.14),
		vpa_model.ResourceMemory: vpa_model.MemoryAmountFromBytes(3.14e9),
	}
)

const testGcPeriod = time.Minute

func addMpa(t *testing.T, cluster *model.ClusterState, mpaID model.MpaID, selector string) *model.Mpa {
	var apiObject mpa_types.MultidimPodAutoscaler
	apiObject.Namespace = mpaID.Namespace
	apiObject.Name = mpaID.MpaName
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	err := cluster.AddOrUpdateMpa(&apiObject, parsedSelector)
	if err != nil {
		t.Fatalf("AddOrUpdateMpa() failed: %v", err)
	}
	return cluster.Mpas[mpaID]
}

func TestMergeContainerStateForCheckpointDropsRecentMemoryPeak(t *testing.T) {
	cluster := model.NewClusterState(testGcPeriod)
	cluster.AddOrUpdatePod(testPodID1, testLabels, v1.PodRunning)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID1, testRequest))
	container := cluster.GetContainer(testContainerID1)

	timeNow := time.Unix(1, 0)
	container.AddSample(&vpa_model.ContainerUsageSample{
		MeasureStart: timeNow,
		Usage:        vpa_model.MemoryAmountFromBytes(1024 * 1024 * 1024),
		Request:      testRequest[vpa_model.ResourceMemory],
		Resource:     vpa_model.ResourceMemory,
	})
	mpa := addMpa(t, cluster, testMpaID1, testSelectorStr)

	// Verify that the current peak is excluded from the aggregation.
	aggregateContainerStateMap := buildAggregateContainerStateMap(mpa, cluster, timeNow)
	if assert.Contains(t, aggregateContainerStateMap, "container-1") {
		assert.True(t, aggregateContainerStateMap["container-1"].AggregateMemoryPeaks.IsEmpty(),
			"Current peak was not excluded from the aggregation.")
	}
	// Verify that an old peak is not excluded from the aggregation.
	timeNow = timeNow.Add(vpa_model.GetAggregationsConfig().MemoryAggregationInterval)
	aggregateContainerStateMap = buildAggregateContainerStateMap(mpa, cluster, timeNow)
	if assert.Contains(t, aggregateContainerStateMap, "container-1") {
		assert.False(t, aggregateContainerStateMap["container-1"].AggregateMemoryPeaks.IsEmpty(),
			"Old peak should not be excluded from the aggregation.")
	}
}

func TestIsFetchingHistory(t *testing.T) {

	testCases := []struct {
		mpa               model.Mpa
		isFetchingHistory bool
	}{
		{
			mpa:               model.Mpa{},
			isFetchingHistory: false,
		},
		{
			mpa: model.Mpa{
				PodSelector: nil,
				Conditions: map[mpa_types.MultidimPodAutoscalerConditionType]mpa_types.MultidimPodAutoscalerCondition{
					mpa_types.FetchingHistory: {
						Type:   mpa_types.FetchingHistory,
						Status: v1.ConditionFalse,
					},
				},
			},
			isFetchingHistory: false,
		},
		{
			mpa: model.Mpa{
				PodSelector: nil,
				Conditions: map[mpa_types.MultidimPodAutoscalerConditionType]mpa_types.MultidimPodAutoscalerCondition{
					mpa_types.FetchingHistory: {
						Type:   mpa_types.FetchingHistory,
						Status: v1.ConditionTrue,
					},
				},
			},
			isFetchingHistory: true,
		},
	}

	for _, tc := range testCases {
		assert.Equalf(t, tc.isFetchingHistory, isFetchingHistory(&tc.mpa), "%+v should have %v as isFetchingHistoryResult", tc.mpa, tc.isFetchingHistory)
	}
}

func TestGetMpasToCheckpointSorts(t *testing.T) {

	time1 := time.Unix(10000, 0)
	time2 := time.Unix(20000, 0)

	genMpaID := func(index int) model.MpaID {
		return model.MpaID{
			MpaName: fmt.Sprintf("mpa-%d", index),
		}
	}
	mpa0 := &model.Mpa{
		ID: genMpaID(0),
	}
	mpa1 := &model.Mpa{
		ID:                genMpaID(1),
		CheckpointWritten: time1,
	}
	mpa2 := &model.Mpa{
		ID:                genMpaID(2),
		CheckpointWritten: time2,
	}
	mpas := make(map[model.MpaID]*model.Mpa)
	addMpa := func(mpa *model.Mpa) {
		mpas[mpa.ID] = mpa
	}
	addMpa(mpa2)
	addMpa(mpa0)
	addMpa(mpa1)
	result := getMpasToCheckpoint(mpas)
	assert.Equal(t, genMpaID(0), result[0].ID)
	assert.Equal(t, genMpaID(1), result[1].ID)
	assert.Equal(t, genMpaID(2), result[2].ID)

}
