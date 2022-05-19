package actuation

import (
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// Actuator is responsible for draining and deleting nodes.
type Actuator struct {
	ctx                 *context.AutoscalingContext
	clusterState        *clusterstate.ClusterStateRegistry
	nodeDeletionTracker *deletiontracker.NodeDeletionTracker
}

// NewActuator returns a new instance of Actuator.
func NewActuator(ctx *context.AutoscalingContext, csr *clusterstate.ClusterStateRegistry, ndr *deletiontracker.NodeDeletionTracker) *Actuator {
	return &Actuator{
		ctx:                 ctx,
		clusterState:        csr,
		nodeDeletionTracker: ndr,
	}
}

// CheckStatus should return an immutable snapshot of ongoing deletions. Before the TODO is addressed, a live object
// is returned instead of an immutable snapshot.
func (a *Actuator) CheckStatus() scaledown.ActuationStatus {
	// TODO: snapshot information from the tracker instead of keeping live
	// updated object.
	return a.nodeDeletionTracker
}

// ClearResultsNotNewerThan removes information about deletions finished before or exactly at the provided timestamp.
func (a *Actuator) ClearResultsNotNewerThan(t time.Time) {
	a.nodeDeletionTracker.ClearResultsNotNewerThan(t)
}

// StartDeletion triggers a new deletion process.
func (a *Actuator) StartDeletion(empty, drain []*apiv1.Node, currentTime time.Time) (*status.ScaleDownStatus, errors.AutoscalerError) {
	defer func() { metrics.UpdateDuration(metrics.ScaleDownNodeDeletion, time.Now().Sub(currentTime)) }()
	results, ts := a.nodeDeletionTracker.DeletionResults()
	scaleDownStatus := &status.ScaleDownStatus{NodeDeleteResults: results, NodeDeleteResultsAsOf: ts}

	emptyToDelete, drainToDelete := a.cropNodesToBudgets(empty, drain)
	if len(emptyToDelete) == 0 && len(drainToDelete) == 0 {
		scaleDownStatus.Result = status.ScaleDownNoNodeDeleted
		return scaleDownStatus, nil
	}

	// Taint empty nodes synchronously, and immediately start deletions asynchronously. Because these nodes are empty, there's no risk that a pod from one
	// to-be-deleted node gets recreated on another.
	emptyScaledDown, err := a.taintSyncDeleteAsyncEmpty(emptyToDelete)
	scaleDownStatus.ScaledDownNodes = append(scaleDownStatus.ScaledDownNodes, emptyScaledDown...)
	if err != nil {
		scaleDownStatus.Result = status.ScaleDownError
		return scaleDownStatus, err
	}

	// Taint all nodes that need drain synchronously, but don't start any drain/deletion yet. Otherwise, pods evicted from one to-be-deleted node
	// could get recreated on another.
	err = a.taintNodesSync(drainToDelete)
	if err != nil {
		scaleDownStatus.Result = status.ScaleDownError
		return scaleDownStatus, err
	}

	// All nodes involved in the scale-down should be tainted now - start draining and deleting nodes asynchronously.
	drainScaledDown := a.deleteAsyncDrain(drainToDelete)
	scaleDownStatus.ScaledDownNodes = append(scaleDownStatus.ScaledDownNodes, drainScaledDown...)

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

// taintSyncDeleteAsyncEmpty synchronously taints the provided empty nodes, and immediately starts deletions asynchronously.
// scaledDownNodes return value contains all nodes for which deletion successfully started. It's valid and should be consumed
// even if err != nil.
func (a *Actuator) taintSyncDeleteAsyncEmpty(empty []*apiv1.Node) (scaledDownNodes []*status.ScaleDownNode, err errors.AutoscalerError) {
	for _, emptyNode := range empty {
		klog.V(0).Infof("Scale-down: removing empty node %q", emptyNode.Name)
		a.ctx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: removing empty node %q", emptyNode.Name)

		err := a.taintNode(emptyNode)
		if err != nil {
			a.ctx.Recorder.Eventf(emptyNode, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
			return scaledDownNodes, errors.NewAutoscalerError(errors.ApiCallError, "couldn't taint node %q with ToBeDeleted", emptyNode.Name)
		}

		if sdNode, err := a.scaleDownNodeToReport(emptyNode, false); err == nil {
			scaledDownNodes = append(scaledDownNodes, sdNode)
		} else {
			klog.Errorf("Scale-down: couldn't report scaled down node, err: %v", err)
		}

		go func(node *apiv1.Node) {
			result := a.deleteNode(node, false)
			if result.Err == nil {
				a.ctx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty", "Scale-down: empty node %s removed", node.Name)
			} else {
				klog.Errorf("Scale-down: couldn't delete empty node, err: %v", err)
				a.ctx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to delete empty node: %v", result.Err)
				_, _ = deletetaint.CleanToBeDeleted(node, a.ctx.ClientSet, a.ctx.CordonNodeBeforeTerminate)
			}
		}(emptyNode)
	}
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
				_, _ = deletetaint.CleanToBeDeleted(taintedNode, a.ctx.ClientSet, a.ctx.CordonNodeBeforeTerminate)
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
	for _, drainNode := range drain {
		if sdNode, err := a.scaleDownNodeToReport(drainNode, true); err == nil {
			klog.V(0).Infof("Scale-down: removing node %s, utilization: %v, pods to reschedule: %s", drainNode.Name, sdNode.UtilInfo, joinPodNames(sdNode.EvictedPods))
			a.ctx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDown", "Scale-down: removing node %s, utilization: %v, pods to reschedule: %s", drainNode.Name, sdNode.UtilInfo, joinPodNames(sdNode.EvictedPods))
			scaledDownNodes = append(scaledDownNodes, sdNode)
		} else {
			klog.Errorf("Scale-down: couldn't report scaled down node, err: %v", err)
		}

		go func(node *apiv1.Node) {
			result := a.deleteNode(node, true)
			if result.Err == nil {
				a.ctx.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDown", "Scale-down: node %s removed with drain", node.Name)
			} else {
				klog.Errorf("Scale-down: couldn't delete node %q with drain, err: %v", node.Name, result.Err)
				a.ctx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to drain and delete node: %v", result.Err)
				_, _ = deletetaint.CleanToBeDeleted(node, a.ctx.ClientSet, a.ctx.CordonNodeBeforeTerminate)
			}
		}(drainNode)
	}
	return scaledDownNodes
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
	utilInfo, err := utilization.Calculate(node, nodeInfo, a.ctx.IgnoreDaemonSetsUtilization, a.ctx.IgnoreMirrorPodsUtilization, a.ctx.CloudProvider.GPULabel(), time.Now())
	if err != nil {
		return nil, err
	}
	var evictedPods []*apiv1.Pod
	if drain {
		_, nonDsPodsToEvict, err := podsToEvict(a.ctx, node.Name)
		if err != nil {
			return nil, err
		}
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
	if err := deletetaint.MarkToBeDeleted(node, a.ctx.ClientSet, a.ctx.CordonNodeBeforeTerminate); err != nil {
		a.ctx.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to mark the node as toBeDeleted/unschedulable: %v", err)
		return errors.ToAutoscalerError(errors.ApiCallError, err)
	}
	a.ctx.Recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "marked the node as toBeDeleted/unschedulable")
	return nil
}

// deleteNode performs the deletion of the provided node. If drain is true, the node is drained before being deleted.
func (a *Actuator) deleteNode(node *apiv1.Node, drain bool) (result status.NodeDeleteResult) {
	// TODO: Implement.
	return status.NodeDeleteResult{}
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
