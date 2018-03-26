/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/kubernetes/pkg/scheduler/schedulercache"

	"github.com/stretchr/testify/assert"
)

func TestUtilization(t *testing.T) {
	pod := BuildTestPod("p1", 100, 200000)
	pod2 := BuildTestPod("p2", -1, -1)

	nodeInfo := schedulercache.NewNodeInfo(pod, pod, pod2)
	node := BuildTestNode("node1", 2000, 2000000)
	SetNodeReadyState(node, true, time.Time{})

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
	SetNodeReadyState(node1, true, time.Time{})
	node2 := BuildTestNode("n2", 1000, 2000000)
	SetNodeReadyState(node2, true, time.Time{})
	nodeInfos["n1"].SetNode(node1)
	nodeInfos["n2"].SetNode(node2)

	oldHints := make(map[string]string)
	newHints := make(map[string]string)
	tracker := NewUsageTracker()

	err := findPlaceFor(
		"x",
		[]*apiv1.Pod{new1, new2},
		[]*apiv1.Node{node1, node2},
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
	SetNodeReadyState(node1, true, time.Time{})

	node2 := BuildTestNode("n2", 1000, 2000000)
	SetNodeReadyState(node2, true, time.Time{})

	nodeInfos["n1"].SetNode(node1)
	nodeInfos["n2"].SetNode(node2)
	nodeInfos["nbad"].SetNode(nodebad)

	oldHints := make(map[string]string)
	newHints := make(map[string]string)
	tracker := NewUsageTracker()

	err := findPlaceFor(
		"nbad",
		[]*apiv1.Pod{new1, new2, new3},
		[]*apiv1.Node{nodebad, node1, node2},
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
	SetNodeReadyState(node1, true, time.Time{})

	node2 := BuildTestNode("n2", 1000, 2000000)
	SetNodeReadyState(node2, true, time.Time{})

	nodeInfos["n1"].SetNode(node1)
	nodeInfos["n2"].SetNode(node2)

	err := findPlaceFor(
		"x",
		[]*apiv1.Pod{},
		[]*apiv1.Node{node1, node2},
		nodeInfos, NewTestPredicateChecker(),
		make(map[string]string),
		make(map[string]string),
		NewUsageTracker(),
		time.Now())
	assert.NoError(t, err)
}

func TestShuffleNodes(t *testing.T) {
	nodes := []*apiv1.Node{
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

func TestFindEmptyNodes(t *testing.T) {
	pod1 := BuildTestPod("p1", 300, 500000)
	pod1.Spec.NodeName = "n1"
	pod2 := BuildTestPod("p2", 300, 500000)
	pod2.Spec.NodeName = "n2"
	pod2.Annotations = map[string]string{
		types.ConfigMirrorAnnotationKey: "",
	}

	node1 := BuildTestNode("n1", 1000, 2000000)
	node2 := BuildTestNode("n2", 1000, 2000000)
	node3 := BuildTestNode("n3", 1000, 2000000)
	node4 := BuildTestNode("n4", 1000, 2000000)

	SetNodeReadyState(node1, true, time.Time{})
	SetNodeReadyState(node2, true, time.Time{})
	SetNodeReadyState(node3, true, time.Time{})
	SetNodeReadyState(node4, true, time.Time{})

	emptyNodes := FindEmptyNodesToRemove([]*apiv1.Node{node1, node2, node3, node4}, []*apiv1.Pod{pod1, pod2})
	assert.Equal(t, []*apiv1.Node{node2, node3, node4}, emptyNodes)
}

type findNodesToRemoveTestConfig struct {
	name        string
	candidates  []*apiv1.Node
	allNodes    []*apiv1.Node
	toRemove    []NodeToBeRemoved
	unremovable []*apiv1.Node
}

func TestFindNodesToRemove(t *testing.T) {
	emptyNode := BuildTestNode("n1", 1000, 2000000)

	// two small pods backed by ReplicaSet
	drainableNode := BuildTestNode("n2", 1000, 2000000)

	// one small pod, not backed by anything
	nonDrainableNode := BuildTestNode("n3", 1000, 2000000)

	// one very large pod
	fullNode := BuildTestNode("n4", 1000, 2000000)

	SetNodeReadyState(emptyNode, true, time.Time{})
	SetNodeReadyState(drainableNode, true, time.Time{})
	SetNodeReadyState(nonDrainableNode, true, time.Time{})
	SetNodeReadyState(fullNode, true, time.Time{})

	ownerRefs := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	pod1 := BuildTestPod("p1", 100, 100000)
	pod1.OwnerReferences = ownerRefs
	pod1.Spec.NodeName = "n2"
	pod2 := BuildTestPod("p2", 100, 100000)
	pod2.OwnerReferences = ownerRefs
	pod2.Spec.NodeName = "n2"
	pod3 := BuildTestPod("p3", 100, 100000)
	pod3.Spec.NodeName = "n3"
	pod4 := BuildTestPod("p4", 1000, 100000)
	pod4.Spec.NodeName = "n4"

	emptyNodeToRemove := NodeToBeRemoved{
		Node:             emptyNode,
		PodsToReschedule: []*apiv1.Pod{},
	}
	drainableNodeToRemove := NodeToBeRemoved{
		Node:             drainableNode,
		PodsToReschedule: []*apiv1.Pod{pod1, pod2},
	}

	pods := []*apiv1.Pod{pod1, pod2, pod3, pod4}
	predicateChecker := NewTestPredicateChecker()
	tracker := NewUsageTracker()

	tests := []findNodesToRemoveTestConfig{
		// just an empty node, should be removed
		{
			name:        "just an empty node, should be removed",
			candidates:  []*apiv1.Node{emptyNode},
			allNodes:    []*apiv1.Node{emptyNode},
			toRemove:    []NodeToBeRemoved{emptyNodeToRemove},
			unremovable: []*apiv1.Node{},
		},
		// just a drainable node, but nowhere for pods to go to
		{
			name:        "just a drainable node, but nowhere for pods to go to",
			candidates:  []*apiv1.Node{drainableNode},
			allNodes:    []*apiv1.Node{drainableNode},
			toRemove:    []NodeToBeRemoved{},
			unremovable: []*apiv1.Node{drainableNode},
		},
		// drainable node, and a mostly empty node that can take its pods
		{
			name:        "drainable node, and a mostly empty node that can take its pods",
			candidates:  []*apiv1.Node{drainableNode, nonDrainableNode},
			allNodes:    []*apiv1.Node{drainableNode, nonDrainableNode},
			toRemove:    []NodeToBeRemoved{drainableNodeToRemove},
			unremovable: []*apiv1.Node{nonDrainableNode},
		},
		// drainable node, and a full node that cannot fit anymore pods
		{
			name:        "drainable node, and a full node that cannot fit anymore pods",
			candidates:  []*apiv1.Node{drainableNode},
			allNodes:    []*apiv1.Node{drainableNode, fullNode},
			toRemove:    []NodeToBeRemoved{},
			unremovable: []*apiv1.Node{drainableNode},
		},
		// 4 nodes, 1 empty, 1 drainable
		{
			name:        "4 nodes, 1 empty, 1 drainable",
			candidates:  []*apiv1.Node{emptyNode, drainableNode},
			allNodes:    []*apiv1.Node{emptyNode, drainableNode, fullNode, nonDrainableNode},
			toRemove:    []NodeToBeRemoved{emptyNodeToRemove, drainableNodeToRemove},
			unremovable: []*apiv1.Node{},
		},
	}

	for _, test := range tests {
		toRemove, unremovable, _, err := FindNodesToRemove(
			test.candidates, test.allNodes, pods, nil,
			predicateChecker, len(test.allNodes), true, map[string]string{},
			tracker, time.Now(), []*policyv1.PodDisruptionBudget{})
		assert.NoError(t, err)
		fmt.Printf("Test scenario: %s, found len(toRemove)=%v, expected len(test.toRemove)=%v\n", test.name, len(toRemove), len(test.toRemove))
		assert.Equal(t, toRemove, test.toRemove)
		assert.Equal(t, unremovable, test.unremovable)
	}

}
