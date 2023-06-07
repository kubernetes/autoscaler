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
	"sort"

	apiv1 "k8s.io/api/core/v1"
)

// CandidatesComparer is an  used for sorting scale down candidates.
type CandidatesComparer interface {
	// ScaleDownEarlierThan return true if node1 should be scaled down earlier than node2.
	ScaleDownEarlierThan(node1, node2 *apiv1.Node) bool
}

// NodeSorter struct contain the list of nodes and the list of processors that should be applied for sorting.
type NodeSorter struct {
	nodes      []*apiv1.Node
	processors []CandidatesComparer
}

// Sort return list of nodes in descending order.
func (n *NodeSorter) Sort() []*apiv1.Node {
	if len(n.processors) == 0 {
		return n.nodes
	}
	sort.Sort(n)
	return n.nodes
}

// Less return true if node with index i is less than node with index j.
func (n *NodeSorter) Less(i, j int) bool {
	node1, node2 := n.nodes[i], n.nodes[j]
	for _, processor := range n.processors {
		if val := processor.ScaleDownEarlierThan(node1, node2); val || processor.ScaleDownEarlierThan(node2, node1) {
			return val
		}
	}
	return false
}

// Swap is swapping the nodes in the list.
func (n *NodeSorter) Swap(i, j int) {
	n.nodes[i], n.nodes[j] = n.nodes[j], n.nodes[i]
}

// Len return the length of node's list.
func (n *NodeSorter) Len() int {
	return len(n.nodes)
}
