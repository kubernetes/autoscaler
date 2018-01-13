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
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
)

// ResourceEstimator is a function from AggregateContainerState to
// model.Resources, e.g. a prediction of resources needed by a group of
// containers.
type ResourceEstimator interface {
	GetResourceEstimation(s *AggregateContainerState) model.Resources
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

// NewConstEstimator returns a new constEstimator with given resources.
func NewConstEstimator(resources model.Resources) ResourceEstimator {
	return &constEstimator{resources}
}

// NewPercentileEstimator returns a new percentileEstimator that uses provided percentiles.
func NewPercentileEstimator(cpuPercentile float64, memoryPercentile float64) ResourceEstimator {
	return &percentileEstimator{cpuPercentile, memoryPercentile}
}

// Returns a constant amount of resources.
func (e *constEstimator) GetResourceEstimation(s *AggregateContainerState) model.Resources {
	return e.resources
}

// Returns specific percentiles of CPU and memory peaks distributions.
func (e *percentileEstimator) GetResourceEstimation(s *AggregateContainerState) model.Resources {
	return model.Resources{
		model.ResourceCPU: model.CPUAmountFromCores(
			s.aggregateCPUUsage.Percentile(e.cpuPercentile)),
		model.ResourceMemory: model.MemoryAmountFromBytes(
			s.aggregateMemoryPeaks.Percentile(e.memoryPercentile)),
	}
}
