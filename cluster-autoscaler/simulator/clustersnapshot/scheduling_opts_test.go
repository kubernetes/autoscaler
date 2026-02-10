/*
Copyright 2025 The Kubernetes Authors.

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
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestLastIndexOrderMapping(t *testing.T) {
	n1 := framework.NewTestNodeInfo(BuildTestNode("n1", 1000, 1000))
	n2 := framework.NewTestNodeInfo(BuildTestNode("n2", 1000, 1000))
	n3 := framework.NewTestNodeInfo(BuildTestNode("n3", 1000, 1000))
	nodes := []*framework.NodeInfo{n1, n2, n3}

	testCases := []struct {
		name             string
		additionalOffset int
		lastMatch        int
		wantOrder        []int
	}{
		{
			name:             "offset 1, lastMatch 0",
			additionalOffset: 1,
			lastMatch:        0,
			wantOrder:        []int{1, 2, 0},
		},
		{
			name:             "offset 1, lastMatch 1",
			additionalOffset: 1,
			lastMatch:        1,
			wantOrder:        []int{2, 0, 1},
		},
		{
			name:             "offset 1, lastMatch 2",
			additionalOffset: 1,
			lastMatch:        2,
			wantOrder:        []int{0, 1, 2},
		},
		{
			name:             "offset 0, lastMatch 1",
			additionalOffset: 0,
			lastMatch:        1,
			wantOrder:        []int{1, 2, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.wantOrder) != len(nodes) {
				t.Fatalf("test case %q has invalid wantOrder length: %d, want %d", tc.name, len(tc.wantOrder), len(nodes))
			}

			mapping := NewLastIndexOrderMapping(tc.additionalOffset)
			mapping.Init(nodes, tc.lastMatch)

			var gotOrder []int
			for i := 0; i < len(nodes); i++ {
				gotOrder = append(gotOrder, mapping.At(i))
			}
			assert.Equal(t, tc.wantOrder, gotOrder)
		})
	}
}

func TestPriorityNodeOrderMapping(t *testing.T) {
	n1 := framework.NewTestNodeInfo(BuildTestNode("n1", 100, 1000))
	n2 := framework.NewTestNodeInfo(BuildTestNode("n2", 200, 1000))
	n3 := framework.NewTestNodeInfo(BuildTestNode("n3", 300, 1000))
	nodes := []*framework.NodeInfo{n1, n2, n3}

	testCases := []struct {
		name      string
		less      func(a, b *framework.NodeInfo) bool
		wantOrder []int
	}{
		{
			name: "sort by name ascending",
			less: func(a, b *framework.NodeInfo) bool {
				return a.Node().Name < b.Node().Name
			},
			wantOrder: []int{0, 1, 2},
		},
		{
			name: "sort by name descending",
			less: func(a, b *framework.NodeInfo) bool {
				return a.Node().Name > b.Node().Name
			},
			wantOrder: []int{2, 1, 0},
		},
		{
			name: "sort by cpu ascending",
			less: func(a, b *framework.NodeInfo) bool {
				return a.Node().Status.Capacity.Cpu().MilliValue() < b.Node().Status.Capacity.Cpu().MilliValue()
			},
			wantOrder: []int{0, 1, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.wantOrder) != len(nodes) {
				t.Fatalf("test case %q has invalid wantOrder length: %d, want %d", tc.name, len(tc.wantOrder), len(nodes))
			}

			mapping := NewPriorityNodeOrderMapping(tc.less)
			mapping.Init(nodes, 1)

			var gotOrder []int
			for i := 0; i < len(nodes); i++ {
				gotOrder = append(gotOrder, mapping.At(i))
			}
			assert.Equal(t, tc.wantOrder, gotOrder)
		})
	}
}
