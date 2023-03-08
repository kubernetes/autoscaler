/*
Copyright 2023 The Kubernetes Authors.

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

package scaledowncandidates

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiv1 "k8s.io/api/core/v1"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

type scoreProcessor struct {
	scores []int
}

func (p *scoreProcessor) ScaleDownEarlierThan(node1, node2 *apiv1.Node) bool {
	idx1, _ := strconv.Atoi(node1.Name[5:])
	idx2, _ := strconv.Atoi(node2.Name[5:])
	return p.scores[idx1] > p.scores[idx2]
}

func TestSort(t *testing.T) {
	testCases := []struct {
		name          string
		numNodes      int
		numProcessors int
		nodeScores    [][]int //2d array that represent the score of node for each processor
		sortedOrder   []int
	}{
		{
			name:        "No score, order the same",
			numNodes:    5,
			sortedOrder: []int{0, 1, 2, 3, 4},
		},
		{
			name:          "One processor, the order has changed",
			numNodes:      5,
			numProcessors: 1,
			nodeScores:    [][]int{{3}, {4}, {1}, {2}, {0}},
			sortedOrder:   []int{1, 0, 3, 2, 4},
		},
		{
			name:          "Two processors, second processor did not affect the order",
			numNodes:      5,
			numProcessors: 2,
			nodeScores:    [][]int{{3, 5}, {4, 0}, {1, 2}, {2, 4}, {0, 5}},
			sortedOrder:   []int{1, 0, 3, 2, 4},
		},
		{
			name:          "Two processors, the first processor has equal scores",
			numNodes:      5,
			numProcessors: 2,
			nodeScores:    [][]int{{4, 5}, {4, 0}, {1, 2}, {2, 4}, {0, 5}},
			sortedOrder:   []int{0, 1, 3, 2, 4},
		},
		{
			name:          "Three processors, all three processors affected the order",
			numNodes:      5,
			numProcessors: 3,
			nodeScores:    [][]int{{1, 1, 1}, {1, 1, 2}, {1, 1, 3}, {1, 2, 1}, {2, 1, 1}},
			sortedOrder:   []int{4, 3, 2, 1, 0},
		},
	}
	for _, test := range testCases {
		nodes := []*apiv1.Node{}
		for i := 0; i < test.numNodes; i++ {
			node := BuildTestNode(fmt.Sprintf("node-%d", i), 10, 100)
			nodes = append(nodes, node)
		}
		processors := []CandidatesComparer{}
		for i := 0; i < test.numProcessors; i++ {
			scores := []int{}
			for _, nodeScore := range test.nodeScores {
				scores = append(scores, nodeScore[i])
			}
			processors = append(processors, &scoreProcessor{scores: scores})
		}
		nd := NodeSorter{nodes: nodes, processors: processors}
		sorted := nd.Sort()
		got := []int{}
		for _, node := range sorted {
			idx, _ := strconv.Atoi(node.Name[5:])
			got = append(got, idx)
		}
		if diff := cmp.Diff(test.sortedOrder, got); diff != "" {
			t.Errorf("%s: NodeSorter.Sort() diff (-want +got):\n%s", test.name, diff)
		}
	}
}
