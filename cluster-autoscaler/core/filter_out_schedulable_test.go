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

package core

import (
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestFilterOutSchedulableByPacking(t *testing.T) {
	// TODO(scheduler_framework_integration) extend/cleanup the test
	// - add more nodes
	// - add better naming for pods/scenarios

	p1 := BuildTestPod("p1", 1500, 200000)
	p2_1 := BuildTestPod("p2_2", 3000, 200000)
	p2_2 := BuildTestPod("p2_2", 3000, 200000)
	p3_1 := BuildTestPod("p3_1", 300, 200000)
	p3_2 := BuildTestPod("p3_2", 300, 200000)

	scheduledPod1 := BuildTestPod("s1", 100, 200000)
	scheduledPod1.Spec.NodeName = "node1"
	scheduledPod2 := BuildTestPod("s2", 1500, 200000)
	scheduledPod2.Spec.NodeName = "node1"
	scheduledPod3 := BuildTestPod("s3", 4000, 200000)
	scheduledPod3.Spec.NodeName = "node1"
	var priority1 int32 = 1
	scheduledPod3.Spec.Priority = &priority1

	podWaitingForPreemption := BuildTestPod("w1", 1500, 200000)
	var priority100 int32 = 100
	podWaitingForPreemption.Spec.Priority = &priority100
	podWaitingForPreemption.Status.NominatedNodeName = "node1"

	p4 := BuildTestPod("p4", 1800, 200000)
	p4.Spec.Priority = &priority100

	node := BuildTestNode("node1", 2000, 2000000)
	SetNodeReadyState(node, true, time.Time{})

	tests := []struct {
		name                   string
		nodes                  []*apiv1.Node
		scheduledPods          []*apiv1.Pod
		pendingPods            []*apiv1.Pod
		expectedPendingPods    []*apiv1.Pod
		expectedPodsInSnapshot []*apiv1.Pod
	}{
		{
			name:                   "scenario 1",
			nodes:                  []*apiv1.Node{node},
			scheduledPods:          []*apiv1.Pod{scheduledPod1, scheduledPod3},
			pendingPods:            []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2},
			expectedPendingPods:    []*apiv1.Pod{p2_1, p2_2, p3_2},
			expectedPodsInSnapshot: []*apiv1.Pod{scheduledPod1, p1, p3_1},
		},
		{
			name:                   "scenario 2",
			nodes:                  []*apiv1.Node{node},
			scheduledPods:          []*apiv1.Pod{scheduledPod1, scheduledPod2, scheduledPod3},
			pendingPods:            []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2},
			expectedPendingPods:    []*apiv1.Pod{p1, p2_1, p2_2, p3_2},
			expectedPodsInSnapshot: []*apiv1.Pod{scheduledPod1, scheduledPod2, p3_1},
		},
		{
			name:                   "scenario 3",
			nodes:                  []*apiv1.Node{node},
			scheduledPods:          []*apiv1.Pod{scheduledPod1, scheduledPod3},
			pendingPods:            []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2, p4},
			expectedPendingPods:    []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2},
			expectedPodsInSnapshot: []*apiv1.Pod{scheduledPod1, p4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicateChecker, err := simulator.NewTestPredicateChecker()
			clusterSnapshot := simulator.NewBasicClusterSnapshot()

			for _, node := range tt.nodes {
				err := clusterSnapshot.AddNode(node)
				assert.NoError(t, err)
			}

			for _, pod := range tt.scheduledPods {
				err = clusterSnapshot.AddPod(pod, pod.Spec.NodeName)
				assert.NoError(t, err)
			}

			stillPendingPods, err := filterOutSchedulableByPacking(tt.pendingPods, clusterSnapshot, predicateChecker, 10)
			assert.NoError(t, err)
			assert.ElementsMatch(t, stillPendingPods, tt.expectedPendingPods)

			// Check if snapshot was correctly modified
			podsInSnapshot, _ := clusterSnapshot.GetAllPods()
			assert.ElementsMatch(t, podsInSnapshot, tt.expectedPodsInSnapshot)
		})
	}
}
