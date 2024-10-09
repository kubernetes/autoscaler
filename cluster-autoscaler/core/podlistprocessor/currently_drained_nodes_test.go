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

package podlistprocessor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestCurrentlyDrainedNodesPodListProcessor(t *testing.T) {
	testCases := []struct {
		name         string
		drainedNodes []string
		nodes        []*apiv1.Node
		pods         []*apiv1.Pod

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
			name:         "single node undergoing deletion",
			drainedNodes: []string{"n"},
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
			name:         "single node undergoing deletion, pods with deletion timestamp set",
			drainedNodes: []string{"n"},
			nodes: []*apiv1.Node{
				BuildTestNode("n", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildTestPod("p2", 200, 1, WithNodeName("n"), WithDeletionTimestamp(time.Now())),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
			},
		},
		{
			name:         "single empty node undergoing deletion",
			drainedNodes: []string{"n"},
			nodes: []*apiv1.Node{
				BuildTestNode("n", 1000, 10),
			},
		},
		{
			name:         "single node undergoing deletion, unschedulable pods",
			drainedNodes: []string{"n"},
			nodes: []*apiv1.Node{
				BuildTestNode("n", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 200, 1, "n"),
			},
			unschedulablePods: []*apiv1.Pod{
				BuildTestPod("p3", 300, 1),
				BuildTestPod("p4", 400, 1),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				BuildTestPod("p2", 200, 1),
				BuildTestPod("p3", 300, 1),
				BuildTestPod("p4", 400, 1),
			},
		},
		{
			name: "single ready node",
			nodes: []*apiv1.Node{
				BuildTestNode("n", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 200, 1, "n"),
			},
		},
		{
			name: "single ready node, unschedulable pods",
			nodes: []*apiv1.Node{
				BuildTestNode("n", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 200, 1, "n"),
			},
			unschedulablePods: []*apiv1.Pod{
				BuildTestPod("p3", 300, 1),
				BuildTestPod("p4", 400, 1),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p3", 300, 1),
				BuildTestPod("p4", 400, 1),
			},
		},
		{
			name:         "multiple nodes, all undergoing deletion",
			drainedNodes: []string{"n1", "n2"},
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n1"),
				BuildScheduledTestPod("p2", 200, 1, "n1"),
				BuildScheduledTestPod("p3", 300, 1, "n2"),
				BuildScheduledTestPod("p4", 400, 1, "n2"),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				BuildTestPod("p2", 200, 1),
				BuildTestPod("p3", 300, 1),
				BuildTestPod("p4", 400, 1),
			},
		},
		{
			name:         "multiple nodes, some undergoing deletion",
			drainedNodes: []string{"n1"},
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n1"),
				BuildScheduledTestPod("p2", 200, 1, "n1"),
				BuildScheduledTestPod("p3", 300, 1, "n2"),
				BuildScheduledTestPod("p4", 400, 1, "n2"),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				BuildTestPod("p2", 200, 1),
			},
		},
		{
			name: "multiple nodes, no undergoing deletion",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n1"),
				BuildScheduledTestPod("p2", 200, 1, "n1"),
				BuildScheduledTestPod("p3", 300, 1, "n2"),
				BuildScheduledTestPod("p4", 400, 1, "n2"),
			},
		},
		{
			name:         "single node, non-recreatable pods filtered out",
			drainedNodes: []string{"n"},
			nodes: []*apiv1.Node{
				BuildTestNode("n", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				SetRSPodSpec(BuildScheduledTestPod("p2", 200, 1, "n"), "rs"),
				SetDSPodSpec(BuildScheduledTestPod("p3", 300, 1, "n")),
				SetMirrorPodSpec(BuildScheduledTestPod("p4", 400, 1, "n")),
				SetStaticPodSpec(BuildScheduledTestPod("p5", 500, 1, "n")),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				SetRSPodSpec(BuildTestPod("p2", 200, 1), "rs"),
			},
		},
		{
			name: "unschedulable pods, non-recreatable pods not filtered out",
			unschedulablePods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				SetRSPodSpec(BuildTestPod("p2", 200, 1), "rs"),
				SetDSPodSpec(BuildTestPod("p3", 300, 1)),
				SetMirrorPodSpec(BuildTestPod("p4", 400, 1)),
				SetStaticPodSpec(BuildTestPod("p5", 500, 1)),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				SetRSPodSpec(BuildTestPod("p2", 200, 1), "rs"),
				SetDSPodSpec(BuildTestPod("p3", 300, 1)),
				SetMirrorPodSpec(BuildTestPod("p4", 400, 1)),
				SetStaticPodSpec(BuildTestPod("p5", 500, 1)),
			},
		},
		{
			name:         "everything works together",
			drainedNodes: []string{"n1", "n3", "n5"},
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
				BuildTestNode("n3", 1000, 10),
				BuildTestNode("n4", 1000, 10),
				BuildTestNode("n5", 1000, 10),
			},
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n1"),
				BuildScheduledTestPod("p2", 200, 1, "n1"),
				SetRSPodSpec(BuildScheduledTestPod("p3", 300, 1, "n1"), "rs"),
				SetDSPodSpec(BuildScheduledTestPod("p4", 400, 1, "n1")),
				BuildScheduledTestPod("p5", 500, 1, "n2"),
				BuildScheduledTestPod("p6", 600, 1, "n2"),
				BuildScheduledTestPod("p7", 700, 1, "n3"),
				SetStaticPodSpec(BuildScheduledTestPod("p8", 800, 1, "n3")),
				SetMirrorPodSpec(BuildScheduledTestPod("p9", 900, 1, "n3")),
			},
			unschedulablePods: []*apiv1.Pod{
				BuildTestPod("p10", 1000, 1),
				SetMirrorPodSpec(BuildTestPod("p11", 1100, 1)),
				SetStaticPodSpec(BuildTestPod("p12", 1200, 1)),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				BuildTestPod("p2", 200, 1),
				SetRSPodSpec(BuildTestPod("p3", 300, 1), "rs"),
				BuildTestPod("p7", 700, 1),
				BuildTestPod("p10", 1000, 1),
				SetMirrorPodSpec(BuildTestPod("p11", 1100, 1)),
				SetStaticPodSpec(BuildTestPod("p12", 1200, 1)),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.AutoscalingContext{
				ScaleDownActuator: &mockActuator{&mockActuationStatus{tc.drainedNodes}},
				ClusterSnapshot:   clustersnapshot.NewBasicClusterSnapshot(framework.TestFrameworkHandleOrDie(t), true),
			}
			clustersnapshot.InitializeClusterSnapshotOrDie(t, ctx.ClusterSnapshot, tc.nodes, tc.pods)

			processor := NewCurrentlyDrainedNodesPodListProcessor()
			pods, err := processor.Process(&ctx, tc.unschedulablePods)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.wantPods, pods)
		})
	}
}

type mockActuator struct {
	status *mockActuationStatus
}

func (m *mockActuator) StartDeletion(_, _ []*apiv1.Node) (status.ScaleDownResult, []*status.ScaleDownNode, errors.AutoscalerError) {
	return status.ScaleDownError, []*status.ScaleDownNode{}, nil
}

func (m *mockActuator) CheckStatus() scaledown.ActuationStatus {
	return m.status
}

func (m *mockActuator) ClearResultsNotNewerThan(time.Time) {

}

func (m *mockActuator) DeletionResults() (map[string]status.NodeDeleteResult, time.Time) {
	return map[string]status.NodeDeleteResult{}, time.Now()
}

type mockActuationStatus struct {
	drainedNodes []string
}

func (m *mockActuationStatus) RecentEvictions() []*apiv1.Pod {
	return nil
}

func (m *mockActuationStatus) DeletionsInProgress() ([]string, []string) {
	return nil, m.drainedNodes
}

func (m *mockActuationStatus) DeletionResults() (map[string]status.NodeDeleteResult, time.Time) {
	return nil, time.Time{}
}

func (m *mockActuationStatus) DeletionsCount(_ string) int {
	return 0
}
