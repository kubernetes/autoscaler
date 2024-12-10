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
	"math"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// ResourceEstimator is a function from AggregateContainerState to
// model.Resources, e.g. a prediction of resources needed by a group of
// containers.
type ResourceEstimator interface {
	GetResourceEstimation(s *model.AggregateContainerState) model.Resources
}

// CPUEstimator predicts CPU resources needed by a container
type CPUEstimator interface {
	GetCPUEstimation(s *model.AggregateContainerState) model.ResourceAmount
}

// MemoryEstimator predicts memory resources needed by a container
type MemoryEstimator interface {
	GetMemoryEstimation(s *model.AggregateContainerState) model.ResourceAmount
}

// combinedEstimator is a ResourceEstimator that combines two estimators: one for CPU and one for memory.
type combinedEstimator struct {
	cpuEstimator    CPUEstimator
	memoryEstimator MemoryEstimator
}

type percentileCPUEstimator struct {
	percentile float64
}

type percentileMemoryEstimator struct {
	percentile float64
}

// margins

type cpuMarginEstimator struct {
	marginFraction float64
	baseEstimator  CPUEstimator
}

type memoryMarginEstimator struct {
	marginFraction float64
	baseEstimator  MemoryEstimator
}

type cpuConfidenceMultiplier struct {
	multiplier    float64
	exponent      float64
	baseEstimator CPUEstimator
}

type memoryConfidenceMultiplier struct {
	multiplier    float64
	exponent      float64
	baseEstimator MemoryEstimator
}

type cpuMinResourceEstimator struct {
	minResource   model.ResourceAmount
	baseEstimator CPUEstimator
}

type memoryMinResourceEstimator struct {
	minResource   model.ResourceAmount
	baseEstimator MemoryEstimator
}

// NewCombinedEstimator returns a new combinedEstimator that uses provided estimators.
func NewCombinedEstimator(cpuEstimator CPUEstimator, memoryEstimator MemoryEstimator) ResourceEstimator {
	return &combinedEstimator{cpuEstimator, memoryEstimator}
}

// NewPercentileCPUEstimator returns a new percentileCPUEstimator that uses provided percentile.
func NewPercentileCPUEstimator(percentile float64) CPUEstimator {
	return &percentileCPUEstimator{percentile}
}

// NewPercentileMemoryEstimator returns a new percentileMemoryEstimator that uses provided percentile.
func NewPercentileMemoryEstimator(percentile float64) MemoryEstimator {
	return &percentileMemoryEstimator{percentile}
}

// NewMemoryEstimator returns a new percentileMemoryEstimator that uses provided percentile.
func NewMemoryEstimator(percentile float64) MemoryEstimator {
	return &percentileMemoryEstimator{percentile}
}

// GetCPUEstimation returns the CPU estimation for the given AggregateContainerState.
func (e *cpuMarginEstimator) GetCPUEstimation(s *model.AggregateContainerState) model.ResourceAmount {
	base := e.baseEstimator.GetCPUEstimation(s)
	margin := model.ScaleResource(base, e.marginFraction)
	return base + margin
}

// GetMemoryEstimation returns the memory estimation for the given AggregateContainerState.
func (e *memoryMarginEstimator) GetMemoryEstimation(s *model.AggregateContainerState) model.ResourceAmount {
	base := e.baseEstimator.GetMemoryEstimation(s)
	margin := model.ScaleResource(base, e.marginFraction)
	return base + margin
}

// WithCPUMargin returns a CPUEstimator that adds a margin to the base estimator.
func WithCPUMargin(marginFraction float64, baseEstimator CPUEstimator) CPUEstimator {
	return &cpuMarginEstimator{marginFraction: marginFraction, baseEstimator: baseEstimator}
}

// WithMemoryMargin returns a MemoryEstimator that adds a margin to the base estimator.
func WithMemoryMargin(marginFraction float64, baseEstimator MemoryEstimator) MemoryEstimator {
	return &memoryMarginEstimator{marginFraction: marginFraction, baseEstimator: baseEstimator}
}

// WithCPUConfidenceMultiplier return a CPUEstimator estimator
func WithCPUConfidenceMultiplier(multiplier, exponent float64, baseEstimator CPUEstimator) CPUEstimator {
	return &cpuConfidenceMultiplier{
		multiplier:    multiplier,
		exponent:      exponent,
		baseEstimator: baseEstimator,
	}
}

// WithMemoryConfidenceMultiplier returns a MemoryEstimator that scales the
func WithMemoryConfidenceMultiplier(multiplier, exponent float64, baseEstimator MemoryEstimator) MemoryEstimator {
	return &memoryConfidenceMultiplier{
		multiplier:    multiplier,
		exponent:      exponent,
		baseEstimator: baseEstimator,
	}
}

func (e *percentileCPUEstimator) GetCPUEstimation(s *model.AggregateContainerState) model.ResourceAmount {
	return model.CPUAmountFromCores(s.AggregateCPUUsage.Percentile(e.percentile))
}

func (e *percentileMemoryEstimator) GetMemoryEstimation(s *model.AggregateContainerState) model.ResourceAmount {
	return model.MemoryAmountFromBytes(s.AggregateMemoryPeaks.Percentile(e.percentile))
}

// Returns resources computed by the underlying estimators, scaled based on the
// confidence metric, which depends on the amount of available historical data.
// Each resource is transformed as follows:
//
//	scaledResource = originalResource * (1 + 1/confidence)^exponent.
//
// This can be used to widen or narrow the gap between the lower and upper bound
// estimators depending on how much input data is available to the estimators.
func (c *combinedEstimator) GetResourceEstimation(s *model.AggregateContainerState) model.Resources {
	return model.Resources{
		model.ResourceCPU:    c.cpuEstimator.GetCPUEstimation(s),
		model.ResourceMemory: c.memoryEstimator.GetMemoryEstimation(s),
	}
}

// Returns a non-negative real number that heuristically measures how much
// confidence the history aggregated in the AggregateContainerState provides.
// For a workload producing a steady stream of samples over N days at the rate
// of 1 sample per minute, this metric is equal to N.
// This implementation is a very simple heuristic which looks at the total count
// of samples and the time between the first and the last sample.
func getConfidence(s *model.AggregateContainerState) float64 {
	// Distance between the first and the last observed sample time, measured in days.
	lifespanInDays := float64(s.LastSampleStart.Sub(s.FirstSampleStart)) / float64(time.Hour*24)
	// Total count of samples normalized such that it equals the number of days for
	// frequency of 1 sample/minute.
	samplesAmount := float64(s.TotalSamplesCount) / (60 * 24)
	return math.Min(lifespanInDays, samplesAmount)
}

func (e *cpuConfidenceMultiplier) GetCPUEstimation(s *model.AggregateContainerState) model.ResourceAmount {
	confidence := getConfidence(s)
	base := e.baseEstimator.GetCPUEstimation(s)
	return model.ScaleResource(base, math.Pow(1.+e.multiplier/confidence, e.exponent))
}

func (e *memoryConfidenceMultiplier) GetMemoryEstimation(s *model.AggregateContainerState) model.ResourceAmount {
	confidence := getConfidence(s)
	base := e.baseEstimator.GetMemoryEstimation(s)
	return model.ScaleResource(base, math.Pow(1.+e.multiplier/confidence, e.exponent))
}

// WithCPUMinResource returns a CPUEstimator that returns at least minResource
func WithCPUMinResource(minResource model.ResourceAmount, baseEstimator CPUEstimator) CPUEstimator {
	return &cpuMinResourceEstimator{minResource, baseEstimator}
}

// WithMemoryMinResource returns a MemoryEstimator that returns at least minResource
func WithMemoryMinResource(minResource model.ResourceAmount, baseEstimator MemoryEstimator) MemoryEstimator {
	return &memoryMinResourceEstimator{minResource, baseEstimator}
}

func (e *cpuMinResourceEstimator) GetCPUEstimation(s *model.AggregateContainerState) model.ResourceAmount {
	return model.ResourceAmountMax(e.baseEstimator.GetCPUEstimation(s), e.minResource)
}

func (e *memoryMinResourceEstimator) GetMemoryEstimation(s *model.AggregateContainerState) model.ResourceAmount {
	return model.ResourceAmountMax(e.baseEstimator.GetMemoryEstimation(s), e.minResource)
}

// NewConstMemoryEstimator returns a Memory estimator that always returns the same value
func NewConstMemoryEstimator(memory model.ResourceAmount) MemoryEstimator {
	return &constMemoryEstimator{memory}
}

type constCPUEstimator struct {
	value model.ResourceAmount
}

type constMemoryEstimator struct {
	value model.ResourceAmount
}

func (e *constCPUEstimator) GetCPUEstimation(_ *model.AggregateContainerState) model.ResourceAmount {
	return e.value
}

func (e *constMemoryEstimator) GetMemoryEstimation(_ *model.AggregateContainerState) model.ResourceAmount {
	return e.value
}

// NewConstCPUEstimator returns a CPU estimator that always returns the same value
func NewConstCPUEstimator(cpu model.ResourceAmount) CPUEstimator {
	return &constCPUEstimator{cpu}
}
