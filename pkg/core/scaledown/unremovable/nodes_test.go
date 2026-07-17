/*
Copyright 2022 The Kubernetes Authors.

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

package unremovable

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

var (
	beforeUpdate = time.UnixMilli(1652455130000)
	updateTime   = time.UnixMilli(1652455131111)
	afterUpdate  = time.UnixMilli(1652455132222)
)

func TestUpdate(t *testing.T) {
	testCases := map[string]struct {
		unremovable map[string]time.Time
		nodes       []string
		want        int
	}{
		"empty": {},
		"one removed via ttl": {
			unremovable: map[string]time.Time{
				"n1": beforeUpdate,
				"n2": afterUpdate,
			},
			nodes: []string{"n1", "n2"},
			want:  1,
		},
		"one missing from cluster": {
			unremovable: map[string]time.Time{
				"n1": afterUpdate,
				"n2": afterUpdate,
			},
			nodes: []string{"n2"},
			want:  1,
		},
		"all stay": {
			unremovable: map[string]time.Time{
				"n1": afterUpdate,
				"n2": afterUpdate,
			},
			nodes: []string{"n1", "n2"},
			want:  2,
		},
	}
	for desc, tc := range testCases {
		n := NewNodes()
		niGetter := newFakeNodeInfoGetter(tc.nodes)
		for name, timeout := range tc.unremovable {
			n.AddTimeout(makeUnremovableNode(name), timeout)
		}
		n.Update(niGetter, updateTime)
		got := len(n.ttls)
		if got != tc.want {
			t.Errorf("%s: got %d nodes, want %d", desc, got, tc.want)
		}
	}
}

func TestContains(t *testing.T) {
	n := NewNodes()
	nodes := []string{"n1", "n2", "n3"}

	n.Add(makeUnremovableNode(nodes[0]))
	n.AddTimeout(makeUnremovableNode(nodes[1]), time.Now())
	n.AddReason(BuildTestNode(nodes[2], 0, 0), simulator.UnremovableReason(1))

	for _, node := range nodes {
		if !n.Contains(node) {
			t.Errorf("n.Contains(%s) return false, want true", node)
		}
	}
	//remove nodes
	n.Update(newFakeNodeInfoGetter(nodes), time.Now().Add(-1*time.Minute))
	for _, node := range nodes {
		if n.Contains(node) {
			t.Errorf("n.Contains(%s) return true, want false", node)
		}
	}
}

type fakeNodeInfoGetter struct {
	names map[string]bool
}

func (f *fakeNodeInfoGetter) GetNodeInfo(name string) (*framework.NodeInfo, error) {
	// We don't actually care about the node info object itself, just its presence.
	_, found := f.names[name]
	if found {
		return nil, nil
	}
	return nil, fmt.Errorf("not found")
}

func newFakeNodeInfoGetter(ns []string) *fakeNodeInfoGetter {
	names := make(map[string]bool, len(ns))
	for _, n := range ns {
		names[n] = true
	}
	return &fakeNodeInfoGetter{names}
}

func makeUnremovableNode(name string) *simulator.UnremovableNode {
	return &simulator.UnremovableNode{
		Node: &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}
}
