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
	"k8s.io/klog"
	"sort"
	"strings"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// PrometheusHistoryProviderConfig allow to select which metrics
// should be queried to get real resource utilization.
type PrometheusHistoryProviderConfig struct {
	Address                                          string
	QueryTimeout                                     time.Duration
	HistoryLength, HistoryResolution                 string
	PodLabelPrefix, PodLabelsMetricName              string
	PodNamespaceLabel, PodNameLabel                  string
	CtrNamespaceLabel, CtrPodNameLabel, CtrNameLabel string
	CadvisorMetricsJobName                           string
}

// PodHistory represents history of usage and labels for a given pod.
type PodHistory struct {
	// Current samples if pod is still alive, last known samples otherwise.
	LastLabels map[string]string
	LastSeen   time.Time
	// A map for container name to a list of its usage samples, in chronological
	// order.
	Samples map[string][]model.ContainerUsageSample
}

func newEmptyHistory() *PodHistory {
	return &PodHistory{LastLabels: map[string]string{}, Samples: map[string][]model.ContainerUsageSample{}}
}

// HistoryProvider gives history of all pods in a cluster.
// TODO(schylek): this interface imposes how history is represented which doesn't work well with checkpoints.
// Consider refactoring to passing ClusterState and create history provider working with checkpoints.
type HistoryProvider interface {
	GetClusterHistory() (map[model.PodID]*PodHistory, error)
}

type prometheusHistoryProvider struct {
	prometheusClient  prometheusv1.API
	config            PrometheusHistoryProviderConfig
	queryTimeout      time.Duration
	historyDuration   prommodel.Duration
	historyResolution prommodel.Duration
}

// NewPrometheusHistoryProvider contructs a history provider that gets data from Prometheus.
func NewPrometheusHistoryProvider(config PrometheusHistoryProviderConfig) (HistoryProvider, error) {
	promClient, err := promapi.NewClient(promapi.Config{
		Address: config.Address,
	})
	if err != nil {
		return &prometheusHistoryProvider{}, err
	}

	// Use Prometheus's model.Duration; this can additionally parse durations in days, weeks and years (as well as seconds, minutes, hours etc)
	historyDuration, err := prommodel.ParseDuration(config.HistoryLength)
	if err != nil {
		return &prometheusHistoryProvider{}, fmt.Errorf("history length %s is not a valid Prometheus duration: %v", config.HistoryLength, err)
	}

	historyResolution, err := prommodel.ParseDuration(config.HistoryResolution)
	if err != nil {
		return &prometheusHistoryProvider{}, fmt.Errorf("history resolution %s is not a valid Prometheus duration: %v", config.HistoryResolution, err)
	}

	return &prometheusHistoryProvider{
		prometheusClient:  prometheusv1.NewAPI(promClient),
		config:            config,
		queryTimeout:      config.QueryTimeout,
		historyDuration:   historyDuration,
		historyResolution: historyResolution,
	}, nil
}

func (p *prometheusHistoryProvider) getContainerIDFromLabels(metric prommodel.Metric) (*model.ContainerID, error) {
	labels := promMetricToLabelMap(metric)
	namespace, ok := labels[p.config.CtrNamespaceLabel]
	if !ok {
		return nil, fmt.Errorf("no %s label", p.config.CtrNamespaceLabel)
	}
	podName, ok := labels[p.config.CtrPodNameLabel]
	if !ok {
		return nil, fmt.Errorf("no %s label", p.config.CtrPodNameLabel)
	}
	containerName, ok := labels[p.config.CtrNameLabel]
	if !ok {
		return nil, fmt.Errorf("no %s label on container data", p.config.CtrNameLabel)
	}
	return &model.ContainerID{
		PodID: model.PodID{
			Namespace: namespace,
			PodName:   podName},
		ContainerName: containerName}, nil
}

func (p *prometheusHistoryProvider) getPodIDFromLabels(metric prommodel.Metric) (*model.PodID, error) {
	labels := promMetricToLabelMap(metric)
	namespace, ok := labels[p.config.PodNamespaceLabel]
	if !ok {
		return nil, fmt.Errorf("no %s label", p.config.PodNamespaceLabel)
	}
	podName, ok := labels[p.config.PodNameLabel]
	if !ok {
		return nil, fmt.Errorf("no %s label", p.config.PodNameLabel)
	}
	return &model.PodID{Namespace: namespace, PodName: podName}, nil
}

func (p *prometheusHistoryProvider) getPodLabelsMap(metric prommodel.Metric) map[string]string {
	podLabels := make(map[string]string)
	for key, value := range metric {
		podLabelKey := strings.TrimPrefix(string(key), p.config.PodLabelPrefix)
		if podLabelKey != string(key) {
			podLabels[podLabelKey] = string(value)
		}
	}
	return podLabels
}

func promMetricToLabelMap(metric prommodel.Metric) map[string]string {
	labels := map[string]string{}
	for k, v := range metric {
		labels[string(k)] = string(v)
	}
	return labels
}

func resourceAmountFromValue(value float64, resource model.ResourceName) model.ResourceAmount {
	// This assumes CPU value is in cores and memory in bytes, which is true
	// for the metrics this class queries from Prometheus.
	switch resource {
	case model.ResourceCPU:
		return model.CPUAmountFromCores(value)
	case model.ResourceMemory:
		return model.MemoryAmountFromBytes(value)
	}
	return model.ResourceAmount(0)
}

func getContainerUsageSamplesFromSamples(samples []prommodel.SamplePair, resource model.ResourceName) []model.ContainerUsageSample {
	res := make([]model.ContainerUsageSample, 0)
	for _, sample := range samples {
		res = append(res, model.ContainerUsageSample{
			MeasureStart: sample.Timestamp.Time(),
			Usage:        resourceAmountFromValue(float64(sample.Value), resource),
			Resource:     resource})
	}
	return res
}

func (p *prometheusHistoryProvider) readResourceHistory(res map[model.PodID]*PodHistory, query string, resource model.ResourceName) error {
	end := time.Now()
	start := end.Add(-time.Duration(p.historyDuration))

	ctx, cancel := context.WithTimeout(context.Background(), p.queryTimeout)
	defer cancel()

	result, err := p.prometheusClient.QueryRange(ctx, query, prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  time.Duration(p.historyResolution),
	})

	if err != nil {
		return fmt.Errorf("cannot get timeseries for %v: %v", resource, err)
	}

	matrix, ok := result.(prommodel.Matrix)
	if !ok {
		return fmt.Errorf("expected query to return a matrix; got result type %T", result)
	}

	for _, ts := range matrix {
		containerID, err := p.getContainerIDFromLabels(ts.Metric)
		if err != nil {
			return fmt.Errorf("cannot get container ID from labels: %v", ts.Metric)
		}

		newSamples := getContainerUsageSamplesFromSamples(ts.Values, resource)
		podHistory, ok := res[containerID.PodID]
		if !ok {
			podHistory = newEmptyHistory()
			res[containerID.PodID] = podHistory
		}
		podHistory.Samples[containerID.ContainerName] = append(
			podHistory.Samples[containerID.ContainerName],
			newSamples...)
	}
	return nil
}

func (p *prometheusHistoryProvider) readLastLabels(res map[model.PodID]*PodHistory, query string) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.queryTimeout)
	defer cancel()

	result, err := p.prometheusClient.Query(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("cannot get timeseries for labels: %v", err)
	}

	matrix, ok := result.(prommodel.Matrix)
	if !ok {
		return fmt.Errorf("expected query to return a matrix; got result type %T", result)
	}

	for _, ts := range matrix {
		podID, err := p.getPodIDFromLabels(ts.Metric)
		if err != nil {
			return fmt.Errorf("cannot get container ID from labels %v: %v", ts.Metric, err)
		}
		podHistory, ok := res[*podID]
		if !ok {
			podHistory = newEmptyHistory()
			res[*podID] = podHistory
		}
		podLabels := p.getPodLabelsMap(ts.Metric)

		// time series results will always be sorted chronologically from oldest to
		// newest, so the last element is the latest sample
		lastSample := ts.Values[len(ts.Values)-1]
		if lastSample.Timestamp.Time().After(podHistory.LastSeen) {
			podHistory.LastSeen = lastSample.Timestamp.Time()
			podHistory.LastLabels = podLabels
		}
	}
	return nil
}

func (p *prometheusHistoryProvider) GetClusterHistory() (map[model.PodID]*PodHistory, error) {
	res := make(map[model.PodID]*PodHistory)
	var podSelector string
	if p.config.CadvisorMetricsJobName != "" {
		podSelector = fmt.Sprintf("job=\"%s\", ", p.config.CadvisorMetricsJobName)
	}
	podSelector = podSelector + fmt.Sprintf("%s=~\".+\", %s!=\"POD\", %s!=\"\"",
		p.config.CtrPodNameLabel, p.config.CtrNameLabel, p.config.CtrNameLabel)

	historicalCpuQuery := fmt.Sprintf("rate(container_cpu_usage_seconds_total{%s}[%s])", podSelector, p.config.HistoryResolution)
	klog.V(4).Infof("Historical CPU usage query used: %s", historicalCpuQuery)
	err := p.readResourceHistory(res, historicalCpuQuery, model.ResourceCPU)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}

	historicalMemoryQuery := fmt.Sprintf("container_memory_working_set_bytes{%s}", podSelector)
	klog.V(4).Infof("Historical memory usage query used: %s", historicalMemoryQuery)
	err = p.readResourceHistory(res, historicalMemoryQuery, model.ResourceMemory)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}
	for _, podHistory := range res {
		for _, samples := range podHistory.Samples {
			sort.Slice(samples, func(i, j int) bool { return samples[i].MeasureStart.Before(samples[j].MeasureStart) })
		}
	}
	p.readLastLabels(res, p.config.PodLabelsMetricName)
	return res, nil
}
