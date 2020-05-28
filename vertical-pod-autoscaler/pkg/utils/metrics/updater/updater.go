/*
Copyright 2018 The Kubernetes Authors.

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

// Package updater (aka metrics_updater) - code for metrics of VPA Updater
package updater

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
)

const (
	metricsNamespace = metrics.TopMetricsNamespace + "updater"
)

// SizeBasedGauge is a wrapper for incrementally recording values indexed by log2(VPA size)
type SizeBasedGauge struct {
	values [metrics.MaxVpaSizeLog]int
	gauge  *prometheus.GaugeVec
}

var (
	controlledCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "controlled_pods_total",
			Help:      "Number of Pods controlled by VPA updater.",
		}, []string{"vpa_size_log2"},
	)

	evictableCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "evictable_pods_total",
			Help:      "Number of Pods matching evicition criteria.",
		}, []string{"vpa_size_log2"},
	)

	evictedCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "evicted_pods_total",
			Help:      "Number of Pods evicted by Updater to apply a new recommendation.",
		}, []string{"vpa_size_log2"},
	)

	vpasWithEvictablePodsCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "vpas_with_evictable_pods_total",
			Help:      "Number of VPA objects with at least one Pod matching evicition criteria.",
		}, []string{"vpa_size_log2"},
	)

	vpasWithEvictedPodsCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "vpas_with_evicted_pods_total",
			Help:      "Number of VPA objects with at least one evicted Pod.",
		}, []string{"vpa_size_log2"},
	)

	functionLatency = metrics.CreateExecutionTimeMetric(metricsNamespace,
		"Time spent in various parts of VPA Updater main loop.")
)

// Register initializes all metrics for VPA Updater
func Register() {
	prometheus.MustRegister(controlledCount, evictableCount, evictedCount, vpasWithEvictablePodsCount, vpasWithEvictedPodsCount, functionLatency)
}

// NewExecutionTimer provides a timer for Updater's RunOnce execution
func NewExecutionTimer() *metrics.ExecutionTimer {
	return metrics.NewExecutionTimer(functionLatency)
}

// newSizeBasedGauge provides a wrapper for counting items in a loop
func newSizeBasedGauge(gauge *prometheus.GaugeVec) *SizeBasedGauge {
	return &SizeBasedGauge{
		values: [metrics.MaxVpaSizeLog]int{},
		gauge:  gauge,
	}
}

// NewControlledPodsCounter returns a wrapper for counting Pods controlled by Updater
func NewControlledPodsCounter() *SizeBasedGauge {
	return newSizeBasedGauge(controlledCount)
}

// NewEvictablePodsCounter returns a wrapper for counting Pods which are matching eviction criteria
func NewEvictablePodsCounter() *SizeBasedGauge {
	return newSizeBasedGauge(evictableCount)
}

// NewVpasWithEvictablePodsCounter returns a wrapper for counting VPA objects with Pods matching eviction criteria
func NewVpasWithEvictablePodsCounter() *SizeBasedGauge {
	return newSizeBasedGauge(vpasWithEvictablePodsCount)
}

// NewVpasWithEvictedPodsCounter returns a wrapper for counting VPA objects with evicted Pods
func NewVpasWithEvictedPodsCounter() *SizeBasedGauge {
	return newSizeBasedGauge(vpasWithEvictedPodsCount)
}

// AddEvictedPod increases the counter of pods evicted by Updater, by given VPA size
func AddEvictedPod(vpaSize int) {
	log2 := metrics.GetVpaSizeLog2(vpaSize)
	evictedCount.WithLabelValues(strconv.Itoa(log2)).Inc()
}

// Add increases the counter for the given VPA size
func (g *SizeBasedGauge) Add(vpaSize int, value int) {
	log2 := metrics.GetVpaSizeLog2(vpaSize)
	g.values[log2] += value
}

// Observe stores the recorded values into metrics object associated with the wrapper
func (g *SizeBasedGauge) Observe() {
	for log2, value := range g.values {
		g.gauge.WithLabelValues(strconv.Itoa(log2)).Set(float64(value))
	}
}
