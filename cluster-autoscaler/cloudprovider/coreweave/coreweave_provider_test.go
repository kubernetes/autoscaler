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
	"errors"
	"fmt"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// Fake manager for testing
type fakeManager struct {
	nodegroups      []cloudprovider.NodeGroup
	getNodeGroupErr bool
}

var (
	//define nodepoolObj
	nodepoolObj = map[string]interface{}{
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
	// Define test nodes
	nodes = []*apiv1.Node{
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
				Labels: map[string]string{coreWeaveNodePoolUID: "uid-2"},
			},
		},
	}
)

// Ensure fakeManager implements CoreWeaveManagerInterface
var _ CoreWeaveManagerInterface = &fakeManager{}

func (f *fakeManager) GetNodeGroup(uid string) (cloudprovider.NodeGroup, error) {
	if f.getNodeGroupErr {
		return nil, fmt.Errorf("error")
	}
	return f.nodegroups[0], nil
}

func (f *fakeManager) ListNodePools() ([]*CoreWeaveNodePool, error) {
	return nil, nil
}

func (f *fakeManager) UpdateNodeGroup() ([]cloudprovider.NodeGroup, error) {
	return f.nodegroups, nil
}

// Add missing Refresh method to satisfy CoreWeaveManagerInterface
func (f *fakeManager) Refresh() error {
	return nil
}

func TestCoreWeaveCloudProvider_Name(t *testing.T) {
	cp := &CoreWeaveCloudProvider{}
	if cp.Name() != cloudprovider.CoreWeaveProviderName {
		t.Errorf("expected provider name %q, got %q", cloudprovider.CoreWeaveProviderName, cp.Name())
	}
}

func TestCoreWeaveCloudProvider_NodeGroups_ManagerNil(t *testing.T) {
	cp := &CoreWeaveCloudProvider{manager: nil}
	if got := cp.NodeGroups(); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestCoreWeaveCloudProvider_NodeGroups_ListError(t *testing.T) {
	cp := &CoreWeaveCloudProvider{manager: &fakeManager{nodegroups: nil}}
	if got := cp.NodeGroups(); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestCoreWeaveCloudProvider_NodeGroups_Success(t *testing.T) {
	item := unstructured.Unstructured{Object: nodepoolObj}
	// Create a test manager with the nodepool item
	manager := makeTestManagerWithNodePools(nodes, item)
	cp := &CoreWeaveCloudProvider{manager: manager}
	groups := cp.NodeGroups()
	if len(groups) != 1 {
		t.Errorf("expected 1 node group, got %d", len(groups))
	}
}

func TestCoreWeaveCloudProvider_NodeGroupForNode_ManagerNil(t *testing.T) {
	cp := &CoreWeaveCloudProvider{manager: nil}
	node := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"}}
	ng, err := cp.NodeGroupForNode(node)
	if ng != nil || err == nil {
		t.Error("expected nil node group and error when manager is nil")
	}
}

func TestCoreWeaveCloudProvider_NodeGroupForNode_GetNodePoolByNameError(t *testing.T) {
	item := unstructured.Unstructured{Object: nodepoolObj}
	// Create a test manager with the nodepool item
	manager := makeTestManagerWithNodePools(nil, item)
	manager.UpdateNodeGroup()
	cp := &CoreWeaveCloudProvider{manager: manager}
	node := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1", Labels: map[string]string{coreWeaveNodePoolUID: "missing"}}}
	nodeGroup, err := cp.NodeGroupForNode(node)
	if err != nil {
		t.Errorf("unexpected error from NodeGroupForNode: %v", err)
	}
	if nodeGroup != nil {
		t.Error("expected nil nodeGroup for non-existent nodepool UID")
	}
}

func TestCoreWeaveCloudProvider_NodeGroupForNode_GetNodePoolByUIDError(t *testing.T) {
	item := unstructured.Unstructured{Object: nodepoolObj}
	// Create a test manager with the nodepool item
	manager := makeTestManagerWithNodePools(nil, item)
	manager.UpdateNodeGroup()
	cp := &CoreWeaveCloudProvider{manager: manager}
	node := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1", Labels: map[string]string{coreWeaveNodePoolUID: "wrong-uid"}}}
	nodeGroup, err := cp.NodeGroupForNode(node)
	if err != nil {
		t.Errorf("unexpected error from NodeGroupForNode: %v", err)
	}
	if nodeGroup != nil {
		t.Error("expected nil nodeGroup for non-existent nodepool UID")
	}
}

func TestCoreWeaveCloudProvider_NodeGroupForNode_Success(t *testing.T) {
	item := unstructured.Unstructured{Object: nodepoolObj}
	node := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1", Labels: map[string]string{coreWeaveNodePoolUID: "uid1"}}}
	// Create a test manager with the nodepool item
	manager := makeTestManagerWithNodePools([]*apiv1.Node{node}, item)
	manager.UpdateNodeGroup()
	cp := &CoreWeaveCloudProvider{manager: manager}

	ng, err := cp.NodeGroupForNode(node)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ng == nil {
		t.Error("expected node group, got nil")
	}
}

func TestCoreWeaveCloudProvider_NotImplementedMethods(t *testing.T) {
	cp := &CoreWeaveCloudProvider{}
	_, err2 := cp.Pricing()
	if !errors.Is(err2, cloudprovider.ErrNotImplemented) {
		t.Error("expected ErrNotImplemented for Pricing")
	}
	_, err3 := cp.GetAvailableMachineTypes()
	if !errors.Is(err3, cloudprovider.ErrNotImplemented) {
		t.Error("expected ErrNotImplemented for GetAvailableMachineTypes")
	}
	_, err4 := cp.NewNodeGroup("", nil, nil, nil, nil)
	if !errors.Is(err4, cloudprovider.ErrNotImplemented) {
		t.Error("expected ErrNotImplemented for NewNodeGroup")
	}
	if cp.GPULabel() != "" {
		t.Error("expected empty string for GPULabel")
	}
	if cp.GetAvailableGPUTypes() != nil {
		t.Error("expected nil for GetAvailableGPUTypes")
	}
	if cp.GetNodeGpuConfig(&apiv1.Node{}) != nil {
		t.Error("expected nil for GetNodeGpuConfig")
	}
	if cp.Cleanup() != nil {
		t.Error("expected nil for Cleanup")
	}
}

func TestCoreWeaveCloudProvider_HasInstance_ManagerNil(t *testing.T) {
	cp := &CoreWeaveCloudProvider{manager: nil}
	node := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"}}
	ok, err := cp.HasInstance(node)
	if ok {
		t.Error("expected false when manager is nil")
	}
	if err == nil || err.Error() != "CoreWeave manager is nil" {
		t.Errorf("expected CoreWeave manager is nil error, got %v", err)
	}
}

func TestCoreWeaveCloudProvider_HasInstance_NoLabels(t *testing.T) {
	cp := &CoreWeaveCloudProvider{manager: &fakeManager{}}
	node := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1", Labels: nil}}
	ok, err := cp.HasInstance(node)
	if ok || err != nil {
		t.Error("expected false, nil for node with nil labels")
	}

	node2 := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n2", Labels: map[string]string{}}}
	ok2, err2 := cp.HasInstance(node2)
	if ok2 || err2 != nil {
		t.Error("expected false, nil for node with empty labels")
	}

	node3 := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n3", Labels: map[string]string{"other": "value"}}}
	ok3, err3 := cp.HasInstance(node3)
	if ok3 || err3 != nil {
		t.Error("expected false, nil for node with unrelated labels")
	}
}

func TestCoreWeaveCloudProvider_HasInstance_NodeGroupError(t *testing.T) {
	cp := &CoreWeaveCloudProvider{manager: &fakeManager{getNodeGroupErr: true}}
	node := &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1", Labels: map[string]string{coreWeaveNodePoolUID: "uid1"}}}
	ok, err := cp.HasInstance(node)
	if ok || err != nil {
		t.Error("expected false, nil when GetNodeGroup returns error")
	}
}
