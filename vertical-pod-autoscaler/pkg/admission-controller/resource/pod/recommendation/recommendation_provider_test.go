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

package recommendation

import (
	"fmt"
	"math"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"

	"github.com/stretchr/testify/assert"
)

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
		WithCPURequest(resource.MustParse("2")).WithMemRequest(resource.MustParse("100Mi")).Get()
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

	targetBelowMinVPA := vpaBuilder.WithTarget("3", "150Mi").WithMinAllowed("4", "300Mi").WithMaxAllowed("5", "1Gi").Get()
	targetAboveMaxVPA := vpaBuilder.WithTarget("7", "2Gi").WithMinAllowed("4", "300Mi").WithMaxAllowed("5", "1Gi").Get()
	vpaWithHighMemory := vpaBuilder.WithTarget("2", "1000Mi").WithMaxAllowed("3", "3Gi").Get()
	vpaWithExabyteRecommendation := vpaBuilder.WithTarget("1Ei", "1Ei").WithMaxAllowed("1Ei", "1Ei").Get()

	resourceRequestsAndLimitsVPA := vpaBuilder.WithControlledValues(vpa_types.ContainerControlledValuesRequestsAndLimits).Get()
	resourceRequestsOnlyVPA := vpaBuilder.WithControlledValues(vpa_types.ContainerControlledValuesRequestsOnly).Get()
	resourceRequestsOnlyVPAHighTarget := vpaBuilder.WithControlledValues(vpa_types.ContainerControlledValuesRequestsOnly).
		WithTarget("3", "500Mi").WithMaxAllowed("5", "1Gi").Get()

	vpaWithEmptyRecommendation := vpaBuilder.Get()
	vpaWithEmptyRecommendation.Status.Recommendation = &vpa_types.RecommendedPodResources{}
	vpaWithNilRecommendation := vpaBuilder.Get()
	vpaWithNilRecommendation.Status.Recommendation = nil

	testCases := []struct {
		name              string
		pod               *apiv1.Pod
		vpa               *vpa_types.VerticalPodAutoscaler
		expectedAction    bool
		expectedError     error
		expectedMem       resource.Quantity
		expectedCPU       resource.Quantity
		expectedCPULimit  *resource.Quantity
		expectedMemLimit  *resource.Quantity
		limitRange        *apiv1.LimitRangeItem
		limitRangeCalcErr error
		annotations       vpa_api_util.ContainerToAnnotationsMap
	}{
		{
			name:           "uninitialized pod",
			pod:            uninitialized,
			vpa:            vpa,
			expectedAction: true,
			expectedMem:    resource.MustParse("200Mi"),
			expectedCPU:    resource.MustParse("2"),
		},
		{
			name:           "target below min",
			pod:            uninitialized,
			vpa:            targetBelowMinVPA,
			expectedAction: true,
			expectedMem:    resource.MustParse("300Mi"), // MinMemory is expected to be used
			expectedCPU:    resource.MustParse("4"),     // MinCpu is expected to be used
			annotations: vpa_api_util.ContainerToAnnotationsMap{
				containerName: []string{"cpu capped to minAllowed", "memory capped to minAllowed"},
			},
		},
		{
			name:           "target above max",
			pod:            uninitialized,
			vpa:            targetAboveMaxVPA,
			expectedAction: true,
			expectedMem:    resource.MustParse("1Gi"), // MaxMemory is expected to be used
			expectedCPU:    resource.MustParse("5"),   // MaxCpu is expected to be used
			annotations: vpa_api_util.ContainerToAnnotationsMap{
				containerName: []string{"cpu capped to maxAllowed", "memory capped to maxAllowed"},
			},
		},
		{
			name:           "initialized pod",
			pod:            initialized,
			vpa:            vpa,
			expectedAction: true,
			expectedMem:    resource.MustParse("200Mi"),
			expectedCPU:    resource.MustParse("2"),
		},
		{
			name:           "high memory",
			pod:            initialized,
			vpa:            vpaWithHighMemory,
			expectedAction: true,
			expectedMem:    resource.MustParse("1000Mi"),
			expectedCPU:    resource.MustParse("2"),
		},
		{
			name:           "empty recommendation",
			pod:            initialized,
			vpa:            vpaWithEmptyRecommendation,
			expectedAction: true,
			expectedMem:    resource.MustParse("0"),
			expectedCPU:    resource.MustParse("0"),
		},
		{
			name:           "nil recommendation",
			pod:            initialized,
			vpa:            vpaWithNilRecommendation,
			expectedAction: true,
			expectedMem:    resource.MustParse("0"),
			expectedCPU:    resource.MustParse("0"),
		},
		{
			name:             "guaranteed resources",
			pod:              limitsMatchRequestsPod,
			vpa:              vpa,
			expectedAction:   true,
			expectedMem:      resource.MustParse("200Mi"),
			expectedCPU:      resource.MustParse("2"),
			expectedCPULimit: mustParseResourcePointer("2"),
			expectedMemLimit: mustParseResourcePointer("200Mi"),
		},
		{
			name:             "guaranteed resources - no request",
			pod:              limitsNoRequestsPod,
			vpa:              vpa,
			expectedAction:   true,
			expectedMem:      resource.MustParse("200Mi"),
			expectedCPU:      resource.MustParse("2"),
			expectedCPULimit: mustParseResourcePointer("2"),
			expectedMemLimit: mustParseResourcePointer("200Mi"),
		},
		{
			name:             "proportional limit - as default",
			pod:              podWithDoubleLimit,
			vpa:              vpa,
			expectedAction:   true,
			expectedCPU:      resource.MustParse("2"),
			expectedMem:      resource.MustParse("200Mi"),
			expectedCPULimit: mustParseResourcePointer("4"),
			expectedMemLimit: mustParseResourcePointer("400Mi"),
		},
		{
			name:             "proportional limit - set explicit",
			pod:              podWithDoubleLimit,
			vpa:              resourceRequestsAndLimitsVPA,
			expectedAction:   true,
			expectedCPU:      resource.MustParse("2"),
			expectedMem:      resource.MustParse("200Mi"),
			expectedCPULimit: mustParseResourcePointer("4"),
			expectedMemLimit: mustParseResourcePointer("400Mi"),
		},
		{
			name:           "disabled limit scaling",
			pod:            podWithDoubleLimit,
			vpa:            resourceRequestsOnlyVPA,
			expectedAction: true,
			expectedCPU:    resource.MustParse("2"),
			expectedMem:    resource.MustParse("200Mi"),
		},
		{
			name:           "disabled limit scaling - requests capped at limit",
			pod:            podWithDoubleLimit,
			vpa:            resourceRequestsOnlyVPAHighTarget,
			expectedAction: true,
			expectedCPU:    resource.MustParse("2"),
			expectedMem:    resource.MustParse("200Mi"),
			annotations: vpa_api_util.ContainerToAnnotationsMap{
				containerName: []string{
					"cpu capped to container limit",
					"memory capped to container limit",
				},
			},
		},
		{
			name:             "limit over int64",
			pod:              podWithTenfoldLimit,
			vpa:              vpaWithExabyteRecommendation,
			expectedAction:   true,
			expectedCPU:      resource.MustParse("1Ei"),
			expectedMem:      resource.MustParse("1Ei"),
			expectedCPULimit: resource.NewMilliQuantity(math.MaxInt64, resource.DecimalExponent),
			expectedMemLimit: resource.NewMilliQuantity(math.MaxInt64, resource.DecimalExponent),
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
			vpa:               vpa,
			limitRangeCalcErr: fmt.Errorf("oh no"),
			expectedAction:    false,
			expectedError:     fmt.Errorf("error getting containerLimitRange: oh no"),
		},
		{
			name:             "proportional limit from default",
			pod:              initialized,
			vpa:              vpa,
			expectedAction:   true,
			expectedCPU:      resource.MustParse("2"),
			expectedMem:      resource.MustParse("200Mi"),
			expectedCPULimit: mustParseResourcePointer("2"),
			expectedMemLimit: mustParseResourcePointer("200Mi"),
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
		t.Run(tc.name, func(t *testing.T) {
			recommendationProvider := &recommendationProvider{
				recommendationProcessor: vpa_api_util.NewCappingRecommendationProcessor(limitrange.NewNoopLimitsCalculator()),
				limitsRangeCalculator: &fakeLimitRangeCalculator{
					containerLimitRange: tc.limitRange,
					containerErr:        tc.limitRangeCalcErr,
				},
			}

			resources, annotations, err := recommendationProvider.GetContainersResourcesForPod(tc.pod, tc.vpa)

			if tc.expectedAction {
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
