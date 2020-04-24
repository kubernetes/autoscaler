/*
Copyright 2020 The Kubernetes Authors.

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

package simulator

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	"github.com/stretchr/testify/assert"
)

var snapshots = map[string]func() ClusterSnapshot{
	"basic": func() ClusterSnapshot { return NewBasicClusterSnapshot() },
	"delta": func() ClusterSnapshot { return NewDeltaClusterSnapshot() },
}

func nodeNames(nodes []*apiv1.Node) []string {
	names := make([]string, len(nodes), len(nodes))
	for i, node := range nodes {
		names[i] = node.Name
	}
	return names
}

func extractNodes(nodeInfos []*schedulerframework.NodeInfo) []*apiv1.Node {
	nodes := []*apiv1.Node{}
	for _, ni := range nodeInfos {
		nodes = append(nodes, ni.Node())
	}
	return nodes
}

type snapshotState struct {
	nodes []*apiv1.Node
	pods  []*apiv1.Pod
}

func compareStates(t *testing.T, a, b snapshotState) {
	assert.ElementsMatch(t, a.nodes, b.nodes)
	assert.ElementsMatch(t, a.pods, b.pods)
}

func getSnapshotState(t *testing.T, snapshot ClusterSnapshot) snapshotState {
	nodes, err := snapshot.NodeInfos().List()
	assert.NoError(t, err)
	var pods []*apiv1.Pod
	for _, nodeInfo := range nodes {
		for _, podInfo := range nodeInfo.Pods {
			pods = append(pods, podInfo.Pod)
		}
	}
	return snapshotState{extractNodes(nodes), pods}
}

func startSnapshot(t *testing.T, snapshotFactory func() ClusterSnapshot, state snapshotState) ClusterSnapshot {
	snapshot := snapshotFactory()
	err := snapshot.AddNodes(state.nodes)
	assert.NoError(t, err)
	for _, pod := range state.pods {
		err := snapshot.AddPod(pod, pod.Spec.NodeName)
		assert.NoError(t, err)
	}
	return snapshot
}

type modificationTestCase struct {
	name          string
	op            func(ClusterSnapshot)
	state         snapshotState
	modifiedState snapshotState
}

func validTestCases(t *testing.T) []modificationTestCase {
	node := BuildTestNode("specialNode", 10, 100)
	pod := BuildTestPod("specialPod", 1, 1)
	pod.Spec.NodeName = node.Name

	testCases := []modificationTestCase{
		{
			name: "add node",
			op: func(snapshot ClusterSnapshot) {
				err := snapshot.AddNode(node)
				assert.NoError(t, err)
			},
			modifiedState: snapshotState{
				nodes: []*apiv1.Node{node},
			},
		},
		{
			name: "add node with pods",
			op: func(snapshot ClusterSnapshot) {
				err := snapshot.AddNodeWithPods(node, []*apiv1.Pod{pod})
				assert.NoError(t, err)
			},
			modifiedState: snapshotState{
				nodes: []*apiv1.Node{node},
				pods:  []*apiv1.Pod{pod},
			},
		},
		{
			name: "remove node",
			state: snapshotState{
				nodes: []*apiv1.Node{node},
			},
			op: func(snapshot ClusterSnapshot) {
				err := snapshot.RemoveNode(node.Name)
				assert.NoError(t, err)
			},
		},
		{
			name: "remove node, then add it back",
			state: snapshotState{
				nodes: []*apiv1.Node{node},
			},
			op: func(snapshot ClusterSnapshot) {
				err := snapshot.RemoveNode(node.Name)
				assert.NoError(t, err)

				err = snapshot.AddNode(node)
				assert.NoError(t, err)
			},
			modifiedState: snapshotState{
				nodes: []*apiv1.Node{node},
			},
		},
		{
			name: "add pod, then remove node",
			state: snapshotState{
				nodes: []*apiv1.Node{node},
			},
			op: func(snapshot ClusterSnapshot) {
				err := snapshot.AddPod(pod, node.Name)
				assert.NoError(t, err)
				err = snapshot.RemoveNode(node.Name)
				assert.NoError(t, err)
			},
		},
	}

	return testCases
}

func TestForking(t *testing.T) {
	testCases := validTestCases(t)

	for name, snapshotFactory := range snapshots {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s: %s base", name, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)
				tc.op(snapshot)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork", name, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				err := snapshot.Fork()
				assert.NoError(t, err)

				tc.op(snapshot)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork & revert", name, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				err := snapshot.Fork()
				assert.NoError(t, err)

				tc.op(snapshot)

				err = snapshot.Revert()
				assert.NoError(t, err)

				// Modifications should no longer be applied.
				compareStates(t, tc.state, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork & commit", name, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				err := snapshot.Fork()
				assert.NoError(t, err)

				tc.op(snapshot)

				err = snapshot.Commit()
				assert.NoError(t, err)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s cache, fork & commit", name, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				// Allow caches to be build.
				_, err := snapshot.NodeInfos().List()
				assert.NoError(t, err)
				_, err = snapshot.NodeInfos().HavePodsWithAffinityList()
				assert.NoError(t, err)

				err = snapshot.Fork()
				assert.NoError(t, err)

				tc.op(snapshot)

				err = snapshot.Commit()
				assert.NoError(t, err)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
		}
	}
}

func TestClear(t *testing.T) {
	// Run with -count=1 to avoid caching.
	localRand := rand.New(rand.NewSource(time.Now().Unix()))

	nodeCount := localRand.Intn(100)
	podCount := localRand.Intn(1000)
	extraNodeCount := localRand.Intn(100)
	extraPodCount := localRand.Intn(1000)

	nodes := createTestNodes(nodeCount)
	pods := createTestPods(podCount)
	assignPodsToNodes(pods, nodes)

	state := snapshotState{nodes, pods}

	extraNodes := createTestNodesWithPrefix("extra", extraNodeCount)

	allNodes := make([]*apiv1.Node, len(nodes)+len(extraNodes), len(nodes)+len(extraNodes))
	copy(allNodes, nodes)
	copy(allNodes[len(nodes):], extraNodes)

	extraPods := createTestPodsWithPrefix("extra", extraPodCount)
	assignPodsToNodes(extraPods, allNodes)

	allPods := make([]*apiv1.Pod, len(pods)+len(extraPods), len(pods)+len(extraPods))
	copy(allPods, pods)
	copy(allPods[len(pods):], extraPods)

	for name, snapshotFactory := range snapshots {
		t.Run(fmt.Sprintf("%s: clear base %d nodes %d pods", name, nodeCount, podCount),
			func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, state)
				compareStates(t, state, getSnapshotState(t, snapshot))

				snapshot.Clear()

				compareStates(t, snapshotState{}, getSnapshotState(t, snapshot))
			})
		t.Run(fmt.Sprintf("%s: clear fork %d nodes %d pods %d extra nodes %d extra pods", name, nodeCount, podCount, extraNodeCount, extraPodCount),
			func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, state)
				compareStates(t, state, getSnapshotState(t, snapshot))

				err := snapshot.Fork()
				assert.NoError(t, err)

				err = snapshot.AddNodes(extraNodes)
				assert.NoError(t, err)

				for _, pod := range extraPods {
					err := snapshot.AddPod(pod, pod.Spec.NodeName)
					assert.NoError(t, err)
				}

				compareStates(t, snapshotState{allNodes, allPods}, getSnapshotState(t, snapshot))

				// Fork()ing twice is not allowed.
				err = snapshot.Fork()
				assert.Error(t, err)

				snapshot.Clear()

				compareStates(t, snapshotState{}, getSnapshotState(t, snapshot))

				// Clear() should break out of forked state.
				err = snapshot.Fork()
				assert.NoError(t, err)
			})
	}
}

func TestNode404(t *testing.T) {
	// Anything and everything that returns errNodeNotFound should be tested here.
	ops := []struct {
		name string
		op   func(ClusterSnapshot) error
	}{
		{"add pod", func(snapshot ClusterSnapshot) error {
			return snapshot.AddPod(BuildTestPod("p1", 0, 0), "node")
		}},
		{"remove pod", func(snapshot ClusterSnapshot) error {
			return snapshot.RemovePod("default", "p1", "node")
		}},
		{"get node", func(snapshot ClusterSnapshot) error {
			_, err := snapshot.NodeInfos().Get("node")
			return err
		}},
		{"remove node", func(snapshot ClusterSnapshot) error {
			return snapshot.RemoveNode("node")
		}},
	}

	for name, snapshotFactory := range snapshots {
		for _, op := range ops {
			t.Run(fmt.Sprintf("%s: %s empty", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory()

					// Empty snapshot - shouldn't be able to operate on nodes that are not here.
					err := op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s fork", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory()

					node := BuildTestNode("node", 10, 100)
					err := snapshot.AddNode(node)
					assert.NoError(t, err)

					err = snapshot.Fork()
					assert.NoError(t, err)

					err = snapshot.RemoveNode("node")
					assert.NoError(t, err)

					// Node deleted after fork - shouldn't be able to operate on it.
					err = op.op(snapshot)
					assert.Error(t, err)

					err = snapshot.Commit()
					assert.NoError(t, err)

					// Node deleted before commit - shouldn't be able to operate on it.
					err = op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s base", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory()

					node := BuildTestNode("node", 10, 100)
					err := snapshot.AddNode(node)
					assert.NoError(t, err)

					err = snapshot.RemoveNode("node")
					assert.NoError(t, err)

					// Node deleted from base - shouldn't be able to operate on it.
					err = op.op(snapshot)
					assert.Error(t, err)
				})
		}
	}
}

func TestNodeAlreadyExists(t *testing.T) {
	node := BuildTestNode("node", 10, 100)
	pod := BuildTestPod("pod", 1, 1)
	pod.Spec.NodeName = node.Name

	ops := []struct {
		name string
		op   func(ClusterSnapshot) error
	}{
		{"add node", func(snapshot ClusterSnapshot) error {
			return snapshot.AddNode(node)
		}},
		{"add node with pod", func(snapshot ClusterSnapshot) error {
			return snapshot.AddNodeWithPods(node, []*apiv1.Pod{pod})
		}},
	}

	for name, snapshotFactory := range snapshots {
		for _, op := range ops {
			t.Run(fmt.Sprintf("%s: %s base", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory()

					err := snapshot.AddNode(node)
					assert.NoError(t, err)

					// Node already in base.
					err = op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s base, forked", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory()

					err := snapshot.AddNode(node)
					assert.NoError(t, err)

					err = snapshot.Fork()
					assert.NoError(t, err)

					// Node already in base, shouldn't be able to add in fork.
					err = op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s fork", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory()

					err := snapshot.Fork()
					assert.NoError(t, err)

					err = snapshot.AddNode(node)
					assert.NoError(t, err)

					// Node already in fork.
					err = op.op(snapshot)
					assert.Error(t, err)
				})
			t.Run(fmt.Sprintf("%s: %s committed", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory()

					err := snapshot.Fork()
					assert.NoError(t, err)

					err = snapshot.AddNode(node)
					assert.NoError(t, err)

					err = snapshot.Commit()
					assert.NoError(t, err)

					// Node already in new base.
					err = op.op(snapshot)
					assert.Error(t, err)
				})
		}
	}
}
