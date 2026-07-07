/*
Copyright 2022 The Kubernetes Authors.

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

package scheduling

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestTrySchedulePods(t *testing.T) {
	testCases := []struct {
		desc            string
		nodes           []*apiv1.Node
		pods            []*apiv1.Pod
		newPods         []*apiv1.Pod
		hints           map[*apiv1.Pod]string
		acceptableNodes func(*framework.NodeInfo) bool
		wantStatuses    []Status
		wantErr         bool
		contextDeadline time.Duration
	}{
		{
			desc: "two new pods, two nodes",
			nodes: []*apiv1.Node{
				buildReadyNode("n1", 1000, 2000000),
				buildReadyNode("n2", 1000, 2000000),
			},
			pods: []*apiv1.Pod{
				buildScheduledPod("p1", 300, 500000, "n1"),
			},
			newPods: []*apiv1.Pod{
				BuildTestPod("p2", 800, 500000),
				BuildTestPod("p3", 500, 500000),
			},
			acceptableNodes: ScheduleAnywhere,
			wantStatuses: []Status{
				{Pod: BuildTestPod("p2", 800, 500000), NodeName: "n2"},
				{Pod: BuildTestPod("p3", 500, 500000), NodeName: "n1"},
			},
		},

		{
			desc: "hinted Node no longer in the cluster doesn't cause an error",
			nodes: []*apiv1.Node{
				buildReadyNode("n1", 1000, 2000000),
				buildReadyNode("n2", 1000, 2000000),
			},
			pods: []*apiv1.Pod{
				buildScheduledPod("p1", 300, 500000, "n1"),
			},
			newPods: []*apiv1.Pod{
				BuildTestPod("p2", 800, 500000),
				BuildTestPod("p3", 500, 500000),
			},
			hints:           map[*apiv1.Pod]string{BuildTestPod("p2", 800, 500000): "non-existing-node"},
			acceptableNodes: ScheduleAnywhere,
			wantStatuses: []Status{
				{Pod: BuildTestPod("p2", 800, 500000), NodeName: "n2"},
				{Pod: BuildTestPod("p3", 500, 500000), NodeName: "n1"},
			},
		},
		{
			desc: "three new pods, two nodes, no fit",
			nodes: []*apiv1.Node{
				buildReadyNode("n1", 1000, 2000000),
				buildReadyNode("n2", 1000, 2000000),
			},
			pods: []*apiv1.Pod{
				buildScheduledPod("p1", 300, 500000, "n1"),
			},
			newPods: []*apiv1.Pod{
				BuildTestPod("p2", 800, 500000),
				BuildTestPod("p3", 500, 500000),
				BuildTestPod("p4", 700, 500000),
			},
			acceptableNodes: ScheduleAnywhere,
			wantStatuses: []Status{
				{Pod: BuildTestPod("p2", 800, 500000), NodeName: "n2"},
				{Pod: BuildTestPod("p3", 500, 500000), NodeName: "n1"},
			},
		},
		{
			desc: "no new pods, two nodes",
			nodes: []*apiv1.Node{
				buildReadyNode("n1", 1000, 2000000),
				buildReadyNode("n2", 1000, 2000000),
			},
			pods: []*apiv1.Pod{
				buildScheduledPod("p1", 300, 500000, "n1"),
			},
			newPods:         []*apiv1.Pod{},
			acceptableNodes: ScheduleAnywhere,
		},
		{
			desc: "two nodes, but only one acceptable",
			nodes: []*apiv1.Node{
				buildReadyNode("n1", 1000, 2000000),
				buildReadyNode("n2", 1000, 2000000),
			},
			pods: []*apiv1.Pod{
				buildScheduledPod("p1", 300, 500000, "n1"),
			},
			newPods: []*apiv1.Pod{
				BuildTestPod("p2", 500, 500000),
				BuildTestPod("p3", 500, 500000),
			},
			acceptableNodes: singleNodeOk("n2"),
			wantStatuses: []Status{
				{Pod: BuildTestPod("p2", 500, 500000), NodeName: "n2"},
				{Pod: BuildTestPod("p3", 500, 500000), NodeName: "n2"},
			},
		},
		{
			desc: "two nodes, but only one acceptable, no fit",
			nodes: []*apiv1.Node{
				buildReadyNode("n1", 1000, 2000000),
				buildReadyNode("n2", 1000, 2000000),
			},
			pods: []*apiv1.Pod{
				buildScheduledPod("p1", 300, 500000, "n1"),
			},
			newPods: []*apiv1.Pod{
				BuildTestPod("p2", 500, 500000),
				BuildTestPod("p3", 500, 500000),
			},
			acceptableNodes: singleNodeOk("n1"),
			wantStatuses: []Status{
				{Pod: BuildTestPod("p2", 500, 500000), NodeName: "n1"},
			},
		},
		{
			desc: "hinted pods are prioritized",
			nodes: []*apiv1.Node{
				buildReadyNode("n1", 1000, 2000000),
			},
			newPods: []*apiv1.Pod{
				BuildTestPod("p1-high", 1100, 1000000, WithPodPriority(100)),
				BuildTestPod("p2-low-not-hinted", 800, 1000000, WithPodPriority(10)),
				BuildTestPod("p2-low-hinted", 800, 1000000, WithPodPriority(10)),
			},
			hints: map[*apiv1.Pod]string{
				BuildTestPod("p2-low-hinted", 800, 1000000, WithPodPriority(10)): "n1",
			},
			acceptableNodes: ScheduleAnywhere,
			wantStatuses: []Status{
				{Pod: BuildTestPod("p2-low-hinted", 800, 1000000, WithPodPriority(10)), NodeName: "n1"},
			},
		},
		{
			desc: "higher priority pods are prioritized",
			nodes: []*apiv1.Node{
				buildReadyNode("n1", 1000, 2000000),
				buildReadyNode("n2", 1000, 2000000),
				buildReadyNode("n3", 1500, 2000000),
			},
			newPods: []*apiv1.Pod{
				BuildTestPod("p1-low", 800, 1000000),
				BuildTestPod("p2-high", 800, 1000000, WithPodPriority(100)),
				BuildTestPod("p3-highest", 1100, 1000000, WithPodPriority(1000)),
			},
			hints: map[*apiv1.Pod]string{
				BuildTestPod("p1-low", 800, 1000000):                             "n1",
				BuildTestPod("p2-high", 800, 1000000, WithPodPriority(100)):      "n1",
				BuildTestPod("p3-highest", 1100, 1000000, WithPodPriority(1000)): "n1",
			},
			acceptableNodes: ScheduleAnywhere,
			wantStatuses: []Status{
				{Pod: BuildTestPod("p3-highest", 1100, 1000000, WithPodPriority(1000)), NodeName: "n3"},
				{Pod: BuildTestPod("p2-high", 800, 1000000, WithPodPriority(100)), NodeName: "n1"},
				{Pod: BuildTestPod("p1-low", 800, 1000000), NodeName: "n2"},
			},
		},
		{
			desc: "first schedule pods with working hints and next schedule pods without hints for the remaining capacity while preserving the input pod order",
			nodes: []*apiv1.Node{
				buildReadyNode("n1", 2100, 4000000),
			},
			newPods: []*apiv1.Pod{
				BuildTestPod("p1", 500, 1000000),
				BuildTestPod("p2", 500, 1000000),
				BuildTestPod("p3", 500, 1000000),
				BuildTestPod("p4", 500, 1000000),
				BuildTestPod("p5", 500, 1000000),
				BuildTestPod("p6", 500, 1000000),
			},
			hints: map[*apiv1.Pod]string{
				BuildTestPod("p2", 500, 1000000): "n1",
				BuildTestPod("p5", 500, 1000000): "n1",
				BuildTestPod("p6", 500, 1000000): "n2",
			},
			acceptableNodes: ScheduleAnywhere,
			wantStatuses: []Status{
				{Pod: BuildTestPod("p2", 500, 1000000), NodeName: "n1"},
				{Pod: BuildTestPod("p5", 500, 1000000), NodeName: "n1"},
				{Pod: BuildTestPod("p1", 500, 1000000), NodeName: "n1"},
				{Pod: BuildTestPod("p3", 500, 1000000), NodeName: "n1"},
			},
		},
		{
			desc: "no simulations when context exceeded",
			nodes: []*apiv1.Node{
				buildReadyNode("n1", 1000, 2000000),
			},
			newPods: []*apiv1.Pod{
				BuildTestPod("p1", 800, 500000),
			},
			acceptableNodes: ScheduleAnywhere,
			wantStatuses:    nil,
			contextDeadline: -2 * time.Minute,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			clusterSnapshot := testsnapshot.NewTestSnapshotOrDie(t)
			clustersnapshot.InitializeClusterSnapshotOrDie(t, clusterSnapshot, tc.nodes, tc.pods)
			s := NewHintingSimulator()

			for pod, nodeName := range tc.hints {
				s.hints.Set(HintKeyFromPod(pod), nodeName)
			}

			ctx := context.Background()
			var cancel context.CancelFunc
			if tc.contextDeadline.Abs() > 0 {
				ctx, cancel = context.WithTimeout(ctx, tc.contextDeadline)
				defer cancel()
			}

			schedulingResult, err := s.TrySchedulePods(ctx, clusterSnapshot, tc.newPods, false, clustersnapshot.SchedulingOptions{IsNodeAcceptable: tc.acceptableNodes})
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatuses, schedulingResult.Statuses)

			numScheduled := countPods(t, clusterSnapshot)
			assert.Equal(t, len(tc.pods)+len(tc.wantStatuses), numScheduled)
			s.DropOldHints()
			// Check if new hints match actually scheduled node names.
			for _, status := range tc.wantStatuses {
				hintedNode, found := s.hints.Get(HintKeyFromPod(status.Pod))
				assert.True(t, found)
				actualNode := nodeNameForPod(t, clusterSnapshot, status.Pod.Name)
				assert.Equal(t, hintedNode, actualNode)
			}
		})
	}
}

func TestPodSchedulesOnHintedNode(t *testing.T) {
	testCases := []struct {
		desc      string
		nodeNames []string
		podNodes  map[string]string
	}{
		{
			desc:      "single hint",
			nodeNames: []string{"n1", "n2", "n3"},
			podNodes:  map[string]string{"p1": "n2"},
		},
		{
			desc:      "all on one node",
			nodeNames: []string{"n1", "n2", "n3"},
			podNodes: map[string]string{
				"p1": "n2",
				"p2": "n2",
				"p3": "n2",
			},
		},
		{
			desc:      "spread across nodes",
			nodeNames: []string{"n1", "n2", "n3"},
			podNodes: map[string]string{
				"p1": "n1",
				"p2": "n2",
				"p3": "n3",
			},
		},
		{
			desc:      "lots of pods",
			nodeNames: []string{"n1", "n2", "n3"},
			podNodes: map[string]string{
				"p1": "n1",
				"p2": "n1",
				"p3": "n1",
				"p4": "n2",
				"p5": "n2",
				"p6": "n2",
				"p7": "n3",
				"p8": "n3",
				"p9": "n3",
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			clusterSnapshot := testsnapshot.NewTestSnapshotOrDie(t)
			nodes := make([]*apiv1.Node, 0, len(tc.nodeNames))
			for _, n := range tc.nodeNames {
				nodes = append(nodes, buildReadyNode(n, 9999, 9999))
			}
			clustersnapshot.InitializeClusterSnapshotOrDie(t, clusterSnapshot, nodes, []*apiv1.Pod{})
			pods := make([]*apiv1.Pod, 0, len(tc.podNodes))
			s := NewHintingSimulator()
			var expectedStatuses []Status
			for p, n := range tc.podNodes {
				pod := BuildTestPod(p, 1, 1)
				pods = append(pods, pod)
				s.hints.Set(HintKeyFromPod(pod), n)
				expectedStatuses = append(expectedStatuses, Status{Pod: pod, NodeName: n})
			}
			schedulingResult, err := s.TrySchedulePods(context.Background(), clusterSnapshot, pods, false, clustersnapshot.SchedulingOptions{})
			assert.NoError(t, err)
			assert.Equal(t, expectedStatuses, schedulingResult.Statuses)

			for p, hinted := range tc.podNodes {
				actual := nodeNameForPod(t, clusterSnapshot, p)
				assert.Equal(t, hinted, actual)
			}
		})
	}
}

func buildReadyNode(name string, cpu, mem int64) *apiv1.Node {
	n := BuildTestNode(name, cpu, mem)
	SetNodeReadyState(n, true, time.Time{})
	return n
}

func buildScheduledPod(name string, cpu, mem int64, nodeName string) *apiv1.Pod {
	p := BuildTestPod(name, cpu, mem)
	p.Spec.NodeName = nodeName
	return p
}

func countPods(t *testing.T, clusterSnapshot clustersnapshot.ClusterSnapshot) int {
	t.Helper()
	count := 0
	nis, err := clusterSnapshot.ListNodeInfos()
	assert.NoError(t, err)
	for _, ni := range nis {
		count += len(ni.Pods())
	}
	return count
}

func nodeNameForPod(t *testing.T, clusterSnapshot clustersnapshot.ClusterSnapshot, pod string) string {
	t.Helper()
	nis, err := clusterSnapshot.ListNodeInfos()
	assert.NoError(t, err)
	for _, ni := range nis {
		for _, pi := range ni.Pods() {
			if pi.Pod.Name == pod {
				return ni.Node().Name
			}
		}
	}
	return ""
}

func singleNodeOk(nodeName string) func(*framework.NodeInfo) bool {
	return func(nodeInfo *framework.NodeInfo) bool {
		return nodeName == nodeInfo.Node().Name
	}
}

func TestNewResult(t *testing.T) {
	p1 := BuildTestPod("p1", 100, 1000)
	p2 := BuildTestPod("p2", 100, 1000)
	p3 := BuildTestPod("p3", 100, 1000)
	p4 := BuildTestPod("p4", 100, 1000)

	testCases := []struct {
		desc                 string
		allPods              []*apiv1.Pod
		unschedulablePodMap  map[types.UID]bool
		statuses             []Status
		similarPodScheduling *SimilarPodsScheduling
		want                 Result
	}{
		{
			desc:                 "empty inputs",
			allPods:              []*apiv1.Pod{},
			unschedulablePodMap:  map[types.UID]bool{},
			statuses:             []Status{},
			similarPodScheduling: &SimilarPodsScheduling{},
			want:                 Result{[]Status{}, nil, 0},
		},
		{
			desc:                 "no overlap",
			allPods:              []*apiv1.Pod{p1, p2},
			unschedulablePodMap:  map[types.UID]bool{p3.UID: true},
			statuses:             []Status{{Pod: p4}},
			similarPodScheduling: &SimilarPodsScheduling{},
			want:                 Result{[]Status{{Pod: p4}}, []*apiv1.Pod{p1, p2}, 0},
		},
		{
			desc:                 "overlap with unschedulablePodMap",
			allPods:              []*apiv1.Pod{p1, p2, p3},
			unschedulablePodMap:  map[types.UID]bool{p2.UID: true, p3.UID: true},
			statuses:             []Status{},
			similarPodScheduling: &SimilarPodsScheduling{},
			want:                 Result{[]Status{}, []*apiv1.Pod{p1}, 0},
		},
		{
			desc:                 "overlap with statuses",
			allPods:              []*apiv1.Pod{p1, p2, p3},
			unschedulablePodMap:  map[types.UID]bool{},
			statuses:             []Status{{Pod: p1}, {Pod: p3}},
			similarPodScheduling: &SimilarPodsScheduling{},
			want:                 Result{[]Status{{Pod: p1}, {Pod: p3}}, []*apiv1.Pod{p2}, 0},
		},
		{
			desc:                 "overlap with both",
			allPods:              []*apiv1.Pod{p1, p2, p3, p4},
			unschedulablePodMap:  map[types.UID]bool{p2.UID: true},
			statuses:             []Status{{Pod: p4}},
			similarPodScheduling: &SimilarPodsScheduling{},
			want:                 Result{[]Status{{Pod: p4}}, []*apiv1.Pod{p1, p3}, 0},
		},
		{
			desc:                 "pods in maps but not in allPods",
			allPods:              []*apiv1.Pod{p1},
			unschedulablePodMap:  map[types.UID]bool{p2.UID: true},
			statuses:             []Status{{Pod: p3}},
			similarPodScheduling: &SimilarPodsScheduling{},
			want:                 Result{[]Status{{Pod: p3}}, []*apiv1.Pod{p1}, 0},
		},
		{
			desc:                 "all pods are either schedulable or unschedulable",
			allPods:              []*apiv1.Pod{p1, p2, p3, p4},
			unschedulablePodMap:  map[types.UID]bool{p1.UID: true, p2.UID: true, p3.UID: true},
			statuses:             []Status{{Pod: p4}},
			similarPodScheduling: &SimilarPodsScheduling{},
			want:                 Result{[]Status{{Pod: p4}}, nil, 0},
		},
		{
			desc:                 "some pods are unprocessed",
			allPods:              []*apiv1.Pod{p1, p2, p3, p4},
			unschedulablePodMap:  map[types.UID]bool{p1.UID: true, p2.UID: true},
			statuses:             []Status{{Pod: p4}},
			similarPodScheduling: &SimilarPodsScheduling{overflowingControllers: map[string]bool{"c1": true, "c2": true}},
			want:                 Result{[]Status{{Pod: p4}}, []*apiv1.Pod{p3}, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := newResult(tc.allPods, tc.unschedulablePodMap, tc.statuses, tc.similarPodScheduling)
			assert.Equal(t, tc.want, got)
		})
	}
}
