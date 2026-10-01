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

package updater

import "github.com/prometheus/client_golang/prometheus"

var (
	pressureQuickUpdates = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "pressure_quick_updates_total",
			Help:      "Pods considered for a quick pressure update, by outcome.",
		}, []string{"result"},
	)

	quickPressureResults = []string{"admitted", "capped", "no_change", "decrease_clamped", "decrease_blocked", "eviction_skipped"}
)

// RecordQuickPressure counts a quick pressure outcome, one of quickPressureResults.
func RecordQuickPressure(result string) {
	pressureQuickUpdates.WithLabelValues(result).Inc()
}
