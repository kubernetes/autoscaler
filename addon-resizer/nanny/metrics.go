/*
Copyright 2024 The Kubernetes Authors.

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

// Package nanny - this file contains metrics for the nanny package.
package nanny

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "k8s.io/component-base/metrics/prometheus/restclient" // for client-go metrics registration
	"k8s.io/klog/v2"
)

const (
	namespace = "addon_resizer"
)

var (
	executionOutcome = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "execution_outcome",
			Help:      "Count of execution loop outcomes.",
		}, []string{"outcome"},
	)
)

// Initialize sets up Prometheus to expose metrics.
func InitializeMetrics(address string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(address, nil)
		klog.Fatalf("Failed to serve metrics: %v", err)
	}()
}

func RegisterMetrics() {
	prometheus.MustRegister(executionOutcome)
}

func ObserveOutcome(outcome string) {
	executionOutcome.WithLabelValues(outcome).Inc()
}
