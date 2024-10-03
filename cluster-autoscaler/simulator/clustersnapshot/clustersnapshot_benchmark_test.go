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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/dynamicresources"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"
)

// TODO(DRA): Add DRA-specific benchmarks.

func createTestNodesWithPrefix(prefix string, n int) []*apiv1.Node {
	nodes := make([]*apiv1.Node, n, n)
	for i := 0; i < n; i++ {
		nodes[i] = BuildTestNode(fmt.Sprintf("%s-%d", prefix, i), 2000, 2000000)
		SetNodeReadyState(nodes[i], true, time.Time{})
	}
	return nodes
}

func createTestNodes(n int) []*apiv1.Node {
	return createTestNodesWithPrefix("n", n)
}

func createTestPodsWithPrefix(prefix string, n int) []*apiv1.Pod {
	pods := make([]*apiv1.Pod, n, n)
	for i := 0; i < n; i++ {
		pods[i] = BuildTestPod(fmt.Sprintf("%s-%d", prefix, i), 1000, 2000000)
	}
	return pods
}

func createTestPods(n int) []*apiv1.Pod {
	return createTestPodsWithPrefix("p", n)
}

func assignPodsToNodes(pods []*apiv1.Pod, nodes []*apiv1.Node) {
	if len(nodes) == 0 {
		return
	}

	j := 0
	for i := 0; i < len(pods); i++ {
		if j >= len(nodes) {
			j = 0
		}
		pods[i].Spec.NodeName = nodes[j].Name
		j++
	}
}

func BenchmarkAddNodeInfo(b *testing.B) {
	testCases := []int{1, 10, 100, 1000, 5000, 15000, 100000}

	for snapshotName, snapshotFactory := range snapshots {
		for _, tc := range testCases {
			nodes := createTestNodes(tc)
			fwHandle, err := framework.TestFrameworkHandle()
			assert.NoError(b, err)
			clusterSnapshot := snapshotFactory(fwHandle)
			b.ResetTimer()
			b.Run(fmt.Sprintf("%s: AddNodeInfo() %d", snapshotName, tc), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					clusterSnapshot.Clear()
					b.StartTimer()
					for _, node := range nodes {
						err := clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
						if err != nil {
							assert.NoError(b, err)
						}
					}
				}
			})
		}
	}
}

func BenchmarkListNodeInfos(b *testing.B) {
	testCases := []int{1, 10, 100, 1000, 5000, 15000, 100000}

	for snapshotName, snapshotFactory := range snapshots {
		for _, tc := range testCases {
			nodes := createTestNodes(tc)
			fwHandle, err := framework.TestFrameworkHandle()
			assert.NoError(b, err)
			clusterSnapshot := snapshotFactory(fwHandle)
			err = clusterSnapshot.Initialize(nodes, nil, dynamicresources.Snapshot{})
			if err != nil {
				assert.NoError(b, err)
			}
			b.ResetTimer()
			b.Run(fmt.Sprintf("%s: List() %d", snapshotName, tc), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					nodeInfos, err := clusterSnapshot.NodeInfos().List()
					if err != nil {
						assert.NoError(b, err)
					}
					if len(nodeInfos) != tc {
						assert.Equal(b, len(nodeInfos), tc)
					}
				}
			})
		}
	}
}

func BenchmarkSchedulePods(b *testing.B) {
	testCases := []int{1, 10, 100, 1000, 5000, 15000}

	for snapshotName, snapshotFactory := range snapshots {
		for _, tc := range testCases {
			nodes := createTestNodes(tc)
			pods := createTestPods(tc * 30)
			assignPodsToNodes(pods, nodes)
			fwHandle, err := framework.TestFrameworkHandle()
			assert.NoError(b, err)
			clusterSnapshot := snapshotFactory(fwHandle)
			err = clusterSnapshot.Initialize(nodes, nil, dynamicresources.Snapshot{})
			assert.NoError(b, err)
			b.ResetTimer()
			b.Run(fmt.Sprintf("%s: SchedulePod() 30*%d", snapshotName, tc), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()

					err = clusterSnapshot.Initialize(nodes, nil, dynamicresources.Snapshot{})
					if err != nil {
						assert.NoError(b, err)
					}
					b.StartTimer()
					for _, pod := range pods {
						err = clusterSnapshot.SchedulePod(pod, pod.Spec.NodeName, nil)
						if err != nil {
							assert.NoError(b, err)
						}
					}
				}
			})
		}
	}
}

func BenchmarkForkAddRevert(b *testing.B) {
	nodeTestCases := []int{1, 10, 100, 1000, 5000, 15000, 100000}
	podTestCases := []int{0, 1, 30}

	for snapshotName, snapshotFactory := range snapshots {
		for _, ntc := range nodeTestCases {
			nodes := createTestNodes(ntc)
			for _, ptc := range podTestCases {
				pods := createTestPods(ntc * ptc)
				assignPodsToNodes(pods, nodes)
				fwHandle, err := framework.TestFrameworkHandle()
				assert.NoError(b, err)
				clusterSnapshot := snapshotFactory(fwHandle)
				err = clusterSnapshot.Initialize(nodes, pods, dynamicresources.Snapshot{})
				assert.NoError(b, err)
				tmpNode1 := BuildTestNode("tmp-1", 2000, 2000000)
				tmpNode2 := BuildTestNode("tmp-2", 2000, 2000000)
				b.ResetTimer()
				b.Run(fmt.Sprintf("%s: ForkAddRevert (%d nodes, %d pods)", snapshotName, ntc, ptc), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						clusterSnapshot.Fork()
						err = clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(tmpNode1))
						if err != nil {
							assert.NoError(b, err)
						}
						clusterSnapshot.Fork()
						err = clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(tmpNode2))
						if err != nil {
							assert.NoError(b, err)
						}
						clusterSnapshot.Revert()
						clusterSnapshot.Revert()
					}
				})
			}
		}
	}
}

func BenchmarkBuildNodeInfoList(b *testing.B) {
	testCases := []struct {
		nodeCount int
	}{
		{
			nodeCount: 1000,
		},
		{
			nodeCount: 5000,
		},
		{
			nodeCount: 15000,
		},
		{
			nodeCount: 100000,
		},
	}

	for _, tc := range testCases {
		b.Run(fmt.Sprintf("fork add 1000 to %d", tc.nodeCount), func(b *testing.B) {
			nodes := createTestNodes(tc.nodeCount + 1000)
			fwHandle, err := framework.TestFrameworkHandle()
			assert.NoError(b, err)
			snapshot := NewDeltaClusterSnapshot(fwHandle, true)
			if err := snapshot.Initialize(nodes[:tc.nodeCount], nil, dynamicresources.Snapshot{}); err != nil {
				assert.NoError(b, err)
			}
			snapshot.Fork()
			for _, node := range nodes[tc.nodeCount:] {
				if err := snapshot.AddNodeInfo(framework.NewTestNodeInfo(node)); err != nil {
					assert.NoError(b, err)
				}
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				list := snapshot.data.buildNodeInfoList()
				if len(list) != tc.nodeCount+1000 {
					assert.Equal(b, len(list), tc.nodeCount+1000)
				}
			}
		})
	}
	for _, tc := range testCases {
		b.Run(fmt.Sprintf("base %d", tc.nodeCount), func(b *testing.B) {
			nodes := createTestNodes(tc.nodeCount)
			fwHandle, err := framework.TestFrameworkHandle()
			assert.NoError(b, err)
			snapshot := NewDeltaClusterSnapshot(fwHandle, true)
			if err := snapshot.Initialize(nodes, nil, dynamicresources.Snapshot{}); err != nil {
				assert.NoError(b, err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				list := snapshot.data.buildNodeInfoList()
				if len(list) != tc.nodeCount {
					assert.Equal(b, len(list), tc.nodeCount)
				}
			}
		})
	}
}
