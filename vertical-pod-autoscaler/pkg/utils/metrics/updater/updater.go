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

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
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

// UpdateModeAndSizeBasedGauge is a wrapper for incrementally recording values
// indexed by log2(VPA size) and update mode
type UpdateModeAndSizeBasedGauge struct {
	values [metrics.MaxVpaSizeLog]map[vpa_types.UpdateMode]int
	gauge  *prometheus.GaugeVec
}

var (
	controlledCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "controlled_pods_total",
			Help:      "Number of Pods controlled by VPA updater.",
		}, []string{"vpa_size_log2", "update_mode"},
	)

	evictableCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "evictable_pods_total",
			Help:      "Number of Pods matching evicition criteria.",
		}, []string{"vpa_size_log2", "update_mode"},
	)

	evictedCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "evicted_pods_total",
			Help:      "Number of Pods evicted by Updater to apply a new recommendation.",
		}, []string{"vpa_size_log2", "update_mode"},
	)

	vpasWithEvictablePodsCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "vpas_with_evictable_pods_total",
			Help:      "Number of VPA objects with at least one Pod matching evicition criteria.",
		}, []string{"vpa_size_log2", "update_mode"},
	)

	vpasWithEvictedPodsCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "vpas_with_evicted_pods_total",
			Help:      "Number of VPA objects with at least one evicted Pod.",
		}, []string{"vpa_size_log2", "update_mode"},
	)

	failedEvictionAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "failed_eviction_attempts_total",
			Help:      "Number of failed attempts to update Pods by eviction",
		}, []string{"vpa_size_log2", "update_mode", "reason", "vpa_name", "vpa_namespace"},
	)

	inPlaceUpdatableCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "in_place_updatable_pods_total",
			Help:      "Number of Pods matching in place update criteria.",
		}, []string{"vpa_size_log2"},
	)

	inPlaceUpdatedCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "in_place_updated_pods_total",
			Help:      "Number of Pods updated in-place by Updater to apply a new recommendation.",
		}, []string{"vpa_size_log2"},
	)

	vpasWithInPlaceUpdatablePodsCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "vpas_with_in_place_updatable_pods_total",
			Help:      "Number of VPA objects with at least one Pod matching in place update criteria.",
		}, []string{"vpa_size_log2"},
	)

	vpasWithInPlaceUpdatedPodsCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "vpas_with_in_place_updated_pods_total",
			Help:      "Number of VPA objects with at least one in-place updated Pod.",
		}, []string{"vpa_size_log2"},
	)

	failedInPlaceUpdateAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "failed_in_place_update_attempts_total",
			Help:      "Number of failed attempts to update Pods in-place.",
		}, []string{"vpa_size_log2", "reason", "vpa_name", "vpa_namespace"},
	)

	functionLatency = metrics.CreateExecutionTimeMetric(metricsNamespace,
		"Time spent in various parts of VPA Updater main loop.")
)

// Register initializes all metrics for VPA Updater
func Register() {
	collectors := []prometheus.Collector{
		controlledCount,
		evictableCount,
		evictedCount,
		vpasWithEvictablePodsCount,
		vpasWithEvictedPodsCount,
		failedEvictionAttempts,
		inPlaceUpdatableCount,
		inPlaceUpdatedCount,
		vpasWithInPlaceUpdatablePodsCount,
		vpasWithInPlaceUpdatedPodsCount,
		failedInPlaceUpdateAttempts,
		functionLatency,
	}
	prometheus.MustRegister(collectors...)
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

// newModeAndSizeBasedGauge provides a wrapper for counting items in a loop
func newModeAndSizeBasedGauge(gauge *prometheus.GaugeVec) *UpdateModeAndSizeBasedGauge {
	g := &UpdateModeAndSizeBasedGauge{
		gauge: gauge,
	}
	for i := range g.values {
		g.values[i] = make(map[vpa_types.UpdateMode]int)
	}
	return g
}

// NewControlledPodsCounter returns a wrapper for counting Pods controlled by Updater
func NewControlledPodsCounter() *UpdateModeAndSizeBasedGauge {
	return newModeAndSizeBasedGauge(controlledCount)
}

// NewEvictablePodsCounter returns a wrapper for counting Pods which are matching eviction criteria
func NewEvictablePodsCounter() *UpdateModeAndSizeBasedGauge {
	return newModeAndSizeBasedGauge(evictableCount)
}

// NewVpasWithEvictablePodsCounter returns a wrapper for counting VPA objects with Pods matching eviction criteria
func NewVpasWithEvictablePodsCounter() *UpdateModeAndSizeBasedGauge {
	return newModeAndSizeBasedGauge(vpasWithEvictablePodsCount)
}

// NewVpasWithEvictedPodsCounter returns a wrapper for counting VPA objects with evicted Pods
func NewVpasWithEvictedPodsCounter() *UpdateModeAndSizeBasedGauge {
	return newModeAndSizeBasedGauge(vpasWithEvictedPodsCount)
}

// AddEvictedPod increases the counter of pods evicted by Updater, by given VPA size
func AddEvictedPod(vpaSize int, mode vpa_types.UpdateMode) {
	log2 := metrics.GetVpaSizeLog2(vpaSize)
	evictedCount.WithLabelValues(strconv.Itoa(log2), string(mode)).Inc()
}

// RecordFailedEviction increases the counter of failed eviction attempts by given VPA size, name, namespace, update mode and reason
func RecordFailedEviction(vpaSize int, vpaName string, vpaNamespace string, mode vpa_types.UpdateMode, reason string) {
	log2 := metrics.GetVpaSizeLog2(vpaSize)
	failedEvictionAttempts.WithLabelValues(strconv.Itoa(log2), string(mode), reason, vpaName, vpaNamespace).Inc()
}

// NewInPlaceUpdatablePodsCounter returns a wrapper for counting Pods which are matching in-place update criteria
func NewInPlaceUpdatablePodsCounter() *SizeBasedGauge {
	return newSizeBasedGauge(inPlaceUpdatableCount)
}

// NewVpasWithInPlaceUpdatablePodsCounter returns a wrapper for counting VPA objects with Pods matching in-place update criteria
func NewVpasWithInPlaceUpdatablePodsCounter() *SizeBasedGauge {
	return newSizeBasedGauge(vpasWithInPlaceUpdatablePodsCount)
}

// NewVpasWithInPlaceUpdatedPodsCounter returns a wrapper for counting VPA objects with in-place updated Pods
func NewVpasWithInPlaceUpdatedPodsCounter() *SizeBasedGauge {
	return newSizeBasedGauge(vpasWithInPlaceUpdatedPodsCount)
}

// AddInPlaceUpdatedPod increases the counter of pods updated in place by Updater, by given VPA size
func AddInPlaceUpdatedPod(vpaSize int) {
	log2 := metrics.GetVpaSizeLog2(vpaSize)
	inPlaceUpdatedCount.WithLabelValues(strconv.Itoa(log2)).Inc()
}

// RecordFailedInPlaceUpdate increases the counter of failed in-place update attempts by given VPA size, name, namespace and reason
func RecordFailedInPlaceUpdate(vpaSize int, vpaName string, vpaNamespace string, reason string) {
	log2 := metrics.GetVpaSizeLog2(vpaSize)
	failedInPlaceUpdateAttempts.WithLabelValues(strconv.Itoa(log2), reason, vpaName, vpaNamespace).Inc()
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

// Add increases the counter for the given VPA size and VPA update mode.
func (g *UpdateModeAndSizeBasedGauge) Add(vpaSize int, vpaUpdateMode vpa_types.UpdateMode, value int) {
	log2 := metrics.GetVpaSizeLog2(vpaSize)
	g.values[log2][vpaUpdateMode] += value
}

// Observe stores the recorded values into metrics object associated with the
// wrapper
func (g *UpdateModeAndSizeBasedGauge) Observe() {
	for log2, valueMap := range g.values {
		for vpaMode, value := range valueMap {
			g.gauge.WithLabelValues(strconv.Itoa(log2), string(vpaMode)).Set(float64(value))
		}
	}
}
