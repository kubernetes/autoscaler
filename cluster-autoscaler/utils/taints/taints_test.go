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
	"slices"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/config"
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
	_, err := MarkToBeDeleted(node, fakeClient, false)
	assert.NoError(t, err)

	updatedNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(updatedNode))
	assert.False(t, HasDeletionCandidateTaint(updatedNode))
}

func TestSoftMarkNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	fakeClient := buildFakeClientWithConflicts(t, node)
	_, err := MarkDeletionCandidate(node, fakeClient)
	assert.NoError(t, err)

	updatedNode := getNode(t, fakeClient, "node")
	assert.False(t, HasToBeDeletedTaint(updatedNode))
	assert.True(t, HasDeletionCandidateTaint(updatedNode))
}

func TestCheckNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	taints := []apiv1.Taint{
		{
			Key:    ToBeDeletedTaint,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "other-taint",
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
	}
	addTaintsToSpec(node, taints, false)
	fakeClient := buildFakeClientWithConflicts(t, node)

	updatedNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(updatedNode))
	assert.True(t, HasTaint(node, "other-taint"))
	assert.False(t, HasDeletionCandidateTaint(updatedNode))
}

func TestSoftCheckNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	taints := []apiv1.Taint{
		{
			Key:    DeletionCandidateTaintKey,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
		{
			Key:    "other-taint",
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
	}
	addTaintsToSpec(node, taints, false)
	fakeClient := buildFakeClientWithConflicts(t, node)

	updatedNode := getNode(t, fakeClient, "node")
	assert.False(t, HasToBeDeletedTaint(updatedNode))
	assert.True(t, HasDeletionCandidateTaint(updatedNode))
}

func TestQueryNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	fakeClient := buildFakeClientWithConflicts(t, node)
	_, err := MarkToBeDeleted(node, fakeClient, false)
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
	_, err := MarkDeletionCandidate(node, fakeClient)
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
	taints := []apiv1.Taint{
		{
			Key:    ToBeDeletedTaint,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "other-taint",
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
	}
	addTaintsToSpec(node, taints, false)
	fakeClient := buildFakeClientWithConflicts(t, node)

	apiNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(apiNode))
	assert.True(t, HasTaint(apiNode, "other-taint"))
	assert.False(t, apiNode.Spec.Unschedulable)

	updatedNode, err := CleanToBeDeleted(node, fakeClient, false)
	cleaned := !slices.Equal(updatedNode.Spec.Taints, node.Spec.Taints)
	assert.True(t, cleaned)
	assert.NoError(t, err)

	apiNode = getNode(t, fakeClient, "node")
	assert.NoError(t, err)
	assert.False(t, HasToBeDeletedTaint(apiNode))
	assert.True(t, HasTaint(apiNode, "other-taint"))
	assert.False(t, apiNode.Spec.Unschedulable)
}

func TestCleanNodesWithCordon(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	taints := []apiv1.Taint{
		{
			Key:    ToBeDeletedTaint,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "other-taint",
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
	}
	addTaintsToSpec(node, taints, true)
	fakeClient := buildFakeClientWithConflicts(t, node)

	apiNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(apiNode))
	assert.True(t, HasTaint(apiNode, "other-taint"))
	assert.True(t, apiNode.Spec.Unschedulable)

	updatedNode, err := CleanToBeDeleted(node, fakeClient, true)
	cleaned := !slices.Equal(updatedNode.Spec.Taints, node.Spec.Taints)
	assert.True(t, cleaned)
	assert.NoError(t, err)

	apiNode = getNode(t, fakeClient, "node")
	assert.NoError(t, err)
	assert.False(t, HasToBeDeletedTaint(apiNode))
	assert.True(t, HasTaint(apiNode, "other-taint"))
	assert.False(t, apiNode.Spec.Unschedulable)
}

func TestCleanNodesWithCordonOnOff(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	taints := []apiv1.Taint{
		{
			Key:    ToBeDeletedTaint,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
		{
			Key:    "other-taint",
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
	}
	addTaintsToSpec(node, taints, true)
	fakeClient := buildFakeClientWithConflicts(t, node)

	apiNode := getNode(t, fakeClient, "node")
	assert.True(t, HasToBeDeletedTaint(apiNode))
	assert.True(t, HasTaint(apiNode, "other-taint"))
	assert.True(t, apiNode.Spec.Unschedulable)

	updatedNode, err := CleanToBeDeleted(node, fakeClient, false)
	cleaned := !slices.Equal(updatedNode.Spec.Taints, node.Spec.Taints)
	assert.True(t, cleaned)
	assert.NoError(t, err)

	apiNode = getNode(t, fakeClient, "node")
	assert.NoError(t, err)
	assert.False(t, HasToBeDeletedTaint(apiNode))
	assert.True(t, HasTaint(apiNode, "other-taint"))
	assert.True(t, apiNode.Spec.Unschedulable)
}

func TestSoftCleanNodes(t *testing.T) {
	defer setConflictRetryInterval(setConflictRetryInterval(time.Millisecond))
	node := BuildTestNode("node", 1000, 1000)
	taints := []apiv1.Taint{
		{
			Key:    DeletionCandidateTaintKey,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
		{
			Key:    "other-taint",
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
	}
	addTaintsToSpec(node, taints, false)
	fakeClient := buildFakeClientWithConflicts(t, node)

	apiNode := getNode(t, fakeClient, "node")
	assert.True(t, HasDeletionCandidateTaint(apiNode))
	assert.True(t, HasTaint(apiNode, "other-taint"))

	updatedNode, err := CleanDeletionCandidate(node, fakeClient)
	cleaned := !slices.Equal(updatedNode.Spec.Taints, node.Spec.Taints)
	assert.True(t, cleaned)
	assert.NoError(t, err)

	apiNode = getNode(t, fakeClient, "node")
	assert.NoError(t, err)
	assert.False(t, HasDeletionCandidateTaint(apiNode))
	assert.True(t, HasTaint(apiNode, "other-taint"))
}

func TestCleanAllToBeDeleted(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 10)
	n2 := BuildTestNode("n2", 1000, 10)
	n2.Spec.Taints = []apiv1.Taint{{Key: ToBeDeletedTaint, Value: strconv.FormatInt(time.Now().Unix()-301, 10)}}

	fakeClient := buildFakeClient(t, n1, n2)
	fakeRecorder := kube_util.CreateEventRecorder(context.TODO(), fakeClient, false)

	assert.Equal(t, 1, len(getNode(t, fakeClient, "n2").Spec.Taints))

	CleanAllToBeDeleted([]*apiv1.Node{n1, n2}, fakeClient, fakeRecorder, false)

	assert.Equal(t, 0, len(getNode(t, fakeClient, "n1").Spec.Taints))
	assert.Equal(t, 0, len(getNode(t, fakeClient, "n2").Spec.Taints))
}

func TestCleanAllDeletionCandidates(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 10)
	n2 := BuildTestNode("n2", 1000, 10)
	n2.Spec.Taints = []apiv1.Taint{{Key: DeletionCandidateTaintKey, Value: strconv.FormatInt(time.Now().Unix()-301, 10)}}

	fakeClient := buildFakeClient(t, n1, n2)
	fakeRecorder := kube_util.CreateEventRecorder(context.TODO(), fakeClient, false)

	assert.Equal(t, 1, len(getNode(t, fakeClient, "n2").Spec.Taints))

	CleanStaleDeletionCandidates([]*apiv1.Node{n1, n2}, fakeClient, fakeRecorder, time.Duration(0))

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
	fakeClient := fake.NewClientset()

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

func TestFilterOutNodesWithStartupTaints(t *testing.T) {
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
		readyNodes            int
		allNodes              int
		startupTaints         TaintKeySet
		startupTaintsPrefixes []string
		node                  *apiv1.Node
	}{
		"empty startup taints, no node": {
			readyNodes:    0,
			allNodes:      0,
			startupTaints: map[string]bool{},
			node:          nil,
		},
		"one startup taint, no node": {
			readyNodes: 0,
			allNodes:   0,
			startupTaints: map[string]bool{
				"my-taint": true,
			},
			node: nil,
		},
		"one startup taint, one ready untainted node": {
			readyNodes: 1,
			allNodes:   1,
			startupTaints: map[string]bool{
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
		"one startup taint, one unready tainted node": {
			readyNodes: 0,
			allNodes:   1,
			startupTaints: map[string]bool{
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
		"no startup taint, one node unready prefixed with startup taint prefix (Compatibility)": {
			readyNodes:            0,
			allNodes:              1,
			startupTaints:         map[string]bool{},
			startupTaintsPrefixes: []string{IgnoreTaintPrefix},
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
		"no startup taint, one node unready prefixed with startup taint prefix": {
			readyNodes:            0,
			allNodes:              1,
			startupTaints:         map[string]bool{},
			startupTaintsPrefixes: []string{StartupTaintPrefix},
			node: &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "notReadyTainted",
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: apiv1.NodeSpec{
					Taints: []apiv1.Taint{
						{
							Key:    StartupTaintPrefix + "another-taint",
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
		"no startup taint, two taints": {
			readyNodes:    1,
			allNodes:      1,
			startupTaints: map[string]bool{},
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
			taintConfig := TaintConfig{
				startupTaints:        tc.startupTaints,
				startupTaintPrefixes: tc.startupTaintsPrefixes,
			}
			allNodes, readyNodes := FilterOutNodesWithStartupTaints(taintConfig, nodes, nodes)
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
					Key:    StatusTaintPrefix + "some-taint",
					Value:  "myValue",
					Effect: apiv1.TaintEffectNoSchedule,
				},
				{
					Key:    StartupTaintPrefix + "some-taint",
					Value:  "myValue",
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
					Key:    "status-me",
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
	taintConfig := TaintConfig{
		startupTaints:        map[string]bool{"ignore-me": true},
		statusTaints:         map[string]bool{"status-me": true},
		startupTaintPrefixes: []string{IgnoreTaintPrefix, StartupTaintPrefix},
	}

	newTaints := SanitizeTaints(node.Spec.Taints, taintConfig)
	require.Equal(t, 2, len(newTaints))
	assert.Equal(t, newTaints[0].Key, StatusTaintPrefix+"some-taint")
	assert.Equal(t, newTaints[1].Key, "test-taint")
}

func TestCountNodeTaints(t *testing.T) {
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-count-node-taints",
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
					Key:    StatusTaintPrefix + "some-taint",
					Value:  "myValue",
					Effect: apiv1.TaintEffectNoSchedule,
				},
				{
					Key:    StartupTaintPrefix + "some-taint",
					Value:  "myValue",
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
					Key:    "status-me",
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
					Value:  "myValue2",
					Effect: apiv1.TaintEffectNoSchedule,
				},
			},
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{},
		},
	}
	node2 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-count-node-taints",
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Spec: apiv1.NodeSpec{
			Taints: []apiv1.Taint{
				{
					Key:    StatusTaintPrefix + "some-taint",
					Value:  "myValue",
					Effect: apiv1.TaintEffectNoSchedule,
				},
				{
					Key:    "node.kubernetes.io/unschedulable",
					Value:  "1",
					Effect: apiv1.TaintEffectNoSchedule,
				},
			},
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{},
		},
	}
	taintConfig := NewTaintConfig(config.AutoscalingOptions{
		StatusTaints:  []string{"status-me"},
		StartupTaints: []string{"ignore-me"},
	})
	want := map[string]int{
		"ignore-taint.cluster-autoscaler.kubernetes.io/": 2,
		"ToBeDeletedByClusterAutoscaler":                 1,
		"node.kubernetes.io/memory-pressure":             1,
		"node.kubernetes.io/unschedulable":               1,
		"other":                                          1,
		"startup-taint":                                  2,
		"status-taint":                                   3,
	}
	got := CountNodeTaints([]*apiv1.Node{node, node2}, taintConfig)
	assert.Equal(t, want, got)
}

func TestAddTaints(t *testing.T) {
	testCases := []struct {
		name           string
		existingTaints []string
		newTaints      []string
		wantTaints     []string
	}{
		{
			name:       "no existing taints",
			newTaints:  []string{"t1", "t2"},
			wantTaints: []string{"t1", "t2"},
		},
		{
			name:           "existing taints - no overlap",
			existingTaints: []string{"t1"},
			newTaints:      []string{"t2", "t3"},
			wantTaints:     []string{"t1", "t2", "t3"},
		},
		{
			name:           "existing taints - duplicates",
			existingTaints: []string{"t1"},
			newTaints:      []string{"t1", "t2"},
			wantTaints:     []string{"t1", "t2"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n := BuildTestNode("node", 1000, 1000)
			existingTaints := make([]apiv1.Taint, len(tc.existingTaints))
			for i, t := range tc.existingTaints {
				existingTaints[i] = apiv1.Taint{
					Key:    t,
					Effect: apiv1.TaintEffectNoSchedule,
				}
			}
			n.Spec.Taints = append([]apiv1.Taint{}, existingTaints...)
			fakeClient := buildFakeClient(t, n)
			newTaints := make([]apiv1.Taint, len(tc.newTaints))
			for i, t := range tc.newTaints {
				newTaints[i] = apiv1.Taint{
					Key:    t,
					Effect: apiv1.TaintEffectNoSchedule,
				}
			}
			updatedNode, err := AddTaints(n, fakeClient, newTaints, false)
			assert.NoError(t, err)
			apiNode := getNode(t, fakeClient, "node")
			for _, want := range tc.wantTaints {
				assert.True(t, HasTaint(updatedNode, want))
				assert.True(t, HasTaint(apiNode, want))
			}
		})
	}
}

func TestCleanTaints(t *testing.T) {
	testCases := []struct {
		name           string
		existingTaints []string
		taintsToRemove []string
		wantTaints     []string
		wantModified   bool
	}{
		{
			name:           "no existing taints",
			taintsToRemove: []string{"t1", "t2"},
			wantTaints:     []string{},
			wantModified:   false,
		},
		{
			name:           "existing taints - no overlap",
			existingTaints: []string{"t1"},
			taintsToRemove: []string{"t2", "t3"},
			wantTaints:     []string{"t1"},
			wantModified:   false,
		},
		{
			name:           "existing taints - remove one",
			existingTaints: []string{"t1", "t2"},
			taintsToRemove: []string{"t1"},
			wantTaints:     []string{"t2"},
			wantModified:   true,
		},
		{
			name:           "existing taints - remove all",
			existingTaints: []string{"t1", "t2"},
			taintsToRemove: []string{"t1", "t2"},
			wantTaints:     []string{},
			wantModified:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n := BuildTestNode("node", 1000, 1000)
			existingTaints := make([]apiv1.Taint, len(tc.existingTaints))
			for i, taintKey := range tc.existingTaints {
				existingTaints[i] = apiv1.Taint{
					Key:    taintKey,
					Effect: apiv1.TaintEffectNoSchedule,
				}
			}
			n.Spec.Taints = append([]apiv1.Taint{}, existingTaints...)
			fakeClient := buildFakeClient(t, n)

			updatedNode, err := CleanTaints(n, fakeClient, tc.taintsToRemove, false)
			modified := !slices.Equal(updatedNode.Spec.Taints, n.Spec.Taints)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantModified, modified)

			apiNode := getNode(t, fakeClient, "node")

			for _, want := range tc.wantTaints {
				assert.True(t, HasTaint(apiNode, want))
				assert.True(t, HasTaint(n, want))
			}

			for _, removed := range tc.taintsToRemove {
				assert.False(t, HasTaint(apiNode, removed))
				assert.False(t, HasTaint(updatedNode, removed), "Taint %s should have been removed from local node object", removed)
			}
		})
	}
}

func TestCleanStaleDeletionCandidates(t *testing.T) {

	currentTime := time.Now()
	deletionCandidateTaint := DeletionCandidateTaint()

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, currentTime)
	nt1 := deletionCandidateTaint
	ntt1 := currentTime.Add(-time.Minute * 2)
	nt1.Value = fmt.Sprint(ntt1.Unix())
	n1.Spec.Taints = append(n1.Spec.Taints, nt1)

	// Node whose DeletionCandidateTaint has lapsed, shouldn't be deleted
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, currentTime)
	nt2 := deletionCandidateTaint
	ntt2 := currentTime.Add(-time.Minute * 10)
	nt2.Value = fmt.Sprint(ntt2.Unix())
	n2.Spec.Taints = append(n2.Spec.Taints, nt2)

	// Node that is marked for deletion, but should have that mark removed
	n3 := BuildTestNode("n3", 1000, 1000)
	SetNodeReadyState(n3, true, currentTime)
	nt3 := deletionCandidateTaint
	ntt3 := currentTime.Add(-time.Minute * 2)
	nt3.Value = fmt.Sprint(ntt3.Unix())
	n3.Spec.Taints = append(n3.Spec.Taints, nt3)

	// Node with invalid DeletionCandidateTaint, taint should be deleted
	n4 := BuildTestNode("n4", 1000, 1000)
	SetNodeReadyState(n4, true, currentTime)
	nt4 := deletionCandidateTaint
	nt4.Value = "invalid-value"
	n4.Spec.Taints = append(n4.Spec.Taints, nt4)

	// Node with no DeletionCandidateTaint, should not be deleted
	n5 := BuildTestNode("n5", 1000, 1000)
	SetNodeReadyState(n5, true, currentTime)

	testCases := []struct {
		name                     string
		allNodes                 []*apiv1.Node
		unneededNodes            []*apiv1.Node
		nodeDeletionCandidateTTL time.Duration
	}{
		{
			name:                     "All deletion candidate nodes with standard TTL",
			allNodes:                 []*apiv1.Node{n1, n2, n3},
			unneededNodes:            []*apiv1.Node{n1, n3},
			nodeDeletionCandidateTTL: time.Minute * 5,
		},
		{
			name:                     "Node without deletion candidate taint should not be deleted",
			allNodes:                 []*apiv1.Node{n5},
			unneededNodes:            []*apiv1.Node{},
			nodeDeletionCandidateTTL: time.Minute * 5,
		},
		{
			name:                     "Node with invalid deletion candidate taint should be deleted",
			allNodes:                 []*apiv1.Node{n4},
			unneededNodes:            []*apiv1.Node{},
			nodeDeletionCandidateTTL: time.Minute * 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := buildFakeClient(t, tc.allNodes...)
			CleanStaleDeletionCandidates(
				tc.allNodes,
				fakeClient,
				kube_util.CreateEventRecorder(context.TODO(), fakeClient, false),
				tc.nodeDeletionCandidateTTL,
			)

			allNodes, err := fakeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			assert.NoError(t, err)
			assert.NotNil(t, allNodes)

			for _, node := range allNodes.Items {
				hasTaint := HasDeletionCandidateTaint(&node)
				isUnneeded := false
				for _, unneededNode := range tc.unneededNodes {
					if unneededNode.Name == node.Name {
						isUnneeded = true
						break
					}
				}

				if isUnneeded {
					assert.True(t, hasTaint, "Node %s should still have deletion candidate taint", node.Name)
				} else {
					assert.False(t, hasTaint, "Node %s should have had deletion candidate taint removed", node.Name)
				}
			}

		})
	}
}
