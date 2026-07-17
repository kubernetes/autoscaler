/*
Copyright The Kubernetes Authors.

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

package emptycandidates

import (
	"fmt"
	"sort"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func BenchmarkEmptySorting(b *testing.B) {
	nodeCount := 5000
	nodeInfos := make(map[string]*framework.NodeInfo)
	nodes := make([]*apiv1.Node, 0, nodeCount)

	for i := 0; i < nodeCount; i++ {
		name := fmt.Sprintf("node-%d", i)
		node := BuildTestNode(name, 1000, 1000)
		nodes = append(nodes, node)
		// Every other node is non-empty
		if i%2 == 0 {
			pod := BuildTestPod(fmt.Sprintf("pod-%d", i), 100, 100)
			nodeInfos[name] = framework.NewTestNodeInfo(node, pod)
		} else {
			nodeInfos[name] = framework.NewTestNodeInfo(node)
		}
	}

	niGetter := &testNodeInfoGetter{m: nodeInfos}
	deleteOptions := options.NodeDeleteOptions{
		SkipNodesWithSystemPods:           true,
		SkipNodesWithLocalStorage:         true,
		SkipNodesWithCustomControllerPods: true,
	}
	p := NewEmptySortingProcessor(niGetter, deleteOptions, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ResetState()
		nodesCopy := make([]*apiv1.Node, len(nodes))
		copy(nodesCopy, nodes)

		sort.SliceStable(nodesCopy, func(i, j int) bool {
			return p.ScaleDownEarlierThan(nodesCopy[i], nodesCopy[j])
		})
	}
}
