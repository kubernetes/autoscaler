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
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestSetGPUAllocatableToCapacity(t *testing.T) {
	nodeGPU := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "nodeGpu"}, Status: apiv1.NodeStatus{Capacity: apiv1.ResourceList{}, Allocatable: apiv1.ResourceList{}}}
	nodeGPU.Status.Allocatable[ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeGPU.Status.Capacity[ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeGPUUnready := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "nodeGpuUnready"}, Status: apiv1.NodeStatus{Capacity: apiv1.ResourceList{}, Allocatable: apiv1.ResourceList{}}}
	nodeGPUUnready.Status.Allocatable[ResourceNvidiaGPU] = *resource.NewQuantity(0, resource.DecimalSI)
	nodeGPUUnready.Status.Capacity[ResourceNvidiaGPU] = *resource.NewQuantity(2, resource.DecimalSI)
	nodeGPUNoAllocatable := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "nodeGpuNoAllocatable"}, Status: apiv1.NodeStatus{Capacity: apiv1.ResourceList{}, Allocatable: apiv1.ResourceList{}}}
	nodeGPUNoAllocatable.Status.Capacity[ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeNoGPU := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "nodeGpuUnready"}, Status: apiv1.NodeStatus{Capacity: apiv1.ResourceList{}, Allocatable: apiv1.ResourceList{}}}
	nodeNoGPU.Status.Allocatable[apiv1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeNoGPU.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)
	result := SetGPUAllocatableToCapacity([]*apiv1.Node{nodeGPU, nodeGPUUnready, nodeGPUNoAllocatable, nodeNoGPU})
	assertAllocatableAndCapacity(t, ResourceNvidiaGPU, 1, 1, result[0])
	assertAllocatableAndCapacity(t, ResourceNvidiaGPU, 2, 2, result[1])
	assertAllocatableAndCapacity(t, ResourceNvidiaGPU, 1, 1, result[2])
	assertAllocatableAndCapacity(t, apiv1.ResourceCPU, 1, 2, result[3])
}

func assertAllocatableAndCapacity(t *testing.T, resourceName apiv1.ResourceName, allocatable, capacity int64, node *apiv1.Node) {
	allocatableResource := *resource.NewQuantity(allocatable, resource.DecimalSI)
	capacityResource := *resource.NewQuantity(capacity, resource.DecimalSI)
	assert.Equal(t, node.Status.Allocatable[resourceName], allocatableResource,
		"Node %v, expected allocatable %v: %v got: %v", node.ObjectMeta.Name, resourceName, node.Status.Allocatable[resourceName], allocatableResource)
	assert.Equal(t, node.Status.Capacity[resourceName], capacityResource,
		"Node %v, expected capacity %v: %v got: %v", node.ObjectMeta.Name, resourceName, node.Status.Capacity[resourceName], capacityResource)
}
