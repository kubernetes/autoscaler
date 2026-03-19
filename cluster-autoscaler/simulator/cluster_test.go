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

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

type simulateNodeRemovalTestConfig struct {
	name        string
	pods        []*apiv1.Pod
	allNodes    []*apiv1.Node
	nodeName    string
	toRemove    *NodeToBeRemoved
	unremovable *UnremovableNode
}

func TestSimulateNodeRemoval(t *testing.T) {
	emptyNode := BuildTestNode("n1", 1000, 2000000)

	// two small pods backed by ReplicaSet
	drainableNode := BuildTestNode("n2", 1000, 2000000)
	drainableNodeInfo := framework.NewTestNodeInfo(drainableNode)

	// one small pod, not backed by anything
	nonDrainableNode := BuildTestNode("n3", 1000, 2000000)
	nonDrainableNodeInfo := framework.NewTestNodeInfo(nonDrainableNode)

	// one very large pod
	fullNode := BuildTestNode("n4", 1000, 2000000)
	fullNodeInfo := framework.NewTestNodeInfo(fullNode)

	// noExistNode it doesn't have any node info in the cluster snapshot.
	noExistNode := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n5"}}

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
	drainableNodeInfo.AddPod(framework.NewPodInfo(pod1, nil))

	pod2 := BuildTestPod("p2", 100, 100000)
	pod2.OwnerReferences = ownerRefs
	pod2.Spec.NodeName = "n2"
	drainableNodeInfo.AddPod(framework.NewPodInfo(pod2, nil))

	pod3 := BuildTestPod("p3", 100, 100000)
	pod3.Spec.NodeName = "n3"
	nonDrainableNodeInfo.AddPod(framework.NewPodInfo(pod3, nil))

	pod4 := BuildTestPod("p4", 1000, 100000)
	pod4.Spec.NodeName = "n4"
	fullNodeInfo.AddPod(framework.NewPodInfo(pod4, nil))

	clusterSnapshot := testsnapshot.NewTestSnapshotOrDie(t)

	topoNode1 := BuildTestNode("topo-n1", 1000, 2000000)
	topoNode2 := BuildTestNode("topo-n2", 1000, 2000000)
	topoNode3 := BuildTestNode("topo-n3", 1000, 2000000)
	topoNode1.Labels = map[string]string{"kubernetes.io/hostname": "topo-n1"}
	topoNode2.Labels = map[string]string{"kubernetes.io/hostname": "topo-n2"}
	topoNode3.Labels = map[string]string{"kubernetes.io/hostname": "topo-n3"}

	SetNodeReadyState(topoNode1, true, time.Time{})
	SetNodeReadyState(topoNode2, true, time.Time{})
	SetNodeReadyState(topoNode3, true, time.Time{})

	minDomains := int32(2)
	maxSkew := int32(1)
	topoConstraint := apiv1.TopologySpreadConstraint{
		MaxSkew:           maxSkew,
		TopologyKey:       "kubernetes.io/hostname",
		WhenUnsatisfiable: apiv1.DoNotSchedule,
		MinDomains:        &minDomains,
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "topo-app",
			},
		},
	}

	pod5 := BuildTestPod("p5", 100, 100000)
	pod5.Labels = map[string]string{"app": "topo-app"}
	pod5.OwnerReferences = ownerRefs
	pod5.Spec.NodeName = "topo-n1"
	pod5.Spec.TopologySpreadConstraints = []apiv1.TopologySpreadConstraint{topoConstraint}

	pod6 := BuildTestPod("p6", 100, 100000)
	pod6.Labels = map[string]string{"app": "topo-app"}
	pod6.OwnerReferences = ownerRefs
	pod6.Spec.NodeName = "topo-n2"
	pod6.Spec.TopologySpreadConstraints = []apiv1.TopologySpreadConstraint{topoConstraint}

	pod7 := BuildTestPod("p7", 100, 100000)
	pod7.Labels = map[string]string{"app": "topo-app"}
	pod7.OwnerReferences = ownerRefs
	pod7.Spec.NodeName = "topo-n3"
	pod7.Spec.TopologySpreadConstraints = []apiv1.TopologySpreadConstraint{topoConstraint}

	blocker1 := BuildTestPod("blocker1", 100, 100000)
	blocker1.Spec.NodeName = "topo-n2"
	blocker2 := BuildTestPod("blocker2", 100, 100000)
	blocker2.Spec.NodeName = "topo-n3"

	tests := []simulateNodeRemovalTestConfig{
		{
			name:        "just an empty node, should be removed",
			nodeName:    emptyNode.Name,
			allNodes:    []*apiv1.Node{emptyNode},
			toRemove:    &NodeToBeRemoved{Node: emptyNode},
			unremovable: nil,
		},
		{
			name:     "just a drainable node, but nowhere for pods to go to",
			pods:     []*apiv1.Pod{pod1, pod2},
			nodeName: drainableNode.Name,
			allNodes: []*apiv1.Node{drainableNode},
			toRemove: nil,
			unremovable: &UnremovableNode{
				Node:   drainableNode,
				Reason: NoPlaceToMovePods,
			},
		},
		{
			name:        "drainable node, and a mostly empty node that can take its pods",
			pods:        []*apiv1.Pod{pod1, pod2, pod3},
			nodeName:    drainableNode.Name,
			allNodes:    []*apiv1.Node{drainableNode, nonDrainableNode},
			toRemove:    &NodeToBeRemoved{Node: drainableNode, PodsToReschedule: []*apiv1.Pod{pod1, pod2}},
			unremovable: nil,
		},
		{
			name:        "drainable node, and a full node that cannot fit anymore pods",
			pods:        []*apiv1.Pod{pod1, pod2, pod4},
			nodeName:    drainableNode.Name,
			allNodes:    []*apiv1.Node{drainableNode, fullNode},
			toRemove:    nil,
			unremovable: &UnremovableNode{Node: drainableNode, Reason: NoPlaceToMovePods},
		},
		{
			name:        "4 nodes, 1 empty, 1 drainable",
			pods:        []*apiv1.Pod{pod1, pod2, pod3, pod4},
			nodeName:    emptyNode.Name,
			allNodes:    []*apiv1.Node{emptyNode, drainableNode, fullNode, nonDrainableNode},
			toRemove:    &NodeToBeRemoved{Node: emptyNode},
			unremovable: nil,
		},
		{
			name:        "topology spread constraint test - one node should be removable",
			pods:        []*apiv1.Pod{pod5, pod6, pod7, blocker1, blocker2},
			allNodes:    []*apiv1.Node{topoNode1, topoNode2, topoNode3},
			nodeName:    topoNode1.Name,
			toRemove:    &NodeToBeRemoved{Node: topoNode1, PodsToReschedule: []*apiv1.Pod{pod5}},
			unremovable: nil,
		},
		{
			name:        "candidate not in clusterSnapshot should be marked unremovable",
			nodeName:    noExistNode.Name,
			allNodes:    []*apiv1.Node{},
			pods:        []*apiv1.Pod{},
			toRemove:    nil,
			unremovable: &UnremovableNode{Node: noExistNode, Reason: NoNodeInfo},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			destinations := make(map[string]bool)
			for _, node := range test.allNodes {
				destinations[node.Name] = true
			}
			clustersnapshot.InitializeClusterSnapshotOrDie(t, clusterSnapshot, test.allNodes, test.pods)
			r := NewRemovalSimulator(registry, clusterSnapshot, testDeleteOptions(), nil, false)
			toRemove, unremovable := r.SimulateNodeRemoval(test.nodeName, destinations, time.Now(), nil)
			fmt.Printf("Test scenario: %s, toRemove=%v, unremovable=%v\n", test.name, toRemove, unremovable)
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
