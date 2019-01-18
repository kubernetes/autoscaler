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

// Package admission (aka metrics_admission) - code for metrics of VPA Admission Controller plugin
package admission

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
)

const (
	metricsNamespace = metrics.TopMetricsNamespace + "admission_controller"
)

// AdmissionLatency measures latency / execution time of Admission Control execution
// usual usage pattern is: timer := NewAdmissionLatency() ; compute ; timer.Observe()
type AdmissionLatency struct {
	histo *prometheus.HistogramVec
	start time.Time
}

// AdmissionStatus describes the result of Admission Control execution
type AdmissionStatus string

// AdmissionResource describes the resource processed by Admission Control execution
type AdmissionResource string

const (
	// Error denotes a failed Admission Control execution
	Error AdmissionStatus = "error"
	// Skipped denotes an Admission Control execution w/o applying a recommendation
	Skipped AdmissionStatus = "skipped"
	// Applied denotes an Admission Control execution when a recommendation was applied
	Applied AdmissionStatus = "applied"
)

const (
	// Unknown means that the resource could not be determined
	Unknown AdmissionResource = "unknown"
	// Pod means Kubernetes Pod
	Pod AdmissionResource = "Pod"
	// Vpa means VerticalPodAutoscaler object (CRD)
	Vpa AdmissionResource = "VPA"
)

var (
	admissionCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "admission_pods_total",
			Help:      "Number of Pods processed by VPA Admission Controller.",
		}, []string{"applied"},
	)

	admissionLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "admission_latency_seconds",
			Help:      "Time spent in VPA Admission Controller.",
			Buckets:   []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0, 10.0, 20.0, 30.0, 60.0, 120.0, 300.0},
		}, []string{"status", "resource"},
	)
)

// Register initializes all metrics for VPA Admission Contoller
func Register() {
	prometheus.MustRegister(admissionCount)
	prometheus.MustRegister(admissionLatency)
}

// OnAdmittedPod increases the counter of pods handled by VPA Admission Controller
func OnAdmittedPod(touched bool) {
	admissionCount.WithLabelValues(fmt.Sprintf("%v", touched)).Add(1)
}

// NewAdmissionLatency provides a timer for admission latency; call Observe() on it to measure
func NewAdmissionLatency() *AdmissionLatency {
	return &AdmissionLatency{
		histo: admissionLatency,
		start: time.Now(),
	}
}

// Observe measures the execution time from when the AdmissionLatency was created
func (t *AdmissionLatency) Observe(status AdmissionStatus, resource AdmissionResource) {
	(*t.histo).WithLabelValues(string(status), string(resource)).Observe(time.Now().Sub(t.start).Seconds())
}
