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

package core

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestFilterOutSchedulableByPacking(t *testing.T) {
	// TODO(scheduler_framework_integration) extend/cleanup the test
	// - add more nodes
	// - add better naming for pods/scenarios

	p1 := BuildTestPod("p1", 1500, 200000)
	p2_1 := BuildTestPod("p2_2", 3000, 200000)
	p2_2 := BuildTestPod("p2_2", 3000, 200000)
	p3_1 := BuildTestPod("p3_1", 300, 200000)
	p3_2 := BuildTestPod("p3_2", 300, 200000)

	scheduledPod1 := BuildTestPod("s1", 100, 200000)
	scheduledPod1.Spec.NodeName = "node1"
	scheduledPod2 := BuildTestPod("s2", 1500, 200000)
	scheduledPod2.Spec.NodeName = "node1"

	podWaitingForPreemption := BuildTestPod("w1", 1500, 200000)
	var priority100 int32 = 100
	podWaitingForPreemption.Spec.Priority = &priority100
	podWaitingForPreemption.Status.NominatedNodeName = "node1"

	p4 := BuildTestPod("p4", 1800, 200000)
	p4.Spec.Priority = &priority100

	node := BuildTestNode("node1", 2000, 2000000)
	SetNodeReadyState(node, true, time.Time{})

	tests := []struct {
		name                   string
		nodes                  []*apiv1.Node
		scheduledPods          []*apiv1.Pod
		pendingPods            []*apiv1.Pod
		expectedPendingPods    []*apiv1.Pod
		expectedPodsInSnapshot []*apiv1.Pod
	}{
		{
			name:                   "scenario 1",
			nodes:                  []*apiv1.Node{node},
			scheduledPods:          []*apiv1.Pod{scheduledPod1},
			pendingPods:            []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2},
			expectedPendingPods:    []*apiv1.Pod{p2_1, p2_2, p3_2},
			expectedPodsInSnapshot: []*apiv1.Pod{scheduledPod1, p1, p3_1},
		},
		{
			name:                   "scenario 2",
			nodes:                  []*apiv1.Node{node},
			scheduledPods:          []*apiv1.Pod{scheduledPod1, scheduledPod2},
			pendingPods:            []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2},
			expectedPendingPods:    []*apiv1.Pod{p1, p2_1, p2_2, p3_2},
			expectedPodsInSnapshot: []*apiv1.Pod{scheduledPod1, scheduledPod2, p3_1},
		},
		{
			name:                   "scenario 3",
			nodes:                  []*apiv1.Node{node},
			scheduledPods:          []*apiv1.Pod{scheduledPod1},
			pendingPods:            []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2, p4},
			expectedPendingPods:    []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2},
			expectedPodsInSnapshot: []*apiv1.Pod{scheduledPod1, p4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicateChecker, err := simulator.NewTestPredicateChecker()
			clusterSnapshot := simulator.NewBasicClusterSnapshot()

			for _, node := range tt.nodes {
				err := clusterSnapshot.AddNode(node)
				assert.NoError(t, err)
			}

			for _, pod := range tt.scheduledPods {
				err = clusterSnapshot.AddPod(pod, pod.Spec.NodeName)
				assert.NoError(t, err)
			}

			stillPendingPods, err := filterOutSchedulableByPacking(tt.pendingPods, clusterSnapshot, predicateChecker)
			assert.NoError(t, err)
			assert.ElementsMatch(t, stillPendingPods, tt.expectedPendingPods, "pending pods differ")

			// Check if snapshot was correctly modified
			podsInSnapshot, err := clusterSnapshot.Pods().List(labels.Everything())
			assert.NoError(t, err)
			assert.ElementsMatch(t, podsInSnapshot, tt.expectedPodsInSnapshot, "pods in snapshot differ")
		})
	}
}

func BenchmarkFilterOutSchedulableByPacking(b *testing.B) {
	// All pending pods in this scenario are unschedulable - predicates will fail.
	tests := []struct {
		name          string
		nodes         int
		scheduledPods int
		pendingPods   int
	}{
		{
			name:          "nothing",
			nodes:         1,
			scheduledPods: 30,
			pendingPods:   1000,
		},
		{
			name:          "small",
			nodes:         10,
			scheduledPods: 300,
			pendingPods:   1000,
		},
		{
			name:          "medium",
			nodes:         100,
			scheduledPods: 3000,
			pendingPods:   1000,
		},
		{
			name:          "large",
			nodes:         200,
			scheduledPods: 200,
			pendingPods:   60000,
		},
		{
			name:          "1k",
			nodes:         1000,
			scheduledPods: 1000,
			pendingPods:   12000,
		},
	}
	snapshots := map[string]func() simulator.ClusterSnapshot{
		"basic": func() simulator.ClusterSnapshot { return simulator.NewBasicClusterSnapshot() },
		"delta": func() simulator.ClusterSnapshot { return simulator.NewDeltaClusterSnapshot() },
	}
	for snapshotName, snapshotFactory := range snapshots {
		for _, tc := range tests {
			b.Run(fmt.Sprintf("%s: %d nodes %d scheduled %d pending", snapshotName, tc.nodes, tc.scheduledPods, tc.pendingPods), func(b *testing.B) {
				pendingPods := make([]*apiv1.Pod, tc.pendingPods, tc.pendingPods)
				for i := 0; i < tc.pendingPods; i++ {
					pendingPods[i] = BuildTestPod(fmt.Sprintf("p-%d", i), 1000, 2000000)
				}
				nodes := make([]*apiv1.Node, tc.nodes, tc.nodes)
				for i := 0; i < tc.nodes; i++ {
					nodes[i] = BuildTestNode(fmt.Sprintf("n-%d", i), 2000, 200000)
					SetNodeReadyState(nodes[i], true, time.Time{})
				}
				scheduledPods := make([]*apiv1.Pod, tc.scheduledPods, tc.scheduledPods)
				j := 0
				for i := 0; i < tc.scheduledPods; i++ {
					scheduledPods[i] = BuildTestPod(fmt.Sprintf("s-%d", i), 1000, 200000)
					scheduledPods[i].Spec.NodeName = nodes[j].Name
					j++
					if j >= tc.nodes {
						j = 0
					}
				}

				predicateChecker, err := simulator.NewTestPredicateChecker()
				assert.NoError(b, err)

				clusterSnapshot := snapshotFactory()
				if err := clusterSnapshot.AddNodes(nodes); err != nil {
					assert.NoError(b, err)
				}

				for _, pod := range scheduledPods {
					if err := clusterSnapshot.AddPod(pod, pod.Spec.NodeName); err != nil {
						assert.NoError(b, err)
					}
				}
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					if stillPending, err := filterOutSchedulableByPacking(pendingPods, clusterSnapshot, predicateChecker); err != nil {
						assert.NoError(b, err)
					} else if len(stillPending) < tc.pendingPods {
						assert.Equal(b, len(stillPending), tc.pendingPods)
					}
				}
			})
		}
	}
}
