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

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/kubelet/types"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

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

	clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot(framework.TestFrameworkHandleOrDie(t), true)
	clustersnapshot.InitializeClusterSnapshotOrDie(t, clusterSnapshot, []*apiv1.Node{nodes[0], nodes[1], nodes[2], nodes[3]}, []*apiv1.Pod{pod1, pod2})
	testTime := time.Date(2020, time.December, 18, 17, 0, 0, 0, time.UTC)
	r := NewRemovalSimulator(nil, clusterSnapshot, nil, testDeleteOptions(), nil, false)
	emptyNodes := r.FindEmptyNodesToRemove(nodeNames, testTime)
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

	replicas := int32(5)
	replicaSets := []*appsv1.ReplicaSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rs",
				Namespace: "default",
				SelfLink:  "api/v1/namespaces/default/replicasets/rs",
			},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
			},
		},
	}
	rsLister, err := kube_util.NewTestReplicaSetLister(replicaSets)
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, rsLister, nil)

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
		Node: emptyNode,
	}
	drainableNodeToRemove := NodeToBeRemoved{
		Node:             drainableNode,
		PodsToReschedule: []*apiv1.Pod{pod1, pod2},
	}

	fwHandle := framework.TestFrameworkHandleOrDie(t)
	clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot(fwHandle, true)
	predicateChecker := predicatechecker.NewSchedulerBasedPredicateChecker(fwHandle)

	tests := []findNodesToRemoveTestConfig{
		{
			name:       "just an empty node, should be removed",
			candidates: []string{emptyNode.Name},
			allNodes:   []*apiv1.Node{emptyNode},
			toRemove:   []NodeToBeRemoved{emptyNodeToRemove},
		},
		{
			name:        "just a drainable node, but nowhere for pods to go to",
			pods:        []*apiv1.Pod{pod1, pod2},
			candidates:  []string{drainableNode.Name},
			allNodes:    []*apiv1.Node{drainableNode},
			unremovable: []*UnremovableNode{{Node: drainableNode, Reason: NoPlaceToMovePods}},
		},
		{
			name:        "drainable node, and a mostly empty node that can take its pods",
			pods:        []*apiv1.Pod{pod1, pod2, pod3},
			candidates:  []string{drainableNode.Name, nonDrainableNode.Name},
			allNodes:    []*apiv1.Node{drainableNode, nonDrainableNode},
			toRemove:    []NodeToBeRemoved{drainableNodeToRemove},
			unremovable: []*UnremovableNode{{Node: nonDrainableNode, Reason: BlockedByPod, BlockingPod: &drain.BlockingPod{Pod: pod3, Reason: drain.NotReplicated}}},
		},
		{
			name:        "drainable node, and a full node that cannot fit anymore pods",
			pods:        []*apiv1.Pod{pod1, pod2, pod4},
			candidates:  []string{drainableNode.Name},
			allNodes:    []*apiv1.Node{drainableNode, fullNode},
			unremovable: []*UnremovableNode{{Node: drainableNode, Reason: NoPlaceToMovePods}},
		},
		{
			name:       "4 nodes, 1 empty, 1 drainable",
			pods:       []*apiv1.Pod{pod1, pod2, pod3, pod4},
			candidates: []string{emptyNode.Name, drainableNode.Name},
			allNodes:   []*apiv1.Node{emptyNode, drainableNode, fullNode, nonDrainableNode},
			toRemove:   []NodeToBeRemoved{emptyNodeToRemove, drainableNodeToRemove},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			destinations := make([]string, 0, len(test.allNodes))
			for _, node := range test.allNodes {
				destinations = append(destinations, node.Name)
			}
			clustersnapshot.InitializeClusterSnapshotOrDie(t, clusterSnapshot, test.allNodes, test.pods)
			r := NewRemovalSimulator(registry, clusterSnapshot, predicateChecker, testDeleteOptions(), nil, false)
			toRemove, unremovable := r.FindNodesToRemove(test.candidates, destinations, time.Now(), nil)
			fmt.Printf("Test scenario: %s, found len(toRemove)=%v, expected len(test.toRemove)=%v\n", test.name, len(toRemove), len(test.toRemove))
			assert.Equal(t, test.toRemove, toRemove)
			assert.Equal(t, test.unremovable, unremovable)
		})
	}
}

func testDeleteOptions() options.NodeDeleteOptions {
	return options.NodeDeleteOptions{
		SkipNodesWithSystemPods:           true,
		SkipNodesWithLocalStorage:         true,
		SkipNodesWithCustomControllerPods: true,
	}
}
