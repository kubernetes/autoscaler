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
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
					Key:    deletetaint.ToBeDeletedTaint,
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
