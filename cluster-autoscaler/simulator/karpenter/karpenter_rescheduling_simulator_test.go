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

func TestKarpenterReschedulingSimulator_TrySchedulePods_BreakOnFailure(t *testing.T) {
	snapshot := testsnapshot.NewTestSnapshotOrDie(t)

	// Node with 2 CPU
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

	// pod1 requires 1.5 CPU (fits on node1)
	pod1 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-1", Namespace: "default"},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{{
				Name: "c",
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("1.5")},
				},
			}},
		},
	}

	// pod2 requires 1 CPU (does not fit if pod1 is scheduled, node1 only has 0.5 CPU left)
	pod2 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-2", Namespace: "default"},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{{
				Name: "c",
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("1")},
				},
			}},
		},
	}

	// pod3 requires 0.1 CPU (would fit if pod2 was not there, or if we scheduled pod1 and pod3 together: 1.5 + 0.1 = 1.6 <= 2)
	pod3 := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-3", Namespace: "default"},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{{
				Name: "c",
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("0.1")},
				},
			}},
		},
	}

	err := snapshot.SetClusterState([]*apiv1.Node{node1}, nil, nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	simulator := NewKarpenterReschedulingSimulator()

	// Try to schedule [pod1, pod2, pod3] with breakOnFailure = true.
	// pod1 fits.
	// pod2 does not fit (failure!).
	// Simulation should STOP. pod3 should NOT be scheduled even though it would fit with pod1.
	statuses, _, err := simulator.TrySchedulePods(snapshot, []*apiv1.Pod{pod1, pod2, pod3}, true, clustersnapshot.SchedulingOptions{})
	assert.NoError(t, err)
	assert.Len(t, statuses, 1)
	assert.Equal(t, "pod-1", statuses[0].Pod.Name)

	// Also verify that only pod1 is committed to the snapshot!
	nodeInfo, err := snapshot.GetNodeInfo("node-1")
	assert.NoError(t, err)
	assert.Len(t, nodeInfo.Pods(), 1)
	assert.Equal(t, "pod-1", nodeInfo.Pods()[0].Pod.Name)
}
