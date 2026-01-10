/*
Copyright 2025 The Kubernetes Authors.

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

package coreweave

import (
	"testing"

	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// Helper to create a minimal nodepool object
func makeTestNodePool(uid, name string, min, max, target int64, option ...NodeGroupOption) *CoreWeaveNodePool {
	dynamicClientset := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: coreWeaveGroup, Version: coreWeaveVersion, Resource: coreWeaveResource}: "kindList",
		},
	)
	obj := map[string]any{
		"metadata": map[string]any{
			"name":      name,
			"namespace": "default",
			"uid":       uid,
		},
		"spec": map[string]any{
			"minNodes":    min,
			"maxNodes":    max,
			"targetNodes": target,
		},
	}

	for _, o := range option {
		o(obj)
	}

	u := &unstructured.Unstructured{Object: obj}
	// add fake kubernetes client
	fakeClient := k8sfake.NewSimpleClientset(
		&apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node1",
				Labels: map[string]string{coreWeaveNodePoolUID: "uid1"},
			},
		},
		&apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node2",
				Labels: map[string]string{coreWeaveNodePoolUID: "uid2"},
			},
		},
	)
	client := fakeClient
	np, _ := NewCoreWeaveNodePool(u, dynamicClientset, client)
	return np
}

func TestGetMinMaxTargetSize(t *testing.T) {
	np := makeTestNodePool("uid1", "np1", 2, 10, 5)
	if np.GetMinSize() != 2 {
		t.Errorf("expected min size 2, got %d", np.GetMinSize())
	}
	if np.GetMaxSize() != 10 {
		t.Errorf("expected max size 10, got %d", np.GetMaxSize())
	}
	if np.GetTargetSize() != 5 {
		t.Errorf("expected target size 5, got %d", np.GetTargetSize())
	}
}

func TestValidateNodes(t *testing.T) {
	nodeGood := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node-good",
			Labels: map[string]string{coreWeaveNodePoolUID: "uid1"},
		},
	}
	nodeBad := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node-bad",
			Labels: map[string]string{coreWeaveNodePoolUID: "other-uid"},
		},
	}
	nodeNoLabel := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node-nolabel",
			Labels: nil,
		},
	}
	np := makeTestNodePool("uid1", "np1", 1, 5, 3)
	// Test cases
	tests := []struct {
		name      string
		nodes     []*apiv1.Node
		expectErr bool
	}{
		{"empty", []*apiv1.Node{}, true},
		{"all good", []*apiv1.Node{nodeGood}, false},
		{"bad uid", []*apiv1.Node{nodeBad}, true},
		{"no label", []*apiv1.Node{nodeNoLabel}, true},
		{"mixed", []*apiv1.Node{nodeGood, nodeBad}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := np.ValidateNodes(tc.nodes)
			if tc.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSetSize_Invalid(t *testing.T) {
	np := makeTestNodePool("uid1", "np1", 1, 5, 3)

	// Negative size
	if err := np.SetSize(-1); err == nil {
		t.Error("expected error for negative size")
	}
	// Out of bounds
	if err := np.SetSize(10); err == nil {
		t.Error("expected error for size out of bounds")
	}
	// No change needed
	if err := np.SetSize(3); err != nil {
		t.Errorf("expected nil error for no change, got %v", err)
	}
}

func TestGetNameAndUID(t *testing.T) {
	np := makeTestNodePool("uid1", "np1", 1, 5, 3)
	if np.GetName() != "np1" {
		t.Errorf("expected name np1, got %s", np.GetName())
	}
	if np.GetUID() != "uid1" {
		t.Errorf("expected uid uid1, got %s", np.GetUID())
	}
}

func TestGetNodes(t *testing.T) {
	np := makeTestNodePool("uid1", "np1", 1, 5, 3)
	nodes, err := np.GetNodes()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != "node1" {
		t.Errorf("expected to get node1, got %+v", nodes)
	}
}

func TestGetInstanceType(t *testing.T) {
	testCases := map[string]struct {
		nodePool *CoreWeaveNodePool
		expected string
	}{
		"empty instance type": {
			makeTestNodePool("uid1", "np1", 1, 5, 3, withInstanceType("")),
			"",
		},
		"instance type present": {
			makeTestNodePool("uid1", "np1", 1, 5, 3, withInstanceType("turin-gp-l")),
			"turin-gp-l",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := tc.nodePool.GetInstanceType()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetNodeLabels(t *testing.T) {
	nodeLabels := map[string]string{
		"foo": "bar",
	}
	testCases := map[string]struct {
		nodePool *CoreWeaveNodePool
		expected map[string]string
	}{
		"empty node labels": {
			makeTestNodePool("uid1", "np1", 1, 5, 3, withNodeLabels(map[string]string{})),
			map[string]string{},
		},
		"node labels does not exist": {
			makeTestNodePool("uid1", "np1", 1, 5, 3),
			map[string]string{},
		},
		"node labels present": {
			makeTestNodePool("uid1", "np1", 1, 5, 3, withNodeLabels(nodeLabels)),
			nodeLabels,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := tc.nodePool.GetNodeLabels()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetNodeTaints(t *testing.T) {
	taints := []apiv1.Taint{
		{
			Key:    "dedicated",
			Value:  "gpu",
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "app",
			Value:  "foo",
			Effect: apiv1.TaintEffectNoExecute,
		},
	}

	testCases := map[string]struct {
		nodePool *CoreWeaveNodePool
		expected []apiv1.Taint
	}{
		"empty node taints": {
			makeTestNodePool("uid1", "np1", 1, 5, 3, withNodeTaints([]apiv1.Taint{})),
			[]apiv1.Taint{},
		},
		"node taints does not exist": {
			makeTestNodePool("uid1", "np1", 1, 5, 3),
			[]apiv1.Taint{},
		},
		"node taints present": {
			makeTestNodePool("uid1", "np1", 1, 5, 3, withNodeTaints(taints)),
			taints,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := tc.nodePool.GetNodeTaints()
			if len(tc.expected) == 0 {
				require.Empty(t, actual)
			} else {
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}
