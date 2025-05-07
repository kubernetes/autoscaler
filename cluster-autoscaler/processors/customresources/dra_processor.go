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
	"k8s.io/apimachinery/pkg/util/sets"
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
		ng, err := context.CloudProvider.NodeGroupForNode(node)
		if err != nil || ng == nil {
			newReadyNodes = append(newReadyNodes, node)
			continue
		}

		nodeInfo, err := ng.TemplateNodeInfo()
		if err != nil {
			newReadyNodes = append(newReadyNodes, node)
			continue
		}

		nodeResourcesSlices, _ := context.ClusterSnapshot.DraSnapshot().NodeResourceSlices(node.Name)
		if isSupersetResourceSlices(nodeResourcesSlices, nodeInfo.LocalResourceSlices) {
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

type deviceDetails struct {
	driver     string
	pool       string
	deviceName string
}

func isSupersetResourceSlices(nodeResourcesSlices []*v1beta1.ResourceSlice, templateResourcesSlices []*v1beta1.ResourceSlice) bool {
	nodeDevices := getResourceSlicesDevicesSet(nodeResourcesSlices)
	templateDevices := getResourceSlicesDevicesSet(templateResourcesSlices)

	return nodeDevices.IsSuperset(templateDevices)
}

func getResourceSlicesDevicesSet(resourcesSlices []*v1beta1.ResourceSlice) sets.Set[deviceDetails] {
	devices := sets.New[deviceDetails]()
	for _, resourceSlice := range resourcesSlices {
		for _, device := range resourceSlice.Spec.Devices {
			devices.Insert(deviceDetails{driver: resourceSlice.Spec.Driver, pool: resourceSlice.Spec.Pool.Name, deviceName: device.Name})
		}
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
