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
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestGetCoreWeaveClient(t *testing.T) {
	// Nil config should return error
	client, err := GetCoreWeaveClient(nil)
	if err == nil || client != nil {
		t.Error("expected error and nil client for nil config")
	}
}

func TestGetCoreWeaveDynamicClient(t *testing.T) {
	// Nil config should return error
	client, err := GetCoreWeaveDynamicClient(nil)
	if err == nil || client != nil {
		t.Error("expected error and nil client for nil config")
	}
}

func TestGetCoreWeaveContext(t *testing.T) {
	ctx, cancel := GetCoreWeaveContext()
	defer cancel()
	if ctx == nil {
		t.Error("expected non-nil context")
	}
}

func TestGetCoreWeaveNodesByNodeGroupUID(t *testing.T) {
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

	// Nil client
	_, err := GetCoreWeaveNodesByNodeGroupUID(nil, "uid1")
	if err == nil {
		t.Error("expected error for nil client")
	}

	// Empty UID
	_, err = GetCoreWeaveNodesByNodeGroupUID(client, "")
	if err == nil {
		t.Error("expected error for empty node group UID")
	}

	// Valid UID
	nodes, err := GetCoreWeaveNodesByNodeGroupUID(client, "uid1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != "node1" {
		t.Errorf("expected 1 node named 'node1', got: %+v", nodes)
	}

	// UID not found
	nodes, err = GetCoreWeaveNodesByNodeGroupUID(client, "missing")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes, got: %+v", nodes)
	}
}
