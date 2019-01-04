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

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"

	"github.com/stretchr/testify/assert"
)

func makePod(cpuPerPod, memoryPerPod int64) *apiv1.Pod {
	return &apiv1.Pod{
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
}

func TestEstimate(t *testing.T) {
	cpuPerPod := int64(500)
	memoryPerPod := int64(1000 * 1024 * 1024)
	pod := makePod(cpuPerPod, memoryPerPod)

	pods := []*apiv1.Pod{}
	for i := 0; i < 5; i++ {
		podCopy := *pod
		pods = append(pods, &podCopy)
	}

	node := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(3*cpuPerPod, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(2*memoryPerPod, resource.DecimalSI),
				apiv1.ResourcePods:   *resource.NewQuantity(10, resource.DecimalSI),
			},
		},
	}
	nodeInfo := schedulercache.NewNodeInfo()
	nodeInfo.SetNode(node)

	estimator := NewBasicNodeEstimator()
	estimate := estimator.Estimate(pods, nodeInfo, []*schedulercache.NodeInfo{})

	// Check result.
	assert.Equal(t, 3, estimate)

	// Check internal state of estimator.
	assert.Equal(t, int64(500*5), estimator.cpuSum.MilliValue())
	assert.Equal(t, int64(5*memoryPerPod), estimator.memorySum.Value())
	assert.Equal(t, 5, estimator.GetCount())
	assert.Contains(t, estimator.GetDebug(), "CPU")
}

func TestEstimateWithComing(t *testing.T) {
	cpuPerPod := int64(500)
	memoryPerPod := int64(1000 * 1024 * 1024)

	pod := makePod(cpuPerPod, memoryPerPod)
	pods := []*apiv1.Pod{}
	for i := 0; i < 5; i++ {
		podCopy := *pod
		pods = append(pods, &podCopy)
	}

	node := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(3*cpuPerPod, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(2*memoryPerPod, resource.DecimalSI),
				apiv1.ResourcePods:   *resource.NewQuantity(10, resource.DecimalSI),
			},
		},
	}
	node.Status.Allocatable = node.Status.Capacity
	nodeInfo := schedulercache.NewNodeInfo()
	nodeInfo.SetNode(node)

	estimator := NewBasicNodeEstimator()
	estimate := estimator.Estimate(pods, nodeInfo, []*schedulercache.NodeInfo{nodeInfo, nodeInfo})

	// Check result.
	assert.Equal(t, 1, estimate)

	// Check internal state of estimator.
	assert.Contains(t, estimator.GetDebug(), "CPU")
	assert.Equal(t, int64(500*5), estimator.cpuSum.MilliValue())
	assert.Equal(t, int64(5*memoryPerPod), estimator.memorySum.Value())
	assert.Equal(t, 5, estimator.GetCount())

}

func TestEstimateWithPorts(t *testing.T) {
	cpuPerPod := int64(500)
	memoryPerPod := int64(1000 * 1024 * 1024)

	pod := makePod(cpuPerPod, memoryPerPod)
	pod.Spec.Containers[0].Ports = []apiv1.ContainerPort{
		{
			HostPort: 5555,
		},
	}

	pods := []*apiv1.Pod{}
	for i := 0; i < 5; i++ {
		pods = append(pods, pod)
	}
	node := &apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(3*cpuPerPod, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(2*memoryPerPod, resource.DecimalSI),
				apiv1.ResourcePods:   *resource.NewQuantity(10, resource.DecimalSI),
			},
		},
	}
	nodeInfo := schedulercache.NewNodeInfo()
	nodeInfo.SetNode(node)

	estimator := NewBasicNodeEstimator()
	estimate := estimator.Estimate(pods, nodeInfo, []*schedulercache.NodeInfo{})
	assert.Contains(t, estimator.GetDebug(), "CPU")
	assert.Equal(t, 5, estimate)
}
