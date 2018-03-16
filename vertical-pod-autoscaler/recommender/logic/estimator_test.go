/*
Copyright 2017 The Kubernetes Authors.

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
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/util"
)

var (
	anyTime = time.Unix(0, 0)
)

// Verifies that the PercentileEstimator returns requested percentiles of CPU
// and memory peaks distributions.
func TestPercentileEstimator(t *testing.T) {
	// Create a sample CPU histogram.
	cpuHistogram := util.NewHistogram(model.CPUHistogramOptions)
	cpuHistogram.AddSample(1.0, 1.0, anyTime)
	cpuHistogram.AddSample(2.0, 1.0, anyTime)
	cpuHistogram.AddSample(3.0, 1.0, anyTime)
	// Create a sample memory histogram.
	memoryPeaksHistogram := util.NewHistogram(model.MemoryHistogramOptions)
	memoryPeaksHistogram.AddSample(1e9, 1.0, anyTime)
	memoryPeaksHistogram.AddSample(2e9, 1.0, anyTime)
	memoryPeaksHistogram.AddSample(3e9, 1.0, anyTime)
	// Create an estimator.
	CPUPercentile := 0.2
	MemoryPercentile := 0.5
	estimator := NewPercentileEstimator(CPUPercentile, MemoryPercentile)

	resourceEstimation := estimator.GetResourceEstimation(
		&AggregateContainerState{
			aggregateCPUUsage:    cpuHistogram,
			aggregateMemoryPeaks: memoryPeaksHistogram,
		})

	assert.InEpsilon(t, 1000, int(resourceEstimation[model.ResourceCPU]), model.HistogramRelativeError)
	assert.InEpsilon(t, 2e9, int(resourceEstimation[model.ResourceMemory]), model.HistogramRelativeError)
}

// Verifies that the confidenceMultiplier calculates the internal
// confidence based on the amount of historical samples and scales the resources
// returned by the base estimator according to the formula, using the calculated
// confidence.
func TestConfidenceMultiplier(t *testing.T) {
	baseEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(3.14),
		model.ResourceMemory: model.MemoryAmountFromBytes(3.14e9),
	})
	testedEstimator := &confidenceMultiplier{0.1, 2.0, baseEstimator}

	container := model.NewContainerState()
	// Add 9 CPU samples at the frequency of 1/(2 mins).
	timestamp := anyTime
	for i := 1; i <= 9; i++ {
		container.AddSample(&model.ContainerUsageSample{
			timestamp, model.CPUAmountFromCores(1.0), model.ResourceCPU})
		timestamp = timestamp.Add(time.Minute * 2)
	}
	s := newAggregateContainerState()
	s.mergeContainerState(container)

	// Expected confidence = 9/(60*24) = 0.00625.
	assert.Equal(t, 0.00625, getConfidence(s))
	// Expected CPU estimation = 3.14 * (1 + 1/confidence)^exponent =
	// 3.14 * (1 + 0.1/0.00625)^2 = 907.46.
	resourceEstimation := testedEstimator.GetResourceEstimation(s)
	assert.Equal(t, 907.46, model.CoresFromCPUAmount(resourceEstimation[model.ResourceCPU]))
}

// Verifies that the confidenceMultiplier works for the case of no
// history. This corresponds to the multiplier of +INF or 0 (depending on the
// sign of the exponent).
func TestConfidenceMultiplierNoHistory(t *testing.T) {
	baseEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(3.14),
		model.ResourceMemory: model.MemoryAmountFromBytes(3.14e9),
	})
	testedEstimator1 := &confidenceMultiplier{1.0, 1.0, baseEstimator}
	testedEstimator2 := &confidenceMultiplier{1.0, -1.0, baseEstimator}
	s := newAggregateContainerState()
	// Expect testedEstimator1 to return the maximum possible resource amount.
	assert.Equal(t, model.ResourceAmount(1e14),
		testedEstimator1.GetResourceEstimation(s)[model.ResourceCPU])
	// Expect testedEstimator2 to return zero.
	assert.Equal(t, model.ResourceAmount(0),
		testedEstimator2.GetResourceEstimation(s)[model.ResourceCPU])
}
