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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unremovable"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpdateClusterState(t *testing.T) {
	testCases := []struct {
		name            string
		nodes           []*apiv1.Node
		pods            []*apiv1.Pod
		actuationStatus *fakeActuationStatus
		eligible        []string
		wantUnneeded    []string
		replicasSets    []*appsv1.ReplicaSet
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
			eligible:     []string{"n1"},
			wantUnneeded: []string{},
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
			eligible:     []string{"n1"},
			wantUnneeded: []string{},
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
			eligible:     []string{"n1"},
			wantUnneeded: []string{"n1"},
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
			eligible:     []string{"n1"},
			wantUnneeded: []string{"n1"},
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
			eligible:     []string{"n1"},
			wantUnneeded: []string{"n1"},
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
			eligible:     []string{"n1"},
			wantUnneeded: []string{"n1"},
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
			eligible:     []string{"n1", "n2"},
			wantUnneeded: []string{"n1"},
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
			eligible:     []string{"n1", "n2"},
			wantUnneeded: []string{"n1", "n2"},
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
			eligible:     []string{"n1", "n2"},
			wantUnneeded: []string{"n1", "n2"},
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
			eligible:     []string{"n1", "n2"},
			wantUnneeded: []string{"n1"},
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
			eligible:     []string{"n1", "n2"},
			wantUnneeded: []string{},
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
			eligible:     []string{"n1", "n2"},
			wantUnneeded: []string{"n1", "n2"},
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
			eligible:     []string{"n1", "n2"},
			wantUnneeded: []string{"n1"},
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
			eligible:     []string{"n1", "n2"},
			wantUnneeded: []string{},
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
			registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, rsLister, nil)
			provider := testprovider.NewTestCloudProvider(nil, nil)
			context, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{ScaleDownSimulationTimeout: 5 * time.Minute}, &fake.Clientset{}, registry, provider, nil, nil)
			assert.NoError(t, err)
			clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, tc.nodes, tc.pods)
			deleteOptions := simulator.NodeDeleteOptions{}
			p := New(&context, NewTestProcessors(&context), deleteOptions)
			p.eligibilityChecker = &fakeEligibilityChecker{eligible: asMap(tc.eligible)}
			// TODO(x13n): test subsets of nodes passed as podDestinations/scaleDownCandidates.
			assert.NoError(t, p.UpdateClusterState(tc.nodes, tc.nodes, tc.actuationStatus, nil, time.Now()))
			wantUnneeded := asMap(tc.wantUnneeded)
			for _, n := range tc.nodes {
				assert.Equal(t, wantUnneeded[n.Name], p.unneededNodes.Contains(n.Name), n.Name)
			}
		})
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
	toBeDeletedTaint := apiv1.Taint{Key: deletetaint.ToBeDeletedTaint, Effect: apiv1.TaintEffectNoSchedule}
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
