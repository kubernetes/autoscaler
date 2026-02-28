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

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limits"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestApplyWithNilVPA(t *testing.T) {
	pod := test.Pod().WithName("pod1").AddContainer(test.Container().WithName("ctr-name").Get()).Get()
	processor := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{})

	res, annotations, err := processor.Apply(nil, pod)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Nil(t, annotations)
}
func TestApplyWithNilPod(t *testing.T) {
	vpa := test.VerticalPodAutoscaler().WithContainer("container").Get()
	processor := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{})

	res, annotations, err := processor.Apply(vpa, nil)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Nil(t, annotations)
}

func TestRecommendationNotAvailable(t *testing.T) {
	pod := test.Pod().WithName("pod1").AddContainer(test.Container().WithName("ctr-name").Get()).Get()

	containerName := "ctr-name-other"
	vpa := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		AppendRecommendation(
			test.Recommendation().
				WithContainer(containerName).
				WithTarget("100m", "50000Mi").
				GetContainerResources()).
		Get()

	res, annotations, err := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{}).Apply(vpa, pod)
	assert.Nil(t, err)
	assert.Empty(t, annotations)
	assert.Empty(t, res.ContainerRecommendations)
}

func TestRecommendationToLimitCapping(t *testing.T) {
	containerName := "ctr-name"
	pod := test.Pod().WithName("pod1").AddContainer(test.Container().WithName(containerName).Get()).Get()
	pod.Spec.Containers[0].Resources.Limits =
		apiv1.ResourceList{
			apiv1.ResourceCPU:    *resource.NewScaledQuantity(3, 1),
			apiv1.ResourceMemory: *resource.NewScaledQuantity(7000, 1),
		}
	podRecommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "ctr-name",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(8000, 1),
				},
				UpperBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(9000, 1),
				},
			},
		},
	}

	vpa := test.VerticalPodAutoscaler().WithContainer(containerName).Get()
	vpa.Status.Recommendation = &podRecommendation

	requestsAndLimits := vpa_types.ContainerControlledValuesRequestsAndLimits
	requestsOnly := vpa_types.ContainerControlledValuesRequestsOnly
	for _, tc := range []struct {
		name               string
		pod                *apiv1.Pod
		policy             vpa_types.PodResourcePolicy
		expectedTarget     apiv1.ResourceList
		expectedUpperBound apiv1.ResourceList
		expectedAnnotation bool
	}{
		{
			name:   "no capping for default policy",
			pod:    pod,
			policy: vpa_types.PodResourcePolicy{},
			expectedTarget: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(8000, 1),
			},
			expectedUpperBound: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(9000, 1),
			},
		}, {
			name: "no capping for RequestsAndLimits policy",
			pod:  pod,
			policy: vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestsAndLimits,
				}},
			},
			expectedTarget: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(8000, 1),
			},
			expectedUpperBound: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(9000, 1),
			},
		}, {
			name: "capping for RequestsOnly policy",
			pod:  pod,
			policy: vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestsOnly,
				}},
			},
			expectedTarget: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(7000, 1),
			},
			expectedUpperBound: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewScaledQuantity(3, 1),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(7000, 1),
			},
			expectedAnnotation: true,
		}, {
			name: "capping for RequestsOnly policy for limits defined in containerStatus",
			pod: func() *apiv1.Pod {
				pod := test.Pod().WithName("pod1").AddContainer(
					test.Container().WithName(containerName).
						WithCPULimit(*resource.NewScaledQuantity(3, 1)).
						WithMemLimit(*resource.NewScaledQuantity(7000, 1)).Get()).Get()
				pod.Status.ContainerStatuses = []apiv1.ContainerStatus{
					test.ContainerStatus().WithName(containerName).
						WithCPULimit(resource.MustParse("2.5")).
						WithMemLimit(*resource.NewScaledQuantity(6000, 1)).Get()}
				return pod
			}(),
			policy: vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestsOnly,
				}},
			},
			expectedTarget: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2.5"),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(6000, 1),
			},
			expectedUpperBound: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2.5"),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(6000, 1),
			},
			expectedAnnotation: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vpa.Spec.ResourcePolicy = &tc.policy
			res, annotations, err := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{}).Apply(vpa, tc.pod)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedTarget, res.ContainerRecommendations[0].Target)

			if tc.expectedAnnotation {
				assert.Contains(t, annotations, "ctr-name")
				assert.Contains(t, annotations["ctr-name"], "memory capped to container limit")
			} else {
				assert.NotContains(t, annotations, "ctr-name")
			}

			assert.Equal(t, tc.expectedUpperBound, res.ContainerRecommendations[0].UpperBound)
		})
	}
}

func TestRecommendationCappedToMinMaxPolicy(t *testing.T) {
	pod := test.Pod().WithName("pod1").AddContainer(test.Container().WithName("ctr-name").Get()).Get()
	podRecommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "ctr-name",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(30, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(5000, 1),
				},
				LowerBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(20, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(4300, 1),
				},
				UpperBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(50, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(5500, 1),
				},
			},
		},
	}
	policy := vpa_types.PodResourcePolicy{
		ContainerPolicies: []vpa_types.ContainerResourcePolicy{
			{
				ContainerName: "ctr-name",
				MinAllowed: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(40, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(4000, 1),
				},
				MaxAllowed: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(45, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(4500, 1),
				},
			},
		},
	}

	containerName := "ctr-name"
	vpa := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		Get()
	vpa.Spec.ResourcePolicy = &policy
	vpa.Status.Recommendation = &podRecommendation

	res, annotations, err := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{}).Apply(vpa, pod)
	assert.Nil(t, err)
	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(40, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(4500, 1),
	}, res.ContainerRecommendations[0].Target)

	assert.Contains(t, annotations, "ctr-name")
	assert.Contains(t, annotations["ctr-name"], "cpu capped to minAllowed")
	assert.Contains(t, annotations["ctr-name"], "memory capped to maxAllowed")

	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(40, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(4300, 1),
	}, res.ContainerRecommendations[0].LowerBound)

	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(45, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(4500, 1),
	}, res.ContainerRecommendations[0].UpperBound)
}

var containerRecommendations1 *vpa_types.RecommendedPodResources = &vpa_types.RecommendedPodResources{
	ContainerRecommendations: []vpa_types.RecommendedContainerResources{
		{
			ContainerName: "ctr-name",
			Target: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewScaledQuantity(5, 1),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(10, 1)},
			LowerBound: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewScaledQuantity(50, 1),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(100, 1)},
			UpperBound: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewScaledQuantity(150, 1),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(200, 1)},
		},
	},
}
var applyTestCases = []struct {
	Recommendations           *vpa_types.RecommendedPodResources
	Policy                    *vpa_types.PodResourcePolicy
	ExpectedPodRecommendation *vpa_types.RecommendedPodResources
	ExpectedError             error
}{
	{
		Recommendations:           nil,
		Policy:                    nil,
		ExpectedPodRecommendation: nil,
		ExpectedError:             nil,
	},
	{
		Recommendations:           containerRecommendations1,
		Policy:                    nil,
		ExpectedPodRecommendation: containerRecommendations1,
		ExpectedError:             nil,
	},
}

func TestApply(t *testing.T) {
	containerName := "ctr-name"
	pod := test.Pod().WithName("pod1").AddContainer(test.Container().WithName(containerName).Get()).Get()

	for _, testCase := range applyTestCases {
		vpa := test.VerticalPodAutoscaler().WithContainer(containerName).Get()
		vpa.Spec.ResourcePolicy = testCase.Policy
		vpa.Status.Recommendation = testCase.Recommendations
		res, _, err := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{}).Apply(
			vpa, pod)
		assert.Equal(t, testCase.ExpectedPodRecommendation, res)
		assert.Equal(t, testCase.ExpectedError, err)
	}
}

var (
	containerRecommendations2 = &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "foo",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("42m"),
					apiv1.ResourceMemory: resource.MustParse("42Mi"),
				},
				LowerBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("31m"),
					apiv1.ResourceMemory: resource.MustParse("31Mi"),
				},
				UpperBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("53m"),
					apiv1.ResourceMemory: resource.MustParse("53Mi"),
				},
				UncappedTarget: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("42m"),
					apiv1.ResourceMemory: resource.MustParse("42Mi"),
				},
			},
		},
	}
)

func TestApplyVPAPolicy(t *testing.T) {
	tests := []struct {
		Name             string
		Recommendations  *vpa_types.RecommendedPodResources
		ResourcePolicy   *vpa_types.PodResourcePolicy
		GlobalMaxAllowed limits.GlobalMaxAllowed
		Expected         *vpa_types.RecommendedPodResources
	}{
		{
			Name:             "recommendation is nil",
			Recommendations:  nil,
			ResourcePolicy:   nil,
			GlobalMaxAllowed: limits.GlobalMaxAllowed{},
			Expected:         nil,
		},
		{
			Name:             "resource policy is nil and global max allowed is nil",
			Recommendations:  containerRecommendations2,
			ResourcePolicy:   nil,
			GlobalMaxAllowed: limits.GlobalMaxAllowed{},
			Expected:         containerRecommendations2.DeepCopy(),
		},
		{
			Name:            "resource policy has min allowed and global max allowed is nil",
			Recommendations: containerRecommendations2,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "foo",
						MinAllowed: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("50m"),
							apiv1.ResourceMemory: resource.MustParse("50Mi"),
						},
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("50m"),
							apiv1.ResourceMemory: resource.MustParse("50Mi"),
						},
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("50m"),
							apiv1.ResourceMemory: resource.MustParse("50Mi"),
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("53m"),
							apiv1.ResourceMemory: resource.MustParse("53Mi"),
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),
							apiv1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
		{
			Name:            "resource policy has max allowed and global max allowed is nil",
			Recommendations: containerRecommendations2,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "foo",
						MaxAllowed: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("40Mi"),
						},
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("40Mi"),
						},
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("31m"),
							apiv1.ResourceMemory: resource.MustParse("31Mi"),
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("40Mi"),
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),
							apiv1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
		{
			Name:            "resource policy is nil and global max allowed is set for cpu and memory",
			Recommendations: containerRecommendations2,
			ResourcePolicy:  nil,
			GlobalMaxAllowed: limits.GlobalMaxAllowed{
				Container: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("40m"),
					apiv1.ResourceMemory: resource.MustParse("40Mi"),
				},
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("40Mi"),
						},
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("31m"),
							apiv1.ResourceMemory: resource.MustParse("31Mi"),
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("40Mi"),
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),
							apiv1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
		{
			Name:            "resource policy has minAllowed and global max allowed is set",
			Recommendations: containerRecommendations2,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "foo",
						MinAllowed: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("35m"),
							apiv1.ResourceMemory: resource.MustParse("35Mi"),
						},
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{
				Container: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("50m"),
					apiv1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),
							apiv1.ResourceMemory: resource.MustParse("42Mi"),
						},
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("35m"),
							apiv1.ResourceMemory: resource.MustParse("35Mi"),
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("50m"),
							apiv1.ResourceMemory: resource.MustParse("50Mi"),
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),
							apiv1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
		{
			Name:            "resource policy has max allowed for cpu and global max allowed is set for memory",
			Recommendations: containerRecommendations2,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "foo",
						MaxAllowed: apiv1.ResourceList{
							apiv1.ResourceCPU: resource.MustParse("40m"),
						},
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{
				Container: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("40Mi"),
				},
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("40Mi"),
						},
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("31m"),
							apiv1.ResourceMemory: resource.MustParse("31Mi"),
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("40Mi"),
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),
							apiv1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
		{
			Name:            "resource policy has max allowed for cpu and global max allowed is set for cpu and memory",
			Recommendations: containerRecommendations2,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "foo",
						MaxAllowed: apiv1.ResourceList{
							apiv1.ResourceCPU: resource.MustParse("40m"),
						},
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{
				Container: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("30m"),
					apiv1.ResourceMemory: resource.MustParse("30Mi"),
				},
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("30Mi"),
						},
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("31m"),
							apiv1.ResourceMemory: resource.MustParse("30Mi"),
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("30Mi"),
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),
							apiv1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			resourcePolicyCopy := tt.ResourcePolicy.DeepCopy()

			actual, err := ApplyVPAPolicy(false, tt.Recommendations, tt.ResourcePolicy, tt.GlobalMaxAllowed)
			assert.Nil(t, err)
			assert.Equal(t, tt.Expected, actual)

			// Make sure that the func does not have a side affect and does not modify the passed resource policy.
			assert.Equal(t, resourcePolicyCopy, tt.ResourcePolicy)
		})
	}
}

var (
	containerRecommendations3 = &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "c1",
				LowerBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("10m"),
					apiv1.ResourceMemory: resource.MustParse("10Mi"),
				},
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("40m"),
					apiv1.ResourceMemory: resource.MustParse("40Mi"),
				},
				UpperBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("50m"),
					apiv1.ResourceMemory: resource.MustParse("50Mi"),
				},
				UncappedTarget: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("40m"),
					apiv1.ResourceMemory: resource.MustParse("40Mi"),
				},
			},
		},
	}
	// sum of container recommendations:
	// LC: 22m, LM: 22Mi
	// TC: 82m, TM: 82Mi
	// UC: 102m, UM: 102Mi
	containerRecommendations4 = &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "c1",
				LowerBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("10m"),
					apiv1.ResourceMemory: resource.MustParse("10Mi"),
				},
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("40m"),
					apiv1.ResourceMemory: resource.MustParse("40Mi"),
				},
				UpperBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("50m"),
					apiv1.ResourceMemory: resource.MustParse("50Mi"),
				},
				UncappedTarget: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("40m"),
					apiv1.ResourceMemory: resource.MustParse("40Mi"),
				},
			},
			{
				ContainerName: "c2",
				LowerBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("12m"),
					apiv1.ResourceMemory: resource.MustParse("12Mi"),
				},
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("42m"),
					apiv1.ResourceMemory: resource.MustParse("42Mi"),
				},
				UpperBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("52m"),
					apiv1.ResourceMemory: resource.MustParse("52Mi"),
				},
				UncappedTarget: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("42m"),
					apiv1.ResourceMemory: resource.MustParse("42Mi"),
				},
			},
		},
	}
)

func TestApplyVPAPolicyWithPodLevelTrue(t *testing.T) {
	tests := []struct {
		Name             string
		Recommendations  *vpa_types.RecommendedPodResources
		ResourcePolicy   *vpa_types.PodResourcePolicy
		GlobalMaxAllowed limits.GlobalMaxAllowed
		Expected         *vpa_types.RecommendedPodResources
	}{
		{
			Name:             "recommendation is nil",
			Recommendations:  nil,
			ResourcePolicy:   nil,
			GlobalMaxAllowed: limits.GlobalMaxAllowed{},
			Expected:         nil,
		},
		{
			Name:             "resource policy is nil and global max allowed is nil",
			Recommendations:  containerRecommendations3,
			ResourcePolicy:   nil,
			GlobalMaxAllowed: limits.GlobalMaxAllowed{},
			Expected:         containerRecommendations3.DeepCopy(),
		},
		{
			Name:            "pod resource policy has min allowed and pod global max allowed is nil",
			Recommendations: containerRecommendations3,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				PodPolicies: &vpa_types.PodResourcePolicies{
					MinAllowed: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("49m"),
						apiv1.ResourceMemory: resource.MustParse("49Mi"),
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "c1",
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(49, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(51380224, resource.BinarySI), // 49Mi
						},
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(49, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(51380224, resource.BinarySI), // 49Mi
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("50m"),  // does not violate Pod-level MinAllowed
							apiv1.ResourceMemory: resource.MustParse("50Mi"), // does not violate Pod-level MinAllowed
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),  // unchanged because we do not cap uncappedTarget
							apiv1.ResourceMemory: resource.MustParse("40Mi"), // unchanged because we do not cap uncappedTarget
						},
					},
				},
			},
		},
		{
			Name:            "pod resource policy has max allowed and pod global max allowed is nil",
			Recommendations: containerRecommendations3,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				PodPolicies: &vpa_types.PodResourcePolicies{
					MaxAllowed: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("39m"),
						apiv1.ResourceMemory: resource.MustParse("39Mi"),
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "c1",
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("10m"),  // does not violate Pod-level MaxAllowed
							apiv1.ResourceMemory: resource.MustParse("10Mi"), // does not violate Pod-level MaxAllowed
						},
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(39, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(40894464, resource.BinarySI), // 39Mi
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(39, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(40894464, resource.BinarySI), // 39Mi
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),  // unchanged because we do not cap uncappedTarget
							apiv1.ResourceMemory: resource.MustParse("40Mi"), // unchanged because we do not cap uncappedTarget
						},
					},
				},
			},
		},
		{
			Name:            "pod resource policy is nil and pod global max allowed is set for cpu and memory",
			Recommendations: containerRecommendations3,
			ResourcePolicy:  nil,
			GlobalMaxAllowed: limits.GlobalMaxAllowed{
				Pod: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("30m"),
					apiv1.ResourceMemory: resource.MustParse("30Mi"),
				},
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "c1",
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("10m"),
							apiv1.ResourceMemory: resource.MustParse("10Mi"),
						},
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(30, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(31457280, resource.BinarySI), // 30Mi
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(30, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(31457280, resource.BinarySI), // 30Mi
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),  // unchanged because we do not cap uncappedTarget
							apiv1.ResourceMemory: resource.MustParse("40Mi"), // unchanged because we do not cap uncappedTarget
						},
					},
				},
			},
		},
		{
			Name:            "pod resource policy has minAllowed and pod global max allowed is set",
			Recommendations: containerRecommendations3,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				PodPolicies: &vpa_types.PodResourcePolicies{
					MinAllowed: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("21m"),
						apiv1.ResourceMemory: resource.MustParse("21Mi"),
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{
				Pod: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("49m"),
					apiv1.ResourceMemory: resource.MustParse("49Mi"),
				},
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "c1",
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(21, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(22020096, resource.BinarySI), // 21Mi
						},
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("40Mi"),
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(49, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(51380224, resource.BinarySI), // 49 Mi
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: resource.MustParse("40Mi"),
						},
					},
				},
			},
		},
		{
			Name:            "pod resource policy has max allowed for cpu and pod global max allowed is set for memory",
			Recommendations: containerRecommendations3,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				PodPolicies: &vpa_types.PodResourcePolicies{
					MaxAllowed: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"),
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{
				Pod: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("33Mi"),
				},
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "c1",
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("10m"),
							apiv1.ResourceMemory: resource.MustParse("10Mi"),
						},
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(22, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(34603008, resource.BinarySI), // 33Mi
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(22, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(34603008, resource.BinarySI), // 33Mi
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),  // unchanged because we do not cap uncappedTarget
							apiv1.ResourceMemory: resource.MustParse("40Mi"), // unchanged because we do not cap uncappedTarget
						},
					},
				},
			},
		},
		{
			Name:            "pod resource policy has max allowed for cpu and pod global max allowed is set for cpu and memory",
			Recommendations: containerRecommendations3,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				PodPolicies: &vpa_types.PodResourcePolicies{
					MaxAllowed: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("44m"), // takes precedence
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{
				Pod: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("33m"),
					apiv1.ResourceMemory: resource.MustParse("33Mi"),
				},
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "c1",
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("10m"),
							apiv1.ResourceMemory: resource.MustParse("10Mi"),
						},
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),
							apiv1.ResourceMemory: *resource.NewQuantity(34603008, resource.BinarySI), // 33Mi
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(44, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(34603008, resource.BinarySI), // 33Mi
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),  // unchanged because we do not cap uncappedTarget
							apiv1.ResourceMemory: resource.MustParse("40Mi"), // unchanged because we do not cap uncappedTarget
						},
					},
				},
			},
		},
		{
			Name:            "pod resource policy and container resource policy are set and pod global max allowed is nil",
			Recommendations: containerRecommendations4,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				PodPolicies: &vpa_types.PodResourcePolicies{
					MaxAllowed: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("66Mi"),
					},
				},
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "c1",
						MinAllowed: apiv1.ResourceList{
							apiv1.ResourceCPU: resource.MustParse("41m"),
						},
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "c1",
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("41m"),
							apiv1.ResourceMemory: resource.MustParse("10Mi"),
						},
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("41m"),
							apiv1.ResourceMemory: *resource.NewQuantity(33759032, resource.BinarySI), //(40*66)/82
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("50m"),
							apiv1.ResourceMemory: *resource.NewQuantity(33759032, resource.BinarySI), //(50*66)/102
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),  // unchanged because we do not cap uncappedTarget
							apiv1.ResourceMemory: resource.MustParse("40Mi"), // unchanged because we do not cap uncappedTarget
						},
					},
					{
						ContainerName: "c2",
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("12m"),
							apiv1.ResourceMemory: resource.MustParse("12Mi"),
						},
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),
							apiv1.ResourceMemory: *resource.NewQuantity(35446983, resource.BinarySI), //(42*66)/82
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("52m"),
							apiv1.ResourceMemory: *resource.NewQuantity(35446983, resource.BinarySI), //(52*66)/102
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),  // unchanged because we do not cap uncappedTarget
							apiv1.ResourceMemory: resource.MustParse("42Mi"), // unchanged because we do not cap uncappedTarget
						},
					},
				},
			},
		},
		{
			Name:            "pod resource policy and container resource policy are set and pod global max allowed is set",
			Recommendations: containerRecommendations4,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				PodPolicies: &vpa_types.PodResourcePolicies{
					MaxAllowed: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("66Mi"),
					},
				},
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "c1",
						MinAllowed: apiv1.ResourceList{
							apiv1.ResourceCPU: resource.MustParse("41m"),
						},
					},
				},
			},
			GlobalMaxAllowed: limits.GlobalMaxAllowed{
				Pod: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("1Mi"), // should be ignored
				},
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "c1",
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("41m"),
							apiv1.ResourceMemory: resource.MustParse("10Mi"),
						},
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("41m"),
							apiv1.ResourceMemory: *resource.NewQuantity(33759032, resource.BinarySI), //(40*66)/82
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("50m"),
							apiv1.ResourceMemory: *resource.NewQuantity(33759032, resource.BinarySI), //(50*66)/102
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("40m"),  // unchanged because we do not cap uncappedTarget
							apiv1.ResourceMemory: resource.MustParse("40Mi"), // unchanged because we do not cap uncappedTarget
						},
					},
					{
						ContainerName: "c2",
						LowerBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("12m"),
							apiv1.ResourceMemory: resource.MustParse("12Mi"),
						},
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),
							apiv1.ResourceMemory: *resource.NewQuantity(35446983, resource.BinarySI), //(42*66)/82
						},
						UpperBound: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("52m"),
							apiv1.ResourceMemory: *resource.NewQuantity(35446983, resource.BinarySI), //(52*66)/102
						},
						UncappedTarget: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("42m"),  // unchanged because we do not cap uncappedTarget
							apiv1.ResourceMemory: resource.MustParse("42Mi"), // unchanged because we do not cap uncappedTarget
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			resourcePolicyCopy := tt.ResourcePolicy.DeepCopy()

			actual, err := ApplyVPAPolicy(true, tt.Recommendations, tt.ResourcePolicy, tt.GlobalMaxAllowed)
			assert.Nil(t, err)
			assert.Equal(t, tt.Expected, actual)

			// Make sure that the func does not have a side affect and does not modify the passed resource policy.
			assert.Equal(t, resourcePolicyCopy, tt.ResourcePolicy)
		})
	}
}

func TestEnsureValidBounds(t *testing.T) {
	tests := []struct {
		name               string
		containerLevelRecs []vpa_types.RecommendedContainerResources
		expected           []vpa_types.RecommendedContainerResources
	}{
		{
			name: "bounds not violated", // i.e. for each container, "LowerBound <= Target <= UpperBound" is true.
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("31m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("42m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("53m"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("11m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("11m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("11m"),
					},
				},
			},
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("31m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("42m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("53m"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("11m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("11m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("11m"),
					},
				},
			},
		},
		{
			name: "UpperBound in c2 is violated", // i.e. for c2 container, "Target <= UpperBound" is false.
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("42m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("53m"),
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("12m"),
					},
				},
			},
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("42m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(43, resource.DecimalSI), // -10m
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"), // UpperBound should be set to Target, i.e. 10 millicores should be taken from other UpperBounds
					},
				},
			},
		},
		{
			name: "LowerBound in c2 is violated", // i.e. for c2 container, "LowerBound <= Target" is false.
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("31m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("42m"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("24m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"),
					},
				},
			},
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(33, resource.DecimalSI), // +2m
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("42m"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"), // LowerBound should be set to Target, i.e. 2 milicores should be re-distributed
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"),
					},
				},
			},
		},
		{
			name: "LowerBound and UpperBound in c2 are violated", // i.e. for c2 container, "LowerBound <= Target <= UpperBound" is false.
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("31m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("42m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("53m"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("25m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("15m"),
					},
				},
			},
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(34, resource.DecimalSI), // +3m
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("42m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(46, resource.DecimalSI), // -7m
					},
				},
				{
					ContainerName: "c2",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"), // -3m
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"),
					},
					UpperBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("22m"), // +7m
					},
				},
			},
		},
		{
			name: "LowerBound in c2 is violated and freed milicores can not be redistributed", // i.e. for c2 container, "LowerBound <= Target" is false.
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("41m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("41m"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("25m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("23m"),
					},
				},
			},
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("41m"),
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("41m"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("23m"), // -2m
					},
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("23m"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ensureBoundsAreValid(tt.containerLevelRecs)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestProcessContainerLevelRecs(t *testing.T) {

	tests := []struct {
		name               string
		podLevelRec        apiv1.ResourceList
		containerLevelRecs []vpa_types.RecommendedContainerResources
		mins               apiv1.ResourceList
		maxs               apiv1.ResourceList
		expected           []vpa_types.RecommendedContainerResources
	}{
		{
			name: "no constrains return same recommendations",
			podLevelRec: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("20m"),
				apiv1.ResourceMemory: resource.MustParse("20Mi"),
			},
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("10m"),
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("30m"),
					},
				},
			},
			mins: nil,
			maxs: nil,
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("10m"),
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("30m"),
					},
				},
			},
		},
		{
			name: "Pod-level minAllowed is set and not violated",
			podLevelRec: apiv1.ResourceList{
				apiv1.ResourceCPU: resource.MustParse("40m"),
			},
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("10m"),
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("30m"),
					},
				},
			},
			mins: apiv1.ResourceList{
				apiv1.ResourceMemory: resource.MustParse("50Mi"),
			},
			maxs: nil,
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("10m"),
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("30m"),
					},
				},
			},
		},
		{
			name: "Pod-level minAllowed is set and violated.",
			podLevelRec: apiv1.ResourceList{
				apiv1.ResourceCPU: resource.MustParse("40m"),
			},
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("10m"),
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("30m"),
					},
				},
			},
			mins: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("50m"),
				apiv1.ResourceMemory: resource.MustParse("50Mi"),
			},
			maxs: nil,
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(12, resource.DecimalSI), // (10x50)/40
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(37, resource.DecimalSI), // (30x50)/40
					},
				},
			},
		},
		{
			name: "Pod-level maxAllowed is set and violated",
			podLevelRec: apiv1.ResourceList{
				apiv1.ResourceCPU: resource.MustParse("40m"),
			},
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("10m"),
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("30m"),
					},
				},
			},
			mins: nil,
			maxs: apiv1.ResourceList{
				apiv1.ResourceCPU: resource.MustParse("20m"),
			},
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(5, resource.DecimalSI), // (10x20)/40
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(15, resource.DecimalSI), // (30x20)/40
					},
				},
			},
		},
		{
			name: "Pod-level minAllowed and maxAllowed are set and not violated",
			podLevelRec: apiv1.ResourceList{
				apiv1.ResourceCPU: resource.MustParse("40m"),
			},
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("10m"),
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("30m"),
					},
				},
			},
			mins: apiv1.ResourceList{
				apiv1.ResourceCPU: resource.MustParse("1m"),
			},
			maxs: apiv1.ResourceList{
				apiv1.ResourceCPU: resource.MustParse("1000m"),
			},
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("10m"),
					},
				},
				{
					ContainerName: "c2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("30m"),
					},
				},
			},
		},
	}
	getTarget := func(rl vpa_types.RecommendedContainerResources) *apiv1.ResourceList { return &rl.Target }
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actual := processContainerLevelRecs(tt.podLevelRec, tt.containerLevelRecs, tt.mins, tt.maxs, getTarget)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

type fakeLimitRangeCalculator struct {
	containerLimitRange apiv1.LimitRangeItem
	podLimitRange       apiv1.LimitRangeItem
}

func (nlrc *fakeLimitRangeCalculator) GetContainerLimitRangeItem(namespace string) (*apiv1.LimitRangeItem, error) {
	return &nlrc.containerLimitRange, nil
}

func (nlrc *fakeLimitRangeCalculator) GetPodLimitRangeItem(namespace string) (*apiv1.LimitRangeItem, error) {
	return &nlrc.podLimitRange, nil
}

func TestApplyCapsToLimitRange(t *testing.T) {
	limitRange := apiv1.LimitRangeItem{
		Type: apiv1.LimitTypeContainer,
		Max: apiv1.ResourceList{
			apiv1.ResourceCPU: resource.MustParse("1"),
		},
		Min: apiv1.ResourceList{
			apiv1.ResourceMemory: resource.MustParse("500M"),
		},
	}

	containerName := "container"
	vpa := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		Get()

	recommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "container",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("2"),
					apiv1.ResourceMemory: resource.MustParse("200M"),
				},
			},
		},
	}
	vpa.Status.Recommendation = &recommendation

	pod := apiv1.Pod{
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name: "container",
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("1G"),
						},
						Limits: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("1G"),
						},
					},
				},
			},
		},
	}
	expectedRecommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "container",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("1000m"),
					apiv1.ResourceMemory: resource.MustParse("500000000"),
				},
			},
		},
	}

	calculator := fakeLimitRangeCalculator{containerLimitRange: limitRange}
	processor := NewCappingRecommendationProcessor(&calculator)
	processedRecommendation, annotations, err := processor.Apply(vpa, &pod)
	assert.NoError(t, err)
	assert.Contains(t, annotations, "container")
	assert.ElementsMatch(t, []string{"cpu capped to fit Max in container LimitRange", "memory capped to fit Min in container LimitRange"}, annotations["container"])
	assert.Equal(t, expectedRecommendation, *processedRecommendation)
}

func TestApplyPodLimitRange(t *testing.T) {
	tests := []struct {
		name         string
		resources    []vpa_types.RecommendedContainerResources
		pod          apiv1.Pod
		limitRange   apiv1.LimitRangeItem
		resourceName apiv1.ResourceName
		expect       []vpa_types.RecommendedContainerResources
	}{
		{
			name: "cap target cpu to max",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("1"),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("1"),
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container1",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU: resource.MustParse("1"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
						{
							Name: "container2",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU: resource.MustParse("1"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			limitRange: apiv1.LimitRangeItem{
				Max: apiv1.ResourceList{
					apiv1.ResourceCPU: resource.MustParse("1"),
				},
			},
			resourceName: apiv1.ResourceCPU,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(500, resource.DecimalSI),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(500, resource.DecimalSI),
					},
				},
			},
		},
		{
			name: "cap cpu to max",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("1"),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse("1"),
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container1",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU: resource.MustParse("1"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
						{
							Name: "container2",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU: resource.MustParse("1"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			limitRange: apiv1.LimitRangeItem{
				Max: apiv1.ResourceList{
					apiv1.ResourceCPU: resource.MustParse("1"),
				},
			},
			resourceName: apiv1.ResourceCPU,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(500, resource.DecimalSI),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceCPU: *resource.NewMilliQuantity(500, resource.DecimalSI),
					},
				},
			},
		},
		{
			name: "cap mem to min",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("1G"),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("1G"),
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container1",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("1"),
								},
							},
						},
						{
							Name: "container2",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			limitRange: apiv1.LimitRangeItem{
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("4G"),
				},
			},
			resourceName: apiv1.ResourceMemory,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("2000000000"),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("2000000000"),
					},
				},
			},
		},
		{
			name: "cap mem request to pod min",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("1G"),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("1G"),
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container1",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("2"),
								},
							},
						},
						{
							Name: "container2",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			limitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypePod,
				Max: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("10G"),
				},
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("4G"),
				},
			},
			resourceName: apiv1.ResourceMemory,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("2000000000"),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("2000000000"),
					},
				},
			},
		},
		{
			name: "cap mem request to pod min, only one container with recommendation",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("1G"),
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container1",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("1G"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("2G"),
								},
							},
						},
						{
							Name: "container2",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("1G"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("2G"),
								},
							},
						},
					},
				},
			},
			limitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypePod,
				Max: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("10G"),
				},
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("4G"),
				},
			},
			resourceName: apiv1.ResourceMemory,
			expect: []vpa_types.RecommendedContainerResources{
				{
					// TODO: This is incorrect; The pod will be rejected by limit range because sum of its
					// pod requests is too small - it's 3Gi(2Gi for `container1` (from recommendation) and 1Gi
					// for `container2` (unchanged incoming request)) and minimum is 4Gi.
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("2000000000"),
					},
				},
			},
		},
		{
			name: "cap mem request to pod min, extra recommendation",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("1G"),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("1G"),
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container2",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			limitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypePod,
				Max: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("10G"),
				},
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("4G"),
				},
			},
			resourceName: apiv1.ResourceMemory,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("1G"),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("4000000000"),
					},
				},
			},
		},
	}
	getTarget := func(rl vpa_types.RecommendedContainerResources) *apiv1.ResourceList { return &rl.Target }
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := applyPodLimitRange(tc.resources, &tc.pod, tc.limitRange, tc.resourceName, getTarget)
			assert.Equal(t, tc.expect, got)
		})
	}
}

func TestApplyLimitRangeMinToRequest(t *testing.T) {
	requestsOnly := vpa_types.ContainerControlledValuesRequestsOnly
	requestAndLimit := vpa_types.ContainerControlledValuesRequestsAndLimits
	tests := []struct {
		name                string
		resources           vpa_types.RecommendedPodResources
		pod                 apiv1.Pod
		containerLimitRange apiv1.LimitRangeItem
		podLimitRange       apiv1.LimitRangeItem
		policy              *vpa_types.PodResourcePolicy
		expect              vpa_types.RecommendedPodResources
		expectAnnotations   map[string][]string
	}{
		{
			name: "caps to min range if above container limit",
			resources: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("600M"),
								},
							},
						},
					},
				},
			},
			containerLimitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypeContainer,
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("500M"),
				},
			},
			expect: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("500M"),
						},
					},
				},
			},
			expectAnnotations: map[string][]string{
				"container": {"memory capped to fit Min in container LimitRange"},
			},
		},
		{
			name: "caps to container limit if below container limit",
			resources: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
				},
			},
			containerLimitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypeContainer,
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("500M"),
				},
			},
			policy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestsOnly,
				}},
			},
			expect: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("100M"),
						},
					},
				},
			},
			expectAnnotations: map[string][]string{
				"container": {
					"memory capped to fit Min in container LimitRange",
					"memory capped to container limit",
				},
			},
		},
		{
			name: "caps to pod limit if below pod limit",
			resources: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
				},
			},
			podLimitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypePod,
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("500M"),
				},
			},
			policy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestsOnly,
				}},
			},
			expect: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("100M"),
						},
					},
				},
			},
			expectAnnotations: map[string][]string{
				"container": {
					"memory capped to container limit",
				},
			},
		},
		{
			name: "Should use container status limits when lower than both recommendation and pod limit range",
			resources: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: *test.Pod().
				AddContainer(test.Container().WithName("container").
					// Requests and Limits defined in the pod spec are ignored because
					// the values in container status have higher priority.
					WithCPURequest(resource.MustParse("10")).
					WithMemRequest(resource.MustParse("1G")).
					WithCPULimit(resource.MustParse("10")).
					WithMemLimit(resource.MustParse("1G")).Get()).
				AddContainerStatus(test.ContainerStatus().WithName("container").
					WithCPURequest(resource.MustParse("5m")).
					WithMemRequest(resource.MustParse("50M")).
					WithCPULimit(resource.MustParse("5m")).
					WithMemLimit(resource.MustParse("100M")).Get()).Get(),
			podLimitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypePod,
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("500M"),
					apiv1.ResourceCPU:    resource.MustParse("2"),
				},
			},
			policy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestsOnly,
				}},
			},
			expect: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("5m"),
							apiv1.ResourceMemory: resource.MustParse("100M"),
						},
					},
				},
			},
			expectAnnotations: map[string][]string{
				"container": {
					"memory capped to container limit",
					"cpu capped to container limit",
				},
			},
		},
		{
			name: "caps to pod limit if below pod limit one container with recommendation and one without",
			resources: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container1",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container1",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
						{
							Name: "container2",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
				},
			},
			podLimitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypePod,
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("500M"),
				},
			},
			policy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestAndLimit,
				}},
			},
			expect: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container1",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("400000000"),
						},
					},
					{
						ContainerName: "container2",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("100000000"),
						},
					},
				},
			},
			expectAnnotations: map[string][]string{},
		},
		{
			name: "caps to pod limit if below pod limit two containers with recommendation",
			resources: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container1",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("200M"),
						},
					},
					{
						ContainerName: "container2",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: apiv1.Pod{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "container1",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
						{
							Name: "container2",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("1"),
									apiv1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
				},
			},
			podLimitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypePod,
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("500M"),
				},
			},
			policy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestAndLimit,
				}},
			},
			expect: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container1",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("250000000"),
						},
					},
					{
						ContainerName: "container2",
						Target: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("250000000"),
						},
					},
				},
			},
			expectAnnotations: map[string][]string{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calculator := fakeLimitRangeCalculator{
				containerLimitRange: tc.containerLimitRange,
				podLimitRange:       tc.podLimitRange,
			}
			processor := NewCappingRecommendationProcessor(&calculator)

			containerName := "ctr-name"
			vpa := test.VerticalPodAutoscaler().WithContainer(containerName).Get()
			vpa.Spec.ResourcePolicy = tc.policy
			vpa.Status.Recommendation = &tc.resources

			processedRecommendation, annotations, err := processor.Apply(vpa, &tc.pod)
			assert.NoError(t, err)
			for containerName, expectedAnnotations := range tc.expectAnnotations {
				if assert.Contains(t, annotations, containerName) {
					assert.ElementsMatch(t, expectedAnnotations, annotations[containerName], "for container '%s'", containerName)
				}
			}
			assert.Equal(t, tc.expect, *processedRecommendation, "for container %s", containerName)
			for containerName := range annotations {
				assert.Contains(t, tc.expectAnnotations, containerName)
			}
		})
	}
}

func TestCapPodMemoryWithUnderByteSplit(t *testing.T) {
	// It's important that recommendations are different so straightforwardly
	// splitting Max between containers results in fractional bytes going to them.
	// We want to ensure we'll round millibytes the right way (rounding down when
	// caping to max, to stay below max, rounding up when caping to min).
	recommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "container-1",
				Target: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			{
				ContainerName: "container-2",
				Target: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
		},
	}
	pod := apiv1.Pod{
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name: "container-1",
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceMemory: resource.MustParse("50M"),
						},
						Limits: apiv1.ResourceList{
							apiv1.ResourceMemory: resource.MustParse("50M"),
						},
					},
				},
				{
					Name: "container-2",
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceMemory: resource.MustParse("50M"),
						},
						Limits: apiv1.ResourceList{
							apiv1.ResourceMemory: resource.MustParse("50M"),
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name                   string
		limitRange             apiv1.LimitRangeItem
		expectedRecommendation vpa_types.RecommendedPodResources
	}{
		{
			name: "cap to max",
			limitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypePod,
				Max: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			expectedRecommendation: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container-1",
						Target: apiv1.ResourceList{
							apiv1.ResourceMemory: *resource.NewQuantity(357913941, resource.BinarySI),
						},
					},
					{
						ContainerName: "container-2",
						Target: apiv1.ResourceList{
							apiv1.ResourceMemory: *resource.NewQuantity(715827882, resource.BinarySI),
						},
					},
				},
			},
		},
		{
			name: "cap to min",
			limitRange: apiv1.LimitRangeItem{
				Type: apiv1.LimitTypePod,
				Min: apiv1.ResourceList{
					apiv1.ResourceMemory: resource.MustParse("4Gi"),
				},
			},
			expectedRecommendation: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container-1",
						Target: apiv1.ResourceList{
							apiv1.ResourceMemory: *resource.NewQuantity(1431655765, resource.BinarySI),
						},
					},
					{
						ContainerName: "container-2",
						Target: apiv1.ResourceList{
							apiv1.ResourceMemory: *resource.NewQuantity(2863311530, resource.BinarySI),
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calculator := fakeLimitRangeCalculator{podLimitRange: tc.limitRange}
			processor := NewCappingRecommendationProcessor(&calculator)

			containerName := "ctr-name"
			vpa := test.VerticalPodAutoscaler().WithContainer(containerName).Get()
			vpa.Status.Recommendation = &recommendation

			processedRecommendation, _, err := processor.Apply(vpa, &pod)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedRecommendation, *processedRecommendation)
		})
	}
}
