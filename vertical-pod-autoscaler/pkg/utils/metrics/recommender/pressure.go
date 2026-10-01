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

package recommender

import "github.com/prometheus/client_golang/prometheus"

var (
	pressureEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "pressure_events_total",
			Help:      "Memory pressure observations, by outcome.",
		}, []string{"result"},
	)

	pressureContainers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "pressure_containers",
			Help:      "Containers meeting both detection conditions in the latest window, by state.",
		}, []string{"state"},
	)

	pressureEstimateRatio = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "pressure_estimate_to_recommendation_ratio",
			Help:      "Accepted pressure estimate divided by the last memory recommendation.",
			Buckets:   []float64{0.5, 0.8, 1, 1.1, 1.25, 1.5, 2, 3, 5},
		},
	)

	pressureFetchDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "pressure_fetch_duration_seconds",
			Help:      "Latency of fetching one node's kubelet Summary, by result.",
			Buckets:   prometheus.ExponentialBuckets(0.005, 2, 12),
		}, []string{"result"},
	)

	pressureNodes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "pressure_nodes",
			Help:      "Eligible nodes, by state of the last fetch.",
		}, []string{"state"},
	)

	pressureNodeRevisit = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "pressure_node_revisit_seconds",
			Help:      "Time between consecutive successful fetches of the same node.",
			Buckets:   []float64{30, 45, 60, 75, 90, 105, 120, 150, 180, 240, 300, 600},
		},
	)
)

var (
	pressureEventResults    = []string{"accepted", "policy_changed", "window_mismatch", "unknown_container", "stale", "identity_changed", "rate_out_of_range", "resize_pending"}
	pressureContainerStates = []string{"injectable", "cooldown", "resize_in_progress", "resize_deferred", "resize_infeasible"}
	pressureNodeStates      = []string{"observed", "missing_psi", "failed", "stale", "undersampled", "unit_mismatch"}
)

// setStateGauge sets one series per state, so a state that stops occurring reads zero instead of
// keeping the count from the last window that had it.
func setStateGauge(g *prometheus.GaugeVec, states []string, byState map[string]int) {
	for _, s := range states {
		g.WithLabelValues(s).Set(float64(byState[s]))
	}
}

func registerPressure() {
	prometheus.MustRegister(pressureEvents, pressureContainers, pressureEstimateRatio, pressureFetchDuration, pressureNodes, pressureNodeRevisit)
	for _, r := range pressureEventResults {
		pressureEvents.WithLabelValues(r).Add(0)
	}
}

// RecordPressureEvent counts an observation outcome, one of pressureEventResults.
func RecordPressureEvent(result string) { pressureEvents.WithLabelValues(result).Inc() }

// SetPressureContainers sets the per-state container gauge.
func SetPressureContainers(byState map[string]int) {
	setStateGauge(pressureContainers, pressureContainerStates, byState)
}

// ObservePressureEstimateRatio records estimate / last recommendation.
func ObservePressureEstimateRatio(r float64) { pressureEstimateRatio.Observe(r) }

// ObservePressureFetch records one fetch.
func ObservePressureFetch(result string, seconds float64) {
	pressureFetchDuration.WithLabelValues(result).Observe(seconds)
}

// SetPressureNodes sets the per-state node gauge.
func SetPressureNodes(byState map[string]int) {
	setStateGauge(pressureNodes, pressureNodeStates, byState)
}

// ObservePressureNodeRevisit records the gap between successful fetches of a node.
func ObservePressureNodeRevisit(seconds float64) { pressureNodeRevisit.Observe(seconds) }
