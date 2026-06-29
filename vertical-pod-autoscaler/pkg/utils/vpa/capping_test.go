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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
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
		corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewScaledQuantity(3, 1),
			corev1.ResourceMemory: *resource.NewScaledQuantity(7000, 1),
		}
	podRecommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "ctr-name",
				Target: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
					corev1.ResourceMemory: *resource.NewScaledQuantity(8000, 1),
				},
				UpperBound: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
					corev1.ResourceMemory: *resource.NewScaledQuantity(9000, 1),
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
		pod                *corev1.Pod
		policy             vpa_types.PodResourcePolicy
		expectedTarget     corev1.ResourceList
		expectedUpperBound corev1.ResourceList
		expectedAnnotation bool
	}{
		{
			name:   "no capping for default policy",
			pod:    pod,
			policy: vpa_types.PodResourcePolicy{},
			expectedTarget: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
				corev1.ResourceMemory: *resource.NewScaledQuantity(8000, 1),
			},
			expectedUpperBound: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
				corev1.ResourceMemory: *resource.NewScaledQuantity(9000, 1),
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
			expectedTarget: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
				corev1.ResourceMemory: *resource.NewScaledQuantity(8000, 1),
			},
			expectedUpperBound: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
				corev1.ResourceMemory: *resource.NewScaledQuantity(9000, 1),
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
			expectedTarget: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
				corev1.ResourceMemory: *resource.NewScaledQuantity(7000, 1),
			},
			expectedUpperBound: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(3, 1),
				corev1.ResourceMemory: *resource.NewScaledQuantity(7000, 1),
			},
			expectedAnnotation: true,
		}, {
			name: "capping for RequestsOnly policy for limits defined in containerStatus",
			pod: func() *corev1.Pod {
				pod := test.Pod().WithName("pod1").AddContainer(
					test.Container().WithName(containerName).
						WithCPULimit(*resource.NewScaledQuantity(3, 1)).
						WithMemLimit(*resource.NewScaledQuantity(7000, 1)).Get()).Get()
				pod.Status.ContainerStatuses = []corev1.ContainerStatus{
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
			expectedTarget: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2.5"),
				corev1.ResourceMemory: *resource.NewScaledQuantity(6000, 1),
			},
			expectedUpperBound: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2.5"),
				corev1.ResourceMemory: *resource.NewScaledQuantity(6000, 1),
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
				Target: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(30, 1),
					corev1.ResourceMemory: *resource.NewScaledQuantity(5000, 1),
				},
				LowerBound: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(20, 1),
					corev1.ResourceMemory: *resource.NewScaledQuantity(4300, 1),
				},
				UpperBound: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(50, 1),
					corev1.ResourceMemory: *resource.NewScaledQuantity(5500, 1),
				},
			},
		},
	}
	policy := vpa_types.PodResourcePolicy{
		ContainerPolicies: []vpa_types.ContainerResourcePolicy{
			{
				ContainerName: "ctr-name",
				MinAllowed: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(40, 1),
					corev1.ResourceMemory: *resource.NewScaledQuantity(4000, 1),
				},
				MaxAllowed: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewScaledQuantity(45, 1),
					corev1.ResourceMemory: *resource.NewScaledQuantity(4500, 1),
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
	assert.Equal(t, corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewScaledQuantity(40, 1),
		corev1.ResourceMemory: *resource.NewScaledQuantity(4500, 1),
	}, res.ContainerRecommendations[0].Target)

	assert.Contains(t, annotations, "ctr-name")
	assert.Contains(t, annotations["ctr-name"], "cpu capped to minAllowed")
	assert.Contains(t, annotations["ctr-name"], "memory capped to maxAllowed")

	assert.Equal(t, corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewScaledQuantity(40, 1),
		corev1.ResourceMemory: *resource.NewScaledQuantity(4300, 1),
	}, res.ContainerRecommendations[0].LowerBound)

	assert.Equal(t, corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewScaledQuantity(45, 1),
		corev1.ResourceMemory: *resource.NewScaledQuantity(4500, 1),
	}, res.ContainerRecommendations[0].UpperBound)
}

var podRecommendation *vpa_types.RecommendedPodResources = &vpa_types.RecommendedPodResources{
	ContainerRecommendations: []vpa_types.RecommendedContainerResources{
		{
			ContainerName: "ctr-name",
			LowerBound: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(5, 1),
				corev1.ResourceMemory: *resource.NewScaledQuantity(10, 1)},
			Target: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(50, 1),
				corev1.ResourceMemory: *resource.NewScaledQuantity(100, 1)},
			UpperBound: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(150, 1),
				corev1.ResourceMemory: *resource.NewScaledQuantity(200, 1)},
		},
	},
}
var applyTestCases = []struct {
	PodRecommendation         *vpa_types.RecommendedPodResources
	Policy                    *vpa_types.PodResourcePolicy
	ExpectedPodRecommendation *vpa_types.RecommendedPodResources
	ExpectedError             error
}{
	{
		PodRecommendation:         nil,
		Policy:                    nil,
		ExpectedPodRecommendation: nil,
		ExpectedError:             nil,
	},
	{
		PodRecommendation:         podRecommendation,
		Policy:                    nil,
		ExpectedPodRecommendation: podRecommendation,
		ExpectedError:             nil,
	},
}

func TestApply(t *testing.T) {
	containerName := "ctr-name"
	pod := test.Pod().WithName("pod1").AddContainer(test.Container().WithName(containerName).Get()).Get()

	for _, testCase := range applyTestCases {
		vpa := test.VerticalPodAutoscaler().WithContainer(containerName).Get()
		vpa.Spec.ResourcePolicy = testCase.Policy
		vpa.Status.Recommendation = testCase.PodRecommendation
		res, _, err := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{}).Apply(
			vpa, pod)
		assert.Equal(t, testCase.ExpectedPodRecommendation, res)
		assert.Equal(t, testCase.ExpectedError, err)
	}
}

var (
	recommendation = &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "foo",
				Target: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("42m"),
					corev1.ResourceMemory: resource.MustParse("42Mi"),
				},
				LowerBound: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("31m"),
					corev1.ResourceMemory: resource.MustParse("31Mi"),
				},
				UpperBound: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("53m"),
					corev1.ResourceMemory: resource.MustParse("53Mi"),
				},
				UncappedTarget: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("42m"),
					corev1.ResourceMemory: resource.MustParse("42Mi"),
				},
			},
		},
	}
)

func TestApplyVPAPolicy(t *testing.T) {
	tests := []struct {
		Name              string
		PodRecommendation *vpa_types.RecommendedPodResources
		ResourcePolicy    *vpa_types.PodResourcePolicy
		GlobalMaxAllowed  corev1.ResourceList
		Expected          *vpa_types.RecommendedPodResources
	}{
		{
			Name:              "recommendation is nil",
			PodRecommendation: nil,
			ResourcePolicy:    nil,
			GlobalMaxAllowed:  nil,
			Expected:          nil,
		},
		{
			Name:              "resource policy is nil and global max allowed is nil",
			PodRecommendation: recommendation,
			ResourcePolicy:    nil,
			GlobalMaxAllowed:  nil,
			Expected:          recommendation.DeepCopy(),
		},
		{
			Name:              "resource policy has min allowed and global max allowed is nil",
			PodRecommendation: recommendation,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "foo",
						MinAllowed: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("50m"),
							corev1.ResourceMemory: resource.MustParse("50Mi"),
						},
					},
				},
			},
			GlobalMaxAllowed: nil,
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("50m"),
							corev1.ResourceMemory: resource.MustParse("50Mi"),
						},
						LowerBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("50m"),
							corev1.ResourceMemory: resource.MustParse("50Mi"),
						},
						UpperBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("53m"),
							corev1.ResourceMemory: resource.MustParse("53Mi"),
						},
						UncappedTarget: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("42m"),
							corev1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
		{
			Name:              "resource policy has max allowed and global max allowed is nil",
			PodRecommendation: recommendation,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "foo",
						MaxAllowed: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("40m"),
							corev1.ResourceMemory: resource.MustParse("40Mi"),
						},
					},
				},
			},
			GlobalMaxAllowed: nil,
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("40m"),
							corev1.ResourceMemory: resource.MustParse("40Mi"),
						},
						LowerBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("31m"),
							corev1.ResourceMemory: resource.MustParse("31Mi"),
						},
						UpperBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("40m"),
							corev1.ResourceMemory: resource.MustParse("40Mi"),
						},
						UncappedTarget: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("42m"),
							corev1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
		{
			Name:              "resource policy is nil and global max allowed is set for cpu and memory",
			PodRecommendation: recommendation,
			ResourcePolicy:    nil,
			GlobalMaxAllowed: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("40m"),
				corev1.ResourceMemory: resource.MustParse("40Mi"),
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("40m"),
							corev1.ResourceMemory: resource.MustParse("40Mi"),
						},
						LowerBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("31m"),
							corev1.ResourceMemory: resource.MustParse("31Mi"),
						},
						UpperBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("40m"),
							corev1.ResourceMemory: resource.MustParse("40Mi"),
						},
						UncappedTarget: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("42m"),
							corev1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
		{
			Name:              "resource policy has maxAllowed and global max allowed is set",
			PodRecommendation: recommendation,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "foo",
						MinAllowed: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("35m"),
							corev1.ResourceMemory: resource.MustParse("35Mi"),
						},
					},
				},
			},
			GlobalMaxAllowed: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("42m"),
							corev1.ResourceMemory: resource.MustParse("42Mi"),
						},
						LowerBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("35m"),
							corev1.ResourceMemory: resource.MustParse("35Mi"),
						},
						UpperBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("50m"),
							corev1.ResourceMemory: resource.MustParse("50Mi"),
						},
						UncappedTarget: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("42m"),
							corev1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
		{
			Name:              "resource policy has max allowed for cpu and global max allowed is set for memory",
			PodRecommendation: recommendation,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "foo",
						MaxAllowed: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("40m"),
						},
					},
				},
			},
			GlobalMaxAllowed: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("40Mi"),
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("40m"),
							corev1.ResourceMemory: resource.MustParse("40Mi"),
						},
						LowerBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("31m"),
							corev1.ResourceMemory: resource.MustParse("31Mi"),
						},
						UpperBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("40m"),
							corev1.ResourceMemory: resource.MustParse("40Mi"),
						},
						UncappedTarget: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("42m"),
							corev1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
		{
			Name:              "resource policy has max allowed for cpu and global max allowed is set for cpu and memory",
			PodRecommendation: recommendation,
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "foo",
						MaxAllowed: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("40m"),
						},
					},
				},
			},
			GlobalMaxAllowed: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("30m"),
				corev1.ResourceMemory: resource.MustParse("30Mi"),
			},
			Expected: &vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "foo",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("40m"),
							corev1.ResourceMemory: resource.MustParse("30Mi"),
						},
						LowerBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("31m"),
							corev1.ResourceMemory: resource.MustParse("30Mi"),
						},
						UpperBound: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("40m"),
							corev1.ResourceMemory: resource.MustParse("30Mi"),
						},
						UncappedTarget: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("42m"),
							corev1.ResourceMemory: resource.MustParse("42Mi"),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			resourcePolicyCopy := tt.ResourcePolicy.DeepCopy()

			actual, err := ApplyVPAPolicy(tt.PodRecommendation, tt.ResourcePolicy, tt.GlobalMaxAllowed)
			assert.Nil(t, err)
			assert.Equal(t, tt.Expected, actual)

			// Make sure that the func does not have a side affect and does not modify the passed resource policy.
			assert.Equal(t, resourcePolicyCopy, tt.ResourcePolicy)
		})
	}
}

type fakeLimitRangeCalculator struct {
	containerLimitRange corev1.LimitRangeItem
	podLimitRange       corev1.LimitRangeItem
}

func (nlrc *fakeLimitRangeCalculator) GetContainerLimitRangeItem(namespace string) (*corev1.LimitRangeItem, error) {
	return &nlrc.containerLimitRange, nil
}

func (nlrc *fakeLimitRangeCalculator) GetPodLimitRangeItem(namespace string) (*corev1.LimitRangeItem, error) {
	return &nlrc.podLimitRange, nil
}

func TestApplyCapsToLimitRange(t *testing.T) {
	limitRange := corev1.LimitRangeItem{
		Type: corev1.LimitTypeContainer,
		Max: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("1"),
		},
		Min: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("500M"),
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
				Target: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("200M"),
				},
			},
		},
	}
	vpa.Status.Recommendation = &recommendation

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1G"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1G"),
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
				Target: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1000m"),
					corev1.ResourceMemory: resource.MustParse("500000000"),
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
		pod          corev1.Pod
		limitRange   corev1.LimitRangeItem
		resourceName corev1.ResourceName
		expect       []vpa_types.RecommendedContainerResources
	}{
		{
			name: "cap target cpu to max",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("1"),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("1"),
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
						{
							Name: "container2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			limitRange: corev1.LimitRangeItem{
				Max: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("1"),
				},
			},
			resourceName: corev1.ResourceCPU,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewMilliQuantity(500, resource.DecimalSI),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewMilliQuantity(500, resource.DecimalSI),
					},
				},
			},
		},
		{
			name: "cap cpu to max",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("1"),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("1"),
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
						{
							Name: "container2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			limitRange: corev1.LimitRangeItem{
				Max: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("1"),
				},
			},
			resourceName: corev1.ResourceCPU,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewMilliQuantity(500, resource.DecimalSI),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewMilliQuantity(500, resource.DecimalSI),
					},
				},
			},
		},
		{
			name: "cap mem to min",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1G"),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1G"),
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1"),
								},
							},
						},
						{
							Name: "container2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			limitRange: corev1.LimitRangeItem{
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("4G"),
				},
			},
			resourceName: corev1.ResourceMemory,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2000000000"),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2000000000"),
					},
				},
			},
		},
		{
			name: "cap mem request to pod min",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1G"),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1G"),
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("2"),
								},
							},
						},
						{
							Name: "container2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			limitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypePod,
				Max: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("10G"),
				},
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("4G"),
				},
			},
			resourceName: corev1.ResourceMemory,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2000000000"),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2000000000"),
					},
				},
			},
		},
		{
			name: "cap mem request to pod min, only one container with recommendation",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1G"),
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1G"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("2G"),
								},
							},
						},
						{
							Name: "container2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1G"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("2G"),
								},
							},
						},
					},
				},
			},
			limitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypePod,
				Max: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("10G"),
				},
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("4G"),
				},
			},
			resourceName: corev1.ResourceMemory,
			expect: []vpa_types.RecommendedContainerResources{
				{
					// TODO: This is incorrect; The pod will be rejected by limit range because sum of its
					// pod requests is too small - it's 3Gi(2Gi for `container1` (from recommendation) and 1Gi
					// for `container2` (unchanged incoming request)) and minimum is 4Gi.
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("2000000000"),
					},
				},
			},
		},
		{
			name: "cap mem request to pod min, extra recommendation",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1G"),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1G"),
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			limitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypePod,
				Max: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("10G"),
				},
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("4G"),
				},
			},
			resourceName: corev1.ResourceMemory,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("1G"),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("4000000000"),
					},
				},
			},
		},
		{
			name: "cap target cpu to pod min",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("15m"),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("100m"),
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1m"),
								},
							},
						},
						{
							Name: "container2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1m"),
								},
							},
						},
					},
				},
			},
			limitRange: corev1.LimitRangeItem{
				Min: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("150m"),
				},
			},
			resourceName: corev1.ResourceCPU,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewMilliQuantity(20, resource.DecimalSI), // ceil((15*150)/115), for more details check PR #8946
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewMilliQuantity(131, resource.DecimalSI), // ceil((100*150)/115)
					},
				},
			},
		},
		{
			name: "cap target cpu to pod max",
			resources: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("15m"),
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("100m"),
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1m"),
								},
							},
						},
						{
							Name: "container2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1m"),
								},
							},
						},
					},
				},
			},
			limitRange: corev1.LimitRangeItem{
				Max: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("90m"),
				},
			},
			resourceName: corev1.ResourceCPU,
			expect: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "container1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewMilliQuantity(11, resource.DecimalSI), // floor((15*90)/115), for more details check PR #8946
					},
				},
				{
					ContainerName: "container2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewMilliQuantity(78, resource.DecimalSI), // floor((100*90)/115)
					},
				},
			},
		},
	}
	getTarget := func(rl vpa_types.RecommendedContainerResources) *corev1.ResourceList { return &rl.Target }
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
		pod                 corev1.Pod
		expectedTarget      corev1.ResourceList
		expectedUpperBound  corev1.ResourceList
		policy              *vpa_types.PodResourcePolicy
		expect              vpa_types.RecommendedPodResources
		expectAnnotations   map[string][]string
		containerLimitRange corev1.LimitRangeItem
		podLimitRange       corev1.LimitRangeItem
	}{
		{
			name: "caps to min range if above container limit",
			resources: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("600M"),
								},
							},
						},
					},
				},
			},
			containerLimitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypeContainer,
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("500M"),
				},
			},
			expect: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("500M"),
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
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
				},
			},
			containerLimitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypeContainer,
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("500M"),
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
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("100M"),
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
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
				},
			},
			podLimitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypePod,
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("500M"),
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
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("100M"),
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
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("200M"),
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
			podLimitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypePod,
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("500M"),
					corev1.ResourceCPU:    resource.MustParse("2"),
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
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("5m"),
							corev1.ResourceMemory: resource.MustParse("100M"),
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
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
						{
							Name: "container2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
				},
			},
			podLimitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypePod,
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("500M"),
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
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("400000000"),
						},
					},
					{
						ContainerName: "container2",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("100000000"),
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
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("200M"),
						},
					},
					{
						ContainerName: "container2",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
						{
							Name: "container2",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("50M"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
				},
			},
			podLimitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypePod,
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("500M"),
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
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("250000000"),
						},
					},
					{
						ContainerName: "container2",
						Target: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("250000000"),
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
				Target: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			{
				ContainerName: "container-2",
				Target: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
		},
	}
	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container-1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("50M"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("50M"),
						},
					},
				},
				{
					Name: "container-2",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("50M"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("50M"),
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name                   string
		limitRange             corev1.LimitRangeItem
		expectedRecommendation vpa_types.RecommendedPodResources
	}{
		{
			name: "cap to max",
			limitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypePod,
				Max: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			expectedRecommendation: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container-1",
						Target: corev1.ResourceList{
							corev1.ResourceMemory: *resource.NewQuantity(357913941, resource.BinarySI),
						},
					},
					{
						ContainerName: "container-2",
						Target: corev1.ResourceList{
							corev1.ResourceMemory: *resource.NewQuantity(715827882, resource.BinarySI),
						},
					},
				},
			},
		},
		{
			name: "cap to min",
			limitRange: corev1.LimitRangeItem{
				Type: corev1.LimitTypePod,
				Min: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
			},
			expectedRecommendation: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						ContainerName: "container-1",
						Target: corev1.ResourceList{
							corev1.ResourceMemory: *resource.NewQuantity(1431655765, resource.BinarySI),
						},
					},
					{
						ContainerName: "container-2",
						Target: corev1.ResourceList{
							corev1.ResourceMemory: *resource.NewQuantity(2863311530, resource.BinarySI),
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

func TestEnsureValidBounds(t *testing.T) {
	tests := []struct {
		name               string
		containerLevelRecs []vpa_types.RecommendedContainerResources
		expected           []vpa_types.RecommendedContainerResources
	}{
		{
			// Here, LowerBound <= Target <= UpperBound holds for all containers, so no action is required.
			name: "no violation",
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
			},
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
			},
		},
		{
			// Here, LowerBound <= Target <= UpperBound holds for all containers, so no action is required.
			name: "no violation",
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
				},
				{
					ContainerName: "c2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
			},
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
				},
				{
					ContainerName: "c2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
			},
		},
		{
			name: "cpu and memory lower bounds are violated in c1, but no additional delta can be added to c2",
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("25m"),
						corev1.ResourceMemory: resource.MustParse("25Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
				},
			},
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),  // -5m
						corev1.ResourceMemory: resource.MustParse("20Mi"), // -5Mi
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
				},
			},
		},
		{
			name: "cpu and memory lower bounds are violated in c1, add delta proportionally",
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("28m"),
						corev1.ResourceMemory: resource.MustParse("28Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
				{
					ContainerName: "c3",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("50m"),
						corev1.ResourceMemory: resource.MustParse("50Mi"),
					},
				},
				{
					ContainerName: "c4",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
			},
			// There is a violation in c1, as the LowerBound is greater than the Target. To fix this violation, we need to lower c1's LowerBound to 20m and 20Mi,
			// since the reference point (i.e. c1's Target) is 20m and 20Mi.
			// Then, we need to add 8 millicores and 8 MiB proportionally to the others so that the sum of LowerBound values still equals 128m and 128Mi.
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),  // -8m
						corev1.ResourceMemory: resource.MustParse("20Mi"), // -8Mi
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(32, resource.DecimalSI), // + 2
						corev1.ResourceMemory: *resource.NewQuantity(33973862, resource.BinarySI), // 31457280 + floor(8388608 x 0,3)
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
				{
					ContainerName: "c3",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(44, resource.DecimalSI), // + 4
						corev1.ResourceMemory: *resource.NewQuantity(45298484, resource.BinarySI), // 41943040 + ceil(8388608 x 0,4)
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("50m"),
						corev1.ResourceMemory: resource.MustParse("50Mi"),
					},
				},
				{
					ContainerName: "c4",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(32, resource.DecimalSI), // + 2
						corev1.ResourceMemory: *resource.NewQuantity(33973862, resource.BinarySI), // 31457280 + floor(8388608 x 0,3)
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
			},
		},
		{
			name: "cpu and memory lower bounds are violated in c1, add delta proportionally, except in c2",
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("28m"),
						corev1.ResourceMemory: resource.MustParse("28Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
				{
					ContainerName: "c3",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("60m"),
						corev1.ResourceMemory: resource.MustParse("60Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("80m"),
						corev1.ResourceMemory: resource.MustParse("80Mi"),
					},
				},
				{
					ContainerName: "c4",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("80m"),
						corev1.ResourceMemory: resource.MustParse("80Mi"),
					},
				},
			},
			// There is a violation in c1, as the LowerBound is greater than the Target. To fix this violation, we need to lower c1's LowerBound to 20m and 20Mi,
			// since the reference point (i.e. c1's Target) is 20m and 20Mi.
			// In this case, we cannot add values to c2, as its Target equals its LowerBound.
			// Then, we need to add 8 millicores and 8 MiB proportionally to the others so that the sum of LowerBound values still equals 168m and 168Mi.
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),  // -8m
						corev1.ResourceMemory: resource.MustParse("20Mi"), // -8Mi
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
				},
				{
					ContainerName: "c2",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(40, resource.DecimalSI), // 40m
						corev1.ResourceMemory: *resource.NewQuantity(41943040, resource.BinarySI), // 40Mi
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
				{
					ContainerName: "c3",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(65, resource.DecimalSI),
						corev1.ResourceMemory: *resource.NewQuantity(67947725, resource.BinarySI), // 62914560 + (8388608 x 0,6)
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("80m"),
						corev1.ResourceMemory: resource.MustParse("80Mi"),
					},
				},
				{
					ContainerName: "c4",
					LowerBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(43, resource.DecimalSI),
						corev1.ResourceMemory: *resource.NewQuantity(45298483, resource.BinarySI), // 41943040 + (8388608 x 0,4)
					},
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("80m"),
						corev1.ResourceMemory: resource.MustParse("80Mi"),
					},
				},
			},
		},
		{
			name: "cpu and memory upper bounds are violated in c1, subtract delta proportionally",
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("12m"),
						corev1.ResourceMemory: resource.MustParse("12Mi"),
					},
				},
				{
					ContainerName: "c2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("10Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
				},
				{
					ContainerName: "c3",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("10Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
				{
					ContainerName: "c4",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("10Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
				},
			},
			// There is a violation in c1, as the UpperBound is lower than the Target. To fix this violation, we need to increase c1's UpperBound to 20m and 20Mi,
			// since the reference point (i.e. c1's Target) is 20m and 20Mi.
			// Then, we need to subtract 8 millicores and 8 MiB proportionally from the others so that the sum of UpperBound values still equals 112m and 112Mi.
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),  // +8
						corev1.ResourceMemory: resource.MustParse("20Mi"), // +8
					},
				},
				{
					ContainerName: "c2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("10Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(28, resource.DecimalSI), // -2
						corev1.ResourceMemory: *resource.NewQuantity(28940698, resource.BinarySI), // 31457280 - floor(8388608 x 0.3)
					},
				},
				{
					ContainerName: "c3",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("10Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(36, resource.DecimalSI), // -4
						corev1.ResourceMemory: *resource.NewQuantity(38587596, resource.BinarySI), // 41943040 - ceil(8388608 x 0.4)
					},
				},
				{
					ContainerName: "c4",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("10Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(28, resource.DecimalSI), // -2
						corev1.ResourceMemory: *resource.NewQuantity(28940698, resource.BinarySI), // 31457280 - floor(8388608 x 0.3)
					},
				},
			},
		},
		{
			name: "cpu and memory upper bounds are violated in c1, subtract delta proportionally, except in c2",
			containerLevelRecs: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("12m"),
						corev1.ResourceMemory: resource.MustParse("12Mi"),
					},
				},
				{
					ContainerName: "c2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
				},
				{
					ContainerName: "c3",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("10Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("40m"),
						corev1.ResourceMemory: resource.MustParse("40Mi"),
					},
				},
				{
					ContainerName: "c4",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("10Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("60m"),
						corev1.ResourceMemory: resource.MustParse("60Mi"),
					},
				},
			},
			// There is a violation in c1, as the UpperBound is lower than the Target. To fix this violation, we need to increase c1's UpperBound to 20m and 20Mi,
			// since the reference point (i.e. c1's Target) is 20m and 20Mi.
			// In this case, we cannot subtract values from c2, as its Target equals its UpperBound.
			// Then, we need to subtract 8m and 8Mi proportionally from the others so that the sum of UpperBound values still equals 142m and 142Mi.
			expected: []vpa_types.RecommendedContainerResources{
				{
					ContainerName: "c1",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),
						corev1.ResourceMemory: resource.MustParse("20Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("20m"),  // +8m
						corev1.ResourceMemory: resource.MustParse("20Mi"), // +8Mi
					},
				},
				{
					ContainerName: "c2",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("30m"),
						corev1.ResourceMemory: resource.MustParse("30Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(30, resource.DecimalSI), // 30m
						corev1.ResourceMemory: *resource.NewQuantity(31457280, resource.BinarySI), // 30Mi
					},
				},
				{
					ContainerName: "c3",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("10Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(37, resource.DecimalSI), // -3
						corev1.ResourceMemory: *resource.NewQuantity(38587597, resource.BinarySI), // 41943040 - floor(8388608 x 0.4)
					},
				},
				{
					ContainerName: "c4",
					Target: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("10Mi"),
					},
					UpperBound: corev1.ResourceList{
						corev1.ResourceCPU:    *resource.NewMilliQuantity(55, resource.DecimalSI), // -5
						corev1.ResourceMemory: *resource.NewQuantity(57881395, resource.BinarySI), // 62914560 - ceil(8388608 x 0.6)
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

func Test_roundPreservingSum(t *testing.T) {
	tests := []struct {
		name           string
		floats         []float64
		expectedFloats []float64
	}{
		{
			name:           "example1",
			floats:         []float64{0, 1, 3.5, 3.5},
			expectedFloats: []float64{0, 1, 4, 3},
		},
		{
			name:           "example2",
			floats:         []float64{0, 1, 2.8, 3.2},
			expectedFloats: []float64{0, 1, 2, 4},
		},
		{
			name:           "example3",
			floats:         []float64{0, 1, 3, 3},
			expectedFloats: []float64{0, 1, 3, 3},
		},
		{
			name:           "example4",
			floats:         []float64{0, 0, 0},
			expectedFloats: []float64{0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			floats := roundPreservingSum(tt.floats)
			assert.Equal(t, tt.expectedFloats, floats)
		})
	}
}
