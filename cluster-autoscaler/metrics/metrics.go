/*
Copyright 2016 The Kubernetes Authors.

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

package metrics

import (
	"reflect"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	caNamespace   = "cluster_autoscaler"
	readyLabel    = "ready"
	unreadyLabel  = "unready"
	startingLabel = "notStarted"
)

var (
	/**** Metrics related to cluster state ****/
	clusterSafeToAutoscale = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "cluster_safe_to_autoscale",
			Help:      "Whether or not cluster is healthy enough for autoscaling. 1 if it is, 0 otherwise.",
		},
	)

	nodesCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "nodes_count",
			Help:      "Number of nodes in cluster.",
		}, []string{"state"},
	)

	unschedulablePodsCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "unschedulable_pods_count",
			Help:      "Number of unschedulable pods in the cluster.",
		},
	)

	lastTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "last_time_seconds",
			Help:      "Last time CA run some main loop fragment.",
		}, []string{"main"},
	)

	lastDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "last_duration_microseconds",
			Help:      "Time spent in last main loop fragments in microseconds.",
		}, []string{"main"},
	)

	duration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: caNamespace,
			Name:      "duration_microseconds",
			Help:      "Time spent in main loop fragments in microseconds.",
		}, []string{"main"},
	)
)

func init() {
	prometheus.MustRegister(duration)
	prometheus.MustRegister(lastDuration)
	prometheus.MustRegister(lastTimestamp)
	prometheus.MustRegister(clusterSafeToAutoscale)
	prometheus.MustRegister(nodesCount)
	prometheus.MustRegister(unschedulablePodsCount)
}

func durationToMicro(start time.Time) float64 {
	return float64(time.Now().Sub(start).Nanoseconds() / 1000)
}

// UpdateDuration records the duration of the step identified by the label
func UpdateDuration(label string, start time.Time) {
	duration.WithLabelValues(label).Observe(durationToMicro(start))
	lastDuration.WithLabelValues(label).Set(durationToMicro(start))
}

// UpdateLastTime records the time the step identified by the label was started
func UpdateLastTime(label string) {
	lastTimestamp.WithLabelValues(label).Set(float64(time.Now().Unix()))
}

// UpdateClusterState updates metrics related to cluster state
func UpdateClusterState(csr *clusterstate.ClusterStateRegistry) {
	if csr == nil || reflect.ValueOf(csr).IsNil() {
		return
	}
	if csr.IsClusterHealthy() {
		clusterSafeToAutoscale.Set(1)
	} else {
		clusterSafeToAutoscale.Set(0)
	}
	readiness := csr.GetClusterReadiness()
	nodesCount.WithLabelValues(readyLabel).Set(float64(readiness.Ready))
	nodesCount.WithLabelValues(unreadyLabel).Set(float64(readiness.Unready + readiness.LongNotStarted))
	nodesCount.WithLabelValues(startingLabel).Set(float64(readiness.NotStarted))
}

// UpdateUnschedulablePodsCount records number of currently unschedulable pods
func UpdateUnschedulablePodsCount(podsCount int) {
	unschedulablePodsCount.Set(float64(podsCount))
}
