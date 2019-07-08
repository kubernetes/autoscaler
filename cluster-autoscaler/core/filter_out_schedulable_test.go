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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/api/testapi"

	"github.com/stretchr/testify/assert"
)

func TestFilterOutSchedulableByPacking(t *testing.T) {
	rc1 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc1",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
			UID:       "12345678-1234-1234-1234-123456789012",
		},
	}

	rc2 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc2",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
			UID:       "12345678-1234-1234-1234-12345678901a",
		},
	}

	p1 := BuildTestPod("p1", 1500, 200000)
	p2_1 := BuildTestPod("p2_2", 3000, 200000)
	p2_1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p2_2 := BuildTestPod("p2_2", 3000, 200000)
	p2_2.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p3_1 := BuildTestPod("p3", 300, 200000)
	p3_1.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	p3_2 := BuildTestPod("p3", 300, 200000)
	p3_2.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	unschedulablePods := []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2}

	scheduledPod1 := BuildTestPod("s1", 100, 200000)
	scheduledPod2 := BuildTestPod("s2", 1500, 200000)
	scheduledPod3 := BuildTestPod("s3", 4000, 200000)
	var priority1 int32 = 1
	scheduledPod3.Spec.Priority = &priority1
	scheduledPod1.Spec.NodeName = "node1"
	scheduledPod2.Spec.NodeName = "node1"
	scheduledPod2.Spec.NodeName = "node1"

	podWaitingForPreemption := BuildTestPod("w1", 1500, 200000)
	var priority100 int32 = 100
	podWaitingForPreemption.Spec.Priority = &priority100
	podWaitingForPreemption.Status.NominatedNodeName = "node1"

	p4 := BuildTestPod("p4", 1800, 200000)
	p4.Spec.Priority = &priority100

	node := BuildTestNode("node1", 2000, 2000000)
	SetNodeReadyState(node, true, time.Time{})

	predicateChecker := simulator.NewTestPredicateChecker()

	res := filterOutSchedulableByPacking(unschedulablePods, []*apiv1.Node{node}, []*apiv1.Pod{scheduledPod1, scheduledPod3}, predicateChecker, 10, true)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, p2_1, res[0])
	assert.Equal(t, p2_2, res[1])
	assert.Equal(t, p3_2, res[2])

	res2 := filterOutSchedulableByPacking(unschedulablePods, []*apiv1.Node{node}, []*apiv1.Pod{scheduledPod1, scheduledPod2, scheduledPod3}, predicateChecker, 10, true)
	assert.Equal(t, 4, len(res2))
	assert.Equal(t, p1, res2[0])
	assert.Equal(t, p2_1, res2[1])
	assert.Equal(t, p2_2, res2[2])
	assert.Equal(t, p3_2, res2[3])

	res4 := filterOutSchedulableByPacking(append(unschedulablePods, p4), []*apiv1.Node{node}, []*apiv1.Pod{scheduledPod1, scheduledPod3}, predicateChecker, 10, true)
	assert.Equal(t, 5, len(res4))
	assert.Equal(t, p1, res4[0])
	assert.Equal(t, p2_1, res4[1])
	assert.Equal(t, p2_2, res4[2])
	assert.Equal(t, p3_1, res4[3])
	assert.Equal(t, p3_2, res4[4])
}

func TestFilterOutSchedulableSimple(t *testing.T) {
	rc1 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc1",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
			UID:       "12345678-1234-1234-1234-123456789012",
		},
	}

	rc2 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc2",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
			UID:       "12345678-1234-1234-1234-12345678901a",
		},
	}

	p1 := BuildTestPod("p1", 1500, 200000)
	p2_1 := BuildTestPod("p2_2", 3000, 200000)
	p2_1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p2_2 := BuildTestPod("p2_2", 3000, 200000)
	p2_2.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p3_1 := BuildTestPod("p3", 100, 200000)
	p3_1.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	p3_2 := BuildTestPod("p3", 100, 200000)
	p3_2.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	unschedulablePods := []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2}

	scheduledPod1 := BuildTestPod("s1", 100, 200000)
	scheduledPod2 := BuildTestPod("s2", 1500, 200000)
	scheduledPod3 := BuildTestPod("s3", 4000, 200000)
	var priority1 int32 = 1
	scheduledPod3.Spec.Priority = &priority1
	scheduledPod1.Spec.NodeName = "node1"
	scheduledPod2.Spec.NodeName = "node1"
	scheduledPod2.Spec.NodeName = "node1"

	podWaitingForPreemption := BuildTestPod("w1", 1500, 200000)
	var priority100 int32 = 100
	podWaitingForPreemption.Spec.Priority = &priority100
	podWaitingForPreemption.Status.NominatedNodeName = "node1"

	node := BuildTestNode("node1", 2000, 2000000)
	SetNodeReadyState(node, true, time.Time{})

	predicateChecker := simulator.NewTestPredicateChecker()

	res := filterOutSchedulableSimple(unschedulablePods, []*apiv1.Node{node}, []*apiv1.Pod{scheduledPod1, scheduledPod3}, predicateChecker, 10)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, p2_1, res[0])
	assert.Equal(t, p2_2, res[1])

	res2 := filterOutSchedulableSimple(unschedulablePods, []*apiv1.Node{node}, []*apiv1.Pod{scheduledPod1, scheduledPod2, scheduledPod3}, predicateChecker, 10)
	assert.Equal(t, 3, len(res2))
	assert.Equal(t, p1, res2[0])
	assert.Equal(t, p2_1, res2[1])
	assert.Equal(t, p2_2, res2[2])

}
