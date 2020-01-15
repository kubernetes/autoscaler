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

func TestFilterOutExpendablePods(t *testing.T) {
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

	res := FilterOutExpendablePods([]*apiv1.Pod{p1, p2, podWaitingForPreemption1, podWaitingForPreemption2}, 0)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, p1, res[0])
	assert.Equal(t, p2, res[1])
	assert.Equal(t, podWaitingForPreemption2, res[2])
}

func TestIsExpendablePod(t *testing.T) {
	pod1 := BuildTestPod("p1", 1500, 200000)
	pod2 := BuildTestPod("w1", 1500, 200000)
	var priority1 int32 = -10
	pod2.Spec.Priority = &priority1
	pod2.Status.NominatedNodeName = "node1"

	assert.False(t, IsExpendablePod(pod1, 0))
	assert.False(t, IsExpendablePod(pod1, -9))
	assert.False(t, IsExpendablePod(pod1, -10))
	assert.False(t, IsExpendablePod(pod1, -11))
	assert.True(t, IsExpendablePod(pod2, 0))
	assert.True(t, IsExpendablePod(pod2, -9))
	assert.False(t, IsExpendablePod(pod2, -10))
	assert.False(t, IsExpendablePod(pod2, -11))

}
