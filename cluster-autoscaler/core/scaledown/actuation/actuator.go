/*
Copyright 2022 The Kubernetes Authors.

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

package actuation

import (
	"context"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/budgets"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/observers/nodegroupchange"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/predicate"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/expiring"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	pastLatencyExpireDuration  = time.Hour
	maxConcurrentNodesTainting = 5
)

// Actuator is responsible for draining and deleting nodes.
type Actuator struct {
	autoscalingCtx        *ca_context.AutoscalingContext
	nodeDeletionTracker   *deletiontracker.NodeDeletionTracker
	nodeDeletionScheduler *GroupDeletionScheduler
	deleteOptions         options.NodeDeleteOptions
	drainabilityRules     rules.Rules
	// TODO: Move budget processor to scaledown planner, potentially merge into PostFilteringScaleDownNodeProcessor
	// This is a larger change to the code structure which impacts some existing actuator unit tests
	// as well as Cluster Autoscaler implementations that may override ScaleDownSetProcessor
	budgetProcessor           *budgets.ScaleDownBudgetProcessor
	configGetter              actuatorNodeGroupConfigGetter
	nodeDeleteDelayAfterTaint time.Duration
	pastLatencies             *expiring.List
}

// actuatorNodeGroupConfigGetter is an interface to limit the functions that can be used
// from NodeGroupConfigProcessor interface
type actuatorNodeGroupConfigGetter interface {
	// GetIgnoreDaemonSetsUtilization returns IgnoreDaemonSetsUtilization value that should be used for a given NodeGroup.
	GetIgnoreDaemonSetsUtilization(nodeGroup cloudprovider.NodeGroup) (bool, error)
}

// NewActuator returns a new instance of Actuator.
func NewActuator(autoscalingCtx *ca_context.AutoscalingContext, scaleStateNotifier nodegroupchange.NodeGroupChangeObserver, ndt *deletiontracker.NodeDeletionTracker, deleteOptions options.NodeDeleteOptions, drainabilityRules rules.Rules, configGetter actuatorNodeGroupConfigGetter) *Actuator {
	ndb := NewNodeDeletionBatcher(autoscalingCtx, scaleStateNotifier, ndt, autoscalingCtx.NodeDeletionBatcherInterval)
	legacyFlagDrainConfig := SingleRuleDrainConfig(autoscalingCtx.MaxGracefulTerminationSec)
	var evictor Evictor
	if len(autoscalingCtx.DrainPriorityConfig) > 0 {
		evictor = NewEvictor(ndt, autoscalingCtx.DrainPriorityConfig, true)
	} else {
		evictor = NewEvictor(ndt, legacyFlagDrainConfig, false)
	}
	return &Actuator{
		autoscalingCtx:            autoscalingCtx,
		nodeDeletionTracker:       ndt,
		nodeDeletionScheduler:     NewGroupDeletionScheduler(autoscalingCtx, ndt, ndb, evictor),
		budgetProcessor:           budgets.NewScaleDownBudgetProcessor(autoscalingCtx),
		deleteOptions:             deleteOptions,
		drainabilityRules:         drainabilityRules,
		configGetter:              configGetter,
		nodeDeleteDelayAfterTaint: autoscalingCtx.NodeDeleteDelayAfterTaint,
		pastLatencies:             expiring.NewList(),
	}
}

// CheckStatus should returns an immutable snapshot of ongoing deletions.
func (a *Actuator) CheckStatus() scaledown.ActuationStatus {
	return a.nodeDeletionTracker.Snapshot()
}

// ClearResultsNotNewerThan removes information about deletions finished before or exactly at the provided timestamp.
func (a *Actuator) ClearResultsNotNewerThan(t time.Time) {
	a.nodeDeletionTracker.ClearResultsNotNewerThan(t)
}

// DeletionResults returns deletion results since the last ClearResultsNotNewerThan call
// in a map form, along with the timestamp of last result.
func (a *Actuator) DeletionResults() (map[string]status.NodeDeleteResult, time.Time) {
	return a.nodeDeletionTracker.DeletionResults()
}

// StartDeletion triggers a new deletion process.
func (a *Actuator) StartDeletion(empty, drain []*apiv1.Node) (status.ScaleDownResult, []*status.ScaleDownNode, errors.AutoscalerError) {
	return a.startDeletion(empty, drain, false)
}

// StartForceDeletion triggers a new forced deletion process. It will bypass PDBs and forcefully delete the pods and the nodes.
func (a *Actuator) StartForceDeletion(empty, drain []*apiv1.Node) (status.ScaleDownResult, []*status.ScaleDownNode, errors.AutoscalerError) {
	return a.startDeletion(empty, drain, true)
}

// startDeletion contains the shared logic for deleting nodes. It handles both
// normal deletions (respecting PDBs) and forced deletions (bypassing PDBs),
// determined by the 'force' parameter.
func (a *Actuator) startDeletion(empty, drain []*apiv1.Node, force bool) (status.ScaleDownResult, []*status.ScaleDownNode, errors.AutoscalerError) {
	a.nodeDeletionScheduler.ResetAndReportMetrics()
	deletionStartTime := time.Now()
	defer func() { metrics.UpdateDuration(metrics.ScaleDownNodeDeletion, time.Since(deletionStartTime)) }()

	var nodesToClean []*apiv1.Node
	defer func() {
		if len(nodesToClean) > 0 {
			a.cleanNodesToBeDeleted(nodesToClean)
		}
	}()

	scaledDownNodes := make([]*status.ScaleDownNode, 0)
	emptyToDelete, drainToDelete := a.budgetProcessor.CropNodes(a.nodeDeletionTracker, empty, drain)
	if len(emptyToDelete) == 0 && len(drainToDelete) == 0 {
		return status.ScaleDownNoNodeDeleted, nil, nil
	}

	if len(emptyToDelete) > 0 {
		// Taint all empty nodes synchronously
		taintRes, err := a.taintNodesSync(emptyToDelete)
		if len(taintRes.nodesToClean) > 0 {
			nodesToClean = append(nodesToClean, taintRes.nodesToClean...)
		}
		if err != nil {
			return status.ScaleDownError, scaledDownNodes, err
		}

		emptyScaledDown := a.deleteAsyncEmpty(taintRes.successfulNodes, taintRes.delayAfterTaint, force)
		scaledDownNodes = append(scaledDownNodes, emptyScaledDown...)
	}

	if len(drainToDelete) > 0 {
		// Taint all nodes that need drain synchronously, but don't start any drain/deletion yet. Otherwise, pods evicted from one to-be-deleted node
		// could get recreated on another.
		taintRes, err := a.taintNodesSync(drainToDelete)
		if len(taintRes.nodesToClean) > 0 {
			nodesToClean = append(nodesToClean, taintRes.nodesToClean...)
		}
		if err != nil {
			return status.ScaleDownError, scaledDownNodes, err
		}

		// All nodes involved in the scale-down should be tainted now - start draining and deleting nodes asynchronously.
		drainScaledDown := a.deleteAsyncDrain(taintRes.successfulNodes, taintRes.delayAfterTaint, force)
		scaledDownNodes = append(scaledDownNodes, drainScaledDown...)
	}

	return status.ScaleDownNodeDeleteStarted, scaledDownNodes, nil
}

// deleteAsyncEmpty immediately starts deletions asynchronously.
// scaledDownNodes return value contains all nodes for which deletion successfully started.
func (a *Actuator) deleteAsyncEmpty(NodeGroupViews []*budgets.NodeGroupView, nodeDeleteDelayAfterTaint time.Duration, force bool) (reportedSDNodes []*status.ScaleDownNode) {
	for _, bucket := range NodeGroupViews {
		for _, node := range bucket.Nodes {
			klog.V(0).Infof("Scale-down: removing empty node %q", node.Name)
			a.autoscalingCtx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: removing empty node %q", node.Name)

			if sdNode, err := a.scaleDownNodeToReport(node, false); err == nil {
				reportedSDNodes = append(reportedSDNodes, sdNode)
			} else {
				klog.Errorf("Scale-down: couldn't report scaled down node, err: %v", err)
			}

			a.nodeDeletionTracker.StartDeletion(bucket.Group.Id(), node.Name)
		}
	}

	for _, bucket := range NodeGroupViews {
		go a.deleteNodesAsync(bucket.Nodes, bucket.Group, false, force, bucket.BatchSize, nodeDeleteDelayAfterTaint)
	}

	return reportedSDNodes
}

type taintNodesResult struct {
	delayAfterTaint time.Duration
	successfulNodes []*budgets.NodeGroupView
	nodesToClean    []*apiv1.Node
}

// taintNodesSync synchronously taints all provided nodes with NoSchedule. If tainting fails for any of the nodes, already
// applied taints are cleaned up. It returns a taintNodesResult containing the actual node delay, the successfully tainted NodeGroupViews, and a list of nodes to clean up, as well as any fatal error.
func (a *Actuator) taintNodesSync(NodeGroupViews []*budgets.NodeGroupView) (taintNodesResult, errors.AutoscalerError) {
	nodesToTaint := make([]*apiv1.Node, 0)
	var updateLatencyTracker *UpdateLatencyTracker
	nodeDeleteDelayAfterTaint := a.nodeDeleteDelayAfterTaint
	if a.autoscalingCtx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
		updateLatencyTracker = NewUpdateLatencyTracker(a.autoscalingCtx.AutoscalingKubeClients.ListerRegistry.AllNodeLister())
		go updateLatencyTracker.Start()
	}

	for _, bucket := range NodeGroupViews {
		for _, node := range bucket.Nodes {
			if a.autoscalingCtx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
				// The start timestamps pushed to the tracker prior to this block are captured before client-go token bucket rate-limiting (QPS/burst).
				// Therefore, this metric inherently measures: [client-side throttle queue time + network round trip + API server propagation time].
				updateLatencyTracker.StartTimeChan <- nodeTaintStartTime{node.Name, time.Now()}
			}
			nodesToTaint = append(nodesToTaint, node)
		}
	}

	var successfulNodeGroupViews []*budgets.NodeGroupView
	var globalNodesToClean []*apiv1.Node
	var retErr errors.AutoscalerError

	if a.autoscalingCtx.AutoscalingOptions.PartialTaintActuationEnabled {
		failedNodes := a.applyTaintsConcurrently(nodesToTaint)

		if a.autoscalingCtx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
			// ExpectedCount relies entirely on the network response. If a node taint call throws a network timeout but implicitly succeeds
			// on the server, expectedCount drops but a start timestamp was already recorded. The tracker loop may hit its expected count early
			// and underestimate max latency. This is a known, accepted compromise.
			expectedCount := len(nodesToTaint) - len(failedNodes)
			if expectedCount == 0 {
				close(updateLatencyTracker.ExpectedNodeCountChan)
			} else {
				updateLatencyTracker.ExpectedNodeCountChan <- expectedCount
				latency, ok := <-updateLatencyTracker.ResultChan
				if ok {
					a.pastLatencies.RegisterElement(latency)
					a.pastLatencies.DropNotNewerThan(time.Now().Add(-1 * pastLatencyExpireDuration))
					nodeDeleteDelayAfterTaint = 2 * maxLatency(a.pastLatencies.ToSlice())
				}
			}
		}

		if len(failedNodes) > 0 {
			klog.Warningf("Couldn't taint %d nodes with ToBeDeleted, proceeding with partial scale down or bucket cleanup", len(failedNodes))
			for _, node := range nodesToTaint {
				if err, ok := failedNodes[node.Name]; ok {
					a.autoscalingCtx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
				}
			}
			successfulNodeGroupViews, globalNodesToClean = a.resolveTaintFailures(NodeGroupViews, failedNodes)
			if len(successfulNodeGroupViews) == 0 {
				retErr = errors.NewAutoscalerErrorf(errors.ApiCallError, "couldn't taint %d nodes with ToBeDeleted and no nodes can be scaled down", len(failedNodes))
			}
		} else {
			successfulNodeGroupViews = NodeGroupViews
		}
	} else {
		failedTaintedNodes := make(chan struct {
			node *apiv1.Node
			err  error
		}, len(nodesToTaint))
		taintedNodes := make(chan *apiv1.Node, len(nodesToTaint))
		workqueue.ParallelizeUntil(context.Background(), maxConcurrentNodesTainting, len(nodesToTaint), func(piece int) {
			node := nodesToTaint[piece]
			err := a.taintNode(node)
			if err != nil {
				failedTaintedNodes <- struct {
					node *apiv1.Node
					err  error
				}{node: node, err: err}
			} else {
				taintedNodes <- node
			}
		})
		close(failedTaintedNodes)
		close(taintedNodes)
		failedCount := len(failedTaintedNodes)
		if failedCount > 0 {
			for nodeWithError := range failedTaintedNodes {
				a.autoscalingCtx.Recorder.Eventf(nodeWithError.node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", nodeWithError.err)
			}
			// Clean up already applied taints in case of issues.
			for taintedNode := range taintedNodes {
				globalNodesToClean = append(globalNodesToClean, taintedNode)
			}
			retErr = errors.NewAutoscalerErrorf(errors.ApiCallError, "couldn't taint %d nodes with ToBeDeleted", failedCount)
		} else {
			successfulNodeGroupViews = NodeGroupViews
		}

		if a.autoscalingCtx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
			if len(successfulNodeGroupViews) == 0 {
				close(updateLatencyTracker.ExpectedNodeCountChan)
			} else {
				expectedCount := 0
				for _, bucket := range successfulNodeGroupViews {
					expectedCount += len(bucket.Nodes)
				}
				updateLatencyTracker.ExpectedNodeCountChan <- expectedCount
				latency, ok := <-updateLatencyTracker.ResultChan
				if ok {
					a.pastLatencies.RegisterElement(latency)
					a.pastLatencies.DropNotNewerThan(time.Now().Add(-1 * pastLatencyExpireDuration))
					nodeDeleteDelayAfterTaint = 2 * maxLatency(a.pastLatencies.ToSlice())
				}
			}
		}
	}

	return taintNodesResult{
		delayAfterTaint: nodeDeleteDelayAfterTaint,
		successfulNodes: successfulNodeGroupViews,
		nodesToClean:    globalNodesToClean,
	}, retErr
}

// applyTaintsConcurrently attempts to add the ToBeDeleted taint to all nodes in the batch.
// It executes via workqueue.ParallelizeUntil to prevent single-node networking latency from bottlenecking the loop.
// Returns a map of node names to errors for nodes that failed the network taint call.
func (a *Actuator) applyTaintsConcurrently(nodesToTaint []*apiv1.Node) map[string]error {
	type taintResult struct {
		node *apiv1.Node
		err  error
	}
	results := make(chan taintResult, len(nodesToTaint))

	workqueue.ParallelizeUntil(context.Background(), maxConcurrentNodesTainting, len(nodesToTaint), func(piece int) {
		node := nodesToTaint[piece]
		err := a.taintNode(node)
		results <- taintResult{node: node, err: err}
	})
	close(results)

	failedNodes := make(map[string]error)
	for res := range results {
		if res.err != nil {
			failedNodes[res.node.Name] = res.err
		}
	}
	return failedNodes
}

// resolveTaintFailures evaluates how to handle taint failures in a batch of nodes.
//
// When actuator fails to apply a taint to a node:
//   - If the node is from an atomic nodegroup, actuator must clean up all successfully applied taints in this nodegroup. The group won't be scaled down.
//   - If the node is from a regular, non-atomic nodegorup, other successfully tainted nodes from this nodegroup can be deleted.
//
// It returns nodes (grouped by nodegroup) that can proceed to deletion, and a list of nodes that require taint cleanup.
func (a *Actuator) resolveTaintFailures(nodeGroupViews []*budgets.NodeGroupView, failedNodes map[string]error) ([]*budgets.NodeGroupView, []*apiv1.Node) {
	var successfulNodeGroupViews []*budgets.NodeGroupView
	var globalNodesToClean []*apiv1.Node

	for _, bucket := range nodeGroupViews {
		opts, err := bucket.Group.GetOptions(a.autoscalingCtx.NodeGroupDefaults)
		isAtomic := false
		if err != nil {
			klog.Warningf("Failed to get options for node group %v: %v, assuming atomic node to be safe", bucket.Group.Id(), err)
			isAtomic = true
		} else if opts != nil && opts.ZeroOrMaxNodeScaling {
			isAtomic = true
		}

		var successfulBucket *budgets.NodeGroupView
		var toClean []*apiv1.Node

		if isAtomic {
			successfulBucket, toClean = resolveAtomicBucketFailures(bucket, failedNodes)
		} else {
			successfulBucket, toClean = resolveNonAtomicBucketFailures(bucket, failedNodes)
		}

		if successfulBucket != nil {
			successfulNodeGroupViews = append(successfulNodeGroupViews, successfulBucket)
		}
		if len(toClean) > 0 {
			globalNodesToClean = append(globalNodesToClean, toClean...)
		}
	}
	return successfulNodeGroupViews, globalNodesToClean
}

// cleanNodesToBeDeleted executes the rollback of ToBeDeleted taints concurrently across all nodes provided.
// This is typically deferred to run at the absolute end of the StartDeletion loop so as not to stall the healthy path.
func (a *Actuator) cleanNodesToBeDeleted(nodes []*apiv1.Node) {
	if a.autoscalingCtx.AutoscalingOptions.PartialTaintActuationEnabled {
		workqueue.ParallelizeUntil(context.Background(), maxConcurrentNodesTainting, len(nodes), func(piece int) {
			_, _ = taints.CleanToBeDeleted(nodes[piece], a.autoscalingCtx.ClientSet, a.autoscalingCtx.CordonNodeBeforeTerminate)
		})
	} else {
		for _, node := range nodes {
			_, _ = taints.CleanToBeDeleted(node, a.autoscalingCtx.ClientSet, a.autoscalingCtx.CordonNodeBeforeTerminate)
		}
	}
}

// deleteAsyncDrain asynchronously starts deletions with drain for all provided nodes. scaledDownNodes return value contains all nodes for which
// deletion successfully started.
func (a *Actuator) deleteAsyncDrain(NodeGroupViews []*budgets.NodeGroupView, nodeDeleteDelayAfterTaint time.Duration, force bool) (reportedSDNodes []*status.ScaleDownNode) {
	for _, bucket := range NodeGroupViews {
		for _, drainNode := range bucket.Nodes {
			if sdNode, err := a.scaleDownNodeToReport(drainNode, true); err == nil {
				klog.V(0).Infof("Scale-down: removing node %s, utilization: %v, pods to reschedule: %s", drainNode.Name, sdNode.UtilInfo, joinPodNames(sdNode.EvictedPods))
				a.autoscalingCtx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDown", "Scale-down: removing node %s, utilization: %v, pods to reschedule: %s", drainNode.Name, sdNode.UtilInfo, joinPodNames(sdNode.EvictedPods))
				reportedSDNodes = append(reportedSDNodes, sdNode)
			} else {
				klog.Errorf("Scale-down: couldn't report scaled down node, err: %v", err)
			}

			a.nodeDeletionTracker.StartDeletionWithDrain(bucket.Group.Id(), drainNode.Name)
		}
	}

	for _, bucket := range NodeGroupViews {
		go a.deleteNodesAsync(bucket.Nodes, bucket.Group, true, force, bucket.BatchSize, nodeDeleteDelayAfterTaint)
	}

	return reportedSDNodes
}

func (a *Actuator) deleteNodesAsync(nodes []*apiv1.Node, nodeGroup cloudprovider.NodeGroup, drain bool, force bool, batchSize int, nodeDeleteDelayAfterTaint time.Duration) {
	var remainingPdbTracker pdb.RemainingPdbTracker
	var registry kube_util.ListerRegistry

	if len(nodes) == 0 {
		return
	}

	if nodeDeleteDelayAfterTaint > time.Duration(0) {
		klog.V(0).Infof("Scale-down: waiting %v before trying to delete nodes", nodeDeleteDelayAfterTaint)
		time.Sleep(nodeDeleteDelayAfterTaint)
	}

	clusterSnapshot, err := a.createSnapshot(nodes)
	if err != nil {
		nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "createSnapshot returned error %v", err)}
		for _, node := range nodes {
			a.nodeDeletionScheduler.AbortNodeDeletionDueToError(node, nodeGroup.Id(), drain, "failed to create delete snapshot", nodeDeleteResult)
		}
		return
	}

	if drain {
		pdbs, err := a.autoscalingCtx.PodDisruptionBudgetLister().List()
		if err != nil {
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "podDisruptionBudgetLister.List returned error %v", err)}
			for _, node := range nodes {
				a.nodeDeletionScheduler.AbortNodeDeletionDueToError(node, nodeGroup.Id(), drain, "failed to fetch pod disruption budgets", nodeDeleteResult)
			}
			return
		}
		remainingPdbTracker = pdb.NewBasicRemainingPdbTracker()
		remainingPdbTracker.SetPdbs(pdbs)
		registry = a.autoscalingCtx.ListerRegistry
	}

	if batchSize == 0 {
		batchSize = len(nodes)
	}

	for _, node := range nodes {
		nodeInfo, err := clusterSnapshot.GetNodeInfo(node.Name)
		if err != nil {
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "nodeInfos.Get for %q returned error: %v", node.Name, err)}
			a.nodeDeletionScheduler.AbortNodeDeletionDueToError(node, nodeGroup.Id(), drain, "failed to get node info", nodeDeleteResult)
			continue
		}

		podMoveInfo, err := simulator.GetPodsToMove(nodeInfo, a.deleteOptions, a.drainabilityRules, registry, remainingPdbTracker, time.Now())
		if err != nil {
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "GetPodsToMove for %q returned error: %v", node.Name, err)}
			a.nodeDeletionScheduler.AbortNodeDeletion(node, nodeGroup.Id(), drain, "failed to get pods to move on node", nodeDeleteResult, true)
			continue
		}

		if !drain {
			if len(podMoveInfo.Pods) != 0 {
				nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "failed to delete empty node %q, new pods scheduled", node.Name)}
				a.nodeDeletionScheduler.AbortNodeDeletion(node, nodeGroup.Id(), drain, "node is not empty", nodeDeleteResult, true)
				continue
			}
			if len(podMoveInfo.OnCompletionPods) != 0 {
				nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "failed to delete empty node %q, active on-completion pods present", node.Name)}
				a.nodeDeletionScheduler.AbortNodeDeletion(node, nodeGroup.Id(), drain, "active on-completion pods present", nodeDeleteResult, true)
				continue
			}
		}

		if force {
			go a.nodeDeletionScheduler.scheduleForceDeletion(nodeInfo, nodeGroup, batchSize, drain)
			continue
		}

		go a.nodeDeletionScheduler.ScheduleDeletion(nodeInfo, nodeGroup, batchSize, drain)
	}
}

func (a *Actuator) scaleDownNodeToReport(node *apiv1.Node, drain bool) (*status.ScaleDownNode, error) {
	nodeGroup, err := a.autoscalingCtx.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		return nil, errors.NewAutoscalerErrorf(errors.InternalError, "failed to get node group for node %s: %v", node.Name, err)
	}
	if nodeGroup == nil {
		return nil, errors.NewAutoscalerErrorf(errors.NodeGroupDoesNotExistError, "no node group for node %s", node.Name)
	}
	nodeInfo, err := a.autoscalingCtx.ClusterSnapshot.GetNodeInfo(node.Name)
	if err != nil {
		return nil, errors.NewAutoscalerErrorf(errors.InternalError, "failed to get node info for %s: %v", node.Name, err)
	}

	ignoreDaemonSetsUtilization, err := a.configGetter.GetIgnoreDaemonSetsUtilization(nodeGroup)
	if err != nil {
		return nil, errors.NewAutoscalerErrorf(errors.InternalError, "failed to get ignoreDaemonSetsUtilization for node group %s: %v", nodeGroup.Id(), err)
	}

	gpuConfig := a.autoscalingCtx.CloudProvider.GetNodeGpuConfig(node)
	utilInfo, err := utilization.Calculate(nodeInfo, ignoreDaemonSetsUtilization, a.autoscalingCtx.IgnoreMirrorPodsUtilization, a.autoscalingCtx.DynamicResourceAllocationEnabled, gpuConfig, time.Now())
	if err != nil {
		return nil, errors.NewAutoscalerErrorf(errors.InternalError, "failed to calculate utilization for %s: %v", node.Name, err)
	}
	var evictedPods []*apiv1.Pod
	if drain {
		_, nonDsPodsToEvict := podsToEvict(nodeInfo, a.autoscalingCtx.DaemonSetEvictionForOccupiedNodes)
		evictedPods = nonDsPodsToEvict
	}
	return &status.ScaleDownNode{
		Node:        node,
		NodeGroup:   nodeGroup,
		EvictedPods: evictedPods,
		UtilInfo:    utilInfo,
	}, nil
}

// taintNode taints the node with NoSchedule to prevent new pods scheduling on it.
func (a *Actuator) taintNode(node *apiv1.Node) error {
	if _, err := taints.MarkToBeDeleted(node, a.autoscalingCtx.ClientSet, a.autoscalingCtx.CordonNodeBeforeTerminate); err != nil {
		a.autoscalingCtx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
		return errors.ToAutoscalerError(errors.ApiCallError, err)
	}
	a.autoscalingCtx.Recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "marked the node as toBeDeleted/unschedulable")
	return nil
}

func (a *Actuator) createSnapshot(nodes []*apiv1.Node) (clustersnapshot.ClusterSnapshot, error) {
	snapshot := predicate.NewPredicateSnapshot(store.NewBasicSnapshotStore(), a.autoscalingCtx.FrameworkHandle, a.autoscalingCtx.DynamicResourceAllocationEnabled, a.autoscalingCtx.PredicateParallelism, a.autoscalingCtx.CSINodeAwareSchedulingEnabled, a.autoscalingCtx.SchedulerVerbosityOffset)
	pods, err := a.autoscalingCtx.AllPodLister().List()
	if err != nil {
		return nil, err
	}

	scheduledPods := kube_util.ScheduledPods(pods)
	nonExpendableScheduledPods := utils.FilterOutExpendablePods(scheduledPods, a.autoscalingCtx.ExpendablePodsPriorityCutoff)

	var draSnapshot *drasnapshot.Snapshot
	if a.autoscalingCtx.DynamicResourceAllocationEnabled && a.autoscalingCtx.DraProvider != nil {
		draSnapshot, err = a.autoscalingCtx.DraProvider.Snapshot()
		if err != nil {
			return nil, err
		}
	}

	var csiSnapshot *csisnapshot.Snapshot
	if a.autoscalingCtx.CSINodeAwareSchedulingEnabled {
		csiSnapshot, err = a.autoscalingCtx.CsiProvider.Snapshot()
		if err != nil {
			return nil, err
		}
	}

	err = snapshot.SetClusterState(nodes, nonExpendableScheduledPods, draSnapshot, csiSnapshot)
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func joinPodNames(pods []*apiv1.Pod) string {
	var names []string
	for _, pod := range pods {
		names = append(names, pod.Name)
	}
	return strings.Join(names, ",")
}

// resolveAtomicBucketFailures handles partial taint failures for zero-or-max atomic NodeGroups (like TPU slices).
// An atomic NodeGroup must scale entirely or not at all; if even one node fails tainting, the entire bucket is rolled back.
func resolveAtomicBucketFailures(bucket *budgets.NodeGroupView, failedNodes map[string]error) (*budgets.NodeGroupView, []*apiv1.Node) {
	bucketFailed := false
	for _, node := range bucket.Nodes {
		if failedNodes[node.Name] != nil {
			bucketFailed = true
			break
		}
	}
	if bucketFailed {
		var toClean []*apiv1.Node
		for _, node := range bucket.Nodes {
			if failedNodes[node.Name] == nil {
				toClean = append(toClean, node)
			}
		}
		return nil, toClean
	}
	return bucket, nil
}

// resolveNonAtomicBucketFailures handles partial taint failures for standard, non-atomic NodeGroups.
// It safely separates the batch, allowing successfully tainted nodes to proceed to immediate deletion while returning failed ones for cleanup.
func resolveNonAtomicBucketFailures(bucket *budgets.NodeGroupView, failedNodes map[string]error) (*budgets.NodeGroupView, []*apiv1.Node) {
	successfulNodes := make([]*apiv1.Node, 0, len(bucket.Nodes))

	for _, node := range bucket.Nodes {
		if failedNodes[node.Name] == nil {
			successfulNodes = append(successfulNodes, node)
		}
	}

	if len(successfulNodes) == len(bucket.Nodes) {
		return bucket, nil
	}

	if len(successfulNodes) > 0 {
		newBucket := *bucket
		newBucket.Nodes = successfulNodes
		return &newBucket, nil
	}

	return nil, nil
}
