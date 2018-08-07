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
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const (
	// TopMetricsNamespace is a prefix for all VPA-related metrics namespaces
	TopMetricsNamespace = "vpa_"
)

// Initialize sets up Prometheus to expose metrics on the given address
func Initialize(address string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(address, nil)
		glog.Fatalf("Failed to start metrics: %v", err)
	}()
}
