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

// Package metrics - common code for metrics of all 3 VPA components
package metrics

import (
	"math"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "k8s.io/component-base/metrics/prometheus/restclient" // for client-go metrics registration
	"k8s.io/klog"
)

// ExecutionTimer measures execution time of a computation, split into major steps
// usual usage pattern is: timer := NewExecutionTimer(...) ; compute ; timer.ObserveStep() ; ... ; timer.ObserveTotal()
type ExecutionTimer struct {
	histo *prometheus.HistogramVec
	start time.Time
	last  time.Time
}

const (
	// TopMetricsNamespace is a prefix for all VPA-related metrics namespaces
	TopMetricsNamespace = "vpa_"

	// MaxVpaSizeLog - The metrics will distinguish VPA sizes up to 2^MaxVpaSizeLog (~1M)
	// Anything above that size will be reported in the top bucket.
	MaxVpaSizeLog = 20
)

// Initialize sets up Prometheus to expose metrics & (optionally) health-check on the given address
func Initialize(address string, healthCheck *HealthCheck) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if healthCheck != nil {
			http.Handle("/health-check", healthCheck)
		}
		err := http.ListenAndServe(address, nil)
		klog.Fatalf("Failed to start metrics: %v", err)
	}()
}

// NewExecutionTimer provides a timer for admission latency; call ObserveXXX() on it to measure
func NewExecutionTimer(histo *prometheus.HistogramVec) *ExecutionTimer {
	now := time.Now()
	return &ExecutionTimer{
		histo: histo,
		start: now,
		last:  now,
	}
}

// ObserveStep measures the execution time from the last call to the ExecutionTimer
func (t *ExecutionTimer) ObserveStep(step string) {
	now := time.Now()
	(*t.histo).WithLabelValues(step).Observe(now.Sub(t.last).Seconds())
	t.last = now
}

// ObserveTotal measures the execution time from the creation of the ExecutionTimer
func (t *ExecutionTimer) ObserveTotal() {
	(*t.histo).WithLabelValues("total").Observe(time.Now().Sub(t.start).Seconds())
}

// CreateExecutionTimeMetric prepares a new histogram labeled with execution step
func CreateExecutionTimeMetric(namespace string, help string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "execution_latency_seconds",
			Help:      help,
			Buckets: []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0, 10.0,
				20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0, 120.0, 150.0, 240.0, 300.0},
		}, []string{"step"},
	)
}

// GetVpaSizeLog2 returns a bucket number for a metric labelled with number of Pods under a given VPA.
// It is basically log2(vpaSize), capped to MaxVpaSizeLog
func GetVpaSizeLog2(vpaSize int) int {
	if vpaSize == 0 {
		return 0
	}

	ret := int(math.Log2(float64(vpaSize)))
	if ret > MaxVpaSizeLog {
		return MaxVpaSizeLog
	}
	return ret
}
