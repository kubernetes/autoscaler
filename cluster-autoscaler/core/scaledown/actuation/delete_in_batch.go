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
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/klog/v2"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

const (
	// MaxKubernetesEmptyNodeDeletionTime is the maximum time needed by Kubernetes to delete an empty node.
	MaxKubernetesEmptyNodeDeletionTime = 3 * time.Minute
	// MaxCloudProviderNodeDeletionTime is the maximum time needed by cloud provider to delete a node.
	MaxCloudProviderNodeDeletionTime = 5 * time.Minute
)

// NodeDeletionBatcher batch scale down candidates for one node group and remove them.
type NodeDeletionBatcher struct {
	sync.Mutex
	ctx                   *context.AutoscalingContext
	clusterState          *clusterstate.ClusterStateRegistry
	nodeDeletionTracker   *deletiontracker.NodeDeletionTracker
	deletionsPerNodeGroup map[string][]*apiv1.Node
	deleteInterval        time.Duration
	drainedNodeDeletions  map[string]bool
}

// NewNodeDeletionBatcher return new NodeBatchDeleter
func NewNodeDeletionBatcher(ctx *context.AutoscalingContext, csr *clusterstate.ClusterStateRegistry, nodeDeletionTracker *deletiontracker.NodeDeletionTracker, deleteInterval time.Duration) *NodeDeletionBatcher {
	return &NodeDeletionBatcher{
		ctx:                   ctx,
		clusterState:          csr,
		nodeDeletionTracker:   nodeDeletionTracker,
		deletionsPerNodeGroup: make(map[string][]*apiv1.Node),
		deleteInterval:        deleteInterval,
		drainedNodeDeletions:  make(map[string]bool),
	}
}

// AddNode adds node to delete candidates and schedule deletion.
func (d *NodeDeletionBatcher) AddNode(node *apiv1.Node, drain bool) error {
	// If delete interval is 0, than instantly start node deletion.
	if d.deleteInterval == 0 {
		nodeGroup, err := deleteNodesFromCloudProvider(d.ctx, []*apiv1.Node{node})
		if err != nil {
			result := status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: err}
			CleanUpAndRecordFailedScaleDownEvent(d.ctx, node, nodeGroup.Id(), drain, d.nodeDeletionTracker, "", result)
		} else {
			RegisterAndRecordSuccessfulScaleDownEvent(d.ctx, d.clusterState, node, nodeGroup, drain, d.nodeDeletionTracker)
		}
		return nil
	}
	nodeGroupId, first, err := d.addNodeToBucket(node, drain)
	if err != nil {
		return err
	}
	if first {
		go func(nodeGroupId string) {
			time.Sleep(d.deleteInterval)
			d.remove(nodeGroupId)
		}(nodeGroupId)
	}
	return nil
}

// AddToBucket adds node to delete candidates and return if it's a first node in the group.
func (d *NodeDeletionBatcher) addNodeToBucket(node *apiv1.Node, drain bool) (string, bool, error) {
	d.Lock()
	defer d.Unlock()
	nodeGroup, err := d.ctx.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		return "", false, err
	}
	d.drainedNodeDeletions[node.Name] = drain
	val, ok := d.deletionsPerNodeGroup[nodeGroup.Id()]
	if !ok || len(val) == 0 {
		d.deletionsPerNodeGroup[nodeGroup.Id()] = []*apiv1.Node{node}
		return nodeGroup.Id(), true, nil
	}
	d.deletionsPerNodeGroup[nodeGroup.Id()] = append(d.deletionsPerNodeGroup[nodeGroup.Id()], node)
	return nodeGroup.Id(), false, nil
}

// remove delete nodes of a given nodeGroup, if successful, the deletion is recorded in CSR, and an event is emitted on the node.
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
		nodeGroup, err := deleteNodesFromCloudProvider(d.ctx, nodes)
		for _, node := range nodes {
			drain := drainedNodeDeletions[node.Name]
			if err != nil {
				result = status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: err}
				CleanUpAndRecordFailedScaleDownEvent(d.ctx, node, nodeGroup.Id(), drain, d.nodeDeletionTracker, "", result)
			} else {
				RegisterAndRecordSuccessfulScaleDownEvent(d.ctx, d.clusterState, node, nodeGroup, drain, d.nodeDeletionTracker)
			}

		}
	}(nodes, drainedNodeDeletions)
	return nil
}

// deleteNodeFromCloudProvider removes the given nodes from cloud provider. No extra pre-deletion actions are executed on
// the Kubernetes side.
func deleteNodesFromCloudProvider(ctx *context.AutoscalingContext, nodes []*apiv1.Node) (cloudprovider.NodeGroup, error) {
	nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(nodes[0])
	if err != nil {
		return nodeGroup, errors.NewAutoscalerError(errors.CloudProviderError, "failed to find node group for %s: %v", nodes[0].Name, err)
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		return nodeGroup, errors.NewAutoscalerError(errors.InternalError, "picked node that doesn't belong to a node group: %s", nodes[0].Name)
	}
	if err = nodeGroup.DeleteNodes(nodes); err != nil {
		return nodeGroup, errors.NewAutoscalerError(errors.CloudProviderError, "failed to delete %s: %v", nodes[0].Name, err)
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
func IsNodeBeingDeleted(node *apiv1.Node, timestamp time.Time) bool {
	deleteTime, _ := taints.GetToBeDeletedTime(node)
	return deleteTime != nil && (timestamp.Sub(*deleteTime) < MaxCloudProviderNodeDeletionTime || timestamp.Sub(*deleteTime) < MaxKubernetesEmptyNodeDeletionTime)
}
