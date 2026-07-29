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
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/component-base/metrics/testutil"
)

func TestActiveMetrics_Lifecycle(t *testing.T) {
	// Dynamically register active metrics (which uses sync.Once internally).
	RegisterActiveMetrics()

	// Update total count
	countsByType := map[string]int{
		"strategy1": 2,
		"strategy2": 1,
	}
	UpdateCapacityBuffersNumber(countsByType)

	// Verify initially both values are correct.
	gaugeTotal1 := capacityBuffersNumberMetric.WithLabelValues("strategy1")
	valTotal1, err := testutil.GetGaugeMetricValue(gaugeTotal1)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), valTotal1)

	gaugeTotal2 := capacityBuffersNumberMetric.WithLabelValues("strategy2")
	valTotal2, err := testutil.GetGaugeMetricValue(gaugeTotal2)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), valTotal2)

	// Update counts with a new map (strategy1 is removed, strategy2 is changed)
	countsByType2 := map[string]int{
		"strategy2": 5,
	}
	UpdateCapacityBuffersNumber(countsByType2)

	// Verify strategy2 value is updated
	gaugeTotal2Updated := capacityBuffersNumberMetric.WithLabelValues("strategy2")
	valTotal2Updated, err := testutil.GetGaugeMetricValue(gaugeTotal2Updated)
	assert.NoError(t, err)
	assert.Equal(t, float64(5), valTotal2Updated)

	// Verify strategy1 has been reset (accessing it after Reset recreates it with value 0)
	gaugeTotal1Reset := capacityBuffersNumberMetric.WithLabelValues("strategy1")
	valTotal1Reset, err := testutil.GetGaugeMetricValue(gaugeTotal1Reset)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), valTotal1Reset)
}
