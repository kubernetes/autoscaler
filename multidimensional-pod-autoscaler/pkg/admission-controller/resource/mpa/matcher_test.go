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

package mpa

import (
	"testing"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	target_mock "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/test"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	test_vpa "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func parseLabelSelector(selector string) labels.Selector {
	labelSelector, _ := meta.ParseToLabelSelector(selector)
	parsedSelector, _ := meta.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

func TestGetMatchingVpa(t *testing.T) {
	podBuilder := test_vpa.Pod().WithName("test-pod").WithLabels(map[string]string{"app": "test"}).
		AddContainer(test.Container().WithName("i-am-container").Get())
	mpaBuilder := test.MultidimPodAutoscaler().WithContainer("i-am-container")
	testCases := []struct {
		name            string
		pod             *core.Pod
		mpas            []*mpa_types.MultidimPodAutoscaler
		labelSelector   string
		expectedFound   bool
		expectedVpaName string
	}{
		{
			name: "matching selector",
			pod:  podBuilder.Get(),
			mpas: []*mpa_types.MultidimPodAutoscaler{
				mpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).WithName("auto-mpa").Get(),
			},
			labelSelector:   "app = test",
			expectedFound:   true,
			expectedVpaName: "auto-mpa",
		}, {
			name: "not matching selector",
			pod:  podBuilder.Get(),
			mpas: []*mpa_types.MultidimPodAutoscaler{
				mpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).WithName("auto-mpa").Get(),
			},
			labelSelector: "app = differentApp",
			expectedFound: false,
		}, {
			name: "off mode",
			pod:  podBuilder.Get(),
			mpas: []*mpa_types.MultidimPodAutoscaler{
				mpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).WithName("off-mpa").Get(),
			},
			labelSelector: "app = test",
			expectedFound: false,
		}, {
			name: "two vpas one in off mode",
			pod:  podBuilder.Get(),
			mpas: []*mpa_types.MultidimPodAutoscaler{
				mpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).WithName("off-mpa").Get(),
				mpaBuilder.WithUpdateMode(vpa_types.UpdateModeAuto).WithName("auto-mpa").Get(),
			},
			labelSelector:   "app = test",
			expectedFound:   true,
			expectedVpaName: "auto-mpa",
		}, {
			name: "initial mode",
			pod:  podBuilder.Get(),
			mpas: []*mpa_types.MultidimPodAutoscaler{
				mpaBuilder.WithUpdateMode(vpa_types.UpdateModeInitial).WithName("initial-mpa").Get(),
			},
			labelSelector:   "app = test",
			expectedFound:   true,
			expectedVpaName: "initial-mpa",
		}, {
			name:          "no vpa objects",
			pod:           podBuilder.Get(),
			mpas:          []*mpa_types.MultidimPodAutoscaler{},
			labelSelector: "app = test",
			expectedFound: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSelectorFetcher := target_mock.NewMockMpaTargetSelectorFetcher(ctrl)

			mpaNamespaceLister := &test.MultidimPodAutoscalerListerMock{}
			mpaNamespaceLister.On("List").Return(tc.mpas, nil)

			mpaLister := &test.MultidimPodAutoscalerListerMock{}
			mpaLister.On("MultidimPodAutoscalers", "default").Return(mpaNamespaceLister)

			mockSelectorFetcher.EXPECT().Fetch(gomock.Any()).AnyTimes().Return(parseLabelSelector(tc.labelSelector), nil)
			matcher := NewMatcher(mpaLister, mockSelectorFetcher)

			mpa := matcher.GetMatchingMPA(tc.pod)
			if tc.expectedFound && assert.NotNil(t, mpa) {
				assert.Equal(t, tc.expectedVpaName, mpa.Name)
			} else {
				assert.Nil(t, mpa)
			}
		})
	}
}
