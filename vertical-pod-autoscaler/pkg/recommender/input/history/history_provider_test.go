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

package history

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

const (
	cpuQuery              = "rate(container_cpu_usage_seconds_total{job=\"kubernetes-cadvisor\", pod_name=~\".+\", name!=\"POD\", name!=\"\"}[30s])"
	memoryQuery           = "container_memory_working_set_bytes{job=\"kubernetes-cadvisor\", pod_name=~\".+\", name!=\"POD\", name!=\"\"}"
	labelsQuery           = "up{job=\"kubernetes-pods\"}"
	cpuNamespacedQuery    = "rate(container_cpu_usage_seconds_total{job=\"kubernetes-cadvisor\", pod_name=~\".+\", name!=\"POD\", name!=\"\", namespace=\"kube-system\"}[30s])"
	memoryNamespacedQuery = "container_memory_working_set_bytes{job=\"kubernetes-cadvisor\", pod_name=~\".+\", name!=\"POD\", name!=\"\", namespace=\"kube-system\"}"
)

func getDefaultPrometheusHistoryProviderConfigForTest() PrometheusHistoryProviderConfig {
	return PrometheusHistoryProviderConfig{
		Address:                "",
		HistoryLength:          "8d",
		HistoryResolution:      "30s",
		PodLabelPrefix:         "pod_label_",
		PodLabelsMetricName:    "up{job=\"kubernetes-pods\"}",
		PodNamespaceLabel:      "kubernetes_namespace",
		PodNameLabel:           "kubernetes_pod_name",
		CtrNamespaceLabel:      "namespace",
		CtrPodNameLabel:        "pod_name",
		CtrNameLabel:           "name",
		CadvisorMetricsJobName: "kubernetes-cadvisor",
	}
}

type mockPrometheusAPI struct {
	mock.Mock
}

func (m mockPrometheusAPI) AlertManagers(ctx context.Context) (prometheusv1.AlertManagersResult, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) Alerts(ctx context.Context) (prometheusv1.AlertsResult, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) CleanTombstones(ctx context.Context) error {
	panic("not implemented")
}

func (m mockPrometheusAPI) Config(ctx context.Context) (prometheusv1.ConfigResult, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) DeleteSeries(ctx context.Context, matches []string, startTime time.Time, endTime time.Time) error {
	panic("not implemented")
}

func (m mockPrometheusAPI) Flags(ctx context.Context) (prometheusv1.FlagsResult, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) LabelNames(ctx context.Context) ([]string, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) LabelValues(ctx context.Context, label string) (prommodel.LabelValues, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) Series(ctx context.Context, matches []string, startTime time.Time, endTime time.Time) ([]prommodel.LabelSet, api.Warnings, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) Rules(ctx context.Context) (prometheusv1.RulesResult, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) Snapshot(ctx context.Context, skipHead bool) (prometheusv1.SnapshotResult, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) Targets(ctx context.Context) (prometheusv1.TargetsResult, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) TargetsMetadata(ctx context.Context, _, _, _ string) ([]prometheusv1.MetricMetadata, error) {
	panic("not implemented")
}

func (m mockPrometheusAPI) Query(ctx context.Context, query string, ts time.Time) (prommodel.Value, api.Warnings, error) {
	args := m.Called(ctx, query, ts)
	var returnArg prommodel.Value
	if args.Get(0) != nil {
		returnArg = args.Get(0).(prommodel.Value)
	}
	return returnArg, nil, args.Error(1)
}

func (m mockPrometheusAPI) QueryRange(ctx context.Context, query string, r prometheusv1.Range) (prommodel.Value, api.Warnings, error) {
	args := m.Called(ctx, query, r)
	var returnArg prommodel.Value
	if args.Get(0) != nil {
		returnArg = args.Get(0).(prommodel.Value)
	}
	return returnArg, nil, args.Error(1)
}

func TestGetEmptyClusterHistory(t *testing.T) {
	mockClient := mockPrometheusAPI{}
	historyProvider := prometheusHistoryProvider{
		config:           getDefaultPrometheusHistoryProviderConfigForTest(),
		prometheusClient: &mockClient,
	}
	mockClient.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Times(1).Return(
		prommodel.Matrix{}, nil)
	mockClient.On("QueryRange", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).Return().Times(2).Return(
		prommodel.Matrix{}, nil)
	tss, err := historyProvider.GetClusterHistory()
	assert.Nil(t, err)
	assert.NotNil(t, tss)
	assert.Empty(t, tss)
}

func TestPrometheusError(t *testing.T) {
	mockClient := mockPrometheusAPI{}
	historyProvider := prometheusHistoryProvider{
		config:           getDefaultPrometheusHistoryProviderConfigForTest(),
		prometheusClient: &mockClient,
	}
	mockClient.On("QueryRange", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("v1.Range")).Return().Times(2).Return(
		nil, fmt.Errorf("bla"))
	_, err := historyProvider.GetClusterHistory()
	assert.NotNil(t, err)
}

func TestGetCPUSamples(t *testing.T) {
	mockClient := mockPrometheusAPI{}
	historyProvider := prometheusHistoryProvider{
		config:           getDefaultPrometheusHistoryProviderConfigForTest(),
		prometheusClient: &mockClient,
	}
	mockClient.On("QueryRange", mock.Anything, cpuQuery, mock.AnythingOfType("v1.Range")).Return().Return(
		prommodel.Matrix{
			{
				Metric: map[prommodel.LabelName]prommodel.LabelValue{
					"namespace": "default",
					"pod_name":  "pod",
					"name":      "container",
				},
				Values: []prommodel.SamplePair{
					{
						Timestamp: prommodel.TimeFromUnix(1),
						Value:     5.5,
					},
				},
			},
		},
		nil)
	mockClient.On("QueryRange", mock.Anything, memoryQuery, mock.AnythingOfType("v1.Range")).Return().Return(prommodel.Matrix{}, nil)
	mockClient.On("Query", mock.Anything, labelsQuery, mock.AnythingOfType("time.Time")).Return(prommodel.Matrix{}, nil)
	podID := model.PodID{Namespace: "default", PodName: "pod"}
	podHistory := &PodHistory{
		LastLabels: map[string]string{},
		Samples: map[string][]model.ContainerUsageSample{"container": {{
			MeasureStart: time.Unix(1, 0),
			Usage:        model.CPUAmountFromCores(5.5),
			Resource:     model.ResourceCPU,
		}}},
	}
	histories, err := historyProvider.GetClusterHistory()
	assert.Nil(t, err)
	assert.Equal(t, histories, map[model.PodID]*PodHistory{podID: podHistory})
}

func TestGetMemorySamples(t *testing.T) {
	mockClient := mockPrometheusAPI{}
	historyProvider := prometheusHistoryProvider{
		config:           getDefaultPrometheusHistoryProviderConfigForTest(),
		prometheusClient: &mockClient,
	}
	mockClient.On("QueryRange", mock.Anything, cpuQuery, mock.AnythingOfType("v1.Range")).Return().Return(prommodel.Matrix{}, nil)
	mockClient.On("QueryRange", mock.Anything, memoryQuery, mock.AnythingOfType("v1.Range")).Return().Return(
		prommodel.Matrix{
			{
				Metric: map[prommodel.LabelName]prommodel.LabelValue{
					"namespace": "default",
					"pod_name":  "pod",
					"name":      "container",
				},
				Values: []prommodel.SamplePair{
					{
						Timestamp: prommodel.TimeFromUnix(1),
						Value:     12345,
					},
				},
			},
		}, nil)
	mockClient.On("Query", mock.Anything, labelsQuery, mock.AnythingOfType("time.Time")).Return(prommodel.Matrix{}, nil)
	podID := model.PodID{Namespace: "default", PodName: "pod"}
	podHistory := &PodHistory{
		LastLabels: map[string]string{},
		Samples: map[string][]model.ContainerUsageSample{"container": {{
			MeasureStart: time.Unix(1, 0),
			Usage:        model.MemoryAmountFromBytes(12345),
			Resource:     model.ResourceMemory,
		}}},
	}
	histories, err := historyProvider.GetClusterHistory()
	assert.Nil(t, err)
	assert.Equal(t, histories, map[model.PodID]*PodHistory{podID: podHistory})
}

func TestGetNamespacedMemorySamples(t *testing.T) {
	mockClient := mockPrometheusAPI{}
	promConfig := getDefaultPrometheusHistoryProviderConfigForTest()
	promConfig.Namespace = "kube-system"
	historyProvider := prometheusHistoryProvider{
		config:           promConfig,
		prometheusClient: &mockClient,
	}

	mockClient.On("QueryRange", mock.Anything, cpuNamespacedQuery, mock.AnythingOfType("v1.Range")).Return().Return(prommodel.Matrix{}, nil)
	mockClient.On("QueryRange", mock.Anything, memoryNamespacedQuery, mock.AnythingOfType("v1.Range")).Return().Return(
		prommodel.Matrix{
			{
				Metric: map[prommodel.LabelName]prommodel.LabelValue{
					"namespace": "kube-system",
					"pod_name":  "pod",
					"name":      "container",
				},
				Values: []prommodel.SamplePair{
					{
						Timestamp: prommodel.TimeFromUnix(1),
						Value:     12345,
					},
				},
			},
		}, nil)
	mockClient.On("Query", mock.Anything, labelsQuery, mock.AnythingOfType("time.Time")).Return(prommodel.Matrix{}, nil)
	podID := model.PodID{Namespace: "kube-system", PodName: "pod"}
	podHistory := &PodHistory{
		LastLabels: map[string]string{},
		Samples: map[string][]model.ContainerUsageSample{"container": {{
			MeasureStart: time.Unix(1, 0),
			Usage:        model.MemoryAmountFromBytes(12345),
			Resource:     model.ResourceMemory,
		}}},
	}
	histories, err := historyProvider.GetClusterHistory()
	assert.Nil(t, err)
	assert.Equal(t, histories, map[model.PodID]*PodHistory{podID: podHistory})
}

func TestGetLabels(t *testing.T) {
	mockClient := mockPrometheusAPI{}
	historyProvider := prometheusHistoryProvider{
		config:           getDefaultPrometheusHistoryProviderConfigForTest(),
		prometheusClient: &mockClient,
	}
	mockClient.On("QueryRange", mock.Anything, cpuQuery, mock.AnythingOfType("v1.Range")).Return().Return(prommodel.Matrix{}, nil)
	mockClient.On("QueryRange", mock.Anything, memoryQuery, mock.AnythingOfType("v1.Range")).Return().Return(prommodel.Matrix{}, nil)
	mockClient.On("Query", mock.Anything, labelsQuery, mock.AnythingOfType("time.Time")).Return(
		prommodel.Matrix{
			{
				Metric: map[prommodel.LabelName]prommodel.LabelValue{
					"kubernetes_namespace": "default",
					"kubernetes_pod_name":  "pod",
					"pod_label_x":          "y",
				},
				Values: []prommodel.SamplePair{
					{
						Timestamp: prommodel.TimeFromUnix(1),
						Value:     12345,
					},
				},
			},
			{
				Metric: map[prommodel.LabelName]prommodel.LabelValue{
					"kubernetes_namespace": "default",
					"kubernetes_pod_name":  "pod",
					"pod_label_x":          "z",
				},
				Values: []prommodel.SamplePair{
					{
						Timestamp: prommodel.TimeFromUnix(20),
						Value:     12345,
					},
				},
			},
		}, nil)
	podID := model.PodID{Namespace: "default", PodName: "pod"}
	podHistory := &PodHistory{
		LastLabels: map[string]string{"x": "z"},
		LastSeen:   time.Unix(20, 0),
		Samples:    map[string][]model.ContainerUsageSample{},
	}
	histories, err := historyProvider.GetClusterHistory()
	assert.Nil(t, err)
	assert.Equal(t, histories, map[model.PodID]*PodHistory{podID: podHistory})
}
