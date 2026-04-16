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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
)

// PrometheusBasicAuthTransport injects basic auth headers into HTTP requests.
type PrometheusBasicAuthTransport struct {
	Username string
	Password string
	Base     http.RoundTripper
}

// RoundTrip function injects the username and password in the request's basic auth header
func (t *PrometheusBasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Use default transport if none specified
	rt := t.Base
	if rt == nil {
		rt = http.DefaultTransport
	}

	// Clone the request before modification to avoid data races and side effects.
	// Original http.Request contains shared fields (Header, URL, Body) that are unsafe to modify directly.
	// Also, RoundTripper interface recommends not to modify the request:
	//   https://cs.opensource.google/go/go/+/refs/tags/go1.24.4:src/net/http/client.go;l=128-132
	// Extra materials: https://pkg.go.dev/net/http#Request.Clone (deep copy requirement)
	//   and https://github.com/golang/go/issues/36095 (concurrency safety discussion)
	cloned := req.Clone(req.Context())
	cloned.SetBasicAuth(t.Username, t.Password)
	return rt.RoundTrip(cloned)
}

// PrometheusBearerTokenAuthTransport injects bearer token into HTTP requests.
type PrometheusBearerTokenAuthTransport struct {
	Token string
	Base  http.RoundTripper
}

// RoundTrip function injects the bearer token in the request's Authorization header
func (bt *PrometheusBearerTokenAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := bt.Base
	if rt == nil {
		rt = http.DefaultTransport
	}

	cloned := req.Clone(req.Context())
	cloned.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bt.Token))
	return rt.RoundTrip(cloned)
}

// PrometheusHistoryProviderConfig allow to select which metrics
// should be queried to get real resource utilization.
type PrometheusHistoryProviderConfig struct {
	Address                                          string
	Insecure                                         bool
	QueryTimeout                                     time.Duration
	HistoryLength, HistoryResolution                 string
	PodLabelPrefix, PodLabelsMetricName              string
	PodNamespaceLabel, PodNameLabel                  string
	CtrNamespaceLabel, CtrPodNameLabel, CtrNameLabel string
	ClusterLabel, ClusterID                          string
	CadvisorMetricsJobName                           string
	Namespace                                        string
	CPUMetricName, MemoryMetricName                  string

	Authentication PrometheusCredentials
}

// PrometheusCredentials keeps credentials for Prometheus API. The Username + Password pair is mutually exclusive with
// the BearerToken field. It's handled in the CLI flags. But if BearerToken is set, it will have priority over the basic auth.
// If both are empty, no authentication is used.
type PrometheusCredentials struct {
	Username    string
	Password    string
	BearerToken string
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
// Consider refactoring to passing clusterState and create history provider working with checkpoints.
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

// NewPrometheusHistoryProvider constructs a history provider that gets data from Prometheus.
func NewPrometheusHistoryProvider(config PrometheusHistoryProviderConfig) (HistoryProvider, error) {
	prometheusTransport := promapi.DefaultRoundTripper

	if config.Insecure {
		prometheusTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if config.Authentication.BearerToken != "" {
		prometheusTransport = &PrometheusBearerTokenAuthTransport{
			Token: config.Authentication.BearerToken,
			Base:  prometheusTransport,
		}
	} else if config.Authentication.Username != "" && config.Authentication.Password != "" {
		prometheusTransport = &PrometheusBasicAuthTransport{
			Username: config.Authentication.Username,
			Password: config.Authentication.Password,
			Base:     prometheusTransport,
		}
	} else {
		// check if env vars for credentials are set
		prometheusUsername := os.Getenv("PROMETHEUS_USERNAME")
		prometheusPassword := os.Getenv("PROMETHEUS_PASSWORD")
		if prometheusUsername != "" && prometheusPassword != "" {
			prometheusTransport = &PrometheusBasicAuthTransport{
				Username: prometheusUsername,
				Password: prometheusPassword,
				Base:     prometheusTransport,
			}
		}
	}

	roundTripper := metrics_recommender.NewPrometheusRoundTripperCounter(
		metrics_recommender.NewPrometheusRoundTripperDuration(prometheusTransport),
	)

	promConfig := promapi.Config{
		Address:      config.Address,
		RoundTripper: roundTripper,
	}

	promClient, err := promapi.NewClient(promConfig)
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
			PodName:   podName,
		},
		ContainerName: containerName,
	}, nil
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
			Resource:     resource,
		})
	}
	return res
}

func (p *prometheusHistoryProvider) readResourceHistory(res map[model.PodID]*PodHistory, query string, resource model.ResourceName) error {
	end := time.Now()
	start := end.Add(-time.Duration(p.historyDuration))

	ctx, cancel := context.WithTimeout(context.Background(), p.queryTimeout)
	defer cancel()

	result, _, err := p.prometheusClient.QueryRange(ctx, query, prometheusv1.Range{
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

	result, _, err := p.prometheusClient.Query(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("cannot get timeseries for labels: %v", err)
	}
	switch value := result.(type) {
	case prommodel.Vector:
		for _, sample := range value {
			if err := p.updatePodLabelsFromMetric(res, sample.Metric, sample.Timestamp.Time()); err != nil {
				return err
			}
		}
	case prommodel.Matrix:
		for _, ts := range value {
			// time series results will always be sorted chronologically from oldest to
			// newest, so the last element is the latest sample
			lastSample := ts.Values[len(ts.Values)-1]
			if err := p.updatePodLabelsFromMetric(res, ts.Metric, lastSample.Timestamp.Time()); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("expected query to return a matrix or vector; got result type %T", result)
	}
	return nil
}

func (p *prometheusHistoryProvider) updatePodLabelsFromMetric(res map[model.PodID]*PodHistory, metric prommodel.Metric, sampleTime time.Time) error {
	if p.config.ClusterLabel != "" {
		clusterLabelValue, ok := metric[prommodel.LabelName(p.config.ClusterLabel)]
		if !ok || string(clusterLabelValue) != p.config.ClusterID {
			return nil
		}
	}
	podID, err := p.getPodIDFromLabels(metric)
	if err != nil {
		return fmt.Errorf("cannot get container ID from labels %v: %v", metric, err)
	}
	podHistory, ok := res[*podID]
	if !ok {
		podHistory = newEmptyHistory()
		res[*podID] = podHistory
	}
	podLabels := p.getPodLabelsMap(metric)
	if sampleTime.After(podHistory.LastSeen) {
		podHistory.LastSeen = sampleTime
		podHistory.LastLabels = podLabels
	}
	return nil
}

type clusterMatcherStatus int

const (
	clusterMatcherMissing clusterMatcherStatus = iota
	clusterMatcherCompatible
	clusterMatcherIncompatible
)

var metricNamePattern = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

func matchRegexOperator(expr, clusterID string) (bool, error) {
	compiled, err := regexp.Compile("^(?:" + expr + ")$")
	if err != nil {
		return false, fmt.Errorf("invalid regex matcher %q: %v", expr, err)
	}
	return compiled.MatchString(clusterID), nil
}

func isRegexExactlyClusterID(expr, clusterID string) bool {
	trimmed := strings.TrimSpace(expr)
	escapedClusterID := regexp.QuoteMeta(clusterID)
	return trimmed == escapedClusterID || trimmed == "^"+escapedClusterID+"$"
}

func isStrictClusterRegex(expr, clusterID string) (bool, error) {
	matchesClusterID, err := matchRegexOperator(expr, clusterID)
	if err != nil {
		return false, err
	}
	if !matchesClusterID {
		return false, nil
	}
	return isRegexExactlyClusterID(expr, clusterID), nil
}

func getClusterMatcherStatus(selector, clusterLabel, clusterID string) (clusterMatcherStatus, error) {
	escapedLabel := regexp.QuoteMeta(clusterLabel)
	matcherPattern := fmt.Sprintf(`(^|,)\s*%s\s*(=~|!~|=|!=)\s*"(?:[^"\\]|\\.)*"\s*(,|$)`, escapedLabel)
	matcherRegex := regexp.MustCompile(matcherPattern)
	matches := matcherRegex.FindAllStringSubmatch(selector, -1)
	if len(matches) == 0 {
		return clusterMatcherMissing, nil
	}
	if len(matches) > 1 {
		return clusterMatcherIncompatible, fmt.Errorf("pod label query selector %q contains multiple matchers for label %q", selector, clusterLabel)
	}

	fullMatcher := strings.TrimSpace(matches[0][0])
	operator := matches[0][2]
	valueRegex := regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
	quotedValue := valueRegex.FindString(fullMatcher)
	if quotedValue == "" {
		return clusterMatcherIncompatible, fmt.Errorf("pod label query selector %q has invalid matcher for label %q", selector, clusterLabel)
	}
	rawValue, err := strconv.Unquote(quotedValue)
	if err != nil {
		return clusterMatcherIncompatible, fmt.Errorf("cannot parse matcher value %q for label %q: %v", quotedValue, clusterLabel, err)
	}

	switch operator {
	case "=":
		if rawValue == clusterID {
			return clusterMatcherCompatible, nil
		}
		return clusterMatcherIncompatible, nil
	case "!=":
		return clusterMatcherIncompatible, nil
	case "=~":
		isStrictMatcher, err := isStrictClusterRegex(rawValue, clusterID)
		if err != nil {
			return clusterMatcherIncompatible, err
		}
		if isStrictMatcher {
			return clusterMatcherCompatible, nil
		}
		return clusterMatcherIncompatible, nil
	case "!~":
		return clusterMatcherIncompatible, nil
	default:
		return clusterMatcherIncompatible, fmt.Errorf("unsupported operator %q for label %q", operator, clusterLabel)
	}
}

func (p *prometheusHistoryProvider) getClusterScopedPodLabelsQuery() (string, error) {
	query := strings.TrimSpace(p.config.PodLabelsMetricName)
	if p.config.ClusterLabel == "" {
		return query, nil
	}

	clusterMatcher := fmt.Sprintf(`%s="%s"`, p.config.ClusterLabel, p.config.ClusterID)
	open := strings.Index(query, "{")
	close := strings.LastIndex(query, "}")
	if open == -1 && close == -1 {
		// Plain metric names (for example "up") are valid selector queries.
		// Convert them into an explicitly cluster-scoped selector.
		if metricNamePattern.MatchString(query) {
			return fmt.Sprintf(`%s{%s}`, query, clusterMatcher), nil
		}
		return "", fmt.Errorf("cannot inject cluster matcher into pod label query %q", query)
	}
	if open == -1 || close == -1 || close < open {
		return "", fmt.Errorf("cannot inject cluster matcher into pod label query %q", query)
	}

	selector := strings.TrimSpace(query[open+1 : close])
	matcherStatus, err := getClusterMatcherStatus(selector, p.config.ClusterLabel, p.config.ClusterID)
	if err != nil {
		return "", err
	}
	if matcherStatus == clusterMatcherCompatible {
		return query, nil
	}

	if selector == "" {
		return query[:open+1] + clusterMatcher + query[close:], nil
	}
	return query[:close] + ", " + clusterMatcher + query[close:], nil
}

func (p *prometheusHistoryProvider) GetClusterHistory() (map[model.PodID]*PodHistory, error) {
	res := make(map[model.PodID]*PodHistory)

	selectorParts := []string{}
	if p.config.CadvisorMetricsJobName != "" {
		selectorParts = append(selectorParts, fmt.Sprintf("job=\"%s\"", p.config.CadvisorMetricsJobName))
	}
	if p.config.ClusterLabel != "" {
		selectorParts = append(selectorParts, fmt.Sprintf("%s=\"%s\"", p.config.ClusterLabel, p.config.ClusterID))
	}
	selectorParts = append(selectorParts, fmt.Sprintf("%s=~\".+\"", p.config.CtrPodNameLabel))
	selectorParts = append(selectorParts, fmt.Sprintf("%s!=\"POD\"", p.config.CtrNameLabel))
	selectorParts = append(selectorParts, fmt.Sprintf("%s!=\"\"", p.config.CtrNameLabel))

	if p.config.Namespace != "" {
		selectorParts = append(selectorParts, fmt.Sprintf("%s=\"%s\"", p.config.CtrNamespaceLabel, p.config.Namespace))
	}
	podSelector := strings.Join(selectorParts, ", ")
	historicalCpuQuery := fmt.Sprintf("rate(%s{%s}[%s])", p.config.CPUMetricName, podSelector, p.config.HistoryResolution)
	klog.V(4).InfoS("Historical CPU usage query", "query", historicalCpuQuery)
	err := p.readResourceHistory(res, historicalCpuQuery, model.ResourceCPU)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}

	historicalMemoryQuery := fmt.Sprintf("%s{%s}", p.config.MemoryMetricName, podSelector)
	klog.V(4).InfoS("Historical memory usage query", "query", historicalMemoryQuery)
	err = p.readResourceHistory(res, historicalMemoryQuery, model.ResourceMemory)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}
	for _, podHistory := range res {
		for _, samples := range podHistory.Samples {
			sort.Slice(samples, func(i, j int) bool { return samples[i].MeasureStart.Before(samples[j].MeasureStart) })
		}
	}
	labelsQuery, err := p.getClusterScopedPodLabelsQuery()
	if err != nil {
		return nil, fmt.Errorf("cannot scope pod labels query by cluster: %v", err)
	}
	err = p.readLastLabels(res, labelsQuery)
	if err != nil {
		return nil, fmt.Errorf("cannot read last labels: %v", err)
	}

	return res, nil
}
