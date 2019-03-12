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

	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	target_mock "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

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
)

func TestLegacySelector(t *testing.T) {

	type testCase struct {
		legacySelector            labels.Selector
		selector                  labels.Selector
		fetchSelectorError        error
		expectedSelector          labels.Selector
		expectedConfigUnsupported *string
		expectedConfigDeprecated  *string
	}

	testCases := []testCase{
		{
			legacySelector:            nil,
			selector:                  nil,
			fetchSelectorError:        fmt.Errorf("targetRef not defined"),
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionTextFromFetcher,
			expectedConfigDeprecated:  nil,
		},
		{
			legacySelector:            nil,
			selector:                  nil,
			fetchSelectorError:        nil,
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionNoExtraText,
			expectedConfigDeprecated:  nil,
		},
		{
			legacySelector:            parseLabelSelector("app = test"),
			selector:                  nil,
			fetchSelectorError:        fmt.Errorf("targetRef not defined"),
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionNoLongerSupported,
			expectedConfigDeprecated:  nil,
		}, {
			// the only valid option since v1beta1 removal
			legacySelector:            nil,
			selector:                  parseLabelSelector("app = test"),
			fetchSelectorError:        nil,
			expectedSelector:          parseLabelSelector("app = test"),
			expectedConfigUnsupported: nil,
			expectedConfigDeprecated:  nil,
		}, {
			legacySelector:            parseLabelSelector("app = test1"),
			selector:                  parseLabelSelector("app = test2"),
			fetchSelectorError:        nil,
			expectedSelector:          labels.Nothing(),
			expectedConfigUnsupported: &unsupportedConditionBothDefined,
			expectedConfigDeprecated:  nil,
		},
	}

	for i, tc := range testCases {

		t.Run(fmt.Sprintf("test case number: %d", i), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			vpa := test.VerticalPodAutoscaler().WithName("testVpa").WithContainer("container").WithNamespace("testNamespace").Get()
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
