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

package predicate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func BenchmarkAddNodeInfo(b *testing.B) {
	testCases := []int{1, 10, 100, 1000, 5000, 15000, 100000}

	for snapshotName, snapshotFactory := range snapshots {
		for _, tc := range testCases {
			nodes := clustersnapshot.CreateTestNodes(tc)
			clusterSnapshot, err := snapshotFactory()
			assert.NoError(b, err)
			b.ResetTimer()
			b.Run(fmt.Sprintf("%s: AddNodeInfo() %d", snapshotName, tc), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					assert.NoError(b, clusterSnapshot.SetClusterState(nil, nil, drasnapshot.Snapshot{}))
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
			nodes := clustersnapshot.CreateTestNodes(tc)
			clusterSnapshot, err := snapshotFactory()
			assert.NoError(b, err)
			err = clusterSnapshot.SetClusterState(nodes, nil, drasnapshot.Snapshot{})
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

func BenchmarkAddPods(b *testing.B) {
	testCases := []int{1, 10, 100, 1000, 5000, 15000}

	for snapshotName, snapshotFactory := range snapshots {
		for _, tc := range testCases {
			nodes := clustersnapshot.CreateTestNodes(tc)
			pods := clustersnapshot.CreateTestPods(tc * 30)
			clustersnapshot.AssignTestPodsToNodes(pods, nodes)
			clusterSnapshot, err := snapshotFactory()
			assert.NoError(b, err)
			err = clusterSnapshot.SetClusterState(nodes, nil, drasnapshot.Snapshot{})
			assert.NoError(b, err)
			b.ResetTimer()
			b.Run(fmt.Sprintf("%s: ForceAddPod() 30*%d", snapshotName, tc), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()

					err = clusterSnapshot.SetClusterState(nodes, nil, drasnapshot.Snapshot{})
					if err != nil {
						assert.NoError(b, err)
					}
					b.StartTimer()
					for _, pod := range pods {
						err = clusterSnapshot.ForceAddPod(pod, pod.Spec.NodeName)
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
			nodes := clustersnapshot.CreateTestNodes(ntc)
			for _, ptc := range podTestCases {
				pods := clustersnapshot.CreateTestPods(ntc * ptc)
				clustersnapshot.AssignTestPodsToNodes(pods, nodes)
				clusterSnapshot, err := snapshotFactory()
				assert.NoError(b, err)
				err = clusterSnapshot.SetClusterState(nodes, pods, drasnapshot.Snapshot{})
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
