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
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus" // for client-go metrics registration

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

// NodeScaleDownReason describes reason for removing node
type NodeScaleDownReason string

// FailedScaleUpReason describes reason of failed scale-up
type FailedScaleUpReason string

// FunctionLabel is a name of Cluster Autoscaler operation for which
// we measure duration
type FunctionLabel string

// NodeGroupType describes node group relation to CA
type NodeGroupType string

const (
	caNamespace           = "cluster_autoscaler"
	readyLabel            = "ready"
	unreadyLabel          = "unready"
	startingLabel         = "notStarted"
	unregisteredLabel     = "unregistered"
	longUnregisteredLabel = "longUnregistered"

	// Underutilized node was removed because of low utilization
	Underutilized NodeScaleDownReason = "underutilized"
	// Empty node was removed
	Empty NodeScaleDownReason = "empty"
	// Unready node was removed
	Unready NodeScaleDownReason = "unready"

	// APIError caused scale-up to fail
	APIError FailedScaleUpReason = "apiCallError"
	// Timeout was encountered when trying to scale-up
	Timeout FailedScaleUpReason = "timeout"
	// Stockout was encountered when trying to scale-up
	Stockout FailedScaleUpReason = "stockout"
	// QuotaExceeded was encoutered when trying to scale-up
	QuotaExceeded FailedScaleUpReason = "quotaExceeded"

	// autoscaledGroup is managed by CA
	autoscaledGroup NodeGroupType = "autoscaled"
	// autoprovisionedGroup have been created by CA (Node Autoprovisioning),
	// is currently autoscaled and can be removed by CA if it's no longer needed
	autoprovisionedGroup NodeGroupType = "autoprovisioned"

	// LogLongDurationThreshold defines the duration after which long function
	// duration will be logged (in addition to being counted in metric).
	// This is meant to help find unexpectedly long function execution times for
	// debugging purposes.
	LogLongDurationThreshold = 5 * time.Second
)

// Names of Cluster Autoscaler operations
const (
	ScaleDown                  FunctionLabel = "scaleDown"
	ScaleDownNodeDeletion      FunctionLabel = "scaleDown:nodeDeletion"
	ScaleDownFindNodesToRemove FunctionLabel = "scaleDown:findNodesToRemove"
	ScaleDownMiscOperations    FunctionLabel = "scaleDown:miscOperations"
	ScaleUp                    FunctionLabel = "scaleUp"
	FindUnneeded               FunctionLabel = "findUnneeded"
	UpdateState                FunctionLabel = "updateClusterState"
	FilterOutSchedulable       FunctionLabel = "filterOutSchedulable"
	Main                       FunctionLabel = "main"
	Poll                       FunctionLabel = "poll"
	Reconfigure                FunctionLabel = "reconfigure"
	Autoscaling                FunctionLabel = "autoscaling"
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

	nodeGroupsCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "node_groups_count",
			Help:      "Number of node groups managed by CA.",
		}, []string{"node_group_type"},
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

	gpuScaleUpCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "scaled_up_gpu_nodes_total",
			Help:      "Number of GPU nodes added by CA, by GPU name.",
		}, []string{"gpu_name"},
	)

	failedScaleUpCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "failed_scale_ups_total",
			Help:      "Number of times scale-up operation has failed.",
		}, []string{"reason"},
	)

	scaleDownCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "scaled_down_nodes_total",
			Help:      "Number of nodes removed by CA.",
		}, []string{"reason"},
	)

	gpuScaleDownCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "scaled_down_gpu_nodes_total",
			Help:      "Number of GPU nodes removed by CA, by reason and GPU name.",
		}, []string{"reason", "gpu_name"},
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

	/**** Metrics related to NodeAutoprovisioning ****/
	napEnabled = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "nap_enabled",
			Help:      "Whether or not Node Autoprovisioning is enabled. 1 if it is, 0 otherwise.",
		},
	)

	nodeGroupCreationCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "created_node_groups_total",
			Help:      "Number of node groups created by Node Autoprovisioning.",
		},
	)

	nodeGroupDeletionCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "deleted_node_groups_total",
			Help:      "Number of node groups deleted by Node Autoprovisioning.",
		},
	)
)

// RegisterAll registers all metrics.
func RegisterAll() {
	prometheus.MustRegister(clusterSafeToAutoscale)
	prometheus.MustRegister(nodesCount)
	prometheus.MustRegister(nodeGroupsCount)
	prometheus.MustRegister(unschedulablePodsCount)
	prometheus.MustRegister(lastActivity)
	prometheus.MustRegister(functionDuration)
	prometheus.MustRegister(errorsCount)
	prometheus.MustRegister(scaleUpCount)
	prometheus.MustRegister(gpuScaleUpCount)
	prometheus.MustRegister(failedScaleUpCount)
	prometheus.MustRegister(scaleDownCount)
	prometheus.MustRegister(gpuScaleDownCount)
	prometheus.MustRegister(evictionsCount)
	prometheus.MustRegister(unneededNodesCount)
	prometheus.MustRegister(napEnabled)
	prometheus.MustRegister(nodeGroupCreationCount)
	prometheus.MustRegister(nodeGroupDeletionCount)
}

// UpdateDurationFromStart records the duration of the step identified by the
// label using start time
func UpdateDurationFromStart(label FunctionLabel, start time.Time) {
	duration := time.Now().Sub(start)
	UpdateDuration(label, duration)
}

// UpdateDuration records the duration of the step identified by the label
func UpdateDuration(label FunctionLabel, duration time.Duration) {
	// TODO(maciekpytel): remove second condition if we manage to get
	// asynchronous node drain
	if duration > LogLongDurationThreshold && label != ScaleDown {
		glog.V(4).Infof("Function %s took %v to complete", label, duration)
	}
	functionDuration.WithLabelValues(string(label)).Observe(duration.Seconds())
}

// UpdateLastTime records the time the step identified by the label was started
func UpdateLastTime(label FunctionLabel, now time.Time) {
	lastActivity.WithLabelValues(string(label)).Set(float64(now.Unix()))
}

// UpdateClusterSafeToAutoscale records if cluster is safe to autoscale
func UpdateClusterSafeToAutoscale(safe bool) {
	if safe {
		clusterSafeToAutoscale.Set(1)
	} else {
		clusterSafeToAutoscale.Set(0)
	}
}

// UpdateNodesCount records the number of nodes in cluster
func UpdateNodesCount(ready, unready, starting, longUnregistered, unregistered int) {
	nodesCount.WithLabelValues(readyLabel).Set(float64(ready))
	nodesCount.WithLabelValues(unreadyLabel).Set(float64(unready))
	nodesCount.WithLabelValues(startingLabel).Set(float64(starting))
	nodesCount.WithLabelValues(longUnregisteredLabel).Set(float64(longUnregistered))
	nodesCount.WithLabelValues(unregisteredLabel).Set(float64(unregistered))
}

// UpdateNodeGroupsCount records the number of node groups managed by CA
func UpdateNodeGroupsCount(autoscaled, autoprovisioned int) {
	nodeGroupsCount.WithLabelValues(string(autoscaledGroup)).Set(float64(autoscaled))
	nodeGroupsCount.WithLabelValues(string(autoprovisionedGroup)).Set(float64(autoprovisioned))
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
func RegisterScaleUp(nodesCount int, gpuType string) {
	scaleUpCount.Add(float64(nodesCount))
	if gpuType != gpu.MetricsNoGPU {
		gpuScaleUpCount.WithLabelValues(gpuType).Add(float64(nodesCount))
	}
}

// RegisterFailedScaleUp records a failed scale-up operation
func RegisterFailedScaleUp(reason FailedScaleUpReason) {
	failedScaleUpCount.WithLabelValues(string(reason)).Inc()
}

// RegisterScaleDown records number of nodes removed by scale down
func RegisterScaleDown(nodesCount int, gpuType string, reason NodeScaleDownReason) {
	scaleDownCount.WithLabelValues(string(reason)).Add(float64(nodesCount))
	if gpuType != gpu.MetricsNoGPU {
		gpuScaleDownCount.WithLabelValues(string(reason), gpuType).Add(float64(nodesCount))
	}
}

// RegisterEvictions records number of evicted pods
func RegisterEvictions(podsCount int) {
	evictionsCount.Add(float64(podsCount))
}

// UpdateUnneededNodesCount records number of currently unneeded nodes
func UpdateUnneededNodesCount(nodesCount int) {
	unneededNodesCount.Set(float64(nodesCount))
}

// UpdateNapEnabled records if NodeAutoprovisioning is enabled
func UpdateNapEnabled(enabled bool) {
	if enabled {
		napEnabled.Set(1)
	} else {
		napEnabled.Set(0)
	}
}

// RegisterNodeGroupCreation registers node group creation
func RegisterNodeGroupCreation() {
	nodeGroupCreationCount.Add(1.0)
}

// RegisterNodeGroupDeletion registers node group deletion
func RegisterNodeGroupDeletion() {
	nodeGroupDeletionCount.Add(1.0)
}
