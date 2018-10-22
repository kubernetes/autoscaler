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
	"github.com/prometheus/client_golang/prometheus"
)

const (
	caNamespace = "cluster_autoscaler"
)

var (
	/**** Metrics related to GCE API usage ****/
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "gce_request_count",
			Help:      "Counter of GCE API requests for each verb and API resource.",
		}, []string{"resource", "verb"},
	)
)

// registerMetrics registers all GCE metrics.
func registerMetrics() {
	prometheus.MustRegister(requestCounter)
}

// registerRequest registers request to GCE API.
func registerRequest(resource string, verb string) {
	requestCounter.WithLabelValues(resource, verb).Add(1.0)
}
