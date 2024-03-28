/*
Copyright 2024 The Kubernetes Authors.

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

package forcescaledown

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestScaleUpUnschedulablePodsProcessor(t *testing.T) {
	testCases := []struct {
		name         string
		forcedNodes  []string
		nodes        []*apiv1.Node
		pods         []*apiv1.Pod
		deletingPods []string

		unschedulablePods []*apiv1.Pod
		wantPods          []*apiv1.Pod
	}{
		{
			name: "no nodes, no unschedulable pods",
		},
		{
			name: "no nodes, some unschedulable pods",
			unschedulablePods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				BuildTestPod("p2", 200, 1),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				BuildTestPod("p2", 200, 1),
			},
		},
		{
			name:        "single node has force-scale-down taint",
			forcedNodes: []string{"n"},
			nodes: []*apiv1.Node{
				BuildTestNode("n", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 200, 1, "n"),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				BuildTestPod("p2", 200, 1),
			},
		},
		{
			name:        "single node does not have force-scale-down taint",
			forcedNodes: []string{},
			nodes: []*apiv1.Node{
				BuildTestNode("n", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 200, 1, "n"),
			},
			wantPods: []*apiv1.Pod{},
		},
		{
			name:        "n1 has force-scale-down taint, and n2 does not have enough capacity",
			forcedNodes: []string{"n1"},
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 350, 10),
				BuildTestNode("n2", 350, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n1"),
				BuildScheduledTestPod("p2", 200, 1, "n1"),
				BuildScheduledTestPod("p3", 100, 1, "n2"),
				BuildScheduledTestPod("p4", 200, 1, "n2"),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				BuildTestPod("p2", 200, 1),
			},
		},
		{
			name:        "n1 has force-scale-down taint, but n2 has enough capacity",
			forcedNodes: []string{"n1"},
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n1"),
				BuildScheduledTestPod("p2", 200, 1, "n1"),
				BuildScheduledTestPod("p3", 100, 1, "n2"),
				BuildScheduledTestPod("p4", 200, 1, "n2"),
			},
			wantPods: []*apiv1.Pod{},
		},
		{
			name:        "3 nodes have force-scale-down taint",
			forcedNodes: []string{"n1", "n2", "n3"},
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n1"),
				BuildScheduledTestPod("p2", 200, 1, "n2"),
				BuildScheduledTestPod("p3", 100, 1, "n3"),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				BuildTestPod("p2", 200, 1),
				BuildTestPod("p3", 100, 1),
			},
		},
		{
			name:        "function input and force-scale-down nodes have duplicated pods",
			forcedNodes: []string{"n2"},
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n1"),
				BuildScheduledTestPod("p2", 200, 1, "n2"),
			},
			unschedulablePods: []*apiv1.Pod{
				BuildTestPod("p2", 200, 1),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p2", 200, 1),
			},
		},
		{
			name:        "force-scale-down pod is being deleted",
			forcedNodes: []string{"n1", "n2"},
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n1"),
				BuildScheduledTestPod("p2", 100, 1, "n2"),
			},
			deletingPods: []string{"p1"},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p2", 100, 1),
			},
		},
		{
			name:        "pod is force-scale-down but reschedulable",
			forcedNodes: []string{"n1"},
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n1"),
				BuildScheduledTestPod("p2", 100, 1, "n2"),
			},
			wantPods: []*apiv1.Pod{},
		},
	}

	for index := range testCases {
		tc := testCases[index]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			forcedNodeNames := map[string]bool{}
			for _, name := range tc.forcedNodes {
				forcedNodeNames[name] = true
			}
			for index := range tc.nodes {
				if !forcedNodeNames[tc.nodes[index].Name] {
					continue
				}
				taint := apiv1.Taint{
					Key:    taints.ForceScaleDownTaint,
					Effect: apiv1.TaintEffectNoSchedule,
				}
				tc.nodes[index].Spec.Taints = append(tc.nodes[index].Spec.Taints, taint)
			}
			deletingPodNames := map[string]bool{}
			for _, name := range tc.deletingPods {
				deletingPodNames[name] = true
			}
			for index := range tc.pods {
				if deletingPodNames[tc.pods[index].Name] {
					tc.pods[index].DeletionTimestamp = &metav1.Time{Time: time.Now()}
				}
			}

			predicateChecker, err := predicatechecker.NewTestPredicateChecker()
			assert.NoError(t, err)
			ctx := context.AutoscalingContext{
				ClusterSnapshot:  clustersnapshot.NewBasicClusterSnapshot(),
				PredicateChecker: predicateChecker,
				AutoscalingKubeClients: context.AutoscalingKubeClients{
					ListerRegistry: newMockListerRegistry(tc.nodes),
				},
			}
			clustersnapshot.InitializeClusterSnapshotOrDie(t, ctx.ClusterSnapshot, tc.nodes, tc.pods)

			processor := NewScaleUpUnschedulablePodsProcessor()
			pods, err := processor.Process(&ctx, tc.unschedulablePods)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.wantPods, pods)
		})
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
