/*
Copyright 2018 The Kubernetes Authors.

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

// Package recommender (aka metrics_recommender) - code for metrics of VPA Recommender
package recommender

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
)

const (
	metricsNamespace = metrics.TopMetricsNamespace + "recommender"
)

var (
	modes = []string{string(vpa_types.UpdateModeOff), string(vpa_types.UpdateModeInitial), string(vpa_types.UpdateModeRecreate), string(vpa_types.UpdateModeAuto)}
)

type apiVersion string

const (
	v1beta1 apiVersion = "v1beta1"
	v1beta2 apiVersion = "v1beta2"
	v1      apiVersion = "v1"
)

var (
	vpaObjectCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "vpa_objects_count",
			Help:      "Number of VPA objects present in the cluster.",
		}, []string{"update_mode", "has_recommendation", "api", "matches_pods", "unsupported_config"},
	)

	recommendationLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "recommendation_latency_seconds",
			Help:      "Time elapsed from creating a valid VPA configuration to the first recommendation.",
			Buckets:   []float64{1.0, 2.0, 5.0, 7.5, 10.0, 20.0, 30.0, 40.00, 50.0, 60.0, 90.0, 120.0, 150.0, 180.0, 240.0, 300.0, 600.0, 900.0, 1800.0},
		},
	)

	functionLatency = metrics.CreateExecutionTimeMetric(metricsNamespace,
		"Time spent in various parts of VPA Recommender main loop.")

	aggregateContainerStatesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "aggregate_container_states_count",
			Help:      "Number of aggregate container states being tracked by the recommender",
		},
	)
)

type objectCounterKey struct {
	mode              string
	has               bool
	matchesPods       bool
	apiVersion        apiVersion
	unsupportedConfig bool
}

// ObjectCounter helps split all VPA objects into buckets
type ObjectCounter struct {
	cnt map[objectCounterKey]int
}

// Register initializes all metrics for VPA Recommender
func Register() {
	prometheus.MustRegister(vpaObjectCount, recommendationLatency, functionLatency, aggregateContainerStatesCount)
}

// NewExecutionTimer provides a timer for Recommender's RunOnce execution
func NewExecutionTimer() *metrics.ExecutionTimer {
	return metrics.NewExecutionTimer(functionLatency)
}

// ObserveRecommendationLatency observes the time it took for the first recommendation to appear
func ObserveRecommendationLatency(created time.Time) {
	recommendationLatency.Observe(time.Now().Sub(created).Seconds())
}

// RecordAggregateContainerStatesCount records the number of containers being tracked by the recommender
func RecordAggregateContainerStatesCount(statesCount int) {
	aggregateContainerStatesCount.Set(float64(statesCount))
}

// NewObjectCounter creates a new helper to split VPA objects into buckets
func NewObjectCounter() *ObjectCounter {
	obj := ObjectCounter{
		cnt: make(map[objectCounterKey]int),
	}

	// initialize with empty data so we can clean stale gauge values in Observe
	for _, m := range modes {
		for _, h := range []bool{false, true} {
			for _, api := range []apiVersion{v1beta1, v1beta2, v1} {
				for _, mp := range []bool{false, true} {
					for _, uc := range []bool{false, true} {
						obj.cnt[objectCounterKey{
							mode:              m,
							has:               h,
							apiVersion:        api,
							matchesPods:       mp,
							unsupportedConfig: uc,
						}] = 0
					}
				}
			}
		}
	}

	return &obj
}

// Add updates the helper state to include the given VPA object
func (oc *ObjectCounter) Add(vpa *model.Vpa) {
	mode := string(vpa_types.UpdateModeAuto)
	if vpa.UpdateMode != nil && string(*vpa.UpdateMode) != "" {
		mode = string(*vpa.UpdateMode)
	}
	// TODO: Maybe report v1 version as well.
	api := v1beta2
	if vpa.IsV1Beta1API {
		api = v1beta1
	}

	key := objectCounterKey{
		mode:              mode,
		has:               vpa.HasRecommendation(),
		apiVersion:        api,
		matchesPods:       vpa.HasMatchedPods(),
		unsupportedConfig: vpa.Conditions.ConditionActive(vpa_types.ConfigUnsupported),
	}
	oc.cnt[key]++
}

// Observe passes all the computed bucket values to metrics
func (oc *ObjectCounter) Observe() {
	for k, v := range oc.cnt {
		vpaObjectCount.WithLabelValues(
			k.mode,
			fmt.Sprintf("%v", k.has),
			string(k.apiVersion),
			fmt.Sprintf("%v", k.matchesPods),
			fmt.Sprintf("%v", k.unsupportedConfig),
		).Set(float64(v))
	}
}
