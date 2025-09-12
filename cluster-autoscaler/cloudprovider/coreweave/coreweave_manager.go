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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// CoreWeaveManager is the main manager for CoreWeave cloud provider operations
type CoreWeaveManager struct {
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
	nodeGroups    []cloudprovider.NodeGroup
}

// CoreWeaveManagerInterface defines the interface for CoreWeave management operations
// This interface allows for mocking in tests and provides a clear contract for CoreWeave management operations
type CoreWeaveManagerInterface interface {
	ListNodePools() ([]*CoreWeaveNodePool, error)
	UpdateNodeGroup() ([]cloudprovider.NodeGroup, error)
	GetNodeGroup(uid string) (cloudprovider.NodeGroup, error)
	Refresh() error
}

// NewCoreWeaveManager creates a new CoreWeaveManager instance
func NewCoreWeaveManager(dynamicClient dynamic.Interface, clientset kubernetes.Interface) (*CoreWeaveManager, error) {
	// check if the dynamic client is nil
	if dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client cannot be nil")
	}
	// check if the clientset is nil
	if clientset == nil {
		return nil, fmt.Errorf("kubernetes client cannot be nil")
	}
	// Create the CoreWeaveManager instance
	return &CoreWeaveManager{
		dynamicClient: dynamicClient,
		clientset:     clientset,
	}, nil
}

// ListNodePools lists all CoreWeave NodePools
func (m *CoreWeaveManager) ListNodePools() ([]*CoreWeaveNodePool, error) {
	ctx, cancel := GetCoreWeaveContext()
	defer cancel()

	resource := m.dynamicClient.Resource(CoreWeaveNodeGroupResource).Namespace(metav1.NamespaceAll)
	list, err := resource.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodepools: %v", err)
	}
	if list == nil || len(list.Items) == 0 {
		klog.V(4).Info("No nodepools found")
		return []*CoreWeaveNodePool{}, nil
	}
	klog.V(4).Infof("Found %d node pools", len(list.Items))
	// Convert unstructured items to CoreWeaveNodePool instances
	var nodePools []*CoreWeaveNodePool
	for _, item := range list.Items {
		nodePool, err := NewCoreWeaveNodePool(&item, m.dynamicClient, m.clientset)
		if err != nil {
			return nil, fmt.Errorf("failed to create node pool from item: %v", err)
		}
		// Ensure nodepool autoscaling is enabled
		if !nodePool.GetAutoscalingEnabled() {
			klog.V(4).Infof("Node pool %s has autoscaling disabled, skipping", nodePool.GetName())
			continue
		}
		// Add the node pool to the list
		klog.V(4).Infof("Found node pool: %s with UID: %s", nodePool.GetName(), nodePool.GetUID())
		// Check if the node
		nodePools = append(nodePools, nodePool)
	}

	return nodePools, nil
}

// Refresh refreshes the node groups in the CoreWeave manager
// This method clears the cached node groups and calls UpdateNodeGroup to repopulate them
// It is useful when the node groups may have changed and need to be reloaded
// It returns an error if the update fails
func (m *CoreWeaveManager) Refresh() error {
	m.nodeGroups = nil
	klog.V(4).Info("Refreshing node groups in CoreWeave manager")
	// Call UpdateNodeGroup to refresh the node groups
	_, err := m.UpdateNodeGroup()
	return err
}

// UpdateNodeGroup updates the specified NodeGroup
func (m *CoreWeaveManager) UpdateNodeGroup() ([]cloudprovider.NodeGroup, error) {
	// check if the nodeGroups are already populated
	if m.nodeGroups != nil {
		klog.V(4).Info("Node groups are already populated, skipping update")
		return m.nodeGroups, nil
	}
	// List all node pools
	nodepools, err := m.ListNodePools()
	if err != nil {
		klog.Errorf("Error listing node pools: %v", err)
		return nil, err
	}
	// If no node pools are found, return an empty slice
	if len(nodepools) == 0 {
		klog.Info("No node pools found, returning empty node groups")
		m.nodeGroups = []cloudprovider.NodeGroup{}
		return m.nodeGroups, nil
	}
	klog.V(4).Infof("Found %d node pools", len(nodepools))
	m.nodeGroups = make([]cloudprovider.NodeGroup, len(nodepools))
	for i, np := range nodepools {
		m.nodeGroups[i] = NewCoreWeaveNodeGroup(np)
	}
	return m.nodeGroups, nil
}

// GetNodeGroup by nodePoolUID retrieves a NodeGroup by its nodepool UID.
func (m *CoreWeaveManager) GetNodeGroup(nodePoolUID string) (cloudprovider.NodeGroup, error) {
	if m.nodeGroups == nil {
		return nil, fmt.Errorf("node groups are not initialized")
	}
	for _, ng := range m.nodeGroups {
		if ng.Id() == nodePoolUID {
			klog.V(4).Infof("Found node group for nodepool UID %s: %s", ng.(*CoreWeaveNodeGroup).Name, nodePoolUID)
			return ng, nil
		}
	}
	// If no node group is found, return nil (not an error) - this handles the case where
	// nodes exist with CoreWeave labels but no corresponding node pools are configured
	klog.V(4).Infof("No node group found for nodepool UID %s", nodePoolUID)
	return nil, nil
}
