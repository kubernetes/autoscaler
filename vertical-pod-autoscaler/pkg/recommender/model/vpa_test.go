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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	labels "k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
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
		expectedAbsent     []vpa_types.VerticalPodAutoscalerConditionType
	}{
		{
			name:              "Has recommendation",
			podsMatched:       true,
			hasRecommendation: true,
			expectedConditions: []vpa_types.VerticalPodAutoscalerCondition{
				{
					Type:    vpa_types.RecommendationProvided,
					Status:  corev1.ConditionTrue,
					Reason:  "",
					Message: "",
				},
			},
			expectedAbsent: []vpa_types.VerticalPodAutoscalerConditionType{vpa_types.NoPodsMatched},
		}, {
			name:              "Has recommendation but no pods matched",
			podsMatched:       false,
			hasRecommendation: true,
			expectedConditions: []vpa_types.VerticalPodAutoscalerCondition{
				{
					Type:    vpa_types.RecommendationProvided,
					Status:  corev1.ConditionTrue,
					Reason:  "",
					Message: "",
				}, {
					Type:    vpa_types.NoPodsMatched,
					Status:  corev1.ConditionTrue,
					Reason:  "NoPodsMatched",
					Message: "No pods match this VPA object",
				},
			},
		}, {
			name:              "No recommendation but pods matched",
			podsMatched:       true,
			hasRecommendation: false,
			expectedConditions: []vpa_types.VerticalPodAutoscalerCondition{
				{
					Type:    vpa_types.RecommendationProvided,
					Status:  corev1.ConditionFalse,
					Reason:  "",
					Message: "",
				},
			},
			expectedAbsent: []vpa_types.VerticalPodAutoscalerConditionType{vpa_types.NoPodsMatched},
		}, {
			name:              "No recommendation no pods matched",
			podsMatched:       false,
			hasRecommendation: false,
			expectedConditions: []vpa_types.VerticalPodAutoscalerCondition{
				{
					Type:    vpa_types.RecommendationProvided,
					Status:  corev1.ConditionFalse,
					Reason:  "NoPodsMatched",
					Message: "No pods match this VPA object",
				}, {
					Type:    vpa_types.NoPodsMatched,
					Status:  corev1.ConditionTrue,
					Reason:  "NoPodsMatched",
					Message: "No pods match this VPA object",
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
				if condition.Status == corev1.ConditionTrue {
					assert.True(t, vpa.Conditions.ConditionActive(condition.Type))
				} else {
					assert.False(t, vpa.Conditions.ConditionActive(condition.Type))
				}
			}
			for _, condition := range tc.expectedAbsent {
				assert.NotContains(t, vpa.Conditions, condition)
				assert.False(t, vpa.Conditions.ConditionActive(condition))
			}
		})
	}
}

func TestUpdateRecommendation(t *testing.T) {
	type simpleRec struct {
		cpu, mem string
	}
	cases := []struct {
		name           string
		containers     map[string]*simpleRec
		recommendation *vpa_types.RecommendedPodResources
		expectedLast   map[string]corev1.ResourceList
	}{
		{
			name:       "New recommendation",
			containers: map[string]*simpleRec{"test-container": nil, "second-container": nil},
			recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("test-container").WithTarget("5", "200").GetContainerResources(),
					test.Recommendation().WithContainer("second-container").WithTarget("200m", "3000").GetContainerResources(),
				}},
			expectedLast: map[string]corev1.ResourceList{
				"test-container":   test.Resources("5", "200"),
				"second-container": test.Resources("200m", "3000"),
			},
		}, {
			name:       "One recommendation updated",
			containers: map[string]*simpleRec{"test-container": {"5", "200"}, "second-container": {"200m", "3000"}},
			recommendation: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					test.Recommendation().WithContainer("test-container").WithTarget("10", "200").GetContainerResources(),
				}},
			expectedLast: map[string]corev1.ResourceList{
				"test-container":   test.Resources("10", "200"),
				"second-container": test.Resources("200m", "3000"),
			},
		}, {
			name:           "Recommendation for container missing",
			containers:     map[string]*simpleRec{"test-container": nil, "second-container": nil},
			recommendation: test.Recommendation().WithContainer("test-container").WithTarget("5", "200").Get(),
			expectedLast: map[string]corev1.ResourceList{
				"test-container": test.Resources("5", "200"),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			namespace := "test-namespace"
			vpa := NewVpa(VpaID{Namespace: namespace, VpaName: "my-favourite-vpa"}, labels.Nothing(), anyTime)
			for container, rec := range tc.containers {
				state := &AggregateContainerState{}
				if rec != nil {
					state.LastRecommendation = corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(rec.cpu),
						corev1.ResourceMemory: resource.MustParse(rec.mem),
					}
				}

				vpa.aggregateContainerStates[aggregateStateKey{
					namespace:     namespace,
					containerName: container,
				}] = state
			}
			vpa.UpdateRecommendation(tc.recommendation)
			assert.Equal(t, vpa.Recommendation, tc.recommendation)
			for key, state := range vpa.aggregateContainerStates {
				expected, ok := tc.expectedLast[key.ContainerName()]
				if !ok {
					assert.Nil(t, state.LastRecommendation)
				} else {
					assert.Equal(t, expected, state.LastRecommendation)
				}
			}
		})
	}
}

func TestUseAggregationIfMatching(t *testing.T) {
	modeOff := vpa_types.UpdateModeOff
	modeAuto := vpa_types.UpdateModeAuto
	cases := []struct {
		name                        string
		aggregations                []string
		vpaSelector                 string
		updateMode                  *vpa_types.UpdateMode
		container                   string
		containerLabels             map[string]string
		isUnderVPA                  bool
		expectedAggregations        []string
		expectedUpdateMode          *vpa_types.UpdateMode
		expectedNeedsRecommendation bool
	}{
		{
			name:                        "First matching aggregation",
			aggregations:                []string{},
			vpaSelector:                 testSelectorStr,
			updateMode:                  &modeOff,
			container:                   "test-container",
			containerLabels:             testLabels,
			isUnderVPA:                  false,
			expectedAggregations:        []string{"test-container"},
			expectedNeedsRecommendation: true,
			expectedUpdateMode:          &modeOff,
		}, {
			name:                        "New matching aggregation",
			aggregations:                []string{"test-container"},
			vpaSelector:                 testSelectorStr,
			updateMode:                  &modeAuto,
			container:                   "second-container",
			containerLabels:             testLabels,
			isUnderVPA:                  false,
			expectedAggregations:        []string{"test-container", "second-container"},
			expectedNeedsRecommendation: true,
			expectedUpdateMode:          &modeAuto,
		}, {
			name:                        "Existing matching aggregation",
			aggregations:                []string{"test-container"},
			vpaSelector:                 testSelectorStr,
			updateMode:                  &modeOff,
			container:                   "test-container",
			containerLabels:             testLabels,
			isUnderVPA:                  true,
			expectedAggregations:        []string{"test-container"},
			expectedNeedsRecommendation: true,
			expectedUpdateMode:          &modeOff,
		}, {
			name:                        "Aggregation not matching",
			aggregations:                []string{"test-container"},
			vpaSelector:                 testSelectorStr,
			updateMode:                  &modeAuto,
			container:                   "second-container",
			containerLabels:             map[string]string{"different": "labels"},
			isUnderVPA:                  false,
			expectedAggregations:        []string{"test-container"},
			expectedNeedsRecommendation: false,
			expectedUpdateMode:          nil,
		},
	}
	for _, tc := range cases {
		if tc.name == "Existing matching aggregation" {
			t.Run(tc.name, func(t *testing.T) {
				namespace := "test-namespace"
				selector, err := labels.Parse(tc.vpaSelector)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				vpa := NewVpa(VpaID{Namespace: namespace, VpaName: "my-favourite-vpa"}, selector, anyTime)
				vpa.UpdateMode = tc.updateMode
				key := mockAggregateStateKey{
					namespace:     namespace,
					containerName: tc.container,
					labels:        labels.Set(tc.containerLabels).String(),
				}
				aggregation := &AggregateContainerState{
					IsUnderVPA: tc.isUnderVPA,
				}
				for _, container := range tc.aggregations {
					containerKey := mockAggregateStateKey{
						namespace:     namespace,
						containerName: container,
						labels:        labels.Set(testLabels).String(),
					}
					if container == tc.container {
						aggregation.UpdateMode = vpa.UpdateMode
						vpa.aggregateContainerStates[containerKey] = aggregation
					} else {
						vpa.aggregateContainerStates[key] = &AggregateContainerState{UpdateMode: vpa.UpdateMode}
					}
				}

				vpa.UseAggregationIfMatching(key, aggregation)
				assert.Equal(t, len(tc.expectedAggregations), len(vpa.aggregateContainerStates), "AggregateContainerStates has unexpected size")
				for _, container := range tc.expectedAggregations {
					found := false
					for key := range vpa.aggregateContainerStates {
						if key.ContainerName() == container {
							found = true
							break
						}
					}
					assert.True(t, found, "Container %s not found in aggregateContainerStates", container)
				}
				assert.Equal(t, tc.expectedNeedsRecommendation, aggregation.NeedsRecommendation())
				if tc.expectedUpdateMode == nil {
					assert.Nil(t, aggregation.GetUpdateMode())
				} else {
					if assert.NotNil(t, aggregation.GetUpdateMode()) {
						assert.Equal(t, *tc.expectedUpdateMode, *aggregation.GetUpdateMode())
					}
				}
			})
		}
	}
}

type mockAggregateStateKey struct {
	namespace     string
	containerName string
	labels        string
}

func (k mockAggregateStateKey) Namespace() string {
	return k.namespace
}

func (k mockAggregateStateKey) ContainerName() string {
	return k.containerName
}

func (k mockAggregateStateKey) Labels() labels.Labels {
	// Should return empty on error
	labels, _ := labels.ConvertSelectorToLabelsMap(k.labels)
	return labels
}
