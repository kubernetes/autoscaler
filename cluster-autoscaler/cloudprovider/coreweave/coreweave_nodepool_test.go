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

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// Helper to create a minimal nodepool object
func makeTestNodePool(uid, name string, min, max, target int64) *CoreWeaveNodePool {
	dynamicClientset := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: coreWeaveGroup, Version: coreWeaveVersion, Resource: coreWeaveResource}: "kindList",
		},
	)
	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": "default",
			"uid":       uid,
		},
		"spec": map[string]interface{}{
			"minNodes":    min,
			"maxNodes":    max,
			"targetNodes": target,
		},
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
