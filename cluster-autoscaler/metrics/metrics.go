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
	"strconv"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/restclient" // for client-go metrics registration

	k8smetrics "k8s.io/component-base/metrics"
	klog "k8s.io/klog/v2"
)

// NodeScaleDownReason describes reason for removing node
type NodeScaleDownReason string

// FailedScaleUpReason describes reason of failed scale-up
type FailedScaleUpReason string

// FunctionLabel is a name of Cluster Autoscaler operation for which
// we measure duration
type FunctionLabel string

// FunctionLabelKey is a key of Cluster Autoscaler operation for which
// we measure duration. Integer based key is used to improve performance
// by avoiding string hashing and enforce zero collisions at runtime.
type FunctionLabelKey int

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
	CloudProviderError FailedScaleUpReason = "cloudProviderError"
	// APIError caused scale-up to fail
	APIError FailedScaleUpReason = "apiCallError"
	// Timeout was encountered when trying to scale-up
	Timeout FailedScaleUpReason = "timeout"

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
	ScaleDown                               FunctionLabel = "scaleDown"
	ScaleDownNodeDeletion                   FunctionLabel = "scaleDown:nodeDeletion"
	ScaleDownSoftTaintUnneeded              FunctionLabel = "scaleDown:softTaintUnneeded"
	ScaleUp                                 FunctionLabel = "scaleUp"
	BuildPodEquivalenceGroups               FunctionLabel = "scaleUp:buildPodEquivalenceGroups"
	Estimate                                FunctionLabel = "scaleUp:estimate"
	FindUnneeded                            FunctionLabel = "findUnneeded"
	UpdateState                             FunctionLabel = "updateClusterState"
	FilterOutSchedulable                    FunctionLabel = "filterOutSchedulable"
	CloudProviderRefresh                    FunctionLabel = "cloudProviderRefresh"
	Main                                    FunctionLabel = "main"
	MainSuccessful                          FunctionLabel = "mainSuccessful"
	Poll                                    FunctionLabel = "poll"
	Reconfigure                             FunctionLabel = "reconfigure"
	Autoscaling                             FunctionLabel = "autoscaling"
	LoopWait                                FunctionLabel = "loopWait"
	BulkListAllGceInstances                 FunctionLabel = "bulkListInstances:listAllInstances"
	BulkListMigInstances                    FunctionLabel = "bulkListInstances:listMigInstances"
	DraSnapshotWrapSchedulerNodeInfo        FunctionLabel = "dra:snapshot:WrapSchedulerNodeInfo"
	DraSnapshotAddClaims                    FunctionLabel = "dra:snapshot:AddClaims"
	SnapshotRemovePodOwnedClaims            FunctionLabel = "dra:snapshot:RemovePodOwnedClaims"
	DraSnapshotPodClaims                    FunctionLabel = "dra:snapshot:PodClaims"
	DraSnapshotReservePodClaims             FunctionLabel = "dra:snapshot:ReservePodClaims"
	DraSnapshotUnreservePodClaims           FunctionLabel = "dra:snapshot:UnreservePodClaims"
	DraSnapshotNodeResourceSlices           FunctionLabel = "dra:snapshot:NodeResourceSlices"
	DraSnapshotAddNodeResourceSlices        FunctionLabel = "dra:snapshot:AddNodeResourceSlices"
	DraSnapshotRemoveNodeResourceSlices     FunctionLabel = "dra:snapshot:RemoveNodeResourceSlices"
	DraSnapshotSignalClaimPendingAllocation FunctionLabel = "dra:snapshot:SignalClaimPendingAllocation"
	DraSnapshotListAllAllocatedDevices      FunctionLabel = "dra:snapshot:ListAllAllocatedDevices"
	DraSnapshotGatherAllocatedState         FunctionLabel = "dra:snapshot:GatherAllocatedState"
	DraSnapshotCommit                       FunctionLabel = "dra:snapshot:Commit"
	DraSnapshotRevert                       FunctionLabel = "dra:snapshot:Revert"
	DraSnapshotFork                         FunctionLabel = "dra:snapshot:Fork"
)

const (
	DraSnapshotWrapSchedulerNodeInfoKey FunctionLabelKey = iota
	DraSnapshotAddClaimsKey
	SnapshotRemovePodOwnedClaimsKey
	DraSnapshotPodClaimsKey
	DraSnapshotReservePodClaimsKey
	DraSnapshotUnreservePodClaimsKey
	DraSnapshotNodeResourceSlicesKey
	DraSnapshotAddNodeResourceSlicesKey
	DraSnapshotRemoveNodeResourceSlicesKey
	DraSnapshotSignalClaimPendingAllocationKey
	DraSnapshotListAllAllocatedDevicesKey
	DraSnapshotGatherAllocatedStateKey
	DraSnapshotCommitKey
	DraSnapshotRevertKey
	DraSnapshotForkKey
	// The key is used for static validation and shouldn't be used for anything else
	_MaxFunctionLabelKey
)

// This is a compile-time check to ensure that the number of function labels
// is not greater than the capacity of the duration counter.
//
// If you are reading this - likely that you've introduced a new function label
// and it doesn't fit in the current capacity of the duration counter.
// In that case, you should increase the capacity of the duration counter or
// re-assess whether we need that many function labels.
var _ [DurationCountersCapacity - int(_MaxFunctionLabelKey)]struct{}

type caMetrics struct {
	registry metrics.KubeRegistry

	clusterSafeToAutoscale *k8smetrics.Gauge
	nodesCount             *k8smetrics.GaugeVec
	nodeGroupsCount        *k8smetrics.GaugeVec
	unschedulablePodsCount *k8smetrics.GaugeVec
	maxNodesCount          *k8smetrics.Gauge
	cpuCurrentCores        *k8smetrics.Gauge
	cpuLimitsCores         *k8smetrics.GaugeVec
	memoryCurrentBytes     *k8smetrics.Gauge
	memoryLimitsBytes      *k8smetrics.GaugeVec
	nodesGroupMinNodes     *k8smetrics.GaugeVec
	nodesGroupMaxNodes     *k8smetrics.GaugeVec
	nodesGroupTargetSize   *k8smetrics.GaugeVec
	nodesGroupHealthiness  *k8smetrics.GaugeVec
	nodeGroupBackOffStatus *k8smetrics.GaugeVec

	// Metrics related to autoscaler execution
	lastActivity               *k8smetrics.GaugeVec
	functionDuration           *k8smetrics.HistogramVec
	functionAggregatedDuration *k8smetrics.HistogramVec
	functionAggregatedTracker  *DurationCounter
	functionDurationSummary    *k8smetrics.SummaryVec
	pendingNodeDeletions       *k8smetrics.Gauge

	// Metrics related to autoscaler operations
	errorsCount                      *k8smetrics.CounterVec
	scaleUpCount                     *k8smetrics.Counter
	gpuScaleUpCount                  *k8smetrics.CounterVec
	failedScaleUpCount               *k8smetrics.CounterVec
	failedGPUScaleUpCount            *k8smetrics.CounterVec
	scaleDownCount                   *k8smetrics.CounterVec
	gpuScaleDownCount                *k8smetrics.CounterVec
	evictionsCount                   *k8smetrics.CounterVec
	unneededNodesCount               *k8smetrics.Gauge
	unremovableNodesCount            *k8smetrics.GaugeVec
	scaleDownInCooldown              *k8smetrics.Gauge
	oldUnregisteredNodesRemovedCount *k8smetrics.Counter
	overflowingControllersCount      *k8smetrics.Gauge
	skippedScaleEventsCount          *k8smetrics.CounterVec
	nodeGroupCreationCount           *k8smetrics.CounterVec
	nodeGroupDeletionCount           *k8smetrics.CounterVec
	nodeTaintsCount                  *k8smetrics.GaugeVec
	inconsistentInstancesMigsCount   *k8smetrics.Gauge
	binpackingHeterogeneity          *k8smetrics.HistogramVec
	maxNodeSkipEvalDurationSeconds   *k8smetrics.Gauge
	scaleDownNodeRemovalLatency      *k8smetrics.HistogramVec
}

func newCaMetrics() *caMetrics {
	return &caMetrics{
		/**** Metrics related to cluster state ****/
		clusterSafeToAutoscale: k8smetrics.NewGauge(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "cluster_safe_to_autoscale",
				Help:      "Whether or not cluster is healthy enough for autoscaling. 1 if it is, 0 otherwise.",
			},
		),

		nodesCount: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "nodes_count",
				Help:      "Number of nodes in cluster.",
			}, []string{"state"},
		),

		nodeGroupsCount: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "node_groups_count",
				Help:      "Number of node groups managed by CA.",
			}, []string{"node_group_type"},
		),

		// Unschedulable pod count can be from scheduler-marked-unschedulable pods or not-yet-processed pods (unknown)
		unschedulablePodsCount: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "unschedulable_pods_count",
				Help:      "Number of unschedulable pods in the cluster.",
			}, []string{"type"},
		),

		maxNodesCount: k8smetrics.NewGauge(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "max_nodes_count",
				Help:      "Maximum number of nodes in all node groups",
			},
		),

		cpuCurrentCores: k8smetrics.NewGauge(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "cluster_cpu_current_cores",
				Help:      "Current number of cores in the cluster, minus deleting nodes.",
			},
		),

		cpuLimitsCores: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "cpu_limits_cores",
				Help:      "Minimum and maximum number of cores in the cluster.",
			}, []string{"direction"},
		),

		memoryCurrentBytes: k8smetrics.NewGauge(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "cluster_memory_current_bytes",
				Help:      "Current number of bytes of memory in the cluster, minus deleting nodes.",
			},
		),

		memoryLimitsBytes: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "memory_limits_bytes",
				Help:      "Minimum and maximum number of bytes of memory in cluster.",
			}, []string{"direction"},
		),

		nodesGroupMinNodes: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "node_group_min_count",
				Help:      "Minimum number of nodes in the node group",
			}, []string{"node_group"},
		),

		nodesGroupMaxNodes: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "node_group_max_count",
				Help:      "Maximum number of nodes in the node group",
			}, []string{"node_group"},
		),

		nodesGroupTargetSize: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "node_group_target_count",
				Help:      "Target number of nodes in the node group by CA.",
			}, []string{"node_group"},
		),

		nodesGroupHealthiness: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "node_group_healthiness",
				Help:      "Whether or not node group is healthy enough for autoscaling. 1 if it is, 0 otherwise.",
			}, []string{"node_group"},
		),

		nodeGroupBackOffStatus: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "node_group_backoff_status",
				Help:      "Whether or not node group is backoff for not autoscaling. 1 if it is, 0 otherwise.",
			}, []string{"node_group", "reason"},
		),

		/**** Metrics related to autoscaler execution ****/
		lastActivity: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "last_activity",
				Help:      "Last time certain part of CA logic executed.",
			}, []string{"activity"},
		),

		functionDuration: k8smetrics.NewHistogramVec(
			&k8smetrics.HistogramOpts{
				Namespace: caNamespace,
				Name:      "function_duration_seconds",
				Help:      "Time taken by various parts of CA main loop.",
				Buckets:   k8smetrics.ExponentialBuckets(0.01, 1.5, 30), // 0.01, 0.015, 0.0225, ..., 852.2269299239293, 1278.3403948858938
			}, []string{"function"},
		),

		functionAggregatedDuration: k8smetrics.NewHistogramVec(
			&k8smetrics.HistogramOpts{
				Namespace: caNamespace,
				Name:      "function_aggregated_duration_seconds",
				Help:      "Time taken by various parts of autoscaler aggregated per loop",
				Buckets:   k8smetrics.ExponentialBuckets(0.01, 1.5, 30), // 0.01, 0.015, 0.0225, ..., 852.2269299239293, 1278.3403948858938
			}, []string{"function"},
		),
		functionAggregatedTracker: newRegisteredDurationCounter(),

		functionDurationSummary: k8smetrics.NewSummaryVec(
			&k8smetrics.SummaryOpts{
				Namespace: caNamespace,
				Name:      "function_duration_quantile_seconds",
				Help:      "Quantiles of time taken by various parts of CA main loop.",
				MaxAge:    time.Hour,
			}, []string{"function"},
		),

		pendingNodeDeletions: k8smetrics.NewGauge(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "pending_node_deletions",
				Help:      "Number of nodes that haven't been removed or aborted after finished scale-down phase.",
			},
		),

		/**** Metrics related to autoscaler operations ****/
		errorsCount: k8smetrics.NewCounterVec(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "errors_total",
				Help:      "The number of CA loops failed due to an error.",
			}, []string{"type"},
		),

		scaleUpCount: k8smetrics.NewCounter(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "scaled_up_nodes_total",
				Help:      "Number of nodes added by CA.",
			},
		),

		gpuScaleUpCount: k8smetrics.NewCounterVec(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "scaled_up_gpu_nodes_total",
				Help:      "Number of GPU nodes added by CA, by GPU name.",
			}, []string{"gpu_resource_name", "gpu_name"},
		),

		failedScaleUpCount: k8smetrics.NewCounterVec(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "failed_scale_ups_total",
				Help:      "Number of times scale-up operation has failed.",
			}, []string{"reason"},
		),

		failedGPUScaleUpCount: k8smetrics.NewCounterVec(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "failed_gpu_scale_ups_total",
				Help:      "Number of times scale-up operation has failed.",
			}, []string{"reason", "gpu_resource_name", "gpu_name"},
		),

		scaleDownCount: k8smetrics.NewCounterVec(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "scaled_down_nodes_total",
				Help:      "Number of nodes removed by CA.",
			}, []string{"reason"},
		),

		gpuScaleDownCount: k8smetrics.NewCounterVec(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "scaled_down_gpu_nodes_total",
				Help:      "Number of GPU nodes removed by CA, by reason and GPU name.",
			}, []string{"reason", "gpu_resource_name", "gpu_name"},
		),

		evictionsCount: k8smetrics.NewCounterVec(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "evicted_pods_total",
				Help:      "Number of pods evicted by CA",
			}, []string{"eviction_result"},
		),

		unneededNodesCount: k8smetrics.NewGauge(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "unneeded_nodes_count",
				Help:      "Number of nodes currently considered unneeded by CA.",
			},
		),

		unremovableNodesCount: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "unremovable_nodes_count",
				Help:      "Number of nodes currently considered unremovable by CA.",
			},
			[]string{"reason"},
		),

		scaleDownInCooldown: k8smetrics.NewGauge(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "scale_down_in_cooldown",
				Help:      "Whether or not the scale down is in cooldown. 1 if its, 0 otherwise.",
			},
		),

		oldUnregisteredNodesRemovedCount: k8smetrics.NewCounter(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "old_unregistered_nodes_removed_count",
				Help:      "Number of unregistered nodes removed by CA.",
			},
		),

		overflowingControllersCount: k8smetrics.NewGauge(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "overflowing_controllers_count",
				Help:      "Number of controllers that own a large set of heterogenous pods, preventing CA from treating these pods as equivalent.",
			},
		),

		skippedScaleEventsCount: k8smetrics.NewCounterVec(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "skipped_scale_events_count",
				Help:      "Count of scaling events that the CA has chosen to skip.",
			},
			[]string{"direction", "reason"},
		),

		nodeGroupCreationCount: k8smetrics.NewCounterVec(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "created_node_groups_total",
				Help:      "Number of node groups created by Node Autoprovisioning.",
			},
			[]string{"group_type"},
		),

		nodeGroupDeletionCount: k8smetrics.NewCounterVec(
			&k8smetrics.CounterOpts{
				Namespace: caNamespace,
				Name:      "deleted_node_groups_total",
				Help:      "Number of node groups deleted by Node Autoprovisioning.",
			},
			[]string{"group_type"},
		),

		nodeTaintsCount: k8smetrics.NewGaugeVec(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "node_taints_count",
				Help:      "Number of taints per type used in the cluster.",
			},
			[]string{"type"},
		),

		inconsistentInstancesMigsCount: k8smetrics.NewGauge(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "inconsistent_instances_migs_count",
				Help:      "Number of migs where instance count according to InstanceGroupManagers.List() differs from the results of Instances.List(). This can happen when some instances are abandoned or a user edits instance 'created-by' metadata.",
			},
		),

		binpackingHeterogeneity: k8smetrics.NewHistogramVec(
			&k8smetrics.HistogramOpts{
				Namespace: caNamespace,
				Name:      "binpacking_heterogeneity",
				Help:      "Number of groups of equivalent pods being processed as a part of the same binpacking simulation.",
				Buckets:   k8smetrics.ExponentialBuckets(1, 2, 6), // 1, 2, 4, ..., 32
			}, []string{"instance_type", "cpu_count", "namespace_count"},
		),

		maxNodeSkipEvalDurationSeconds: k8smetrics.NewGauge(
			&k8smetrics.GaugeOpts{
				Namespace: caNamespace,
				Name:      "max_node_skip_eval_duration_seconds",
				Help:      "Maximum evaluation time of a node being skipped during ScaleDown.",
			},
		),

		scaleDownNodeRemovalLatency: k8smetrics.NewHistogramVec(
			&k8smetrics.HistogramOpts{
				Namespace: caNamespace,
				Name:      "node_removal_latency_seconds",
				Help:      "Latency from when an unneeded node is eligible for scale down until it is removed (deleted=true) or it became needed again (deleted=false).",
				Buckets:   k8smetrics.ExponentialBuckets(1, 1.5, 19), // ~1s → ~24min
			}, []string{"deleted"},
		),
	}
}


// newRegisteredDurationCounter builds a duration counter and registers all the known function labels into it
// after this call - it would be ready to accept duration increases.
func newRegisteredDurationCounter() *DurationCounter {
	dc := NewDurationCounter()
	dc.RegisterLabel(int(DraSnapshotWrapSchedulerNodeInfoKey), string(DraSnapshotWrapSchedulerNodeInfo))
	dc.RegisterLabel(int(DraSnapshotAddClaimsKey), string(DraSnapshotAddClaims))
	dc.RegisterLabel(int(SnapshotRemovePodOwnedClaimsKey), string(SnapshotRemovePodOwnedClaims))
	dc.RegisterLabel(int(DraSnapshotPodClaimsKey), string(DraSnapshotPodClaims))
	dc.RegisterLabel(int(DraSnapshotReservePodClaimsKey), string(DraSnapshotReservePodClaims))
	dc.RegisterLabel(int(DraSnapshotUnreservePodClaimsKey), string(DraSnapshotUnreservePodClaims))
	dc.RegisterLabel(int(DraSnapshotNodeResourceSlicesKey), string(DraSnapshotNodeResourceSlices))
	dc.RegisterLabel(int(DraSnapshotAddNodeResourceSlicesKey), string(DraSnapshotAddNodeResourceSlices))
	dc.RegisterLabel(int(DraSnapshotRemoveNodeResourceSlicesKey), string(DraSnapshotRemoveNodeResourceSlices))
	dc.RegisterLabel(int(DraSnapshotSignalClaimPendingAllocationKey), string(DraSnapshotSignalClaimPendingAllocation))
	dc.RegisterLabel(int(DraSnapshotListAllAllocatedDevicesKey), string(DraSnapshotListAllAllocatedDevices))
	dc.RegisterLabel(int(DraSnapshotGatherAllocatedStateKey), string(DraSnapshotGatherAllocatedState))
	dc.RegisterLabel(int(DraSnapshotCommitKey), string(DraSnapshotCommit))
	dc.RegisterLabel(int(DraSnapshotRevertKey), string(DraSnapshotRevert))
	dc.RegisterLabel(int(DraSnapshotForkKey), string(DraSnapshotFork))
	return dc
}

func (m *caMetrics) mustRegister(cs ...k8smetrics.Registerable) {
	if m.registry != nil {
		m.registry.MustRegister(cs...)
		return
	}
	legacyregistry.MustRegister(cs...)
}

// RegisterAll registers all metrics.
func (m *caMetrics) RegisterAll(emitPerNodeGroupMetrics bool) {
	m.mustRegister(m.clusterSafeToAutoscale)
	m.mustRegister(m.nodesCount)
	m.mustRegister(m.nodeGroupsCount)
	m.mustRegister(m.unschedulablePodsCount)
	m.mustRegister(m.maxNodesCount)
	m.mustRegister(m.cpuCurrentCores)
	m.mustRegister(m.cpuLimitsCores)
	m.mustRegister(m.memoryCurrentBytes)
	m.mustRegister(m.memoryLimitsBytes)
	m.mustRegister(m.lastActivity)
	m.mustRegister(m.functionDuration)
	m.mustRegister(m.functionDurationSummary)
	m.mustRegister(m.functionAggregatedDuration)
	m.mustRegister(m.errorsCount)
	m.mustRegister(m.scaleUpCount)
	m.mustRegister(m.gpuScaleUpCount)
	m.mustRegister(m.failedScaleUpCount)
	m.mustRegister(m.failedGPUScaleUpCount)
	m.mustRegister(m.scaleDownCount)
	m.mustRegister(m.gpuScaleDownCount)
	m.mustRegister(m.evictionsCount)
	m.mustRegister(m.unneededNodesCount)
	m.mustRegister(m.unremovableNodesCount)
	m.mustRegister(m.scaleDownInCooldown)
	m.mustRegister(m.oldUnregisteredNodesRemovedCount)
	m.mustRegister(m.overflowingControllersCount)
	m.mustRegister(m.skippedScaleEventsCount)
	m.mustRegister(m.nodeGroupCreationCount)
	m.mustRegister(m.nodeGroupDeletionCount)
	m.mustRegister(m.pendingNodeDeletions)
	m.mustRegister(m.nodeTaintsCount)
	m.mustRegister(m.inconsistentInstancesMigsCount)
	m.mustRegister(m.binpackingHeterogeneity)
	m.mustRegister(m.maxNodeSkipEvalDurationSeconds)
	m.mustRegister(m.scaleDownNodeRemovalLatency)

	if emitPerNodeGroupMetrics {
		m.mustRegister(m.nodesGroupMinNodes)
		m.mustRegister(m.nodesGroupMaxNodes)
		m.mustRegister(m.nodesGroupTargetSize)
		m.mustRegister(m.nodesGroupHealthiness)
		m.mustRegister(m.nodeGroupBackOffStatus)
	}
}

// InitMetrics initializes all metrics
func (m *caMetrics) InitMetrics() {
	for _, errorType := range []errors.AutoscalerErrorType{errors.CloudProviderError, errors.ApiCallError, errors.InternalError, errors.TransientError, errors.ConfigurationError, errors.NodeGroupDoesNotExistError, errors.UnexpectedScaleDownStateError} {
		m.errorsCount.WithLabelValues(string(errorType)).Add(0)
	}

	for _, reason := range []FailedScaleUpReason{CloudProviderError, APIError, Timeout} {
		m.scaleDownCount.WithLabelValues(string(reason)).Add(0)
		m.failedScaleUpCount.WithLabelValues(string(reason)).Add(0)
	}

	for _, result := range []PodEvictionResult{PodEvictionSucceed, PodEvictionFailed} {
		m.evictionsCount.WithLabelValues(string(result)).Add(0)
	}

	m.skippedScaleEventsCount.WithLabelValues(DirectionScaleDown, CpuResourceLimit).Add(0)
	m.skippedScaleEventsCount.WithLabelValues(DirectionScaleDown, MemoryResourceLimit).Add(0)
	m.skippedScaleEventsCount.WithLabelValues(DirectionScaleUp, CpuResourceLimit).Add(0)
	m.skippedScaleEventsCount.WithLabelValues(DirectionScaleUp, MemoryResourceLimit).Add(0)

}

// UpdateDurationFromStart records the duration of the step identified by the
// label using start time
func (m *caMetrics) UpdateDurationFromStart(label FunctionLabel, start time.Time) {
	m.UpdateDuration(label, time.Since(start))
}

// UpdateDuration records the duration of the step identified by the label
func (m *caMetrics) UpdateDuration(label FunctionLabel, duration time.Duration) {
	if duration > LogLongDurationThreshold {
		klog.V(4).Infof("Function %s took %v to complete", label, duration)
	}
	m.functionDuration.WithLabelValues(string(label)).Observe(duration.Seconds())
	m.functionDurationSummary.WithLabelValues(string(label)).Observe(duration.Seconds())
}

// UpdateDurationAggregated records the duration of the step identified by the
// label key.
//
// Warning: if the label key is not registered, it will be ignored when flushing the
// aggregated durations.
func (m *caMetrics) UpdateDurationAggregated(labelKey FunctionLabelKey, duration time.Duration) {
	m.functionAggregatedTracker.Increment(int(labelKey), duration)
}

// UpdateDurationAggregatedFromStart records the duration of the step identified by the
// label key using start time.
//
// Warning: if the label key is not registered, it will be ignored when flushing the
// aggregated durations.
func (m *caMetrics) UpdateDurationAggregatedFromStart(labelKey FunctionLabelKey, start time.Time) {
	m.functionAggregatedTracker.Increment(int(labelKey), time.Since(start))
}

// FlushAggregatedDurations records aggregated durations of the steps tracked
// by UpdateDuration, resets the tracker and updates the metric.
func (m *caMetrics) FlushAggregatedDurations() {
	snapshot := m.functionAggregatedTracker.Snapshot()
	for label, duration := range snapshot {
		m.functionAggregatedDuration.WithLabelValues(label).Observe(duration.Seconds())
	}
	m.functionAggregatedTracker.Reset()
}

// UpdateLastTime records the time the step identified by the label was started
func (m *caMetrics) UpdateLastTime(label FunctionLabel, now time.Time) {
	m.lastActivity.WithLabelValues(string(label)).Set(float64(now.Unix()))
}

// UpdateClusterSafeToAutoscale records if cluster is safe to autoscale
func (m *caMetrics) UpdateClusterSafeToAutoscale(safe bool) {
	if safe {
		m.clusterSafeToAutoscale.Set(1)
	} else {
		m.clusterSafeToAutoscale.Set(0)
	}
}

// UpdateNodesCount records the number of nodes in cluster
func (m *caMetrics) UpdateNodesCount(ready, unready, starting, longUnregistered, unregistered int) {
	m.nodesCount.WithLabelValues(readyLabel).Set(float64(ready))
	m.nodesCount.WithLabelValues(unreadyLabel).Set(float64(unready))
	m.nodesCount.WithLabelValues(startingLabel).Set(float64(starting))
	m.nodesCount.WithLabelValues(longUnregisteredLabel).Set(float64(longUnregistered))
	m.nodesCount.WithLabelValues(unregisteredLabel).Set(float64(unregistered))
}

// UpdateNodeGroupsCount records the number of node groups managed by CA
func (m *caMetrics) UpdateNodeGroupsCount(autoscaled, autoprovisioned int) {
	m.nodeGroupsCount.WithLabelValues(string(autoscaledGroup)).Set(float64(autoscaled))
	m.nodeGroupsCount.WithLabelValues(string(autoprovisionedGroup)).Set(float64(autoprovisioned))
}

// UpdateUnschedulablePodsCount records number of currently unschedulable pods
func (m *caMetrics) UpdateUnschedulablePodsCount(uschedulablePodsCount, schedulerUnprocessedCount int) {
	m.UpdateUnschedulablePodsCountWithLabel(uschedulablePodsCount, "unschedulable")
	m.UpdateUnschedulablePodsCountWithLabel(schedulerUnprocessedCount, "scheduler_unprocessed")
}

// UpdateUnschedulablePodsCountWithLabel records number of currently unschedulable pods wil label "type" value "label"
func (m *caMetrics) UpdateUnschedulablePodsCountWithLabel(uschedulablePodsCount int, label string) {
	m.unschedulablePodsCount.WithLabelValues(label).Set(float64(uschedulablePodsCount))
}

// UpdateMaxNodesCount records the current maximum number of nodes being set for all node groups
func (m *caMetrics) UpdateMaxNodesCount(nodesCount int) {
	m.maxNodesCount.Set(float64(nodesCount))
}

// UpdateClusterCPUCurrentCores records the number of cores in the cluster, minus deleting nodes
func (m *caMetrics) UpdateClusterCPUCurrentCores(coresCount int64) {
	m.cpuCurrentCores.Set(float64(coresCount))
}

// UpdateCPULimitsCores records the minimum and maximum number of cores in the cluster
func (m *caMetrics) UpdateCPULimitsCores(minCoresCount int64, maxCoresCount int64) {
	m.cpuLimitsCores.WithLabelValues("minimum").Set(float64(minCoresCount))
	m.cpuLimitsCores.WithLabelValues("maximum").Set(float64(maxCoresCount))
}

// UpdateClusterMemoryCurrentBytes records the number of bytes of memory in the cluster, minus deleting nodes
func (m *caMetrics) UpdateClusterMemoryCurrentBytes(memoryCount int64) {
	m.memoryCurrentBytes.Set(float64(memoryCount))
}

// UpdateMemoryLimitsBytes records the minimum and maximum bytes of memory in the cluster
func (m *caMetrics) UpdateMemoryLimitsBytes(minMemoryCount int64, maxMemoryCount int64) {
	m.memoryLimitsBytes.WithLabelValues("minimum").Set(float64(minMemoryCount))
	m.memoryLimitsBytes.WithLabelValues("maximum").Set(float64(maxMemoryCount))
}

// UpdateNodeGroupMin records the node group minimum allowed number of nodes
func (m *caMetrics) UpdateNodeGroupMin(nodeGroup string, minNodes int) {
	m.nodesGroupMinNodes.WithLabelValues(nodeGroup).Set(float64(minNodes))
}

// UpdateNodeGroupMax records the node group maximum allowed number of nodes
func (m *caMetrics) UpdateNodeGroupMax(nodeGroup string, maxNodes int) {
	m.nodesGroupMaxNodes.WithLabelValues(nodeGroup).Set(float64(maxNodes))
}

// UpdateNodeGroupTargetSize records the node group target size
func (m *caMetrics) UpdateNodeGroupTargetSize(targetSizes map[string]int) {
	for nodeGroup, targetSize := range targetSizes {
		m.nodesGroupTargetSize.WithLabelValues(nodeGroup).Set(float64(targetSize))
	}
}

// UpdateNodeGroupHealthStatus records if node group is healthy to autoscaling
func (m *caMetrics) UpdateNodeGroupHealthStatus(nodeGroup string, healthy bool) {
	if healthy {
		m.nodesGroupHealthiness.WithLabelValues(nodeGroup).Set(1)
	} else {
		m.nodesGroupHealthiness.WithLabelValues(nodeGroup).Set(0)
	}
}

// UpdateNodeGroupBackOffStatus records if node group is backoff for not autoscaling
func (m *caMetrics) UpdateNodeGroupBackOffStatus(nodeGroup string, backoffReasonStatus map[string]bool) {
	if len(backoffReasonStatus) == 0 {
		m.nodeGroupBackOffStatus.WithLabelValues(nodeGroup, "").Set(0)
	} else {
		for reason, backoff := range backoffReasonStatus {
			if backoff {
				m.nodeGroupBackOffStatus.WithLabelValues(nodeGroup, reason).Set(1)
			} else {
				m.nodeGroupBackOffStatus.WithLabelValues(nodeGroup, reason).Set(0)
			}
		}
	}
}

// RegisterError records any errors preventing Cluster Autoscaler from working.
// No more than one error should be recorded per loop.
func (m *caMetrics) RegisterError(err errors.AutoscalerError) {
	m.errorsCount.WithLabelValues(string(err.Type())).Add(1.0)
}

// RegisterScaleUp records number of nodes added by scale up
func (m *caMetrics) RegisterScaleUp(nodesCount int, gpuResourceName, gpuType string) {
	m.scaleUpCount.Add(float64(nodesCount))
	if gpuType != gpu.MetricsNoGPU {
		m.gpuScaleUpCount.WithLabelValues(gpuResourceName, gpuType).Add(float64(nodesCount))
	}
}

// RegisterFailedScaleUp records a failed scale-up operation
func (m *caMetrics) RegisterFailedScaleUp(reason FailedScaleUpReason, gpuResourceName, gpuType string) {
	m.failedScaleUpCount.WithLabelValues(string(reason)).Inc()
	if gpuType != gpu.MetricsNoGPU {
		m.failedGPUScaleUpCount.WithLabelValues(string(reason), gpuResourceName, gpuType).Inc()
	}
}

// RegisterScaleDown records number of nodes removed by scale down
func (m *caMetrics) RegisterScaleDown(nodesCount int, gpuResourceName, gpuType string, reason NodeScaleDownReason) {
	m.scaleDownCount.WithLabelValues(string(reason)).Add(float64(nodesCount))
	if gpuType != gpu.MetricsNoGPU {
		m.gpuScaleDownCount.WithLabelValues(string(reason), gpuResourceName, gpuType).Add(float64(nodesCount))
	}
}

// RegisterEvictions records number of evicted pods succeed or failed
func (m *caMetrics) RegisterEvictions(podsCount int, result PodEvictionResult) {
	m.evictionsCount.WithLabelValues(string(result)).Add(float64(podsCount))
}

// UpdateUnneededNodesCount records number of currently unneeded nodes
func (m *caMetrics) UpdateUnneededNodesCount(nodesCount int) {
	m.unneededNodesCount.Set(float64(nodesCount))
}

// UpdateUnremovableNodesCount records number of currently unremovable nodes
func (m *caMetrics) UpdateUnremovableNodesCount(unremovableReasonCounts map[string]int) {
	for reason, count := range unremovableReasonCounts {
		m.unremovableNodesCount.WithLabelValues(reason).Set(float64(count))
	}
}

// RegisterNodeGroupCreation registers node group creation
func (m *caMetrics) RegisterNodeGroupCreation() {
	m.RegisterNodeGroupCreationWithLabelValues("")
}

// RegisterNodeGroupCreationWithLabelValues registers node group creation with the provided labels
func (m *caMetrics) RegisterNodeGroupCreationWithLabelValues(groupType string) {
	m.nodeGroupCreationCount.WithLabelValues(groupType).Add(1.0)
}

// RegisterNodeGroupDeletion registers node group deletion
func (m *caMetrics) RegisterNodeGroupDeletion() {
	m.RegisterNodeGroupDeletionWithLabelValues("")
}

// RegisterNodeGroupDeletionWithLabelValues registers node group deletion with the provided labels
func (m *caMetrics) RegisterNodeGroupDeletionWithLabelValues(groupType string) {
	m.nodeGroupDeletionCount.WithLabelValues(groupType).Add(1.0)
}

// UpdateScaleDownInCooldown registers if the cluster autoscaler
// scaledown is in cooldown
func (m *caMetrics) UpdateScaleDownInCooldown(inCooldown bool) {
	if inCooldown {
		m.scaleDownInCooldown.Set(1.0)
	} else {
		m.scaleDownInCooldown.Set(0.0)
	}
}

// RegisterOldUnregisteredNodesRemoved records number of old unregistered
// nodes that have been removed by the cluster autoscaler
func (m *caMetrics) RegisterOldUnregisteredNodesRemoved(nodesCount int) {
	m.oldUnregisteredNodesRemovedCount.Add(float64(nodesCount))
}

// UpdateOverflowingControllers sets the number of controllers that could not
// have their pods cached.
func (m *caMetrics) UpdateOverflowingControllers(count int) {
	m.overflowingControllersCount.Set(float64(count))
}

// RegisterSkippedScaleDownCPU increases the count of skipped scale outs because of CPU resource limits
func (m *caMetrics) RegisterSkippedScaleDownCPU() {
	m.skippedScaleEventsCount.WithLabelValues(DirectionScaleDown, CpuResourceLimit).Add(1.0)
}

// RegisterSkippedScaleDownMemory increases the count of skipped scale outs because of Memory resource limits
func (m *caMetrics) RegisterSkippedScaleDownMemory() {
	m.skippedScaleEventsCount.WithLabelValues(DirectionScaleDown, MemoryResourceLimit).Add(1.0)
}

// RegisterSkippedScaleUpCPU increases the count of skipped scale outs because of CPU resource limits
func (m *caMetrics) RegisterSkippedScaleUpCPU() {
	m.skippedScaleEventsCount.WithLabelValues(DirectionScaleUp, CpuResourceLimit).Add(1.0)
}

// RegisterSkippedScaleUpMemory increases the count of skipped scale outs because of Memory resource limits
func (m *caMetrics) RegisterSkippedScaleUpMemory() {
	m.skippedScaleEventsCount.WithLabelValues(DirectionScaleUp, MemoryResourceLimit).Add(1.0)
}

// ObservePendingNodeDeletions records the current value of nodes_pending_deletion metric
func (m *caMetrics) ObservePendingNodeDeletions(value int) {
	m.pendingNodeDeletions.Set(float64(value))
}

// ObserveNodeTaintsCount records the node taints count of given type.
func (m *caMetrics) ObserveNodeTaintsCount(taintType string, count float64) {
	m.nodeTaintsCount.WithLabelValues(taintType).Set(count)
}

// UpdateInconsistentInstancesMigsCount records the observed number of migs where instance count
// according to InstanceGroupManagers.List() differs from the results of Instances.List().
// This can happen when some instances are abandoned or a user edits instance 'created-by' metadata.
func (m *caMetrics) UpdateInconsistentInstancesMigsCount(migCount int) {
	m.inconsistentInstancesMigsCount.Set(float64(migCount))
}

// ObserveBinpackingHeterogeneity records the number of pod equivalence groups
// considered in a single binpacking estimation.
func (m *caMetrics) ObserveBinpackingHeterogeneity(instanceType, cpuCount, namespaceCount string, pegCount int) {
	m.binpackingHeterogeneity.WithLabelValues(instanceType, cpuCount, namespaceCount).Observe(float64(pegCount))
}

// UpdateScaleDownNodeRemovalLatency records the time after which node was deleted/needed
// again after being marked unneded
func (m *caMetrics) UpdateScaleDownNodeRemovalLatency(deleted bool, duration time.Duration) {
	m.scaleDownNodeRemovalLatency.WithLabelValues(strconv.FormatBool(deleted)).Observe(duration.Seconds())
}

// ObserveMaxNodeSkipEvalDurationSeconds records the longest time during which node was skipped during ScaleDown.
// If a node is skipped multiple times consecutively, we store only the earliest timestamp.
func (m *caMetrics) ObserveMaxNodeSkipEvalDurationSeconds(duration time.Duration) {
	m.maxNodeSkipEvalDurationSeconds.Set(duration.Seconds())
}
