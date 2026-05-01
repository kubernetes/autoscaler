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
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/version"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/utils/dump"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

func mustParseResourcePointer(val string) *resource.Quantity {
	q := resource.MustParse(val)
	return &q
}

type fakeLimitRangeCalculator struct {
	containerLimitRange *corev1.LimitRangeItem
	containerErr        error
	podLimitRange       *corev1.LimitRangeItem
	podErr              error
}

func (nlrc *fakeLimitRangeCalculator) GetContainerLimitRangeItem(namespace string) (*corev1.LimitRangeItem, error) {
	return nlrc.containerLimitRange, nlrc.containerErr
}

func (nlrc *fakeLimitRangeCalculator) GetPodLimitRangeItem(namespace string) (*corev1.LimitRangeItem, error) {
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
		WithMinAllowed(containerName, "1", "100Mi").
		WithMaxAllowed(containerName, "3", "1Gi").
		WithTargetResource("", "666") // Testing that this weird/empty resource will be purged
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

	targetBelowMinVPA := vpaBuilder.WithTarget("3", "150Mi").WithMinAllowed(containerName, "4", "300Mi").WithMaxAllowed(containerName, "5", "1Gi").Get()
	targetAboveMaxVPA := vpaBuilder.WithTarget("7", "2Gi").WithMinAllowed(containerName, "4", "300Mi").WithMaxAllowed(containerName, "5", "1Gi").Get()
	vpaWithHighMemory := vpaBuilder.WithTarget("2", "1000Mi").WithMinAllowed(containerName, "", "").WithMaxAllowed(containerName, "3", "3Gi").Get()
	vpaWithExabyteRecommendation := vpaBuilder.WithTarget("1Ei", "1Ei").WithMinAllowed(containerName, "", "").WithMaxAllowed(containerName, "1Ei", "1Ei").Get()

	resourceRequestsAndLimitsVPA := vpaBuilder.WithControlledValues(containerName, vpa_types.ContainerControlledValuesRequestsAndLimits).Get()
	resourceRequestsOnlyVPA := vpaBuilder.WithControlledValues(containerName, vpa_types.ContainerControlledValuesRequestsOnly).Get()
	resourceRequestsOnlyVPAHighTarget := vpaBuilder.WithControlledValues(containerName, vpa_types.ContainerControlledValuesRequestsOnly).
		WithTarget("3", "500Mi").WithMaxAllowed(containerName, "5", "1Gi").Get()

	vpaWithEmptyRecommendation := vpaBuilder.Get()
	vpaWithEmptyRecommendation.Status.Recommendation = &vpa_types.RecommendedPodResources{}
	vpaWithNilRecommendation := vpaBuilder.Get()
	vpaWithNilRecommendation.Status.Recommendation = nil

	testCases := []struct {
		name              string
		pod               *corev1.Pod
		vpa               *vpa_types.VerticalPodAutoscaler
		expectedAction    bool
		expectedError     error
		expectedMem       resource.Quantity
		expectedCPU       resource.Quantity
		expectedCPULimit  *resource.Quantity
		expectedMemLimit  *resource.Quantity
		limitRange        *corev1.LimitRangeItem
		limitRangeCalcErr error
		annotations       utils.ContainerToAnnotationsMap
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
			annotations: utils.ContainerToAnnotationsMap{
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
			annotations: utils.ContainerToAnnotationsMap{
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
			annotations: utils.ContainerToAnnotationsMap{
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
			expectedMemLimit: resource.NewQuantity(math.MaxInt64, resource.DecimalExponent),
			annotations: utils.ContainerToAnnotationsMap{
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
			limitRangeCalcErr: errors.New("oh no"),
			expectedAction:    false,
			expectedError:     errors.New("error getting containerLimitRange: oh no"),
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
			limitRange: &corev1.LimitRangeItem{
				Type: corev1.LimitTypeContainer,
				Default: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
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

			resources, annotations, _, err := recommendationProvider.GetContainersResourcesForPod(tc.pod, tc.vpa)

			if tc.expectedAction {
				assert.Nil(t, err)
				if !assert.Equal(t, len(resources), 1) {
					return
				}

				assert.NotContains(t, resources, "", "expected empty resource to be removed")

				cpuRequest := resources[0].Requests[corev1.ResourceCPU]
				assert.Equal(t, tc.expectedCPU.Value(), cpuRequest.Value(), "cpu request doesn't match")

				memoryRequest := resources[0].Requests[corev1.ResourceMemory]
				assert.Equal(t, tc.expectedMem.Value(), memoryRequest.Value(), "memory request doesn't match")

				cpuLimit, cpuLimitPresent := resources[0].Limits[corev1.ResourceCPU]
				if tc.expectedCPULimit == nil {
					assert.False(t, cpuLimitPresent, "expected no cpu limit, got %s", cpuLimit.String())
				} else {
					if assert.True(t, cpuLimitPresent, "expected cpu limit, but it's missing") {
						assert.Equal(t, tc.expectedCPULimit.MilliValue(), cpuLimit.MilliValue(), "cpu limit doesn't match")
					}
				}

				memLimit, memLimitPresent := resources[0].Limits[corev1.ResourceMemory]
				if tc.expectedMemLimit == nil {
					assert.False(t, memLimitPresent, "expected no memory limit, got %s", memLimit.String())
				} else {
					if assert.True(t, memLimitPresent, "expected memory limit, but it's missing") {
						assert.Equal(t, tc.expectedMemLimit.MilliValue(), memLimit.MilliValue(), "memory limit doesn't match")
					}
				}

				assert.Len(t, annotations.Container, len(tc.annotations))
				if len(tc.annotations) > 0 {
					for annotationKey, annotationValues := range tc.annotations {
						assert.Len(t, annotations.Container[annotationKey], len(annotationValues))
						for _, annotation := range annotationValues {
							assert.Contains(t, annotations.Container[annotationKey], annotation)
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

func TestGetContainersResources(t *testing.T) {
	testCases := []struct {
		name              string
		container         corev1.Container
		containerStatus   corev1.ContainerStatus
		vpa               *vpa_types.VerticalPodAutoscaler
		expectedResources []vpa_api_util.ContainerResources
		expectedCPU       *resource.Quantity
		expectedMem       *resource.Quantity
		expectedCPULimit  *resource.Quantity
		expectedMemLimit  *resource.Quantity
		addAll            bool
	}{
		{
			name:      "CPU and Memory recommendation, request and limits set",
			container: test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("1M")).WithCPULimit(resource.MustParse("10")).WithMemLimit(resource.MustParse("10M")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTarget("2", "2M").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20"),
						corev1.ResourceMemory: resource.MustParse("20M"),
					},
				},
			},
			addAll: true,
		},
		{
			name:      "CPU and Memory recommendation, only request set",
			container: test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("1M")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTarget("2", "2M").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
				},
			},
			addAll: true,
		},
		{
			name:      "CPU only recommendation, request and limits set",
			container: test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("1M")).WithCPULimit(resource.MustParse("10")).WithMemLimit(resource.MustParse("10M")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTargetResource(corev1.ResourceCPU, "2").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("1M"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20"),
						corev1.ResourceMemory: resource.MustParse("10M"),
					},
				},
			},
			addAll: true,
		},
		{
			name:      "CPU only recommendation, only request set",
			container: test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("1M")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTargetResource(corev1.ResourceCPU, "2").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("1M"),
					},
				},
			},
			addAll: true,
		},
		{
			name:      "CPU only recommendation, only CPU request set",
			container: test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTargetResource(corev1.ResourceCPU, "2").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("2"),
					},
				},
			},
			addAll: true,
		},
		{
			name:      "CPU only recommendation, only CPU request and limit set",
			container: test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithCPULimit(resource.MustParse("10")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTargetResource(corev1.ResourceCPU, "2").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("2"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("20"),
					},
				},
			},
			addAll: true,
		},
		{
			name: "CPU only recommendation, only CPU request and limit set, ContainerControlledValuesRequestOnly",
			container: test.Container().WithName("container").
				WithCPURequest(resource.MustParse("1")).
				WithCPULimit(resource.MustParse("10")).
				Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer("container").
				WithControlledValues("container", vpa_types.ContainerControlledValuesRequestsOnly).
				WithTargetResource(corev1.ResourceCPU, "2").
				Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("2"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("10"),
					},
				},
			},
			addAll: true,
		},
		{
			name:      "Memory only recommendation, request and limits set",
			container: test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("1M")).WithCPULimit(resource.MustParse("10")).WithMemLimit(resource.MustParse("10M")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTargetResource(corev1.ResourceMemory, "2M").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10"),
						corev1.ResourceMemory: resource.MustParse("20M"),
					},
				},
			},
			addAll: true,
		},
		{
			name:      "Memory only recommendation, only request set",
			container: test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("1M")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTargetResource(corev1.ResourceMemory, "2M").Get(),

			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
				},
			},
			addAll: true,
		},
		{
			name:      "Memory only recommendation, only memory request set",
			container: test.Container().WithName("container").WithMemRequest(resource.MustParse("1M")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTargetResource(corev1.ResourceMemory, "2M").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
				},
			},
			addAll: true,
		},
		{
			name:      "Memory only recommendation, only memory request and limit set",
			container: test.Container().WithName("container").WithMemRequest(resource.MustParse("1M")).WithMemLimit(resource.MustParse("10M")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTargetResource(corev1.ResourceMemory, "2M").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("20M"),
					},
				},
			},
			addAll: true,
		},
		{
			name: "Memory only recommendation, only memory request and limit set, ContainerControlledValuesRequestOnly",
			container: test.Container().WithName("container").
				WithMemRequest(resource.MustParse("1M")).
				WithMemLimit(resource.MustParse("10M")).
				Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer("container").
				WithControlledValues("container", vpa_types.ContainerControlledValuesRequestsOnly).
				WithTargetResource(corev1.ResourceMemory, "2M").
				Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("10M"),
					},
				},
			},
			addAll: true,
		},
		{
			name:      "CPU and Memory recommendation, request and limits set, addAll false",
			container: test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("1M")).WithCPULimit(resource.MustParse("10")).WithMemLimit(resource.MustParse("10M")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTarget("2", "2M").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20"),
						corev1.ResourceMemory: resource.MustParse("20M"),
					},
				},
			},
			addAll: false,
		},
		{
			name:      "CPU and Memory recommendation, only request set, addAll false",
			container: test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("1M")).Get(),
			vpa:       test.VerticalPodAutoscaler().WithContainer("container").WithTarget("2", "2M").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
				},
			},
			addAll: false,
		},
		{
			name:             "CPU only recommendation, request and limits set, addAll false",
			container:        test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("1M")).WithCPULimit(resource.MustParse("10")).WithMemLimit(resource.MustParse("10M")).Get(),
			vpa:              test.VerticalPodAutoscaler().WithContainer("container").WithTargetResource(corev1.ResourceCPU, "2").Get(),
			expectedCPU:      mustParseResourcePointer("2"),
			expectedCPULimit: mustParseResourcePointer("20"),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("2"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("20"),
					},
				},
			},
			addAll: false,
		},
		{
			name:             "Memory only recommendation, request and limits set, addAll false",
			container:        test.Container().WithName("container").WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("1M")).WithCPULimit(resource.MustParse("10")).WithMemLimit(resource.MustParse("10M")).Get(),
			vpa:              test.VerticalPodAutoscaler().WithContainer("container").WithTargetResource(corev1.ResourceMemory, "2M").Get(),
			expectedMem:      mustParseResourcePointer("2M"),
			expectedMemLimit: mustParseResourcePointer("20M"),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("20M"),
					},
				},
			},
			addAll: false,
		},
		{
			name:      "CPU and memory recommendation, request and limits only set in containerStatus, addAll true",
			container: test.Container().WithName("container").Get(),
			containerStatus: test.ContainerStatus().WithName("container").
				WithCPURequest(resource.MustParse("1")).
				WithMemRequest(resource.MustParse("1M")).
				WithCPULimit(resource.MustParse("10")).
				WithMemLimit(resource.MustParse("3M")).Get(),
			vpa: test.VerticalPodAutoscaler().WithContainer("container").WithTarget("3", "2M").Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("3"),
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30"),
						corev1.ResourceMemory: resource.MustParse("6M"),
					},
				},
			},
			addAll: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pod := test.Pod().WithName("pod").AddContainer(tc.container).AddContainerStatus(tc.containerStatus).Get()
			resources := GetContainersResources(pod, tc.vpa.Spec.ResourcePolicy, *tc.vpa.Status.Recommendation, nil, tc.addAll, utils.ContainerToAnnotationsMap{})
			compareRecommendations(t, resources, tc.expectedResources)
		})
	}
}

func TestGetContainersResources_VPAPodLevelResources(t *testing.T) {
	testCases := []struct {
		name                        string
		container                   corev1.Container
		containerStatus             corev1.ContainerStatus
		vpa                         *vpa_types.VerticalPodAutoscaler
		expectedResources           []vpa_api_util.ContainerResources
		addAll                      bool
		VPAPodLevelResourcesEnabled bool
	}{
		{
			name:      "feature flag turned off, RecommendationOnly mode should fallback to Auto",
			container: test.Container().WithName("container").Get(),
			containerStatus: test.ContainerStatus().WithName("container").
				WithCPURequest(resource.MustParse("1")).
				WithMemRequest(resource.MustParse("1M")).
				WithCPULimit(resource.MustParse("10")).
				WithMemLimit(resource.MustParse("3M")).Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer("container").
				AppendRecommendation(
					test.Recommendation().
						WithContainer("container").
						WithTarget("3", "2M").
						GetContainerResources()).
				WithScalingMode("container", vpa_types.ContainerScalingModeRecsOnly).
				Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("3"),
						corev1.ResourceMemory: resource.MustParse("2M"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30"),
						corev1.ResourceMemory: resource.MustParse("6M"),
					},
				},
			},
			VPAPodLevelResourcesEnabled: false,
		},
		{
			name:      "feature flag turned on, RecommendationOnly mode should be ignored",
			container: test.Container().WithName("container").Get(),
			containerStatus: test.ContainerStatus().WithName("container").
				WithCPURequest(resource.MustParse("1")).
				WithMemRequest(resource.MustParse("1M")).
				WithCPULimit(resource.MustParse("10")).
				WithMemLimit(resource.MustParse("3M")).Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer("container").
				AppendRecommendation(
					test.Recommendation().
						WithContainer("container").
						WithTarget("3", "2M").
						GetContainerResources()).
				WithScalingMode("container", vpa_types.ContainerScalingModeRecsOnly).
				Get(),
			expectedResources: []vpa_api_util.ContainerResources{
				{
					Requests: nil,
					Limits:   nil,
				},
			},
			VPAPodLevelResourcesEnabled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.VPAPodLevelResourcesEnabled {
				featuregatetesting.SetFeatureGateEmulationVersionDuringTest(t, features.MutableFeatureGate, version.MustParse("1.6"))
			}
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.VPAPodLevelResources, tc.VPAPodLevelResourcesEnabled)

			pod := test.Pod().WithName("pod").AddContainer(tc.container).AddContainerStatus(tc.containerStatus).Get()
			resources := GetContainersResources(pod, tc.vpa.Spec.ResourcePolicy, *tc.vpa.Status.Recommendation, nil, tc.addAll, utils.ContainerToAnnotationsMap{})
			compareRecommendations(t, resources, tc.expectedResources)
		})
	}
}

func compareRecommendations(t *testing.T, got, want []vpa_api_util.ContainerResources) {
	// If the overall slice shape differs, fail fast
	if !assert.Len(t, got, len(want), "container resources length mismatch") {
		t.Logf("got:\n%s", dump.Pretty(got))
		t.Logf("want:\n%s", dump.Pretty(want))
		return
	}

	for i := range want {
		gotReq := got[i].Requests
		wantReq := want[i].Requests
		gotLim := got[i].Limits
		wantLim := want[i].Limits

		if len(wantReq) == 0 || wantReq.Cpu().IsZero() {
			_, present := gotReq[corev1.ResourceCPU]
			assert.False(t, present, "expected no cpu request")
		} else {
			q, present := gotReq[corev1.ResourceCPU]
			if assert.True(t, present, "expected cpu request, but it's missing") {
				assert.Equal(t, wantReq.Cpu().MilliValue(), q.MilliValue(), "cpu request doesn't match")
			}
		}

		if len(wantReq) == 0 || wantReq.Memory().IsZero() {
			_, present := gotReq[corev1.ResourceMemory]
			assert.False(t, present, "expected no memory request")
		} else {
			q, present := gotReq[corev1.ResourceMemory]
			if assert.True(t, present, "expected memory request, but it's missing") {
				assert.Equal(t, wantReq.Memory().MilliValue(), q.MilliValue(), "memory request doesn't match")
			}
		}

		if len(wantLim) == 0 || wantLim.Cpu().IsZero() {
			_, present := gotLim[corev1.ResourceCPU]
			assert.False(t, present, "expected no cpu limit")
		} else {
			q, present := gotLim[corev1.ResourceCPU]
			if assert.True(t, present, "expected cpu limit, but it's missing") {
				assert.Equal(t, wantLim.Cpu().MilliValue(), q.MilliValue(), "cpu limit doesn't match")
			}
		}

		if len(wantLim) == 0 || wantLim.Memory().IsZero() {
			_, present := gotLim[corev1.ResourceMemory]
			assert.False(t, present, "expected no memory limit")
		} else {
			q, present := gotLim[corev1.ResourceMemory]
			if assert.True(t, present, "expected memory limit, but it's missing") {
				assert.Equal(t, wantLim.Memory().MilliValue(), q.MilliValue(), "memory limit doesn't match")
			}
		}
	}
}

func TestGetPodResources(t *testing.T) {
	c1 := "container1"
	c2 := "container2"

	testCases := []struct {
		name             string
		pod              *corev1.Pod
		vpa              *vpa_types.VerticalPodAutoscaler
		expectedCPU      resource.Quantity
		expectedMem      resource.Quantity
		expectedCPULimit resource.Quantity
		expectedMemLimit resource.Quantity
	}{
		{
			name: "no pod level target is set no requests or limits are returned",
			// Pod-level resources with a 1:2 request-to-limit ratio for both CPU and memory
			pod: test.Pod().
				WithCPURequest(resource.MustParse("10m")).
				WithMemRequest(resource.MustParse("12Mi")).
				WithCPULimit(resource.MustParse("20m")).
				WithMemLimit(resource.MustParse("24Mi")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer(c1).
				WithContainer(c2).
				Get(),
		},
		{
			name: "should return cpu targets and limits according to the request to limit ratio",
			// Pod-level resources with a 1:2 request-to-limit ratio for both CPU and memory
			pod: test.Pod().
				WithCPURequest(resource.MustParse("10m")).
				WithMemRequest(resource.MustParse("12Mi")).
				WithCPULimit(resource.MustParse("20m")).
				WithMemLimit(resource.MustParse("24Mi")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer(c1).
				WithContainer(c2).
				WithPodLevelTarget("155m", ""). // memory target is omitted
				Get(),
			expectedCPU:      resource.MustParse("155m"),
			expectedCPULimit: *resource.NewMilliQuantity(310, resource.DecimalSI),
		},
		{
			name: "should return both cpu and memory targets and limits according to the request to limit ratio",
			// Pod-level resources with a 1:2 request-to-limit ratio for both CPU and memory
			pod: test.Pod().
				WithCPURequest(resource.MustParse("10m")).
				WithMemRequest(resource.MustParse("12Mi")).
				WithCPULimit(resource.MustParse("20m")).
				WithMemLimit(resource.MustParse("24Mi")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer(c1).
				WithContainer(c2).
				WithPodLevelTarget("155m", "155Mi").
				Get(),
			expectedCPU:      resource.MustParse("155m"),
			expectedMem:      resource.MustParse("155Mi"),
			expectedCPULimit: *resource.NewMilliQuantity(310, resource.DecimalSI),
			expectedMemLimit: *resource.NewQuantity(325058560, resource.BinarySI), // 310Mi
		},
		{
			name: "RequestsOnly is set return the requests and limits from the pod spec",
			// Pod-level resources with a 1:2 request-to-limit ratio for both CPU and memory
			pod: test.Pod().
				WithCPURequest(resource.MustParse("10m")).
				WithMemRequest(resource.MustParse("12Mi")).
				WithCPULimit(resource.MustParse("20m")).
				WithMemLimit(resource.MustParse("24Mi")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer(c1).
				WithContainer(c2).
				WithPodLevelTarget("11m", "11Mi").
				WithPodLevelControlledValues(vpa_types.ContainerControlledValuesRequestsOnly).
				Get(),
			expectedCPU:      resource.MustParse("11m"),
			expectedMem:      resource.MustParse("11Mi"),
			expectedCPULimit: resource.MustParse("20m"),
			expectedMemLimit: *resource.NewQuantity(25165824, resource.BinarySI), // 24Mi
		},
		{
			name: "RequestsOnly is set return only the requests",
			// Pod-level resources with a 1:2 request-to-limit ratio for both CPU and memory
			pod: test.Pod().
				WithCPURequest(resource.MustParse("10m")).
				WithMemRequest(resource.MustParse("12Mi")).
				Get(),
			vpa: test.VerticalPodAutoscaler().
				WithContainer(c1).
				WithContainer(c2).
				WithPodLevelTarget("11m", "11Mi").
				WithPodLevelControlledValues(vpa_types.ContainerControlledValuesRequestsOnly).
				Get(),
			expectedCPU: resource.MustParse("11m"),
			expectedMem: resource.MustParse("11Mi"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resources := GetPodResources(tc.pod, tc.vpa.Spec.ResourcePolicy, *tc.vpa.Status.Recommendation)

			cpu, cpuPresent := resources.Requests[corev1.ResourceCPU]
			if tc.expectedCPU.IsZero() {
				assert.False(t, cpuPresent, "expected no cpu request, got %s", cpu.String())
			} else {
				assert.True(t, cpuPresent, "expected cpu request, but it's missing")
				assert.Equal(t, tc.expectedCPU, cpu, "cpu request doesn't match")
			}

			mem, memPresent := resources.Requests[corev1.ResourceMemory]
			if tc.expectedMem.IsZero() {
				assert.False(t, memPresent, "expected no memory request, got %s", mem.String())
			} else {
				assert.True(t, memPresent, "expected memory request, but it's missing")
				assert.Equal(t, tc.expectedMem, mem, "memory request doesn't match")
			}

			cpuLimit, cpuLimitPresent := resources.Limits[corev1.ResourceCPU]
			if tc.expectedCPULimit.IsZero() {
				assert.False(t, cpuLimitPresent, "expected no cpu limit, got %s", cpuLimit.String())
			} else {
				assert.True(t, cpuLimitPresent, "expected cpu limit, but it's missing")
				assert.Equal(t, tc.expectedCPULimit, cpuLimit, "cpu limit doesn't match")
			}

			memLimit, memLimitPresent := resources.Limits[corev1.ResourceMemory]
			if tc.expectedMemLimit.IsZero() {
				assert.False(t, memLimitPresent, "expected no memory limit, got %s", memLimit.String())
			} else {
				assert.True(t, memLimitPresent, "expected memory limit, but it's missing")
				assert.Equal(t, tc.expectedMemLimit, memLimit, "memory limit doesn't match")
			}
		})
	}
}
