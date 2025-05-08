/*
Copyright 2025 The Kubernetes Authors.

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

package customresources

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/resource/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

// DraCustomResourcesProcessor handles DRA custom resource. It assumes,
// that the DRA resources may not become allocatable immediately after the node creation.
type DraCustomResourcesProcessor struct {
}

// FilterOutNodesWithUnreadyResources removes nodes that should have DRA resource, but don't have
// it in allocatable from ready nodes list and updates their status to unready on all nodes list.
func (p *DraCustomResourcesProcessor) FilterOutNodesWithUnreadyResources(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyDraResources := make(map[string]*apiv1.Node)

	for _, node := range readyNodes {
		nodeResourcesSlices, nodeWithDraResources := context.DraSnapShot.NodeResourceSlices(node.Name)

		if nodeWithDraResources && nodeHasUnreadyDraResource(nodeResourcesSlices, node) {
			nodesWithUnreadyDraResources[node.Name] = kubernetes.GetUnreadyNodeCopy(node, kubernetes.ResourceUnready)
		} else {
			newReadyNodes = append(newReadyNodes, node)
		}
	}

	// Override any node with unready DRA resources with its "unready" copy
	for _, node := range allNodes {
		if newNode, found := nodesWithUnreadyDraResources[node.Name]; found {
			newAllNodes = append(newAllNodes, newNode)
		} else {
			newAllNodes = append(newAllNodes, node)
		}
	}
	return newAllNodes, newReadyNodes
}

func nodeHasUnreadyDraResource(nodeResourcesSlices []*v1beta1.ResourceSlice, node *apiv1.Node) bool {
	nodesAllocatableResources := make(map[string]resource.Quantity)

	for resourceName, resourceQuantity := range node.Status.Allocatable {
		nodesAllocatableResources[resourceName.String()] = resourceQuantity
	}

	for _, resourceSlice := range nodeResourcesSlices {
		allocatableResourceQuantity, found := nodesAllocatableResources[resourceSlice.Spec.Driver]
		if !found || allocatableResourceQuantity.IsZero() {
			return true
		}
	}

	return false
}

// GetNodeResourceTargets returns the resource targets for DRA resource slices, and empty list otherwise.
func (p *DraCustomResourcesProcessor) GetNodeResourceTargets(context *context.AutoscalingContext, node *apiv1.Node, _ cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError) {
	customResourceTargets := []CustomResourceTarget{}
	allocatableResources := make(map[string]bool)

	if context == nil {
		return customResourceTargets, nil
	}
	nodeResourcesSlices, nodeWithDraResources := context.DraSnapShot.NodeResourceSlices(node.Name)
	if !nodeWithDraResources {
		return customResourceTargets, nil
	}

	nodesAllocatableResources := make(map[string]resource.Quantity)
	for resourceName, resourceQuantity := range node.Status.Allocatable {
		nodesAllocatableResources[resourceName.String()] = resourceQuantity
	}

	for _, resourceSlice := range nodeResourcesSlices {
		allocatableResourceQuantity, found := nodesAllocatableResources[resourceSlice.Name]
		if found && !allocatableResourceQuantity.IsZero() {
			customResourceTargets = append(customResourceTargets, CustomResourceTarget{resourceSlice.Name, allocatableResourceQuantity.Value()})
			allocatableResources[resourceSlice.Name] = true
		}
	}
	return customResourceTargets, nil
}

// CleanUp cleans up processor's internal structures.
func (p *DraCustomResourcesProcessor) CleanUp() {
}
