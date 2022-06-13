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
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/stretchr/testify/assert"
)

func makePods(cpuPerPod int64, memoryPerPod int64, hostport int32, podCount int) []*apiv1.Pod {
	pod := &apiv1.Pod{
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpuPerPod, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(memoryPerPod*units.MiB, resource.DecimalSI),
						},
					},
				},
			},
		},
	}
	if hostport != 0 {
		pod.Spec.Containers[0].Ports = []apiv1.ContainerPort{
			{
				HostPort: hostport,
			},
		}
	}
	pods := []*apiv1.Pod{}
	for i := 0; i < podCount; i++ {
		pods = append(pods, pod)
	}
	return pods
}

func TestBinpackingEstimate(t *testing.T) {
	testCases := []struct {
		name            string
		millicores      int64
		memory          int64
		maxNodes        int
		pods            []*apiv1.Pod
		expectNodeCount int
		expectPodCount  int
	}{
		{
			name:            "simple resource-based binpacking",
			millicores:      350*3 - 50,
			memory:          2 * 1000,
			pods:            makePods(350, 1000, 0, 10),
			expectNodeCount: 5,
			expectPodCount:  10,
		},
		{
			name:            "pods-per-node bound binpacking",
			millicores:      10000,
			memory:          20000,
			pods:            makePods(10, 100, 0, 20),
			expectNodeCount: 2,
			expectPodCount:  20,
		},
		{
			name:            "hostport conflict forces pod-per-node",
			millicores:      1000,
			memory:          5000,
			pods:            makePods(200, 1000, 5555, 8),
			expectNodeCount: 8,
			expectPodCount:  8,
		},
		{
			name:            "limiter cuts binpacking",
			millicores:      1000,
			memory:          5000,
			pods:            makePods(500, 1000, 0, 20),
			maxNodes:        5,
			expectNodeCount: 5,
			expectPodCount:  10,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limiter := NewFakeEstimationLimiter(tc.maxNodes)
			estimator := newBinPackingEstimator(t, limiter)
			node := &apiv1.Node{
				Status: apiv1.NodeStatus{
					Capacity: apiv1.ResourceList{
						apiv1.ResourceCPU:    *resource.NewMilliQuantity(tc.millicores, resource.DecimalSI),
						apiv1.ResourceMemory: *resource.NewQuantity(tc.memory*units.MiB, resource.DecimalSI),
						apiv1.ResourcePods:   *resource.NewQuantity(10, resource.DecimalSI),
					},
				},
			}
			node.Status.Allocatable = node.Status.Capacity
			SetNodeReadyState(node, true, time.Time{})

			nodeInfo := schedulerframework.NewNodeInfo()
			nodeInfo.SetNode(node)
			estimatedNodes, estimatedPods := estimator.Estimate(tc.pods, nodeInfo, nil)
			assert.Equal(t, tc.expectNodeCount, estimatedNodes)
			assert.Equal(t, tc.expectPodCount, len(estimatedPods))
		})
	}
}

func newBinPackingEstimator(t *testing.T, l EstimationLimiter) *BinpackingNodeEstimator {
	predicateChecker, err := simulator.NewTestPredicateChecker()
	clusterSnapshot := simulator.NewBasicClusterSnapshot()
	assert.NoError(t, err)
	estimator := NewBinpackingNodeEstimator(predicateChecker, clusterSnapshot, l)
	return estimator
}
