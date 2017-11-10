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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/kubernetes/pkg/api"

	"github.com/golang/glog"
)

const (
	// ResourceNvidiaGPU is the name of the Nvidia GPU resource.
	ResourceNvidiaGPU = "nvidia.com/gpu"
	// GPULabel is the label added to nodes with GPU resource on GKE.
	GPULabel = "cloud.google.com/gke-accelerator"
)

// SetGPUAllocatableToCapacity allows us to tolerate the fact that nodes with
// GPUs can have allocatable set to 0 for multiple minutes after becoming ready
// Without this workaround, Cluster Autoscaler will trigger an unnecessary
// additional scale up before the node is fully operational.
// TODO: Remove this once we handle dynamically privisioned resources well.
func SetGPUAllocatableToCapacity(nodes []*apiv1.Node, cloudProvider cloudprovider.CloudProvider) []*apiv1.Node {
	templateCache := make(map[string]resource.Quantity)
	result := []*apiv1.Node{}
	for _, node := range nodes {
		newNode := node
		if _, found := node.Labels[GPULabel]; found {
			if gpuAllocatable, ok := node.Status.Allocatable[ResourceNvidiaGPU]; !ok || gpuAllocatable.IsZero() {
				var capacity resource.Quantity
				updateCapacity := false
				if gpuCapacity, found := node.Status.Capacity[ResourceNvidiaGPU]; found {
					capacity = gpuCapacity
					updateCapacity = true
				} else {
					// The node has a label suggesting it should have GPU, but it's not in capacity / allocatable
					// and we don't know how many GPUs are there. Let's just take it from node template for this group.
					// this is hacky and evil
					// like, really, really evil
					// voldemort level evil
					ng, err := cloudProvider.NodeGroupForNode(node)
					if err != nil {
						glog.Errorf("Failed to get node group for node when getting GPU estimation for %v: %v",
							node.ObjectMeta.Name, err)
					} else {
						ngId := ng.Id()
						if gpuCapacity, found := templateCache[ngId]; found {
							capacity = gpuCapacity
							updateCapacity = true
						} else {
							nodeInfo, err := ng.TemplateNodeInfo()
							if err != nil {
								glog.Errorf("Failed to build template for getting GPU estimation for node %v: %v",
									node.ObjectMeta.Name, err)
							} else if gpuCapacity, found := nodeInfo.Node().Status.Capacity[ResourceNvidiaGPU]; found {
								capacity = gpuCapacity
								templateCache[ngId] = gpuCapacity
								updateCapacity = true
							}
						}
					}
				}

				if updateCapacity {
					nodeCopy, err := api.Scheme.DeepCopy(node)
					if err != nil {
						glog.Errorf("Failed to make a copy of node %v", node.ObjectMeta.Name)
					} else {
						newNode = nodeCopy.(*apiv1.Node)
						newNode.Status.Allocatable[ResourceNvidiaGPU] = capacity.DeepCopy()
					}
				}
			}
		}
		result = append(result, newNode)
	}
	return result
}
