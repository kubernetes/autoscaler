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
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/model"
	target_mock "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/test"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/spec"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
)

type fakeControllerFetcher struct {
	key *controllerfetcher.ControllerKeyWithAPIVersion
	err error
	mapper restmapper.DeferredDiscoveryRESTMapper
	scaleNamespacer scale.ScalesGetter
}

func (f *fakeControllerFetcher) FindTopMostWellKnownOrScalable(_ *controllerfetcher.ControllerKeyWithAPIVersion) (*controllerfetcher.ControllerKeyWithAPIVersion, error) {
	return f.key, f.err
}

func (f *fakeControllerFetcher) GetRESTMappings(groupKind schema.GroupKind) ([]*apimeta.RESTMapping, error) {
	return f.mapper.RESTMappings(groupKind)
}

func (f *fakeControllerFetcher) Scales(namespace string) (scale.ScaleInterface) {
	return f.scaleNamespacer.Scales(namespace)
}

func parseLabelSelector(selector string) labels.Selector {
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

var (
	recommenderName                     = "name"
	empty                               = ""
	unsupportedConditionTextFromFetcher = "Cannot read scaleTargetRef. Reason: scaleTargetRef not defined"
	unsupportedConditionNoExtraText     = "Cannot read scaleTargetRef"
	unsupportedConditionNoTargetRef     = "Cannot read scaleTargetRef"
	unsupportedConditionMudaMudaMuda    = "Error checking if target is a topmost well-known or scalable controller: muda muda muda"
	unsupportedTargetRefHasParent       = "The scaleTargetRef controller has a parent but it should point to a topmost well-known or scalable controller"
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
		scaleTargetRef                      *autoscalingv1.CrossVersionObjectReference
		topMostWellKnownOrScalableKey       *controllerfetcher.ControllerKeyWithAPIVersion
		findTopMostWellKnownOrScalableError error
		expectedSelector                    labels.Selector
		expectedConfigUnsupported           *string
		expectedConfigDeprecated            *string
		expectedMpaFetch                    bool
		recommenderName                     *string
		recommender                         string
	}

	testCases := []testCase{
		{
			name:                      "no selector",
			selector:                  nil,
			fetchSelectorError:        fmt.Errorf("scaleTargetRef not defined"),
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionTextFromFetcher,
			expectedConfigDeprecated:  nil,
			expectedMpaFetch:          true,
		},
		{
			name:                      "also no selector but no error",
			selector:                  nil,
			fetchSelectorError:        nil,
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionNoExtraText,
			expectedConfigDeprecated:  nil,
			expectedMpaFetch:          true,
		},
		{
			name:               "scaleTargetRef selector",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			scaleTargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			expectedMpaFetch:          true,
		},
		{
			name:                      "no scaleTargetRef",
			selector:                  parseLabelSelector("app = test"),
			fetchSelectorError:        nil,
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
			expectedMpaFetch:          true,
		},
		{
			name:               "can't decide if top-level-ref",
			selector:           nil,
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			scaleTargetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			expectedConfigUnsupported: &unsupportedConditionNoTargetRef,
			expectedMpaFetch:          true,
		},
		{
			name:               "non-top-level scaleTargetRef",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			scaleTargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			expectedMpaFetch:          true,
		},
		{
			name:               "error checking if top-level-ref",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			scaleTargetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind:       "doestar",
				Name:       "doseph-doestar",
				APIVersion: "taxonomy",
			},
			expectedConfigUnsupported:           &unsupportedConditionMudaMudaMuda,
			expectedMpaFetch:                    true,
			findTopMostWellKnownOrScalableError: fmt.Errorf("muda muda muda"),
		},
		{
			name:               "top-level target ref",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   parseLabelSelector("app = test"),
			scaleTargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			expectedMpaFetch:          true,
		},
		{
			name:               "no recommenderName",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			scaleTargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			expectedMpaFetch:          false,
			recommenderName:           &empty,
		},
		{
			name:               "recommenderName doesn't match recommender",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			scaleTargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			expectedMpaFetch:          false,
			recommenderName:           &recommenderName,
			recommender:               "other",
		},
		{
			name:               "recommenderName matches recommender",
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			scaleTargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			expectedMpaFetch:          true,
			recommenderName:           &recommenderName,
			recommender:               recommenderName,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mpaBuilder := test.MultidimPodAutoscaler().WithName("testMpa").WithContainer("container").WithNamespace("testNamespace").WithScaleTargetRef(tc.scaleTargetRef)
			if tc.recommender != "" {
				mpaBuilder = mpaBuilder.WithRecommender(tc.recommender)
			}
			mpa := mpaBuilder.Get()
			mpaLister := &test.MultidimPodAutoscalerListerMock{}
			mpaLister.On("List").Return([]*mpa_types.MultidimPodAutoscaler{mpa}, nil)

			targetSelectorFetcher := target_mock.NewMockMpaTargetSelectorFetcher(ctrl)
			clusterState := model.NewClusterState(testGcPeriod)

			clusterStateFeeder := clusterStateFeeder{
				mpaLister:       mpaLister,
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

			if tc.expectedMpaFetch {
				targetSelectorFetcher.EXPECT().Fetch(mpa).Return(tc.selector, tc.fetchSelectorError)
			}
			clusterStateFeeder.LoadMPAs()

			mpaID := model.MpaID{
				Namespace: mpa.Namespace,
				MpaName:   mpa.Name,
			}

			if !tc.expectedMpaFetch {
				assert.NotContains(t, clusterState.Mpas, mpaID)
				return
			}
			assert.Contains(t, clusterState.Mpas, mpaID)
			storedMpa := clusterState.Mpas[mpaID]
			if tc.expectedSelector != nil {
				assert.NotNil(t, storedMpa.PodSelector)
				assert.Equal(t, tc.expectedSelector.String(), storedMpa.PodSelector.String())
			} else {
				assert.Nil(t, storedMpa.PodSelector)
			}

			if tc.expectedConfigDeprecated != nil {
				assert.Contains(t, storedMpa.Conditions, mpa_types.ConfigDeprecated)
				assert.Equal(t, *tc.expectedConfigDeprecated, storedMpa.Conditions[mpa_types.ConfigDeprecated].Message)
			} else {
				assert.NotContains(t, storedMpa.Conditions, mpa_types.ConfigDeprecated)
			}

			if tc.expectedConfigUnsupported != nil {
				assert.Contains(t, storedMpa.Conditions, mpa_types.ConfigUnsupported)
				assert.Equal(t, *tc.expectedConfigUnsupported, storedMpa.Conditions[mpa_types.ConfigUnsupported].Message)
			} else {
				assert.NotContains(t, storedMpa.Conditions, mpa_types.ConfigUnsupported)
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
			ID:        vpa_model.PodID{Namespace: "default", PodName: fmt.Sprintf("pod-%d", i)},
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
		MPALabelSelectors []string
		PodLabels         []map[string]string
		TrackedPods       int
	}{
		{
			Name:              "simple",
			MPALabelSelectors: []string{"name=mpa-pod"},
			PodLabels: []map[string]string{
				{"name": "mpa-pod"},
				{"type": "stateful"},
			},
			TrackedPods: 1,
		},
		{
			Name:              "multiple",
			MPALabelSelectors: []string{"name=mpa-pod,type=stateful"},
			PodLabels: []map[string]string{
				{"name": "mpa-pod", "type": "stateful"},
				{"type": "stateful"},
				{"name": "mpa-pod"},
			},
			TrackedPods: 1,
		},
		{
			Name:              "no matches",
			MPALabelSelectors: []string{"name=mpa-pod"},
			PodLabels: []map[string]string{
				{"name": "non-mpa-pod", "type": "stateful"},
			},
			TrackedPods: 0,
		},
		{
			Name:              "set based",
			MPALabelSelectors: []string{"environment in (staging, qa),name=mpa-pod"},
			PodLabels: []map[string]string{
				{"name": "mpa-pod", "environment": "staging"},
				{"name": "mpa-pod", "environment": "production"},
				{"name": "non-mpa-pod", "environment": "staging"},
				{"name": "non-mpa-pod", "environment": "production"},
			},
			TrackedPods: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			clusterState := model.NewClusterState(testGcPeriod)
			for i, selector := range tc.MPALabelSelectors {
				mpaLabel, err := labels.Parse(selector)
				assert.NoError(t, err)
				clusterState.Mpas = map[model.MpaID]*model.Mpa{
					{MpaName: fmt.Sprintf("test-mpa-%d", i), Namespace: "default"}: {PodSelector: mpaLabel},
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
	history map[vpa_model.PodID]*history.PodHistory
	err     error
}

func (fhp *fakeHistoryProvider) GetClusterHistory() (map[vpa_model.PodID]*history.PodHistory, error) {
	return fhp.history, fhp.err
}

func TestClusterStateFeeder_InitFromHistoryProvider(t *testing.T) {
	pod1 := vpa_model.PodID{
		Namespace: "ns",
		PodName:   "a-pod",
	}
	memAmount := vpa_model.ResourceAmount(128 * 1024 * 1024)
	t0 := time.Date(2021, time.August, 30, 10, 21, 0, 0, time.UTC)
	containerCpu := "containerCpu"
	containerMem := "containerMem"
	pod1History := history.PodHistory{
		LastLabels: map[string]string{},
		LastSeen:   t0,
		Samples: map[string][]vpa_model.ContainerUsageSample{
			containerCpu: {
				{
					MeasureStart: t0,
					Usage:        10,
					Request:      101,
					Resource:     vpa_model.ResourceCPU,
				},
			},
			containerMem: {
				{
					MeasureStart: t0,
					Usage:        memAmount,
					Request:      1024 * 1024 * 1024,
					Resource:     vpa_model.ResourceMemory,
				},
			},
		},
	}
	provider := fakeHistoryProvider{
		history: map[vpa_model.PodID]*history.PodHistory{
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
