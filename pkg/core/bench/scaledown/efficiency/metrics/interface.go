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

package metrics

import (
	"testing"
	"time"

	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/status"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/framework"
)

// ClusterState holds cluster state scaleDownMetric use for computing the values.
type ClusterState struct {
	scaledDownNodes []*status.ScaleDownNode
	nodeInfos       []*framework.NodeInfo
	currentTime     time.Time
}

// NewClusterState returns a new instance of ClusterState.
func NewClusterState(scaledDownNodes []*status.ScaleDownNode, nodeInfos []*framework.NodeInfo, currentTime time.Time) ClusterState {
	return ClusterState{
		scaledDownNodes: scaledDownNodes,
		nodeInfos:       nodeInfos,
		currentTime:     currentTime,
	}
}

// scaleDownMetric is the unified interface all metrics implement.
type scaleDownMetric interface {
	// Name returns name of the scaleDownMetric.
	Name() string
	// Compute calculates single cluster-level metric value using clusterState.
	Compute(clusterState ClusterState) (float64, error)
	// Summarize processes metric recorded timeline and derives final benchmarking values returned in a map.
	Summarize(metricValues []float64) map[string]float64
}

// Tracker orchestrates the computation, storing, and reporting of multiple scaleDownMetric.
type Tracker struct {
	metrics         []scaleDownMetric
	metricsTimeline map[string][]float64
}

// NewTracker creates a new instance of Tracker with given list of scaleDownMetrics.
func NewTracker(metrics ...scaleDownMetric) *Tracker {
	return &Tracker{
		metrics:         metrics,
		metricsTimeline: make(map[string][]float64),
	}
}

// ReportToBenchmark registers final benchmark values to benchmark runner across a set of configured scaleDownMetrics.
func (mt *Tracker) ReportToBenchmark(b *testing.B) {
	for _, m := range mt.metrics {
		metricValues := mt.metricsTimeline[m.Name()]
		benchVals := m.Summarize(metricValues)
		for name, val := range benchVals {
			b.ReportMetric(val, name)
		}
	}
}

// ComputeMetrics computes metric value across a set of configured scaleDownMetrics.
// Watch out: If a metric fails to be computed, final metricsTimeline for each configured scaleDownMetric will diverge in length.
func (mt *Tracker) ComputeMetrics(clusterState ClusterState, t testing.TB) {
	for _, m := range mt.metrics {
		val, err := m.Compute(clusterState)
		if err != nil {
			t.Errorf("failed to compute metric %s, err %v", m.Name(), err)
			continue
		}
		mt.metricsTimeline[m.Name()] = append(mt.metricsTimeline[m.Name()], val)
	}
}
