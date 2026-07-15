/*
Copyright 2023 The Kubernetes Authors.

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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	corev1 "k8s.io/api/core/v1"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

func TestObserveRecommendationLatency(t *testing.T) {
	before := readHistogram(t, recommendationLatency)

	ObserveRecommendationLatency(time.Now().Add(-time.Second))

	after := readHistogram(t, recommendationLatency)
	if got, want := after.GetSampleCount(), before.GetSampleCount()+1; got != want {
		t.Errorf("unexpected sample count: got %d, want %d", got, want)
	}
	if elapsed := after.GetSampleSum() - before.GetSampleSum(); elapsed < 1 || elapsed > 30 {
		t.Errorf("unexpected observed recommendation latency: got %f seconds, want between 1 and 30 seconds", elapsed)
	}
}

func TestRecordAggregateContainerStatesCount(t *testing.T) {
	t.Cleanup(func() {
		aggregateContainerStatesCount.Set(0)
	})

	RecordAggregateContainerStatesCount(5)

	if got, want := testutil.ToFloat64(aggregateContainerStatesCount), float64(5); got != want {
		t.Errorf("unexpected aggregate container states count: got %f, want %f", got, want)
	}
}

func TestRecordMetricsServerResponse(t *testing.T) {
	metricServerResponses.Reset()
	t.Cleanup(metricServerResponses.Reset)

	RecordMetricsServerResponse(nil, "default-client")
	RecordMetricsServerResponse(nil, "default-client")
	RecordMetricsServerResponse(fmt.Errorf("metrics server unavailable"), "default-client")
	RecordMetricsServerResponse(fmt.Errorf("metrics server unavailable"), "fallback-client")

	cases := []struct {
		name       string
		isError    string
		clientName string
		want       float64
	}{
		{name: "successful default client responses", isError: "false", clientName: "default-client", want: 2},
		{name: "failed default client responses", isError: "true", clientName: "default-client", want: 1},
		{name: "failed fallback client responses", isError: "true", clientName: "fallback-client", want: 1},
		{name: "missing successful fallback client responses", isError: "false", clientName: "fallback-client", want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := testutil.ToFloat64(metricServerResponses.WithLabelValues(tc.isError, tc.clientName))
			if got != tc.want {
				t.Errorf("unexpected response count: got %f, want %f", got, tc.want)
			}
		})
	}
}

func TestObjectCounter(t *testing.T) {
	// We verify that other update modes are handled correctly as validation
	// may not happen if there are issues with the admission controller.
	updateModeUserDefined := vpa_types.UpdateMode("userDefined")

	cases := []struct {
		name        string
		add         []*model.Vpa
		wantMetrics map[string]float64
	}{
		{
			name: "set empty api on metric if it is missing on the VPA",
			add: []*model.Vpa{
				{
					APIVersion: "",
				},
			},
			wantMetrics: map[string]float64{
				"api=,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report api version v1beta1",
			add: []*model.Vpa{
				{
					APIVersion: "v1beta1",
				},
			},
			wantMetrics: map[string]float64{
				"api=v1beta1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report api version v1beta2",
			add: []*model.Vpa{
				{
					APIVersion: "v1beta2",
				},
			},
			wantMetrics: map[string]float64{
				"api=v1beta2,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report api version v1",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "default update mode to auto",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					UpdateMode: nil,
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report update mode user defined",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					UpdateMode: &updateModeUserDefined,
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=userDefined,": 1,
			},
		},
		{
			name: "report has recommendation as false on missing recommendations",
			add: []*model.Vpa{
				func() *model.Vpa {
					vpa := model.NewVpa(model.VpaID{}, nil, time.Time{})
					vpa.APIVersion = "v1"
					return vpa
				}(),
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report has recommendation as false on missing container recommendations",
			add: []*model.Vpa{
				func() *model.Vpa {
					vpa := model.NewVpa(model.VpaID{}, nil, time.Time{})
					vpa.APIVersion = "v1"
					vpa.SetRecommendationDirect(&vpa_types.RecommendedPodResources{
						ContainerRecommendations: nil,
					})
					return vpa
				}(),
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report has recommendation as true on existing container recommendations",
			add: []*model.Vpa{
				func() *model.Vpa {
					vpa := model.NewVpa(model.VpaID{}, nil, time.Time{})
					vpa.APIVersion = "v1"
					vpa.SetRecommendationDirect(&vpa_types.RecommendedPodResources{
						ContainerRecommendations: []vpa_types.RecommendedContainerResources{{}},
					})
					return vpa
				}(),
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=true,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report has matches pods as true on missing condition",
			add: []*model.Vpa{
				func() *model.Vpa {
					vpa := model.NewVpa(model.VpaID{}, nil, time.Time{})
					vpa.APIVersion = "v1"
					return vpa
				}(),
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report has matches pods as false on NoPodsMatched condition",
			add: []*model.Vpa{
				func() *model.Vpa {
					vpa := model.NewVpa(model.VpaID{}, nil, time.Time{})
					vpa.APIVersion = "v1"
					vpa.SetConditionsMap(map[vpa_types.VerticalPodAutoscalerConditionType]vpa_types.VerticalPodAutoscalerCondition{
						vpa_types.NoPodsMatched: {
							Status: corev1.ConditionTrue,
						},
					})
					return vpa
				}(),
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=false,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report unsupported config as false on missing condition",
			add: []*model.Vpa{
				func() *model.Vpa {
					vpa := model.NewVpa(model.VpaID{}, nil, time.Time{})
					vpa.APIVersion = "v1"
					return vpa
				}(),
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report unsupported config as true on ConfigUnsupported condition",
			add: []*model.Vpa{
				func() *model.Vpa {
					vpa := model.NewVpa(model.VpaID{}, nil, time.Time{})
					vpa.APIVersion = "v1"
					vpa.SetConditionsMap(map[vpa_types.VerticalPodAutoscalerConditionType]vpa_types.VerticalPodAutoscalerCondition{
						vpa_types.ConfigUnsupported: {
							Status: corev1.ConditionTrue,
						},
					})
					return vpa
				}(),
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=true,update_mode=Recreate,": 1,
			},
		},
	}

	for mode := range vpa_types.GetUpdateModes() {
		cases = append(cases, struct {
			name        string
			add         []*model.Vpa
			wantMetrics map[string]float64
		}{
			name: fmt.Sprintf("report update mode %s", mode),
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					UpdateMode: &mode,
				},
			},
			wantMetrics: map[string]float64{
				fmt.Sprintf("api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=%s,", mode): 1,
			},
		})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			counter := NewObjectCounter()
			for _, add := range tc.add {
				counter.Add(add)
			}

			t.Cleanup(func() {
				// Reset the metric after the test to avoid collisions.
				vpaObjectCount.Reset()
			})
			counter.Observe()

			metrics := make(chan prometheus.Metric)

			go func() {
				vpaObjectCount.Collect(metrics)
				close(metrics)
			}()

			gotMetrics := make(map[string]float64)
			for metric := range metrics {
				var metricProto dto.Metric
				if err := metric.Write(&metricProto); err != nil {
					t.Errorf("failed to write metric: %v", err)
				}

				key := labelsToKey(metricProto.GetLabel())
				gotMetrics[key] = *metricProto.GetGauge().Value
			}

			for wantKey, wantValue := range tc.wantMetrics {
				gotValue, gotKey := gotMetrics[wantKey]
				if !gotKey {
					t.Errorf("missing metrics sample %q, want value %f", wantKey, wantValue)
				}
				if gotValue != wantValue {
					t.Errorf("incorrect metrics sample %q, want value %f, got value %f", wantKey, wantValue, gotValue)
				}
			}

			// If a test case only covers specific metric cases we expect all other metrics to be 0.
			for gotKey, gotValue := range gotMetrics {
				_, wantKey := tc.wantMetrics[gotKey]
				if wantKey {
					continue
				}
				if gotValue != 0 {
					t.Errorf("incorrect metrics sample %q, want value %f, got value %f", gotKey, 0.0, gotValue)
				}
			}
		})
	}
}

func labelsToKey(labels []*dto.LabelPair) string {
	key := strings.Builder{}
	for _, label := range labels {
		key.WriteString(*label.Name)
		key.WriteRune('=')
		key.WriteString(*label.Value)
		key.WriteRune(',')
	}
	return key.String()
}

func readHistogram(t *testing.T, histogram prometheus.Histogram) *dto.Histogram {
	t.Helper()

	metric := &dto.Metric{}
	if err := histogram.Write(metric); err != nil {
		t.Fatalf("failed to read histogram: %v", err)
	}
	return metric.GetHistogram()
}

func TestObjectCounterResetsAllUpdateModes(t *testing.T) {
	for mode := range vpa_types.GetUpdateModes() {
		t.Run(string(mode), func(t *testing.T) {
			t.Cleanup(func() {
				vpaObjectCount.Reset()
			})

			key := "api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=" + string(mode) + ","

			// first loop add VPAs to increment the counter
			counter1 := NewObjectCounter()
			for range 3 {
				vpa := model.Vpa{
					APIVersion: "v1",
					UpdateMode: &mode,
				}
				counter1.Add(&vpa)
			}
			counter1.Observe()
			collectMetricsAndVerifyCount(t, key, 3)

			// next loop no VPAs
			counter2 := NewObjectCounter()
			counter2.Observe()
			collectMetricsAndVerifyCount(t, key, 0)
		})
	}
}

func collectMetricsAndVerifyCount(t *testing.T, key string, expectedCount float64) {
	metrics := make(chan prometheus.Metric)
	go func() {
		vpaObjectCount.Collect(metrics)
		close(metrics)
	}()

	liveMetrics := make(map[string]float64)
	for metric := range metrics {
		var metricProto dto.Metric
		if err := metric.Write(&metricProto); err != nil {
			t.Errorf("failed to write metric: %v", err)
		}
		liveMetrics[labelsToKey(metricProto.GetLabel())] = *metricProto.GetGauge().Value
	}

	if actualCount := liveMetrics[key]; actualCount != expectedCount {
		t.Errorf("key=%s expectedCount=%v actualCount=%v", key, expectedCount, actualCount)
	}
}
