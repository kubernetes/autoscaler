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

package clustersnapshot

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/dynamicresources"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/stretchr/testify/assert"
)

// TODO(DRA): Add DRA-specific tests.

var snapshots = map[string]func(fwHandle *framework.Handle) ClusterSnapshot{
	"basic": func(fwHandle *framework.Handle) ClusterSnapshot {
		return NewBasicClusterSnapshot(fwHandle, true)
	},
	"delta": func(fwHandle *framework.Handle) ClusterSnapshot {
		return NewDeltaClusterSnapshot(fwHandle, true)
	},
}

func nodeNames(nodes []*apiv1.Node) []string {
	names := make([]string, len(nodes), len(nodes))
	for i, node := range nodes {
		names[i] = node.Name
	}
	return names
}

func extractNodes(nodeInfos []*framework.NodeInfo) []*apiv1.Node {
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
	nodes, err := snapshot.ListNodeInfos()
	assert.NoError(t, err)
	var pods []*apiv1.Pod
	for _, nodeInfo := range nodes {
		for _, podInfo := range nodeInfo.Pods {
			pods = append(pods, podInfo.Pod)
		}
	}
	return snapshotState{extractNodes(nodes), pods}
}

func startSnapshot(t *testing.T, snapshotFactory func(fwHandle *framework.Handle) ClusterSnapshot, state snapshotState) ClusterSnapshot {
	snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))
	err := snapshot.Initialize(state.nodes, state.pods, dynamicresources.Snapshot{})
	assert.NoError(t, err)
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
			name: "add empty nodeInfo",
			op: func(snapshot ClusterSnapshot) {
				err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
				assert.NoError(t, err)
			},
			modifiedState: snapshotState{
				nodes: []*apiv1.Node{node},
			},
		},
		{
			name: "add nodeInfo",
			op: func(snapshot ClusterSnapshot) {
				err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node, pod))
				assert.NoError(t, err)
			},
			modifiedState: snapshotState{
				nodes: []*apiv1.Node{node},
				pods:  []*apiv1.Pod{pod},
			},
		},
		{
			name: "remove nodeInfo",
			state: snapshotState{
				nodes: []*apiv1.Node{node},
			},
			op: func(snapshot ClusterSnapshot) {
				err := snapshot.RemoveNodeInfo(node.Name)
				assert.NoError(t, err)
			},
		},
		{
			name: "remove nodeInfo, then add it back",
			state: snapshotState{
				nodes: []*apiv1.Node{node},
			},
			op: func(snapshot ClusterSnapshot) {
				err := snapshot.RemoveNodeInfo(node.Name)
				assert.NoError(t, err)

				err = snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
				assert.NoError(t, err)
			},
			modifiedState: snapshotState{
				nodes: []*apiv1.Node{node},
			},
		},
		{
			name: "add pod, then remove nodeInfo",
			state: snapshotState{
				nodes: []*apiv1.Node{node},
			},
			op: func(snapshot ClusterSnapshot) {
				err := snapshot.SchedulePod(pod, node.Name, nil)
				assert.NoError(t, err)
				err = snapshot.RemoveNodeInfo(node.Name)
				assert.NoError(t, err)
			},
		},
	}

	return testCases
}

func TestForking(t *testing.T) {
	testCases := validTestCases(t)
	node := BuildTestNode("specialNode-2", 10, 100)

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

				snapshot.Fork()

				tc.op(snapshot)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork & revert", name, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()

				tc.op(snapshot)

				snapshot.Revert()

				// Modifications should no longer be applied.
				compareStates(t, tc.state, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork & fork & revert & revert", name, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()
				tc.op(snapshot)
				snapshot.Fork()

				snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))

				snapshot.Revert()
				snapshot.Revert()

				// Modifications should no longer be applied.
				compareStates(t, tc.state, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork & commit", name, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()
				tc.op(snapshot)

				err := snapshot.Commit()
				assert.NoError(t, err)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s fork & fork & commit & revert", name, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()
				snapshot.Fork()
				tc.op(snapshot)

				err := snapshot.Commit()
				assert.NoError(t, err)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))

				snapshot.Revert()
				// Modifications should no longer be applied.
				compareStates(t, tc.state, getSnapshotState(t, snapshot))

			})
			t.Run(fmt.Sprintf("%s: %s fork & fork & revert & commit", name, tc.name), func(t *testing.T) {
				snapshot := startSnapshot(t, snapshotFactory, tc.state)

				snapshot.Fork()
				tc.op(snapshot)
				snapshot.Fork()
				snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
				snapshot.Revert()
				err := snapshot.Commit()
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

				snapshot.Fork()
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

	nodeCount := localRand.Intn(99) + 1
	podCount := localRand.Intn(999) + 1
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

				snapshot.Fork()

				for _, node := range extraNodes {
					err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
					assert.NoError(t, err)
				}

				for _, pod := range extraPods {
					err := snapshot.SchedulePod(pod, pod.Spec.NodeName, nil)
					assert.NoError(t, err)
				}

				compareStates(t, snapshotState{allNodes, allPods}, getSnapshotState(t, snapshot))

				snapshot.Clear()

				compareStates(t, snapshotState{}, getSnapshotState(t, snapshot))

				// Clear() should break out of forked state.
				snapshot.Fork()
			})
	}
}

func TestNode404(t *testing.T) {
	// Anything and everything that returns errNodeNotFound should be tested here.
	ops := []struct {
		name string
		op   func(ClusterSnapshot) error
	}{
		{"schedule pod", func(snapshot ClusterSnapshot) error {
			return snapshot.SchedulePod(BuildTestPod("p1", 0, 0), "node", nil)
		}},
		{"unschedule pod", func(snapshot ClusterSnapshot) error {
			return snapshot.UnschedulePod("default", "p1", "node")
		}},
		{"get node", func(snapshot ClusterSnapshot) error {
			_, err := snapshot.NodeInfos().Get("node")
			return err
		}},
		{"remove nodeInfo", func(snapshot ClusterSnapshot) error {
			return snapshot.RemoveNodeInfo("node")
		}},
	}

	for name, snapshotFactory := range snapshots {
		for _, op := range ops {
			t.Run(fmt.Sprintf("%s: %s empty", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))

					// Empty snapshot - shouldn't be able to operate on nodes that are not here.
					err := op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s fork", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))

					node := BuildTestNode("node", 10, 100)
					err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
					assert.NoError(t, err)

					snapshot.Fork()
					assert.NoError(t, err)

					err = snapshot.RemoveNodeInfo("node")
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
					snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))

					node := BuildTestNode("node", 10, 100)
					err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
					assert.NoError(t, err)

					err = snapshot.RemoveNodeInfo("node")
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
		{"add nodeInfo", func(snapshot ClusterSnapshot) error {
			return snapshot.AddNodeInfo(framework.NewTestNodeInfo(node, pod))
		}},
	}

	for name, snapshotFactory := range snapshots {
		for _, op := range ops {
			t.Run(fmt.Sprintf("%s: %s base", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))

					err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
					assert.NoError(t, err)

					// Node already in base.
					err = op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s base, forked", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))

					err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
					assert.NoError(t, err)

					snapshot.Fork()
					assert.NoError(t, err)

					// Node already in base, shouldn't be able to add in fork.
					err = op.op(snapshot)
					assert.Error(t, err)
				})

			t.Run(fmt.Sprintf("%s: %s fork", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))

					snapshot.Fork()

					err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
					assert.NoError(t, err)

					// Node already in fork.
					err = op.op(snapshot)
					assert.Error(t, err)
				})
			t.Run(fmt.Sprintf("%s: %s committed", name, op.name),
				func(t *testing.T) {
					snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))

					snapshot.Fork()

					err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
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

func TestPVCUsedByPods(t *testing.T) {
	node := BuildTestNode("node", 10, 10)
	pod1 := BuildTestPod("pod1", 10, 10)
	pod1.Spec.NodeName = node.Name
	pod1.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim1",
				},
			},
		},
	}
	pod2 := BuildTestPod("pod2", 10, 10)
	pod2.Spec.NodeName = node.Name
	pod2.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim1",
				},
			},
		},
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim2",
				},
			},
		},
	}
	nonPvcPod := BuildTestPod("pod3", 10, 10)
	nonPvcPod.Spec.NodeName = node.Name
	nonPvcPod.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				NFS: &apiv1.NFSVolumeSource{
					Server:   "",
					Path:     "",
					ReadOnly: false,
				},
			},
		},
	}
	testcase := []struct {
		desc              string
		node              *apiv1.Node
		pods              []*apiv1.Pod
		claimName         string
		exists            bool
		removePod         string
		existsAfterRemove bool
	}{
		{
			desc:      "pvc new pod with volume fetch",
			node:      node,
			pods:      []*apiv1.Pod{pod1},
			claimName: "claim1",
			exists:    true,
			removePod: "",
		},
		{
			desc:      "pvc new pod with incorrect volume fetch",
			node:      node,
			pods:      []*apiv1.Pod{pod1},
			claimName: "incorrect-claim",
			exists:    false,
			removePod: "",
		},
		{
			desc:      "new pod with non-pvc volume fetch",
			node:      node,
			pods:      []*apiv1.Pod{nonPvcPod},
			claimName: "incorrect-claim",
			exists:    false,
			removePod: "",
		},
		{
			desc:              "pvc new pod with delete volume fetch",
			node:              node,
			pods:              []*apiv1.Pod{pod1},
			claimName:         "claim1",
			exists:            true,
			removePod:         "pod1",
			existsAfterRemove: false,
		},
		{
			desc:              "pvc two pods with duplicated volume, delete one pod, fetch",
			node:              node,
			pods:              []*apiv1.Pod{pod1, pod2},
			claimName:         "claim1",
			exists:            true,
			removePod:         "pod1",
			existsAfterRemove: true,
		},
		{
			desc:              "pvc and non-pvc pods, fetch and delete non-pvc pod",
			node:              node,
			pods:              []*apiv1.Pod{pod1, nonPvcPod},
			claimName:         "claim1",
			exists:            true,
			removePod:         "pod3",
			existsAfterRemove: true,
		},
		{
			desc:              "pvc and non-pvc pods, delete pvc pod and fetch",
			node:              node,
			pods:              []*apiv1.Pod{pod1, nonPvcPod},
			claimName:         "claim1",
			exists:            true,
			removePod:         "pod1",
			existsAfterRemove: false,
		},
	}

	for snapshotName, snapshotFactory := range snapshots {
		for _, tc := range testcase {
			t.Run(fmt.Sprintf("%s with snapshot (%s)", tc.desc, snapshotName), func(t *testing.T) {
				snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))
				err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(tc.node, tc.pods...))
				assert.NoError(t, err)

				volumeExists := snapshot.StorageInfos().IsPVCUsedByPods(schedulerframework.GetNamespacedName("default", tc.claimName))
				assert.Equal(t, tc.exists, volumeExists)

				if tc.removePod != "" {
					err = snapshot.UnschedulePod("default", tc.removePod, "node")
					assert.NoError(t, err)

					volumeExists = snapshot.StorageInfos().IsPVCUsedByPods(schedulerframework.GetNamespacedName("default", tc.claimName))
					assert.Equal(t, tc.existsAfterRemove, volumeExists)
				}
			})
		}
	}
}

func TestPVCClearAndFork(t *testing.T) {
	node := BuildTestNode("node", 10, 10)
	pod1 := BuildTestPod("pod1", 10, 10)
	pod1.Spec.NodeName = node.Name
	pod1.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim1",
				},
			},
		},
	}
	pod2 := BuildTestPod("pod2", 10, 10)
	pod2.Spec.NodeName = node.Name
	pod2.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim1",
				},
			},
		},
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "claim2",
				},
			},
		},
	}
	nonPvcPod := BuildTestPod("pod3", 10, 10)
	nonPvcPod.Spec.NodeName = node.Name
	nonPvcPod.Spec.Volumes = []apiv1.Volume{
		{
			Name: "v1",
			VolumeSource: apiv1.VolumeSource{
				NFS: &apiv1.NFSVolumeSource{
					Server:   "",
					Path:     "",
					ReadOnly: false,
				},
			},
		},
	}

	for snapshotName, snapshotFactory := range snapshots {
		t.Run(fmt.Sprintf("fork and revert snapshot with pvc pods with snapshot: %s", snapshotName), func(t *testing.T) {
			snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))
			err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node, pod1))
			assert.NoError(t, err)
			volumeExists := snapshot.StorageInfos().IsPVCUsedByPods(schedulerframework.GetNamespacedName("default", "claim1"))
			assert.Equal(t, true, volumeExists)

			snapshot.Fork()
			assert.NoError(t, err)
			volumeExists = snapshot.StorageInfos().IsPVCUsedByPods(schedulerframework.GetNamespacedName("default", "claim1"))
			assert.Equal(t, true, volumeExists)

			err = snapshot.SchedulePod(pod2, "node", nil)
			assert.NoError(t, err)

			volumeExists = snapshot.StorageInfos().IsPVCUsedByPods(schedulerframework.GetNamespacedName("default", "claim2"))
			assert.Equal(t, true, volumeExists)

			snapshot.Revert()

			volumeExists = snapshot.StorageInfos().IsPVCUsedByPods(schedulerframework.GetNamespacedName("default", "claim2"))
			assert.Equal(t, false, volumeExists)

		})

		t.Run(fmt.Sprintf("clear snapshot with pvc pods with snapshot: %s", snapshotName), func(t *testing.T) {
			snapshot := snapshotFactory(framework.TestFrameworkHandleOrDie(t))
			err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node, pod1))
			assert.NoError(t, err)
			volumeExists := snapshot.StorageInfos().IsPVCUsedByPods(schedulerframework.GetNamespacedName("default", "claim1"))
			assert.Equal(t, true, volumeExists)

			snapshot.Clear()
			volumeExists = snapshot.StorageInfos().IsPVCUsedByPods(schedulerframework.GetNamespacedName("default", "claim1"))
			assert.Equal(t, false, volumeExists)

		})
	}
}

func TestWithForkedSnapshot(t *testing.T) {
	testCases := validTestCases(t)
	err := fmt.Errorf("some error")
	for name, snapshotFactory := range snapshots {
		for _, tc := range testCases {
			snapshot := startSnapshot(t, snapshotFactory, tc.state)
			successFunc := func() (bool, error) {
				tc.op(snapshot)
				return true, err
			}
			failedFunc := func() (bool, error) {
				tc.op(snapshot)
				return false, err
			}
			t.Run(fmt.Sprintf("%s: %s WithForkedSnapshot for failed function", name, tc.name), func(t *testing.T) {
				err1, err2 := WithForkedSnapshot(snapshot, failedFunc)
				assert.Error(t, err1)
				assert.NoError(t, err2)

				// Modifications should not be applied.
				compareStates(t, tc.state, getSnapshotState(t, snapshot))
			})
			t.Run(fmt.Sprintf("%s: %s WithForkedSnapshot for success function", name, tc.name), func(t *testing.T) {
				err1, err2 := WithForkedSnapshot(snapshot, successFunc)
				assert.Error(t, err1)
				assert.NoError(t, err2)

				// Modifications should be applied.
				compareStates(t, tc.modifiedState, getSnapshotState(t, snapshot))
			})
		}
	}
}
