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
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/controller_fetcher"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/spec"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	target_mock "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

type fakeControllerFetcher struct {
	key *controllerfetcher.ControllerKeyWithAPIVersion
	err error
}

func (f *fakeControllerFetcher) FindTopLevel(controller *controllerfetcher.ControllerKeyWithAPIVersion) (*controllerfetcher.ControllerKeyWithAPIVersion, error) {
	return f.key, f.err
}

func parseLabelSelector(selector string) labels.Selector {
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

var (
	unsupportedConditionNoLongerSupported = "Label selector is no longer supported, please migrate to targetRef"
	unsupportedConditionTextFromFetcher   = "Cannot read targetRef. Reason: targetRef not defined"
	unsupportedConditionNoExtraText       = "Cannot read targetRef"
	unsupportedConditionBothDefined       = "Both targetRef and label selector defined. Please remove label selector"
	unsupportedConditionNoTargetRef       = "Cannot read targetRef"
	unsupportedConditionMudaMudaMuda      = "Error checking if target is a top level controller: muda muda muda"
	unsupportedTargetRefHasParent         = "The targetRef controller has a parent but it should point to a top-level controller"
)

const (
	kind       = "dodokind"
	name1      = "dotaro"
	name2      = "doseph"
	namespace  = "testNamespace"
	apiVersion = "stardust"
)

func TestLegacySelector(t *testing.T) {

	type testCase struct {
		name                      string
		legacySelector            labels.Selector
		selector                  labels.Selector
		fetchSelectorError        error
		targetRef                 *v1.CrossVersionObjectReference
		topLevelKey               *controllerfetcher.ControllerKeyWithAPIVersion
		findTopLevelError         error
		expectedSelector          labels.Selector
		expectedConfigUnsupported *string
		expectedConfigDeprecated  *string
	}

	testCases := []testCase{
		{
			name:                      "no selector",
			legacySelector:            nil,
			selector:                  nil,
			fetchSelectorError:        fmt.Errorf("targetRef not defined"),
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionTextFromFetcher,
			expectedConfigDeprecated:  nil,
		},
		{
			name:                      "also no selector but no error",
			legacySelector:            nil,
			selector:                  nil,
			fetchSelectorError:        nil,
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionNoExtraText,
			expectedConfigDeprecated:  nil,
		},
		{
			name:                      "legacy selector no ref",
			legacySelector:            parseLabelSelector("app = test"),
			selector:                  nil,
			fetchSelectorError:        fmt.Errorf("targetRef not defined"),
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionNoLongerSupported,
			expectedConfigDeprecated:  nil,
		}, {
			name: "targetRef selector",
			// the only valid option since v1beta1 removal
			legacySelector:     nil,
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			targetRef: &v1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topLevelKey: &controllerfetcher.ControllerKeyWithAPIVersion{
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
		}, {
			name:                      "new and legacy selector",
			legacySelector:            parseLabelSelector("app = test1"),
			selector:                  parseLabelSelector("app = test2"),
			fetchSelectorError:        nil,
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionBothDefined,
			expectedConfigDeprecated:  nil,
		},
		{
			name:               "can't decide if top-level-ref",
			legacySelector:     nil,
			selector:           nil,
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			targetRef: &v1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			expectedConfigUnsupported: &unsupportedConditionNoTargetRef,
		},
		{
			name:               "non-top-level targetRef",
			legacySelector:     nil,
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			targetRef: &v1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topLevelKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name2,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedConfigUnsupported: &unsupportedTargetRefHasParent,
		},
		{
			name:               "error checking if top-level-ref",
			legacySelector:     nil,
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   labels.Nothing(),
			targetRef: &v1.CrossVersionObjectReference{
				Kind:       "doestar",
				Name:       "doseph-doestar",
				APIVersion: "taxonomy",
			},
			expectedConfigUnsupported: &unsupportedConditionMudaMudaMuda,
			findTopLevelError:         fmt.Errorf("muda muda muda"),
		},
		{
			name:               "top-level target ref",
			legacySelector:     nil,
			selector:           parseLabelSelector("app = test"),
			fetchSelectorError: nil,
			expectedSelector:   parseLabelSelector("app = test"),
			targetRef: &v1.CrossVersionObjectReference{
				Kind:       kind,
				Name:       name1,
				APIVersion: apiVersion,
			},
			topLevelKey: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Kind:      kind,
					Name:      name1,
					Namespace: namespace,
				},
				ApiVersion: apiVersion,
			},
			expectedConfigUnsupported: nil,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			vpa := test.VerticalPodAutoscaler().WithName("testVpa").WithContainer("container").WithNamespace("testNamespace").WithTargetRef(tc.targetRef).Get()
			vpaLister := &test.VerticalPodAutoscalerListerMock{}
			vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpa}, nil)

			legacyTargetSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)
			targetSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)

			clusterState := model.NewClusterState()

			clusterStateFeeder := clusterStateFeeder{
				vpaLister:             vpaLister,
				clusterState:          clusterState,
				legacySelectorFetcher: legacyTargetSelectorFetcher,
				selectorFetcher:       targetSelectorFetcher,
				controllerFetcher: &fakeControllerFetcher{
					key: tc.topLevelKey,
					err: tc.findTopLevelError,
				},
			}

			// legacyTargetSelectorFetcher is called twice:
			// - one time to determine ultimate selector
			// - one time to check if object uses deprecated API
			legacyTargetSelectorFetcher.EXPECT().Fetch(vpa).Times(2).Return(tc.legacySelector, nil)
			targetSelectorFetcher.EXPECT().Fetch(vpa).Return(tc.selector, tc.fetchSelectorError)
			clusterStateFeeder.LoadVPAs()

			vpaID := model.VpaID{
				Namespace: vpa.Namespace,
				VpaName:   vpa.Name,
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
			clusterState := model.NewClusterState()
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
