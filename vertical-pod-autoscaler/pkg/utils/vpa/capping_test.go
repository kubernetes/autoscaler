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
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
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
	limitRange apiv1.LimitRangeItem
}

func (nlrc *fakeLimitRangeCalculator) GetContainerLimitRangeItem(namespace string) (*apiv1.LimitRangeItem, error) {
	return &nlrc.limitRange, nil
}

func TestApplyCapsToLimitRange(t *testing.T) {
	limitRange := apiv1.LimitRangeItem{
		Type: apiv1.LimitTypeContainer,
		Max: apiv1.ResourceList{
			apiv1.ResourceCPU:    resource.MustParse("1"),
			apiv1.ResourceMemory: resource.MustParse("1G"),
		},
	}
	recommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				ContainerName: "container",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("2"),
					apiv1.ResourceMemory: resource.MustParse("10G"),
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
				LowerBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewQuantity(0, resource.DecimalSI),
					apiv1.ResourceMemory: *resource.NewQuantity(0, resource.BinarySI),
				},
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("1000m"),
					apiv1.ResourceMemory: resource.MustParse("1000000000000m"),
				},
				UpperBound: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewQuantity(0, resource.DecimalSI),
					apiv1.ResourceMemory: *resource.NewQuantity(0, resource.BinarySI),
				},
			},
		},
	}

	calculator := fakeLimitRangeCalculator{limitRange}
	processor := NewCappingRecommendationProcessor(&calculator)
	processedRecommendation, annotations, err := processor.Apply(&recommendation, nil, nil, &pod)
	assert.NoError(t, err)
	assert.Equal(t, map[string][]string{"container": {"changed CPU limit to fit within limit range", "changed memory limit to fit within limit range"}}, annotations)
	assert.Equal(t, expectedRecommendation, *processedRecommendation)
}
