/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package simulator

import (
	"testing"
	"time"

	. "k8s.io/contrib/cluster-autoscaler/utils/test"

	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/stretchr/testify/assert"
)

func TestUtilization(t *testing.T) {
	pod := BuildTestPod("p1", 100, 200000)
	pod2 := BuildTestPod("p2", -1, -1)

	nodeInfo := schedulercache.NewNodeInfo(pod, pod, pod2)
	node := BuildTestNode("node1", 2000, 2000000)

	utilization, err := CalculateUtilization(node, nodeInfo)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilization, 0.01)

	node2 := BuildTestNode("node1", 2000, -1)

	_, err = CalculateUtilization(node2, nodeInfo)
	assert.Error(t, err)
}

func TestFindPlaceAllOk(t *testing.T) {
	pod1 := BuildTestPod("p1", 300, 500000)
	new1 := BuildTestPod("p2", 600, 500000)
	new2 := BuildTestPod("p3", 500, 500000)

	nodeInfos := map[string]*schedulercache.NodeInfo{
		"n1": schedulercache.NewNodeInfo(pod1),
		"n2": schedulercache.NewNodeInfo(),
	}
	node1 := BuildTestNode("n1", 1000, 2000000)
	node2 := BuildTestNode("n2", 1000, 2000000)
	nodeInfos["n1"].SetNode(node1)
	nodeInfos["n2"].SetNode(node2)

	oldHints := make(map[string]string)
	newHints := make(map[string]string)
	tracker := NewUsageTracker()

	err := findPlaceFor(
		"x",
		[]*kube_api.Pod{new1, new2},
		[]*kube_api.Node{node1, node2},
		nodeInfos, NewTestPredicateChecker(),
		oldHints, newHints, tracker, time.Now())

	assert.Len(t, newHints, 2)
	assert.Contains(t, newHints, new1.Namespace+"/"+new1.Name)
	assert.Contains(t, newHints, new2.Namespace+"/"+new2.Name)
	assert.NoError(t, err)
}

func TestFindPlaceAllBas(t *testing.T) {
	pod1 := BuildTestPod("p1", 300, 500000)
	new1 := BuildTestPod("p2", 600, 500000)
	new2 := BuildTestPod("p3", 500, 500000)
	new3 := BuildTestPod("p4", 700, 500000)

	nodeInfos := map[string]*schedulercache.NodeInfo{
		"n1":   schedulercache.NewNodeInfo(pod1),
		"n2":   schedulercache.NewNodeInfo(),
		"nbad": schedulercache.NewNodeInfo(),
	}
	nodebad := BuildTestNode("nbad", 1000, 2000000)
	node1 := BuildTestNode("n1", 1000, 2000000)
	node2 := BuildTestNode("n2", 1000, 2000000)
	nodeInfos["n1"].SetNode(node1)
	nodeInfos["n2"].SetNode(node2)
	nodeInfos["nbad"].SetNode(nodebad)

	oldHints := make(map[string]string)
	newHints := make(map[string]string)
	tracker := NewUsageTracker()

	err := findPlaceFor(
		"nbad",
		[]*kube_api.Pod{new1, new2, new3},
		[]*kube_api.Node{nodebad, node1, node2},
		nodeInfos, NewTestPredicateChecker(),
		oldHints, newHints, tracker, time.Now())

	assert.Error(t, err)
	assert.True(t, len(newHints) == 2)
	assert.Contains(t, newHints, new1.Namespace+"/"+new1.Name)
	assert.Contains(t, newHints, new2.Namespace+"/"+new2.Name)
}

func TestFindNone(t *testing.T) {
	pod1 := BuildTestPod("p1", 300, 500000)

	nodeInfos := map[string]*schedulercache.NodeInfo{
		"n1": schedulercache.NewNodeInfo(pod1),
		"n2": schedulercache.NewNodeInfo(),
	}
	node1 := BuildTestNode("n1", 1000, 2000000)
	node2 := BuildTestNode("n2", 1000, 2000000)
	nodeInfos["n1"].SetNode(node1)
	nodeInfos["n2"].SetNode(node2)

	err := findPlaceFor(
		"x",
		[]*kube_api.Pod{},
		[]*kube_api.Node{node1, node2},
		nodeInfos, NewTestPredicateChecker(),
		make(map[string]string),
		make(map[string]string),
		NewUsageTracker(),
		time.Now())
	assert.NoError(t, err)
}

func TestShuffleNodes(t *testing.T) {
	nodes := []*kube_api.Node{
		BuildTestNode("n1", 0, 0),
		BuildTestNode("n2", 0, 0),
		BuildTestNode("n3", 0, 0)}
	gotPermutation := false
	for i := 0; i < 10000; i++ {
		shuffled := shuffleNodes(nodes)
		if shuffled[0].Name == "n2" && shuffled[1].Name == "n3" && shuffled[2].Name == "n1" {
			gotPermutation = true
			break
		}
	}
	assert.True(t, gotPermutation)
}
