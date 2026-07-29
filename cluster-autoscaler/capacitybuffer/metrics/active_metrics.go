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
	"sync"

	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

var (
	capacityBuffersNumberMetric *k8smetrics.GaugeVec
	registerOnce                sync.Once
)

// RegisterActiveMetrics registers the capacity buffers metrics in legacyregistry.
// This is done on-demand (only if the capacity buffers feature is active/enabled).
func RegisterActiveMetrics() {
	registerOnce.Do(func() {
		capacityBuffersNumberMetric = k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: "cluster_autoscaler",
				Name:      "capacity_buffers_count",
				Help:      "Number of capacity buffers in the cluster",
			},
			[]string{"provisioning_strategy"},
		)

		legacyregistry.MustRegister(capacityBuffersNumberMetric)
	})
}

// UpdateCapacityBuffersNumber records the number of capacity buffers in the cluster.
func UpdateCapacityBuffersNumber(countsByType map[string]int) {
	if capacityBuffersNumberMetric == nil {
		return
	}
	capacityBuffersNumberMetric.Reset()
	for strategy, count := range countsByType {
		capacityBuffersNumberMetric.WithLabelValues(strategy).Set(float64(count))
	}
}
