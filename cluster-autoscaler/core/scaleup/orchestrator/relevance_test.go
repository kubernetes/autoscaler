/*
Copyright 2024 The Kubernetes Authors.

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

package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestFilterRelevantNodeInfos(t *testing.T) {
	podLabel := map[string]string{"app": "foo"}
	pendingLabel := map[string]string{"app": "pending"}

	testCases := []struct {
		name          string
		pendingPods   []*apiv1.Pod
		allNodeInfos  []*framework.NodeInfo
		expectedNodes []string
	}{
		{
			name: "no affinities, node with pod is irrelevant",
			pendingPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 100),
			},
			allNodeInfos: []*framework.NodeInfo{
				framework.NewTestNodeInfo(BuildTestNode("n1", 1000, 1000), BuildTestPod("e1", 100, 100, WithLabels(podLabel))),
			},
			expectedNodes: []string{},
		},
		{
			name: "pod affinity (pending to existing)",
			pendingPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 100, WithPodAffinity(podLabel, "topology")),
			},
			allNodeInfos: []*framework.NodeInfo{
				framework.NewTestNodeInfo(BuildTestNode("n1", 1000, 1000), BuildTestPod("e1", 100, 100, WithLabels(podLabel))),
				framework.NewTestNodeInfo(BuildTestNode("n2", 1000, 1000)),
			},
			expectedNodes: []string{"n1"},
		},
		{
			name: "pod anti-affinity (pending to existing)",
			pendingPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 100, WithPodAntiAffinity(podLabel, "topology")),
			},
			allNodeInfos: []*framework.NodeInfo{
				framework.NewTestNodeInfo(BuildTestNode("n1", 1000, 1000), BuildTestPod("e1", 100, 100, WithLabels(podLabel))),
			},
			expectedNodes: []string{"n1"},
		},
		{
			name: "pod anti-affinity (existing to pending)",
			pendingPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 100, WithLabels(pendingLabel)),
			},
			allNodeInfos: []*framework.NodeInfo{
				framework.NewTestNodeInfo(BuildTestNode("n1", 1000, 1000), BuildTestPod("e1", 100, 100, WithPodAntiAffinity(pendingLabel, "topology"))),
			},
			expectedNodes: []string{"n1"},
		},
		{
			name: "self-matching hostname affinity (keeps one node)",
			pendingPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 100, WithLabels(podLabel), WithPodAffinity(podLabel, apiv1.LabelHostname)),
			},
			allNodeInfos: []*framework.NodeInfo{
				framework.NewTestNodeInfo(BuildTestNode("n1", 1000, 1000), BuildTestPod("e1", 100, 100, WithLabels(podLabel))),
				framework.NewTestNodeInfo(BuildTestNode("n2", 1000, 1000), BuildTestPod("e2", 100, 100, WithLabels(podLabel))),
			},
			expectedNodes: []string{"n1"}, // Only first one should be kept
		},
		{
			name: "TSC preservation",
			pendingPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 100, WithHardTopologySpreadConstraint(1, "topology", podLabel)),
			},
			allNodeInfos: []*framework.NodeInfo{
				framework.NewTestNodeInfo(BuildTestNode("n1", 1000, 1000, WithNodeLabel("topology", "zone1"))),
				framework.NewTestNodeInfo(BuildTestNode("n2", 1000, 1000, WithNodeLabel("topology", "zone2"))),
				framework.NewTestNodeInfo(BuildTestNode("n1-dup", 1000, 1000, WithNodeLabel("topology", "zone1"))),
			},
			expectedNodes: []string{"n1", "n2"}, // n1-dup has same signature as n1
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := FilterRelevantNodeInfos(tc.pendingPods, tc.allNodeInfos)
			gotNames := make([]string, 0, len(got))
			for _, ni := range got {
				gotNames = append(gotNames, ni.Node().Name)
			}
			assert.ElementsMatch(t, tc.expectedNodes, gotNames)
		})
	}
}

func TestApplyRelevanceFilter(t *testing.T) {
	podLabel := map[string]string{"app": "foo"}
	pendingPod := BuildTestPod("p1", 100, 100, WithPodAffinity(podLabel, "topology"))

	n1 := BuildTestNode("n1", 1000, 1000)
	ni1 := framework.NewTestNodeInfo(n1, BuildTestPod("e1", 100, 100, WithLabels(podLabel)))
	n2 := BuildTestNode("n2", 1000, 1000)
	ni2 := framework.NewTestNodeInfo(n2)

	snapshot := testsnapshot.NewTestSnapshotOrDie(t)
	snapshot.AddNodeInfo(ni1)
	snapshot.AddNodeInfo(ni2)

	revert := ApplyRelevanceFilter(snapshot, []*apiv1.Pod{pendingPod})

	// After filter, only n1 should be in snapshot
	allInfos, _ := snapshot.ListNodeInfos()
	assert.Equal(t, 1, len(allInfos))
	assert.Equal(t, "n1", allInfos[0].Node().Name)

	revert()

	// After revert, both should be back
	allInfos, _ = snapshot.ListNodeInfos()
	assert.Equal(t, 2, len(allInfos))
}
