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

	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
)

const (
	historyLength  = "8d"
	podLabelPrefix = "pod_label_"
)

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
type HistoryProvider interface {
	GetClusterHistory() (map[model.PodID]*PodHistory, error)
}

type prometheusHistoryProvider struct {
	prometheusClient PrometheusClient
}

// NewPrometheusHistoryProvider contructs a history provider that gets data from Prometheus.
func NewPrometheusHistoryProvider(prometheusAddress string) HistoryProvider {
	return &prometheusHistoryProvider{
		prometheusClient: NewPrometheusClient(&http.Client{}, prometheusAddress),
	}
}

func getContainerIDFromLabels(labels map[string]string) (*model.ContainerID, error) {
	namespace, ok := labels["namespace"]
	if !ok {
		return nil, fmt.Errorf("no namespace label")
	}
	podName, ok := labels["pod_name"]
	if !ok {
		return nil, fmt.Errorf("no pod_name label")
	}
	containerName, ok := labels["name"]
	if !ok {
		return nil, fmt.Errorf("no name label on container data")
	}
	return &model.ContainerID{
		PodID: model.PodID{
			Namespace: namespace,
			PodName:   podName},
		ContainerName: containerName}, nil
}

func getPodIDFromLabels(labels map[string]string) (*model.PodID, error) {
	namespace, ok := labels["kubernetes_namespace"]
	if !ok {
		return nil, fmt.Errorf("no kubernetes_namespace label")
	}
	podName, ok := labels["kubernetes_pod_name"]
	if !ok {
		return nil, fmt.Errorf("no kubernetes_pod_name label")
	}
	return &model.PodID{Namespace: namespace, PodName: podName}, nil
}

func getPodLabelsMap(metricLabels map[string]string) map[string]string {
	podLabels := make(map[string]string)
	for key, value := range metricLabels {
		podLabelKey := strings.TrimPrefix(key, podLabelPrefix)
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
		containerID, err := getContainerIDFromLabels(ts.Labels)
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
		podID, err := getPodIDFromLabels(ts.Labels)
		if err != nil {
			return fmt.Errorf("cannot get container ID from labels: %v", ts.Labels)
		}
		podHistory, ok := res[*podID]
		if !ok {
			podHistory = newEmptyHistory()
			res[*podID] = podHistory
		}
		podLabels := getPodLabelsMap(ts.Labels)
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
	podSelector := "job=\"kubernetes-cadvisor\", pod_name=~\".+\""
	err := p.readResourceHistory(res, fmt.Sprintf("container_cpu_usage_seconds_total{%s}[%s]", podSelector, historyLength), model.ResourceCPU)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}
	err = p.readResourceHistory(res, fmt.Sprintf("container_memory_usage_bytes{%s}[%s]", podSelector, historyLength), model.ResourceMemory)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}
	for _, podHistory := range res {
		for _, samples := range podHistory.Samples {
			sort.Slice(samples, func(i, j int) bool { return samples[i].MeasureStart.Before(samples[j].MeasureStart) })
		}
	}
	p.readLastLabels(res, fmt.Sprintf("up{job=\"kubernetes-pods\"}[%s]", historyLength))
	return res, nil
}
