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
	"fmt"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/simulator"

	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	_ "k8s.io/component-base/metrics/prometheus/restclient" // for client-go metrics registration

	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
	klog "k8s.io/klog/v2"
)

// NodeScaleDownReason describes reason for removing node
type NodeScaleDownReason string

// FailedScaleReason describes reason of failed scale-up
type FailedScaleReason string

// FunctionLabel is a name of Cluster Autoscaler operation for which
// we measure duration
type FunctionLabel string

// NodeGroupType describes node group relation to CA
type NodeGroupType string

// PodEvictionResult describes result of the pod eviction attempt
type PodEvictionResult string

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

	// CloudProviderError caused scale-up to fail
	CloudProviderError FailedScaleReason = "cloudProviderError"
	// APIError caused scale-up to fail
	APIError FailedScaleReason = "apiCallError"
	// Timeout was encountered when trying to scale-up
	Timeout FailedScaleReason = "timeout"

	// DirectionScaleDown is the direction of skipped scaling event when scaling in (shrinking)
	DirectionScaleDown string = "down"
	// DirectionScaleUp is the direction of skipped scaling event when scaling out (growing)
	DirectionScaleUp string = "up"

	// CpuResourceLimit minimum or maximum reached, check the direction label to determine min or max
	CpuResourceLimit string = "CpuResourceLimit"
	// MemoryResourceLimit minimum or maximum reached, check the direction label to determine min or max
	MemoryResourceLimit string = "MemoryResourceLimit"

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
	// PodEvictionSucceed means creation of the pod eviction object succeed
	PodEvictionSucceed PodEvictionResult = "succeeded"
	// PodEvictionFailed means creation of the pod eviction object failed
	PodEvictionFailed PodEvictionResult = "failed"
)

// Names of Cluster Autoscaler operations
const (
	ScaleDown                  FunctionLabel = "scaleDown"
	ScaleDownNodeDeletion      FunctionLabel = "scaleDown:nodeDeletion"
	ScaleDownFindNodesToRemove FunctionLabel = "scaleDown:findNodesToRemove"
	ScaleDownMiscOperations    FunctionLabel = "scaleDown:miscOperations"
	ScaleDownSoftTaintUnneeded FunctionLabel = "scaleDown:softTaintUnneeded"
	ScaleUp                    FunctionLabel = "scaleUp"
	BuildPodEquivalenceGroups  FunctionLabel = "scaleUp:buildPodEquivalenceGroups"
	Estimate                   FunctionLabel = "scaleUp:estimate"
	FindUnneeded               FunctionLabel = "findUnneeded"
	UpdateState                FunctionLabel = "updateClusterState"
	FilterOutSchedulable       FunctionLabel = "filterOutSchedulable"
	CloudProviderRefresh       FunctionLabel = "cloudProviderRefresh"
	Main                       FunctionLabel = "main"
	Poll                       FunctionLabel = "poll"
	Reconfigure                FunctionLabel = "reconfigure"
	Autoscaling                FunctionLabel = "autoscaling"
	LoopWait                   FunctionLabel = "loopWait"
	BulkListAllGceInstances    FunctionLabel = "bulkListInstances:listAllInstances"
	BulkListMigInstances       FunctionLabel = "bulkListInstances:listMigInstances"
)

var (
	/**** Metrics related to cluster state ****/
	clusterSafeToAutoscale = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "cluster_safe_to_autoscale",
			Help:      "Whether or not cluster is healthy enough for autoscaling. 1 if it is, 0 otherwise.",
		},
	)

	nodesCount = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "nodes_count",
			Help:      "Number of nodes in cluster.",
		}, []string{"state"},
	)

	nodeGroupsCount = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "node_groups_count",
			Help:      "Number of node groups managed by CA.",
		}, []string{"node_group_type"},
	)

	// Unschedulable pod count can be from scheduler-marked-unschedulable pods or not-yet-processed pods (unknown)
	unschedulablePodsCount = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "unschedulable_pods_count",
			Help:      "Number of unschedulable pods in the cluster.",
		}, []string{"type"},
	)

	maxNodesCount = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "max_nodes_count",
			Help:      "Maximum number of nodes in all node groups",
		},
	)

	cpuCurrentCores = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "cluster_cpu_current_cores",
			Help:      "Current number of cores in the cluster, minus deleting nodes.",
		},
	)

	cpuLimitsCores = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "cpu_limits_cores",
			Help:      "Minimum and maximum number of cores in the cluster.",
		}, []string{"direction"},
	)

	memoryCurrentBytes = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "cluster_memory_current_bytes",
			Help:      "Current number of bytes of memory in the cluster, minus deleting nodes.",
		},
	)

	memoryLimitsBytes = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "memory_limits_bytes",
			Help:      "Minimum and maximum number of bytes of memory in cluster.",
		}, []string{"direction"},
	)

	nodesGroupMinNodes = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "node_group_min_count",
			Help:      "Minimum number of nodes in the node group",
		}, []string{"node_group"},
	)

	nodesGroupMaxNodes = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "node_group_max_count",
			Help:      "Maximum number of nodes in the node group",
		}, []string{"node_group"},
	)

	nodesGroupTargetSize = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "node_group_target_count",
			Help:      "Target number of nodes in the node group by CA.",
		}, []string{"node_group"},
	)

	nodesGroupHealthiness = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "node_group_healthiness",
			Help:      "Whether or not node group is healthy enough for autoscaling. 1 if it is, 0 otherwise.",
		}, []string{"node_group"},
	)

	nodeGroupBackOffStatus = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "node_group_backoff_status",
			Help:      "Whether or not node group is backoff for not autoscaling. 1 if it is, 0 otherwise.",
		}, []string{"node_group", "reason"},
	)

	/**** Metrics related to autoscaler execution ****/
	lastActivity = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "last_activity",
			Help:      "Last time certain part of CA logic executed.",
		}, []string{"activity"},
	)

	functionDuration = k8smetrics.NewHistogramVec(
		&k8smetrics.HistogramOpts{
			Namespace: caNamespace,
			Name:      "function_duration_seconds",
			Help:      "Time taken by various parts of CA main loop.",
			Buckets:   k8smetrics.ExponentialBuckets(0.01, 1.5, 30), // 0.01, 0.015, 0.0225, ..., 852.2269299239293, 1278.3403948858938
		}, []string{"function"},
	)

	functionDurationSummary = k8smetrics.NewSummaryVec(
		&k8smetrics.SummaryOpts{
			Namespace: caNamespace,
			Name:      "function_duration_quantile_seconds",
			Help:      "Quantiles of time taken by various parts of CA main loop.",
			MaxAge:    time.Hour,
		}, []string{"function"},
	)

	pendingNodeDeletions = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "pending_node_deletions",
			Help:      "Number of nodes that haven't been removed or aborted after finished scale-down phase.",
		},
	)

	/**** Metrics related to autoscaler operations ****/
	errorsCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "errors_total",
			Help:      "The number of CA loops failed due to an error.",
		}, []string{"type"},
	)

	scaleUpCount = k8smetrics.NewCounter(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "scaled_up_nodes_total",
			Help:      "Number of nodes added by CA.",
		},
	)

	gpuScaleUpCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "scaled_up_gpu_nodes_total",
			Help:      "Number of GPU nodes added by CA, by GPU name.",
		}, []string{"gpu_resource_name", "gpu_name"},
	)

	failedScaleUpCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "failed_scale_ups_total",
			Help:      "Number of times scale-up operation has failed.",
		}, []string{"reason"},
	)

	failedGPUScaleUpCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "failed_gpu_scale_ups_total",
			Help:      "Number of times scale-up operation has failed.",
		}, []string{"reason", "gpu_resource_name", "gpu_name"},
	)

	failedScaleDownCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "failed_scale_down_total",
			Help:      "Number of times scale-down operation has failed.",
		}, []string{"reason"},
	)

	failedGPUScaleDownCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "failed_gpu_scale_down_total",
			Help:      "Number of times scale-down operation has failed.",
		}, []string{"reason", "gpu_resource_name", "gpu_name"},
	)

	scaleDownCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "scaled_down_nodes_total",
			Help:      "Number of nodes removed by CA.",
		}, []string{"reason"},
	)

	gpuScaleDownCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "scaled_down_gpu_nodes_total",
			Help:      "Number of GPU nodes removed by CA, by reason and GPU name.",
		}, []string{"reason", "gpu_resource_name", "gpu_name"},
	)

	evictionsCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "evicted_pods_total",
			Help:      "Number of pods evicted by CA",
		}, []string{"eviction_result"},
	)

	unneededNodesCount = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "unneeded_nodes_count",
			Help:      "Number of nodes currently considered unneeded by CA.",
		},
	)

	unremovableNodesCount = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "unremovable_nodes_count",
			Help:      "Number of nodes currently considered unremovable by CA.",
		},
		[]string{"reason"},
	)

	scaleDownInCooldown = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "scale_down_in_cooldown",
			Help:      "Whether or not the scale down is in cooldown. 1 if its, 0 otherwise.",
		},
	)

	oldUnregisteredNodesRemovedCount = k8smetrics.NewCounter(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "old_unregistered_nodes_removed_count",
			Help:      "Number of unregistered nodes removed by CA.",
		},
	)

	overflowingControllersCount = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "overflowing_controllers_count",
			Help:      "Number of controllers that own a large set of heterogenous pods, preventing CA from treating these pods as equivalent.",
		},
	)

	skippedScaleEventsCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "skipped_scale_events_count",
			Help:      "Count of scaling events that the CA has chosen to skip.",
		},
		[]string{"direction", "reason"},
	)

	/**** Metrics related to NodeAutoprovisioning ****/
	napEnabled = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "nap_enabled",
			Help:      "Whether or not Node Autoprovisioning is enabled. 1 if it is, 0 otherwise.",
		},
	)

	nodeGroupCreationCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "created_node_groups_total",
			Help:      "Number of node groups created by Node Autoprovisioning.",
		},
		[]string{"group_type"},
	)

	nodeGroupDeletionCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "deleted_node_groups_total",
			Help:      "Number of node groups deleted by Node Autoprovisioning.",
		},
		[]string{"group_type"},
	)

	nodeTaintsCount = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "node_taints_count",
			Help:      "Number of taints per type used in the cluster.",
		},
		[]string{"type"},
	)

	inconsistentInstancesMigsCount = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "inconsistent_instances_migs_count",
			Help:      "Number of migs where instance count according to InstanceGroupManagers.List() differs from the results of Instances.List(). This can happen when some instances are abandoned or a user edits instance 'created-by' metadata.",
		},
	)
)

// RegisterAll registers all metrics.
func RegisterAll(emitPerNodeGroupMetrics bool) {
	legacyregistry.MustRegister(clusterSafeToAutoscale)
	legacyregistry.MustRegister(nodesCount)
	legacyregistry.MustRegister(nodeGroupsCount)
	legacyregistry.MustRegister(unschedulablePodsCount)
	legacyregistry.MustRegister(maxNodesCount)
	legacyregistry.MustRegister(cpuCurrentCores)
	legacyregistry.MustRegister(cpuLimitsCores)
	legacyregistry.MustRegister(memoryCurrentBytes)
	legacyregistry.MustRegister(memoryLimitsBytes)
	legacyregistry.MustRegister(lastActivity)
	legacyregistry.MustRegister(functionDuration)
	legacyregistry.MustRegister(functionDurationSummary)
	legacyregistry.MustRegister(errorsCount)
	legacyregistry.MustRegister(scaleUpCount)
	legacyregistry.MustRegister(gpuScaleUpCount)
	legacyregistry.MustRegister(failedScaleUpCount)
	legacyregistry.MustRegister(failedGPUScaleUpCount)
	legacyregistry.MustRegister(failedScaleDownCount)
	legacyregistry.MustRegister(failedGPUScaleDownCount)
	legacyregistry.MustRegister(scaleDownCount)
	legacyregistry.MustRegister(gpuScaleDownCount)
	legacyregistry.MustRegister(evictionsCount)
	legacyregistry.MustRegister(unneededNodesCount)
	legacyregistry.MustRegister(unremovableNodesCount)
	legacyregistry.MustRegister(scaleDownInCooldown)
	legacyregistry.MustRegister(oldUnregisteredNodesRemovedCount)
	legacyregistry.MustRegister(overflowingControllersCount)
	legacyregistry.MustRegister(skippedScaleEventsCount)
	legacyregistry.MustRegister(napEnabled)
	legacyregistry.MustRegister(nodeGroupCreationCount)
	legacyregistry.MustRegister(nodeGroupDeletionCount)
	legacyregistry.MustRegister(pendingNodeDeletions)
	legacyregistry.MustRegister(nodeTaintsCount)
	legacyregistry.MustRegister(inconsistentInstancesMigsCount)

	if emitPerNodeGroupMetrics {
		legacyregistry.MustRegister(nodesGroupMinNodes)
		legacyregistry.MustRegister(nodesGroupMaxNodes)
		legacyregistry.MustRegister(nodesGroupTargetSize)
		legacyregistry.MustRegister(nodesGroupHealthiness)
		legacyregistry.MustRegister(nodeGroupBackOffStatus)
	}
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
		klog.V(4).Infof("Function %s took %v to complete", label, duration)
	}
	functionDuration.WithLabelValues(string(label)).Observe(duration.Seconds())
	functionDurationSummary.WithLabelValues(string(label)).Observe(duration.Seconds())
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
func UpdateUnschedulablePodsCount(uschedulablePodsCount, schedulerUnprocessedCount int) {
	UpdateUnschedulablePodsCountWithLabel(uschedulablePodsCount, "unschedulable")
	UpdateUnschedulablePodsCountWithLabel(schedulerUnprocessedCount, "scheduler_unprocessed")
}

// UpdateUnschedulablePodsCountWithLabel records number of currently unschedulable pods wil label "type" value "label"
func UpdateUnschedulablePodsCountWithLabel(uschedulablePodsCount int, label string) {
	unschedulablePodsCount.WithLabelValues(label).Set(float64(uschedulablePodsCount))
}

// UpdateMaxNodesCount records the current maximum number of nodes being set for all node groups
func UpdateMaxNodesCount(nodesCount int) {
	maxNodesCount.Set(float64(nodesCount))
}

// UpdateClusterCPUCurrentCores records the number of cores in the cluster, minus deleting nodes
func UpdateClusterCPUCurrentCores(coresCount int64) {
	cpuCurrentCores.Set(float64(coresCount))
}

// UpdateCPULimitsCores records the minimum and maximum number of cores in the cluster
func UpdateCPULimitsCores(minCoresCount int64, maxCoresCount int64) {
	cpuLimitsCores.WithLabelValues("minimum").Set(float64(minCoresCount))
	cpuLimitsCores.WithLabelValues("maximum").Set(float64(maxCoresCount))
}

// UpdateClusterMemoryCurrentBytes records the number of bytes of memory in the cluster, minus deleting nodes
func UpdateClusterMemoryCurrentBytes(memoryCount int64) {
	memoryCurrentBytes.Set(float64(memoryCount))
}

// UpdateMemoryLimitsBytes records the minimum and maximum bytes of memory in the cluster
func UpdateMemoryLimitsBytes(minMemoryCount int64, maxMemoryCount int64) {
	memoryLimitsBytes.WithLabelValues("minimum").Set(float64(minMemoryCount))
	memoryLimitsBytes.WithLabelValues("maximum").Set(float64(maxMemoryCount))
}

// UpdateNodeGroupMin records the node group minimum allowed number of nodes
func UpdateNodeGroupMin(nodeGroup string, minNodes int) {
	nodesGroupMinNodes.WithLabelValues(nodeGroup).Set(float64(minNodes))
}

// UpdateNodeGroupMax records the node group maximum allowed number of nodes
func UpdateNodeGroupMax(nodeGroup string, maxNodes int) {
	nodesGroupMaxNodes.WithLabelValues(nodeGroup).Set(float64(maxNodes))
}

// UpdateNodeGroupTargetSize records the node group target size
func UpdateNodeGroupTargetSize(targetSizes map[string]int) {
	for nodeGroup, targetSize := range targetSizes {
		nodesGroupTargetSize.WithLabelValues(nodeGroup).Set(float64(targetSize))
	}
}

// UpdateNodeGroupHealthStatus records if node group is healthy to autoscaling
func UpdateNodeGroupHealthStatus(nodeGroup string, healthy bool) {
	if healthy {
		nodesGroupHealthiness.WithLabelValues(nodeGroup).Set(1)
	} else {
		nodesGroupHealthiness.WithLabelValues(nodeGroup).Set(0)
	}
}

// UpdateNodeGroupBackOffStatus records if node group is backoff for not autoscaling
func UpdateNodeGroupBackOffStatus(nodeGroup string, backoffReasonStatus map[string]bool) {
	if len(backoffReasonStatus) == 0 {
		nodeGroupBackOffStatus.WithLabelValues(nodeGroup, "").Set(0)
	} else {
		for reason, backoff := range backoffReasonStatus {
			if backoff {
				nodeGroupBackOffStatus.WithLabelValues(nodeGroup, reason).Set(1)
			} else {
				nodeGroupBackOffStatus.WithLabelValues(nodeGroup, reason).Set(0)
			}
		}
	}
}

// RegisterError records any errors preventing Cluster Autoscaler from working.
// No more than one error should be recorded per loop.
func RegisterError(err errors.AutoscalerError) {
	errorsCount.WithLabelValues(string(err.Type())).Add(1.0)
}

// RegisterScaleUp records number of nodes added by scale up
func RegisterScaleUp(nodesCount int, gpuResourceName, gpuType string) {
	scaleUpCount.Add(float64(nodesCount))
	if gpuType != gpu.MetricsNoGPU {
		gpuScaleUpCount.WithLabelValues(gpuResourceName, gpuType).Add(float64(nodesCount))
	}
}

// RegisterFailedScaleUp records a failed scale-up operation
func RegisterFailedScaleUp(reason FailedScaleReason, gpuResourceName, gpuType string) {
	failedScaleUpCount.WithLabelValues(string(reason)).Inc()
	if gpuType != gpu.MetricsNoGPU {
		failedGPUScaleUpCount.WithLabelValues(string(reason), gpuResourceName, gpuType).Inc()
	}
}

// RegisterScaleDown records number of nodes removed by scale down
func RegisterScaleDown(nodesCount int, gpuResourceName, gpuType string, reason NodeScaleDownReason) {
	scaleDownCount.WithLabelValues(string(reason)).Add(float64(nodesCount))
	if gpuType != gpu.MetricsNoGPU {
		gpuScaleDownCount.WithLabelValues(string(reason), gpuResourceName, gpuType).Add(float64(nodesCount))
	}
}

// RegisterFailedScaleDown records a failed scale-down operation
func RegisterFailedScaleDown(reason FailedScaleReason, gpuResourceName, gpuType string) {
	failedScaleDownCount.WithLabelValues(string(reason)).Inc()
	if gpuType != gpu.MetricsNoGPU {
		failedGPUScaleDownCount.WithLabelValues(string(reason), gpuResourceName, gpuType).Inc()
	}
}

// RegisterEvictions records number of evicted pods succeed or failed
func RegisterEvictions(podsCount int, result PodEvictionResult) {
	evictionsCount.WithLabelValues(string(result)).Add(float64(podsCount))
}

// UpdateUnneededNodesCount records number of currently unneeded nodes
func UpdateUnneededNodesCount(nodesCount int) {
	unneededNodesCount.Set(float64(nodesCount))
}

// UpdateUnremovableNodesCount records number of currently unremovable nodes
func UpdateUnremovableNodesCount(unremovableReasonCounts map[simulator.UnremovableReason]int) {
	for reason, count := range unremovableReasonCounts {
		unremovableNodesCount.WithLabelValues(fmt.Sprintf("%v", reason)).Set(float64(count))
	}
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
	RegisterNodeGroupCreationWithLabelValues("")
}

// RegisterNodeGroupCreationWithLabelValues registers node group creation with the provided labels
func RegisterNodeGroupCreationWithLabelValues(groupType string) {
	nodeGroupCreationCount.WithLabelValues(groupType).Add(1.0)
}

// RegisterNodeGroupDeletion registers node group deletion
func RegisterNodeGroupDeletion() {
	RegisterNodeGroupDeletionWithLabelValues("")
}

// RegisterNodeGroupDeletionWithLabelValues registers node group deletion with the provided labels
func RegisterNodeGroupDeletionWithLabelValues(groupType string) {
	nodeGroupDeletionCount.WithLabelValues(groupType).Add(1.0)
}

// UpdateScaleDownInCooldown registers if the cluster autoscaler
// scaledown is in cooldown
func UpdateScaleDownInCooldown(inCooldown bool) {
	if inCooldown {
		scaleDownInCooldown.Set(1.0)
	} else {
		scaleDownInCooldown.Set(0.0)
	}
}

// RegisterOldUnregisteredNodesRemoved records number of old unregistered
// nodes that have been removed by the cluster autoscaler
func RegisterOldUnregisteredNodesRemoved(nodesCount int) {
	oldUnregisteredNodesRemovedCount.Add(float64(nodesCount))
}

// UpdateOverflowingControllers sets the number of controllers that could not
// have their pods cached.
func UpdateOverflowingControllers(count int) {
	overflowingControllersCount.Set(float64(count))
}

// RegisterSkippedScaleDownCPU increases the count of skipped scale outs because of CPU resource limits
func RegisterSkippedScaleDownCPU() {
	skippedScaleEventsCount.WithLabelValues(DirectionScaleDown, CpuResourceLimit).Add(1.0)
}

// RegisterSkippedScaleDownMemory increases the count of skipped scale outs because of Memory resource limits
func RegisterSkippedScaleDownMemory() {
	skippedScaleEventsCount.WithLabelValues(DirectionScaleDown, MemoryResourceLimit).Add(1.0)
}

// RegisterSkippedScaleUpCPU increases the count of skipped scale outs because of CPU resource limits
func RegisterSkippedScaleUpCPU() {
	skippedScaleEventsCount.WithLabelValues(DirectionScaleUp, CpuResourceLimit).Add(1.0)
}

// RegisterSkippedScaleUpMemory increases the count of skipped scale outs because of Memory resource limits
func RegisterSkippedScaleUpMemory() {
	skippedScaleEventsCount.WithLabelValues(DirectionScaleUp, MemoryResourceLimit).Add(1.0)
}

// ObservePendingNodeDeletions records the current value of nodes_pending_deletion metric
func ObservePendingNodeDeletions(value int) {
	pendingNodeDeletions.Set(float64(value))
}

// ObserveNodeTaintsCount records the node taints count of given type.
func ObserveNodeTaintsCount(taintType string, count float64) {
	nodeTaintsCount.WithLabelValues(taintType).Set(count)
}

// UpdateInconsistentInstancesMigsCount records the observed number of migs where instance count
// according to InstanceGroupManagers.List() differs from the results of Instances.List().
// This can happen when some instances are abandoned or a user edits instance 'created-by' metadata.
func UpdateInconsistentInstancesMigsCount(migCount int) {
	inconsistentInstancesMigsCount.Set(float64(migCount))
}
