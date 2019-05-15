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
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	"k8s.io/klog"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricsNamespace = metrics.TopMetricsNamespace + "quality"
)

var (
	// Buckets between 0.01 and 655.36
	cpuBuckets = prometheus.ExponentialBuckets(0.01, 2., 17)
	// Buckets between 1MB and
	memoryBuckets = prometheus.ExponentialBuckets(1e6, 2., 17)
)

var (
	usageRecommendationRelativeDiff = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "usage_recommendation_relative_diffs",
			Help:      "Diffs between usage and recommendation, normalized by recommendation value",
			Buckets: []float64{-1., -.75, -.5, -.25, -.1, -.05, -0.025, -.01, -.005, -0.0025, -.001, 0.,
				.001, .0025, .005, .01, .025, .05, .1, .25, .5, .75, 1., 2.5, 5., 10., 25., 50., 100.},
		}, []string{"update_mode", "resource", "is_oom"},
	)
	usageMissingRecommendationCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "usage_sample_missing_recommendation_count",
			Help:      "Count of usage samples when a recommendation should be present but is missing",
		}, []string{"update_mode", "resource", "is_oom"},
	)
	cpuUsageRecommendationDiff = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "cpu_usage_recommendation_diffs_cores",
			Help:      "Absolute diffs between usage and recommendation for CPU",
			Buckets:   prependNegativeBuckets(cpuBuckets),
		}, []string{"update_mode", "recommendation_missing"},
	)
	memoryUsageRecommendationDiff = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "mem_usage_recommendation_diffs_bytes",
			Help:      "Absolute diffs between usage and recommendation for memory",
			Buckets:   prependNegativeBuckets(memoryBuckets),
		}, []string{"update_mode", "recommendation_missing", "is_oom"},
	)
	cpuRecommendations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "cpu_recommendations_cores",
			Help:      "CPU recommendation values as observed on recorded usage sample",
			Buckets:   cpuBuckets,
		}, []string{"update_mode"},
	)
	memoryRecommendations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "mem_recommendations_bytes",
			Help:      "Memory recommendation values as observed on recorded usage sample",
			Buckets:   memoryBuckets,
		}, []string{"update_mode", "is_oom"},
	)
)

// Register initializes all VPA quality metrics
func Register() {
	prometheus.MustRegister(usageRecommendationRelativeDiff)
	prometheus.MustRegister(usageMissingRecommendationCounter)
	prometheus.MustRegister(cpuUsageRecommendationDiff)
	prometheus.MustRegister(memoryUsageRecommendationDiff)
	prometheus.MustRegister(cpuRecommendations)
	prometheus.MustRegister(memoryRecommendations)
}

// observeUsageRecommendationRelativeDiff records relative diff between usage and
// recommendation if recommendation has a positive value.
func observeUsageRecommendationRelativeDiff(usage, recommendation float64, isOOM bool, resource corev1.ResourceName, updateMode *vpa_types.UpdateMode) {
	if recommendation > 0 {
		usageRecommendationRelativeDiff.WithLabelValues(updateModeToString(updateMode), string(resource), strconv.FormatBool(isOOM)).Observe((usage - recommendation) / recommendation)
	}
}

// observeMissingRecommendation counts usage samples with missing recommendations.
func observeMissingRecommendation(isOOM bool, resource corev1.ResourceName, updateMode *vpa_types.UpdateMode) {
	usageMissingRecommendationCounter.WithLabelValues(updateModeToString(updateMode), string(resource), strconv.FormatBool(isOOM)).Inc()
}

// observeUsageRecommendationDiff records absolute diff between usage and
// recommendation.
func observeUsageRecommendationDiff(usage, recommendation float64, isRecommendationMissing, isOOM bool, resource corev1.ResourceName, updateMode *vpa_types.UpdateMode) {
	switch resource {
	case corev1.ResourceCPU:
		cpuUsageRecommendationDiff.WithLabelValues(updateModeToString(updateMode), strconv.FormatBool(isRecommendationMissing)).Observe(usage - recommendation)
	case corev1.ResourceMemory:
		memoryUsageRecommendationDiff.WithLabelValues(updateModeToString(updateMode),
			strconv.FormatBool(isRecommendationMissing), strconv.FormatBool(isOOM)).Observe(usage - recommendation)
	default:
		klog.Warningf("Unknown resource: %v", resource)
	}
}

// observeRecommendation records the value of recommendation.
func observeRecommendation(recommendation float64, isOOM bool, resource corev1.ResourceName, updateMode *vpa_types.UpdateMode) {
	switch resource {
	case corev1.ResourceCPU:
		cpuRecommendations.WithLabelValues(updateModeToString(updateMode)).Observe(recommendation)
	case corev1.ResourceMemory:
		memoryRecommendations.WithLabelValues(updateModeToString(updateMode), strconv.FormatBool(isOOM)).Observe(recommendation)
	default:
		klog.Warningf("Unknown resource: %v", resource)
	}
}

// ObserveQualityMetrics records all quality metrics that we can derive from usage and recommendation.
func ObserveQualityMetrics(usage, recommendation float64, isOOM bool, resource corev1.ResourceName, updateMode *vpa_types.UpdateMode) {
	observeRecommendation(recommendation, isOOM, resource, updateMode)
	observeUsageRecommendationDiff(usage, recommendation, false, isOOM, resource, updateMode)
	observeUsageRecommendationRelativeDiff(usage, recommendation, isOOM, resource, updateMode)
}

// ObserveQualityMetricsRecommendationMissing records all quality metrics that we can derive from usage when recommendation is missing.
func ObserveQualityMetricsRecommendationMissing(usage float64, isOOM bool, resource corev1.ResourceName, updateMode *vpa_types.UpdateMode) {
	observeMissingRecommendation(isOOM, resource, updateMode)
	observeUsageRecommendationDiff(usage, 0, true, isOOM, resource, updateMode)
}

func prependNegativeBuckets(buckets []float64) []float64 {
	bucketCount := len(buckets)
	result := make([]float64, bucketCount*2+1)
	result[bucketCount] = 0
	for i, bucket := range buckets {
		if bucket < 0 {
			panic(fmt.Sprintf("Positive buckets expected. Got: %v", bucket))
		}
		result[bucketCount-i-1] = -bucket
		result[bucketCount+i+1] = bucket
	}
	return result
}

func updateModeToString(updateMode *vpa_types.UpdateMode) string {
	if updateMode == nil {
		return ""
	}
	return string(*updateMode)
}
