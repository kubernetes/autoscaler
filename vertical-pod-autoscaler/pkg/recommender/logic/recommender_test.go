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
	"time"

	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

func TestMinResourcesApplied(t *testing.T) {
	minCPUMillicores := 25.
	minMemoryMb := 250.
	constCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(0.001))
	constMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(1e6))

	recommender := podResourceRecommender{
		targetCPU:        constCPUEstimator,
		targetMemory:     constMemoryEstimator,
		lowerBoundCPU:    constCPUEstimator,
		lowerBoundMemory: constMemoryEstimator,
		upperBoundCPU:    constCPUEstimator,
		upperBoundMemory: constMemoryEstimator,
		minCPUMillicores: minCPUMillicores,
		minMemoryMb:      minMemoryMb,
	}

	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		"container-1": &model.AggregateContainerState{},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
	assert.Equal(t, model.CPUAmountFromCores(minCPUMillicores/1000), recommendedResources["container-1"].Target[model.ResourceCPU])
	assert.Equal(t, model.MemoryAmountFromBytes(minMemoryMb*1024*1024), recommendedResources["container-1"].Target[model.ResourceMemory])
}

// Verifies that CreatePodResourceRecommender wires each per-container percentile
// override to its own estimator: overriding one of the six (lower/target/upper
// for cpu/memory) moves only that bound and leaves the other five untouched.
func TestCreatePodResourceRecommenderPerContainerPercentiles(t *testing.T) {
	aggregations := model.GetAggregationsConfig()
	cpuHistogram := util.NewHistogram(aggregations.CPUHistogramOptions)
	memHistogram := util.NewHistogram(aggregations.MemoryHistogramOptions)
	for i := 1; i <= 100; i++ {
		cpuHistogram.AddSample(float64(i), 1.0, anyTime)
		memHistogram.AddSample(float64(i)*1e9, 1.0, anyTime)
	}
	// A week of history keeps the lower/upper bound confidence multipliers
	// finite, and identical across the states compared below.
	newState := func() *model.AggregateContainerState {
		return &model.AggregateContainerState{
			AggregateCPUUsage:    cpuHistogram,
			AggregateMemoryPeaks: memHistogram,
			FirstSampleStart:     anyTime,
			LastSampleStart:      anyTime.Add(7 * 24 * time.Hour),
			TotalSamplesCount:    7 * 24 * 60,
		}
	}

	recommender := CreatePodResourceRecommender(RecommendationConfig{
		SafetyMarginFraction:       0.15,
		TargetCPUPercentile:        0.2,
		LowerBoundCPUPercentile:    0.2,
		UpperBoundCPUPercentile:    0.2,
		TargetMemoryPercentile:     0.2,
		LowerBoundMemoryPercentile: 0.2,
		UpperBoundMemoryPercentile: 0.2,
		ConfidenceIntervalCPU:      24 * time.Hour,
		ConfidenceIntervalMemory:   24 * time.Hour,
	})
	const container = "container"
	recommend := func(s *model.AggregateContainerState) RecommendedContainerResources {
		return recommender.GetRecommendedPodResources(model.ContainerNameToAggregateStateMap{container: s})[container]
	}
	baseline := recommend(newState())

	slots := []struct {
		name     string
		override func(*model.AggregateContainerState)
		pick     func(RecommendedContainerResources) model.ResourceAmount
	}{
		{"cpu lower bound", func(s *model.AggregateContainerState) { s.LowerBoundCPUPercentile = 0.9 }, func(r RecommendedContainerResources) model.ResourceAmount { return r.LowerBound[model.ResourceCPU] }},
		{"cpu target", func(s *model.AggregateContainerState) { s.TargetCPUPercentile = 0.9 }, func(r RecommendedContainerResources) model.ResourceAmount { return r.Target[model.ResourceCPU] }},
		{"cpu upper bound", func(s *model.AggregateContainerState) { s.UpperBoundCPUPercentile = 0.9 }, func(r RecommendedContainerResources) model.ResourceAmount { return r.UpperBound[model.ResourceCPU] }},
		{"memory lower bound", func(s *model.AggregateContainerState) { s.LowerBoundMemoryPercentile = 0.9 }, func(r RecommendedContainerResources) model.ResourceAmount { return r.LowerBound[model.ResourceMemory] }},
		{"memory target", func(s *model.AggregateContainerState) { s.TargetMemoryPercentile = 0.9 }, func(r RecommendedContainerResources) model.ResourceAmount { return r.Target[model.ResourceMemory] }},
		{"memory upper bound", func(s *model.AggregateContainerState) { s.UpperBoundMemoryPercentile = 0.9 }, func(r RecommendedContainerResources) model.ResourceAmount { return r.UpperBound[model.ResourceMemory] }},
	}
	for i, tc := range slots {
		t.Run(tc.name, func(t *testing.T) {
			s := newState()
			tc.override(s)
			got := recommend(s)
			assert.Greater(t, tc.pick(got), tc.pick(baseline), "%s should follow its override", tc.name)
			for j, other := range slots {
				if j != i {
					assert.Equal(t, other.pick(baseline), other.pick(got), "%s should be unchanged", other.name)
				}
			}
		})
	}
}

func TestMinResourcesSplitAcrossContainers(t *testing.T) {
	minCPUMillicores := 25.
	minMemoryMb := 250.
	constCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(0.001))
	constMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(1e6))

	recommender := podResourceRecommender{
		targetCPU:        constCPUEstimator,
		targetMemory:     constMemoryEstimator,
		lowerBoundCPU:    constCPUEstimator,
		lowerBoundMemory: constMemoryEstimator,
		upperBoundCPU:    constCPUEstimator,
		upperBoundMemory: constMemoryEstimator,
		minCPUMillicores: minCPUMillicores,
		minMemoryMb:      minMemoryMb,
	}

	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		"container-1": &model.AggregateContainerState{},
		"container-2": &model.AggregateContainerState{},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
	assert.Equal(t, model.CPUAmountFromCores((minCPUMillicores/1000)/2), recommendedResources["container-1"].Target[model.ResourceCPU])
	assert.Equal(t, model.CPUAmountFromCores((minCPUMillicores/1000)/2), recommendedResources["container-2"].Target[model.ResourceCPU])
	assert.Equal(t, model.MemoryAmountFromBytes((minMemoryMb*1024*1024)/2), recommendedResources["container-1"].Target[model.ResourceMemory])
	assert.Equal(t, model.MemoryAmountFromBytes((minMemoryMb*1024*1024)/2), recommendedResources["container-2"].Target[model.ResourceMemory])
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
			outRecommendations := MapToListOfRecommendedContainerResources(tc.resources, RecommendationFormat{RoundCPUMillicores: 1, RoundMemoryBytes: 1})
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
