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

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var rSetLabels = map[string]string{
	"app": "rs",
}

func TestUpdateClusterState(t *testing.T) {
	testCases := []struct {
		name            string
		nodes           []*apiv1.Node
		pods            []*apiv1.Pod
		actuationStatus *fakeActuationStatus
		eligible        []string
		wantUnneeded    []string
		wantErr         bool
		replicasSets    []*appsv1.ReplicaSet
	}{
		{
			name: "all eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				BuildTestNode("n4", 1000, 10),
			},
			eligible:        []string{"n1", "n2", "n3", "n4"},
			actuationStatus: &fakeActuationStatus{},
			wantUnneeded:    []string{"n1", "n2", "n3", "n4"},
		},
		{
			name: "none eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				BuildTestNode("n4", 1000, 10),
			},
			eligible:        []string{},
			actuationStatus: &fakeActuationStatus{},
			wantUnneeded:    []string{},
		},
		{
			name: "some eligible",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				BuildTestNode("n4", 1000, 10),
			},
			eligible:        []string{"n1", "n3"},
			actuationStatus: &fakeActuationStatus{},
			wantUnneeded:    []string{"n1", "n3"},
		},
		{
			name: "pods from already drained node can schedule elsewhere",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 2000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 500, 1, "n2", "rs"),
				scheduledPod("p2", 500, 1, "n2", "rs"),
			},
			eligible: []string{"n1"},
			actuationStatus: &fakeActuationStatus{
				currentlyDrained: []string{"n2"},
			},
			wantUnneeded: []string{},
		},
		{
			name: "pods from already drained node can't schedule elsewhere",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 2000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 500, 1, "n2", "rs"),
				scheduledPod("p2", 500, 1, "n2", "rs"),
				scheduledPod("p3", 500, 1, "n2", "rs"),
			},
			eligible: []string{"n1"},
			actuationStatus: &fakeActuationStatus{
				currentlyDrained: []string{"n2"},
			},
			wantUnneeded: []string{},
			wantErr:      true,
		},
		{
			name: "pods from multiple drained nodes can schedule elsewhere",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 2000, 10),
				BuildTestNode("n3", 1000, 10),
				nodeUndergoingDeletion("n4", 2000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 500, 1, "n2", "rs"),
				scheduledPod("p2", 500, 1, "n2", "rs"),
				scheduledPod("p4", 500, 1, "n4", "rs"),
				scheduledPod("p5", 500, 1, "n4", "rs"),
			},
			eligible: []string{"n1", "n3"},
			actuationStatus: &fakeActuationStatus{
				currentlyDrained: []string{"n2", "n4"},
			},
			wantUnneeded: []string{},
		},
		{
			name: "pods from multiple drained nodes can't schedule elsewhere",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 2000, 10),
				BuildTestNode("n3", 1000, 10),
				nodeUndergoingDeletion("n4", 2000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 500, 1, "n2", "rs"),
				scheduledPod("p2", 500, 1, "n2", "rs"),
				scheduledPod("p3", 500, 1, "n2", "rs"),
				scheduledPod("p4", 500, 1, "n4", "rs"),
				scheduledPod("p5", 500, 1, "n4", "rs"),
			},
			eligible: []string{"n1", "n3"},
			actuationStatus: &fakeActuationStatus{
				currentlyDrained: []string{"n2", "n4"},
			},
			wantUnneeded: []string{},
			wantErr:      true,
		},
		{
			name: "multiple drained nodes but new candidates found",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 2000, 10),
				nodeUndergoingDeletion("n2", 2000, 10),
				BuildTestNode("n3", 2000, 10),
				nodeUndergoingDeletion("n4", 2000, 10),
				BuildTestNode("n5", 2000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 400, 1, "n1", "rs"),
				scheduledPod("p2", 400, 1, "n2", "rs"),
				scheduledPod("p3", 400, 1, "n3", "rs"),
				scheduledPod("p4", 400, 1, "n4", "rs"),
				scheduledPod("p5", 400, 1, "n5", "rs"),
			},
			eligible: []string{"n1", "n3", "n5"},
			actuationStatus: &fakeActuationStatus{
				currentlyDrained: []string{"n2", "n4"},
			},
			wantUnneeded: []string{"n1", "n3"},
		},
		{
			name: "recently evicted pods can schedule elsewhere, node uneeded",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 500, 1, "n2", "rs"),
				scheduledPod("p2", 500, 1, "n2", "rs"),
			},
			eligible: []string{"n1", "n2"},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					scheduledPod("p3", 500, 1, "n4", "rs"),
				},
			},
			wantUnneeded: []string{"n1"},
		},
		{
			name: "recently evicted pods can schedule elsewhere, no unneeded",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 500, 1, "n2", "rs"),
				scheduledPod("p2", 500, 1, "n2", "rs"),
			},
			eligible: []string{"n1", "n2"},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					scheduledPod("p3", 500, 1, "n4", "rs"),
					scheduledPod("p4", 500, 1, "n4", "rs"),
					scheduledPod("p5", 500, 1, "n4", "rs"),
				},
			},
			wantUnneeded: []string{},
		},
		{
			name: "recently evicted pods can't schedule elsewhere",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 500, 1, "n1", "rs"),
				scheduledPod("p2", 500, 1, "n1", "rs"),
			},
			eligible: []string{"n1", "n2"},
			actuationStatus: &fakeActuationStatus{
				recentEvictions: []*apiv1.Pod{
					scheduledPod("p3", 500, 1, "n3", "rs"),
					scheduledPod("p4", 500, 1, "n3", "rs"),
					scheduledPod("p5", 500, 1, "n3", "rs"),
				},
			},
			wantUnneeded: []string{},
			wantErr:      true,
		},
		{
			name: "multiple drained nodes and recent evictions, no unneeded",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				nodeUndergoingDeletion("n4", 1000, 10),
				BuildTestNode("n5", 1000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 200, 1, "n1", "rs"),
				scheduledPod("p2", 200, 1, "n2", "rs"),
				scheduledPod("p3", 200, 1, "n3", "rs"),
				scheduledPod("p4", 200, 1, "n4", "rs"),
				scheduledPod("p5", 200, 1, "n5", "rs"),
			},
			eligible: []string{"n1", "n3", "n5"},
			actuationStatus: &fakeActuationStatus{
				currentlyDrained: []string{"n2", "n4"},
				recentEvictions: []*apiv1.Pod{
					scheduledPod("p6", 600, 1, "n6", "rs"),
					scheduledPod("p7", 600, 1, "n6", "rs"),
				},
			},
			wantUnneeded: []string{},
		},
		{
			name: "multiple drained nodes and recent evictions, one unneeded",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				nodeUndergoingDeletion("n4", 1000, 10),
				BuildTestNode("n5", 1000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 200, 1, "n1", "rs"),
				scheduledPod("p2", 200, 1, "n2", "rs"),
				scheduledPod("p3", 200, 1, "n3", "rs"),
				scheduledPod("p4", 200, 1, "n4", "rs"),
				scheduledPod("p5", 200, 1, "n5", "rs"),
			},
			eligible: []string{"n1", "n3", "n5"},
			actuationStatus: &fakeActuationStatus{
				currentlyDrained: []string{"n2", "n4"},
				recentEvictions: []*apiv1.Pod{
					scheduledPod("p6", 600, 1, "n6", "rs"),
				},
			},
			wantUnneeded: []string{"n1"},
		},
		{
			name: "multiple drained nodes and recent evictions, replicas rescheduled, two nodes unneeded",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				nodeUndergoingDeletion("n4", 1000, 10),
				BuildTestNode("n5", 1000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 200, 1, "n1", "rs"),
				scheduledPod("p2", 200, 1, "n2", "rs"),
				scheduledPod("p3", 200, 1, "n3", "rs"),
				scheduledPod("p4", 200, 1, "n4", "rs"),
				scheduledPod("p5", 200, 1, "n5", "rs"),
			},
			eligible: []string{"n1", "n3", "n5"},
			actuationStatus: &fakeActuationStatus{
				currentlyDrained: []string{"n2", "n4"},
				recentEvictions: []*apiv1.Pod{
					scheduledPod("p6", 600, 1, "n1", "rs1"),
					scheduledPod("p7", 600, 1, "n3", "rs1"),
				},
			},
			replicasSets: append(generateReplicaSetWithReplicas("rs1", 2, 2, rSetLabels), generateReplicaSets("rs", 5)...),
			wantUnneeded: []string{"n1", "n3"},
		},
		{
			name: "multiple drained nodes and recent evictions, some replicas rescheduled, one node unneeded",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				nodeUndergoingDeletion("n4", 1000, 10),
				BuildTestNode("n5", 1000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 200, 1, "n1", "rs"),
				scheduledPod("p2", 200, 1, "n2", "rs"),
				scheduledPod("p3", 200, 1, "n3", "rs"),
				scheduledPod("p4", 200, 1, "n4", "rs"),
				scheduledPod("p5", 200, 1, "n5", "rs"),
			},
			eligible: []string{"n1", "n3", "n5"},
			actuationStatus: &fakeActuationStatus{
				currentlyDrained: []string{"n2", "n4"},
				recentEvictions: []*apiv1.Pod{
					scheduledPod("p6", 600, 1, "n1", "rs1"),
					scheduledPod("p7", 600, 1, "n3", "rs1"),
				},
			},
			replicasSets: append(generateReplicaSetWithReplicas("rs1", 2, 1, rSetLabels), generateReplicaSets("rs", 5)...),
			wantUnneeded: []string{"n1"},
		},
		{
			name: "multiple drained nodes and recent evictions, pods belonging to ds",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				nodeUndergoingDeletion("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				nodeUndergoingDeletion("n4", 1000, 10),
				BuildTestNode("n5", 1000, 10),
			},
			pods: []*apiv1.Pod{
				scheduledPod("p1", 200, 1, "n1", "rs"),
				scheduledPod("p2", 200, 1, "n2", "rs"),
				scheduledPod("p3", 200, 1, "n3", "rs"),
				scheduledPod("p4", 200, 1, "n4", "rs"),
				scheduledPod("p5", 200, 1, "n5", "rs"),
			},
			eligible: []string{"n1", "n3", "n5"},
			actuationStatus: &fakeActuationStatus{
				currentlyDrained: []string{"n2", "n4"},
				recentEvictions: []*apiv1.Pod{
					scheduledDSPod("p6", 600, 1, "n1"),
					scheduledDSPod("p7", 600, 1, "n3"),
				},
			},
			wantUnneeded: []string{"n1", "n3"},
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
			context, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{}, &fake.Clientset{}, registry, provider, nil, nil)
			assert.NoError(t, err)
			clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, tc.nodes, tc.pods)
			deleteOptions := simulator.NodeDeleteOptions{}
			p := New(&context, NewTestProcessors(&context), deleteOptions)
			p.eligibilityChecker = &fakeEligibilityChecker{eligible: asMap(tc.eligible)}
			// TODO(x13n): test subsets of nodes passed as podDestinations/scaleDownCandidates.
			aErr := p.UpdateClusterState(tc.nodes, tc.nodes, tc.actuationStatus, nil, time.Now())
			if tc.wantErr {
				assert.Error(t, aErr)
			} else {
				assert.NoError(t, aErr)
			}
			wantUnneeded := asMap(tc.wantUnneeded)
			for _, n := range tc.nodes {
				if wantUnneeded[n.Name] {
					assert.True(t, p.unneededNodes.Contains(n.Name), n.Name)
				} else {
					assert.False(t, p.unneededNodes.Contains(n.Name), n.Name)
				}
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
				UID:       rSetUID(name),
			},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
			},
		},
	}
}

func generateReplicaSetWithReplicas(name string, specReplicas, statusReplicas int32, labels map[string]string) []*appsv1.ReplicaSet {
	return []*appsv1.ReplicaSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
				UID:       rSetUID(name),
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

func rSetUID(name string) types.UID {
	return types.UID(fmt.Sprintf("api/v1/namespaces/default/replicasets/%s", name))
}

func scheduledDSPod(name string, cpu, memory int64, nodeName string) *apiv1.Pod {
	p := BuildTestPod(name, cpu, memory)
	p.OwnerReferences = GenerateOwnerReferences("ds", "DaemonSet", "extensions/v1beta1", "api/v1/namespaces/default/daemonsets/ds")
	p.Spec.NodeName = nodeName
	p.Namespace = "default"
	p.Labels = rSetLabels
	return p
}

func scheduledPod(name string, cpu, memory int64, nodeName, rSetName string) *apiv1.Pod {
	p := BuildTestPod(name, cpu, memory)
	p.OwnerReferences = GenerateOwnerReferences(rSetName, "ReplicaSet", "extensions/v1beta1", rSetUID(rSetName))
	p.Spec.NodeName = nodeName
	p.Namespace = "default"
	p.Labels = rSetLabels
	return p
}

func nodeUndergoingDeletion(name string, cpu, memory int64) *apiv1.Node {
	n := BuildTestNode(name, cpu, memory)
	toBeDeletedTaint := apiv1.Taint{Key: deletetaint.ToBeDeletedTaint, Effect: apiv1.TaintEffectNoSchedule}
	n.Spec.Taints = append(n.Spec.Taints, toBeDeletedTaint)
	return n
}

type fakeActuationStatus struct {
	recentEvictions  []*apiv1.Pod
	currentlyDrained []string
}

func (f *fakeActuationStatus) RecentEvictions() []*apiv1.Pod {
	return f.recentEvictions
}

func (f *fakeActuationStatus) DeletionsInProgress() ([]string, []string) {
	return nil, f.currentlyDrained
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
