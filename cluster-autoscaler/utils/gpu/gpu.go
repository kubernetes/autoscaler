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
	"k8s.io/kubernetes/pkg/api"

	"github.com/golang/glog"
)

const (
	// ResourceNvidiaGPU is the name of the Nvidia GPU resource.
	ResourceNvidiaGPU = "nvidia.com/gpu"
)

// SetGPUAllocatableToCapacity allows us to tolerate the fact that nodes with
// GPUs can have allocatable set to 0 for multiple minutes after becoming ready
// Without this workaround, Cluster Autoscaler will trigger an unnecessary
// additional scale up before the node is fully operational.
// TODO: Remove this once we handle dynamically privisioned resources well.
func SetGPUAllocatableToCapacity(nodes []*apiv1.Node) []*apiv1.Node {
	result := []*apiv1.Node{}
	for _, node := range nodes {
		newNode := node
		if gpuCapacity, ok := node.Status.Capacity[ResourceNvidiaGPU]; ok {
			if gpuAllocatable, ok := node.Status.Allocatable[ResourceNvidiaGPU]; !ok || gpuAllocatable.IsZero() {
				nodeCopy, err := api.Scheme.DeepCopy(node)
				if err != nil {
					glog.Errorf("Failed to make a copy of node %v", node.ObjectMeta.Name)
				} else {
					newNode = nodeCopy.(*apiv1.Node)
					newNode.Status.Allocatable[ResourceNvidiaGPU] = gpuCapacity.DeepCopy()
				}
			}
		}
		result = append(result, newNode)
	}
	return result
}
