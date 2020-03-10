/*
Copyright 2020 The Kubernetes Authors.

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

package vpa

import (
	"testing"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	target_mock "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func parseLabelSelector(selector string) labels.Selector {
	labelSelector, _ := meta.ParseToLabelSelector(selector)
	parsedSelector, _ := meta.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

func TestGetMatchingVpa(t *testing.T) {
	podBuilder := test.Pod().WithName("test-pod").WithLabels(map[string]string{"app": "test"}).
		AddContainer(test.Container().WithName("i-am-container").Get())
	vpaBuilder := test.VerticalPodAutoscaler().WithContainer("i-am-container")
	testCases := []struct {
		name            string
		pod             *core.Pod
		vpas            []*vpa_types.VerticalPodAutoscaler
		labelSelector   string
		expectedFound   bool
		expectedVpaName string
	}{
		{
			name: "matching selector",
			pod:  podBuilder.Get(),
			vpas: []*vpa_types.VerticalPodAutoscaler{
				vpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).WithName("auto-vpa").Get(),
			},
			labelSelector:   "app = test",
			expectedFound:   true,
			expectedVpaName: "auto-vpa",
		}, {
			name: "not matching selector",
			pod:  podBuilder.Get(),
			vpas: []*vpa_types.VerticalPodAutoscaler{
				vpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).WithName("auto-vpa").Get(),
			},
			labelSelector: "app = differentApp",
			expectedFound: false,
		}, {
			name: "off mode",
			pod:  podBuilder.Get(),
			vpas: []*vpa_types.VerticalPodAutoscaler{
				vpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).WithName("off-vpa").Get(),
			},
			labelSelector: "app = test",
			expectedFound: false,
		}, {
			name: "two vpas one in off mode",
			pod:  podBuilder.Get(),
			vpas: []*vpa_types.VerticalPodAutoscaler{
				vpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).WithName("off-vpa").Get(),
				vpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).WithName("auto-vpa").Get(),
			},
			labelSelector:   "app = test",
			expectedFound:   true,
			expectedVpaName: "auto-vpa",
		}, {
			name: "initial mode",
			pod:  podBuilder.Get(),
			vpas: []*vpa_types.VerticalPodAutoscaler{
				vpaBuilder.WithUpdateMode(vpa_types.UpdateModeInitial).WithName("initial-vpa").Get(),
			},
			labelSelector:   "app = test",
			expectedFound:   true,
			expectedVpaName: "initial-vpa",
		}, {
			name:          "no vpa objects",
			pod:           podBuilder.Get(),
			vpas:          []*vpa_types.VerticalPodAutoscaler{},
			labelSelector: "app = test",
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)

			vpaNamespaceLister := &test.VerticalPodAutoscalerListerMock{}
			vpaNamespaceLister.On("List").Return(tc.vpas, nil)

			vpaLister := &test.VerticalPodAutoscalerListerMock{}
			vpaLister.On("VerticalPodAutoscalers", "default").Return(vpaNamespaceLister)

			mockSelectorFetcher.EXPECT().Fetch(gomock.Any()).AnyTimes().Return(parseLabelSelector(tc.labelSelector), nil)
			matcher := NewMatcher(vpaLister, mockSelectorFetcher)

			vpa := matcher.GetMatchingVPA(tc.pod)
			if tc.expectedFound && assert.NotNil(t, vpa) {
				assert.Equal(t, tc.expectedVpaName, vpa.Name)
			} else {
				assert.Nil(t, vpa)
			}
		})
	}
}
