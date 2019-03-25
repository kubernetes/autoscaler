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
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// PrometheusHistoryProviderConfig allow to select which metrics
// should be queried to get real resource utilization.
type PrometheusHistoryProviderConfig struct {
	Address                                            string
	HistoryLength, PodLabelPrefix, PodLabelsMetricName string
	PodNamespaceLabel, PodNameLabel                    string
	CtrNamespaceLabel, CtrPodNameLabel, CtrNameLabel   string
	CadvisorMetricsJobName                             string
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
	prometheusClient PrometheusClient
	config           PrometheusHistoryProviderConfig
}

// NewPrometheusHistoryProvider contructs a history provider that gets data from Prometheus.
func NewPrometheusHistoryProvider(config PrometheusHistoryProviderConfig) HistoryProvider {
	return &prometheusHistoryProvider{
		prometheusClient: NewPrometheusClient(&http.Client{}, config.Address),
		config:           config,
	}
}

func (p *prometheusHistoryProvider) getContainerIDFromLabels(labels map[string]string) (*model.ContainerID, error) {
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

func (p *prometheusHistoryProvider) getPodIDFromLabels(labels map[string]string) (*model.PodID, error) {
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

func (p *prometheusHistoryProvider) getPodLabelsMap(metricLabels map[string]string) map[string]string {
	podLabels := make(map[string]string)
	for key, value := range metricLabels {
		podLabelKey := strings.TrimPrefix(key, p.config.PodLabelPrefix)
		if podLabelKey != key {
			podLabels[podLabelKey] = value
		}
	}
	return podLabels
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

func getContainerUsageSamplesFromSamples(samples []Sample, resource model.ResourceName) []model.ContainerUsageSample {
	res := make([]model.ContainerUsageSample, 0)
	for _, sample := range samples {
		res = append(res, model.ContainerUsageSample{
			MeasureStart: sample.Timestamp,
			Usage:        resourceAmountFromValue(sample.Value, resource),
			Resource:     resource})
	}
	return res
}

func (p *prometheusHistoryProvider) readResourceHistory(res map[model.PodID]*PodHistory, query string, resource model.ResourceName) error {
	tss, err := p.prometheusClient.GetTimeseries(query)
	if err != nil {
		return fmt.Errorf("cannot get timeseries for %v: %v", resource, err)
	}
	for _, ts := range tss {
		containerID, err := p.getContainerIDFromLabels(ts.Labels)
		if err != nil {
			return fmt.Errorf("cannot get container ID from labels: %v", ts.Labels)
		}
		newSamples := getContainerUsageSamplesFromSamples(ts.Samples, resource)
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
	tss, err := p.prometheusClient.GetTimeseries(query)
	if err != nil {
		return fmt.Errorf("cannot get timeseries for labels: %v", err)
	}
	for _, ts := range tss {
		podID, err := p.getPodIDFromLabels(ts.Labels)
		if err != nil {
			return fmt.Errorf("cannot get container ID from labels: %v", ts.Labels)
		}
		podHistory, ok := res[*podID]
		if !ok {
			podHistory = newEmptyHistory()
			res[*podID] = podHistory
		}
		podLabels := p.getPodLabelsMap(ts.Labels)
		for _, sample := range ts.Samples {
			if sample.Timestamp.After(podHistory.LastSeen) {
				podHistory.LastSeen = sample.Timestamp
				podHistory.LastLabels = podLabels
			}
		}
	}
	return nil
}

func (p *prometheusHistoryProvider) GetClusterHistory() (map[model.PodID]*PodHistory, error) {
	res := make(map[model.PodID]*PodHistory)
	podSelector := fmt.Sprintf("job=\"%s\", %s=~\".+\", %s!=\"POD\", %s!=\"\"",
		p.config.CadvisorMetricsJobName, p.config.CtrPodNameLabel,
		p.config.CtrNameLabel, p.config.CtrNameLabel)
	err := p.readResourceHistory(res, fmt.Sprintf("rate(container_cpu_usage_seconds_total{%s}[%s])", podSelector, p.config.HistoryLength), model.ResourceCPU)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}
	err = p.readResourceHistory(res, fmt.Sprintf("container_memory_usage_bytes{%s}[%s]", podSelector, p.config.HistoryLength), model.ResourceMemory)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}
	for _, podHistory := range res {
		for _, samples := range podHistory.Samples {
			sort.Slice(samples, func(i, j int) bool { return samples[i].MeasureStart.Before(samples[j].MeasureStart) })
		}
	}
	p.readLastLabels(res, fmt.Sprintf("%s[%s]", p.config.PodLabelsMetricName, p.config.HistoryLength))
	return res, nil
}
