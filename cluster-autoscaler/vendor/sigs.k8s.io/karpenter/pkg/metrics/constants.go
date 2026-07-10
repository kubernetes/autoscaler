/*
Copyright The Kubernetes Authors.

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
	"time"

	opmetrics "github.com/awslabs/operatorpkg/metrics"
)

const (
	// Common namespace for application metrics.
	Namespace = "karpenter"

	NodePoolLabel         = "nodepool"
	ReasonLabel           = "reason"
	ResourceTypeLabel     = "resource_type"
	CapacityTypeLabel     = "capacity_type"
	ZoneLabel             = "zone"
	MinValuesRelaxedLabel = "min_values_relaxed"

	// Reasons for CREATE/DELETE shared metrics
	ProvisionedReason = "provisioned"
	ExpiredReason     = "expired"
	UnhealthyReason   = "unhealthy"
)

// DurationBuckets returns a []float64 of default threshold values for duration histograms.
// Each returned slice is new and may be modified without impacting other bucket definitions.
func DurationBuckets() []float64 {
	// Use same bucket thresholds as controller-runtime.
	// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.10.0/pkg/internal/controller/metrics/metrics.go#L47-L48
	// Add in values larger than 60 for singleton controllers that do not have a timeout.
	return []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0,
		1.25, 1.5, 1.75, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 40, 50, 60, 120, 150, 300, 450, 600,
	}
}

// Returns a map of summary objectives (quantile-error pairs)
func SummaryObjectives() map[float64]float64 {
	const epsilon = 0.01
	objectives := make(map[float64]float64)
	for _, quantile := range []float64{0.0, 0.5, 0.9, 0.99, 1.0} {
		objectives[quantile] = epsilon
	}
	return objectives
}

// Measure returns a deferrable function that observes the duration between the
// defer statement and the end of the function.
func Measure(observer opmetrics.ObservationMetric, labels map[string]string) func() {
	start := time.Now()
	return func() { observer.Observe(time.Since(start).Seconds(), labels) }
}
