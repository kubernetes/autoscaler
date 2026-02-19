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

package store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
)

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
			nodes := clustersnapshot.CreateTestNodes(tc.nodeCount + 1000)
			deltaStore := NewDeltaSnapshotStore(16)
			if err := deltaStore.SetClusterState(nodes[:tc.nodeCount], nil, nil, nil); err != nil {
				assert.NoError(b, err)
			}
			deltaStore.Fork()
			for _, node := range nodes[tc.nodeCount:] {
				schedNodeInfo := schedulerimpl.NewNodeInfo()
				schedNodeInfo.SetNode(node)
				if err := deltaStore.AddSchedulerNodeInfo(schedNodeInfo); err != nil {
					assert.NoError(b, err)
				}
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				list := deltaStore.data.buildNodeInfoList()
				assert.Equal(b, tc.nodeCount+1000, len(list))
			}
		})
	}
	for _, tc := range testCases {
		b.Run(fmt.Sprintf("base %d", tc.nodeCount), func(b *testing.B) {
			nodes := clustersnapshot.CreateTestNodes(tc.nodeCount)
			deltaStore := NewDeltaSnapshotStore(16)
			if err := deltaStore.SetClusterState(nodes, nil, nil, nil); err != nil {
				assert.NoError(b, err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				list := deltaStore.data.buildNodeInfoList()
				assert.Equal(b, tc.nodeCount, len(list))
			}
		})
	}
}
