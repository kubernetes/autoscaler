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
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/budgets"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/dynamicresources"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/observers/nodegroupchange"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/expiring"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/klog/v2"
)

const (
	pastLatencyExpireDuration = time.Hour
)

// Actuator is responsible for draining and deleting nodes.
type Actuator struct {
	ctx                   *context.AutoscalingContext
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
	draProvider               *dynamicresources.Provider
}

// actuatorNodeGroupConfigGetter is an interface to limit the functions that can be used
// from NodeGroupConfigProcessor interface
type actuatorNodeGroupConfigGetter interface {
	// GetIgnoreDaemonSetsUtilization returns IgnoreDaemonSetsUtilization value that should be used for a given NodeGroup.
	GetIgnoreDaemonSetsUtilization(nodeGroup cloudprovider.NodeGroup) (bool, error)
}

// NewActuator returns a new instance of Actuator.
func NewActuator(ctx *context.AutoscalingContext, scaleStateNotifier nodegroupchange.NodeGroupChangeObserver, ndt *deletiontracker.NodeDeletionTracker, deleteOptions options.NodeDeleteOptions, drainabilityRules rules.Rules, configGetter actuatorNodeGroupConfigGetter, draProvider *dynamicresources.Provider) *Actuator {
	ndb := NewNodeDeletionBatcher(ctx, scaleStateNotifier, ndt, ctx.NodeDeletionBatcherInterval)
	legacyFlagDrainConfig := SingleRuleDrainConfig(ctx.MaxGracefulTerminationSec)
	var evictor Evictor
	if len(ctx.DrainPriorityConfig) > 0 {
		evictor = NewEvictor(ndt, ctx.DrainPriorityConfig, true)
	} else {
		evictor = NewEvictor(ndt, legacyFlagDrainConfig, false)
	}
	return &Actuator{
		ctx:                       ctx,
		nodeDeletionTracker:       ndt,
		nodeDeletionScheduler:     NewGroupDeletionScheduler(ctx, ndt, ndb, evictor),
		budgetProcessor:           budgets.NewScaleDownBudgetProcessor(ctx),
		deleteOptions:             deleteOptions,
		drainabilityRules:         drainabilityRules,
		configGetter:              configGetter,
		nodeDeleteDelayAfterTaint: ctx.NodeDeleteDelayAfterTaint,
		pastLatencies:             expiring.NewList(),
		draProvider:               draProvider,
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
	a.nodeDeletionScheduler.ResetAndReportMetrics()
	deletionStartTime := time.Now()
	defer func() { metrics.UpdateDuration(metrics.ScaleDownNodeDeletion, time.Since(deletionStartTime)) }()

	scaledDownNodes := make([]*status.ScaleDownNode, 0)
	emptyToDelete, drainToDelete := a.budgetProcessor.CropNodes(a.nodeDeletionTracker, empty, drain)
	if len(emptyToDelete) == 0 && len(drainToDelete) == 0 {
		return status.ScaleDownNoNodeDeleted, nil, nil
	}

	if len(emptyToDelete) > 0 {
		// Taint all empty nodes synchronously
		nodeDeleteDelayAfterTaint, err := a.taintNodesSync(emptyToDelete)
		if err != nil {
			return status.ScaleDownError, scaledDownNodes, err
		}

		emptyScaledDown := a.deleteAsyncEmpty(emptyToDelete, nodeDeleteDelayAfterTaint)
		scaledDownNodes = append(scaledDownNodes, emptyScaledDown...)
	}

	if len(drainToDelete) > 0 {
		// Taint all nodes that need drain synchronously, but don't start any drain/deletion yet. Otherwise, pods evicted from one to-be-deleted node
		// could get recreated on another.
		nodeDeleteDelayAfterTaint, err := a.taintNodesSync(drainToDelete)
		if err != nil {
			return status.ScaleDownError, scaledDownNodes, err
		}

		// All nodes involved in the scale-down should be tainted now - start draining and deleting nodes asynchronously.
		drainScaledDown := a.deleteAsyncDrain(drainToDelete, nodeDeleteDelayAfterTaint)
		scaledDownNodes = append(scaledDownNodes, drainScaledDown...)
	}

	return status.ScaleDownNodeDeleteStarted, scaledDownNodes, nil
}

// deleteAsyncEmpty immediately starts deletions asynchronously.
// scaledDownNodes return value contains all nodes for which deletion successfully started.
func (a *Actuator) deleteAsyncEmpty(NodeGroupViews []*budgets.NodeGroupView, nodeDeleteDelayAfterTaint time.Duration) (reportedSDNodes []*status.ScaleDownNode) {
	for _, bucket := range NodeGroupViews {
		for _, node := range bucket.Nodes {
			klog.V(0).Infof("Scale-down: removing empty node %q", node.Name)
			a.ctx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: removing empty node %q", node.Name)

			if sdNode, err := a.scaleDownNodeToReport(node, false); err == nil {
				reportedSDNodes = append(reportedSDNodes, sdNode)
			} else {
				klog.Errorf("Scale-down: couldn't report scaled down node, err: %v", err)
			}

			a.nodeDeletionTracker.StartDeletion(bucket.Group.Id(), node.Name)
		}
	}

	for _, bucket := range NodeGroupViews {
		go a.deleteNodesAsync(bucket.Nodes, bucket.Group, false, bucket.BatchSize, nodeDeleteDelayAfterTaint)
	}

	return reportedSDNodes
}

// taintNodesSync synchronously taints all provided nodes with NoSchedule. If tainting fails for any of the nodes, already
// applied taints are cleaned up.
func (a *Actuator) taintNodesSync(NodeGroupViews []*budgets.NodeGroupView) (time.Duration, errors.AutoscalerError) {
	var taintedNodes []*apiv1.Node
	var updateLatencyTracker *UpdateLatencyTracker
	nodeDeleteDelayAfterTaint := a.nodeDeleteDelayAfterTaint
	if a.ctx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
		updateLatencyTracker = NewUpdateLatencyTracker(a.ctx.AutoscalingKubeClients.ListerRegistry.AllNodeLister())
		go updateLatencyTracker.Start()
	}
	for _, bucket := range NodeGroupViews {
		for _, node := range bucket.Nodes {
			if a.ctx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
				updateLatencyTracker.StartTimeChan <- nodeTaintStartTime{node.Name, time.Now()}
			}
			err := a.taintNode(node)
			if err != nil {
				a.ctx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
				// Clean up already applied taints in case of issues.
				for _, taintedNode := range taintedNodes {
					_, _ = taints.CleanToBeDeleted(taintedNode, a.ctx.ClientSet, a.ctx.CordonNodeBeforeTerminate)
				}
				if a.ctx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
					close(updateLatencyTracker.AwaitOrStopChan)
				}
				return nodeDeleteDelayAfterTaint, errors.NewAutoscalerError(errors.ApiCallError, "couldn't taint node %q with ToBeDeleted", node)
			}
			taintedNodes = append(taintedNodes, node)
		}
	}
	if a.ctx.AutoscalingOptions.DynamicNodeDeleteDelayAfterTaintEnabled {
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
func (a *Actuator) deleteAsyncDrain(NodeGroupViews []*budgets.NodeGroupView, nodeDeleteDelayAfterTaint time.Duration) (reportedSDNodes []*status.ScaleDownNode) {
	for _, bucket := range NodeGroupViews {
		for _, drainNode := range bucket.Nodes {
			if sdNode, err := a.scaleDownNodeToReport(drainNode, true); err == nil {
				klog.V(0).Infof("Scale-down: removing node %s, utilization: %v, pods to reschedule: %s", drainNode.Name, sdNode.UtilInfo, joinPodNames(sdNode.EvictedPods))
				a.ctx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDown", "Scale-down: removing node %s, utilization: %v, pods to reschedule: %s", drainNode.Name, sdNode.UtilInfo, joinPodNames(sdNode.EvictedPods))
				reportedSDNodes = append(reportedSDNodes, sdNode)
			} else {
				klog.Errorf("Scale-down: couldn't report scaled down node, err: %v", err)
			}

			a.nodeDeletionTracker.StartDeletionWithDrain(bucket.Group.Id(), drainNode.Name)
		}
	}

	for _, bucket := range NodeGroupViews {
		go a.deleteNodesAsync(bucket.Nodes, bucket.Group, true, bucket.BatchSize, nodeDeleteDelayAfterTaint)
	}

	return reportedSDNodes
}

func (a *Actuator) deleteNodesAsync(nodes []*apiv1.Node, nodeGroup cloudprovider.NodeGroup, drain bool, batchSize int, nodeDeleteDelayAfterTaint time.Duration) {
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
		klog.Errorf("Scale-down: couldn't create delete snapshot, err: %v", err)
		nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "createSnapshot returned error %v", err)}
		for _, node := range nodes {
			a.nodeDeletionScheduler.AbortNodeDeletion(node, nodeGroup.Id(), drain, "failed to create delete snapshot", nodeDeleteResult)
		}
		return
	}

	if drain {
		pdbs, err := a.ctx.PodDisruptionBudgetLister().List()
		if err != nil {
			klog.Errorf("Scale-down: couldn't fetch pod disruption budgets, err: %v", err)
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "podDisruptionBudgetLister.List returned error %v", err)}
			for _, node := range nodes {
				a.nodeDeletionScheduler.AbortNodeDeletion(node, nodeGroup.Id(), drain, "failed to fetch pod disruption budgets", nodeDeleteResult)
			}
			return
		}
		remainingPdbTracker = pdb.NewBasicRemainingPdbTracker()
		remainingPdbTracker.SetPdbs(pdbs)
		registry = a.ctx.ListerRegistry
	}

	if batchSize == 0 {
		batchSize = len(nodes)
	}

	for _, node := range nodes {
		nodeInfo, err := clusterSnapshot.GetNodeInfo(node.Name)
		if err != nil {
			klog.Errorf("Scale-down: can't retrieve node %q from snapshot, err: %v", node.Name, err)
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "nodeInfos.Get for %q returned error: %v", node.Name, err)}
			a.nodeDeletionScheduler.AbortNodeDeletion(node, nodeGroup.Id(), drain, "failed to get node info", nodeDeleteResult)
			continue
		}

		podsToRemove, _, _, err := simulator.GetPodsToMove(nodeInfo, a.deleteOptions, a.drainabilityRules, registry, remainingPdbTracker, time.Now())
		if err != nil {
			klog.Errorf("Scale-down: couldn't delete node %q, err: %v", node.Name, err)
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "GetPodsToMove for %q returned error: %v", node.Name, err)}
			a.nodeDeletionScheduler.AbortNodeDeletion(node, nodeGroup.Id(), drain, "failed to get pods to move on node", nodeDeleteResult)
			continue
		}

		if !drain && len(podsToRemove) != 0 {
			klog.Errorf("Scale-down: couldn't delete empty node %q, new pods got scheduled", node.Name)
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "failed to delete empty node %q, new pods scheduled", node.Name)}
			a.nodeDeletionScheduler.AbortNodeDeletion(node, nodeGroup.Id(), drain, "node is not empty", nodeDeleteResult)
			continue
		}

		go a.nodeDeletionScheduler.ScheduleDeletion(nodeInfo, nodeGroup, batchSize, drain)
	}
}

func (a *Actuator) scaleDownNodeToReport(node *apiv1.Node, drain bool) (*status.ScaleDownNode, error) {
	nodeGroup, err := a.ctx.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		return nil, err
	}
	nodeInfo, err := a.ctx.ClusterSnapshot.GetNodeInfo(node.Name)
	if err != nil {
		return nil, err
	}

	ignoreDaemonSetsUtilization, err := a.configGetter.GetIgnoreDaemonSetsUtilization(nodeGroup)
	if err != nil {
		return nil, err
	}

	gpuConfig := a.ctx.CloudProvider.GetNodeGpuConfig(node)
	utilInfo, err := utilization.Calculate(nodeInfo, ignoreDaemonSetsUtilization, a.ctx.IgnoreMirrorPodsUtilization, a.ctx.EnableDynamicResources, gpuConfig, time.Now())
	if err != nil {
		return nil, err
	}
	var evictedPods []*apiv1.Pod
	if drain {
		_, nonDsPodsToEvict := podsToEvict(nodeInfo, a.ctx.DaemonSetEvictionForOccupiedNodes)
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
	if err := taints.MarkToBeDeleted(node, a.ctx.ClientSet, a.ctx.CordonNodeBeforeTerminate); err != nil {
		a.ctx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
		return errors.ToAutoscalerError(errors.ApiCallError, err)
	}
	a.ctx.Recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "marked the node as toBeDeleted/unschedulable")
	return nil
}

func (a *Actuator) createSnapshot(nodes []*apiv1.Node) (clustersnapshot.ClusterSnapshot, error) {
	snapshot := clustersnapshot.NewBasicClusterSnapshot(a.ctx.FrameworkHandle, a.ctx.EnableDynamicResources)
	pods, err := a.ctx.AllPodLister().List()
	if err != nil {
		return nil, err
	}

	scheduledPods := kube_util.ScheduledPods(pods)
	expendableScheduledPods, nonExpendableScheduledPods := utils.SplitExpendablePods(scheduledPods, a.ctx.ExpendablePodsPriorityCutoff)

	draSnapshot := dynamicresources.Snapshot{}
	if a.ctx.EnableDynamicResources {
		// Grab a live snapshot of DRA objects.
		draSnap, err := a.draProvider.Snapshot()
		if err != nil {
			klog.Warningf("Couldn't retrieve DRA objects, this probably means that DRA is misconfigured in the cluster. Scaling involving DRA pods won't work, proceeding. Error: %v", err)
		} else {
			draSnapshot = draSnap
		}

		for _, expendablePod := range expendableScheduledPods {
			draSnapshot.RemovePodClaims(expendablePod)
		}
	}

	err = snapshot.Initialize(nodes, nonExpendableScheduledPods, draSnapshot)
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
