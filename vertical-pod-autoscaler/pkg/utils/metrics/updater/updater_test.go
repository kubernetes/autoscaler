/*
Copyright 2025 The Kubernetes Authors.

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

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

func TestAddEvictedPod(t *testing.T) {
	testCases := []struct {
		desc    string
		vpaSize int
		mode    vpa_types.UpdateMode
		log2    string
	}{
		{
			desc:    "VPA size 5, mode Auto",
			vpaSize: 5,
			mode:    vpa_types.UpdateModeAuto,
			log2:    "2",
		},
		{
			desc:    "VPA size 10, mode Off",
			vpaSize: 10,
			mode:    vpa_types.UpdateModeOff,
			log2:    "3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Cleanup(evictedCount.Reset)
			AddEvictedPod(tc.vpaSize, tc.mode)
			val := testutil.ToFloat64(evictedCount.WithLabelValues(tc.log2, string(tc.mode)))
			if val != 1 {
				t.Errorf("Unexpected value for evictedCount metric with labels (%s, %s): got %v, want 1", tc.log2, string(tc.mode), val)
			}
		})
	}
}

func TestRecordFailedEviction(t *testing.T) {
	testCases := []struct {
		desc         string
		vpaSize      int
		mode         vpa_types.UpdateMode
		reason       string
		log2         string
		vpaName      string
		vpaNamespace string
	}{
		{
			desc:         "VPA size 2, some reason",
			vpaSize:      2,
			reason:       "some_reason",
			log2:         "1",
			vpaName:      "vpa-2",
			vpaNamespace: "vpa-2-ns",
		},
		{
			desc:         "VPA size 20, another reason",
			vpaSize:      20,
			reason:       "another_reason",
			log2:         "4",
			vpaName:      "vpa-20",
			vpaNamespace: "vpa-20-ns",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Cleanup(failedEvictionAttempts.Reset)
			RecordFailedEviction(tc.vpaSize, tc.vpaName, tc.vpaNamespace, tc.mode, tc.reason)
			val := testutil.ToFloat64(failedEvictionAttempts.WithLabelValues(tc.log2, string(tc.mode), tc.reason, tc.vpaName, tc.vpaNamespace))
			if val != 1 {
				t.Errorf("Unexpected value for FailedEviction metric with labels (%s, %s): got %v, want 1", tc.log2, tc.reason, val)
			}
		})
	}
}

func TestAddInPlaceUpdatedPod(t *testing.T) {
	testCases := []struct {
		desc    string
		vpaSize int
		log2    string
	}{
		{
			desc:    "VPA size 10",
			vpaSize: 10,
			log2:    "3",
		},
		{
			desc:    "VPA size 1",
			vpaSize: 1,
			log2:    "0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Cleanup(inPlaceUpdatedCount.Reset)
			AddInPlaceUpdatedPod(tc.vpaSize)
			val := testutil.ToFloat64(inPlaceUpdatedCount.WithLabelValues(tc.log2))
			if val != 1 {
				t.Errorf("Unexpected value for InPlaceUpdatedPod metric with labels (%s): got %v, want 1", tc.log2, val)
			}
		})
	}
}

func TestRecordFailedInPlaceUpdate(t *testing.T) {
	testCases := []struct {
		desc         string
		vpaSize      int
		reason       string
		log2         string
		vpaName      string
		vpaNamespace string
	}{
		{
			desc:         "VPA size 2, some reason",
			vpaSize:      2,
			reason:       "some_reason",
			log2:         "1",
			vpaName:      "vpa-2",
			vpaNamespace: "vpa-2-ns",
		},
		{
			desc:         "VPA size 20, another reason",
			vpaSize:      20,
			reason:       "another_reason",
			log2:         "4",
			vpaName:      "vpa-20",
			vpaNamespace: "vpa-20-ns",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Cleanup(failedInPlaceUpdateAttempts.Reset)
			RecordFailedInPlaceUpdate(tc.vpaSize, tc.vpaName, tc.vpaNamespace, tc.reason)
			val := testutil.ToFloat64(failedInPlaceUpdateAttempts.WithLabelValues(tc.log2, tc.reason, tc.vpaName, tc.vpaNamespace))
			if val != 1 {
				t.Errorf("Unexpected value for FailedInPlaceUpdate metric with labels (%s, %s): got %v, want 1", tc.log2, tc.reason, val)
			}
		})
	}
}

func TestUpdateModeAndSizeBasedGauge(t *testing.T) {
	type addition struct {
		vpaSize int
		mode    vpa_types.UpdateMode
		value   int
	}
	type expectation struct {
		labels []string
		value  float64
	}
	testCases := []struct {
		desc            string
		newCounter      func() *UpdateModeAndSizeBasedGauge
		metric          *prometheus.GaugeVec
		metricName      string
		additions       []addition
		expectedMetrics []expectation
	}{
		{
			desc:       "ControlledPodsCounter",
			newCounter: NewControlledPodsCounter,
			metric:     controlledCount,
			metricName: "vpa_updater_controlled_pods_total",
			additions: []addition{
				{1, vpa_types.UpdateModeAuto, 5},
				{2, vpa_types.UpdateModeOff, 10},
				{2, vpa_types.UpdateModeAuto, 2},
				{2, vpa_types.UpdateModeAuto, 7},
			},
			expectedMetrics: []expectation{
				{[]string{"0" /* log2(1) */, "Auto"}, 5},
				{[]string{"1" /* log2(2) */, "Auto"}, 9},
				{[]string{"1" /* log2(2) */, "Off"}, 10},
			},
		},
		{
			desc:       "EvictablePodsCounter",
			newCounter: NewEvictablePodsCounter,
			metric:     evictableCount,
			metricName: "vpa_updater_evictable_pods_total",
			additions: []addition{
				{4, vpa_types.UpdateModeAuto, 3},
				{1, vpa_types.UpdateModeRecreate, 8},
			},
			expectedMetrics: []expectation{
				{[]string{"2" /* log2(4) */, "Auto"}, 3},
				{[]string{"0" /* log2(1) */, "Recreate"}, 8},
			},
		},
		{
			desc:       "VpasWithEvictablePodsCounter",
			newCounter: NewVpasWithEvictablePodsCounter,
			metric:     vpasWithEvictablePodsCount,
			metricName: "vpa_updater_vpas_with_evictable_pods_total",
			additions: []addition{
				{1, vpa_types.UpdateModeOff, 1},
				{2, vpa_types.UpdateModeAuto, 1},
			},
			expectedMetrics: []expectation{
				{[]string{"0" /* log2(1) */, "Off"}, 1},
				{[]string{"1" /* log2(2) */, "Auto"}, 1},
			},
		},
		{
			desc:       "VpasWithEvictedPodsCounter",
			newCounter: NewVpasWithEvictedPodsCounter,
			metric:     vpasWithEvictedPodsCount,
			metricName: "vpa_updater_vpas_with_evicted_pods_total",
			additions: []addition{
				{1, vpa_types.UpdateModeAuto, 2},
				{1, vpa_types.UpdateModeAuto, 3},
			},
			expectedMetrics: []expectation{
				{[]string{"0" /* log2(1) */, "Auto"}, 5},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Cleanup(tc.metric.Reset)
			counter := tc.newCounter()
			for _, add := range tc.additions {
				counter.Add(add.vpaSize, add.mode, add.value)
			}
			counter.Observe()
			for _, expected := range tc.expectedMetrics {
				val := testutil.ToFloat64(tc.metric.WithLabelValues(expected.labels...))
				if val != expected.value {
					t.Errorf("Unexpected value for metric %s with labels %v: got %v, want %v", tc.metricName, expected.labels, val, expected.value)
				}
			}
		})
	}
}

func TestSizeBasedGauge(t *testing.T) {
	type addition struct {
		vpaSize int
		value   int
	}
	type expectation struct {
		labels []string
		value  float64
	}
	testCases := []struct {
		desc            string
		newCounter      func() *SizeBasedGauge
		metric          *prometheus.GaugeVec
		metricName      string
		additions       []addition
		expectedMetrics []expectation
	}{
		{
			desc:       "InPlaceUpdatablePodsCounter",
			newCounter: NewInPlaceUpdatablePodsCounter,
			metric:     inPlaceUpdatableCount,
			metricName: "vpa_updater_in_place_updatable_pods_total",
			additions: []addition{
				{1, 5},
				{2, 10},
			},
			expectedMetrics: []expectation{
				{[]string{"0" /* log2(1) */}, 5},
				{[]string{"1" /* log2(2) */}, 10},
			},
		},
		{
			desc:       "VpasWithInPlaceUpdatablePodsCounter",
			newCounter: NewVpasWithInPlaceUpdatablePodsCounter,
			metric:     vpasWithInPlaceUpdatablePodsCount,
			metricName: "vpa_updater_vpas_with_in_place_updatable_pods_total",
			additions: []addition{
				{10, 1},
				{20, 1},
			},
			expectedMetrics: []expectation{
				{[]string{"3" /* log2(10) */}, 1},
				{[]string{"4" /* log2(20) */}, 1},
			},
		},
		{
			desc:       "VpasWithInPlaceUpdatedPodsCounter",
			newCounter: NewVpasWithInPlaceUpdatedPodsCounter,
			metric:     vpasWithInPlaceUpdatedPodsCount,
			metricName: "vpa_updater_vpas_with_in_place_updated_pods_total",
			additions: []addition{
				{2, 4},
				{4, 5},
			},
			expectedMetrics: []expectation{
				{[]string{"1" /* log2(2) */}, 4},
				{[]string{"2" /* log2(4) */}, 5},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Cleanup(tc.metric.Reset)
			counter := tc.newCounter()
			for _, add := range tc.additions {
				counter.Add(add.vpaSize, add.value)
			}
			counter.Observe()
			for _, expected := range tc.expectedMetrics {
				val := testutil.ToFloat64(tc.metric.WithLabelValues(expected.labels...))
				if val != expected.value {
					t.Errorf("Unexpected value for metric %s with labels %v: got %v, want %v", tc.metricName, expected.labels, val, expected.value)
				}
			}
		})
	}
}
