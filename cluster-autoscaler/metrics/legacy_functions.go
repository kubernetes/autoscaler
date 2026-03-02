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

	"k8s.io/autoscaler/cluster-autoscaler/simulator"

	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	_ "k8s.io/component-base/metrics/prometheus/restclient" // for client-go metrics registration
)

// DefaultMetrics is a default implementation using global legacyregistry
var DefaultMetrics = newCaMetrics()

// InitMetrics initializes all metrics
func InitMetrics() {
	DefaultMetrics.InitMetrics()
}

// RegisterAll registers all metrics
func RegisterAll(emitPerNodeGroupMetrics bool) {
	DefaultMetrics.RegisterAll(emitPerNodeGroupMetrics)
}

// UpdateDurationFromStart records the duration of the step identified by the
// label using start time
func UpdateDurationFromStart(label FunctionLabel, start time.Time) {
	DefaultMetrics.UpdateDurationFromStart(label, start)
}

// UpdateDuration records the duration of the step identified by the label
func UpdateDuration(label FunctionLabel, duration time.Duration) {
	DefaultMetrics.UpdateDuration(label, duration)
}

// UpdateLastTime records the time the step identified by the label was started
func UpdateLastTime(label FunctionLabel, now time.Time) {
	DefaultMetrics.UpdateLastTime(label, now)
}

// UpdateClusterSafeToAutoscale records if cluster is safe to autoscale
func UpdateClusterSafeToAutoscale(safe bool) {
	DefaultMetrics.UpdateClusterSafeToAutoscale(safe)
}

// UpdateNodesCount records the number of nodes in cluster
func UpdateNodesCount(ready, unready, starting, longUnregistered, unregistered int) {
	DefaultMetrics.UpdateNodesCount(ready, unready, starting, longUnregistered, unregistered)
}

// UpdateNodeGroupsCount records the number of node groups managed by CA
func UpdateNodeGroupsCount(autoscaled, autoprovisioned int) {
	DefaultMetrics.UpdateNodeGroupsCount(autoscaled, autoprovisioned)
}

// UpdateUnschedulablePodsCount records number of currently unschedulable pods
func UpdateUnschedulablePodsCount(uschedulablePodsCount, schedulerUnprocessedCount int) {
	DefaultMetrics.UpdateUnschedulablePodsCount(uschedulablePodsCount, schedulerUnprocessedCount)
}

// UpdateUnschedulablePodsCountWithLabel records number of currently unschedulable pods wil label "type" value "label"
func UpdateUnschedulablePodsCountWithLabel(uschedulablePodsCount int, label string) {
	DefaultMetrics.UpdateUnschedulablePodsCountWithLabel(uschedulablePodsCount, label)
}

// UpdateMaxNodesCount records the current maximum number of nodes being set for all node groups
func UpdateMaxNodesCount(nodesCount int) {
	DefaultMetrics.UpdateMaxNodesCount(nodesCount)
}

// UpdateClusterCPUCurrentCores records the number of cores in the cluster, minus deleting nodes
func UpdateClusterCPUCurrentCores(coresCount int64) {
	DefaultMetrics.UpdateClusterCPUCurrentCores(coresCount)
}

// UpdateCPULimitsCores records the minimum and maximum number of cores in the cluster
func UpdateCPULimitsCores(minCoresCount int64, maxCoresCount int64) {
	DefaultMetrics.UpdateCPULimitsCores(minCoresCount, maxCoresCount)
}

// UpdateClusterMemoryCurrentBytes records the number of bytes of memory in the cluster, minus deleting nodes
func UpdateClusterMemoryCurrentBytes(memoryCount int64) {
	DefaultMetrics.UpdateClusterMemoryCurrentBytes(memoryCount)
}

// UpdateMemoryLimitsBytes records the minimum and maximum bytes of memory in the cluster
func UpdateMemoryLimitsBytes(minMemoryCount int64, maxMemoryCount int64) {
	DefaultMetrics.UpdateMemoryLimitsBytes(minMemoryCount, maxMemoryCount)
}

// UpdateNodeGroupMin records the node group minimum allowed number of nodes
func UpdateNodeGroupMin(nodeGroup string, minNodes int) {
	DefaultMetrics.UpdateNodeGroupMin(nodeGroup, minNodes)
}

// UpdateNodeGroupMax records the node group maximum allowed number of nodes
func UpdateNodeGroupMax(nodeGroup string, maxNodes int) {
	DefaultMetrics.UpdateNodeGroupMax(nodeGroup, maxNodes)
}

// UpdateNodeGroupTargetSize records the node group target size
func UpdateNodeGroupTargetSize(targetSizes map[string]int) {
	DefaultMetrics.UpdateNodeGroupTargetSize(targetSizes)
}

// UpdateNodeGroupHealthStatus records if node group is healthy to autoscaling
func UpdateNodeGroupHealthStatus(nodeGroup string, healthy bool) {
	DefaultMetrics.UpdateNodeGroupHealthStatus(nodeGroup, healthy)
}

// UpdateNodeGroupBackOffStatus records if node group is backoff for not autoscaling
func UpdateNodeGroupBackOffStatus(nodeGroup string, backoffReasonStatus map[string]bool) {
	DefaultMetrics.UpdateNodeGroupBackOffStatus(nodeGroup, backoffReasonStatus)
}

// RegisterError records any errors preventing Cluster Autoscaler from working.
// No more than one error should be recorded per loop.
func RegisterError(err errors.AutoscalerError) {
	DefaultMetrics.RegisterError(err)
}

// RegisterScaleUp records number of nodes added by scale up
func RegisterScaleUp(nodesCount int, gpuResourceName, gpuType, draDrivers string) {
	DefaultMetrics.RegisterScaleUp(nodesCount, gpuResourceName, gpuType, draDrivers)
}

// RegisterFailedScaleUp records a failed scale-up operation
func RegisterFailedScaleUp(reason FailedScaleUpReason, gpuResourceName, gpuType, draDrivers string) {
	DefaultMetrics.RegisterFailedScaleUp(reason, gpuResourceName, gpuType, draDrivers)
}

// RegisterScaleDown records number of nodes removed by scale down
func RegisterScaleDown(nodesCount int, gpuResourceName, gpuType string, reason NodeScaleDownReason, draDrivers string) {
	DefaultMetrics.RegisterScaleDown(nodesCount, gpuResourceName, gpuType, reason, draDrivers)
}

// RegisterEvictions records number of evicted pods succeed or failed
func RegisterEvictions(podsCount int, result PodEvictionResult) {
	DefaultMetrics.RegisterEvictions(podsCount, result)
}

// UpdateUnneededNodesCount records number of currently unneeded nodes
func UpdateUnneededNodesCount(nodesCount int) {
	DefaultMetrics.UpdateUnneededNodesCount(nodesCount)
}

// UpdateUnremovableNodesCount records number of currently unremovable nodes
func UpdateUnremovableNodesCount(unremovableReasonCounts map[simulator.UnremovableReason]int) {
	DefaultMetrics.UpdateUnremovableNodesCount(unremovableReasonCounts)
}

// RegisterNodeGroupCreation registers node group creation
func RegisterNodeGroupCreation() {
	DefaultMetrics.RegisterNodeGroupCreation()
}

// RegisterNodeGroupCreationWithLabelValues registers node group creation with the provided labels
func RegisterNodeGroupCreationWithLabelValues(groupType string) {
	DefaultMetrics.RegisterNodeGroupCreationWithLabelValues(groupType)
}

// RegisterNodeGroupDeletion registers node group deletion
func RegisterNodeGroupDeletion() {
	DefaultMetrics.RegisterNodeGroupDeletion()
}

// RegisterNodeGroupDeletionWithLabelValues registers node group deletion with the provided labels
func RegisterNodeGroupDeletionWithLabelValues(groupType string) {
	DefaultMetrics.RegisterNodeGroupDeletionWithLabelValues(groupType)
}

// UpdateScaleDownInCooldown registers if the cluster autoscaler
// scaledown is in cooldown
func UpdateScaleDownInCooldown(inCooldown bool) {
	DefaultMetrics.UpdateScaleDownInCooldown(inCooldown)
}

// RegisterOldUnregisteredNodesRemoved records number of old unregistered
// nodes that have been removed by the cluster autoscaler
func RegisterOldUnregisteredNodesRemoved(nodesCount int) {
	DefaultMetrics.RegisterOldUnregisteredNodesRemoved(nodesCount)
}

// UpdateOverflowingControllers sets the number of controllers that could not
// have their pods cached.
func UpdateOverflowingControllers(count int) {
	DefaultMetrics.UpdateOverflowingControllers(count)
}

// RegisterSkippedScaleDownCPU increases the count of skipped scale outs because of CPU resource limits
func RegisterSkippedScaleDownCPU() {
	DefaultMetrics.RegisterSkippedScaleDownCPU()
}

// RegisterSkippedScaleDownMemory increases the count of skipped scale outs because of Memory resource limits
func RegisterSkippedScaleDownMemory() {
	DefaultMetrics.RegisterSkippedScaleDownMemory()
}

// RegisterSkippedScaleUpCPU increases the count of skipped scale outs because of CPU resource limits
func RegisterSkippedScaleUpCPU() {
	DefaultMetrics.RegisterSkippedScaleUpCPU()
}

// RegisterSkippedScaleUpMemory increases the count of skipped scale outs because of Memory resource limits
func RegisterSkippedScaleUpMemory() {
	DefaultMetrics.RegisterSkippedScaleUpMemory()
}

// ObservePendingNodeDeletions records the current value of nodes_pending_deletion metric
func ObservePendingNodeDeletions(value int) {
	DefaultMetrics.ObservePendingNodeDeletions(value)
}

// ObserveNodeTaintsCount records the node taints count of given type.
func ObserveNodeTaintsCount(taintType string, count float64) {
	DefaultMetrics.ObserveNodeTaintsCount(taintType, count)
}

// UpdateInconsistentInstancesMigsCount records the observed number of migs where instance count
// according to InstanceGroupManagers.List() differs from the results of Instances.List().
// This can happen when some instances are abandoned or a user edits instance 'created-by' metadata.
func UpdateInconsistentInstancesMigsCount(migCount int) {
	DefaultMetrics.UpdateInconsistentInstancesMigsCount(migCount)
}

// ObserveBinpackingHeterogeneity records the number of pod equivalence groups
// considered in a single binpacking estimation.
func ObserveBinpackingHeterogeneity(instanceType, cpuCount, namespaceCount string, pegCount int) {
	DefaultMetrics.ObserveBinpackingHeterogeneity(instanceType, cpuCount, namespaceCount, pegCount)
}

// UpdateScaleDownNodeRemovalLatency records the time after which node was deleted/needed
// again after being marked unneded
func UpdateScaleDownNodeRemovalLatency(deleted bool, duration time.Duration) {
	DefaultMetrics.UpdateScaleDownNodeRemovalLatency(deleted, duration)
}

// ObserveMaxNodeSkipEvalDurationSeconds records the longest time during which node was skipped during ScaleDown.
// If a node is skipped multiple times consecutively, we store only the earliest timestamp.
func ObserveMaxNodeSkipEvalDurationSeconds(duration time.Duration) {
	DefaultMetrics.ObserveMaxNodeSkipEvalDurationSeconds(duration)
}
