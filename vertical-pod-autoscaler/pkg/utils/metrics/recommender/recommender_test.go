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
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	apiv1 "k8s.io/api/core/v1"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

func TestObjectCounter(t *testing.T) {
	updateModeOff := vpa_types.UpdateModeOff
	updateModeInitial := vpa_types.UpdateModeInitial
	updateModeRecreate := vpa_types.UpdateModeRecreate
	updateModeAuto := vpa_types.UpdateModeAuto
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
				"api=,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
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
				"api=v1beta1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
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
				"api=v1beta2,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
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
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
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
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
			},
		},
		{
			name: "report update mode auto",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					UpdateMode: &updateModeAuto,
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
			},
		},
		{
			name: "report update mode initial",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					UpdateMode: &updateModeInitial,
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Initial,": 1,
			},
		},
		{
			name: "report update mode recreate",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					UpdateMode: &updateModeRecreate,
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Recreate,": 1,
			},
		},
		{
			name: "report update mode off",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					UpdateMode: &updateModeOff,
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Off,": 1,
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
				{
					APIVersion:     "v1",
					Recommendation: nil,
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
			},
		},
		{
			name: "report has recommendation as false on missing container recommendations",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					Recommendation: &vpa_types.RecommendedPodResources{
						ContainerRecommendations: nil,
					},
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
			},
		},
		{
			name: "report has recommendation as true on existing container recommendations",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					Recommendation: &vpa_types.RecommendedPodResources{
						ContainerRecommendations: []vpa_types.RecommendedContainerResources{{}},
					},
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=true,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
			},
		},
		{
			name: "report has matches pods as true on missing condition",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					Conditions: nil,
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
			},
		},
		{
			name: "report has matches pods as false on NoPodsMatched condition",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					Conditions: map[vpa_types.VerticalPodAutoscalerConditionType]vpa_types.VerticalPodAutoscalerCondition{
						vpa_types.NoPodsMatched: {
							Status: apiv1.ConditionTrue,
						},
					},
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=false,unsupported_config=false,update_mode=Auto,": 1,
			},
		},
		{
			name: "report unsupported config as false on missing condition",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					Conditions: nil,
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=false,update_mode=Auto,": 1,
			},
		},
		{
			name: "report unsupported config as true on ConfigUnsupported condition",
			add: []*model.Vpa{
				{
					APIVersion: "v1",
					Conditions: map[vpa_types.VerticalPodAutoscalerConditionType]vpa_types.VerticalPodAutoscalerCondition{
						vpa_types.ConfigUnsupported: {
							Status: apiv1.ConditionTrue,
						},
					},
				},
			},
			wantMetrics: map[string]float64{
				"api=v1,has_recommendation=false,matches_pods=true,unsupported_config=true,update_mode=Auto,": 1,
			},
		},
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
