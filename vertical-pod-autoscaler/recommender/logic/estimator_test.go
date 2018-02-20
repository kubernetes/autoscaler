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
	anyTime = time.Time{}
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
		&AggregateContainerState{cpuHistogram, memoryPeaksHistogram})

	assert.InEpsilon(t, 1000, int(resourceEstimation[model.ResourceCPU]), model.HistogramRelativeError)
	assert.InEpsilon(t, 2e9, int(resourceEstimation[model.ResourceMemory]), model.HistogramRelativeError)
}
