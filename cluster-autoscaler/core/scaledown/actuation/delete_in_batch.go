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
	"fmt"
	"reflect"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/observers/nodegroupchange"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/klog/v2"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

const (
	// MaxKubernetesEmptyNodeDeletionTime is the maximum time needed by Kubernetes to delete an empty node.
	MaxKubernetesEmptyNodeDeletionTime = 3 * time.Minute
)

// NodeDeletionBatcher batch scale down candidates for one node group and remove them.
type NodeDeletionBatcher struct {
	sync.Mutex
	ctx                   *context.AutoscalingContext
	scaleStateNotifier    nodegroupchange.NodeGroupChangeObserver
	nodeDeletionTracker   *deletiontracker.NodeDeletionTracker
	deletionsPerNodeGroup map[string][]*apiv1.Node
	deleteInterval        time.Duration
	drainedNodeDeletions  map[string]bool
}

// NewNodeDeletionBatcher return new NodeBatchDeleter
func NewNodeDeletionBatcher(ctx *context.AutoscalingContext, scaleStateNotifier nodegroupchange.NodeGroupChangeObserver, nodeDeletionTracker *deletiontracker.NodeDeletionTracker, deleteInterval time.Duration) *NodeDeletionBatcher {
	return &NodeDeletionBatcher{
		ctx:                   ctx,
		nodeDeletionTracker:   nodeDeletionTracker,
		deletionsPerNodeGroup: make(map[string][]*apiv1.Node),
		deleteInterval:        deleteInterval,
		drainedNodeDeletions:  make(map[string]bool),
		scaleStateNotifier:    scaleStateNotifier,
	}
}

// AddNodes adds node list to delete candidates and schedules deletion. The deletion is performed asynchronously.
func (d *NodeDeletionBatcher) AddNodes(nodes []*apiv1.Node, nodeGroup cloudprovider.NodeGroup, drain bool) {
	// If delete interval is 0, than instantly start node deletion.
	if d.deleteInterval == 0 {
		go d.deleteNodesAndRegisterStatus(nodes, nodeGroup.Id(), drain)
		return
	}
	first := d.addNodesToBucket(nodes, nodeGroup, drain)
	if first {
		// Just in case a node group implementation is not thread-safe, the async "remove" function will obtain a new instance of it to preform deletion.
		go func(nodeGroupId string) {
			time.Sleep(d.deleteInterval)
			d.remove(nodeGroupId)
		}(nodeGroup.Id())
	}
}

func (d *NodeDeletionBatcher) deleteNodesAndRegisterStatus(nodes []*apiv1.Node, nodeGroupId string, drain bool) {
	nodeGroup, err := deleteNodesFromCloudProvider(d.ctx, d.scaleStateNotifier, nodes)
	for _, node := range nodes {
		if err != nil {
			result := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: err}
			CleanUpAndRecordFailedScaleDownEvent(d.ctx, node, nodeGroupId, drain, d.nodeDeletionTracker, "", result)
		} else {
			RegisterAndRecordSuccessfulScaleDownEvent(d.ctx, d.scaleStateNotifier, node, nodeGroup, drain, d.nodeDeletionTracker)
		}
	}
}

// AddToBucket adds node to delete candidates and return if it's a first node in the group.
func (d *NodeDeletionBatcher) addNodesToBucket(nodes []*apiv1.Node, nodeGroup cloudprovider.NodeGroup, drain bool) bool {
	d.Lock()
	defer d.Unlock()
	for _, node := range nodes {
		d.drainedNodeDeletions[node.Name] = drain
	}
	val, ok := d.deletionsPerNodeGroup[nodeGroup.Id()]
	if !ok || len(val) == 0 {
		d.deletionsPerNodeGroup[nodeGroup.Id()] = nodes
		return true
	}
	d.deletionsPerNodeGroup[nodeGroup.Id()] = append(d.deletionsPerNodeGroup[nodeGroup.Id()], nodes...)
	return false
}

// remove deletes nodes of a given nodeGroup, if successful, the deletion is recorded in CSR, and an event is emitted on the node.
func (d *NodeDeletionBatcher) remove(nodeGroupId string) error {
	d.Lock()
	defer d.Unlock()
	nodes, ok := d.deletionsPerNodeGroup[nodeGroupId]
	if !ok {
		return fmt.Errorf("Node Group %s is not present in the batch deleter", nodeGroupId)
	}
	delete(d.deletionsPerNodeGroup, nodeGroupId)
	drainedNodeDeletions := make(map[string]bool)
	for _, node := range nodes {
		drainedNodeDeletions[node.Name] = d.drainedNodeDeletions[node.Name]
		delete(d.drainedNodeDeletions, node.Name)
	}

	go func(nodes []*apiv1.Node, drainedNodeDeletions map[string]bool) {
		var result status.NodeDeleteResult
		nodeGroup, err := deleteNodesFromCloudProvider(d.ctx, d.scaleStateNotifier, nodes)
		for _, node := range nodes {
			drain := drainedNodeDeletions[node.Name]
			if err != nil {
				result = status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: err}
				CleanUpAndRecordFailedScaleDownEvent(d.ctx, node, nodeGroupId, drain, d.nodeDeletionTracker, "", result)
			} else {
				RegisterAndRecordSuccessfulScaleDownEvent(d.ctx, d.scaleStateNotifier, node, nodeGroup, drain, d.nodeDeletionTracker)
			}
		}
	}(nodes, drainedNodeDeletions)
	return nil
}

// deleteNodeFromCloudProvider removes the given nodes from cloud provider. No extra pre-deletion actions are executed on
// the Kubernetes side.
func deleteNodesFromCloudProvider(ctx *context.AutoscalingContext, scaleStateNotifier nodegroupchange.NodeGroupChangeObserver, nodes []*apiv1.Node) (cloudprovider.NodeGroup, error) {
	nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(nodes[0])
	if err != nil {
		return nodeGroup, errors.NewAutoscalerError(errors.CloudProviderError, "failed to find node group for %s: %v", nodes[0].Name, err)
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		return nil, errors.NewAutoscalerError(errors.InternalError, "picked node that doesn't belong to a node group: %s", nodes[0].Name)
	}
	if err := nodeGroup.DeleteNodes(nodes); err != nil {
		scaleStateNotifier.RegisterFailedScaleDown(nodeGroup,
			string(errors.CloudProviderError),
			time.Now())
		return nodeGroup, errors.NewAutoscalerError(errors.CloudProviderError, "failed to delete nodes from group %s: %v", nodeGroup.Id(), err)
	}
	return nodeGroup, nil
}

func nodeScaleDownReason(node *apiv1.Node, drain bool) metrics.NodeScaleDownReason {
	readiness, err := kubernetes.GetNodeReadiness(node)
	if err != nil {
		klog.Errorf("Couldn't determine node %q readiness while scaling down - assuming unready: %v", node.Name, err)
		return metrics.Unready
	}
	if !readiness.Ready {
		return metrics.Unready
	}
	// Node is ready.
	if drain {
		return metrics.Underutilized
	}
	return metrics.Empty
}

// IsNodeBeingDeleted returns true iff a given node is being deleted.
func IsNodeBeingDeleted(ctx *context.AutoscalingContext, node *apiv1.Node, timestamp time.Time) bool {
	deleteTime, _ := taints.GetToBeDeletedTime(node)
	return deleteTime != nil && (timestamp.Sub(*deleteTime) < ctx.MaxCloudProviderNodeDeletionTime || timestamp.Sub(*deleteTime) < MaxKubernetesEmptyNodeDeletionTime)
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
func RegisterAndRecordSuccessfulScaleDownEvent(ctx *context.AutoscalingContext, scaleStateNotifier nodegroupchange.NodeGroupChangeObserver, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup, drain bool, nodeDeletionTracker *deletiontracker.NodeDeletionTracker) {
	ctx.Recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "nodes removed by cluster autoscaler")
	currentTime := time.Now()
	expectedDeleteTime := time.Now().Add(ctx.MaxCloudProviderNodeDeletionTime)
	scaleStateNotifier.RegisterScaleDown(nodeGroup, node.Name, currentTime, expectedDeleteTime)
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
