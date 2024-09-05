/*
Copyright 2019 The Kubernetes Authors.

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

package input

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/spec"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	target_mock "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

type fakeControllerFetcher struct {
	key *controllerfetcher.ControllerKeyWithAPIVersion
	err error
}

func (f *fakeControllerFetcher) FindTopMostWellKnownOrScalable(_ context.Context, _ *controllerfetcher.ControllerKeyWithAPIVersion) (*controllerfetcher.ControllerKeyWithAPIVersion, error) {
	return f.key, f.err
}

func parseLabelSelector(selector string) labels.Selector {
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

var (
	recommenderName                     = "name"
	empty                               = ""
	unsupportedConditionTextFromFetcher = "Cannot read targetRef. Reason: targetRef not defined"
	unsupportedConditionNoExtraText     = "Cannot read targetRef"
	unsupportedConditionNoTargetRef     = "Cannot read targetRef"
	unsupportedConditionMudaMudaMuda    = "Error checking if target is a topmost well-known or scalable controller: muda muda muda"
	unsupportedTargetRefHasParent       = "The targetRef controller has a parent but it should point to a topmost well-known or scalable controller"
)

const (
	kind         = "dodokind"
	name1        = "dotaro"
	name2        = "doseph"
	namespace    = "testNamespace"
	apiVersion   = "stardust"
	testGcPeriod = time.Minute
)

func TestLoadPods(t *testing.T) {

	type testCase struct {
		name                                string
		selector                            labels.Selector
		fetchSelectorError                  error
		targetRef                           *autoscalingv1.CrossVersionObjectReference
		topMostWellKnownOrScalableKey       *controllerfetcher.ControllerKeyWithAPIVersion
		findTopMostWellKnownOrScalableError error
		expectedSelector                    labels.Selector
		expectedConfigUnsupported           *string
		expectedConfigDeprecated            *string
		expectedVpaFetch                    bool
		recommenderName                     *string
		recommender                         string
	}

	testCases := []testCase{
		{
			name:                      "no selector",
			selector:                  nil,
			fetchSelectorError:        fmt.Errorf("targetRef not defined"),
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionTextFromFetcher,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          true,
		},
		{
			name:                      "also no selector but no error",
			selector:                  nil,
			fetchSelectorError:        nil,
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionNoExtraText,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          true,
		},
		{
			name:               "targetRef selector",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedSelector:          parseLabelSelector("app = test"),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          true,
		},
		{
			name:                      "no targetRef",
			selector:                  parseLabelSelector("app = test"),
			fetchSelectorError:        nil,
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          true,
		},
		{
			name:               "can't decide if top-level-ref",
			selector:           nil,
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			expectedConfigUnsupported: &unsupportedConditionNoTargetRef,
			expectedVpaFetch:          true,
		},
		{
			name:               "non-top-level targetRef",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name2,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedConfigUnsupported: &unsupportedTargetRefHasParent,
			expectedVpaFetch:          true,
		},
		{
			name:               "error checking if top-level-ref",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       "doestar",
				Name:       "doseph-doestar",
				APIVersion: "taxonomy",
			},
			expectedConfigUnsupported:           &unsupportedConditionMudaMudaMuda,
			expectedVpaFetch:                    true,
			findTopMostWellKnownOrScalableError: fmt.Errorf("muda muda muda"),
		},
		{
			name:               "top-level target ref",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   parseLabelSelector("app = test"),
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedConfigUnsupported: nil,
			expectedVpaFetch:          true,
		},
		{
			name:               "no recommenderName",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedSelector:          parseLabelSelector("app = test"),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          false,
			recommenderName:           &empty,
		},
		{
			name:               "recommenderName doesn't match recommender",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedSelector:          parseLabelSelector("app = test"),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          false,
			recommenderName:           &recommenderName,
			recommender:               "other",
		},
		{
			name:               "recommenderName matches recommender",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			targetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topMostWellKnownOrScalableKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedSelector:          parseLabelSelector("app = test"),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedVpaFetch:          true,
			recommenderName:           &recommenderName,
			recommender:               recommenderName,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			vpaBuilder := test.VerticalPodAutoscaler().WithName("testVpa").WithContainer("container").WithNamespace("testNamespace").WithTargetRef(tc.targetRef)
			if tc.recommender != "" {
				vpaBuilder = vpaBuilder.WithRecommender(tc.recommender)
			}
			vpa := vpaBuilder.Get()
			vpaLister := &test.VerticalPodAutoscalerListerMock{}
			vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpa}, nil)

			targetSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)
			clusterState := model.NewClusterState(testGcPeriod)

			clusterStateFeeder := clusterStateFeeder{
				vpaLister:       vpaLister,
				clusterState:    clusterState,
				selectorFetcher: targetSelectorFetcher,
				controllerFetcher: &fakeControllerFetcher{
					key: tc.topMostWellKnownOrScalableKey,
					err: tc.findTopMostWellKnownOrScalableError,
				},
			}
			if tc.recommenderName == nil {
				clusterStateFeeder.recommenderName = DefaultRecommenderName
			} else {
				clusterStateFeeder.recommenderName = *tc.recommenderName
			}

			if tc.expectedVpaFetch {
				targetSelectorFetcher.EXPECT().Fetch(vpa).Return(tc.selector, tc.fetchSelectorError)
			}
			clusterStateFeeder.LoadVPAs(context.Background())

			vpaID := model.VpaID{
				Namespace: vpa.Namespace,
				VpaName:   vpa.Name,
			}

			if !tc.expectedVpaFetch {
				assert.NotContains(t, clusterState.Vpas, vpaID)
				return
			}
			assert.Contains(t, clusterState.Vpas, vpaID)
			storedVpa := clusterState.Vpas[vpaID]
			if tc.expectedSelector != nil {
				assert.NotNil(t, storedVpa.PodSelector)
				assert.Equal(t, tc.expectedSelector.String(), storedVpa.PodSelector.String())
			} else {
				assert.Nil(t, storedVpa.PodSelector)
			}

			if tc.expectedConfigDeprecated != nil {
				assert.Contains(t, storedVpa.Conditions, vpa_types.ConfigDeprecated)
				assert.Equal(t, *tc.expectedConfigDeprecated, storedVpa.Conditions[vpa_types.ConfigDeprecated].Message)
			} else {
				assert.NotContains(t, storedVpa.Conditions, vpa_types.ConfigDeprecated)
			}

			if tc.expectedConfigUnsupported != nil {
				assert.Contains(t, storedVpa.Conditions, vpa_types.ConfigUnsupported)
				assert.Equal(t, *tc.expectedConfigUnsupported, storedVpa.Conditions[vpa_types.ConfigUnsupported].Message)
			} else {
				assert.NotContains(t, storedVpa.Conditions, vpa_types.ConfigUnsupported)
			}

		})
	}
}

type testSpecClient struct {
	pods []*spec.BasicPodSpec
}

func (c *testSpecClient) GetPodSpecs() ([]*spec.BasicPodSpec, error) {
	return c.pods, nil
}

func makeTestSpecClient(podLabels []map[string]string) spec.SpecClient {
	pods := make([]*spec.BasicPodSpec, len(podLabels))
	for i, l := range podLabels {
		pods[i] = &spec.BasicPodSpec{
			ID:        model.PodID{Namespace: "default", PodName: fmt.Sprintf("pod-%d", i)},
			PodLabels: l,
		}
	}
	return &testSpecClient{
		pods: pods,
	}
}

func TestClusterStateFeeder_LoadPods(t *testing.T) {
	for _, tc := range []struct {
		Name              string
		VPALabelSelectors []string
		PodLabels         []map[string]string
		TrackedPods       int
	}{
		{
			Name:              "simple",
			VPALabelSelectors: []string{"name=vpa-pod"},
			PodLabels: []map[string]string{
				{"name": "vpa-pod"},
				{"type": "stateful"},
			},
			TrackedPods: 1,
		},
		{
			Name:              "multiple",
			VPALabelSelectors: []string{"name=vpa-pod,type=stateful"},
			PodLabels: []map[string]string{
				{"name": "vpa-pod", "type": "stateful"},
				{"type": "stateful"},
				{"name": "vpa-pod"},
			},
			TrackedPods: 1,
		},
		{
			Name:              "no matches",
			VPALabelSelectors: []string{"name=vpa-pod"},
			PodLabels: []map[string]string{
				{"name": "non-vpa-pod", "type": "stateful"},
			},
			TrackedPods: 0,
		},
		{
			Name:              "set based",
			VPALabelSelectors: []string{"environment in (staging, qa),name=vpa-pod"},
			PodLabels: []map[string]string{
				{"name": "vpa-pod", "environment": "staging"},
				{"name": "vpa-pod", "environment": "production"},
				{"name": "non-vpa-pod", "environment": "staging"},
				{"name": "non-vpa-pod", "environment": "production"},
			},
			TrackedPods: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			clusterState := model.NewClusterState(testGcPeriod)
			for i, selector := range tc.VPALabelSelectors {
				vpaLabel, err := labels.Parse(selector)
				assert.NoError(t, err)
				clusterState.Vpas = map[model.VpaID]*model.Vpa{
					{VpaName: fmt.Sprintf("test-vpa-%d", i), Namespace: "default"}: {PodSelector: vpaLabel},
				}
			}

			feeder := clusterStateFeeder{
				specClient:     makeTestSpecClient(tc.PodLabels),
				memorySaveMode: true,
				clusterState:   clusterState,
			}

			feeder.LoadPods()
			assert.Len(t, feeder.clusterState.Pods, tc.TrackedPods, "number of pods is not %d", tc.TrackedPods)

			feeder = clusterStateFeeder{
				specClient:     makeTestSpecClient(tc.PodLabels),
				memorySaveMode: false,
				clusterState:   clusterState,
			}

			feeder.LoadPods()
			assert.Len(t, feeder.clusterState.Pods, len(tc.PodLabels), "number of pods is not %d", len(tc.PodLabels))
		})
	}
}

type fakeHistoryProvider struct {
	history map[model.PodID]*history.PodHistory
	err     error
}

func (fhp *fakeHistoryProvider) GetClusterHistory() (map[model.PodID]*history.PodHistory, error) {
	return fhp.history, fhp.err
}

func TestClusterStateFeeder_InitFromHistoryProvider(t *testing.T) {
	pod1 := model.PodID{
		Namespace: "ns",
		PodName:   "a-pod",
	}
	memAmount := model.ResourceAmount(128 * 1024 * 1024)
	t0 := time.Date(2021, time.August, 30, 10, 21, 0, 0, time.UTC)
	containerCpu := "containerCpu"
	containerMem := "containerMem"
	pod1History := history.PodHistory{
		LastLabels: map[string]string{},
		LastSeen:   t0,
		Samples: map[string][]model.ContainerUsageSample{
			containerCpu: {
				{
					MeasureStart: t0,
					Usage:        10,
					Request:      101,
					Resource:     model.ResourceCPU,
				},
			},
			containerMem: {
				{
					MeasureStart: t0,
					Usage:        memAmount,
					Request:      1024 * 1024 * 1024,
					Resource:     model.ResourceMemory,
				},
			},
		},
	}
	provider := fakeHistoryProvider{
		history: map[model.PodID]*history.PodHistory{
			pod1: &pod1History,
		},
	}

	clusterState := model.NewClusterState(testGcPeriod)
	feeder := clusterStateFeeder{
		clusterState: clusterState,
	}
	feeder.InitFromHistoryProvider(&provider)
	if !assert.Contains(t, feeder.clusterState.Pods, pod1) {
		return
	}
	pod1State := feeder.clusterState.Pods[pod1]
	if !assert.Contains(t, pod1State.Containers, containerCpu) {
		return
	}
	containerState := pod1State.Containers[containerCpu]
	if !assert.NotNil(t, containerState) {
		return
	}
	assert.Equal(t, t0, containerState.LastCPUSampleStart)
	if !assert.Contains(t, pod1State.Containers, containerMem) {
		return
	}
	containerState = pod1State.Containers[containerMem]
	if !assert.NotNil(t, containerState) {
		return
	}
	assert.Equal(t, memAmount, containerState.GetMaxMemoryPeak())
}

func TestFilterVPAs(t *testing.T) {
	recommenderName := "test-recommender"
	defaultRecommenderName := "default-recommender"

	vpa1 := &vpa_types.VerticalPodAutoscaler{
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: defaultRecommenderName},
			},
		},
	}
	vpa2 := &vpa_types.VerticalPodAutoscaler{
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: recommenderName},
			},
		},
	}
	vpa3 := &vpa_types.VerticalPodAutoscaler{
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: "another-recommender"},
			},
		},
	}

	allVpaCRDs := []*vpa_types.VerticalPodAutoscaler{vpa1, vpa2, vpa3}

	feeder := &clusterStateFeeder{
		recommenderName: recommenderName,
	}

	// Set expected results
	expectedResult := []*vpa_types.VerticalPodAutoscaler{vpa2}

	// Run the filterVPAs function
	result := filterVPAs(feeder, allVpaCRDs)

	if len(result) != len(expectedResult) {
		t.Fatalf("expected %d VPAs, got %d", len(expectedResult), len(result))
	}

	assert.ElementsMatch(t, expectedResult, result)
}

func TestFilterVPAsIgnoreNamespaces(t *testing.T) {

	vpa1 := &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace1",
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: DefaultRecommenderName},
			},
		},
	}
	vpa2 := &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "namespace2",
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: DefaultRecommenderName},
			},
		},
	}
	vpa3 := &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ignore1",
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: DefaultRecommenderName},
			},
		},
	}
	vpa4 := &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ignore2",
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: DefaultRecommenderName},
			},
		},
	}

	allVpaCRDs := []*vpa_types.VerticalPodAutoscaler{vpa1, vpa2, vpa3, vpa4}

	feeder := &clusterStateFeeder{
		recommenderName:   DefaultRecommenderName,
		ignoredNamespaces: []string{"ignore1", "ignore2"},
	}

	// Set expected results
	expectedResult := []*vpa_types.VerticalPodAutoscaler{vpa1, vpa2}

	// Run the filterVPAs function
	result := filterVPAs(feeder, allVpaCRDs)

	if len(result) != len(expectedResult) {
		t.Fatalf("expected %d VPAs, got %d", len(expectedResult), len(result))
	}

	assert.ElementsMatch(t, expectedResult, result)
}
