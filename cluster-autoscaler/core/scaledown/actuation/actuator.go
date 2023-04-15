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
	"reflect"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// Actuator is responsible for draining and deleting nodes.
type Actuator struct {
	ctx                 *context.AutoscalingContext
	clusterState        *clusterstate.ClusterStateRegistry
	nodeDeletionTracker *deletiontracker.NodeDeletionTracker
	nodeDeletionBatcher *NodeDeletionBatcher
	evictor             Evictor
	deleteOptions       simulator.NodeDeleteOptions
}

// NewActuator returns a new instance of Actuator.
func NewActuator(ctx *context.AutoscalingContext, csr *clusterstate.ClusterStateRegistry, ndt *deletiontracker.NodeDeletionTracker, deleteOptions simulator.NodeDeleteOptions) *Actuator {
	nbd := NewNodeDeletionBatcher(ctx, csr, ndt, ctx.NodeDeletionBatcherInterval)
	return &Actuator{
		ctx:                 ctx,
		clusterState:        csr,
		nodeDeletionTracker: ndt,
		nodeDeletionBatcher: nbd,
		evictor:             NewDefaultEvictor(deleteOptions, ndt),
		deleteOptions:       deleteOptions,
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

// StartDeletion triggers a new deletion process.
func (a *Actuator) StartDeletion(empty, drain []*apiv1.Node) (*status.ScaleDownStatus, errors.AutoscalerError) {
	deletionStartTime := time.Now()
	defer func() { metrics.UpdateDuration(metrics.ScaleDownNodeDeletion, time.Now().Sub(deletionStartTime)) }()

	results, ts := a.nodeDeletionTracker.DeletionResults()
	scaleDownStatus := &status.ScaleDownStatus{NodeDeleteResults: results, NodeDeleteResultsAsOf: ts}

	emptyToDelete, drainToDelete := a.cropNodesToBudgets(empty, drain)
	if len(emptyToDelete) == 0 && len(drainToDelete) == 0 {
		scaleDownStatus.Result = status.ScaleDownNoNodeDeleted
		return scaleDownStatus, nil
	}

	if len(emptyToDelete) > 0 {
		// Taint all empty nodes synchronously
		if err := a.taintNodesSync(emptyToDelete); err != nil {
			scaleDownStatus.Result = status.ScaleDownError
			return scaleDownStatus, err
		}

		emptyScaledDown, err := a.deleteAsyncEmpty(emptyToDelete)
		scaleDownStatus.ScaledDownNodes = append(scaleDownStatus.ScaledDownNodes, emptyScaledDown...)
		if err != nil {
			scaleDownStatus.Result = status.ScaleDownError
			return scaleDownStatus, err
		}
	}

	if len(drainToDelete) > 0 {
		// Taint all nodes that need drain synchronously, but don't start any drain/deletion yet. Otherwise, pods evicted from one to-be-deleted node
		// could get recreated on another.
		if err := a.taintNodesSync(drainToDelete); err != nil {
			scaleDownStatus.Result = status.ScaleDownError
			return scaleDownStatus, err
		}

		// All nodes involved in the scale-down should be tainted now - start draining and deleting nodes asynchronously.
		drainScaledDown := a.deleteAsyncDrain(drainToDelete)
		scaleDownStatus.ScaledDownNodes = append(scaleDownStatus.ScaledDownNodes, drainScaledDown...)
	}

	scaleDownStatus.Result = status.ScaleDownNodeDeleteStarted
	return scaleDownStatus, nil
}

// cropNodesToBudgets crops the provided node lists to respect scale-down max parallelism budgets.
func (a *Actuator) cropNodesToBudgets(empty, needDrain []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	emptyInProgress, drainInProgress := a.nodeDeletionTracker.DeletionsInProgress()
	parallelismBudget := a.ctx.MaxScaleDownParallelism - len(emptyInProgress) - len(drainInProgress)
	drainBudget := a.ctx.MaxDrainParallelism - len(drainInProgress)

	var emptyToDelete []*apiv1.Node
	for _, node := range empty {
		if len(emptyToDelete) >= parallelismBudget {
			break
		}
		emptyToDelete = append(emptyToDelete, node)
	}

	parallelismBudgetLeft := parallelismBudget - len(emptyToDelete)
	drainBudget = min(parallelismBudgetLeft, drainBudget)

	var drainToDelete []*apiv1.Node
	for _, node := range needDrain {
		if len(drainToDelete) >= drainBudget {
			break
		}
		drainToDelete = append(drainToDelete, node)
	}

	return emptyToDelete, drainToDelete
}

// deleteAsyncEmpty immediately starts deletions asynchronously.
// scaledDownNodes return value contains all nodes for which deletion successfully started. It's valid and should be consumed
// even if err != nil.
func (a *Actuator) deleteAsyncEmpty(empty []*apiv1.Node) (scaledDownNodes []*status.ScaleDownNode, err errors.AutoscalerError) {
	var groupIds []string
	var validNodes []*apiv1.Node
	for _, emptyNode := range empty {
		klog.V(0).Infof("Scale-down: removing empty node %q", emptyNode.Name)
		a.ctx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: removing empty node %q", emptyNode.Name)

		nodeGroup, err := a.ctx.CloudProvider.NodeGroupForNode(emptyNode)
		if err != nil || nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.Errorf("Failed to find node group for %s: %v", emptyNode.Name, err)
			continue
		}

		if sdNode, err := a.scaleDownNodeToReport(emptyNode, false); err == nil {
			scaledDownNodes = append(scaledDownNodes, sdNode)
		} else {
			klog.Errorf("Scale-down: couldn't report scaled down node, err: %v", err)
		}

		a.nodeDeletionTracker.StartDeletion(nodeGroup.Id(), emptyNode.Name)
		groupIds = append(groupIds, nodeGroup.Id())
		validNodes = append(validNodes, emptyNode)
	}

	go a.deleteNodesAsync(validNodes, groupIds, false)

	return scaledDownNodes, nil
}

// taintNodesSync synchronously taints all provided nodes with NoSchedule. If tainting fails for any of the nodes, already
// applied taints are cleaned up.
func (a *Actuator) taintNodesSync(nodes []*apiv1.Node) errors.AutoscalerError {
	var taintedNodes []*apiv1.Node
	for _, node := range nodes {
		err := a.taintNode(node)
		if err != nil {
			a.ctx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
			// Clean up already applied taints in case of issues.
			for _, taintedNode := range taintedNodes {
				_, _ = taints.CleanToBeDeleted(taintedNode, a.ctx.ClientSet, a.ctx.CordonNodeBeforeTerminate)
			}
			return errors.NewAutoscalerError(errors.ApiCallError, "couldn't taint node %q with ToBeDeleted", node)
		}
		taintedNodes = append(taintedNodes, node)
	}
	return nil
}

// deleteAsyncDrain asynchronously starts deletions with drain for all provided nodes. scaledDownNodes return value contains all nodes for which
// deletion successfully started.
func (a *Actuator) deleteAsyncDrain(drain []*apiv1.Node) (scaledDownNodes []*status.ScaleDownNode) {
	var groupIds []string
	var validNodes []*apiv1.Node
	for _, drainNode := range drain {
		if sdNode, err := a.scaleDownNodeToReport(drainNode, true); err == nil {
			klog.V(0).Infof("Scale-down: removing node %s, utilization: %v, pods to reschedule: %s", drainNode.Name, sdNode.UtilInfo, joinPodNames(sdNode.EvictedPods))
			a.ctx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDown", "Scale-down: removing node %s, utilization: %v, pods to reschedule: %s", drainNode.Name, sdNode.UtilInfo, joinPodNames(sdNode.EvictedPods))
			scaledDownNodes = append(scaledDownNodes, sdNode)
		} else {
			klog.Errorf("Scale-down: couldn't report scaled down node, err: %v", err)
		}

		nodeGroup, err := a.ctx.CloudProvider.NodeGroupForNode(drainNode)
		if err != nil || nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.Errorf("Failed to find node group for %s: %v", drainNode.Name, err)
			continue
		}

		a.nodeDeletionTracker.StartDeletionWithDrain(nodeGroup.Id(), drainNode.Name)
		groupIds = append(groupIds, nodeGroup.Id())
		validNodes = append(validNodes, drainNode)
	}

	go a.deleteNodesAsync(validNodes, groupIds, true)

	return scaledDownNodes
}

func (a *Actuator) deleteNodesAsync(nodes []*apiv1.Node, groupIds []string, drain bool) {
	var pdbs []*policyv1.PodDisruptionBudget
	var registry kube_util.ListerRegistry

	if len(nodes) == 0 {
		return
	}

	if a.ctx.NodeDeleteDelayAfterTaint > time.Duration(0) {
		klog.V(0).Infof("Scale-down: waiting %v before trying to delete nodes", a.ctx.NodeDeleteDelayAfterTaint)
		time.Sleep(a.ctx.NodeDeleteDelayAfterTaint)
	}

	clusterSnapshot, err := a.createSnapshot(nodes)

	if err != nil {
		klog.Errorf("Scale-down: couldn't create delete snapshot, err: %v", err)
		nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "createSnapshot returned error %v", err)}
		for i, node := range nodes {
			CleanUpAndRecordFailedScaleDownEvent(a.ctx, node, groupIds[i], drain, a.nodeDeletionTracker, "failed to create delete snapshot", nodeDeleteResult)
		}
		return
	}

	if drain {
		pdbs, err = a.ctx.PodDisruptionBudgetLister().List()
		if err != nil {
			klog.Errorf("Scale-down: couldn't fetch pod disruption budgets, err: %v", err)
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "podDisruptionBudgetLister.List returned error %v", err)}
			for i, node := range nodes {
				CleanUpAndRecordFailedScaleDownEvent(a.ctx, node, groupIds[i], drain, a.nodeDeletionTracker, "failed to fetch pod disruption budgets", nodeDeleteResult)
			}
			return
		}

		registry = a.ctx.ListerRegistry
	}

	for i, node := range nodes {
		nodeInfo, err := clusterSnapshot.NodeInfos().Get(node.Name)
		if err != nil {
			klog.Errorf("Scale-down: can't retrieve node %q from snapshot, err: %v", node.Name, err)
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "nodeInfos.Get for %q returned error: %v", node.Name, err)}
			CleanUpAndRecordFailedScaleDownEvent(a.ctx, node, groupIds[i], drain, a.nodeDeletionTracker, "failed to get node info", nodeDeleteResult)
			continue
		}

		podsToRemove, _, _, err := simulator.GetPodsToMove(nodeInfo, a.deleteOptions, registry, pdbs, time.Now())
		if err != nil {
			klog.Errorf("Scale-down: couldn't delete node %q, err: %v", node.Name, err)
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "GetPodsToMove for %q returned error: %v", node.Name, err)}
			CleanUpAndRecordFailedScaleDownEvent(a.ctx, node, groupIds[i], drain, a.nodeDeletionTracker, "failed to get pods to move on node", nodeDeleteResult)
			continue
		}

		if !drain && len(podsToRemove) != 0 {
			klog.Errorf("Scale-down: couldn't delete empty node %q, new pods got scheduled", node.Name)
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "failed to delete empty node %q, new pods scheduled", node.Name)}
			CleanUpAndRecordFailedScaleDownEvent(a.ctx, node, groupIds[i], drain, a.nodeDeletionTracker, "node is not empty", nodeDeleteResult)
			continue
		}

		go a.scheduleDeletion(nodeInfo, groupIds[i], drain)
	}
}

func (a *Actuator) scaleDownNodeToReport(node *apiv1.Node, drain bool) (*status.ScaleDownNode, error) {
	nodeGroup, err := a.ctx.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		return nil, err
	}
	nodeInfo, err := a.ctx.ClusterSnapshot.NodeInfos().Get(node.Name)
	if err != nil {
		return nil, err
	}
	gpuConfig := a.ctx.CloudProvider.GetNodeGpuConfig(node)
	utilInfo, err := utilization.Calculate(nodeInfo, a.ctx.IgnoreDaemonSetsUtilization, a.ctx.IgnoreMirrorPodsUtilization, gpuConfig, time.Now())
	if err != nil {
		return nil, err
	}
	var evictedPods []*apiv1.Pod
	if drain {
		_, nonDsPodsToEvict := podsToEvict(a.ctx, nodeInfo)
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

func (a *Actuator) prepareNodeForDeletion(nodeInfo *framework.NodeInfo, drain bool) status.NodeDeleteResult {
	node := nodeInfo.Node()
	if drain {
		if evictionResults, err := a.evictor.DrainNode(a.ctx, nodeInfo); err != nil {
			return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToEvictPods, Err: err, PodEvictionResults: evictionResults}
		}
	} else {
		if err := a.evictor.EvictDaemonSetPods(a.ctx, nodeInfo, time.Now()); err != nil {
			// Evicting DS pods is best-effort, so proceed with the deletion even if there are errors.
			klog.Warningf("Error while evicting DS pods from an empty node %q: %v", node.Name, err)
		}
	}
	if err := WaitForDelayDeletion(node, a.ctx.ListerRegistry.AllNodeLister(), a.ctx.AutoscalingOptions.NodeDeletionDelayTimeout); err != nil {
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: err}
	}
	return status.NodeDeleteResult{ResultType: status.NodeDeleteOk}
}

// scheduleDeletion schedule the deletion on of the provided node by adding a node to NodeDeletionBatcher. If drain is true, the node is drained before being deleted.
func (a *Actuator) scheduleDeletion(nodeInfo *framework.NodeInfo, nodeGroupId string, drain bool) {
	node := nodeInfo.Node()
	nodeDeleteResult := a.prepareNodeForDeletion(nodeInfo, drain)
	if nodeDeleteResult.Err != nil {
		CleanUpAndRecordFailedScaleDownEvent(a.ctx, node, nodeGroupId, drain, a.nodeDeletionTracker, "prepareNodeForDeletion failed", nodeDeleteResult)
		return
	}
	err := a.nodeDeletionBatcher.AddNode(node, drain)
	if err != nil {
		klog.Errorf("Couldn't add node to nodeDeletionBatcher, err: %v", err)
		nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "nodeDeletionBatcher.AddNode for %s returned error: %v", node.Name, err)}
		CleanUpAndRecordFailedScaleDownEvent(a.ctx, node, nodeGroupId, drain, a.nodeDeletionTracker, "failed add node to the nodeDeletionBatche", nodeDeleteResult)
	}
}

func (a *Actuator) createSnapshot(nodes []*apiv1.Node) (clustersnapshot.ClusterSnapshot, error) {
	knownNodes := make(map[string]bool)
	snapshot := clustersnapshot.NewBasicClusterSnapshot()

	scheduledPods, err := a.ctx.ScheduledPodLister().List()
	if err != nil {
		return nil, err
	}

	nonExpendableScheduledPods := utils.FilterOutExpendablePods(scheduledPods, a.ctx.ExpendablePodsPriorityCutoff)

	for _, node := range nodes {
		if err := snapshot.AddNode(node); err != nil {
			return nil, err
		}

		knownNodes[node.Name] = true
	}

	for _, pod := range nonExpendableScheduledPods {
		if knownNodes[pod.Spec.NodeName] {
			if err := snapshot.AddPod(pod, pod.Spec.NodeName); err != nil {
				return nil, err
			}
		}
	}

	return snapshot, nil
}

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}

func joinPodNames(pods []*apiv1.Pod) string {
	var names []string
	for _, pod := range pods {
		names = append(names, pod.Name)
	}
	return strings.Join(names, ",")
}

// CleanUpAndRecordFailedScaleDownEvent record failed scale down event and log an error.
func CleanUpAndRecordFailedScaleDownEvent(ctx *context.AutoscalingContext, node *apiv1.Node, nodeGroupId string, drain bool, nodeDeletionTracker *deletiontracker.NodeDeletionTracker, errMsg string, status status.NodeDeleteResult) {
	if drain {
		klog.Errorf("Scale-down: couldn't delete node %q with drain, %v, status error: %v", node.Name, errMsg, status.Err)
		ctx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to drain and delete node: %v", status.Err)

	} else {
		klog.Errorf("Scale-down: couldn't delete empty node, %v, status error: %v", errMsg, status.Err)
		ctx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to delete empty node: %v", status.Err)
	}
	taints.CleanToBeDeleted(node, ctx.ClientSet, ctx.CordonNodeBeforeTerminate)
	nodeDeletionTracker.EndDeletion(nodeGroupId, node.Name, status)
}

// RegisterAndRecordSuccessfulScaleDownEvent register scale down and record successful scale down event.
func RegisterAndRecordSuccessfulScaleDownEvent(ctx *context.AutoscalingContext, csr *clusterstate.ClusterStateRegistry, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup, drain bool, nodeDeletionTracker *deletiontracker.NodeDeletionTracker) {
	ctx.Recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "nodes removed by cluster autoscaler")
	csr.RegisterScaleDown(&clusterstate.ScaleDownRequest{
		NodeGroup:          nodeGroup,
		NodeName:           node.Name,
		Time:               time.Now(),
		ExpectedDeleteTime: time.Now().Add(MaxCloudProviderNodeDeletionTime),
	})
	gpuConfig := ctx.CloudProvider.GetNodeGpuConfig(node)
	metricResourceName, metricGpuType := gpu.GetGpuInfoForMetrics(gpuConfig, ctx.CloudProvider.GetAvailableGPUTypes(), node, nodeGroup)
	metrics.RegisterScaleDown(1, metricResourceName, metricGpuType, nodeScaleDownReason(node, drain))
	if drain {
		ctx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDown", "Scale-down: node %s removed with drain", node.Name)
	} else {
		ctx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: empty node %s removed", node.Name)
	}
	nodeDeletionTracker.EndDeletion(nodeGroup.Id(), node.Name, status.NodeDeleteResult{ResultType: status.NodeDeleteOk})
}
