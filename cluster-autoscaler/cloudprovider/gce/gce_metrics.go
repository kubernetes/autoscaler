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

package gce

import (
	"time"

	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	caNamespace = "cluster_autoscaler"
)

var (
	/**** Metrics related to GCE API usage ****/
	requestCounter = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "gce_request_count",
			Help:      "Deprecated: Counter of GCE API requests for each verb and API resource. Use request_latencies instead.",
		}, []string{"resource", "verb"},
	)

	requestLatencies = k8smetrics.NewHistogramVec(
		&k8smetrics.HistogramOpts{
			Namespace: caNamespace,
			Name:      "request_latencies",
			Help:      "Latencies of requests made by Cluster Autoscaler, measured in seconds.",
			Buckets: []float64{
				0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 3, 4, 5, 6, 7, 8, 16, 32, 64, 128,
				256, 512, 1024, 2048, 4096,
			},
		}, []string{"service", "resource", "verb", "status"},
	)
)

// RegisterMetrics registers all GCE metrics.
func RegisterMetrics() {
	legacyregistry.MustRegister(requestCounter)
	legacyregistry.MustRegister(requestLatencies)
}

// registerRequest registers request to GCE API.
func registerRequest(resource string, verb string) {
	requestCounter.WithLabelValues(resource, verb).Add(1.0)
}

// EmitRequestLatencyMetric emits to request_latencies histogram metric.
// It is exported to allow downstream GKE cloudprovider packages to emit
// latency metrics with appropriate service labels ("gce", "gke", "service_control").
func EmitRequestLatencyMetric(service, resource, verb string, response any, err error, start time.Time, withReason bool) {
	duration := time.Since(start)
	status := determineResponseStatusForLatencyMetric(response, err, withReason)
	requestLatencies.WithLabelValues(service, resource, verb, status).Observe(duration.Seconds())
	if service == "gce" {
		registerRequest(resource, verb)
	}
}

// emitGceLatency emits both request latency and legacy request counter for GCE API calls.
func emitGceLatency(resource, verb string, response any, err error, start time.Time) {
	EmitRequestLatencyMetric("gce", resource, verb, response, err, start, false)
}
