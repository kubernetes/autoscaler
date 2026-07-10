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

package metrics

import (
	"fmt"

	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var (
	reconcileLatency api.Float64Histogram
	reconcileErrors  api.Int64Counter
)

// ReconcileLatency returns the histogram for reconcile latency.
func ReconcileLatency() api.Float64Histogram {
	if reconcileLatency == nil {
		return noop.Float64Histogram{}
	}
	return reconcileLatency
}

// ReconcileErrors returns the counter for reconcile errors.
func ReconcileErrors() api.Int64Counter {
	if reconcileErrors == nil {
		return noop.Int64Counter{}
	}
	return reconcileErrors
}

func setupReconcileLatency(meter api.Meter) error {
	m, err := meter.Float64Histogram(
		"provider.reconcile.duration",
		api.WithUnit("s"),
		api.WithDescription("Measures the duration of reconcile operations in seconds."),
		api.WithExplicitBucketBoundaries(0.1, 0.2, 0.5, 1, 5, 10, 60, 300, 600),
	)

	if err != nil {
		return fmt.Errorf("create provider.reconcile.duration histogram: %w", err)
	}

	reconcileLatency = m

	return nil
}

func setupReconcileErrors(meter api.Meter) error {
	c, err := meter.Int64Counter(
		"provider.reconcile.errors.counter",
		api.WithDescription("Measures the number of errors during reconcile operations."),
	)

	if err != nil {
		return fmt.Errorf("create provider.reconcile.errors.counter: %w", err)
	}

	reconcileErrors = c

	return nil
}
