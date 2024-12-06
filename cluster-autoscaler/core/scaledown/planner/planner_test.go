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

package planner

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unremovable"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	processorstest "k8s.io/autoscaler/cluster-autoscaler/processors/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpdateClusterState(t *testing.T) {
	testCases := []struct {
		name                string
		nodes               []*apiv1.Node
		pods                []*apiv1.Pod
		actuationStatus     *fakeActuationStatus
		eligible            []string
		wantUnneeded        []string
		wantUnremovable     []string
		replicasSets        []*appsv1.ReplicaSet
		isSimulationTimeout bool
	}{
		{
			name: "empty nodes, all eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{"n1", "n2", "n3"},
			wantUnneeded:    []string{"n1", "n2", "n3"},
		},
		{
			name: "empty nodes, some eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{"n1", "n2"},
			wantUnremovable: []string{"n3"},
		},
		{
			name: "empty nodes, none eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{},
			wantUnneeded:    []string{},
			wantUnremovable: []string{"n1", "n2", "n3"},
		},
		{
			name: "single utilised node, not eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n1"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{"n1"},
			wantUnneeded:    []string{},
			wantUnremovable: []string{"n1"},
		},
		{
			name: "pods cannot schedule on node undergoing deletion, not eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p2", 500, 1, "n1"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{"n1"},
			wantUnneeded:    []string{},
			wantUnremovable: []string{"n1", "n2"},
		},
		{
			name: "pods can schedule on non-eligible node, eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p2", 500, 1, "n1"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{"n1"},
			wantUnneeded:    []string{"n1"},
			wantUnremovable: []string{"n2"},
		},
		{
			name: "pods can schedule on eligible node, eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p2", 500, 1, "n1"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{"n1"},
			wantUnremovable: []string{"n2"},
		},
		{
			name: "pods cannot schedule anywhere, not eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 2000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 500, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 1000, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p2", 1000, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p3", 1000, 1, "n2"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{},
			wantUnremovable: []string{"n1", "n2", "n3"},
		},
		{
			name: "all pods from multiple nodes can schedule elsewhere, all eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 2000, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p2", 500, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p3", 500, 1, "n2"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p4", 500, 1, "n2"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{"n1", "n2"},
			wantUnremovable: []string{"n3"},
		},
		{
			name: "some pods from multiple nodes can schedule elsewhere, some eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 2000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 1000, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p2", 1000, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p3", 500, 1, "n2"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p4", 500, 1, "n2"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{"n2"},
			wantUnremovable: []string{"n1", "n3"},
		},
		{
			name: "no pods from multiple nodes can schedule elsewhere, no eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 500, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p2", 500, 1, "n1"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p3", 500, 1, "n2"), "rs"),
				SetRSPodSpec(BuildScheduledTestPod("p4", 500, 1, "n2"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{},
			wantUnremovable: []string{"n1", "n2", "n3"},
		},
		{
			name: "recently evicted RS pod, not eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 2000, 10),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetRSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n2"), "rs"),
				},
			},
			eligible:        []string{"n1"},
			wantUnneeded:    []string{},
			wantUnremovable: []string{"n1", "n2"},
		},
		{
			name: "recently evicted pod without owner, not eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					BuildTestPod("p1", 1000, 1),
				},
			},
			eligible:        []string{"n1"},
			wantUnneeded:    []string{},
			wantUnremovable: []string{"n1"},
		},
		{
			name: "recently evicted static pod, eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 2000, 10),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetStaticPodSpec(BuildScheduledTestPod("p1", 500, 1, "n2")),
				},
			},
			eligible:        []string{"n1"},
			wantUnneeded:    []string{"n1"},
			wantUnremovable: []string{"n2"},
		},
		{
			name: "recently evicted mirror pod, eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 2000, 10),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetMirrorPodSpec(BuildScheduledTestPod("p1", 500, 1, "n2")),
				},
			},
			eligible:        []string{"n1"},
			wantUnneeded:    []string{"n1"},
			wantUnremovable: []string{"n2"},
		},
		{
			name: "recently evicted DS pod, eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 2000, 10),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetDSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n2")),
				},
			},
			eligible:        []string{"n1"},
			wantUnneeded:    []string{"n1"},
			wantUnremovable: []string{"n2"},
		},
		{
			name: "recently evicted pod can schedule on non-eligible node, eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				nodeUndergoingDeletion("n3", 1000, 10),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetRSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n3"), "rs"),
				},
			},
			eligible:        []string{"n1"},
			wantUnneeded:    []string{"n1"},
			wantUnremovable: []string{"n2", "n3"},
		},
		{
			name: "recently evicted pod can schedule on eligible node, eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				nodeUndergoingDeletion("n3", 1000, 10),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetRSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n3"), "rs"),
				},
			},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{"n1"},
			wantUnremovable: []string{"n2", "n3"},
		},
		{
			name: "recently evicted pod too large to schedule anywhere, all eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				nodeUndergoingDeletion("n3", 2000, 10),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetRSPodSpec(BuildScheduledTestPod("p1", 2000, 1, "n3"), "rs"),
				},
			},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{"n1", "n2"},
			wantUnremovable: []string{"n3"},
		},
		{
			name: "all recently evicted pod got rescheduled, all eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				nodeUndergoingDeletion("n3", 2000, 10),
			},
			replicasSets: append(generateReplicaSetWithReplicas("rs1", 2, 2, nil), generateReplicaSets("rs", 5)...),
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetRSPodSpec(BuildScheduledTestPod("p1", 1000, 1, "n3"), "rs1"),
					SetRSPodSpec(BuildScheduledTestPod("p2", 1000, 1, "n3"), "rs1"),
				},
			},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{"n1", "n2"},
			wantUnremovable: []string{"n3"},
		},
		{
			name: "some recently evicted pod got rescheduled, some eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				nodeUndergoingDeletion("n3", 2000, 10),
			},
			replicasSets: append(generateReplicaSetWithReplicas("rs1", 2, 1, nil), generateReplicaSets("rs", 5)...),
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetRSPodSpec(BuildScheduledTestPod("p1", 1000, 1, "n3"), "rs1"),
					SetRSPodSpec(BuildScheduledTestPod("p2", 1000, 1, "n3"), "rs1"),
				},
			},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{"n1"},
			wantUnremovable: []string{"n2", "n3"},
		},
		{
			name: "no recently evicted pod got rescheduled, no eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				nodeUndergoingDeletion("n3", 2000, 10),
			},
			replicasSets: append(generateReplicaSetWithReplicas("rs1", 2, 0, nil), generateReplicaSets("rs", 5)...),
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetRSPodSpec(BuildScheduledTestPod("p1", 1000, 1, "n3"), "rs1"),
					SetRSPodSpec(BuildScheduledTestPod("p2", 1000, 1, "n3"), "rs1"),
				},
			},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{},
			wantUnremovable: []string{"n1", "n2", "n3"},
		},
		{
			name: "all scheduled and recently evicted pods can schedule elsewhere, all eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				nodeUndergoingDeletion("n4", 2000, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 250, 1, "n1"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetRSPodSpec(BuildScheduledTestPod("p2", 250, 1, "n4"), "rs"),
					SetRSPodSpec(BuildScheduledTestPod("p3", 250, 1, "n4"), "rs"),
				},
			},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{"n1", "n2"},
			wantUnremovable: []string{"n3", "n4"},
		},
		{
			name: "some scheduled and recently evicted pods can schedule elsewhere, some eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				nodeUndergoingDeletion("n4", 2000, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 500, 1, "n1"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetRSPodSpec(BuildScheduledTestPod("p2", 500, 1, "n4"), "rs"),
					SetRSPodSpec(BuildScheduledTestPod("p3", 500, 1, "n4"), "rs"),
				},
			},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{"n1"},
			wantUnremovable: []string{"n2", "n3", "n4"},
		},
		{
			name: "scheduled and recently evicted pods take all capacity, no eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				nodeUndergoingDeletion("n4", 2000, 10),
			},
			pods: []*apiv1.Pod{
				SetRSPodSpec(BuildScheduledTestPod("p1", 1000, 1, "n1"), "rs"),
			},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					SetRSPodSpec(BuildScheduledTestPod("p2", 1000, 1, "n4"), "rs"),
					SetRSPodSpec(BuildScheduledTestPod("p3", 1000, 1, "n4"), "rs"),
				},
			},
			eligible:        []string{"n1", "n2"},
			wantUnneeded:    []string{},
			wantUnremovable: []string{"n1", "n2", "n3", "n4"},
		},
		{
			name: "Simulation timeout is hitted",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
			},
			actuationStatus:     &fakeActuationStatus{},
			eligible:            []string{"n1", "n2", "n3"},
			wantUnneeded:        []string{"n1"},
			isSimulationTimeout: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.replicasSets == nil {
				tc.replicasSets = generateReplicaSets("rs", 5)
			}
			rsLister, err := kube_util.NewTestReplicaSetLister(tc.replicasSets)
			assert.NoError(t, err)
			registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, rsLister, nil)
			provider := testprovider.NewTestCloudProvider(nil, nil)
			provider.AddNodeGroup("ng1", 0, 0, 0)
			for _, node := range tc.nodes {
				provider.AddNode("ng1", node)
			}
			context, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{
				NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
					ScaleDownUnneededTime: 10 * time.Minute,
				},
				ScaleDownSimulationTimeout: 1 * time.Second,
				MaxScaleDownParallelism:    10,
			}, &fake.Clientset{}, registry, provider, nil, nil)
			assert.NoError(t, err)
			clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, tc.nodes, tc.pods)
			deleteOptions := options.NodeDeleteOptions{}
			p := New(&context, processorstest.NewTestProcessors(&context), deleteOptions, nil)
			p.eligibilityChecker = &fakeEligibilityChecker{eligible: asMap(tc.eligible)}
			if tc.isSimulationTimeout {
				context.AutoscalingOptions.ScaleDownSimulationTimeout = 1 * time.Second
				rs := &fakeRemovalSimulator{
					nodes: tc.nodes,
					sleep: 2 * time.Second,
				}
				p.rs = rs
			}
			// TODO(x13n): test subsets of nodes passed as podDestinations/scaleDownCandidates.
			assert.NoError(t, p.UpdateClusterState(tc.nodes, tc.nodes, tc.actuationStatus, time.Now()))
			wantUnneeded := asMap(tc.wantUnneeded)
			wantUnremovable := asMap(tc.wantUnremovable)
			for _, n := range tc.nodes {
				assert.Equal(t, wantUnneeded[n.Name], p.unneededNodes.Contains(n.Name), []string{n.Name, "unneeded"})
				assert.Equal(t, wantUnremovable[n.Name], p.unremovableNodes.Contains(n.Name), []string{n.Name, "unremovable"})
			}
		})
	}
}

func TestUpdateClusterStatUnneededNodesLimit(t *testing.T) {
	testCases := []struct {
		name               string
		previouslyUnneeded int
		nodes              int
		opts               *config.NodeGroupAutoscalingOptions
		maxParallelism     int
		maxUnneededTime    time.Duration
		updateInterval     time.Duration
		wantUnneeded       int
	}{
		{
			name:               "no unneeded, default settings",
			previouslyUnneeded: 0,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     10 * time.Second,
			wantUnneeded:       20,
		},
		{
			name:               "some unneeded, default settings",
			previouslyUnneeded: 3,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     10 * time.Second,
			wantUnneeded:       23,
		},
		{
			name:               "max unneeded, default settings",
			previouslyUnneeded: 70,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     10 * time.Second,
			wantUnneeded:       70,
		},
		{
			name:               "too many unneeded, default settings",
			previouslyUnneeded: 77,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     10 * time.Second,
			wantUnneeded:       70,
		},
		{
			name:               "instant kill nodes",
			previouslyUnneeded: 0,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    0 * time.Minute,
			updateInterval:     10 * time.Second,
			wantUnneeded:       20,
		},
		{
			name:               "quick loops",
			previouslyUnneeded: 13,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     1 * time.Second,
			wantUnneeded:       33,
		},
		{
			name:               "slow loops",
			previouslyUnneeded: 13,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     30 * time.Second,
			wantUnneeded:       30,
		},
		{
			name:               "atomic sclale down - default settings",
			previouslyUnneeded: 5,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     10 * time.Second,
			wantUnneeded:       100,
			opts: &config.NodeGroupAutoscalingOptions{
				ZeroOrMaxNodeScaling: true,
			},
		},
		{
			name:               "atomic sclale down - quick loops",
			previouslyUnneeded: 5,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     1 * time.Second,
			wantUnneeded:       100,
			opts: &config.NodeGroupAutoscalingOptions{
				ZeroOrMaxNodeScaling: true,
			},
		},
		{
			name:               "atomic sclale down - slow loops",
			previouslyUnneeded: 5,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     30 * time.Second,
			wantUnneeded:       100,
			opts: &config.NodeGroupAutoscalingOptions{
				ZeroOrMaxNodeScaling: true,
			},
		},
		{
			name:               "atomic sclale down - too many unneeded",
			previouslyUnneeded: 77,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     30 * time.Second,
			wantUnneeded:       100,
			opts: &config.NodeGroupAutoscalingOptions{
				ZeroOrMaxNodeScaling: true,
			},
		},
		{
			name:               "atomic sclale down - no uneeded",
			previouslyUnneeded: 0,
			nodes:              100,
			maxParallelism:     10,
			maxUnneededTime:    1 * time.Minute,
			updateInterval:     30 * time.Second,
			wantUnneeded:       100,
			opts: &config.NodeGroupAutoscalingOptions{
				ZeroOrMaxNodeScaling: true,
			},
		},
		{
			name:               "atomic sclale down - short uneeded time and short update interval",
			previouslyUnneeded: 0,
			nodes:              500,
			maxParallelism:     1,
			maxUnneededTime:    1 * time.Second,
			updateInterval:     1 * time.Second,
			wantUnneeded:       500,
			opts: &config.NodeGroupAutoscalingOptions{
				ZeroOrMaxNodeScaling: true,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			nodes := make([]*apiv1.Node, tc.nodes)
			for i := 0; i < tc.nodes; i++ {
				nodes[i] = BuildTestNode(fmt.Sprintf("n%d", i), 1000, 10)
			}
			previouslyUnneeded := make([]simulator.NodeToBeRemoved, tc.previouslyUnneeded)
			for i := 0; i < tc.previouslyUnneeded; i++ {
				previouslyUnneeded[i] = simulator.NodeToBeRemoved{Node: nodes[i]}
			}
			provider := testprovider.NewTestCloudProvider(nil, nil)
			provider.AddNodeGroupWithCustomOptions("ng1", 0, 0, 0, tc.opts)
			for _, node := range nodes {
				provider.AddNode("ng1", node)
			}
			context, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{
				NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
					ScaleDownUnneededTime: tc.maxUnneededTime,
				},
				ScaleDownSimulationTimeout: 1 * time.Hour,
				MaxScaleDownParallelism:    tc.maxParallelism,
			}, &fake.Clientset{}, nil, provider, nil, nil)
			assert.NoError(t, err)
			clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, nil)
			deleteOptions := options.NodeDeleteOptions{}
			p := New(&context, processorstest.NewTestProcessors(&context), deleteOptions, nil)
			p.eligibilityChecker = &fakeEligibilityChecker{eligible: asMap(nodeNames(nodes))}
			p.minUpdateInterval = tc.updateInterval
			p.unneededNodes.Update(previouslyUnneeded, time.Now())
			assert.NoError(t, p.UpdateClusterState(nodes, nodes, &fakeActuationStatus{}, time.Now()))
			assert.Equal(t, tc.wantUnneeded, len(p.unneededNodes.AsList()))
		})
	}
}

func TestNodesToDelete(t *testing.T) {
	testCases := []struct {
		name      string
		nodes     map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved
		wantEmpty []*apiv1.Node
		wantDrain []*apiv1.Node
	}{
		{
			name:      "empty",
			nodes:     map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{},
			wantEmpty: []*apiv1.Node{},
			wantDrain: []*apiv1.Node{},
		},
		{
			name: "single empty",
			nodes: map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{
				sizedNodeGroup("test-ng", 3, false): {
					buildRemovableNode("test-node", 0),
				},
			},
			wantEmpty: []*apiv1.Node{
				buildRemovableNode("test-node", 0).Node,
			},
			wantDrain: []*apiv1.Node{},
		},
		{
			name: "single drain",
			nodes: map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{
				sizedNodeGroup("test-ng", 3, false): {
					buildRemovableNode("test-node", 1),
				},
			},
			wantEmpty: []*apiv1.Node{},
			wantDrain: []*apiv1.Node{
				buildRemovableNode("test-node", 1).Node,
			},
		},
		{
			name: "single empty atomic",
			nodes: map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{
				sizedNodeGroup("atomic-ng", 3, true): {
					buildRemovableNode("node-1", 0),
				},
			},
			wantEmpty: []*apiv1.Node{},
			wantDrain: []*apiv1.Node{},
		},
		{
			name: "all empty atomic",
			nodes: map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{
				sizedNodeGroup("atomic-ng", 3, true): {
					buildRemovableNode("node-1", 0),
					buildRemovableNode("node-2", 0),
					buildRemovableNode("node-3", 0),
				},
			},
			wantEmpty: []*apiv1.Node{
				buildRemovableNode("node-1", 0).Node,
				buildRemovableNode("node-2", 0).Node,
				buildRemovableNode("node-3", 0).Node,
			},
			wantDrain: []*apiv1.Node{},
		},
		{
			name: "some drain atomic",
			nodes: map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{
				sizedNodeGroup("atomic-ng", 3, true): {
					buildRemovableNode("node-1", 0),
					buildRemovableNode("node-2", 0),
					buildRemovableNode("node-3", 1),
				},
			},
			wantEmpty: []*apiv1.Node{
				buildRemovableNode("node-1", 0).Node,
				buildRemovableNode("node-2", 0).Node,
			},
			wantDrain: []*apiv1.Node{
				buildRemovableNode("node-3", 1).Node,
			},
		},
		{
			name: "different groups",
			nodes: map[cloudprovider.NodeGroup][]simulator.NodeToBeRemoved{
				sizedNodeGroup("standard-empty-ng", 3, false): {
					buildRemovableNode("node-1", 0),
					buildRemovableNode("node-2", 0),
					buildRemovableNode("node-3", 0),
				},
				sizedNodeGroup("standard-drain-ng", 3, false): {
					buildRemovableNode("node-4", 1),
					buildRemovableNode("node-5", 2),
					buildRemovableNode("node-6", 3),
				},
				sizedNodeGroup("standard-mixed-ng", 3, false): {
					buildRemovableNode("node-7", 0),
					buildRemovableNode("node-8", 1),
					buildRemovableNode("node-9", 2),
				},
				sizedNodeGroup("atomic-empty-ng", 3, true): {
					buildRemovableNode("node-10", 0),
					buildRemovableNode("node-11", 0),
					buildRemovableNode("node-12", 0),
				},
				sizedNodeGroup("atomic-mixed-ng", 3, true): {
					buildRemovableNode("node-13", 0),
					buildRemovableNode("node-14", 1),
					buildRemovableNode("node-15", 2),
				},
				sizedNodeGroup("atomic-partial-ng", 3, true): {
					buildRemovableNode("node-16", 0),
					buildRemovableNode("node-17", 1),
				},
			},
			wantEmpty: []*apiv1.Node{
				buildRemovableNode("node-1", 0).Node,
				buildRemovableNode("node-2", 0).Node,
				buildRemovableNode("node-3", 0).Node,
				buildRemovableNode("node-7", 0).Node,
				buildRemovableNode("node-10", 0).Node,
				buildRemovableNode("node-11", 0).Node,
				buildRemovableNode("node-12", 0).Node,
				buildRemovableNode("node-13", 0).Node,
			},
			wantDrain: []*apiv1.Node{
				buildRemovableNode("node-4", 0).Node,
				buildRemovableNode("node-5", 0).Node,
				buildRemovableNode("node-6", 0).Node,
				buildRemovableNode("node-8", 0).Node,
				buildRemovableNode("node-9", 0).Node,
				buildRemovableNode("node-14", 0).Node,
				buildRemovableNode("node-15", 0).Node,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			provider := testprovider.NewTestCloudProvider(nil, nil)
			allNodes := []*apiv1.Node{}
			allRemovables := []simulator.NodeToBeRemoved{}
			for ng, nodes := range tc.nodes {
				provider.InsertNodeGroup(ng)
				for _, removable := range nodes {
					allNodes = append(allNodes, removable.Node)
					allRemovables = append(allRemovables, removable)
					provider.AddNode(ng.Id(), removable.Node)
				}
			}
			context, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{
				NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
					ScaleDownUnneededTime: 10 * time.Minute,
					ScaleDownUnreadyTime:  0 * time.Minute,
				},
			}, &fake.Clientset{}, nil, provider, nil, nil)
			assert.NoError(t, err)
			clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, nil)
			deleteOptions := options.NodeDeleteOptions{}
			p := New(&context, processorstest.NewTestProcessors(&context), deleteOptions, nil)
			p.latestUpdate = time.Now()
			p.scaleDownContext.ActuationStatus = deletiontracker.NewNodeDeletionTracker(0 * time.Second)
			p.unneededNodes.Update(allRemovables, time.Now().Add(-1*time.Hour))
			p.eligibilityChecker = &fakeEligibilityChecker{eligible: asMap(nodeNames(allNodes))}
			empty, drain := p.NodesToDelete(time.Now())
			assert.ElementsMatch(t, tc.wantEmpty, empty)
			assert.ElementsMatch(t, tc.wantDrain, drain)
		})
	}
}

func sizedNodeGroup(id string, size int, atomic bool) cloudprovider.NodeGroup {
	ng := testprovider.NewTestNodeGroup(id, 10000, 0, size, true, false, "n1-standard-2", nil, nil)
	ng.SetOptions(&config.NodeGroupAutoscalingOptions{
		ZeroOrMaxNodeScaling: atomic,
	})
	return ng
}

func buildRemovableNode(name string, podCount int) simulator.NodeToBeRemoved {
	podsToReschedule := []*apiv1.Pod{}
	for i := 0; i < podCount; i++ {
		podsToReschedule = append(podsToReschedule, &apiv1.Pod{})
	}
	return simulator.NodeToBeRemoved{
		Node:             BuildTestNode(name, 1000, 10),
		PodsToReschedule: podsToReschedule,
	}
}

func generateReplicaSets(name string, replicas int32) []*appsv1.ReplicaSet {
	return []*appsv1.ReplicaSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
				UID:       types.UID(name),
			},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
			},
		},
	}
}

func generateReplicaSetWithReplicas(name string, specReplicas, statusReplicas int32, labels map[string]string) []*appsv1.ReplicaSet {
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["app"] = "rs"
	return []*appsv1.ReplicaSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
				UID:       types.UID(name),
			},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &specReplicas,
				Selector: metav1.SetAsLabelSelector(labels),
			},
			Status: appsv1.ReplicaSetStatus{
				Replicas: statusReplicas,
			},
		},
	}
}

func nodeUndergoingDeletion(name string, cpu, memory int64) *apiv1.Node {
	n := BuildTestNode(name, cpu, memory)
	toBeDeletedTaint := apiv1.Taint{Key: taints.ToBeDeletedTaint, Effect: apiv1.TaintEffectNoSchedule}
	n.Spec.Taints = append(n.Spec.Taints, toBeDeletedTaint)
	return n
}

type fakeActuationStatus struct {
	recentEvictions []*apiv1.Pod
}

func (f *fakeActuationStatus) RecentEvictions() []*apiv1.Pod {
	return f.recentEvictions
}

func (f *fakeActuationStatus) DeletionsInProgress() ([]string, []string) {
	return nil, nil
}

func (f *fakeActuationStatus) DeletionResults() (map[string]status.NodeDeleteResult, time.Time) {
	return nil, time.Time{}
}

func (f *fakeActuationStatus) DeletionsCount(nodeGroup string) int {
	return 0
}

type fakeEligibilityChecker struct {
	eligible map[string]bool
}

func (f *fakeEligibilityChecker) FilterOutUnremovable(context *context.AutoscalingContext, scaleDownCandidates []*apiv1.Node, timestamp time.Time, unremovableNodes *unremovable.Nodes) ([]string, map[string]utilization.Info, []*simulator.UnremovableNode) {
	eligible := []string{}
	utilMap := make(map[string]utilization.Info)
	for _, n := range scaleDownCandidates {
		if f.eligible[n.Name] {
			eligible = append(eligible, n.Name)
			utilMap[n.Name] = utilization.Info{}
		} else {
			unremovableNodes.AddReason(n, simulator.UnexpectedError)
		}
	}
	return eligible, utilMap, nil
}

type fakeRemovalSimulator struct {
	nodes []*apiv1.Node
	sleep time.Duration
}

func (r *fakeRemovalSimulator) DropOldHints() {}

func (r *fakeRemovalSimulator) SimulateNodeRemoval(name string, _ map[string]bool, _ time.Time, _ pdb.RemainingPdbTracker) (*simulator.NodeToBeRemoved, *simulator.UnremovableNode) {
	time.Sleep(r.sleep)
	node := &apiv1.Node{}
	for _, n := range r.nodes {
		if n.Name == name {
			node = n
		}
	}
	return &simulator.NodeToBeRemoved{Node: node}, nil
}
