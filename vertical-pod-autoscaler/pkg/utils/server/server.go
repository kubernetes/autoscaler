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

// Package server - common code for mux of all 3 VPA components
package server

import (
	"net/http"
	"net/http/pprof"

	"k8s.io/klog/v2"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
)

// Initialize sets up Prometheus to expose metrics & (optionally) health-check and profiling on the given address
func Initialize(enableProfiling *bool, healthCheck *metrics.HealthCheck, address *string) {
	go func() {
		mux := http.NewServeMux()

		mux.Handle("/metrics", promhttp.Handler())
		if healthCheck != nil {
			mux.Handle("/health-check", healthCheck)
		}

		if *enableProfiling {
			mux.HandleFunc("/debug/pprof/", http.HandlerFunc(pprof.Index))
			mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		}

		err := http.ListenAndServe(*address, mux)
		klog.Fatalf("Failed to start metrics: %v", err)
	}()
}
