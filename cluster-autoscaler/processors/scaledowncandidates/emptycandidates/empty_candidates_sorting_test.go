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

package emptycandidates

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

var err = fmt.Errorf("error")

type testNodeInfoGetter struct {
	m map[string]*framework.NodeInfo
}

func (t *testNodeInfoGetter) GetNodeInfo(nodeName string) (*framework.NodeInfo, error) {
	if nodeInfo, ok := t.m[nodeName]; ok {
		return nodeInfo, nil
	}
	return nil, err
}

func TestScaleDownEarlierThan(t *testing.T) {
	nodeEmptyName := "nodeEmpty"
	nodeEmpty := BuildTestNode(nodeEmptyName, 0, 100)
	niEmpty := framework.NewNodeInfo(nodeEmpty, nil)

	nodeEmptyName2 := "nodeEmpty2"
	nodeEmpty2 := BuildTestNode(nodeEmptyName2, 0, 100)
	niEmpty2 := framework.NewNodeInfo(nodeEmpty2, nil)

	nodeNonEmptyName := "nodeNonEmpty"
	nodeNonEmpty := BuildTestNode(nodeNonEmptyName, 0, 100)
	niNonEmpty := framework.NewNodeInfo(nodeNonEmpty, nil)
	pod := BuildTestPod("p1", 0, 100)
	niNonEmpty.AddPod(&framework.PodInfo{Pod: pod})

	noNodeInfoNode := BuildTestNode("n1", 0, 100)

	niGetter := testNodeInfoGetter{map[string]*framework.NodeInfo{nodeEmptyName: niEmpty, nodeNonEmptyName: niNonEmpty, nodeEmptyName2: niEmpty2}}

	deleteOptions := options.NodeDeleteOptions{
		SkipNodesWithSystemPods:           true,
		SkipNodesWithLocalStorage:         true,
		SkipNodesWithCustomControllerPods: true,
	}
	p := EmptySorting{
		nodeInfoGetter: &niGetter,
		deleteOptions:  deleteOptions,
	}

	tests := []struct {
		name        string
		node1       *v1.Node
		node2       *v1.Node
		wantEarlier bool
	}{
		{
			name:        "Empty node earlier that non-empty node",
			node1:       nodeEmpty,
			node2:       nodeNonEmpty,
			wantEarlier: true,
		},
		{
			name:        "Non-empty node is not earlier that empty node",
			node1:       nodeEmpty,
			node2:       nodeNonEmpty,
			wantEarlier: true,
		},
		{
			name:        "Empty node earlier that node without nodeInfo",
			node1:       nodeEmpty,
			node2:       noNodeInfoNode,
			wantEarlier: true,
		},
		{
			name:  "Non-empty node is not earlier that node without nodeInfo",
			node1: nodeNonEmpty,
			node2: noNodeInfoNode,
		},
		{
			name:  "Node without nodeInfo is not earlier that non-empty node",
			node1: noNodeInfoNode,
			node2: nodeNonEmpty,
		},
		{
			name:  "Empty node is not earlier that another empty node",
			node1: nodeEmpty,
			node2: nodeEmpty2,
		},
	}
	for _, test := range tests {
		gotEarlier := p.ScaleDownEarlierThan(test.node1, test.node2)
		if gotEarlier != test.wantEarlier {
			t.Errorf("%s: want %v, got %v", test.name, test.wantEarlier, gotEarlier)
		}
	}
}
