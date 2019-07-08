/*
Copyright 2016 The Kubernetes Authors.

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

package estimator

import (
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/stretchr/testify/assert"
)

func TestOldBinpackingEstimate(t *testing.T) {
	estimator := NewOldBinpackingNodeEstimator(simulator.NewTestPredicateChecker())

	cpuPerPod := int64(350)
	memoryPerPod := int64(1000 * units.MiB)
	pod := makePod(cpuPerPod, memoryPerPod)

	pods := make([]*apiv1.Pod, 0)
	for i := 0; i < 10; i++ {
		pods = append(pods, pod)
	}
	node := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpuPerPod*3-50, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(2*memoryPerPod, resource.DecimalSI),
				apiv1.ResourcePods:   *resource.NewQuantity(10, resource.DecimalSI),
			},
		},
	}
	node.Status.Allocatable = node.Status.Capacity
	SetNodeReadyState(node, true, time.Time{})

	nodeInfo := schedulernodeinfo.NewNodeInfo()
	nodeInfo.SetNode(node)
	estimate := estimator.Estimate(pods, nodeInfo, []*schedulernodeinfo.NodeInfo{})
	assert.Equal(t, 5, estimate)
}

func TestOldBinpackingEstimateComingNodes(t *testing.T) {
	estimator := NewOldBinpackingNodeEstimator(simulator.NewTestPredicateChecker())

	cpuPerPod := int64(350)
	memoryPerPod := int64(1000 * units.MiB)
	pod := makePod(cpuPerPod, memoryPerPod)

	pods := make([]*apiv1.Pod, 0)
	for i := 0; i < 10; i++ {
		pods = append(pods, pod)
	}
	node := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpuPerPod*3-50, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(2*memoryPerPod, resource.DecimalSI),
				apiv1.ResourcePods:   *resource.NewQuantity(10, resource.DecimalSI),
			},
		},
	}
	node.Status.Allocatable = node.Status.Capacity
	SetNodeReadyState(node, true, time.Time{})

	nodeInfo := schedulernodeinfo.NewNodeInfo()
	nodeInfo.SetNode(node)
	estimate := estimator.Estimate(pods, nodeInfo, []*schedulernodeinfo.NodeInfo{nodeInfo, nodeInfo})
	// 5 - 2 nodes that are coming.
	assert.Equal(t, 3, estimate)
}

func TestOldBinpackingEstimateWithPorts(t *testing.T) {
	estimator := NewOldBinpackingNodeEstimator(simulator.NewTestPredicateChecker())

	cpuPerPod := int64(200)
	memoryPerPod := int64(1000 * units.MiB)
	pod := makePod(cpuPerPod, memoryPerPod)
	pod.Spec.Containers[0].Ports = []apiv1.ContainerPort{
		{
			HostPort: 5555,
		},
	}
	pods := make([]*apiv1.Pod, 0)
	for i := 0; i < 8; i++ {
		pods = append(pods, pod)
	}
	node := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(5*cpuPerPod, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(5*memoryPerPod, resource.DecimalSI),
				apiv1.ResourcePods:   *resource.NewQuantity(10, resource.DecimalSI),
			},
		},
	}
	node.Status.Allocatable = node.Status.Capacity
	SetNodeReadyState(node, true, time.Time{})

	nodeInfo := schedulernodeinfo.NewNodeInfo()
	nodeInfo.SetNode(node)
	estimate := estimator.Estimate(pods, nodeInfo, []*schedulernodeinfo.NodeInfo{})
	assert.Equal(t, 8, estimate)
}
