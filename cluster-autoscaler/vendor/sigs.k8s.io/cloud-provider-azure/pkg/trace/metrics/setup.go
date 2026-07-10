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
	"go.opentelemetry.io/otel/metric"
)

// Setup sets up cloud-provider metrics.
func Setup(meter metric.Meter) error {
	setups := []func(metric.Meter) error{
		// For reconcile metrics
		setupReconcileLatency,
		setupReconcileErrors,
	}

	for _, setup := range setups {
		if err := setup(meter); err != nil {
			return err
		}
	}

	return nil
}
