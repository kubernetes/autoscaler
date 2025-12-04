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

	"github.com/stretchr/testify/require"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func makeTestNodeGroup(name string, uid string, min, max, target int64) *CoreWeaveNodeGroup {
	// Create a dynamic client with a fake nodepool object
	// This is a minimal setup to test the CoreWeaveNodeGroup functionality
	dynamicClientset := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: coreWeaveGroup, Version: coreWeaveVersion, Resource: coreWeaveResource}: "kindList",
		},
	)
	obj := map[string]interface{}{
		"apiVersion": coreWeaveGroup + "/" + coreWeaveVersion,
		"kind":       coreWeaveResource,
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
	_, _ = dynamicClientset.Resource(CoreWeaveNodeGroupResource).Namespace("default").Create(context.TODO(), u.DeepCopy(), metav1.CreateOptions{})
	// fake a Kubernetes client to simulate node operations
	fakeClient := k8sfake.NewSimpleClientset(
		&apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node1",
				Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
			},
		},
		&apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node2",
				Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
			},
		},
	)
	client := fakeClient
	// Create the CoreWeaveNodePool instance
	np, _ := NewCoreWeaveNodePool(u, dynamicClientset, client)
	return NewCoreWeaveNodeGroup(np)
}

func TestId(t *testing.T) {
	ng := makeTestNodeGroup("ng-1", "uid-1", 1, 5, 3)
	if ng.Id() != "uid-1" {
		t.Errorf("expected id 'uid-1', got %s", ng.Id())
	}
}

func TestMinMaxTargetSize(t *testing.T) {
	ng := makeTestNodeGroup("ng-1", "uid-1", 2, 10, 5)
	if ng.MinSize() != 2 {
		t.Errorf("expected min size 2, got %d", ng.MinSize())
	}
	if ng.MaxSize() != 10 {
		t.Errorf("expected max size 10, got %d", ng.MaxSize())
	}
	size, err := ng.TargetSize()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if size != 5 {
		t.Errorf("expected target size 5, got %d", size)
	}
}

func TestIncreaseSize(t *testing.T) {
	ng := makeTestNodeGroup("ng-1", "uid-1", 1, 5, 3)
	err := ng.IncreaseSize(2)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		t.Errorf("expected ErrNotImplemented or nil, got %v", err)
	}
}

func TestDeleteNodes(t *testing.T) {
	initialTargetSize := int64(3)

	testCases := map[string]struct {
		nodesToDelete      []*apiv1.Node
		expectedTargetSize int
		expectedError      error
	}{
		"reduce-target-size-by-one-node": {
			nodesToDelete: []*apiv1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node1",
						Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
					},
				},
			},
			expectedTargetSize: 2,
		},
		"reduce-target-size-by-three-node": {
			nodesToDelete: []*apiv1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node1",
						Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node2",
						Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node3",
						Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
					},
				},
			},
			expectedTargetSize: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ng := makeTestNodeGroup("ng-1", "uid-1", 0, 5, initialTargetSize)

			err := ng.DeleteNodes(tc.nodesToDelete)
			if tc.expectedError != nil {
				require.Equal(t, tc.expectedError, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, ng.nodepool.GetTargetSize(), tc.expectedTargetSize)
		})
	}
}

func TestDecreaseTargetSize(t *testing.T) {
	testCases := map[string]struct {
		delta              int
		expectedTargetSize int
		expectedError      error
	}{
		"positive-delta": {
			delta:              2,
			expectedTargetSize: 1,
		},
		"negative-delta": {
			delta:              -2,
			expectedTargetSize: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ng := makeTestNodeGroup("ng-1", "uid-1", 1, 5, 3)

			err := ng.DecreaseTargetSize(tc.delta)
			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedTargetSize, ng.nodepool.GetTargetSize())
		})
	}
}
