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
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestRecommendationNotAvailable(t *testing.T) {
	pod := test.Pod().WithName("pod1").AddContainer(test.BuildTestContainer("ctr-name", "", "")).Get()
	podRecommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "ctr-name-other",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(100, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(50000, 1),
				},
			},
		},
	}
	policy := vpa_types.PodResourcePolicy{}

	res, annotations, err := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{}).Apply(&podRecommendation, &policy, nil, pod)
	assert.Nil(t, err)
	assert.Empty(t, annotations)
	assert.Empty(t, res.ContainerRecommendations)
}

func TestRecommendationCappedToLimit(t *testing.T) {
	pod := test.Pod().WithName("pod1").AddContainer(test.BuildTestContainer("ctr-name", "", "")).Get()
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
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(8000, 1),
				},
				UpperBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(9000, 1),
				},
			},
		},
	}
	policy := vpa_types.PodResourcePolicy{}
	res, annotations, err := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{}).Apply(&podRecommendation, &policy, nil, pod)
	assert.Nil(t, err)
	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(3, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(7000, 1),
	}, res.ContainerRecommendations[0].Target)

	assert.Contains(t, annotations, "ctr-name")
	assert.Contains(t, annotations["ctr-name"], "memory capped to container limit")

	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(7000, 1),
	}, res.ContainerRecommendations[0].UpperBound)
}

func TestRecommendationCappedToMinMaxPolicy(t *testing.T) {
	pod := test.Pod().WithName("pod1").AddContainer(test.BuildTestContainer("ctr-name", "", "")).Get()
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

	res, annotations, err := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{}).Apply(&podRecommendation, &policy, nil, pod)
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

var podRecommendation *vpa_types.RecommendedPodResources = &vpa_types.RecommendedPodResources{
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
	pod := test.Pod().WithName("pod1").AddContainer(test.BuildTestContainer("ctr-name", "", "")).Get()

	for _, testCase := range applyTestCases {
		res, _, err := NewCappingRecommendationProcessor(&fakeLimitRangeCalculator{}).Apply(
			testCase.PodRecommendation, testCase.Policy, nil, pod)
		assert.Equal(t, testCase.ExpectedPodRecommendation, res)
		assert.Equal(t, testCase.ExpectedError, err)
	}
}

func TestApplyVpa(t *testing.T) {
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
				UncappedTarget: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(30, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(5000, 1),
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

	res, err := ApplyVPAPolicy(&podRecommendation, &policy)
	assert.Nil(t, err)
	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(40, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(4500, 1),
	}, res.ContainerRecommendations[0].Target)

	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(30, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(5000, 1),
	}, res.ContainerRecommendations[0].UncappedTarget)

	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(40, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(4300, 1),
	}, res.ContainerRecommendations[0].LowerBound)

	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(45, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(4500, 1),
	}, res.ContainerRecommendations[0].UpperBound)
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
					apiv1.ResourceMemory: resource.MustParse("500000000000m"),
				},
			},
		},
	}

	calculator := fakeLimitRangeCalculator{containerLimitRange: limitRange}
	processor := NewCappingRecommendationProcessor(&calculator)
	processedRecommendation, annotations, err := processor.Apply(&recommendation, nil, nil, &pod)
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
						apiv1.ResourceMemory: resource.MustParse("2000000000000m"),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("2000000000000m"),
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
						apiv1.ResourceMemory: resource.MustParse("2000000000000m"),
					},
				},
				{
					ContainerName: "container2",
					Target: apiv1.ResourceList{
						apiv1.ResourceMemory: resource.MustParse("2000000000000m"),
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
	tests := []struct {
		name              string
		resources         vpa_types.RecommendedPodResources
		pod               apiv1.Pod
		limitRange        apiv1.LimitRangeItem
		expect            vpa_types.RecommendedPodResources
		expectAnnotations []string
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
			limitRange: apiv1.LimitRangeItem{
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
			expectAnnotations: []string{
				"memory capped to fit Min in container LimitRange",
			},
		}, {
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
			limitRange: apiv1.LimitRangeItem{
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
							apiv1.ResourceMemory: resource.MustParse("100M"),
						},
					},
				},
			},
			expectAnnotations: []string{
				"memory capped to fit Min in container LimitRange",
				"memory capped to container limit",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calculator := fakeLimitRangeCalculator{containerLimitRange: tc.limitRange}
			processor := NewCappingRecommendationProcessor(&calculator)
			processedRecommendation, annotations, err := processor.Apply(&tc.resources, nil, nil, &tc.pod)
			assert.NoError(t, err)
			assert.Contains(t, annotations, "container")
			assert.ElementsMatch(t, tc.expectAnnotations, annotations["container"])
			assert.Equal(t, tc.expect, *processedRecommendation)
		})
	}
}
