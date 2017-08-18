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
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

// NodeScaleDownReason describes reason for removing node
type NodeScaleDownReason string

const (
	caNamespace   = "cluster_autoscaler"
	readyLabel    = "ready"
	unreadyLabel  = "unready"
	startingLabel = "notStarted"

	// Underutilized node was removed because of low utilization
	Underutilized NodeScaleDownReason = "underutilized"
	// Empty node was removed
	Empty NodeScaleDownReason = "empty"
	// Unready node was removed
	Unready NodeScaleDownReason = "unready"

	// LogLongDurationThreshold defines the duration after which long function
	// duration will be logged (in addition to being counted in metric).
	// This is meant to help find unexpectedly long function execution times for
	// debugging purposes.
	LogLongDurationThreshold = 5 * time.Second
)

const (
	ScaleDownNodeDeletion      = "scaleDown:nodeDeletion"
	ScaleDownFindNodesToRemove = "scaleDown:findNodesToRemove"
	ScaleDownMiscOperations    = "scaleDown:miscOperations"
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

	/**** Metrics related to autoscaler execution ****/
	lastActivity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "last_activity",
			Help:      "Last time certain part of CA logic executed.",
		}, []string{"activity"},
	)

	functionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: caNamespace,
			Name:      "function_duration_seconds",
			Help:      "Time taken by various parts of CA main loop.",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 7.5, 10.0, 12.5, 15.0, 17.5, 20.0, 22.5, 25.0, 27.5, 30.0, 50.0, 75.0, 100.0, 1000.0},
		}, []string{"function"},
	)

	/**** Metrics related to autoscaler operations ****/
	errorsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "errors_total",
			Help:      "The number of CA loops failed due to an error.",
		}, []string{"type"},
	)

	scaleUpCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "scaled_up_nodes_total",
			Help:      "Number of nodes added by CA.",
		},
	)

	scaleDownCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "scaled_down_nodes_total",
			Help:      "Number of nodes removed by CA.",
		}, []string{"reason"},
	)

	evictionsCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "evicted_pods_total",
			Help:      "Number of pods evicted by CA",
		},
	)

	unneededNodesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "unneeded_nodes_count",
			Help:      "Number of nodes currently considered unneeded by CA.",
		},
	)
)

func init() {
	prometheus.MustRegister(clusterSafeToAutoscale)
	prometheus.MustRegister(nodesCount)
	prometheus.MustRegister(unschedulablePodsCount)
	prometheus.MustRegister(lastActivity)
	prometheus.MustRegister(functionDuration)
	prometheus.MustRegister(errorsCount)
	prometheus.MustRegister(scaleUpCount)
	prometheus.MustRegister(scaleDownCount)
	prometheus.MustRegister(evictionsCount)
	prometheus.MustRegister(unneededNodesCount)
}

// UpdateDurationFromStart records the duration of the step identified by the
// label using start time
func UpdateDurationFromStart(label string, start time.Time) {
	duration := time.Now().Sub(start)
	UpdateDuration(label, duration)
}

// UpdateDuration records the duration of the step identified by the label
func UpdateDuration(label string, duration time.Duration) {
	// TODO(maciekpytel): remove second condition if we manage to get
	// asynchronous node drain
	if duration > LogLongDurationThreshold && label != "scaleDown" {
		glog.Infof("Function %s took %v to complete", label, duration)
	}
	functionDuration.WithLabelValues(label).Observe(duration.Seconds())
}

// UpdateLastTime records the time the step identified by the label was started
func UpdateLastTime(label string, now time.Time) {
	lastActivity.WithLabelValues(label).Set(float64(now.Unix()))
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

// RegisterError records any errors preventing Cluster Autoscaler from working.
// No more than one error should be recorded per loop.
func RegisterError(err errors.AutoscalerError) {
	errorsCount.WithLabelValues(string(err.Type())).Add(1.0)
}

// RegisterScaleUp records number of nodes added by scale up
func RegisterScaleUp(nodesCount int) {
	scaleUpCount.Add(float64(nodesCount))
}

// RegisterScaleDown records number of nodes removed by scale down
func RegisterScaleDown(nodesCount int, reason NodeScaleDownReason) {
	scaleDownCount.WithLabelValues(string(reason)).Add(float64(nodesCount))
}

// RegisterEvictions records number of evicted pods
func RegisterEvictions(podsCount int) {
	evictionsCount.Add(float64(podsCount))
}

// UpdateUnneededNodesCount records number of currently unneeded nodes
func UpdateUnneededNodesCount(nodesCount int) {
	unneededNodesCount.Set(float64(nodesCount))
}
