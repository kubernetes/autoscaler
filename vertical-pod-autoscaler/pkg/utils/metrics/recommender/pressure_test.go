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

package recommender

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// collectGauge returns the current value of every series of a gauge, by its single label value.
func collectGauge(t *testing.T, g *prometheus.GaugeVec) map[string]float64 {
	metrics := make(chan prometheus.Metric)
	go func() {
		g.Collect(metrics)
		close(metrics)
	}()
	got := map[string]float64{}
	for m := range metrics {
		var proto dto.Metric
		if err := m.Write(&proto); err != nil {
			t.Errorf("failed to write metric: %v", err)
			continue
		}
		got[proto.GetLabel()[0].GetValue()] = proto.GetGauge().GetValue()
	}
	return got
}

func TestPressureStateGauges(t *testing.T) {
	testCases := []struct {
		name  string
		set   func(map[string]int)
		gauge *prometheus.GaugeVec
		// states is written out here rather than read from the package, so that dropping a state
		// the recommender reports fails this test instead of silently losing the series.
		states []string
	}{
		{
			name:   "containers",
			set:    SetPressureContainers,
			gauge:  pressureContainers,
			states: []string{"injectable", "cooldown", "resize_in_progress", "resize_deferred", "resize_infeasible"},
		},
		{
			name:   "nodes",
			set:    SetPressureNodes,
			gauge:  pressureNodes,
			states: []string{"observed", "missing_psi", "failed", "stale", "undersampled", "unit_mismatch"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(tc.gauge.Reset)
			want := map[string]float64{}
			for _, s := range tc.states {
				want[s] = 0
			}
			want[tc.states[0]] = 7

			tc.set(map[string]int{tc.states[0]: 7})

			assert.Equal(t, want, collectGauge(t, tc.gauge), "every state is reported, and a state absent from the window reads zero")
		})
	}
}
