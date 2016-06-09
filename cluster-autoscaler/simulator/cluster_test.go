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

	err := findPlaceFor(
		"x",
		[]*kube_api.Pod{new1, new2},
		[]*kube_api.Node{node1, node2},
		nodeInfos, NewTestPredicateChecker())
	assert.NoError(t, err)
}

func TestFindPlaceAllBas(t *testing.T) {
	pod1 := BuildTestPod("p1", 300, 500000)
	new1 := BuildTestPod("p2", 600, 500000)
	new2 := BuildTestPod("p3", 500, 500000)
	new3 := BuildTestPod("p4", 700, 500000)

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
		[]*kube_api.Pod{new1, new2, new3},
		[]*kube_api.Node{node1, node2},
		nodeInfos, NewTestPredicateChecker())
	assert.Error(t, err)
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
		nodeInfos, NewTestPredicateChecker())
	assert.NoError(t, err)
}
