/*
Copyright 2023 The Kubernetes Authors.

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
	"sync"

	apiv1 "k8s.io/api/core/v1"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

type batcher interface {
	AddNodes(ctx context.Context, nodes []*apiv1.Node, nodeGroup cloudprovider.NodeGroup, drain bool)
}

// GroupDeletionScheduler is a wrapper over NodeDeletionBatcher responsible for grouping nodes for deletion
// and rolling back deletion of all nodes from a group in case deletion fails for any of the other nodes.
type GroupDeletionScheduler struct {
	sync.Mutex
	autoscalingCtx      *ca_context.AutoscalingContext
	nodeDeletionTracker *deletiontracker.NodeDeletionTracker
	nodeDeletionBatcher batcher
	evictor             Evictor
	nodeQueue           map[string][]*apiv1.Node
	failuresForGroup    map[string]bool
}

// NewGroupDeletionScheduler creates an instance of GroupDeletionScheduler.
func NewGroupDeletionScheduler(autoscalingCtx *ca_context.AutoscalingContext, ndt *deletiontracker.NodeDeletionTracker, b batcher, evictor Evictor) *GroupDeletionScheduler {
	return &GroupDeletionScheduler{
		autoscalingCtx:      autoscalingCtx,
		nodeDeletionTracker: ndt,
		nodeDeletionBatcher: b,
		evictor:             evictor,
		nodeQueue:           map[string][]*apiv1.Node{},
		failuresForGroup:    map[string]bool{},
	}
}

// ResetAndReportMetrics should be invoked for GroupDeletionScheduler before each scale-down phase.
func (ds *GroupDeletionScheduler) ResetAndReportMetrics() {
	ds.Lock()
	defer ds.Unlock()
	pendingNodeDeletions := 0
	for _, nodes := range ds.nodeQueue {
		pendingNodeDeletions += len(nodes)
	}
	ds.failuresForGroup = map[string]bool{}
	// Since the nodes are deleted asynchronously, it's easier to
	// monitor the pending ones at the beginning of the next scale-down phase.
	metrics.ObservePendingNodeDeletions(pendingNodeDeletions)
}

// ScheduleDeletion schedules deletion of the node. Nodes that should be deleted in groups are queued until whole group is scheduled for deletion,
// other nodes are passed over to NodeDeletionBatcher immediately.
func (ds *GroupDeletionScheduler) ScheduleDeletion(ctx context.Context, nodeInfo *framework.NodeInfo, nodeGroup cloudprovider.NodeGroup, batchSize int, drain bool) {
	ds.scheduleDeletion(ctx, nodeInfo, nodeGroup, batchSize, drain, false)
}

// scheduleForceDeletion schedules forced node deletion, similar to ScheduleDeletion but bypassing eviction errors and PDB checks.
func (ds *GroupDeletionScheduler) scheduleForceDeletion(ctx context.Context, nodeInfo *framework.NodeInfo, nodeGroup cloudprovider.NodeGroup, batchSize int, drain bool) {
	ds.scheduleDeletion(ctx, nodeInfo, nodeGroup, batchSize, drain, true)
}

// scheduleDeletion handles the common logic for scheduling node deletion, supporting
// both normal and forced deletion based on the 'force' parameter.
func (ds *GroupDeletionScheduler) scheduleDeletion(ctx context.Context, nodeInfo *framework.NodeInfo, nodeGroup cloudprovider.NodeGroup, batchSize int, drain bool, force bool) {
	logger := klog.FromContext(ctx)
	opts, err := nodeGroup.GetOptions(ds.autoscalingCtx.NodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerErrorf(errors.InternalError, "GetOptions returned error %v", err)}
		ds.AbortNodeDeletionDueToError(ctx, nodeInfo.Node(), nodeGroup.Id(), drain, "failed to get autoscaling options for a node group", nodeDeleteResult)
		return
	}
	if opts == nil {
		opts = &config.NodeGroupAutoscalingOptions{}
	}

	nodeDeleteResult := ds.prepareNodeForDeletion(ctx, nodeInfo, drain, force)
	if nodeDeleteResult.Err != nil {
		if force {
			logger.Info("Starting force deletion of node", "name", nodeInfo.Node().Name)
			if err := nodeGroup.ForceDeleteNodes([]*apiv1.Node{nodeInfo.Node()}); err != nil {
				focrefulNodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: err}
				ds.AbortNodeDeletion(ctx, nodeInfo.Node(), nodeGroup.Id(), drain, "forceful node deletion failed", focrefulNodeDeleteResult, true)
				return
			}
		} else {
			ds.AbortNodeDeletion(ctx, nodeInfo.Node(), nodeGroup.Id(), drain, "prepareNodeForDeletion failed", nodeDeleteResult, true)
			return
		}
	}

	ds.addToBatcher(ctx, nodeInfo, nodeGroup, batchSize, drain, opts.ZeroOrMaxNodeScaling)
}

// prepareNodeForDeletion is a long-running operation, so it needs to avoid locking the AtomicDeletionScheduler object
func (ds *GroupDeletionScheduler) prepareNodeForDeletion(ctx context.Context, nodeInfo *framework.NodeInfo, drain bool, force bool) status.NodeDeleteResult {
	logger := klog.FromContext(ctx)
	node := nodeInfo.Node()
	if drain {
		var evictionResults map[string]status.PodEvictionResult
		var err error
		if force {
			evictionResults, err = ds.evictor.drainNodeForce(ctx, ds.autoscalingCtx, nodeInfo)
		} else {
			evictionResults, err = ds.evictor.DrainNode(ctx, ds.autoscalingCtx, nodeInfo)
		}
		if err != nil {
			return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToEvictPods, Err: err, PodEvictionResults: evictionResults}
		}
	} else {
		if _, err := ds.evictor.EvictDaemonSetPods(ctx, ds.autoscalingCtx, nodeInfo); err != nil {
			logger.
				// Evicting DS pods is best-effort, so proceed with the deletion even if there are errors.
				Error(err, "Error while evicting DS pods from an empty node", "node", node.Name)
		}
	}
	if err := WaitForDelayDeletion(ctx, node, ds.autoscalingCtx.ListerRegistry.AllNodeLister(), ds.autoscalingCtx.AutoscalingOptions.NodeDeletionDelayTimeout); err != nil {
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: err}
	}
	return status.NodeDeleteResult{ResultType: status.NodeDeleteOk}
}

func (ds *GroupDeletionScheduler) addToBatcher(ctx context.Context, nodeInfo *framework.NodeInfo, nodeGroup cloudprovider.NodeGroup, batchSize int, drain, atomic bool) {
	ds.Lock()
	defer ds.Unlock()
	ds.nodeQueue[nodeGroup.Id()] = append(ds.nodeQueue[nodeGroup.Id()], nodeInfo.Node())
	if atomic {
		if ds.failuresForGroup[nodeGroup.Id()] {
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: errors.NewAutoscalerError(errors.TransientError, "couldn't scale down other nodes in this node group")}
			CleanUpAndRecordErrorForFailedScaleDownEvent(ctx, ds.autoscalingCtx, nodeInfo.Node(), nodeGroup.Id(), drain, ds.nodeDeletionTracker, "scale down failed for node group as a whole", nodeDeleteResult)
			delete(ds.nodeQueue, nodeGroup.Id())
		}
		if len(ds.nodeQueue[nodeGroup.Id()]) < batchSize {
			// node group should be scaled down atomically, but not all nodes are ready yet
			return
		}
	}
	ds.nodeDeletionBatcher.AddNodes(ctx, ds.nodeQueue[nodeGroup.Id()], nodeGroup, drain)
	ds.nodeQueue[nodeGroup.Id()] = []*apiv1.Node{}
}

// AbortNodeDeletionDueToError frees up a node that couldn't be deleted successfully. If it was a part of a group, the same is applied for other nodes queued for deletion.
func (ds *GroupDeletionScheduler) AbortNodeDeletionDueToError(ctx context.Context, node *apiv1.Node, nodeGroupId string, drain bool, errMsg string, result status.NodeDeleteResult) {
	ds.AbortNodeDeletion(ctx, node, nodeGroupId, drain, errMsg, result, false)
}

// AbortNodeDeletion frees up a node that couldn't be deleted successfully. If it was a part of a group, the same is applied for other nodes queued for deletion.
func (ds *GroupDeletionScheduler) AbortNodeDeletion(ctx context.Context, node *apiv1.Node, nodeGroupId string, drain bool, errMsg string, result status.NodeDeleteResult, logAsWarning bool) {
	ds.Lock()
	defer ds.Unlock()
	ds.failuresForGroup[nodeGroupId] = true
	CleanUpAndRecordFailedScaleDownEvent(ctx, ds.autoscalingCtx, node, nodeGroupId, drain, ds.nodeDeletionTracker, errMsg, result, logAsWarning)
	for _, otherNode := range ds.nodeQueue[nodeGroupId] {
		if otherNode == node {
			continue
		}
		nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: errors.NewAutoscalerError(errors.TransientError, "couldn't scale down other nodes in this node group")}
		CleanUpAndRecordFailedScaleDownEvent(ctx, ds.autoscalingCtx, otherNode, nodeGroupId, drain, ds.nodeDeletionTracker, "scale down failed for node group as a whole", nodeDeleteResult, logAsWarning)
	}
	delete(ds.nodeQueue, nodeGroupId)
}
