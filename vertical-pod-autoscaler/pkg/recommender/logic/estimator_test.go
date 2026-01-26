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

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/util"
)

var (
	anyTime                   = time.Unix(0, 0)
	defaultConfidenceInterval = time.Hour * 24
)

// Verifies that the PercentileEstimator returns requested percentiles of CPU
// and memory peaks distributions.
func TestPercentileEstimator(t *testing.T) {
	config := model.GetAggregationsConfig()
	// Create a sample CPU histogram.
	cpuHistogram := util.NewHistogram(config.CPUHistogramOptions)
	cpuHistogram.AddSample(1.0, 1.0, anyTime)
	cpuHistogram.AddSample(2.0, 1.0, anyTime)
	cpuHistogram.AddSample(3.0, 1.0, anyTime)
	// Create a sample memory histogram.
	memoryPeaksHistogram := util.NewHistogram(config.MemoryHistogramOptions)
	memoryPeaksHistogram.AddSample(1e9, 1.0, anyTime)
	memoryPeaksHistogram.AddSample(2e9, 1.0, anyTime)
	memoryPeaksHistogram.AddSample(3e9, 1.0, anyTime)
	// Create an estimator.
	cpuPercentile := 0.2
	memoryPercentile := 0.5
	cpuEstimator := NewPercentileCPUEstimator(cpuPercentile)
	memoryEstimator := NewPercentileMemoryEstimator(memoryPercentile)
	combinedEstimator := NewCombinedEstimator(cpuEstimator, memoryEstimator)

	resourceEstimation := combinedEstimator.GetResourceEstimation(
		&model.AggregateContainerState{
			AggregateCPUUsage:    cpuHistogram,
			AggregateMemoryPeaks: memoryPeaksHistogram,
		})
	maxRelativeError := 0.05 // Allow 5% relative error to account for histogram rounding.
	assert.InEpsilon(t, 1.0, model.CoresFromCPUAmount(resourceEstimation[model.ResourceCPU]), maxRelativeError)
	assert.InEpsilon(t, 2e9, model.BytesFromMemoryAmount(resourceEstimation[model.ResourceMemory]), maxRelativeError)
}

// Verifies that the confidenceMultiplier calculates the internal
// confidence based on the amount of historical samples and scales the resources
// returned by the base estimator according to the formula, using the calculated
// confidence.
func TestConfidenceMultiplier(t *testing.T) {
	baseCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(3.14))
	baseMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(3.14e9))
	testedCPU1 := WithCPUConfidenceMultiplier(1.0, 1.0, baseCPUEstimator, defaultConfidenceInterval)
	testedMemory1 := WithMemoryConfidenceMultiplier(1.0, 1.0, baseMemoryEstimator, defaultConfidenceInterval)
	testedEstimator1 := NewCombinedEstimator(testedCPU1, testedMemory1)

	testedCPU2 := WithCPUConfidenceMultiplier(1.0, -1.0, baseCPUEstimator, defaultConfidenceInterval)
	testedMemory2 := WithMemoryConfidenceMultiplier(1.0, -1.0, baseMemoryEstimator, defaultConfidenceInterval)
	testedEstimator2 := NewCombinedEstimator(testedCPU2, testedMemory2)

	s := model.NewAggregateContainerState()
	// Expect testedEstimator1 to return the maximum possible resource amount.
	assert.Equal(t, model.ResourceAmount(1e14),
		testedEstimator1.GetResourceEstimation(s)[model.ResourceCPU])
	// Expect testedEstimator2 to return zero.
	assert.Equal(t, model.ResourceAmount(0),
		testedEstimator2.GetResourceEstimation(s)[model.ResourceCPU])

	timestamp := anyTime
	testedCPU3 := WithCPUConfidenceMultiplier(0.1, 2.0, baseCPUEstimator, defaultConfidenceInterval)
	testedMemory3 := WithMemoryConfidenceMultiplier(0.1, 2.0, baseMemoryEstimator, defaultConfidenceInterval)
	testedEstimator3 := NewCombinedEstimator(testedCPU3, testedMemory3)

	for i := 1; i <= 9; i++ {
		s.AddSample(&model.ContainerUsageSample{
			MeasureStart: timestamp,
			Usage:        model.CPUAmountFromCores(1.0),
			Resource:     model.ResourceCPU,
		})
		timestamp = timestamp.Add(time.Minute * 2)
	}

	// Expected confidence = 9/(60*24) = 0.00625.
	assert.Equal(t, 0.00625, getConfidence(s, defaultConfidenceInterval))
	// Expected CPU estimation = 3.14 * (1 + multiplier/confidence)^exponent =
	// 3.14 * (1 + 0.1/0.00625)^2 = 907.46.
	// Expected Memory estimation =
	// 3140000000 * (1 + 0.1/0.00625)^2 = 907460000000
	resourceEstimation := testedEstimator3.GetResourceEstimation(s)
	assert.Equal(t, 907.46, model.CoresFromCPUAmount(resourceEstimation[model.ResourceCPU]))
	assert.Equal(t, 9.0746e+11, model.BytesFromMemoryAmount(resourceEstimation[model.ResourceMemory]))
}

// TestConfidenceMultiplierWithCustomInterval verifies that the confidenceMultiplier calculates the internal
// confidence based on the amount of historical samples and the provided ConfidenceInterval.
func TestConfidenceMultiplierWithCustomInterval(t *testing.T) {
	customConfidenceInterval := time.Minute * 18
	baseCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(3.14))
	baseMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(3.14e9))

	s := model.NewAggregateContainerState()

	timestamp := anyTime
	testedCPU4 := WithCPUConfidenceMultiplier(0.1, 2.0, baseCPUEstimator, customConfidenceInterval)
	testedMemory4 := WithMemoryConfidenceMultiplier(0.1, 2.0, baseMemoryEstimator, customConfidenceInterval)
	testedEstimator4 := NewCombinedEstimator(testedCPU4, testedMemory4)

	for i := 1; i <= 9; i++ {
		s.AddSample(&model.ContainerUsageSample{
			MeasureStart: timestamp,
			Usage:        model.CPUAmountFromCores(1.0),
			Resource:     model.ResourceCPU,
		})
		timestamp = timestamp.Add(time.Minute * 2)
	}

	// Expected confidence = 9/18 = 0.5
	assert.Equal(t, 0.5, getConfidence(s, customConfidenceInterval))
	// Expected CPU estimation = 3.14 * (1 + multiplier/confidence)^exponent =
	// 3.14 * (1 + 0.1/0.5)^2 = 4.5216
	// Expected Memory estimation =
	// 3140000000 * (1 + 0.1/0.5)^2 = 4521600000
	resourceEstimation := testedEstimator4.GetResourceEstimation(s)
	assert.Equal(t, 4.521, model.CoresFromCPUAmount(resourceEstimation[model.ResourceCPU]))
	assert.Equal(t, 4.5216e+9, model.BytesFromMemoryAmount(resourceEstimation[model.ResourceMemory]))
}

// Verifies that the confidenceMultiplier works for the case of no
// history. This corresponds to the multiplier of +INF or 0 (depending on the
// sign of the exponent).
func TestConfidenceMultiplierNoHistory(t *testing.T) {
	baseCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(3.14))
	baseMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(3.14e9))
	testedCPU := WithCPUConfidenceMultiplier(0.1, 2.0, baseCPUEstimator, defaultConfidenceInterval)
	testedMemory := WithMemoryConfidenceMultiplier(0.1, 2.0, baseMemoryEstimator, defaultConfidenceInterval)
	testedEstimator := NewCombinedEstimator(testedCPU, testedMemory)
	s := model.NewAggregateContainerState()
	// Expect testedEstimator to return the maximum possible resource amount.
	assert.Equal(t, model.ResourceAmount(1e14),
		testedEstimator.GetResourceEstimation(s)[model.ResourceCPU])
}

// Verifies that the MarginEstimator adds margin to the originally
// estimated resources.
func TestMarginEstimator(t *testing.T) {
	// Use 10% margin on top of the recommended resources.
	marginFraction := 0.1
	baseCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(3.14))
	baseMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(3.14e9))
	testedCPU := WithCPUMargin(marginFraction, baseCPUEstimator)
	testedMemory := WithMemoryMargin(marginFraction, baseMemoryEstimator)
	testedEstimator := NewCombinedEstimator(testedCPU, testedMemory)
	s := model.NewAggregateContainerState()
	resourceEstimation := testedEstimator.GetResourceEstimation(s)
	assert.Equal(t, 3.14*1.1, model.CoresFromCPUAmount(resourceEstimation[model.ResourceCPU]))
	assert.Equal(t, 3.14e9*1.1, model.BytesFromMemoryAmount(resourceEstimation[model.ResourceMemory]))
}

// Verifies that the MinResourcesEstimator returns at least MinResources.
func TestMinResourcesEstimator(t *testing.T) {
	constCPUEstimator := NewConstCPUEstimator(model.CPUAmountFromCores(3.14))
	minCPU := model.CPUAmountFromCores(0.2)
	cpuEstimator := WithCPUMinResource(minCPU, constCPUEstimator)
	s := model.NewAggregateContainerState()
	cpuEstimation := cpuEstimator.GetCPUEstimation(s)
	assert.Equal(t, 3.14, model.CoresFromCPUAmount(cpuEstimation))

	constMemoryEstimator := NewConstMemoryEstimator(model.MemoryAmountFromBytes(4e8))
	minMemory := model.MemoryAmountFromBytes(2e7)
	memoryEstimator := WithMemoryMinResource(minMemory, constMemoryEstimator)
	memoryEstimation := memoryEstimator.GetMemoryEstimation(s)
	assert.Equal(t, 4e8, model.BytesFromMemoryAmount(memoryEstimation))
}
