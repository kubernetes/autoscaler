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
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/budgets"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/deletiontracker"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/pdb"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/status"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/utils"
	"sigs.k8s.io/cluster-autoscaler/pkg/metrics"
	"sigs.k8s.io/cluster-autoscaler/pkg/observers/nodegroupchange"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/clustersnapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/clustersnapshot/predicate"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/clustersnapshot/store"
	csisnapshot "sigs.k8s.io/cluster-autoscaler/pkg/simulator/csi/snapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/drainability/rules"
	drasnapshot "sigs.k8s.io/cluster-autoscaler/pkg/simulator/dynamicresources/snapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/options"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/utilization"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/expiring"
	kube_util "sigs.k8s.io/cluster-autoscaler/pkg/utils/kubernetes"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/taints"

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
	GetIgnoreDaemonSetsUtilization(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (bool, error)
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
func (a *Actuator) StartDeletion(ctx context.Context, empty, drain []*apiv1.Node) (status.ScaleDownResult, []*status.ScaleDownNode, errors.AutoscalerError) {
	return a.startDeletion(ctx, empty, drain, false)
}

// StartForceDeletion triggers a new forced deletion process. It will bypass PDBs and forcefully delete the pods and the nodes.
func (a *Actuator) StartForceDeletion(ctx context.Context, empty, drain []*apiv1.Node) (status.ScaleDownResult, []*status.ScaleDownNode, errors.AutoscalerError) {
	return a.startDeletion(ctx, empty, drain, true)
}

// startDeletion contains the shared logic for deleting nodes. It handles both
// normal deletions (respecting PDBs) and forced deletions (bypassing PDBs),
// determined by the 'force' parameter.
func (a *Actuator) startDeletion(ctx context.Context, empty, drain []*apiv1.Node, force bool) (status.ScaleDownResult, []*status.ScaleDownNode, errors.AutoscalerError) {
	a.nodeDeletionScheduler.ResetAndReportMetrics()
	deletionStartTime := time.Now()
	defer func() { metrics.UpdateDuration(ctx, metrics.ScaleDownNodeDeletion, time.Since(deletionStartTime)) }()

	scaledDownNodes := make([]*status.ScaleDownNode, 0)
	emptyToDelete, drainToDelete := a.budgetProcessor.CropNodes(ctx, a.nodeDeletionTracker, empty, drain)
	if len(emptyToDelete) == 0 && len(drainToDelete) == 0 {
		return status.ScaleDownNoNodeDeleted, nil, nil
	}

	if len(emptyToDelete) > 0 {
		// Taint all empty nodes synchronously
		nodeDeleteDelayAfterTaint, err := a.taintNodesSync(ctx, emptyToDelete)
		if err != nil {
			return status.ScaleDownError, scaledDownNodes, err
		}

		emptyScaledDown := a.deleteAsyncEmpty(ctx, emptyToDelete, nodeDeleteDelayAfterTaint, force)
		scaledDownNodes = append(scaledDownNodes, emptyScaledDown...)
	}

	if len(drainToDelete) > 0 {
		// Taint all nodes that need drain synchronously, but don't start any drain/deletion yet. Otherwise, pods evicted from one to-be-deleted node
		// could get recreated on another.
		nodeDeleteDelayAfterTaint, err := a.taintNodesSync(ctx, drainToDelete)
		if err != nil {
			return status.ScaleDownError, scaledDownNodes, err
		}

		// All nodes involved in the scale-down should be tainted now - start draining and deleting nodes asynchronously.
		drainScaledDown := a.deleteAsyncDrain(ctx, drainToDelete, nodeDeleteDelayAfterTaint, force)
		scaledDownNodes = append(scaledDownNodes, drainScaledDown...)
	}

	return status.ScaleDownNodeDeleteStarted, scaledDownNodes, nil
}

// deleteAsyncEmpty immediately starts deletions asynchronously.
// scaledDownNodes return value contains all nodes for which deletion successfully started.
func (a *Actuator) deleteAsyncEmpty(ctx context.Context, NodeGroupViews []*budgets.NodeGroupView, nodeDeleteDelayAfterTaint time.Duration, force bool) (reportedSDNodes []*status.ScaleDownNode) {
	for _, bucket := range NodeGroupViews {
		for _, node := range bucket.Nodes {
			klog.V(0).Infof("Scale-down: removing empty node %q", node.Name)
			a.autoscalingCtx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: removing empty node %q", node.Name)

			if sdNode, err := a.scaleDownNodeToReport(ctx, node, false); err == nil {
				reportedSDNodes = append(reportedSDNodes, sdNode)
			} else {
				klog.Errorf("Scale-down: couldn't report scaled down node, err: %v", err)
			}

			a.nodeDeletionTracker.StartDeletion(bucket.Group.Id(), node.Name)
		}
	}

	for _, bucket := range NodeGroupViews {
		go a.deleteNodesAsync(ctx, bucket.Nodes, bucket.Group, false, force, bucket.BatchSize, nodeDeleteDelayAfterTaint)
	}

	return reportedSDNodes
}

// taintNodesSync synchronously taints all provided nodes with NoSchedule. If tainting fails for any of the nodes, already
// applied taints are cleaned up.
func (a *Actuator) taintNodesSync(ctx context.Context, NodeGroupViews []*budgets.NodeGroupView) (time.Duration, errors.AutoscalerError) {
	nodesToTaint := make([]*apiv1.Node, 0)
	var updateLatencyTracker *UpdateLatencyTracker
	nodeDeleteDelayAfterTaint := a.nodeDeleteDelayAfterTaint
	if a.autoscalingCtx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
		updateLatencyTracker = NewUpdateLatencyTracker(a.autoscalingCtx.AutoscalingKubeClients.ListerRegistry.AllNodeLister())
		go updateLatencyTracker.Start(ctx)
	}

	for _, bucket := range NodeGroupViews {
		for _, node := range bucket.Nodes {
			if a.autoscalingCtx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
				updateLatencyTracker.StartTimeChan <- nodeTaintStartTime{node.Name, time.Now()}
			}
			nodesToTaint = append(nodesToTaint, node)
		}
	}
	failedTaintedNodes := make(chan struct {
		node *apiv1.Node
		err  error
	}, len(nodesToTaint))
	taintedNodes := make(chan *apiv1.Node, len(nodesToTaint))
	workqueue.ParallelizeUntil(context.Background(), maxConcurrentNodesTainting, len(nodesToTaint), func(piece int) {
		node := nodesToTaint[piece]
		err := a.taintNode(ctx, node)
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
	if len(failedTaintedNodes) > 0 {
		for nodeWithError := range failedTaintedNodes {
			a.autoscalingCtx.Recorder.Eventf(nodeWithError.node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", nodeWithError.err)
		}
		// Clean up already applied taints in case of issues.
		for taintedNode := range taintedNodes {
			_, _ = taints.CleanToBeDeleted(ctx, taintedNode, a.autoscalingCtx.ClientSet, a.autoscalingCtx.CordonNodeBeforeTerminate)
		}
		if a.autoscalingCtx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
			close(updateLatencyTracker.AwaitOrStopChan)
		}
		return nodeDeleteDelayAfterTaint, errors.NewAutoscalerErrorf(errors.ApiCallError, "couldn't taint %d nodes with ToBeDeleted", len(failedTaintedNodes))
	}

	if a.autoscalingCtx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
		updateLatencyTracker.AwaitOrStopChan <- true
		latency, ok := <-updateLatencyTracker.ResultChan
		if ok {
			a.pastLatencies.RegisterElement(latency)
			a.pastLatencies.DropNotNewerThan(time.Now().Add(-1 * pastLatencyExpireDuration))
			// CA is expected to wait 3 times the round-trip time between CA and the api-server.
			// At this point, we have already tainted all the nodes.
			// Therefore, the nodeDeleteDelayAfterTaint is set 2 times the maximum latency observed during the last hour.
			nodeDeleteDelayAfterTaint = 2 * maxLatency(a.pastLatencies.ToSlice())
		}
	}
	return nodeDeleteDelayAfterTaint, nil
}

// deleteAsyncDrain asynchronously starts deletions with drain for all provided nodes. scaledDownNodes return value contains all nodes for which
// deletion successfully started.
func (a *Actuator) deleteAsyncDrain(ctx context.Context, NodeGroupViews []*budgets.NodeGroupView, nodeDeleteDelayAfterTaint time.Duration, force bool) (reportedSDNodes []*status.ScaleDownNode) {
	for _, bucket := range NodeGroupViews {
		for _, drainNode := range bucket.Nodes {
			if sdNode, err := a.scaleDownNodeToReport(ctx, drainNode, true); err == nil {
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
		go a.deleteNodesAsync(ctx, bucket.Nodes, bucket.Group, true, force, bucket.BatchSize, nodeDeleteDelayAfterTaint)
	}

	return reportedSDNodes
}

func (a *Actuator) deleteNodesAsync(ctx context.Context, nodes []*apiv1.Node, nodeGroup cloudprovider.NodeGroup, drain bool, force bool, batchSize int, nodeDeleteDelayAfterTaint time.Duration) {
	var remainingPdbTracker pdb.RemainingPdbTracker
	var registry kube_util.ListerRegistry

	if len(nodes) == 0 {
		return
	}

	if nodeDeleteDelayAfterTaint > time.Duration(0) {
		klog.V(0).Infof("Scale-down: waiting %v before trying to delete nodes", nodeDeleteDelayAfterTaint)
		time.Sleep(nodeDeleteDelayAfterTaint)
	}

	clusterSnapshot, err := a.createSnapshot(ctx, nodes)
	if err != nil {
		nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "createSnapshot returned error %v", err)}
		for _, node := range nodes {
			a.nodeDeletionScheduler.AbortNodeDeletionDueToError(ctx, node, nodeGroup.Id(), drain, "failed to create delete snapshot", nodeDeleteResult)
		}
		return
	}

	if drain {
		pdbs, err := a.autoscalingCtx.PodDisruptionBudgetLister().List()
		if err != nil {
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "podDisruptionBudgetLister.List returned error %v", err)}
			for _, node := range nodes {
				a.nodeDeletionScheduler.AbortNodeDeletionDueToError(ctx, node, nodeGroup.Id(), drain, "failed to fetch pod disruption budgets", nodeDeleteResult)
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
			a.nodeDeletionScheduler.AbortNodeDeletionDueToError(ctx, node, nodeGroup.Id(), drain, "failed to get node info", nodeDeleteResult)
			continue
		}

		podMoveInfo, err := simulator.GetPodsToMove(ctx, nodeInfo, a.deleteOptions, a.drainabilityRules, registry, remainingPdbTracker, time.Now())
		if err != nil {
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "GetPodsToMove for %q returned error: %v", node.Name, err)}
			a.nodeDeletionScheduler.AbortNodeDeletion(ctx, node, nodeGroup.Id(), drain, "failed to get pods to move on node", nodeDeleteResult, true)
			continue
		}

		if !drain {
			if len(podMoveInfo.Pods) != 0 {
				nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "failed to delete empty node %q, new pods scheduled", node.Name)}
				a.nodeDeletionScheduler.AbortNodeDeletion(ctx, node, nodeGroup.Id(), drain, "node is not empty", nodeDeleteResult, true)
				continue
			}
			if len(podMoveInfo.OnCompletionPods) != 0 {
				nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "failed to delete empty node %q, active on-completion pods present", node.Name)}
				a.nodeDeletionScheduler.AbortNodeDeletion(ctx, node, nodeGroup.Id(), drain, "active on-completion pods present", nodeDeleteResult, true)
				continue
			}
		}

		if force {
			go a.nodeDeletionScheduler.scheduleForceDeletion(ctx, nodeInfo, nodeGroup, batchSize, drain)
			continue
		}

		go a.nodeDeletionScheduler.ScheduleDeletion(ctx, nodeInfo, nodeGroup, batchSize, drain)
	}
}

func (a *Actuator) scaleDownNodeToReport(ctx context.Context, node *apiv1.Node, drain bool) (*status.ScaleDownNode, error) {
	nodeGroup, err := a.autoscalingCtx.CloudProvider.NodeGroupForNode(ctx, node)
	if err != nil {
		return nil, err
	}
	if nodeGroup == nil {
		return nil, errors.NewAutoscalerErrorf(errors.NodeGroupDoesNotExistError, "no node group for node %s", node.Name)
	}
	nodeInfo, err := a.autoscalingCtx.ClusterSnapshot.GetNodeInfo(node.Name)
	if err != nil {
		return nil, err
	}

	ignoreDaemonSetsUtilization, err := a.configGetter.GetIgnoreDaemonSetsUtilization(ctx, nodeGroup)
	if err != nil {
		return nil, err
	}

	gpuConfig := a.autoscalingCtx.CloudProvider.GetNodeGpuConfig(ctx, node)
	utilInfo, err := utilization.Calculate(ctx, nodeInfo, ignoreDaemonSetsUtilization, a.autoscalingCtx.IgnoreMirrorPodsUtilization, a.autoscalingCtx.DynamicResourceAllocationEnabled, gpuConfig, time.Now())
	if err != nil {
		return nil, err
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
func (a *Actuator) taintNode(ctx context.Context, node *apiv1.Node) error {
	if _, err := taints.MarkToBeDeleted(ctx, node, a.autoscalingCtx.ClientSet, a.autoscalingCtx.CordonNodeBeforeTerminate); err != nil {
		a.autoscalingCtx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
		return errors.ToAutoscalerError(errors.ApiCallError, err)
	}
	a.autoscalingCtx.Recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "marked the node as toBeDeleted/unschedulable")
	return nil
}

func (a *Actuator) createSnapshot(ctx context.Context, nodes []*apiv1.Node) (clustersnapshot.ClusterSnapshot, error) {
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

	err = snapshot.SetClusterState(ctx, nodes, nonExpendableScheduledPods, draSnapshot, csiSnapshot)
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
