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
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
)

const (
	metricsNamespace = metrics.TopMetricsNamespace + "updater"
)

var (
	evictedCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "evicted_pods_total",
			Help:      "Number of Pods evicted by Updater to apply a new recommendation.",
		},
	)
)

// Register initializes all metrics for VPA Updater
func Register() {
	prometheus.MustRegister(evictedCount)
}

// AddEvictedPod increases the counter of pods evicted by VPA
func AddEvictedPod() {
	evictedCount.Add(1)
}
