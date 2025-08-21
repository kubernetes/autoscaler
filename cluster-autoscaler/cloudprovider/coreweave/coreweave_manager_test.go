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
	"context"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// Helper to create a fake manager with a dynamic client and optional nodepool items
func makeTestManagerWithNodePools(nodes []*apiv1.Node, items ...unstructured.Unstructured) *CoreWeaveManager {

	scheme := runtime.NewScheme()
	// Register the CoreWeave NodePool resource
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{
			Group:   coreWeaveGroup,
			Version: coreWeaveVersion,
			Kind:    coreWeaveResource,
		},
		&unstructured.Unstructured{},
	)
	// Create a fake dynamic client with the custom list kind
	dynClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			{Group: coreWeaveGroup, Version: coreWeaveVersion, Resource: coreWeaveResource}: "List",
		},
	)
	for _, item := range items {
		// Create each item individually
		_, _ = dynClient.Resource(CoreWeaveNodeGroupResource).Namespace("default").Create(context.TODO(), item.DeepCopy(), metav1.CreateOptions{})
	}
	// fake a Kubernetes client to simulate node operations
	fakeClient := k8sfake.NewSimpleClientset()
	// Add nodes to the fake client
	for _, node := range nodes {
		_, _ = fakeClient.CoreV1().Nodes().Create(context.TODO(), node.DeepCopy(), metav1.CreateOptions{})
	}
	client := fakeClient
	return &CoreWeaveManager{dynamicClient: dynClient, clientset: client}
}

func TestNewCoreWeaveManager(t *testing.T) {
	// Create a fake dynamic client
	scheme := runtime.NewScheme()
	dynClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			{Group: coreWeaveGroup, Version: coreWeaveVersion, Resource: coreWeaveResource}: "List",
		},
	)
	// Create a fake Kubernetes client
	fakeClient := k8sfake.NewSimpleClientset()

	manager, err := NewCoreWeaveManager(dynClient, fakeClient)
	if err != nil {
		t.Fatalf("unexpected error creating CoreWeaveManager: %v", err)
	}
	if manager == nil {
		t.Fatal("expected CoreWeaveManager to be initialized")
	}

	if manager.dynamicClient == nil {
		t.Error("expected dynamic client to be initialized")
	}
	if manager.clientset == nil {
		t.Error("expected Kubernetes client to be initialized")
	}
}
func TestListNodePools_Success(t *testing.T) {
	// Create a valid nodepool object
	obj := map[string]interface{}{
		"apiVersion": coreWeaveGroup + "/" + coreWeaveVersion,
		"kind":       coreWeaveResource,
		"metadata": map[string]interface{}{
			"name":      "np1",
			"namespace": "default",
			"uid":       "uid1",
		},
		"spec": map[string]interface{}{
			"minNodes":    int64(1),
			"maxNodes":    int64(5),
			"targetNodes": int64(3),
			"autoscaling": true,
		},
	}
	item := unstructured.Unstructured{Object: obj}
	// Create a test manager with the nodepool item
	manager := makeTestManagerWithNodePools(nil, item)
	nodePools, err := manager.ListNodePools()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodePools) != 1 {
		t.Fatalf("expected 1 nodepool, got %d", len(nodePools))
	}
	if nodePools[0].GetName() != "np1" {
		t.Errorf("expected nodepool name 'np1', got %s", nodePools[0].GetName())
	}
}

func TestListNodePools_ListError(t *testing.T) {
	// Patch the dynamic client to return an error on List
	scheme := runtime.NewScheme()
	// create list with a single item
	dynClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			{Group: coreWeaveGroup, Version: coreWeaveVersion, Resource: coreWeaveResource}: "kindList",
		},
	)
	manager := &CoreWeaveManager{
		dynamicClient: dynClient,
	}
	_, err := manager.ListNodePools()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestListNodePools_InvalidNodePool(t *testing.T) {
	// Item missing required metadata fields
	obj := map[string]interface{}{
		"apiVersion": coreWeaveGroup + "/" + coreWeaveVersion,
		"kind":       coreWeaveResource,
	}
	item := unstructured.Unstructured{Object: obj}
	manager := makeTestManagerWithNodePools(nil, item)
	_, err := manager.ListNodePools()
	if err == nil {
		t.Error("expected error for invalid nodepool, got nil")
	}
}

func TestListNodePools_AutoscalingDisabled(t *testing.T) {
	// Create a nodepool with autoscaling disabled
	obj := map[string]interface{}{
		"apiVersion": coreWeaveGroup + "/" + coreWeaveVersion,
		"kind":       coreWeaveResource,
		"metadata": map[string]interface{}{
			"name":      "np1",
			"namespace": "default",
			"uid":       "uid1",
		},
		"spec": map[string]interface{}{
			"minNodes":    int64(1),
			"maxNodes":    int64(5),
			"targetNodes": int64(3),
			"autoscaling": false, // Autoscaling disabled
		},
	}
	item := unstructured.Unstructured{Object: obj}
	manager := makeTestManagerWithNodePools(nil, item)
	nodePools, err := manager.ListNodePools()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodePools) != 0 {
		t.Errorf("expected no nodepools with autoscaling disabled, got %d", len(nodePools))
	}
}
func TestListNodePools_EmptyList(t *testing.T) {
	// Create a manager with no nodepool items
	manager := makeTestManagerWithNodePools(nil)
	nodePools, err := manager.ListNodePools()
	if err == nil {
		t.Errorf("expected error for empty nodepools list, got nil")
	}
	if len(nodePools) != 0 {
		t.Errorf("expected empty nodepools list, got %d", len(nodePools))
	}
}
