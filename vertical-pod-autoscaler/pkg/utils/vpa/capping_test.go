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
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
)

func TestRecommendationNotAvailable(t *testing.T) {
	container := apiv1.Container{Name: "ctr-name"}
	podRecommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				Name: "ctr-name-other",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(100, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(50000, 1),
				},
			},
		},
	}
	policy := vpa_types.PodResourcePolicy{}

	res, err := GetCappedRecommendationForContainer(container, &podRecommendation, &policy)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestRecommendationChosenForProperContainer(t *testing.T) {
	container := apiv1.Container{Name: "ctr-name"}
	podRecommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				Name: "ctr-name-other",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(100, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(50000, 1),
				},
			},
			{
				Name: "ctr-name",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(5000, 1),
				},
			},
		},
	}
	policy := vpa_types.PodResourcePolicy{}

	res, err := GetCappedRecommendationForContainer(container, &podRecommendation, &policy)
	assert.Nil(t, err)
	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(5000, 1),
	}, res.Target)
}

func TestRecommendationCappedToLimit(t *testing.T) {
	container := apiv1.Container{
		Name: "ctr-name",
		Resources: apiv1.ResourceRequirements{
			Limits: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewScaledQuantity(3, 1),
				apiv1.ResourceMemory: *resource.NewScaledQuantity(7000, 1),
			},
		},
	}
	podRecommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				Name: "ctr-name",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(5000, 1),
				},
				MaxRecommended: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(9000, 1),
				},
			},
		},
	}
	policy := vpa_types.PodResourcePolicy{}

	res, err := GetCappedRecommendationForContainer(container, &podRecommendation, &policy)
	assert.Nil(t, err)
	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(3, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(5000, 1),
	}, res.Target)
	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(2, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(7000, 1),
	}, res.MaxRecommended)
}

func TestRecommendationCappedToMinMaxPolicy(t *testing.T) {
	container := apiv1.Container{Name: "ctr-name"}
	podRecommendation := vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{
			{
				Name: "ctr-name",
				Target: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(10, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(5000, 1),
				},
				MinRecommended: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewScaledQuantity(50, 1),
					apiv1.ResourceMemory: *resource.NewScaledQuantity(4300, 1),
				},
			},
		},
	}
	policy := vpa_types.PodResourcePolicy{
		ContainerPolicies: []vpa_types.ContainerResourcePolicy{
			{
				Name: "ctr-name",
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

	res, err := GetCappedRecommendationForContainer(container, &podRecommendation, &policy)
	assert.Nil(t, err)
	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(40, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(4500, 1),
	}, res.Target)
	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourceCPU:    *resource.NewScaledQuantity(45, 1),
		apiv1.ResourceMemory: *resource.NewScaledQuantity(4300, 1),
	}, res.MinRecommended)
}
