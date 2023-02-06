/*
Copyright 2020 The Kubernetes Authors.

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

package taints

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	fakeClient := buildFakeClientWithConflicts(t, node)
	err := MarkToBeDeleted(node, fakeClient, false)
	assert.NoError(t, err)

	updatedNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(updatedNode))
	assert.False(t, HasDeletionCandidateTaint(updatedNode))
}

func TestSoftMarkNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	fakeClient := buildFakeClientWithConflicts(t, node)
	err := MarkDeletionCandidate(node, fakeClient)
	assert.NoError(t, err)

	updatedNode := getNode(t, fakeClient, "node")
	assert.False(t, HasToBeDeletedTaint(updatedNode))
	assert.True(t, HasDeletionCandidateTaint(updatedNode))
}

func TestCheckNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	addTaintToSpec(node, ToBeDeletedTaint, apiv1.TaintEffectNoSchedule, false)
	fakeClient := buildFakeClientWithConflicts(t, node)

	updatedNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(updatedNode))
	assert.False(t, HasDeletionCandidateTaint(updatedNode))
}

func TestSoftCheckNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	addTaintToSpec(node, DeletionCandidateTaint, apiv1.TaintEffectPreferNoSchedule, false)
	fakeClient := buildFakeClientWithConflicts(t, node)

	updatedNode := getNode(t, fakeClient, "node")
	assert.False(t, HasToBeDeletedTaint(updatedNode))
	assert.True(t, HasDeletionCandidateTaint(updatedNode))
}

func TestQueryNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	fakeClient := buildFakeClientWithConflicts(t, node)
	err := MarkToBeDeleted(node, fakeClient, false)
	assert.NoError(t, err)

	updatedNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(updatedNode))

	val, err := GetToBeDeletedTime(updatedNode)
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.True(t, time.Now().Sub(*val) < 10*time.Second)
}

func TestSoftQueryNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	fakeClient := buildFakeClientWithConflicts(t, node)
	err := MarkDeletionCandidate(node, fakeClient)
	assert.NoError(t, err)

	updatedNode := getNode(t, fakeClient, "node")
	assert.True(t, HasDeletionCandidateTaint(updatedNode))

	val, err := GetDeletionCandidateTime(updatedNode)
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.True(t, time.Now().Sub(*val) < 10*time.Second)
}

func TestCleanNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	addTaintToSpec(node, ToBeDeletedTaint, apiv1.TaintEffectNoSchedule, false)
	fakeClient := buildFakeClientWithConflicts(t, node)

	updatedNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(updatedNode))
	assert.False(t, updatedNode.Spec.Unschedulable)

	cleaned, err := CleanToBeDeleted(node, fakeClient, false)
	assert.True(t, cleaned)
	assert.NoError(t, err)

	updatedNode = getNode(t, fakeClient, "node")
	assert.NoError(t, err)
	assert.False(t, HasToBeDeletedTaint(updatedNode))
	assert.False(t, updatedNode.Spec.Unschedulable)
}

func TestCleanNodesWithCordon(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	addTaintToSpec(node, ToBeDeletedTaint, apiv1.TaintEffectNoSchedule, true)
	fakeClient := buildFakeClientWithConflicts(t, node)

	updatedNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(updatedNode))
	assert.True(t, updatedNode.Spec.Unschedulable)

	cleaned, err := CleanToBeDeleted(node, fakeClient, true)
	assert.True(t, cleaned)
	assert.NoError(t, err)

	updatedNode = getNode(t, fakeClient, "node")
	assert.NoError(t, err)
	assert.False(t, HasToBeDeletedTaint(updatedNode))
	assert.False(t, updatedNode.Spec.Unschedulable)
}

func TestCleanNodesWithCordonOnOff(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	addTaintToSpec(node, ToBeDeletedTaint, apiv1.TaintEffectNoSchedule, true)
	fakeClient := buildFakeClientWithConflicts(t, node)

	updatedNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(updatedNode))
	assert.True(t, updatedNode.Spec.Unschedulable)

	cleaned, err := CleanToBeDeleted(node, fakeClient, false)
	assert.True(t, cleaned)
	assert.NoError(t, err)

	updatedNode = getNode(t, fakeClient, "node")
	assert.NoError(t, err)
	assert.False(t, HasToBeDeletedTaint(updatedNode))
	assert.True(t, updatedNode.Spec.Unschedulable)
}

func TestSoftCleanNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	addTaintToSpec(node, DeletionCandidateTaint, apiv1.TaintEffectPreferNoSchedule, false)
	fakeClient := buildFakeClientWithConflicts(t, node)

	updatedNode := getNode(t, fakeClient, "node")
	assert.True(t, HasDeletionCandidateTaint(updatedNode))

	cleaned, err := CleanDeletionCandidate(node, fakeClient)
	assert.True(t, cleaned)
	assert.NoError(t, err)

	updatedNode = getNode(t, fakeClient, "node")
	assert.NoError(t, err)
	assert.False(t, HasDeletionCandidateTaint(updatedNode))
}

func TestCleanAllToBeDeleted(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 10)
	n2 := BuildTestNode("n2", 1000, 10)
	n2.Spec.Taints = []apiv1.Taint{{Key: ToBeDeletedTaint, Value: strconv.FormatInt(time.Now().Unix()-301, 10)}}

	fakeClient := buildFakeClient(t, n1, n2)
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient, false)

	assert.Equal(t, 1, len(getNode(t, fakeClient, "n2").Spec.Taints))

	CleanAllToBeDeleted([]*apiv1.Node{n1, n2}, fakeClient, fakeRecorder, false)

	assert.Equal(t, 0, len(getNode(t, fakeClient, "n1").Spec.Taints))
	assert.Equal(t, 0, len(getNode(t, fakeClient, "n2").Spec.Taints))
}

func TestCleanAllDeletionCandidates(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 10)
	n2 := BuildTestNode("n2", 1000, 10)
	n2.Spec.Taints = []apiv1.Taint{{Key: DeletionCandidateTaint, Value: strconv.FormatInt(time.Now().Unix()-301, 10)}}

	fakeClient := buildFakeClient(t, n1, n2)
	fakeRecorder := kube_util.CreateEventRecorder(fakeClient, false)

	assert.Equal(t, 1, len(getNode(t, fakeClient, "n2").Spec.Taints))

	CleanAllDeletionCandidates([]*apiv1.Node{n1, n2}, fakeClient, fakeRecorder)

	assert.Equal(t, 0, len(getNode(t, fakeClient, "n1").Spec.Taints))
	assert.Equal(t, 0, len(getNode(t, fakeClient, "n2").Spec.Taints))
}

func setConflictRetryInterval(interval time.Duration) time.Duration {
	before := conflictRetryInterval
	conflictRetryInterval = interval
	return before
}

func getNode(t *testing.T, client kube_client.Interface, name string) *apiv1.Node {
	t.Helper()
	node, err := client.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to retrieve node %v: %v", name, err)
	}
	return node
}

func buildFakeClient(t *testing.T, nodes ...*apiv1.Node) *fake.Clientset {
	t.Helper()
	fakeClient := fake.NewSimpleClientset()

	for _, node := range nodes {
		_, err := fakeClient.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{})
		assert.NoError(t, err)
	}

	return fakeClient
}

func buildFakeClientWithConflicts(t *testing.T, nodes ...*apiv1.Node) *fake.Clientset {
	fakeClient := buildFakeClient(t, nodes...)

	// return a 'Conflict' error on the first upadte, then pass it through, then return a Conflict again
	var returnedConflict int32
	fakeClient.Fake.PrependReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)

		if atomic.LoadInt32(&returnedConflict) == 0 {
			// allow the next update
			atomic.StoreInt32(&returnedConflict, 1)
			return true, nil, errors.NewConflict(apiv1.Resource("node"), obj.GetName(), fmt.Errorf("concurrent update on %s", obj.GetName()))
		}

		// return a conflict on next update
		atomic.StoreInt32(&returnedConflict, 0)
		return false, nil, nil
	})

	return fakeClient
}

func TestFilterOutNodesWithIgnoredTaints(t *testing.T) {
	isReady := func(t *testing.T, node *apiv1.Node) bool {
		for _, condition := range node.Status.Conditions {
			if condition.Type == apiv1.NodeReady {
				return condition.Status == apiv1.ConditionTrue
			}
		}
		t.Fatalf("failed to find condition/NodeReady")
		return false
	}

	readyCondition := apiv1.NodeCondition{
		Type:               apiv1.NodeReady,
		Status:             apiv1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	for name, tc := range map[string]struct {
		readyNodes    int
		allNodes      int
		ignoredTaints TaintKeySet
		node          *apiv1.Node
	}{
		"empty ignored taints, no node": {
			readyNodes:    0,
			allNodes:      0,
			ignoredTaints: map[string]bool{},
			node:          nil,
		},
		"one ignored taint, no node": {
			readyNodes: 0,
			allNodes:   0,
			ignoredTaints: map[string]bool{
				"my-taint": true,
			},
			node: nil,
		},
		"one ignored taint, one ready untainted node": {
			readyNodes: 1,
			allNodes:   1,
			ignoredTaints: map[string]bool{
				"my-taint": true,
			},
			node: &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "readyNoTaint",
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: apiv1.NodeSpec{
					Taints: []apiv1.Taint{},
				},
				Status: apiv1.NodeStatus{
					Conditions: []apiv1.NodeCondition{readyCondition},
				},
			},
		},
		"one ignored taint, one unready tainted node": {
			readyNodes: 0,
			allNodes:   1,
			ignoredTaints: map[string]bool{
				"my-taint": true,
			},
			node: &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "notReadyTainted",
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: apiv1.NodeSpec{
					Taints: []apiv1.Taint{
						{
							Key:    "my-taint",
							Value:  "myValue",
							Effect: apiv1.TaintEffectNoSchedule,
						},
					},
				},
				Status: apiv1.NodeStatus{
					Conditions: []apiv1.NodeCondition{readyCondition},
				},
			},
		},
		"no ignored taint, one unready prefixed tainted node": {
			readyNodes:    0,
			allNodes:      1,
			ignoredTaints: map[string]bool{},
			node: &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "notReadyTainted",
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: apiv1.NodeSpec{
					Taints: []apiv1.Taint{
						{
							Key:    IgnoreTaintPrefix + "another-taint",
							Value:  "myValue",
							Effect: apiv1.TaintEffectNoSchedule,
						},
					},
				},
				Status: apiv1.NodeStatus{
					Conditions: []apiv1.NodeCondition{readyCondition},
				},
			},
		},
		"no ignored taint, two taints": {
			readyNodes:    1,
			allNodes:      1,
			ignoredTaints: map[string]bool{},
			node: &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "ReadyTainted",
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: apiv1.NodeSpec{
					Taints: []apiv1.Taint{
						{
							Key:    "first-taint",
							Value:  "myValue",
							Effect: apiv1.TaintEffectNoSchedule,
						},
						{
							Key:    "second-taint",
							Value:  "myValue",
							Effect: apiv1.TaintEffectNoSchedule,
						},
					},
				},
				Status: apiv1.NodeStatus{
					Conditions: []apiv1.NodeCondition{readyCondition},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var nodes []*apiv1.Node
			if tc.node != nil {
				nodes = append(nodes, tc.node)
			}
			allNodes, readyNodes := FilterOutNodesWithIgnoredTaints(tc.ignoredTaints, nodes, nodes)
			assert.Equal(t, tc.allNodes, len(allNodes))
			assert.Equal(t, tc.readyNodes, len(readyNodes))

			allNodesSet := make(map[string]struct{}, len(allNodes))
			for _, node := range allNodes {
				_, ok := allNodesSet[node.Name]
				assert.False(t, ok)
				allNodesSet[node.Name] = struct{}{}
				nodeIsReady := isReady(t, node)
				assert.Equal(t, tc.allNodes == tc.readyNodes, nodeIsReady)
			}

			readyNodesSet := make(map[string]struct{}, len(allNodes))
			for _, node := range readyNodes {
				_, ok := readyNodesSet[node.Name]
				assert.False(t, ok)
				readyNodesSet[node.Name] = struct{}{}
				_, ok = allNodesSet[node.Name]
				assert.True(t, ok)
				nodeIsReady := isReady(t, node)
				assert.True(t, nodeIsReady)
			}
		})
	}
}

func TestSanitizeTaints(t *testing.T) {
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-sanitize-taints",
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Spec: apiv1.NodeSpec{
			Taints: []apiv1.Taint{
				{
					Key:    IgnoreTaintPrefix + "another-taint",
					Value:  "myValue",
					Effect: apiv1.TaintEffectNoSchedule,
				},
				{
					Key:    ReschedulerTaintKey,
					Value:  "test1",
					Effect: apiv1.TaintEffectNoSchedule,
				},
				{
					Key:    "test-taint",
					Value:  "test2",
					Effect: apiv1.TaintEffectNoSchedule,
				},
				{
					Key:    ToBeDeletedTaint,
					Value:  "1",
					Effect: apiv1.TaintEffectNoSchedule,
				},
				{
					Key:    "ignore-me",
					Value:  "1",
					Effect: apiv1.TaintEffectNoSchedule,
				},
				{
					Key:    "node.kubernetes.io/memory-pressure",
					Value:  "1",
					Effect: apiv1.TaintEffectNoSchedule,
				},
				{
					Key:    "ignore-taint.cluster-autoscaler.kubernetes.io/to-be-ignored",
					Value:  "I-am-the-invisible-man-Incredible-how-you-can",
					Effect: apiv1.TaintEffectNoSchedule,
				},
			},
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{},
		},
	}
	ignoredTaints := map[string]bool{"ignore-me": true}

	newTaints := SanitizeTaints(node.Spec.Taints, ignoredTaints)
	require.Equal(t, len(newTaints), 1)
	assert.Equal(t, newTaints[0].Key, "test-taint")
}
