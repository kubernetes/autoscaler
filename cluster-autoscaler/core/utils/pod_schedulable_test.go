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

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/api/testapi"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/stretchr/testify/assert"
	"time"
)

func TestPodSchedulableMap(t *testing.T) {
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

	pMap := make(PodSchedulableMap)

	podInRc1_1 := BuildTestPod("podInRc1_1", 500, 1000)
	podInRc1_1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)

	podInRc2 := BuildTestPod("podInRc2", 500, 1000)
	podInRc2.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)

	// Basic sanity checks
	_, found := pMap.Get(podInRc1_1)
	assert.False(t, found)
	pMap.Set(podInRc1_1, nil)
	err, found := pMap.Get(podInRc1_1)
	assert.True(t, found)
	assert.Nil(t, err)

	cpuErr := &simulator.PredicateError{}

	// Pod in different RC
	_, found = pMap.Get(podInRc2)
	assert.False(t, found)
	pMap.Set(podInRc2, cpuErr)
	err, found = pMap.Get(podInRc2)
	assert.True(t, found)
	assert.Equal(t, cpuErr, err)

	// Another replica in rc1
	podInRc1_2 := BuildTestPod("podInRc1_1", 500, 1000)
	podInRc1_2.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	err, found = pMap.Get(podInRc1_2)
	assert.True(t, found)
	assert.Nil(t, err)

	// A pod in rc1, but with different requests
	differentPodInRc1 := BuildTestPod("differentPodInRc1", 1000, 1000)
	differentPodInRc1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	_, found = pMap.Get(differentPodInRc1)
	assert.False(t, found)
	pMap.Set(differentPodInRc1, cpuErr)
	err, found = pMap.Get(differentPodInRc1)
	assert.True(t, found)
	assert.Equal(t, cpuErr, err)

	// A non-replicated pod
	nonReplicatedPod := BuildTestPod("nonReplicatedPod", 1000, 1000)
	_, found = pMap.Get(nonReplicatedPod)
	assert.False(t, found)
	pMap.Set(nonReplicatedPod, err)
	_, found = pMap.Get(nonReplicatedPod)
	assert.False(t, found)

	// Verify information about first pod has not been overwritten by adding
	// other pods
	err, found = pMap.Get(podInRc1_1)
	assert.True(t, found)
	assert.Nil(t, err)
}

func TestFilterSchedulablePodsForNode(t *testing.T) {
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
	p2_1 := BuildTestPod("p2_1", 3000, 200000)
	p2_1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p2_2 := BuildTestPod("p2_2", 3000, 200000)
	p2_2.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p3_1 := BuildTestPod("p3_1", 100, 200000)
	p3_1.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	p3_2 := BuildTestPod("p3_2", 100, 200000)
	p3_2.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	unschedulablePods := []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2}

	tn := BuildTestNode("T1-abc", 2000, 2000000)
	SetNodeReadyState(tn, true, time.Time{})
	tni := schedulernodeinfo.NewNodeInfo()
	tni.SetNode(tn)

	context := &context.AutoscalingContext{
		PredicateChecker: simulator.NewTestPredicateChecker(),
	}

	checker := NewPodsSchedulableOnNodeChecker(context, unschedulablePods)
	res := checker.CheckPodsSchedulableOnNode("T1-abc", tni)
	wantedSchedulable := []*apiv1.Pod{p1, p3_1, p3_2}
	wantedUnschedulable := []*apiv1.Pod{p2_1, p2_2}

	assert.Equal(t, 5, len(res))
	for _, pod := range wantedSchedulable {
		err, found := res[pod]
		assert.True(t, found)
		assert.Nil(t, err)
	}
	for _, pod := range wantedUnschedulable {
		err, found := res[pod]
		assert.True(t, found)
		assert.NotNil(t, err)
	}
}
