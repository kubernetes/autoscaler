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
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

type batcher interface {
	AddNodes(nodes []*apiv1.Node, nodeGroup cloudprovider.NodeGroup, drain bool)
}

// GroupDeletionScheduler is a wrapper over NodeDeletionBatcher responsible for grouping nodes for deletion
// and rolling back deletion of all nodes from a group in case deletion fails for any of the other nodes.
type GroupDeletionScheduler struct {
	sync.Mutex
	ctx                 *context.AutoscalingContext
	nodeDeletionTracker *deletiontracker.NodeDeletionTracker
	nodeDeletionBatcher batcher
	evictor             Evictor
	nodeQueue           map[string][]*apiv1.Node
	failuresForGroup    map[string]bool
}

// NewGroupDeletionScheduler creates an instance of GroupDeletionScheduler.
func NewGroupDeletionScheduler(ctx *context.AutoscalingContext, ndt *deletiontracker.NodeDeletionTracker, b batcher, evictor Evictor) *GroupDeletionScheduler {
	return &GroupDeletionScheduler{
		ctx:                 ctx,
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
func (ds *GroupDeletionScheduler) ScheduleDeletion(nodeInfo *framework.NodeInfo, nodeGroup cloudprovider.NodeGroup, batchSize int, drain bool) {
	opts, err := nodeGroup.GetOptions(ds.ctx.NodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorInternal, Err: errors.NewAutoscalerError(errors.InternalError, "GetOptions returned error %v", err)}
		ds.AbortNodeDeletion(nodeInfo.Node(), nodeGroup.Id(), drain, "failed to get autoscaling options for a node group", nodeDeleteResult)
		return
	}
	if opts == nil {
		opts = &config.NodeGroupAutoscalingOptions{}
	}

	nodeDeleteResult := ds.prepareNodeForDeletion(nodeInfo, drain)
	if nodeDeleteResult.Err != nil {
		ds.AbortNodeDeletion(nodeInfo.Node(), nodeGroup.Id(), drain, "prepareNodeForDeletion failed", nodeDeleteResult)
		return
	}

	ds.addToBatcher(nodeInfo, nodeGroup, batchSize, drain, opts.ZeroOrMaxNodeScaling)
}

// prepareNodeForDeletion is a long-running operation, so it needs to avoid locking the AtomicDeletionScheduler object
func (ds *GroupDeletionScheduler) prepareNodeForDeletion(nodeInfo *framework.NodeInfo, drain bool) status.NodeDeleteResult {
	node := nodeInfo.Node()
	if drain {
		if evictionResults, err := ds.evictor.DrainNode(ds.ctx, nodeInfo); err != nil {
			return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToEvictPods, Err: err, PodEvictionResults: evictionResults}
		}
	} else {
		if _, err := ds.evictor.EvictDaemonSetPods(ds.ctx, nodeInfo); err != nil {
			// Evicting DS pods is best-effort, so proceed with the deletion even if there are errors.
			klog.Warningf("Error while evicting DS pods from an empty node %q: %v", node.Name, err)
		}
	}
	if err := WaitForDelayDeletion(node, ds.ctx.ListerRegistry.AllNodeLister(), ds.ctx.AutoscalingOptions.NodeDeletionDelayTimeout); err != nil {
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: err}
	}
	return status.NodeDeleteResult{ResultType: status.NodeDeleteOk}
}

func (ds *GroupDeletionScheduler) addToBatcher(nodeInfo *framework.NodeInfo, nodeGroup cloudprovider.NodeGroup, batchSize int, drain, atomic bool) {
	ds.Lock()
	defer ds.Unlock()
	ds.nodeQueue[nodeGroup.Id()] = append(ds.nodeQueue[nodeGroup.Id()], nodeInfo.Node())
	if atomic {
		if ds.failuresForGroup[nodeGroup.Id()] {
			nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: errors.NewAutoscalerError(errors.TransientError, "couldn't scale down other nodes in this node group")}
			CleanUpAndRecordFailedScaleDownEvent(ds.ctx, nodeInfo.Node(), nodeGroup.Id(), drain, ds.nodeDeletionTracker, "scale down failed for node group as a whole", nodeDeleteResult)
			delete(ds.nodeQueue, nodeGroup.Id())
		}
		if len(ds.nodeQueue[nodeGroup.Id()]) < batchSize {
			// node group should be scaled down atomically, but not all nodes are ready yet
			return
		}
	}
	ds.nodeDeletionBatcher.AddNodes(ds.nodeQueue[nodeGroup.Id()], nodeGroup, drain)
	ds.nodeQueue[nodeGroup.Id()] = []*apiv1.Node{}
}

// AbortNodeDeletion frees up a node that couldn't be deleted successfully. If it was a part of a group, the same is applied for other nodes queued for deletion.
func (ds *GroupDeletionScheduler) AbortNodeDeletion(node *apiv1.Node, nodeGroupId string, drain bool, errMsg string, result status.NodeDeleteResult) {
	ds.Lock()
	defer ds.Unlock()
	ds.failuresForGroup[nodeGroupId] = true
	CleanUpAndRecordFailedScaleDownEvent(ds.ctx, node, nodeGroupId, drain, ds.nodeDeletionTracker, errMsg, result)
	for _, otherNode := range ds.nodeQueue[nodeGroupId] {
		if otherNode == node {
			continue
		}
		nodeDeleteResult := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: errors.NewAutoscalerError(errors.TransientError, "couldn't scale down other nodes in this node group")}
		CleanUpAndRecordFailedScaleDownEvent(ds.ctx, otherNode, nodeGroupId, drain, ds.nodeDeletionTracker, "scale down failed for node group as a whole", nodeDeleteResult)
	}
	delete(ds.nodeQueue, nodeGroupId)
}
