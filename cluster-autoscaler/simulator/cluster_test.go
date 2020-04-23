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
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/kubelet/types"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestUtilization(t *testing.T) {
	gpuLabel := GetGPULabel()
	pod := BuildTestPod("p1", 100, 200000)
	pod2 := BuildTestPod("p2", -1, -1)

	nodeInfo := schedulerframework.NewNodeInfo(pod, pod, pod2)
	node := BuildTestNode("node1", 2000, 2000000)
	SetNodeReadyState(node, true, time.Time{})

	utilInfo, err := CalculateUtilization(node, nodeInfo, false, false, gpuLabel)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	node2 := BuildTestNode("node1", 2000, -1)

	_, err = CalculateUtilization(node2, nodeInfo, false, false, gpuLabel)
	assert.Error(t, err)

	daemonSetPod3 := BuildTestPod("p3", 100, 200000)
	daemonSetPod3.OwnerReferences = GenerateOwnerReferences("ds", "DaemonSet", "apps/v1", "")

	daemonSetPod4 := BuildTestPod("p4", 100, 200000)
	daemonSetPod4.OwnerReferences = GenerateOwnerReferences("ds", "CustomDaemonSet", "crd/v1", "")
	daemonSetPod4.Annotations = map[string]string{"cluster-autoscaler.kubernetes.io/daemonset-pod": "true"}

	nodeInfo = schedulerframework.NewNodeInfo(pod, pod, pod2, daemonSetPod3, daemonSetPod4)
	utilInfo, err = CalculateUtilization(node, nodeInfo, true, false, gpuLabel)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	nodeInfo = schedulerframework.NewNodeInfo(pod, pod2, daemonSetPod3)
	utilInfo, err = CalculateUtilization(node, nodeInfo, false, false, gpuLabel)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	mirrorPod4 := BuildTestPod("p4", 100, 200000)
	mirrorPod4.Annotations = map[string]string{
		types.ConfigMirrorAnnotationKey: "",
	}

	nodeInfo = schedulerframework.NewNodeInfo(pod, pod, pod2, mirrorPod4)
	utilInfo, err = CalculateUtilization(node, nodeInfo, false, true, gpuLabel)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	nodeInfo = schedulerframework.NewNodeInfo(pod, pod2, mirrorPod4)
	utilInfo, err = CalculateUtilization(node, nodeInfo, false, false, gpuLabel)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	gpuNode := BuildTestNode("gpu_node", 2000, 2000000)
	AddGpusToNode(gpuNode, 1)
	gpuPod := BuildTestPod("gpu_pod", 100, 200000)
	RequestGpuForPod(gpuPod, 1)
	TolerateGpuForPod(gpuPod)
	nodeInfo = schedulerframework.NewNodeInfo(pod, pod, gpuPod)
	utilInfo, err = CalculateUtilization(gpuNode, nodeInfo, false, false, gpuLabel)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1/1, utilInfo.Utilization, 0.01)

	// Node with Unready GPU
	gpuNode = BuildTestNode("gpu_node", 2000, 2000000)
	AddGpuLabelToNode(gpuNode)
	nodeInfo = schedulerframework.NewNodeInfo(pod, pod)
	utilInfo, err = CalculateUtilization(gpuNode, nodeInfo, false, false, gpuLabel)
	assert.NoError(t, err)
	assert.Zero(t, utilInfo.Utilization)
}

func nodeInfos(nodes []*apiv1.Node) []*schedulerframework.NodeInfo {
	result := make([]*schedulerframework.NodeInfo, len(nodes))
	for i, node := range nodes {
		ni := schedulerframework.NewNodeInfo()
		ni.SetNode(node)
		result[i] = ni
	}
	return result
}

func TestFindPlaceAllOk(t *testing.T) {
	node1 := BuildTestNode("n1", 1000, 2000000)
	SetNodeReadyState(node1, true, time.Time{})
	node2 := BuildTestNode("n2", 1000, 2000000)
	SetNodeReadyState(node2, true, time.Time{})

	pod1 := BuildTestPod("p1", 300, 500000)
	pod1.Spec.NodeName = "n1"
	new1 := BuildTestPod("p2", 600, 500000)
	new2 := BuildTestPod("p3", 500, 500000)

	oldHints := make(map[string]string)
	newHints := make(map[string]string)
	destinations := map[string]bool{
		"n1": true,
		"n2": true,
	}
	tracker := NewUsageTracker()

	clusterSnapshot := NewBasicClusterSnapshot()
	predicateChecker, err := NewTestPredicateChecker()
	assert.NoError(t, err)
	InitializeClusterSnapshotOrDie(t, clusterSnapshot,
		[]*apiv1.Node{node1, node2},
		[]*apiv1.Pod{pod1})

	err = findPlaceFor(
		"x",
		[]*apiv1.Pod{new1, new2},
		destinations,
		clusterSnapshot,
		predicateChecker,
		oldHints, newHints, tracker, time.Now())

	assert.Len(t, newHints, 2)
	assert.Contains(t, newHints, new1.Namespace+"/"+new1.Name)
	assert.Contains(t, newHints, new2.Namespace+"/"+new2.Name)
	assert.NoError(t, err)
}

func TestFindPlaceAllBas(t *testing.T) {
	node1 := BuildTestNode("n1", 1000, 2000000)
	SetNodeReadyState(node1, true, time.Time{})
	node2 := BuildTestNode("n2", 1000, 2000000)
	SetNodeReadyState(node2, true, time.Time{})

	pod1 := BuildTestPod("p1", 300, 500000)
	pod1.Spec.NodeName = "n1"
	new1 := BuildTestPod("p2", 600, 500000)
	new2 := BuildTestPod("p3", 500, 500000)
	new3 := BuildTestPod("p4", 700, 500000)

	oldHints := make(map[string]string)
	newHints := make(map[string]string)
	destinations := map[string]bool{
		"nbad": true,
		"n1":   true,
		"n2":   true,
	}
	tracker := NewUsageTracker()

	clusterSnapshot := NewBasicClusterSnapshot()
	predicateChecker, err := NewTestPredicateChecker()
	assert.NoError(t, err)
	InitializeClusterSnapshotOrDie(t, clusterSnapshot,
		[]*apiv1.Node{node1, node2},
		[]*apiv1.Pod{pod1})

	err = findPlaceFor(
		"nbad",
		[]*apiv1.Pod{new1, new2, new3},
		destinations,
		clusterSnapshot, predicateChecker,
		oldHints, newHints, tracker, time.Now())

	assert.Error(t, err)
	assert.True(t, len(newHints) == 2)
	assert.Contains(t, newHints, new1.Namespace+"/"+new1.Name)
	assert.Contains(t, newHints, new2.Namespace+"/"+new2.Name)
}

func TestFindNone(t *testing.T) {
	node1 := BuildTestNode("n1", 1000, 2000000)
	SetNodeReadyState(node1, true, time.Time{})
	node2 := BuildTestNode("n2", 1000, 2000000)
	SetNodeReadyState(node2, true, time.Time{})

	pod1 := BuildTestPod("p1", 300, 500000)
	pod1.Spec.NodeName = "n1"

	destinations := map[string]bool{
		"n1": true,
		"n2": true,
	}

	clusterSnapshot := NewBasicClusterSnapshot()
	predicateChecker, err := NewTestPredicateChecker()
	assert.NoError(t, err)
	InitializeClusterSnapshotOrDie(t, clusterSnapshot,
		[]*apiv1.Node{node1, node2},
		[]*apiv1.Pod{pod1})

	err = findPlaceFor(
		"x",
		[]*apiv1.Pod{},
		destinations,
		clusterSnapshot, predicateChecker,
		make(map[string]string),
		make(map[string]string),
		NewUsageTracker(),
		time.Now())
	assert.NoError(t, err)
}

func TestFindEmptyNodes(t *testing.T) {
	nodes := []*apiv1.Node{}
	nodeNames := []string{}
	for i := 0; i < 4; i++ {
		nodeName := fmt.Sprintf("n%d", i)
		node := BuildTestNode(nodeName, 1000, 2000000)
		SetNodeReadyState(node, true, time.Time{})
		nodes = append(nodes, node)
		nodeNames = append(nodeNames, nodeName)
	}

	pod1 := BuildTestPod("p1", 300, 500000)
	pod1.Spec.NodeName = "n1"

	pod2 := BuildTestPod("p2", 300, 500000)
	pod2.Spec.NodeName = "n2"
	pod2.Annotations = map[string]string{
		types.ConfigMirrorAnnotationKey: "",
	}

	clusterSnapshot := NewBasicClusterSnapshot()
	InitializeClusterSnapshotOrDie(t, clusterSnapshot, []*apiv1.Node{nodes[0], nodes[1], nodes[2], nodes[3]}, []*apiv1.Pod{pod1, pod2})

	emptyNodes := FindEmptyNodesToRemove(clusterSnapshot, nodeNames)
	assert.Equal(t, []string{nodeNames[0], nodeNames[2], nodeNames[3]}, emptyNodes)
}

type findNodesToRemoveTestConfig struct {
	name        string
	pods        []*apiv1.Pod
	allNodes    []*apiv1.Node
	candidates  []string
	toRemove    []NodeToBeRemoved
	unremovable []*UnremovableNode
}

func TestFindNodesToRemove(t *testing.T) {
	emptyNode := BuildTestNode("n1", 1000, 2000000)
	emptyNodeInfo := schedulerframework.NewNodeInfo()
	emptyNodeInfo.SetNode(emptyNode)

	// two small pods backed by ReplicaSet
	drainableNode := BuildTestNode("n2", 1000, 2000000)
	drainableNodeInfo := schedulerframework.NewNodeInfo()
	drainableNodeInfo.SetNode(drainableNode)

	// one small pod, not backed by anything
	nonDrainableNode := BuildTestNode("n3", 1000, 2000000)
	nonDrainableNodeInfo := schedulerframework.NewNodeInfo()
	nonDrainableNodeInfo.SetNode(nonDrainableNode)

	// one very large pod
	fullNode := BuildTestNode("n4", 1000, 2000000)
	fullNodeInfo := schedulerframework.NewNodeInfo()
	fullNodeInfo.SetNode(fullNode)

	SetNodeReadyState(emptyNode, true, time.Time{})
	SetNodeReadyState(drainableNode, true, time.Time{})
	SetNodeReadyState(nonDrainableNode, true, time.Time{})
	SetNodeReadyState(fullNode, true, time.Time{})

	ownerRefs := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	pod1 := BuildTestPod("p1", 100, 100000)
	pod1.OwnerReferences = ownerRefs
	pod1.Spec.NodeName = "n2"
	drainableNodeInfo.AddPod(pod1)

	pod2 := BuildTestPod("p2", 100, 100000)
	pod2.OwnerReferences = ownerRefs
	pod2.Spec.NodeName = "n2"
	drainableNodeInfo.AddPod(pod2)

	pod3 := BuildTestPod("p3", 100, 100000)
	pod3.Spec.NodeName = "n3"
	nonDrainableNodeInfo.AddPod(pod3)

	pod4 := BuildTestPod("p4", 1000, 100000)
	pod4.Spec.NodeName = "n4"
	fullNodeInfo.AddPod(pod4)

	emptyNodeToRemove := NodeToBeRemoved{
		Node:             emptyNode,
		PodsToReschedule: []*apiv1.Pod{},
	}
	drainableNodeToRemove := NodeToBeRemoved{
		Node:             drainableNode,
		PodsToReschedule: []*apiv1.Pod{pod1, pod2},
	}

	clusterSnapshot := NewBasicClusterSnapshot()
	predicateChecker, err := NewTestPredicateChecker()
	assert.NoError(t, err)
	tracker := NewUsageTracker()

	tests := []findNodesToRemoveTestConfig{
		// just an empty node, should be removed
		{
			name:        "just an empty node, should be removed",
			pods:        []*apiv1.Pod{},
			candidates:  []string{emptyNode.Name},
			allNodes:    []*apiv1.Node{emptyNode},
			toRemove:    []NodeToBeRemoved{emptyNodeToRemove},
			unremovable: []*UnremovableNode{},
		},
		// just a drainable node, but nowhere for pods to go to
		{
			name:        "just a drainable node, but nowhere for pods to go to",
			pods:        []*apiv1.Pod{pod1, pod2},
			candidates:  []string{drainableNode.Name},
			allNodes:    []*apiv1.Node{drainableNode},
			toRemove:    []NodeToBeRemoved{},
			unremovable: []*UnremovableNode{{Node: drainableNode, Reason: NoPlaceToMovePods}},
		},
		// drainable node, and a mostly empty node that can take its pods
		{
			name:        "drainable node, and a mostly empty node that can take its pods",
			pods:        []*apiv1.Pod{pod1, pod2, pod3},
			candidates:  []string{drainableNode.Name, nonDrainableNode.Name},
			allNodes:    []*apiv1.Node{drainableNode, nonDrainableNode},
			toRemove:    []NodeToBeRemoved{drainableNodeToRemove},
			unremovable: []*UnremovableNode{{Node: nonDrainableNode, Reason: BlockedByPod, BlockingPod: &drain.BlockingPod{Pod: pod3, Reason: drain.NotReplicated}}},
		},
		// drainable node, and a full node that cannot fit anymore pods
		{
			name:        "drainable node, and a full node that cannot fit anymore pods",
			pods:        []*apiv1.Pod{pod1, pod2, pod4},
			candidates:  []string{drainableNode.Name},
			allNodes:    []*apiv1.Node{drainableNode, fullNode},
			toRemove:    []NodeToBeRemoved{},
			unremovable: []*UnremovableNode{{Node: drainableNode, Reason: NoPlaceToMovePods}},
		},
		// 4 nodes, 1 empty, 1 drainable
		{
			name:        "4 nodes, 1 empty, 1 drainable",
			pods:        []*apiv1.Pod{pod1, pod2, pod3, pod4},
			candidates:  []string{emptyNode.Name, drainableNode.Name},
			allNodes:    []*apiv1.Node{emptyNode, drainableNode, fullNode, nonDrainableNode},
			toRemove:    []NodeToBeRemoved{emptyNodeToRemove, drainableNodeToRemove},
			unremovable: []*UnremovableNode{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			destinations := make([]string, 0, len(test.allNodes))
			for _, node := range test.allNodes {
				destinations = append(destinations, node.Name)
			}
			InitializeClusterSnapshotOrDie(t, clusterSnapshot, test.allNodes, test.pods)
			toRemove, unremovable, _, err := FindNodesToRemove(
				test.candidates, destinations, nil,
				clusterSnapshot, predicateChecker, len(test.allNodes), true, map[string]string{},
				tracker, time.Now(), []*policyv1.PodDisruptionBudget{})
			assert.NoError(t, err)
			fmt.Printf("Test scenario: %s, found len(toRemove)=%v, expected len(test.toRemove)=%v\n", test.name, len(toRemove), len(test.toRemove))
			assert.Equal(t, toRemove, test.toRemove)
			assert.Equal(t, unremovable, test.unremovable)
		})
	}
}
