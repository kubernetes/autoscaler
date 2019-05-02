/*
Copyright 2019 The Kubernetes Authors.

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

// Package quality (aka metrics_quality) - code for VPA quality metrics
package quality

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricsNamespace = metrics.TopMetricsNamespace + "quality"
)

var (
	usageRecommendationRelativeDiff = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "usage_recommendation_relative_diffs",
			Help:      "Diffs between usage and recommendation, normalized by recommendation value",
			Buckets: []float64{-1., -.75, -.5, -.25, -.1, -.05, -0.025, -.01, -.005, -0.0025, -.001, 0.,
				.001, .0025, .005, .01, .025, .05, .1, .25, .5, .75, 1., 2.5, 5., 10., 25., 50., 100.},
		}, []string{"resource", "is_oom"},
	)
	usageMissingRecommendationCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "usage_sample_missing_recommendation_count",
			Help:      "Count of usage samples when a recommendation should be present but is missing.",
		}, []string{"resource", "is_oom"},
	)
)

// Register initializes all VPA quality metrics
func Register() {
	prometheus.MustRegister(usageRecommendationRelativeDiff)
	prometheus.MustRegister(usageMissingRecommendationCounter)
}

// ObserveUsageRecommendationRelativeDiff records relative diff between usage and
// recommendation if recommendation has a positive value.
func ObserveUsageRecommendationRelativeDiff(usage, recommendation float64, isOOM bool, resource corev1.ResourceName) {
	if recommendation > 0 {
		usageRecommendationRelativeDiff.WithLabelValues(string(resource), strconv.FormatBool(isOOM)).Observe((usage - recommendation) / recommendation)
	}
}

// ObserveMissingRecommendation counts usage samples with missing recommendations.
func ObserveMissingRecommendation(isOOM bool, resource corev1.ResourceName) {
	usageMissingRecommendationCounter.WithLabelValues(string(resource), strconv.FormatBool(isOOM)).Inc()
}
