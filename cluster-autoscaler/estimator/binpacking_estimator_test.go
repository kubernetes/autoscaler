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

	"k8s.io/contrib/cluster-autoscaler/simulator"
	"k8s.io/kubernetes/pkg/api/resource"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/stretchr/testify/assert"
)

func TestBinpackingEstimate(t *testing.T) {
	estimator := NewBinpackingNodeEstimator(simulator.NewTestPredicateChecker())

	cpuPerPod := int64(350)
	memoryPerPod := int64(1000 * 1024 * 1024)
	pod := &apiv1.Pod{
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpuPerPod, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(memoryPerPod, resource.DecimalSI),
						},
					},
				},
			},
		},
	}

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

	nodeInfo := schedulercache.NewNodeInfo()
	nodeInfo.SetNode(node)
	estimate := estimator.Estimate(pods, nodeInfo)
	assert.Equal(t, 5, estimate)
}

func TestBinpackingEstimateWithPorts(t *testing.T) {
	estimator := NewBinpackingNodeEstimator(simulator.NewTestPredicateChecker())

	cpuPerPod := int64(200)
	memoryPerPod := int64(1000 * 1024 * 1024)
	pod := &apiv1.Pod{
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpuPerPod, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(memoryPerPod, resource.DecimalSI),
						},
					},
					Ports: []apiv1.ContainerPort{
						{
							HostPort: 5555,
						},
					},
				},
			},
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

	nodeInfo := schedulercache.NewNodeInfo()
	nodeInfo.SetNode(node)
	estimate := estimator.Estimate(pods, nodeInfo)
	assert.Equal(t, 8, estimate)
}
