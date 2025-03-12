/*
Copyright 2020 The Kubernetes Authors.

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

package metrics

import (
	"strings"
	"time"

	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

var (
	metricLabels = []string{
		"request",         // API function that is being invoked
		"resource_group",  // Resource group of the resource being monitored
		"subscription_id", // Subscription ID of the resource being monitored
		"source",          // Operation source(optional)
	}

	apiMetrics       = registerAPIMetrics(metricLabels...)
	operationMetrics = registerOperationMetrics(metricLabels...)
)

// apiCallMetrics is the metrics measuring the performance of a single API call
// e.g., GET, POST ...
type apiCallMetrics struct {
	latency          *metrics.HistogramVec
	errors           *metrics.CounterVec
	rateLimitedCount *metrics.CounterVec
	throttledCount   *metrics.CounterVec
}

// operationCallMetrics is the metrics measuring the performance of a whole operation
// e.g., the create / update / delete process of a loadbalancer or route.
type operationCallMetrics struct {
	operationLatency      *metrics.HistogramVec
	operationFailureCount *metrics.CounterVec
}

// MetricContext indicates the context for Azure client metrics.
type MetricContext struct {
	start      time.Time
	attributes []string
	// log level in ObserveOperationWithResult
	LogLevel int32
}

// NewMetricContext creates a new MetricContext.
func NewMetricContext(prefix, request, resourceGroup, subscriptionID, source string) *MetricContext {
	return &MetricContext{
		start:      time.Now(),
		attributes: []string{prefix + "_" + request, strings.ToLower(resourceGroup), subscriptionID, source},
		LogLevel:   3,
	}
}

// RateLimitedCount records the metrics for rate limited request count.
func (mc *MetricContext) RateLimitedCount() {
	apiMetrics.rateLimitedCount.WithLabelValues(mc.attributes...).Inc()
}

// ThrottledCount records the metrics for throttled request count.
func (mc *MetricContext) ThrottledCount() {
	apiMetrics.throttledCount.WithLabelValues(mc.attributes...).Inc()
}

// Observe observes the request latency and failed requests.
func (mc *MetricContext) Observe(rerr *retry.Error, labelAndValues ...interface{}) {
	latency := time.Since(mc.start).Seconds()
	apiMetrics.latency.WithLabelValues(mc.attributes...).Observe(latency)
	if rerr != nil {
		errorCode := rerr.ServiceErrorCode()
		attributes := append(mc.attributes, errorCode)
		apiMetrics.errors.WithLabelValues(attributes...).Inc()
	}
	mc.logLatency(6, latency, append(labelAndValues, "error_code", rerr.ServiceErrorCode())...)
}

// ObserveOperationWithResult observes the request latency and failed requests of an operation.
func (mc *MetricContext) ObserveOperationWithResult(isOperationSucceeded bool, labelAndValues ...interface{}) {
	latency := time.Since(mc.start).Seconds()
	operationMetrics.operationLatency.WithLabelValues(mc.attributes...).Observe(latency)
	resultCode := "succeeded"
	if !isOperationSucceeded {
		resultCode = "failed"
		if len(mc.attributes) > 0 {
			resultCode += mc.attributes[0][strings.Index(mc.attributes[0], "_"):]
		}
		mc.CountFailedOperation()
	}
	mc.logLatency(mc.LogLevel, latency, append(labelAndValues, "result_code", resultCode)...)
}

func (mc *MetricContext) logLatency(logLevel int32, latency float64, additionalKeysAndValues ...interface{}) {
	keysAndValues := []interface{}{"latency_seconds", latency}
	for i, label := range metricLabels {
		keysAndValues = append(keysAndValues, label, mc.attributes[i])
	}
	klog.V(klog.Level(logLevel)).InfoS("Observed Request Latency", append(keysAndValues, additionalKeysAndValues...)...)
}

// CountFailedOperation increase the number of failed operations
func (mc *MetricContext) CountFailedOperation() {
	operationMetrics.operationFailureCount.WithLabelValues(mc.attributes...).Inc()
}

// registerAPIMetrics registers the API metrics.
func registerAPIMetrics(attributes ...string) *apiCallMetrics {
	metrics := &apiCallMetrics{
		latency: metrics.NewHistogramVec(
			&metrics.HistogramOpts{
				Namespace:      consts.AzureMetricsNamespace,
				Name:           "api_request_duration_seconds",
				Help:           "Latency of an Azure API call",
				Buckets:        []float64{.1, .25, .5, 1, 2.5, 5, 10, 15, 25, 50, 120, 300, 600, 1200},
				StabilityLevel: metrics.ALPHA,
			},
			attributes,
		),
		errors: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Namespace:      consts.AzureMetricsNamespace,
				Name:           "api_request_errors",
				Help:           "Number of errors for an Azure API call",
				StabilityLevel: metrics.ALPHA,
			},
			append(attributes, "code"),
		),
		rateLimitedCount: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Namespace:      consts.AzureMetricsNamespace,
				Name:           "api_request_ratelimited_count",
				Help:           "Number of rate limited Azure API calls",
				StabilityLevel: metrics.ALPHA,
			},
			attributes,
		),
		throttledCount: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Namespace:      consts.AzureMetricsNamespace,
				Name:           "api_request_throttled_count",
				Help:           "Number of throttled Azure API calls",
				StabilityLevel: metrics.ALPHA,
			},
			attributes,
		),
	}

	legacyregistry.MustRegister(metrics.latency)
	legacyregistry.MustRegister(metrics.errors)
	legacyregistry.MustRegister(metrics.rateLimitedCount)
	legacyregistry.MustRegister(metrics.throttledCount)

	return metrics
}

// registerOperationMetrics registers the operation metrics.
func registerOperationMetrics(attributes ...string) *operationCallMetrics {
	metrics := &operationCallMetrics{
		operationLatency: metrics.NewHistogramVec(
			&metrics.HistogramOpts{
				Namespace:      consts.AzureMetricsNamespace,
				Name:           "op_duration_seconds",
				Help:           "Latency of an Azure service operation",
				StabilityLevel: metrics.ALPHA,
				Buckets:        []float64{0.1, 0.2, 0.5, 1, 5, 10, 15, 20, 30, 40, 50, 60, 100, 200, 300, 600, 1200},
			},
			attributes,
		),
		operationFailureCount: metrics.NewCounterVec(
			&metrics.CounterOpts{
				Namespace:      consts.AzureMetricsNamespace,
				Name:           "op_failure_count",
				Help:           "Number of failed Azure service operations",
				StabilityLevel: metrics.ALPHA,
			},
			attributes,
		),
	}

	legacyregistry.MustRegister(metrics.operationLatency)
	legacyregistry.MustRegister(metrics.operationFailureCount)

	return metrics
}
