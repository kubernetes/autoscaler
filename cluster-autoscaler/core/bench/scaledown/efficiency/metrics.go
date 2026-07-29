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

package efficiency

import (
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

type metricsDependencies struct {
	scaledDownNodes []*status.ScaleDownNode
	nodeInfos       []*framework.NodeInfo
	currentTime     time.Time
}

// scaleDownMetric is the unified interface all metrics implement.
type scaleDownMetric interface {
	// Name returns name of the scaleDownMetric.
	Name() string
	// Compute calculates single cluster-level metric value using metricsDependencies.
	Compute(dependencies metricsDependencies) (float64, error)
	// Summarize processes metric recorded timeline and derives final benchmarking values returned in a map.
	Summarize(metricValues []float64) map[string]float64
}

// metricsTracker orchestrates the computation, storing, and reporting of multiple scaleDownMetric.
type metricsTracker struct {
	metrics         []scaleDownMetric
	metricsTimeline map[string][]float64
}

// NewMetricsTracker creates a new instance of metricsTracker with given list of scaleDownMetrics.
func NewMetricsTracker(metrics ...scaleDownMetric) *metricsTracker {
	return &metricsTracker{
		metrics:         metrics,
		metricsTimeline: make(map[string][]float64),
	}
}

// ReportToBenchmark registers final benchmark values to benchmark runner across a set of configured scaleDownMetrics.
func (mt *metricsTracker) ReportToBenchmark(b *testing.B) {
	for _, m := range mt.metrics {
		metricValues := mt.metricsTimeline[m.Name()]
		benchVals := m.Summarize(metricValues)
		for name, val := range benchVals {
			b.ReportMetric(val, name)
		}
	}
}

// ComputeMetrics computes metric value across a set of configured scaleDownMetrics.
func (mt *metricsTracker) ComputeMetrics(dependencies metricsDependencies, t testing.TB) {
	for _, m := range mt.metrics {
		val, err := m.Compute(dependencies)
		if err != nil {
			t.Errorf("failed to compute metric %s, err %v", m.Name(), err)
			continue
		}
		mt.metricsTimeline[m.Name()] = append(mt.metricsTimeline[m.Name()], val)
	}
}
