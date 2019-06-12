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

package logic

import (
	"fmt"
	"math"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	target_mock "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func parseLabelSelector(selector string) labels.Selector {
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

func mustParseResourcePointer(val string) *resource.Quantity {
	q := resource.MustParse(val)
	return &q
}

type fakeLimitRangeCalculator struct {
	containerLimitRange *apiv1.LimitRangeItem
	containerErr        error
	podLimitRange       *apiv1.LimitRangeItem
	podErr              error
}

func (nlrc *fakeLimitRangeCalculator) GetContainerLimitRangeItem(namespace string) (*apiv1.LimitRangeItem, error) {
	return nlrc.containerLimitRange, nlrc.containerErr
}

func (nlrc *fakeLimitRangeCalculator) GetPodLimitRangeItem(namespace string) (*apiv1.LimitRangeItem, error) {
	return nlrc.podLimitRange, nlrc.podErr
}

func TestUpdateResourceRequests(t *testing.T) {
	containerName := "container1"
	vpaName := "vpa1"
	labels := map[string]string{"app": "testingApp"}
	vpaBuilder := test.VerticalPodAutoscaler().
		WithName(vpaName).
		WithContainer(containerName).
		WithTarget("2", "200Mi").
		WithMinAllowed("1", "100Mi").
		WithMaxAllowed("3", "1Gi")
	vpa := vpaBuilder.Get()

	uninitialized := test.Pod().WithName("test_uninitialized").
		AddContainer(test.Container().WithName(containerName).Get()).
		WithLabels(labels).Get()

	initializedContainer := test.Container().WithName(containerName).
		WithCPURequest(resource.MustParse("1")).WithCPURequest(resource.MustParse("2")).WithMemRequest(resource.MustParse("100Mi")).Get()
	initialized := test.Pod().WithName("test_initialized").
		AddContainer(initializedContainer).WithLabels(labels).Get()

	limitsMatchRequestsContainer := test.Container().WithName(containerName).
		WithCPURequest(resource.MustParse("2")).WithCPULimit(resource.MustParse("2")).
		WithMemRequest(resource.MustParse("200Mi")).WithMemLimit(resource.MustParse("200Mi")).Get()
	limitsMatchRequestsPod := test.Pod().WithName("test_initialized").
		AddContainer(limitsMatchRequestsContainer).WithLabels(labels).Get()

	containerWithDoubleLimit := test.Container().WithName(containerName).
		WithCPURequest(resource.MustParse("1")).WithCPULimit(resource.MustParse("2")).
		WithMemRequest(resource.MustParse("100Mi")).WithMemLimit(resource.MustParse("200Mi")).Get()
	podWithDoubleLimit := test.Pod().WithName("test_initialized").
		AddContainer(containerWithDoubleLimit).WithLabels(labels).Get()

	containerWithTenfoldLimit := test.Container().WithName(containerName).
		WithCPURequest(resource.MustParse("1")).WithCPULimit(resource.MustParse("10")).
		WithMemRequest(resource.MustParse("100Mi")).WithMemLimit(resource.MustParse("1000Mi")).Get()
	podWithTenfoldLimit := test.Pod().WithName("test_initialized").
		AddContainer(containerWithTenfoldLimit).WithLabels(labels).Get()

	limitsNoRequestsContainer := test.Container().WithName(containerName).
		WithCPULimit(resource.MustParse("2")).WithMemLimit(resource.MustParse("200Mi")).Get()
	limitsNoRequestsPod := test.Pod().WithName("test_initialized").
		AddContainer(limitsNoRequestsContainer).WithLabels(labels).Get()

	offVPA := vpaBuilder.WithUpdateMode(vpa_types.UpdateModeOff).Get()

	targetBelowMinVPA := vpaBuilder.WithTarget("3", "150Mi").WithMinAllowed("4", "300Mi").WithMaxAllowed("5", "1Gi").Get()
	targetAboveMaxVPA := vpaBuilder.WithTarget("7", "2Gi").WithMinAllowed("4", "300Mi").WithMaxAllowed("5", "1Gi").Get()
	vpaWithHighMemory := vpaBuilder.WithTarget("2", "1000Mi").WithMaxAllowed("3", "3Gi").Get()
	vpaWithExabyteRecommendation := vpaBuilder.WithTarget("1Ei", "1Ei").WithMaxAllowed("1Ei", "1Ei").Get()

	vpaWithEmptyRecommendation := vpaBuilder.Get()
	vpaWithEmptyRecommendation.Status.Recommendation = &vpa_types.RecommendedPodResources{}
	vpaWithNilRecommendation := vpaBuilder.Get()
	vpaWithNilRecommendation.Status.Recommendation = nil

	testCases := []struct {
		name              string
		pod               *apiv1.Pod
		vpas              []*vpa_types.VerticalPodAutoscaler
		expectedAction    bool
		expectedError     error
		expectedMem       resource.Quantity
		expectedCPU       resource.Quantity
		expectedCPULimit  *resource.Quantity
		expectedMemLimit  *resource.Quantity
		limitRange        *apiv1.LimitRangeItem
		limitRangeCalcErr error
		annotations       vpa_api_util.ContainerToAnnotationsMap
		labelSelector     string
	}{
		{
			name:           "uninitialized pod",
			pod:            uninitialized,
			vpas:           []*vpa_types.VerticalPodAutoscaler{vpa},
			expectedAction: true,
			expectedMem:    resource.MustParse("200Mi"),
			expectedCPU:    resource.MustParse("2"),
			labelSelector:  "app = testingApp",
		},
		{
			name:           "target below min",
			pod:            uninitialized,
			vpas:           []*vpa_types.VerticalPodAutoscaler{targetBelowMinVPA},
			expectedAction: true,
			expectedMem:    resource.MustParse("300Mi"), // MinMemory is expected to be used
			expectedCPU:    resource.MustParse("4"),     // MinCpu is expected to be used
			annotations: vpa_api_util.ContainerToAnnotationsMap{
				containerName: []string{"cpu capped to minAllowed", "memory capped to minAllowed"},
			},
			labelSelector: "app = testingApp",
		},
		{
			name:           "target above max",
			pod:            uninitialized,
			vpas:           []*vpa_types.VerticalPodAutoscaler{targetAboveMaxVPA},
			expectedAction: true,
			expectedMem:    resource.MustParse("1Gi"), // MaxMemory is expected to be used
			expectedCPU:    resource.MustParse("5"),   // MaxCpu is expected to be used
			annotations: vpa_api_util.ContainerToAnnotationsMap{
				containerName: []string{"cpu capped to maxAllowed", "memory capped to maxAllowed"},
			},
			labelSelector: "app = testingApp",
		},
		{
			name:           "initialized pod",
			pod:            initialized,
			vpas:           []*vpa_types.VerticalPodAutoscaler{vpa},
			expectedAction: true,
			expectedMem:    resource.MustParse("200Mi"),
			expectedCPU:    resource.MustParse("2"),
			labelSelector:  "app = testingApp",
		},
		{
			name:           "high memory",
			pod:            initialized,
			vpas:           []*vpa_types.VerticalPodAutoscaler{vpaWithHighMemory},
			expectedAction: true,
			expectedMem:    resource.MustParse("1000Mi"),
			expectedCPU:    resource.MustParse("2"),
			labelSelector:  "app = testingApp",
		},
		{
			name:           "not matching selecetor",
			pod:            uninitialized,
			vpas:           []*vpa_types.VerticalPodAutoscaler{vpa},
			expectedAction: false,
			labelSelector:  "app = differentApp",
		},
		{
			name:           "off mode",
			pod:            uninitialized,
			vpas:           []*vpa_types.VerticalPodAutoscaler{offVPA},
			expectedAction: false,
			labelSelector:  "app = testingApp",
		},
		{
			name:           "two vpas one in off mode",
			pod:            uninitialized,
			vpas:           []*vpa_types.VerticalPodAutoscaler{offVPA, vpa},
			expectedAction: true,
			expectedMem:    resource.MustParse("200Mi"),
			expectedCPU:    resource.MustParse("2"),
			labelSelector:  "app = testingApp",
		},
		{
			name:           "empty recommendation",
			pod:            initialized,
			vpas:           []*vpa_types.VerticalPodAutoscaler{vpaWithEmptyRecommendation},
			expectedAction: true,
			expectedMem:    resource.MustParse("0"),
			expectedCPU:    resource.MustParse("0"),
			labelSelector:  "app = testingApp",
		},
		{
			pod:            initialized,
			vpas:           []*vpa_types.VerticalPodAutoscaler{vpaWithNilRecommendation},
			expectedAction: true,
			expectedMem:    resource.MustParse("0"),
			expectedCPU:    resource.MustParse("0"),
			labelSelector:  "app = testingApp",
		},
		{
			name:             "guaranteed resources",
			pod:              limitsMatchRequestsPod,
			vpas:             []*vpa_types.VerticalPodAutoscaler{vpa},
			expectedAction:   true,
			expectedMem:      resource.MustParse("200Mi"),
			expectedCPU:      resource.MustParse("2"),
			expectedCPULimit: mustParseResourcePointer("2"),
			expectedMemLimit: mustParseResourcePointer("200Mi"),
			labelSelector:    "app = testingApp",
		},
		{
			name:             "guaranteed resources - no request",
			pod:              limitsNoRequestsPod,
			vpas:             []*vpa_types.VerticalPodAutoscaler{vpa},
			expectedAction:   true,
			expectedMem:      resource.MustParse("200Mi"),
			expectedCPU:      resource.MustParse("2"),
			expectedCPULimit: mustParseResourcePointer("2"),
			expectedMemLimit: mustParseResourcePointer("200Mi"),
			labelSelector:    "app = testingApp",
		},
		{
			name:             "proportional limit",
			pod:              podWithDoubleLimit,
			vpas:             []*vpa_types.VerticalPodAutoscaler{vpa},
			expectedAction:   true,
			expectedCPU:      resource.MustParse("2"),
			expectedMem:      resource.MustParse("200Mi"),
			expectedCPULimit: mustParseResourcePointer("4"),
			expectedMemLimit: mustParseResourcePointer("400Mi"),
			labelSelector:    "app = testingApp",
		},
		{
			name:             "limit over int64",
			pod:              podWithTenfoldLimit,
			vpas:             []*vpa_types.VerticalPodAutoscaler{vpaWithExabyteRecommendation},
			expectedAction:   true,
			expectedCPU:      resource.MustParse("1Ei"),
			expectedMem:      resource.MustParse("1Ei"),
			expectedCPULimit: resource.NewMilliQuantity(math.MaxInt64, resource.DecimalExponent),
			expectedMemLimit: resource.NewMilliQuantity(math.MaxInt64, resource.DecimalExponent),
			labelSelector:    "app = testingApp",
			annotations: vpa_api_util.ContainerToAnnotationsMap{
				containerName: []string{
					"cpu: failed to keep limit to request ratio; capping limit to int64",
					"memory: failed to keep limit to request ratio; capping limit to int64",
				},
			},
		},
		{
			name:              "limit range calculation error",
			pod:               initialized,
			vpas:              []*vpa_types.VerticalPodAutoscaler{vpa},
			limitRangeCalcErr: fmt.Errorf("oh no"),
			expectedAction:    false,
			expectedError:     fmt.Errorf("error getting containerLimitRange: oh no"),
		},
		{
			name:             "proportional limit from default",
			pod:              initialized,
			vpas:             []*vpa_types.VerticalPodAutoscaler{vpa},
			expectedAction:   true,
			expectedCPU:      resource.MustParse("2"),
			expectedMem:      resource.MustParse("200Mi"),
			expectedCPULimit: mustParseResourcePointer("2"),
			expectedMemLimit: mustParseResourcePointer("200Mi"),
			labelSelector:    "app = testingApp",
			limitRange: &apiv1.LimitRangeItem{
				Type: apiv1.LimitTypeContainer,
				Default: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("2"),
					apiv1.ResourceMemory: resource.MustParse("100Mi"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf(tc.name), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)

			vpaNamespaceLister := &test.VerticalPodAutoscalerListerMock{}
			vpaNamespaceLister.On("List").Return(tc.vpas, nil)

			vpaLister := &test.VerticalPodAutoscalerListerMock{}
			vpaLister.On("VerticalPodAutoscalers", "default").Return(vpaNamespaceLister)

			mockSelectorFetcher.EXPECT().Fetch(gomock.Any()).AnyTimes().Return(parseLabelSelector(tc.labelSelector), nil)

			recommendationProvider := &recommendationProvider{
				vpaLister:               vpaLister,
				recommendationProcessor: vpa_api_util.NewCappingRecommendationProcessor(limitrange.NewNoopLimitsCalculator()),
				selectorFetcher:         mockSelectorFetcher,
				limitsRangeCalculator: &fakeLimitRangeCalculator{
					containerLimitRange: tc.limitRange,
					containerErr:        tc.limitRangeCalcErr,
				},
			}

			resources, annotations, name, err := recommendationProvider.GetContainersResourcesForPod(tc.pod)

			if tc.expectedAction {
				assert.Equal(t, vpaName, name)
				assert.Nil(t, err)
				if !assert.Equal(t, len(resources), 1) {
					return
				}

				cpuRequest := resources[0].Requests[apiv1.ResourceCPU]
				assert.Equal(t, tc.expectedCPU.Value(), cpuRequest.Value(), "cpu request doesn't match")

				memoryRequest := resources[0].Requests[apiv1.ResourceMemory]
				assert.Equal(t, tc.expectedMem.Value(), memoryRequest.Value(), "memory request doesn't match")

				cpuLimit, cpuLimitPresent := resources[0].Limits[apiv1.ResourceCPU]
				if tc.expectedCPULimit == nil {
					assert.False(t, cpuLimitPresent, "expected no cpu limit, got %s", cpuLimit.String())
				} else {
					if assert.True(t, cpuLimitPresent, "expected cpu limit, but it's missing") {
						assert.Equal(t, tc.expectedCPULimit.MilliValue(), cpuLimit.MilliValue(), "cpu limit doesn't match")
					}
				}

				memLimit, memLimitPresent := resources[0].Limits[apiv1.ResourceMemory]
				if tc.expectedMemLimit == nil {
					assert.False(t, memLimitPresent, "expected no memory limit, got %s", memLimit.String())
				} else {
					if assert.True(t, memLimitPresent, "expected memory limit, but it's missing") {
						assert.Equal(t, tc.expectedMemLimit.MilliValue(), memLimit.MilliValue(), "memory limit doesn't match")
					}
				}

				assert.Len(t, annotations, len(tc.annotations))
				if len(tc.annotations) > 0 {
					for annotationKey, annotationValues := range tc.annotations {
						assert.Len(t, annotations[annotationKey], len(annotationValues))
						for _, annotation := range annotationValues {
							assert.Contains(t, annotations[annotationKey], annotation)
						}
					}
				}
			} else {
				assert.Empty(t, resources)
				if tc.expectedError != nil {
					assert.Error(t, err)
					assert.Equal(t, tc.expectedError.Error(), err.Error())
				} else {
					assert.NoError(t, err)
				}
			}

		})

	}
}
