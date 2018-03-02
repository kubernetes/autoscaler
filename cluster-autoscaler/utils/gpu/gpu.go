/*
Copyright 2017 The Kubernetes Authors.

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

package gpu

import (
	apiv1 "k8s.io/api/core/v1"

	"github.com/golang/glog"
)

const (
	// ResourceNvidiaGPU is the name of the Nvidia GPU resource.
	ResourceNvidiaGPU = "nvidia.com/gpu"
	// GPULabel is the label added to nodes with GPU resource on GKE.
	GPULabel = "cloud.google.com/gke-accelerator"
)

// FilterOutNodesWithUnreadyGpus removes nodes that should have GPU, but don't have it in allocatable
// from ready nodes list and updates their status to unready on all nodes list.
// This is a hack/workaround for nodes with GPU coming up without installed drivers, resulting
// in GPU missing from their allocatable and capacity.
func FilterOutNodesWithUnreadyGpus(allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyGpu := make(map[string]*apiv1.Node)
	for _, node := range readyNodes {
		isUnready := false
		_, hasGpuLabel := node.Labels[GPULabel]
		gpuAllocatable, hasGpuAllocatable := node.Status.Allocatable[ResourceNvidiaGPU]
		// We expect node to have GPU based on label, but it doesn't show up
		// on node object. Assume the node is still not fully started (installing
		// GPU drivers).
		if hasGpuLabel && (!hasGpuAllocatable || gpuAllocatable.IsZero()) {
			newNode, err := getUnreadyNodeCopy(node)
			if err != nil {
				glog.Errorf("Failed to override status of node %v with unready GPU: %v",
					node.Name, err)
			} else {
				glog.V(3).Infof("Overriding status of node %v, which seems to have unready GPU",
					node.Name)
				nodesWithUnreadyGpu[newNode.Name] = newNode
				isUnready = true
			}
		}
		if !isUnready {
			newReadyNodes = append(newReadyNodes, node)
		}
	}
	// Override any node with unready GPU with its "unready" copy
	for _, node := range allNodes {
		if newNode, found := nodesWithUnreadyGpu[node.Name]; found {
			newAllNodes = append(newAllNodes, newNode)
		} else {
			newAllNodes = append(newAllNodes, node)
		}
	}
	return newAllNodes, newReadyNodes
}

func getUnreadyNodeCopy(node *apiv1.Node) (*apiv1.Node, error) {
	newNode := node.DeepCopy()
	newReadyCondition := apiv1.NodeCondition{
		Type:               apiv1.NodeReady,
		Status:             apiv1.ConditionFalse,
		LastTransitionTime: node.CreationTimestamp,
	}
	newNodeConditions := []apiv1.NodeCondition{newReadyCondition}
	for _, condition := range newNode.Status.Conditions {
		if condition.Type != apiv1.NodeReady {
			newNodeConditions = append(newNodeConditions, condition)
		}
	}
	newNode.Status.Conditions = newNodeConditions
	return newNode, nil
}

// NodeHasGpu returns true if a given node has GPU hardware.
// The result will be true if there is hardware capability. It doesn't matter
// if the drivers are installed and GPU is ready to use.
func NodeHasGpu(node *apiv1.Node) bool {
	_, hasGpuLabel := node.Labels[GPULabel]
	gpuAllocatable, hasGpuAllocatable := node.Status.Allocatable[ResourceNvidiaGPU]
	return hasGpuLabel || (hasGpuAllocatable && !gpuAllocatable.IsZero())
}
