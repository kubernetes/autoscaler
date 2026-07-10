/*
Copyright 2026 The Kubernetes Authors.

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

package karpenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
)

func TestKarpenterReschedulingSimulator_TrySchedulePods(t *testing.T) {
	// Create a test cluster snapshot
	snapshot := testsnapshot.NewTestSnapshotOrDie(t)

	// Define some nodes and pods
	node1 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-1",
			Labels: map[string]string{
				"kubernetes.io/hostname": "node-1",
			},
		},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("4Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	pod1 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "default",
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:  "container",
					Image: "nginx",
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("100m"),
							apiv1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
				},
			},
		},
	}

	err := snapshot.SetClusterState([]*apiv1.Node{node1}, nil, nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	simulator := NewKarpenterReschedulingSimulator()

	// 1. Try to schedule pod-1 on node-1
	statuses, overflowing, err := simulator.TrySchedulePods(snapshot, []*apiv1.Pod{pod1}, false, clustersnapshot.SchedulingOptions{})
	assert.NoError(t, err)
	assert.Equal(t, 0, overflowing)
	assert.Len(t, statuses, 1)
	assert.Equal(t, "pod-1", statuses[0].Pod.Name)
	assert.Equal(t, "node-1", statuses[0].NodeName)

	// 2. Try to schedule a pod that is too large to fit
	hugePod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "huge-pod",
			Namespace: "default",
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:  "container",
					Image: "nginx",
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU: resource.MustParse("10"),
						},
					},
				},
			},
		},
	}

	statuses, overflowing, err = simulator.TrySchedulePods(snapshot, []*apiv1.Pod{hugePod}, false, clustersnapshot.SchedulingOptions{})
	assert.NoError(t, err)
	assert.Len(t, statuses, 0)
	assert.Equal(t, 1, overflowing) // huge-pod fails to schedule, representing 1 overflowing controller
}
