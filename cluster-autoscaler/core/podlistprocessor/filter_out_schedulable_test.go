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

package podlistprocessor

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestFilterOutSchedulable(t *testing.T) {
	node := buildReadyTestNode("node", 2000, 100)
	matchesAllNodes := func(*framework.NodeInfo) bool { return true }
	matchesNoNodes := func(*framework.NodeInfo) bool { return false }

	testCases := map[string]struct {
		nodesWithPods           map[*apiv1.Node][]*apiv1.Pod
		unschedulableCandidates []*apiv1.Pod
		expectedScheduledPods   []*apiv1.Pod
		expectedUnscheduledPods []*apiv1.Pod
		nodeFilter              func(*framework.NodeInfo) bool
	}{
		"single empty node, no pods": {
			nodesWithPods: map[*apiv1.Node][]*apiv1.Pod{node: {}},
			nodeFilter:    matchesAllNodes,
		},
		"single empty node, single schedulable pod": {
			nodesWithPods: map[*apiv1.Node][]*apiv1.Pod{node: {}},
			unschedulableCandidates: []*apiv1.Pod{
				BuildTestPod("pod", 500, 10),
			},
			expectedScheduledPods: []*apiv1.Pod{
				BuildTestPod("pod", 500, 10),
			},
			nodeFilter: matchesAllNodes,
		},
		"single empty node, many schedulable pods": {
			nodesWithPods: map[*apiv1.Node][]*apiv1.Pod{node: {}},
			unschedulableCandidates: []*apiv1.Pod{
				BuildTestPod("pod1", 200, 10),
				BuildTestPod("pod2", 500, 10),
				BuildTestPod("pod3", 800, 10),
			},
			expectedScheduledPods: []*apiv1.Pod{
				BuildTestPod("pod1", 200, 10),
				BuildTestPod("pod2", 500, 10),
				BuildTestPod("pod3", 800, 10),
			},
			nodeFilter: matchesAllNodes,
		},
		"single empty node, single unschedulable pod": {
			nodesWithPods: map[*apiv1.Node][]*apiv1.Pod{node: {}},
			unschedulableCandidates: []*apiv1.Pod{
				BuildTestPod("pod1", 3000, 10),
			},
			expectedUnscheduledPods: []*apiv1.Pod{
				BuildTestPod("pod1", 3000, 10),
			},
			nodeFilter: matchesAllNodes,
		},
		"single empty node, various pods": {
			nodesWithPods: map[*apiv1.Node][]*apiv1.Pod{node: {}},
			unschedulableCandidates: []*apiv1.Pod{
				BuildTestPod("pod1", 200, 10),
				BuildTestPod("pod2", 500, 10),
				BuildTestPod("pod3", 1800, 10),
			},
			expectedScheduledPods: []*apiv1.Pod{
				BuildTestPod("pod1", 200, 10),
				BuildTestPod("pod2", 500, 10),
			},
			expectedUnscheduledPods: []*apiv1.Pod{
				BuildTestPod("pod3", 1800, 10),
			},
			nodeFilter: matchesAllNodes,
		},
		"single empty node, some priority pods": {
			nodesWithPods: map[*apiv1.Node][]*apiv1.Pod{node: {}},
			unschedulableCandidates: []*apiv1.Pod{
				BuildTestPod("pod1", 200, 10),
				buildPriorityTestPod("pod2", 500, 10, 10),
				buildPriorityTestPod("pod3", 1800, 10, 20),
			},
			expectedScheduledPods: []*apiv1.Pod{
				buildPriorityTestPod("pod3", 1800, 10, 20),
				BuildTestPod("pod1", 200, 10),
			},
			expectedUnscheduledPods: []*apiv1.Pod{
				buildPriorityTestPod("pod2", 500, 10, 10),
			},
			nodeFilter: matchesAllNodes,
		},
		"non-empty node with a single pods scheduled": {
			nodesWithPods: map[*apiv1.Node][]*apiv1.Pod{
				node: {
					BuildTestPod("pod1", 500, 10),
				},
			},
			unschedulableCandidates: []*apiv1.Pod{
				BuildTestPod("pod2", 1000, 10),
				BuildTestPod("pod3", 300, 10),
				BuildTestPod("pod4", 300, 10),
			},
			expectedScheduledPods: []*apiv1.Pod{
				BuildTestPod("pod2", 1000, 10),
				BuildTestPod("pod3", 300, 10),
			},
			expectedUnscheduledPods: []*apiv1.Pod{
				BuildTestPod("pod4", 300, 10),
			},
			nodeFilter: matchesAllNodes,
		},
		"non-empty node with many pods scheduled": {
			nodesWithPods: map[*apiv1.Node][]*apiv1.Pod{
				node: {
					BuildTestPod("pod1", 500, 10),
					BuildTestPod("pod2", 1000, 10),
				},
			},
			unschedulableCandidates: []*apiv1.Pod{
				BuildTestPod("pod3", 1000, 10),
				BuildTestPod("pod4", 300, 10),
				BuildTestPod("pod5", 300, 10),
			},
			expectedScheduledPods: []*apiv1.Pod{
				BuildTestPod("pod4", 300, 10),
			},
			expectedUnscheduledPods: []*apiv1.Pod{
				BuildTestPod("pod3", 1000, 10),
				BuildTestPod("pod5", 300, 10),
			},
			nodeFilter: matchesAllNodes,
		},
		"single empty node, various pods, node should not be considered": {
			nodesWithPods: map[*apiv1.Node][]*apiv1.Pod{node: {}},
			unschedulableCandidates: []*apiv1.Pod{
				BuildTestPod("pod1", 200, 10),
				BuildTestPod("pod2", 500, 10),
				BuildTestPod("pod3", 1800, 10),
			},
			expectedScheduledPods: []*apiv1.Pod{},
			expectedUnscheduledPods: []*apiv1.Pod{
				BuildTestPod("pod1", 200, 10),
				BuildTestPod("pod2", 500, 10),
				BuildTestPod("pod3", 1800, 10),
			},
			nodeFilter: matchesNoNodes,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot()
			predicateChecker, err := predicatechecker.NewTestPredicateChecker()
			assert.NoError(t, err)

			var allExpectedScheduledPods []*apiv1.Pod
			allExpectedScheduledPods = append(allExpectedScheduledPods, tc.expectedScheduledPods...)

			for node, pods := range tc.nodesWithPods {
				for _, pod := range pods {
					pod.Spec.NodeName = node.Name
					allExpectedScheduledPods = append(allExpectedScheduledPods, pod)
				}
				err := clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(node, pods...))
				assert.NoError(t, err)
			}

			clusterSnapshot.Fork()

			processor := NewFilterOutSchedulablePodListProcessor(predicateChecker, tc.nodeFilter)
			unschedulablePods, err := processor.filterOutSchedulableByPacking(tc.unschedulableCandidates, clusterSnapshot)

			assert.NoError(t, err)
			assert.ElementsMatch(t, unschedulablePods, tc.expectedUnscheduledPods, "unschedulable pods differ")

			nodeInfos, err := clusterSnapshot.ListNodeInfos()
			assert.NoError(t, err)
			var scheduledPods []*apiv1.Pod
			for _, nodeInfo := range nodeInfos {
				for _, podInfo := range nodeInfo.Pods() {
					scheduledPods = append(scheduledPods, podInfo.Pod)
				}
			}
			assert.ElementsMatch(t, scheduledPods, allExpectedScheduledPods, "scheduled pods differ")
		})
	}
}

func BenchmarkFilterOutSchedulable(b *testing.B) {
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
	snapshots := map[string]func() clustersnapshot.ClusterSnapshot{
		"basic": func() clustersnapshot.ClusterSnapshot { return clustersnapshot.NewBasicClusterSnapshot() },
		"delta": func() clustersnapshot.ClusterSnapshot { return clustersnapshot.NewDeltaClusterSnapshot() },
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

				predicateChecker, err := predicatechecker.NewTestPredicateChecker()
				assert.NoError(b, err)

				clusterSnapshot := snapshotFactory()
				if err := clusterSnapshot.SetClusterState(nodes, scheduledPods); err != nil {
					assert.NoError(b, err)
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					processor := NewFilterOutSchedulablePodListProcessor(predicateChecker, scheduling.ScheduleAnywhere)
					if stillPending, err := processor.filterOutSchedulableByPacking(pendingPods, clusterSnapshot); err != nil {
						assert.NoError(b, err)
					} else if len(stillPending) < tc.pendingPods {
						assert.Equal(b, len(stillPending), tc.pendingPods)
					}
				}
			})
		}
	}
}

func buildReadyTestNode(name string, cpu, mem int64) *apiv1.Node {
	node := BuildTestNode(name, cpu, mem)
	SetNodeReadyState(node, true, time.Time{})
	return node
}

func buildPriorityTestPod(name string, cpu, mem int64, priority int32) *apiv1.Pod {
	pod := BuildTestPod(name, cpu, mem)
	pod.Spec.Priority = &priority
	return pod
}
