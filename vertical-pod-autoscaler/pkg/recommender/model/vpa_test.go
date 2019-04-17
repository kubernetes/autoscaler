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

package model

import (
	"testing"
	"time"

	core "k8s.io/api/core/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"

	"github.com/stretchr/testify/assert"
)

var (
	anyTime = time.Unix(0, 0)
)

func TestMergeAggregateContainerState(t *testing.T) {

	containersInitialAggregateState := ContainerNameToAggregateStateMap{}
	containersInitialAggregateState["test"] = NewAggregateContainerState()
	vpa := NewVpa(VpaID{}, nil, anyTime)
	vpa.ContainersInitialAggregateState = containersInitialAggregateState

	containerNameToAggregateStateMap := ContainerNameToAggregateStateMap{}
	vpa.MergeCheckpointedState(containerNameToAggregateStateMap)

	assert.Contains(t, containerNameToAggregateStateMap, "test")
}

func TestUpdateConditions(t *testing.T) {
	cases := []struct {
		name               string
		podsMatched        bool
		hasRecommendation  bool
		expectedConditions []vpa_types.VerticalPodAutoscalerCondition
	}{
		{
			name:              "Has recommendation",
			podsMatched:       true,
			hasRecommendation: true,
			expectedConditions: []vpa_types.VerticalPodAutoscalerCondition{
				{
					Type:    vpa_types.RecommendationProvided,
					Status:  core.ConditionTrue,
					Reason:  "",
					Message: "",
				},
			},
		}, {
			name:              "Has recommendation but no pods matched",
			podsMatched:       false,
			hasRecommendation: true,
			expectedConditions: []vpa_types.VerticalPodAutoscalerCondition{
				{
					Type:    vpa_types.RecommendationProvided,
					Status:  core.ConditionTrue,
					Reason:  "",
					Message: "",
				}, {
					Type:    vpa_types.NoPodsMatched,
					Status:  core.ConditionTrue,
					Reason:  "NoPodsMatched",
					Message: "No live pods match this VPA object",
				},
			},
		}, {
			name:              "No recommendation but pods matched",
			podsMatched:       true,
			hasRecommendation: false,
			expectedConditions: []vpa_types.VerticalPodAutoscalerCondition{
				{
					Type:    vpa_types.RecommendationProvided,
					Status:  core.ConditionFalse,
					Reason:  "",
					Message: "",
				},
			},
		}, {
			name:              "No recommendation no pods matched",
			podsMatched:       false,
			hasRecommendation: false,
			expectedConditions: []vpa_types.VerticalPodAutoscalerCondition{
				{
					Type:    vpa_types.RecommendationProvided,
					Status:  core.ConditionFalse,
					Reason:  "NoPodsMatched",
					Message: "No live pods match this VPA object",
				}, {
					Type:    vpa_types.NoPodsMatched,
					Status:  core.ConditionTrue,
					Reason:  "NoPodsMatched",
					Message: "No live pods match this VPA object",
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			containerName := "container"
			vpa := NewVpa(VpaID{Namespace: "test-namespace", VpaName: "my-facourite-vpa"}, labels.Nothing(), time.Unix(0, 0))
			if tc.hasRecommendation {
				vpa.Recommendation = test.Recommendation().WithContainer(containerName).WithTarget("5", "200").Get()
			}
			vpa.UpdateConditions(tc.podsMatched)
			for _, condition := range tc.expectedConditions {
				assert.Contains(t, vpa.Conditions, condition.Type)
				actualCondition := vpa.Conditions[condition.Type]
				assert.Equal(t, condition.Status, actualCondition.Status, "Condition: %v", condition.Type)
				assert.Equal(t, condition.Reason, actualCondition.Reason, "Condition: %v", condition.Type)
				assert.Equal(t, condition.Message, actualCondition.Message, "Condition: %v", condition.Type)
			}
		})
	}
}
