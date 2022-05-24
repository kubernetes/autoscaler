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
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"

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

// DeleteNodeFromCloudProvider removes the given node from cloud provider. No extra pre-deletion actions are executed on
// the Kubernetes side. If successful, the deletion is recorded in CSR, and an event is emitted on the node.
func DeleteNodeFromCloudProvider(ctx *context.AutoscalingContext, node *apiv1.Node, registry *clusterstate.ClusterStateRegistry) errors.AutoscalerError {
	nodeGroup, err := ctx.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		return errors.NewAutoscalerError(errors.CloudProviderError, "failed to find node group for %s: %v", node.Name, err)
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		return errors.NewAutoscalerError(errors.InternalError, "picked node that doesn't belong to a node group: %s", node.Name)
	}
	if err = nodeGroup.DeleteNodes([]*apiv1.Node{node}); err != nil {
		return errors.NewAutoscalerError(errors.CloudProviderError, "failed to delete %s: %v", node.Name, err)
	}
	ctx.Recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "node removed by cluster autoscaler")
	registry.RegisterScaleDown(&clusterstate.ScaleDownRequest{
		NodeGroup:          nodeGroup,
		NodeName:           node.Name,
		Time:               time.Now(),
		ExpectedDeleteTime: time.Now().Add(MaxCloudProviderNodeDeletionTime),
	})
	return nil
}

// IsNodeBeingDeleted returns true iff a given node is being deleted.
func IsNodeBeingDeleted(node *apiv1.Node, timestamp time.Time) bool {
	deleteTime, _ := deletetaint.GetToBeDeletedTime(node)
	return deleteTime != nil && (timestamp.Sub(*deleteTime) < MaxCloudProviderNodeDeletionTime || timestamp.Sub(*deleteTime) < MaxKubernetesEmptyNodeDeletionTime)
}
