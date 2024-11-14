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

package podlistprocessor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestFilterOutExpendable(t *testing.T) {
	testCases := []struct {
		name               string
		pods               []*apiv1.Pod
		wantPods           []*apiv1.Pod
		wantPodsInSnapshot []*apiv1.Pod
		priorityCutoff     int
		nodes              []*apiv1.Node
	}{
		{
			name: "no pods",
		},
		{
			name: "single non-expendable pod",
			pods: []*apiv1.Pod{
				test.BuildTestPod("p", 1000, 1),
			},
			wantPods: []*apiv1.Pod{
				test.BuildTestPod("p", 1000, 1),
			},
		},
		{
			name: "non-expendable pods with priority >= to cutoff priority",
			pods: []*apiv1.Pod{
				test.BuildTestPod("p1", 1000, 1, priority(2)),
				test.BuildTestPod("p2", 1000, 1, priority(3)),
			},
			wantPods: []*apiv1.Pod{
				test.BuildTestPod("p1", 1000, 1, priority(2)),
				test.BuildTestPod("p2", 1000, 1, priority(3)),
			},
			priorityCutoff: 2,
		},
		{
			name: "single expednable pod",
			pods: []*apiv1.Pod{
				test.BuildTestPod("p", 1000, 1, priority(2)),
			},
			priorityCutoff: 3,
		},
		{
			name: "single waiting-for-low-priority-preemption pod",
			pods: []*apiv1.Pod{
				test.BuildTestPod("p", 1000, 1, nominatedNodeName("node-1")),
			},
			nodes: []*apiv1.Node{
				test.BuildTestNode("node-1", 2400, 2400),
			},
			wantPodsInSnapshot: []*apiv1.Pod{
				test.BuildTestPod("p", 1000, 1, nominatedNodeName("node-1")),
			},
		},
		{
			name: "mixed expendable, non-expendable & waiting-for-low-priority-preemption pods",
			pods: []*apiv1.Pod{
				test.BuildTestPod("p1", 1000, 1, priority(3)),
				test.BuildTestPod("p2", 1000, 1, priority(4)),
				test.BuildTestPod("p3", 1000, 1, priority(1)),
				test.BuildTestPod("p4", 1000, 1),
				test.BuildTestPod("p5", 1000, 1, nominatedNodeName("node-1")),
			},
			priorityCutoff: 2,
			wantPods: []*apiv1.Pod{
				test.BuildTestPod("p1", 1000, 1, priority(3)),
				test.BuildTestPod("p2", 1000, 1, priority(4)),
				test.BuildTestPod("p4", 1000, 1),
			},
			wantPodsInSnapshot: []*apiv1.Pod{
				test.BuildTestPod("p5", 1000, 1, nominatedNodeName("node-1")),
			},
			nodes: []*apiv1.Node{
				test.BuildTestNode("node-1", 2400, 2400),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			processor := NewFilterOutExpendablePodListProcessor()
			snapshot := testsnapshot.NewTestSnapshotOrDie(t)
			err := snapshot.SetClusterState(tc.nodes, nil)
			assert.NoError(t, err)

			pods, err := processor.Process(&context.AutoscalingContext{
				ClusterSnapshot: snapshot,
				AutoscalingOptions: config.AutoscalingOptions{
					ExpendablePodsPriorityCutoff: tc.priorityCutoff,
				},
				AutoscalingKubeClients: context.AutoscalingKubeClients{
					ListerRegistry: newMockListerRegistry(tc.nodes),
				},
			}, tc.pods)

			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.wantPods, pods)

			var podsInSnapshot []*apiv1.Pod
			// Get pods in snapshot
			for _, n := range tc.nodes {
				nodeInfo, err := snapshot.GetNodeInfo(n.Name)
				assert.NoError(t, err)
				assert.NotEqual(t, nodeInfo.Pods(), nil)
				for _, podInfo := range nodeInfo.Pods() {
					podsInSnapshot = append(podsInSnapshot, podInfo.Pod)
				}
			}

			assert.ElementsMatch(t, tc.wantPodsInSnapshot, podsInSnapshot)
		})
	}
}

func priority(priority int32) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.Spec.Priority = &priority
	}
}
func nominatedNodeName(nodeName string) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.Status.NominatedNodeName = nodeName
	}
}

type mockListerRegistry struct {
	kube_util.ListerRegistry
	nodes []*apiv1.Node
}

func newMockListerRegistry(nodes []*apiv1.Node) *mockListerRegistry {
	return &mockListerRegistry{
		nodes: nodes,
	}
}

func (mlr mockListerRegistry) AllNodeLister() kube_util.NodeLister {
	return &mockNodeLister{nodes: mlr.nodes}
}

type mockNodeLister struct {
	nodes []*apiv1.Node
}

func (mnl *mockNodeLister) List() ([]*apiv1.Node, error) {
	return mnl.nodes, nil
}
func (mnl *mockNodeLister) Get(name string) (*apiv1.Node, error) {
	return nil, fmt.Errorf("Unsupported operation")
}
