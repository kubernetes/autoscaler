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
	resourceapi "k8s.io/api/resource/v1"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"
)

// DraCustomResourcesProcessor handles DRA custom resource. It assumes,
// that the DRA resources may not become allocatable immediately after the node creation.
type DraCustomResourcesProcessor struct {
}

// FilterOutNodesWithUnreadyResources removes nodes that should have DRA resource, but don't have
// it in allocatable from ready nodes list and updates their status to unready on all nodes list.
func (p *DraCustomResourcesProcessor) FilterOutNodesWithUnreadyResources(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node, draSnapshot *snapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyDraResources := make(map[string]*apiv1.Node)
	if draSnapshot == nil {
		klog.Warningf("Cannot filter out nodes with unready DRA resources. The DRA snapshot is nil. Processing will be skipped.")
		return allNodes, readyNodes
	}

	for _, node := range readyNodes {
		ng, err := context.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			klog.Warningf("Failed to get node group for node %s, Skipping DRA readiness check and keeping node in ready list. Error: %v", node.Name, err)
			continue
		}
		if ng == nil {
			newReadyNodes = append(newReadyNodes, node)
			continue
		}

		nodeInfo, err := ng.TemplateNodeInfo()
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			klog.Warningf("Failed to get template node info for node group %s with error: %v", ng.Id(), err)
			continue
		}

		nodeResourcesSlices, _ := draSnapshot.NodeResourceSlices(node.Name)
		if isEqualResourceSlices(nodeResourcesSlices, nodeInfo.LocalResourceSlices) {
			newReadyNodes = append(newReadyNodes, node)
		} else {
			nodesWithUnreadyDraResources[node.Name] = kubernetes.GetUnreadyNodeCopy(node, kubernetes.ResourceUnready)
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

type resourceSliceSpecs struct {
	driver string
	pool   string
}

func isEqualResourceSlices(nodeResourcesSlices []*resourceapi.ResourceSlice, templateResourcesSlices []*resourceapi.ResourceSlice) bool {
	tempSlicesByPools := getDevicesBySpecs(templateResourcesSlices)
	nodeSlicesByPools := getDevicesBySpecs(nodeResourcesSlices)

	for templSpecs, tempDevicesSet := range tempSlicesByPools {
		matched := false
		for nodeSpecs, nodeDevicesSet := range nodeSlicesByPools {
			if templSpecs.driver == nodeSpecs.driver && nodeDevicesSet.Equal(tempDevicesSet) {
				delete(nodeSlicesByPools, nodeSpecs)
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func getDevicesBySpecs(resourcesSlices []*resourceapi.ResourceSlice) map[resourceSliceSpecs]sets.Set[string] {
	slicesGroupedByPoolAndDriver := make(map[resourceSliceSpecs]sets.Set[string])
	for _, rs := range resourcesSlices {
		rsSpecs := resourceSliceSpecs{
			pool:   rs.Spec.Pool.Name,
			driver: rs.Spec.Driver,
		}
		slicesGroupedByPoolAndDriver[rsSpecs] = getResourceSliceDevicesSet(rs)
	}
	return slicesGroupedByPoolAndDriver
}

func getResourceSliceDevicesSet(resourcesSlice *resourceapi.ResourceSlice) sets.Set[string] {
	devices := sets.New[string]()
	for _, device := range resourcesSlice.Spec.Devices {
		devices.Insert(device.Name)
	}
	return devices
}

// GetNodeResourceTargets returns the resource targets for DRA resource slices, not implemented.
func (p *DraCustomResourcesProcessor) GetNodeResourceTargets(_ *context.AutoscalingContext, _ *apiv1.Node, _ cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError) {
	// TODO(DRA): Figure out resource limits for DRA here.
	return []CustomResourceTarget{}, nil
}

// CleanUp cleans up processor's internal structures.
func (p *DraCustomResourcesProcessor) CleanUp() {
}
