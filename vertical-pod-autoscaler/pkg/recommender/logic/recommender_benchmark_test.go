/*
Copyright The Kubernetes Authors.

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
	"fmt"
	"testing"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/config"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// newBenchmarkRecommender returns a PodResourceRecommender wired the same way
// the production recommender is (see DefaultRecommenderConfig and
// routines.RecommenderFactory).
func newBenchmarkRecommender() PodResourceRecommender {
	defaults := config.DefaultRecommenderConfig()
	return CreatePodResourceRecommender(RecommendationConfig{
		SafetyMarginFraction:       defaults.SafetyMarginFraction,
		PodMinCPUMillicores:        defaults.PodMinCPUMillicores,
		PodMinMemoryMb:             defaults.PodMinMemoryMb,
		TargetCPUPercentile:        defaults.TargetCPUPercentile,
		LowerBoundCPUPercentile:    defaults.LowerBoundCPUPercentile,
		UpperBoundCPUPercentile:    defaults.UpperBoundCPUPercentile,
		ConfidenceIntervalCPU:      defaults.ConfidenceIntervalCPU,
		TargetMemoryPercentile:     defaults.TargetMemoryPercentile,
		LowerBoundMemoryPercentile: defaults.LowerBoundMemoryPercentile,
		UpperBoundMemoryPercentile: defaults.UpperBoundMemoryPercentile,
		ConfidenceIntervalMemory:   defaults.ConfidenceIntervalMemory,
	})
}

// newBenchmarkAggregateState returns an AggregateContainerState populated with
// sampleCount CPU and memory usage samples, one pair per minute, with usage
// varying deterministically so that the histograms have many non-empty buckets.
func newBenchmarkAggregateState(sampleCount int) *model.AggregateContainerState {
	state := model.NewAggregateContainerState()
	timestamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range sampleCount {
		state.AddSample(&model.ContainerUsageSample{
			MeasureStart: timestamp,
			Usage:        model.CPUAmountFromCores(0.1 + 0.005*float64(i%100)),
			Resource:     model.ResourceCPU,
		})
		state.AddSample(&model.ContainerUsageSample{
			MeasureStart: timestamp,
			Usage:        model.MemoryAmountFromBytes(float64(100+i%100) * 1024 * 1024),
			Resource:     model.ResourceMemory,
		})
		timestamp = timestamp.Add(time.Minute)
	}
	return state
}

// BenchmarkGetRecommendedPodResources measures the cost of computing a
// recommendation for a single VPA from already-aggregated container usage
// histograms. This is the per-VPA hot path of the recommender's main loop
// (UpdateVPAs).
func BenchmarkGetRecommendedPodResources(b *testing.B) {
	// One day of per-minute usage samples per container.
	const samplesPerContainer = 24 * 60
	for _, containerCount := range []int{1, 2, 10} {
		b.Run(fmt.Sprintf("containers=%d", containerCount), func(b *testing.B) {
			recommender := newBenchmarkRecommender()
			containerNameToAggregateStateMap := make(model.ContainerNameToAggregateStateMap, containerCount)
			for i := range containerCount {
				containerNameToAggregateStateMap[fmt.Sprintf("container-%d", i)] = newBenchmarkAggregateState(samplesPerContainer)
			}
			b.ReportAllocs()
			for b.Loop() {
				recommendation := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
				if len(recommendation) != containerCount {
					b.Fatalf("expected recommendations for %d containers, got %d", containerCount, len(recommendation))
				}
			}
		})
	}
}
