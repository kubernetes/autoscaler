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
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestMinResourcesApplied(t *testing.T) {
	constCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(0.001))
	constMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(1e6))

	recommender := podResourceRecommender{
		targetCPU:        constCPUEstimator,
		targetMemory:     constMemoryEstimator,
		lowerBoundCPU:    constCPUEstimator,
		lowerBoundMemory: constMemoryEstimator,
		upperBoundCPU:    constCPUEstimator,
		upperBoundMemory: constMemoryEstimator,
	}

	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		"container-1": &model.AggregateContainerState{},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
	assert.Equal(t, model.CPUAmountFromCores(*podMinCPUMillicores/1000), recommendedResources["container-1"].Target[model.ResourceCPU])
	assert.Equal(t, model.MemoryAmountFromBytes(*podMinMemoryMb*1024*1024), recommendedResources["container-1"].Target[model.ResourceMemory])
}

func TestMinResourcesSplitAcrossContainers(t *testing.T) {
	constCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(0.001))
	constMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(1e6))

	recommender := podResourceRecommender{
		targetCPU:        constCPUEstimator,
		targetMemory:     constMemoryEstimator,
		lowerBoundCPU:    constCPUEstimator,
		lowerBoundMemory: constMemoryEstimator,
		upperBoundCPU:    constCPUEstimator,
		upperBoundMemory: constMemoryEstimator,
	}

	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		"container-1": &model.AggregateContainerState{},
		"container-2": &model.AggregateContainerState{},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
	assert.Equal(t, model.CPUAmountFromCores((*podMinCPUMillicores/1000)/2), recommendedResources["container-1"].Target[model.ResourceCPU])
	assert.Equal(t, model.CPUAmountFromCores((*podMinCPUMillicores/1000)/2), recommendedResources["container-2"].Target[model.ResourceCPU])
	assert.Equal(t, model.MemoryAmountFromBytes((*podMinMemoryMb*1024*1024)/2), recommendedResources["container-1"].Target[model.ResourceMemory])
	assert.Equal(t, model.MemoryAmountFromBytes((*podMinMemoryMb*1024*1024)/2), recommendedResources["container-2"].Target[model.ResourceMemory])
}

func TestControlledResourcesFiltered(t *testing.T) {
	constCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(0.001))
	constMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(1e6))

	recommender := podResourceRecommender{
		targetCPU:        constCPUEstimator,
		targetMemory:     constMemoryEstimator,
		lowerBoundCPU:    constCPUEstimator,
		lowerBoundMemory: constMemoryEstimator,
		upperBoundCPU:    constCPUEstimator,
		upperBoundMemory: constMemoryEstimator,
	}

	containerName := "container-1"
	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		containerName: &model.AggregateContainerState{
			ControlledResources: &[]model.ResourceName{model.ResourceMemory},
		},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
	assert.Contains(t, recommendedResources[containerName].Target, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].LowerBound, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].UpperBound, model.ResourceMemory)
	assert.NotContains(t, recommendedResources[containerName].Target, model.ResourceCPU)
	assert.NotContains(t, recommendedResources[containerName].LowerBound, model.ResourceCPU)
	assert.NotContains(t, recommendedResources[containerName].UpperBound, model.ResourceCPU)
}

func TestControlledResourcesFilteredDefault(t *testing.T) {
	constCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(0.001))
	constMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(1e6))

	recommender := podResourceRecommender{
		targetCPU:        constCPUEstimator,
		targetMemory:     constMemoryEstimator,
		lowerBoundCPU:    constCPUEstimator,
		lowerBoundMemory: constMemoryEstimator,
		upperBoundCPU:    constCPUEstimator,
		upperBoundMemory: constMemoryEstimator,
	}

	containerName := "container-1"
	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		containerName: &model.AggregateContainerState{
			ControlledResources: &[]model.ResourceName{model.ResourceMemory, model.ResourceCPU},
		},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
	assert.Contains(t, recommendedResources[containerName].Target, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].LowerBound, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].UpperBound, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].Target, model.ResourceCPU)
	assert.Contains(t, recommendedResources[containerName].LowerBound, model.ResourceCPU)
	assert.Contains(t, recommendedResources[containerName].UpperBound, model.ResourceCPU)
}

func TestMapToListOfRecommendedContainerResources(t *testing.T) {
	cases := []struct {
		name         string
		resources    RecommendedPodResources
		expectedLast []string
	}{
		{
			name: "All recommendations sorted",
			resources: RecommendedPodResources{
				"a-container": RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(1), model.ResourceMemory: model.MemoryAmountFromBytes(1e6)}},
				"b-container": RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(2), model.ResourceMemory: model.MemoryAmountFromBytes(2e6)}},
				"c-container": RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(3), model.ResourceMemory: model.MemoryAmountFromBytes(3e6)}},
				"d-container": RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(4), model.ResourceMemory: model.MemoryAmountFromBytes(4e6)}},
			},
			expectedLast: []string{
				"a-container",
				"b-container",
				"c-container",
				"d-container",
			},
		},
		{
			name: "All recommendations unsorted",
			resources: RecommendedPodResources{
				"b-container": RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(1), model.ResourceMemory: model.MemoryAmountFromBytes(1e6)}},
				"a-container": RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(2), model.ResourceMemory: model.MemoryAmountFromBytes(2e6)}},
				"d-container": RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(3), model.ResourceMemory: model.MemoryAmountFromBytes(3e6)}},
				"c-container": RecommendedContainerResources{Target: model.Resources{model.ResourceCPU: model.CPUAmountFromCores(4), model.ResourceMemory: model.MemoryAmountFromBytes(4e6)}},
			},
			expectedLast: []string{
				"a-container",
				"b-container",
				"c-container",
				"d-container",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outRecommendations := MapToListOfRecommendedContainerResources(tc.resources)
			for i, outRecommendation := range outRecommendations.ContainerRecommendations {
				containerName := tc.expectedLast[i]
				assert.Equal(t, containerName, outRecommendation.ContainerName)
				// also check that the recommendation is not changed
				assert.Equal(t, int64(tc.resources[containerName].Target[model.ResourceCPU]), outRecommendation.Target.Cpu().MilliValue())
				assert.Equal(t, int64(tc.resources[containerName].Target[model.ResourceMemory]), outRecommendation.Target.Memory().Value())
			}
		})
	}
}

func TestCalculatePodlevelRecommendations(t *testing.T) {
	cases := []struct {
		name                       string
		containerRecommendations   []vpa_types.RecommendedContainerResources
		expectedPodRecommendations vpa_types.RecommendedPodRes
	}{
		{
			name: "no container recommendations",
			containerRecommendations: []vpa_types.RecommendedContainerResources{
				test.Recommendation().
					WithContainer("container1").
					WithTarget("", "").
					WithLowerBound("", "").
					WithUpperBound("", "").
					GetContainerResources(),
			},
			expectedPodRecommendations: vpa_types.RecommendedPodRes{
				Target: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("0"),
					v1.ResourceMemory: resource.MustParse("0"),
				},
				LowerBound: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("0"),
					v1.ResourceMemory: resource.MustParse("0"),
				},
				UpperBound: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("0"),
					v1.ResourceMemory: resource.MustParse("0"),
				},
				UncappedTarget: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("0"),
					v1.ResourceMemory: resource.MustParse("0"),
				},
			},
		},
		{
			name: "one container resource does not include a recommendations",
			containerRecommendations: []vpa_types.RecommendedContainerResources{
				test.Recommendation().
					WithContainer("container1").
					WithTarget("2m", "10Mi").
					WithLowerBound("2m", "10Mi").
					WithUpperBound("2m", "10Mi").
					GetContainerResources(),
				test.Recommendation().
					WithContainer("container2").
					WithTarget("8m", "").
					WithLowerBound("8m", "").
					WithUpperBound("8m", "").
					GetContainerResources(),
			},
			expectedPodRecommendations: vpa_types.RecommendedPodRes{
				Target: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("10Mi"),
				},
				LowerBound: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("10Mi"),
				},
				UpperBound: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("10Mi"),
				},
				UncappedTarget: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("10Mi"),
				},
			},
		},
		{
			name: "all containers contain recommendations",
			containerRecommendations: []vpa_types.RecommendedContainerResources{
				test.Recommendation().
					WithContainer("container1").
					WithTarget("4m", "4Mi").
					WithLowerBound("4m", "4Mi").
					WithUpperBound("4m", "4Mi").
					GetContainerResources(),
				test.Recommendation().
					WithContainer("container2").
					WithTarget("4m", "4Mi").
					WithLowerBound("4m", "4Mi").
					WithUpperBound("4m", "4Mi").
					GetContainerResources(),
				test.Recommendation().
					WithContainer("container3").
					WithTarget("2m", "2Mi").
					WithLowerBound("2m", "2Mi").
					WithUpperBound("2m", "2Mi").
					GetContainerResources(),
			},
			expectedPodRecommendations: vpa_types.RecommendedPodRes{
				Target: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("10Mi"),
				},
				LowerBound: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("10Mi"),
				},
				UpperBound: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("10Mi"),
				},
				UncappedTarget: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("10Mi"),
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outRecommendations := CalculatePodlevelRecommendations(tc.containerRecommendations)
			assert.Equal(t, tc.expectedPodRecommendations.Target.Cpu().MilliValue(), outRecommendations.PodRecommendations.Target.Cpu().MilliValue())
			assert.Equal(t, tc.expectedPodRecommendations.Target.Memory().Value(), outRecommendations.PodRecommendations.Target.Memory().Value())

			assert.Equal(t, tc.expectedPodRecommendations.LowerBound.Cpu().MilliValue(), outRecommendations.PodRecommendations.LowerBound.Cpu().MilliValue())
			assert.Equal(t, tc.expectedPodRecommendations.LowerBound.Memory().Value(), outRecommendations.PodRecommendations.LowerBound.Memory().Value())

			assert.Equal(t, tc.expectedPodRecommendations.UpperBound.Cpu().MilliValue(), outRecommendations.PodRecommendations.UpperBound.Cpu().MilliValue())
			assert.Equal(t, tc.expectedPodRecommendations.UpperBound.Memory().Value(), outRecommendations.PodRecommendations.UpperBound.Memory().Value())

			assert.Equal(t, tc.expectedPodRecommendations.UncappedTarget.Cpu().MilliValue(), outRecommendations.PodRecommendations.UncappedTarget.Cpu().MilliValue())
			assert.Equal(t, tc.expectedPodRecommendations.UncappedTarget.Memory().Value(), outRecommendations.PodRecommendations.UncappedTarget.Memory().Value())
		})
	}
}
