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
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

func TestMinResourcesApplied(t *testing.T) {
	constEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.001),
		model.ResourceMemory: model.MemoryAmountFromBytes(1e6),
	})
	recommender := podResourceRecommender{
		constEstimator,
		constEstimator,
		constEstimator}

	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		"container-1": &model.AggregateContainerState{},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap, len(containerNameToAggregateStateMap))
	assert.Equal(t, model.CPUAmountFromCores(*podMinCPUMillicores/1000), recommendedResources["container-1"].Target[model.ResourceCPU])
	assert.Equal(t, model.MemoryAmountFromBytes(*podMinMemoryMb*1024*1024), recommendedResources["container-1"].Target[model.ResourceMemory])
}

func TestMinResourcesSplitAcrossContainers(t *testing.T) {
	constEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.001),
		model.ResourceMemory: model.MemoryAmountFromBytes(1e6),
	})
	recommender := podResourceRecommender{
		constEstimator,
		constEstimator,
		constEstimator}

	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		"container-1": &model.AggregateContainerState{},
		"container-2": &model.AggregateContainerState{},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap, len(containerNameToAggregateStateMap))
	assert.Equal(t, model.CPUAmountFromCores((*podMinCPUMillicores/1000)/2), recommendedResources["container-1"].Target[model.ResourceCPU])
	assert.Equal(t, model.CPUAmountFromCores((*podMinCPUMillicores/1000)/2), recommendedResources["container-2"].Target[model.ResourceCPU])
	assert.Equal(t, model.MemoryAmountFromBytes((*podMinMemoryMb*1024*1024)/2), recommendedResources["container-1"].Target[model.ResourceMemory])
	assert.Equal(t, model.MemoryAmountFromBytes((*podMinMemoryMb*1024*1024)/2), recommendedResources["container-2"].Target[model.ResourceMemory])
}

func TestControlledResourcesFiltered(t *testing.T) {
	constEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.001),
		model.ResourceMemory: model.MemoryAmountFromBytes(1e6),
	})
	recommender := podResourceRecommender{
		constEstimator,
		constEstimator,
		constEstimator}

	containerName := "container-1"
	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		containerName: &model.AggregateContainerState{
			ControlledResources: &[]model.ResourceName{model.ResourceMemory},
		},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap, len(containerNameToAggregateStateMap))
	assert.Contains(t, recommendedResources[containerName].Target, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].LowerBound, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].UpperBound, model.ResourceMemory)
	assert.NotContains(t, recommendedResources[containerName].Target, model.ResourceCPU)
	assert.NotContains(t, recommendedResources[containerName].LowerBound, model.ResourceCPU)
	assert.NotContains(t, recommendedResources[containerName].UpperBound, model.ResourceCPU)
}

func TestControlledResourcesFilteredDefault(t *testing.T) {
	constEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.001),
		model.ResourceMemory: model.MemoryAmountFromBytes(1e6),
	})
	recommender := podResourceRecommender{
		constEstimator,
		constEstimator,
		constEstimator}

	containerName := "container-1"
	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		containerName: &model.AggregateContainerState{
			ControlledResources: &[]model.ResourceName{model.ResourceMemory, model.ResourceCPU},
		},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap, len(containerNameToAggregateStateMap))
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
