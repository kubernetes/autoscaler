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

// TODO: Split the estimator to have a separate estimator object for CPU and memory.

// ResourceEstimator is a function from AggregateContainerState to
// model.Resources, e.g. a prediction of resources needed by a group of
// containers.
type ResourceEstimator interface {
	GetResourceEstimation(s *model.AggregateContainerState) model.Resources
}

// Implementation of ResourceEstimator that returns constant amount of
// resources. This can be used as by a fake recommender for test purposes.
type constEstimator struct {
	resources model.Resources
}

// Simple implementation of the ResourceEstimator interface. It returns specific
// percentiles of CPU usage distribution and memory peaks distribution.
type percentileEstimator struct {
	cpuPercentile    float64
	memoryPercentile float64
}

type marginEstimator struct {
	marginFraction float64
	baseEstimator  ResourceEstimator
}

type minResourcesEstimator struct {
	minResources  model.Resources
	baseEstimator ResourceEstimator
}

type confidenceMultiplier struct {
	multiplier    float64
	exponent      float64
	baseEstimator ResourceEstimator
}

// NewConstEstimator returns a new constEstimator with given resources.
func NewConstEstimator(resources model.Resources) ResourceEstimator {
	return &constEstimator{resources}
}

// NewPercentileEstimator returns a new percentileEstimator that uses provided percentiles.
func NewPercentileEstimator(cpuPercentile float64, memoryPercentile float64) ResourceEstimator {
	return &percentileEstimator{cpuPercentile, memoryPercentile}
}

// WithMargin returns a given ResourceEstimator with margin applied.
// The returned resources are equal to the original resources plus (originalResource * marginFraction)
func WithMargin(marginFraction float64, baseEstimator ResourceEstimator) ResourceEstimator {
	return &marginEstimator{marginFraction, baseEstimator}
}

// WithMinResources returns a given ResourceEstimator with minResources applied.
// The returned resources are equal to the max(original resources, minResources)
func WithMinResources(minResources model.Resources, baseEstimator ResourceEstimator) ResourceEstimator {
	return &minResourcesEstimator{minResources, baseEstimator}
}

// WithConfidenceMultiplier returns a given ResourceEstimator with confidenceMultiplier applied.
func WithConfidenceMultiplier(multiplier, exponent float64, baseEstimator ResourceEstimator) ResourceEstimator {
	return &confidenceMultiplier{multiplier, exponent, baseEstimator}
}

// Returns a constant amount of resources.
func (e *constEstimator) GetResourceEstimation(s *model.AggregateContainerState) model.Resources {
	return e.resources
}

// Returns specific percentiles of CPU and memory peaks distributions.
func (e *percentileEstimator) GetResourceEstimation(s *model.AggregateContainerState) model.Resources {
	return model.Resources{
		model.ResourceCPU: model.CPUAmountFromCores(
			s.AggregateCPUUsage.Percentile(e.cpuPercentile)),
		model.ResourceMemory: model.MemoryAmountFromBytes(
			s.AggregateMemoryPeaks.Percentile(e.memoryPercentile)),
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

// Returns resources computed by the underlying estimator, scaled based on the
// confidence metric, which depends on the amount of available historical data.
// Each resource is transformed as follows:
//     scaledResource = originalResource * (1 + 1/confidence)^exponent.
// This can be used to widen or narrow the gap between the lower and upper bound
// estimators depending on how much input data is available to the estimators.
func (e *confidenceMultiplier) GetResourceEstimation(s *model.AggregateContainerState) model.Resources {
	confidence := getConfidence(s)
	originalResources := e.baseEstimator.GetResourceEstimation(s)
	scaledResources := make(model.Resources)
	for resource, resourceAmount := range originalResources {
		scaledResources[resource] = model.ScaleResource(
			resourceAmount, math.Pow(1.+e.multiplier/confidence, e.exponent))
	}
	return scaledResources
}

func (e *marginEstimator) GetResourceEstimation(s *model.AggregateContainerState) model.Resources {
	originalResources := e.baseEstimator.GetResourceEstimation(s)
	newResources := make(model.Resources)
	for resource, resourceAmount := range originalResources {
		margin := model.ScaleResource(resourceAmount, e.marginFraction)
		newResources[resource] = originalResources[resource] + margin
	}
	return newResources
}

func (e *minResourcesEstimator) GetResourceEstimation(s *model.AggregateContainerState) model.Resources {
	originalResources := e.baseEstimator.GetResourceEstimation(s)
	newResources := make(model.Resources)
	for resource, resourceAmount := range originalResources {
		if resourceAmount < e.minResources[resource] {
			resourceAmount = e.minResources[resource]
		}
		newResources[resource] = resourceAmount
	}
	return newResources
}
