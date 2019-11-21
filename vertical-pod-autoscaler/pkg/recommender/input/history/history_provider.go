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
	"math"
	"sort"
	"strings"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/klog"
)

const Day = 24 * time.Hour

// PrometheusHistoryProviderConfig allow to select which metrics
// should be queried to get real resource utilization.
type PrometheusHistoryProviderConfig struct {
	Address                                                               string
	HistoryLength, HistoryResolution, PodLabelPrefix, PodLabelsMetricName string
	PodNamespaceLabel, PodNameLabel                                       string
	CtrNamespaceLabel, CtrPodNameLabel, CtrNameLabel                      string
	CadvisorMetricsJobName                                                string
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
	prometheusClient prometheusv1.API
	config           PrometheusHistoryProviderConfig
	queryTimeout     time.Duration
}

// NewPrometheusHistoryProvider contructs a history provider that gets data from Prometheus.
func NewPrometheusHistoryProvider(config PrometheusHistoryProviderConfig) HistoryProvider {
	promClient, err := promapi.NewClient(promapi.Config{
		Address: config.Address,
	})
	if err != nil {
		panic(err)
	}
	return &prometheusHistoryProvider{
		prometheusClient: prometheusv1.NewAPI(promClient),
		config:           config,
		queryTimeout:     5 * time.Minute,
	}
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

// splitTimeShards splits a larger range of time into smaller time periods
// (currently of a day each, but this may change in future).
// It returns a slice of v1.Ranges with the first Range starting at the given
// time and the last Range ending at the given end time.
// Each Range will be no more than a day.
func splitTimeShards(start, end time.Time, step time.Duration) []prometheusv1.Range {
	timeRange := end.Sub(start)

	if timeRange < Day {
		// no need to split up time range
		return []prometheusv1.Range{
			{
				Start: start,
				End:   end,
				Step:  step,
			},
		}
	}

	// split time range up into days, save the remainder
	days := int(math.Floor(timeRange.Hours() / 24.))
	remainder := timeRange - (time.Duration(days) * Day)

	res := []prometheusv1.Range{}

	rangeStart := start
	rangeEnd := start.Add(remainder)

	// If there's a remainder bit, put it first
	if int(remainder) > 0 {
		res = append(res, prometheusv1.Range{
			Start: rangeStart,
			End:   rangeEnd,
			Step:  step,
		})
	}

	// Add ranges of a day each
	for i := 0; i < days; i++ {
		rangeStart = rangeEnd
		rangeEnd = rangeStart.Add(Day)
		res = append(res, prometheusv1.Range{
			Start: rangeStart,
			End:   rangeEnd,
			Step:  step,
		})
	}

	return res
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
	// Use Prometheus's model.Duration; this can parse durations in days, weeks and years
	historyDuration, err := prommodel.ParseDuration(p.config.HistoryLength)
	if err != nil {
		return fmt.Errorf("history length %s is not a valid duration: %v", p.config.HistoryLength, err)
	}

	end := time.Now()
	start := end.Add(-time.Duration(historyDuration))

	step, err := prommodel.ParseDuration(p.config.HistoryResolution)
	if err != nil {
		return fmt.Errorf("history resolution %s is not a valid duration: %v", p.config.HistoryResolution, err)
	}

	for _, r := range splitTimeShards(start, end, time.Duration(step)) {
		ctx, cancel := context.WithTimeout(context.Background(), p.queryTimeout)
		defer cancel()

		klog.V(4).Infof("Running query for %s between %s and %s with step %s", query, r.Start.Format(time.RFC3339), r.End.Format(time.RFC3339), r.Step.String())
		v, err := p.prometheusClient.QueryRange(ctx, query, r)

		if err != nil {
			return fmt.Errorf("cannot get timeseries for %v: %v", resource, err)
		}

		matrix, ok := v.(prommodel.Matrix)
		if !ok {
			return fmt.Errorf("expected query to return a matrix; got result type %T", v)
		}
		klog.V(4).Infof("Got %d results for query", len(matrix))

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
	}
	return nil
}

func (p *prometheusHistoryProvider) readLastLabels(res map[model.PodID]*PodHistory, query string) error {
	historyDuration, err := prommodel.ParseDuration(p.config.HistoryLength)
	if err != nil {
		return fmt.Errorf("history length %s is not a valid duration: %v", p.config.HistoryLength, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.queryTimeout)
	defer cancel()

	end := time.Now()
	start := end.Add(-time.Duration(historyDuration))

	step, err := prommodel.ParseDuration(p.config.HistoryResolution)
	if err != nil {
		return fmt.Errorf("history resolution %s is not a valid duration: %v", p.config.HistoryResolution, err)
	}

	for _, r := range splitTimeShards(start, end, time.Duration(step)) {

		klog.V(4).Infof("Running query for %s between %s and %s with step %s", query, r.Start.Format(time.RFC3339), r.End.Format(time.RFC3339), r.Step.String())

		v, err := p.prometheusClient.QueryRange(ctx, query, r)
		if err != nil {
			return fmt.Errorf("cannot get timeseries for labels: %v", err)
		}

		matrix, ok := v.(prommodel.Matrix)
		if !ok {
			return fmt.Errorf("expected query to return a matrix; got result type %T", v)
		}
		klog.V(4).Infof("Got %d results for query", len(matrix))

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
			lastSample := ts.Values[len(ts.Values)-1]
			if lastSample.Timestamp.Time().After(podHistory.LastSeen) {
				podHistory.LastSeen = lastSample.Timestamp.Time()
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
	err := p.readResourceHistory(res, fmt.Sprintf("rate(container_cpu_usage_seconds_total{%s}[%s])", podSelector, p.config.HistoryResolution), model.ResourceCPU)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}
	err = p.readResourceHistory(res, fmt.Sprintf("container_memory_working_set_bytes{%s}", podSelector), model.ResourceMemory)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}
	for _, podHistory := range res {
		for _, samples := range podHistory.Samples {
			sort.Slice(samples, func(i, j int) bool { return samples[i].MeasureStart.Before(samples[j].MeasureStart) })
		}
	}
	err = p.readLastLabels(res, p.config.PodLabelsMetricName)
	if err != nil {
		return nil, err
	}

	return res, nil
}
