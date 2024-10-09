/*
Copyright 2019 The Kubernetes Authors.

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

package utils

import (
	"testing"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
)

func TestFilterOutExpendableAndSplit(t *testing.T) {
	var priority1 int32 = 1
	var priority100 int32 = 100

	p1 := BuildTestPod("p1", 1000, 200000)
	p1.Spec.Priority = &priority1
	p2 := BuildTestPod("p2", 1000, 200000)
	p2.Spec.Priority = &priority100
	n1 := BuildTestNode("node1", 10, 10)
	n2 := BuildTestNode("node2", 10, 10)

	podWaitingForPreemption1 := BuildTestPod("w1", 1000, 200000)
	podWaitingForPreemption1.Spec.Priority = &priority1
	podWaitingForPreemption1.Status.NominatedNodeName = "node1"
	podWaitingForPreemption2 := BuildTestPod("w2", 1000, 200000)
	podWaitingForPreemption2.Spec.Priority = &priority100
	podWaitingForPreemption2.Status.NominatedNodeName = "node2"

	res1, res2 := FilterOutExpendableAndSplit([]*apiv1.Pod{p1, p2, podWaitingForPreemption1, podWaitingForPreemption2}, []*apiv1.Node{n1, n2}, 0)
	assert.Equal(t, 2, len(res1))
	assert.Equal(t, p1, res1[0])
	assert.Equal(t, p2, res1[1])
	assert.Equal(t, 2, len(res2))
	assert.Equal(t, podWaitingForPreemption1, res2[0])
	assert.Equal(t, podWaitingForPreemption2, res2[1])

	res1, res2 = FilterOutExpendableAndSplit([]*apiv1.Pod{p1, p2, podWaitingForPreemption1, podWaitingForPreemption2}, []*apiv1.Node{n1, n2}, 10)
	assert.Equal(t, 1, len(res1))
	assert.Equal(t, p2, res1[0])
	assert.Equal(t, 1, len(res2))
	assert.Equal(t, podWaitingForPreemption2, res2[0])

	// if node2 is missing podWaitingForPreemption2 should be treated as standard pod not one waiting for preemption
	res1, res2 = FilterOutExpendableAndSplit([]*apiv1.Pod{p1, p2, podWaitingForPreemption1, podWaitingForPreemption2}, []*apiv1.Node{n1}, 0)
	assert.Equal(t, 3, len(res1))
	assert.Equal(t, p1, res1[0])
	assert.Equal(t, p2, res1[1])
	assert.Equal(t, podWaitingForPreemption2, res1[2])
	assert.Equal(t, 1, len(res2))
	assert.Equal(t, podWaitingForPreemption1, res2[0])
}

func TestSplitExpendablePods(t *testing.T) {
	p1 := BuildTestPod("p1", 1500, 200000)
	p2 := BuildTestPod("p2", 3000, 200000)

	podWaitingForPreemption1 := BuildTestPod("w1", 1500, 200000)
	var priority1 int32 = -10
	podWaitingForPreemption1.Spec.Priority = &priority1
	podWaitingForPreemption1.Status.NominatedNodeName = "node1"

	podWaitingForPreemption2 := BuildTestPod("w1", 1500, 200000)
	var priority2 int32 = 10
	podWaitingForPreemption2.Spec.Priority = &priority2
	podWaitingForPreemption2.Status.NominatedNodeName = "node1"

	exp, nonExp := SplitExpendablePods([]*apiv1.Pod{p1, p2, podWaitingForPreemption1, podWaitingForPreemption2}, 0)
	assert.Equal(t, 3, len(nonExp))
	assert.Equal(t, p1, nonExp[0])
	assert.Equal(t, p2, nonExp[1])
	assert.Equal(t, podWaitingForPreemption2, nonExp[2])
	assert.Equal(t, 1, len(exp))
	assert.Equal(t, podWaitingForPreemption1, exp[0])
}

func TestIsExpandablePod(t *testing.T) {
	preemptLowerPriorityPolicy := apiv1.PreemptLowerPriority
	neverPolicy := apiv1.PreemptNever

	testCases := []struct {
		name   string
		pod    *apiv1.Pod
		cutoff int
		want   bool
	}{
		{
			name:   "pod priority not set, zero cutoff",
			pod:    BuildTestPod("p", 0, 0),
			cutoff: 0,
			want:   false,
		},
		{
			name:   "pod priority not set, negative cutoff",
			pod:    BuildTestPod("p", 0, 0),
			cutoff: -1,
			want:   false,
		},
		{
			name:   "pod priority set, default preemption policy, higher cutoff",
			pod:    withPodPriority(BuildTestPod("p", 0, 0), -1, nil),
			cutoff: 0,
			want:   true,
		},
		{
			name:   "pod priority set, default preemption policy, equal cutoff",
			pod:    withPodPriority(BuildTestPod("p", 0, 0), -1, nil),
			cutoff: -1,
			want:   false,
		},
		{
			name:   "pod priority set, default preemption policy, smaller cutoff",
			pod:    withPodPriority(BuildTestPod("p", 0, 0), -1, nil),
			cutoff: -2,
			want:   false,
		},
		{
			name:   "pod priority set, preempt lower priority preemption policy, higher cutoff",
			pod:    withPodPriority(BuildTestPod("p", 0, 0), -1, &preemptLowerPriorityPolicy),
			cutoff: 0,
			want:   true,
		},
		{
			name:   "pod priority set, preempt lower priority preemption policy, equal cutoff",
			pod:    withPodPriority(BuildTestPod("p", 0, 0), -1, &preemptLowerPriorityPolicy),
			cutoff: -1,
			want:   false,
		},
		{
			name:   "pod priority set, preempt lower priority preemption policy, smaller cutoff",
			pod:    withPodPriority(BuildTestPod("p", 0, 0), -1, &preemptLowerPriorityPolicy),
			cutoff: -2,
			want:   false,
		},
		{
			name:   "pod priority set, never preemption policy, higher cutoff",
			pod:    withPodPriority(BuildTestPod("p", 0, 0), -1, &neverPolicy),
			cutoff: 0,
			want:   false,
		},
		{
			name:   "pod priority set, never preemption policy, equal cutoff",
			pod:    withPodPriority(BuildTestPod("p", 0, 0), -1, &neverPolicy),
			cutoff: -1,
			want:   false,
		},
		{
			name:   "pod priority set, never preemption policy, smaller cutoff",
			pod:    withPodPriority(BuildTestPod("p", 0, 0), -1, &neverPolicy),
			cutoff: -2,
			want:   false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsExpendablePod(tc.pod, tc.cutoff))
		})
	}
}

func withPodPriority(pod *apiv1.Pod, priority int32, preemptionPolicy *apiv1.PreemptionPolicy) *apiv1.Pod {
	pod.Spec.Priority = &priority
	pod.Spec.PreemptionPolicy = preemptionPolicy
	return pod
}
