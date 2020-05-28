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
	scalingModeAuto := vpa_types.ContainerScalingModeAuto
	scalingModeOff := vpa_types.ContainerScalingModeOff
	cases := []struct {
		name                        string
		aggregations                []string
		vpaSelector                 string
		resourcePolicy              *vpa_types.PodResourcePolicy
		updateMode                  *vpa_types.UpdateMode
		container                   string
		containerLabels             map[string]string
		expectedUpdateMode          *vpa_types.UpdateMode
		expectedNeedsRecommendation map[string]bool
	}{
		{
			name:                        "First matching aggregation",
			aggregations:                []string{},
			vpaSelector:                 testSelectorStr,
			updateMode:                  &modeOff,
			container:                   "test-container",
			containerLabels:             testLabels,
			expectedNeedsRecommendation: map[string]bool{"test-container": true},
			expectedUpdateMode:          &modeOff,
		}, {
			name:                        "New matching aggregation",
			aggregations:                []string{"test-container"},
			vpaSelector:                 testSelectorStr,
			updateMode:                  &modeAuto,
			container:                   "second-container",
			containerLabels:             testLabels,
			expectedNeedsRecommendation: map[string]bool{"test-container": true, "second-container": true},
			expectedUpdateMode:          &modeAuto,
		}, {
			name:                        "Existing matching aggregation",
			aggregations:                []string{"test-container"},
			vpaSelector:                 testSelectorStr,
			updateMode:                  &modeOff,
			container:                   "test-container",
			containerLabels:             testLabels,
			expectedNeedsRecommendation: map[string]bool{"test-container": true},
			expectedUpdateMode:          &modeOff,
		}, {
			name:                        "Aggregation not matching",
			aggregations:                []string{"test-container"},
			vpaSelector:                 testSelectorStr,
			updateMode:                  &modeAuto,
			container:                   "second-container",
			containerLabels:             map[string]string{"different": "labels"},
			expectedNeedsRecommendation: map[string]bool{"test-container": true},
			expectedUpdateMode:          nil,
		}, {
			name:         "New matching aggregation with scaling mode Off",
			aggregations: []string{"test-container"},
			vpaSelector:  testSelectorStr,
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "second-container",
						Mode:          &scalingModeOff,
					},
				},
			},
			updateMode:                  &modeAuto,
			container:                   "second-container",
			containerLabels:             testLabels,
			expectedNeedsRecommendation: map[string]bool{"second-container": false, "test-container": true},
			expectedUpdateMode:          &modeAuto,
		}, {
			name:         "New matching aggregation with default scaling mode Off",
			aggregations: []string{"test-container"},
			vpaSelector:  testSelectorStr,
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "*",
						Mode:          &scalingModeOff,
					},
				},
			},
			updateMode:                  &modeAuto,
			container:                   "second-container",
			containerLabels:             testLabels,
			expectedNeedsRecommendation: map[string]bool{"second-container": false, "test-container": false},
			expectedUpdateMode:          &modeAuto,
		}, {
			name:         "New matching aggregation with scaling mode Off with default Auto",
			aggregations: []string{"test-container"},
			vpaSelector:  testSelectorStr,
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "*",
						Mode:          &scalingModeAuto,
					},
					{
						ContainerName: "second-container",
						Mode:          &scalingModeOff,
					},
				},
			},
			updateMode:                  &modeAuto,
			container:                   "second-container",
			containerLabels:             testLabels,
			expectedNeedsRecommendation: map[string]bool{"second-container": false, "test-container": true},
			expectedUpdateMode:          &modeAuto,
		}, {
			name:         "New matching aggregation with scaling mode Auto with default Off",
			aggregations: []string{"test-container"},
			vpaSelector:  testSelectorStr,
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "*",
						Mode:          &scalingModeOff,
					},
					{
						ContainerName: "second-container",
						Mode:          &scalingModeAuto,
					},
				},
			},
			updateMode:                  &modeAuto,
			container:                   "second-container",
			containerLabels:             testLabels,
			expectedNeedsRecommendation: map[string]bool{"second-container": true, "test-container": false},
			expectedUpdateMode:          &modeAuto,
		},
	}
	for _, tc := range cases {
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
			aggregationUnderTest := &AggregateContainerState{}
			for _, container := range tc.aggregations {
				containerKey, aggregation := testAggregation(vpa, container, labels.Set(testLabels).String())
				vpa.aggregateContainerStates[containerKey] = aggregation
				if container == tc.container {
					aggregationUnderTest = aggregation
				}
			}
			vpa.SetResourcePolicy(tc.resourcePolicy)

			vpa.UseAggregationIfMatching(key, aggregationUnderTest)
			assert.Len(t, vpa.aggregateContainerStates, len(tc.expectedNeedsRecommendation), "AggregateContainerStates has unexpected size")
			for container, expectedNeedsRecommendation := range tc.expectedNeedsRecommendation {
				found := false
				for key, state := range vpa.aggregateContainerStates {
					if key.ContainerName() == container {
						found = true
						assert.Equal(t, expectedNeedsRecommendation, state.NeedsRecommendation(),
							"Unexpected NeedsRecommendation result for container %s", container)
					}
				}
				assert.True(t, found, "Container %s not found in aggregateContainerStates", container)
			}
			assert.Equal(t, tc.expectedUpdateMode, aggregationUnderTest.GetUpdateMode())
		})
	}
}

func TestDeleteAggregation(t *testing.T) {
	cases := []struct {
		name                     string
		aggregateContainerStates aggregateContainerStatesMap
		delet                    AggregateStateKey
	}{
		{
			name: "delet dis",
			aggregateContainerStates: aggregateContainerStatesMap{
				aggregateStateKey{
					namespace:     "ns",
					containerName: "container",
					labelSetKey:   "labelSetKey",
					labelSetMap:   nil,
				}: &AggregateContainerState{},
			},
			delet: aggregateStateKey{
				namespace:     "ns",
				containerName: "container",
				labelSetKey:   "labelSetKey",
				labelSetMap:   nil,
			},
		},
		{
			name: "no delet",
			delet: aggregateStateKey{
				namespace:     "ns",
				containerName: "container",
				labelSetKey:   "labelSetKey",
				labelSetMap:   nil,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vpa := Vpa{
				aggregateContainerStates: tc.aggregateContainerStates,
			}
			vpa.DeleteAggregation(tc.delet)
			assert.Equal(t, 0, len(vpa.aggregateContainerStates))
		})
	}
}

func TestSetResourcePolicy(t *testing.T) {
	scalingModeAuto := vpa_types.ContainerScalingModeAuto
	scalingModeOff := vpa_types.ContainerScalingModeOff
	cases := []struct {
		name                        string
		containers                  []string
		resourcePolicy              *vpa_types.PodResourcePolicy
		expectedScalingMode         map[string]*vpa_types.ContainerScalingMode
		expectedNeedsRecommendation map[string]bool
	}{
		{
			name:           "Nil policy",
			containers:     []string{"container1"},
			resourcePolicy: nil,
			expectedScalingMode: map[string]*vpa_types.ContainerScalingMode{
				"container1": &scalingModeAuto,
			},
			expectedNeedsRecommendation: map[string]bool{"container1": true},
		}, {
			name:       "Default policy with no scaling mode",
			containers: []string{"container1", "container2"},
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "*",
					},
				},
			},
			expectedScalingMode: map[string]*vpa_types.ContainerScalingMode{
				"container1": &scalingModeAuto, "container2": &scalingModeAuto,
			},
			expectedNeedsRecommendation: map[string]bool{
				"container1": true, "container2": true},
		}, {
			name:       "Default policy with scaling mode auto",
			containers: []string{"container1", "container2"},
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "*",
						Mode:          &scalingModeAuto,
					},
				},
			},
			expectedScalingMode: map[string]*vpa_types.ContainerScalingMode{
				"container1": &scalingModeAuto, "container2": &scalingModeAuto,
			},
			expectedNeedsRecommendation: map[string]bool{
				"container1": true, "container2": true},
		}, {
			name:       "Default policy with scaling mode off",
			containers: []string{"container1", "container2"},
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "*",
						Mode:          &scalingModeOff,
					},
				},
			},
			expectedScalingMode: map[string]*vpa_types.ContainerScalingMode{
				"container1": &scalingModeOff, "container2": &scalingModeOff,
			},
			expectedNeedsRecommendation: map[string]bool{
				"container1": false, "container2": false},
		}, {
			name:       "One container has scaling mode Off",
			containers: []string{"container1", "container2"},
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "container2",
						Mode:          &scalingModeOff,
					},
				},
			},
			expectedScalingMode: map[string]*vpa_types.ContainerScalingMode{
				"container1": &scalingModeAuto, "container2": &scalingModeOff,
			},
			expectedNeedsRecommendation: map[string]bool{
				"container1": true, "container2": false},
		}, {
			name:       "One container has scaling mode Auto with default off",
			containers: []string{"container1", "container2"},
			resourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "*",
						Mode:          &scalingModeOff,
					}, {
						ContainerName: "container2",
						Mode:          &scalingModeAuto,
					},
				},
			},
			expectedScalingMode: map[string]*vpa_types.ContainerScalingMode{
				"container1": &scalingModeOff, "container2": &scalingModeAuto,
			},
			expectedNeedsRecommendation: map[string]bool{
				"container1": false, "container2": true},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			selector, err := labels.Parse(testSelectorStr)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			vpa := NewVpa(VpaID{Namespace: "test-namespace", VpaName: "my-favourite-vpa"}, selector, anyTime)
			for _, container := range tc.containers {
				containerKey, aggregation := testAggregation(vpa, container, labels.Set(testLabels).String())
				vpa.aggregateContainerStates[containerKey] = aggregation
			}
			vpa.SetResourcePolicy(tc.resourcePolicy)
			assert.Equal(t, tc.resourcePolicy, vpa.ResourcePolicy)
			for key, state := range vpa.aggregateContainerStates {
				containerName := key.ContainerName()
				assert.Equal(t, tc.expectedScalingMode[containerName], state.GetScalingMode(), "Unexpected scaling mode for container %s", containerName)
				assert.Equal(t, tc.expectedNeedsRecommendation[containerName], state.NeedsRecommendation(), "Unexpected needs recommendation for container %s", containerName)
			}
		})
	}
}

func testAggregation(vpa *Vpa, containerName, labels string) (mockAggregateStateKey, *AggregateContainerState) {
	scalingModeAuto := vpa_types.ContainerScalingModeAuto
	containerKey := mockAggregateStateKey{
		namespace:     vpa.ID.Namespace,
		containerName: containerName,
		labels:        labels,
	}
	aggregation := &AggregateContainerState{}
	aggregation.UpdateMode = vpa.UpdateMode
	aggregation.IsUnderVPA = true
	aggregation.ScalingMode = &scalingModeAuto
	return containerKey, aggregation
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
